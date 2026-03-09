# wxterm

A terminal-based weather application built in Go. View current conditions, hourly and daily forecasts, and animated precipitation radar — all from your terminal.

> **Personal Note:** This project began as one of my first from-scratch builds using Claude and has been a personal idea I’ve wanted to explore for quite some time. While the current state is functional, I do not yet consider it fully complete—this initial commit should be viewed as roughly **v0.99**. The next milestone is to implement an improved map renderer with a look and feel closer to the approach used in [mapscii](https://github.com/rastapasta/mapscii/).


![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)

## Features

- **Current Weather** — Temperature, "feels like", humidity, wind, pressure, cloud cover, precipitation, visibility, and sunrise/sunset times with ASCII-art weather icons
- **Hourly Forecast** — 24-hour forecast table with temperature, humidity, wind speed, and precipitation
- **Daily Forecast** — 7-day outlook with high/low temps, precipitation probability, and wind speeds
- **Animated Radar** — Live precipitation radar rendered with Unicode Braille characters, featuring pan, zoom, and animation playback controls
- **Location Search** — Search by city name or enter coordinates manually; save favorite locations for quick access
- **Auto-Detection** — Automatically detects your location via IP geolocation on startup
- **Unit Toggle** — Switch between metric (°C) and imperial (°F) units on the fly
- **Cross-Platform** — Works on Windows, macOS, and Linux
- **No API Keys Required** — Uses free, open APIs (Open-Meteo, RainViewer, OpenStreetMap)
- **Mouse Support** — Full mouse cell motion support via the Bubble Tea framework

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later
- A terminal emulator with Unicode and 256-color support (most modern terminals work)

### Build from Source

```bash
git clone https://github.com/psteger/wxterm.git
cd wxterm
go mod download
```

**Windows:**
```bash
go build -o wxterm.exe .
```

**macOS / Linux:**
```bash
go build -o wxterm .
```

### Run Directly (Development)

```bash
go run .
```

## Usage

Launch wxterm by running the built binary:

```bash
./wxterm        # macOS / Linux
wxterm.exe      # Windows
```

The app starts in full-screen mode and auto-detects your location. No command-line arguments are needed.

### Navigation

| Key                 | Action                                   |
| ------------------- | ---------------------------------------- |
| `Tab` / `Shift+Tab` | Cycle through views                      |
| `1` `2` `3` `4`     | Jump to Current / Hourly / Daily / Radar |
| `s`                 | Search for a city                        |
| `Ctrl+S`            | Save current location                    |
| `l`                 | Enter coordinates manually               |
| `r`                 | Refresh weather data                     |
| `u`                 | Toggle Metric / Imperial units           |
| `?`                 | Show help screen                         |
| `q` / `Ctrl+C`      | Quit                                     |

### Radar Controls

| Key             | Action                                             |
| --------------- | -------------------------------------------------- |
| Arrow keys      | Pan the map                                        |
| `+` / `=`       | Zoom in (max 12)                                   |
| `-`             | Zoom out (min 3)                                   |
| `Space`         | Pause / resume animation                           |
| `p`             | Cycle precipitation legend (Rain / Snow / Frz/Mix) |

## Configuration

wxterm stores its configuration at:

**Windows:**
```
%AppData%\wxterm\wxterm.json
```

**macOS / Linux:**
```
~/.config/wxterm/wxterm.json
```

### Options

| Field              | Type      | Description                                |
| ------------------ | --------- | ------------------------------------------ |
| `default_location` | object    | Fallback location if auto-detection fails  |
| `saved_locations`  | array     | List of saved favorite locations           |
| `use_fahrenheit`   | boolean   | `true` for °F, `false` for °C (default)    |

### Example Configuration

```json
{
  "default_location": {
    "name": "New York",
    "latitude": 40.7128,
    "longitude": -74.006,
    "country": "United States",
    "admin1": "New York"
  },
  "saved_locations": [
    {
      "name": "San Francisco",
      "latitude": 37.7749,
      "longitude": -122.4194,
      "country": "United States",
      "admin1": "California"
    }
  ],
  "use_fahrenheit": false
}
```

## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create a branch** for your feature or fix: `git checkout -b my-feature`
3. **Make your changes** and add tests where appropriate
4. **Run tests** to make sure everything passes:
   ```bash
   go test ./...
   ```
5. **Commit** with a clear message describing what you changed
6. **Open a pull request** against `main`

### Reporting Issues

If you find a bug or have a feature request, please [open an issue](https://github.com/psteger/wxterm/issues) with:

- A clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Your OS and terminal emulator

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — UI components (spinner, text input)
- [Open-Meteo](https://open-meteo.com/) — Free weather forecast and geocoding API
- [RainViewer](https://www.rainviewer.com/) — Precipitation radar tile data
- [OpenStreetMap](https://www.openstreetmap.org/) — Base map tiles
- [ip-api.com](http://ip-api.com/) — IP-based geolocation
