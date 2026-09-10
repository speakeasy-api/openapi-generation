package generate

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/register"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/security"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type operationSecurityTestCase struct {
	securityFixesFeb2025 bool
	transform            func(*ast.Security) *ast.Security // Optional transform function
}

type operationSecurityTest struct {
	name             string // Name will be concatenated with securityFixesFeb2025 true or false
	openAPIYAML      string
	operationPath    string
	operationMethod  string
	fixes            *config.Fixes
	mockFeatures     map[features.Feature]bool
	expectedSecurity *ast.Security // Base expected result
	testCases        []operationSecurityTestCase
}

// Helper function to create operation context stack
func newOperationContextStack(operationName string) ast.ContextStack {
	return ast.ContextStack{
		{
			Type:       ast.ContextTypeOperation,
			Identifier: operationName,
			Used:       true,
		},
	}
}

// Helper methods for common security test patterns

// newSecurityClass creates a standard "Test_Security" class TypeDef with the given fields
func newSecurityClass(fields ...*ast.FieldDef) *ast.TypeDef {
	return testutils.NewTypeDef(ast.DataTypeClass).
		WithName("Test_Security").
		WithOriginalName("Test_Security").
		WithOriginalNameFrozen().
		WithContextStack(newOperationContextStack("Test")).
		WithScope(ast.ScopeShared).
		WithRegistered().
		WithFields(fields...).
		Build()
}

// newStringSecurityField creates a string-based security field (ApiKey, OAuth2, etc.)
func newStringSecurityField(name, originalName, fieldName, secType, subType, schemeKey string, optional, securityOption bool) *ast.FieldDef {
	// Determine the description based on the security type and subtype
	var description string
	switch secType {
	case "apiKey":
		description = "API Key"
	case "http":
		if subType == "bearer" {
			description = "HTTP Bearer"
		}
	case "oauth2":
		switch name {
		case "ClientID":
			description = "OAuth2 Client ID"
		case "ClientSecret":
			description = "OAuth2 Client Secret"
		default:
			description = "OAuth2 Authorization"
		}
	case "openIdConnect":
		description = "OpenID Connect Authorization"
	}

	typeDef := testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
		WithExamples()

	// Add comments if we have a description
	if description != "" {
		typeDef = typeDef.WithComments(&ast.Comment{Description: description})
	}

	return testutils.NewFieldDef(name,
		typeDef.Build()).
		WithOriginalName(originalName).
		WithOptional(optional).
		WithAnnotations(testutils.NewSecurityAnnotation(
			fieldName, secType, subType, schemeKey, true, securityOption), &ast.NeedsCasingAnnotation{}).
		Build()
}

func withComposite(f *ast.FieldDef) *ast.FieldDef {
	f.Annotations.Get(ast.AnnotationTypeSecurity).(*ast.SecurityAnnotation).Composite = true
	return f
}

// newBasicAuthClass creates a BasicAuth class with Username and Password fields
func newBasicAuthClass() *ast.TypeDef {
	return testutils.NewTypeDefWithoutValidations(ast.DataTypeClass).
		WithName("SchemeBasicAuth").
		WithOriginalName("SchemeBasicAuth").
		WithOriginalNameFrozen().
		WithContextStack(
			testutils.NewContextStack().
				WithRefName("SchemeBasicAuth").
				WithIdentifierForNaming(nil).
				Build()).
		WithScope(ast.ScopeShared).
		WithIsComponent(true).
		WithRegistered().
		WithComments(&ast.Comment{Description: ""}).
		WithFields(
			testutils.NewFieldDef("Username",
				testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind: yaml.ScalarNode,
							Tag:  "!!str",
						},
					}).
					WithComments(&ast.Comment{Description: "HTTP Basic username"}).
					Build()).
				WithOriginalName("username").
				WithAnnotations(&ast.SecurityAnnotation{FieldName: "username"}, &ast.NeedsCasingAnnotation{}).
				Build(),
			testutils.NewFieldDef("Password",
				testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind: yaml.ScalarNode,
							Tag:  "!!str",
						},
					}).
					WithComments(&ast.Comment{Description: "HTTP Basic password"}).
					Build()).
				WithOriginalName("password").
				WithAnnotations(&ast.SecurityAnnotation{FieldName: "password"}, &ast.NeedsCasingAnnotation{}).
				Build(),
		).
		Build()
}

// newBasicAuthField creates a BasicAuth field with the given optionality and security option
func newBasicAuthField(optional, securityOption bool) *ast.FieldDef {
	return testutils.NewFieldDef("BasicAuth", newBasicAuthClass()).
		WithOriginalName("BasicAuth").
		WithOptional(optional).
		WithAnnotations(testutils.NewSecurityAnnotation(
			"", "http", "basic", "BasicAuth", true, securityOption), &ast.NeedsCasingAnnotation{}).
		Build()
}

