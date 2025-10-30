package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SyncConfig represents the sync configuration
type SyncConfig struct {
	Provider string                 `json:"provider"`
	Settings map[string]interface{} `json:"settings"`
}

// NextcloudConfig represents Nextcloud-specific configuration
type NextcloudConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // Optional: can use keychain
	Path     string `json:"path"`
}

// GetDefaultConfigPath returns the default sync config path
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".secret-vault")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "sync-config.json"), nil
}

// LoadSyncConfig loads sync configuration from file
func LoadSyncConfig(configPath string) (*SyncConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("sync not configured: run 'vault sync configure' first")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config SyncConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveSyncConfig saves sync configuration to file
func SaveSyncConfig(configPath string, config *SyncConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ConfigExists checks if a sync config file exists
func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}

// ParseNextcloudConfig parses Nextcloud settings from config
func ParseNextcloudConfig(config *SyncConfig) (*NextcloudConfig, error) {
	if config.Provider != "nextcloud" {
		return nil, fmt.Errorf("config is not for Nextcloud provider")
	}

	// Convert settings map to struct
	settingsJSON, err := json.Marshal(config.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	var ncConfig NextcloudConfig
	if err := json.Unmarshal(settingsJSON, &ncConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Nextcloud config: %w", err)
	}

	// Validate required fields
	if ncConfig.URL == "" || ncConfig.Username == "" || ncConfig.Path == "" {
		return nil, fmt.Errorf("incomplete Nextcloud configuration: url, username, and path are required")
	}

	return &ncConfig, nil
}
