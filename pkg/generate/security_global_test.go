package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/require"
)

type globalSecurityTestCase struct {
	securityFixesFeb2025 bool
	hoistGlobalSecurity  bool
	transform            func(*ast.Security) *ast.Security // Optional transform function
}

type globalSecurityTest struct {
	name             string // Name will be concatenated with securityFixesFeb2025 true or false
	openAPIYAML      string
	fixes            *config.Fixes
	mockFeatures     map[features.Feature]bool
	expectedSecurity *ast.Security // Base expected result
	testCases        []globalSecurityTestCase
}

// Helper function to create global context stack
func newGlobalContextStack() ast.ContextStack {
	return ast.ContextStack{}
}

// newGlobalSecurityClass creates a standard "Security" class TypeDef with the given fields for global security
func newGlobalSecurityClass(fields ...*ast.FieldDef) *ast.TypeDef {
	return testutils.NewTypeDef(ast.DataTypeClass).
		WithName("Security").
		WithOriginalName("Security").
		WithOriginalNameFrozen().
		WithContextStack(newGlobalContextStack()).
		WithScope(ast.ScopeShared).
		WithRegistered().
		WithFields(fields...).
		Build()
}

// newGlobalSecurityWrapper creates the main Security field wrapper for global security
func newGlobalSecurityWrapper(securityClass *ast.TypeDef, optional bool) *ast.FieldDef {
	return testutils.NewFieldDef("Security", securityClass).
		WithOriginalName("security").
		WithOptional(optional).
		WithAnnotations(&ast.NeedsCasingAnnotation{}).
		Build()
}

// TestHandleGlobalSecurity tests the handleGlobalSecurity method with various global-level security scenarios
func TestHandleGlobalSecurity(t *testing.T) {
	tests := []globalSecurityTest{
		{
			name: "NoGlobalSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			expectedSecurity: nil,
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "BasicGlobalSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false),
					), false)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "GlobalSecurityWithOperationDisabled",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
paths:
  /test1:
    get:
      security: []  # Explicitly disable security
      responses:
        '200':
          description: OK
  /test2:
    get:
      security:
        - {}  # Empty security requirement (also disables security)
      responses:
        '200':
          description: OK
  /test3:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, false),
					), true)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonOperationOverride).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "GlobalSecurityWithOperationOverride",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
