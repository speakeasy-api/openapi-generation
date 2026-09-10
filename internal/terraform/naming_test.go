package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   map[string]any
		expected string
	}{
		{
			name: "packageName only",
			config: map[string]any{
				"packageName": "testing",
			},
			expected: "testing",
		},
		{
			name: "packageName with hyphen",
			config: map[string]any{
				"packageName": "my-provider",
			},
			expected: "my-provider",
		},
		{
			name: "packageName with mixed casing normalized to kebab",
			config: map[string]any{
				"packageName": "MyProvider",
			},
			expected: "my-provider",
		},
		{
			name: "providerTypeNameOverride takes precedence",
			config: map[string]any{
				"packageName":              "myprovider-beta",
				"providerTypeNameOverride": "myprovider",
			},
			expected: "myprovider",
		},
		{
			name: "empty providerTypeNameOverride falls back to packageName",
			config: map[string]any{
				"packageName":              "myprovider",
				"providerTypeNameOverride": "",
			},
			expected: "myprovider",
		},
		{
			name:     "no config",
			config:   map[string]any{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ProviderTypeName(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		providerTypeName string
		entityName       string
		expected         string
	}{
		{
			name:             "PascalCase entity",
			providerTypeName: "testing",
			entityName:       "ExampleResource",
			expected:         "testing_example_resource",
		},
		{
			name:             "single word entity",
			providerTypeName: "testing",
			entityName:       "Thing",
			expected:         "testing_thing",
		},
		{
			name:             "snake_case entity preserved",
			providerTypeName: "testing",
			entityName:       "example_resource",
			expected:         "testing_example_resource",
		},
		{
			name:             "acronyms handled",
			providerTypeName: "testing",
			entityName:       "HTTPClient",
			expected:         "testing_http_client",
		},
		{
			name:             "hyphenated provider type name",
			providerTypeName: "my-provider",
			entityName:       "Thing",
			expected:         "my-provider_thing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TypeName(tt.providerTypeName, tt.entityName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAttributeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "camelCase to snake_case",
			input:    "myFieldName",
			expected: "my_field_name",
		},
		{
			name:     "PascalCase to snake_case",
			input:    "MyFieldName",
			expected: "my_field_name",
		},
		{
			name:     "already snake_case",
			input:    "my_field_name",
			expected: "my_field_name",
		},
		{
			name:     "removes trailing underscore",
			input:    "myField_",
			expected: "my_field",
		},
		{
			name:     "handles acronyms",
			input:    "HTTPResponse",
			expected: "http_response",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "field",
			expected: "field",
		},
		{
			name:     "single uppercase word",
			input:    "Field",
			expected: "field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := AttributeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
