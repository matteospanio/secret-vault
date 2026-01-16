package clipboard

import (
	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the given text to the system clipboard
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ClearClipboard clears the system clipboard by writing an empty string
func ClearClipboard() error {
	return clipboard.WriteAll("")
}

// GetClipboardContent retrieves the current content from the system clipboard
func GetClipboardContent() (string, error) {
	return clipboard.ReadAll()
}
