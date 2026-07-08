package vmap

import (
	"math"
	"sync"
)

// Renderer is a port of mapscii's Renderer: fetches the 3x3 tile
// neighborhood around a center, then draws styled features layer by
// layer onto a braille canvas.
type Renderer struct {
	ts *TileSource
}

var (
	defaultRenderer *Renderer
	defaultErr      error
	defaultOnce     sync.Once
)

// Default returns the process-wide renderer with its shared tile cache.
func Default() (*Renderer, error) {
	defaultOnce.Do(func() {
		defaultRenderer, defaultErr = NewRenderer()
	})
	return defaultRenderer, defaultErr
}

func NewRenderer() (*Renderer, error) {
	ts, err := NewTileSource()
	if err != nil {
		return nil, err
	}
	return &Renderer{ts: ts}, nil
}

// fetchedTile pairs a decoded tile with its raw (unwrapped) grid
// coordinates, so positions can be computed for any view size at draw time.
type fetchedTile struct {
	rawX, rawY int
	tile       *Tile
}

// TileSet holds the prefetched tiles for one radar viewport.
type TileSet struct {
	Lat, Lon float64
	Zoom     int
	tiles    []fetchedTile
}

// FetchTiles downloads and decodes the 3x3 tile neighborhood around the
// center, mirroring Renderer._visibleTiles + _getTile. Blocking — call it
// from an async command, then Draw with the result.
func (r *Renderer) FetchTiles(lat, lon float64, zoom int) (*TileSet, error) {
	z := baseZoom(float64(zoom))
	centerX, centerY := ll2tile(lon, lat, z)
	gridSize := 1 << z

	type job struct{ rawX, rawY, fetchX, fetchY int }
	var jobs []job
	for rawY := int(math.Floor(centerY)) - 1; rawY <= int(math.Floor(centerY))+1; rawY++ {
		if rawY < 0 || rawY >= gridSize {
			continue
		}
		for rawX := int(math.Floor(centerX)) - 1; rawX <= int(math.Floor(centerX))+1; rawX++ {
			fetchX := ((rawX % gridSize) + gridSize) % gridSize
			jobs = append(jobs, job{rawX, rawY, fetchX, rawY})
		}
	}

	ts := &TileSet{Lat: lat, Lon: lon, Zoom: zoom}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			tile, err := r.ts.GetTile(z, j.fetchX, j.fetchY)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			ts.tiles = append(ts.tiles, fetchedTile{rawX: j.rawX, rawY: j.rawY, tile: tile})
		}(j)
	}
	wg.Wait()

	if len(ts.tiles) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return ts, nil
}

// Frame is a drawn map the caller can overlay cell backgrounds onto
// (rain colors) before rendering to a string.
type Frame struct {
	canvas *canvas
}

// SetCellBackgroundRGB sets a terminal cell's background to the nearest
// xterm-256 color.
func (f *Frame) SetCellBackgroundRGB(cellX, cellY int, r, g, b uint8) {
	f.canvas.background(cellX*2, cellY*4, x256([3]int{int(r), int(g), int(b)}))
}

// Render composes the final ANSI string.
func (f *Frame) Render() string {
	return f.canvas.frame()
}

// positionedTile mirrors the descriptors _visibleTiles builds: a tile
// plus its pixel position in the current view.
type positionedTile struct {
	tile       *Tile
	posX, posY float64
	size       float64
	zoom       float64
}

// Draw renders the tile set onto a widthPx x heightPx braille canvas,
// mirroring Renderer.draw/_renderTiles/_drawFeature.
func (r *Renderer) Draw(ts *TileSet, widthPx, heightPx int) *Frame {
	cv := newCanvas(widthPx, heightPx)
	if bg := r.ts.Styler().BackgroundColor(); bg != "" {
		if rgb, err := hex2rgb(bg); err == nil {
			cv.setBackground(x256(rgb))
		}
	}

	if ts == nil {
		return &Frame{canvas: cv}
	}

	zoom := float64(ts.Zoom)
	z := baseZoom(zoom)
	centerX, centerY := ll2tile(ts.Lon, ts.Lat, z)
	tileSize := tilesizeAtZoom(zoom)

	var tiles []positionedTile
	for _, ft := range ts.tiles {
		posX := float64(widthPx)/2 - (centerX-float64(ft.rawX))*tileSize
		posY := float64(heightPx)/2 - (centerY-float64(ft.rawY))*tileSize
		if posX+tileSize < 0 || posY+tileSize < 0 ||
			posX > float64(widthPx) || posY > float64(heightPx) {
			continue
		}
		tiles = append(tiles, positionedTile{
			tile: ft.tile, posX: posX, posY: posY, size: tileSize, zoom: zoom,
		})
	}

	labels := labelBuffer{}
	type pendingLabel struct {
		tile    *positionedTile
		feature *feature
		scale   float64
	}
	var pendingLabels []pendingLabel

	// mapscii sorts collected labels with a comparator that always
	// returns NaN (a.feature.sorty is a typo), so effective order is
	// collection order — preserved here.
	for _, layerID := range drawOrder(zoom) {
		for i := range tiles {
			pt := &tiles[i]
			layer, ok := pt.tile.layers[layerID]
			if !ok {
				continue
			}
			scale := layer.extent / tilesizeAtZoom(pt.zoom)
			minX := -pt.posX * scale
			minY := -pt.posY * scale
			maxX := (float64(widthPx) - pt.posX) * scale
			maxY := (float64(heightPx) - pt.posY) * scale
			for _, ft := range layer.features {
				if !ft.intersects(minX, minY, maxX, maxY) {
					continue
				}
				if ft.style.Type == "symbol" {
					pendingLabels = append(pendingLabels, pendingLabel{pt, ft, scale})
				} else {
					r.drawFeature(cv, &labels, pt, ft, scale, widthPx, heightPx)
				}
			}
		}
	}
	for _, pl := range pendingLabels {
		r.drawFeature(cv, &labels, pl.tile, pl.feature, pl.scale, widthPx, heightPx)
	}

	return &Frame{canvas: cv}
}

