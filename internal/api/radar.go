package api

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"net/http"
	"time"
)

// ErrOutsideCONUS is returned when radar is requested for a location outside the continental US
var ErrOutsideCONUS = errors.New("radar not available outside continental US")

// Continental US bounding box (approximate)
const (
	conusMinLat = 24.0  // Southern tip of Florida/Texas
	conusMaxLat = 50.0  // Northern border with Canada
	conusMinLon = -125.0 // Pacific coast
	conusMaxLon = -66.0  // Atlantic coast (Maine)
)

// IsInContinentalUS returns true if the coordinates are within the continental US
func IsInContinentalUS(lat, lon float64) bool {
	return lat >= conusMinLat && lat <= conusMaxLat &&
		lon >= conusMinLon && lon <= conusMaxLon
}

// Base URL for RIDGE II standard radar GIFs
const ridgeStandardBaseURL = "https://radar.weather.gov/ridge/standard"

// Regional sector definitions for RIDGE radar mosaics
type RadarSector struct {
	ID      string  // Sector ID used in URL (e.g., "CENTGRLAKES")
	Name    string  // Display name (e.g., "Central Great Lakes")
	MinLat  float64 // Bounding box
	MaxLat  float64
	MinLon  float64
	MaxLon  float64
}

// Regional sectors available in RIDGE radar
var radarSectors = []RadarSector{
	{"NORTHEAST", "Northeast", 38.0, 48.0, -80.0, -66.0},
	{"SOUTHEAST", "Southeast", 24.0, 38.0, -92.0, -75.0},
	{"CENTGRLAKES", "Central Great Lakes", 38.0, 48.0, -92.0, -80.0},
	{"UPPERMISSVLY", "Upper Mississippi Valley", 40.0, 50.0, -100.0, -88.0},
	{"SOUTHMISSVLY", "South Mississippi Valley", 28.0, 40.0, -100.0, -88.0},
	{"SOUTHPLAINS", "Southern Plains", 26.0, 40.0, -108.0, -95.0},
	{"NORTHROCKIES", "Northern Rockies", 40.0, 50.0, -118.0, -100.0},
	{"SOUTHROCKIES", "Southern Rockies", 30.0, 42.0, -118.0, -102.0},
	{"PACNORTHWEST", "Pacific Northwest", 40.0, 50.0, -128.0, -115.0},
	{"PACSOUTHWEST", "Pacific Southwest", 32.0, 42.0, -125.0, -114.0},
}

// RadarMode represents the radar view mode
type RadarMode int

const (
	RadarModeLocal    RadarMode = iota // Local MRMS view
	RadarModeRegional                  // Regional/CONUS view
)

// String returns the display name for the radar mode
func (m RadarMode) String() string {
	switch m {
	case RadarModeLocal:
		return "Local"
	case RadarModeRegional:
		return "Regional"
	default:
		return "Unknown"
	}
}

// RadarFrame represents a single frame of radar data
type RadarFrame struct {
	Image     image.Image
	Timestamp time.Time
}

// RadarData contains radar frames and metadata
type RadarData struct {
	Frames      []RadarFrame
	StationID   string
	StationName string
	CenterLat   float64
	CenterLon   float64
	Mode        RadarMode
}

// Image returns the first frame's image (for backward compatibility)
func (r *RadarData) Image() image.Image {
	if len(r.Frames) > 0 {
		return r.Frames[0].Image
	}
	return nil
}

// Timestamp returns the current time
func (r *RadarData) Timestamp() time.Time {
	return time.Now()
}

// NWSOffice represents a National Weather Service office
type NWSOffice struct {
	ID      string  // e.g., "iln" (lowercase for URL)
	RadarID string  // NEXRAD radar ID e.g., "ILN" (for ridge radar fallback)
	Name    string  // e.g., "Cincinnati, OH"
	Lat     float64
	Lon     float64
}

