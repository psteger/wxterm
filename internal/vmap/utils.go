// Package vmap is a Go port of the mapscii terminal vector-tile renderer
// (https://github.com/rastapasta/mapscii, MIT license by Michael
// Strassburger). It draws Mapbox vector tiles as braille dots and text
// labels on a terminal character grid, following mapscii's pipeline:
// TileSource -> Tile -> Styler -> Renderer -> Canvas -> BrailleBuffer.
package vmap

import (
	"fmt"
	"math"
)

// Constants mirroring mapscii's config.js
const (
	tileRange   = 14
	projectSize = 256
	labelMargin = 5
	poiMarker   = "◉"
	language    = "en"
	tilePadding = 64
)

// layerConfig mirrors config.layers in mapscii's config.js
type layerConfig struct {
	margin  int
	cluster bool
}

// Keyed by OpenMapTiles source-layer names (mapscii's config uses the
// v7 equivalents housenum_label / poi_label / place_label).
var layerConfigs = map[string]layerConfig{
	"housenumber": {margin: 4},
	"poi":         {cluster: true, margin: 5},
	"place":       {cluster: true},
}

// baseZoom mirrors utils.baseZoom
func baseZoom(zoom float64) int {
	return int(math.Min(tileRange, math.Max(0, math.Floor(zoom))))
}

// tilesizeAtZoom mirrors utils.tilesizeAtZoom
func tilesizeAtZoom(zoom float64) float64 {
	return projectSize * math.Exp2(zoom-float64(baseZoom(zoom)))
}

// ll2tile mirrors utils.ll2tile (Web Mercator)
func ll2tile(lon, lat float64, zoom int) (x, y float64) {
	n := math.Exp2(float64(zoom))
	x = (lon + 180) / 360 * n
	latRad := lat * math.Pi / 180
	y = (1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2 * n
	return
}

// hex2rgb mirrors utils.hex2rgb
func hex2rgb(color string) ([3]int, error) {
	if len(color) == 0 || color[0] != '#' {
		return [3]int{255, 0, 0}, fmt.Errorf("%q isn't a supported hex color", color)
	}
	hexPart := color[1:]
	var v int
	if _, err := fmt.Sscanf(hexPart, "%x", &v); err != nil {
		return [3]int{255, 0, 0}, fmt.Errorf("%q isn't a supported hex color", color)
	}
	if len(hexPart) == 3 {
		r := (v >> 8) & 15
		g := (v >> 4) & 15
		b := v & 15
		return [3]int{r + r<<4, g + g<<4, b + b<<4}, nil
	}
	return [3]int{(v >> 16) & 255, (v >> 8) & 255, v & 255}, nil
}

// x256 finds the nearest xterm-256 color code for an RGB triple, like the
// x256 npm module mapscii uses. The 16 system colors are skipped because
// terminals theme them unpredictably; the 6x6x6 cube and grayscale ramp
// cover the same space.
func x256(rgb [3]int) uint8 {
	best := -1
	bestDist := math.MaxFloat64
	check := func(code int, r, g, b int) {
		dr := float64(rgb[0] - r)
		dg := float64(rgb[1] - g)
		db := float64(rgb[2] - b)
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			bestDist = d
			best = code
		}
	}
	levels := [6]int{0, 95, 135, 175, 215, 255}
	for ri, r := range levels {
		for gi, g := range levels {
			for bi, b := range levels {
				check(16+36*ri+6*gi+bi, r, g, b)
			}
		}
	}
	for i := 0; i < 24; i++ {
		v := 8 + i*10
		check(232+i, v, v, v)
	}
	return uint8(best & 0xff)
}
