# Secret Vault

A secure command-line application for storing and managing API tokens, secrets, and sensitive credentials. Keep your tokens encrypted and easily accessible from the terminal.

## Features

- 🔒 **AES-256-GCM Encryption**: Military-grade encryption for your secrets
- 🔑 **Password-Based Protection**: Master password secures all stored secrets
- 🖥️ **Interactive TUI**: Beautiful terminal user interface with keyboard shortcuts
- 📦 **Simple CLI Interface**: Built with Cobra for robust command parsing
- 💾 **Portable Vault**: Encrypted vault file can be synced across machines
- ☁️ **Cloud Sync**: Built-in Nextcloud/WebDAV sync with conflict detection and resolution
- 🛡️ **Secure Storage**: Vault file has restrictive permissions (600)
- ⚡ **Fast Access**: Quick retrieval of tokens when you need them
- 🔧 **Flexible Configuration**: Support for flags, environment variables, and config files via Viper
- 🚀 **Shell Completion**: Auto-completion support for bash, zsh, fish, and PowerShell
- 🔗 **Git Integration**: Seamless Git workflow integration via hooks with domain-based token suggestions
- 🏷️ **Categories & Tags**: Organize secrets with categories and tags for easy filtering
- ⚠️ **Aging Warnings**: Visual indicators for secrets older than 1 year
- 📋 **Clipboard Integration**: Copy secrets directly to clipboard with one keystroke
- 🔍 **Smart Search**: Fuzzy search and advanced filtering capabilities
- 📤 **Import/Export**: Backup and migrate secrets using JSON or YAML formats with `jq`/`yq` integration

## Installation

### From Source

Requires Go 1.21 or higher.

```bash
git clone https://github.com/matteospanio/secret-vault.git
cd secret-vault
go build -o secretvault ./cmd/secretvault
sudo mv secretvault /usr/local/bin/  # Optional: install system-wide
```

## Quick Start

### 1. Initialize a New Vault

```bash
secretvault init
```

You'll be prompted to create a master password. This password encrypts your vault.

### 2. Add a Secret

```bash
secretvault add github-token
```

Enter your secret value when prompted (input is hidden for security).

### 3. Retrieve a Secret

```bash
secretvault get github-token
```

The secret value will be printed to stdout (suitable for piping to other commands).

### 4. List All Secrets

```bash
secretvault list
```

Shows all stored secret names with descriptions and timestamps.

### 5. Remove a Secret

```bash
secretvault remove github-token
```

Permanently deletes the secret from your vault.

### 6. Launch TUI (Terminal User Interface)

```bash
secretvault tui
```

Opens an interactive terminal interface for managing secrets with keyboard shortcuts, categories, tags, and visual organization.

## Terminal User Interface (TUI)

Secret Vault includes a powerful Terminal User Interface (TUI) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) that provides an intuitive way to manage your secrets.

### Launching the TUI

```bash
secretvault tui
```

You'll be prompted for your master password, then enter the interactive interface.

### TUI Features

- **List View**: Browse all secrets with search and filtering
- **Detail View**: View secret metadata with age warnings
- **Edit View**: Add or modify secrets with categories and tags
- **Filter View**: Advanced multi-criteria filtering
- **Real-time Search**: Fuzzy search as you type
- **Clipboard Integration**: Copy secrets with one keystroke
- **Aging Warnings**: Visual indicators (⚠) for secrets older than 1 year

### Keyboard Shortcuts

#### Global (All Views)
- `?` - Toggle help view
- `q` - Quit application
- `esc` - Go back / Cancel

#### List View
- `↑/↓` - Navigate secrets
- `enter` - View secret details
- `/` - Quick search (built-in filtering)
- `a` - Add new secret
- `d` - Delete selected secret
- `f` - Open advanced filter
- `F` - Clear all filters
- `C` - Clear clipboard

#### Detail View
- `r` - Reveal/hide secret value
- `c` - Copy secret to clipboard
- `C` - Clear clipboard
- `e` - Edit secret
- `esc` - Return to list

#### Edit View
- `tab`/`↓` - Next field
- `shift+tab`/`↑` - Previous field
- `ctrl+s` - Save secret
- `enter` - Save (on last field)
- `esc` - Cancel without saving

#### Filter View
- `tab`/`↓` - Next field
- `shift+tab`/`↑` - Previous field
- `ctrl+o` - Toggle 'old secrets only'
- `enter` - Apply filters
- `ctrl+r` - Clear all filters
- `esc` - Cancel without applying

### TUI Workflow Examples

#### Adding a Secret with Categories and Tags

