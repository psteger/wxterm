package api

import "time"

// WeatherData contains all weather information from Open-Meteo
type WeatherData struct {
	Latitude  float64
	Longitude float64
	Timezone  string
	Current   CurrentWeather
	Hourly    HourlyForecast
	Daily     DailyForecast
}

// CurrentWeather represents current conditions
type CurrentWeather struct {
	Time             time.Time
	Temperature      float64
	FeelsLike        float64
	Humidity         int
	WindSpeed        float64
	WindDirection    int
	WeatherCode      int
	IsDay            bool
	Precipitation    float64
	CloudCover       int
	Pressure         float64
	Visibility       float64
}

// HourlyForecast contains 24-hour forecast data
type HourlyForecast struct {
	Time          []time.Time
	Temperature   []float64
	FeelsLike     []float64
	Humidity      []int
	WeatherCode   []int
	WindSpeed     []float64
	Precipitation []float64
	CloudCover    []int
}

// DailyForecast contains 7-day forecast data
type DailyForecast struct {
	Time              []time.Time
	WeatherCode       []int
	TemperatureMax    []float64
	TemperatureMin    []float64
	PrecipitationProb []int
	WindSpeedMax      []float64
	WindSpeedMean     []float64
	Sunrise           []time.Time
	Sunset            []time.Time
}

// GeoLocation represents a location from geocoding search
type GeoLocation struct {
	Name      string
	Latitude  float64
	Longitude float64
	Country   string
	Admin1    string // State/Province
	Timezone  string
}

// WeatherCodeDescription returns a human-readable description for WMO weather codes
func WeatherCodeDescription(code int) string {
	descriptions := map[int]string{
		0:  "Clear sky",
		1:  "Mainly clear",
		2:  "Partly cloudy",
		3:  "Overcast",
		45: "Foggy",
		48: "Depositing rime fog",
		51: "Light drizzle",
		53: "Moderate drizzle",
		55: "Dense drizzle",
		56: "Light freezing drizzle",
		57: "Dense freezing drizzle",
		61: "Slight rain",
		63: "Moderate rain",
		65: "Heavy rain",
		66: "Light freezing rain",
		67: "Heavy freezing rain",
		71: "Slight snow",
		73: "Moderate snow",
		75: "Heavy snow",
		77: "Snow grains",
		80: "Slight rain showers",
		81: "Moderate rain showers",
		82: "Violent rain showers",
		85: "Slight snow showers",
		86: "Heavy snow showers",
		95: "Thunderstorm",
		96: "Thunderstorm with slight hail",
		99: "Thunderstorm with heavy hail",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Unknown"
}

// WeatherCodeCategory returns a simplified category for ASCII art selection
func WeatherCodeCategory(code int) string {
	switch {
	case code == 0:
		return "sunny"
	case code >= 1 && code <= 3:
		return "cloudy"
	case code >= 45 && code <= 48:
		return "foggy"
	case code >= 51 && code <= 67:
		return "rainy"
	case code >= 71 && code <= 77:
		return "snowy"
	case code >= 80 && code <= 82:
		return "rainy"
	case code >= 85 && code <= 86:
		return "snowy"
	case code >= 95:
		return "stormy"
	default:
		return "cloudy"
	}
}
