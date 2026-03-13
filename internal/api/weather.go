package api

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// openMeteoResponse represents the raw API response
type openMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Time                string  `json:"time"`
		Temperature2m       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		RelativeHumidity2m  int     `json:"relative_humidity_2m"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
		WindDirection10m    int     `json:"wind_direction_10m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               int     `json:"is_day"`
		Precipitation       float64 `json:"precipitation"`
		CloudCover          int     `json:"cloud_cover"`
		PressureMsl         float64 `json:"pressure_msl"`
		Visibility          float64 `json:"visibility"`
	} `json:"current"`
	Hourly struct {
		Time                []string  `json:"time"`
		Temperature2m       []float64 `json:"temperature_2m"`
		ApparentTemperature []float64 `json:"apparent_temperature"`
		RelativeHumidity2m  []int     `json:"relative_humidity_2m"`
		WeatherCode         []int     `json:"weather_code"`
		WindSpeed10m        []float64 `json:"wind_speed_10m"`
		Precipitation            []float64 `json:"precipitation"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
		CloudCover               []int     `json:"cloud_cover"`
		WindDirection10m         []int     `json:"wind_direction_10m"`
		WindGusts10m             []float64 `json:"wind_gusts_10m"`
		UVIndex                  []float64 `json:"uv_index"`
		Visibility               []float64 `json:"visibility"`
		SurfacePressure          []float64 `json:"surface_pressure"`
	} `json:"hourly"`
	Daily struct {
		Time                 []string  `json:"time"`
		WeatherCode          []int     `json:"weather_code"`
		Temperature2mMax     []float64 `json:"temperature_2m_max"`
		Temperature2mMin     []float64 `json:"temperature_2m_min"`
		PrecipitationProbMax []int     `json:"precipitation_probability_max"`
		WindSpeed10mMax      []float64 `json:"wind_speed_10m_max"`
		WindSpeed10mMean     []float64 `json:"wind_speed_10m_mean"`
		PrecipitationSum   []float64 `json:"precipitation_sum"`
		UVIndexMax         []float64 `json:"uv_index_max"`
		WindGusts10mMax    []float64 `json:"wind_gusts_10m_max"`
		DaylightDuration   []float64 `json:"daylight_duration"`
		PrecipitationHours []float64 `json:"precipitation_hours"`
		Sunrise            []string  `json:"sunrise"`
		Sunset             []string  `json:"sunset"`
	} `json:"daily"`
}

// FetchWeather retrieves weather data for a given location
func (c *Client) FetchWeather(lat, lon float64) (*WeatherData, error) {
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%.4f", lat))
	params.Set("longitude", fmt.Sprintf("%.4f", lon))
	params.Set("current", "temperature_2m,relative_humidity_2m,apparent_temperature,is_day,precipitation,weather_code,cloud_cover,pressure_msl,wind_speed_10m,wind_direction_10m,visibility")
	params.Set("hourly", "temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,precipitation_probability,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,wind_gusts_10m,uv_index,visibility,surface_pressure")
	params.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,wind_speed_10m_mean,sunrise,sunset,precipitation_sum,uv_index_max,wind_gusts_10m_max,daylight_duration,precipitation_hours")
	params.Set("timezone", "auto")
	params.Set("forecast_days", "7")

	reqURL := fmt.Sprintf("%s?%s", weatherBaseURL, params.Encode())

	var resp openMeteoResponse
	if err := c.get(reqURL, &resp); err != nil {
		return nil, err
	}

	return parseWeatherResponse(&resp)
}

func parseWeatherResponse(resp *openMeteoResponse) (*WeatherData, error) {
	data := &WeatherData{
		Latitude:  resp.Latitude,
		Longitude: resp.Longitude,
		Timezone:  resp.Timezone,
	}

	// Parse current weather
	currentTime, _ := time.Parse("2006-01-02T15:04", resp.Current.Time)
	data.Current = CurrentWeather{
		Time:          currentTime,
		Temperature:   resp.Current.Temperature2m,
		FeelsLike:     resp.Current.ApparentTemperature,
		Humidity:      resp.Current.RelativeHumidity2m,
		WindSpeed:     resp.Current.WindSpeed10m,
		WindDirection: resp.Current.WindDirection10m,
		WeatherCode:   resp.Current.WeatherCode,
		IsDay:         resp.Current.IsDay == 1,
		Precipitation: resp.Current.Precipitation,
		CloudCover:    resp.Current.CloudCover,
		Pressure:      resp.Current.PressureMsl,
		Visibility:    resp.Current.Visibility,
	}

	// Parse hourly forecast (limit to 24 hours, starting from current hour)
	// Find the starting index based on current time
	startIdx := 0
	now := time.Now()
	for i, timeStr := range resp.Hourly.Time {
		t, _ := time.Parse("2006-01-02T15:04", timeStr)
		if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() && t.Hour() >= now.Hour() {
			startIdx = i
			break
		}
		// Also handle case where we've moved past today
		if t.After(now) {
			startIdx = i
			break
		}
	}

	hourlyLen := min(24, len(resp.Hourly.Time)-startIdx)
	data.Hourly = HourlyForecast{
		Time:          make([]time.Time, hourlyLen),
		Temperature:   make([]float64, hourlyLen),
		FeelsLike:     make([]float64, hourlyLen),
		Humidity:      make([]int, hourlyLen),
		WeatherCode:   make([]int, hourlyLen),
		WindSpeed:          make([]float64, hourlyLen),
		WindDirection:      make([]int, hourlyLen),
		WindGusts:          make([]float64, hourlyLen),
		Precipitation:     make([]float64, hourlyLen),
		PrecipitationProb: make([]int, hourlyLen),
		CloudCover:        make([]int, hourlyLen),
		UVIndex:           make([]float64, hourlyLen),
		Visibility:        make([]float64, hourlyLen),
		SurfacePressure:   make([]float64, hourlyLen),
	}

	for i := 0; i < hourlyLen; i++ {
		srcIdx := startIdx + i
		t, _ := time.Parse("2006-01-02T15:04", resp.Hourly.Time[srcIdx])
		data.Hourly.Time[i] = t
		data.Hourly.Temperature[i] = resp.Hourly.Temperature2m[srcIdx]
		data.Hourly.FeelsLike[i] = resp.Hourly.ApparentTemperature[srcIdx]
		data.Hourly.Humidity[i] = resp.Hourly.RelativeHumidity2m[srcIdx]
		data.Hourly.WeatherCode[i] = resp.Hourly.WeatherCode[srcIdx]
		data.Hourly.WindSpeed[i] = resp.Hourly.WindSpeed10m[srcIdx]
		if srcIdx < len(resp.Hourly.WindDirection10m) {
			data.Hourly.WindDirection[i] = resp.Hourly.WindDirection10m[srcIdx]
		}
		if srcIdx < len(resp.Hourly.WindGusts10m) {
			data.Hourly.WindGusts[i] = resp.Hourly.WindGusts10m[srcIdx]
		}
		data.Hourly.Precipitation[i] = resp.Hourly.Precipitation[srcIdx]
		if srcIdx < len(resp.Hourly.PrecipitationProbability) {
			data.Hourly.PrecipitationProb[i] = resp.Hourly.PrecipitationProbability[srcIdx]
		}
		data.Hourly.CloudCover[i] = resp.Hourly.CloudCover[srcIdx]
		if srcIdx < len(resp.Hourly.UVIndex) {
			data.Hourly.UVIndex[i] = resp.Hourly.UVIndex[srcIdx]
		}
		if srcIdx < len(resp.Hourly.Visibility) {
			data.Hourly.Visibility[i] = resp.Hourly.Visibility[srcIdx]
		}
		if srcIdx < len(resp.Hourly.SurfacePressure) {
			data.Hourly.SurfacePressure[i] = resp.Hourly.SurfacePressure[srcIdx]
		}
	}

	// Parse daily forecast
	dailyLen := len(resp.Daily.Time)
	data.Daily = DailyForecast{
		Time:              make([]time.Time, dailyLen),
		WeatherCode:       make([]int, dailyLen),
		TemperatureMax:    make([]float64, dailyLen),
		TemperatureMin:    make([]float64, dailyLen),
		PrecipitationProb: make([]int, dailyLen),
		PrecipitationSum:  make([]float64, dailyLen),
		PrecipTiming:      make([]string, dailyLen),
		WindSpeedMax:      make([]float64, dailyLen),
		WindSpeedMean:     make([]float64, dailyLen),
		WindGustsMax:      make([]float64, dailyLen),
		UVIndexMax:        make([]float64, dailyLen),
		DaylightDuration:  make([]float64, dailyLen),
		Sunrise:           make([]time.Time, dailyLen),
		Sunset:            make([]time.Time, dailyLen),
	}

	for i := 0; i < dailyLen; i++ {
		t, _ := time.Parse("2006-01-02", resp.Daily.Time[i])
		data.Daily.Time[i] = t
		data.Daily.WeatherCode[i] = resp.Daily.WeatherCode[i]
		data.Daily.TemperatureMax[i] = resp.Daily.Temperature2mMax[i]
		data.Daily.TemperatureMin[i] = resp.Daily.Temperature2mMin[i]
		if i < len(resp.Daily.PrecipitationProbMax) {
			data.Daily.PrecipitationProb[i] = resp.Daily.PrecipitationProbMax[i]
		}
		data.Daily.WindSpeedMax[i] = resp.Daily.WindSpeed10mMax[i]
		if i < len(resp.Daily.WindSpeed10mMean) {
			data.Daily.WindSpeedMean[i] = resp.Daily.WindSpeed10mMean[i]
		}
		if i < len(resp.Daily.WindGusts10mMax) {
			data.Daily.WindGustsMax[i] = resp.Daily.WindGusts10mMax[i]
		}
		if i < len(resp.Daily.UVIndexMax) {
			data.Daily.UVIndexMax[i] = resp.Daily.UVIndexMax[i]
		}
		if i < len(resp.Daily.PrecipitationSum) {
			data.Daily.PrecipitationSum[i] = resp.Daily.PrecipitationSum[i]
		}
		if i < len(resp.Daily.DaylightDuration) {
			data.Daily.DaylightDuration[i] = resp.Daily.DaylightDuration[i]
		}
		sunrise, _ := time.Parse("2006-01-02T15:04", resp.Daily.Sunrise[i])
		sunset, _ := time.Parse("2006-01-02T15:04", resp.Daily.Sunset[i])
		data.Daily.Sunrise[i] = sunrise
		data.Daily.Sunset[i] = sunset
	}

	// Compute precipitation timing from hourly data
	for i := 0; i < dailyLen; i++ {
		precipHours := 0.0
		if i < len(resp.Daily.PrecipitationHours) {
			precipHours = resp.Daily.PrecipitationHours[i]
		}
		data.Daily.PrecipTiming[i] = computePrecipTiming(resp, data.Daily.Time[i], precipHours)
	}

	return data, nil
}

func computePrecipTiming(resp *openMeteoResponse, dayDate time.Time, precipHours float64) string {
	if precipHours == 0 {
		return "None"
	}

	hasNight, hasMorning, hasAfternoon, hasEvening := false, false, false, false
	matched := false

	for j, timeStr := range resp.Hourly.Time {
		t, err := time.Parse("2006-01-02T15:04", timeStr)
		if err != nil {
			continue
		}
		if t.Year() == dayDate.Year() && t.YearDay() == dayDate.YearDay() {
			matched = true
			if j < len(resp.Hourly.Precipitation) && resp.Hourly.Precipitation[j] > 0 {
				switch {
				case t.Hour() < 6:
					hasNight = true
				case t.Hour() < 12:
					hasMorning = true
				case t.Hour() < 18:
					hasAfternoon = true
				default:
					hasEvening = true
				}
			}
		} else if matched {
			break
		}
	}

	count := 0
	if hasNight {
		count++
	}
	if hasMorning {
		count++
	}
	if hasAfternoon {
		count++
	}
	if hasEvening {
		count++
	}

	if count == 0 {
		if precipHours < 2 {
			return "Brief"
		}
		return "Scattered"
	}
	if count == 4 {
		return "All day"
	}

	var parts []string
	if hasNight {
		parts = append(parts, "Night")
	}
	if hasMorning {
		parts = append(parts, "AM")
	}
	if hasAfternoon {
		parts = append(parts, "PM")
	}
	if hasEvening {
		parts = append(parts, "Eve")
	}

	return strings.Join(parts, ", ")
}
