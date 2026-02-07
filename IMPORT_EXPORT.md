# Import/Export Functionality

The Secret Vault CLI supports exporting and importing secrets in **JSON** and **YAML** formats, enabling backups, migrations, and integration with external tools like `jq` and `yq`.

## Export

### Basic Export (Metadata Only - Safe)

Export secret metadata without values (safe for version control or sharing structure):

```bash
secretvault export -f json
secretvault export -f yaml
```

Output includes: names, descriptions, categories, tags, and timestamps, but **NOT** secret values.

### Export with Values (Requires Confirmation)

Export secrets including their values (for backups or migration):

```bash
secretvault export -f json --include-values -o backup.json
secretvault export -f yaml --include-values -o backup.yaml
```

⚠️ You will be prompted to confirm before exporting values in plain text.

### Export Options

- `-f, --format`: Output format (`json` or `yaml`) [default: json]
- `-o, --output`: Output file path (defaults to stdout)
- `--include-values`: Include secret values in export (requires confirmation)

### Export Examples

**Export to stdout (pretty-printed JSON):**
```bash
secretvault export
```

**Export metadata to a file:**
```bash
secretvault export -f yaml -o secrets-structure.yaml
```

**Full backup with values:**
```bash
secretvault export --include-values -o vault-backup-$(date +%Y%m%d).json
```

## Import

### Basic Import

Import secrets from a JSON or YAML file:

```bash
secretvault import backup.json
secretvault import backup.yaml
```

Format is auto-detected from file extension. Use `-f` to specify explicitly.

### Import Modes

Control how existing secrets are handled:

- **skip** (default): Skip secrets that already exist
- **overwrite**: Replace all existing secrets with imported versions
- **merge**: Keep the newer version based on `updated_at` timestamp

```bash
secretvault import backup.json -m skip       # Skip duplicates
secretvault import backup.json -m overwrite  # Overwrite all
secretvault import backup.json -m merge      # Smart merge (keep newer)
```

### Import Options

- `-f, --format`: Input format (`json` or `yaml`) - auto-detected if not specified
- `-m, --mode`: Import mode (`skip`, `overwrite`, or `merge`) [default: skip]

### Import Examples

**Import new secrets only:**
```bash
secretvault import team-secrets.json
```

**Force overwrite all secrets:**
```bash
secretvault import backup.json -m overwrite
```

**Merge with keeping newer versions:**
```bash
secretvault import updated-secrets.yaml -m merge
```

## Integration with External Tools

### Using jq (JSON Query Tool)

Filter and manipulate exported JSON data:

**Select secrets by category:**
```bash
secretvault export -f json | jq '.secrets[] | select(.category == "api")'
```

**Get secret names with specific tag:**
```bash
secretvault export | jq '.secrets[] | select(.tags[] | contains("production")) | .name'
```

**Count secrets by category:**
```bash
secretvault export | jq '.secrets | group_by(.category) | map({category: .[0].category, count: length})'
```

**Extract only secret names:**
```bash
secretvault export | jq -r '.secrets[].name'
```

### Using yq (YAML Query Tool)

Filter and manipulate exported YAML data:

**Select secrets by tag:**
```bash
secretvault export -f yaml | yq '.secrets[] | select(.tags[] | contains("production"))'
```

**Get all secrets in cloud category:**
```bash
secretvault export -f yaml | yq '.secrets[] | select(.category == "cloud")'
```

**List secret names and categories:**
```bash
secretvault export -f yaml | yq '.secrets[] | [.name, .category]'
```

## Export/Import Data Format

### JSON Structure

```json
{
  "version": "1.0",
  "exported_at": "2024-02-07T12:00:00Z",
  "secrets": [
    {
      "name": "github-token",
      "value": "ghp_...",           // Only present with --include-values
      "description": "GitHub API token",
      "category": "api",
      "tags": ["github", "api"],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-02-01T14:20:00Z"
    }
  ]
}
```

### YAML Structure

```yaml
version: "1.0"
exported_at: 2024-02-07T12:00:00Z
secrets:
  - name: github-token
    value: ghp_...                   # Only present with --include-values
    description: GitHub API token
    category: api
    tags:
      - github
      - api
    created_at: 2024-01-15T10:30:00Z
    updated_at: 2024-02-01T14:20:00Z
```

## Use Cases

### Backup Workflow

```bash
# Create a dated backup with values
secretvault export --include-values -o "vault-backup-$(date +%Y%m%d).json"

# Restore from backup
secretvault import vault-backup-20240207.json -m overwrite
```

### Team Secret Sharing (Metadata Only)

```bash
# Export structure for team review (no values)
secretvault export -f yaml -o team-secrets-structure.yaml

# Share via Git (safe - no values)
git add team-secrets-structure.yaml
git commit -m "Add secrets structure documentation"
```

### Migration Between Machines

```bash
# On source machine
secretvault export --include-values -o vault-export.json

# Transfer file securely (e.g., via SCP, encrypted USB)
scp vault-export.json user@newmachine:/tmp/

# On destination machine
secretvault import /tmp/vault-export.json -m overwrite
rm /tmp/vault-export.json  # Clean up
```

### CI/CD Integration

```bash
# Extract specific secrets for CI/CD
secretvault export | jq -r '.secrets[] | select(.category == "ci") | "\(.name)=\(.value)"' > .env
```

## Security Considerations

1. **Default Export Excludes Values**: Metadata-only export is safe for sharing structure
2. **Explicit Confirmation Required**: Exporting values requires typing "yes" to confirm
3. **Secure File Permissions**: Export files are created with 0600 permissions (owner read/write only)
4. **Encrypted Storage**: The main vault remains encrypted; exports are plain text
5. **Clean Up Exports**: Delete export files after use if they contain values
6. **Merge Mode Safety**: Automatically keeps newer versions to prevent data loss

## Tips

- Use `--include-values` with caution - only when necessary for backups/migration
- Pipe to `jq` or `yq` for powerful filtering and transformation
- Use merge mode when syncing secrets across environments
- Export metadata to version control for documentation purposes
- Always secure transport of files containing secret values (encryption, secure channels)
