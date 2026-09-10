package server

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/oas"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"gopkg.in/yaml.v3"
)

func TestLookupDefinition(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		line           int
		character      int
		expectError    bool
		expectDocMatch string // substring that should be in documentation
	}{
		{
			name: "openapi version hover",
			yamlContent: `openapi: 3.0.3
info:
  title: Test API`,
			line:           0,
			character:      1, // position on "openapi"
			expectError:    false,
			expectDocMatch: "OpenAPI Version",
		},
		{
			name: "info title hover",
			yamlContent: `openapi: 3.0.3
info:
  title: Test API`,
			line:           2,
			character:      3, // position on "title"
			expectError:    false,
			expectDocMatch: "API Title",
		},
		{
			name: "speakeasy extension hover",
			yamlContent: `openapi: 3.0.3
info:
  title: Test API
  x-speakeasy-name-override: CustomAPI`,
			line:           3,
			character:      3, // position on "x-speakeasy-name-override"
			expectError:    false,
			expectDocMatch: "Name Override",
		},
		{
			name: "schema type hover",
			yamlContent: `openapi: 3.0.3
components:
  schemas:
    User:
      type: object`,
			line:           4,
			character:      7, // position on "type"
			expectError:    false,
			expectDocMatch: "Schema Type",
		},
		{
			name: "position not found",
			yamlContent: `openapi: 3.0.3
info:
  title: Test API`,
			line:           10,
			character:      0, // beyond document
			expectError:    true,
			expectDocMatch: "",
		},
		{
			name: "invalid yaml",
			yamlContent: `openapi: 3.0.3
info:
  title: Test API
  invalid: [unclosed`,
			line:           2,
			character:      2,
			expectError:    true,
			expectDocMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock document
			doc := &Document{
				URI:             "file:///test.yaml",
				Content:         tt.yamlContent,
				ReservedChecker: oas.NewDynamicReservedChecker(),
			}

			// Create hover parameters
			params := &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: doc.URI,
					},
					Position: protocol.Position{
						Line:      protocol.UInteger(tt.line),
						Character: protocol.UInteger(tt.character),
					},
				},
			}

			// Call lookupDefinition
			definition, err := lookupDefinition(doc, params)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if definition == nil {
				t.Errorf("Expected definition but got nil")
				return
			}

			if definition.node == nil {
				t.Errorf("Expected definition.node but got nil")
				return
			}

			if definition.documentation == "" {
				t.Errorf("Expected documentation but got empty string")
				return
			}

			if tt.expectDocMatch != "" {
				if !contains(definition.documentation, tt.expectDocMatch) {
					t.Errorf("Expected documentation to contain %q, got %q",
						tt.expectDocMatch, definition.documentation)
				}
			}
		})
	}
}

func TestGetEnhancedDocumentation(t *testing.T) {
	tests := []struct {
		name           string
		path           []string
		expectDocMatch string
		expectEmpty    bool
	}{
		{
			name:           "empty path",
			path:           []string{},
			expectDocMatch: "",
			expectEmpty:    true,
		},
		{
			name:           "openapi property",
			path:           []string{"openapi"},
			expectDocMatch: "OpenAPI Version",
			expectEmpty:    false,
		},
		{
			name:           "speakeasy extension",
			path:           []string{"info", "x-speakeasy-name-override"},
			expectDocMatch: "Name Override",
			expectEmpty:    false,
		},
		{
			name:           "nested path",
			path:           []string{"paths", "/users", "get", "summary"},
			expectDocMatch: "Summary",
			expectEmpty:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy node for testing
			node := &yaml.Node{
				Kind:   yaml.ScalarNode,
				Value:  "test",
				Line:   1,
				Column: 1,
			}

			result := getEnhancedDocumentation(node, tt.path, nil)

			if tt.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty result, got %q", result)
				}
			} else {
				if result == "" {
					t.Errorf("Expected non-empty result")
					return
				}

				if tt.expectDocMatch != "" && !contains(result, tt.expectDocMatch) {
					t.Errorf("Expected result to contain %q, got %q", tt.expectDocMatch, result)
				}
			}
		})
	}
}

func TestGetLibOpenAPIContext(t *testing.T) {
	tests := []struct {
		name           string
		path           []string
		expectContains []string // strings that should be in the result
	}{
		{
			name:           "schema component path",
			path:           []string{"components", "schemas", "User", "properties", "name"},
			expectContains: []string{"Path"},
		},
		{
			name:           "operation path",
			path:           []string{"paths", "/users", "get"},
			expectContains: []string{"Path"},
		},
		{
			name:           "simple property",
			path:           []string{"info", "title"},
			expectContains: []string{"Path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy node for testing
			node := &yaml.Node{
				Kind:   yaml.ScalarNode,
				Value:  "test",
				Line:   1,
				Column: 1,
			}

			// Test with nil rolodex (should still work but with limited info)
			result := getLibOpenAPIContext(node, tt.path, nil)

			// Should at least have path information
			for _, expected := range tt.expectContains {
				if !contains(result, expected) {
					t.Errorf("Expected result to contain %q, got %q", expected, result)
				}
			}
		})
	}
}
