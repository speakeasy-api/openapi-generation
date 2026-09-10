package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidation_BuiltInValidation tests that built-in OpenAPI validation
// during parsing catches JSON schema violations. ValidateJSONSchema rule has been
// removed as it's now handled by the OpenAPI library's built-in validation.
func TestValidation_BuiltInValidation(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "empty schema",
			args: args{
				schema: ``,
			},
			wantErrs: []string{
				"empty document",
			},
		},
		{
			name: "not an openapi document",
			args: args{
				schema: `something: else`,
			},
			wantErrs: []string{
				"validation error: [line 1:1] speakeasy-validate-document - document must have a `paths` or `webhooks` object with at least one entry",
				"validation warn: [line 1:1] validation-required-field - `openapi.info` is required",
				"validation error: [line 1:1] validation-required-field - `openapi.openapi` is required",
				"validation error: [line 1:1] validation-supported-version - openapi.openapi invalid OpenAPI version : invalid version ",
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation hint: [line 1:1] generator-validate-servers - no servers found in document, either add servers to the document or set a `baseServerUrl` in the `gen.yaml` config file",
				"validation warn: [line 1:1] validation-unknown-properties - unknown property `something` found",
			},
		},
		{
			name: "missing `info` section in openapi specification now reports error (stricter than old system)",
			args: args{
				schema: `openapi: 3.1.0`,
			},
			wantErrs: []string{
				"validation error: [line 1:1] speakeasy-validate-document - document must have a `paths` or `webhooks` object with at least one entry",
				"validation warn: [line 1:1] validation-required-field - `openapi.info` is required",
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation hint: [line 1:1] generator-validate-servers - no servers found in document, either add servers to the document or set a `baseServerUrl` in the `gen.yaml` config file",
			},
		},
		{
			name: "missing `required` field for path param now reports validation error (stricter)",
			args: args{
				schema: utils.Dedent(`
openapi: 3.1.0
info:
  title: FastAPI
  version: 0.1.0
  description: stop complaining
servers:
  - url: 'https://api.example.com'
    description: Production server
paths:
  '/items/{random_id}':
    get:
      summary: Read Item
      operationId: read_item_items__item_id__get
      parameters:
        - name: random_id
          in: path
          schema:
            type: boolean
            title: random id
      responses:
        '200':
          description: Successful Response
          content:
            application/json:
              schema: {}
`),
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation warn: [line 15:11] validation-required-field - `parameter.in=path` requires `required=true`",
				"validation hint: [line 21:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
		{
			name: "invalid $ref in MediaType now caught by built-in validation",
			args: args{
				schema: `openapi: 3.0.3
info:
  title: Test API
  description: test
  version: 1.0.0
servers:
  - url: https://api.example.com
    description: Test API
paths:
  /widgets:
    get:
      summary: All Widgets
      operationId: listWidgets
      parameters:
        - name: limit
          in: query
          description: Number of widgets to return
          required: false
          schema:
            type: integer
        - name: offset
          in: query
          description: Index to start widget list at
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: A list of widgets
          content:
            application/json:
              $ref: '#/components/schemas/WidgetsResponse'

components:
  schemas:
    Widgets:
      type: array
      items:
        type: string
        example: '00000000-0000-4000-8000-000000000001'

    WidgetsResponse:
      type: object
      properties:
        paging:
          type: object
          properties:
            count:
              type: integer
            offset:
              type: integer
        widgets:
          $ref: '#/components/schemas/Widgets'
`,
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation hint: [line 12:7] generator-pagination - pagination might be supported by this operation - consider adding `x-speakeasy-pagination` extension",
				"validation hint: [line 28:9] generator-missing-error-response - an error response should be defined for all operations",
				"validation warn: [line 32:15] validation-unknown-properties - unknown property `$ref` found",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			// Use standard ruleset - built-in validation happens during parsing
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
