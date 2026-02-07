package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault/pkg/vault"
)

// Helper to create a test secret
func createTestSecret(name string, daysOld int) *vault.Secret {
	createdAt := time.Now().Add(-time.Duration(daysOld) * 24 * time.Hour)
	return &vault.Secret{
		Name:        name,
		Value:       "secret-value-123",
		Description: "Test description",
		Category:    "test-category",
		Tags:        []string{"tag1", "tag2"},
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}

// Helper to create a minimal test secret
func createMinimalSecret(name string) *vault.Secret {
	now := time.Now()
	return &vault.Secret{
		Name:      name,
		Value:     "minimal-value",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// --- Construction Tests ---

func TestNewDetailView(t *testing.T) {
	secret := createTestSecret("test-secret", 30)
	dv := NewDetailView(secret, 80, 24)

	if dv == nil {
		t.Fatal("NewDetailView returned nil")
	}
	if dv.secret != secret {
		t.Error("DetailView should store the provided secret")
	}
	if dv.width != 80 {
		t.Errorf("width = %d, want 80", dv.width)
	}
	if dv.height != 24 {
		t.Errorf("height = %d, want 24", dv.height)
	}
	if dv.revealed {
		t.Error("revealed should be false by default")
	}
}

func TestNewDetailViewNilSecret(t *testing.T) {
	dv := NewDetailView(nil, 80, 24)

	if dv == nil {
		t.Fatal("NewDetailView should not return nil even with nil secret")
	}
	if dv.secret != nil {
		t.Error("secret should be nil")
	}
}

// --- State Management Tests ---

func TestDetailViewIsRevealed(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	if dv.IsRevealed() {
		t.Error("IsRevealed should return false initially")
	}
}

func TestDetailViewToggleReveal(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	dv.ToggleReveal()
	if !dv.IsRevealed() {
		t.Error("IsRevealed should return true after first toggle")
	}

	dv.ToggleReveal()
	if dv.IsRevealed() {
		t.Error("IsRevealed should return false after second toggle")
	}
}

func TestDetailViewSetSecret(t *testing.T) {
	secret1 := createTestSecret("secret1", 30)
	secret2 := createTestSecret("secret2", 60)
	dv := NewDetailView(secret1, 80, 24)

	dv.SetSecret(secret2)
	if dv.secret != secret2 {
		t.Error("SetSecret should update the secret")
	}
}

func TestDetailViewSetSecretResetsRevealed(t *testing.T) {
	secret1 := createTestSecret("secret1", 30)
	secret2 := createTestSecret("secret2", 60)
	dv := NewDetailView(secret1, 80, 24)

	// Reveal the value
	dv.ToggleReveal()
	if !dv.IsRevealed() {
		t.Fatal("revealed should be true after toggle")
	}

	// Set a new secret - should reset revealed state
	dv.SetSecret(secret2)
	if dv.IsRevealed() {
		t.Error("SetSecret should reset revealed to false")
	}
}

func TestDetailViewSetSize(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	dv.SetSize(120, 40)
	if dv.width != 120 {
		t.Errorf("width = %d, want 120", dv.width)
	}
	if dv.height != 40 {
		t.Errorf("height = %d, want 40", dv.height)
	}
}

func TestDetailViewGetSecret(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	if dv.GetSecret() != secret {
		t.Error("GetSecret should return the stored secret")
	}
}

// --- formatAge Helper Tests ---

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "less than a day"},
		{"half day", 12 * time.Hour, "less than a day"},
		{"one day", 25 * time.Hour, "1 day"},
		{"few days", 5 * 24 * time.Hour, "5 days"},
		{"almost month", 29 * 24 * time.Hour, "29 days"},
		{"one month", 32 * 24 * time.Hour, "1 month"},
		{"few months", 90 * 24 * time.Hour, "3 months"},
		{"almost year", 350 * 24 * time.Hour, "11 months"},
		{"one year", 380 * 24 * time.Hour, "1 year"},
		{"multiple years", 800 * 24 * time.Hour, "2 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.duration)
			if result != tt.expected {
				t.Errorf("formatAge(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// --- View Rendering Tests ---

func TestDetailViewViewNoSecret(t *testing.T) {
	dv := NewDetailView(nil, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "No secret selected") {
		t.Errorf("View with nil secret should show 'No secret selected', got %q", view)
	}
}

func TestDetailViewViewBasicFields(t *testing.T) {
	secret := createMinimalSecret("my-api-key")
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	// Check name is displayed
	if !strings.Contains(view, "my-api-key") {
		t.Error("View should contain the secret name")
	}

	// Check value is masked
	if !strings.Contains(view, "********") {
		t.Error("View should contain masked value")
	}

	// Check timestamps are displayed
	if !strings.Contains(view, "Created:") {
		t.Error("View should contain 'Created:' label")
	}
	if !strings.Contains(view, "Updated:") {
		t.Error("View should contain 'Updated:' label")
	}
}

func TestDetailViewViewWithDescription(t *testing.T) {
	secret := createTestSecret("test", 30)
	secret.Description = "This is a test description"
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "This is a test description") {
		t.Error("View should contain the description")
	}
}

func TestDetailViewViewWithoutDescription(t *testing.T) {
	secret := createMinimalSecret("test")
	secret.Description = ""
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	// View should not have double empty lines where description would be
	// This is a soft check - we're mainly testing it doesn't crash
	if secret.Description != "" {
		t.Error("Test setup error - description should be empty")
	}
	_ = view // Just ensure it renders
}

func TestDetailViewViewWithCategory(t *testing.T) {
	secret := createTestSecret("test", 30)
	secret.Category = "work-secrets"
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "Category:") {
		t.Error("View should contain 'Category:' label")
	}
	if !strings.Contains(view, "work-secrets") {
		t.Error("View should contain the category value")
	}
}

func TestDetailViewViewWithoutCategory(t *testing.T) {
	secret := createMinimalSecret("test")
	secret.Category = ""
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if strings.Contains(view, "Category:") {
		t.Error("View should not contain 'Category:' label when category is empty")
	}
}

func TestDetailViewViewWithTags(t *testing.T) {
	secret := createTestSecret("test", 30)
	secret.Tags = []string{"github", "api", "production"}
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "Tags:") {
		t.Error("View should contain 'Tags:' label")
	}
	if !strings.Contains(view, "github") {
		t.Error("View should contain tag 'github'")
	}
	if !strings.Contains(view, "api") {
		t.Error("View should contain tag 'api'")
	}
	if !strings.Contains(view, "production") {
		t.Error("View should contain tag 'production'")
	}
}

func TestDetailViewViewWithoutTags(t *testing.T) {
	secret := createMinimalSecret("test")
	secret.Tags = nil
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if strings.Contains(view, "Tags:") {
		t.Error("View should not contain 'Tags:' label when tags is nil/empty")
	}
}

func TestDetailViewViewMaskedValue(t *testing.T) {
	secret := createTestSecret("test", 30)
	secret.Value = "super-secret-value"
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if strings.Contains(view, "super-secret-value") {
		t.Error("View should NOT contain the actual secret value when masked")
	}
	if !strings.Contains(view, "********") {
		t.Error("View should contain masked value '********'")
	}
}

func TestDetailViewViewRevealedValue(t *testing.T) {
	secret := createTestSecret("test", 30)
	secret.Value = "super-secret-value"
	dv := NewDetailView(secret, 80, 24)
	dv.ToggleReveal()
	view := dv.View()

	if !strings.Contains(view, "super-secret-value") {
		t.Error("View should contain the actual secret value when revealed")
	}
}

func TestDetailViewViewOldSecretWarning(t *testing.T) {
	// Create secret older than 1 year (400 days)
	secret := createTestSecret("old-secret", 400)
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "⚠") {
		t.Error("View should contain warning symbol for old secret")
	}
	if !strings.Contains(view, "1 year old") {
		t.Errorf("View should mention age warning, got: %s", view)
	}
}

func TestDetailViewViewNewSecretNoWarning(t *testing.T) {
	// Create recent secret (30 days old)
	secret := createTestSecret("new-secret", 30)
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if strings.Contains(view, "⚠") {
		t.Error("View should NOT contain warning symbol for new secret")
	}
}

func TestDetailViewViewAgeDisplay(t *testing.T) {
	secret := createTestSecret("test", 90) // 3 months old
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "Age:") {
		t.Error("View should contain 'Age:' label")
	}
	if !strings.Contains(view, "3 months") {
		t.Errorf("View should show age as '3 months', got: %s", view)
	}
}

