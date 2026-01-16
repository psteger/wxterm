package location

// Location represents a saved or current location
type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country,omitempty"`
	Admin1    string  `json:"admin1,omitempty"`
}

// IsValid returns true if the location has coordinates
func (l Location) IsValid() bool {
	return l.Latitude != 0 || l.Longitude != 0
}

// DisplayName returns a formatted location name
func (l Location) DisplayName() string {
	if l.Name == "" {
		return "Unknown Location"
	}
	if l.Admin1 != "" && l.Country != "" {
		return l.Name + ", " + l.Admin1 + ", " + l.Country
	}
	if l.Country != "" {
		return l.Name + ", " + l.Country
	}
	return l.Name
}