// NWS Weather Forecast Offices with corresponding NEXRAD radar IDs
var nwsOffices = []NWSOffice{
	// Northeast
	{"okx", "OKX", "New York, NY", 40.8656, -72.8639},
	{"box", "BOX", "Boston, MA", 41.9558, -71.1369},
	{"aly", "ENX", "Albany, NY", 42.5864, -74.0639},
	{"bgm", "BGM", "Binghamton, NY", 42.1997, -75.985},
	{"buf", "BUF", "Buffalo, NY", 42.9489, -78.7369},
	{"btv", "CXX", "Burlington, VT", 44.5111, -73.1667},
	{"phi", "DIX", "Philadelphia, PA", 39.9469, -74.4108},
	{"pbz", "PBZ", "Pittsburgh, PA", 40.5317, -80.0183},
	{"ctp", "CCX", "State College, PA", 40.9231, -78.0039},
	{"lwx", "LWX", "Washington DC", 38.9753, -77.4778},

	// Southeast
	{"mhx", "MHX", "Morehead City, NC", 34.7761, -76.8764},
	{"rah", "RAX", "Raleigh, NC", 35.6656, -78.4897},
	{"ilm", "LTX", "Wilmington, NC", 33.9892, -78.4292},
	{"gsp", "GSP", "Greenville, SC", 34.8833, -82.2203},
	{"cae", "CAE", "Columbia, SC", 33.9486, -81.1183},
	{"chs", "CLX", "Charleston, SC", 32.6556, -81.0422},
	{"ffc", "FFC", "Atlanta, GA", 33.3636, -84.5658},
	{"jax", "JAX", "Jacksonville, FL", 30.4847, -81.7019},
	{"mlb", "MLB", "Melbourne, FL", 28.1131, -80.6542},
	{"mfl", "AMX", "Miami, FL", 25.6111, -80.4128},
	{"tbw", "TBW", "Tampa, FL", 27.7056, -82.4017},
	{"key", "BYX", "Key West, FL", 24.5975, -81.7033},
	{"tae", "TLH", "Tallahassee, FL", 30.3975, -84.3289},
	{"mob", "MOB", "Mobile, AL", 30.6794, -88.2397},
	{"bmx", "BMX", "Birmingham, AL", 33.1722, -86.7697},
	{"hun", "HTX", "Huntsville, AL", 34.9306, -86.0836},
	{"mrx", "MRX", "Knoxville, TN", 36.1686, -83.4017},
	{"meg", "NQA", "Memphis, TN", 35.3447, -89.8733},
	{"ohx", "OHX", "Nashville, TN", 36.2472, -86.5625},
	{"jkl", "JKL", "Jackson, KY", 37.5908, -83.3131},
	{"lmk", "LVX", "Louisville, KY", 37.9753, -85.9439},
	{"pah", "PAH", "Paducah, KY", 37.0683, -88.7719},

	// Midwest
	{"iln", "ILN", "Cincinnati, OH", 39.4203, -83.8217},
	{"cle", "CLE", "Cleveland, OH", 41.4131, -81.8597},
	{"dtx", "DTX", "Detroit, MI", 42.6997, -83.4719},
	{"grr", "GRR", "Grand Rapids, MI", 42.8939, -85.5447},
	{"apx", "APX", "Gaylord, MI", 44.9072, -84.7197},
	{"mqt", "MQT", "Marquette, MI", 46.5311, -87.5486},
	{"grb", "GRB", "Green Bay, WI", 44.4986, -88.1111},
	{"mkx", "MKX", "Milwaukee, WI", 42.9678, -88.5506},
	{"arx", "ARX", "La Crosse, WI", 43.8228, -91.1911},
	{"dlh", "DLH", "Duluth, MN", 46.8369, -92.21},
	{"mpx", "MPX", "Minneapolis, MN", 44.8489, -93.5653},
	{"fsd", "FSD", "Sioux Falls, SD", 43.5878, -96.7294},
	{"abr", "ABR", "Aberdeen, SD", 45.4558, -98.4131},
	{"unr", "UDX", "Rapid City, SD", 44.125, -102.83},
	{"bis", "BIS", "Bismarck, ND", 46.7711, -100.7606},
	{"fgf", "MVX", "Grand Forks, ND", 47.5278, -97.325},
	{"lot", "LOT", "Chicago, IL", 41.6044, -88.0847},
	{"ilx", "ILX", "Lincoln, IL", 40.1506, -89.3367},
	{"ind", "IND", "Indianapolis, IN", 39.7075, -86.2803},
	{"iwx", "IWX", "Fort Wayne, IN", 41.3586, -85.7},

	// Great Plains
	{"oax", "OAX", "Omaha, NE", 41.3203, -96.3667},
	{"lbf", "LNX", "North Platte, NE", 41.9578, -100.5761},
	{"gid", "UEX", "Hastings, NE", 40.3208, -98.4419},
	{"gld", "GLD", "Goodland, KS", 39.3669, -101.7003},
	{"ddc", "DDC", "Dodge City, KS", 37.7608, -99.9689},
	{"ict", "ICT", "Wichita, KS", 37.6544, -97.4431},
	{"top", "TWX", "Topeka, KS", 38.9969, -96.2325},
	{"eax", "EAX", "Kansas City, MO", 38.8103, -94.2644},
	{"sgf", "SGF", "Springfield, MO", 37.2353, -93.4006},
	{"lsx", "LSX", "St. Louis, MO", 38.6989, -90.6828},
	{"tsa", "INX", "Tulsa, OK", 36.175, -95.5647},
	{"oun", "TLX", "Oklahoma City, OK", 35.3331, -97.2778},
	{"ama", "AMA", "Amarillo, TX", 35.2333, -101.7092},
	{"lub", "LBB", "Lubbock, TX", 33.6536, -101.8142},
	{"maf", "MAF", "Midland, TX", 31.9433, -102.1894},
	{"sjt", "SJT", "San Angelo, TX", 31.3711, -100.4925},
	{"fwd", "FWS", "Dallas/Fort Worth, TX", 32.5731, -97.3031},
	{"ewx", "EWX", "Austin/San Antonio, TX", 29.7039, -98.0286},
	{"hgx", "HGX", "Houston, TX", 29.4719, -95.0789},
	{"crp", "CRP", "Corpus Christi, TX", 27.7842, -97.5111},
	{"bro", "BRO", "Brownsville, TX", 25.9161, -97.4189},
	{"epz", "EPZ", "El Paso, TX", 31.8731, -106.6981},

	// Mountain West
	{"bou", "FTG", "Denver, CO", 39.7867, -104.5458},
	{"pub", "PUX", "Pueblo, CO", 38.4594, -104.1814},
	{"gjt", "GJX", "Grand Junction, CO", 39.0622, -108.2139},
	{"cys", "CYS", "Cheyenne, WY", 41.1519, -104.8061},
	{"riw", "RIW", "Riverton, WY", 43.0661, -108.4772},
	{"slc", "MTX", "Salt Lake City, UT", 40.9694, -111.9303},
	{"fgz", "FSX", "Flagstaff, AZ", 34.5744, -111.1983},
	{"psr", "IWA", "Phoenix, AZ", 33.2892, -111.6697},
	{"twc", "EMX", "Tucson, AZ", 31.8936, -110.6303},
	{"abq", "ABX", "Albuquerque, NM", 35.1497, -106.8239},
	{"lkn", "LRX", "Elko, NV", 40.7397, -116.8025},
	{"vef", "ESX", "Las Vegas, NV", 35.7011, -114.8919},
	{"rev", "RGX", "Reno, NV", 39.7542, -119.4622},

	// Pacific
	{"lox", "VTX", "Los Angeles, CA", 34.4117, -119.1792},
	{"sgx", "NKX", "San Diego, CA", 32.9189, -117.0419},
	{"mtr", "MUX", "San Francisco, CA", 37.155, -121.8983},
	{"sto", "DAX", "Sacramento, CA", 38.5011, -121.6778},
	{"eka", "BHX", "Eureka, CA", 40.4986, -124.2919},
	{"mfr", "MAX", "Medford, OR", 42.0811, -122.7172},
	{"pdt", "PDT", "Pendleton, OR", 45.6906, -118.8528},
	{"pqr", "RTX", "Portland, OR", 45.715, -122.9656},
	{"sew", "ATX", "Seattle, WA", 48.1947, -122.4958},
	{"otx", "OTX", "Spokane, WA", 47.6803, -117.6267},

	// Note: Alaska and Hawaii removed - no RIDGE radar coverage
}

