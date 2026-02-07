package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

func TestSecretItemFilterValue(t *testing.T) {
	secret := &vault.Secret{Name: "test-secret", Value: "secret-value"}
	item := SecretItem{secret: secret}

	if item.FilterValue() != "test-secret" {
		t.Errorf("FilterValue should return secret name, got %q", item.FilterValue())
	}
}

func TestSecretItemTitle(t *testing.T) {
	secret := &vault.Secret{Name: "my-secret", Value: "value"}
	item := SecretItem{secret: secret}

	if item.Title() != "my-secret" {
		t.Errorf("Title should return secret name, got %q", item.Title())
	}
}

func TestSecretItemDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{"with description", "A test secret", "A test secret"},
		{"empty description", "", "No description"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := &vault.Secret{Name: "test", Value: "value", Description: tt.description}
			item := SecretItem{secret: secret}

			if item.Description() != tt.expected {
				t.Errorf("Description() = %q, want %q", item.Description(), tt.expected)
			}
		})
	}
}

func TestSecretItemSecret(t *testing.T) {
	secret := &vault.Secret{Name: "test-secret", Value: "secret-value"}
	item := SecretItem{secret: secret}

	if item.Secret() != secret {
		t.Error("Secret() should return the underlying secret")
	}
}

func TestSecretItemAgeWarningInDelegate(t *testing.T) {
	// This test verifies that the secretItemDelegate renders age warning for old secrets
	// The actual rendering is done in the Render method which writes to an io.Writer
	// We can't easily test the visual output, but we verify the IsOld logic is used

	now := time.Now()
	oldSecret := &vault.Secret{
		Name:      "old-secret",
		Value:     "value",
		CreatedAt: now.Add(-400 * 24 * time.Hour), // ~13 months ago
		UpdatedAt: now.Add(-400 * 24 * time.Hour),
	}
	newSecret := &vault.Secret{
		Name:      "new-secret",
		Value:     "value",
		CreatedAt: now.Add(-30 * 24 * time.Hour), // 1 month ago
		UpdatedAt: now.Add(-30 * 24 * time.Hour),
	}

	// Verify the IsOld method works correctly (used in delegate)
	if !oldSecret.IsOld(vault.DefaultAgeThreshold) {
		t.Error("Old secret (400 days) should be marked as old")
	}
	if newSecret.IsOld(vault.DefaultAgeThreshold) {
		t.Error("New secret (30 days) should not be marked as old")
	}
}

func TestNewListViewEmpty(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	if lv == nil {
		t.Fatal("NewListView should not return nil")
	}

	if lv.ItemCount() != 0 {
		t.Errorf("Empty vault should have 0 items, got %d", lv.ItemCount())
	}
}

func TestNewListViewWithSecrets(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value-a", "Description A", "work", nil)
	v.AddSecret("secret-b", "value-b", "Description B", "personal", []string{"tag1"})
	v.AddSecret("secret-c", "value-c", "", "", nil)

	lv := NewListView(v, 80, 24)

	if lv.ItemCount() != 3 {
		t.Errorf("List should have 3 items, got %d", lv.ItemCount())
	}
}

func TestListViewSelectedItem(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value-a", "Description A", "", nil)
	v.AddSecret("secret-b", "value-b", "Description B", "", nil)

	lv := NewListView(v, 80, 24)

	// First item should be selected by default
	item := lv.SelectedItem()
	if item == nil {
		t.Fatal("SelectedItem should not be nil")
	}

	// Items are sorted alphabetically
	if item.Title() != "secret-a" {
		t.Errorf("First selected item should be 'secret-a', got %q", item.Title())
	}
}

func TestListViewSelectedItemEmpty(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	item := lv.SelectedItem()
	if item != nil {
		t.Error("SelectedItem should be nil for empty list")
	}
}

func TestListViewSelectedSecret(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("my-secret", "my-value", "My description", "", nil)

	lv := NewListView(v, 80, 24)

	secret := lv.SelectedSecret()
	if secret == nil {
		t.Fatal("SelectedSecret should not be nil")
	}

	if secret.Name != "my-secret" {
		t.Errorf("Selected secret name should be 'my-secret', got %q", secret.Name)
	}
}

