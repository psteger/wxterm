package geo

// City represents a city with its coordinates
type City struct {
	Name    string
	Lat     float64
	Lon     float64
	Country string
}

// MajorCities contains a list of major world cities
// This is a curated list of ~200 significant cities for radar overlay
var MajorCities = []City{
	// North America
	{"New York", 40.7128, -74.0060, "US"},
	{"Los Angeles", 34.0522, -118.2437, "US"},
	{"Chicago", 41.8781, -87.6298, "US"},
	{"Houston", 29.7604, -95.3698, "US"},
	{"Phoenix", 33.4484, -112.0740, "US"},
	{"Philadelphia", 39.9526, -75.1652, "US"},
	{"San Antonio", 29.4241, -98.4936, "US"},
	{"San Diego", 32.7157, -117.1611, "US"},
	{"Dallas", 32.7767, -96.7970, "US"},
	{"San Jose", 37.3382, -121.8863, "US"},
	{"Austin", 30.2672, -97.7431, "US"},
	{"Seattle", 47.6062, -122.3321, "US"},
	{"Denver", 39.7392, -104.9903, "US"},
	{"Boston", 42.3601, -71.0589, "US"},
	{"Miami", 25.7617, -80.1918, "US"},
	{"Atlanta", 33.7490, -84.3880, "US"},
	{"Minneapolis", 44.9778, -93.2650, "US"},
	{"Detroit", 42.3314, -83.0458, "US"},
	{"Portland", 45.5051, -122.6750, "US"},
	{"Las Vegas", 36.1699, -115.1398, "US"},
	{"Memphis", 35.1495, -90.0490, "US"},
	{"Louisville", 38.2527, -85.7585, "US"},
	{"Baltimore", 39.2904, -76.6122, "US"},
	{"Milwaukee", 43.0389, -87.9065, "US"},
	{"Albuquerque", 35.0844, -106.6504, "US"},
	{"Nashville", 36.1627, -86.7816, "US"},
	{"Kansas City", 39.0997, -94.5786, "US"},
	{"Indianapolis", 39.7684, -86.1581, "US"},
	{"Cleveland", 41.4993, -81.6944, "US"},
	{"New Orleans", 29.9511, -90.0715, "US"},
	{"Toronto", 43.6532, -79.3832, "CA"},
	{"Montreal", 45.5017, -73.5673, "CA"},
	{"Vancouver", 49.2827, -123.1207, "CA"},
	{"Calgary", 51.0447, -114.0719, "CA"},
	{"Ottawa", 45.4215, -75.6972, "CA"},
	{"Edmonton", 53.5461, -113.4938, "CA"},
	{"Winnipeg", 49.8951, -97.1384, "CA"},
	{"Mexico City", 19.4326, -99.1332, "MX"},
	{"Guadalajara", 20.6597, -103.3496, "MX"},
	{"Monterrey", 25.6866, -100.3161, "MX"},
	{"Tijuana", 32.5149, -117.0382, "MX"},

	// Europe
	{"London", 51.5074, -0.1278, "UK"},
	{"Paris", 48.8566, 2.3522, "FR"},
	{"Berlin", 52.5200, 13.4050, "DE"},
	{"Madrid", 40.4168, -3.7038, "ES"},
	{"Rome", 41.9028, 12.4964, "IT"},
	{"Amsterdam", 52.3676, 4.9041, "NL"},
	{"Vienna", 48.2082, 16.3738, "AT"},
	{"Barcelona", 41.3851, 2.1734, "ES"},
	{"Munich", 48.1351, 11.5820, "DE"},
	{"Milan", 45.4642, 9.1900, "IT"},
	{"Prague", 50.0755, 14.4378, "CZ"},
	{"Dublin", 53.3498, -6.2603, "IE"},
	{"Brussels", 50.8503, 4.3517, "BE"},
	{"Warsaw", 52.2297, 21.0122, "PL"},
	{"Budapest", 47.4979, 19.0402, "HU"},
	{"Stockholm", 59.3293, 18.0686, "SE"},
	{"Oslo", 59.9139, 10.7522, "NO"},
	{"Copenhagen", 55.6761, 12.5683, "DK"},
	{"Helsinki", 60.1699, 24.9384, "FI"},
	{"Lisbon", 38.7223, -9.1393, "PT"},
	{"Athens", 37.9838, 23.7275, "GR"},
	{"Zurich", 47.3769, 8.5417, "CH"},
	{"Geneva", 46.2044, 6.1432, "CH"},
	{"Hamburg", 53.5511, 9.9937, "DE"},
	{"Frankfurt", 50.1109, 8.6821, "DE"},
	{"Edinburgh", 55.9533, -3.1883, "UK"},
	{"Manchester", 53.4808, -2.2426, "UK"},
	{"Birmingham", 52.4862, -1.8904, "UK"},
	{"Glasgow", 55.8642, -4.2518, "UK"},
	{"Kyiv", 50.4501, 30.5234, "UA"},
	{"Bucharest", 44.4268, 26.1025, "RO"},
	{"Sofia", 42.6977, 23.3219, "BG"},
	{"Belgrade", 44.7866, 20.4489, "RS"},
	{"Zagreb", 45.8150, 15.9819, "HR"},

	// Asia
	{"Tokyo", 35.6762, 139.6503, "JP"},
	{"Delhi", 28.7041, 77.1025, "IN"},
	{"Shanghai", 31.2304, 121.4737, "CN"},
	{"Beijing", 39.9042, 116.4074, "CN"},
	{"Mumbai", 19.0760, 72.8777, "IN"},
	{"Osaka", 34.6937, 135.5023, "JP"},
	{"Seoul", 37.5665, 126.9780, "KR"},
	{"Bangkok", 13.7563, 100.5018, "TH"},
	{"Hong Kong", 22.3193, 114.1694, "HK"},
	{"Singapore", 1.3521, 103.8198, "SG"},
	{"Taipei", 25.0330, 121.5654, "TW"},
	{"Kuala Lumpur", 3.1390, 101.6869, "MY"},
	{"Jakarta", 6.2088, 106.8456, "ID"},
	{"Manila", 14.5995, 120.9842, "PH"},
	{"Ho Chi Minh", 10.8231, 106.6297, "VN"},
	{"Hanoi", 21.0278, 105.8342, "VN"},
	{"Bangalore", 12.9716, 77.5946, "IN"},
	{"Chennai", 13.0827, 80.2707, "IN"},
	{"Kolkata", 22.5726, 88.3639, "IN"},
	{"Hyderabad", 17.3850, 78.4867, "IN"},
	{"Shenzhen", 22.5431, 114.0579, "CN"},
	{"Guangzhou", 23.1291, 113.2644, "CN"},
	{"Chengdu", 30.5728, 104.0668, "CN"},
	{"Wuhan", 30.5928, 114.3055, "CN"},
	{"Tianjin", 39.3434, 117.3616, "CN"},
	{"Nagoya", 35.1815, 136.9066, "JP"},
	{"Busan", 35.1796, 129.0756, "KR"},

	// Middle East
	{"Dubai", 25.2048, 55.2708, "AE"},
	{"Istanbul", 41.0082, 28.9784, "TR"},
	{"Tehran", 35.6892, 51.3890, "IR"},
	{"Riyadh", 24.7136, 46.6753, "SA"},
	{"Baghdad", 33.3152, 44.3661, "IQ"},
	{"Tel Aviv", 32.0853, 34.7818, "IL"},
	{"Jerusalem", 31.7683, 35.2137, "IL"},
	{"Ankara", 39.9334, 32.8597, "TR"},
	{"Abu Dhabi", 24.4539, 54.3773, "AE"},
	{"Doha", 25.2854, 51.5310, "QA"},
	{"Kuwait City", 29.3759, 47.9774, "KW"},

	// Africa
	{"Cairo", 30.0444, 31.2357, "EG"},
	{"Lagos", 6.5244, 3.3792, "NG"},
	{"Johannesburg", 26.2041, 28.0473, "ZA"},
	{"Cape Town", 33.9249, 18.4241, "ZA"},
	{"Nairobi", 1.2921, 36.8219, "KE"},
	{"Casablanca", 33.5731, -7.5898, "MA"},
	{"Algiers", 36.7538, 3.0588, "DZ"},
	{"Addis Ababa", 9.0320, 38.7469, "ET"},
	{"Accra", 5.6037, -0.1870, "GH"},
	{"Durban", 29.8587, 31.0218, "ZA"},
	{"Tunis", 36.8065, 10.1815, "TN"},

	// South America
	{"Sao Paulo", -23.5505, -46.6333, "BR"},
	{"Buenos Aires", -34.6037, -58.3816, "AR"},
	{"Rio de Janeiro", -22.9068, -43.1729, "BR"},
	{"Lima", -12.0464, -77.0428, "PE"},
	{"Bogota", 4.7110, -74.0721, "CO"},
	{"Santiago", -33.4489, -70.6693, "CL"},
	{"Caracas", 10.4806, -66.9036, "VE"},
	{"Medellin", 6.2442, -75.5812, "CO"},
	{"Brasilia", -15.8267, -47.9218, "BR"},
	{"Montevideo", -34.9011, -56.1645, "UY"},
	{"Quito", -0.1807, -78.4678, "EC"},
	{"La Paz", -16.4897, -68.1193, "BO"},

	// Oceania
	{"Sydney", -33.8688, 151.2093, "AU"},
	{"Melbourne", -37.8136, 144.9631, "AU"},
	{"Brisbane", -27.4698, 153.0251, "AU"},
	{"Perth", -31.9505, 115.8605, "AU"},
	{"Auckland", -36.8509, 174.7645, "NZ"},
	{"Wellington", -41.2866, 174.7756, "NZ"},
	{"Adelaide", -34.9285, 138.6007, "AU"},

	// Russia
	{"Moscow", 55.7558, 37.6173, "RU"},
	{"St Petersburg", 59.9311, 30.3609, "RU"},
	{"Novosibirsk", 55.0084, 82.9357, "RU"},
	{"Yekaterinburg", 56.8389, 60.6057, "RU"},
	{"Vladivostok", 43.1332, 131.9113, "RU"},
}

