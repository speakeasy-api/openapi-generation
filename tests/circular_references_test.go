package main

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
)

const openapiHeaderNoPaths = `openapi: 3.1.0
info:
  title: FailureCases
  version: 0.1.0
servers:
  - url: http://localhost:35123
    description: The default server.
`

const openapiHeader = openapiHeaderNoPaths + `paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`

func TestValidCircularLoops_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple non-required property loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Obj:
      x-speakeasy-include: true
      type: object
      properties:
        self:
          $ref: '#/components/schemas/Obj'
`,
			},
		},
		{
			name: "deep non-required property loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Obj:
      x-speakeasy-include: true
      type: object
      properties:
        other:
          $ref: '#/components/schemas/Obj2'
    Obj2:
      type: object
      properties:
        other:
          $ref: '#/components/schemas/Obj'
      required:
        - other
`,
			},
		},
		{
			name: "array with no minItems loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Obj:
      type: object
      properties:
        children:
          type: array
          items:
            $ref: '#/components/schemas/Obj'
      required:
        - children
`,
			},
		},
		{
			name: "array with no minItems loop and examples succeeds",
			args: args{
				schema: openapiHeaderNoPaths + `
paths:
  /testcircular:
    get:
      operationId: testCircular
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Obj'
      responses:
        '200':
          description: OK
components:
  schemas:
    Obj:
      type: object
      properties:
        name:
          type: string
        children:
          type: array
          items:
            $ref: '#/components/schemas/Obj'
      required:
        - children
      example:
        name: level1
        children:
          - name: level2
            children:
              - name: level3
                children:
                  - name: level4
                    children: []
`,
			},
		},
		{
			name: "deep array with no minItems loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Obj:
      type: object
      properties:
        other:
          $ref: '#/components/schemas/Obj2'
      required:
        - other
    Obj2:
      type: object
      properties:
        children:
          type: array
          items:
            $ref: '#/components/schemas/Obj'
      required:
        - children
`,
			},
		},
		{
			name: "loop through additionalProperties succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Obj:
      x-speakeasy-include: true
      type: object
      additionalProperties:
          $ref: '#/components/schemas/Obj'
`,
			},
		},
		{
			name: "Banking events allOf loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    eventHistory:
      x-speakeasy-include: true
      title: Event History (v1.2.0)
      allOf:
        - $ref: "#/components/schemas/abstractPagedBody"
        - $ref: "#/components/schemas/eventMessages"
    abstractPagedBody:
      title: Abstract Paged Body (v0.5.0)
      type: object
      allOf:
        - $ref: "#/components/schemas/abstractBody"
        - $ref: "#/components/schemas/abstractPagedBodyFields"
        - type: object
          properties:
            start:
              type: string
    abstractBody:
      title: Abstract Body (v0.2.0)
      type: object
      properties: {}
    abstractPagedBodyFields:
      title: Abstract Paged Body Fields (v1.1.0)
      type: object
      required:
        - limit
      properties:
        limit:
          type: integer
        nextPage_url:
          type: string
    eventMessages:
      title: Event Messages (v0.7.0)
      type: object
      required:
        - version
        - items
      allOf:
        - $ref: "#/components/schemas/abstractBody"
        - type: object
          properties:
            version:
              type: string
`,
			},
		},
		{
			name: "External accounting TrackingCategoryTree array allOf loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    TrackingCategoryTree:
      x-speakeasy-include: true
      title: Tracking category tree
      description: The full structure of a specific tracking category including any child or subcategories.
      type: object
      allOf:
        - type: object
          properties:
            id:
              type: string
              description: 'The identifier for the item, unique per tracking category'
              nullable: true
            name:
              type: string
              description: The name of the tracking category
              nullable: true
            parentId:
              type: string
              description: The identifier for this item's immediate parent
              nullable: true
            hasChildren:
              type: boolean
              description: Boolean value indicating whether this category has SubCategories
            subCategories:
              type: array
              nullable: true
              description: A collection of subcategories that are nested beneath this category.
              items:
                $ref: '#/components/schemas/TrackingCategoryTree'
        - $ref: '#/components/schemas/Bill'
    Bill:
      allOf:
        - type: object
          properties:
            modifiedDate:
              type: string
        - type: object
          properties:
            createdData:
              type: string
`,
			},
		},
		{
			name: "External accounting TrackingCategoryTree array allOf loop succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Object:
      type: object
      properties:
        child:
          $ref: '#/components/schemas/Child'
    OtherObject:
      type: object
    Child:
      type: object
      oneOf:
        - $ref: '#/components/schemas/Object'
        - $ref: '#/components/schemas/OtherObject'`,
			},
		},
		{
			name: "recursive oneOf as object property succeeds",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    Wrapper:
      x-speakeasy-include: true
      type: object
      properties:
        value:
          $ref: '#/components/schemas/RecursiveOneOf'
    RecursiveOneOf:
      oneOf:
        - type: string
        - type: array
          items:
            $ref: '#/components/schemas/RecursiveOneOf'
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := testGenerate(t, []byte(tt.args.schema))
			require.Empty(t, errs)
		})
	}
}

func TestInvalidCircularLoops_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "simple allOf loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    AllOfLoop:
      x-speakeasy-include: true
      allOf:
       - $ref: '#/components/schemas/AllOfLoop'
       - type: object
         properties:
           name:
             type: string
