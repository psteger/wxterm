package components

import (
	"fmt"
	"image"
	"strings"

	"wxterm/internal/api"
	"wxterm/internal/vmap"

	"github.com/charmbracelet/lipgloss"
)

var (
	radarTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	radarInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	radarTimestampStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))
)

// RenderRadar renders the radar view: a mapscii-style vector map with the
// rain overlay as colored cell backgrounds.
func RenderRadar(radar *api.RadarData, width, height int, frameIndex int, legendIndex int) string {
	if radar == nil {
		return radarInfoStyle.Render("No radar data available. Press 'r' to refresh.")
	}

	var b strings.Builder

	// Header
	b.WriteString(radarTitleStyle.Render("Weather Radar"))
	b.WriteString("  ")
	b.WriteString(radarInfoStyle.Render(fmt.Sprintf("Zoom: %d", radar.ZoomLevel)))
	b.WriteString("  ")
	b.WriteString(radarInfoStyle.Render(fmt.Sprintf("%.2f°, %.2f°", radar.CenterLat, radar.CenterLon)))

	var frameIdx int
	if len(radar.RainFrames) > 0 {
		frameIdx = clampIndex(frameIndex, len(radar.RainFrames))
		frame := radar.RainFrames[frameIdx]
		b.WriteString("  ")
		b.WriteString(radarTimestampStyle.Render(frame.Timestamp.Format("15:04")))
		b.WriteString("  ")
		b.WriteString(radarInfoStyle.Render(fmt.Sprintf("Frame %d/%d", frameIdx+1, len(radar.RainFrames))))
	}
	b.WriteString("\n")
	b.WriteString(radarInfoStyle.Render("(arrows: pan, +/-: zoom, space: pause, p: precip type, r: refresh)"))
	b.WriteString("\n")

	// Calculate display area
	displayWidth := width
	displayHeight := height - 6
	if displayHeight < 5 {
		displayHeight = 5
	}
	if displayWidth < 10 {
		displayWidth = 10
	}

	// Get rain image for current frame
	var rainImg image.Image
	if len(radar.RainFrames) > 0 {
		rainImg = radar.RainFrames[frameIdx].Image
	}

	b.WriteString(renderVectorMap(radar, rainImg, displayWidth, displayHeight))

	b.WriteString(renderRainLegend(legendIndex))

	return b.String()
}

func clampIndex(idx, length int) int {
	if idx < 0 {
		return 0
	}
	if idx >= length {
		return length - 1
	}
	return idx
}

// renderVectorMap draws the vector base map with the mapscii renderer,
// then overlays rain as per-cell background colors.
func renderVectorMap(radar *api.RadarData, rainImg image.Image, charWidth, charHeight int) string {
	renderer, err := vmap.Default()
	if err != nil {
		return radarInfoStyle.Render("map renderer unavailable: " + err.Error())
	}

	widthPx := charWidth * 2
	heightPx := charHeight * 4
	frame := renderer.Draw(radar.VMap, widthPx, heightPx)

	if rainImg != nil {
		// The rain composite shares the Web Mercator pixel grid the
		// canvas is drawn in; CenterPX/PY anchor the view within it.
		startX := radar.CenterPX - widthPx/2
		startY := radar.CenterPY - heightPx/2
		rb := rainImg.Bounds()
		for cy := 0; cy < charHeight; cy++ {
			for cx := 0; cx < charWidth; cx++ {
				rainPX := startX + cx*2 + 1
				rainPY := startY + cy*4 + 2
				if rainPX < rb.Min.X || rainPX >= rb.Max.X ||
					rainPY < rb.Min.Y || rainPY >= rb.Max.Y {
					continue
				}
				r, g, bl, a := rainImg.At(rainPX, rainPY).RGBA()
				if a > 0x1000 {
					frame.SetCellBackgroundRGB(cx, cy, uint8((r>>8)&0xff), uint8((g>>8)&0xff), uint8((bl>>8)&0xff))
				}
			}
		}
	}

	return frame.Render()
}

