package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultAgeThreshold is the default threshold for considering a secret "old" (1 year)
const DefaultAgeThreshold = 365 * 24 * time.Hour

// Secret represents a single secret entry
type Secret struct {
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetAge returns the duration since the secret was last updated
func (s *Secret) GetAge() time.Duration {
	return time.Since(s.UpdatedAt)
}

// IsOld returns true if the secret's age exceeds the given threshold
func (s *Secret) IsOld(threshold time.Duration) bool {
	return s.GetAge() > threshold
}

// Vault represents the collection of secrets
type Vault struct {
	Secrets map[string]Secret `json:"secrets"`
	Version string            `json:"version"`
}

// NewVault creates a new empty vault
func NewVault() *Vault {
	return &Vault{
		Secrets: make(map[string]Secret),
		Version: "1.0",
	}
}

// AddSecret adds or updates a secret in the vault
func (v *Vault) AddSecret(name, value, description, category string, tags []string) {
	now := time.Now()
	secret := Secret{
		Name:        name,
		Value:       value,
		Description: description,
		Category:    category,
		Tags:        tags,
		UpdatedAt:   now,
	}

	// Preserve creation time if updating existing secret
	if existing, exists := v.Secrets[name]; exists {
		secret.CreatedAt = existing.CreatedAt
	} else {
		secret.CreatedAt = now
	}

	v.Secrets[name] = secret
}

// GetSecret retrieves a secret from the vault
func (v *Vault) GetSecret(name string) (Secret, error) {
	secret, exists := v.Secrets[name]
	if !exists {
		return Secret{}, fmt.Errorf("secret '%s' not found", name)
	}
	return secret, nil
}

// RemoveSecret removes a secret from the vault
func (v *Vault) RemoveSecret(name string) error {
	if _, exists := v.Secrets[name]; !exists {
		return fmt.Errorf("secret '%s' not found", name)
	}
	delete(v.Secrets, name)
	return nil
}

// ListSecrets returns all secret names
func (v *Vault) ListSecrets() []string {
	names := make([]string, 0, len(v.Secrets))
	for name := range v.Secrets {
		names = append(names, name)
	}
	return names
}

// FilterByCategory returns all secrets matching the given category (case-insensitive)
func (v *Vault) FilterByCategory(category string) []Secret {
	result := make([]Secret, 0)
	categoryLower := strings.ToLower(category)

	for _, secret := range v.Secrets {
		if strings.ToLower(secret.Category) == categoryLower {
			result = append(result, secret)
		}
	}
	return result
}

// FilterByTag returns all secrets containing the given tag (case-insensitive)
func (v *Vault) FilterByTag(tag string) []Secret {
	result := make([]Secret, 0)
	if tag == "" {
		return result
	}
	tagLower := strings.ToLower(tag)

	for _, secret := range v.Secrets {
		for _, t := range secret.Tags {
			if strings.ToLower(t) == tagLower {
				result = append(result, secret)
				break
			}
		}
	}
	return result
}

// FilterByAge returns all secrets older than the given threshold
func (v *Vault) FilterByAge(threshold time.Duration) []Secret {
	result := make([]Secret, 0)

	for _, secret := range v.Secrets {
		if secret.IsOld(threshold) {
			result = append(result, secret)
		}
	}
	return result
}

// Search returns all secrets matching the query in name or description (case-insensitive)
func (v *Vault) Search(query string) []Secret {
	result := make([]Secret, 0)

	// Empty query returns all secrets
	if query == "" {
		for _, secret := range v.Secrets {
			result = append(result, secret)
		}
		return result
	}

	queryLower := strings.ToLower(query)

	for _, secret := range v.Secrets {
		nameLower := strings.ToLower(secret.Name)
		descLower := strings.ToLower(secret.Description)

		if strings.Contains(nameLower, queryLower) || strings.Contains(descLower, queryLower) {
			result = append(result, secret)
		}
	}
	return result
}

// ToJSON serializes the vault to JSON
func (v *Vault) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON deserializes the vault from JSON
func FromJSON(data []byte) (*Vault, error) {
	vault := NewVault()
	err := json.Unmarshal(data, vault)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// GetDefaultVaultPath returns the default vault file path
func GetDefaultVaultPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	vaultDir := filepath.Join(homeDir, ".secret-vault")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(vaultDir, "vault.enc"), nil
}
