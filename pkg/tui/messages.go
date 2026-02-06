package tui

import "github.com/matteospanio/secret-vault-cli/pkg/vault"

// Message types for Bubble Tea

// SecretSelectedMsg is sent when a secret is selected from the list
type SecretSelectedMsg struct {
	Secret *vault.Secret
}

// SecretCopiedMsg is sent when a secret value is copied to clipboard
type SecretCopiedMsg struct {
	Name string
}

// ClipboardClearedMsg is sent when the clipboard is cleared
type ClipboardClearedMsg struct{}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// ViewChangeMsg is sent when the view should change
type ViewChangeMsg struct {
	View ViewType
}

// FilterAppliedMsg is sent when filters are applied
type FilterAppliedMsg struct {
	Category string
	Tag      string
	Query    string
	OldOnly  bool
}

// FilterClearedMsg is sent when filters are cleared
type FilterClearedMsg struct{}
