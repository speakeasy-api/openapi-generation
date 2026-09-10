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

// TestHandleString tests the handleString method with various string scenarios
func TestHandleString(t *testing.T) {
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
			name: "BasicString",
			schemasYAML: `TestSchema:
  type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithMinMaxLength",
			schemasYAML: `TestSchema:
  type: string
  minLength: 1
  maxLength: 100`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						MinLength: pointer.From[int64](1),
						MaxLength: pointer.From[int64](100),
					}).
					Build()).
				Build(),
		},
		{
			name: "StringWithPattern",
			schemasYAML: `TestSchema:
  type: string
  pattern: "^[a-zA-Z0-9]+$"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{
						Pattern: pointer.From("^[a-zA-Z0-9]+$"),
					}).
					Build()).
				Build(),
		},
		{
			name: "StringWithDateFormat",
			schemasYAML: `TestSchema:
  type: string
  format: date`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("date",
				testutils.NewTypeDef(ast.DataTypeDate).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithDateTimeFormat",
			schemasYAML: `TestSchema:
  type: string
  format: date-time`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("date-time",
				testutils.NewTypeDef(ast.DataTypeDateTime).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithBinaryFormat",
			schemasYAML: `TestSchema:
  type: string
  format: binary`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("bytes",
				testutils.NewTypeDef(ast.DataTypeBytes).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithInt64Format",
			schemasYAML: `TestSchema:
  type: string
  format: int64`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureStringNumberFormats: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithFormat("string").
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithFloat64Format",
			schemasYAML: `TestSchema:
  type: string
  format: float64`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureStringNumberFormats: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithFormat("string").
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithBigIntFormat",
			schemasYAML: `TestSchema:
  type: string
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
					WithFormat("string").
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithDecimalFormat",
			schemasYAML: `TestSchema:
  type: string
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
					WithFormat("string").
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithCustomFormat",
			schemasYAML: `TestSchema:
  type: string
  format: custom-format`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithFormat("custom-format").
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithNullable",
			schemasYAML: `TestSchema:
  type: string
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "StringWithExample",
			schemasYAML: `TestSchema:
  type: string
  example: "example value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind:   yaml.ScalarNode,
							Style:  yaml.DoubleQuotedStyle,
							Tag:    "!!str",
							Value:  "example value",
							Line:   10,
							Column: 16,
						},
					}).
					Build()).
				Build(),
		},
		{
			name: "StringWithExtensions",
			schemasYAML: `TestSchema:
  type: string
  x-custom: "value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					WithExtensions(map[string]interface{}{
						"x-custom": "value",
					}).
					Build()).
				Build(),
		},
		{
			name: "StringWithDefault",
			schemasYAML: `TestSchema:
  type: string
  default: "default value"`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithOptional(true).
				WithDefault(&ast.AnyValue{Value: "default value"}).
				Build(),
		},
		{
			name: "StringWithConst",
			schemasYAML: `TestSchema:
  type: string
  const: "constant value"`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithConst(&ast.AnyValue{Value: "constant value"}).
				Build(),
		},
		{
			name: "StringReference",
			schemasYAML: `StringType:
  type: string
TestSchema:
  $ref: "#/components/schemas/StringType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("StringType",
				testutils.NewTypeDef(ast.DataTypeString).
					WithName("StringType").
					WithOriginalName("").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithRequestStream",
			schemasYAML: `TestSchema:
  type: string
  format: binary`,
			schemaName: "TestSchema",
			isRequest:  pointer.From(true),
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureUploadStreams: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("request-stream",
				testutils.NewTypeDef(ast.DataTypeRequestStream).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithValidations(&ast.Validations{}).
					Build()).
				Build(),
		},
		{
			name: "StringWithResponseStream",
			schemasYAML: `TestSchema:
  type: string
  format: binary`,
			schemaName: "TestSchema",
			isRequest:  pointer.From(false),
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureDownloadStreams: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("response-stream",
				testutils.NewTypeDef(ast.DataTypeResponseStream).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithOutput(true).
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
				params.Depth = 1 // Set depth for stream detection
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