// newSecurityWrapper creates the main Security field wrapper
func newSecurityWrapper(securityClass *ast.TypeDef, optional bool) *ast.FieldDef {
	return testutils.NewFieldDef("Security", securityClass).
		WithOriginalName("security").
		WithOptional(optional).
		WithAnnotations(&ast.OperationSecurityAnnotation{}, &ast.NeedsCasingAnnotation{}).
		Build()
}

func newEnterpriseContext() context.Context {
	return generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
}

// TestHandleSecurity tests the handleSecurity method with various operation-level security scenarios
func TestHandleSecurity(t *testing.T) {
	tests := []operationSecurityTest{
		{
			name: "NoOperationSecurity",
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
			operationPath:    "/test",
			operationMethod:  "get",
			expectedSecurity: nil,
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "BasicOperationSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false),
					), false)).
				WithSchemes("ApiKeyAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "DisabledOperationSecurity",
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
      security: []  # Explicitly disable security
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithDisabled(true).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "EmptyObjectOperationSecurityDisables",
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
        - {}
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithDisabled(true).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "MultipleEmptyObjectsOperationSecurityDisables",
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
        - {}
        - {}
        - {}
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithDisabled(true).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "MultipleOperationSecurityOptions",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - ApiKeyAuth: []
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, true),
						newBasicAuthField(true, true),
					), false)).
				WithSchemes("ApiKeyAuth", "BasicAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "CombinedOperationSecurityRequirement",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - ApiKeyAuth: []
          BasicAuth: []
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						withComposite(newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false)),
						withComposite(newBasicAuthField(false, false)),
					), false)).
				WithSchemes("ApiKeyAuth-BasicAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OAuth2AuthorizationCodeOperation",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - OAuth2: [read, write]
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
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
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsOperation",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - OAuth2ClientCredentials: [read, write]
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OAuth2ClientCredentials:
      description: Client Credentials
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			fixes: &config.Fixes{
				SecurityFeb2025: false, // Will be overridden by test case
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "Client Credentials", []string{"read", "write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
					transform: func(security *ast.Security) *ast.Security {
						// SecurityFeb2025 version creates full OAuth2 client credentials structure
						securityClass := newSecurityClass(
							testutils.NewFieldDef("ClientID",
								testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
									WithComments(&ast.Comment{Description: "Client Credentials client identifier"}).
									Build()).
								WithOriginalName("clientID").
								WithAnnotations(testutils.NewSecurityAnnotation(
									"clientID", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
								Build(),
							testutils.NewFieldDef("ClientSecret",
								testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
									WithComments(&ast.Comment{Description: "Client Credentials client secret"}).
									Build()).
								WithOriginalName("clientSecret").
								WithAnnotations(testutils.NewSecurityAnnotation(
									"clientSecret", "oauth2", "client_credentials", "OAuth2ClientCredentials", true, false), &ast.NeedsCasingAnnotation{}).
								Build(),
							testutils.NewFieldDef("TokenURL",
								testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
									WithComments(&ast.Comment{Description: "Client Credentials token URL"}).
									Build()).
								WithOriginalName("tokenURL").
								WithDefault(&ast.AnyValue{Value: "https://example.com/oauth/token"}).
								Build(),
						)

						// Set annotations on TokenURL field to match actual behavior
						securityClass.Fields[2].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}

						// Add extensions to the security class
						securityClass.Extensions.All = map[string]any{
							"x-speakeasy-token-endpoint-authentication": "client_secret_post",
						}

						return testutils.NewSecurity().
							WithSecurity(newSecurityWrapper(securityClass, false)).
							WithSchemes("OAuth2ClientCredentials").
							WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "Client Credentials", []string{"read", "write"},
								[]ast.OAuth2Scope{
									{Name: "read", Comments: ast.Comment{Description: "Read access"}},
									{Name: "write", Comments: ast.Comment{Description: "Write access"}},
								}).
							WithOptionalityReason(ast.SecOptReasonNotOptional).
							Build()
					},
				},
			},
		},
		{
			name: "BearerTokenOperationSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - BearerAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
`,
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("BearerAuth", "BearerAuth", "Authorization", "http", "bearer", "BearerAuth", false, false),
					), false)).
				WithSchemes("BearerAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OpenIdConnectOperationSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - OpenIdConnect: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OpenIdConnect:
      type: openIdConnect
      openIdConnectUrl: https://example.com/.well-known/openid_configuration
`,
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("OpenIdConnect", "OpenIdConnect", "Authorization", "openIdConnect", "", "OpenIdConnect", false, false),
					), false)).
				WithSchemes("OpenIdConnect").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OperationOverridesGlobalSecurity",
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						testutils.NewFieldDef("Username",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind: yaml.ScalarNode,
										Tag:  "!!str",
									},
								}).
								WithComments(&ast.Comment{Description: "HTTP Basic username"}).
								Build()).
							WithOriginalName("username").
							WithAnnotations(&ast.SecurityAnnotation{
								FieldName: "username",
								SecType:   "http",
								SubType:   "basic",
								Scheme:    true,
								SchemeKey: "BasicAuth",
							}, &ast.NeedsCasingAnnotation{}).
							Build(),
						testutils.NewFieldDef("Password",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind: yaml.ScalarNode,
										Tag:  "!!str",
									},
								}).
								WithComments(&ast.Comment{Description: "HTTP Basic password"}).
								Build()).
							WithOriginalName("password").
							WithAnnotations(&ast.SecurityAnnotation{
								FieldName: "password",
								SecType:   "http",
								SubType:   "basic",
								Scheme:    true,
								SchemeKey: "BasicAuth",
							}, &ast.NeedsCasingAnnotation{}).
							Build(),
					), false)).
				WithSchemes().
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
					transform: func(security *ast.Security) *ast.Security {
						// SecurityFeb2025 populates Requirements with resolved scheme keys
						result := *security // Copy the struct
						result.Requirements = []ast.SecurityRequirement{{ast.SecurityScheme("BasicAuth")}}
						return &result
					},
				},
			},
		},
		{
			name: "OperationSecurityWithGlobalSchemes",
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
        - ApiKeyAuth: []  # Same as global, should be hoisted
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithHoistedSecurityConfig(true, []ast.HoistedSecurityField{
					{Name: "ApiKeyAuth", Index: 0, Group: 0},
				}).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "MissingOperationSecurityScheme",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - NonExistentAuth: []
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("NonExistentAuth", "NonExistentAuth", "Authorization", "apiKey", "header", "NonExistentAuth", false, false),
					), false)).
				WithSchemes("NonExistentAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OptionalOperationSecurity",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - ApiKeyAuth: []
        - {}  # Makes security optional
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", true, false),
					), true)).
				WithRequirements(ast.SecurityRequirement{ast.SecurityScheme("ApiKeyAuth")}).
				WithOptionalityReason(ast.SecOptReasonOptionalScheme).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "ZeroOptionsSecurityWithSchemes",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - {}
        - {}
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
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithDisabled(true).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OAuth2PasswordFlowOperation",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - OAuth2Password: [read, write]
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    OAuth2Password:
      description: Resource Owner Password Flow
      type: oauth2
      flows:
        password:
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2Password: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("OAuth2Password", "password", "Resource Owner Password Flow", []string{"read", "write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
					transform: func(security *ast.Security) *ast.Security {
						// SecurityFeb2025 version creates complex OAuth2 password union structure
						// Step 1: Create the OAuth2Password_Token type (simple string)
						tokenType := testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
							WithName("OAuth2Password_Token").
							WithOriginalName("Token").
							WithComments(&ast.Comment{Description: "Resource Owner Password Flow token"}).
							Build()

						// Step 2: Create the OAuth2Password_Credentials type (complex class with 5 fields)
						credentialsType := testutils.NewTypeDefWithoutValidations(ast.DataTypeClass).
							WithName("OAuth2Password_Credentials").
							WithOriginalName("Credentials").
							WithOriginalNameFrozen().
							WithContextStack(ast.ContextStack{}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithIsComponent(true).
							WithComments(&ast.Comment{Description: "Resource Owner Password Flow credentials"}).
							WithFields(
								// Username field
								testutils.NewFieldDef("Username",
									testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
										WithComments(&ast.Comment{Description: "Resource Owner Password Flow username"}).
										Build()).
									WithOriginalName("username").
									WithAnnotations(&ast.SecurityAnnotation{
										FieldName: "username",
									}, &ast.NeedsCasingAnnotation{}).
									Build(),
								// Password field
								testutils.NewFieldDef("Password",
									testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
										WithComments(&ast.Comment{Description: "Resource Owner Password Flow password"}).
										Build()).
									WithOriginalName("password").
									WithAnnotations(&ast.SecurityAnnotation{
										FieldName: "password",
									}, &ast.NeedsCasingAnnotation{}).
									Build(),
								// ClientID field (optional)
								testutils.NewFieldDef("ClientID",
									testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
										WithComments(&ast.Comment{Description: "Resource Owner Password Flow client identifier"}).
										Build()).
									WithOriginalName("clientID").
									WithOptional(true).
									WithAnnotations(&ast.SecurityAnnotation{
										FieldName: "clientID",
									}, &ast.NeedsCasingAnnotation{}).
									Build(),
								// ClientSecret field (optional)
								testutils.NewFieldDef("ClientSecret",
									testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
										WithComments(&ast.Comment{Description: "Resource Owner Password Flow client secret"}).
										Build()).
									WithOriginalName("clientSecret").
									WithOptional(true).
									WithAnnotations(&ast.SecurityAnnotation{
										FieldName: "clientSecret",
									}, &ast.NeedsCasingAnnotation{}).
									Build(),
								// TokenURL field
								testutils.NewFieldDef("TokenURL",
									testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
										WithComments(&ast.Comment{Description: "Resource Owner Password Flow token URL"}).
										Build()).
									WithOriginalName("tokenURL").
									WithAnnotations(&ast.SecurityAnnotation{
										FieldName: "tokenURL",
									}, &ast.NeedsCasingAnnotation{}).
									WithDefault(&ast.AnyValue{Value: "https://example.com/oauth/token"}).
									Build(),
							).
							Build()

						// Step 3: Create the union type OAuth2Password_Input
						unionType := testutils.NewTypeDefWithoutValidations(ast.DataTypeUnion).
							WithName("OAuth2Password_Input").
							WithOriginalName("OAuth2Password_Input").
							WithOriginalNameFrozen().
							WithContextStack(ast.ContextStack{}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithIsComponent(true).
							WithAssociatedTypes(credentialsType, tokenType).
							WithComments(&ast.Comment{Description: "Resource Owner Password Flow"}).
							Build()

						// Step 4: Create the OAuth2Password field with the union type
						oauth2Field := testutils.NewFieldDef("OAuth2Password", unionType).
							WithOriginalName("OAuth2Password").
							WithAnnotations(testutils.NewSecurityAnnotation(
								"Authorization", "oauth2", "password", "OAuth2Password", true, false), &ast.NeedsCasingAnnotation{}).
							Build()

						// Step 5: Create the security class with the OAuth2Password field
						securityClass := newSecurityClass(oauth2Field)

						return testutils.NewSecurity().
							WithSecurity(newSecurityWrapper(securityClass, false)).
							WithSchemes("OAuth2Password").
							WithOAuth2Config("OAuth2Password", "password", "Resource Owner Password Flow", []string{"read", "write"},
								[]ast.OAuth2Scope{
									{Name: "read", Comments: ast.Comment{Description: "Read access"}},
									{Name: "write", Comments: ast.Comment{Description: "Write access"}},
								}).
							WithOptionalityReason(ast.SecOptReasonNotOptional).
							Build()
					},
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsHoisted",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - OAuth2ClientCredentials: [read, write]
  - BearerAuth: []
paths:
  /test:
    get:
      security:
        - OAuth2ClientCredentials: [read, write]  # Subset of global - only OAuth2 hoisted, not BearerAuth
      responses:
        '200':
          description: OK
  /test2:
    get:
      operationId: bothSecurities
      security:
        - OAuth2ClientCredentials: [read, write]
        - BearerAuth: []
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
    BearerAuth:
      type: http
      scheme: bearer
`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true, // Enable OAuth2 flow to capture scopes
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "", []string{"read", "write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithHoistedSecurityConfig(false, []ast.HoistedSecurityField{
					{Name: "OAuth2ClientCredentials", Index: 0},
				}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsBothSecuritiesHoistedDifferentOrder",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - OAuth2ClientCredentials: [read, write]
  - BearerAuth: []
paths:
  /test:
    get:
      operationId: bothSecurities
      security:
        - BearerAuth: []
        - OAuth2ClientCredentials: [read, write]
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
    BearerAuth:
      type: http
      scheme: bearer
`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "", []string{"read", "write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithHoistedSecurityConfig(false, []ast.HoistedSecurityField{
					{Name: "BearerAuth", Index: 1, Group: 1},
					{Name: "OAuth2ClientCredentials", Index: 0, Group: 0},
				}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "MultipleOptionsWithComplexRequirements",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - ApiKeyAuth: []
          BasicAuth: []  # Combined requirement (AND)
        - BearerAuth: []   # Alternative requirement (OR)
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
    BearerAuth:
      type: http
      scheme: bearer
`,
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						testutils.NewFieldDef("Option1",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeClass).
								WithName("Test_SecurityOption1").
								WithOriginalName("Test_SecurityOption1").
								WithOriginalNameFrozen().
								WithContextStack(newOperationContextStack("Test")).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithFields(
									withComposite(newStringSecurityField("ApiKeyAuth", "ApiKeyAuth", "X-API-Key", "apiKey", "header", "ApiKeyAuth", false, false)),
									withComposite(newBasicAuthField(false, false)),
								).
								Build()).
							WithOriginalName("Option1").
							WithOptional(true).
							WithAnnotations(&ast.SecurityAnnotation{Option: true, SecurityOption: true}, &ast.NeedsCasingAnnotation{}).
							Build(),
						testutils.NewFieldDef("Option2",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeClass).
								WithName("Test_SecurityOption2").
								WithOriginalName("Test_SecurityOption2").
								WithOriginalNameFrozen().
								WithContextStack(newOperationContextStack("Test")).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithFields(
									newStringSecurityField("BearerAuth", "BearerAuth", "Authorization", "http", "bearer", "BearerAuth", false, false),
								).
								Build()).
							WithOriginalName("Option2").
							WithOptional(true).
							WithAnnotations(&ast.SecurityAnnotation{Option: true, SecurityOption: true}, &ast.NeedsCasingAnnotation{}).
							Build(),
					), false)).
				WithSchemes("ApiKeyAuth-BasicAuth", "BearerAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "CustomSecurityScheme",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - CustomAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    CustomAuth:
      type: http
      scheme: custom
      x-speakeasy-custom-security-scheme:
        schema:
          type: object
          properties:
            customField:
              type: string
          required:
            - customField
`,
			operationPath:   "/test",
			operationMethod: "get",
			fixes: &config.Fixes{
				SecurityFeb2025: false, // Will be overridden by test case
			},
			mockFeatures: map[features.Feature]bool{
				features.FeatureCustomSecuritySchemes: true,
				features.FeatureSDKHooks:              true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						testutils.NewFieldDef("CustomAuth",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithValidations(&ast.Validations{}).
								WithLocation(&yaml.Node{Line: 24, Column: 15}).
								Build()).
							WithOriginalName("CustomAuth").
							WithAnnotations(
								&ast.JSONAnnotation{FieldName: "customField"},
								&ast.SecurityAnnotation{
									FieldName: "customField",
									SecType:   "http",
									SubType:   "custom",
									Scheme:    true,
									SchemeKey: "CustomAuth",
								},
								&ast.NeedsCasingAnnotation{}).
							Build(),
					), false)).
				WithSchemes("CustomAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "SecurityWithExamples",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - ApiKeyAuth: []
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
      x-speakeasy-example: "test-api-key"
    BasicAuth:
      type: http
      scheme: basic
      x-speakeasy-example: "testuser;testpass"
`,
			operationPath:   "/test",
			operationMethod: "get",
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						testutils.NewFieldDef("ApiKeyAuth",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind:   yaml.ScalarNode,
										Style:  yaml.DoubleQuotedStyle,
										Tag:    "!!str",
										Value:  "test-api-key",
										Line:   21,
										Column: 28,
									},
								}).
								WithComments(&ast.Comment{Description: "API Key"}).
								Build()).
							WithOriginalName("ApiKeyAuth").
							WithOptional(true).
							WithAnnotations(testutils.NewSecurityAnnotation(
								"X-API-Key", "apiKey", "header", "ApiKeyAuth", true, true), &ast.NeedsCasingAnnotation{}).
							Build(),
						testutils.NewFieldDef("BasicAuth",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeClass).
								WithName("SchemeBasicAuth").
								WithOriginalName("SchemeBasicAuth").
								WithOriginalNameFrozen().
								WithContextStack(
									testutils.NewContextStack().
										WithRefName("SchemeBasicAuth").
										WithIdentifierForNaming(nil).
										Build()).
								WithScope(ast.ScopeShared).
								WithIsComponent(true).
								WithRegistered().
								WithComments(&ast.Comment{Description: ""}).
								WithFields(
									testutils.NewFieldDef("Username",
										testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
											WithExamples(&ast.Example{
												Value: &yaml.Node{
													Kind:  yaml.ScalarNode,
													Tag:   "!!str",
													Value: "testuser",
												},
											}).
											WithComments(&ast.Comment{Description: "HTTP Basic username"}).
											Build()).
										WithOriginalName("username").
										WithAnnotations(&ast.SecurityAnnotation{FieldName: "username"}, &ast.NeedsCasingAnnotation{}).
										Build(),
									testutils.NewFieldDef("Password",
										testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
											WithExamples(&ast.Example{
												Value: &yaml.Node{
													Kind:  yaml.ScalarNode,
													Tag:   "!!str",
													Value: "testpass",
												},
											}).
											WithComments(&ast.Comment{Description: "HTTP Basic password"}).
											Build()).
										WithOriginalName("password").
										WithAnnotations(&ast.SecurityAnnotation{FieldName: "password"}, &ast.NeedsCasingAnnotation{}).
										Build(),
								).
								Build()).
							WithOriginalName("BasicAuth").
							WithOptional(true).
							WithAnnotations(testutils.NewSecurityAnnotation(
								"", "http", "basic", "BasicAuth", true, true), &ast.NeedsCasingAnnotation{}).
							Build(),
					), false)).
				WithSchemes("ApiKeyAuth", "BasicAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "CustomSecuritySchemeNotSupported",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - CustomAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    CustomAuth:
      type: http
      scheme: custom
      x-speakeasy-custom-security-scheme:
        schema:
          type: object
          properties:
            customField:
              type: string
          required:
            - customField
`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureCustomSecuritySchemes: false, // Not supported - should fallback
				features.FeatureSDKHooks:              false,
			},
			expectedSecurity: testutils.NewSecurity().
				WithSecurity(newSecurityWrapper(
					newSecurityClass(
						testutils.NewFieldDef("Username",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind: yaml.ScalarNode,
										Tag:  "!!str",
									},
								}).
								WithComments(&ast.Comment{Description: "HTTP Basic username"}).
								Build()).
							WithOriginalName("username").
							WithAnnotations(&ast.SecurityAnnotation{
								FieldName: "username",
								SecType:   "http",
								SubType:   "custom",
								Scheme:    true,
								SchemeKey: "CustomAuth",
							}, &ast.NeedsCasingAnnotation{}).
							Build(),
						testutils.NewFieldDef("Password",
							testutils.NewTypeDefWithoutValidations(ast.DataTypeString).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind: yaml.ScalarNode,
										Tag:  "!!str",
									},
								}).
								WithComments(&ast.Comment{Description: "HTTP Basic password"}).
								Build()).
							WithOriginalName("password").
							WithAnnotations(&ast.SecurityAnnotation{
								FieldName: "password",
								SecType:   "http",
								SubType:   "custom",
								Scheme:    true,
								SchemeKey: "CustomAuth",
							}, &ast.NeedsCasingAnnotation{}).
							Build(),
					), false)).
				WithSchemes("CustomAuth").
				WithOptionalityReason(ast.SecOptReasonNotOptional).
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "OAuth2ClientCredentialsOperationWithOverridableScopes",
			openAPIYAML: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      security:
        - OAuth2ClientCredentials: [read, write]
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
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			fixes: &config.Fixes{
				SecurityFeb2025: true,
			},
			expectedSecurity: func() *ast.Security {
				// Create security class with Scopes field since overridable-scopes is enabled
				securityClass := newSecurityClass(
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

				// Set annotations on TokenURL and Scopes fields to match actual behavior
				securityClass.Fields[2].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}
				securityClass.Fields[3].Annotations = ast.Annotations{&ast.NeedsCasingAnnotation{}}

				// Add extensions to the security class
				securityClass.Extensions.All = map[string]any{
					"x-speakeasy-token-endpoint-authentication": "client_secret_post",
				}
				securityClass.Extensions.OverridableOAuth2Scopes = true

				return testutils.NewSecurity().
					WithSecurity(newSecurityWrapper(securityClass, false)).
					WithSchemes("OAuth2ClientCredentials").
					WithOAuth2Config("OAuth2ClientCredentials", "client_credentials", "", []string{"read", "write"},
						[]ast.OAuth2Scope{
							{Name: "read", Comments: ast.Comment{Description: "Read access"}},
							{Name: "write", Comments: ast.Comment{Description: "Write access"}},
						}).
					WithOptionalityReason(ast.SecOptReasonNotOptional).
					Build()
			}(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "HoistedSecurityOptions",
			openAPIYAML: `
openapi: 3.1.0
info:
  version: 1.0.0
security:
  - authA1: []
    authA2: []
  - authB: []
  - authC1: [read]
    authC2: []
paths:
  /test:
    get:
      security:
        - authC1: [write]
          authC2: []
        - authA2: []
          authA1: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    authA1:
      type: apiKey
      in: header
      name: X-Auth-1
    authA2:
      type: http
      scheme: basic
    authB:
      type: http
      scheme: bearer
      x-speakeasy-example: <YOUR_JWT>
    authC1:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: /oauth2/authorize
          scopes:
            read: Read access
            write: Write access
    authC2:
      type: apiKey
      in: header
      name: X-Project-Id`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("authC1", "client_credentials", "", []string{"write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithHoistedSecurityConfig(false, []ast.HoistedSecurityField{
					{Name: "Option3", Index: 2, Group: 2},
					{Name: "Option1", Index: 0, Group: 0},
				}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "HoistedClientCredentialsBasic",
			openAPIYAML: `
openapi: 3.1.0
info:
  version: 1.0.0
security:
  - clientCredentials: [read]
    basicAuth: []
paths:
  /test:
    get:
      security:
        - basicAuth: []
          clientCredentials: [write]
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
    clientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: /oauth2/authorize
          scopes:
            read: Read access
            write: Write access`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithOAuth2Config("clientCredentials", "client_credentials", "", []string{"write"},
					[]ast.OAuth2Scope{
						{Name: "read", Comments: ast.Comment{Description: "Read access"}},
						{Name: "write", Comments: ast.Comment{Description: "Write access"}},
					}).
				WithHoistedSecurityConfig(true, []ast.HoistedSecurityField{
					{Name: "basicAuth", Index: 1},
					{Name: "clientCredentials", Index: 0},
				}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
				},
			},
		},
		{
			name: "HoistedFlattenedBasicAuth",
			openAPIYAML: `
openapi: 3.1.0
info:
  version: 1.0.0
security:
  - basicAuth: []
paths:
  /test:
    get:
      security:
        - basicAuth: []
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
    clientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: /oauth2/authorize
          scopes:
            read: Read access
            write: Write access`,
			operationPath:   "/test",
			operationMethod: "get",
			mockFeatures: map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			},
			expectedSecurity: testutils.NewSecurity().
				WithHoistedSecurityConfig(true, []ast.HoistedSecurityField{
					{Name: "Username", Index: 0},
				}).
				WithOptionalityReason("").
				Build(),
			testCases: []operationSecurityTestCase{
				{
					securityFixesFeb2025: false,
				},
				{
					securityFixesFeb2025: true,
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
			t.Run(testName, func(t *testing.T) {
				runOperationSecurityTest(t, tt, tc)
			})
		}
	}
}

