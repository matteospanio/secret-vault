package sync

import (
	"fmt"
	"os"

	"github.com/studio-b12/gowebdav"
)

// NextcloudProvider implements SyncProvider for Nextcloud using WebDAV
type NextcloudProvider struct {
	client     *gowebdav.Client
	config     *NextcloudConfig
	remotePath string
}

// NewNextcloudProvider creates a new Nextcloud sync provider
func NewNextcloudProvider(config *NextcloudConfig) *NextcloudProvider {
	return &NextcloudProvider{
		config:     config,
		remotePath: config.Path,
	}
}

// Connect establishes connection to Nextcloud
func (n *NextcloudProvider) Connect() error {
	if n.config.URL == "" || n.config.Username == "" {
		return fmt.Errorf("invalid Nextcloud configuration: URL and username required")
	}

	// Create WebDAV client
	n.client = gowebdav.NewClient(n.config.URL, n.config.Username, n.config.Password)

	// Test connection
	if err := n.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Nextcloud: %w", err)
	}

	return nil
}

// Upload uploads the vault file to Nextcloud
func (n *NextcloudProvider) Upload(localPath string) error {
	if n.client == nil {
		return fmt.Errorf("not connected: call Connect() first")
	}

	// Read local file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Upload to Nextcloud
	if err := n.client.Write(n.remotePath, data, 0644); err != nil {
		return fmt.Errorf("failed to upload to Nextcloud: %w", err)
	}

	return nil
}

// Download downloads the vault file from Nextcloud
func (n *NextcloudProvider) Download(localPath string) error {
	if n.client == nil {
		return fmt.Errorf("not connected: call Connect() first")
	}

	// Download from Nextcloud
	data, err := n.client.Read(n.remotePath)
	if err != nil {
		return fmt.Errorf("failed to download from Nextcloud: %w", err)
	}

	// Write to local file with secure permissions
	if err := os.WriteFile(localPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

// GetRemoteMetadata retrieves metadata about the remote vault
func (n *NextcloudProvider) GetRemoteMetadata() (*VaultMetadata, error) {
	if n.client == nil {
		return nil, fmt.Errorf("not connected: call Connect() first")
	}

	// Get file info
	info, err := n.client.Stat(n.remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote file info: %w", err)
	}

	// Download file to calculate checksum
	data, err := n.client.Read(n.remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote file for checksum: %w", err)
	}

	// Calculate checksum
	checksum, err := calculateChecksumFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return &VaultMetadata{
		LastModified: info.ModTime(),
		Checksum:     checksum,
		Size:         info.Size(),
	}, nil
}

// Exists checks if the vault file exists on Nextcloud
func (n *NextcloudProvider) Exists() (bool, error) {
	if n.client == nil {
		return false, fmt.Errorf("not connected: call Connect() first")
	}

	_, err := n.client.Stat(n.remotePath)
	if err != nil {
		// Check if it's a "not found" error
		if os.IsNotExist(err) || err.Error() == "404" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check remote file: %w", err)
	}

	return true, nil
}

// Disconnect closes the connection to Nextcloud
func (n *NextcloudProvider) Disconnect() error {
	// WebDAV client doesn't need explicit disconnect
	n.client = nil
	return nil
}

// calculateChecksumFromData computes SHA-256 checksum from data
func calculateChecksumFromData(data []byte) (string, error) {
	// Create a temporary file to use the existing checksum function
	tmpfile, err := os.CreateTemp("", "vault-checksum-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.Write(data); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return calculateChecksum(tmpfile.Name())
}
