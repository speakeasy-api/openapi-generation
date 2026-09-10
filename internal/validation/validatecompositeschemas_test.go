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

func TestValidation_ValidateCompositeSchemas_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "single item in anyOf/allOf/oneOf",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    AnyOf:
      anyOf:
        - type: boolean
    AllOf:
      oneOf:
        - type: boolean
    OneOf:
      oneOf:
        - type: boolean`,
			},
		},
		{
			name: "\"type: 'null'\" in anyOf/allOf/oneOf",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    AnyOf:
      anyOf:
        - type: integer
        - type: 'null'
    AllOf:
      allOf:
        - type: integer
        - type: 'null'
    OneOf:
      oneOf:
        - type: integer
        - type: 'null'`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration,
				validation.WithFilteredRules([]string{(&validation.ValidateCompositeSchemas{}).ID()}),
			)
			require.NoError(t, err)
			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func TestValidation_ValidateCompositeSchemas_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "ensure not duplicate references in allOf",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
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
                allOf:
                  - $ref: '#/components/schemas/test'
                  - $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: object`,
			},
			wantErrs: []string{
				"validation error: [line 18:21] generator-validate-composite-schemas - duplicate schema reference found in `allOf` (first occurrence at index `0`, duplicate at index `1`)",
			},
		},
		{
			name: "ensure not duplicate references in anyOf",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
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
                anyOf:
                  - $ref: '#/components/schemas/test'
                  - $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: object`,
			},
			wantErrs: []string{
				"validation error: [line 18:21] generator-validate-composite-schemas - duplicate schema reference found in `anyOf` (first occurrence at index `0`, duplicate at index `1`)",
			},
		},
		{
			name: "ensure not duplicate references in oneOf",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
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
                oneOf:
                  - $ref: '#/components/schemas/test'
                  - $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: object`,
			},
			wantErrs: []string{
				"validation error: [line 18:21] generator-validate-composite-schemas - duplicate schema reference found in `oneOf` (first occurrence at index `0`, duplicate at index `1`)",
			},
		},
		{
			name: "ensure not empty anyOf/allOf/oneOf arrays",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    AnyOf:
      anyOf: []
    OneOf:
      oneOf: []
    AllOf:
      allOf: []`,
			},
			wantErrs: []string{
				"validation error: [line 8:14] validation-invalid-schema - schema.anyOf minItems: got 0, want 1",
				"validation error: [line 10:14] validation-invalid-schema - schema.oneOf minItems: got 0, want 1",
				"validation error: [line 12:14] validation-invalid-schema - schema.allOf minItems: got 0, want 1",
			},
		},
		{
			name: "ensure not null anyOf/allOf/oneOf (original YAML format)",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    AnyOf:
      anyOf:
    OneOf:
      oneOf:
    AllOf:
      allOf:`,
			},
			// Note: Null composites are type mismatches caught by built-in validation
			// These become parse errors before our custom rule runs
			// The built-in linter reports 2 errors per field (array type + sequence type)
			wantErrs: []string{
				"validation error: [line 8:7] validation-type-mismatch - schema.anyOf expected `array`, got `null`",
				"validation error: [line 8:13] validation-type-mismatch - schema.anyOf expected `sequence`, got ``",
				"validation error: [line 10:7] validation-type-mismatch - schema.oneOf expected `array`, got `null`",
				"validation error: [line 10:13] validation-type-mismatch - schema.oneOf expected `sequence`, got ``",
				"validation error: [line 12:7] validation-type-mismatch - schema.allOf expected `array`, got `null`",
				"validation error: [line 12:13] validation-type-mismatch - schema.allOf expected `sequence`, got ``",
			},
		},
		{
			name: "ensure not simply \"type: 'null'\"",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    AnyOf:
      anyOf:
        - type: 'null'
    OneOf:
      oneOf:
        - type: 'null'
    AllOf:
      allOf:
        - type: 'null'`,
			},
			wantErrs: []string{
				"validation warn: [line 9:11] generator-validate-composite-schemas - `anyOf` cannot contain `type: null` only",
				"validation warn: [line 12:11] generator-validate-composite-schemas - `oneOf` cannot contain `type: null` only",
				"validation warn: [line 15:11] generator-validate-composite-schemas - `allOf` cannot contain `type: null` only",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateCompositeSchemas{}).ID()}))
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
