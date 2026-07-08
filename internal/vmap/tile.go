package vmap

import (
	"bytes"
	"compress/gzip"
	"io"
)

// feature is a drawable feature with its resolved style, mirroring the
// nodes mapscii's Tile._loadLayers builds.
type feature struct {
	layer string
	style *StyleLayer
	label string
	color uint8
	// fill features keep all rings; line/symbol features hold one part each
	points [][]point
	// bounding box in tile-geometry coordinates (from the outer ring)
	minX, minY, maxX, maxY int
}

// tileLayer holds a layer's styled features and geometry extent.
type tileLayer struct {
	extent   float64
	features []*feature
}

// Tile is a decoded, styled vector tile.
type Tile struct {
	layers map[string]*tileLayer
}

// loadTile mirrors mapscii's Tile.load: gunzip if needed, decode the
// protobuf, then keep only features that match a style, annotated with
// color, label, and bounding box.
func loadTile(styler *Styler, data []byte) (*Tile, error) {
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		unzipped, err := io.ReadAll(zr)
		if cerr := zr.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return nil, err
		}
		data = unzipped
	}

	raw, err := decodeMVT(data)
	if err != nil {
		return nil, err
	}

	geomTypes := [4]string{"", "Point", "LineString", "Polygon"}
	colorCache := map[string]uint8{}

	tile := &Tile{layers: map[string]*tileLayer{}}
	for name, layer := range raw {
		tl := &tileLayer{extent: layer.extent}
		for i := range layer.features {
			ft := &layer.features[i]
			props := layer.properties(ft)
			if ft.geomType >= 0 && ft.geomType < 4 {
				props["$type"] = geomTypes[ft.geomType]
			}

			style := styler.GetStyleFor(name, props)
			if style == nil {
				continue
			}

			colorHex := style.Color()
			code, ok := colorCache[colorHex]
			if !ok {
				rgb, err := hex2rgb(colorHex)
				if err != nil {
					rgb = [3]int{255, 0, 0}
				}
				code = x256(rgb)
				colorCache[colorHex] = code
			}

			var label string
			if style.Type == "symbol" {
				// OpenMapTiles name keys, with mapscii's v7 fallbacks
				for _, key := range []string{"name:" + language, "name_" + language,
					"name_en", "name", "name:latin", "housenumber"} {
					if v, ok := props[key].(string); ok && v != "" {
						label = v
						break
					}
				}
			}

			geometry := ft.loadGeometry()
			if len(geometry) == 0 {
				continue
			}

			if style.Type == "fill" {
				f := &feature{
					layer:  name,
					style:  style,
					label:  label,
					color:  code,
					points: geometry,
				}
				f.computeBoundsForRings(geometry)
				tl.features = append(tl.features, f)
			} else {
				// one node per line part, like mapscii
				for _, part := range geometry {
					f := &feature{
						layer:  name,
						style:  style,
						label:  label,
						color:  code,
						points: [][]point{part},
					}
					f.computeBounds(part)
					tl.features = append(tl.features, f)
				}
			}
		}
		tile.layers[name] = tl
	}
	return tile, nil
}

func (f *feature) computeBoundsForRings(rings [][]point) {
	f.minX, f.minY = 1<<30, 1<<30
	f.maxX, f.maxY = -(1 << 30), -(1 << 30)
	for _, ring := range rings {
		f.expandBounds(ring)
	}
}

// computeBounds mirrors Tile._addBoundaries
func (f *feature) computeBounds(points []point) {
	f.minX, f.minY = 1<<30, 1<<30
	f.maxX, f.maxY = -(1 << 30), -(1 << 30)
	f.expandBounds(points)
}

func (f *feature) expandBounds(points []point) {
	for _, p := range points {
		if p.x < f.minX {
			f.minX = p.x
		}
		if p.x > f.maxX {
			f.maxX = p.x
		}
		if p.y < f.minY {
			f.minY = p.y
		}
		if p.y > f.maxY {
			f.maxY = p.y
		}
	}
}

func (f *feature) intersects(minX, minY, maxX, maxY float64) bool {
	return float64(f.minX) <= maxX && float64(f.maxX) >= minX &&
		float64(f.minY) <= maxY && float64(f.maxY) >= minY
}
