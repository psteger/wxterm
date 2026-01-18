package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

const (
	// Maximum distance in km for local radar coverage (500 miles)
	maxLocalRadarDistanceKm = 804.672
	// Maximum distance in km for regional radar coverage (1000 miles)
	maxRegionalRadarDistanceKm = 1609.344
)

// TestAllRadarStationsReachable tests that all NWS office radar URLs return valid responses
func TestAllRadarStationsReachable(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var failures []string

	for _, office := range nwsOffices {
		url := fmt.Sprintf("%s/K%s_loop.gif", ridgeStandardBaseURL, office.RadarID)

		t.Run(office.Name, func(t *testing.T) {
			resp, err := client.Head(url)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s (%s): connection error - %v", office.Name, office.RadarID, err))
				t.Errorf("Failed to reach %s (%s): %v", office.Name, office.RadarID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				failures = append(failures, fmt.Sprintf("%s (%s): HTTP %d", office.Name, office.RadarID, resp.StatusCode))
				t.Errorf("Radar for %s (%s) returned HTTP %d - URL: %s", office.Name, office.RadarID, resp.StatusCode, url)
			}
		})
	}

	if len(failures) > 0 {
		t.Logf("\n=== SUMMARY: %d radar stations unreachable ===", len(failures))
		for _, f := range failures {
			t.Logf("  - %s", f)
		}
	}
}

// TestAllRegionalSectorsReachable tests that all regional sector radar URLs return valid responses
func TestAllRegionalSectorsReachable(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var failures []string

	for _, sector := range radarSectors {
		url := fmt.Sprintf("%s/%s_loop.gif", ridgeStandardBaseURL, sector.ID)

		t.Run(sector.Name, func(t *testing.T) {
			resp, err := client.Head(url)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s (%s): connection error - %v", sector.Name, sector.ID, err))
				t.Errorf("Failed to reach %s (%s): %v", sector.Name, sector.ID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				failures = append(failures, fmt.Sprintf("%s (%s): HTTP %d", sector.Name, sector.ID, resp.StatusCode))
				t.Errorf("Sector %s (%s) returned HTTP %d - URL: %s", sector.Name, sector.ID, resp.StatusCode, url)
			}
		})
	}

	if len(failures) > 0 {
		t.Logf("\n=== SUMMARY: %d regional sectors unreachable ===", len(failures))
		for _, f := range failures {
			t.Logf("  - %s", f)
		}
	}
}

// TestRadarStationsByRegion groups radar tests by geographic region for easier debugging
func TestRadarStationsByRegion(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	regions := map[string][]NWSOffice{
		"Northeast": {},
		"Southeast": {},
		"Midwest":   {},
		"Plains":    {},
		"Mountain":  {},
		"Pacific":   {},
	}

	// Categorize offices by region based on their coordinates
	for _, office := range nwsOffices {
		switch {
		case office.Lon > -80:
			regions["Northeast"] = append(regions["Northeast"], office)
		case office.Lon > -92 && office.Lat < 38:
			regions["Southeast"] = append(regions["Southeast"], office)
		case office.Lon > -100 && office.Lat >= 38:
			regions["Midwest"] = append(regions["Midwest"], office)
		case office.Lon > -108:
			regions["Plains"] = append(regions["Plains"], office)
		case office.Lon > -118:
			regions["Mountain"] = append(regions["Mountain"], office)
		default:
			regions["Pacific"] = append(regions["Pacific"], office)
		}
	}

	for regionName, offices := range regions {
		t.Run(regionName, func(t *testing.T) {
			for _, office := range offices {
				url := fmt.Sprintf("%s/K%s_loop.gif", ridgeStandardBaseURL, office.RadarID)

				t.Run(office.Name, func(t *testing.T) {
					resp, err := client.Head(url)
					if err != nil {
						t.Errorf("Connection error for %s: %v", office.Name, err)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						t.Errorf("HTTP %d for %s (RadarID: %s) - URL: %s",
							resp.StatusCode, office.Name, office.RadarID, url)
					}
				})
			}
		})
	}
}

