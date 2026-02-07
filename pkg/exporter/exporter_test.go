package exporter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/matteospanio/secret-vault/pkg/vault"
	"gopkg.in/yaml.v3"
)

func TestExportJSON(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "secret-value", "test description", "category1", []string{"tag1", "tag2"})

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
		PrettyPrint:   true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify it's valid JSON
	var exported ExportedVault
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify content
	if len(exported.Secrets) != 1 {
		t.Errorf("Expected 1 secret, got %d", len(exported.Secrets))
	}

	secret := exported.Secrets[0]
	if secret.Name != "test-secret" {
		t.Errorf("Expected name 'test-secret', got '%s'", secret.Name)
	}
	if secret.Value != "secret-value" {
		t.Errorf("Expected value 'secret-value', got '%s'", secret.Value)
	}
	if secret.Description != "test description" {
		t.Errorf("Expected description 'test description', got '%s'", secret.Description)
	}
	if secret.Category != "category1" {
		t.Errorf("Expected category 'category1', got '%s'", secret.Category)
	}
	if len(secret.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(secret.Tags))
	}
}

func TestExportYAML(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "secret-value", "test description", "category1", []string{"tag1"})

	opts := ExportOptions{
		Format:        FormatYAML,
		IncludeValues: true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify it's valid YAML
	var exported ExportedVault
	if err := yaml.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Verify content
	if len(exported.Secrets) != 1 {
		t.Errorf("Expected 1 secret, got %d", len(exported.Secrets))
	}

	secret := exported.Secrets[0]
	if secret.Name != "test-secret" {
		t.Errorf("Expected name 'test-secret', got '%s'", secret.Name)
	}
	if secret.Value != "secret-value" {
		t.Errorf("Expected value 'secret-value', got '%s'", secret.Value)
	}
}

func TestExportWithoutValues(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "secret-value", "test description", "", nil)

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: false, // Don't include values
		PrettyPrint:   true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var exported ExportedVault
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	secret := exported.Secrets[0]
	if secret.Value != "" {
		t.Errorf("Expected empty value, got '%s'", secret.Value)
	}
	if secret.Name != "test-secret" {
		t.Errorf("Expected name to be preserved")
	}
}

func TestExportMultipleSecrets(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret1", "value1", "desc1", "cat1", []string{"tag1"})
	v.AddSecret("secret2", "value2", "desc2", "cat2", []string{"tag2"})
	v.AddSecret("secret3", "value3", "desc3", "cat1", []string{"tag1", "tag3"})

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
		PrettyPrint:   false,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var exported ExportedVault
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(exported.Secrets) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(exported.Secrets))
	}
}

func TestExportNilVault(t *testing.T) {
	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
	}

	_, err := Export(nil, opts)
	if err == nil {
		t.Error("Expected error when exporting nil vault")
	}
}

func TestExportUnsupportedFormat(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test", "value", "", "", nil)

	opts := ExportOptions{
		Format:        "xml", // Unsupported
		IncludeValues: true,
	}

	_, err := Export(v, opts)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected ExportFormat
		wantErr  bool
	}{
		{"json", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"xml", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
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

func TestExportPreservesTimestamps(t *testing.T) {
	v := vault.NewVault()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	// Manually create a secret with specific timestamps
	secret := vault.Secret{
		Name:        "test-secret",
		Value:       "test-value",
		Description: "test",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	v.Secrets["test-secret"] = secret

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var exported ExportedVault
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	exportedSecret := exported.Secrets[0]
	
	// Allow small time difference due to marshaling
	if exportedSecret.CreatedAt.Sub(createdAt).Abs() > time.Second {
		t.Errorf("CreatedAt not preserved correctly")
	}
	if exportedSecret.UpdatedAt.Sub(updatedAt).Abs() > time.Second {
		t.Errorf("UpdatedAt not preserved correctly")
	}
}

func TestExportJSONPrettyPrint(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test", "value", "", "", nil)

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
		PrettyPrint:   true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Pretty printed JSON should contain newlines and indentation
	dataStr := string(data)
	if !strings.Contains(dataStr, "\n") {
		t.Error("Expected pretty-printed JSON to contain newlines")
	}
	if !strings.Contains(dataStr, "  ") {
		t.Error("Expected pretty-printed JSON to contain indentation")
	}
}

func TestExportEmptyVault(t *testing.T) {
	v := vault.NewVault()

	opts := ExportOptions{
		Format:        FormatJSON,
		IncludeValues: true,
		PrettyPrint:   true,
	}

	data, err := Export(v, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var exported ExportedVault
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(exported.Secrets) != 0 {
		t.Errorf("Expected 0 secrets in empty vault, got %d", len(exported.Secrets))
	}
	if exported.Version == "" {
		t.Error("Expected version to be set")
	}
}
