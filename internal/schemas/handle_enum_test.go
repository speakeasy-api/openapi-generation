package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleEnum tests the handleEnum method with various enum scenarios
func TestHandleEnum(t *testing.T) {
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
			name: "StringEnum",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "active"
    - "inactive"
    - "pending"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive", "pending"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerEnum",
			schemasYAML: `TestSchema:
  type: integer
  enum:
    - 1
    - 2
    - 3`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeInteger, []string{"1", "2", "3"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "Int32Enum",
			schemasYAML: `TestSchema:
  type: integer
  format: int32
  enum:
    - 10
    - 20
    - 30`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeInt32, []string{"10", "20", "30"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithNullable",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "red"
    - "green"
    - "blue"
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"red", "green", "blue"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "EnumWithSingleNullValue",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - null`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				WithConst(&ast.AnyValue{Value: nil}).
				Build(),
		},
		{
			name: "EnumWithCustomNames",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "ACTIVE"
    - "INACTIVE"
    - "PENDING"
  x-speakeasy-enums:
    - Active
    - Inactive
    - Pending`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"ACTIVE", "INACTIVE", "PENDING"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithEnumNames([]string{"Active", "Inactive", "Pending"}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-enums": []interface{}{"Active", "Inactive", "Pending"},
					}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithDuplicateValues",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "active"
    - "inactive"
    - "active"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OpenEnum",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "known1"
    - "known2"
  x-speakeasy-open-enum: true`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig().WithTarget(types.Target{Target: "go"})
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureOpenEnums: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"known1", "known2"}, "").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-open-enum": true,
					}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithDescription",
			schemasYAML: `TestSchema:
  type: string
  description: "Status enumeration"
  enum:
    - "active"
    - "inactive"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithComments(&ast.Comment{
						Description:      "Status enumeration",
						ExtendedComments: make(map[string]*ast.ExtendedComment),
					}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithTitle",
			schemasYAML: `TestSchema:
  type: string
  title: "Status"
  enum:
    - "active"
    - "inactive"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("Status",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithName("Status").
					WithOriginalName("Status").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithExample",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "red"
    - "green"
    - "blue"
  example: "red"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"red", "green", "blue"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind:   yaml.ScalarNode,
							Style:  yaml.DoubleQuotedStyle,
							Tag:    "!!str",
							Value:  "red",
							Line:   14,
							Column: 16,
						},
					}).
					Build()).
				Build(),
		},
		{
			name: "EnumWithExtensions",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "value1"
    - "value2"
  x-custom-extension: "custom-value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"value1", "value2"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-custom-extension": "custom-value",
					}).
					Build()).
				Build(),
		},
		{
			name: "EnumReference",
			schemasYAML: `StatusEnum:
  type: string
  enum:
    - "active"
    - "inactive"
TestSchema:
  $ref: "#/components/schemas/StatusEnum"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("StatusEnum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithName("StatusEnum").
					WithOriginalName("StatusEnum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("StatusEnum").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "EnumWithMixedTypes",
			schemasYAML: `TestSchema:
  enum:
    - "string_value"
    - 42
    - true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"string_value", "42", "true"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "EnumUnionFormat",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "option1"
    - "option2"
  x-speakeasy-enum-format: "union"`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig().WithTarget(types.Target{Target: "go"})
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureEnumUnions: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"option1", "option2"}, "union").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-enum-format": "union",
					}).
					Build()).
				Build(),
		},
		{
			name: "SingleValueEnumWithConstProperty",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "single_value"`,
			schemaName: "TestSchema",
			fixes: &config.Fixes{
				NameResolutionFeb2025: true,
			},
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"single_value"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithConstProperty("single_value").WithIdentifierForNaming(pointer.From("SingleValue")).Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "StringEnumWithNullMember",
			schemasYAML: `TestSchema:
  type: string
  enum:
    - "active"
    - "inactive"
    - null`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerEnumWithNullMember",
			schemasYAML: `TestSchema:
  type: integer
  enum:
    - 1
    - 2
    - null`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeInteger, []string{"1", "2"}, "enum").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
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
			params := CreateTestParams(testSchema, common.DocInfo)
			if tt.isRequest != nil {
				params.IsRequest = *tt.isRequest
			}

			// Call HandleSchema - let it handle nullable extraction and routing
			result, err := schemas.HandleSchema(context.Background(), params)
			require.NoError(t, err)

			// Use deep comparison from the working test
			isEqual := testutils.DeepCompare(t, tt.expected, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