// findRegionalSector finds the best matching regional sector for coordinates
func findRegionalSector(lat, lon float64) RadarSector {
	// First, try to find a sector that contains this point
	for _, sector := range radarSectors {
		if lat >= sector.MinLat && lat <= sector.MaxLat &&
			lon >= sector.MinLon && lon <= sector.MaxLon {
			return sector
		}
	}

	// If no exact match, find the nearest sector center
	var nearest RadarSector
	minDist := math.MaxFloat64

	for _, sector := range radarSectors {
		centerLat := (sector.MinLat + sector.MaxLat) / 2
		centerLon := (sector.MinLon + sector.MaxLon) / 2
		dist := haversineDistance(lat, lon, centerLat, centerLon)
		if dist < minDist {
			minDist = dist
			nearest = sector
		}
	}

	// Default to CONUS if nothing found (shouldn't happen for US locations)
	if nearest.ID == "" {
		return RadarSector{ID: "CONUS", Name: "Continental US"}
	}

	return nearest
}

// FetchRadar fetches radar imagery from weather.gov for a location
func (c *Client) FetchRadar(lat, lon float64, mode RadarMode) (*RadarData, error) {
	// Check if location is within continental US
	if !IsInContinentalUS(lat, lon) {
		return nil, ErrOutsideCONUS
	}

	var url string
	var stationID string
	var stationName string

	if mode == RadarModeRegional {
		// Regional mode: fetch sector mosaic
		sector := findRegionalSector(lat, lon)
		url = fmt.Sprintf("%s/%s_loop.gif", ridgeStandardBaseURL, sector.ID)
		stationID = sector.ID
		stationName = sector.Name
	} else {
		// Local mode: fetch individual radar station
		office := findNearestOffice(lat, lon)
		url = fmt.Sprintf("%s/K%s_loop.gif", ridgeStandardBaseURL, office.RadarID)
		stationID = office.RadarID
		stationName = office.Name
	}

	frames, err := c.fetchAndCompositeGIF(url)
	if err != nil || len(frames) == 0 {
		return nil, fmt.Errorf("failed to fetch radar from %s: %w", stationName, err)
	}

	return &RadarData{
		Frames:      frames,
		StationID:   stationID,
		StationName: stationName,
		CenterLat:   lat,
		CenterLon:   lon,
		Mode:        mode,
	}, nil
}

