package components

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"weatherterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

// ASCII character gradient from light to heavy (like the Python implementation)
// Space = no precipitation, then increasing density
const asciiGradient = " ░░▒▒▓▓██"

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

// imageToColoredASCII converts an image to colored ASCII art
// Uses the original image colors with ASCII density characters
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
	cropBottomLines := 1
	cropTopPixels := int(float64(cropTopLines) * float64(imgHeight) / float64(height+cropTopLines+cropBottomLines))
	cropBottomPixels := int(float64(cropBottomLines) * float64(imgHeight) / float64(height+cropTopLines+cropBottomLines))

	// Adjust bounds to crop the image
	croppedMinY := bounds.Min.Y + cropTopPixels
	croppedMaxY := bounds.Max.Y - cropBottomPixels
	croppedHeight := croppedMaxY - croppedMinY

	if croppedHeight <= 0 {
		return ""
	}

	// Calculate scaling factors
	// Terminal characters are roughly twice as tall as wide, so we need to
	// sample more horizontal pixels per character to compensate
	// We also halve the output width since we're sampling 2x horizontally
	outputWidth := width / 2
	scaleX := float64(imgWidth) / float64(outputWidth)
	scaleY := float64(croppedHeight) / float64(height)

	var b strings.Builder
	gradientRunes := []rune(asciiGradient)
	gradientLen := len(gradientRunes)

	var lastColor string

	for y := 0; y < height; y++ {
		for x := 0; x < outputWidth; x++ {
			// Sample the source image from the cropped region
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			if srcX >= imgWidth {
				srcX = imgWidth - 1
			}
			if srcY >= croppedHeight {
				srcY = croppedHeight - 1
			}

			// Get pixel color - handles both paletted and RGBA images
			pixel := img.At(srcX+bounds.Min.X, srcY+croppedMinY)
			r8, g8, b8, a8 := colorToRGBA(pixel)

			// Calculate grayscale intensity for character selection
			gray := (int(r8) + int(g8) + int(b8)) / 3

			// Map intensity to character
			charIdx := gray * (gradientLen - 1) / 255
			if charIdx >= gradientLen {
				charIdx = gradientLen - 1
			}
			char := string(gradientRunes[charIdx])

			// Skip fully transparent pixels
			if a8 < 30 {
				b.WriteString(" ")
				continue
			}

			// Use true color ANSI escape for the character
			// Format: \033[38;2;R;G;Bm
			colorCode := fmt.Sprintf("\033[38;2;%d;%d;%dm", r8, g8, b8)

			// Optimization: only emit color code if it changed
			if colorCode != lastColor {
				b.WriteString("\033[0m") // Reset first
				b.WriteString(colorCode)
				lastColor = colorCode
			}
			b.WriteString(char)
		}
		b.WriteString("\033[0m\n") // Reset at end of line
		lastColor = ""
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
