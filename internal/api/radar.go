package api

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/http"
	"sync"
	"time"
)

// Ensure png format is available
var _ = png.Decode

// LatLonToTileXY converts lat/lon to fractional tile coordinates at given zoom
func LatLonToTileXY(lat, lon float64, zoom int) (x, y float64) {
	n := math.Pow(2, float64(zoom))
	x = (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y = (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return
}

// TileXYToLatLon converts fractional tile coordinates back to lat/lon
func TileXYToLatLon(x, y float64, zoom int) (lat, lon float64) {
	n := math.Pow(2, float64(zoom))
	lon = x/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*y/n)))
	lat = latRad * 180.0 / math.Pi
	return
}

// PanCenter shifts a lat/lon center by the given pixel deltas at the specified zoom level
func PanCenter(lat, lon float64, zoom int, dxPixels, dyPixels float64) (newLat, newLon float64) {
	tx, ty := LatLonToTileXY(lat, lon, zoom)
	tx += dxPixels / float64(tileSize)
	ty += dyPixels / float64(tileSize)
	return TileXYToLatLon(tx, ty, zoom)
}

// RadarData holds the map and rain data for the radar view
type RadarData struct {
	MapImage   image.Image // Composited map tiles
	RainFrames []RainFrame // Rain overlay frames for animation

	CenterLat float64
	CenterLon float64
	ZoomLevel int

	// Center pixel position within MapImage
	CenterPX int
	CenterPY int
}

// RainFrame contains composited rain tile data for one timestamp
type RainFrame struct {
	Image     image.Image
	Timestamp time.Time
}

// RainViewerResponse is the JSON response from RainViewer weather maps API
type RainViewerResponse struct {
	Version   string `json:"version"`
	Generated int64  `json:"generated"`
	Host      string `json:"host"`
	Radar     struct {
		Past    []RainViewerEntry `json:"past"`
		Nowcast []RainViewerEntry `json:"nowcast"`
	} `json:"radar"`
}

// RainViewerEntry represents a single radar timestamp
type RainViewerEntry struct {
	Time int64  `json:"time"`
	Path string `json:"path"`
}

// TileCache provides thread-safe caching of tile images
type TileCache struct {
	mu    sync.RWMutex
	tiles map[string]image.Image
}

// NewTileCache creates a new tile cache
func NewTileCache() *TileCache {
	return &TileCache{tiles: make(map[string]image.Image)}
}

// Get retrieves a cached tile
func (tc *TileCache) Get(key string) (image.Image, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	img, ok := tc.tiles[key]
	return img, ok
}

// Set stores a tile in the cache, evicting all entries if the cache is too large.
func (tc *TileCache) Set(key string, img image.Image) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if len(tc.tiles) >= maxCacheEntries {
		clear(tc.tiles)
	}
	tc.tiles[key] = img
}

const (
	osmTileURL        = "https://tile.openstreetmap.org/%d/%d/%d.png"
	rainViewerMapsURL = "https://api.rainviewer.com/public/weather-maps.json"
	tileSize          = 256
	maxRainZoom       = 7     // RainViewer supports up to zoom 7
	rainColorScheme   = 6     // NEXRAD Level III
	rainOptions       = "1_1" // smooth + snow
	maxCacheEntries   = 512   // evict all when exceeded
)

// tileGrid holds the computed tile range and composite dimensions for a viewport.
type tileGrid struct {
	minX, minY, maxX, maxY int
	compW, compH           int
	numTiles               int
}

// calcTileGrid determines which tiles cover the given pixel bounds at the specified zoom.
func calcTileGrid(left, top, right, bottom float64, zoom int) tileGrid {
	minX := int(math.Floor(left / float64(tileSize)))
	minY := int(math.Floor(top / float64(tileSize)))
	maxX := int(math.Floor((right - 1) / float64(tileSize)))
	maxY := int(math.Floor((bottom - 1) / float64(tileSize)))

	maxTile := (1 << zoom) - 1
	if minY < 0 {
		minY = 0
	}
	if maxY > maxTile {
		maxY = maxTile
	}

	return tileGrid{
		minX: minX, minY: minY, maxX: maxX, maxY: maxY,
		compW:    (maxX - minX + 1) * tileSize,
		compH:    (maxY - minY + 1) * tileSize,
		numTiles: maxTile + 1,
	}
}