`,
			},
			wantErrs: []string{"validation error: [line 20:10] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/AllOfLoop -> openapi.yaml#/components/schemas/AllOfLoop"},
		},
		{
			name: "allOf that includes itself and another ref fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    AllOfLoop:
      x-speakeasy-include: true
      allOf:
       - $ref: '#/components/schemas/AllOfLoop'
       - $ref: '#/components/schemas/OtherObject'
    OtherObject:
      type: object
      properties:
        name:
          type: string
`,
			},
			wantErrs: []string{"validation error: [line 20:10] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/AllOfLoop -> openapi.yaml#/components/schemas/AllOfLoop"},
		},
		{
			name: "allOf through oneOf loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    AllOf:
      x-speakeasy-include: true
      allOf:
       - $ref: '#/components/schemas/OneOf'
       - type: object
         properties:
           name:
             type: string
    OneOf:
      x-speakeasy-include: true
      oneOf:
       - $ref: '#/components/schemas/AllOf'
`,
			},
			wantErrs: []string{"validation error: [line 26:7] circular-reference-invalid - non-terminating circular reference: all oneOf branches recurse with no base case"},
		},
		{
			name: "allOf through other allOf loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    AllOf1:
      x-speakeasy-include: true
      allOf:
       - $ref: '#/components/schemas/AllOf2'
    AllOf2:
      allOf:
       - $ref: '#/components/schemas/AllOf3'
    AllOf3:
      allOf:
       - $ref: '#/components/schemas/AllOf1'
`,
			},
			wantErrs: []string{
				"validation error: [line 20:10] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/AllOf2 -> openapi.yaml#/components/schemas/AllOf3 -> openapi.yaml#/components/schemas/AllOf1 -> openapi.yaml#/components/schemas/AllOf2",
			},
		},
		{
			name: "required property loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    ObjectWithRequiredProperty:
      x-speakeasy-include: true
      type: object
      properties:
        self:
          $ref: '#/components/schemas/ObjectWithRequiredProperty'
      required:
        - self
`,
			},
			wantErrs: []string{
				"validation error: [line 22:11] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/ObjectWithRequiredProperty -> openapi.yaml#/components/schemas/ObjectWithRequiredProperty",
			},
		},
		{
			name: "required property loop across objects fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    ObjectA:
      x-speakeasy-include: true
      type: object
      properties:
        b:
          $ref: '#/components/schemas/ObjectB'
      required:
        - b
    ObjectB:
      type: object
      properties:
        a:
          $ref: '#/components/schemas/ObjectA'
      required:
        - a
`,
			},
			wantErrs: []string{
				"validation error: [line 22:11] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/ObjectB -> openapi.yaml#/components/schemas/ObjectA -> openapi.yaml#/components/schemas/ObjectB",
			},
		},
		{
			name: "array with minItems loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    ObjectWithArray:
      x-speakeasy-include: true
      type: object
      properties:
        children:
          type: array
          minItems: 1
          items:
            $ref: '#/components/schemas/ObjectWithArray'
      required:
        - children
`,
			},
			wantErrs: []string{"validation error: [line 25:13] circular-reference-invalid - non-terminating circular reference detected: openapi.yaml#/components/schemas/ObjectWithArray -> openapi.yaml#/components/schemas/ObjectWithArray"},
		},
		{
			name: "single oneOf loop fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    ObjectWithOneOf:
      x-speakeasy-include: true
      type: object
      properties:
        child:
          oneOf:
            - $ref: '#/components/schemas/ObjectWithOneOf'
      required:
        - child
`,
			},
			wantErrs: []string{"validation error: [line 22:11] circular-reference-invalid - non-terminating circular reference: all oneOf branches recurse with no base case"},
		},
		{
			name: "recursive oneOf as direct schema reference fails",
			args: args{
				schema: openapiHeader + `
components:
  schemas:
    RecursiveOneOf:
      x-speakeasy-include: true
      oneOf:
        - type: string
        - type: array
          items:
            $ref: '#/components/schemas/RecursiveOneOf'
`,
			},
			wantErrs: []string{"validation error: [line 23:13] circular reference:  -> #/components/schemas/RecursiveOneOf -> #/components/schemas/RecursiveOneOf"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := testGenerate(t, []byte(tt.args.schema))
			receivedErrs := []string{}
			for _, err := range errs {
				errStr := err.Error()
				// Filter out validation hints, keep only validation errors
				if strings.HasPrefix(errStr, "validation error:") {
					// Normalize temp directory paths
					errStr = normalizeErrorPath(errStr)
					receivedErrs = append(receivedErrs, errStr)
				}
			}

			assert.Equal(t, tt.wantErrs, receivedErrs)
		})
	}
}

// normalizeErrorPath removes temp directory paths from error messages
func normalizeErrorPath(errMsg string) string {
	// Remove any absolute directory prefix, keeping just openapi.yaml#ref
	re := regexp.MustCompile(`[^\s]*?/openapi\.yaml`)
	return re.ReplaceAllString(errMsg, "openapi.yaml")
}

func testGenerate(t *testing.T, contents []byte) []error {
	t.Helper()

	outDir := t.TempDir()

	schemaPath := filepath.Join(outDir, "openapi.yaml")
	if err := os.WriteFile(schemaPath, contents, 0o644); err != nil {
		return []error{err}
	}

	return generate.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), generate.GenerateOptions{
		OutDir:      outDir,
		Lang:        "go",
		SchemaPath:  schemaPath,
		SkipCompile: true,
		Logger:      logger,
	})
}
