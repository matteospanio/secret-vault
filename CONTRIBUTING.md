# Contributing to Secret Vault CLI

Thank you for your interest in contributing to Secret Vault CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Testing Philosophy](#testing-philosophy)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Adding New Features](#adding-new-features)

## Getting Started

### Prerequisites

- **Go 1.21 or higher**
- **Git**
- **A terminal that supports 256 colors** (for TUI development)
- **Clipboard utilities** (Linux: `xclip` or `xsel`)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/matteospanio/secret-vault.git
   cd secret-vault
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/matteospanio/secret-vault.git
   ```

## Development Setup

### Install Dependencies

```bash
go mod download
```

### Build the Project

```bash
go build -o secretvault ./cmd/secretvault
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./pkg/vault/...
go test ./pkg/tui/...

# Run tests verbosely
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run the Application

```bash
# After building
./secretvault tui

# Or run directly without building
go run ./cmd/secretvault tui
```

## Project Structure

```
secret-vault/
├── cmd/
│   └── secretvault/          # Main CLI application entry point
│       └── main.go
├── pkg/
│   ├── clipboard/            # Cross-platform clipboard operations
│   ├── crypto/               # Encryption/decryption logic
│   ├── git/                  # Git workflow integration
│   ├── sync/                 # Cloud sync functionality
│   ├── tui/                  # Terminal User Interface (Bubble Tea)
│   └── vault/                # Vault data model and storage
├── plan/                     # Project planning documents
│   └── project.md
├── go.mod
├── go.sum
├── README.md
└── CONTRIBUTING.md
```

### Key Packages

- **`pkg/vault`**: Core data model, secret storage, filtering, and querying
- **`pkg/crypto`**: AES-256-GCM encryption/decryption with PBKDF2 key derivation
- **`pkg/tui`**: Bubble Tea-based terminal user interface with multiple views
- **`pkg/clipboard`**: Cross-platform clipboard integration
- **`pkg/sync`**: Cloud sync with provider abstraction
- **`pkg/git`**: Git hooks and workflow integration

## Testing Philosophy

This project follows a **pragmatic Test-Driven Development (TDD)** approach:

### When to Write Tests First

✅ **Business logic** (vault operations, encryption, filtering)
✅ **Data models** (secret structure, vault structure)
✅ **Core algorithms** (fuzzy matching, filtering, sorting)
✅ **API boundaries** (external integrations)

### When Tests Can Follow Implementation

✅ **UI components** (TUI views and rendering)
✅ **Simple wrappers** (thin wrappers around libraries)
✅ **Integration code** (glue code between components)

### Coverage Targets

- **Overall**: >75%
- **`pkg/vault`**: >90%
- **`pkg/crypto`**: >95%
- **`pkg/clipboard`**: >80%
- **`pkg/tui`**: >70%

### Test Structure

Use table-driven tests for comprehensive coverage:

```go
func TestVaultFilterByCategory(t *testing.T) {
    tests := []struct {
        name     string
        vault    *Vault
        category string
        want     int // expected number of results
    }{
        {"empty vault", NewVault(), "work", 0},
        {"matching category", vaultWithSecrets(), "work", 2},
        {"no match", vaultWithSecrets(), "nonexistent", 0},
        {"case insensitive", vaultWithSecrets(), "WORK", 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.vault.FilterByCategory(tt.category)
            if len(got) != tt.want {
                t.Errorf("got %v secrets, want %v", len(got), tt.want)
            }
        })
    }
}
```

### Testing TUI Components

For TUI components, focus on:
- State transitions
- Message handling
- View switching logic
- Keyboard shortcut handling

```go
func TestModelViewTransition(t *testing.T) {
    m := NewModel(vault.NewVault())

    // Test transition from List to Help
    m.SetView(ViewHelp)
    if m.GetCurrentView() != ViewHelp {
        t.Error("View should transition to Help")
    }

    // Test back navigation
    m.GoBack()
    if m.GetCurrentView() != ViewList {
        t.Error("Should return to previous view")
    }
}
```

## Code Style

### General Guidelines

- Follow standard Go conventions
- Use `gofmt` to format code
- Run `go vet` to catch common issues
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `vault`, `crypto`, `tui`)
- **Files**: lowercase with underscores for tests (e.g., `vault.go`, `vault_test.go`)
- **Types**: PascalCase (e.g., `SecretVault`, `FilterCriteria`)
- **Functions**: camelCase for private, PascalCase for exported (e.g., `filterByAge`, `FilterByCategory`)
- **Constants**: PascalCase or ALL_CAPS for package-level (e.g., `DefaultAgeThreshold`)