// wrapTileX wraps a tile X coordinate into the valid range [0, numTiles).
func wrapTileX(tx, numTiles int) int {
	return ((tx % numTiles) + numTiles) % numTiles
}

// FetchRadar fetches map tiles and rain overlay data for the radar view
func (c *Client) FetchRadar(lat, lon float64, zoom int, viewWidthChars, viewHeightChars int) (*RadarData, error) {
	// Each braille character covers 2 pixels wide, 4 pixels tall
	bpxWidth := viewWidthChars * 2
	bpxHeight := viewHeightChars * 4

	// Find center position in tile-pixel space
	centerTX, centerTY := LatLonToTileXY(lat, lon, zoom)
	centerPXf := centerTX * float64(tileSize)
	centerPYf := centerTY * float64(tileSize)

	// Calculate the tile-pixel bounds of our view
	left := centerPXf - float64(bpxWidth)/2
	top := centerPYf - float64(bpxHeight)/2
	right := centerPXf + float64(bpxWidth)/2
	bottom := centerPYf + float64(bpxHeight)/2

	// Determine which tiles we need
	grid := calcTileGrid(left, top, right, bottom, zoom)

	// Create map composite with neutral background
	mapComposite := image.NewRGBA(image.Rect(0, 0, grid.compW, grid.compH))
	draw.Draw(mapComposite, mapComposite.Bounds(),
		&image.Uniform{color.RGBA{240, 238, 233, 255}}, image.Point{}, draw.Src)

	// Fetch map tiles in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for ty := grid.minY; ty <= grid.maxY; ty++ {
		for tx := grid.minX; tx <= grid.maxX; tx++ {
			wg.Add(1)
			go func(tx, ty int) {
				defer wg.Done()
				img, err := c.fetchMapTile(zoom, wrapTileX(tx, grid.numTiles), ty)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				destX := (tx - grid.minX) * tileSize
				destY := (ty - grid.minY) * tileSize
				mu.Lock()
				draw.Draw(mapComposite,
					image.Rect(destX, destY, destX+tileSize, destY+tileSize),
					img, img.Bounds().Min, draw.Src)
				mu.Unlock()
			}(tx, ty)
		}
	}
	wg.Wait()

	if firstErr != nil {
		return nil, fmt.Errorf("failed to fetch map tiles: %w", firstErr)
	}

	// Calculate center pixel in composite coordinates
	centerPX := int(centerPXf - float64(grid.minX)*float64(tileSize))
	centerPY := int(centerPYf - float64(grid.minY)*float64(tileSize))

	result := &RadarData{
		MapImage:  mapComposite,
		CenterLat: lat,
		CenterLon: lon,
		ZoomLevel: zoom,
		CenterPX:  centerPX,
		CenterPY:  centerPY,
	}

	// Fetch RainViewer data (non-fatal if it fails)
	rainResp, err := c.fetchRainViewerMaps()
	if err != nil {
		return result, nil
	}

	// Collect timestamps (past + nowcast)
	var entries []RainViewerEntry
	entries = append(entries, rainResp.Radar.Past...)
	entries = append(entries, rainResp.Radar.Nowcast...)

	maxFrames := 6
	if len(entries) > maxFrames {
		entries = entries[len(entries)-maxFrames:]
	}

	// RainViewer supports up to zoom 7; for higher zooms, fetch at 7 and upscale
	rainZoom := zoom
	if rainZoom > maxRainZoom {
		rainZoom = maxRainZoom
	}
	needsUpscale := rainZoom < zoom

	// Calculate rain tile grid (may differ from map tile grid at high zoom)
	var rainGrid tileGrid
	if !needsUpscale {
		rainGrid = grid
	} else {
		scaleFactor := math.Pow(2, float64(zoom-rainZoom))
		rainGrid = calcTileGrid(left/scaleFactor, top/scaleFactor, right/scaleFactor, bottom/scaleFactor, rainZoom)
	}

	// Fetch rain tiles for each timestamp
	for _, entry := range entries {
		rainSmall := image.NewRGBA(image.Rect(0, 0, rainGrid.compW, rainGrid.compH))

		var rwg sync.WaitGroup
		for ty := rainGrid.minY; ty <= rainGrid.maxY; ty++ {
			for tx := rainGrid.minX; tx <= rainGrid.maxX; tx++ {
				rwg.Add(1)
				go func(tx, ty int) {
					defer rwg.Done()
					img, err := c.fetchRainTile(rainResp.Host, entry.Path, rainZoom, wrapTileX(tx, rainGrid.numTiles), ty)
					if err != nil {
						return // Skip failed rain tiles
					}
					destX := (tx - rainGrid.minX) * tileSize
					destY := (ty - rainGrid.minY) * tileSize
					mu.Lock()
					draw.Draw(rainSmall,
						image.Rect(destX, destY, destX+tileSize, destY+tileSize),
						img, img.Bounds().Min, draw.Over)
					mu.Unlock()
				}(tx, ty)
			}
		}
		rwg.Wait()

		var rainFrame image.Image
		if needsUpscale {
			rainFrame = upscaleRainToMapSpace(rainSmall, grid, rainGrid, zoom, rainZoom)
		} else {
			rainFrame = rainSmall
		}

		result.RainFrames = append(result.RainFrames, RainFrame{
			Image:     rainFrame,
			Timestamp: time.Unix(entry.Time, 0),
		})
	}

	return result, nil
}

