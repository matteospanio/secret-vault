package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Secret represents a single secret entry
type Secret struct {
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
func (v *Vault) AddSecret(name, value, description string) {
	now := time.Now()
	secret := Secret{
		Name:        name,
		Value:       value,
		Description: description,
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
