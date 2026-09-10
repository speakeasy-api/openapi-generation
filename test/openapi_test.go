package test

import (
	"bytes"
	"testing"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOpenAPIDocument(t *testing.T) {
	// Define a simple OpenAPI 3.0 document
	spec := `openapi: 3.0.0
info:
  title: Simple Test API
  version: 1.0.0
paths:
  /hello:
    get:
      summary: Say hello
      operationId: sayHello
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        gender:
          type: string
          enum:
            - male
            - female
            - other
`

	// Load the document
	doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(spec))
	require.NoError(t, err, "Failed to load OpenAPI document")

	foundSchemas := []string{}

	for item := range openapi.Walk(t.Context(), doc) {
		_ = item.Match(openapi.Matcher{
			Schema: func(j *oas3.JSONSchema[oas3.Referenceable]) error {
				foundSchemas = append(foundSchemas, item.Location.ToJSONPointer().String())
				return nil
			},
		})
	}

	assert.Len(t, foundSchemas, 4)
	assert.Contains(t, foundSchemas, "/components/schemas/User")
	assert.Contains(t, foundSchemas, "/components/schemas/User/properties/id")
	assert.Contains(t, foundSchemas, "/components/schemas/User/properties/gender")
}
