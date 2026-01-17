package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// ViewType represents the current view being displayed
type ViewType int

const (
	ViewList ViewType = iota
	ViewDetail
	ViewEdit
	ViewFilter
	ViewHelp
)

// Model is the main TUI model that implements tea.Model
type Model struct {
	vault          *vault.Vault
	listView       *ListView
	currentView    ViewType
	previousView   ViewType
	selectedSecret *vault.Secret
	statusMessage  string
	statusIsError  bool
	width          int
	height         int
	ready          bool
}

// NewModel creates a new TUI model with the given vault
func NewModel(v *vault.Vault) Model {
	return Model{
		vault:       v,
		currentView: ViewList,
		ready:       false,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear status message on any key press
		m.ClearStatus()

		// Handle global keys first
		switch msg.String() {
		case "q", "ctrl+c":
			// Don't quit if filtering in list view
			if m.currentView == ViewList && m.listView != nil && m.listView.IsFiltering() {
				break
			}
			return m, tea.Quit

		case "?":
			// Don't show help if filtering
			if m.currentView == ViewList && m.listView != nil && m.listView.IsFiltering() {
				break
			}
			// Toggle help view
			if m.currentView == ViewHelp {
				m.GoBack()
			} else {
				m.SetView(ViewHelp)
			}
			return m, nil

		case "esc":
			// If filtering, let the list handle it
			if m.currentView == ViewList && m.listView != nil && m.listView.IsFiltering() {
				break
			}
			// Go back from current view
			if m.currentView != ViewList {
				m.GoBack()
			}
			return m, nil

		case "enter":
			// Handle selection in list view
			if m.currentView == ViewList && m.listView != nil && !m.listView.IsFiltering() {
				if secret := m.listView.SelectedSecret(); secret != nil {
					m.SelectSecret(secret)
					m.SetView(ViewDetail)
					return m, nil
				}
			}
		}

		// Pass key messages to list view when in list mode
		if m.currentView == ViewList && m.listView != nil {
			cmd := m.listView.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Initialize or resize list view
		if m.listView == nil {
			m.listView = NewListView(m.vault, msg.Width, msg.Height-4)
		} else {
			m.listView.SetSize(msg.Width, msg.Height-4)
		}

		// Pass window size to list view
		if m.listView != nil {
			cmd := m.listView.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ViewChangeMsg:
		m.SetView(msg.View)

	case SecretSelectedMsg:
		m.SelectSecret(msg.Secret)
		m.SetView(ViewDetail)

	case ErrorMsg:
		if msg.Err != nil {
			m.SetStatus(msg.Err.Error(), true)
		}

	case SecretCopiedMsg:
		m.SetStatus(fmt.Sprintf("Secret '%s' copied to clipboard", msg.Name), false)

	case ClipboardClearedMsg:
		m.SetStatus("Clipboard cleared", false)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var content string

	switch m.currentView {
	case ViewList:
		if m.listView != nil {
			content = m.listView.View()
		} else {
			content = m.renderWelcome()
		}
	case ViewDetail:
		content = m.renderDetail()
	case ViewHelp:
		content = m.renderHelp()
	default:
		content = m.renderWelcome()
	}

	// Append status message if present
	if m.statusMessage != "" {
		var statusStyle = successStyle
		if m.statusIsError {
			statusStyle = errorStyle
		}
		content += "\n\n" + statusStyle.Render(m.statusMessage)
	}

	return content
}

// renderDetail renders the secret detail view
func (m Model) renderDetail() string {
	if m.selectedSecret == nil {
		return dimStyle.Render("No secret selected")
	}

	s := m.selectedSecret
	title := titleStyle.Render(s.Name)

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	if s.Description != "" {
		lines = append(lines, subtitleStyle.Render(s.Description))
		lines = append(lines, "")
	}

	lines = append(lines, dimStyle.Render("Value: ")+"********")

	if s.Category != "" {
		lines = append(lines, dimStyle.Render("Category: ")+s.Category)
	}

	if len(s.Tags) > 0 {
		tags := ""
		for i, tag := range s.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += tag
		}
		lines = append(lines, dimStyle.Render("Tags: ")+tags)
	}

	lines = append(lines, "")
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Created: %s", s.CreatedAt.Format("2006-01-02 15:04"))))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Updated: %s", s.UpdatedAt.Format("2006-01-02 15:04"))))

	// Check if secret is old
	if s.IsOld(vault.DefaultAgeThreshold) {
		lines = append(lines, "")
		lines = append(lines, warningStyle.Render("⚠ This secret is over 1 year old"))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("r: reveal • c: copy • esc: back • q: quit"))

	result := ""
	for _, line := range lines {
		result += "\n" + line
	}
	return result
}

