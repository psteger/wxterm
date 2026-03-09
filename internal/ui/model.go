package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"wxterm/internal/api"
	"wxterm/internal/config"
	"wxterm/internal/location"
	"wxterm/internal/ui/components"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType represents different views in the app
type ViewType int

const (
	ViewCurrent ViewType = iota
	ViewHourly
	ViewDaily
	ViewRadar
)

// Mode represents the current interaction mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeCoordinates
	ModeSavedLocations
	ModeHelp
)

// Model represents the main application state
type Model struct {
	// View state
	activeView ViewType
	mode       Mode

	// Data
	location      location.Location
	weather       *api.WeatherData
	radar         *api.RadarData
	searchResults []api.GeoLocation
	config        *config.Config

	// UI state
	loading       bool
	err           error
	width         int
	height        int
	selectedIndex int

	// Input components
	searchInput textinput.Model
	latInput    textinput.Model
	lonInput    textinput.Model
	focusLat    bool
	spinner     spinner.Model

	// API client
	apiClient *api.Client

	// Keys
	keys KeyMap

	// Units
	useImperial bool

	// Radar viewport and animation
	radarZoom        int
	radarCenterLat   float64
	radarCenterLon   float64
	radarFrameIndex  int
	radarAnimating   bool
	radarGeneration  int // Incremented when new radar data loads to invalidate old ticks
	radarLegendIndex int // Cycles through precipitation type legends
}

// Messages
type weatherLoadedMsg struct {
	weather *api.WeatherData
}

type locationDetectedMsg struct {
	location location.Location
}

type searchResultsMsg struct {
	results []api.GeoLocation
}

type radarLoadedMsg struct {
	radar *api.RadarData
}

type errMsg struct {
	err error
}

type radarTickMsg struct {
	generation int // Which radar generation this tick belongs to
}

// radarTick returns a command that ticks the radar animation
func radarTick(generation int) tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return radarTickMsg{generation: generation}
	})
}