// findNearestOffice finds the nearest NWS office to the given coordinates
func findNearestOffice(lat, lon float64) NWSOffice {
	var nearest NWSOffice
	minDist := math.MaxFloat64

	for _, office := range nwsOffices {
		dist := haversineDistance(lat, lon, office.Lat, office.Lon)
		if dist < minDist {
			minDist = dist
			nearest = office
		}
	}

	return nearest
}

// haversineDistance calculates the distance between two coordinates in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in km

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// fetchAndCompositeGIF fetches a GIF and properly composites all frames
// This handles GIF disposal methods correctly so frames render properly
func (c *Client) fetchAndCompositeGIF(url string) ([]RadarFrame, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	g, err := gif.DecodeAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode GIF: %w", err)
	}

	if len(g.Image) == 0 {
		return nil, fmt.Errorf("GIF has no frames")
	}

	// Create a canvas to composite frames onto
	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)

	// Fill with background color if specified
	if int(g.BackgroundIndex) < len(g.Image[0].Palette) {
		bgColor := g.Image[0].Palette[g.BackgroundIndex]
		draw.Draw(canvas, bounds, &image.Uniform{bgColor}, image.Point{}, draw.Src)
	}

	frames := make([]RadarFrame, len(g.Image))
	now := time.Now()

	for i, srcFrame := range g.Image {
		// Get the frame bounds (where this frame should be drawn)
		frameBounds := srcFrame.Bounds()

		// Draw this frame onto the canvas
		draw.Draw(canvas, frameBounds, srcFrame, frameBounds.Min, draw.Over)

		// Create a copy of the current canvas state for this frame
		frameCopy := image.NewRGBA(bounds)
		draw.Draw(frameCopy, bounds, canvas, image.Point{}, draw.Src)

		frames[i] = RadarFrame{
			Image:     frameCopy,
			Timestamp: now.Add(time.Duration(-len(g.Image)+i+1) * 5 * time.Minute),
		}

		// Handle disposal method for next frame
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				// Clear the frame area to background
				if int(g.BackgroundIndex) < len(srcFrame.Palette) {
					bgColor := srcFrame.Palette[g.BackgroundIndex]
					draw.Draw(canvas, frameBounds, &image.Uniform{bgColor}, image.Point{}, draw.Src)
				} else {
					draw.Draw(canvas, frameBounds, &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
				}
			case gif.DisposalPrevious:
				// Restore to previous state - for simplicity, we don't track this
				// Most weather radar GIFs don't use this disposal method
			case gif.DisposalNone:
				// Leave canvas as-is (frame persists)
			}
		}
	}

	return frames, nil
}
