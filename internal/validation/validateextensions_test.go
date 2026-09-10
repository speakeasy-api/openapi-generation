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

func TestValidation_ValidateExtensions_Success(t *testing.T) {
	type args struct {
		schema        string
		baseServerURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "must have name",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - name: globalQueryParam
      in: query
      required: true
      schema:
        type: string
    - name: globalPathParam
      in: path
      required: true
      schema:
        type: string`,
				baseServerURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{
					BaseServerURL: tt.args.baseServerURL,
				},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateExtensions{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func TestValidation_ValidateExtensions_Errors(t *testing.T) {
	type args struct {
		schema        string
		baseServerURL string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "must have name",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - in: query
      required: true
      schema:
        type: string`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation error: [line 16:7] validation-required-field - `parameter.name` is required",
			},
		},
		{
			name: "must have schema",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - name: globalQueryParam
      in: query
      required: true`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation error: [line 16:7] generator-validate-extensions - `x-speakeasy-globals` parameter must have `schema`",
			},
		},
		{
			name: "only primitives allow",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - name: globalQueryParam
      in: query
      required: true
      schema:
        type: object`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation error: [line 20:9] generator-validate-extensions - only primitive types are allowed in `x-speakeasy-globals` parameters: found `object`",
			},
		},
		{
			name: "only unique names allowed",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - name: testSomething
      in: query
      required: true
      schema:
        type: string
    - name: test_something
      in: path
      required: true
      schema:
        type: integer`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation error: [line 21:7] generator-validate-extensions - `test_something` (`TestSomething`) will collide with `testSomething` (`TestSomething`) [line `16`], global parameters must be unique",
			},
		},
		{
			name: "allow duplicate names when schemas match",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: TestSomething
      responses:
        '200':
          description: OK
x-speakeasy-globals:
  parameters:
    - name: orgId
      in: query
      required: true
      schema:
        $ref: "#/components/schemas/OrgId"
    - name: orgId
      in: path
      required: true
      schema:
        $ref: "#/components/schemas/OrgId"
components:
  schemas:
    OrgId:
      type: string`,
				baseServerURL: "",
			},
			wantErrs: []string{},
		},
		{
			name: "fails if global parameter conflicts with server variable",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080/{testParam}
    variables:
      testParam:
        default: test
x-speakeasy-globals:
  parameters:
    - name: testParam
      in: query
      required: true
      schema:
        type: string`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation error: [line 12:7] generator-validate-extensions - global parameter `testParam` (`TestParam`) will collide with server variable `testParam` (`TestParam`) [line `8`], global parameters must be unique",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{
					BaseServerURL: tt.args.baseServerURL,
				},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateExtensions{}).ID()}))
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
