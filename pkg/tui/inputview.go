package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// InputView provides a text input with autocomplete suggestions
type InputView struct {
	input           textinput.Model
	vault           *vault.Vault
	suggestions     []string
	selectedIdx     int
	maxSuggestions  int
	showSuggestions bool
	width           int
	height          int
	placeholder     string
	prompt          string
}

// NewInputView creates a new input view with autocomplete
func NewInputView(v *vault.Vault, placeholder, prompt string) *InputView {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	iv := &InputView{
		input:           ti,
		vault:           v,
		suggestions:     []string{},
		selectedIdx:     0,
		maxSuggestions:  5,
		showSuggestions: true,
		width:           80,
		height:          24,
		placeholder:     placeholder,
		prompt:          prompt,
	}

	// Initialize suggestions with all secrets
	iv.updateSuggestions()

	return iv
}

// Update handles messages for the input view
// Returns: command, whether a selection was made, selected value
func (iv *InputView) Update(msg tea.Msg) (tea.Cmd, bool, string) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			// Navigate suggestions up
			if iv.showSuggestions && len(iv.suggestions) > 0 {
				iv.selectedIdx--
				if iv.selectedIdx < 0 {
					iv.selectedIdx = len(iv.suggestions) - 1
				}
				return nil, false, ""
			}

		case "down":
			// Navigate suggestions down
			if iv.showSuggestions && len(iv.suggestions) > 0 {
				iv.selectedIdx++
				if iv.selectedIdx >= len(iv.suggestions) {
					iv.selectedIdx = 0
				}
				return nil, false, ""
			}

		case "tab":
			// Complete with selected suggestion
			if iv.showSuggestions && len(iv.suggestions) > 0 && iv.selectedIdx < len(iv.suggestions) {
				iv.input.SetValue(iv.suggestions[iv.selectedIdx])
				iv.input.CursorEnd()
				iv.updateSuggestions()
				return nil, false, ""
			}

		case "enter":
			// Select current suggestion or use input value
			var selected string
			if iv.showSuggestions && len(iv.suggestions) > 0 && iv.selectedIdx < len(iv.suggestions) {
				selected = iv.suggestions[iv.selectedIdx]
			} else {
				selected = iv.input.Value()
			}
			if selected != "" {
				return nil, true, selected
			}
			return nil, false, ""

		case "esc":
			// Clear input or cancel
			if iv.input.Value() != "" {
				iv.input.SetValue("")
				iv.updateSuggestions()
				return nil, false, ""
			}
			// Return empty to signal cancel (handled by parent)
			return nil, false, ""
		}

		// Pass other key messages to text input
		var cmd tea.Cmd
		iv.input, cmd = iv.input.Update(msg)

		// Update suggestions based on new input
		iv.updateSuggestions()

		return cmd, false, ""
	}

	// Pass non-key messages to text input
	var cmd tea.Cmd
	iv.input, cmd = iv.input.Update(msg)
	return cmd, false, ""
}

// updateSuggestions updates the suggestion list based on current input
func (iv *InputView) updateSuggestions() {
	if iv.vault == nil {
		iv.suggestions = []string{}
		return
	}

	query := iv.input.Value()
	allSecrets := iv.vault.ListSecrets()

	// Sort and filter by score
	iv.suggestions = SortByScore(allSecrets, query)

	// Limit to max suggestions
	if len(iv.suggestions) > iv.maxSuggestions {
		iv.suggestions = iv.suggestions[:iv.maxSuggestions]
	}

	// Reset selection if out of bounds
	if iv.selectedIdx >= len(iv.suggestions) {
		iv.selectedIdx = 0
	}
}

