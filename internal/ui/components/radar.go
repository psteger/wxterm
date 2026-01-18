package components

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"weatherterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	radarTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	radarInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	radarStationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#60A5FA"))

	radarTimestampStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))
)

// RenderRadar renders a single frame of radar as colored ASCII art
func RenderRadar(radar *api.RadarData, width, height int, frameIndex int) string {
	if radar == nil || len(radar.Frames) == 0 {
		return radarInfoStyle.Render("No radar data available. Press 'r' to refresh.")
	}

	// Clamp frame index
	if frameIndex < 0 {
		frameIndex = 0
	}
	if frameIndex >= len(radar.Frames) {
		frameIndex = len(radar.Frames) - 1
	}

	frame := radar.Frames[frameIndex]

	var b strings.Builder

	// Header with station info
	b.WriteString(radarTitleStyle.Render("Weather Radar"))
	b.WriteString("  ")
	b.WriteString(radarStationStyle.Render(fmt.Sprintf("%s - %s", radar.StationID, radar.StationName)))
	b.WriteString("\n")
	b.WriteString(radarInfoStyle.Render(fmt.Sprintf("[%s]", radar.Mode.String())))
	b.WriteString("  ")
	b.WriteString(radarTimestampStyle.Render(frame.Timestamp.Format("15:04")))
	b.WriteString("  ")
	b.WriteString(radarInfoStyle.Render(fmt.Sprintf("Frame %d/%d", frameIndex+1, len(radar.Frames))))
	b.WriteString("  ")
	b.WriteString(radarInfoStyle.Render("(m: mode, space: pause, r: refresh)"))
	b.WriteString("\n")

	// Calculate display dimensions - reserve space for header and legend
	displayHeight := height - 10
	displayWidth := width - 2

	if displayHeight < 10 {
		displayHeight = 10
	}
	if displayWidth < 20 {
		displayWidth = 20
	}

	// Render the frame as ASCII art with true colors
	asciiArt := imageToColoredASCII(frame.Image, displayWidth, displayHeight)
	b.WriteString(asciiArt)

	// Legend
	b.WriteString(renderRadarLegend())

	return b.String()
}

// imageToColoredASCII converts an image to colored ASCII art using half-block characters
// Each character cell represents 2 vertical pixels using ▀ with foreground (top) and background (bottom) colors
func imageToColoredASCII(img image.Image, width, height int) string {
	if img == nil {
		return ""
	}

	bounds := img.Bounds()
	imgWidth := bounds.Max.X - bounds.Min.X
	imgHeight := bounds.Max.Y - bounds.Min.Y

	if imgWidth == 0 || imgHeight == 0 {
		return ""
	}

	// Crop the image to remove top 2 lines and bottom 1 line worth of pixels
	// Calculate how many pixels to crop based on output height ratio
	cropTopLines := 2
	cropBottomLines := 2
	// With half-blocks, each output row represents 2 pixel rows
	effectiveHeight := height * 2
	cropTopPixels := int(float64(cropTopLines*2) * float64(imgHeight) / float64(effectiveHeight+cropTopLines*2+cropBottomLines*2))
	cropBottomPixels := int(float64(cropBottomLines*2) * float64(imgHeight) / float64(effectiveHeight+cropTopLines*2+cropBottomLines*2))

	// Adjust bounds to crop the image
	croppedMinY := bounds.Min.Y + cropTopPixels
	croppedMaxY := bounds.Max.Y - cropBottomPixels
	croppedHeight := croppedMaxY - croppedMinY

	if croppedHeight <= 0 {
		return ""
	}

	// Calculate scaling factors
	// Terminal characters are roughly twice as tall as wide, so we halve the output width
	// Each character represents 2 vertical pixels (top and bottom half)
	outputWidth := width / 2
	scaleX := float64(imgWidth) / float64(outputWidth)
	scaleY := float64(croppedHeight) / float64(height*2) // *2 because each row = 2 pixel rows

	var b strings.Builder

	for y := 0; y < height; y++ {
		for x := 0; x < outputWidth; x++ {
			// Sample two vertical pixels for this character cell
			srcX := int(float64(x) * scaleX)
			srcYTop := int(float64(y*2) * scaleY)
			srcYBottom := int(float64(y*2+1) * scaleY)

			if srcX >= imgWidth {
				srcX = imgWidth - 1
			}
			if srcYTop >= croppedHeight {
				srcYTop = croppedHeight - 1
			}
			if srcYBottom >= croppedHeight {
				srcYBottom = croppedHeight - 1
			}

			// Get top and bottom pixel colors
			topPixel := img.At(srcX+bounds.Min.X, srcYTop+croppedMinY)
			bottomPixel := img.At(srcX+bounds.Min.X, srcYBottom+croppedMinY)

			rTop, gTop, bTop, aTop := colorToRGBA(topPixel)
			rBot, gBot, bBot, aBot := colorToRGBA(bottomPixel)

			// Check transparency
			topTransparent := aTop < 30
			bottomTransparent := aBot < 30

			if topTransparent && bottomTransparent {
				// Both transparent - just a space
				b.WriteString(" ")
			} else if topTransparent {
				// Only bottom visible - use lower half block with foreground color
				b.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm▄\033[0m", rBot, gBot, bBot))
			} else if bottomTransparent {
				// Only top visible - use upper half block with foreground color
				b.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm▀\033[0m", rTop, gTop, bTop))
			} else {
				// Both visible - use upper half block with foreground (top) and background (bottom)
				// Format: \033[38;2;R;G;Bm for foreground, \033[48;2;R;G;Bm for background
				b.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀\033[0m",
					rTop, gTop, bTop, rBot, gBot, bBot))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// colorToRGBA converts any color.Color to RGBA uint8 values
func colorToRGBA(c color.Color) (r, g, b, a uint8) {
	rr, gg, bb, aa := c.RGBA()
	return uint8(rr >> 8), uint8(gg >> 8), uint8(bb >> 8), uint8(aa >> 8)
}

// renderRadarLegend renders a legend for precipitation intensity
func renderRadarLegend() string {
	var b strings.Builder

	b.WriteString(radarInfoStyle.Render("Intensity: "))

	// Show gradient with typical radar colors
	levels := []struct {
		char  string
		color string
		label string
	}{
		{"░", "#00ff00", "Light"},
		{"▒", "#ffff00", "Mod"},
		{"▓", "#ff8800", "Heavy"},
		{"█", "#ff0000", "Severe"},
	}

	for i, level := range levels {
		if i > 0 {
			b.WriteString(" ")
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(level.color))
		b.WriteString(style.Render(level.char))
		b.WriteString(radarInfoStyle.Render(level.label))
	}

	return b.String()
}
