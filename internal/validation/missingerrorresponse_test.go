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

func TestValidation_ValidationMissingErrorResponses(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "has error response - is valid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        '500':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
          `,
			},
			wantErrs: []string{},
		},
		{
			name: "missing error response - is invalid",
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
      operationId: TestSomething
      responses:
        200:
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 12:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
		{
			name: "wildcard error response - 4XX - is valid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        '4XX':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{},
		},
		{
			name: "wildcard error response - 5XX - is valid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        '5XX':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{},
		},
		{
			name: "default response - is valid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        'default':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{},
		},
		{
			name: "unrecognised error response series - is invalid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        '9XX':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 12:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
		{
			name: "single digit error response code - is invalid",
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
      operationId: TestSomething
      responses:
        '200':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string
        '5':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 12:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
		{
			name: "unrecognised 4XX status code",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: string
        '499':
          description: Bad Request`,
			},
			wantErrs: []string{
				"validation hint: [line 9:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
		{
			name: "unrecognised 5XX status code",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: string
        '599':
          description: Bad Request`,
			},
			wantErrs: []string{
				"validation hint: [line 9:9] generator-missing-error-response - an error response should be defined for all operations",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.MissingErrorResponse{}).ID()}))
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
