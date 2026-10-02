//go:build !js

package config

import (
	"os"
	"path/filepath"
)

// readConfig returns the raw config file contents. A missing file or an
// unavailable config directory is reported as os.ErrNotExist.
func readConfig() ([]byte, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, os.ErrNotExist
	}

	// #nosec G304 -- path is os.UserConfigDir() plus fixed names; no user input
	return os.ReadFile(configPath)
}

func writeConfig(data []byte) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "wxterm", configFileName), nil
}
