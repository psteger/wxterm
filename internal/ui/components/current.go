package components

import (
	"fmt"
	"strings"

	"wxterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	currentTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED"))

	bigTempStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	iconBoxStyle = lipgloss.NewStyle().
			Width(16).
			Align(lipgloss.Center)
)

// RenderCurrentWeather renders the current weather view
func RenderCurrentWeather(weather *api.WeatherData, width int, useImperial bool) string {
	if weather == nil {
		return "No weather data available"
	}

	current := weather.Current
	category := api.WeatherCodeCategory(current.WeatherCode)
	icon := WeatherIcon(category, current.IsDay)
	description := api.WeatherCodeDescription(current.WeatherCode)

	// Build the layout
	var b strings.Builder

	// Main temperature and icon row
	tempStr := formatTemp(current.Temperature, useImperial)
	feelsLike := fmt.Sprintf("Feels like %s", formatTemp(current.FeelsLike, useImperial))

	tempSection := lipgloss.JoinVertical(
		lipgloss.Left,
		bigTempStyle.Render(tempStr),
		labelStyle.Render(feelsLike),
		"",
		valueStyle.Render(description),
	)

	iconSection := iconBoxStyle.Render(icon)

	mainRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		iconSection,
		"    ",
		tempSection,
	)

	b.WriteString(mainRow)
	b.WriteString("\n\n")

	// Details grid
	details := []struct {
		label string
		value string
	}{
		{"Humidity", fmt.Sprintf("%d%%", current.Humidity)},
		{"Wind", fmt.Sprintf("%s %s", formatWindSpeed(current.WindSpeed, useImperial), WindDirection(current.WindDirection))},
		{"Pressure", fmt.Sprintf("%.0f hPa", current.Pressure)},
		{"Cloud Cover", fmt.Sprintf("%d%%", current.CloudCover)},
		{"Precipitation", formatPrecipitation(current.Precipitation, useImperial)},
		{"Visibility", formatVisibility(current.Visibility, useImperial)},
	}

	// Render details in 2 columns
	detailWidth := 24
	detailStyle := lipgloss.NewStyle().Width(detailWidth)

	for i := 0; i < len(details); i += 2 {
		left := detailStyle.Render(
			labelStyle.Render(details[i].label+": ") + valueStyle.Render(details[i].value),
		)
		right := ""
		if i+1 < len(details) {
			right = detailStyle.Render(
				labelStyle.Render(details[i+1].label+": ") + valueStyle.Render(details[i+1].value),
			)
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
		b.WriteString("\n")
	}

	// Sunrise/Sunset
	if len(weather.Daily.Sunrise) > 0 {
		b.WriteString("\n")
		sunrise := weather.Daily.Sunrise[0].Format("15:04")
		sunset := weather.Daily.Sunset[0].Format("15:04")
		b.WriteString(labelStyle.Render("Sunrise: ") + valueStyle.Render(sunrise))
		b.WriteString("    ")
		b.WriteString(labelStyle.Render("Sunset: ") + valueStyle.Render(sunset))
	}

	return b.String()
}

// Unit conversion helpers
func formatTemp(celsius float64, useImperial bool) string {
	if useImperial {
		f := celsius*9/5 + 32
		return fmt.Sprintf("%.0f°F", f)
	}
	return fmt.Sprintf("%.0f°C", celsius)
}

func formatTempShort(celsius float64, useImperial bool) string {
	if useImperial {
		f := celsius*9/5 + 32
		return fmt.Sprintf("%.0f°", f)
	}
	return fmt.Sprintf("%.0f°", celsius)
}

func formatWindSpeed(kmh float64, useImperial bool) string {
	if useImperial {
		mph := kmh * 0.621371
		return fmt.Sprintf("%.0f mph", mph)
	}
	return fmt.Sprintf("%.0f km/h", kmh)
}

func formatVisibility(meters float64, useImperial bool) string {
	if useImperial {
		miles := meters / 1609.34
		return fmt.Sprintf("%.1f mi", miles)
	}
	return fmt.Sprintf("%.1f km", meters/1000)
}

func formatPrecipitation(mm float64, useImperial bool) string {
	if useImperial {
		inches := mm / 25.4
		return fmt.Sprintf("%.2f in", inches)
	}
	return fmt.Sprintf("%.1f mm", mm)
}
