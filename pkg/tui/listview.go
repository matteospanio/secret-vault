package tui

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matteospanio/secret-vault-cli/pkg/vault"
)

// SecretItem represents a secret in the list view
type SecretItem struct {
	secret *vault.Secret
}

// FilterValue implements list.Item
func (i SecretItem) FilterValue() string {
	return i.secret.Name
}

// Title returns the secret name
func (i SecretItem) Title() string {
	return i.secret.Name
}

// Description returns the secret description or a placeholder
func (i SecretItem) Description() string {
	if i.secret.Description != "" {
		return i.secret.Description
	}
	return "No description"
}

// Secret returns the underlying secret
func (i SecretItem) Secret() *vault.Secret {
	return i.secret
}

// secretItemDelegate handles rendering of list items
type secretItemDelegate struct{}

func (d secretItemDelegate) Height() int                             { return 2 }
func (d secretItemDelegate) Spacing() int                            { return 1 }
func (d secretItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d secretItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(SecretItem)
	if !ok {
		return
	}

	// Build the item display
	var titleStr string
	var descStr string

	if index == m.Index() {
		// Selected item
		titleStr = listSelectedStyle.Render("> " + item.Title())
		descStr = "  " + dimStyle.Render(item.Description())
	} else {
		titleStr = listItemStyle.Render("  " + item.Title())
		descStr = "  " + dimStyle.Render(item.Description())
	}

	// Add age warning indicator
	if item.secret.IsOld(vault.DefaultAgeThreshold) {
		titleStr += " " + warningStyle.Render("⚠")
	}

	fmt.Fprintf(w, "%s\n%s", titleStr, descStr)
}

// ListView wraps bubbles/list for displaying secrets
type ListView struct {
	list   list.Model
	vault  *vault.Vault
	width  int
	height int
}

// NewListView creates a new list view with secrets from the vault
func NewListView(v *vault.Vault, width, height int) *ListView {
	items := secretsToItems(v)

	// Create the list model
	delegate := secretItemDelegate{}
	l := list.New(items, delegate, width, height)
	l.Title = "Secrets"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We use our own help

	// Style the list
	l.Styles.Title = titleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 0)

	return &ListView{
		list:   l,
		vault:  v,
		width:  width,
		height: height,
	}
}

// secretsToItems converts vault secrets to list items
func secretsToItems(v *vault.Vault) []list.Item {
	names := v.ListSecrets()
	sort.Strings(names)

	items := make([]list.Item, 0, len(names))
	for _, name := range names {
		secret, err := v.GetSecret(name)
		if err != nil {
			continue
		}
		// Create a copy to avoid pointer issues
		secretCopy := secret
		items = append(items, SecretItem{secret: &secretCopy})
	}
	return items
}

// Update handles messages for the list view
func (lv *ListView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	lv.list, cmd = lv.list.Update(msg)
	return cmd
}

// View renders the list view
func (lv *ListView) View() string {
	if len(lv.list.Items()) == 0 {
		return lv.renderEmpty()
	}
	return lv.list.View()
}

// renderEmpty renders the empty vault message
func (lv *ListView) renderEmpty() string {
	title := titleStyle.Render("Secret Vault")
	subtitle := subtitleStyle.Render("Secure token and secret storage")
	empty := dimStyle.Render("No secrets stored yet")
	hint := helpStyle.Render("Use 'secretvault add <name>' to add your first secret")

	return fmt.Sprintf("\n%s\n%s\n\n%s\n\n%s", title, subtitle, empty, hint)
}

// SetSize updates the list dimensions
func (lv *ListView) SetSize(width, height int) {
	lv.width = width
	lv.height = height
	lv.list.SetSize(width, height)
}

// SelectedItem returns the currently selected item
func (lv *ListView) SelectedItem() *SecretItem {
	if len(lv.list.Items()) == 0 {
		return nil
	}
	item, ok := lv.list.SelectedItem().(SecretItem)
	if !ok {
		return nil
	}
	return &item
}

// SelectedSecret returns the secret of the currently selected item
func (lv *ListView) SelectedSecret() *vault.Secret {
	item := lv.SelectedItem()
	if item == nil {
		return nil
	}
	return item.secret
}

// Refresh reloads the secrets from the vault
func (lv *ListView) Refresh() {
	items := secretsToItems(lv.vault)
	lv.list.SetItems(items)
}

// SetFilteredItems sets the list items from a pre-filtered slice of secrets
func (lv *ListView) SetFilteredItems(secrets []vault.Secret) {
	items := make([]list.Item, 0, len(secrets))
	for i := range secrets {
		s := secrets[i]
		items = append(items, SecretItem{secret: &s})
	}
	lv.list.SetItems(items)
}

// ItemCount returns the number of items in the list
func (lv *ListView) ItemCount() int {
	return len(lv.list.Items())
}

// FilterState returns the current filter state
func (lv *ListView) FilterState() list.FilterState {
	return lv.list.FilterState()
}

// IsFiltering returns true if the list is currently in filtering mode
func (lv *ListView) IsFiltering() bool {
	return lv.list.FilterState() == list.Filtering
}
