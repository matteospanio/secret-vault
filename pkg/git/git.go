package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitRemoteURL returns the URL of the remote repository
func GetGitRemoteURL(remoteName string) (string, error) {
	cmd := exec.Command("git", "config", "--get", fmt.Sprintf("remote.%s.url", remoteName))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetGitRootDir returns the root directory of the git repository
func GetGitRootDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ExtractDomain extracts the domain from a git remote URL
func ExtractDomain(remoteURL string) string {
	// Handle HTTPS URLs: https://github.com/user/repo.git
	if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
		parts := strings.Split(remoteURL, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	// Handle SSH URLs: git@github.com:user/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, "@")
		if len(parts) >= 2 {
			hostPart := strings.Split(parts[1], ":")[0]
			return hostPart
		}
	}

	return ""
}

// InstallHooks installs git hooks in the current repository
func InstallHooks(hooksPath, vaultPath string) error {
	gitRoot, err := GetGitRootDir()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(gitRoot, ".git", "hooks")

	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Install pre-push hook
	prePushPath := filepath.Join(hooksDir, "pre-push")
	if err := installPrePushHook(prePushPath, vaultPath); err != nil {
		return err
	}

	return nil
}

// installPrePushHook creates and installs the pre-push hook
func installPrePushHook(hookPath, vaultPath string) error {
	hookContent := generatePrePushHook(vaultPath)

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Backup existing hook
		backupPath := hookPath + ".backup"
		if err := os.Rename(hookPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing hook: %w", err)
		}
	}

	// Write new hook
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	return nil
}

// generatePrePushHook generates the content for the pre-push hook
func generatePrePushHook(vaultPath string) string {
	vaultPathArg := ""
	if vaultPath != "" {
		vaultPathArg = fmt.Sprintf(" --vault-path %s", vaultPath)
	}

	hookContent := `#!/bin/sh
# Secret Vault CLI - Git Hook
# This hook integrates with Secret Vault CLI to help manage authentication

# Get the remote name from the first argument
remote="$1"
url="$2"

# Check if Secret Vault CLI is available
if ! command -v secretvault >/dev/null 2>&1; then
    echo "Note: secretvault command not found. Skipping token check." >&2
    exit 0
fi

# Notify user that Secret Vault integration is available
echo "🔐 Secret Vault CLI: Git hook active" >&2
echo "   Remote: $remote ($url)" >&2

# Extract domain from URL
domain=""
case "$url" in
    https://*)
        domain=$(echo "$url" | sed 's|https://||' | cut -d'/' -f1)
        ;;
    http://*)
        domain=$(echo "$url" | sed 's|http://||' | cut -d'/' -f1)
        ;;
    git@*)
        domain=$(echo "$url" | sed 's|git@||' | cut -d':' -f1)
        ;;
esac

if [ -n "$domain" ]; then
    echo "   Domain detected: $domain" >&2
    echo "   Tip: Use 'secretvault get <token-name>' to retrieve tokens" >&2
    echo "   Or add tokens with 'secretvault add ${domain}-token'" >&2
fi

# Allow push to continue
exit 0
`
	// Note: vaultPathArg is currently unused but reserved for future enhancement
	// where the hook could directly query the vault for matching tokens
	_ = vaultPathArg
	return hookContent
}

// UninstallHooks removes git hooks installed by Secret Vault
func UninstallHooks() error {
	gitRoot, err := GetGitRootDir()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(gitRoot, ".git", "hooks")
	prePushPath := filepath.Join(hooksDir, "pre-push")

	// Check if it's our hook
	if _, err := os.Stat(prePushPath); err == nil {
		content, err := os.ReadFile(prePushPath)
		if err == nil && strings.Contains(string(content), "Secret Vault CLI - Git Hook") {
			// Remove the hook
			if err := os.Remove(prePushPath); err != nil {
				return fmt.Errorf("failed to remove hook: %w", err)
			}

			// Restore backup if exists
			backupPath := prePushPath + ".backup"
			if _, err := os.Stat(backupPath); err == nil {
				if err := os.Rename(backupPath, prePushPath); err != nil {
					return fmt.Errorf("failed to restore backup: %w", err)
				}
			}
		}
	}

	return nil
}
