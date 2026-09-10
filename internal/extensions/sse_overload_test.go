package extensions

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckSchemaForStreamField covers the schema walker used to validate that
// the SSE-overload-selector field exists and supports a boolean type.
func TestCheckSchemaForStreamField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schemaYAML string
		fieldName  string
		expected   bool
	}{
		{
			name: "DirectProperty",
			schemaYAML: `
type: object
properties:
  stream:
    type: boolean`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "DirectPropertyAbsent",
			schemaYAML: `
type: object
properties:
  other:
    type: boolean`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "AllOfFieldInOneMember",
			schemaYAML: `
allOf:
  - type: object
    properties:
      stream:
        type: boolean
  - type: object
    properties:
      other:
        type: string`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "OneOfAllMembersHaveField",
			schemaYAML: `
oneOf:
  - type: object
    properties:
      stream:
        type: boolean
      a:
        type: string
  - type: object
    properties:
      stream:
        type: boolean
      b:
        type: string`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "OneOfOneBranchMissingField",
			schemaYAML: `
oneOf:
  - type: object
    properties:
      stream:
        type: boolean
  - type: object
    properties:
      other:
        type: string`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "AnyOfAllMembersHaveField",
			schemaYAML: `
anyOf:
  - type: object
    properties:
      stream:
        type: boolean
  - type: object
    properties:
      stream:
        type: boolean`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "AnyOfOneBranchMissingField",
			schemaYAML: `
anyOf:
  - type: object
    properties:
      stream:
        type: boolean
  - type: object
    properties:
      other:
        type: string`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "OneOfWithAllOfMember",
			schemaYAML: `
oneOf:
  - allOf:
      - type: object
        properties:
          stream:
            type: boolean
      - type: object
        properties:
          a:
            type: string
  - type: object
    properties:
      stream:
        type: boolean
      b:
        type: string`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "EmptyOneOf",
			schemaYAML: `
oneOf: []`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "NullableViaAnyOfPropertyType",
			schemaYAML: `
type: object
properties:
  stream:
    anyOf:
      - type: boolean
      - type: "null"`,
			fieldName: "stream",
			expected:  true,
		},
		{
			name: "FieldWithStringType",
			schemaYAML: `
type: object
properties:
  stream:
    type: string`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "AllOfPropertyMixedTypesUnsatisfiable",
			schemaYAML: `
type: object
properties:
  stream:
    allOf:
      - type: boolean
      - type: string`,
			fieldName: "stream",
			expected:  false,
		},
		{
			name: "AllOfPropertyBoolean",
			schemaYAML: `
type: object
properties:
  stream:
    allOf:
      - type: boolean
      - type: boolean`,
			fieldName: "stream",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			op, docInfo := loadOperationWithRequestSchema(t, tt.schemaYAML)
			bodyContent := op.GetRequestBody().GetObject().GetContent().GetOrZero("application/json")
			require.NotNil(t, bodyContent)
			schema := bodyContent.GetSchema()
			require.NotNil(t, schema)

			got, err := checkSchemaForStreamField(context.Background(), schema, docInfo, tt.fieldName)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestOperationCanHaveSSEOverload_SurfacesResolverError verifies that broken
// $ref errors from the body probe are surfaced to the caller instead of being
// masked by the generic "missing stream" validation when the query fallback
// also fails to find an eligible parameter.
func TestOperationCanHaveSSEOverload_SurfacesResolverError(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: testOp
      x-speakeasy-sse-overload: true
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/DoesNotExist"
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
            text/event-stream:
              schema:
                type: object`

	op, docInfo := loadOperationFromSpec(t, spec)
	e := &Extensions{}
	_, err := e.OperationCanHaveSSEOverload(context.Background(), op, docInfo)
	require.EqualError(t, err,
		"validation error: [line 15:15] error resolving schema: not found -- struct is nil at /components/schemas")
}

// TestCheckSchemaForStreamField_ResolveErrors verifies that resolution failures in
// $ref'd subschemas surface as errors instead of being swallowed as "field absent".
func TestCheckSchemaForStreamField_ResolveErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		schemaYAML     string
		componentsYAML string
		expectErr      string
		expectFound    bool
	}{
		{
			name:       "ValidRefResolves",
			schemaYAML: `$ref: "#/components/schemas/Body"`,
			componentsYAML: `Body:
      type: object
      properties:
        stream:
          type: boolean`,
			expectFound: true,
		},
		{
			name:       "BrokenRefSurfacesError",
			schemaYAML: `$ref: "#/components/schemas/DoesNotExist"`,
			expectErr:  "validation error: [line 14:15] error resolving schema: not found -- struct is nil at /components/schemas",
		},
		{
			name: "OneOfWithRefMembers",
			schemaYAML: `oneOf:
                - $ref: "#/components/schemas/MemberA"
                - $ref: "#/components/schemas/MemberB"`,
			componentsYAML: `MemberA:
      type: object
      properties:
        stream:
          type: boolean
    MemberB:
      type: object
      properties:
        stream:
          type: boolean`,
			expectFound: true,
		},
		{
			name: "OneOfWithBrokenRefMember",
			schemaYAML: `oneOf:
                - $ref: "#/components/schemas/MemberA"
                - $ref: "#/components/schemas/DoesNotExist"`,
			componentsYAML: `MemberA:
      type: object
      properties:
        stream:
          type: boolean`,
			expectErr: "validation error: [line 16:19] error resolving schema: not found -- key DoesNotExist not found in sequencedmap.Map",
		},
		{
			name: "PropertyWithRefToBoolean",
			schemaYAML: `type: object
              properties:
                stream:
                  $ref: "#/components/schemas/Bool"`,
			componentsYAML: `Bool:
      type: boolean`,
			expectFound: true,
		},
		{
			name: "PropertyWithBrokenRef",
			schemaYAML: `type: object
              properties:
                stream:
                  $ref: "#/components/schemas/DoesNotExist"`,
			expectErr: "validation error: [line 17:19] error resolving schema: not found -- struct is nil at /components/schemas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fullYAML := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: testOp
      requestBody:
        required: true
        content:
          application/json:
            schema:
              ` + tt.schemaYAML + `
      responses:
        '200':
          description: ok`
			if tt.componentsYAML != "" {
				fullYAML += "\ncomponents:\n  schemas:\n    " + tt.componentsYAML
			}

			op, docInfo := loadOperationFromSpec(t, fullYAML)
			bodyContent := op.GetRequestBody().GetObject().GetContent().GetOrZero("application/json")
			require.NotNil(t, bodyContent)
			got, err := checkSchemaForStreamField(context.Background(), bodyContent.GetSchema(), docInfo, "stream")
			if tt.expectErr != "" {
				require.EqualError(t, err, tt.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectFound, got)
		})
	}
}

