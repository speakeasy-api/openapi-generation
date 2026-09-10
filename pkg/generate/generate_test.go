package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLegacyGeneratedFiles(t *testing.T) {
	tests := []struct {
		name                 string
		additionalProperties map[string]any
		expected             []string
	}{
		{
			name:                 "nil additionalProperties returns nil",
			additionalProperties: nil,
			expected:             nil,
		},
		{
			name:                 "empty additionalProperties returns nil",
			additionalProperties: map[string]any{},
			expected:             nil,
		},
		{
			name: "missing generatedFiles key returns nil",
			additionalProperties: map[string]any{
				"otherKey": "value",
			},
			expected: nil,
		},
		{
			name: "generatedFiles with wrong type returns nil",
			additionalProperties: map[string]any{
				"generatedFiles": "not-a-list",
			},
			expected: nil,
		},
		{
			name: "generatedFiles with empty list returns empty slice",
			additionalProperties: map[string]any{
				"generatedFiles": []any{},
			},
			expected: []string{},
		},
		{
			name: "generatedFiles with valid file paths",
			additionalProperties: map[string]any{
				"generatedFiles": []any{
					"src/sdk.go",
					"src/client.go",
					"README.md",
				},
			},
			expected: []string{"src/sdk.go", "src/client.go", "README.md"},
		},
		{
			name: "generatedFiles filters out non-string values",
			additionalProperties: map[string]any{
				"generatedFiles": []any{
					"src/sdk.go",
					123,
					"src/client.go",
					nil,
					true,
					"README.md",
				},
			},
			expected: []string{"src/sdk.go", "src/client.go", "README.md"},
		},
		{
			name: "generatedFiles with only non-string values returns empty slice",
			additionalProperties: map[string]any{
				"generatedFiles": []any{
					123,
					nil,
					true,
					map[string]any{"key": "value"},
				},
			},
			expected: []string{},
		},
		{
			name: "preserves other additionalProperties keys",
			additionalProperties: map[string]any{
				"ast":            "some-ast-data",
				"astversion":     "1.0.0",
				"generatedFiles": []any{"file1.go", "file2.go"},
			},
			expected: []string{"file1.go", "file2.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLegacyGeneratedFiles(tt.additionalProperties)
			assert.Equal(t, tt.expected, result)
		})
	}
}
