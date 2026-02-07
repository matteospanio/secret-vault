package vault

import (
	"fmt"
	"os"

	"github.com/matteospanio/secret-vault/pkg/crypto"
)

// LoadVault loads and decrypts a vault from a file
func LoadVault(path, password string) (*Vault, error) {
	// Read encrypted data
	encryptedData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	// Decrypt data
	decryptedData, err := crypto.Decrypt(encryptedData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	// Parse JSON
	vault, err := FromJSON(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault: %w", err)
	}

	return vault, nil
}

// SaveVault encrypts and saves a vault to a file
func SaveVault(vault *Vault, path, password string) error {
	// Serialize to JSON
	jsonData, err := vault.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize vault: %w", err)
	}

	// Encrypt data
	encryptedData, err := crypto.Encrypt(jsonData, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	// Write to file with secure permissions
	err = os.WriteFile(path, encryptedData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	return nil
}

// VaultExists checks if a vault file exists
func VaultExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
