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

func TestValidation_ValidateTypes_Arrays_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "arrays using boolean items with prefixItems attribute inline valid",
			args: args{
				schema: `openapi: 3.1.0
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
          content:
            application/json:
              schema:
                type: array
                prefixItems:
                  - type: string
                  - type: number
                items: true
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "arrays using boolean items with prefixItems attribute in component valid",
			args: args{
				schema: `openapi: 3.1.0
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
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
        '400':
          description: Bad Request
components:
  schemas:
    test:
      type: array
      prefixItems:
        - type: string
        - type: number
      items: true`,
			},
		},
		{
			name: "arrays using boolean items with prefixItems attribute inline nested valid",
			args: args{
				schema: `openapi: 3.1.0
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
          content:
            application/json:
              schema:
                type: object
                properties:
                  test:
                    type: array
                    prefixItems:
                      - type: string
                      - type: number
                    items: true
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "arrays with items attribute in 3.0.X documents valid",
			args: args{
				schema: `openapi: 3.0.3
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
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
        '400':
          description: Bad Request`,
			},
		},
		{
			name: "array-valued examples are skipped during validation",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Object:
      type: object
      properties:
        items:
          type: array
          items:
            type: object
            properties:
              type:
                type: string
      example:
        items:
          - nested:
              type: pass`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateTypes{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func TestValidation_ValidateTypes_Arrays_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "arrays using boolean items attribute inline",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: true`,
			},
			wantErrs: []string{
				"validation error: [line 17:17] generator-validate-types - `items` as `boolean` only valid when used with `prefixItems`",
			},
		},
		{
			name: "arrays using boolean items attribute in component",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: array
      items: true`,
			},
			wantErrs: []string{
				"validation error: [line 21:7] generator-validate-types - `items` as `boolean` only valid when used with `prefixItems`",
			},
		},
		{
			name: "arrays using boolean items attribute inline nested",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  test:
                    type: array
                    items: true`,
			},
			wantErrs: []string{
				"validation error: [line 20:21] generator-validate-types - `items` as `boolean` only valid when used with `prefixItems`",
			},
		},
		{
			name: "arrays require items attribute in 3.0.X documents",
			args: args{
				schema: `openapi: 3.0.3
info:
  title: Test
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array`,
			},
			wantErrs: []string{
				"validation error: [line 16:17] generator-validate-types - `items` attribute required for arrays in OpenAPI 3.0.X documents",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateTypes{}).ID()}))
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
