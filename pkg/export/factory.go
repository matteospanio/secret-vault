package export

import (
	"fmt"
	"strings"
)

// GetFormatter returns an ExportFormat implementation based on the format name
func GetFormatter(format string) (ExportFormat, error) {
	switch strings.ToLower(format) {
	case "json":
		return NewJSONFormat(), nil
	case "yaml", "yml":
		return NewYAMLFormat(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported formats: json, yaml)", format)
	}
}

// SupportedFormats returns a list of supported format names
func SupportedFormats() []string {
	return []string{"json", "yaml"}
}
