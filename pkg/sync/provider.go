package sync

import (
	"fmt"
	"time"
)

// SyncProvider defines the interface for cloud sync providers
type SyncProvider interface {
	// Connect establishes connection to the cloud provider
	Connect() error

	// Upload uploads the vault file to the cloud
	Upload(localPath string) error

	// Download downloads the vault file from the cloud
	Download(localPath string) error

	// GetRemoteMetadata retrieves metadata about the remote vault
	GetRemoteMetadata() (*VaultMetadata, error)

	// Exists checks if the vault file exists on the remote
	Exists() (bool, error)

	// Disconnect closes the connection to the cloud provider
	Disconnect() error
}

// VaultMetadata contains version and sync information
type VaultMetadata struct {
	LastModified time.Time `json:"last_modified"`
	Checksum     string    `json:"checksum"`
	Size         int64     `json:"size"`
}

// ConflictResolution represents how to resolve sync conflicts
type ConflictResolution int

const (
	// KeepLocal keeps the local version
	KeepLocal ConflictResolution = iota
	// KeepRemote keeps the remote version
	KeepRemote
	// Abort aborts the sync operation
	Abort
)

// ErrConflict is returned when there's a sync conflict
type ErrConflict struct {
	Local  *VaultMetadata
	Remote *VaultMetadata
}

func (e *ErrConflict) Error() string {
	return fmt.Sprintf("sync conflict detected: local modified at %v, remote modified at %v",
		e.Local.LastModified, e.Remote.LastModified)
}
