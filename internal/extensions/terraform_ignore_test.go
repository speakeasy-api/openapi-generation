package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseTerraformIgnore(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *TerraformIgnore
		wantErr  bool
	}{
		{
			name: "valid boolean true",
			yaml: `true`,
			expected: &TerraformIgnore{
				DataModel: true,
				Schema:    true,
			},
			wantErr: false,
		},
		{
			name: "valid boolean false",
			yaml: `false`,
			expected: &TerraformIgnore{
				DataModel: false,
				Schema:    false,
			},
			wantErr: false,
		},
		{
			name: "valid quoted string true",
			yaml: `"true"`,
			expected: &TerraformIgnore{
				DataModel: true,
				Schema:    true,
			},
			wantErr: false,
		},
		{
			name: "valid quoted string false",
			yaml: `"false"`,
			expected: &TerraformIgnore{
				DataModel: false,
				Schema:    false,
			},
			wantErr: false,
		},
		{
			name: "valid schema string",
			yaml: `schema`,
			expected: &TerraformIgnore{
				DataModel: false,
				Schema:    true,
			},
			wantErr: false,
		},
		{
			name: "valid quoted schema string",
			yaml: `"schema"`,
			expected: &TerraformIgnore{
				DataModel: false,
				Schema:    true,
			},
			wantErr: false,
		},
		{
			name:     "invalid string value",
			yaml:     `"invalid_value"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid unquoted string value",
			yaml:     `dataModel`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid null node",
			yaml:     `null`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid numeric node",
			yaml:     `42`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid mapping node",
			yaml: `
dataModel: true
schema: false`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node",
			yaml: `
- item1
- item2`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid empty mapping node",
			yaml:     `{}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid empty sequence node",
			yaml:     `[]`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			require.NoError(t, err)

			// The root node is a document node, we need the content node
			require.Len(t, node.Content, 1, "Expected exactly one content node")
			contentNode := node.Content[0]

			e := &Extensions{}
			result, err := e.parseTerraformIgnore(contentNode)

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

func TestTerraformIgnore_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ignore   *TerraformIgnore
		testFunc func(t *testing.T, original, cloned *TerraformIgnore)
	}{
		{
			name:   "nil ignore",
			ignore: nil,
			testFunc: func(t *testing.T, original, cloned *TerraformIgnore) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:   "empty ignore",
			ignore: &TerraformIgnore{},
			testFunc: func(t *testing.T, original, cloned *TerraformIgnore) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.False(t, cloned.DataModel)
				assert.False(t, cloned.Schema)
			},
		},
		{
			name: "ignore with all true",
			ignore: &TerraformIgnore{
				DataModel: true,
				Schema:    true,
			},
			testFunc: func(t *testing.T, original, cloned *TerraformIgnore) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.True(t, cloned.DataModel)
				assert.True(t, cloned.Schema)

				// Verify deep copy by modifying original
				original.DataModel = false
				assert.True(t, cloned.DataModel)
			},
		},
		{
			name: "ignore with schema only",
			ignore: &TerraformIgnore{
				DataModel: false,
				Schema:    true,
			},
			testFunc: func(t *testing.T, original, cloned *TerraformIgnore) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.False(t, cloned.DataModel)
				assert.True(t, cloned.Schema)

				// Verify deep copy by modifying original
				original.Schema = false
				assert.True(t, cloned.Schema)
			},
		},
		{
			name: "ignore with data model only",
			ignore: &TerraformIgnore{
				DataModel: true,
				Schema:    false,
			},
			testFunc: func(t *testing.T, original, cloned *TerraformIgnore) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.True(t, cloned.DataModel)
				assert.False(t, cloned.Schema)

				// Verify deep copy by modifying original
				original.DataModel = false
				assert.True(t, cloned.DataModel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.ignore.Clone()
			tt.testFunc(t, tt.ignore, cloned)
		})
	}
}
