package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

func newTestVaultWithSecrets() *vault.Vault {
	v := vault.NewVault()
	v.AddSecret("github-token", "ghp_abc123", "GitHub API token", "work", []string{"api", "github"})
	v.AddSecret("aws-key", "AKIA123", "AWS access key", "work", []string{"api", "aws"})
	v.AddSecret("personal-email", "mypassword", "Email password", "personal", []string{"email"})
	v.AddSecret("db-password", "dbpass123", "Database password", "work", []string{"database"})
	return v
}

func TestNewFilterView(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	if fv.FocusIndex() != filterFieldSearch {
		t.Errorf("Initial focus should be on search field, got %d", fv.FocusIndex())
	}
	if fv.IsOldOnly() {
		t.Error("Old-only should be false initially")
	}
	if fv.GetSearchQuery() != "" {
		t.Error("Search query should be empty initially")
	}
	if fv.GetCategoryFilter() != "" {
		t.Error("Category filter should be empty initially")
	}
	if fv.GetTagFilter() != "" {
		t.Error("Tag filter should be empty initially")
	}
}

func TestFilterCriteriaIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		criteria FilterCriteria
		expected bool
	}{
		{"empty criteria", FilterCriteria{}, true},
		{"with query", FilterCriteria{Query: "test"}, false},
		{"with category", FilterCriteria{Category: "work"}, false},
		{"with tag", FilterCriteria{Tag: "api"}, false},
		{"with old only", FilterCriteria{OldOnly: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.criteria.IsEmpty() != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", tt.criteria.IsEmpty(), tt.expected)
			}
		})
	}
}

func TestFilterViewTabNavigation(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	tabMsg := tea.KeyMsg{Type: tea.KeyTab}

	// Tab through fields
	fv.Update(tabMsg)
	if fv.FocusIndex() != filterFieldCategory {
		t.Errorf("After tab, focus should be on category, got %d", fv.FocusIndex())
	}

	fv.Update(tabMsg)
	if fv.FocusIndex() != filterFieldTag {
		t.Errorf("After 2nd tab, focus should be on tag, got %d", fv.FocusIndex())
	}

	// Wrap around
	fv.Update(tabMsg)
	if fv.FocusIndex() != filterFieldSearch {
		t.Errorf("After wrapping, focus should be on search, got %d", fv.FocusIndex())
	}
}

func TestFilterViewShiftTabNavigation(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}

	// Shift+Tab from first should wrap to last
	fv.Update(shiftTabMsg)
	if fv.FocusIndex() != filterFieldTag {
		t.Errorf("Shift+tab should wrap to tag field, got %d", fv.FocusIndex())
	}
}

func TestFilterViewToggleOldOnly(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	if fv.IsOldOnly() {
		t.Error("Initially should not be old-only")
	}

	// Toggle with ctrl+o
	ctrlOMsg := tea.KeyMsg{Type: tea.KeyCtrlO}
	fv.Update(ctrlOMsg)
	if !fv.IsOldOnly() {
		t.Error("After ctrl+o, should be old-only")
	}

	// Toggle back
	fv.Update(ctrlOMsg)
	if fv.IsOldOnly() {
		t.Error("After 2nd ctrl+o, should not be old-only")
	}
}

func TestFilterViewClearAll(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	// Set some values
	fv.inputs[filterFieldSearch].SetValue("test")
	fv.inputs[filterFieldCategory].SetValue("work")
	fv.inputs[filterFieldTag].SetValue("api")
	fv.oldOnly = true

	// Clear all
	ctrlRMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	fv.Update(ctrlRMsg)

	if fv.GetSearchQuery() != "" {
		t.Error("Search should be cleared")
	}
	if fv.GetCategoryFilter() != "" {
		t.Error("Category should be cleared")
	}
	if fv.GetTagFilter() != "" {
		t.Error("Tag should be cleared")
	}
	if fv.IsOldOnly() {
		t.Error("Old-only should be cleared")
	}
}

func TestFilterViewApplyFilters(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	fv.inputs[filterFieldSearch].SetValue("github")
	fv.inputs[filterFieldCategory].SetValue("work")

	// Press enter to apply
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd, handled := fv.Update(enterMsg)
	if !handled {
		t.Error("Enter should be handled")
	}
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	// Execute command
	msg := cmd()
	appliedMsg, ok := msg.(FilterAppliedMsg)
	if !ok {
		t.Fatalf("Expected FilterAppliedMsg, got %T", msg)
	}
	if appliedMsg.Query != "github" {
		t.Errorf("Query should be 'github', got %q", appliedMsg.Query)
	}
	if appliedMsg.Category != "work" {
		t.Errorf("Category should be 'work', got %q", appliedMsg.Category)
	}
}

func TestFilterViewGetCriteria(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	fv.inputs[filterFieldSearch].SetValue("  test  ")
	fv.inputs[filterFieldCategory].SetValue("  work  ")
	fv.inputs[filterFieldTag].SetValue("  api  ")
	fv.oldOnly = true

	criteria := fv.GetCriteria()
	if criteria.Query != "test" {
		t.Errorf("Query should be trimmed to 'test', got %q", criteria.Query)
	}
	if criteria.Category != "work" {
		t.Errorf("Category should be trimmed to 'work', got %q", criteria.Category)
	}
	if criteria.Tag != "api" {
		t.Errorf("Tag should be trimmed to 'api', got %q", criteria.Tag)
	}
	if !criteria.OldOnly {
		t.Error("OldOnly should be true")
	}
}

