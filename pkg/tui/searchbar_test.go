package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

func TestNewSearchBar(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	if sb == nil {
		t.Fatal("NewSearchBar should return non-nil")
	}
	if sb.IsActive() {
		t.Error("Search bar should not be active initially")
	}
	if sb.Query() != "" {
		t.Error("Search bar query should be empty initially")
	}
}

func TestSearchBarFocusBlur(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	sb.Focus()
	if !sb.IsActive() {
		t.Error("Search bar should be active after Focus()")
	}

	sb.Blur()
	if sb.IsActive() {
		t.Error("Search bar should not be active after Blur()")
	}
}

func TestSearchBarSetQuery(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	sb.SetQuery("test")
	if sb.Query() != "test" {
		t.Errorf("Query should be 'test', got %q", sb.Query())
	}
}

func TestSearchBarClear(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	sb.SetQuery("test")
	sb.Clear()
	if sb.Query() != "" {
		t.Errorf("Query should be empty after Clear(), got %q", sb.Query())
	}
}

func TestSearchBarUpdateInactive(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	// When inactive, Update should not handle messages
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	_, changed := sb.Update(msg)
	if changed {
		t.Error("Update should not report changes when inactive")
	}
}

func TestSearchBarUpdateActive(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)
	sb.Focus()

	// Type a character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	_, changed := sb.Update(msg)
	if !changed {
		t.Error("Update should report changes when typing")
	}
	if sb.Query() != "a" {
		t.Errorf("Query should be 'a', got %q", sb.Query())
	}
}

func TestSearchBarView(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	// Inactive, no query
	view := sb.View()
	if !contains(view, "Type to search") {
		t.Errorf("Inactive view should contain hint, got %q", view)
	}

	// Inactive with query
	sb.SetQuery("test")
	view = sb.View()
	if !contains(view, "test") {
		t.Errorf("View with query should contain the query, got %q", view)
	}

	// Active
	sb.Focus()
	view = sb.View()
	if view == "" {
		t.Error("Active view should not be empty")
	}
}

func TestSearchBarSetSize(t *testing.T) {
	v := vault.NewVault()
	sb := NewSearchBar(v, 80)

	sb.SetSize(120)
	if sb.width != 120 {
		t.Errorf("Width should be 120, got %d", sb.width)
	}
}

func TestSearchBarFilterSecretsByQuery(t *testing.T) {
	v := vault.NewVault()
	now := time.Now()

	secrets := []vault.Secret{
		{Name: "github-token", Value: "abc", Description: "GitHub API token", CreatedAt: now, UpdatedAt: now},
		{Name: "aws-key", Value: "def", Description: "AWS access key", CreatedAt: now, UpdatedAt: now},
		{Name: "gitlab-token", Value: "ghi", Description: "GitLab CI token", CreatedAt: now, UpdatedAt: now},
		{Name: "database-password", Value: "jkl", Description: "DB credentials", CreatedAt: now, UpdatedAt: now},
	}

	sb := NewSearchBar(v, 80)

	t.Run("empty query returns all", func(t *testing.T) {
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 4 {
			t.Errorf("Empty query should return all secrets, got %d", len(result))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		sb.SetQuery("token")
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 2 {
			t.Errorf("Should match 2 secrets with 'token', got %d", len(result))
		}
	})

	t.Run("filter by description", func(t *testing.T) {
		sb.SetQuery("AWS")
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 1 {
			t.Errorf("Should match 1 secret with 'AWS', got %d", len(result))
		}
		if result[0].Name != "aws-key" {
			t.Errorf("Should match 'aws-key', got %q", result[0].Name)
		}
	})

	t.Run("fuzzy match", func(t *testing.T) {
		sb.SetQuery("ghtoken")
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 1 {
			t.Errorf("Should fuzzy match 1 secret with 'ghtoken', got %d", len(result))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		sb.SetQuery("zzzzz")
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 0 {
			t.Errorf("Should match 0 secrets, got %d", len(result))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		sb.SetQuery("GITHUB")
		result := sb.FilterSecretsByQuery(secrets)
		if len(result) != 1 {
			t.Errorf("Should match 1 secret case-insensitively, got %d", len(result))
		}
	})
}
