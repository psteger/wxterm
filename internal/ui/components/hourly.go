package components

import (
	"fmt"
	"strings"

	"wxterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	hourHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#60A5FA")).
			Width(6)

	hourTempStyle = lipgloss.NewStyle().
			Bold(true).
			Width(6)

	hourDetailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(6)

	hourCellStyle = lipgloss.NewStyle().
			Width(8).
			Align(lipgloss.Center)

	currentHourStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF"))
)

// RenderHourlyForecast renders the hourly forecast view
func RenderHourlyForecast(weather *api.WeatherData, width, height int, useImperial bool) string {
	if weather == nil {
		return "No weather data available"
	}

	hourly := weather.Hourly
	if len(hourly.Time) == 0 {
		return "No hourly data available"
	}

	var b strings.Builder

	// Calculate how many hours we can fit
	cellWidth := 9
	maxHours := (width - 12) / cellWidth
	if maxHours > 24 {
		maxHours = 24
	}
	if maxHours > len(hourly.Time) {
		maxHours = len(hourly.Time)
	}

	// Header row (times)
	var timeRow strings.Builder
	timeRow.WriteString(lipgloss.NewStyle().Width(10).Render(""))
	for i := 0; i < maxHours; i++ {
		timeStr := hourly.Time[i].Format("15:00")
		cell := hourCellStyle.Render(timeStr)
		timeRow.WriteString(cell)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true).Render(timeRow.String()))
	b.WriteString("\n")

	// Weather icon row
	var iconRow strings.Builder
	iconRow.WriteString(lipgloss.NewStyle().Width(10).Render(""))
	for i := 0; i < maxHours; i++ {
		category := api.WeatherCodeCategory(hourly.WeatherCode[i])
		icon := SmallWeatherIcon(category, true)
		cell := hourCellStyle.Render(icon)
		iconRow.WriteString(cell)
	}
	b.WriteString(iconRow.String())
	b.WriteString("\n")

	// Temperature row
	var tempRow strings.Builder
	tempRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Temp"))
	for i := 0; i < maxHours; i++ {
		temp := formatTempShort(hourly.Temperature[i], useImperial)
		style := getTempStyle(hourly.Temperature[i])
		cell := hourCellStyle.Render(style.Render(temp))
		tempRow.WriteString(cell)
	}
	b.WriteString(tempRow.String())
	b.WriteString("\n")

	// Feels like row
	var feelsRow strings.Builder
	feelsRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Feels"))
	for i := 0; i < maxHours; i++ {
		temp := formatTempShort(hourly.FeelsLike[i], useImperial)
		cell := hourCellStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(temp))
		feelsRow.WriteString(cell)
	}
	b.WriteString(feelsRow.String())
	b.WriteString("\n")

	// Humidity row
	var humidRow strings.Builder
	humidRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Humid"))
	for i := 0; i < maxHours; i++ {
		humid := fmt.Sprintf("%d%%", hourly.Humidity[i])
		cell := hourCellStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Render(humid))
		humidRow.WriteString(cell)
	}
	b.WriteString(humidRow.String())
	b.WriteString("\n")

	// Wind direction row
	var windDirRow strings.Builder
	windDirRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Wind Dir"))
	for i := 0; i < maxHours; i++ {
		arrow := windDirectionArrow(hourly.WindDirection[i], hourly.WindSpeed[i])
		cell := hourCellStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(arrow))
		windDirRow.WriteString(cell)
	}
	b.WriteString(windDirRow.String())
	b.WriteString("\n")

	// Wind speed row
	var windRow strings.Builder
	windLabel := "km/h"
	if useImperial {
		windLabel = "mph"
	}
	windRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render(windLabel))
	for i := 0; i < maxHours; i++ {
		windVal := hourly.WindSpeed[i]
		if useImperial {
			windVal = windVal * 0.621371
		}
		wind := fmt.Sprintf("%.0f", windVal)
		cell := hourCellStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(wind))
		windRow.WriteString(cell)
	}
	b.WriteString(windRow.String())
	b.WriteString("\n")

	// Wind gusts row
	var gustRow strings.Builder
	gustLabel := "Gusts"
	gustRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render(gustLabel))
	for i := 0; i < maxHours; i++ {
		gustVal := hourly.WindGusts[i]
		if useImperial {
			gustVal = gustVal * 0.621371
		}
		gust := fmt.Sprintf("%.0f", gustVal)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		if gustVal > 40 {
			style = style.Foreground(lipgloss.Color("#F97316"))
		}
		cell := hourCellStyle.Render(style.Render(gust))
		gustRow.WriteString(cell)
	}
	b.WriteString(gustRow.String())
	b.WriteString("\n")

	// Rain probability row
	var rainProbRow strings.Builder
	rainProbRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Precip %"))
	for i := 0; i < maxHours; i++ {
		prob := fmt.Sprintf("%d%%", hourly.PrecipitationProb[i])
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		if hourly.PrecipitationProb[i] > 50 {
			style = style.Foreground(lipgloss.Color("#60A5FA"))
		}
		cell := hourCellStyle.Render(style.Render(prob))
		rainProbRow.WriteString(cell)
	}
	b.WriteString(rainProbRow.String())
	b.WriteString("\n")

	// Precipitation row
	var precipRow strings.Builder
	precipLabel := "mm"
	if useImperial {
		precipLabel = "in"
	}
	precipRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render(precipLabel))
	for i := 0; i < maxHours; i++ {
		precipVal := hourly.Precipitation[i]
		var precip string
		if useImperial {
			precip = fmt.Sprintf("%.2f", precipVal/25.4)
		} else {
			precip = fmt.Sprintf("%.1f", precipVal)
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		if hourly.Precipitation[i] > 0 {
			style = style.Foreground(lipgloss.Color("#60A5FA"))
		}
		cell := hourCellStyle.Render(style.Render(precip))
		precipRow.WriteString(cell)
	}
	b.WriteString(precipRow.String())
	b.WriteString("\n")

	// UV index row
	var uvRow strings.Builder
	uvRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("UV"))
	for i := 0; i < maxHours; i++ {
		uv := fmt.Sprintf("%.0f", hourly.UVIndex[i])
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		if hourly.UVIndex[i] >= 8 {
			style = style.Foreground(lipgloss.Color("#EF4444"))
		} else if hourly.UVIndex[i] >= 6 {
			style = style.Foreground(lipgloss.Color("#F97316"))
		} else if hourly.UVIndex[i] >= 3 {
			style = style.Foreground(lipgloss.Color("#F59E0B"))
		}
		cell := hourCellStyle.Render(style.Render(uv))
		uvRow.WriteString(cell)
	}
	b.WriteString(uvRow.String())
	b.WriteString("\n")

	// Visibility row
	var visRow strings.Builder
	visLabel := "Vis km"
	if useImperial {
		visLabel = "Vis mi"
	}
	visRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render(visLabel))
	for i := 0; i < maxHours; i++ {
		visVal := hourly.Visibility[i] / 1000.0 // meters to km
		if useImperial {
			visVal = hourly.Visibility[i] / 1609.34 // meters to miles
		}
		vis := fmt.Sprintf("%.0f", visVal)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		if visVal < 5 {
			style = style.Foreground(lipgloss.Color("#F59E0B"))
		}
		cell := hourCellStyle.Render(style.Render(vis))
		visRow.WriteString(cell)
	}
	b.WriteString(visRow.String())
	b.WriteString("\n")

	// Surface pressure row
	var pressRow strings.Builder
	pressRow.WriteString(lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("#6B7280")).Render("Press hPa"))
	for i := 0; i < maxHours; i++ {
		press := fmt.Sprintf("%.0f", hourly.SurfacePressure[i])
		cell := hourCellStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(press))
		pressRow.WriteString(cell)
	}
	b.WriteString(pressRow.String())

	return b.String()
}

func windDirectionArrow(degrees int, windSpeed float64) string {
	if windSpeed < 1 {
		return "·"
	}
	// degrees is where wind comes FROM; arrow shows where it blows TO
	arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
	idx := ((degrees + 22) / 45) % 8
	return arrows[idx]
}

func getTempStyle(tempCelsius float64) lipgloss.Style {
	switch {
	case tempCelsius >= 30:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")).Bold(true)
	case tempCelsius >= 20:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	case tempCelsius <= 0:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Bold(true)
	case tempCelsius <= 10:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	}
}
