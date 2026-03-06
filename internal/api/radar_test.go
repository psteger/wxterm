package api

import (
	"math"
	"testing"
)

func TestLatLonToTileXY(t *testing.T) {
	tests := []struct {
		name       string
		lat, lon   float64
		zoom       int
		wantX      float64
		wantY      float64
		tolerance  float64
	}{
		{"Origin at zoom 0", 0, 0, 0, 0.5, 0.5, 0.01},
		{"NYC at zoom 5", 40.7128, -74.006, 5, 9.49, 12.07, 0.1},
		{"London at zoom 5", 51.5074, -0.1278, 5, 15.99, 10.63, 0.1},
		{"Sydney at zoom 5", -33.8688, 151.2093, 5, 29.44, 19.20, 0.1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y := LatLonToTileXY(tc.lat, tc.lon, tc.zoom)
			if math.Abs(x-tc.wantX) > tc.tolerance || math.Abs(y-tc.wantY) > tc.tolerance {
				t.Errorf("LatLonToTileXY(%f, %f, %d) = (%f, %f), want ~(%f, %f)",
					tc.lat, tc.lon, tc.zoom, x, y, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestTileXYToLatLonRoundtrip(t *testing.T) {
	tests := []struct {
		lat, lon float64
		zoom     int
	}{
		{40.7128, -74.006, 5},
		{51.5074, -0.1278, 8},
		{35.6762, 139.6503, 6},
		{-33.8688, 151.2093, 7},
		{0, 0, 4},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			x, y := LatLonToTileXY(tc.lat, tc.lon, tc.zoom)
			lat, lon := TileXYToLatLon(x, y, tc.zoom)

			if math.Abs(lat-tc.lat) > 0.0001 || math.Abs(lon-tc.lon) > 0.0001 {
				t.Errorf("Roundtrip (%f, %f) -> tile (%f, %f) -> (%f, %f)",
					tc.lat, tc.lon, x, y, lat, lon)
			}
		})
	}
}

func TestPanCenter(t *testing.T) {
	lat, lon := 40.0, -74.0
	zoom := 6

	// Pan right should increase longitude
	_, newLon := PanCenter(lat, lon, zoom, 128, 0)
	if newLon <= lon {
		t.Errorf("Pan right: expected lon > %f, got %f", lon, newLon)
	}

	// Pan down should decrease latitude (south)
	newLat, _ := PanCenter(lat, lon, zoom, 0, 128)
	if newLat >= lat {
		t.Errorf("Pan down: expected lat < %f, got %f", lat, newLat)
	}

	// Pan left should decrease longitude
	_, newLon = PanCenter(lat, lon, zoom, -128, 0)
	if newLon >= lon {
		t.Errorf("Pan left: expected lon < %f, got %f", lon, newLon)
	}

	// Pan up should increase latitude (north)
	newLat, _ = PanCenter(lat, lon, zoom, 0, -128)
	if newLat <= lat {
		t.Errorf("Pan up: expected lat > %f, got %f", lat, newLat)
	}
}

func TestTileCacheGetSet(t *testing.T) {
	cache := NewTileCache()

	_, ok := cache.Get("test/1/2/3")
	if ok {
		t.Error("Expected cache miss for new key")
	}

	// We can't easily create an image.Image in a test without more imports,
	// but we can verify the cache doesn't panic on nil
	cache.Set("test/1/2/3", nil)
	img, ok := cache.Get("test/1/2/3")
	if !ok {
		t.Error("Expected cache hit after set")
	}
	if img != nil {
		t.Error("Expected nil image from cache")
	}
}
