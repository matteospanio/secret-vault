package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

func TestSetViewAndGoBack(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initial view is ViewList
	if m.GetCurrentView() != ViewList {
		t.Errorf("Initial view should be ViewList, got %v", m.GetCurrentView())
	}

	// Set view to ViewDetail
	m.SetView(ViewDetail)
	if m.GetCurrentView() != ViewDetail {
		t.Errorf("View should be ViewDetail after SetView, got %v", m.GetCurrentView())
	}

	// Go back should return to ViewList
	m.GoBack()
	if m.GetCurrentView() != ViewList {
		t.Errorf("View should be ViewList after GoBack, got %v", m.GetCurrentView())
	}
}

func TestStatusMessages(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initially no status
	if m.GetStatusMessage() != "" {
		t.Error("Initial status should be empty")
	}
	if m.IsStatusError() {
		t.Error("Initial status should not be error")
	}

	// Set success status
	m.SetStatus("Success!", false)
	if m.GetStatusMessage() != "Success!" {
		t.Errorf("Status message should be 'Success!', got %q", m.GetStatusMessage())
	}
	if m.IsStatusError() {
		t.Error("Status should not be error")
	}

	// Set error status
	m.SetStatus("Error occurred", true)
	if m.GetStatusMessage() != "Error occurred" {
		t.Errorf("Status message should be 'Error occurred', got %q", m.GetStatusMessage())
	}
	if !m.IsStatusError() {
		t.Error("Status should be error")
	}

	// Clear status
	m.ClearStatus()
	if m.GetStatusMessage() != "" {
		t.Error("Status should be empty after clear")
	}
	if m.IsStatusError() {
		t.Error("Status should not be error after clear")
	}
}

func TestSelectSecret(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initially no secret selected
	if m.GetSelectedSecret() != nil {
		t.Error("Initially no secret should be selected")
	}

	// Select a secret
	secret := &vault.Secret{Name: "test-secret", Value: "secret-value"}
	m.SelectSecret(secret)

	if m.GetSelectedSecret() != secret {
		t.Error("GetSelectedSecret should return the selected secret")
	}

	if m.GetSelectedSecret().Name != "test-secret" {
		t.Errorf("Selected secret name should be 'test-secret', got %q", m.GetSelectedSecret().Name)
	}
}

func TestDimensionGetters(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initially 0
	if m.GetWidth() != 0 {
		t.Errorf("Initial width should be 0, got %d", m.GetWidth())
	}
	if m.GetHeight() != 0 {
		t.Errorf("Initial height should be 0, got %d", m.GetHeight())
	}

	// After window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.GetWidth() != 120 {
		t.Errorf("Width should be 120, got %d", m.GetWidth())
	}
	if m.GetHeight() != 40 {
		t.Errorf("Height should be 40, got %d", m.GetHeight())
	}
}

func TestUpdateHelpToggle(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press ? to open help
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	newModel, _ = m.Update(helpMsg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewHelp {
		t.Errorf("View should be ViewHelp after pressing ?, got %v", m.GetCurrentView())
	}

	// Press ? again to close help
	newModel, _ = m.Update(helpMsg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewList {
		t.Errorf("View should be ViewList after pressing ? again, got %v", m.GetCurrentView())
	}
}

func TestUpdateEscGoBack(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Go to help view first
	m.SetView(ViewHelp)

	// Press esc to go back
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewList {
		t.Errorf("View should be ViewList after pressing esc, got %v", m.GetCurrentView())
	}
}

func TestUpdateEscDoesNothingOnList(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// On ViewList, esc should do nothing
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewList {
		t.Errorf("View should still be ViewList, got %v", m.GetCurrentView())
	}
}

func TestUpdateViewChangeMsg(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Send ViewChangeMsg
	msg := ViewChangeMsg{View: ViewFilter}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewFilter {
		t.Errorf("View should be ViewFilter, got %v", m.GetCurrentView())
	}
}

func TestUpdateSecretSelectedMsg(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	secret := &vault.Secret{Name: "my-secret"}
	msg := SecretSelectedMsg{Secret: secret}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewDetail {
		t.Errorf("View should be ViewDetail after selecting secret, got %v", m.GetCurrentView())
	}
	if m.GetSelectedSecret() != secret {
		t.Error("Selected secret should be set")
	}
}

func TestUpdateErrorMsg(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	err := fmt.Errorf("test error")
	msg := ErrorMsg{Err: err}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.GetStatusMessage() != "test error" {
		t.Errorf("Status message should be 'test error', got %q", m.GetStatusMessage())
	}
	if !m.IsStatusError() {
		t.Error("Status should be error")
	}
}

func TestUpdateSecretCopiedMsg(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	msg := SecretCopiedMsg{Name: "my-secret"}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if !contains(m.GetStatusMessage(), "my-secret") {
		t.Errorf("Status message should contain secret name, got %q", m.GetStatusMessage())
	}
	if !contains(m.GetStatusMessage(), "clipboard") {
		t.Errorf("Status message should mention clipboard, got %q", m.GetStatusMessage())
	}
	if m.IsStatusError() {
		t.Error("Status should not be error")
	}
}

func TestUpdateClipboardClearedMsg(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	msg := ClipboardClearedMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if !contains(m.GetStatusMessage(), "Clipboard") {
		t.Errorf("Status message should mention clipboard, got %q", m.GetStatusMessage())
	}
	if m.IsStatusError() {
		t.Error("Status should not be error")
	}
}

func TestClearClipboardKeyHandler(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press 'C' to clear clipboard
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Error("'C' key should return a command for clearing clipboard")
	}
}

func TestRenderDetailNoSecret(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Set view to detail without selecting a secret
	m.SetView(ViewDetail)
	view := m.View()

	if !contains(view, "No secret selected") {
		t.Errorf("View should show 'No secret selected', got %q", view)
	}
}

func TestRenderDetailWithSecret(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "secret-value", "A test secret", "work", []string{"api", "test"})
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Select the secret
	secret, _ := v.GetSecret("test-secret")
	m.SelectSecret(&secret)
	m.SetView(ViewDetail)

	view := m.View()

	if !contains(view, "test-secret") {
		t.Errorf("View should contain secret name, got %q", view)
	}
	if !contains(view, "A test secret") {
		t.Errorf("View should contain description, got %q", view)
	}
	if !contains(view, "work") {
		t.Errorf("View should contain category, got %q", view)
	}
}

func TestRenderHelp(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	m.SetView(ViewHelp)
	view := m.View()

	if !contains(view, "Keyboard Shortcuts") {
		t.Errorf("Help view should contain 'Keyboard Shortcuts', got %q", view)
	}
	if !contains(view, "Quit") {
		t.Errorf("Help view should mention Quit, got %q", view)
	}
}

func TestViewWithStatusMessage(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Make model ready
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Set a status message
	m.SetStatus("Operation successful", false)

	view := m.View()
	if !contains(view, "Operation successful") {
		t.Errorf("View should contain status message, got %q", view)
	}
}

func TestKeyPressClearsStatus(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Set a status message
	m.SetStatus("Some status", false)

	// Press any key (that doesn't quit)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Status should be cleared (we're checking before the view switch happens)
	// Actually the status is cleared then help view is shown
	// Let's press esc which will clear status but not do anything on list
	m.SetStatus("Some status", false)
	m.SetView(ViewList)
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.GetStatusMessage() != "" {
		t.Errorf("Status should be cleared on key press, got %q", m.GetStatusMessage())
	}
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
