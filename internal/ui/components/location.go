package components

import (
	"fmt"
	"strings"

	"wxterm/internal/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1)

	searchResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				PaddingLeft(2)

	searchSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF")).
				PaddingLeft(2).
				Bold(true)

	searchHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)
)

// RenderLocationSearch renders the search results list
func RenderLocationSearch(results []api.GeoLocation, selectedIndex int) string {
	if len(results) == 0 {
		return searchHintStyle.Render("No results found. Try a different search term.")
	}

	var b strings.Builder

	for i, loc := range results {
		line := loc.FormatLocation()
		if i == selectedIndex {
			b.WriteString(searchSelectedStyle.Render(fmt.Sprintf("> %s", line)))
		} else {
			b.WriteString(searchResultStyle.Render(fmt.Sprintf("  %s", line)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(searchHintStyle.Render("↑/↓ to navigate • Enter to select • Esc to cancel"))

	return b.String()
}

// RenderCoordinateInput renders the manual coordinate input dialog
func RenderCoordinateInput(latInput, lonInput string, focusLat bool) string {
	var b strings.Builder

	b.WriteString(searchTitleStyle.Render("Enter Coordinates"))
	b.WriteString("\n\n")

	latLabel := "Latitude: "
	lonLabel := "Longitude: "

	if focusLat {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render(latLabel))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(latInput + "█"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(latLabel))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(latInput))
	}

	b.WriteString("\n")

	if !focusLat {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render(lonLabel))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(lonInput + "█"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(lonLabel))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(lonInput))
	}

	b.WriteString("\n\n")
	b.WriteString(searchHintStyle.Render("Tab to switch fields • Enter to confirm • Esc to cancel"))

	return b.String()
}

// RenderSavedLocations renders the list of saved locations
func RenderSavedLocations(locations []string, selectedIndex int) string {
	if len(locations) == 0 {
		return searchHintStyle.Render("No saved locations. Press Ctrl+S to save current location.")
	}

	var b strings.Builder

	b.WriteString(searchTitleStyle.Render("Saved Locations"))
	b.WriteString("\n\n")

	for i, loc := range locations {
		if i == selectedIndex {
			b.WriteString(searchSelectedStyle.Render(fmt.Sprintf("> %s", loc)))
		} else {
			b.WriteString(searchResultStyle.Render(fmt.Sprintf("  %s", loc)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(searchHintStyle.Render("↑/↓ to navigate • Enter to select • D to delete • Esc to cancel"))

	return b.String()
}
