package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVault(t *testing.T) {
	v := NewVault()
	if v == nil {
		t.Fatal("NewVault returned nil")
	}
	if v.Secrets == nil {
		t.Fatal("Secrets map is nil")
	}
	if v.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", v.Version)
	}
}

func TestAddSecret(t *testing.T) {
	v := NewVault()
	v.AddSecret("test-key", "test-value", "Test description")

	if len(v.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(v.Secrets))
	}

	secret, exists := v.Secrets["test-key"]
	if !exists {
		t.Fatal("Secret not found")
	}

	if secret.Name != "test-key" {
		t.Errorf("Expected name 'test-key', got '%s'", secret.Name)
	}
	if secret.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", secret.Value)
	}
	if secret.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", secret.Description)
	}
}

func TestGetSecret(t *testing.T) {
	v := NewVault()
	v.AddSecret("test-key", "test-value", "")

	secret, err := v.GetSecret("test-key")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	if secret.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", secret.Value)
	}

	_, err = v.GetSecret("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent secret")
	}
}

func TestRemoveSecret(t *testing.T) {
	v := NewVault()
	v.AddSecret("test-key", "test-value", "")

	err := v.RemoveSecret("test-key")
	if err != nil {
		t.Fatalf("RemoveSecret failed: %v", err)
	}

	if len(v.Secrets) != 0 {
		t.Errorf("Expected 0 secrets, got %d", len(v.Secrets))
	}

	err = v.RemoveSecret("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent secret")
	}
}

func TestListSecrets(t *testing.T) {
	v := NewVault()
	v.AddSecret("key1", "value1", "")
	v.AddSecret("key2", "value2", "")
	v.AddSecret("key3", "value3", "")

	names := v.ListSecrets()
	if len(names) != 3 {
		t.Fatalf("Expected 3 secret names, got %d", len(names))
	}

	// Check all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["key1"] || !nameMap["key2"] || !nameMap["key3"] {
		t.Error("Not all secret names found in list")
	}
}

func TestJSONSerialization(t *testing.T) {
	v := NewVault()
	v.AddSecret("test-key", "test-value", "Test description")

	jsonData, err := v.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	v2, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if len(v2.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(v2.Secrets))
	}

	secret, exists := v2.Secrets["test-key"]
	if !exists {
		t.Fatal("Secret not found after deserialization")
	}

	if secret.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", secret.Value)
	}
}

func TestGetDefaultVaultPath(t *testing.T) {
	path, err := GetDefaultVaultPath()
	if err != nil {
		t.Fatalf("GetDefaultVaultPath failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty path")
	}

	// Check that the directory exists
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Vault directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Vault path is not a directory")
	}
}
