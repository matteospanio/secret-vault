package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault/pkg/clipboard"
	"github.com/matteospanio/secret-vault/pkg/vault"
)

// DetailView displays secret details with reveal/copy functionality
type DetailView struct {
	secret   *vault.Secret
	revealed bool
	width    int
	height   int
}

// NewDetailView creates a new detail view for the given secret
func NewDetailView(secret *vault.Secret, width, height int) *DetailView {
	return &DetailView{
		secret:   secret,
		revealed: false,
		width:    width,
		height:   height,
	}
}

// Update handles key messages for the detail view
// Returns: command to execute, whether the message was handled
func (dv *DetailView) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			dv.ToggleReveal()
			return nil, true
		case "c":
			return copySecretCmd(dv.secret), true
		case "e":
			return func() tea.Msg {
				return ViewChangeMsg{View: ViewEdit}
			}, true
		}
	}
	return nil, false
}

// copySecretCmd returns a command that copies the secret value to clipboard
func copySecretCmd(secret *vault.Secret) tea.Cmd {
	return func() tea.Msg {
		if secret == nil {
			return ErrorMsg{Err: fmt.Errorf("no secret to copy")}
		}
		err := clipboard.CopyToClipboard(secret.Value)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}
		return SecretCopiedMsg{Name: secret.Name}
	}
}

// View renders the detail view
func (dv *DetailView) View() string {
	if dv.secret == nil {
		return dimStyle.Render("No secret selected")
	}

	s := dv.secret
	var lines []string

	// Title (secret name)
	lines = append(lines, titleStyle.Render(s.Name))
	lines = append(lines, "")

	// Description (if present)
	if s.Description != "" {
		lines = append(lines, subtitleStyle.Render(s.Description))
		lines = append(lines, "")
	}

	// Value (masked or revealed)
	valueDisplay := "********"
	if dv.revealed {
		valueDisplay = s.Value
	}
	lines = append(lines, dimStyle.Render("Value: ")+valueDisplay)

	// Category (if present)
	if s.Category != "" {
		lines = append(lines, dimStyle.Render("Category: ")+s.Category)
	}

	// Tags (if present)
	if len(s.Tags) > 0 {
		tags := strings.Join(s.Tags, ", ")
		lines = append(lines, dimStyle.Render("Tags: ")+tags)
	}

	lines = append(lines, "")

	// Timestamps
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Created: %s", s.CreatedAt.Format("2006-01-02 15:04"))))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Updated: %s", s.UpdatedAt.Format("2006-01-02 15:04"))))

	// Age information
	age := s.GetAge()
	ageStr := formatAge(age)
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Age: %s", ageStr)))

	// Age warning if old
	if s.IsOld(vault.DefaultAgeThreshold) {
		lines = append(lines, "")
		lines = append(lines, warningStyle.Render("⚠ This secret is over 1 year old"))
	}

	// Help footer
	lines = append(lines, "")
	revealText := "r: reveal"
	if dv.revealed {
		revealText = "r: hide"
	}
	lines = append(lines, helpStyle.Render(fmt.Sprintf("%s • c: copy • e: edit • esc: back • q: quit", revealText)))

	// Join with newlines
	return "\n" + strings.Join(lines, "\n")
}

// formatAge returns a human-readable age string
func formatAge(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days < 1 {
		return "less than a day"
	}
	if days == 1 {
		return "1 day"
	}
	if days < 30 {
		return fmt.Sprintf("%d days", days)
	}
	months := days / 30
	if months == 1 {
		return "1 month"
	}
	if months < 12 {
		return fmt.Sprintf("%d months", months)
	}
	years := months / 12
	if years == 1 {
		return "1 year"
	}
	return fmt.Sprintf("%d years", years)
}

// SetSize updates the view dimensions
func (dv *DetailView) SetSize(width, height int) {
	dv.width = width
	dv.height = height
}

// SetSecret updates the displayed secret and resets revealed state
func (dv *DetailView) SetSecret(secret *vault.Secret) {
	dv.secret = secret
	dv.revealed = false
}

// GetSecret returns the current secret
func (dv *DetailView) GetSecret() *vault.Secret {
	return dv.secret
}

// IsRevealed returns whether the value is currently revealed
func (dv *DetailView) IsRevealed() bool {
	return dv.revealed
}

// ToggleReveal toggles the revealed state
func (dv *DetailView) ToggleReveal() {
	dv.revealed = !dv.revealed
}
