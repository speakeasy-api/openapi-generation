package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleAnyType tests the handleAnyType method with various any type scenarios
func TestHandleAnyType(t *testing.T) {
	tests := []struct {
		name         string
		schemasYAML  string
		schemaName   string
		fixes        *config.Fixes                        // nil means use default
		mockFeatures func() *testutils.MockFeaturesConfig // nil means use default (all features supported)
		isRequest    *bool                                // nil means use default (false), otherwise use specified value
		expected     *ast.FieldDef
	}{
		{
			name: "BasicAnyType",
			schemasYAML: `TestSchema:
  type: object`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithEmptyFields().
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithoutType",
			schemasYAML: `TestSchema:
  description: "Any type without explicit type"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithExample",
			schemasYAML: `TestSchema:
  description: "Any type with example"
  example: {"key": "value"}`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExamples(
						&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.MappingNode,
								Style:  yaml.FlowStyle,
								Tag:    "!!map",
								Line:   10,
								Column: 16,
								Content: []*yaml.Node{
									{
										Kind:   yaml.ScalarNode,
										Style:  yaml.DoubleQuotedStyle,
										Tag:    "!!str",
										Value:  "key",
										Line:   10,
										Column: 17,
									},
									{
										Kind:   yaml.ScalarNode,
										Style:  yaml.DoubleQuotedStyle,
										Tag:    "!!str",
										Value:  "value",
										Line:   10,
										Column: 24,
									},
								},
							},
						},
					).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithExtensions",
			schemasYAML: `TestSchema:
  description: "Any type with extensions"
  x-custom-extension: "custom-value"
  x-another-extension: 42`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-custom-extension":  "custom-value",
						"x-another-extension": 42,
					}).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithTitle",
			schemasYAML: `TestSchema:
  title: "Custom Any Type"
  description: "Any type with title"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("Custom Any Type",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithNullable",
			schemasYAML: `TestSchema:
  description: "Any type with nullable"
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "AnyTypeReference",
			schemasYAML: `AnyTypeSchema:
  description: "Referenced any type"
TestSchema:
  $ref: "#/components/schemas/AnyTypeSchema"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("AnyTypeSchema",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithDefault",
			schemasYAML: `TestSchema:
  description: "Any type with default value"
  default: "default_value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				WithOptional(true).
				WithDefault(&ast.AnyValue{Value: "default_value"}).
				Build(),
		},
		{
			name: "AnyTypeWithConst",
			schemasYAML: `TestSchema:
  description: "Any type with const value"
  const: "constant_value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithValidations(&ast.Validations{}).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				WithConst(&ast.AnyValue{Value: "constant_value"}).
				Build(),
		},
		{
			name:        "EmptySchema",
			schemasYAML: `TestSchema: {}`,
			schemaName:  "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 8, Column: 17}).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithMultipleExamples",
			schemasYAML: `TestSchema:
  description: "Any type with multiple examples"
  examples:
    - "string_example"
    - 42
    - true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExamples(
						&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Style:  yaml.DoubleQuotedStyle,
								Tag:    "!!str",
								Value:  "string_example",
								Line:   11,
								Column: 11,
							},
						},
						&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Tag:    "!!int",
								Value:  "42",
								Line:   12,
								Column: 11,
							},
						},
						&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Tag:    "!!bool",
								Value:  "true",
								Line:   13,
								Column: 11,
							},
						},
					).
					Build()).
				Build(),
		},
		{
			name: "AnyTypeWithRequestResponseComponentNames",
			schemasYAML: `TestSchema:
  description: "Any type for request/response component names test"`,
			schemaName: "TestSchema",
			fixes: &config.Fixes{
				RequestResponseComponentNamesFeb2024: true,
			},
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithComplexAny(false).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using common testutils
			fullYAML := testutils.CreateOpenAPIDoc(tt.schemasYAML)

			var mockFeatures *testutils.MockFeaturesConfig
			if tt.mockFeatures != nil {
				mockFeatures = tt.mockFeatures()
			} else {
				mockFeatures = testutils.NewMockFeaturesConfig().WithSupportAllFeatures()
			}

			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  fullYAML,
				Fixes:        tt.fixes,
				MockFeatures: mockFeatures,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Create the Schemas instance
			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get the test schema
			testSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get(tt.schemaName)
			require.True(t, exists)
			require.NotNil(t, testSchema)

			// Create test parameters
			isRequest := false
			if tt.isRequest != nil {
				isRequest = *tt.isRequest
			}

			params := Params{
				ContextStack:        ast.ContextStack{},
				SerializationMethod: ast.SerializationMethodJSON,
				Schema:              testSchema,
				Scope:               ast.ScopeShared,
				IsRequest:           isRequest,
				Depth:               0,
				MaxDepth:            10,
				Parents:             []string{},
				TypeDefCache:        make(map[string]*ast.TypeDef),
				LoopContext:         []LoopFrame{},
				Nullable:            false,
				CircularReference:   false,
				DocInfo:             common.DocInfo,
			}

			// Call the top-level HandleSchema method instead of handleAnyType directly
			ctx := context.Background()
			result, err := schemas.HandleSchema(ctx, params)

			// Verify the result
			require.NoError(t, err)
			require.NotNil(t, result)

			// Use our deep comparison function to find exact differences
			isEqual := testutils.DeepCompare(t, tt.expected, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
