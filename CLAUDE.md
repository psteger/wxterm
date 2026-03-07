# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
- `go build -o wxterm.exe .` - Build the application (Windows)
- `go build -o wxterm .` - Build the application (Linux/macOS)
- `go run .` - Run directly for development

### Test
- `go test ./...` - Run all tests
- `go test ./... -v` - Run all tests with verbose output
- `go test -run TestName ./path/to/package` - Run a specific test

### Dependencies
- `go mod download` - Download dependencies
- `go mod tidy` - Clean up dependencies

## Architecture

wxterm is a terminal-based weather application built with the Bubble Tea TUI framework. It displays weather data from Open-Meteo API with ASCII-art radar visualization.

### Project Layout

- `main.go` - Entry point; creates Bubble Tea program with alt-screen and mouse support
- `internal/ui/` - TUI layer: model, views, keys, styles
- `internal/api/` - HTTP clients for Open-Meteo (weather + geocoding) and RainViewer (radar tiles)
- `internal/ui/components/` - View renderers for each tab (current, hourly, daily, radar, location)
- `internal/geo/` - Static geographic data (world borders/coastlines, major cities) for radar overlay
- `internal/location/` - Location types and IP-based geolocation via ip-api.com
- `internal/config/` - User preferences stored in `~/.config/wxterm/wxterm.json`

### Bubble Tea Pattern

The app follows the standard Elm architecture (`Init` / `Update` / `View`):

- **Model** (`internal/ui/model.go`) holds all application state including the active view (`ViewCurrent`, `ViewHourly`, `ViewDaily`, `ViewRadar`) and interaction mode (`ModeNormal`, `ModeSearch`, `ModeCoordinates`, `ModeSavedLocations`, `ModeHelp`)
- **Messages** are typed structs (`weatherLoadedMsg`, `radarLoadedMsg`, `searchResultsMsg`, `locationDetectedMsg`, `errMsg`, `radarTickMsg`) returned from async commands
- **Commands** are `tea.Cmd` functions that perform async work (API calls, location detection) and return messages
- Radar animation uses a generation counter (`radarGeneration`) to invalidate stale `radarTickMsg` ticks when new data loads

### Radar Rendering

The radar view converts tile images to terminal output using Unicode Braille characters (U+2800..U+28FF). Each braille character encodes a 2x4 pixel grid, so one terminal cell represents 2px wide x 4px tall of tile imagery. The rendering pipeline in `internal/ui/components/radar.go`:

1. Map tile pixels are thresholded to binary (dark = dot on) to form braille patterns
2. Rain overlay pixels are sampled at cell centers; non-transparent rain colors become ANSI background colors
3. Cells with both map features and rain get white braille on colored background

Tile fetching (`internal/api/radar.go`) uses Web Mercator projection (`LatLonToTileXY` / `TileXYToLatLon`) with a thread-safe `TileCache`. RainViewer only supports up to zoom 7, so higher zooms fetch at zoom 7 and nearest-neighbor upscale.

### Data Flow

1. App initializes and detects location via IP geolocation
2. Weather data fetched from Open-Meteo API on location change
3. Radar data (RainViewer tiles) fetched on-demand when Radar view selected
4. Radar rendering layers: map tiles (braille) -> precipitation (colored backgrounds) -> legend

### External APIs
- **Open-Meteo** (`api.open-meteo.com`) - Weather forecasts and geocoding (no API key required)
- **RainViewer** (`rainviewer.com`) - Radar precipitation tiles (no API key required)
- **ip-api.com** - IP-based location detection (no API key required)
