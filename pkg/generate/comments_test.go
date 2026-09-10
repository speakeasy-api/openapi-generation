package generate

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// CommentsTestCase defines the structure for comments test cases
type CommentsTestCase struct {
	Name           string
	OpenAPIYAML    string
	PathName       string
	OperationName  string
	Tag            *openapi.Tag
	Extensions     *sequencedmap.Map[string, *yaml.Node]
	ExpectedResult *ast.Comment
	ExpectError    bool
	ExpectedError  string
}

// runCommentsTest is a flexible helper method to reduce duplication in comments tests
func runCommentsTest(t *testing.T, testCase CommentsTestCase, testFunc func(*testutils.CommonTestSetup, *Generator) (*ast.Comment, error)) {
	t.Helper()

	t.Run(testCase.Name, func(t *testing.T) {
		// Set up test environment
		setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
			OpenAPIYAML: testCase.OpenAPIYAML,
		})
		require.NoError(t, err)

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

		// Call the test function
		result, err := testFunc(setup, generator)

		// Verify the result
		if testCase.ExpectError {
			require.Error(t, err)
			if testCase.ExpectedError != "" {
				assert.Contains(t, err.Error(), testCase.ExpectedError)
			}
			return
		}

		require.NoError(t, err)

		// Use deep comparison for complex structures
		isEqual := testutils.DeepCompare(t, testCase.ExpectedResult, result)
		if !isEqual {
			t.Errorf("Deep comparison failed for test case: %s", testCase.Name)
		}
		assert.Equal(t, testCase.ExpectedResult, result)
	})
}

