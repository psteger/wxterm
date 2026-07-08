package vmap

import (
	"os"
	"strings"
	"testing"
)

func loadFixtureTile(t *testing.T) (*Styler, *Tile) {
	t.Helper()
	styler, err := NewStyler()
	if err != nil {
		t.Fatalf("NewStyler: %v", err)
	}
	data, err := os.ReadFile("testdata/ofm_6_18_24.pbf")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	tile, err := loadTile(styler, data)
	if err != nil {
		t.Fatalf("loadTile: %v", err)
	}
	return styler, tile
}

func TestHex2RGB(t *testing.T) {
	rgb, err := hex2rgb("#fff")
	if err != nil || rgb != [3]int{255, 255, 255} {
		t.Errorf("#fff -> %v, %v", rgb, err)
	}
	rgb, err = hex2rgb("#12ab34")
	if err != nil || rgb != [3]int{0x12, 0xab, 0x34} {
		t.Errorf("#12ab34 -> %v, %v", rgb, err)
	}
}

func TestX256(t *testing.T) {
	if got := x256([3]int{255, 0, 0}); got != 196 {
		t.Errorf("red -> %d, want 196", got)
	}
	if got := x256([3]int{0, 0, 0}); got != 16 {
		t.Errorf("black -> %d, want 16", got)
	}
	if got := x256([3]int{255, 255, 255}); got != 231 {
		t.Errorf("white -> %d, want 231", got)
	}
}

func TestLabelBufferCollision(t *testing.T) {
	lb := labelBuffer{}
	if !lb.writeIfPossible("Philadelphia", 40, 40, 5) {
		t.Fatal("first label should place")
	}
	if lb.writeIfPossible("Camden", 44, 42, 5) {
		t.Error("overlapping label should be rejected")
	}
	if !lb.writeIfPossible("Pittsburgh", 40, 200, 5) {
		t.Error("distant label should place")
	}
}

// A named label appearing in several tiles (e.g. a lake name in every
// tile the lake spans) must draw once per frame; generic poi markers are
// never deduped. mapscii intended this but its guard tested the truthy
// poiMarker constant, so the dedupe never fired.
func TestSymbolDedupeAcrossTiles(t *testing.T) {
	r := &Renderer{}
	st := &StyleLayer{Type: "symbol"}
	near := &positionedTile{posX: 0, posY: 0, size: 256, zoom: 6}
	far := &positionedTile{posX: 60, posY: 40, size: 256, zoom: 6}

	lake := &feature{layer: "water_name", style: st, label: "Lake Erie", color: 45,
		points: [][]point{{{x: 800, y: 400}}}}
	cv := newCanvas(160, 96)
	labels := labelBuffer{}
	r.drawFeature(cv, &labels, near, lake, 16, 160, 96)
	r.drawFeature(cv, &labels, far, lake, 16, 160, 96)
	if got := strings.Count(cv.frame(), "Lake Erie"); got != 1 {
		t.Errorf("named label drawn %d times, want 1", got)
	}

	marker := &feature{layer: "housenumber", style: st, color: 45,
		points: [][]point{{{x: 800, y: 400}}}}
	cv = newCanvas(160, 96)
	labels = labelBuffer{}
	r.drawFeature(cv, &labels, near, marker, 16, 160, 96)
	r.drawFeature(cv, &labels, far, marker, 16, 160, 96)
	if got := strings.Count(cv.frame(), poiMarker); got != 2 {
		t.Errorf("generic marker drawn %d times, want 2", got)
	}
}

// The mid-Atlantic fixture tile (OpenFreeMap) must decode into the
// OpenMapTiles layers the style draws.
func TestMVTDecodeFixture(t *testing.T) {
	_, tile := loadFixtureTile(t)
	for _, name := range []string{"water", "boundary", "place"} {
		layer, ok := tile.layers[name]
		if !ok {
			t.Errorf("layer %q missing (got %d layers)", name, len(tile.layers))
			continue
		}
		if len(layer.features) == 0 {
			t.Errorf("layer %q has no styled features", name)
		}
	}
}

func TestFillFeatureBoundsIncludeAllRings(t *testing.T) {
	f := &feature{}
	rings := [][]point{
		{
			{x: 1000, y: 1000},
			{x: 1100, y: 1000},
			{x: 1100, y: 1100},
			{x: 1000, y: 1100},
			{x: 1000, y: 1000},
		},
		{
			{x: 10, y: 20},
			{x: 30, y: 20},
			{x: 30, y: 40},
			{x: 10, y: 40},
			{x: 10, y: 20},
		},
	}

	f.computeBoundsForRings(rings)

	if !f.intersects(0, 0, 40, 50) {
		t.Fatal("fill feature bounds should include visible rings beyond the first ring")
	}
}

// Drawing the fixture tile centered on Philadelphia must produce text
// labels as real characters and braille geometry — with no rasterized
// text artifacts possible, since only vector features are drawn.
func TestDrawFixture(t *testing.T) {
	_, tile := loadFixtureTile(t)
	r := &Renderer{}
	var err error
	r.ts, err = NewTileSource()
	if err != nil {
		t.Fatalf("NewTileSource: %v", err)
	}

	ts := &TileSet{
		Lat: 39.9526, Lon: -75.1652, Zoom: 6,
		tiles: []fetchedTile{{rawX: 18, rawY: 24, tile: tile}},
	}
	frame := r.Draw(ts, 160, 96).Render()

	if !strings.Contains(frame, "Philadelphia") {
		t.Error("expected 'Philadelphia' place label in frame")
	}
	hasBraille := false
	for _, ch := range frame {
		if ch > 0x2800 && ch <= 0x28FF {
			hasBraille = true
			break
		}
	}
	if !hasBraille {
		t.Error("expected braille geometry in frame")
	}
}
