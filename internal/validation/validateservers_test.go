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

func TestValidation_ValidateServers_Success(t *testing.T) {
	type args struct {
		schema        string
		baseServerURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "passes validation with global servers",
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
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "passes validation with global servers and x-speakeasy-server-id",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
    x-speakeasy-server-id: test
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "passes validation with path servers",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
paths:
  /test:
    servers:
      - url: http://localhost:8080
    get:
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "passes validation with op servers",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
paths:
  /test:
    get:
      servers:
        - url: http://localhost:8080
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "passes validation with base server url",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 0.0.1
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`,
				baseServerURL: "http://localhost:8080",
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
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateServers{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func TestValidation_ValidateServers_Errors(t *testing.T) {
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
			name: "no servers found",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths: {}`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation hint: [line 1:1] generator-validate-servers - no servers found in document, either add servers to the document or set a `baseServerUrl` in the `gen.yaml` config file",
			},
		},
		{
			name: "no servers found for operations",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
    post:
      servers:
        - url: http://localhost:8080
      responses:
        '200':
          description: OK`,
				baseServerURL: "",
			},
			wantErrs: []string{
				"validation hint: [line 7:5] generator-validate-servers - no servers found for operation `get` `/test` either locally, on the path or globally",
			},
		},
		{
			name: "fails with unmatched x-speakeasy-server-ids",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
    x-speakeasy-server-id: test
  - url: http://localhost:8081
paths: {}`,
			},
			wantErrs: []string{
				"validation error: [line 5:1] generator-validate-servers - if using server IDs (via `x-speakeasy-server-id` extension), all servers must have a unique ID",
				"validation error: [line 8:5] generator-validate-servers - server is missing an ID (add `x-speakeasy-server-id` extension)",
			},
		},
		{
			name: "fails with non-unique x-speakeasy-server-ids",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
    x-speakeasy-server-id: test
  - url: http://localhost:8081
    x-speakeasy-server-id: test
paths: {}`,
			},
			wantErrs: []string{
				"validation error: [line 5:1] generator-validate-servers - if using server IDs (via `x-speakeasy-server-id` extension), all servers must have a unique ID",
				"validation error: [line 8:5] generator-validate-servers - server ID `test` is not unique",
			},
		},
		{
			name: "fails with unmatched server names in OpenAPI 3.2",
			args: args{
				schema: `openapi: 3.2.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
    name: production
  - url: http://localhost:8081
paths: {}`,
			},
			wantErrs: []string{
				"validation error: [line 5:1] generator-validate-servers - if using server IDs (via `x-speakeasy-server-id` extension or `name` field), all servers must have a unique ID",
				"validation error: [line 8:5] generator-validate-servers - server is missing an ID (add `x-speakeasy-server-id` extension or `name` field)",
			},
		},
		{
			name: "fails with non-unique server names in OpenAPI 3.2",
			args: args{
				schema: `openapi: 3.2.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
    name: production
  - url: http://localhost:8081
    name: production
paths: {}`,
			},
			wantErrs: []string{
				"validation error: [line 5:1] generator-validate-servers - if using server IDs (via `x-speakeasy-server-id` extension or `name` field), all servers must have a unique ID",
				"validation error: [line 8:5] generator-validate-servers - server ID `production` is not unique",
			},
		},
		{
			name: "fails if variable with same name is of a different type in different servers",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://{hostname}:8080
    variables:
      hostname:
        default: localhost
        enum:
          - localhost
          - api.com
  - url: http://{hostname}:8081
    variables:
      hostname:
        default: localhost
paths: {}`,
			},
			wantErrs: []string{
				"validation hint: [line 15:7] generator-validate-servers - variable `hostname` (type: `string`) has different types in different servers, conflicts with variable (type: `enum`) at line `8`",
			},
		},
		{
			name: "fails if missing default value in enum",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://{hostname}:8080
    variables:
      hostname:
        default: test
        enum:
          - localhost
          - api.com
paths: {}`,
			},
			wantErrs: []string{
				// Built-in validation handles this check
				"validation error: [line 9:18] validation-allowed-values - serverVariable.default must be one of [`localhost, api.com`]",
			},
		},
		{
			// Reference: internal issue reference
			name: "fails if invalid url variable syntax - double curly braces",
			args: args{
				schema: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://{{hostname}}:8080
    variables:
      hostname:
        default: example.com
        enum:
          - example.com
paths: {}`,
			},
			wantErrs: []string{
				// Built-in validation handles double curly braces check
				"validation error: [line 6:10] validation-invalid-syntax - server variable `{hostname}` is not defined. Use single curly braces for variable substitution",
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
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateServers{}).ID()}))
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
