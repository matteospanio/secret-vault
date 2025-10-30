package sync

import (
	"fmt"
)

// Manager handles sync operations with conflict detection
type Manager struct {
	provider      SyncProvider
	metadataStore *LocalMetadataStore
	vaultPath     string
}

// NewManager creates a new sync manager
func NewManager(provider SyncProvider, vaultPath string) *Manager {
	return &Manager{
		provider:      provider,
		metadataStore: NewLocalMetadataStore(vaultPath),
		vaultPath:     vaultPath,
	}
}

// Push uploads the local vault to the remote
func (m *Manager) Push(force bool) error {
	// Get local metadata
	localMeta, err := m.metadataStore.GetLocalMetadata(m.vaultPath)
	if err != nil {
		return fmt.Errorf("failed to get local metadata: %w", err)
	}

	// Check if remote exists
	exists, err := m.provider.Exists()
	if err != nil {
		return fmt.Errorf("failed to check remote: %w", err)
	}

	if exists && !force {
		// Get remote metadata
		remoteMeta, err := m.provider.GetRemoteMetadata()
		if err != nil {
			return fmt.Errorf("failed to get remote metadata: %w", err)
		}

		// Load last sync metadata
		lastSyncMeta, err := m.metadataStore.LoadSyncMetadata()
		if err != nil {
			return fmt.Errorf("failed to load sync metadata: %w", err)
		}

		// Detect conflict
		if m.metadataStore.DetectConflict(localMeta, remoteMeta, lastSyncMeta) {
			return &ErrConflict{
				Local:  localMeta,
				Remote: remoteMeta,
			}
		}
	}

	// Upload
	if err := m.provider.Upload(m.vaultPath); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	// Save sync metadata
	if err := m.metadataStore.SaveSyncMetadata(localMeta); err != nil {
		return fmt.Errorf("failed to save sync metadata: %w", err)
	}

	return nil
}

// Pull downloads the remote vault to local
func (m *Manager) Pull(force bool) error {
	// Check if remote exists
	exists, err := m.provider.Exists()
	if err != nil {
		return fmt.Errorf("failed to check remote: %w", err)
	}

	if !exists {
		return fmt.Errorf("no remote vault found: use 'push' to upload first")
	}

	// Get remote metadata
	remoteMeta, err := m.provider.GetRemoteMetadata()
	if err != nil {
		return fmt.Errorf("failed to get remote metadata: %w", err)
	}

	if !force {
		// Get local metadata (if exists)
		localMeta, err := m.metadataStore.GetLocalMetadata(m.vaultPath)
		if err == nil {
			// Load last sync metadata
			lastSyncMeta, err := m.metadataStore.LoadSyncMetadata()
			if err != nil {
				return fmt.Errorf("failed to load sync metadata: %w", err)
			}

			// Detect conflict
			if m.metadataStore.DetectConflict(localMeta, remoteMeta, lastSyncMeta) {
				return &ErrConflict{
					Local:  localMeta,
					Remote: remoteMeta,
				}
			}
		}
	}

	// Download
	if err := m.provider.Download(m.vaultPath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Save sync metadata
	if err := m.metadataStore.SaveSyncMetadata(remoteMeta); err != nil {
		return fmt.Errorf("failed to save sync metadata: %w", err)
	}

	return nil
}

// Status returns the sync status
func (m *Manager) Status() (*SyncStatus, error) {
	status := &SyncStatus{}

	// Check local vault
	localMeta, err := m.metadataStore.GetLocalMetadata(m.vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get local metadata: %w", err)
	}
	status.LocalExists = true
	status.LocalMetadata = localMeta

	// Check remote vault
	exists, err := m.provider.Exists()
	if err != nil {
		return nil, fmt.Errorf("failed to check remote: %w", err)
	}
	status.RemoteExists = exists

	if exists {
		remoteMeta, err := m.provider.GetRemoteMetadata()
		if err != nil {
			return nil, fmt.Errorf("failed to get remote metadata: %w", err)
		}
		status.RemoteMetadata = remoteMeta

		// Load last sync metadata
		lastSyncMeta, err := m.metadataStore.LoadSyncMetadata()
		if err == nil {
			status.LastSyncMetadata = lastSyncMeta
			status.InSync = (localMeta.Checksum == remoteMeta.Checksum)
			status.HasConflict = m.metadataStore.DetectConflict(localMeta, remoteMeta, lastSyncMeta)
		}
	}

	return status, nil
}

// SyncStatus represents the current sync status
type SyncStatus struct {
	LocalExists      bool
	RemoteExists     bool
	InSync           bool
	HasConflict      bool
	LocalMetadata    *VaultMetadata
	RemoteMetadata   *VaultMetadata
	LastSyncMetadata *VaultMetadata
}
