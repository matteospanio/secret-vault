package export

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// JSONFormat implements the ExportFormat interface for JSON
type JSONFormat struct{}

// NewJSONFormat creates a new JSON formatter
func NewJSONFormat() *JSONFormat {
	return &JSONFormat{}
}

// Marshal converts a vault to JSON format
func (f *JSONFormat) Marshal(v *vault.Vault, includeValues bool) ([]byte, error) {
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

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// Unmarshal converts JSON data back to a vault
func (f *JSONFormat) Unmarshal(data []byte) (*vault.Vault, error) {
	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
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

// FileExtension returns the file extension for JSON format
func (f *JSONFormat) FileExtension() string {
	return ".json"
}

// Name returns the human-readable name of the format
func (f *JSONFormat) Name() string {
	return "JSON"
}
