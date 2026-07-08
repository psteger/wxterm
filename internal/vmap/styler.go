package vmap

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// omt-dark.json is mapscii's dark style (MIT licensed, from
// https://github.com/rastapasta/mapscii) translated to the OpenMapTiles
// schema that OpenFreeMap serves.
//
//go:embed styles/omt-dark.json
var darkStyleJSON []byte

// StyleLayer is one compiled style layer, mirroring what mapscii's Styler
// attaches to features.
type StyleLayer struct {
	ID          string
	Type        string // "line", "fill", "symbol", "background"
	SourceLayer string
	MinZoom     float64 // 0 = unset
	MaxZoom     float64 // 0 = unset
	appliesTo   func(props map[string]any) bool
	paint       map[string]any
}

// Color returns the layer's paint color (line || fill || text), resolving
// zoom-stop objects to their first stop like mapscii's Tile._loadLayers.
func (s *StyleLayer) Color() string {
	for _, key := range []string{"line-color", "fill-color", "text-color", "background-color"} {
		if v, ok := s.paint[key]; ok {
			return resolveStops(v)
		}
	}
	return ""
}

// LineWidth resolves paint's line-width the way mapscii's Renderer does.
func (s *StyleLayer) LineWidth() int {
	v, ok := s.paint["line-width"]
	if !ok {
		return 1
	}
	switch w := v.(type) {
	case float64:
		return int(w)
	case map[string]any:
		if stops, ok := w["stops"].([]any); ok && len(stops) > 0 {
			if pair, ok := stops[0].([]any); ok && len(pair) > 1 {
				if f, ok := pair[1].(float64); ok {
					return int(f)
				}
			}
		}
	}
	return 1
}

func resolveStops(v any) string {
	switch c := v.(type) {
	case string:
		return c
	case map[string]any:
		if stops, ok := c["stops"].([]any); ok && len(stops) > 0 {
			if pair, ok := stops[0].([]any); ok && len(pair) > 1 {
				if s, ok := pair[1].(string); ok {
					return s
				}
			}
		}
	}
	return ""
}

// Styler is a port of mapscii's Styler: compiles the style's layer
// filters and answers "which style applies to this feature?".
type Styler struct {
	byLayer map[string][]*StyleLayer
	byID    map[string]*StyleLayer
}

// NewStyler loads and compiles the embedded mapscii dark style.
func NewStyler() (*Styler, error) {
	var doc struct {
		Constants map[string]any   `json:"constants"`
		Layers    []map[string]any `json:"layers"`
	}
	if err := json.Unmarshal(darkStyleJSON, &doc); err != nil {
		return nil, fmt.Errorf("parse style: %w", err)
	}

	for _, layer := range doc.Layers {
		replaceConstants(doc.Constants, layer)
	}

	s := &Styler{
		byLayer: map[string][]*StyleLayer{},
		byID:    map[string]*StyleLayer{},
	}

	for _, layer := range doc.Layers {
		// ref inheritance, mirroring Styler's constructor
		if ref, ok := layer["ref"].(string); ok {
			if parent, ok := s.byID[ref]; ok {
				if _, has := layer["type"]; !has {
					layer["type"] = parent.Type
				}
				if _, has := layer["source-layer"]; !has {
					layer["source-layer"] = parent.SourceLayer
				}
				if _, has := layer["minzoom"]; !has && parent.MinZoom != 0 {
					layer["minzoom"] = parent.MinZoom
				}
				if _, has := layer["maxzoom"]; !has && parent.MaxZoom != 0 {
					layer["maxzoom"] = parent.MaxZoom
				}
			}
		}

		sl := &StyleLayer{
			ID:          str(layer["id"]),
			Type:        str(layer["type"]),
			SourceLayer: str(layer["source-layer"]),
			appliesTo:   compileFilter(layer["filter"]),
			paint:       map[string]any{},
		}
		if mz, ok := layer["minzoom"].(float64); ok {
			sl.MinZoom = mz
		}
		if mz, ok := layer["maxzoom"].(float64); ok {
			sl.MaxZoom = mz
		}
		if paint, ok := layer["paint"].(map[string]any); ok {
			sl.paint = paint
		}

		s.byLayer[sl.SourceLayer] = append(s.byLayer[sl.SourceLayer], sl)
		s.byID[sl.ID] = sl
	}
	return s, nil
}

