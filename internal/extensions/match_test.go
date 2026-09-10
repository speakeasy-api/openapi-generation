package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		expected *MatchConfig
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
			expected: &MatchConfig{Path: ptr(""), UsePriorState: false},
			wantErr:  false,
		},
		{
			name:     "string path (backward compatibility)",
			yaml:     `test`,
			expected: &MatchConfig{Path: ptr("test"), UsePriorState: false},
			wantErr:  false,
		},
		{
			name:     "number as string path",
			yaml:     `123`,
			expected: &MatchConfig{Path: ptr("123"), UsePriorState: false},
			wantErr:  false,
		},
		{
			name:     "boolean as string path",
			yaml:     `true`,
			expected: &MatchConfig{Path: ptr("true"), UsePriorState: false},
			wantErr:  false,
		},
		{
			name:     "object with usePriorState only",
			yaml:     `{usePriorState: true}`,
			expected: &MatchConfig{Path: nil, UsePriorState: true},
			wantErr:  false,
		},
		{
			name:     "object with path only",
			yaml:     `{path: "id"}`,
			expected: &MatchConfig{Path: ptr("id"), UsePriorState: false},
			wantErr:  false,
		},
		{
			name:     "object with both path and usePriorState",
			yaml:     `{path: "object.id", usePriorState: true}`,
			expected: &MatchConfig{Path: ptr("object.id"), UsePriorState: true},
			wantErr:  false,
		},
		{
			name:     "invalid empty object",
			yaml:     `{}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid usePriorState false only",
			yaml:     `{usePriorState: false}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid sequence",
			yaml:     `[item1, item2]`,
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
			result, err := e.parseMatch(contentNode)

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