// GetCitiesInBounds returns all cities within the given bounds
func GetCitiesInBounds(north, south, east, west float64) []City {
	var cities []City
	for _, city := range MajorCities {
		if city.Lat >= south && city.Lat <= north &&
			city.Lon >= west && city.Lon <= east {
			cities = append(cities, city)
		}
	}
	return cities
}

// GetNearestCities returns the N nearest cities to a given point
func GetNearestCities(lat, lon float64, count int) []City {
	type cityDist struct {
		city City
		dist float64
	}

	var distances []cityDist
	for _, city := range MajorCities {
		d := haversineDistance(lat, lon, city.Lat, city.Lon)
		distances = append(distances, cityDist{city, d})
	}

	// Sort by distance
	for i := 0; i < len(distances)-1; i++ {
		for j := i + 1; j < len(distances); j++ {
			if distances[j].dist < distances[i].dist {
				distances[i], distances[j] = distances[j], distances[i]
			}
		}
	}

	// Return top N
	var result []City
	for i := 0; i < count && i < len(distances); i++ {
		result = append(result, distances[i].city)
	}
	return result
}

// haversineDistance calculates distance between two points in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Simplified distance calculation (good enough for sorting)
	latDiff := lat2 - lat1
	lonDiff := lon2 - lon1
	return latDiff*latDiff + lonDiff*lonDiff
}
