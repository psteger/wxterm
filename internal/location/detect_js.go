//go:build js

package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ip-api.com is plain HTTP only, which browsers block from an HTTPS page, so
// the browser build uses ipwho.is (HTTPS, CORS-enabled).
const ipAPIURL = "https://ipwho.is/"

type ipWhoResponse struct {
	Success   bool    `json:"success"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Message   string  `json:"message,omitempty"`
}

// DetectFromIP attempts to detect location from IP address
func DetectFromIP() (Location, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(ipAPIURL)
	if err != nil {
		return Location{}, fmt.Errorf("failed to detect location: %w", err)
	}
	defer resp.Body.Close()

	var result ipWhoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Location{}, fmt.Errorf("failed to parse location response: %w", err)
	}

	if !result.Success {
		return Location{}, fmt.Errorf("location detection failed: %s", result.Message)
	}

	return Location{
		Name:      result.City,
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		Country:   result.Country,
		Admin1:    result.Region,
	}, nil
}
