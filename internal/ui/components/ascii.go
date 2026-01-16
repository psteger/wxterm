package components

// WeatherIcon returns ASCII art for a weather condition
func WeatherIcon(category string, isDay bool) string {
	switch category {
	case "sunny":
		if isDay {
			return sunnyDay
		}
		return clearNight
	case "cloudy":
		return cloudy
	case "foggy":
		return foggy
	case "rainy":
		return rainy
	case "snowy":
		return snowy
	case "stormy":
		return stormy
	default:
		return cloudy
	}
}

// SmallWeatherIcon returns a compact ASCII icon for tables/lists
func SmallWeatherIcon(category string, isDay bool) string {
	switch category {
	case "sunny":
		if isDay {
			return "  *  "
		}
		return "  C  "
	case "cloudy":
		return " .-. "
	case "foggy":
		return " = = "
	case "rainy":
		return " ' ' "
	case "snowy":
		return " * * "
	case "stormy":
		return " /\\/ "
	default:
		return " .-. "
	}
}

const sunnyDay = `    \   /
     .-.
  - (   ) -
     '-'
    /   \    `

const clearNight = `       *
    *
  *     *
     C
    *   *    `

const cloudy = `
     .--.
  .-(    ).
 (___.__)__)
             `

const foggy = `
 _ - _ - _ -
  _ - _ - _
 _ - _ - _ -
             `

const rainy = `     .-.
    (   ).
   (___(__)
    ' ' ' '
   ' ' ' '   `

const snowy = `     .-.
    (   ).
   (___(__)
    * * * *
   * * * *   `

const stormy = `     .-.
    (   ).
   (___(__)
   ⚡' '⚡
   ' ' ' '   `

// WindDirection returns an arrow for wind direction
func WindDirection(degrees int) string {
	// Normalize to 0-360
	degrees = degrees % 360
	if degrees < 0 {
		degrees += 360
	}

	// 8-point compass
	switch {
	case degrees < 23 || degrees >= 338:
		return "N"
	case degrees < 68:
		return "NE"
	case degrees < 113:
		return "E"
	case degrees < 158:
		return "SE"
	case degrees < 203:
		return "S"
	case degrees < 248:
		return "SW"
	case degrees < 293:
		return "W"
	default:
		return "NW"
	}
}
