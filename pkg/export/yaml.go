package export

import (
	"fmt"
	"time"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
	"gopkg.in/yaml.v3"
)

// YAMLFormat implements the ExportFormat interface for YAML
type YAMLFormat struct{}

// NewYAMLFormat creates a new YAML formatter
func NewYAMLFormat() *YAMLFormat {
	return &YAMLFormat{}
}

// Marshal converts a vault to YAML format
func (f *YAMLFormat) Marshal(v *vault.Vault, includeValues bool) ([]byte, error) {
	exportData := ExportData{
		Version: v.Version,
		Secrets: make([]ExportSecret, 0, len(v.Secrets)),
	}

	for _, secret := range v.Secrets {
		exportSecret := ExportSecret{
			Name:        secret.Name,
			Description: secret.Description,
			CreatedAt:   secret.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   secret.UpdatedAt.Format(time.RFC3339),
		}
		
		if includeValues {
			exportSecret.Value = secret.Value
		}
		
		exportData.Secrets = append(exportData.Secrets, exportSecret)
	}

	data, err := yaml.Marshal(exportData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return data, nil
}

// Unmarshal converts YAML data back to a vault
func (f *YAMLFormat) Unmarshal(data []byte) (*vault.Vault, error) {
	var exportData ExportData
	if err := yaml.Unmarshal(data, &exportData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	v := vault.NewVault()
	v.Version = exportData.Version

	for _, exportSecret := range exportData.Secrets {
		createdAt, err := time.Parse(time.RFC3339, exportSecret.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid created_at for secret '%s': %w", exportSecret.Name, err)
		}

		updatedAt, err := time.Parse(time.RFC3339, exportSecret.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_at for secret '%s': %w", exportSecret.Name, err)
		}

		secret := vault.Secret{
			Name:        exportSecret.Name,
			Value:       exportSecret.Value,
			Description: exportSecret.Description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		v.Secrets[secret.Name] = secret
	}

	return v, nil
}

// FileExtension returns the file extension for YAML format
func (f *YAMLFormat) FileExtension() string {
	return ".yaml"
}

// Name returns the human-readable name of the format
func (f *YAMLFormat) Name() string {
	return "YAML"
}
