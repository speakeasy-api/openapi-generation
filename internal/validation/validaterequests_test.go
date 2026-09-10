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

func Test_ValidateRequests_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "invalid request content type",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    post:
      requestBody:
        content:
          wrong//:
            schema:
              type: string
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{
				"validation error: [line 13:13] generator-validate-requests - content type not valid - found `wrong//`: `mime: expected token after slash`",
			},
		},
		{
			name: "invalid request content type from $ref",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    post:
      requestBody:
        $ref: '#/components/requestBodies/test'
      responses:
        '200':
          description: OK
components:
  requestBodies:
    test:
      content:
        wrong//:
          schema:
            type: string`,
			},
			wantErrs: []string{
				"validation error: [line 20:11] generator-validate-requests - content type not valid - found `wrong//`: `mime: expected token after slash`",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateRequests{}).ID()}))
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