// nexradColors contains the NEXRAD Level III (scheme 6) color table
// from the RainViewer API colors table, indexed by dBZ value (0–75).
// Source: https://www.rainviewer.com/files/rainviewer_api_colors_table.csv
var nexradColors = [76][3]uint8{
	/* 0  */ {0x04, 0xe9, 0xe7} /* 1  */, {0x03, 0xea, 0xe7} /* 2  */, {0x02, 0xeb, 0xe7} /* 3  */, {0x01, 0xec, 0xe7},
	/* 4  */ {0x00, 0xed, 0xe7} /* 5  */, {0x00, 0xef, 0xe7} /* 6  */, {0x00, 0xde, 0xea} /* 7  */, {0x00, 0xcd, 0xed},
	/* 8  */ {0x00, 0xbd, 0xf0} /* 9  */, {0x00, 0xac, 0xf3} /* 10 */, {0x00, 0x9c, 0xf7} /* 11 */, {0x00, 0x7c, 0xf7},
	/* 12 */ {0x00, 0x5d, 0xf7} /* 13 */, {0x00, 0x3e, 0xf7} /* 14 */, {0x00, 0x1f, 0xf7} /* 15 */, {0x00, 0x00, 0xf7},
	/* 16 */ {0x00, 0x33, 0xc5} /* 17 */, {0x00, 0x66, 0x94} /* 18 */, {0x00, 0x99, 0x62} /* 19 */, {0x00, 0xcc, 0x31},
	/* 20 */ {0x00, 0xff, 0x00} /* 21 */, {0x00, 0xf0, 0x00} /* 22 */, {0x01, 0xe2, 0x01} /* 23 */, {0x01, 0xd3, 0x01},
	/* 24 */ {0x02, 0xc5, 0x02} /* 25 */, {0x03, 0xb7, 0x03} /* 26 */, {0x04, 0xa9, 0x03} /* 27 */, {0x05, 0x9b, 0x03},
	/* 28 */ {0x06, 0x8e, 0x04} /* 29 */, {0x07, 0x80, 0x04} /* 30 */, {0x08, 0x73, 0x05} /* 31 */, {0x39, 0x8f, 0x04},
	/* 32 */ {0x6a, 0xab, 0x03} /* 33 */, {0x9c, 0xc7, 0x02} /* 34 */, {0xcd, 0xe3, 0x00} /* 35 */, {0xff, 0xff, 0x00},
	/* 36 */ {0xfb, 0xf5, 0x00} /* 37 */, {0xf7, 0xeb, 0x00} /* 38 */, {0xf3, 0xe1, 0x00} /* 39 */, {0xef, 0xd7, 0x00},
	/* 40 */ {0xec, 0xce, 0x00} /* 41 */, {0xef, 0xc2, 0x00} /* 42 */, {0xf3, 0xb6, 0x00} /* 43 */, {0xf6, 0xaa, 0x00},
	/* 44 */ {0xfa, 0x9e, 0x00} /* 45 */, {0xfe, 0x93, 0x00} /* 46 */, {0xfe, 0x75, 0x00} /* 47 */, {0xfe, 0x58, 0x00},
	/* 48 */ {0xfe, 0x3a, 0x00} /* 49 */, {0xfe, 0x1d, 0x00} /* 50 */, {0xff, 0x00, 0x00} /* 51 */, {0xf1, 0x00, 0x00},
	/* 52 */ {0xe4, 0x00, 0x00} /* 53 */, {0xd7, 0x00, 0x00} /* 54 */, {0xca, 0x00, 0x00} /* 55 */, {0xbd, 0x00, 0x00},
	/* 56 */ {0xbd, 0x00, 0x00} /* 57 */, {0xbd, 0x00, 0x00} /* 58 */, {0xbd, 0x00, 0x00} /* 59 */, {0xbd, 0x00, 0x00},
	/* 60 */ {0xbd, 0x00, 0x00} /* 61 */, {0xca, 0x00, 0x32} /* 62 */, {0xd7, 0x00, 0x65} /* 63 */, {0xe4, 0x00, 0x98},
	/* 64 */ {0xf1, 0x00, 0xcb} /* 65 */, {0xfe, 0x00, 0xfe} /* 66 */, {0xea, 0x10, 0xf2} /* 67 */, {0xd6, 0x20, 0xe7},
	/* 68 */ {0xc3, 0x31, 0xdc} /* 69 */, {0xaf, 0x41, 0xd1} /* 70 */, {0x9c, 0x52, 0xc6} /* 71 */, {0xaf, 0x74, 0xd1},
	/* 72 */ {0xc3, 0x96, 0xdc} /* 73 */, {0xd6, 0xb9, 0xe7} /* 74 */, {0xea, 0xdb, 0xf2} /* 75 */, {0xfe, 0xfe, 0xfe},
}

// legendEntry maps a dBZ value to a NEXRAD color and intensity label.
type legendEntry struct {
	dBZ     int
	r, g, b uint8
	label   string
}

// precipLegend groups legend entries under a precipitation type heading.
type precipLegend struct {
	name    string
	entries []legendEntry
}

// radarLegends defines the color legends for each precipitation type.
// The same NEXRAD colors appear on the radar, but different dBZ thresholds
// correspond to different intensities depending on precipitation type.
var radarLegends = []precipLegend{
	{"Rain", []legendEntry{
		{5, 0x00, 0xef, 0xe7, "Drizzle"},
		{15, 0x00, 0x00, 0xf7, "Light"},
		{25, 0x03, 0xb7, 0x03, "Mod"},
		{35, 0xff, 0xff, 0x00, "Heavy"},
		{45, 0xfe, 0x93, 0x00, "V.Heavy"},
		{50, 0xff, 0x00, 0x00, "Intense"},
		{60, 0xbd, 0x00, 0x00, "Extreme"},
		{63, 0xe4, 0x00, 0x98, "Severe"},
		{65, 0xfe, 0x00, 0xfe, "Hail"},
	}},
	{"Snow", []legendEntry{
		{5, 0x00, 0xef, 0xe7, "Flurry"},
		{10, 0x00, 0x9c, 0xf7, "Light"},
		{20, 0x00, 0xff, 0x00, "Mod"},
		{30, 0x08, 0x73, 0x05, "Heavy"},
		{35, 0xff, 0xff, 0x00, "Intense"},
	}},
	{"Frz/Mix", []legendEntry{
		{10, 0x00, 0x9c, 0xf7, "Light"},
		{20, 0x00, 0xff, 0x00, "Mod"},
		{30, 0x08, 0x73, 0x05, "Heavy"},
	}},
}

// NumRadarLegends returns the number of available precipitation legends.
func NumRadarLegends() int {
	return len(radarLegends)
}

func renderRainLegend(legendIndex int) string {
	idx := legendIndex % len(radarLegends)
	legend := radarLegends[idx]

	var b strings.Builder
	b.WriteString(radarInfoStyle.Render(legend.name + ": "))
	for j, e := range legend.entries {
		if j > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "\033[48;2;%d;%d;%dm  \033[0m", e.r, e.g, e.b)
		b.WriteString(radarInfoStyle.Render(e.label))
	}

	return b.String()
}
