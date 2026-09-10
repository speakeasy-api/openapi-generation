package extensions

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleOperationPaginationExtension tests the HandleOperationPaginationExtension method with various pagination scenarios
func TestHandleOperationPaginationExtension(t *testing.T) {
	tests := []struct {
		name          string
		operationYAML string
		expected      *Pagination
		expectError   bool
		errorContains string
	}{
		{
			name: "NoPaginationExtension",
			operationYAML: `
operationId: testOperation
responses:
  '200':
    description: Success`,
			expected:    nil,
			expectError: false,
		},
		{
			name: "EmptyPaginationExtension",
			operationYAML: `
operationId: testOperation
x-speakeasy-pagination: {}
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type:    "",
				Inputs:  nil,
				Outputs: PaginationOutputs{},
			},
			expectError: false,
		},
		{
			name: "ValidOffsetLimitPaginationWithPage",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: page
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: offsetLimit
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: page
      in: parameters
      type: page
  outputs:
    results: $.data
    numPages: $.totalPages
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeOffsetLimit,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: true,
					},
					{
						Name:     "page",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypePage,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
					NumPages:          "$.totalPages",
					NumPagesDot:       "totalPages",
				},
			},
			expectError: false,
		},
		{
			name: "ValidOffsetLimitPaginationWithOffset",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: offset
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: offsetLimit
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: offset
      in: parameters
      type: offset
  outputs:
    results: $.data
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeOffsetLimit,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: true,
					},
					{
						Name:     "offset",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeOffset,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
				},
			},
			expectError: false,
		},
		{
			name: "ValidCursorPagination",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: cursor
    in: query
    schema:
      type: string
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: cursor
      in: parameters
      type: cursor
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeCursor,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: true,
					},
					{
						Name:     "cursor",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeCursor,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
					NextCursor:        "$.nextCursor",
					NextCursorDot:     "nextCursor",
				},
			},
			expectError: false,
		},
		{
			name: "ValidURLPagination",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: url
  inputs:
    - name: limit
      in: parameters
      type: limit
  outputs:
    results: $.data
    nextUrl: $.nextUrl
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeURL,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
					NextURL:           "$.nextUrl",
					NextURLDot:        "nextUrl",
				},
			},
			expectError: false,
		},
		{
			name: "ValidRequestBodyPagination",
			operationYAML: `
operationId: testOperation
requestBody:
  content:
    application/json:
      schema:
        type: object
        properties:
          limit:
            type: integer
          cursor:
            type: string
        required:
          - limit
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: requestBody
      type: limit
    - name: cursor
      in: requestBody
      type: cursor
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeCursor,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeRequestBody,
						Type:     PaginationInputTypeLimit,
						Optional: false, // Required in request body
					},
					{
						Name:     "cursor",
						In:       PaginationInputInTypeRequestBody,
						Type:     PaginationInputTypeCursor,
						Optional: true, // Not required in request body
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
					NextCursor:        "$.nextCursor",
					NextCursorDot:     "nextCursor",
				},
			},
			expectError: false,
		},
		{
			name: "RequiredParameterPagination",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    required: true
    schema:
      type: integer
  - name: cursor
    in: query
    schema:
      type: string
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: cursor
      in: parameters
      type: cursor
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeCursor,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: false, // Required parameter
					},
					{
						Name:     "cursor",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeCursor,
						Optional: true, // Optional parameter
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: true,
					Results:           "$.data",
					ResultsDot:        "data",
					NextCursor:        "$.nextCursor",
					NextCursorDot:     "nextCursor",
				},
			},
			expectError: false,
		},
		{
			name: "ComplexJSONPathPagination",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: cursor
    in: query
    schema:
      type: string
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: cursor
      in: parameters
      type: cursor
  outputs:
    results: $.response.data[*]
    nextCursor: $.response.pagination.nextCursor
responses:
  '200':
    description: Success`,
			expected: &Pagination{
				Type: PaginationTypeCursor,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: true,
					},
					{
						Name:     "cursor",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeCursor,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					CanUseDotNotation: false, // Complex JSONPath cannot use dot notation
					Results:           "$.response.data[*]",
					NextCursor:        "$.response.pagination.nextCursor",
				},
			},
			expectError: false,
		},
		{
			name: "InvalidPaginationType",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: invalid
  inputs:
    - name: limit
      in: parameters
      type: limit
  outputs:
    results: $.data
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "invalid x-speakeasy-pagination.type 'invalid'",
		},
		{
			name: "MissingParameterError",
			operationYAML: `
operationId: testOperation
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: nonexistent
      in: parameters
      type: limit
    - name: cursor
      in: parameters
      type: cursor
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "parameter nonexistent not found for operation testOperation",
		},
		{
			name: "MissingRequestBodyFieldError",
			operationYAML: `
operationId: testOperation
requestBody:
  content:
    application/json:
      schema:
        type: object
        properties:
          limit:
            type: integer
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: requestBody
      type: limit
    - name: nonexistent
      in: requestBody
      type: cursor
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "request body field nonexistent not found for operation testOperation",
		},
		{
			name: "CursorTypeMissingNextCursor",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: cursor
    in: query
    schema:
      type: string
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: cursor
      in: parameters
      type: cursor
  outputs:
    results: $.data
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type cursor requires output.nextCursor",
		},
		{
			name: "URLTypeMissingNextURL",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: url
  inputs:
    - name: limit
      in: parameters
      type: limit
  outputs:
    results: $.data
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type url requires output.nextUrl",
		},
		{
			name: "OffsetLimitMissingPageAndOffset",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: offsetLimit
  inputs:
    - name: limit
      in: parameters
      type: limit
  outputs:
    results: $.data
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type offsetLimit requires either input.page input.offset",
		},
		{
			name: "OffsetLimitPageMissingRequiredOutputs",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: page
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: offsetLimit
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: page
      in: parameters
      type: page
  outputs: {}
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type cursor requires output.numPages OR output.Results",
		},
		{
			name: "OffsetLimitOffsetMissingResults",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
  - name: offset
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: offsetLimit
  inputs:
    - name: limit
      in: parameters
      type: limit
    - name: offset
      in: parameters
      type: offset
  outputs: {}
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type cursor requires output.Results",
		},
		{
			name: "InvalidInputInType",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: invalid
      type: limit
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "invalid field 'in' for x-speakeasy-pagination input limit 'invalid'",
		},
		{
			name: "MissingCursorInput",
			operationYAML: `
operationId: testOperation
parameters:
  - name: limit
    in: query
    schema:
      type: integer
x-speakeasy-pagination:
  type: cursor
  inputs:
    - name: limit
      in: parameters
      type: limit
  outputs:
    results: $.data
    nextCursor: $.nextCursor
responses:
  '200':
    description: Success`,
			expected:      nil,
			expectError:   true,
			errorContains: "x-speakeasy-pagination type  requires cursor input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a full OpenAPI document with the operation
			fullYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
` + indentYAML(tt.operationYAML, "      ")

			ctx := context.Background()

			doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(fullYAML)))
			require.NoError(t, err)

			// Get the operation
			pi, exists := doc.Paths.Get("/test")
			require.True(t, exists)
			pathItem := pi.MustGetObject()
			require.NotNil(t, pathItem)

			operation := pathItem.Get()

			// Create Extensions instance
			extensions := New(types.Target{Target: "go"})

			// Call the method under test
			result, err := extensions.HandleOperationPaginationExtension(ctx, operation, &document.DocumentInfo{
				Doc:        doc,
				Schema:     []byte(fullYAML),
				SchemaPath: "test.yaml",
				IsRemote:   false,
			})

			// Check error expectations
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Helper function to indent YAML
func indentYAML(yaml string, indent string) string {
	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

func TestPagination_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pagination *Pagination
		testFunc   func(t *testing.T, original, cloned *Pagination)
	}{
		{
			name:       "nil pagination",
			pagination: nil,
			testFunc: func(t *testing.T, original, cloned *Pagination) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:       "empty pagination",
			pagination: &Pagination{},
			testFunc: func(t *testing.T, original, cloned *Pagination) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, PaginationType(""), cloned.Type)
				assert.Empty(t, cloned.Inputs)
			},
		},
		{
			name: "pagination with inputs",
			pagination: &Pagination{
				Type: PaginationTypeOffsetLimit,
				Inputs: []PaginationInputs{
					{
						Name:     "limit",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeLimit,
						Optional: false,
					},
					{
						Name:     "offset",
						In:       PaginationInputInTypeParameters,
						Type:     PaginationInputTypeOffset,
						Optional: true,
					},
				},
				Outputs: PaginationOutputs{
					Results: "$.data",
				},
			},
			testFunc: func(t *testing.T, original, cloned *Pagination) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Type, cloned.Type)
				assert.Equal(t, original.Outputs, cloned.Outputs)
				assert.Equal(t, original.Inputs, cloned.Inputs)
				for i := range original.Inputs {
					assert.NotSame(t, &original.Inputs[i], &cloned.Inputs[i])
				}

				// Verify deep copy by modifying original slice
				original.Inputs = append(original.Inputs, PaginationInputs{Name: "new"})
				assert.Len(t, cloned.Inputs, 2)
			},
		},
		{
			name: "pagination with nil inputs",
			pagination: &Pagination{
				Type:    PaginationTypeCursor,
				Inputs:  nil,
				Outputs: PaginationOutputs{NextCursor: "$.next"},
			},
			testFunc: func(t *testing.T, original, cloned *Pagination) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Type, cloned.Type)
				assert.Equal(t, original.Outputs, cloned.Outputs)
				assert.Empty(t, cloned.Inputs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.pagination.Clone()
			tt.testFunc(t, tt.pagination, cloned)
		})
	}
}
