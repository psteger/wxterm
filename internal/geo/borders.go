package geo

// Point represents a geographic coordinate
type Point struct {
	Lat float64
	Lon float64
}

// LineSegment represents a line between two points (for borders/coastlines)
type LineSegment struct {
	Start Point
	End   Point
	Type  string // "coast", "border", "lake"
}

// GetSegmentsInBounds returns all line segments that intersect the given bounds
func GetSegmentsInBounds(north, south, east, west float64) []LineSegment {
	var segments []LineSegment
	for _, seg := range CoastlinesAndBorders {
		// Check if segment intersects bounds (simplified check)
		if segmentIntersectsBounds(seg, north, south, east, west) {
			segments = append(segments, seg)
		}
	}
	return segments
}

func segmentIntersectsBounds(seg LineSegment, north, south, east, west float64) bool {
	// Simple bounding box check for the segment
	minLat := min(seg.Start.Lat, seg.End.Lat)
	maxLat := max(seg.Start.Lat, seg.End.Lat)
	minLon := min(seg.Start.Lon, seg.End.Lon)
	maxLon := max(seg.Start.Lon, seg.End.Lon)

	// Check if bounding boxes overlap
	return !(maxLat < south || minLat > north || maxLon < west || minLon > east)
}

// CoastlinesAndBorders contains simplified coastline and border data
// These are major continental outlines and country borders
var CoastlinesAndBorders = []LineSegment{
	// North America - East Coast (simplified)
	{Point{25.0, -80.0}, Point{30.0, -81.0}, "coast"},
	{Point{30.0, -81.0}, Point{32.0, -80.5}, "coast"},
	{Point{32.0, -80.5}, Point{35.0, -75.5}, "coast"},
	{Point{35.0, -75.5}, Point{37.0, -76.0}, "coast"},
	{Point{37.0, -76.0}, Point{39.0, -74.5}, "coast"},
	{Point{39.0, -74.5}, Point{40.5, -74.0}, "coast"},
	{Point{40.5, -74.0}, Point{41.5, -71.0}, "coast"},
	{Point{41.5, -71.0}, Point{42.5, -70.5}, "coast"},
	{Point{42.5, -70.5}, Point{44.0, -69.0}, "coast"},
	{Point{44.0, -69.0}, Point{45.0, -67.0}, "coast"},
	{Point{45.0, -67.0}, Point{47.0, -64.0}, "coast"},

	// North America - Gulf Coast
	{Point{25.0, -80.0}, Point{26.0, -82.0}, "coast"},
	{Point{26.0, -82.0}, Point{28.0, -83.0}, "coast"},
	{Point{28.0, -83.0}, Point{30.0, -84.0}, "coast"},
	{Point{30.0, -84.0}, Point{30.0, -88.0}, "coast"},
	{Point{30.0, -88.0}, Point{29.0, -90.0}, "coast"},
	{Point{29.0, -90.0}, Point{29.5, -94.0}, "coast"},
	{Point{29.5, -94.0}, Point{26.0, -97.0}, "coast"},

	// North America - West Coast
	{Point{32.5, -117.0}, Point{34.0, -118.5}, "coast"},
	{Point{34.0, -118.5}, Point{35.0, -121.0}, "coast"},
	{Point{35.0, -121.0}, Point{37.5, -122.5}, "coast"},
	{Point{37.5, -122.5}, Point{40.0, -124.0}, "coast"},
	{Point{40.0, -124.0}, Point{42.0, -124.5}, "coast"},
	{Point{42.0, -124.5}, Point{46.0, -124.0}, "coast"},
	{Point{46.0, -124.0}, Point{48.5, -123.0}, "coast"},
	{Point{48.5, -123.0}, Point{49.0, -123.5}, "coast"},

	// US-Canada Border (simplified)
	{Point{49.0, -123.0}, Point{49.0, -117.0}, "border"},
	{Point{49.0, -117.0}, Point{49.0, -110.0}, "border"},
	{Point{49.0, -110.0}, Point{49.0, -100.0}, "border"},
	{Point{49.0, -100.0}, Point{49.0, -95.0}, "border"},
	{Point{49.0, -95.0}, Point{48.0, -88.0}, "border"},
	{Point{48.0, -88.0}, Point{46.0, -84.0}, "border"},
	{Point{46.0, -84.0}, Point{43.0, -79.0}, "border"},
	{Point{43.0, -79.0}, Point{45.0, -75.0}, "border"},
	{Point{45.0, -75.0}, Point{45.0, -71.0}, "border"},
	{Point{45.0, -71.0}, Point{47.0, -67.0}, "border"},

	// US-Mexico Border
	{Point{32.5, -117.0}, Point{32.0, -111.0}, "border"},
	{Point{32.0, -111.0}, Point{31.5, -106.5}, "border"},
	{Point{31.5, -106.5}, Point{29.5, -104.0}, "border"},
	{Point{29.5, -104.0}, Point{26.0, -97.0}, "border"},

	// Great Lakes (simplified)
	{Point{46.0, -84.5}, Point{47.5, -85.0}, "lake"},
	{Point{47.5, -85.0}, Point{48.0, -88.0}, "lake"},
	{Point{48.0, -88.0}, Point{46.5, -92.0}, "lake"},
	{Point{46.5, -92.0}, Point{46.0, -84.5}, "lake"},
	{Point{42.0, -87.5}, Point{43.5, -87.0}, "lake"},
	{Point{43.5, -87.0}, Point{46.0, -84.5}, "lake"},
	{Point{42.0, -87.5}, Point{42.0, -83.0}, "lake"},
	{Point{42.0, -83.0}, Point{43.5, -82.5}, "lake"},
	{Point{43.5, -82.5}, Point{46.0, -84.5}, "lake"},

	// Europe - Western Coast
	{Point{36.0, -6.0}, Point{37.0, -9.0}, "coast"},
	{Point{37.0, -9.0}, Point{40.0, -9.0}, "coast"},
	{Point{40.0, -9.0}, Point{43.5, -8.0}, "coast"},
	{Point{43.5, -8.0}, Point{44.0, -1.5}, "coast"},
	{Point{44.0, -1.5}, Point{46.0, -1.5}, "coast"},
	{Point{46.0, -1.5}, Point{48.5, -5.0}, "coast"},
	{Point{48.5, -5.0}, Point{49.5, -1.5}, "coast"},
	{Point{49.5, -1.5}, Point{51.0, 1.5}, "coast"},
	{Point{51.0, 1.5}, Point{53.0, 5.0}, "coast"},
	{Point{53.0, 5.0}, Point{55.0, 8.0}, "coast"},
	{Point{55.0, 8.0}, Point{58.0, 6.0}, "coast"},

	// UK
	{Point{50.0, -5.5}, Point{51.5, -3.0}, "coast"},
	{Point{51.5, -3.0}, Point{53.5, -3.0}, "coast"},
	{Point{53.5, -3.0}, Point{55.0, -5.0}, "coast"},
	{Point{55.0, -5.0}, Point{58.5, -5.0}, "coast"},
	{Point{58.5, -5.0}, Point{58.5, -3.0}, "coast"},
	{Point{50.0, -5.5}, Point{50.5, 0.0}, "coast"},
	{Point{50.5, 0.0}, Point{53.0, 0.5}, "coast"},
	{Point{53.0, 0.5}, Point{55.5, -1.5}, "coast"},

	// Mediterranean
	{Point{36.0, -6.0}, Point{36.0, 0.0}, "coast"},
	{Point{36.0, 0.0}, Point{37.0, 5.0}, "coast"},
	{Point{37.0, 5.0}, Point{43.5, 7.5}, "coast"},
	{Point{43.5, 7.5}, Point{44.0, 9.5}, "coast"},
	{Point{44.0, 9.5}, Point{40.5, 14.0}, "coast"},
	{Point{40.5, 14.0}, Point{38.0, 16.0}, "coast"},
	{Point{38.0, 16.0}, Point{40.0, 20.0}, "coast"},
	{Point{40.0, 20.0}, Point{39.5, 26.0}, "coast"},
	{Point{39.5, 26.0}, Point{41.0, 29.0}, "coast"},

	// Scandinavia
	{Point{58.0, 6.0}, Point{60.0, 5.0}, "coast"},
	{Point{60.0, 5.0}, Point{62.0, 6.0}, "coast"},
	{Point{62.0, 6.0}, Point{65.0, 12.0}, "coast"},
	{Point{65.0, 12.0}, Point{70.0, 20.0}, "coast"},
	{Point{70.0, 20.0}, Point{70.0, 28.0}, "coast"},
	{Point{58.0, 10.0}, Point{56.0, 12.0}, "coast"},
	{Point{56.0, 12.0}, Point{56.0, 16.0}, "coast"},
	{Point{56.0, 16.0}, Point{60.0, 18.0}, "coast"},
	{Point{60.0, 18.0}, Point{66.0, 24.0}, "coast"},

	// Australia (simplified outline)
	{Point{-12.0, 130.0}, Point{-14.0, 127.0}, "coast"},
	{Point{-14.0, 127.0}, Point{-18.0, 122.0}, "coast"},
	{Point{-18.0, 122.0}, Point{-22.0, 114.0}, "coast"},
	{Point{-22.0, 114.0}, Point{-28.0, 114.0}, "coast"},
	{Point{-28.0, 114.0}, Point{-34.0, 116.0}, "coast"},
	{Point{-34.0, 116.0}, Point{-35.0, 137.0}, "coast"},
	{Point{-35.0, 137.0}, Point{-38.5, 145.0}, "coast"},
	{Point{-38.5, 145.0}, Point{-37.5, 150.0}, "coast"},
	{Point{-37.5, 150.0}, Point{-33.5, 151.5}, "coast"},
	{Point{-33.5, 151.5}, Point{-28.0, 153.5}, "coast"},
	{Point{-28.0, 153.5}, Point{-23.5, 151.0}, "coast"},
	{Point{-23.5, 151.0}, Point{-19.0, 146.5}, "coast"},
	{Point{-19.0, 146.5}, Point{-16.0, 145.5}, "coast"},
	{Point{-16.0, 145.5}, Point{-12.0, 142.0}, "coast"},
	{Point{-12.0, 142.0}, Point{-12.0, 130.0}, "coast"},

	// Japan (simplified)
	{Point{31.0, 130.0}, Point{33.0, 130.0}, "coast"},
	{Point{33.0, 130.0}, Point{35.5, 134.0}, "coast"},
	{Point{35.5, 134.0}, Point{36.0, 136.0}, "coast"},
	{Point{36.0, 136.0}, Point{37.5, 138.5}, "coast"},
	{Point{37.5, 138.5}, Point{40.0, 140.0}, "coast"},
	{Point{40.0, 140.0}, Point{41.5, 141.0}, "coast"},
	{Point{41.5, 141.0}, Point{43.0, 145.5}, "coast"},
	{Point{35.5, 140.0}, Point{34.5, 139.0}, "coast"},
	{Point{34.5, 139.0}, Point{33.0, 136.0}, "coast"},
	{Point{33.0, 136.0}, Point{31.0, 131.5}, "coast"},

	// South America - East Coast
	{Point{5.0, -52.0}, Point{0.0, -50.0}, "coast"},
	{Point{0.0, -50.0}, Point{-5.0, -35.0}, "coast"},
	{Point{-5.0, -35.0}, Point{-13.0, -39.0}, "coast"},
	{Point{-13.0, -39.0}, Point{-23.0, -43.0}, "coast"},
	{Point{-23.0, -43.0}, Point{-30.0, -51.0}, "coast"},
	{Point{-30.0, -51.0}, Point{-35.0, -57.0}, "coast"},

	// South America - West Coast
	{Point{-5.0, -81.0}, Point{-12.0, -77.0}, "coast"},
	{Point{-12.0, -77.0}, Point{-18.0, -70.5}, "coast"},
	{Point{-18.0, -70.5}, Point{-24.0, -70.5}, "coast"},
	{Point{-24.0, -70.5}, Point{-33.0, -72.0}, "coast"},
	{Point{-33.0, -72.0}, Point{-42.0, -73.0}, "coast"},
	{Point{-42.0, -73.0}, Point{-55.0, -68.0}, "coast"},

	// Africa - North Coast
	{Point{35.5, -6.0}, Point{37.0, 10.0}, "coast"},
	{Point{37.0, 10.0}, Point{33.0, 12.0}, "coast"},
	{Point{33.0, 12.0}, Point{31.5, 25.0}, "coast"},
	{Point{31.5, 25.0}, Point{31.5, 32.0}, "coast"},

	// Africa - West Coast
	{Point{35.5, -6.0}, Point{28.0, -13.0}, "coast"},
	{Point{28.0, -13.0}, Point{21.0, -17.0}, "coast"},
	{Point{21.0, -17.0}, Point{15.0, -17.0}, "coast"},
	{Point{15.0, -17.0}, Point{5.0, -5.0}, "coast"},
	{Point{5.0, -5.0}, Point{5.0, 10.0}, "coast"},
	{Point{5.0, 10.0}, Point{-5.0, 12.0}, "coast"},

	// Africa - East Coast
	{Point{31.5, 32.0}, Point{22.0, 37.0}, "coast"},
	{Point{22.0, 37.0}, Point{12.0, 44.0}, "coast"},
	{Point{12.0, 44.0}, Point{-5.0, 40.0}, "coast"},
	{Point{-5.0, 40.0}, Point{-12.0, 40.0}, "coast"},
	{Point{-12.0, 40.0}, Point{-26.0, 33.0}, "coast"},
	{Point{-26.0, 33.0}, Point{-35.0, 20.0}, "coast"},
	{Point{-35.0, 20.0}, Point{-34.5, 18.5}, "coast"},

	// China coast (simplified)
	{Point{40.0, 120.0}, Point{37.0, 122.5}, "coast"},
	{Point{37.0, 122.5}, Point{31.0, 122.0}, "coast"},
	{Point{31.0, 122.0}, Point{28.0, 121.5}, "coast"},
	{Point{28.0, 121.5}, Point{24.5, 118.5}, "coast"},
	{Point{24.5, 118.5}, Point{22.0, 114.5}, "coast"},
	{Point{22.0, 114.5}, Point{21.5, 110.0}, "coast"},
	{Point{21.5, 110.0}, Point{20.0, 110.0}, "coast"},

	// India (simplified)
	{Point{23.0, 68.5}, Point{20.0, 73.0}, "coast"},
	{Point{20.0, 73.0}, Point{15.0, 74.0}, "coast"},
	{Point{15.0, 74.0}, Point{8.0, 77.0}, "coast"},
	{Point{8.0, 77.0}, Point{10.0, 80.0}, "coast"},
	{Point{10.0, 80.0}, Point{16.0, 82.5}, "coast"},
	{Point{16.0, 82.5}, Point{21.0, 87.0}, "coast"},
	{Point{21.0, 87.0}, Point{22.0, 88.5}, "coast"},
}