// renderHelp renders the help view
func (m Model) renderHelp() string {
	title := titleStyle.Render("Keyboard Shortcuts")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"?", "Toggle this help"},
		{"q", "Quit"},
		{"esc", "Go back"},
		{"↑/↓", "Navigate list"},
		{"enter", "Select item"},
		{"c", "Copy secret to clipboard"},
		{"C", "Clear clipboard"},
		{"r", "Reveal/hide secret value"},
		{"/", "Search"},
		{"f", "Filter"},
	}

	var lines []string
	lines = append(lines, "", title, "")

	for _, s := range shortcuts {
		lines = append(lines, fmt.Sprintf("  %s  %s", dimStyle.Render(fmt.Sprintf("%6s", s.key)), s.desc))
	}

	lines = append(lines, "", helpStyle.Render("Press ? or esc to close"))

	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

// renderWelcome renders the welcome screen with vault statistics
func (m Model) renderWelcome() string {
	secretCount := len(m.vault.ListSecrets())

	title := titleStyle.Render("Secret Vault")
	subtitle := subtitleStyle.Render("Secure token and secret storage")

	var stats string
	if secretCount == 0 {
		stats = dimStyle.Render("No secrets stored yet")
	} else if secretCount == 1 {
		stats = dimStyle.Render("1 secret stored")
	} else {
		stats = dimStyle.Render(fmt.Sprintf("%d secrets stored", secretCount))
	}

	help := helpStyle.Render("Press q to quit")

	return fmt.Sprintf("\n%s\n%s\n\n%s\n\n%s", title, subtitle, stats, help)
}

// GetVault returns the vault instance
func (m Model) GetVault() *vault.Vault {
	return m.vault
}

// GetCurrentView returns the current view type
func (m Model) GetCurrentView() ViewType {
	return m.currentView
}

// IsReady returns whether the model is ready (received window size)
func (m Model) IsReady() bool {
	return m.ready
}

// SetView transitions to a new view, saving the previous view
func (m *Model) SetView(view ViewType) {
	m.previousView = m.currentView
	m.currentView = view
}

// GoBack returns to the previous view
func (m *Model) GoBack() {
	m.currentView = m.previousView
}

// SetStatus sets a status message to display
func (m *Model) SetStatus(message string, isError bool) {
	m.statusMessage = message
	m.statusIsError = isError
}

// ClearStatus clears the status message
func (m *Model) ClearStatus() {
	m.statusMessage = ""
	m.statusIsError = false
}

// GetStatusMessage returns the current status message
func (m Model) GetStatusMessage() string {
	return m.statusMessage
}

// IsStatusError returns whether the status message is an error
func (m Model) IsStatusError() bool {
	return m.statusIsError
}

// SelectSecret sets the currently selected secret
func (m *Model) SelectSecret(secret *vault.Secret) {
	m.selectedSecret = secret
}

// GetSelectedSecret returns the currently selected secret
func (m Model) GetSelectedSecret() *vault.Secret {
	return m.selectedSecret
}

// GetWidth returns the terminal width
func (m Model) GetWidth() int {
	return m.width
}

// GetHeight returns the terminal height
func (m Model) GetHeight() int {
	return m.height
}

// GetListView returns the list view instance
func (m Model) GetListView() *ListView {
	return m.listView
}
