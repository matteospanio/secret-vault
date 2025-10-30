package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetLocalMetadata(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.vault")
	testData := []byte("test vault data")

	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewLocalMetadataStore(testFile)
	meta, err := store.GetLocalMetadata(testFile)
	if err != nil {
		t.Fatalf("GetLocalMetadata failed: %v", err)
	}

	if meta.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), meta.Size)
	}

	if meta.Checksum == "" {
		t.Error("Expected non-empty checksum")
	}

	if meta.LastModified.IsZero() {
		t.Error("Expected non-zero last modified time")
	}
}

func TestSaveAndLoadSyncMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.vault")

	store := NewLocalMetadataStore(testFile)

	// Create test metadata
	originalMeta := &VaultMetadata{
		Checksum: "abc123",
		Size:     1024,
	}

	// Save metadata
	if err := store.SaveSyncMetadata(originalMeta); err != nil {
		t.Fatalf("SaveSyncMetadata failed: %v", err)
	}

	// Load metadata
	loadedMeta, err := store.LoadSyncMetadata()
	if err != nil {
		t.Fatalf("LoadSyncMetadata failed: %v", err)
	}

	if loadedMeta.Checksum != originalMeta.Checksum {
		t.Errorf("Expected checksum %s, got %s", originalMeta.Checksum, loadedMeta.Checksum)
	}

	if loadedMeta.Size != originalMeta.Size {
		t.Errorf("Expected size %d, got %d", originalMeta.Size, loadedMeta.Size)
	}
}

func TestLoadSyncMetadataNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.vault")

	store := NewLocalMetadataStore(testFile)

	// Try to load non-existent metadata
	meta, err := store.LoadSyncMetadata()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if meta != nil {
		t.Error("Expected nil metadata for non-existent file")
	}
}

func TestDetectConflict(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.vault")
	store := NewLocalMetadataStore(testFile)

	tests := []struct {
		name           string
		localMeta      *VaultMetadata
		remoteMeta     *VaultMetadata
		lastSyncMeta   *VaultMetadata
		expectConflict bool
	}{
		{
			name: "No conflict - checksums match",
			localMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			remoteMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			lastSyncMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			expectConflict: false,
		},
		{
			name: "No conflict - no previous sync",
			localMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			remoteMeta: &VaultMetadata{
				Checksum: "def456",
			},
			lastSyncMeta:   nil,
			expectConflict: false,
		},
		{
			name: "Conflict - both changed",
			localMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			remoteMeta: &VaultMetadata{
				Checksum: "def456",
			},
			lastSyncMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			expectConflict: true,
		},
		{
			name: "No conflict - only local changed",
			localMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			remoteMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			lastSyncMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			expectConflict: false,
		},
		{
			name: "No conflict - only remote changed",
			localMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			remoteMeta: &VaultMetadata{
				Checksum: "abc123",
			},
			lastSyncMeta: &VaultMetadata{
				Checksum: "xyz789",
			},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := store.DetectConflict(tt.localMeta, tt.remoteMeta, tt.lastSyncMeta)
			if conflict != tt.expectConflict {
				t.Errorf("Expected conflict=%v, got %v", tt.expectConflict, conflict)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("test data")

	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum1, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	if checksum1 == "" {
		t.Error("Expected non-empty checksum")
	}

	// Same file should give same checksum
	checksum2, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("Same file should produce same checksum")
	}

	// Different file should give different checksum
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("different data"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum3, err := calculateChecksum(testFile2)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	if checksum1 == checksum3 {
		t.Error("Different files should produce different checksums")
	}
}
