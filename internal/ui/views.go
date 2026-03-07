package ui

import (
	"strings"

	"wxterm/internal/ui/components"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMainView() string {
	if m.loading {
		if m.activeView == ViewRadar {
			return m.spinner.View() + " Loading radar data..."
		}
		return m.spinner.View() + " Loading weather data..."
	}

	switch m.activeView {
	case ViewCurrent:
		if m.weather == nil {
			return mutedStyle.Render("No weather data. Press 's' to search for a location.")
		}
		return components.RenderCurrentWeather(m.weather, m.width, m.useImperial)
	case ViewHourly:
		if m.weather == nil {
			return mutedStyle.Render("No weather data. Press 's' to search for a location.")
		}
		return components.RenderHourlyForecast(m.weather, m.width, m.height, m.useImperial)
	case ViewDaily:
		if m.weather == nil {
			return mutedStyle.Render("No weather data. Press 's' to search for a location.")
		}
		return components.RenderDailyForecast(m.weather, m.width, m.useImperial)
	case ViewRadar:
		return components.RenderRadar(m.radar, m.width, m.height, m.radarFrameIndex, m.radarLegendIndex)
	default:
		return "Unknown view"
	}
}

func (m Model) renderSearchView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search Location"))
	b.WriteString("\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Searching...")
	} else if len(m.searchResults) > 0 {
		b.WriteString(components.RenderLocationSearch(m.searchResults, m.selectedIndex))
	} else if m.searchInput.Value() != "" {
		b.WriteString(mutedStyle.Render("Press Enter to search"))
	}

	return b.String()
}

func (m Model) renderCoordinatesView() string {
	return components.RenderCoordinateInput(
		m.latInput.Value(),
		m.lonInput.Value(),
		m.focusLat,
	)
}

func (m Model) renderSavedLocationsView() string {
	if m.config == nil || len(m.config.SavedLocations) == 0 {
		return mutedStyle.Render("No saved locations. Press Ctrl+S to save current location.")
	}

	var names []string
	for _, loc := range m.config.SavedLocations {
		names = append(names, loc.DisplayName())
	}

	return components.RenderSavedLocations(names, m.selectedIndex)
}

func (m Model) renderHelpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Help"))
	b.WriteString("\n\n")

	helpContent := `Navigation
──────────
←/→/tab/shift+tab Cycle through views
1, 2, 3, 4        Jump to Current / Hourly / Daily / Radar

Location
────────
s                 Search for a city
l                 Enter coordinates manually
ctrl+s            Save current location
r                 Refresh weather/radar data

Radar
─────
arrows            Pan radar map
+/-/=             Zoom radar in/out
space             Pause/play radar animation
p                 Cycle precipitation legend

General
───────
u                 Toggle metric/imperial units
?                 Show this help
q / ctrl+c        Quit

In Search Mode
──────────────
↑/↓               Navigate results
enter             Select location / Search
esc               Cancel

In Coordinate Mode
──────────────────
tab               Switch between lat/lon fields
enter             Confirm coordinates
esc               Cancel`

	// Style the help content
	lines := strings.Split(helpContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "──") {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(line))
		} else if !strings.Contains(line, "  ") && line != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true).Render(line))
		} else {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				key := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Width(18).Render(parts[0])
				desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(strings.TrimSpace(parts[1]))
				b.WriteString(key + desc)
			} else {
				b.WriteString(line)
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Press ? or Esc to close"))

	return b.String()
}