// drawOrder mirrors Renderer._generateDrawOrder, with mapscii's v7 layer
// names translated to their OpenMapTiles equivalents. Symbol features are
// deferred to the label pass by style type (v7 label layers had "label"
// in their names; OpenMapTiles ones don't), so this ordering governs
// geometry stacking and label placement priority.
func drawOrder(zoom float64) []string {
	if zoom < 2 {
		return []string{"boundary", "water", "place", "water_name"}
	}
	return []string{
		"landcover", "landuse", "park", "water", "water_name", "building",
		"transportation", "boundary", "place", "aerodrome_label", "poi",
		"transportation_name", "housenumber",
	}
}

// drawFeature mirrors Renderer._drawFeature
func (r *Renderer) drawFeature(cv *canvas, labels *labelBuffer, pt *positionedTile,
	ft *feature, scale float64, widthPx, heightPx int) {

	if ft.style.MinZoom != 0 && pt.zoom < ft.style.MinZoom {
		return
	}
	if ft.style.MaxZoom != 0 && pt.zoom > ft.style.MaxZoom {
		return
	}

	switch ft.style.Type {
	case "line":
		width := ft.style.LineWidth()
		points := scaleAndReduce(pt, ft.points[0], scale, widthPx, heightPx, true, false)
		if len(points) > 0 {
			cv.polyline(points, ft.color, width)
		}
	case "fill":
		rings := make([][]point, 0, len(ft.points))
		for _, ring := range ft.points {
			rings = append(rings, scaleAndReduce(pt, ring, scale, widthPx, heightPx, false, false))
		}
		cv.polygon(rings, ft.color)
	case "symbol":
		text := ft.label
		if text == "" {
			text = poiMarker
		}
		// mapscii intends each named label to draw once per frame, but its
		// guard (`this._seen[text] && !genericSymbol`) tests the poiMarker
		// constant — always truthy — so the dedupe never fires and labels
		// spanning tiles repeat. Apply the intended check: dedupe names,
		// never generic markers.
		if text != poiMarker && labels.placed(text) {
			return
		}
		points := scaleAndReduce(pt, ft.points[0], scale, widthPx, heightPx, true, true)
		lc := layerConfigs[ft.layer]
		margin := lc.margin
		if margin == 0 {
			margin = labelMargin
		}
		for _, p := range points {
			x := p.x - len([]rune(text))
			if labels.writeIfPossible(text, x, p.y, margin) {
				cv.text(text, x, p.y, ft.color)
				if text != poiMarker {
					labels.markPlaced(text)
				}
				break
			} else if lc.cluster && labels.writeIfPossible(poiMarker, p.x, p.y, 3) {
				cv.text(poiMarker, p.x, p.y, ft.color)
				if text != poiMarker {
					labels.markPlaced(text)
				}
				break
			}
		}
	}
}

// scaleAndReduce mirrors Renderer._scaleAndReduce: scale geometry into
// view pixels, drop consecutive duplicates, and collapse runs of points
// outside the padded viewport.
func scaleAndReduce(pt *positionedTile, points []point, scale float64,
	widthPx, heightPx int, filter bool, symbol bool) []point {

	minX := -tilePadding
	minY := -tilePadding
	maxX := widthPx + tilePadding
	maxY := heightPx + tilePadding

	var scaled []point
	var lastX, lastY int
	first := true
	outside := false

	for _, p := range points {
		x := int(math.Floor(pt.posX + float64(p.x)/scale))
		y := int(math.Floor(pt.posY + float64(p.y)/scale))
		if !first && lastX == x && lastY == y {
			continue
		}
		first = false
		lastX, lastY = x, y
		if filter {
			if x < minX || x > maxX || y < minY || y > maxY {
				if outside {
					continue
				}
				outside = true
			} else if outside {
				outside = false
				scaled = append(scaled, point{lastX, lastY})
			}
		}
		scaled = append(scaled, point{x, y})
	}

	if !symbol && len(scaled) < 2 {
		return nil
	}
	return scaled
}