### Error Handling

Always handle errors explicitly:

```go
// ❌ Bad
result, _ := vault.GetSecret("name")

// ✅ Good
result, err := vault.GetSecret("name")
if err != nil {
    return fmt.Errorf("failed to get secret: %w", err)
}
```

### Comments

- Add package-level comments for all packages
- Document all exported functions, types, and constants
- Use GoDoc format for documentation

```go
// FilterByCategory returns all secrets matching the given category.
// The match is case-insensitive. Returns an empty slice if no matches found.
func (v *Vault) FilterByCategory(category string) []Secret {
    // implementation
}
```

## Pull Request Process

### Before Submitting

1. **Run tests**: `go test ./...`
2. **Check coverage**: `go test -cover ./...`
3. **Format code**: `gofmt -w .`
4. **Run linter**: `go vet ./...`
5. **Test manually**: Build and test the feature in the TUI

### PR Guidelines

1. **Create a focused PR**: One feature or fix per PR
2. **Write a clear description**: Explain what and why, not just how
3. **Reference issues**: Link to related issues with "Fixes #123"
4. **Add tests**: Include tests for new functionality
5. **Update docs**: Update README.md if adding user-facing features
6. **Keep commits clean**: Squash WIP commits before merging

### PR Title Format

Use conventional commit format:

- `feat: Add category filtering to list command`
- `fix: Prevent clipboard crash on headless systems`
- `docs: Update TUI keyboard shortcuts in README`
- `test: Add tests for fuzzy matching algorithm`
- `refactor: Simplify filter view state management`

### PR Description Template

```markdown
## Description
Brief description of the change.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests you ran and their results.

## Checklist
- [ ] My code follows the project's code style
- [ ] I have added tests that prove my fix/feature works
- [ ] I have updated the documentation (if applicable)
- [ ] All tests pass locally
- [ ] I have run `go vet` and `gofmt`
```

## Adding New Features

### TUI Features

When adding a new TUI view or component:

1. **Create the view file**: `pkg/tui/myview.go`
2. **Implement Bubble Tea interface**:
   - `Update(msg tea.Msg)` for handling messages
   - `View()` for rendering
3. **Add message types**: Define in `pkg/tui/messages.go`
4. **Integrate with model**: Update `pkg/tui/model.go` to handle view
5. **Add tests**: Create `pkg/tui/myview_test.go`
6. **Update help view**: Add shortcuts to `renderHelp()` in `model.go`

### CLI Features

When adding a new CLI command:

1. **Add command**: Update `cmd/secretvault/main.go`
2. **Define flags**: Use Cobra's flag system
3. **Implement handler**: Create handler function
4. **Add help text**: Document the command
5. **Test manually**: Verify command works as expected
6. **Update README**: Document the new command

### Data Model Changes

When modifying the vault structure:

1. **Update `pkg/vault/vault.go`**: Modify structs
2. **Use `omitempty`**: For backward compatibility
3. **Add migration logic**: If needed for existing vaults
4. **Write tests**: Verify old vaults still load
5. **Update storage tests**: Test serialization/deserialization

## Development Workflow

### Typical Feature Development Flow

1. **Check the plan**: Review `plan/project.md` for context
2. **Create a branch**: `git checkout -b feat/my-feature`
3. **Write tests**: Start with tests for business logic
4. **Implement feature**: Write minimal code to pass tests
5. **Add UI tests**: Test UI components if applicable
6. **Manual testing**: Test in the actual TUI
7. **Update docs**: Update README.md and help view
8. **Run full test suite**: `go test ./...`
9. **Commit changes**: Use conventional commit format
10. **Push and create PR**: Push to your fork and open a PR

### Debugging Tips

**Debug TUI rendering:**
```go
// Add to any view's View() method
fmt.Fprintf(os.Stderr, "Debug: %+v\n", someVariable)
```

**Test TUI without running:**
```go
// In tests, create model and send messages
m := NewModel(vault)
m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
```

**Profile performance:**
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/tui/...
go tool pprof cpu.prof
```

## Questions?

- **Open an issue**: For bugs, feature requests, or questions
- **Check existing issues**: Someone may have already asked
- **Read the code**: The codebase is well-documented

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a positive community

Thank you for contributing to Secret Vault CLI! 🎉
