package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	v.AddSecret("test-key", "test-value", "Test description", "", nil)

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
	v.AddSecret("test-key", "test-value", "", "", nil)

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
	v.AddSecret("test-key", "test-value", "", "", nil)

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
	v.AddSecret("key1", "value1", "", "", nil)
	v.AddSecret("key2", "value2", "", "", nil)
	v.AddSecret("key3", "value3", "", "", nil)

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
	v.AddSecret("test-key", "test-value", "Test description", "", nil)

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

func TestAddSecretWithCategoryAndTags(t *testing.T) {
	v := NewVault()
	v.AddSecret("api-key", "secret123", "My API key", "work", []string{"github", "api"})

	secret, exists := v.Secrets["api-key"]
	if !exists {
		t.Fatal("Secret not found")
	}

	if secret.Category != "work" {
		t.Errorf("Expected category 'work', got '%s'", secret.Category)
	}

	if len(secret.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(secret.Tags))
	}

	if secret.Tags[0] != "github" || secret.Tags[1] != "api" {
		t.Errorf("Expected tags ['github', 'api'], got %v", secret.Tags)
	}
}

func TestAddSecretWithoutCategoryAndTags(t *testing.T) {
	v := NewVault()
	v.AddSecret("simple-key", "value123", "Simple secret", "", nil)

	secret, exists := v.Secrets["simple-key"]
	if !exists {
		t.Fatal("Secret not found")
	}

	if secret.Category != "" {
		t.Errorf("Expected empty category, got '%s'", secret.Category)
	}

	if secret.Tags != nil && len(secret.Tags) != 0 {
		t.Errorf("Expected nil or empty tags, got %v", secret.Tags)
	}
}

func TestAddSecretPreservesCategoryAndTagsOnUpdate(t *testing.T) {
	v := NewVault()
	v.AddSecret("my-secret", "old-value", "desc", "personal", []string{"important"})

	// Update the secret with new value but same category/tags
	v.AddSecret("my-secret", "new-value", "new desc", "personal", []string{"important", "updated"})

	secret, _ := v.Secrets["my-secret"]

	if secret.Value != "new-value" {
		t.Errorf("Expected value 'new-value', got '%s'", secret.Value)
	}

	if secret.Category != "personal" {
		t.Errorf("Expected category 'personal', got '%s'", secret.Category)
	}

	if len(secret.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(secret.Tags))
	}
}

func TestJSONSerializationWithCategoryAndTags(t *testing.T) {
	v := NewVault()
	v.AddSecret("test-key", "test-value", "desc", "work", []string{"tag1", "tag2"})

	jsonData, err := v.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	v2, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	secret, exists := v2.Secrets["test-key"]
	if !exists {
		t.Fatal("Secret not found after deserialization")
	}

	if secret.Category != "work" {
		t.Errorf("Expected category 'work', got '%s'", secret.Category)
	}

	if len(secret.Tags) != 2 {
		t.Fatalf("Expected 2 tags after deserialization, got %d", len(secret.Tags))
	}

	if secret.Tags[0] != "tag1" || secret.Tags[1] != "tag2" {
		t.Errorf("Expected tags ['tag1', 'tag2'], got %v", secret.Tags)
	}
}

func TestBackwardCompatibilityOldVaultFormat(t *testing.T) {
	// Simulate old vault JSON without category and tags
	oldVaultJSON := []byte(`{
		"secrets": {
			"old-secret": {
				"name": "old-secret",
				"value": "old-value",
				"description": "old desc",
				"created_at": "2023-01-15T10:30:00Z",
				"updated_at": "2023-06-20T14:00:00Z"
			}
		},
		"version": "1.0"
	}`)

	v, err := FromJSON(oldVaultJSON)
	if err != nil {
		t.Fatalf("Failed to load old vault format: %v", err)
	}

	secret, exists := v.Secrets["old-secret"]
	if !exists {
		t.Fatal("Secret not found in old vault")
	}

	if secret.Name != "old-secret" {
		t.Errorf("Expected name 'old-secret', got '%s'", secret.Name)
	}

	if secret.Value != "old-value" {
		t.Errorf("Expected value 'old-value', got '%s'", secret.Value)
	}

	// Category and tags should be empty/nil for old secrets
	if secret.Category != "" {
		t.Errorf("Expected empty category for old secret, got '%s'", secret.Category)
	}

	if secret.Tags != nil && len(secret.Tags) > 0 {
		t.Errorf("Expected nil/empty tags for old secret, got %v", secret.Tags)
	}
}