func TestDetailViewViewHelpFooter(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)
	view := dv.View()

	if !strings.Contains(view, "r: reveal") {
		t.Error("View should contain 'r: reveal' in help footer")
	}
	if !strings.Contains(view, "c: copy") {
		t.Error("View should contain 'c: copy' in help footer")
	}
	if !strings.Contains(view, "esc: back") {
		t.Error("View should contain 'esc: back' in help footer")
	}
}

func TestDetailViewViewHelpFooterRevealed(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)
	dv.ToggleReveal()
	view := dv.View()

	if strings.Contains(view, "r: reveal") {
		t.Error("View should NOT contain 'r: reveal' when value is revealed")
	}
	if !strings.Contains(view, "r: hide") {
		t.Error("View should contain 'r: hide' when value is revealed")
	}
}

// --- Update/Keyboard Handling Tests ---

func TestDetailViewUpdateRevealKey(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	cmd, handled := dv.Update(msg)

	if !handled {
		t.Error("'r' key should be handled by DetailView")
	}
	if cmd != nil {
		t.Error("'r' key should not return a command (state change only)")
	}
	if !dv.IsRevealed() {
		t.Error("'r' key should toggle revealed state to true")
	}

	// Toggle again
	cmd, handled = dv.Update(msg)
	if !handled {
		t.Error("'r' key should still be handled")
	}
	if dv.IsRevealed() {
		t.Error("second 'r' key should toggle revealed state to false")
	}
}

