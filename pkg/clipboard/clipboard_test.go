package clipboard

import (
	"testing"
)

// Note: These tests require a working clipboard environment.
// On headless Linux servers, they may be skipped if no display is available.

func TestCopyToClipboard(t *testing.T) {
	testText := "test-clipboard-content-12345"

	err := CopyToClipboard(testText)
	if err != nil {
		t.Skipf("Clipboard not available (headless environment?): %v", err)
	}

	// Verify by reading back
	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent failed: %v", err)
	}

	if content != testText {
		t.Errorf("Expected clipboard content %q, got %q", testText, content)
	}
}

func TestClearClipboard(t *testing.T) {
	// First, set some content
	err := CopyToClipboard("content-to-clear")
	if err != nil {
		t.Skipf("Clipboard not available (headless environment?): %v", err)
	}

	// Clear the clipboard
	err = ClearClipboard()
	if err != nil {
		t.Fatalf("ClearClipboard failed: %v", err)
	}

	// Verify clipboard is empty
	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent failed: %v", err)
	}

	if content != "" {
		t.Errorf("Expected empty clipboard, got %q", content)
	}
}

func TestGetClipboardContent(t *testing.T) {
	testText := "get-clipboard-test-content"

	err := CopyToClipboard(testText)
	if err != nil {
		t.Skipf("Clipboard not available (headless environment?): %v", err)
	}

	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent failed: %v", err)
	}

	if content != testText {
		t.Errorf("Expected %q, got %q", testText, content)
	}
}

func TestCopyEmptyString(t *testing.T) {
	err := CopyToClipboard("")
	if err != nil {
		t.Skipf("Clipboard not available (headless environment?): %v", err)
	}

	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent failed: %v", err)
	}

	if content != "" {
		t.Errorf("Expected empty string, got %q", content)
	}
}

func TestCopyLargeText(t *testing.T) {
	// Create a large text string (10KB)
	largeText := make([]byte, 10*1024)
	for i := range largeText {
		largeText[i] = byte('A' + (i % 26))
	}
	testText := string(largeText)

	err := CopyToClipboard(testText)
	if err != nil {
		t.Skipf("Clipboard not available (headless environment?): %v", err)
	}

	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent failed: %v", err)
	}

	if content != testText {
		t.Errorf("Large text copy/paste failed: lengths differ (expected %d, got %d)", len(testText), len(content))
	}
}

func TestCopySpecialCharacters(t *testing.T) {
	testCases := []string{
		"hello\nworld",          // newlines
		"tab\there",             // tabs
		"unicode: \u00e9\u00e8", // accented chars
		"emoji: \U0001F600",     // emoji
		"quotes: \"'`",          // quotes
		"special: !@#$%^&*()[]", // special chars
	}

	for _, testText := range testCases {
		t.Run(testText[:min(10, len(testText))], func(t *testing.T) {
			err := CopyToClipboard(testText)
			if err != nil {
				t.Skipf("Clipboard not available: %v", err)
			}

			content, err := GetClipboardContent()
			if err != nil {
				t.Fatalf("GetClipboardContent failed: %v", err)
			}

			if content != testText {
				t.Errorf("Expected %q, got %q", testText, content)
			}
		})
	}
}
