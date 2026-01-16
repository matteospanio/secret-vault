package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadVault(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.enc")
	password := "test-password-123"

	// Create and save a vault
	v1 := NewVault()
	v1.AddSecret("key1", "value1", "Description 1", "", nil)
	v1.AddSecret("key2", "value2", "Description 2", "", nil)

	err = SaveVault(v1, vaultPath, password)
	if err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Verify file was created
	if !VaultExists(vaultPath) {
		t.Fatal("Vault file was not created")
	}

	// Load the vault
	v2, err := LoadVault(vaultPath, password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	// Verify the data
	if len(v2.Secrets) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(v2.Secrets))
	}

	secret1, err := v2.GetSecret("key1")
	if err != nil {
		t.Fatalf("GetSecret key1 failed: %v", err)
	}
	if secret1.Value != "value1" {
		t.Errorf("Expected value1, got %s", secret1.Value)
	}

	secret2, err := v2.GetSecret("key2")
	if err != nil {
		t.Fatalf("GetSecret key2 failed: %v", err)
	}
	if secret2.Value != "value2" {
		t.Errorf("Expected value2, got %s", secret2.Value)
	}
}

func TestLoadVaultWithWrongPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.enc")
	correctPassword := "correct-password"
	wrongPassword := "wrong-password"

	// Create and save a vault
	v1 := NewVault()
	v1.AddSecret("key1", "value1", "", "", nil)

	err = SaveVault(v1, vaultPath, correctPassword)
	if err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Try to load with wrong password
	_, err = LoadVault(vaultPath, wrongPassword)
	if err == nil {
		t.Error("Expected error when loading vault with wrong password")
	}
}

func TestLoadNonexistentVault(t *testing.T) {
	_, err := LoadVault("/nonexistent/path/vault.enc", "password")
	if err == nil {
		t.Error("Expected error when loading nonexistent vault")
	}
}

func TestVaultExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	existingPath := filepath.Join(tmpDir, "existing.enc")
	nonexistentPath := filepath.Join(tmpDir, "nonexistent.enc")

	// Create a file
	err = os.WriteFile(existingPath, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !VaultExists(existingPath) {
		t.Error("VaultExists should return true for existing file")
	}

	if VaultExists(nonexistentPath) {
		t.Error("VaultExists should return false for nonexistent file")
	}
}

func TestSaveVaultPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.enc")
	password := "test-password"

	v := NewVault()
	v.AddSecret("key1", "value1", "", "", nil)

	err = SaveVault(v, vaultPath, password)
	if err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Check file permissions (should be 0600 - read/write for owner only)
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault file: %v", err)
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)
	if mode != expectedMode {
		t.Errorf("Expected file permissions %v, got %v", expectedMode, mode)
	}
}

func TestUpdateVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.enc")
	password := "test-password"

	// Create initial vault
	v1 := NewVault()
	v1.AddSecret("key1", "value1", "", "", nil)
	err = SaveVault(v1, vaultPath, password)
	if err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Load, modify, and save again
	v2, err := LoadVault(vaultPath, password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	v2.AddSecret("key2", "value2", "", "", nil)
	v2.RemoveSecret("key1")

	err = SaveVault(v2, vaultPath, password)
	if err != nil {
		t.Fatalf("SaveVault (update) failed: %v", err)
	}

	// Load again and verify
	v3, err := LoadVault(vaultPath, password)
	if err != nil {
		t.Fatalf("LoadVault (after update) failed: %v", err)
	}

	if len(v3.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(v3.Secrets))
	}

	_, err = v3.GetSecret("key2")
	if err != nil {
		t.Error("key2 should exist")
	}

	_, err = v3.GetSecret("key1")
	if err == nil {
		t.Error("key1 should not exist")
	}
}
