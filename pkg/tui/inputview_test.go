package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// Helper to create a vault with test secrets
func createTestVault() *vault.Vault {
	v := vault.NewVault()
	v.AddSecret("github-token", "ghp_123", "GitHub personal access token", "work", []string{"api"})
	v.AddSecret("gitlab-api-key", "glpat_456", "GitLab API key", "work", []string{"api"})
	v.AddSecret("aws-secret", "AKIA123", "AWS secret key", "cloud", []string{"aws"})
	v.AddSecret("azure-key", "azure_789", "Azure key", "cloud", []string{"azure"})
	v.AddSecret("personal-github", "ghp_personal", "Personal GitHub", "personal", nil)
	return v
}

// --- Construction Tests ---

func TestNewInputView(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "placeholder", "prompt")

	if iv == nil {
		t.Fatal("NewInputView should not return nil")
	}
	if iv.vault != v {
		t.Error("InputView should store the vault")
	}
	if iv.GetPlaceholder() != "placeholder" {
		t.Errorf("placeholder = %q, want 'placeholder'", iv.GetPlaceholder())
	}
	if iv.GetPrompt() != "prompt" {
		t.Errorf("prompt = %q, want 'prompt'", iv.GetPrompt())
	}
}

func TestNewInputViewNilVault(t *testing.T) {
	iv := NewInputView(nil, "placeholder", "prompt")

	if iv == nil {
		t.Fatal("NewInputView should not return nil even with nil vault")
	}
	if len(iv.Suggestions()) != 0 {
		t.Error("Suggestions should be empty with nil vault")
	}
}

func TestNewInputViewInitialSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	// Should have suggestions from vault (up to maxSuggestions)
	if iv.SuggestionCount() == 0 {
		t.Error("Should have initial suggestions from vault")
	}
	if iv.SuggestionCount() > iv.maxSuggestions {
		t.Errorf("Suggestions count %d exceeds max %d", iv.SuggestionCount(), iv.maxSuggestions)
	}
}

// --- State Management Tests ---

func TestInputViewValue(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if iv.Value() != "" {
		t.Error("Initial value should be empty")
	}

	iv.SetValue("test")
	if iv.Value() != "test" {
		t.Errorf("Value() = %q, want 'test'", iv.Value())
	}
}

func TestInputViewSetValueUpdatesSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	initialCount := iv.SuggestionCount()

	iv.SetValue("github")
	newCount := iv.SuggestionCount()

	// Should have filtered suggestions
	if newCount >= initialCount && initialCount > 2 {
		t.Error("Setting value should filter suggestions")
	}
}

func TestInputViewReset(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetValue("test")
	iv.Reset()

	if iv.Value() != "" {
		t.Error("Reset should clear input value")
	}
	if iv.SelectedIndex() != 0 {
		t.Error("Reset should reset selected index")
	}
}

func TestInputViewIsEmpty(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.IsEmpty() {
		t.Error("New input should be empty")
	}

	iv.SetValue("test")
	if iv.IsEmpty() {
		t.Error("Input with value should not be empty")
	}
}

func TestInputViewSetSize(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetSize(120, 40)
	if iv.width != 120 {
		t.Errorf("width = %d, want 120", iv.width)
	}
	if iv.height != 40 {
		t.Errorf("height = %d, want 40", iv.height)
	}
}

func TestInputViewMaxSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetMaxSuggestions(3)
	if iv.SuggestionCount() > 3 {
		t.Errorf("Suggestions count %d exceeds max 3", iv.SuggestionCount())
	}

	iv.SetMaxSuggestions(0) // Should be clamped to 1
	if iv.maxSuggestions != 1 {
		t.Errorf("maxSuggestions should be clamped to 1, got %d", iv.maxSuggestions)
	}
}

func TestInputViewShowSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetShowSuggestions(false)
	view := iv.View()

	// Suggestions should not appear in view
	if strings.Contains(view, "github-token") {
		t.Error("View should not show suggestions when disabled")
	}
}

// --- Suggestion Selection Tests ---

func TestInputViewSelectedIndex(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if iv.SelectedIndex() != 0 {
		t.Error("Initial selected index should be 0")
	}
}

