package validation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const paginationNullableRequestBodySpecTemplate = `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    put:
      operationId: paginateNullableBody
      requestBody:
        required: false
        content:
          application/json:
            schema:
              %s
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      type: string
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - name: page
            in: requestBody
            type: page
        outputs:
          results: $.results
components:
  schemas:
    NullableRequiredBody:
      type: object
      nullable: true
      required:
        - page
      properties:
        page:
          type: integer
    NullableOptionalBody:
      type: object
      nullable: true
      properties:
        page:
          type: integer
    RequiredBody:
      type: object
      required:
        - page
      properties:
        page:
          type: integer
`

const paginationNullableRequestBodyParamsOnlySpec = `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    put:
      operationId: paginateNullableBodyParamsOnly
      parameters:
        - name: page
          in: query
          schema:
            type: integer
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
              nullable: true
              required:
                - name
              properties:
                name:
                  type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      type: string
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - name: page
            in: parameters
            type: page
        outputs:
          results: $.results
`

func TestValidation_PaginationNullableRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		wantErrs []string
	}{
		{
			name: "inline nullable body with required properties - rejected",
			schema: `type: object
              nullable: true
              required:
                - page
              properties:
                page:
                  type: integer`,
			wantErrs: []string{
				"validation error: [line 16:15] generator-validate-pagination-nullable-request-body - paginated operation has a nullable request body with required properties ([page]) - pages after the first need a body object, which cannot be constructed when the caller passes null",
			},
		},
		{
			name: "type array nullable body with required properties - rejected",
			schema: `type: ["object", "null"]
              required:
                - page
              properties:
                page:
                  type: integer`,
			wantErrs: []string{
				"validation error: [line 16:15] generator-validate-pagination-nullable-request-body - paginated operation has a nullable request body with required properties ([page]) - pages after the first need a body object, which cannot be constructed when the caller passes null",
			},
		},
		{
			name: "reference with sibling type-array nullability - rejected",
			schema: `type: ["object", "null"]
              $ref: '#/components/schemas/RequiredBody'`,
			wantErrs: []string{
				"validation error: [line 55:7] generator-validate-pagination-nullable-request-body - paginated operation has a nullable request body with required properties ([page]) - pages after the first need a body object, which cannot be constructed when the caller passes null",
			},
		},
		{
			name:   "referenced nullable body with required properties - rejected",
			schema: `$ref: '#/components/schemas/NullableRequiredBody'`,
			wantErrs: []string{
				"validation error: [line 40:7] generator-validate-pagination-nullable-request-body - paginated operation has a nullable request body with required properties ([page]) - pages after the first need a body object, which cannot be constructed when the caller passes null",
			},
		},
		{
			name: "inline nullable body with only optional properties - allowed",
			schema: `type: object
              nullable: true
              properties:
                page:
                  type: integer`,
			wantErrs: []string{},
		},
		{
			name:     "referenced nullable body with only optional properties - allowed",
			schema:   `$ref: '#/components/schemas/NullableOptionalBody'`,
			wantErrs: []string{},
		},
		{
			name: "non-nullable body with required properties - allowed",
			schema: `type: object
              required:
                - page
              properties:
                page:
                  type: integer`,
			wantErrs: []string{},
		},
		{
			name:     "parameter-only pagination with a nullable required body - allowed",
			schema:   "",
			wantErrs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := paginationNullableRequestBodyParamsOnlySpec
			if tt.schema != "" {
				spec = fmt.Sprintf(paginationNullableRequestBodySpecTemplate, tt.schema)
			}
			schema := []byte(spec)

			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.PaginationNullableRequestBody{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), schema, "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