1. Launch TUI: `secretvault tui`
2. Press `a` to add a new secret
3. Fill in the form:
   - Name: `github-token`
   - Value: (your token - masked)
   - Description: `GitHub PAT for CLI access`
   - Category: `work`
   - Tags: `github,api,development`
4. Press `ctrl+s` to save

#### Filtering Secrets

1. In list view, press `f` to open filter view
2. Enter filter criteria:
   - Search query: (optional text search)
   - Category: `work`
   - Tag: `api`
   - Toggle `ctrl+o` for old secrets only
3. Press `enter` to apply filters
4. Press `F` to clear all filters

#### Copying a Secret

1. Navigate to a secret with `↑/↓`
2. Press `enter` to view details
3. Press `c` to copy to clipboard
4. Use secret in another application
5. Press `C` to clear clipboard when done

#### Search and Quick Access

1. In list view, press `/` to activate quick search
2. Start typing to filter secrets in real-time
3. Use `↑/↓` to navigate filtered results
4. Press `esc` to clear search
5. Press `enter` on a secret to view details

## Usage Examples

### Store a GitHub Personal Access Token

```bash
$ secretvault add github-token
Enter master password:
Enter secret value:
Enter description (optional): GitHub PAT for CLI access
✓ Secret 'github-token' added successfully
```

### Store a Secret with Categories and Tags

```bash
$ secretvault add github-token --category "work" --tags "github,api,development"
Enter master password:
Enter secret value:
Enter description (optional): GitHub PAT for CLI access
✓ Secret 'github-token' added successfully
```

### Use Token in Git Operations

```bash
# Retrieve token and use it
TOKEN=$(secretvault get github-token)
git clone https://$TOKEN@github.com/user/repo.git
```

### Store Multiple Service Tokens

```bash
secretvault add npm-token
secretvault add pypi-token
secretvault add gitlab-token
secretvault add aws-access-key
```

### List All Stored Secrets

```bash
$ secretvault list
Enter master password:
NAME            DESCRIPTION              CREATED     UPDATED
----            -----------              -------     -------
github-token    GitHub PAT for CLI      2025-01-15  2025-01-15
npm-token       NPM publish token       2025-01-15  2025-01-15
pypi-token      PyPI upload token       2025-01-15  2025-01-15
```

### Filter Secrets by Category or Tag

```bash
# List secrets in "work" category
$ secretvault list --filter-category "work"

# List secrets with "api" tag
$ secretvault list --filter-tag "api"

# List old secrets (older than 1 year)
$ secretvault list --old-only

# Combine filters
$ secretvault list --filter-category "work" --filter-tag "github"
```

### Copy Secret to Clipboard

```bash
# Copy secret value directly to clipboard
$ secretvault get github-token --copy
Enter master password:
✓ Secret copied to clipboard

# Clear clipboard when done
$ secretvault clear-clipboard
✓ Clipboard cleared
```

## Import/Export

Secret Vault supports exporting and importing secrets in **JSON** and **YAML** formats for backups, migrations, and integration with external tools.

### Export Secrets

**Export metadata only (safe for documentation):**
```bash
secretvault export -f json
secretvault export -f yaml
```

**Export with values (for backup/migration):**
```bash
secretvault export --include-values -o backup.json
secretvault export --include-values -f yaml -o backup.yaml
```

### Import Secrets

**Import from file:**
```bash
secretvault import backup.json              # Skip existing secrets
secretvault import backup.json -m overwrite  # Overwrite all
secretvault import backup.yaml -m merge      # Smart merge (keep newer)
```

### Integration with jq and yq

**Filter secrets using jq:**
```bash
# Get all API secrets
secretvault export | jq '.secrets[] | select(.category == "api")'

# Get secrets with production tag
secretvault export | jq '.secrets[] | select(.tags[] | contains("production"))'
```

**Filter secrets using yq:**
```bash
# Get cloud secrets
secretvault export -f yaml | yq '.secrets[] | select(.category == "cloud")'
```

See [IMPORT_EXPORT.md](IMPORT_EXPORT.md) for detailed documentation and examples.

## Git Workflow Integration

Secret Vault integrates with Git workflows through Git hooks, providing notifications and suggestions when you interact with remote repositories.

### Installing Git Hooks

To enable Git integration in a repository:

```bash
cd /path/to/your/git/repo
secretvault git install-hooks
```

This installs a `pre-push` hook that:
- Detects when you're pushing to a remote repository
- Identifies the remote domain (e.g., github.com, gitlab.com, bitbucket.org)
- Suggests relevant tokens from your vault based on the domain
- Provides helpful reminders about using Secret Vault for authentication

### Example Hook Output

When you push to a remote, you'll see:

