package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// SearchBar provides a persistent search input for real-time filtering
type SearchBar struct {
	input  textinput.Model
	vault  *vault.Vault
	active bool
	width  int
}

// NewSearchBar creates a new search bar
func NewSearchBar(v *vault.Vault, width int) *SearchBar {
	input := textinput.New()
	input.Placeholder = "Type to search secrets..."
	input.CharLimit = 256
	input.Width = width - 12
	input.Prompt = "/ "

	return &SearchBar{
		input: input,
		vault: v,
		width: width,
	}
}

// Focus activates the search bar
func (sb *SearchBar) Focus() {
	sb.active = true
	sb.input.Focus()
}

// Blur deactivates the search bar
func (sb *SearchBar) Blur() {
	sb.active = false
	sb.input.Blur()
}

// IsActive returns whether the search bar is focused
func (sb *SearchBar) IsActive() bool {
	return sb.active
}

// Query returns the current search query
func (sb *SearchBar) Query() string {
	return sb.input.Value()
}

// SetQuery sets the search query
func (sb *SearchBar) SetQuery(query string) {
	sb.input.SetValue(query)
}

// Clear resets the search query
func (sb *SearchBar) Clear() {
	sb.input.SetValue("")
}

// Update handles key messages for the search bar
// Returns: command to execute, whether the input changed
func (sb *SearchBar) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !sb.active {
		return nil, false
	}

	prevValue := sb.input.Value()
	var cmd tea.Cmd
	sb.input, cmd = sb.input.Update(msg)
	changed := sb.input.Value() != prevValue

	return cmd, changed
}

// View renders the search bar
func (sb *SearchBar) View() string {
	if sb.active {
		return sb.input.View()
	}
	query := sb.input.Value()
	if query != "" {
		return dimStyle.Render("/ ") + query
	}
	return dimStyle.Render("/ Type to search (press / to focus)")
}

// SetSize updates the search bar width
func (sb *SearchBar) SetSize(width int) {
	sb.width = width
	sb.input.Width = width - 12
}

// FilterSecretsByQuery applies fuzzy search against vault secrets and returns matching ones
func (sb *SearchBar) FilterSecretsByQuery(secrets []vault.Secret) []vault.Secret {
	query := strings.TrimSpace(sb.input.Value())
	if query == "" {
		return secrets
	}

	var results []vault.Secret
	for _, s := range secrets {
		// Match against name and description using fuzzy matching
		if FuzzyMatch(query, s.Name) || FuzzyMatch(query, s.Description) {
			results = append(results, s)
		}
	}

	// Sort by score (best matches first)
	sort.Slice(results, func(i, j int) bool {
		scoreI := FuzzyScore(query, results[i].Name)
		scoreJ := FuzzyScore(query, results[j].Name)
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		return results[i].Name < results[j].Name
	})

	return results
}