// TestHandleDocumentComments tests the handleDocumentComments method with various scenarios
func TestHandleDocumentComments(t *testing.T) {
	tests := []CommentsTestCase{
		{
			Name: "DocumentWithTitleAndDescription",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  description: This is a test API for demonstration purposes
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				Description:      "Test API: This is a test API for demonstration purposes",
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithTitleAndSummary",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  summary: A simple test API
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				Summary:          "Test API: A simple test API",
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithSummaryAndDescription",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  summary: A simple test API
  description: This is a test API for demonstration purposes
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				Summary:          "Test API: A simple test API",
				Description:      "This is a test API for demonstration purposes",
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithOnlyTitle",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithOnlySummary",
			OpenAPIYAML: `openapi: 3.0.0
info:
  summary: A simple test API
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				Summary:          "A simple test API",
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithOnlyDescription",
			OpenAPIYAML: `openapi: 3.0.0
info:
  description: This is a test API for demonstration purposes
  version: 1.0.0
paths: {}`,
			ExpectedResult: &ast.Comment{
				Description:      "This is a test API for demonstration purposes",
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithExternalDocs",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  description: This is a test API
  version: 1.0.0
externalDocs:
  url: https://example.com/docs
  description: External documentation
paths: {}`,
			ExpectedResult: &ast.Comment{
				Description: "Test API: This is a test API",
				ExternalDocs: &ast.ExternalDocs{
					URL:         "https://example.com/docs",
					Description: "External documentation",
				},
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithExternalDocsURLOnly",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  description: This is a test API
  version: 1.0.0
externalDocs:
  url: https://example.com/docs
paths: {}`,
			ExpectedResult: &ast.Comment{
				Description: "Test API: This is a test API",
				ExternalDocs: &ast.ExternalDocs{
					URL: "https://example.com/docs",
				},
				ExtendedComments: nil,
			},
		},
		{
			Name: "DocumentWithEmptyInfo",
			OpenAPIYAML: `openapi: 3.0.0
info:
  version: 1.0.0
paths: {}`,
			ExpectedResult: nil,
		},
		{
			Name: "DocumentWithNoInfo",
			OpenAPIYAML: `openapi: 3.0.0
paths: {}`,
			ExpectedResult: nil,
		},
	}

	for _, tt := range tests {
		runCommentsTest(t, tt, func(setup *testutils.CommonTestSetup, generator *Generator) (*ast.Comment, error) {
			return generator.handleDocumentComments(context.Background(), setup.DocInfo.Doc, nil)
		})
	}
}

// TestHandleOperationComments tests the handleOperationComments method with various scenarios
func TestHandleOperationComments(t *testing.T) {
	tests := []CommentsTestCase{
		{
			Name: "OperationWithSummaryAndDescription",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      description: Retrieve a list of users from the system
      responses:
        '200':
          description: Success`,
			PathName:      "/users",
			OperationName: "get",
			ExpectedResult: &ast.Comment{
				Summary:          "Get users",
				Description:      "Retrieve a list of users from the system",
				ExtendedComments: nil,
			},
		},
		{
			Name: "OperationWithOnlySummary",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success`,
			PathName:      "/users",
			OperationName: "get",
			ExpectedResult: &ast.Comment{
				Summary:          "Get users",
				ExtendedComments: nil,
			},
		},
		{
			Name: "OperationWithOnlyDescription",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      description: Retrieve a list of users from the system
      responses:
        '200':
          description: Success`,
			PathName:      "/users",
			OperationName: "get",
			ExpectedResult: &ast.Comment{
				Summary:          "Retrieve a list of users from the system",
				ExtendedComments: nil,
			},
		},
		{
			Name: "OperationWithDeprecated",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      description: Retrieve a list of users from the system
      deprecated: true
      responses:
        '200':
          description: Success`,
			PathName:      "/users",
			OperationName: "get",
			ExpectedResult: &ast.Comment{
				Summary:            "Get users",
				Description:        "Retrieve a list of users from the system",
				Deprecated:         true,
				DeprecationMessage: "This will be removed in a future release, please migrate away from it as soon as possible",
				ExtendedComments:   nil,
			},
		},
		{
			Name: "OperationWithExternalDocs",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      description: Retrieve a list of users from the system
      externalDocs:
        url: https://example.com/users-docs
        description: User API documentation
      responses:
        '200':
          description: Success`,
			PathName:      "/users",
			OperationName: "get",
			ExpectedResult: &ast.Comment{
				Summary:     "Get users",
				Description: "Retrieve a list of users from the system",
				ExternalDocs: &ast.ExternalDocs{
					URL:         "https://example.com/users-docs",
					Description: "User API documentation",
				},
				ExtendedComments: nil,
			},
		},
		{
			Name: "OperationWithNoCommentableFields",
			OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: Success`,
			PathName:       "/users",
			OperationName:  "get",
			ExpectedResult: nil,
		},
	}

	for _, tt := range tests {
		runCommentsTest(t, tt, func(setup *testutils.CommonTestSetup, generator *Generator) (*ast.Comment, error) {
			// Get the operation
			pi, exists := setup.DocInfo.Doc.GetPaths().Get(tt.PathName)
			require.True(t, exists, "Path %s not found", tt.PathName)
			pathItem, err := resolution.Resolve(context.Background(), pi, setup.DocInfo)
			require.NoError(t, err)

			var operation *openapi.Operation
			switch tt.OperationName {
			case "get":
				operation = pathItem.Get()
			case "post":
				operation = pathItem.Post()
			case "put":
				operation = pathItem.Put()
			case "delete":
				operation = pathItem.Delete()
			case "patch":
				operation = pathItem.Patch()
			case "head":
				operation = pathItem.Head()
			case "options":
				operation = pathItem.Options()
			case "trace":
				operation = pathItem.Trace()
			}

			require.NotNil(t, operation, "Operation %s not found", tt.OperationName)

			return generator.handleOperationComments(context.Background(), operation)
		})
	}
}

// TestHandleTagComments tests the handleTagComments method with various scenarios
func TestHandleTagComments(t *testing.T) {
	tests := []struct {
		name           string
		tag            *openapi.Tag
		expectedResult *ast.Comment
	}{
		{
			name: "TagWithDescription",
			tag: &openapi.Tag{
				Name:        "users",
				Description: pointer.From("Operations related to user management"),
			},
			expectedResult: &ast.Comment{
				Description:      "Operations related to user management",
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithDescriptionAndExternalDocs",
			tag: &openapi.Tag{
				Name:        "users",
				Description: pointer.From("Operations related to user management"),
				ExternalDocs: &oas3.ExternalDocumentation{
					URL:         "https://example.com/users-docs",
					Description: pointer.From("User management documentation"),
				},
			},
			expectedResult: &ast.Comment{
				Description: "Operations related to user management",
				ExternalDocs: &ast.ExternalDocs{
					URL:         "https://example.com/users-docs",
					Description: "User management documentation",
				},
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithOnlyExternalDocs",
			tag: &openapi.Tag{
				Name: "users",
				ExternalDocs: &oas3.ExternalDocumentation{
					URL:         "https://example.com/users-docs",
					Description: pointer.From("User management documentation"),
				},
			},
			expectedResult: &ast.Comment{
				ExternalDocs: &ast.ExternalDocs{
					URL:         "https://example.com/users-docs",
					Description: "User management documentation",
				},
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithExternalDocsURLOnly",
			tag: &openapi.Tag{
				Name:        "users",
				Description: pointer.From("Operations related to user management"),
				ExternalDocs: &oas3.ExternalDocumentation{
					URL: "https://example.com/users-docs",
				},
			},
			expectedResult: &ast.Comment{
				Description: "Operations related to user management",
				ExternalDocs: &ast.ExternalDocs{
					URL: "https://example.com/users-docs",
				},
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithNoCommentableFields",
			tag: &openapi.Tag{
				Name: "users",
			},
			expectedResult: nil,
		},
		{
			name:           "NilTag",
			tag:            nil,
			expectedResult: nil,
		},
		{
			name: "TagWithEmptyExternalDocsURL",
			tag: &openapi.Tag{
				Name:        "users",
				Description: pointer.From("Operations related to user management"),
				ExternalDocs: &oas3.ExternalDocumentation{
					URL:         "",
					Description: pointer.From("User management documentation"),
				},
			},
			expectedResult: &ast.Comment{
				Description:      "Operations related to user management",
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithSummary",
			tag: &openapi.Tag{
				Name:    "users",
				Summary: pointer.From("User Operations"),
			},
			expectedResult: &ast.Comment{
				Summary:          "User Operations",
				ExtendedComments: nil,
			},
		},
		{
			name: "TagWithSummaryAndDescription",
			tag: &openapi.Tag{
				Name:        "users",
				Summary:     pointer.From("User Operations"),
				Description: pointer.From("Operations related to user management"),
			},
			expectedResult: &ast.Comment{
				Summary:          "User Operations",
				Description:      "Operations related to user management",
				ExtendedComments: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up minimal test environment
			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			})
			require.NoError(t, err)

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

			// Call handleTagComments
			result, err := generator.handleTagComments(context.Background(), tt.tag)

			// Verify the result
			require.NoError(t, err)

			// Use deep comparison for complex structures
			isEqual := testutils.DeepCompare(t, tt.expectedResult, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestHandleDocsExtension tests the handleDocsExtension method with various scenarios
func TestHandleDocsExtension(t *testing.T) {
	tests := []struct {
		name           string
		extensions     *extensions.Extensions
		expectedResult map[string]*ast.ExtendedComment
	}{
		{
			name:           "NoExtensions",
			extensions:     extensions.New(),
			expectedResult: nil,
		},
		{
			name:           "NilExtensions",
			extensions:     nil,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up minimal test environment
			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success`,
			})
			require.NoError(t, err)

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

			// Call handleDocsExtension
			result, err := generator.handleDocsExtension(context.Background(), tt.extensions)

			// Verify the result
			require.NoError(t, err)

			// Use deep comparison for complex structures
			isEqual := testutils.DeepCompare(t, tt.expectedResult, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
