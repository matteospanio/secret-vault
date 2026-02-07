# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Secret Vault is a secure command-line application for storing and managing API tokens, secrets, and sensitive credentials with AES-256-GCM encryption. The project features both a traditional CLI interface and an interactive Terminal User Interface (TUI) built with Bubble Tea.

## Development Commands

### Building
```bash
go build -o secretvault ./cmd/secretvault
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/vault/...
go test ./pkg/tui/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests verbosely
go test -v ./...

# Run a single test
go test -v ./pkg/vault -run TestVaultFilterByCategory
```

### Running
```bash
# Run TUI directly
go run ./cmd/secretvault tui

# Run after building
./secretvault tui
```

## Architecture Overview

### Core Data Flow

The application has a layered architecture:

1. **Vault Layer** (`pkg/vault`): Core data model and business logic
   - `Secret` struct with fields: Name, Value, Description, Category, Tags, CreatedAt, UpdatedAt
   - `Vault` struct manages a map of secrets with version tracking
   - Filtering methods: `FilterByCategory()`, `FilterByTag()`, `FilterByAge()`, `Search()`
   - Age detection: `IsOld()` checks if secret > 1 year old (configurable threshold)

2. **Storage Layer** (`pkg/vault/storage.go`): Persistence with encryption
   - Encrypts entire vault JSON before writing to disk
   - Vault file created with 0600 permissions (owner read/write only)
   - Backward compatibility with `omitempty` JSON tags

3. **Crypto Layer** (`pkg/crypto`): AES-256-GCM encryption
   - PBKDF2 key derivation with SHA-256 (100,000 iterations)
   - Random 32-byte salt and 12-byte nonce per encryption
   - Authenticated encryption with GCM

4. **TUI Layer** (`pkg/tui`): Bubble Tea-based interactive interface
5. **CLI Layer** (`cmd/secretvault/main.go`): Cobra commands

### TUI Architecture (Bubble Tea)

The TUI uses a state machine pattern with multiple views:

**Main Model** (`pkg/tui/model.go`):
- Central coordinator implementing `tea.Model` interface
- Manages view transitions and state
- `ViewType` enum: ViewList, ViewDetail, ViewEdit, ViewFilter, ViewHelp
- Maintains references to all sub-views

**View Components**:
- `ListView`: Browse secrets with navigation and search (uses `bubbles/list`)
- `DetailView`: Display secret with reveal/copy functionality
- `EditView`: Form for creating/editing secrets with categories and tags
- `FilterView`: Advanced multi-criteria filtering UI
- `SearchBar`: Real-time fuzzy search component

**Message Passing**:
- Messages defined in `pkg/tui/messages.go` (e.g., `SecretSelectedMsg`, `FilterAppliedMsg`)
- Views return `tea.Cmd` for async operations
- Model routes messages to appropriate views

**Key Patterns**:
- Views use `(tea.Cmd, bool)` return pattern where bool indicates if message was handled
- `Update()` delegates to current view, then handles view-level messages
- `View()` renders based on `currentView` state
- Help footers show context-sensitive keyboard shortcuts

### Keyboard Shortcuts System

Shortcuts are standardized across views:
- **Global**: `?` (help), `q` (quit), `esc` (back/cancel)
- **Navigation**: `↑/↓`, `enter`, `tab/shift+tab`
- **Actions**: View-specific (e.g., `c` copy, `r` reveal, `f` filter)

Each view has a help footer showing available shortcuts. Press `?` for context-sensitive help.

## Testing Philosophy

**Pragmatic TDD approach**:
- Write tests FIRST for business logic (vault, crypto, algorithms)
- Tests can FOLLOW for UI components (views, rendering)

**Coverage Targets**:
- `pkg/vault`: >90% (currently 92.1%)
- `pkg/crypto`: >95% (currently 82.9%)
- `pkg/tui`: >70% (currently 89.2%)
- `pkg/clipboard`: >80% (currently 100%)

**Test Patterns**:
- Use table-driven tests for comprehensive coverage
- TUI tests focus on state transitions and message handling
- Mock vault data for TUI component tests

## Security Considerations

- **Never log or print decrypted vault contents in tests or debug code**
- Vault file permissions are enforced at 0600
- Master password is never stored
- All encryption uses authenticated encryption (GCM)
- Use `omitempty` for backward compatibility when adding fields

## Key Dependencies

- **Bubble Tea**: TUI framework (Elm-inspired architecture)
- **Bubbles**: Pre-built TUI components (list, textinput)
- **Lipgloss**: Terminal styling library
- **Cobra**: CLI framework
- **Viper**: Configuration management
- **atotto/clipboard**: Cross-platform clipboard operations

## Adding New Features

### Adding a TUI View
1. Create `pkg/tui/myview.go` with struct implementing view logic
2. Add `Update(msg tea.Msg) (tea.Cmd, bool)` for message handling
3. Add `View() string` for rendering
4. Define message types in `pkg/tui/messages.go`
5. Integrate with `Model` in `pkg/tui/model.go`
6. Update help view in `renderHelp()` with new shortcuts
7. Create `pkg/tui/myview_test.go` with state transition tests

### Modifying Vault Data Model
1. Update `Secret` struct in `pkg/vault/vault.go`
2. Use `omitempty` JSON tags for backward compatibility
3. Add migration logic if needed
4. Write tests verifying old vaults still load
5. Update storage tests for serialization/deserialization

## Common Pitfalls

- **Don't break backward compatibility**: Old vault files must load without errors
- **TUI state management**: Views must not modify vault directly - use messages
- **Persistence in TUI**: The vault is saved to disk when TUI exits (in `handleTUI()`), not on every change. The `finalModel` is captured after `p.Run()` and `SaveVault()` is called to persist changes.
- **Clipboard in headless environments**: Clipboard operations gracefully fail in headless systems
- **Test isolation**: Reset vault state between tests to avoid cross-contamination
