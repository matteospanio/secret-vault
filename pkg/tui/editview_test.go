package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

func TestNewEditViewForNewSecret(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	if ev.IsEditing() {
		t.Error("Should not be editing when secret is nil")
	}
	if ev.FocusIndex() != editFieldName {
		t.Errorf("Initial focus should be on name field, got %d", ev.FocusIndex())
	}
	if ev.GetName() != "" {
		t.Error("Name should be empty for new secret")
	}
	if ev.GetValue() != "" {
		t.Error("Value should be empty for new secret")
	}
}

func TestNewEditViewForExistingSecret(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("my-token", "secret123", "My API token", "work", []string{"api", "github"})
	secret, _ := v.GetSecret("my-token")

	ev := NewEditView(v, &secret, 80, 24)

	if !ev.IsEditing() {
		t.Error("Should be editing when secret is provided")
	}
	if ev.GetName() != "my-token" {
		t.Errorf("Name should be 'my-token', got %q", ev.GetName())
	}
	if ev.GetValue() != "secret123" {
		t.Errorf("Value should be 'secret123', got %q", ev.GetValue())
	}
	if ev.GetDescription() != "My API token" {
		t.Errorf("Description should be 'My API token', got %q", ev.GetDescription())
	}
	if ev.GetCategory() != "work" {
		t.Errorf("Category should be 'work', got %q", ev.GetCategory())
	}
	if ev.GetTagsRaw() != "api, github" {
		t.Errorf("Tags should be 'api, github', got %q", ev.GetTagsRaw())
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", nil},
		{"whitespace only", "   ", nil},
		{"single tag", "api", []string{"api"}},
		{"multiple tags", "api, github, work", []string{"api", "github", "work"}},
		{"tags with extra spaces", "  api ,  github  , work  ", []string{"api", "github", "work"}},
		{"trailing comma", "api,github,", []string{"api", "github"}},
		{"only commas", ",,,", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("Tag %d: expected %q, got %q", i, tt.expected[i], tag)
				}
			}
		})
	}
}

func TestEditViewTabNavigation(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	if ev.FocusIndex() != editFieldName {
		t.Fatalf("Initial focus should be on name field")
	}

	// Tab to next field
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	ev.Update(tabMsg)
	if ev.FocusIndex() != editFieldValue {
		t.Errorf("After tab, focus should be on value field, got %d", ev.FocusIndex())
	}

	// Tab again
	ev.Update(tabMsg)
	if ev.FocusIndex() != editFieldDescription {
		t.Errorf("After 2nd tab, focus should be on description field, got %d", ev.FocusIndex())
	}

	// Tab through all fields and wrap around
	ev.Update(tabMsg) // category
	ev.Update(tabMsg) // tags
	ev.Update(tabMsg) // wraps to name
	if ev.FocusIndex() != editFieldName {
		t.Errorf("After wrapping, focus should be on name field, got %d", ev.FocusIndex())
	}
}

func TestEditViewShiftTabNavigation(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Shift+Tab from first field should wrap to last
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	ev.Update(shiftTabMsg)
	if ev.FocusIndex() != editFieldTags {
		t.Errorf("Shift+tab from first field should wrap to tags, got %d", ev.FocusIndex())
	}

	// Shift+Tab again
	ev.Update(shiftTabMsg)
	if ev.FocusIndex() != editFieldCategory {
		t.Errorf("Shift+tab from tags should go to category, got %d", ev.FocusIndex())
	}
}

func TestEditViewDownUpNavigation(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Down key should move to next field
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	ev.Update(downMsg)
	if ev.FocusIndex() != editFieldValue {
		t.Errorf("Down should move to value field, got %d", ev.FocusIndex())
	}

	// Up key should move to previous field
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	ev.Update(upMsg)
	if ev.FocusIndex() != editFieldName {
		t.Errorf("Up should move back to name field, got %d", ev.FocusIndex())
	}
}

func TestEditViewSaveValidation(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Try to save with empty fields
	ctrlSMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	cmd, _ := ev.Update(ctrlSMsg)
	if cmd != nil {
		t.Error("Save with empty name should not return a command")
	}
	if ev.GetError() != "Name is required" {
		t.Errorf("Expected 'Name is required' error, got %q", ev.GetError())
	}

	// Set name but not value
	ev.inputs[editFieldName].SetValue("my-secret")
	cmd, _ = ev.Update(ctrlSMsg)
	if cmd != nil {
		t.Error("Save with empty value should not return a command")
	}
	if ev.GetError() != "Value is required" {
		t.Errorf("Expected 'Value is required' error, got %q", ev.GetError())
	}
}

func TestEditViewSaveSuccess(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Set required fields
	ev.inputs[editFieldName].SetValue("my-secret")
	ev.inputs[editFieldValue].SetValue("my-value")
	ev.inputs[editFieldDescription].SetValue("My description")
	ev.inputs[editFieldCategory].SetValue("work")
	ev.inputs[editFieldTags].SetValue("api, github")

	// Save
	ctrlSMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	cmd, handled := ev.Update(ctrlSMsg)
	if !handled {
		t.Error("Save should be handled")
	}
	if cmd == nil {
		t.Fatal("Save should return a command")
	}

	// Execute the command and check the message
	msg := cmd()
	savedMsg, ok := msg.(SecretSavedMsg)
	if !ok {
		t.Fatalf("Expected SecretSavedMsg, got %T", msg)
	}
	if savedMsg.Name != "my-secret" {
		t.Errorf("Saved name should be 'my-secret', got %q", savedMsg.Name)
	}

	// Verify secret was saved in vault
	secret, err := v.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("Secret should exist in vault: %v", err)
	}
	if secret.Value != "my-value" {
		t.Errorf("Secret value should be 'my-value', got %q", secret.Value)
	}
	if secret.Description != "My description" {
		t.Errorf("Description should be 'My description', got %q", secret.Description)
	}
	if secret.Category != "work" {
		t.Errorf("Category should be 'work', got %q", secret.Category)
	}
	if len(secret.Tags) != 2 || secret.Tags[0] != "api" || secret.Tags[1] != "github" {
		t.Errorf("Tags should be [api, github], got %v", secret.Tags)
	}
}