func TestHandleGlobalSecurityMultipleBuiltInOAuth2Flows(t *testing.T) {
	const openAPIYAML = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ClientCredentials: [client-read]
  - Password: [user-read]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ClientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/client-token
          scopes:
            client-read: Client access
    Password:
      type: oauth2
      flows:
        password:
          tokenUrl: https://example.com/password-token
          scopes:
            user-read: User access
`

	for _, securityFixesFeb2025 := range []bool{false, true} {
		t.Run(fmt.Sprintf("SecurityFeb2025=%t", securityFixesFeb2025), func(t *testing.T) {
			setup := setupSecurityTest(t, openAPIYAML, nil, map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
				features.FeatureOAuth2Password:          true,
			}, securityFixesFeb2025, true)

			globalSecurity, _, err := setup.generator.handleGlobalSecurity(setup.ctx, setup.docInfo, &ast.AST{})
			require.NoError(t, err)
			require.NotNil(t, globalSecurity)
			require.Len(t, globalSecurity.SecurityConfig.OAuth2Config, 2)

			clientCredentials := globalSecurity.SecurityConfig.OAuth2Config["ClientCredentials"]
			assert.Equal(t, ast.OAuth2FlowClientCredentials, clientCredentials.Flow)
			assert.Equal(t, []string{"client-read"}, clientCredentials.RequiredScopes)

			password := globalSecurity.SecurityConfig.OAuth2Config["Password"]
			assert.Equal(t, ast.OAuth2FlowPassword, password.Flow)
			assert.Equal(t, []string{"user-read"}, password.RequiredScopes)
		})
	}
}

func TestHandleGlobalSecurityAcceptsReorderedScopesForSameOAuth2Scheme(t *testing.T) {
	const openAPIYAML = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ClientCredentials: [read, write, read]
  - ClientCredentials: [write, read]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ClientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/client-token
          scopes:
            read: Read access
            write: Write access
`

	for _, securityFixesFeb2025 := range []bool{false, true} {
		t.Run(fmt.Sprintf("SecurityFeb2025=%t", securityFixesFeb2025), func(t *testing.T) {
			setup := setupSecurityTest(t, openAPIYAML, nil, map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			}, securityFixesFeb2025, true)

			globalSecurity, _, err := setup.generator.handleGlobalSecurity(setup.ctx, setup.docInfo, &ast.AST{})
			require.NoError(t, err)
			require.NotNil(t, globalSecurity)
			require.Len(t, globalSecurity.SecurityConfig.OAuth2Config, 1)
		})
	}
}

