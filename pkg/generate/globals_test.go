package generate

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleGlobals tests the handleGlobals method with various scenarios
func TestHandleGlobals(t *testing.T) {
	tests := []struct {
		name                string
		openAPIYAML         string
		globalNameOverrides []*extensions.NameOverride
		fixes               *config.Fixes
		supportedFeatures   map[features.Feature]bool
		enabledFeatures     map[features.Feature]bool
		expected            *ast.TypeDef
		expectError         bool
	}{
		{
			name: "NoGlobalsExtension",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			expected: nil,
		},
		{
			name: "EmptyGlobalsExtension",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters: []`,
			expected: nil,
		},
		{
			name: "SingleQueryParameter",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters:
    - name: apiKey
      in: query
      required: true
      schema:
        type: string
      description: API key for authentication`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("apiKey",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 12, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API key for authentication",
							}).
							Build()).
						WithOriginalName("apiKey").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       true,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 12, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API key for authentication",
								}).
								Build(),
							Name:                 "apiKey",
							ParamType:            ast.ParamTypeQueryParam,
							RequiredForOperation: true,
							Style:                "form",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API key for authentication",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
				).
				Build(),
		},
		{
			name: "SinglePathParameter",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters:
    - name: version
      in: path
      required: true
      schema:
        type: string
      description: API version`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("version",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 12, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API version",
							}).
							Build()).
						WithOriginalName("version").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       false,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 12, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API version",
								}).
								Build(),
							Name:                 "version",
							ParamType:            ast.ParamTypePathParam,
							RequiredForOperation: true,
							Style:                "simple",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API version",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
				).
				Build(),
		},
		{
			name: "SingleHeaderParameter",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters:
    - name: X-Client-Version
      in: header
      required: false
      schema:
        type: string
      description: Client version header`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("X-Client-Version",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 12, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "Client version header",
							}).
							Build()).
						WithOriginalName("X-Client-Version").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       false,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 12, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "Client version header",
								}).
								Build(),
							Name:                 "X-Client-Version",
							ParamType:            ast.ParamTypeHeader,
							RequiredForOperation: false,
							Style:                "simple",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "Client version header",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
				).
				Build(),
		},
		{
			name: "MultipleParameters",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters:
    - name: apiKey
      in: query
      required: true
      schema:
        type: string
      description: API key for authentication
    - name: version
      in: path
      required: true
      schema:
        type: string
      description: API version
    - name: X-Client-ID
      in: header
      required: false
      schema:
        type: string
      description: Client identifier`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("X-Client-ID",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 24, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "Client identifier",
							}).
							Build()).
						WithOriginalName("X-Client-ID").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       false,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 24, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "Client identifier",
								}).
								Build(),
							Name:                 "X-Client-ID",
							ParamType:            ast.ParamTypeHeader,
							RequiredForOperation: false,
							Style:                "simple",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "Client identifier",
						}).
						WithParameterIndex(pointer.From(2)).
						Build(),
					testutils.NewFieldDef("apiKey",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 12, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API key for authentication",
							}).
							Build()).
						WithOriginalName("apiKey").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       true,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 12, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API key for authentication",
								}).
								Build(),
							Name:                 "apiKey",
							ParamType:            ast.ParamTypeQueryParam,
							RequiredForOperation: true,
							Style:                "form",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API key for authentication",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
					testutils.NewFieldDef("version",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 18, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API version",
							}).
							Build()).
						WithOriginalName("version").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       false,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 18, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API version",
								}).
								Build(),
							Name:                 "version",
							ParamType:            ast.ParamTypePathParam,
							RequiredForOperation: true,
							Style:                "simple",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API version",
						}).
						WithParameterIndex(pointer.From(1)).
						Build(),
				).
				Build(),
		},
		{
			name: "ParameterWithComplexType",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
x-speakeasy-globals:
  parameters:
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        minimum: 1
        maximum: 100
      description: Maximum number of results`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("limit",
						testutils.NewTypeDef(ast.DataTypeInteger).
							WithLocation(&yaml.Node{Line: 12, Column: 9}).
							WithValidations(&ast.Validations{
								Minimum: pointer.From[float64](1),
								Maximum: pointer.From[float64](100),
							}).
							WithComments(&ast.Comment{
								Description: "Maximum number of results",
							}).
							Build()).
						WithOriginalName("limit").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       true,
							FieldType: testutils.NewTypeDef(ast.DataTypeInteger).
								WithLocation(&yaml.Node{Line: 12, Column: 9}).
								WithValidations(&ast.Validations{
									Minimum: pointer.From[float64](1),
									Maximum: pointer.From[float64](100),
								}).
								WithComments(&ast.Comment{
									Description: "Maximum number of results",
								}).
								Build(),
							Name:                 "limit",
							ParamType:            ast.ParamTypeQueryParam,
							RequiredForOperation: false,
							Style:                "form",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "Maximum number of results",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
				).
				Build(),
		},
		{
			name: "ParameterWithRef",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  parameters:
    ApiKeyParam:
      name: apiKey
      in: query
      required: true
      schema:
        type: string
      description: API key parameter from components
x-speakeasy-globals:
  parameters:
    - $ref: "#/components/parameters/ApiKeyParam"`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("apiKey",
						testutils.NewTypeDef(ast.DataTypeString).
							WithName("ApiKeyParam").
							WithOriginalName("").
							WithLocation(&yaml.Node{Line: 13, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API key parameter from components",
							}).
							Build()).
						WithOriginalName("apiKey").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       true,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithName("ApiKeyParam").
								WithOriginalName("").
								WithLocation(&yaml.Node{Line: 13, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API key parameter from components",
								}).
								Build(),
							Name:                 "apiKey",
							ParamType:            ast.ParamTypeQueryParam,
							RequiredForOperation: true,
							Style:                "form",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API key parameter from components",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
				).
				Build(),
		},
		{
			name: "MixedParametersWithRef",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  parameters:
    VersionParam:
      name: version
      in: path
      required: true
      schema:
        type: string
      description: API version from components
x-speakeasy-globals:
  parameters:
    - name: apiKey
      in: query
      required: true
      schema:
        type: string
      description: Direct API key parameter
    - $ref: "#/components/parameters/VersionParam"`,
			expected: testutils.NewTypeDef(ast.DataTypeClass).
				WithName("Globals").
				WithScope(ast.ScopeGlobals).
				WithOriginalNameFrozen().
				WithContextStack(ast.ContextStack{}).
				WithRegistered().
				WithFields(
					testutils.NewFieldDef("apiKey",
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 21, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "Direct API key parameter",
							}).
							Build()).
						WithOriginalName("apiKey").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       true,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithLocation(&yaml.Node{Line: 21, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "Direct API key parameter",
								}).
								Build(),
							Name:                 "apiKey",
							ParamType:            ast.ParamTypeQueryParam,
							RequiredForOperation: true,
							Style:                "form",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "Direct API key parameter",
						}).
						WithParameterIndex(pointer.From(0)).
						Build(),
					testutils.NewFieldDef("version",
						testutils.NewTypeDef(ast.DataTypeString).
							WithName("VersionParam").
							WithOriginalName("").
							WithLocation(&yaml.Node{Line: 13, Column: 9}).
							WithValidations(&ast.Validations{}).
							WithComments(&ast.Comment{
								Description: "API version from components",
							}).
							Build()).
						WithOriginalName("version").
						WithOptional(true).
						WithAnnotations(&ast.ParamAnnotation{
							AllowReserved: false,
							Explode:       false,
							FieldType: testutils.NewTypeDef(ast.DataTypeString).
								WithName("VersionParam").
								WithOriginalName("").
								WithLocation(&yaml.Node{Line: 13, Column: 9}).
								WithValidations(&ast.Validations{}).
								WithComments(&ast.Comment{
									Description: "API version from components",
								}).
								Build(),
							Name:                 "version",
							ParamType:            ast.ParamTypePathParam,
							RequiredForOperation: true,
							Style:                "simple",
							IsGlobal:             true,
						}).
						WithComments(&ast.Comment{
							Description: "API version from components",
						}).
						WithParameterIndex(pointer.From(1)).
						Build(),
				).
				Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using the helper
			supportedFeatures := tt.supportedFeatures
			enabledFeatures := tt.enabledFeatures

			if supportedFeatures == nil {
				supportedFeatures = make(map[features.Feature]bool)
				enabledFeatures = make(map[features.Feature]bool)
				// Support all features by default
				for i := 0; i < 100; i++ { // Rough estimate of feature count
					feature := features.Feature(i)
					supportedFeatures[feature] = true
					enabledFeatures[feature] = true
				}
			}

			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML:       tt.openAPIYAML,
				Fixes:             tt.fixes,
				SupportedFeatures: supportedFeatures,
				EnabledFeatures:   enabledFeatures,
			})
			require.NoError(t, err)

			// Debug: Check if extensions exist
			t.Logf("Document extensions: %v", setup.DocInfo.Doc.GetExtensions())

			// Debug: Try to get globals extension directly
			globalVariables, err := setup.Subsystem.Extensions.HandleGlobalsExtension(context.Background(), setup.DocInfo.Doc)
			switch {
			case err != nil:
				t.Logf("Error getting globals extension: %v", err)
			case globalVariables == nil:
				t.Logf("No global variables found")
			default:
				t.Logf("Found %d global parameters", len(globalVariables.Parameters))
			}

			// Create Generator from setup components
			schemasInstance := &schemas.Schemas{
				Config:    setup.Config,
				Target:    setup.Target,
				Subsystem: setup.Subsystem,
				Namer:     setup.Namer,
			}

			generator := &Generator{
				subsystem: setup.Subsystem,
				schemas:   schemasInstance,
			}
			a := ast.NewAST()
			a.MainSDK = ast.NewMainSDK(nil)

			// Call handleGlobals
			result, err := generator.handleGlobals(context.Background(), setup.DocInfo.Doc, a, tt.globalNameOverrides, setup.DocInfo)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			if result == nil {
				t.Logf("Result is nil for test case: %s", tt.name)
				t.Logf("OpenAPI YAML: %s", tt.openAPIYAML)
			}

			require.NotNil(t, result)
			assert.NotNil(t, a.MainSDK.Globals)

			// Use deep comparison
			isEqual := testutils.DeepCompare(t, tt.expected, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