func TestListViewSelectedSecretEmpty(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	secret := lv.SelectedSecret()
	if secret != nil {
		t.Error("SelectedSecret should be nil for empty list")
	}
}

func TestListViewSetSize(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	lv.SetSize(120, 40)

	if lv.width != 120 {
		t.Errorf("Width should be 120, got %d", lv.width)
	}
	if lv.height != 40 {
		t.Errorf("Height should be 40, got %d", lv.height)
	}
}

func TestListViewRefresh(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("initial-secret", "value", "", "", nil)

	lv := NewListView(v, 80, 24)

	if lv.ItemCount() != 1 {
		t.Errorf("Initial count should be 1, got %d", lv.ItemCount())
	}

	// Add another secret to the vault
	v.AddSecret("new-secret", "value2", "", "", nil)

	// Refresh the list
	lv.Refresh()

	if lv.ItemCount() != 2 {
		t.Errorf("After refresh, count should be 2, got %d", lv.ItemCount())
	}
}

func TestListViewViewEmpty(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	view := lv.View()

	if !contains(view, "No secrets stored yet") {
		t.Errorf("Empty view should contain 'No secrets stored yet', got %q", view)
	}
}

func TestListViewViewWithSecrets(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "value", "A test secret", "", nil)

	lv := NewListView(v, 80, 24)

	view := lv.View()

	if !contains(view, "test-secret") {
		t.Errorf("View should contain secret name, got %q", view)
	}
}

func TestListViewUpdate(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value", "", "", nil)
	v.AddSecret("secret-b", "value", "", "", nil)

	lv := NewListView(v, 80, 24)

	// Initial selection should be first item
	if lv.SelectedItem().Title() != "secret-a" {
		t.Errorf("Initial selection should be 'secret-a', got %q", lv.SelectedItem().Title())
	}

	// Simulate down key press
	msg := tea.KeyMsg{Type: tea.KeyDown}
	lv.Update(msg)

	// After down, second item should be selected
	if lv.SelectedItem().Title() != "secret-b" {
		t.Errorf("After down key, selection should be 'secret-b', got %q", lv.SelectedItem().Title())
	}
}

func TestListViewFilterState(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "value", "", "", nil)

	lv := NewListView(v, 80, 24)

	// Initially not filtering
	if lv.IsFiltering() {
		t.Error("Initially should not be filtering")
	}
}

func TestSecretsToItems(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("zebra", "value", "", "", nil)
	v.AddSecret("apple", "value", "", "", nil)
	v.AddSecret("mango", "value", "", "", nil)

	items := secretsToItems(v)

	if len(items) != 3 {
		t.Errorf("Should have 3 items, got %d", len(items))
	}

	// Items should be sorted alphabetically
	if items[0].(SecretItem).Title() != "apple" {
		t.Errorf("First item should be 'apple', got %q", items[0].(SecretItem).Title())
	}
	if items[1].(SecretItem).Title() != "mango" {
		t.Errorf("Second item should be 'mango', got %q", items[1].(SecretItem).Title())
	}
	if items[2].(SecretItem).Title() != "zebra" {
		t.Errorf("Third item should be 'zebra', got %q", items[2].(SecretItem).Title())
	}
}

func TestSecretItemDelegateHeight(t *testing.T) {
	d := secretItemDelegate{}
	if d.Height() != 2 {
		t.Errorf("Delegate height should be 2, got %d", d.Height())
	}
}

func TestSecretItemDelegateSpacing(t *testing.T) {
	d := secretItemDelegate{}
	if d.Spacing() != 1 {
		t.Errorf("Delegate spacing should be 1, got %d", d.Spacing())
	}
}

func TestModelWithListView(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "value", "Description", "", nil)

	m := NewModel(v)

	// Initially list view is nil (not initialized until window size received)
	if m.GetListView() != nil {
		t.Error("List view should be nil before window size message")
	}

	// Send window size message to initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	if m.GetListView() == nil {
		t.Fatal("List view should be initialized after window size message")
	}

	if m.GetListView().ItemCount() != 1 {
		t.Errorf("List view should have 1 item, got %d", m.GetListView().ItemCount())
	}
}