func TestDetailViewUpdateCopyKey(t *testing.T) {
	secret := createTestSecret("test-secret", 30)
	dv := NewDetailView(secret, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	cmd, handled := dv.Update(msg)

	if !handled {
		t.Error("'c' key should be handled by DetailView")
	}
	if cmd == nil {
		t.Error("'c' key should return a command for clipboard operation")
	}
}

func TestDetailViewUpdateCopyKeyNilSecret(t *testing.T) {
	dv := NewDetailView(nil, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	cmd, handled := dv.Update(msg)

	if !handled {
		t.Error("'c' key should be handled even with nil secret")
	}
	if cmd == nil {
		t.Error("'c' key should return a command (error command)")
	}

	// Execute the command and check for error message
	result := cmd()
	if _, ok := result.(ErrorMsg); !ok {
		t.Errorf("command should return ErrorMsg, got %T", result)
	}
}

func TestDetailViewUpdateEditKey(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
	cmd, handled := dv.Update(msg)

	if !handled {
		t.Error("'e' key should be handled by DetailView")
	}
	if cmd == nil {
		t.Error("'e' key should return a command")
	}

	// Execute the command and check it returns ViewChangeMsg
	result := cmd()
	viewMsg, ok := result.(ViewChangeMsg)
	if !ok {
		t.Errorf("command should return ViewChangeMsg, got %T", result)
	}
	if viewMsg.View != ViewEdit {
		t.Errorf("ViewChangeMsg.View = %v, want ViewEdit", viewMsg.View)
	}
}

func TestDetailViewUpdateUnhandledKey(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	// Test various keys that should NOT be handled by DetailView
	unhandledKeys := []string{"q", "esc", "?", "x", "enter"}

	for _, key := range unhandledKeys {
		t.Run(key, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			_, handled := dv.Update(msg)

			if handled {
				t.Errorf("'%s' key should NOT be handled by DetailView", key)
			}
		})
	}
}

func TestDetailViewUpdateNonKeyMsg(t *testing.T) {
	secret := createTestSecret("test", 30)
	dv := NewDetailView(secret, 80, 24)

	// Test with WindowSizeMsg - should not be handled
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_, handled := dv.Update(msg)

	if handled {
		t.Error("WindowSizeMsg should not be handled by DetailView")
	}
}

// --- Integration Tests ---

func TestModelDetailViewIntegration(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "test-value", "desc", "cat", []string{"tag"})
	m := NewModel(v)

	// Simulate window size message
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := m.Update(windowMsg)
	m = updated.(Model)

	// Check detailView was created
	if m.GetDetailView() == nil {
		t.Error("DetailView should be created after WindowSizeMsg")
	}
}

func TestModelSelectSecretUpdatesDetailView(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "test-value", "desc", "cat", []string{"tag"})
	m := NewModel(v)

	// Initialize with window size
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := m.Update(windowMsg)
	m = updated.(Model)

	// Get the secret and select it
	secret, _ := v.GetSecret("test-secret")
	m.SelectSecret(&secret)

	// Check detail view has the secret
	dv := m.GetDetailView()
	if dv == nil {
		t.Fatal("DetailView should exist")
	}
	if dv.GetSecret() == nil {
		t.Error("DetailView should have the selected secret")
	}
	if dv.GetSecret().Name != "test-secret" {
		t.Errorf("DetailView secret name = %q, want 'test-secret'", dv.GetSecret().Name)
	}
}

func TestModelDetailViewKeyHandling(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "test-value", "desc", "cat", []string{"tag"})
	m := NewModel(v)

	// Initialize with window size
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := m.Update(windowMsg)
	m = updated.(Model)

	// Select a secret and switch to detail view
	secret, _ := v.GetSecret("test-secret")
	m.SelectSecret(&secret)
	m.SetView(ViewDetail)

	// Press 'r' to reveal
	revealMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updated, _ = m.Update(revealMsg)
	m = updated.(Model)

	// Check that the detail view's revealed state changed
	dv := m.GetDetailView()
	if dv == nil {
		t.Fatal("DetailView should exist")
	}
	if !dv.IsRevealed() {
		t.Error("DetailView should be revealed after 'r' key in detail view")
	}
}