func TestApplyFilterCriteriaNoFilters(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{})

	if len(results) != 4 {
		t.Errorf("No filters should return all 4 secrets, got %d", len(results))
	}
}

func TestApplyFilterCriteriaByQuery(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Query: "token"})

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'token', got %d", len(results))
		return
	}
	if results[0].Name != "github-token" {
		t.Errorf("Expected github-token, got %q", results[0].Name)
	}
}

func TestApplyFilterCriteriaByCategory(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Category: "work"})

	if len(results) != 3 {
		t.Errorf("Expected 3 work secrets, got %d", len(results))
	}
}

func TestApplyFilterCriteriaByCategoryPersonal(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Category: "personal"})

	if len(results) != 1 {
		t.Errorf("Expected 1 personal secret, got %d", len(results))
		return
	}
	if results[0].Name != "personal-email" {
		t.Errorf("Expected personal-email, got %q", results[0].Name)
	}
}

func TestApplyFilterCriteriaByTag(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Tag: "api"})

	if len(results) != 2 {
		t.Errorf("Expected 2 secrets with 'api' tag, got %d", len(results))
	}
}

func TestApplyFilterCriteriaCombined(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Category: "work", Tag: "api"})

	if len(results) != 2 {
		t.Errorf("Expected 2 secrets with category=work and tag=api, got %d", len(results))
	}
}

func TestApplyFilterCriteriaQueryAndCategory(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Query: "password", Category: "work"})

	if len(results) != 1 {
		t.Errorf("Expected 1 result for query=password category=work, got %d", len(results))
		return
	}
	if results[0].Name != "db-password" {
		t.Errorf("Expected db-password, got %q", results[0].Name)
	}
}

func TestApplyFilterCriteriaOldOnly(t *testing.T) {
	v := vault.NewVault()
	// Add an old secret
	v.Secrets["old-secret"] = vault.Secret{
		Name:      "old-secret",
		Value:     "old",
		UpdatedAt: time.Now().Add(-2 * 365 * 24 * time.Hour),
	}
	// Add a new secret
	v.AddSecret("new-secret", "new", "", "", nil)

	results := ApplyFilterCriteria(v, FilterCriteria{OldOnly: true})

	if len(results) != 1 {
		t.Errorf("Expected 1 old secret, got %d", len(results))
		return
	}
	if results[0].Name != "old-secret" {
		t.Errorf("Expected old-secret, got %q", results[0].Name)
	}
}

func TestApplyFilterCriteriaNoResults(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{Category: "nonexistent"})

	if len(results) != 0 {
		t.Errorf("Expected 0 results for nonexistent category, got %d", len(results))
	}
}

func TestApplyFilterCriteriaSortedByName(t *testing.T) {
	v := newTestVaultWithSecrets()
	results := ApplyFilterCriteria(v, FilterCriteria{})

	for i := 1; i < len(results); i++ {
		if results[i-1].Name >= results[i].Name {
			t.Errorf("Results should be sorted by name: %q >= %q", results[i-1].Name, results[i].Name)
		}
	}
}

func TestFilterViewRender(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	view := fv.View()

	if !contains(view, "Filter Secrets") {
		t.Error("View should show 'Filter Secrets' title")
	}
	if !contains(view, "Search:") {
		t.Error("View should show 'Search:' label")
	}
	if !contains(view, "Category:") {
		t.Error("View should show 'Category:' label")
	}
	if !contains(view, "Tag:") {
		t.Error("View should show 'Tag:' label")
	}
	if !contains(view, "[ ]") {
		t.Error("View should show unchecked old-only toggle")
	}
	if !contains(view, "enter: apply") {
		t.Error("View should show apply hint")
	}
}

func TestFilterViewRenderOldOnlyChecked(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)
	fv.oldOnly = true

	view := fv.View()

	if !contains(view, "[x]") {
		t.Error("View should show checked old-only toggle")
	}
}

func TestFilterViewRenderActiveFilters(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	fv.inputs[filterFieldSearch].SetValue("test")
	fv.inputs[filterFieldCategory].SetValue("work")

	view := fv.View()

	if !contains(view, "Active filters") {
		t.Error("View should show active filters summary")
	}
	if !contains(view, "search=") {
		t.Error("View should show search filter")
	}
	if !contains(view, "category=") {
		t.Error("View should show category filter")
	}
}

func TestFilterViewSetSize(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	fv.SetSize(120, 40)
	if fv.width != 120 {
		t.Errorf("Width should be 120, got %d", fv.width)
	}
	if fv.height != 40 {
		t.Errorf("Height should be 40, got %d", fv.height)
	}
}

func TestFilterViewApplyToVault(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	fv.inputs[filterFieldCategory].SetValue("work")
	results := fv.ApplyToVault()

	if len(results) != 3 {
		t.Errorf("Expected 3 work secrets, got %d", len(results))
	}
}

func TestFilterViewDownUpNavigation(t *testing.T) {
	v := newTestVaultWithSecrets()
	fv := NewFilterView(v, 80, 24)

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	fv.Update(downMsg)
	if fv.FocusIndex() != filterFieldCategory {
		t.Errorf("Down should move to category, got %d", fv.FocusIndex())
	}

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	fv.Update(upMsg)
	if fv.FocusIndex() != filterFieldSearch {
		t.Errorf("Up should move back to search, got %d", fv.FocusIndex())
	}
}
