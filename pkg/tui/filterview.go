package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

const (
	filterFieldSearch = iota
	filterFieldCategory
	filterFieldTag
	filterFieldCount
)

// FilterCriteria holds the active filter settings
type FilterCriteria struct {
	Query    string
	Category string
	Tag      string
	OldOnly  bool
}

// IsEmpty returns true if no filters are active
func (fc FilterCriteria) IsEmpty() bool {
	return fc.Query == "" && fc.Category == "" && fc.Tag == "" && !fc.OldOnly
}

// FilterView provides a panel for filtering secrets by various criteria
type FilterView struct {
	inputs     []textinput.Model
	focusIndex int
	oldOnly    bool
	vault      *vault.Vault
	width      int
	height     int
}

// NewFilterView creates a new filter view
func NewFilterView(v *vault.Vault, width, height int) *FilterView {
	inputs := make([]textinput.Model, filterFieldCount)

	// Search query input
	inputs[filterFieldSearch] = textinput.New()
	inputs[filterFieldSearch].Placeholder = "Search by name or description..."
	inputs[filterFieldSearch].CharLimit = 256
	inputs[filterFieldSearch].Width = 40

	// Category input
	inputs[filterFieldCategory] = textinput.New()
	inputs[filterFieldCategory].Placeholder = "Filter by category..."
	inputs[filterFieldCategory].CharLimit = 128
	inputs[filterFieldCategory].Width = 40

	// Tag input
	inputs[filterFieldTag] = textinput.New()
	inputs[filterFieldTag].Placeholder = "Filter by tag..."
	inputs[filterFieldTag].CharLimit = 128
	inputs[filterFieldTag].Width = 40

	// Focus the first field
	inputs[filterFieldSearch].Focus()

	return &FilterView{
		inputs:     inputs,
		focusIndex: filterFieldSearch,
		oldOnly:    false,
		vault:      v,
		width:      width,
		height:     height,
	}
}

// Update handles key messages for the filter view
// Returns: command to execute, whether the message was handled
func (fv *FilterView) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			fv.nextField()
			return nil, true

		case "shift+tab", "up":
			fv.prevField()
			return nil, true

		case "ctrl+o":
			// Toggle old-only filter
			fv.oldOnly = !fv.oldOnly
			return nil, true

		case "enter":
			// Apply filters
			return fv.applyFilters(), true

		case "ctrl+r":
			// Clear all filters
			fv.clearAll()
			return nil, true
		}

		// Pass other key messages to the focused input
		var cmd tea.Cmd
		fv.inputs[fv.focusIndex], cmd = fv.inputs[fv.focusIndex].Update(msg)
		return cmd, true
	}
	return nil, false
}

// applyFilters creates and returns a command with the current filter criteria
func (fv *FilterView) applyFilters() tea.Cmd {
	criteria := fv.GetCriteria()
	return func() tea.Msg {
		return FilterAppliedMsg{
			Category: criteria.Category,
			Tag:      criteria.Tag,
			Query:    criteria.Query,
			OldOnly:  criteria.OldOnly,
		}
	}
}

// clearAll resets all filter fields
func (fv *FilterView) clearAll() {
	for i := range fv.inputs {
		fv.inputs[i].SetValue("")
	}
	fv.oldOnly = false
}

// GetCriteria returns the current filter criteria
func (fv *FilterView) GetCriteria() FilterCriteria {
	return FilterCriteria{
		Query:    strings.TrimSpace(fv.inputs[filterFieldSearch].Value()),
		Category: strings.TrimSpace(fv.inputs[filterFieldCategory].Value()),
		Tag:      strings.TrimSpace(fv.inputs[filterFieldTag].Value()),
		OldOnly:  fv.oldOnly,
	}
}

// ApplyToVault applies the current filter criteria to the vault and returns matching secrets
func (fv *FilterView) ApplyToVault() []vault.Secret {
	criteria := fv.GetCriteria()
	return ApplyFilterCriteria(fv.vault, criteria)
}

