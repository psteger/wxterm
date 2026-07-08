package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#60A5FA")
	accentColor    = lipgloss.Color("#F59E0B")
	mutedColor     = lipgloss.Color("#6B7280")
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2)

	tabGap = lipgloss.NewStyle().Width(1)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Location style
	locationStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Version style
	versionStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Update notice style
	updateStyle = lipgloss.NewStyle().
			Foreground(accentColor)
)
