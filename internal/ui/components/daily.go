package components

import (
	"fmt"
	"strings"

	"wxterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	dayNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#60A5FA")).
			Width(12)

	dayDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(8)

	dayTempHighStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F97316")).
				Width(6)

	dayTempLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Width(6)

	dayConditionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Width(20)

	dayPrecipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Width(8)

	dayWindStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Width(8)

	dayRowStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	todayStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")).
			PaddingLeft(1).
			PaddingRight(1)
)

// RenderDailyForecast renders the 7-day forecast view
func RenderDailyForecast(weather *api.WeatherData, width int, useImperial bool) string {
	if weather == nil {
		return "No weather data available"
	}

	daily := weather.Daily
	if len(daily.Time) == 0 {
		return "No daily data available"
	}

	var b strings.Builder

	// Header
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		dayNameStyle.Render("Day"),
		dayDateStyle.Render("Date"),
		lipgloss.NewStyle().Width(7).Render(""),
		lipgloss.NewStyle().Width(6).Render("High"),
		lipgloss.NewStyle().Width(6).Render("Low"),
		lipgloss.NewStyle().Width(8).Render("Precip"),
		lipgloss.NewStyle().Width(8).Render("Avg Wind"),
		lipgloss.NewStyle().Width(8).Render("Max Wind"),
		dayConditionStyle.Render("Condition"),
	)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(header))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(width-4, 75)))
	b.WriteString("\n")

	// Days
	for i := 0; i < len(daily.Time); i++ {
		dayName := daily.Time[i].Format("Monday")
		if i == 0 {
			dayName = "Today"
		} else if i == 1 {
			dayName = "Tomorrow"
		}

		dateStr := daily.Time[i].Format("Jan 2")
		category := api.WeatherCodeCategory(daily.WeatherCode[i])
		icon := SmallWeatherIcon(category, true)
		highTemp := formatTempShort(daily.TemperatureMax[i], useImperial)
		lowTemp := formatTempShort(daily.TemperatureMin[i], useImperial)
		precip := fmt.Sprintf("%d%%", daily.PrecipitationProb[i])
		condition := api.WeatherCodeDescription(daily.WeatherCode[i])

		windUnit := "km/h"
		if useImperial {
			windUnit = "mph"
		}
		avgWind := daily.WindSpeedMean[i]
		maxWind := daily.WindSpeedMax[i]
		if useImperial {
			avgWind = avgWind * 0.621371
			maxWind = maxWind * 0.621371
		}
		avgWindStr := fmt.Sprintf("%.0f %s", avgWind, windUnit)
		maxWindStr := fmt.Sprintf("%.0f %s", maxWind, windUnit)

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			dayNameStyle.Render(dayName),
			dayDateStyle.Render(dateStr),
			lipgloss.NewStyle().Width(7).Render(icon),
			getTempStyle(daily.TemperatureMax[i]).Width(6).Render(highTemp),
			getTempStyle(daily.TemperatureMin[i]).Width(6).Render(lowTemp),
			dayPrecipStyle.Render(precip),
			dayWindStyle.Render(avgWindStr),
			dayWindStyle.Render(maxWindStr),
			dayConditionStyle.Render(condition),
		)

		if i == 0 {
			b.WriteString(todayStyle.Render(row))
		} else {
			b.WriteString(dayRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}