// ApplyFilterCriteria applies the given criteria to the vault and returns matching secrets
func ApplyFilterCriteria(v *vault.Vault, criteria FilterCriteria) []vault.Secret {
	// Start with all secrets
	var results []vault.Secret
	if criteria.Query != "" {
		results = v.Search(criteria.Query)
	} else {
		for _, s := range v.Secrets {
			results = append(results, s)
		}
	}

	// Filter by category
	if criteria.Category != "" {
		categoryLower := strings.ToLower(criteria.Category)
		filtered := make([]vault.Secret, 0)
		for _, s := range results {
			if strings.ToLower(s.Category) == categoryLower {
				filtered = append(filtered, s)
			}
		}
		results = filtered
	}

	// Filter by tag
	if criteria.Tag != "" {
		tagLower := strings.ToLower(criteria.Tag)
		filtered := make([]vault.Secret, 0)
		for _, s := range results {
			for _, t := range s.Tags {
				if strings.ToLower(t) == tagLower {
					filtered = append(filtered, s)
					break
				}
			}
		}
		results = filtered
	}

	// Filter by age
	if criteria.OldOnly {
		filtered := make([]vault.Secret, 0)
		for _, s := range results {
			if s.IsOld(vault.DefaultAgeThreshold) {
				filtered = append(filtered, s)
			}
		}
		results = filtered
	}

	// Sort by name for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// View renders the filter view
func (fv *FilterView) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Filter Secrets"))
	b.WriteString("\n\n")

	labels := []string{"Search:", "Category:", "Tag:"}

	for i, label := range labels {
		if i == fv.focusIndex {
			b.WriteString(subtitleStyle.Render(label))
		} else {
			b.WriteString(dimStyle.Render(label))
		}
		b.WriteString("\n")
		b.WriteString(fv.inputs[i].View())
		b.WriteString("\n\n")
	}

	// Old-only toggle
	oldCheck := "[ ]"
	if fv.oldOnly {
		oldCheck = "[x]"
	}
	oldLabel := fmt.Sprintf("%s Show old secrets only (>1 year)", oldCheck)
	b.WriteString(dimStyle.Render(oldLabel))
	b.WriteString("\n\n")

	// Active filters summary
	criteria := fv.GetCriteria()
	if !criteria.IsEmpty() {
		b.WriteString(subtitleStyle.Render("Active filters: "))
		var parts []string
		if criteria.Query != "" {
			parts = append(parts, fmt.Sprintf("search=%q", criteria.Query))
		}
		if criteria.Category != "" {
			parts = append(parts, fmt.Sprintf("category=%q", criteria.Category))
		}
		if criteria.Tag != "" {
			parts = append(parts, fmt.Sprintf("tag=%q", criteria.Tag))
		}
		if criteria.OldOnly {
			parts = append(parts, "old only")
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n\n")
	}

	// Help footer
	b.WriteString(helpStyle.Render("tab/↓: next • shift+tab/↑: prev • ctrl+o: toggle old • enter: apply • ctrl+r: clear • esc: cancel"))

	return "\n" + b.String()
}

// nextField moves focus to the next field
func (fv *FilterView) nextField() {
	fv.inputs[fv.focusIndex].Blur()
	fv.focusIndex = (fv.focusIndex + 1) % filterFieldCount
	fv.inputs[fv.focusIndex].Focus()
}

// prevField moves focus to the previous field
func (fv *FilterView) prevField() {
	fv.inputs[fv.focusIndex].Blur()
	fv.focusIndex = (fv.focusIndex - 1 + filterFieldCount) % filterFieldCount
	fv.inputs[fv.focusIndex].Focus()
}

// SetSize updates the view dimensions
func (fv *FilterView) SetSize(width, height int) {
	fv.width = width
	fv.height = height
	for i := range fv.inputs {
		fv.inputs[i].Width = width - 4
	}
}

// SetCriteria pre-fills the filter view with existing criteria
func (fv *FilterView) SetCriteria(criteria FilterCriteria) {
	fv.inputs[filterFieldSearch].SetValue(criteria.Query)
	fv.inputs[filterFieldCategory].SetValue(criteria.Category)
	fv.inputs[filterFieldTag].SetValue(criteria.Tag)
	fv.oldOnly = criteria.OldOnly
}

// FocusIndex returns the currently focused field index
func (fv *FilterView) FocusIndex() int {
	return fv.focusIndex
}

// IsOldOnly returns whether the old-only filter is active
func (fv *FilterView) IsOldOnly() bool {
	return fv.oldOnly
}

// SetOldOnly sets the old-only filter
func (fv *FilterView) SetOldOnly(old bool) {
	fv.oldOnly = old
}

// GetSearchQuery returns the search query value
func (fv *FilterView) GetSearchQuery() string {
	return fv.inputs[filterFieldSearch].Value()
}

// GetCategoryFilter returns the category filter value
func (fv *FilterView) GetCategoryFilter() string {
	return fv.inputs[filterFieldCategory].Value()
}

// GetTagFilter returns the tag filter value
func (fv *FilterView) GetTagFilter() string {
	return fv.inputs[filterFieldTag].Value()
}
