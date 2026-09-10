package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseIgnore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   *types.Target
		yaml     string
		expected *bool
		wantErr  bool
	}{
		{
			name:     "true value",
			yaml:     `true`,
			expected: ptr(true),
		},
		{
			name:     "false value",
			yaml:     `false`,
			expected: ptr(false),
		},
		{
			name:     "True value",
			yaml:     `True`,
			expected: ptr(true),
		},
		{
			name:     "False value",
			yaml:     `False`,
			expected: ptr(false),
		},
		{
			name:     "1 value",
			yaml:     `1`,
			expected: ptr(true),
		},
		{
			name:     "0 value",
			yaml:     `0`,
			expected: ptr(false),
		},
		{
			name:     "target list with matching target",
			target:   &types.Target{Target: "go"},
			yaml:     `go,python,typescript`,
			expected: ptr(true),
		},
		{
			name:     "target list with matching target and spaces",
			target:   &types.Target{Target: "python"},
			yaml:     `go, python, typescript`,
			expected: ptr(true),
		},
		{
			name:     "target list with non-matching target",
			target:   &types.Target{Target: "java"},
			yaml:     `go,python,typescript`,
			expected: ptr(false),
		},
		{
			name:     "single target matching",
			target:   &types.Target{Target: "terraform"},
			yaml:     `terraform`,
			expected: ptr(true),
		},
		{
			name:     "single target not matching",
			target:   &types.Target{Target: "go"},
			yaml:     `terraform`,
			expected: ptr(false),
		},
		{
			name:     "target list with empty target",
			target:   &types.Target{Target: ""},
			yaml:     `go,python,typescript`,
			expected: ptr(false),
		},
		{
			name:     "target list without target set",
			yaml:     `go,python,typescript`,
			expected: ptr(false),
		},
		{
			name:    "invalid sequence",
			yaml:    `[item1, item2]`,
			wantErr: true,
		},
		{
			name:    "invalid mapping",
			yaml:    `{key: value}`,
			wantErr: true,
		},
		{
			name: "invalid mapping with multiple keys",
			yaml: `
key1: value1
key2: value2`,
			wantErr: true,
		},
		{
			name: "invalid nested sequence",
			yaml: `
- item1
- item2
- item3`,
			wantErr: true,
		},
		{
			name: "invalid nested mapping",
			yaml: `
parent:
  child: value`,
			wantErr: true,
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

			var target types.Target
			if tt.target != nil {
				target = *tt.target
			}

			e := &Extensions{target: target}
			result, err := e.parseIgnore(contentNode)

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
