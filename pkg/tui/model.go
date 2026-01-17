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
	vault       *vault.Vault
	currentView ViewType
	width       int
	height      int
	ready       bool
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	return m.renderWelcome()
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
