package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseTerraformAliasTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		expected *string
		wantErr  bool
	}{
		{
			name:     "null value",
			yaml:     `null`,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "empty string",
			yaml:     `""`,
			expected: ptr(""),
			wantErr:  false,
		},
		{
			name:     "string",
			yaml:     `test`,
			expected: ptr("test"),
			wantErr:  false,
		},
		{
			name:     "number",
			yaml:     `123`,
			expected: ptr("123"),
			wantErr:  false,
		},
		{
			name:     "boolean",
			yaml:     `true`,
			expected: ptr("true"),
			wantErr:  false,
		},
		{
			name:     "invalid sequence",
			yaml:     `[item1, item2]`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid mapping",
			yaml:     `{key: value}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid mapping with multiple keys",
			yaml: `
key1: value1
key2: value2`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid nested sequence",
			yaml: `
- item1
- item2
- item3`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid nested mapping",
			yaml: `
parent:
  child: value`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			require.NoError(t, err)

			// The root node is a document node, we need the content node
			require.Len(t, node.Content, 1, "Expected exactly one content node")
			contentNode := node.Content[0]

			e := &Extensions{}
			result, err := e.parseTerraformAliasTo(contentNode)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
