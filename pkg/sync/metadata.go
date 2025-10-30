package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalMetadataStore manages local metadata for sync tracking
type LocalMetadataStore struct {
	metadataPath string
}

// NewLocalMetadataStore creates a new metadata store
func NewLocalMetadataStore(vaultPath string) *LocalMetadataStore {
	// Store metadata in the same directory as vault with .metadata suffix
	metadataPath := vaultPath + ".metadata"
	return &LocalMetadataStore{
		metadataPath: metadataPath,
	}
}

// GetLocalMetadata computes metadata for the local vault file
func (m *LocalMetadataStore) GetLocalMetadata(vaultPath string) (*VaultMetadata, error) {
	info, err := os.Stat(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat vault file: %w", err)
	}

	checksum, err := calculateChecksum(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return &VaultMetadata{
		LastModified: info.ModTime(),
		Checksum:     checksum,
		Size:         info.Size(),
	}, nil
}

// SaveSyncMetadata saves sync metadata to disk
func (m *LocalMetadataStore) SaveSyncMetadata(metadata *VaultMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.metadataPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	if err := os.WriteFile(m.metadataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// LoadSyncMetadata loads sync metadata from disk
func (m *LocalMetadataStore) LoadSyncMetadata() (*VaultMetadata, error) {
	data, err := os.ReadFile(m.metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata yet
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata VaultMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// DetectConflict checks if there's a sync conflict
func (m *LocalMetadataStore) DetectConflict(localMeta, remoteMeta, lastSyncMeta *VaultMetadata) bool {
	// No conflict if no previous sync
	if lastSyncMeta == nil {
		return false
	}

	// No conflict if checksums match
	if localMeta.Checksum == remoteMeta.Checksum {
		return false
	}

	// Conflict if both local and remote have changed since last sync
	localChanged := localMeta.Checksum != lastSyncMeta.Checksum
	remoteChanged := remoteMeta.Checksum != lastSyncMeta.Checksum

	return localChanged && remoteChanged
}

// calculateChecksum computes SHA-256 checksum of a file
func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