func TestJSONOmitsEmptyCategoryAndTags(t *testing.T) {
	v := NewVault()
	v.AddSecret("minimal", "value", "", "", nil)

	jsonData, err := v.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	jsonStr := string(jsonData)

	// Category and tags with omitempty should not appear in JSON when empty
	if contains(jsonStr, `"category":""`) {
		t.Error("Empty category should be omitted from JSON")
	}

	if contains(jsonStr, `"tags":null`) || contains(jsonStr, `"tags":[]`) {
		t.Error("Empty/nil tags should be omitted from JSON")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSecretGetAge(t *testing.T) {
	tests := []struct {
		name        string
		updatedAt   time.Time
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{
			name:        "newly created secret",
			updatedAt:   time.Now(),
			minExpected: 0,
			maxExpected: time.Second,
		},
		{
			name:        "one day old secret",
			updatedAt:   time.Now().Add(-24 * time.Hour),
			minExpected: 24*time.Hour - time.Second,
			maxExpected: 24*time.Hour + time.Second,
		},
		{
			name:        "one year old secret",
			updatedAt:   time.Now().Add(-365 * 24 * time.Hour),
			minExpected: 365*24*time.Hour - time.Second,
			maxExpected: 365*24*time.Hour + time.Second,
		},
		{
			name:        "very old secret (5 years)",
			updatedAt:   time.Now().Add(-5 * 365 * 24 * time.Hour),
			minExpected: 5*365*24*time.Hour - time.Second,
			maxExpected: 5*365*24*time.Hour + time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := Secret{
				Name:      "test",
				Value:     "value",
				UpdatedAt: tt.updatedAt,
			}

			age := secret.GetAge()

			if age < tt.minExpected || age > tt.maxExpected {
				t.Errorf("GetAge() = %v, expected between %v and %v", age, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestSecretGetAgeWithZeroTime(t *testing.T) {
	secret := Secret{
		Name:      "test",
		Value:     "value",
		UpdatedAt: time.Time{}, // zero time
	}

	age := secret.GetAge()

	// Age from zero time should be very large (since epoch)
	if age < 50*365*24*time.Hour {
		t.Errorf("GetAge() with zero time should return age since epoch, got %v", age)
	}
}

func TestSecretIsOld(t *testing.T) {
	tests := []struct {
		name      string
		updatedAt time.Time
		threshold time.Duration
		expected  bool
	}{
		{
			name:      "new secret is not old",
			updatedAt: time.Now(),
			threshold: DefaultAgeThreshold,
			expected:  false,
		},
		{
			name:      "secret just under threshold is not old",
			updatedAt: time.Now().Add(-364 * 24 * time.Hour),
			threshold: DefaultAgeThreshold,
			expected:  false,
		},
		{
			name:      "secret just over threshold is old",
			updatedAt: time.Now().Add(-366 * 24 * time.Hour),
			threshold: DefaultAgeThreshold,
			expected:  true,
		},
		{
			name:      "very old secret is old",
			updatedAt: time.Now().Add(-5 * 365 * 24 * time.Hour),
			threshold: DefaultAgeThreshold,
			expected:  true,
		},
		{
			name:      "custom threshold - 30 days",
			updatedAt: time.Now().Add(-31 * 24 * time.Hour),
			threshold: 30 * 24 * time.Hour,
			expected:  true,
		},
		{
			name:      "custom threshold - secret under 30 days",
			updatedAt: time.Now().Add(-29 * 24 * time.Hour),
			threshold: 30 * 24 * time.Hour,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := Secret{
				Name:      "test",
				Value:     "value",
				UpdatedAt: tt.updatedAt,
			}

			result := secret.IsOld(tt.threshold)

			if result != tt.expected {
				t.Errorf("IsOld(%v) = %v, expected %v", tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestDefaultAgeThreshold(t *testing.T) {
	expectedDuration := 365 * 24 * time.Hour

	if DefaultAgeThreshold != expectedDuration {
		t.Errorf("DefaultAgeThreshold = %v, expected %v (1 year)", DefaultAgeThreshold, expectedDuration)
	}
}

func TestSecretIsOldBoundaryCondition(t *testing.T) {
	// Test exact boundary: secret updated exactly at threshold
	threshold := 24 * time.Hour
	exactlyAtThreshold := time.Now().Add(-threshold)

	secret := Secret{
		Name:      "test",
		Value:     "value",
		UpdatedAt: exactlyAtThreshold,
	}

	// At exactly the threshold, IsOld should return false (uses > not >=)
	// Due to time passing during test execution, we need some tolerance
	// The secret age will be >= threshold, so it might be just over
	result := secret.IsOld(threshold)

	// This is a boundary test - the exact result depends on timing
	// The important thing is that the function works consistently
	if secret.GetAge() > threshold && !result {
		t.Errorf("IsOld() should return true when age > threshold")
	}
	if secret.GetAge() <= threshold && result {
		t.Errorf("IsOld() should return false when age <= threshold")
	}
}