// GetStyleFor mirrors Styler.getStyleFor: first style whose filter
// matches the feature's properties wins.
func (s *Styler) GetStyleFor(layer string, props map[string]any) *StyleLayer {
	for _, style := range s.byLayer[layer] {
		if style.appliesTo(props) {
			return style
		}
	}
	return nil
}

// BackgroundColor returns the style's background paint color, if any.
func (s *Styler) BackgroundColor() string {
	if bg, ok := s.byID["background"]; ok {
		return bg.Color()
	}
	return ""
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// replaceConstants mirrors Styler._replaceConstants: swap "@name" strings
// for their constant values throughout the layer tree.
func replaceConstants(constants map[string]any, tree map[string]any) {
	for key, node := range tree {
		switch n := node.(type) {
		case map[string]any:
			replaceConstants(constants, n)
		case []any:
			for i, item := range n {
				if s, ok := item.(string); ok && len(s) > 0 && s[0] == '@' {
					n[i] = constants[s]
				} else if m, ok := item.(map[string]any); ok {
					replaceConstants(constants, m)
				}
			}
		case string:
			if len(n) > 0 && n[0] == '@' {
				tree[key] = constants[n]
			}
		}
	}
}

// compileFilter mirrors Styler._compileFilter, turning Mapbox filter
// expressions into predicate closures.
func compileFilter(filter any) func(map[string]any) bool {
	f, ok := filter.([]any)
	if !ok || len(f) == 0 {
		return func(map[string]any) bool { return true }
	}
	op, _ := f[0].(string)
	switch op {
	case "all":
		subs := compileSubFilters(f[1:])
		return func(p map[string]any) bool {
			for _, sub := range subs {
				if !sub(p) {
					return false
				}
			}
			return true
		}
	case "any":
		subs := compileSubFilters(f[1:])
		return func(p map[string]any) bool {
			for _, sub := range subs {
				if sub(p) {
					return true
				}
			}
			return false
		}
	case "none":
		subs := compileSubFilters(f[1:])
		return func(p map[string]any) bool {
			for _, sub := range subs {
				if sub(p) {
					return false
				}
			}
			return true
		}
	case "==":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool { return jsEqual(p[key], val) }
	case "!=":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool { return !jsEqual(p[key], val) }
	case "in":
		key, vals := str(f[1]), f[2:]
		return func(p map[string]any) bool {
			for _, v := range vals {
				if jsEqual(p[key], v) {
					return true
				}
			}
			return false
		}
	case "!in":
		key, vals := str(f[1]), f[2:]
		return func(p map[string]any) bool {
			for _, v := range vals {
				if jsEqual(p[key], v) {
					return false
				}
			}
			return true
		}
	case "has":
		key := str(f[1])
		return func(p map[string]any) bool { return truthy(p[key]) }
	case "!has":
		key := str(f[1])
		return func(p map[string]any) bool { return !truthy(p[key]) }
	case ">":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool {
			cmp, ok := compareNum(p[key], val)
			return ok && cmp > 0
		}
	case ">=":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool {
			cmp, ok := compareNum(p[key], val)
			return ok && cmp >= 0
		}
	case "<":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool {
			cmp, ok := compareNum(p[key], val)
			return ok && cmp < 0
		}
	case "<=":
		key, val := str(f[1]), f[2]
		return func(p map[string]any) bool {
			cmp, ok := compareNum(p[key], val)
			return ok && cmp <= 0
		}
	default:
		return func(map[string]any) bool { return true }
	}
}

func compileSubFilters(subs []any) []func(map[string]any) bool {
	out := make([]func(map[string]any) bool, len(subs))
	for i, sub := range subs {
		out[i] = compileFilter(sub)
	}
	return out
}

// jsEqual approximates JS === for the value types found in tiles and
// style filters: strings, numbers, and booleans.
func jsEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	default:
		return a == nil && b == nil
	}
}

func truthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return t != ""
	case float64:
		return t != 0
	case bool:
		return t
	default:
		return true
	}
}

// compareNum returns -1/0/1 for numeric comparison; ok is false when
// either side isn't a number, which makes every ordering operator false —
// like JS comparisons against undefined.
func compareNum(a, b any) (int, bool) {
	av, aok := a.(float64)
	bv, bok := b.(float64)
	if !aok || !bok {
		return 0, false
	}
	switch {
	case av < bv:
		return -1, true
	case av > bv:
		return 1, true
	default:
		return 0, true
	}
}
