package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type securityTestCase struct {
	name            string
	scheme          string
	expectedSubType string // Expected normalized subtype (always lowercase)
	validateFields  func(t *testing.T, fields []*ast.FieldDef)
}

// TestSecurityCaseInsensitiveHTTPSchemes tests that HTTP authentication schemes
// are validated case-insensitively, accepting "Basic", "Bearer", "basic", "bearer", etc.
func TestSecurityCaseInsensitiveHTTPSchemes(t *testing.T) {
	validateBasicAuth := func(t *testing.T, fields []*ast.FieldDef) {
		t.Helper()
		fieldNames := make([]string, len(fields))
		for i, field := range fields {
			fieldNames[i] = field.Name
		}
		assert.Contains(t, fieldNames, "Username", "Should have Username field for basic auth")
		assert.Contains(t, fieldNames, "Password", "Should have Password field for basic auth")
	}

	validateBearerAuth := func(t *testing.T, fields []*ast.FieldDef) {
		t.Helper()
		require.Len(t, fields, 1, "Bearer auth should have exactly one field")
		assert.Equal(t, "TestAuth", fields[0].Name, "Bearer auth field should be named after the scheme")
	}

	tests := []securityTestCase{
		{
			name:            "BasicLowercase",
			scheme:          "basic",
			expectedSubType: "basic",
			validateFields:  validateBasicAuth,
		},
		{
			name:            "BasicCapitalized",
			scheme:          "Basic",
			expectedSubType: "basic",
			validateFields:  validateBasicAuth,
		},
		{
			name:            "BasicUppercase",
			scheme:          "BASIC",
			expectedSubType: "basic",
			validateFields:  validateBasicAuth,
		},
		{
			name:            "BasicMixedCase",
			scheme:          "bAsIc",
			expectedSubType: "basic",
			validateFields:  validateBasicAuth,
		},
		{
			name:            "BearerLowercase",
			scheme:          "bearer",
			expectedSubType: "bearer",
			validateFields:  validateBearerAuth,
		},
		{
			name:            "BearerCapitalized",
			scheme:          "Bearer",
			expectedSubType: "bearer",
			validateFields:  validateBearerAuth,
		},
		{
			name:            "BearerUppercase",
			scheme:          "BEARER",
			expectedSubType: "bearer",
			validateFields:  validateBearerAuth,
		},
		{
			name:            "BearerMixedCase",
			scheme:          "BeArEr",
			expectedSubType: "bearer",
			validateFields:  validateBearerAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create OpenAPI spec with the test scheme
			rawSpec := utils.Dedent(`{
				"openapi": "3.1.0",
				"info": { "title": "Test API", "version": "1.0.0" },
				"paths": {
					"/test": {
						"get": {
							"operationId": "getTest",
							"security": [
								{ "TestAuth": [] }
							],
							"responses": {
								"200": {
									"description": "OK",
									"content": {
										"application/json": {
											"schema": { "type": "object" }
										}
									}
								}
							}
						}
					}
				},
				"components": {
					"securitySchemes": {
						"TestAuth": {
							"type": "http",
							"scheme": "` + tt.scheme + `"
						}
					}
				}
			}`)

			// Use ResolveAST to generate the full AST
			_, astTree, err := TestResolveAST(TestResolveASTInput{OpenAPIContents: []byte(rawSpec)})
			require.NoError(t, err)

			// Since security is hoisted to global level, access it directly from MainSDK
			require.NotNil(t, astTree.MainSDK, "MainSDK should exist")
			require.NotNil(t, astTree.MainSDK.Security, "MainSDK should have security configured")

			securityClass := astTree.MainSDK.Security.Type
			require.NotNil(t, securityClass, "Security class should be generated")

			// Verify the security class has the expected structure
			assert.Equal(t, ast.DataTypeClass, securityClass.Type, "Security should be a class")
			require.NotEmpty(t, securityClass.Fields, "Security class should have fields")

			// Use the test case's validation function to check field structure
			tt.validateFields(t, securityClass.Fields)

			// Verify that the scheme was processed correctly by checking annotations
			foundValidAnnotation := false
			for _, field := range securityClass.Fields {
				if annotation := field.Annotations.Get(ast.AnnotationTypeSecurity); annotation != nil {
					if secAnnotation, ok := annotation.(*ast.SecurityAnnotation); ok {
						assert.Equal(t, "http", secAnnotation.SecType, "Security type should be http")
						assert.Equal(t, tt.expectedSubType, secAnnotation.SubType, "SubType should be normalized to lowercase")
						foundValidAnnotation = true
					}
				}
			}
			assert.True(t, foundValidAnnotation, "Should find at least one valid security annotation")

			// Verify that the security field is properly configured
			assert.Equal(t, "Security", astTree.MainSDK.Security.Name, "Security field should be named 'Security'")
		})
	}
}
