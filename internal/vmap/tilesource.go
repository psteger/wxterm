package vmap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// tileJSONURL is OpenFreeMap's TileJSON endpoint. The actual tile URL
// template inside it contains a dated path that changes with each weekly
// data update, so it must be resolved at runtime rather than hardcoded.
const tileJSONURL = "https://tiles.openfreemap.org/planet"

const maxTileCacheEntries = 64

// TileSource fetches, decodes, and caches vector tiles, mirroring
// mapscii's TileSource in HTTP mode — pointed at OpenFreeMap
// (OpenMapTiles schema, weekly OSM updates, no API key).
type TileSource struct {
	styler     *Styler
	httpClient *http.Client

	tmplOnce sync.Once
	tmpl     string
	tmplErr  error

	mu    sync.Mutex
	cache map[string]*Tile
}

// tileURLTemplate resolves the {z}/{x}/{y} tile URL template from the
// TileJSON document, once per process.
func (ts *TileSource) tileURLTemplate() (string, error) {
	ts.tmplOnce.Do(func() {
		req, err := http.NewRequest("GET", tileJSONURL, nil)
		if err != nil {
			ts.tmplErr = err
			return
		}
		req.Header.Set("User-Agent", "wxterm/1.0 (terminal weather app)")
		resp, err := ts.httpClient.Do(req)
		if err != nil {
			ts.tmplErr = err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			ts.tmplErr = fmt.Errorf("HTTP %d for TileJSON %s", resp.StatusCode, tileJSONURL)
			return
		}
		var doc struct {
			Tiles []string `json:"tiles"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			ts.tmplErr = fmt.Errorf("parse TileJSON: %w", err)
			return
		}
		if len(doc.Tiles) == 0 {
			ts.tmplErr = fmt.Errorf("TileJSON %s has no tile URLs", tileJSONURL)
			return
		}
		ts.tmpl = doc.Tiles[0]
	})
	return ts.tmpl, ts.tmplErr
}

// NewTileSource creates a tile source with the embedded mapscii style.
func NewTileSource() (*TileSource, error) {
	styler, err := NewStyler()
	if err != nil {
		return nil, err
	}
	return &TileSource{
		styler:     styler,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cache:      map[string]*Tile{},
	}, nil
}

// Styler exposes the compiled style (for the background color).
func (ts *TileSource) Styler() *Styler { return ts.styler }

// GetTile fetches and decodes one vector tile, served from cache when
// possible.
func (ts *TileSource) GetTile(z, x, y int) (*Tile, error) {
	key := fmt.Sprintf("%d/%d/%d", z, x, y)
	ts.mu.Lock()
	if tile, ok := ts.cache[key]; ok {
		ts.mu.Unlock()
		return tile, nil
	}
	ts.mu.Unlock()

	tmpl, err := ts.tileURLTemplate()
	if err != nil {
		return nil, err
	}
	url := strings.NewReplacer(
		"{z}", fmt.Sprint(z), "{x}", fmt.Sprint(x), "{y}", fmt.Sprint(y),
	).Replace(tmpl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wxterm/1.0 (terminal weather app)")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for vector tile %s", resp.StatusCode, key)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tile, err := loadTile(ts.styler, data)
	if err != nil {
		return nil, err
	}

	ts.mu.Lock()
	if len(ts.cache) >= maxTileCacheEntries {
		clear(ts.cache)
	}
	ts.cache[key] = tile
	ts.mu.Unlock()
	return tile, nil
}
