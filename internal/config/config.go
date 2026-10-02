package config

import (
	"encoding/json"
	"os"

	"wxterm/internal/location"
)

const configFileName = "wxterm.json"

// Config stores user preferences and saved locations
type Config struct {
	DefaultLocation *location.Location  `json:"default_location,omitempty"`
	SavedLocations  []location.Location `json:"saved_locations,omitempty"`
	UseFahrenheit   bool                `json:"use_fahrenheit"`
}

// Load reads the config file from the user's config directory
func Load() (*Config, error) {
	data, err := readConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, nil
	}

	return &cfg, nil
}

// Save writes the config to the user's config directory
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return writeConfig(data)
}

// AddSavedLocation adds a location to saved locations if not already present
func (c *Config) AddSavedLocation(loc location.Location) {
	for _, saved := range c.SavedLocations {
		if saved.Latitude == loc.Latitude && saved.Longitude == loc.Longitude {
			return
		}
	}
	c.SavedLocations = append(c.SavedLocations, loc)
}

// RemoveSavedLocation removes a location from saved locations
func (c *Config) RemoveSavedLocation(idx int) {
	if idx >= 0 && idx < len(c.SavedLocations) {
		c.SavedLocations = append(c.SavedLocations[:idx], c.SavedLocations[idx+1:]...)
	}
}

// SetDefaultLocation sets the default location
func (c *Config) SetDefaultLocation(loc location.Location) {
	c.DefaultLocation = &loc
}
