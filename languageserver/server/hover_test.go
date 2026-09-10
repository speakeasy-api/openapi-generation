package server

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFindNodeAtPosition(t *testing.T) {
	// Simple YAML for testing
	yamlContent := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users`

	var root yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &root)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	tests := []struct {
		name        string
		line        int
		character   int
		expectFound bool
		expectPath  []string
		expectValue string
	}{
		{
			name:        "find openapi version",
			line:        0, // 0-based, corresponds to YAML line 1
			character:   1, // position on "openapi" key
			expectFound: true,
			expectPath:  []string{"openapi"},
			expectValue: "openapi",
		},
		{
			name:        "find info title",
			line:        2, // 0-based, corresponds to YAML line 3
			character:   3, // position on "title"
			expectFound: true,
			expectPath:  []string{"info", "title"},
			expectValue: "title",
		},
		{
			name:        "find version value",
			line:        3,  // 0-based, corresponds to YAML line 4
			character:   11, // position on "1.0.0"
			expectFound: true,
			expectPath:  []string{"info", "version"},
			expectValue: "1.0.0",
		},
		{
			name:        "find path get method",
			line:        6, // 0-based, corresponds to YAML line 7
			character:   5, // position on "get"
			expectFound: true,
			expectPath:  []string{"paths", "/users", "get"},
			expectValue: "get",
		},
		{
			name:        "position not found",
			line:        10,
			character:   0, // beyond document
			expectFound: false,
			expectPath:  nil,
			expectValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, path := findNodeAtPosition(&root, tt.line, tt.character, []string{})

			if tt.expectFound {
				if node == nil {
					t.Errorf("Expected to find node but got nil")
					return
				}

				if node.Value != tt.expectValue {
					t.Errorf("Expected node value %q, got %q", tt.expectValue, node.Value)
				}

				if len(path) != len(tt.expectPath) {
					t.Errorf("Expected path length %d, got %d", len(tt.expectPath), len(path))
					return
				}

				for i, expectedPathPart := range tt.expectPath {
					if i >= len(path) || path[i] != expectedPathPart {
						t.Errorf("Expected path[%d] = %q, got %q", i, expectedPathPart, path[i])
					}
				}
			} else if node != nil {
				t.Errorf("Expected not to find node but got %v", node)
			}
		})
	}
}

func TestGetOpenAPIPropertyDoc(t *testing.T) {
	tests := []struct {
		name        string
		property    string
		path        []string
		expectMatch string // substring that should be in the result
	}{
		{
			name:        "openapi version",
			property:    "openapi",
			path:        []string{"openapi"},
			expectMatch: "OpenAPI Version",
		},
		{
			name:        "info object",
			property:    "info",
			path:        []string{"info"},
			expectMatch: "API Information",
		},
		{
			name:        "title in info context",
			property:    "title",
			path:        []string{"info", "title"},
			expectMatch: "API Title",
		},
		{
			name:        "title in other context",
			property:    "title",
			path:        []string{"other", "title"},
			expectMatch: "A title for the object",
		},
		{
			name:        "version in info context",
			property:    "version",
			path:        []string{"info", "version"},
			expectMatch: "API Version",
		},
		{
			name:        "paths object",
			property:    "paths",
			path:        []string{"paths"},
			expectMatch: "available paths and operations",
		},
		{
			name:        "get operation",
			property:    "get",
			path:        []string{"paths", "/users", "get"},
			expectMatch: "HTTP GET Operation",
		},
		{
			name:        "schema type",
			property:    "type",
			path:        []string{"components", "schemas", "User", "type"},
			expectMatch: "Schema Type",
		},
		{
			name:        "unknown property",
			property:    "unknown",
			path:        []string{"unknown"},
			expectMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOpenAPIPropertyDoc(tt.property, tt.path)

			if tt.expectMatch == "" {
				if result != "" {
					t.Errorf("Expected empty result for unknown property, got %q", result)
				}
			} else {
				if result == "" {
					t.Errorf("Expected non-empty result, got empty")
					return
				}

				if !contains(result, tt.expectMatch) {
					t.Errorf("Expected result to contain %q, got %q", tt.expectMatch, result)
				}
			}
		})
	}
}

func TestGetSpeakeasyExtensionDoc(t *testing.T) {
	tests := []struct {
		name        string
		property    string
		expectMatch string
	}{
		{
			name:        "name override extension",
			property:    "x-speakeasy-name-override",
			expectMatch: "Name Override",
		},
		{
			name:        "errors extension",
			property:    "x-speakeasy-errors",
			expectMatch: "Errors",
		},
		{
			name:        "retries extension",
			property:    "x-speakeasy-retries",
			expectMatch: "Retries",
		},
		{
			name:        "pagination extension",
			property:    "x-speakeasy-pagination",
			expectMatch: "Pagination",
		},
		{
			name:        "param name extension",
			property:    "x-speakeasy-param-name",
			expectMatch: "Parameter Name Override",
		},
		{
			name:        "usage example extension",
			property:    "x-speakeasy-usage-example",
			expectMatch: "Usage Example",
		},
		{
			name:        "unknown speakeasy extension",
			property:    "x-speakeasy-unknown",
			expectMatch: "Custom Speakeasy extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := getSpeakeasyExtensionDoc(tt.property, []string{})

			if result == "" {
				t.Errorf("Expected non-empty result for %s", tt.property)
				return
			}

			if !contains(result, tt.expectMatch) {
				t.Errorf("Expected result to contain %q, got %q", tt.expectMatch, result)
			}
		})
	}
}

func TestGetPropertyContext(t *testing.T) {
	tests := []struct {
		name           string
		path           []string
		expectedResult string
	}{
		{
			name:           "empty path",
			path:           []string{},
			expectedResult: "",
		},
		{
			name:           "single element path",
			path:           []string{"openapi"},
			expectedResult: "",
		},
		{
			name:           "two element path",
			path:           []string{"info", "title"},
			expectedResult: "info",
		},
		{
			name:           "multiple element path",
			path:           []string{"paths", "/users", "get", "responses"},
			expectedResult: "get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPropertyContext(tt.path)
			if result != tt.expectedResult {
				t.Errorf("Expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
