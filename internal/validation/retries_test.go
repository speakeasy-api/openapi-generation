package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_ValidateRetries(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "existing global retries",
			args: args{
				schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
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
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error
x-speakeasy-retries:
  strategy: backoff
  backoff:
    initialInterval: 500        # 500 milliseconds
    maxInterval: 60000          # 60 seconds
    maxElapsedTime: 3600000     # 5 minutes
    exponent: 1.5
  statusCodes:
    - 5XX
  retryConnectionErrors: true`,
			},
			wantErrs: []string{},
		},
		{
			name: "no global retries",
			args: args{
				schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      operationId: testPost
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
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{
				"validation hint: [line 2:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
			},
		},
		{
			name: "retry extension applied to at least one operation",
			args: args{
				schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      x-speakeasy-retries:
        strategy: backoff
        backoff:
          initialInterval: 500        # 500 milliseconds
          maxInterval: 60000          # 60 seconds
          maxElapsedTime: 3600000     # 5 minutes
          exponent: 1.5
        statusCodes:
          - 5XX
        retryConnectionErrors: true
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
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{},
		},
		{
			name: "attempt count retry extension",
			args: args{
				schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      x-speakeasy-retries:
        strategy: attempt-count-backoff
        maxRetries: 2
        backoff:
          initialInterval: 500
          maxInterval: 60000
          maxElapsedTime: 3600000
          exponent: 2
        statusCodes:
          - 5XX
        retryConnectionErrors: true
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{},
		},
		{
			name: "attempt count retry extension allows zero max retries",
			args: args{
				schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      x-speakeasy-retries:
        strategy: attempt-count-backoff
        maxRetries: 0
        statusCodes:
          - 5XX
      responses:
        '200':
          description: OK
        '500':
          description: Internal Server Error`,
			},
			wantErrs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.Retries{}).ID()}))
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
