package export

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

func TestJSONFormat_Marshal_WithValues(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-key", "test-value", "Test description")

	formatter := NewJSONFormat()
	data, err := formatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify it's valid JSON
	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if len(exportData.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(exportData.Secrets))
	}

	secret := exportData.Secrets[0]
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

func TestJSONFormat_Marshal_WithoutValues(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-key", "test-value", "Test description")

	formatter := NewJSONFormat()
	data, err := formatter.Marshal(v, false)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	secret := exportData.Secrets[0]
	if secret.Value != "" {
		t.Errorf("Expected empty value, got '%s'", secret.Value)
	}
	if secret.Name != "test-key" {
		t.Errorf("Expected name 'test-key', got '%s'", secret.Name)
	}
}

func TestJSONFormat_Unmarshal(t *testing.T) {
	now := time.Now()
	jsonData := `{
		"version": "1.0",
		"secrets": [
			{
				"name": "test-key",
				"value": "test-value",
				"description": "Test description",
				"created_at": "` + now.Format(time.RFC3339) + `",
				"updated_at": "` + now.Format(time.RFC3339) + `"
			}
		]
	}`

	formatter := NewJSONFormat()
	v, err := formatter.Unmarshal([]byte(jsonData))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(v.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(v.Secrets))
	}

	secret, exists := v.Secrets["test-key"]
	if !exists {
		t.Fatal("Secret 'test-key' not found")
	}

	if secret.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", secret.Value)
	}
	if secret.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", secret.Description)
	}
}

func TestJSONFormat_RoundTrip(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("key1", "value1", "Description 1")
	v.AddSecret("key2", "value2", "Description 2")

	formatter := NewJSONFormat()

	// Marshal
	data, err := formatter.Marshal(v, true)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	v2, err := formatter.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if len(v2.Secrets) != 2 {
		t.Fatalf("Expected 2 secrets, got %d", len(v2.Secrets))
	}

	for name, secret := range v.Secrets {
		secret2, exists := v2.Secrets[name]
		if !exists {
			t.Errorf("Secret '%s' not found after round trip", name)
			continue
		}

		if secret.Value != secret2.Value {
			t.Errorf("Value mismatch for '%s': expected '%s', got '%s'", name, secret.Value, secret2.Value)
		}
		if secret.Description != secret2.Description {
			t.Errorf("Description mismatch for '%s': expected '%s', got '%s'", name, secret.Description, secret2.Description)
		}
	}
}

func TestJSONFormat_FileExtension(t *testing.T) {
	formatter := NewJSONFormat()
	if ext := formatter.FileExtension(); ext != ".json" {
		t.Errorf("Expected extension '.json', got '%s'", ext)
	}
}

func TestJSONFormat_Name(t *testing.T) {
	formatter := NewJSONFormat()
	if name := formatter.Name(); name != "JSON" {
		t.Errorf("Expected name 'JSON', got '%s'", name)
	}
}
