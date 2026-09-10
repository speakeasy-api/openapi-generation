package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// September 2024: This test sensechecks any rules that we want to deprioritize from error to warn severity.
func TestValidation_DeprioritizeErrors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "additional property keys validation - should be warning",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      operationId: test
      requestBody:
        content:
          application/json:
            schema:
              type: object
              example: foobar
              properties:
                test:
                  type: string
                  description: Test
                  name: foo
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation warn: [line 19:19] validation-unknown-properties - unknown property `name` found",
			},
		},
		{
			name: "ignore required field for path parameters",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test/{test}:
    get:
      operationId: test
      parameters:
        - name: test
          in: path
          required: false
          schema:
            type: string
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation warn: [line 14:21] validation-required-field - `parameter.in=path` requires `required=true`",
			},
		},
		{
			name: "additionalProperties: true - should be warning",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      operationId: test
      requestBody:
        content:
          application/json:
            schema:
              type: object
              example: foobar
              properties:
                additionalProperties: true
                test:
                  type: string
                  description: Test
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			// New linter doesn't validate this case (treating additionalProperties as a property name is technically valid)
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{}))
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
