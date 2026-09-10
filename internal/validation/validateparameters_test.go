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

func TestValidation_ValidateParameters_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "different `in` values don't conflict",
			args: args{
				schema: utils.Dedent(`
        openapi: 3.0.0
        info:
          title: Test
          version: 0.0.1
        servers:
          - url: http://localhost:8080
        paths:
          /test:
            get:
              parameters:
                - name: action
                  in: query
                  schema:
                    type: string
                - name: action
                  in: header
                  schema:
                    type: string
              responses:
                '200':
                  description: OK`),
			},
		},
		{
			name: "ref parameters with same name and same `in` value should not conflict",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
 - url: http://localhost:8080
paths:
 /test:
   get:
     parameters:
       - $ref: '#/components/parameters/action'
       - name: action
         in: query
         schema:
           type: string
     responses:
       '200':
         description: OK
components:
 parameters:
   action:
     name: action
     in: query
     schema:
       type: string`),
			},
		},
		{
			name: "ref parameter names with different casing does not conflict",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
     parameters:
       - $ref: '#/components/parameters/action_test'
       - $ref: '#/components/parameters/actionTest'
     responses:
       '200':
         description: OK
components:
 parameters:
   action_test:
     name: action_test
     in: query
     schema:
       type: string
   actionTest:
     name: actionTest
     in: query
     schema:
       type: string`),
			},
		},
		{
			name: "overrides bypass conflicts",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: startTime
          in: query
          schema:
            type: string
        - name: start_time
          x-speakeasy-name-override: startTimeDeprecated
          in: query
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
		},
		{
			name: "valid header parameter with style simple",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: x-header-1
          in: header
          style: simple
          schema:
            type: string
        - name: x-header-2
          in: header
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateParameters{}).ID(), (&validation.PathParams{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func TestValidation_ValidateParameters_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "empty name",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name:
          in: query
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{
				"validation error: [line 11:11] generator-validate-parameters - parameter must have a non-empty `name` property",
				"validation error: [line 11:16] validation-type-mismatch - parameter.name expected `string`, got `null`",
			},
		},
		{
			name: "reserved header name",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: Accept
          in: header
          schema:
            type: string
        - name: Content-Type
          in: header
          schema:
            type: string
        - name: Authorization
          in: header
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{
				"validation warn: [line 11:11] generator-validate-parameters - header name `Accept` is reserved and will be ignored",
				"validation warn: [line 15:11] generator-validate-parameters - header name `Content-Type` is reserved and will be ignored",
				"validation warn: [line 19:11] generator-validate-parameters - header name `Authorization` is reserved and will be ignored",
			},
		},
		{
			name: "parameter names with different casing should not conflict",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: startTime
          in: query
          schema:
            type: string
        - name: start_time
          in: query
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{},
		},
		{
			name: "parameters with same name and same `in` value should not conflict",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: action
          in: query
          schema:
            type: string
        - name: action
          in: query
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{
				"validation warn: [line 15:11] validation-operation-parameters - parameter \"action\" is duplicated in GET operation at path \"/test\"",
			},
		},
		{
			name: "header parameter must use style simple",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
        - name: x-action
          in: header
          style: spaceDelimited
          schema:
            type: string
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{
				"validation error: [line 13:18] validation-allowed-values - parameter.style must be one of [`simple`] for in=header",
			},
		},
		{
			name: "Duplicate x-speakeasy-name-override gives errors ",
			args: args{
				schema: utils.Dedent(`
openapi: 3.1.0
info:
  title: FastAPI
  version: 0.1.0
  description: stop complaining
servers:
  - url: https://api.example.com
    description: Production server
paths:
  /items/{random_id}:
    get:
      summary: Read Item
      operationId: read_item_items__item_id__get
      parameters:
        - name: random_id
          in: path
          required: true
          schema:
            type: boolean
            title: random id
        - name: random_id
          x-speakeasy-name-override: new_random_id
          in: query
          required: true
          schema:
            type: boolean
            title: random id
        - name: random_id
          x-speakeasy-name-override: new_random_id
          in: query
          required: true
          schema:
            type: integer
            title: random id
      responses:
        "200":
          description: Successful Response
          content:
            application/json:
              schema: {}`),
			},
			wantErrs: []string{
				"validation error: [line 28:11] generator-validate-parameters - `new_random_id` (`NewRandomID`) will collide with `new_random_id` (`NewRandomID`) [line `21`] when converted to field name",
				"validation warn: [line 28:11] validation-operation-parameters - parameter \"random_id\" is duplicated in GET operation at path \"/items/{random_id}\"",
			},
		},
		{
			name: "x-speakeasy-allow-empty-value on non-query parameter",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          x-speakeasy-allow-empty-value: true
      responses:
        '200':
          description: OK`),
			},
			wantErrs: []string{
				"validation error: [line 11:11] generator-validate-parameters - parameter with `in` == `path` cannot use `x-speakeasy-allow-empty-value` extension, only query parameters are supported",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateParameters{}).ID(), (&validation.PathParams{}).ID()}))
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

func TestValidation_ValidateParameterRefs_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "same ref across operations don't conflict",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
 - url: http://localhost:8080
paths:
 /test:
   get:
     parameters:
       - $ref: '#/components/parameters/action'
     responses:
       '200':
         description: OK
   post:
     parameters:
       - $ref: '#/components/parameters/action'
     responses:
       '200':
         description: OK

components:
 parameters:
   action:
     name: action
     in: query
     schema:
       type: string`),
			},
		},
		{
			name: "different in values don't conflict in refs",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
 - url: http://localhost:8080
paths:
 /test:
   get:
     parameters:
       - $ref: '#/components/parameters/action'
       - $ref: '#/components/parameters/action2'
     responses:
       '200':
         description: OK
components:
 parameters:
   action:
     name: action
     in: query
     schema:
       type: string
   action2:
     name: action
     in: header
     schema:
       type: string`),
			},
		},
		{
			name: "overrides bypass conflicts in ref",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
       - $ref: '#/components/parameters/action'
       - name: action
         in: query
         schema:
           type: string
      responses:
        '200':
          description: OK
components:
 parameters:
   action:
     name: action
     in: query
     x-speakeasy-name-override: OverrideActionName
     schema:
       type: string`),
			},
		},
		{
			name: "overrides bypass conflicts outside ref",
			args: args{
				schema: utils.Dedent(`
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      parameters:
       - $ref: '#/components/parameters/action'
       - name: action
         in: query
         schema:
           type: string
         x-speakeasy-name-override: OverrideActionName
      responses:
        '200':
          description: OK
components:
 parameters:
   action:
     name: action
     in: query
     schema:
       type: string`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateParameters{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}