// NewModel creates a new Model with default values
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search city..."
	ti.CharLimit = 100
	ti.Width = 40

	latInput := textinput.New()
	latInput.Placeholder = "e.g., 40.7128"
	latInput.CharLimit = 20
	latInput.Width = 20

	lonInput := textinput.New()
	lonInput.Placeholder = "e.g., -74.0060"
	lonInput.CharLimit = 20
	lonInput.Width = 20

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	cfg, _ := config.Load()

	useImperial := false
	if cfg != nil {
		useImperial = cfg.UseFahrenheit
	}

	return Model{
		activeView:     ViewCurrent,
		mode:           ModeNormal,
		searchInput:    ti,
		latInput:       latInput,
		lonInput:       lonInput,
		focusLat:       true,
		spinner:        s,
		apiClient:      api.NewClient(),
		keys:           DefaultKeyMap(),
		config:         cfg,
		useImperial:    useImperial,
		radarAnimating: true, // Start with animation enabled
		radarZoom:      6,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Use saved default location if available, otherwise detect via IP
	if m.config != nil && m.config.DefaultLocation != nil && m.config.DefaultLocation.IsValid() {
		loc := *m.config.DefaultLocation
		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg { return locationDetectedMsg{loc} },
		)
	}
	return tea.Batch(
		m.spinner.Tick,
		detectLocation(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle mode-specific keys first
		switch m.mode {
		case ModeSearch:
			return m.handleSearchMode(msg)
		case ModeCoordinates:
			return m.handleCoordinatesMode(msg)
		case ModeSavedLocations:
			return m.handleSavedLocationsMode(msg)
		case ModeHelp:
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				m.mode = ModeNormal
				return m, nil
			}
		}

		// Normal mode keys
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Radar pan (intercept arrow keys when on radar view)
		case "up":
			if m.activeView == ViewRadar {
				return m.radarPan(0, -1)
			}
		case "down":
			if m.activeView == ViewRadar {
				return m.radarPan(0, 1)
			}
		case "right":
			if m.activeView == ViewRadar {
				return m.radarPan(1, 0)
			}
			m.activeView = (m.activeView + 1) % 4
			return m.ensureRadarView()
		case "left":
			if m.activeView == ViewRadar {
				return m.radarPan(-1, 0)
			}
			m.activeView = (m.activeView + 3) % 4
			return m.ensureRadarView()

		// Radar zoom
		case "+", "=":
			if m.activeView == ViewRadar && m.radarZoom < 12 {
				m.radarZoom++
				m.radar = nil
				m.loading = true
				return m, m.radarViewCmd()
			}
		case "-":
			if m.activeView == ViewRadar && m.radarZoom > 3 {
				m.radarZoom--
				m.radar = nil
				m.loading = true
				return m, m.radarViewCmd()
			}

		case "tab":
			m.activeView = (m.activeView + 1) % 4
			return m.ensureRadarView()
		case "shift+tab":
			m.activeView = (m.activeView + 3) % 4
			return m.ensureRadarView()
		case "1":
			m.activeView = ViewCurrent
		case "2":
			m.activeView = ViewHourly
		case "3":
			m.activeView = ViewDaily
		case "4":
			m.activeView = ViewRadar
			return m.ensureRadarView()
		case "s":
			m.mode = ModeSearch
			m.searchInput.Focus()
			m.searchResults = nil
			m.selectedIndex = 0
			return m, textinput.Blink
		case "l":
			m.mode = ModeCoordinates
			m.latInput.SetValue("")
			m.lonInput.SetValue("")
			m.focusLat = true
			m.latInput.Focus()
			return m, textinput.Blink
		case "r":
			if m.location.IsValid() {
				m.loading = true
				if m.activeView == ViewRadar {
					m.radar = nil
					return m, m.radarViewCmd()
				}
				return m, fetchWeather(m.apiClient, m.location.Latitude, m.location.Longitude)
			}
		case "ctrl+s":
			if m.location.IsValid() && m.config != nil {
				m.config.AddSavedLocation(m.location)
				m.config.SetDefaultLocation(m.location)
				m.config.Save()
			}
		case "?":
			m.mode = ModeHelp
		case "u":
			m.useImperial = !m.useImperial
			if m.config != nil {
				m.config.UseFahrenheit = m.useImperial
				m.config.Save()
			}
		case " ":
			// Toggle radar animation pause/play
			if m.activeView == ViewRadar {
				m.radarAnimating = !m.radarAnimating
				if m.radarAnimating {
					m.radarGeneration++
					return m, radarTick(m.radarGeneration)
				}
			}
		case "p":
			// Cycle precipitation legend type
			if m.activeView == ViewRadar {
				m.radarLegendIndex = (m.radarLegendIndex + 1) % components.NumRadarLegends()
			}
		}

	case weatherLoadedMsg:
		m.loading = false
		m.weather = msg.weather
		m.err = nil
		// If on radar view and radar was cleared, fetch radar data
		if m.activeView == ViewRadar && m.radar == nil && m.location.IsValid() {
			m.loading = true
			return m, m.radarViewCmd()
		}

	case radarLoadedMsg:
		m.loading = false
		m.radar = msg.radar
		m.radarFrameIndex = 0
		m.radarGeneration++ // Increment generation to invalidate any pending ticks from old data
		m.err = nil
		// Start animation if enabled
		if m.radarAnimating && m.activeView == ViewRadar {
			return m, radarTick(m.radarGeneration)
		}

	case locationDetectedMsg:
		m.location = msg.location
		m.resetRadarForLocation(msg.location.Latitude, msg.location.Longitude)
		m.loading = true
		return m, fetchWeather(m.apiClient, m.location.Latitude, m.location.Longitude)

	case searchResultsMsg:
		m.loading = false
		m.searchResults = msg.results
		m.selectedIndex = 0

	case errMsg:
		m.loading = false
		m.err = msg.err

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case radarTickMsg:
		// Advance radar animation frame, but only if this tick is from the current generation
		if msg.generation != m.radarGeneration {
			// Stale tick from old radar data, ignore it
			return m, nil
		}
		if m.activeView == ViewRadar && m.radarAnimating && m.radar != nil && len(m.radar.RainFrames) > 0 {
			m.radarFrameIndex = (m.radarFrameIndex + 1) % len(m.radar.RainFrames)
			return m, radarTick(m.radarGeneration)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchInput.Blur()
		return m, nil
	case "enter":
		if len(m.searchResults) == 0 {
			// Perform search
			query := m.searchInput.Value()
			if query != "" {
				m.loading = true
				return m, searchLocation(m.apiClient, query)
			}
		} else {
			// Select result
			if m.selectedIndex < len(m.searchResults) {
				result := m.searchResults[m.selectedIndex]
				m.location = location.Location{
					Name:      result.Name,
					Latitude:  result.Latitude,
					Longitude: result.Longitude,
					Country:   result.Country,
					Admin1:    result.Admin1,
				}
				m.mode = ModeNormal
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.searchResults = nil
				m.resetRadarForLocation(result.Latitude, result.Longitude)
				m.loading = true
				return m, fetchWeather(m.apiClient, m.location.Latitude, m.location.Longitude)
			}
		}
	case "up":
		if len(m.searchResults) > 0 && m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil
	case "down":
		if m.selectedIndex < len(m.searchResults)-1 {
			m.selectedIndex++
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) handleCoordinatesMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "tab":
		m.focusLat = !m.focusLat
		if m.focusLat {
			m.latInput.Focus()
			m.lonInput.Blur()
		} else {
			m.lonInput.Focus()
			m.latInput.Blur()
		}
		return m, textinput.Blink
	case "enter":
		lat, latErr := strconv.ParseFloat(m.latInput.Value(), 64)
		lon, lonErr := strconv.ParseFloat(m.lonInput.Value(), 64)
		if latErr == nil && lonErr == nil {
			m.location = location.Location{
				Name:      fmt.Sprintf("%.4f, %.4f", lat, lon),
				Latitude:  lat,
				Longitude: lon,
			}
			m.mode = ModeNormal
			m.resetRadarForLocation(lat, lon)
			m.loading = true
			return m, fetchWeather(m.apiClient, lat, lon)
		}
		m.err = fmt.Errorf("invalid coordinates")
		return m, nil
	}

	var cmd tea.Cmd
	if m.focusLat {
		m.latInput, cmd = m.latInput.Update(msg)
	} else {
		m.lonInput, cmd = m.lonInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleSavedLocationsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil
	case "down":
		if m.config != nil && m.selectedIndex < len(m.config.SavedLocations)-1 {
			m.selectedIndex++
		}
		return m, nil
	case "enter":
		if m.config != nil && m.selectedIndex < len(m.config.SavedLocations) {
			loc := m.config.SavedLocations[m.selectedIndex]
			m.location = loc
			m.mode = ModeNormal
			m.resetRadarForLocation(loc.Latitude, loc.Longitude)
			m.loading = true
			return m, fetchWeather(m.apiClient, loc.Latitude, loc.Longitude)
		}
	case "d":
		if m.config != nil && m.selectedIndex < len(m.config.SavedLocations) {
			m.config.RemoveSavedLocation(m.selectedIndex)
			m.config.Save()
			if m.selectedIndex >= len(m.config.SavedLocations) && m.selectedIndex > 0 {
				m.selectedIndex--
			}
		}
		return m, nil
	}
	return m, nil
}

