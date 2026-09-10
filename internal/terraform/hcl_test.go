package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONToHCLExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty object",
			input:    `{}`,
			expected: `jsonencode({})`,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: `jsonencode([])`,
		},
		{
			name:     "string value",
			input:    `"hello"`,
			expected: `jsonencode("hello")`,
		},
		{
			name:     "integer number",
			input:    `42`,
			expected: `jsonencode(42)`,
		},
		{
			name:     "float number",
			input:    `3.14`,
			expected: `jsonencode(3.14)`,
		},
		{
			name:     "boolean true",
			input:    `true`,
			expected: `jsonencode(true)`,
		},
		{
			name:     "boolean false",
			input:    `false`,
			expected: `jsonencode(false)`,
		},
		{
			name:     "null",
			input:    `null`,
			expected: `jsonencode(null)`,
		},
		{
			name:  "simple object",
			input: `{"key": "value"}`,
			expected: `jsonencode({
  key = "value"
})`,
		},
		{
			name:  "object with multiple keys",
			input: `{"alpha": "a", "beta": "b"}`,
			expected: `jsonencode({
  alpha = "a"
  beta  = "b"
})`,
		},
		{
			name:  "object with mixed types",
			input: `{"name": "test", "count": 5, "enabled": true}`,
			expected: `jsonencode({
  count   = 5
  enabled = true
  name    = "test"
})`,
		},
		{
			name:     "simple array",
			input:    `["a", "b", "c"]`,
			expected: `jsonencode(["a", "b", "c"])`,
		},
		{
			name:     "array with mixed types",
			input:    `["hello", 42, true, null]`,
			expected: `jsonencode(["hello", 42, true, null])`,
		},
		{
			name:  "nested object",
			input: `{"outer": {"inner": "value"}}`,
			expected: `jsonencode({
  outer = {
    inner = "value"
  }
})`,
		},
		{
			name:  "object with array value",
			input: `{"tags": ["a", "b"]}`,
			expected: `jsonencode({
  tags = ["a", "b"]
})`,
		},
		{
			name:     "invalid JSON falls back to quoted string",
			input:    `{not valid json`,
			expected: `"{not valid json"`,
		},
		{
			name:     "empty string falls back to quoted string",
			input:    ``,
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := JSONToHCLExpression(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