```bash
$ git push origin main
🔐 Secret Vault: Git hook active
   Remote: origin (https://github.com/user/repo.git)
   Domain detected: github.com
   Tip: Use 'secretvault get <token-name>' to retrieve tokens
   Or add tokens with 'secretvault add github.com-token'
```

### Uninstalling Hooks

To remove Git integration:

```bash
secretvault git uninstall-hooks
```

This removes the hooks and restores any previously existing hooks that were backed up.

### Best Practices for Git Integration

1. **Organize tokens by domain**: Name your tokens with the domain for easy identification:
   ```bash
   secretvault add github.com-token
   secretvault add gitlab.com-token
   secretvault add bitbucket.org-token
   ```

2. **Use tokens in Git operations**: Retrieve and use tokens for authentication:
   ```bash
   TOKEN=$(secretvault get github.com-token)
   git clone https://$TOKEN@github.com/user/repo.git
   ```

3. **Per-repository installation**: Install hooks on a per-repository basis to maintain flexibility.

## Configuration

The CLI supports multiple configuration methods (in order of precedence):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Default values** (lowest priority)

### Command-line Flags

All commands support global flags:

```bash
secretvault [command] --vault-path /path/to/vault.enc --password mypassword
```

Available global flags:
- `--vault-path`: Custom location for vault file (default: `~/.secret-vault/vault.enc`)
- `--password`: Master password (if set, skips password prompt - use with caution)

### Environment Variables

- `VAULT_PATH` or `VAULT_VAULT_PATH`: Custom location for vault file
- `VAULT_PASSWORD`: Master password (if set, skips password prompt - use with caution)

### Custom Vault Location

Using environment variables:
```bash
export VAULT_PATH=/path/to/my/vault.enc
secretvault init
```

Using flags:
```bash
secretvault init --vault-path /path/to/my/vault.enc
```

### Shell Completion

Generate shell completion scripts for your shell:

```bash
# Bash
secretvault completion bash > /etc/bash_completion.d/secretvault

# Zsh
secretvault completion zsh > "${fpath[1]}/_secretvault"

# Fish
secretvault completion fish > ~/.config/fish/completions/secretvault.fish

# PowerShell
secretvault completion powershell > secretvault.ps1
```

### Cloud Sync (Nextcloud/WebDAV)

Secret Vault CLI includes built-in cloud sync support for Nextcloud and other WebDAV-compatible services. This provides automatic synchronization across multiple devices while maintaining encryption.

#### Configure Sync

```bash
secretvault sync configure
# Follow prompts to enter:
# - Provider: nextcloud
# - Nextcloud URL: https://cloud.example.com
# - Username: your-username
# - Password: your-password
# - Remote path: /Vaults/vault.enc
```

#### Push Vault to Cloud

```bash
secretvault sync push
```

Uploads your encrypted vault to the configured cloud storage. The vault remains encrypted during transit and storage.

#### Pull Vault from Cloud

```bash
secretvault sync pull
```

Downloads the vault from cloud storage. Useful when setting up a new machine or retrieving updates.

#### Check Sync Status

```bash
secretvault sync status
```

Shows the current sync state, including:
- Local and remote vault metadata
- Last sync time
- Whether vaults are in sync
- Any conflicts that need resolution

#### Conflict Resolution

If both local and remote vaults have changed since the last sync, you'll see a conflict error:

```bash
$ secretvault sync push
Error: Sync conflict detected!
  Local modified: 2025-01-15 14:30:00
  Remote modified: 2025-01-15 14:35:00

To resolve, you can:
  1. Pull remote changes: secretvault sync pull
  2. Force push (overwrites remote): secretvault sync push --force
```

### Manual Syncing Across Machines

You can also manually sync by placing the vault file in cloud storage:

```bash
# On machine 1
export VAULT_PATH=~/Dropbox/vault.enc
secretvault init
secretvault add my-token

# On machine 2
export VAULT_PATH=~/Dropbox/vault.enc
secretvault list  # Access the same vault
```

## Security

### Encryption Details

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: PBKDF2 with SHA-256 (100,000 iterations)
- **Salt**: 32 bytes, randomly generated per encryption
- **Nonce**: 12 bytes, randomly generated per encryption
- **Authentication**: GCM provides authenticated encryption

### Best Practices

1. **Strong Master Password**: Use a unique, strong password for your vault
2. **Secure Storage**: The vault file is encrypted, but store it securely
3. **No Plaintext**: Secrets are never stored in plaintext
4. **Limited Exposure**: Secret values are only shown when explicitly requested
5. **File Permissions**: Vault file is created with 0600 permissions (owner read/write only)

### Security Considerations

