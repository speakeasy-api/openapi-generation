package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseTerraformCustomDefault(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *TerraformCustomDefault
		wantErr  bool
	}{
		{
			name: "valid mapping with imports and schema definition",
			yaml: `
imports:
  - github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault
  - github.com/hashicorp/terraform-plugin-framework/types
schemaDefinition: stringdefault.StaticValue(types.StringValue("test"))`,
			expected: &TerraformCustomDefault{
				Imports: []string{
					"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
					"github.com/hashicorp/terraform-plugin-framework/types",
				},
				SchemaDefinition: "stringdefault.StaticValue(types.StringValue(\"test\"))",
			},
			wantErr: false,
		},
		{
			name: "valid mapping with no imports",
			yaml: `
schemaDefinition: CustomDefault()`,
			expected: &TerraformCustomDefault{
				Imports:          nil,
				SchemaDefinition: "CustomDefault()",
			},
			wantErr: false,
		},
		{
			name: "valid mapping with empty imports array",
			yaml: `
imports: []
schemaDefinition: SimpleDefault()`,
			expected: &TerraformCustomDefault{
				Imports:          []string{},
				SchemaDefinition: "SimpleDefault()",
			},
			wantErr: false,
		},
		{
			name: "valid mapping with single import",
			yaml: `
imports:
  - example.com/customdefaults
schemaDefinition: customdefaults.Example()`,
			expected: &TerraformCustomDefault{
				Imports: []string{
					"example.com/customdefaults",
				},
				SchemaDefinition: "customdefaults.Example()",
			},
			wantErr: false,
		},
		{
			name: "valid mapping with multiline schema definition",
			yaml: `
imports:
  - github.com/hashicorp/terraform-plugin-framework/types
schemaDefinition: |
  CustomDefault(
      types.StringValue("test"),
  )`,
			expected: &TerraformCustomDefault{
				Imports: []string{
					"github.com/hashicorp/terraform-plugin-framework/types",
				},
				SchemaDefinition: "CustomDefault(\n    types.StringValue(\"test\"),\n)",
			},
			wantErr: false,
		},
		{
			name: "valid mapping with only imports",
			yaml: `
imports:
  - fmt
  - strings`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "mapping with null imports",
			yaml: `
imports: null
schemaDefinition: DefaultFunc()`,
			expected: &TerraformCustomDefault{
				Imports:          nil,
				SchemaDefinition: "DefaultFunc()",
			},
			wantErr: false,
		},
		{
			name: "mapping with null schema definition",
			yaml: `
schemaDefinition: null`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "mapping with empty string schema definition",
			yaml: `
schemaDefinition: ""`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid scalar node - string",
			yaml:     `"invalid_scalar_value"`,
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
			name:     "invalid boolean node",
			yaml:     `false`,
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
			name:     "invalid numeric node",
			yaml:     `42`,
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
			name: "mapping with imports as string instead of array",
			yaml: `
imports: "single_import"
schemaDefinition: Test()`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "mapping with additional unknown fields",
			yaml: `
schemaDefinition: DefaultFunc()
unknownField: should_be_ignored`,
			expected: &TerraformCustomDefault{
				SchemaDefinition: "DefaultFunc()",
			},
			wantErr: false,
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
			result, err := e.parseTerraformCustomDefault(contentNode)

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

func TestTerraformCustomDefault_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		defaults *TerraformCustomDefault
		testFunc func(t *testing.T, original, cloned *TerraformCustomDefault)
	}{
		{
			name:     "nil defaults",
			defaults: nil,
			testFunc: func(t *testing.T, original, cloned *TerraformCustomDefault) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:     "empty defaults",
			defaults: &TerraformCustomDefault{},
			testFunc: func(t *testing.T, original, cloned *TerraformCustomDefault) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Empty(t, cloned.SchemaDefinition)
				assert.Nil(t, cloned.Imports)
			},
		},
		{
			name: "defaults with schema definition only",
			defaults: &TerraformCustomDefault{
				SchemaDefinition: "stringdefault.StaticValue(types.StringValue(\"test\"))",
			},
			testFunc: func(t *testing.T, original, cloned *TerraformCustomDefault) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.SchemaDefinition, cloned.SchemaDefinition)
				assert.Nil(t, cloned.Imports)

				// Verify deep copy by modifying original
				original.SchemaDefinition = "modified"
				assert.Equal(t, "stringdefault.StaticValue(types.StringValue(\"test\"))", cloned.SchemaDefinition)
			},
		},
		{
			name: "defaults with imports and schema definition",
			defaults: &TerraformCustomDefault{
				Imports: []string{
					"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
					"github.com/hashicorp/terraform-plugin-framework/types",
				},
				SchemaDefinition: "stringdefault.StaticValue(types.StringValue(\"test\"))",
			},
			testFunc: func(t *testing.T, original, cloned *TerraformCustomDefault) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.SchemaDefinition, cloned.SchemaDefinition)
				assert.Equal(t, original.Imports, cloned.Imports)
				for i := range original.Imports {
					assert.NotSame(t, &original.Imports[i], &cloned.Imports[i])
				}

				// Verify deep copy by modifying original
				original.Imports = append(original.Imports, "new.import")
				assert.Len(t, cloned.Imports, 2)
			},
		},
		{
			name: "defaults with nil imports",
			defaults: &TerraformCustomDefault{
				Imports:          nil,
				SchemaDefinition: "test",
			},
			testFunc: func(t *testing.T, original, cloned *TerraformCustomDefault) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.SchemaDefinition, cloned.SchemaDefinition)
				assert.Nil(t, cloned.Imports)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.defaults.Clone()
			tt.testFunc(t, tt.defaults, cloned)
		})
	}
}
