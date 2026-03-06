package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	weatherBaseURL    = "https://api.open-meteo.com/v1/forecast"
	geocodingBaseURL  = "https://geocoding-api.open-meteo.com/v1/search"
	defaultTimeout    = 10 * time.Second
)

// Client handles API requests to Open-Meteo
type Client struct {
	httpClient *http.Client
	tileCache  *TileCache
}

// NewClient creates a new API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		tileCache: NewTileCache(),
	}
}

// get performs a GET request and decodes the JSON response
func (c *Client) get(url string, result interface{}) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
