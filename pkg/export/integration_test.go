package export

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// TestJSONExportImportIntegration tests full export/import cycle with JSON
func TestJSONExportImportIntegration(t *testing.T) {
	// Create a test vault
	v := vault.NewVault()
	v.AddSecret("test-key-1", "test-value-1", "Description 1")
	v.AddSecret("test-key-2", "test-value-2", "Description 2")
	v.AddSecret("test-key-3", "test-value-3", "")

	// Export to JSON with values
	formatter := NewJSONFormat()
	data, err := formatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	exportFile := filepath.Join(tmpDir, "export.json")
	if err := os.WriteFile(exportFile, data, 0600); err != nil {
		t.Fatalf("Failed to write export file: %v", err)
	}

	// Read and import
	importData, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	importedVault, err := formatter.Unmarshal(importData)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify all secrets were imported correctly
	if len(importedVault.Secrets) != 3 {
		t.Fatalf("Expected 3 secrets, got %d", len(importedVault.Secrets))
	}

	for name, originalSecret := range v.Secrets {
		importedSecret, exists := importedVault.Secrets[name]
		if !exists {
			t.Errorf("Secret '%s' not found in imported vault", name)
			continue
		}

		if importedSecret.Value != originalSecret.Value {
			t.Errorf("Value mismatch for '%s': expected '%s', got '%s'",
				name, originalSecret.Value, importedSecret.Value)
		}

		if importedSecret.Description != originalSecret.Description {
			t.Errorf("Description mismatch for '%s': expected '%s', got '%s'",
				name, originalSecret.Description, importedSecret.Description)
		}
	}
}

// TestYAMLExportImportIntegration tests full export/import cycle with YAML
func TestYAMLExportImportIntegration(t *testing.T) {
	// Create a test vault
	v := vault.NewVault()
	v.AddSecret("yaml-key-1", "yaml-value-1", "YAML Description 1")
	v.AddSecret("yaml-key-2", "yaml-value-2", "YAML Description 2")

	// Export to YAML with values
	formatter := NewYAMLFormat()
	data, err := formatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	exportFile := filepath.Join(tmpDir, "export.yaml")
	if err := os.WriteFile(exportFile, data, 0600); err != nil {
		t.Fatalf("Failed to write export file: %v", err)
	}

	// Read and import
	importData, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	importedVault, err := formatter.Unmarshal(importData)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify all secrets were imported correctly
	if len(importedVault.Secrets) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(importedVault.Secrets))
	}

	for name, originalSecret := range v.Secrets {
		importedSecret, exists := importedVault.Secrets[name]
		if !exists {
			t.Errorf("Secret '%s' not found in imported vault", name)
			continue
		}

		if importedSecret.Value != originalSecret.Value {
			t.Errorf("Value mismatch for '%s': expected '%s', got '%s'",
				name, originalSecret.Value, importedSecret.Value)
		}

		if importedSecret.Description != originalSecret.Description {
			t.Errorf("Description mismatch for '%s': expected '%s', got '%s'",
				name, originalSecret.Description, importedSecret.Description)
		}
	}
}

// TestMetadataOnlyExport tests that metadata-only exports don't include values
func TestMetadataOnlyExport(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-key", "secret-value", "Secret Description")

	// Test JSON
	jsonFormatter := NewJSONFormat()
	jsonData, err := jsonFormatter.Marshal(v, false)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	importedJSON, err := jsonFormatter.Unmarshal(jsonData)
	if err != nil {
		t.Fatalf("JSON import failed: %v", err)
	}

	if len(importedJSON.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(importedJSON.Secrets))
	}

	secret := importedJSON.Secrets["secret-key"]
	if secret.Value != "" {
		t.Errorf("Expected empty value in metadata-only export, got '%s'", secret.Value)
	}
	if secret.Name != "secret-key" {
		t.Errorf("Expected name 'secret-key', got '%s'", secret.Name)
	}
	if secret.Description != "Secret Description" {
		t.Errorf("Expected description 'Secret Description', got '%s'", secret.Description)
	}

	// Test YAML
	yamlFormatter := NewYAMLFormat()
	yamlData, err := yamlFormatter.Marshal(v, false)
	if err != nil {
		t.Fatalf("YAML export failed: %v", err)
	}

	importedYAML, err := yamlFormatter.Unmarshal(yamlData)
	if err != nil {
		t.Fatalf("YAML import failed: %v", err)
	}

	if len(importedYAML.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(importedYAML.Secrets))
	}

	secretYAML := importedYAML.Secrets["secret-key"]
	if secretYAML.Value != "" {
		t.Errorf("Expected empty value in metadata-only export, got '%s'", secretYAML.Value)
	}
}

// TestCrossFormatCompatibility tests that the same vault can be exported to JSON and YAML
func TestCrossFormatCompatibility(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("cross-key-1", "cross-value-1", "Cross Description 1")
	v.AddSecret("cross-key-2", "cross-value-2", "Cross Description 2")

	// Export to JSON
	jsonFormatter := NewJSONFormat()
	jsonData, err := jsonFormatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	// Export to YAML
	yamlFormatter := NewYAMLFormat()
	yamlData, err := yamlFormatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("YAML export failed: %v", err)
	}

	// Import both
	jsonVault, err := jsonFormatter.Unmarshal(jsonData)
	if err != nil {
		t.Fatalf("JSON import failed: %v", err)
	}

	yamlVault, err := yamlFormatter.Unmarshal(yamlData)
	if err != nil {
		t.Fatalf("YAML import failed: %v", err)
	}

	// Verify both imported vaults are identical
	if len(jsonVault.Secrets) != len(yamlVault.Secrets) {
		t.Fatalf("Secret count mismatch: JSON=%d, YAML=%d",
			len(jsonVault.Secrets), len(yamlVault.Secrets))
	}

	for name, jsonSecret := range jsonVault.Secrets {
		yamlSecret, exists := yamlVault.Secrets[name]
		if !exists {
			t.Errorf("Secret '%s' in JSON but not in YAML", name)
			continue
		}

		if jsonSecret.Value != yamlSecret.Value {
			t.Errorf("Value mismatch for '%s': JSON='%s', YAML='%s'",
				name, jsonSecret.Value, yamlSecret.Value)
		}
	}
}

// TestLargeVaultExportImport tests export/import with many secrets
func TestLargeVaultExportImport(t *testing.T) {
	v := vault.NewVault()

	// Add 100 secrets
	for i := 0; i < 100; i++ {
		v.AddSecret(
			formatKeyName(i),
			formatValue(i),
			formatDescription(i),
		)
	}

	// Export to JSON
	formatter := NewJSONFormat()
	data, err := formatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import
	importedVault, err := formatter.Unmarshal(data)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify count
	if len(importedVault.Secrets) != 100 {
		t.Fatalf("Expected 100 secrets, got %d", len(importedVault.Secrets))
	}

	// Spot check a few secrets
	for name, originalSecret := range v.Secrets {
		importedSecret, exists := importedVault.Secrets[name]
		if !exists {
			t.Errorf("Secret '%s' not found in imported vault", name)
			continue
		}

		if importedSecret.Value != originalSecret.Value {
			t.Errorf("Value mismatch for '%s'", name)
			break
		}
	}
}

// Helper functions for test data generation

func formatKeyName(i int) string {
	return string(rune('a'+i%26)) + string(rune('a'+(i/26)%26)) + "-key"
}

func formatValue(i int) string {
	return "value-" + string(rune('0'+i%10))
}

func formatDescription(i int) string {
	return "Description " + string(rune('0'+i%10))
}