- Master password is not stored anywhere - keep it safe
- Use `VAULT_PASSWORD` environment variable carefully (avoid in shared environments)
- The vault file can be safely backed up and synced
- Consider using different vaults for different security contexts

## Architecture

### Project Structure

```
secret-vault/
├── cmd/
│   └── secretvault/      # Main CLI application
│       └── main.go
├── pkg/
│   ├── clipboard/        # Cross-platform clipboard operations
│   │   ├── clipboard.go
│   │   └── clipboard_test.go
│   ├── crypto/           # Encryption/decryption logic
│   │   ├── crypto.go
│   │   └── crypto_test.go
│   ├── git/              # Git workflow integration
│   │   ├── git.go
│   │   └── git_test.go
│   ├── sync/             # Cloud sync functionality
│   │   ├── provider.go   # SyncProvider interface
│   │   ├── nextcloud.go  # Nextcloud/WebDAV implementation
│   │   ├── config.go     # Configuration management
│   │   ├── config_test.go
│   │   ├── metadata.go   # Version tracking and checksums
│   │   ├── metadata_test.go
│   │   └── manager.go    # Sync orchestration
│   ├── tui/              # Terminal User Interface
│   │   ├── model.go      # Main TUI state machine
│   │   ├── listview.go   # Secret list with navigation
│   │   ├── detailview.go # Secret detail display
│   │   ├── editview.go   # Secret creation/editing
│   │   ├── filterview.go # Advanced filtering
│   │   ├── inputview.go  # Input with autocomplete
│   │   ├── searchbar.go  # Quick search component
│   │   ├── autocomplete.go # Fuzzy matching logic
│   │   ├── styles.go     # UI styling
│   │   ├── messages.go   # Bubble Tea messages
│   │   └── *_test.go     # Comprehensive test suite
│   └── vault/            # Vault data structure and storage
│       ├── vault.go
│       ├── storage.go
│       ├── vault_test.go
│       └── storage_test.go
├── go.mod
└── README.md
```

### Core Components

1. **Vault Package**: Manages secret storage and retrieval with support for categories, tags, and filtering
2. **Crypto Package**: Handles AES-256-GCM encryption/decryption
3. **TUI Package**: Interactive terminal user interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
   - **Model**: Main state machine coordinating all views
   - **Views**: List, Detail, Edit, Filter, and Help views
   - **Components**: Search bar, autocomplete, and input handling
   - **Styling**: Consistent theming with [Lipgloss](https://github.com/charmbracelet/lipgloss)
4. **Clipboard Package**: Cross-platform clipboard operations using [atotto/clipboard](https://github.com/atotto/clipboard)
5. **Sync Package**: Modular cloud sync with provider abstraction
   - **SyncProvider Interface**: Extensible design for multiple cloud providers
   - **NextcloudProvider**: WebDAV-based sync for Nextcloud/ownCloud
   - **Metadata System**: Checksum verification and conflict detection
6. **Git Package**: Git workflow integration with hooks
7. **CLI**: Built with [Cobra](https://github.com/spf13/cobra) for robust command structure and [Viper](https://github.com/spf13/viper) for flexible configuration management

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o secretvault ./cmd/secretvault
```

### Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Troubleshooting

### TUI Issues

**TUI doesn't display correctly**
- Ensure your terminal supports 256 colors
- Try setting `TERM=xterm-256color`
- Minimum terminal size: 80x24

**Clipboard not working**
- **Linux**: Install `xclip` or `xsel` (`sudo apt install xclip`)
- **macOS**: Clipboard works natively
- **Windows**: Clipboard works natively
- For headless systems, clipboard functionality will be unavailable

**Secrets not filtering**
- Press `F` to clear all active filters
- Check that you're not in search mode (press `esc` to exit)
- Verify filter criteria in filter view (`f` key)

### CLI Issues

**"vault not initialized" error**
- Run `secretvault init` to create a new vault
- Or specify vault path: `secretvault --vault-path /path/to/vault.enc list`

**Permission denied errors**
- Vault file should have 0600 permissions (owner read/write only)
- Check vault directory permissions: `ls -la ~/.secret-vault/`

**Sync conflicts**
- Use `secretvault sync status` to check sync state
- Pull latest changes: `secretvault sync pull`
- Or force push: `secretvault sync push --force` (overwrites remote)

### Performance

**TUI slow with many secrets**
- The TUI uses virtual scrolling and should handle thousands of secrets
- If experiencing issues, report with vault size in GitHub issues

**Build from source fails**
- Ensure Go 1.21 or higher: `go version`
- Clear module cache: `go clean -modcache`
- Rebuild: `go build -o secretvault ./cmd/secretvault`

## License

See [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/matteospanio/secret-vault).
