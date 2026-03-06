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

// Set stores a tile in the cache
func (tc *TileCache) Set(key string, img image.Image) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tiles[key] = img
}

const (
	osmTileURL        = "https://tile.openstreetmap.org/%d/%d/%d.png"
	rainViewerMapsURL = "https://api.rainviewer.com/public/weather-maps.json"
	tileSize          = 256
	maxRainZoom       = 7 // RainViewer supports up to zoom 7
)

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
	tileMinX := int(math.Floor(left / float64(tileSize)))
	tileMinY := int(math.Floor(top / float64(tileSize)))
	tileMaxX := int(math.Floor((right - 1) / float64(tileSize)))
	tileMaxY := int(math.Floor((bottom - 1) / float64(tileSize)))

	// Clamp Y to valid range (latitude doesn't wrap)
	maxTile := (1 << zoom) - 1
	if tileMinY < 0 {
		tileMinY = 0
	}
	if tileMaxY > maxTile {
		tileMaxY = maxTile
	}

	cols := tileMaxX - tileMinX + 1
	rows := tileMaxY - tileMinY + 1
	compositeWidth := cols * tileSize
	compositeHeight := rows * tileSize

	// Create map composite with neutral background
	mapComposite := image.NewRGBA(image.Rect(0, 0, compositeWidth, compositeHeight))
	draw.Draw(mapComposite, mapComposite.Bounds(),
		&image.Uniform{color.RGBA{240, 238, 233, 255}}, image.Point{}, draw.Src)

	// Fetch map tiles in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	numTiles := maxTile + 1

	for ty := tileMinY; ty <= tileMaxY; ty++ {
		for tx := tileMinX; tx <= tileMaxX; tx++ {
			wg.Add(1)
			go func(tx, ty int) {
				defer wg.Done()
				wrappedTX := ((tx % numTiles) + numTiles) % numTiles
				img, err := c.fetchMapTile(zoom, wrappedTX, ty)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				destX := (tx - tileMinX) * tileSize
				destY := (ty - tileMinY) * tileSize
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
	originPX := float64(tileMinX) * float64(tileSize)
	originPY := float64(tileMinY) * float64(tileSize)
	centerPX := int(centerPXf - originPX)
	centerPY := int(centerPYf - originPY)

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
	var rainTileMinX, rainTileMinY, rainTileMaxX, rainTileMaxY int
	var rainCompW, rainCompH, rainNumTiles int

	if !needsUpscale {
		rainTileMinX = tileMinX
		rainTileMinY = tileMinY
		rainTileMaxX = tileMaxX
		rainTileMaxY = tileMaxY
		rainCompW = compositeWidth
		rainCompH = compositeHeight
		rainNumTiles = numTiles
	} else {
		scaleFactor := math.Pow(2, float64(zoom-rainZoom))
		rLeft := left / scaleFactor
		rTop := top / scaleFactor
		rRight := right / scaleFactor
		rBottom := bottom / scaleFactor

		rainTileMinX = int(math.Floor(rLeft / float64(tileSize)))
		rainTileMinY = int(math.Floor(rTop / float64(tileSize)))
		rainTileMaxX = int(math.Floor((rRight - 1) / float64(tileSize)))
		rainTileMaxY = int(math.Floor((rBottom - 1) / float64(tileSize)))

		rainMaxTile := (1 << rainZoom) - 1
		rainNumTiles = rainMaxTile + 1
		if rainTileMinY < 0 {
			rainTileMinY = 0
		}
		if rainTileMaxY > rainMaxTile {
			rainTileMaxY = rainMaxTile
		}

		rainCols := rainTileMaxX - rainTileMinX + 1
		rainRows := rainTileMaxY - rainTileMinY + 1
		rainCompW = rainCols * tileSize
		rainCompH = rainRows * tileSize
	}

	// Fetch rain tiles for each timestamp
	for _, entry := range entries {
		rainSmall := image.NewRGBA(image.Rect(0, 0, rainCompW, rainCompH))

		var rwg sync.WaitGroup
		for ty := rainTileMinY; ty <= rainTileMaxY; ty++ {
			for tx := rainTileMinX; tx <= rainTileMaxX; tx++ {
				rwg.Add(1)
				go func(tx, ty int) {
					defer rwg.Done()
					wrappedTX := ((tx % rainNumTiles) + rainNumTiles) % rainNumTiles
					img, err := c.fetchRainTile(rainResp.Host, entry.Path, rainZoom, wrappedTX, ty)
					if err != nil {
						return // Skip failed rain tiles
					}
					destX := (tx - rainTileMinX) * tileSize
					destY := (ty - rainTileMinY) * tileSize
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
			rainFrame = upscaleRainToMapSpace(rainSmall,
				tileMinX, tileMinY, compositeWidth, compositeHeight,
				rainTileMinX, rainTileMinY, zoom, rainZoom)
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
func upscaleRainToMapSpace(rain *image.RGBA,
	mapTileMinX, mapTileMinY, mapW, mapH int,
	rainTileMinX, rainTileMinY, mapZoom, rainZoom int,
) *image.RGBA {
	shift := uint(mapZoom - rainZoom)
	out := image.NewRGBA(image.Rect(0, 0, mapW, mapH))
	rainW := rain.Bounds().Dx()
	rainH := rain.Bounds().Dy()
	rainBaseX := rainTileMinX * tileSize
	rainBaseY := rainTileMinY * tileSize

	for y := 0; y < mapH; y++ {
		// Global pixel Y at mapZoom, then floor-divide to rainZoom pixel space
		srcY := ((mapTileMinY*tileSize + y) >> shift) - rainBaseY
		if srcY < 0 || srcY >= rainH {
			continue
		}
		for x := 0; x < mapW; x++ {
			srcX := ((mapTileMinX*tileSize + x) >> shift) - rainBaseX
			if srcX < 0 || srcX >= rainW {
				continue
			}
			out.SetRGBA(x, y, rain.RGBAAt(srcX, srcY))
		}
	}
	return out
}

func (c *Client) fetchMapTile(zoom, x, y int) (image.Image, error) {
	key := fmt.Sprintf("map/%d/%d/%d", zoom, x, y)
	if img, ok := c.tileCache.Get(key); ok {
		return img, nil
	}

	url := fmt.Sprintf(osmTileURL, zoom, x, y)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WeatherTerm/1.0 (terminal weather app)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for tile %d/%d/%d", resp.StatusCode, zoom, x, y)
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	c.tileCache.Set(key, img)
	return img, nil
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
	if img, ok := c.tileCache.Get(key); ok {
		return img, nil
	}

	// RainViewer tile URL: {host}{path}/{size}/{z}/{x}/{y}/{color}/{options}.png
	// color 6 = NEXRAD Level III, options 1_1 = smooth + snow
	url := fmt.Sprintf("%s%s/%d/%d/%d/%d/6/1_1.png", host, path, tileSize, zoom, x, y)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for rain tile", resp.StatusCode)
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	c.tileCache.Set(key, img)
	return img, nil
}