// resetRadarForLocation updates the radar center and invalidates cached radar data.
func (m *Model) resetRadarForLocation(lat, lon float64) {
	m.radarCenterLat = lat
	m.radarCenterLon = lon
	m.radar = nil
}

// Commands
func detectLocation() tea.Cmd {
	return func() tea.Msg {
		loc, err := location.DetectFromIP()
		if err != nil {
			return errMsg{err}
		}
		return locationDetectedMsg{loc}
	}
}

func fetchWeather(client *api.Client, lat, lon float64) tea.Cmd {
	return func() tea.Msg {
		weather, err := client.FetchWeather(lat, lon)
		if err != nil {
			return errMsg{err}
		}
		return weatherLoadedMsg{weather}
	}
}

func fetchRadar(client *api.Client, lat, lon float64, zoom, viewWidth, viewHeight int) tea.Cmd {
	return func() tea.Msg {
		radar, err := client.FetchRadar(lat, lon, zoom, viewWidth, viewHeight)
		if err != nil {
			return errMsg{err}
		}
		return radarLoadedMsg{radar}
	}
}

// radarDisplaySize returns the clamped display dimensions for the radar viewport.
func (m Model) radarDisplaySize() (width, height int) {
	width = m.width
	if width < 10 {
		width = 10
	}
	height = m.height - 10
	if height < 10 {
		height = 10
	}
	return
}

