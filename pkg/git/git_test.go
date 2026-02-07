package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		wantDomain string
	}{
		{
			name:       "HTTPS GitHub URL",
			remoteURL:  "https://github.com/user/repo.git",
			wantDomain: "github.com",
		},
		{
			name:       "HTTP URL",
			remoteURL:  "http://gitlab.com/user/repo.git",
			wantDomain: "gitlab.com",
		},
		{
			name:       "SSH GitHub URL",
			remoteURL:  "git@github.com:user/repo.git",
			wantDomain: "github.com",
		},
		{
			name:       "SSH GitLab URL",
			remoteURL:  "git@gitlab.com:user/repo.git",
			wantDomain: "gitlab.com",
		},
		{
			name:       "HTTPS with port",
			remoteURL:  "https://github.com:443/user/repo.git",
			wantDomain: "github.com:443",
		},
		{
			name:       "Invalid URL",
			remoteURL:  "invalid-url",
			wantDomain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain := ExtractDomain(tt.remoteURL)
			if gotDomain != tt.wantDomain {
				t.Errorf("ExtractDomain() = %v, want %v", gotDomain, tt.wantDomain)
			}
		})
	}
}

func TestInstallHooks(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Change to temp directory for the test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Install hooks
	vaultPath := "/tmp/test-vault.enc"
	err = InstallHooks(tmpDir, vaultPath)
	if err != nil {
		t.Fatalf("InstallHooks failed: %v", err)
	}

	// Check if pre-push hook was created
	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-push")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("Hook file was not created: %v", err)
	}

	// Check if hook is executable
	if info.Mode()&0111 == 0 {
		t.Error("Hook file is not executable")
	}

	// Read hook content
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}

	// Verify hook contains expected content
	hookStr := string(content)
	if !strings.Contains(hookStr, "Secret Vault - Git Hook") {
		t.Error("Hook does not contain Secret Vault marker")
	}
	if !strings.Contains(hookStr, "secretvault") {
		t.Error("Hook does not reference secretvault command")
	}
}

func TestUninstallHooks(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Change to temp directory for the test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Install hooks first
	err = InstallHooks(tmpDir, "")
	if err != nil {
		t.Fatalf("InstallHooks failed: %v", err)
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-push")

	// Verify hook exists
	if _, err := os.Stat(hookPath); err != nil {
		t.Fatalf("Hook was not installed: %v", err)
	}

	// Uninstall hooks
	err = UninstallHooks()
	if err != nil {
		t.Fatalf("UninstallHooks failed: %v", err)
	}

	// Verify hook was removed
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("Hook was not removed")
	}
}

func TestInstallHooksWithExistingHook(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Change to temp directory for the test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create an existing hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	existingHookPath := filepath.Join(hooksDir, "pre-push")
	existingContent := "#!/bin/sh\necho 'existing hook'\n"
	err = os.WriteFile(existingHookPath, []byte(existingContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create existing hook: %v", err)
	}

	// Install hooks (should backup existing)
	err = InstallHooks(tmpDir, "")
	if err != nil {
		t.Fatalf("InstallHooks failed: %v", err)
	}

	// Check if backup was created
	backupPath := existingHookPath + ".backup"
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("Backup was not created: %v", err)
	}

	// Verify backup contains original content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(backupContent) != existingContent {
		t.Error("Backup content does not match original hook")
	}
}

func TestGeneratePrePushHook(t *testing.T) {
	tests := []struct {
		name      string
		vaultPath string
		contains  []string
	}{
		{
			name:      "With vault path",
			vaultPath: "/custom/vault.enc",
			contains:  []string{"#!/bin/sh", "Secret Vault", "secretvault"},
		},
		{
			name:      "Without vault path",
			vaultPath: "",
			contains:  []string{"#!/bin/sh", "Secret Vault", "secretvault"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := generatePrePushHook(tt.vaultPath)
			for _, expected := range tt.contains {
				if !strings.Contains(hook, expected) {
					t.Errorf("Generated hook does not contain expected string: %s", expected)
				}
			}
			// Verify hook is executable
			if !strings.HasPrefix(hook, "#!/bin/sh") {
				t.Error("Hook does not start with proper shebang")
			}
		})
	}
}