// TestLocalRadarWithin500Miles verifies that findNearestOffice returns a station
// within 500 miles for various continental US locations
func TestLocalRadarWithin500Miles(t *testing.T) {
	// Test locations across the continental US
	testLocations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		// Major cities
		{"New York City", 40.7128, -74.0060},
		{"Los Angeles", 34.0522, -118.2437},
		{"Chicago", 41.8781, -87.6298},
		{"Houston", 29.7604, -95.3698},
		{"Phoenix", 33.4484, -112.0740},
		{"Philadelphia", 39.9526, -75.1652},
		{"San Antonio", 29.4241, -98.4936},
		{"San Diego", 32.7157, -117.1611},
		{"Dallas", 32.7767, -96.7970},
		{"San Jose", 37.3382, -121.8863},
		{"Austin", 30.2672, -97.7431},
		{"Jacksonville", 30.3322, -81.6557},
		{"Fort Worth", 32.7555, -97.3308},
		{"Columbus", 39.9612, -82.9988},
		{"Charlotte", 35.2271, -80.8431},
		{"San Francisco", 37.7749, -122.4194},
		{"Indianapolis", 39.7684, -86.1581},
		{"Seattle", 47.6062, -122.3321},
		{"Denver", 39.7392, -104.9903},
		{"Boston", 42.3601, -71.0589},
		{"Miami", 25.7617, -80.1918},
		{"Atlanta", 33.7490, -84.3880},
		{"Minneapolis", 44.9778, -93.2650},
		{"New Orleans", 29.9511, -90.0715},
		{"Detroit", 42.3314, -83.0458},
		{"Portland", 45.5152, -122.6784},
		{"Las Vegas", 36.1699, -115.1398},
		{"Salt Lake City", 40.7608, -111.8910},

		// Edge cases - remote areas
		{"Rural Montana", 47.0, -110.0},
		{"Rural Wyoming", 43.0, -107.5},
		{"Rural Nevada", 40.0, -117.0},
		{"West Texas", 31.0, -104.0},
		{"Northern Maine", 46.5, -68.5},
		{"Florida Keys", 24.7, -81.5},
		{"Upper Peninsula MI", 46.5, -87.5},
	}

	for _, loc := range testLocations {
		t.Run(loc.name, func(t *testing.T) {
			office := findNearestOffice(loc.lat, loc.lon)
			distance := haversineDistance(loc.lat, loc.lon, office.Lat, office.Lon)

			if distance > maxLocalRadarDistanceKm {
				t.Errorf("Location %s (%.4f, %.4f) - nearest radar %s is %.1f km away (%.1f miles), exceeds 500 mile limit",
					loc.name, loc.lat, loc.lon, office.Name, distance, distance/1.60934)
			} else {
				t.Logf("%s -> %s: %.1f km (%.1f miles)", loc.name, office.Name, distance, distance/1.60934)
			}
		})
	}
}

// TestRegionalRadarWithin1000Miles verifies that findRegionalSector returns a sector
// within 1000 miles for various continental US locations
func TestRegionalRadarWithin1000Miles(t *testing.T) {
	// Test locations across the continental US
	testLocations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		// Major cities
		{"New York City", 40.7128, -74.0060},
		{"Los Angeles", 34.0522, -118.2437},
		{"Chicago", 41.8781, -87.6298},
		{"Houston", 29.7604, -95.3698},
		{"Phoenix", 33.4484, -112.0740},
		{"Seattle", 47.6062, -122.3321},
		{"Denver", 39.7392, -104.9903},
		{"Miami", 25.7617, -80.1918},
		{"Minneapolis", 44.9778, -93.2650},
		{"Dallas", 32.7767, -96.7970},

		// Edge locations
		{"San Diego", 32.7157, -117.1611},
		{"Maine", 46.5, -68.5},
		{"Florida Keys", 24.7, -81.5},
		{"El Paso", 31.7619, -106.4850},
	}

	for _, loc := range testLocations {
		t.Run(loc.name, func(t *testing.T) {
			sector := findRegionalSector(loc.lat, loc.lon)

			// Calculate distance to sector center
			centerLat := (sector.MinLat + sector.MaxLat) / 2
			centerLon := (sector.MinLon + sector.MaxLon) / 2
			distance := haversineDistance(loc.lat, loc.lon, centerLat, centerLon)

			if distance > maxRegionalRadarDistanceKm {
				t.Errorf("Location %s (%.4f, %.4f) - sector %s center is %.1f km away (%.1f miles), exceeds 1000 mile limit",
					loc.name, loc.lat, loc.lon, sector.Name, distance, distance/1.60934)
			} else {
				t.Logf("%s -> %s: %.1f km (%.1f miles)", loc.name, sector.Name, distance, distance/1.60934)
			}
		})
	}
}

