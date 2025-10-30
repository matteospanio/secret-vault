package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSyncConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sync-config.json")

	// Create test config
	config := &SyncConfig{
		Provider: "nextcloud",
		Settings: map[string]interface{}{
			"url":      "https://cloud.example.com",
			"username": "testuser",
			"password": "testpass",
			"path":     "/Vaults/test.vault",
		},
	}

	// Save config
	if err := SaveSyncConfig(configPath, config); err != nil {
		t.Fatalf("SaveSyncConfig failed: %v", err)
	}

	// Verify file exists
	if !ConfigExists(configPath) {
		t.Error("Config file should exist after saving")
	}

	// Load config
	loadedConfig, err := LoadSyncConfig(configPath)
	if err != nil {
		t.Fatalf("LoadSyncConfig failed: %v", err)
	}

	if loadedConfig.Provider != config.Provider {
		t.Errorf("Expected provider %s, got %s", config.Provider, loadedConfig.Provider)
	}

	if loadedConfig.Settings["url"] != config.Settings["url"] {
		t.Errorf("Expected url %s, got %s", config.Settings["url"], loadedConfig.Settings["url"])
	}
}

func TestLoadSyncConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")

	_, err := LoadSyncConfig(configPath)
	if err == nil {
		t.Error("Expected error for non-existent config")
	}
}

func TestConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.json")

	// Should not exist initially
	if ConfigExists(configPath) {
		t.Error("Config should not exist initially")
	}

	// Create file
	if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should exist now
	if !ConfigExists(configPath) {
		t.Error("Config should exist after creation")
	}
}

func TestParseNextcloudConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *SyncConfig
		expectError bool
	}{
		{
			name: "Valid Nextcloud config",
			config: &SyncConfig{
				Provider: "nextcloud",
				Settings: map[string]interface{}{
					"url":      "https://cloud.example.com",
					"username": "testuser",
					"password": "testpass",
					"path":     "/Vaults/test.vault",
				},
			},
			expectError: false,
		},
		{
			name: "Wrong provider",
			config: &SyncConfig{
				Provider: "dropbox",
				Settings: map[string]interface{}{
					"url":      "https://cloud.example.com",
					"username": "testuser",
					"password": "testpass",
					"path":     "/Vaults/test.vault",
				},
			},
			expectError: true,
		},
		{
			name: "Missing URL",
			config: &SyncConfig{
				Provider: "nextcloud",
				Settings: map[string]interface{}{
					"username": "testuser",
					"password": "testpass",
					"path":     "/Vaults/test.vault",
				},
			},
			expectError: true,
		},
		{
			name: "Missing username",
			config: &SyncConfig{
				Provider: "nextcloud",
				Settings: map[string]interface{}{
					"url":      "https://cloud.example.com",
					"password": "testpass",
					"path":     "/Vaults/test.vault",
				},
			},
			expectError: true,
		},
		{
			name: "Missing path",
			config: &SyncConfig{
				Provider: "nextcloud",
				Settings: map[string]interface{}{
					"url":      "https://cloud.example.com",
					"username": "testuser",
					"password": "testpass",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ncConfig, err := ParseNextcloudConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if ncConfig == nil {
					t.Error("Expected non-nil config")
				}
				if ncConfig != nil {
					if ncConfig.URL != tt.config.Settings["url"] {
						t.Errorf("Expected URL %s, got %s", tt.config.Settings["url"], ncConfig.URL)
					}
					if ncConfig.Username != tt.config.Settings["username"] {
						t.Errorf("Expected username %s, got %s", tt.config.Settings["username"], ncConfig.Username)
					}
				}
			}
		})
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path, err := GetDefaultConfigPath()
	if err != nil {
		t.Fatalf("GetDefaultConfigPath failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty config path")
	}

	// Should contain .secret-vault directory
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}
