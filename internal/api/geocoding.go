package api

import (
	"fmt"
	"net/url"
)

// geocodingResponse represents the raw geocoding API response
type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
		Timezone  string  `json:"timezone"`
	} `json:"results"`
}

// SearchLocation searches for locations by city name
func (c *Client) SearchLocation(query string) ([]GeoLocation, error) {
	params := url.Values{}
	params.Set("name", query)
	params.Set("count", "10")
	params.Set("language", "en")
	params.Set("format", "json")

	reqURL := fmt.Sprintf("%s?%s", geocodingBaseURL, params.Encode())

	var resp geocodingResponse
	if err := c.get(reqURL, &resp); err != nil {
		return nil, err
	}

	locations := make([]GeoLocation, len(resp.Results))
	for i, r := range resp.Results {
		locations[i] = GeoLocation{
			Name:      r.Name,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
			Country:   r.Country,
			Admin1:    r.Admin1,
			Timezone:  r.Timezone,
		}
	}

	return locations, nil
}

// FormatLocation returns a human-readable location string
func (loc GeoLocation) FormatLocation() string {
	if loc.Admin1 != "" {
		return fmt.Sprintf("%s, %s, %s", loc.Name, loc.Admin1, loc.Country)
	}
	return fmt.Sprintf("%s, %s", loc.Name, loc.Country)
}
