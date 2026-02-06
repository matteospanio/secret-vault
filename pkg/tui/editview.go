package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

const (
	editFieldName = iota
	editFieldValue
	editFieldDescription
	editFieldCategory
	editFieldTags
	editFieldCount
)

// SecretSavedMsg is sent when a secret is saved from the edit view
type SecretSavedMsg struct {
	Name string
}

// EditView provides a form for adding/editing secrets with category and tag fields
type EditView struct {
	inputs     []textinput.Model
	focusIndex int
	vault      *vault.Vault
	editing    *vault.Secret // nil if adding new secret
	width      int
	height     int
	err        string
}

// NewEditView creates a new edit view. If secret is non-nil, it's an edit; otherwise it's a new secret.
func NewEditView(v *vault.Vault, secret *vault.Secret, width, height int) *EditView {
	inputs := make([]textinput.Model, editFieldCount)

	// Name field
	inputs[editFieldName] = textinput.New()
	inputs[editFieldName].Placeholder = "Secret name (required)"
	inputs[editFieldName].CharLimit = 256
	inputs[editFieldName].Width = 40

	// Value field
	inputs[editFieldValue] = textinput.New()
	inputs[editFieldValue].Placeholder = "Secret value (required)"
	inputs[editFieldValue].CharLimit = 4096
	inputs[editFieldValue].Width = 40
	inputs[editFieldValue].EchoMode = textinput.EchoPassword

	// Description field
	inputs[editFieldDescription] = textinput.New()
	inputs[editFieldDescription].Placeholder = "Description (optional)"
	inputs[editFieldDescription].CharLimit = 512
	inputs[editFieldDescription].Width = 40

	// Category field
	inputs[editFieldCategory] = textinput.New()
	inputs[editFieldCategory].Placeholder = "Category (optional)"
	inputs[editFieldCategory].CharLimit = 128
	inputs[editFieldCategory].Width = 40

	// Tags field
	inputs[editFieldTags] = textinput.New()
	inputs[editFieldTags].Placeholder = "Tags, comma-separated (optional)"
	inputs[editFieldTags].CharLimit = 512
	inputs[editFieldTags].Width = 40

	ev := &EditView{
		inputs:     inputs,
		focusIndex: editFieldName,
		vault:      v,
		editing:    secret,
		width:      width,
		height:     height,
	}

	// Pre-fill fields when editing an existing secret
	if secret != nil {
		inputs[editFieldName].SetValue(secret.Name)
		inputs[editFieldValue].SetValue(secret.Value)
		inputs[editFieldDescription].SetValue(secret.Description)
		inputs[editFieldCategory].SetValue(secret.Category)
		if len(secret.Tags) > 0 {
			inputs[editFieldTags].SetValue(strings.Join(secret.Tags, ", "))
		}
	}

	// Focus the first field
	inputs[editFieldName].Focus()

	return ev
}

// Update handles key messages for the edit view
// Returns: command to execute, whether the message was handled
func (ev *EditView) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			ev.nextField()
			return nil, true

		case "shift+tab", "up":
			ev.prevField()
			return nil, true

		case "ctrl+s":
			// Validate and save
			cmd := ev.save()
			return cmd, true

		case "enter":
			// On the last field, enter saves; otherwise move to next field
			if ev.focusIndex == editFieldTags {
				cmd := ev.save()
				return cmd, true
			}
			ev.nextField()
			return nil, true
		}

		// Pass other key messages to the focused input
		var cmd tea.Cmd
		ev.inputs[ev.focusIndex], cmd = ev.inputs[ev.focusIndex].Update(msg)
		return cmd, true
	}
	return nil, false
}

// save validates and saves the secret
func (ev *EditView) save() tea.Cmd {
	name := strings.TrimSpace(ev.inputs[editFieldName].Value())
	value := strings.TrimSpace(ev.inputs[editFieldValue].Value())

	// Validate required fields
	if name == "" {
		ev.err = "Name is required"
		return nil
	}
	if value == "" {
		ev.err = "Value is required"
		return nil
	}

	description := strings.TrimSpace(ev.inputs[editFieldDescription].Value())
	category := strings.TrimSpace(ev.inputs[editFieldCategory].Value())
	tags := parseTags(ev.inputs[editFieldTags].Value())

	ev.vault.AddSecret(name, value, description, category, tags)
	ev.err = ""

	return func() tea.Msg {
		return SecretSavedMsg{Name: name}
	}
}

// parseTags splits a comma-separated string into a trimmed tag slice
func parseTags(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

// View renders the edit view
func (ev *EditView) View() string {
	var b strings.Builder

	// Title
	if ev.editing != nil {
		b.WriteString(titleStyle.Render("Edit Secret"))
	} else {
		b.WriteString(titleStyle.Render("Add Secret"))
	}
	b.WriteString("\n\n")

	labels := []string{"Name:", "Value:", "Description:", "Category:", "Tags:"}

	for i, label := range labels {
		// Highlight the focused label
		if i == ev.focusIndex {
			b.WriteString(subtitleStyle.Render(label))
		} else {
			b.WriteString(dimStyle.Render(label))
		}
		b.WriteString("\n")
		b.WriteString(ev.inputs[i].View())
		b.WriteString("\n\n")
	}

	// Show validation error
	if ev.err != "" {
		b.WriteString(errorStyle.Render("Error: " + ev.err))
		b.WriteString("\n\n")
	}

	// Help footer
	b.WriteString(helpStyle.Render("tab/↓: next field • shift+tab/↑: prev field • ctrl+s: save • esc: cancel"))

	return "\n" + b.String()
}

// nextField moves focus to the next field
func (ev *EditView) nextField() {
	ev.inputs[ev.focusIndex].Blur()
	ev.focusIndex = (ev.focusIndex + 1) % editFieldCount
	ev.inputs[ev.focusIndex].Focus()
}

// prevField moves focus to the previous field
func (ev *EditView) prevField() {
	ev.inputs[ev.focusIndex].Blur()
	ev.focusIndex = (ev.focusIndex - 1 + editFieldCount) % editFieldCount
	ev.inputs[ev.focusIndex].Focus()
}

// SetSize updates the view dimensions
func (ev *EditView) SetSize(width, height int) {
	ev.width = width
	ev.height = height
	for i := range ev.inputs {
		ev.inputs[i].Width = width - 4
	}
}

// IsEditing returns true if the view is editing an existing secret
func (ev *EditView) IsEditing() bool {
	return ev.editing != nil
}

// FocusIndex returns the currently focused field index
func (ev *EditView) FocusIndex() int {
	return ev.focusIndex
}

// GetError returns the current validation error message
func (ev *EditView) GetError() string {
	return ev.err
}

// GetName returns the current name field value
func (ev *EditView) GetName() string {
	return ev.inputs[editFieldName].Value()
}

// GetValue returns the current value field value
func (ev *EditView) GetValue() string {
	return ev.inputs[editFieldValue].Value()
}

// GetDescription returns the current description field value
func (ev *EditView) GetDescription() string {
	return ev.inputs[editFieldDescription].Value()
}

// GetCategory returns the current category field value
func (ev *EditView) GetCategory() string {
	return ev.inputs[editFieldCategory].Value()
}

// GetTagsRaw returns the raw tags field value
func (ev *EditView) GetTagsRaw() string {
	return ev.inputs[editFieldTags].Value()
}

// GetParsedTags returns the parsed tags
func (ev *EditView) GetParsedTags() []string {
	return parseTags(ev.inputs[editFieldTags].Value())
}
