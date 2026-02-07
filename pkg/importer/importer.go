package importer

import (
	"encoding/json"
	"fmt"

	"github.com/matteospanio/secret-vault/pkg/exporter"
	"github.com/matteospanio/secret-vault/pkg/vault"
	"gopkg.in/yaml.v3"
)

// ImportMode defines how to handle existing secrets
type ImportMode string

const (
	ModeSkip      ImportMode = "skip"      // Skip existing secrets
	ModeOverwrite ImportMode = "overwrite" // Overwrite existing secrets
	ModeMerge     ImportMode = "merge"     // Merge, keeping newer version
)

// ImportOptions contains configuration for import operations
type ImportOptions struct {
	Mode   ImportMode
	Format exporter.ExportFormat
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	Imported int
	Skipped  int
	Updated  int
	Errors   []string
}

// Import imports secrets from the provided data into the vault
func Import(v *vault.Vault, data []byte, opts ImportOptions) (*ImportResult, error) {
	if v == nil {
		return nil, fmt.Errorf("vault cannot be nil")
	}

	// Parse the imported data
	var importedVault exporter.ExportedVault
	var err error

	switch opts.Format {
	case exporter.FormatJSON:
		err = json.Unmarshal(data, &importedVault)
	case exporter.FormatYAML:
		err = yaml.Unmarshal(data, &importedVault)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", opts.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse import data: %w", err)
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Import each secret
	for _, es := range importedVault.Secrets {
		// Validate secret
		if es.Name == "" {
			result.Errors = append(result.Errors, "skipping secret with empty name")
			continue
		}

		// Check if secret exists
		existing, exists := v.Secrets[es.Name]

		if exists {
			switch opts.Mode {
			case ModeSkip:
				result.Skipped++
				continue
			case ModeOverwrite:
				// Always overwrite
				v.AddSecret(es.Name, es.Value, es.Description, es.Category, es.Tags)
				result.Updated++
			case ModeMerge:
				// Keep newer version based on UpdatedAt
				if es.UpdatedAt.After(existing.UpdatedAt) {
					v.AddSecret(es.Name, es.Value, es.Description, es.Category, es.Tags)
					result.Updated++
				} else {
					result.Skipped++
				}
			}
		} else {
			// New secret
			v.AddSecret(es.Name, es.Value, es.Description, es.Category, es.Tags)
			result.Imported++
		}
	}

	return result, nil
}

// ValidateImportData validates the import data without importing it
func ValidateImportData(data []byte, format exporter.ExportFormat) error {
	var importedVault exporter.ExportedVault
	var err error

	switch format {
	case exporter.FormatJSON:
		err = json.Unmarshal(data, &importedVault)
	case exporter.FormatYAML:
		err = yaml.Unmarshal(data, &importedVault)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("invalid import data: %w", err)
	}

	// Validate structure
	if importedVault.Version == "" {
		return fmt.Errorf("missing version field")
	}

	if len(importedVault.Secrets) == 0 {
		return fmt.Errorf("no secrets found in import data")
	}

	// Validate each secret
	for i, secret := range importedVault.Secrets {
		if secret.Name == "" {
			return fmt.Errorf("secret at index %d has empty name", i)
		}
	}

	return nil
}

// ParseMode converts a string to ImportMode
func ParseMode(mode string) (ImportMode, error) {
	switch mode {
	case "skip":
		return ModeSkip, nil
	case "overwrite":
		return ModeOverwrite, nil
	case "merge":
		return ModeMerge, nil
	default:
		return "", fmt.Errorf("unsupported mode: %s (supported: skip, overwrite, merge)", mode)
	}
}
