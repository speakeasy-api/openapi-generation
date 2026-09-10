package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleNumber tests the handleNumber method with various number scenarios
func TestHandleNumber(t *testing.T) {
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
			name: "BasicNumber",
			schemasYAML: `TestSchema:
  type: number`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithMinMax",
			schemasYAML: `TestSchema:
  type: number
  minimum: 0.0
  maximum: 100.0`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						Minimum: pointer.From[float64](0.0),
						Maximum: pointer.From[float64](100.0),
					}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithFloatFormat",
			schemasYAML: `TestSchema:
  type: number
  format: float`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("float32",
				testutils.NewTypeDef(ast.DataTypeFloat32).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithDecimalFormat",
			schemasYAML: `TestSchema:
  type: number
  format: decimal`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureDecimal: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("decimal",
				testutils.NewTypeDef(ast.DataTypeDecimal).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithDecimalFormatUnsupported",
			schemasYAML: `TestSchema:
  type: number
  format: decimal`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureDecimal: false,
				}
				return config
			},
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithNullable",
			schemasYAML: `TestSchema:
  type: number
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "NumberWithExample",
			schemasYAML: `TestSchema:
  type: number
  example: 42.5`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExamples(testutils.NewFloatExample("42.5")).
					Build()).
				Build(),
		},
		{
			name: "NumberWithExtensions",
			schemasYAML: `TestSchema:
  type: number
  x-custom: "value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExtensions(map[string]interface{}{
						"x-custom": "value",
					}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithDefault",
			schemasYAML: `TestSchema:
  type: number
  default: 10.5`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithOptional(true).
				WithDefault(&ast.AnyValue{Value: 10.5}).
				Build(),
		},
		{
			name: "NumberWithConst",
			schemasYAML: `TestSchema:
  type: number
  const: 3.14159`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithConst(&ast.AnyValue{Value: 3.14159}).
				Build(),
		},
		{
			name: "NumberReference",
			schemasYAML: `NumberType:
  type: number
TestSchema:
  $ref: "#/components/schemas/NumberType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("NumberType",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithName("NumberType").
					WithOriginalName("").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithExclusiveMinMax",
			schemasYAML: `TestSchema:
  type: number
  minimum: 0.0
  maximum: 100.0
  exclusiveMinimum: true
  exclusiveMaximum: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						Minimum: pointer.From[float64](0.0),
						Maximum: pointer.From[float64](100.0),
					}).
					Build()).
				Build(),
		},
		{
			name: "NumberWithMultipleOf",
			schemasYAML: `TestSchema:
  type: number
  multipleOf: 0.5`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
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
