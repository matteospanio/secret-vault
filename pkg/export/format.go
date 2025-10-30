package export

import (
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// ExportFormat defines the interface for different export formats
// This allows for easy extension to support additional formats in the future
type ExportFormat interface {
	// Marshal converts a vault to the specific format
	Marshal(v *vault.Vault, includeValues bool) ([]byte, error)
	
	// Unmarshal converts data from the specific format back to a vault
	Unmarshal(data []byte) (*vault.Vault, error)
	
	// FileExtension returns the recommended file extension for this format
	FileExtension() string
	
	// Name returns the human-readable name of the format
	Name() string
}

// ExportData represents the structure of exported vault data
type ExportData struct {
	Version string                 `json:"version" yaml:"version"`
	Secrets []ExportSecret         `json:"secrets" yaml:"secrets"`
}

// ExportSecret represents a secret in export format
type ExportSecret struct {
	Name        string `json:"name" yaml:"name"`
	Value       string `json:"value,omitempty" yaml:"value,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   string `json:"created_at" yaml:"created_at"`
	UpdatedAt   string `json:"updated_at" yaml:"updated_at"`
}
