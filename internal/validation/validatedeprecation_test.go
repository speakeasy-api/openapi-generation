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

func Test_ValidateDeprecation_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "errors when operation deprecated attribute is missing",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: testGet
      x-speakeasy-deprecation-replacement: test2Get
      x-speakeasy-deprecation-message: test message
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{
				"validation error: [line 10:7] generator-validate-deprecation - operation must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-replacement`",
				"validation error: [line 10:7] generator-validate-deprecation - operation must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-message`",
			},
		},
		{
			name: "errors when operation deprecated attribute is not set to true",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: testGet
      deprecated: false
      x-speakeasy-deprecation-replacement: test2Get
      x-speakeasy-deprecation-message: test message
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{
				"validation error: [line 10:7] generator-validate-deprecation - operation must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-replacement`",
				"validation error: [line 10:7] generator-validate-deprecation - operation must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-message`",
			},
		},
		{
			name: "error when unable to find replacement operation",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: testGet
      x-speakeasy-deprecation-replacement: test2Get
      deprecated: true
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 10:7] generator-validate-deprecation - `x-speakeasy-deprecation-replacement` not found, operation with id `test2Get` doesn't exist"},
		},
		{
			name: "errors when parameters deprecated attribute is missing",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      parameters:
        - name: testParam
          in: query
          schema:
            type: string
          x-speakeasy-deprecation-replacement: test2Param
          x-speakeasy-deprecation-message: test message
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{
				"validation error: [line 11:11] generator-validate-deprecation - parameter must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-replacement`",
				"validation error: [line 11:11] generator-validate-deprecation - parameter must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-message`",
			},
		},
		{
			name: "errors when parameters deprecated attribute is not set to true",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      parameters:
        - name: testParam
          in: query
          schema:
            type: string
          deprecated: false
          x-speakeasy-deprecation-replacement: test2Param
          x-speakeasy-deprecation-message: test message
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{
				"validation error: [line 11:11] generator-validate-deprecation - parameter must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-replacement`",
				"validation error: [line 11:11] generator-validate-deprecation - parameter must have a `deprecated` property set to `true` to use `x-speakeasy-deprecation-message`",
			},
		},
		{
			name: "error when unable to find replacement parameter",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      parameters:
        - name: testParam
          in: query
          schema:
            type: string
          x-speakeasy-deprecation-replacement: test2Param
          deprecated: true
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 11:11] generator-validate-deprecation - `x-speakeasy-deprecation-replacement` not found, parameter with name `test2Param` doesn't exist"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateDeprecation{}).ID()}))
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