// TestAllOfficesWithin500MilesOfThemselves ensures every NWS office
// would select itself (or nearby station) as nearest radar
func TestAllOfficesWithin500MilesOfThemselves(t *testing.T) {
	for _, office := range nwsOffices {
		t.Run(office.Name, func(t *testing.T) {
			nearest := findNearestOffice(office.Lat, office.Lon)
			distance := haversineDistance(office.Lat, office.Lon, nearest.Lat, nearest.Lon)

			if distance > maxLocalRadarDistanceKm {
				t.Errorf("Office %s at (%.4f, %.4f) - nearest radar %s is %.1f km away (%.1f miles)",
					office.Name, office.Lat, office.Lon, nearest.Name, distance, distance/1.60934)
			}

			// The nearest office to any office should be itself
			if nearest.ID != office.ID {
				t.Logf("Note: Office %s selected %s as nearest (distance: %.1f km)",
					office.Name, nearest.Name, distance)
			}
		})
	}
}

// TestIsInContinentalUS tests the CONUS boundary check
func TestIsInContinentalUS(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected bool
	}{
		// Should be in CONUS
		{"New York City", 40.7128, -74.0060, true},
		{"Los Angeles", 34.0522, -118.2437, true},
		{"Chicago", 41.8781, -87.6298, true},
		{"Miami", 25.7617, -80.1918, true},
		{"Seattle", 47.6062, -122.3321, true},
		{"Key West", 24.5551, -81.7800, true},
		{"Northern Maine", 47.0, -68.0, true},

		// Should NOT be in CONUS (outside bounding box)
		{"Honolulu, HI", 21.3069, -157.8583, false},
		{"Anchorage, AK", 61.2181, -149.9003, false},
		{"Juneau, AK", 58.3019, -134.4197, false},
		{"San Juan, PR", 18.4655, -66.1057, false},
		{"Edmonton, Canada", 53.5461, -113.4938, false}, // Well north of border
		{"Cancun, Mexico", 21.1619, -86.8515, false},   // South of border
		{"London, UK", 51.5074, -0.1278, false},
		{"Tokyo, Japan", 35.6762, 139.6503, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsInContinentalUS(tc.lat, tc.lon)
			if result != tc.expected {
				t.Errorf("IsInContinentalUS(%f, %f) = %v, expected %v",
					tc.lat, tc.lon, result, tc.expected)
			}
		})
	}
}

// TestFetchRadarRejectsNonCONUS tests that FetchRadar returns ErrOutsideCONUS for non-CONUS locations
func TestFetchRadarRejectsNonCONUS(t *testing.T) {
	client := NewClient()

	nonCONUSLocations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Honolulu, HI", 21.3069, -157.8583},
		{"Anchorage, AK", 61.2181, -149.9003},
		{"San Juan, PR", 18.4655, -66.1057},
	}

	for _, loc := range nonCONUSLocations {
		t.Run(loc.name, func(t *testing.T) {
			_, err := client.FetchRadar(loc.lat, loc.lon, RadarModeLocal)
			if err != ErrOutsideCONUS {
				t.Errorf("Expected ErrOutsideCONUS for %s, got: %v", loc.name, err)
			}
		})
	}
}

// TestFindNearestOffice tests that findNearestOffice returns expected results
func TestFindNearestOffice(t *testing.T) {
	tests := []struct {
		name        string
		lat         float64
		lon         float64
		expectedID  string
		description string
	}{
		{"New York City", 40.7128, -74.0060, "phi", "Should find PHI for NYC"},
		{"Los Angeles", 34.0522, -118.2437, "lox", "Should find LOX for LA"},
		{"Chicago", 41.8781, -87.6298, "lot", "Should find LOT for Chicago"},
		{"Miami", 25.7617, -80.1918, "mfl", "Should find MFL for Miami"},
		{"Seattle", 47.6062, -122.3321, "sew", "Should find SEW for Seattle"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			office := findNearestOffice(tc.lat, tc.lon)
			if office.ID != tc.expectedID {
				t.Errorf("%s: expected office ID %s, got %s (%s)",
					tc.description, tc.expectedID, office.ID, office.Name)
			}
		})
	}
}

// TestFindRegionalSector tests regional sector lookup
func TestFindRegionalSector(t *testing.T) {
	tests := []struct {
		name       string
		lat        float64
		lon        float64
		expectedID string
	}{
		{"New York", 40.7128, -74.0060, "NORTHEAST"},
		{"Miami", 25.7617, -80.1918, "SOUTHEAST"},
		{"Chicago", 41.8781, -87.6298, "CENTGRLAKES"},
		{"Denver", 39.7392, -104.9903, "SOUTHPLAINS"},
		{"Seattle", 47.6062, -122.3321, "PACNORTHWEST"},
		{"Los Angeles", 34.0522, -118.2437, "PACSOUTHWEST"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sector := findRegionalSector(tc.lat, tc.lon)
			if sector.ID != tc.expectedID {
				t.Errorf("Expected sector %s for %s, got %s",
					tc.expectedID, tc.name, sector.ID)
			}
		})
	}
}