func TestEditViewSaveUpdatesExisting(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("existing", "old-value", "old desc", "old-cat", []string{"old-tag"})

	secret, _ := v.GetSecret("existing")
	ev := NewEditView(v, &secret, 80, 24)

	// Update value
	ev.inputs[editFieldValue].SetValue("new-value")
	ev.inputs[editFieldCategory].SetValue("new-cat")

	ctrlSMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	cmd, _ := ev.Update(ctrlSMsg)
	if cmd == nil {
		t.Fatal("Save should return a command")
	}

	// Execute command
	cmd()

	// Verify update
	updated, _ := v.GetSecret("existing")
	if updated.Value != "new-value" {
		t.Errorf("Value should be updated to 'new-value', got %q", updated.Value)
	}
	if updated.Category != "new-cat" {
		t.Errorf("Category should be updated to 'new-cat', got %q", updated.Category)
	}
}

func TestEditViewRenderNewSecret(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	view := ev.View()

	if !contains(view, "Add Secret") {
		t.Error("View should show 'Add Secret' for new secret")
	}
	if !contains(view, "Name:") {
		t.Error("View should show 'Name:' label")
	}
	if !contains(view, "Value:") {
		t.Error("View should show 'Value:' label")
	}
	if !contains(view, "Category:") {
		t.Error("View should show 'Category:' label")
	}
	if !contains(view, "Tags:") {
		t.Error("View should show 'Tags:' label")
	}
	if !contains(view, "ctrl+s: save") {
		t.Error("View should show save hint")
	}
}

func TestEditViewRenderEditSecret(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test", "val", "", "", nil)
	secret, _ := v.GetSecret("test")
	ev := NewEditView(v, &secret, 80, 24)

	view := ev.View()

	if !contains(view, "Edit Secret") {
		t.Error("View should show 'Edit Secret' for existing secret")
	}
}

func TestEditViewRenderWithError(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Trigger validation error
	ctrlSMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	ev.Update(ctrlSMsg)

	view := ev.View()
	if !contains(view, "Error:") {
		t.Error("View should show validation error")
	}
	if !contains(view, "Name is required") {
		t.Error("View should show 'Name is required' error")
	}
}

func TestEditViewSetSize(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	ev.SetSize(120, 40)
	if ev.width != 120 {
		t.Errorf("Width should be 120, got %d", ev.width)
	}
	if ev.height != 40 {
		t.Errorf("Height should be 40, got %d", ev.height)
	}
}

func TestEditViewEnterOnLastField(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Navigate to last field (tags)
	for i := 0; i < editFieldTags; i++ {
		tabMsg := tea.KeyMsg{Type: tea.KeyTab}
		ev.Update(tabMsg)
	}

	if ev.FocusIndex() != editFieldTags {
		t.Fatalf("Should be on tags field, got %d", ev.FocusIndex())
	}

	// Set required fields to allow save
	ev.inputs[editFieldName].SetValue("test")
	ev.inputs[editFieldValue].SetValue("value")

	// Enter on last field should trigger save
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd, handled := ev.Update(enterMsg)
	if !handled {
		t.Error("Enter on last field should be handled")
	}
	if cmd == nil {
		t.Error("Enter on last field should trigger save")
	}
}

func TestEditViewEnterOnNonLastField(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	// Focus is on name field (first field)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	ev.Update(enterMsg)

	// Should move to next field, not save
	if ev.FocusIndex() != editFieldValue {
		t.Errorf("Enter on non-last field should move to next, got %d", ev.FocusIndex())
	}
}

func TestEditViewGetParsedTags(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	ev.inputs[editFieldTags].SetValue("api, github, test")
	tags := ev.GetParsedTags()

	if len(tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "api" || tags[1] != "github" || tags[2] != "test" {
		t.Errorf("Unexpected tags: %v", tags)
	}
}

func TestEditViewSaveTrimsWhitespace(t *testing.T) {
	v := vault.NewVault()
	ev := NewEditView(v, nil, 80, 24)

	ev.inputs[editFieldName].SetValue("  my-secret  ")
	ev.inputs[editFieldValue].SetValue("  my-value  ")
	ev.inputs[editFieldDescription].SetValue("  desc  ")
	ev.inputs[editFieldCategory].SetValue("  work  ")

	ctrlSMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	cmd, _ := ev.Update(ctrlSMsg)
	if cmd != nil {
		cmd()
	}

	secret, err := v.GetSecret("my-secret")
	if err != nil {
		t.Fatalf("Secret should exist: %v", err)
	}
	if secret.Value != "my-value" {
		t.Errorf("Value should be trimmed, got %q", secret.Value)
	}
	if secret.Description != "desc" {
		t.Errorf("Description should be trimmed, got %q", secret.Description)
	}
	if secret.Category != "work" {
		t.Errorf("Category should be trimmed, got %q", secret.Category)
	}
}
