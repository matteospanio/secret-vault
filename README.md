# Secret Vault CLI

A secure command-line application for storing and managing API tokens, secrets, and sensitive credentials. Keep your tokens encrypted and easily accessible from the terminal.

## Features

- 🔒 **AES-256-GCM Encryption**: Military-grade encryption for your secrets
- 🔑 **Password-Based Protection**: Master password secures all stored secrets
- 📦 **Simple CLI Interface**: Built with Cobra for robust command parsing
- 💾 **Portable Vault**: Encrypted vault file can be synced across machines
- ☁️ **Cloud Sync**: Built-in Nextcloud/WebDAV sync with conflict detection and resolution
- 🛡️ **Secure Storage**: Vault file has restrictive permissions (600)
- ⚡ **Fast Access**: Quick retrieval of tokens when you need them
- 🔧 **Flexible Configuration**: Support for flags, environment variables, and config files via Viper
- 🚀 **Shell Completion**: Auto-completion support for bash, zsh, fish, and PowerShell
- 🔗 **Git Integration**: Seamless Git workflow integration via hooks with domain-based token suggestions

## Installation

### From Source

Requires Go 1.21 or higher.

```bash
git clone https://github.com/matteospanio/secret-vault-cli.git
cd secret-vault-cli
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

## Usage Examples

### Store a GitHub Personal Access Token

```bash
$ secretvault add github-token
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

## Git Workflow Integration

Secret Vault CLI integrates with Git workflows through Git hooks, providing notifications and suggestions when you interact with remote repositories.

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
🔐 Secret Vault CLI: Git hook active
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
secret-vault-cli/
├── cmd/
│   └── secretvault/      # Main CLI application
│       └── main.go
├── pkg/
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
│   └── vault/            # Vault data structure and storage
│       ├── vault.go
│       ├── storage.go
│       ├── vault_test.go
│       └── storage_test.go
├── go.mod
└── README.md
```

### Core Components

1. **Vault Package**: Manages secret storage and retrieval
2. **Crypto Package**: Handles AES-256-GCM encryption/decryption
3. **Sync Package**: Modular cloud sync with provider abstraction
   - **SyncProvider Interface**: Extensible design for multiple cloud providers
   - **NextcloudProvider**: WebDAV-based sync for Nextcloud/ownCloud
   - **Metadata System**: Checksum verification and conflict detection
4. **Git Package**: Git workflow integration with hooks
5. **CLI**: Built with [Cobra](https://github.com/spf13/cobra) for robust command structure and [Viper](https://github.com/spf13/viper) for flexible configuration management

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

## Roadmap

### Phase 1 - Core Vault ✅

- [x] Design vault data structure
- [x] Implement AES-256-GCM encryption and decryption
- [x] Build CLI commands for CRUD operations
- [x] Store configuration securely
- [x] Cross-platform support (Linux/macOS/Windows)
- [x] Comprehensive tests

### Phase 2 - Enhanced Features

- [x] Git workflow integration (git hooks)
- [x] Token suggestion mechanism based on context (via hooks)
- [x] Cloud sync with Nextcloud/WebDAV
- [x] Conflict detection and resolution
- [x] Modular sync architecture for future providers
- [ ] Shell completions (bash, zsh, fish) *(partial - via Cobra)*
- [ ] Auto-fill capabilities (advanced)
- [ ] Import/export functionality
- [ ] Secret rotation reminders

### Phase 3 - Advanced Features (Future)

- [ ] Additional sync providers (Dropbox, Google Drive, OneDrive)

- [ ] System keychain integration (macOS Keychain, Windows Credential Manager)
- [ ] Cloud secret manager integration (AWS Secrets Manager, HashiCorp Vault)
- [ ] SSH key management
- [ ] OAuth token management
- [ ] GUI front-end
- [ ] Browser extension integration

## License

See [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/matteospanio/secret-vault-cli).
