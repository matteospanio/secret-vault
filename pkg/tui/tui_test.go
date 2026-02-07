package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matteospanio/secret-vault/pkg/vault"
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

func TestModelFilterApplied(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("work-secret", "val", "", "work", nil)
	v.AddSecret("personal-secret", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Should store active filters
	if m.GetActiveFilters() == nil {
		t.Fatal("Active filters should be set after FilterAppliedMsg")
	}
	if m.GetActiveFilters().Category != "work" {
		t.Errorf("Active filter category should be 'work', got %q", m.GetActiveFilters().Category)
	}

	// List should be filtered
	if m.GetListView().ItemCount() != 1 {
		t.Errorf("Filtered list should have 1 item, got %d", m.GetListView().ItemCount())
	}
}

func TestModelFilterCleared(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("work-secret", "val", "", "work", nil)
	v.AddSecret("personal-secret", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter first
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Clear filters via empty FilterAppliedMsg
	clearMsg := FilterAppliedMsg{}
	newModel, _ = m.Update(clearMsg)
	m = newModel.(Model)

	if m.GetActiveFilters() != nil {
		t.Error("Active filters should be nil after clearing")
	}
	if m.GetListView().ItemCount() != 2 {
		t.Errorf("List should show all 2 items after clearing, got %d", m.GetListView().ItemCount())
	}
}

func TestModelFilterClearedMsg(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "val", "", "work", nil)
	v.AddSecret("secret-b", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Set active filters
	m.activeFilters = &FilterCriteria{Category: "work"}

	// Send FilterClearedMsg
	clearMsg := FilterClearedMsg{}
	newModel, _ = m.Update(clearMsg)
	m = newModel.(Model)

	if m.GetActiveFilters() != nil {
		t.Error("Active filters should be nil after FilterClearedMsg")
	}
}

func TestModelClearFiltersKey(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("secret-a", "val", "", "work", nil)
	v.AddSecret("secret-b", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Press 'F' to clear filters
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(Model)

	if m.GetActiveFilters() != nil {
		t.Error("Active filters should be nil after pressing F")
	}
	if !contains(m.GetStatusMessage(), "Filters cleared") {
		t.Errorf("Status should say 'Filters cleared', got %q", m.GetStatusMessage())
	}
}

func TestModelSearchBarInitialized(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Before window size, search bar is nil
	if m.GetSearchBar() != nil {
		t.Error("Search bar should be nil before window size")
	}

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	if m.GetSearchBar() == nil {
		t.Error("Search bar should be initialized after window size")
	}
}

func TestModelSearchBarSlashKey(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "val", "", "", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Press '/' to focus search bar
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	newModel, _ = m.Update(slashMsg)
	m = newModel.(Model)

	if !m.GetSearchBar().IsActive() {
		t.Error("Search bar should be active after pressing /")
	}
}

func TestModelSearchBarEscClear(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "val", "", "", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Focus search bar and type
	m.searchBar.Focus()
	m.searchBar.SetQuery("test")

	// Press esc should clear the query first
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.GetSearchBar().Query() != "" {
		t.Error("Search bar query should be cleared after first esc")
	}

	// Press esc again to blur (re-focus first)
	m.searchBar.Focus()
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.GetSearchBar().IsActive() {
		t.Error("Search bar should be inactive after esc with empty query")
	}
}

func TestModelSearchBarEnterBlurs(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("test-secret", "val", "", "", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Focus search bar
	m.searchBar.Focus()
	m.searchBar.SetQuery("test")

	// Press enter should blur search bar
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.Update(enterMsg)
	m = newModel.(Model)

	if m.GetSearchBar().IsActive() {
		t.Error("Search bar should be blurred after enter")
	}
	if m.GetSearchBar().Query() != "test" {
		t.Error("Search bar query should be preserved after enter")
	}
}

func TestModelSearchBarNoQuitWhenActive(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Focus search bar
	m.searchBar.Focus()

	// Press 'q' should not quit
	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := m.Update(qMsg)

	// cmd should not be tea.Quit
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("Should not quit when search bar is active")
		}
	}
}

func TestModelFilterPersistsAcrossViews(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("work-secret", "val", "", "work", nil)
	v.AddSecret("personal-secret", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Go to detail view and back
	m.SetView(ViewDetail)
	m.GoBack()

	// Active filters should persist
	if m.GetActiveFilters() == nil {
		t.Error("Active filters should persist across view changes")
	}
	if m.GetActiveFilters().Category != "work" {
		t.Errorf("Active filter category should still be 'work', got %q", m.GetActiveFilters().Category)
	}
}

func TestModelFilterViewPreFilled(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("work-secret", "val", "", "work", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Open filter view - should pre-fill with existing filters
	fMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	newModel, _ = m.Update(fMsg)
	m = newModel.(Model)

	if m.GetCurrentView() != ViewFilter {
		t.Errorf("Should be in filter view, got %v", m.GetCurrentView())
	}

	fv := m.GetFilterView()
	if fv == nil {
		t.Fatal("Filter view should be initialized")
	}
	criteria := fv.GetCriteria()
	if criteria.Category != "work" {
		t.Errorf("Filter view should be pre-filled with category 'work', got %q", criteria.Category)
	}
}

func TestFilterViewSetCriteria(t *testing.T) {
	v := vault.NewVault()
	fv := NewFilterView(v, 80, 24)

	criteria := FilterCriteria{
		Query:    "test",
		Category: "work",
		Tag:      "api",
		OldOnly:  true,
	}
	fv.SetCriteria(criteria)

	got := fv.GetCriteria()
	if got.Query != "test" {
		t.Errorf("Query should be 'test', got %q", got.Query)
	}
	if got.Category != "work" {
		t.Errorf("Category should be 'work', got %q", got.Category)
	}
	if got.Tag != "api" {
		t.Errorf("Tag should be 'api', got %q", got.Tag)
	}
	if !got.OldOnly {
		t.Error("OldOnly should be true")
	}
}

func TestModelSecretSavedReappliesFilters(t *testing.T) {
	v := vault.NewVault()
	v.AddSecret("work-secret", "val", "", "work", nil)
	v.AddSecret("personal-secret", "val", "", "personal", nil)
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Apply filter
	filterMsg := FilterAppliedMsg{Category: "work"}
	newModel, _ = m.Update(filterMsg)
	m = newModel.(Model)

	// Add a new secret to the vault (simulating a save)
	v.AddSecret("new-work-secret", "val2", "", "work", nil)

	// Send SecretSavedMsg
	savedMsg := SecretSavedMsg{Name: "new-work-secret"}
	newModel, _ = m.Update(savedMsg)
	m = newModel.(Model)

	// Should show 2 work secrets (reapplied filter)
	if m.GetListView().ItemCount() != 2 {
		t.Errorf("After save with active filter, should show 2 filtered items, got %d", m.GetListView().ItemCount())
	}
}

func TestModelHelpShowsClearFilters(t *testing.T) {
	v := vault.NewVault()
	m := NewModel(v)

	// Initialize
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	m.SetView(ViewHelp)
	view := m.View()

	if !contains(view, "Clear all filters") {
		t.Errorf("Help view should mention Clear all filters, got %q", view)
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