func TestModelListViewSelection(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "value", "Description", "", nil)

	m := NewModel(v)

	// Initialize with window size
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press enter to select
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(Model)

	// Should switch to detail view
	if m.GetCurrentView() != ViewDetail {
		t.Errorf("View should be ViewDetail after enter, got %v", m.GetCurrentView())
	}

	// Secret should be selected
	if m.GetSelectedSecret() == nil {
		t.Fatal("Secret should be selected after enter")
	}

	if m.GetSelectedSecret().Name != "test-secret" {
		t.Errorf("Selected secret name should be 'test-secret', got %q", m.GetSelectedSecret().Name)
	}
}

func TestModelListViewNavigation(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value", "", "", nil)
	v.AddSecret("secret-b", "value", "", "", nil)

	m := NewModel(v)

	// Initialize with window size
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press down to navigate
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = m.Update(downMsg)
	m = newModel.(Model)

	// The list should have moved selection
	if m.GetListView().SelectedItem().Title() != "secret-b" {
		t.Errorf("After down, should select 'secret-b', got %q", m.GetListView().SelectedItem().Title())
	}
}

func TestListViewActiveFilters(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value", "", "work", nil)
	v.AddSecret("secret-b", "value", "", "personal", nil)

	lv := NewListView(v, 80, 24)

	// Initially no active filters
	if lv.HasActiveFilters() {
		t.Error("Initially should have no active filters")
	}
	if lv.GetActiveFilters() != nil {
		t.Error("GetActiveFilters should return nil initially")
	}

	// Set active filters
	criteria := &FilterCriteria{Category: "work"}
	lv.SetActiveFilters(criteria)

	if !lv.HasActiveFilters() {
		t.Error("Should have active filters after SetActiveFilters")
	}
	if lv.GetActiveFilters() != criteria {
		t.Error("GetActiveFilters should return the set criteria")
	}

	// Clear filters
	lv.ClearFilters()
	if lv.HasActiveFilters() {
		t.Error("Should not have active filters after ClearFilters")
	}
	// After clear, list should be refreshed with all secrets
	if lv.ItemCount() != 2 {
		t.Errorf("After ClearFilters, should have all 2 items, got %d", lv.ItemCount())
	}
}

func TestListViewFilterHeader(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value", "", "work", []string{"api"})

	lv := NewListView(v, 80, 24)

	// No filter header without active filters
	header := lv.renderFilterHeader()
	if header != "" {
		t.Error("Should have no header without active filters")
	}

	// Set filters and check header
	criteria := &FilterCriteria{Category: "work", Tag: "api"}
	lv.SetActiveFilters(criteria)

	header = lv.renderFilterHeader()
	if !contains(header, "work") {
		t.Errorf("Header should contain category, got %q", header)
	}
	if !contains(header, "api") {
		t.Errorf("Header should contain tag, got %q", header)
	}
}

func TestListViewNoResultsWithFilters(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	// Set active filters
	criteria := &FilterCriteria{Category: "nonexistent"}
	lv.SetActiveFilters(criteria)

	// With active filters and no items, should show "no results" message
	view := lv.View()
	if !contains(view, "No secrets match the current filters") {
		t.Errorf("Should show no results message, got %q", view)
	}
}

func TestListViewFilteredItemsWithHeader(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "value", "", "work", nil)
	v.AddSecret("secret-b", "value", "", "personal", nil)

	lv := NewListView(v, 80, 24)

	// Set filtered items and active filters
	filtered := []vault.Secret{{Name: "secret-a", Value: "value", Category: "work"}}
	lv.SetFilteredItems(filtered)
	lv.SetActiveFilters(&FilterCriteria{Category: "work"})

	view := lv.View()
	if !contains(view, "work") {
		t.Errorf("Filtered view should show filter header with 'work', got %q", view)
	}
}

func TestListViewHasActiveFiltersWithEmptyCriteria(t *testing.T) {
	v := vault.NewVault()
	lv := NewListView(v, 80, 24)

	// Set empty criteria
	lv.SetActiveFilters(&FilterCriteria{})
	if lv.HasActiveFilters() {
		t.Error("HasActiveFilters should return false for empty criteria")
	}
}

func TestModelEmptyListEnter(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initialize with window size
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press enter on empty list
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(Model)

	// Should stay on list view (nothing to select)
	if m.GetCurrentView() != ViewList {
		t.Errorf("View should still be ViewList for empty list, got %v", m.GetCurrentView())
	}
}
