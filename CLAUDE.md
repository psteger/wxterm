# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
- `go build -o weatherterm.exe .` - Build the application (Windows)
- `go build -o weatherterm .` - Build the application (Linux/macOS)
- `go run .` - Run directly for development

### Test
- `go test ./...` - Run all tests
- `go test ./... -v` - Run all tests with verbose output
- `go test -run TestName ./path/to/package` - Run a specific test

### Dependencies
- `go mod download` - Download dependencies
- `go mod tidy` - Clean up dependencies

## Architecture

WeatherTerm is a terminal-based weather application built with the Bubble Tea TUI framework. It displays weather data from Open-Meteo API with ASCII-art radar visualization.

### Core Components

**Entry Point & TUI Framework**
- `main.go` - Initializes Bubble Tea program with alt-screen and mouse support
- `internal/ui/model.go` - Main application state (Model) implementing tea.Model interface with view types (Current, Hourly, Daily, Radar) and interaction modes (Normal, Search, Coordinates, SavedLocations, Help)
- `internal/ui/views.go` - View rendering dispatch and modal views (search, coordinates, help)
- `internal/ui/keys.go` - Keyboard bindings
- `internal/ui/styles.go` - Lipgloss styling definitions

**API Layer** (`internal/api/`)
- `client.go` - HTTP client for Open-Meteo weather and geocoding APIs
- `weather.go` - Weather data fetching with hourly/daily forecasts
- `geocoding.go` - Location search functionality
- `radar.go` - Radar tile fetching from RainViewer API with Web Mercator coordinate conversion
- `types.go` - Weather data structures and WMO weather code mappings

**UI Components** (`internal/ui/components/`)
- `current.go` - Current conditions display with ASCII weather art
- `hourly.go` - 24-hour forecast with temperature graph
- `daily.go` - 7-day forecast view
- `radar.go` - ASCII radar rendering with layered precipitation, coastlines, borders, and city markers
- `ascii.go` - Weather condition ASCII art
- `location.go` - Location search results and saved locations UI

**Geographic Data** (`internal/geo/`)
- `borders.go` - Simplified world coastline and border line segments for radar overlay
- `cities.go` - Major world cities database for radar labels

**Location & Config**
- `internal/location/` - Location types and IP-based geolocation (ip-api.com)
- `internal/config/` - User preferences stored in `~/.config/weatherterm/weatherterm.json`

### Data Flow

1. App initializes and detects location via IP geolocation
2. Weather data fetched from Open-Meteo API on location change
3. Radar data (RainViewer tiles) fetched on-demand when Radar view selected
4. Radar rendering layers: geography (borders/coasts) -> precipitation -> city labels

### External APIs
- **Open-Meteo** (`api.open-meteo.com`) - Weather forecasts and geocoding
- **RainViewer** (`rainviewer.com`) - Radar precipitation tiles
- **ip-api.com** - IP-based location detection
