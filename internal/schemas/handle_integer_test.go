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

// TestHandleInteger tests the handleInteger method with various integer scenarios
func TestHandleInteger(t *testing.T) {
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
			name: "BasicInteger",
			schemasYAML: `TestSchema:
  type: integer`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithMinMax",
			schemasYAML: `TestSchema:
  type: integer
  minimum: 0
  maximum: 100`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						Minimum: pointer.From[float64](0),
						Maximum: pointer.From[float64](100),
					}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithInt32Format",
			schemasYAML: `TestSchema:
  type: integer
  format: int32`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("int32",
				testutils.NewTypeDef(ast.DataTypeInt32).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithBigIntFormat",
			schemasYAML: `TestSchema:
  type: integer
  format: bigint`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureBigInt: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("bigint",
				testutils.NewTypeDef(ast.DataTypeBigInt).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithBigIntFormatUnsupported",
			schemasYAML: `TestSchema:
  type: integer
  format: bigint`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureBigInt: false,
				}
				return config
			},
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithNullable",
			schemasYAML: `TestSchema:
  type: integer
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "IntegerWithExample",
			schemasYAML: `TestSchema:
  type: integer
  example: 42`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExamples(testutils.NewIntExample("42")).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithExtensions",
			schemasYAML: `TestSchema:
  type: integer
  x-custom: "value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExtensions(map[string]interface{}{
						"x-custom": "value",
					}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithDefault",
			schemasYAML: `TestSchema:
  type: integer
  default: 10`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithOptional(true).
				WithDefault(&ast.AnyValue{Value: 10}).
				Build(),
		},
		{
			name: "IntegerWithConst",
			schemasYAML: `TestSchema:
  type: integer
  const: 42`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithConst(&ast.AnyValue{Value: 42}).
				Build(),
		},
		{
			name: "IntegerReference",
			schemasYAML: `IntegerType:
  type: integer
TestSchema:
  $ref: "#/components/schemas/IntegerType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("IntegerType",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithName("IntegerType").
					WithOriginalName("").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithExclusiveMinMax",
			schemasYAML: `TestSchema:
  type: integer
  minimum: 0
  maximum: 100
  exclusiveMinimum: true
  exclusiveMaximum: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						Minimum: pointer.From[float64](0),
						Maximum: pointer.From[float64](100),
					}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithMultipleOf",
			schemasYAML: `TestSchema:
  type: integer
  multipleOf: 5`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "IntegerWithInt64Format",
			schemasYAML: `TestSchema:
  type: integer
  format: int64`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using common testutils

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