// View renders the input view with suggestions
func (iv *InputView) View() string {
	var b strings.Builder

	// Prompt
	if iv.prompt != "" {
		b.WriteString(titleStyle.Render(iv.prompt))
		b.WriteString("\n\n")
	}

	// Text input
	b.WriteString(iv.input.View())
	b.WriteString("\n")

	// Suggestions
	if iv.showSuggestions && len(iv.suggestions) > 0 {
		b.WriteString("\n")
		for i, suggestion := range iv.suggestions {
			if i == iv.selectedIdx {
				b.WriteString(listSelectedStyle.Render("> " + suggestion))
			} else {
				b.WriteString(listItemStyle.Render("  " + suggestion))
			}
			b.WriteString("\n")
		}
	} else if iv.showSuggestions && iv.input.Value() != "" && len(iv.suggestions) == 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  No matching secrets"))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • tab: complete • enter: select • esc: cancel"))

	return b.String()
}

// SetSize updates the view dimensions
func (iv *InputView) SetSize(width, height int) {
	iv.width = width
	iv.height = height
	iv.input.Width = width - 4
}

// Value returns the current input value
func (iv *InputView) Value() string {
	return iv.input.Value()
}

// SetValue sets the input value
func (iv *InputView) SetValue(value string) {
	iv.input.SetValue(value)
	iv.updateSuggestions()
}

// Focus focuses the text input
func (iv *InputView) Focus() tea.Cmd {
	return iv.input.Focus()
}

// Blur removes focus from the text input
func (iv *InputView) Blur() {
	iv.input.Blur()
}

// IsFocused returns whether the input is focused
func (iv *InputView) IsFocused() bool {
	return iv.input.Focused()
}

// Suggestions returns the current suggestions list
func (iv *InputView) Suggestions() []string {
	return iv.suggestions
}

// SelectedIndex returns the currently selected suggestion index
func (iv *InputView) SelectedIndex() int {
	return iv.selectedIdx
}

// SelectedSuggestion returns the currently selected suggestion, or empty string
func (iv *InputView) SelectedSuggestion() string {
	if iv.selectedIdx < len(iv.suggestions) {
		return iv.suggestions[iv.selectedIdx]
	}
	return ""
}

// SetMaxSuggestions sets the maximum number of suggestions to display
func (iv *InputView) SetMaxSuggestions(max int) {
	if max < 1 {
		max = 1
	}
	iv.maxSuggestions = max
	iv.updateSuggestions()
}

// SetShowSuggestions enables or disables showing suggestions
func (iv *InputView) SetShowSuggestions(show bool) {
	iv.showSuggestions = show
}

// SecretSelectedMsg is sent when a secret is selected via autocomplete
type SecretInputSelectedMsg struct {
	Name string
}

// secretInputCmd returns a command that sends a selection message
func secretInputCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return SecretInputSelectedMsg{Name: name}
	}
}

// Reset clears the input and resets to initial state
func (iv *InputView) Reset() {
	iv.input.SetValue("")
	iv.selectedIdx = 0
	iv.updateSuggestions()
}

// GetPrompt returns the current prompt text
func (iv *InputView) GetPrompt() string {
	return iv.prompt
}

// SetPrompt sets the prompt text
func (iv *InputView) SetPrompt(prompt string) {
	iv.prompt = prompt
}

// GetPlaceholder returns the current placeholder text
func (iv *InputView) GetPlaceholder() string {
	return iv.placeholder
}

// SetPlaceholder sets the placeholder text
func (iv *InputView) SetPlaceholder(placeholder string) {
	iv.placeholder = placeholder
	iv.input.Placeholder = placeholder
}

// HasSuggestions returns true if there are suggestions available
func (iv *InputView) HasSuggestions() bool {
	return len(iv.suggestions) > 0
}

// SuggestionCount returns the number of current suggestions
func (iv *InputView) SuggestionCount() int {
	return len(iv.suggestions)
}

// IsEmpty returns true if the input is empty
func (iv *InputView) IsEmpty() bool {
	return iv.input.Value() == ""
}

// CreateForSecretSearch creates an InputView configured for secret search
func CreateForSecretSearch(v *vault.Vault) *InputView {
	return NewInputView(v, "Type to search secrets...", "Search Secrets")
}

// CreateForSecretSelection creates an InputView configured for secret selection
func CreateForSecretSelection(v *vault.Vault) *InputView {
	return NewInputView(v, "Enter secret name...", "Select Secret")
}

// Ensure the unused import doesn't cause issues
var _ = fmt.Sprintf
