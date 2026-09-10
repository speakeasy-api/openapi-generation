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

func TestValidation_ValidateDuplicatedEntryInEnum(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "enum with duplicated entry",
			args: args{
				schema: utils.Dedent(`
openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: testOp
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: string
                enum:
                    - foo
                    - bar
                    - foo
              example: "foo"
        '500':
          description: Error
          content:
            application/json:
              schema:
                type: object
              example: {}
`),
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
				"validation warn: [line 21:23] semantic-duplicated-enum - enum contains a duplicate: `foo`",
			},
		},
		{
			name: "enum without duplicated entry",
			args: args{
				schema: utils.Dedent(`
openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: testOp
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: string
                enum:
                  - foo
                  - bar
              example: "foo"
        '500':
          description: Error
          content:
            application/json:
              schema:
                type: object
              example: {}
`),
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended)
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