// upscaleRainToMapSpace scales a rain composite fetched at rainZoom into the
// pixel space of the map composite at mapZoom using nearest-neighbor interpolation.
func upscaleRainToMapSpace(rain *image.RGBA, mapGrid, rainGrid tileGrid, mapZoom, rainZoom int) *image.RGBA {
	shift := uint(mapZoom - rainZoom)
	out := image.NewRGBA(image.Rect(0, 0, mapGrid.compW, mapGrid.compH))
	rainW := rain.Bounds().Dx()
	rainH := rain.Bounds().Dy()
	rainBaseX := rainGrid.minX * tileSize
	rainBaseY := rainGrid.minY * tileSize

	for y := 0; y < mapGrid.compH; y++ {
		srcY := ((mapGrid.minY*tileSize + y) >> shift) - rainBaseY
		if srcY < 0 || srcY >= rainH {
			continue
		}
		srcRowOff := srcY * rain.Stride
		dstRowOff := y * out.Stride
		for x := 0; x < mapGrid.compW; x++ {
			srcX := ((mapGrid.minX*tileSize + x) >> shift) - rainBaseX
			if srcX < 0 || srcX >= rainW {
				continue
			}
			srcOff := srcRowOff + srcX*4
			dstOff := dstRowOff + x*4
			copy(out.Pix[dstOff:dstOff+4], rain.Pix[srcOff:srcOff+4])
		}
	}
	return out
}

func (c *Client) fetchMapTile(zoom, x, y int) (image.Image, error) {
	key := fmt.Sprintf("map/%d/%d/%d", zoom, x, y)
	url := fmt.Sprintf(osmTileURL, zoom, x, y)
	return c.fetchPNGTile(key, url)
}

func (c *Client) fetchRainViewerMaps() (*RainViewerResponse, error) {
	var result RainViewerResponse
	if err := c.get(rainViewerMapsURL, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) fetchRainTile(host, path string, zoom, x, y int) (image.Image, error) {
	key := fmt.Sprintf("rain/%s/%d/%d/%d", path, zoom, x, y)
	url := fmt.Sprintf("%s%s/%d/%d/%d/%d/%d/%s.png", host, path, tileSize, zoom, x, y, rainColorScheme, rainOptions)
	return c.fetchPNGTile(key, url)
}

// fetchPNGTile fetches a PNG tile image by URL with caching.
func (c *Client) fetchPNGTile(key, url string) (image.Image, error) {
	if img, ok := c.tileCache.Get(key); ok {
		return img, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wxterm/1.0 (terminal weather app)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for tile %s", resp.StatusCode, key)
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	c.tileCache.Set(key, img)
	return img, nil
}
