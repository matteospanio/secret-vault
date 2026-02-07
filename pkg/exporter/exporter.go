package exporter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/matteospanio/secret-vault/pkg/vault"
	"gopkg.in/yaml.v3"
)

// ExportFormat represents the supported export formats
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
)

// ExportOptions contains configuration for export operations
type ExportOptions struct {
	Format          ExportFormat
	IncludeValues   bool // Include secret values in export
	PrettyPrint     bool // Format JSON/YAML with indentation
}

// ExportedSecret represents a secret in the export format
type ExportedSecret struct {
	Name        string    `json:"name" yaml:"name"`
	Value       string    `json:"value,omitempty" yaml:"value,omitempty"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Category    string    `json:"category,omitempty" yaml:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty" yaml:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// ExportedVault represents the vault structure for export
type ExportedVault struct {
	Version    string           `json:"version" yaml:"version"`
	ExportedAt time.Time        `json:"exported_at" yaml:"exported_at"`
	Secrets    []ExportedSecret `json:"secrets" yaml:"secrets"`
}

// Export exports the vault to the specified format
func Export(v *vault.Vault, opts ExportOptions) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("vault cannot be nil")
	}

	// Convert vault secrets to export format
	exportedSecrets := make([]ExportedSecret, 0, len(v.Secrets))
	for _, secret := range v.Secrets {
		es := ExportedSecret{
			Name:        secret.Name,
			Description: secret.Description,
			Category:    secret.Category,
			Tags:        secret.Tags,
			CreatedAt:   secret.CreatedAt,
			UpdatedAt:   secret.UpdatedAt,
		}
		
		// Include values only if requested
		if opts.IncludeValues {
			es.Value = secret.Value
		}
		
		exportedSecrets = append(exportedSecrets, es)
	}

	exportedVault := ExportedVault{
		Version:    v.Version,
		ExportedAt: time.Now(),
		Secrets:    exportedSecrets,
	}

	// Serialize based on format
	switch opts.Format {
	case FormatJSON:
		if opts.PrettyPrint {
			return json.MarshalIndent(exportedVault, "", "  ")
		}
		return json.Marshal(exportedVault)
	case FormatYAML:
		return yaml.Marshal(exportedVault)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ParseFormat converts a string to ExportFormat
func ParseFormat(format string) (ExportFormat, error) {
	switch format {
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: json, yaml)", format)
	}
}