func TestInputViewSelectedSuggestion(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if iv.HasSuggestions() {
		selected := iv.SelectedSuggestion()
		if selected == "" {
			t.Error("SelectedSuggestion should return first suggestion")
		}
		if selected != iv.Suggestions()[0] {
			t.Error("SelectedSuggestion should match first suggestion")
		}
	}
}

func TestInputViewSelectedSuggestionEmpty(t *testing.T) {
	iv := NewInputView(nil, "", "")

	selected := iv.SelectedSuggestion()
	if selected != "" {
		t.Error("SelectedSuggestion should be empty with no suggestions")
	}
}

// --- Navigation Tests ---

func TestInputViewNavigateDown(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to navigate")
	}

	initialIdx := iv.SelectedIndex()
	msg := tea.KeyMsg{Type: tea.KeyDown}
	iv.Update(msg)

	if iv.SelectedIndex() != initialIdx+1 {
		t.Errorf("Down should increment index from %d to %d, got %d", initialIdx, initialIdx+1, iv.SelectedIndex())
	}
}

func TestInputViewNavigateUp(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to navigate")
	}

	// Move down first
	msg := tea.KeyMsg{Type: tea.KeyDown}
	iv.Update(msg)

	newIdx := iv.SelectedIndex()
	msg = tea.KeyMsg{Type: tea.KeyUp}
	iv.Update(msg)

	if iv.SelectedIndex() != newIdx-1 {
		t.Errorf("Up should decrement index from %d to %d, got %d", newIdx, newIdx-1, iv.SelectedIndex())
	}
}

func TestInputViewNavigateUpWrap(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to navigate")
	}

	// Navigate up from 0 should wrap to last
	msg := tea.KeyMsg{Type: tea.KeyUp}
	iv.Update(msg)

	expectedIdx := iv.SuggestionCount() - 1
	if iv.SelectedIndex() != expectedIdx {
		t.Errorf("Up from 0 should wrap to %d, got %d", expectedIdx, iv.SelectedIndex())
	}
}

func TestInputViewNavigateDownWrap(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to navigate")
	}

	// Navigate to last item
	for i := 0; i < iv.SuggestionCount()-1; i++ {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		iv.Update(msg)
	}

	// Navigate down one more should wrap to 0
	msg := tea.KeyMsg{Type: tea.KeyDown}
	iv.Update(msg)

	if iv.SelectedIndex() != 0 {
		t.Errorf("Down from last should wrap to 0, got %d", iv.SelectedIndex())
	}
}

// --- Tab Completion Tests ---

func TestInputViewTabComplete(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to complete")
	}

	expected := iv.SelectedSuggestion()
	msg := tea.KeyMsg{Type: tea.KeyTab}
	iv.Update(msg)

	if iv.Value() != expected {
		t.Errorf("Tab should set value to %q, got %q", expected, iv.Value())
	}
}

func TestInputViewTabCompleteNavigated(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if iv.SuggestionCount() < 2 {
		t.Skip("Need at least 2 suggestions")
	}

	// Navigate to second suggestion
	msg := tea.KeyMsg{Type: tea.KeyDown}
	iv.Update(msg)

	expected := iv.SelectedSuggestion()
	msg = tea.KeyMsg{Type: tea.KeyTab}
	iv.Update(msg)

	if iv.Value() != expected {
		t.Errorf("Tab should complete to selected suggestion %q, got %q", expected, iv.Value())
	}
}

// --- Enter Selection Tests ---

func TestInputViewEnterSelectsSuggestion(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to select")
	}

	expected := iv.SelectedSuggestion()
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, selected, value := iv.Update(msg)

	if !selected {
		t.Error("Enter should indicate selection")
	}
	if value != expected {
		t.Errorf("Enter should return selected value %q, got %q", expected, value)
	}
}

func TestInputViewEnterWithTypedValue(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	// Type a value that doesn't match anything
	iv.SetValue("nonexistent-secret")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, selected, value := iv.Update(msg)

	if !selected {
		t.Error("Enter with typed value should indicate selection")
	}
	if value != "nonexistent-secret" {
		t.Errorf("Enter should return typed value, got %q", value)
	}
}

