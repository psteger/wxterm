package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"weatherterm/internal/location"
)

const configFileName = "weatherterm.json"

// Config stores user preferences and saved locations
type Config struct {
	DefaultLocation *location.Location  `json:"default_location,omitempty"`
	SavedLocations  []location.Location `json:"saved_locations,omitempty"`
	UseFahrenheit   bool                `json:"use_fahrenheit"`
}

// Load reads the config file from the user's config directory
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
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
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
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

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "weatherterm", configFileName), nil
}