func TestHandleGlobalSecurityRejectsConflictingScopesForSameOAuth2Scheme(t *testing.T) {
	const openAPIYAML = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
security:
  - ClientCredentials: [read]
  - ClientCredentials: [write]
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    ClientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://example.com/client-token
          scopes:
            read: Read access
            write: Write access
`

	for _, securityFixesFeb2025 := range []bool{false, true} {
		t.Run(fmt.Sprintf("SecurityFeb2025=%t", securityFixesFeb2025), func(t *testing.T) {
			setup := setupSecurityTest(t, openAPIYAML, nil, map[features.Feature]bool{
				features.FeatureOAuth2ClientCredentials: true,
			}, securityFixesFeb2025, true)

			_, _, err := setup.generator.handleGlobalSecurity(setup.ctx, setup.docInfo, &ast.AST{})
			require.ErrorContains(t, err, "multiple security requirements for OAuth2 scheme \"ClientCredentials\" with different scopes are not currently supported")
		})
	}
}

// Common test setup structures and functions

type securityTestSetup struct {
	generator *Generator
	docInfo   *document.DocumentInfo
	//nolint:containedctx
	ctx context.Context
}

// setupSecurityTest creates common test infrastructure for both operation and global security tests
func setupSecurityTest(t *testing.T, openAPIYAML string, fixes *config.Fixes, mockFeatures map[features.Feature]bool, securityFixesFeb2025 bool, hoistGlobalSecurity bool) *securityTestSetup {
	t.Helper()
	// Create enterprise context for licensing
	ctx := newEnterpriseContext()

	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(openAPIYAML)))
	require.NoError(t, err)

	// Set up fixes
	if fixes == nil {
		fixes = &config.Fixes{
			SecurityFeb2025: securityFixesFeb2025,
		}
	} else {
		fixes.SecurityFeb2025 = securityFixesFeb2025
	}

	// Create test configuration
	baseConfig := &config.Configuration{
		Generation: config.Generation{
			MaintainOpenAPIOrder: true,
			Fixes:                fixes,
			Auth: &config.Auth{
				HoistGlobalSecurity:            hoistGlobalSecurity,
				OAuth2ClientCredentialsEnabled: mockFeatures != nil && mockFeatures[features.FeatureOAuth2ClientCredentials],
				OAuth2PasswordEnabled:          mockFeatures != nil && mockFeatures[features.FeatureOAuth2Password],
			},
		},
		Languages: map[string]config.LanguageConfig{
			"go": {
				Cfg: map[string]any{
					"imports": configuration.ImportConfig{
						Option: configuration.ImportOptionOpenAPI,
					},
				},
			},
		},
	}

	cfg := configuration.New(baseConfig, "go")
	target := types.Target{Target: "go"}

	// Set up mock features
	supportedFeatures := map[features.Feature]bool{
		features.FeatureGlobalSecurity:          true,
		features.FeatureOAuth2ClientCredentials: false,
		features.FeatureOAuth2Password:          false,
		features.FeatureEnvVarSecurityUsage:     false,
		features.FeatureCustomSecuritySchemes:   false,
		features.FeatureSDKHooks:                false,
	}
	enabledFeatures := map[features.Feature]bool{}

	for feature, enabled := range mockFeatures {
		supportedFeatures[feature] = enabled
		enabledFeatures[feature] = enabled
	}

	// Create subsystem
	exts := extensions.New(target)
	reg := register.New()
	reg.Config = cfg

	mockSubsystem := &subsystem.Subsystem{
		Extensions: exts,
		Features: testutils.NewMockFeatures(&testutils.MockFeaturesConfig{
			Target:            target,
			SupportedFeatures: supportedFeatures,
			EnabledFeatures:   enabledFeatures,
		}),
		Register: reg,
		Config:   cfg,
		Target:   target,
	}

	// Create schemas handler
	schemasHandler := &schemas.Schemas{
		Config:    cfg,
		Target:    target,
		Subsystem: mockSubsystem,
		Namer:     namer.New(mockSubsystem),
	}

	// Create generator
	generator := &Generator{
		subsystem: mockSubsystem,
		schemas:   schemasHandler,
		log:       logging.NewLogger(logging.LevelFromEnv()),
	}

	return &securityTestSetup{
		generator: generator,
		docInfo: &document.DocumentInfo{
			Doc:        doc,
			Schema:     []byte(openAPIYAML),
			SchemaPath: "test.yaml",
			IsRemote:   false,
		},
		ctx: ctx,
	}
}

// assertSecurityResult performs common security test assertions
func assertSecurityResult(t *testing.T, testName string, expected, actual *ast.Security, transform func(*ast.Security) *ast.Security) {
	t.Helper()

	// Get expected security (base + optional transform)
	expectedSecurity := expected
	if transform != nil {
		expectedSecurity = transform(expectedSecurity)
	}

	isEqual := testutils.DeepCompare(t, expectedSecurity, actual)
	if !isEqual {
		t.Errorf("Deep comparison failed for test case: %s", testName)
	}
	assert.EqualExportedValues(t, expectedSecurity, actual, "Security should match expected value")
}

func runOperationSecurityTest(t *testing.T, tt operationSecurityTest, tc operationSecurityTestCase) {
	t.Helper()

	setup := setupSecurityTest(t, tt.openAPIYAML, tt.fixes, tt.mockFeatures, tc.securityFixesFeb2025, false)

	// Get the operation
	pi, exists := setup.docInfo.Doc.GetPaths().Get(tt.operationPath)
	require.True(t, exists, "Path %s should exist", tt.operationPath)

	pathItem := pi.MustGetObject()

	operation := pathItem.GetOperation(openapi.HTTPMethod(tt.operationMethod))
	require.NotNil(t, operation, "Operation %s should exist", tt.operationMethod)

	// Create context stack for operation
	contextStack := ast.ContextStack{
		{
			Type:       ast.ContextTypeOperation,
			Identifier: "Test",
		},
	}

	// Build global security for hoisting detection
	globalSecurity, err := setup.generator.handleSecurity(
		setup.ctx,
		setup.docInfo,
		ast.ContextStack{},
		setup.docInfo.Doc.GetSecurity(),
		setup.docInfo.Doc.GetComponents().GetSecuritySchemes(),
		ast.SecOptReasonNotOptional,
		nil,
		nil,
		ast.ScopeShared,
		true,
	)
	require.NoError(t, err)

	globalSecurityMatcher := security.NewGlobalSecurityMatcher(globalSecurity)

	security, err := setup.generator.handleSecurity(
		setup.ctx,
		setup.docInfo,
		contextStack,
		operation.Security,
		setup.docInfo.Doc.GetComponents().GetSecuritySchemes(),
		ast.SecOptReasonNotOptional,
		globalSecurity,
		globalSecurityMatcher,
		ast.ScopeShared,
		false,
	)

	require.NoError(t, err)

	assertSecurityResult(t, tt.name, tt.expectedSecurity, security, tc.transform)
}
