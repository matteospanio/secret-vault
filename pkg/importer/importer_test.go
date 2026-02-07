package importer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/matteospanio/secret-vault/pkg/exporter"
	"github.com/matteospanio/secret-vault/pkg/vault"
)

func TestImportJSON(t *testing.T) {
	// Create export data
	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:        "test-secret",
				Value:       "secret-value",
				Description: "test description",
				Category:    "category1",
				Tags:        []string{"tag1", "tag2"},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	data, err := json.Marshal(exportedVault)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	v := vault.NewVault()
	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 1 {
		t.Errorf("Expected 1 imported secret, got %d", result.Imported)
	}

	secret, err := v.GetSecret("test-secret")
	if err != nil {
		t.Fatalf("Failed to get imported secret: %v", err)
	}

	if secret.Value != "secret-value" {
		t.Errorf("Expected value 'secret-value', got '%s'", secret.Value)
	}
	if secret.Category != "category1" {
		t.Errorf("Expected category 'category1', got '%s'", secret.Category)
	}
}

func TestImportYAML(t *testing.T) {
	yamlData := `
version: "1.0"
exported_at: 2024-01-01T00:00:00Z
secrets:
  - name: test-secret
    value: secret-value
    description: test description
    category: category1
    tags:
      - tag1
    created_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z
`

	v := vault.NewVault()
	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatYAML,
	}

	result, err := Import(v, []byte(yamlData), opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 1 {
		t.Errorf("Expected 1 imported secret, got %d", result.Imported)
	}

	secret, err := v.GetSecret("test-secret")
	if err != nil {
		t.Fatalf("Failed to get imported secret: %v", err)
	}

	if secret.Value != "secret-value" {
		t.Errorf("Expected value 'secret-value', got '%s'", secret.Value)
	}
}

func TestImportModeSkip(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("existing-secret", "original-value", "original", "", nil)

	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "existing-secret",
				Value:     "new-value",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeSkip,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped secret, got %d", result.Skipped)
	}
	if result.Imported != 0 {
		t.Errorf("Expected 0 imported secrets, got %d", result.Imported)
	}

	secret, _ := v.GetSecret("existing-secret")
	if secret.Value != "original-value" {
		t.Errorf("Expected original value to be preserved, got '%s'", secret.Value)
	}
}

func TestImportModeOverwrite(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("existing-secret", "original-value", "original", "", nil)

	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:        "existing-secret",
				Value:       "new-value",
				Description: "new description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Updated != 1 {
		t.Errorf("Expected 1 updated secret, got %d", result.Updated)
	}

	secret, _ := v.GetSecret("existing-secret")
	if secret.Value != "new-value" {
		t.Errorf("Expected new value, got '%s'", secret.Value)
	}
	if secret.Description != "new description" {
		t.Errorf("Expected new description, got '%s'", secret.Description)
	}
}

func TestImportModeMerge(t *testing.T) {
	v := vault.NewVault()
	
	// Add an old secret
	oldTime := time.Now().Add(-24 * time.Hour)
	oldSecret := vault.Secret{
		Name:      "test-secret",
		Value:     "old-value",
		CreatedAt: oldTime,
		UpdatedAt: oldTime,
	}
	v.Secrets["test-secret"] = oldSecret

	// Import newer version
	newTime := time.Now()
	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: newTime,
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "test-secret",
				Value:     "new-value",
				CreatedAt: newTime,
				UpdatedAt: newTime,
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeMerge,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Updated != 1 {
		t.Errorf("Expected 1 updated secret, got %d", result.Updated)
	}

	secret, _ := v.GetSecret("test-secret")
	if secret.Value != "new-value" {
		t.Errorf("Expected newer value to be kept, got '%s'", secret.Value)
	}
}

func TestImportModeMergeKeepsNewer(t *testing.T) {
	v := vault.NewVault()
	
	// Add a recent secret
	newTime := time.Now()
	newSecret := vault.Secret{
		Name:      "test-secret",
		Value:     "new-value",
		CreatedAt: newTime,
		UpdatedAt: newTime,
	}
	v.Secrets["test-secret"] = newSecret

	// Try to import older version
	oldTime := time.Now().Add(-24 * time.Hour)
	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: oldTime,
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "test-secret",
				Value:     "old-value",
				CreatedAt: oldTime,
				UpdatedAt: oldTime,
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeMerge,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped secret, got %d", result.Skipped)
	}

	secret, _ := v.GetSecret("test-secret")
	if secret.Value != "new-value" {
		t.Errorf("Expected newer value to be preserved, got '%s'", secret.Value)
	}
}

func TestImportMultipleSecrets(t *testing.T) {
	v := vault.NewVault()

	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "secret1",
				Value:     "value1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Name:      "secret2",
				Value:     "value2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Name:      "secret3",
				Value:     "value3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 3 {
		t.Errorf("Expected 3 imported secrets, got %d", result.Imported)
	}
}

func TestImportInvalidJSON(t *testing.T) {
	v := vault.NewVault()
	invalidJSON := []byte(`{"invalid": json}`)

	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	_, err := Import(v, invalidJSON, opts)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestImportNilVault(t *testing.T) {
	data := []byte(`{"version":"1.0","exported_at":"2024-01-01T00:00:00Z","secrets":[]}`)
	
	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	_, err := Import(nil, data, opts)
	if err == nil {
		t.Error("Expected error when importing to nil vault")
	}
}

func TestValidateImportDataValid(t *testing.T) {
	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "test-secret",
				Value:     "value",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	err := ValidateImportData(data, exporter.FormatJSON)
	if err != nil {
		t.Errorf("Expected valid data to pass validation, got error: %v", err)
	}
}

func TestValidateImportDataInvalid(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "invalid JSON",
			data: `{"invalid": json}`,
		},
		{
			name: "missing version",
			data: `{"exported_at":"2024-01-01T00:00:00Z","secrets":[{"name":"test","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]}`,
		},
		{
			name: "no secrets",
			data: `{"version":"1.0","exported_at":"2024-01-01T00:00:00Z","secrets":[]}`,
		},
		{
			name: "secret with empty name",
			data: `{"version":"1.0","exported_at":"2024-01-01T00:00:00Z","secrets":[{"name":"","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImportData([]byte(tt.data), exporter.FormatJSON)
			if err == nil {
				t.Error("Expected validation to fail")
			}
		})
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input    string
		expected ImportMode
		wantErr  bool
	}{
		{"skip", ModeSkip, false},
		{"overwrite", ModeOverwrite, false},
		{"merge", ModeMerge, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, got)
				}
			}
		})
	}
}

func TestImportSkipsEmptyNames(t *testing.T) {
	v := vault.NewVault()

	exportedVault := exporter.ExportedVault{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Secrets: []exporter.ExportedSecret{
			{
				Name:      "",
				Value:     "value",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Name:      "valid-secret",
				Value:     "value",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	data, _ := json.Marshal(exportedVault)

	opts := ImportOptions{
		Mode:   ModeOverwrite,
		Format: exporter.FormatJSON,
	}

	result, err := Import(v, data, opts)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 1 {
		t.Errorf("Expected 1 imported secret, got %d", result.Imported)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error for empty name, got %d", len(result.Errors))
	}
}
