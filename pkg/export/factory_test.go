package export

import (
	"testing"
)

func TestGetFormatter_JSON(t *testing.T) {
	formatter, err := GetFormatter("json")
	if err != nil {
		t.Fatalf("GetFormatter failed: %v", err)
	}

	if formatter == nil {
		t.Fatal("Expected non-nil formatter")
	}

	if _, ok := formatter.(*JSONFormat); !ok {
		t.Error("Expected JSONFormat type")
	}
}

func TestGetFormatter_YAML(t *testing.T) {
	formatter, err := GetFormatter("yaml")
	if err != nil {
		t.Fatalf("GetFormatter failed: %v", err)
	}

	if formatter == nil {
		t.Fatal("Expected non-nil formatter")
	}

	if _, ok := formatter.(*YAMLFormat); !ok {
		t.Error("Expected YAMLFormat type")
	}
}

func TestGetFormatter_YML(t *testing.T) {
	formatter, err := GetFormatter("yml")
	if err != nil {
		t.Fatalf("GetFormatter failed: %v", err)
	}

	if formatter == nil {
		t.Fatal("Expected non-nil formatter")
	}

	if _, ok := formatter.(*YAMLFormat); !ok {
		t.Error("Expected YAMLFormat type")
	}
}

func TestGetFormatter_CaseInsensitive(t *testing.T) {
	tests := []string{"JSON", "Json", "YAML", "Yaml", "YML", "Yml"}
	
	for _, format := range tests {
		formatter, err := GetFormatter(format)
		if err != nil {
			t.Errorf("GetFormatter('%s') failed: %v", format, err)
		}
		if formatter == nil {
			t.Errorf("Expected non-nil formatter for '%s'", format)
		}
	}
}

func TestGetFormatter_UnsupportedFormat(t *testing.T) {
	formatter, err := GetFormatter("xml")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if formatter != nil {
		t.Error("Expected nil formatter for unsupported format")
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	
	if len(formats) != 2 {
		t.Errorf("Expected 2 supported formats, got %d", len(formats))
	}
	
	// Check that json and yaml are in the list
	hasJSON := false
	hasYAML := false
	
	for _, format := range formats {
		if format == "json" {
			hasJSON = true
		}
		if format == "yaml" {
			hasYAML = true
		}
	}
	
	if !hasJSON {
		t.Error("Expected 'json' in supported formats")
	}
	if !hasYAML {
		t.Error("Expected 'yaml' in supported formats")
	}
}
