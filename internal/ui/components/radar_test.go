package components

import (
	"strings"
	"testing"

	"wxterm/internal/api"
)

func TestRenderRadarNil(t *testing.T) {
	out := RenderRadar(nil, 80, 24, 0, 0)
	if !strings.Contains(out, "No radar data") {
		t.Errorf("expected placeholder message, got %q", out)
	}
}

// Without prefetched vector tiles the radar view must still render a
// full-size frame (background only) rather than erroring.
func TestRenderRadarWithoutTiles(t *testing.T) {
	radar := &api.RadarData{
		CenterLat: 40.7128, CenterLon: -74.0060,
		ZoomLevel: 6, CenterPX: 256, CenterPY: 256,
	}
	out := RenderRadar(radar, 80, 30, 0, 0)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// header (2) + map rows (30-6) + legend (1)
	if len(lines) != 2+24+1 {
		t.Errorf("expected %d lines, got %d", 2+24+1, len(lines))
	}
}
