package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// TestImports verifies that all TUI dependencies are correctly installed and importable.
func TestImports(t *testing.T) {
	// Verify bubbletea import
	_ = tea.Quit

	// Verify bubbles import
	_ = list.New(nil, list.NewDefaultDelegate(), 0, 0)

	// Verify lipgloss import
	_ = lipgloss.NewStyle()
}

func TestNewModel(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	if m.vault != v {
		t.Error("NewModel should store the vault reference")
	}

	if m.currentView != ViewList {
		t.Errorf("NewModel should start with ViewList, got %v", m.currentView)
	}

	if m.ready {
		t.Error("NewModel should not be ready initially")
	}
}

func TestModelInit(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModelUpdateQuit(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	tests := []struct {
		name string
		key  string
	}{
		{"q key", "q"},
		{"ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := m.Update(msg)
			if cmd == nil {
				t.Errorf("Update(%s) should return tea.Quit command", tt.key)
			}
		})
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	if m.ready {
		t.Error("Model should not be ready before receiving WindowSizeMsg")
	}

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.width != 80 {
		t.Errorf("Width should be 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Height should be 24, got %d", m.height)
	}

	if !m.ready {
		t.Error("Model should be ready after receiving WindowSizeMsg")
	}
}

func TestModelViewNotReady(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	view := m.View()
	if view != "Loading..." {
		t.Errorf("View() should return 'Loading...' when not ready, got %q", view)
	}
}

func TestModelViewWelcome(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Simulate window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() should return non-empty string when ready")
	}

	// Should contain title
	if !contains(view, "Secret Vault") {
		t.Error("View() should contain 'Secret Vault' title")
	}
}

func TestModelGetters(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	if m.GetVault() != v {
		t.Error("GetVault() should return the vault")
	}

	if m.GetCurrentView() != ViewList {
		t.Error("GetCurrentView() should return ViewList")
	}

	if m.IsReady() {
		t.Error("IsReady() should return false initially")
	}
}

func TestViewTypes(t *testing.T) {
	tests := []struct {
		view ViewType
		name string
	}{
		{ViewList, "ViewList"},
		{ViewDetail, "ViewDetail"},
		{ViewEdit, "ViewEdit"},
		{ViewFilter, "ViewFilter"},
		{ViewHelp, "ViewHelp"},
	}

	for i, tt := range tests {
		if int(tt.view) != i {
			t.Errorf("%s should have value %d, got %d", tt.name, i, tt.view)
		}
	}
}

func TestStylesExist(t *testing.T) {
	// Verify styles are defined and usable
	_ = titleStyle.Render("test")
	_ = subtitleStyle.Render("test")
	_ = dimStyle.Render("test")
	_ = helpStyle.Render("test")
	_ = warningStyle.Render("test")
	_ = errorStyle.Render("test")
	_ = successStyle.Render("test")
	_ = listItemStyle.Render("test")
	_ = listSelectedStyle.Render("test")
	_ = headerStyle.Render("test")
	_ = footerStyle.Render("test")
}

func TestMessagesExist(t *testing.T) {
	// Verify message types can be instantiated
	secret := &vault.Secret{Name: "test"}
	_ = SecretSelectedMsg{Secret: secret}
	_ = SecretCopiedMsg{Name: "test"}
	_ = ClipboardClearedMsg{}
	_ = ErrorMsg{Err: nil}
	_ = ViewChangeMsg{View: ViewList}
	_ = FilterAppliedMsg{Category: "work", Tag: "api", Query: "test"}
	_ = FilterClearedMsg{}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
