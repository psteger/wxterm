package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ipAPIURL = "http://ip-api.com/json/"

// ipAPIResponse represents the response from ip-api.com
type ipAPIResponse struct {
	Status      string  `json:"status"`
	City        string  `json:"city"`
	RegionName  string  `json:"regionName"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	Message     string  `json:"message,omitempty"`
}

// DetectFromIP attempts to detect location from IP address
func DetectFromIP() (Location, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(ipAPIURL)
	if err != nil {
		return Location{}, fmt.Errorf("failed to detect location: %w", err)
	}
	defer resp.Body.Close()

	var result ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Location{}, fmt.Errorf("failed to parse location response: %w", err)
	}

	if result.Status != "success" {
		return Location{}, fmt.Errorf("location detection failed: %s", result.Message)
	}

	return Location{
		Name:      result.City,
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		Country:   result.Country,
		Admin1:    result.RegionName,
	}, nil
}