func loadOperationFromSpec(t *testing.T, fullYAML string) (*openapi.Operation, *document.DocumentInfo) {
	t.Helper()
	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(fullYAML)))
	require.NoError(t, err)

	pi, ok := doc.Paths.Get("/test")
	require.True(t, ok)
	pathItem := pi.MustGetObject()
	require.NotNil(t, pathItem)

	operation := pathItem.Post()
	require.NotNil(t, operation, "operation not parsed from spec; check YAML:\n%s", fullYAML)

	docInfo := &document.DocumentInfo{
		Doc:        doc,
		Schema:     []byte(fullYAML),
		SchemaPath: "test.yaml",
		IsRemote:   false,
	}
	return operation, docInfo
}

// loadOperationWithRequestSchema wraps a schema fragment as the application/json
// request body of a synthetic operation so test rows only specify the schema.
func loadOperationWithRequestSchema(t *testing.T, schemaYAML string) (*openapi.Operation, *document.DocumentInfo) {
	t.Helper()

	indented := indentSSETestYAML(schemaYAML, "              ")
	fullYAML := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: testOp
      requestBody:
        required: true
        content:
          application/json:
            schema:
` + indented + `
      responses:
        '200':
          description: ok`
	return loadOperationFromSpec(t, fullYAML)
}

func indentSSETestYAML(yaml string, indent string) string {
	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}