paths:
  /test:
    get:
      security:
        - BasicAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
    BasicAuth:
      type: http
      scheme: basic
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, false),
					), true)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonOperationOverride).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "GlobalSecurityOptional",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
  - {}  # Makes security optional
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, false),
					), true)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonOptionalScheme).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "MultipleGlobalSecurityOptions",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
  - BasicAuth: []
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
    BasicAuth:
      type: http
      scheme: basic
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, true),
						newBasicAuthField(true, true),
					),
					false)).
				WithSchemes("ApiKeyAuth", "BasicAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "CombinedGlobalSecurityRequirement",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
    BasicAuth: []
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
    BasicAuth:
      type: http
      scheme: basic
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						withComposite(newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false)),
						withComposite(newBasicAuthField(false, false)),
					), false)).
				WithSchemes("ApiKeyAuth-BasicAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "OAuth2GlobalSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - OAuth2: [read, write]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OAuth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/oauth/authorize
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("OAuth2", "OAuth2", "Authorization", "oauth2", "", "OAuth2", false, false),
					), false)).
				WithSchemes("OAuth2").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				WithOAuth2Config("OAuth2", "authorization_code", "", []string{"read", "write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "GlobalSecurityWithEnvVarFeature",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ApiKeyAuth: []
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			mockFeatures: map[features.Feature]bool{
				features.FeatureEnvVarSecurityUsage: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, false),
					), true)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonEnvVar).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
			},
		},
		{
			name: "GlobalSecurityHoistingFromOperations",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test1:
    get:
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: OK
  /test2:
    get:
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: OK
  /test3:
    get:
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`,
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newGlobalSecurityWrapper(
					newGlobalSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false),
					), false)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  false,
					transform: func(security *ast.Security) *ast.Security {
						// When hoisting is disabled, no global security should be created
						return nil
					},
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  false,
					transform: func(security *ast.Security) *ast.Security {
						// When hoisting is disabled, no global security should be created
						return nil
					},
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsGlobal",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - OAuth2ClientCredentials: [read, write]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OAuth2ClientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
`,
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			fixes: &config.Fixes{
				SecurityFeb2025: false, // Will be overridden by test case
			},
			expectedSecurity: func() *ast.Security {
				// Create base security class with TokenURL having empty annotations
				securityClass := newGlobalSecurityClass(
					testutils.NewFieldDef("ClientID",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow client identifier"}).
							Build()).
						WithOriginalName("clientID").
						WithAnnotations(testutils.NewSecurityAnnotation(
							"clientID", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
						Build(),
					testutils.NewFieldDef("ClientSecret",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow client secret"}).
							Build()).
						WithOriginalName("clientSecret").
						WithAnnotations(testutils.NewSecurityAnnotation(
							"clientSecret", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
						Build(),
					testutils.NewFieldDef("TokenURL",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow token URL"}).
							Build()).
						WithOriginalName("tokenURL").
						WithDefault(&ast.AnyValue{Value: "https://example.com/oauth/token"}).
						Build(),
				)

				// Set annotations on TokenURL field to match actual behavior
				securityClass.Fields[2].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}

				return testutils.NewSecurity().
					WithSecurity(newGlobalSecurityWrapper(securityClass, false)).
					WithSchemes("OAuth2ClientCredentials").
					WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "", []string{"read", "write"},
						[]ast.OAuth2Scope{
							{Name: "read", Comments: ast.Comment{Description: "Read access"}},
							{Name: "write", Comments: ast.Comment{Description: "Write access"}},
						}).
					WithOptionalityReason(ast.SecOptReasonNotOptional).
					Build()
			}(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: false,
					hoistGlobalSecurity:  true,
				},
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
					transform: func(security *ast.Security) *ast.Security {
						// SecurityFeb2025 version adds extensions to the security class
						security.Security.Type.Extensions.All = map[string]any{
							"x-speakeasy-token-endpoint-authentication": "client_secret_post",
						}
						return security
					},
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsGlobalWithOverridableScopes",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - OAuth2ClientCredentials: [read, write]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OAuth2ClientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
          x-speakeasy-overridable-scopes: true
`,
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			fixes: &config.Fixes{
				SecurityFeb2025: true,
			},
			expectedSecurity: func() *ast.Security {
				// Create base security class with Scopes field since overridable-scopes is enabled
				securityClass := newGlobalSecurityClass(
					testutils.NewFieldDef("ClientID",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow client identifier"}).
							Build()).
						WithOriginalName("clientID").
						WithAnnotations(testutils.NewSecurityAnnotation(
							"clientID", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
						Build(),
					testutils.NewFieldDef("ClientSecret",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow client secret"}).
							Build()).
						WithOriginalName("clientSecret").
						WithAnnotations(testutils.NewSecurityAnnotation(
							"clientSecret", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
						Build(),
					testutils.NewFieldDef("TokenURL",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow token URL"}).
							Build()).
						WithOriginalName("tokenURL").
						WithDefault(&ast.AnyValue{Value: "https://example.com/oauth/token"}).
						Build(),
					testutils.NewFieldDef("Scopes",
						testutils.NewTypeDefWithoutValidations(ast.DataTypeArray).
							WithItemType(testutils.NewTypeDefWithoutValidations(ast.DataTypeString).Build()).
							WithComments(&ast.Comment{Description: "OAuth2 Client Credentials Flow scopes override (optional)"}).
							Build()).
						WithOriginalName("scopes").
						WithOptional(true).
						Build(),
				)

				// Set annotations on Scopes and TokenURL fields to match actual behavior
				securityClass.Fields[2].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}
				securityClass.Fields[3].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}

				return testutils.NewSecurity().
					WithSecurity(newGlobalSecurityWrapper(securityClass, false)).
					WithSchemes("OAuth2ClientCredentials").
					WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "", []string{"read", "write"},
						[]ast.OAuth2Scope{
							{Name: "read", Comments: ast.Comment{Description: "Read access"}},
							{Name: "write", Comments: ast.Comment{Description: "Write access"}},
						}).
					WithOptionalityReason(ast.SecOptReasonNotOptional).
					Build()
			}(),
			testCases: []globalSecurityTestCase{
				{
					securityFixesFeb2025: true,
					hoistGlobalSecurity:  true,
					transform: func(security *ast.Security) *ast.Security {
						// SecurityFeb2025 version adds extensions to the security class
						security.Security.Type.Extensions.All = map[string]any{
							"x-speakeasy-token-endpoint-authentication": "client_secret_post",
						}

						// Mark that scopes are overridable in extensions
						security.Security.Type.Extensions.OverridableOAuth2Scopes = true
						return security
					},
				},
			},
		},
	}

	for _, tt := range tests {
		for _, tc := range tt.testCases {
			testName := tt.name
			if tc.securityFixesFeb2025 {
				testName += "_SecurityFeb2025"
			} else {
				testName += "_Legacy"
			}
			if tc.hoistGlobalSecurity {
				testName += "_HoistEnabled"
			} else {
				testName += "_HoistDisabled"
			}
			t.Run(testName, func(t *testing.T) {
				runGlobalSecurityTest(t, tt, tc)
			})
		}
	}
}

func runGlobalSecurityTest(t *testing.T, tt globalSecurityTest, tc globalSecurityTestCase) {
	t.Helper()

	setup := setupSecurityTest(t, tt.openAPIYAML, tt.fixes, tt.mockFeatures, tc.securityFixesFeb2025, tc.hoistGlobalSecurity)
	a := ast.NewAST()
	a.MainSDK = ast.NewMainSDK(nil)

	security, _, err := setup.generator.handleGlobalSecurity(setup.ctx, setup.docInfo, a)

	require.NoError(t, err)

	assertSecurityResult(t, tt.name, tt.expectedSecurity, security, tc.transform)
}