func TestInputViewEnterEmptyNoSelection(t *testing.T) {
	iv := NewInputView(nil, "", "")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, selected, value := iv.Update(msg)

	if selected {
		t.Error("Enter with empty input and no suggestions should not select")
	}
	if value != "" {
		t.Errorf("Value should be empty, got %q", value)
	}
}

// --- Escape Tests ---

func TestInputViewEscapeClearsInput(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetValue("test")
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	iv.Update(msg)

	if iv.Value() != "" {
		t.Error("Escape should clear input value")
	}
}

func TestInputViewEscapeEmptyInput(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	// Escape on empty input should signal cancel (empty return)
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, selected, value := iv.Update(msg)

	if selected {
		t.Error("Escape should not indicate selection")
	}
	if value != "" {
		t.Error("Escape should return empty value")
	}
}

// --- View Rendering Tests ---

func TestInputViewView(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "placeholder", "Test Prompt")

	view := iv.View()

	// Should contain prompt
	if !strings.Contains(view, "Test Prompt") {
		t.Error("View should contain prompt")
	}

	// Should contain help text
	if !strings.Contains(view, "navigate") {
		t.Error("View should contain help text")
	}
}

func TestInputViewViewShowsSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	view := iv.View()

	// Should show some suggestions
	hasAnySuggestion := false
	for _, s := range iv.Suggestions() {
		if strings.Contains(view, s) {
			hasAnySuggestion = true
			break
		}
	}
	if !hasAnySuggestion && iv.HasSuggestions() {
		t.Error("View should display suggestions")
	}
}

func TestInputViewViewNoMatchesMessage(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetValue("xyznonexistent")
	view := iv.View()

	if !strings.Contains(view, "No matching") {
		t.Error("View should show 'No matching' message when no suggestions")
	}
}

func TestInputViewViewSelectedHighlight(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Skip("No suggestions to highlight")
	}

	view := iv.View()

	// Should contain ">" for selected item
	if !strings.Contains(view, ">") {
		t.Error("View should highlight selected suggestion with '>'")
	}
}

// --- Focus Tests ---

func TestInputViewFocus(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	// Should be focused initially
	if !iv.IsFocused() {
		t.Error("InputView should be focused initially")
	}

	iv.Blur()
	if iv.IsFocused() {
		t.Error("InputView should not be focused after Blur")
	}

	iv.Focus()
	// Note: Focus returns a command, actual focus state depends on tea runtime
}

// --- Factory Function Tests ---

func TestCreateForSecretSearch(t *testing.T) {
	v := createTestVault()
	iv := CreateForSecretSearch(v)

	if iv == nil {
		t.Fatal("CreateForSecretSearch should not return nil")
	}
	if !strings.Contains(iv.GetPrompt(), "Search") {
		t.Error("Search input should have 'Search' in prompt")
	}
}

func TestCreateForSecretSelection(t *testing.T) {
	v := createTestVault()
	iv := CreateForSecretSelection(v)

	if iv == nil {
		t.Fatal("CreateForSecretSelection should not return nil")
	}
	if !strings.Contains(iv.GetPrompt(), "Select") {
		t.Error("Selection input should have 'Select' in prompt")
	}
}

// --- HasSuggestions Tests ---

func TestInputViewHasSuggestions(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	if !iv.HasSuggestions() {
		t.Error("Should have suggestions with populated vault")
	}

	iv.SetValue("xyznonexistent")
	if iv.HasSuggestions() {
		t.Error("Should not have suggestions with non-matching query")
	}
}

// --- Prompt and Placeholder Tests ---

func TestInputViewSetPrompt(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetPrompt("New Prompt")
	if iv.GetPrompt() != "New Prompt" {
		t.Errorf("prompt = %q, want 'New Prompt'", iv.GetPrompt())
	}
}

func TestInputViewSetPlaceholder(t *testing.T) {
	v := createTestVault()
	iv := NewInputView(v, "", "")

	iv.SetPlaceholder("New Placeholder")
	if iv.GetPlaceholder() != "New Placeholder" {
		t.Errorf("placeholder = %q, want 'New Placeholder'", iv.GetPlaceholder())
	}
}
