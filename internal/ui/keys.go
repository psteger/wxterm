package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings
type KeyMap struct {
	NextView    key.Binding
	PrevView    key.Binding
	View1       key.Binding
	View2       key.Binding
	View3       key.Binding
	Search      key.Binding
	Location    key.Binding
	Refresh     key.Binding
	Save        key.Binding
	Help        key.Binding
	Quit        key.Binding
	Enter       key.Binding
	Escape      key.Binding
	Up          key.Binding
	Down        key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextView: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		PrevView: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev view"),
		),
		View1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "current"),
		),
		View2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "hourly"),
		),
		View3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "daily"),
		),
		Search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search city"),
		),
		Location: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "enter coords"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save location"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "down"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextView, k.Search, k.Refresh, k.Help, k.Quit}
}

// FullHelp returns the full help string
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextView, k.PrevView, k.View1, k.View2, k.View3},
		{k.Search, k.Location, k.Refresh, k.Save},
		{k.Help, k.Quit},
	}
}