// radarViewCmd creates a command to fetch radar tiles for the current viewport
func (m Model) radarViewCmd() tea.Cmd {
	w, h := m.radarDisplaySize()
	return fetchRadar(m.apiClient, m.radarCenterLat, m.radarCenterLon, m.radarZoom, w, h)
}

// radarPan pans the radar view by the given direction (dx, dy each -1, 0, or 1)
func (m Model) radarPan(dx, dy int) (tea.Model, tea.Cmd) {
	_, displayHeight := m.radarDisplaySize()
	// Pan by 1/4 of the viewport in braille pixels
	panH := float64(m.width) / 2.0 // charWidth * 2 / 4
	panV := float64(displayHeight) // charHeight * 4 / 4
	m.radarCenterLat, m.radarCenterLon = api.PanCenter(
		m.radarCenterLat, m.radarCenterLon, m.radarZoom,
		float64(dx)*panH, float64(dy)*panV)
	m.radar = nil
	m.loading = true
	return m, m.radarViewCmd()
}

// ensureRadarView loads radar data if needed when switching to the radar view
func (m Model) ensureRadarView() (tea.Model, tea.Cmd) {
	if m.activeView == ViewRadar {
		if m.radar == nil && m.location.IsValid() {
			m.loading = true
			return m, m.radarViewCmd()
		}
		if m.radarAnimating && m.radar != nil {
			m.radarGeneration++
			return m, radarTick(m.radarGeneration)
		}
	}
	return m, nil
}

func searchLocation(client *api.Client, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.SearchLocation(query)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{results}
	}
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Main content based on mode
	switch m.mode {
	case ModeSearch:
		b.WriteString(m.renderSearchView())
	case ModeCoordinates:
		b.WriteString(m.renderCoordinatesView())
	case ModeSavedLocations:
		b.WriteString(m.renderSavedLocationsView())
	case ModeHelp:
		b.WriteString(m.renderHelpView())
	default:
		b.WriteString(m.renderMainView())
	}

	// Footer
	b.WriteString("\n\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("wxterm")
	loc := ""
	if m.location.IsValid() {
		loc = locationStyle.Render(m.location.DisplayName())
	}

	tabs := m.renderTabs()

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		"  ",
		loc,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
	)
}

func (m Model) renderTabs() string {
	tabs := []string{"Current", "Hourly", "Daily", "Radar"}
	var rendered []string

	for i, tab := range tabs {
		if ViewType(i) == m.activeView {
			rendered = append(rendered, activeTabStyle.Render(tab))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderFooter() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	units := "°C"
	if m.useImperial {
		units = "°F"
	}

	help := helpStyle.Render("tab: switch view • s: search • u: units [" + units + "] • r: refresh • ?: help • q: quit")
	return help
}
