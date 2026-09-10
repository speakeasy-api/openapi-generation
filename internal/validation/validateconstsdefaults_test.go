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

func TestValidation_ValidateConstsDefaults_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "consts and defaults used successfully succeed",
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
        '400':
          description: Bad Request
components:
  schemas:
    test:
      type: object
      properties:
        stringConst:
          type: string
          const: test
        stringDefault:
          type: string
          default: test
        stringNullableConst:
          type: string
          nullable: true
          const: null
        stringNullConst:
          type: [string, 'null']
          const: null
        stringNullableDefault:
          type: string
          nullable: true
          default: null
        stringNullDefault:
          type: [string, 'null']
          default: null
        integerConst:
          type: integer
          const: 1
        integerDefault:
          type: integer
          default: 1
        numberConst:
          type: number
          const: 1.1
        numberDefault:
          type: number
          default: 1.1
        numberIntConst:
          type: number
          const: 1
        numberIntDefault:
          type: number
          default: 1
        booleanConst:
          type: boolean
          const: true
        booleanDefault:
          type: boolean
          default: true
        nullConst:
          type: 'null'
          const: null
        nullDefault:
          type: 'null'
          default: null
        multipleTypesConst:
          type: [string, integer]
          const: 1
        multipleTypesDefault:
          type: [string, integer]
          default: test
        arrayConst:
          type: array
          const: []
        arrayDefault:
          type: array
          default: []
        objConst:
          type: array
          const: {}
        objDefault:
          type: array
          default: {}
        stringEnumDefault:
          type: string
          enum: [test, test2]
          default: test
        stringEnumConst:
          type: string
          enum: [test, test2]
          const: test
        integerEnumDefault:
          type: integer
          enum: [1, 2]
          default: 2
        integerEnumConst:
          type: integer
          enum: [1, 2]
          const: 2`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateConstsDefaults{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			assert.Empty(t, res.GetValidationErrors())
		})
	}
}

func TestValidation_ValidateConstsDefaults_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "consts and defaults used successfully succeed",
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
components:
  schemas:
    test:
      type: object
      properties:
        stringConstInvalid:
          type: string
          const: 1
        stringDefaultInvalidTypeArray:
          type: string
          default: ["test"]
        stringDefaultInvalidTypeNumber:
          type: string
          default: 1.1
        stringNullableConstInvalid:
          type: string
          nullable: true
          const: true
        stringNullConstInvalid:
          type: [string, 'null']
          const: 1
        stringNullableDefaultInvalid:
          type: string
          nullable: true
          default: 1.1
        stringNullDefaultInvalid:
          type: [string, 'null']
          default: false
        integerConstInvalid:
          type: integer
          const: '1'
        integerDefaultInvalid:
          type: integer
          default: 1.1
        numberConstInvalid:
          type: number
          const: '1.1'
        numberDefaultInvalid:
          type: number
          default: false
        booleanConstInvalid:
          type: boolean
          const: 'true'
        booleanDefaultInvalid:
          type: boolean
          default: 1
        nullConstInvalid:
          type: 'null'
          const: test
        nullDefaultInvalid:
          type: 'null'
          default: 1
        multipleTypesConstInvalid:
          type: [string, integer]
          const: true
        multipleTypesDefault:
          type: [string, integer]
          default: null
        stringEnumDefaultNotInEnum:
          type: string
          enum: [test, test2]
          default: test3
        stringEnumDefaultInEnumWrongType:
          type: string
          enum: [test, test2]
          default: [test, test2]
        stringEnumConst:
          type: string
          enum: [test, test2]
          const: ""
        integerEnumDefault:
          type: integer
          enum: [1, 2]
          default: 3
        integerEnumConst:
          type: integer
          enum: [1, 2]
          const: -1`,
			},
			wantErrs: []string{
				"validation warn: [line 20:11] generator-validate-consts-defaults - `const` value `1` does not match type `string`",
				"validation warn: [line 23:11] generator-validate-consts-defaults - `default` value `[test]` does not match type `string`",
				"validation warn: [line 26:11] generator-validate-consts-defaults - `default` value `1.1` does not match type `string`",
				"validation warn: [line 30:11] generator-validate-consts-defaults - `const` value `true` does not match type `string`",
				"validation warn: [line 33:11] generator-validate-consts-defaults - `const` value `1` does not match type `string`",
				"validation warn: [line 37:11] generator-validate-consts-defaults - `default` value `1.1` does not match type `string`",
				"validation warn: [line 40:11] generator-validate-consts-defaults - `default` value `false` does not match type `string`",
				"validation warn: [line 43:11] generator-validate-consts-defaults - `const` value `1` does not match type `integer`",
				"validation warn: [line 46:11] generator-validate-consts-defaults - `default` value `1.1` does not match type `integer`",
				"validation warn: [line 49:11] generator-validate-consts-defaults - `const` value `1.1` does not match type `number`",
				"validation warn: [line 52:11] generator-validate-consts-defaults - `default` value `false` does not match type `number`",
				"validation warn: [line 55:11] generator-validate-consts-defaults - `const` value `true` does not match type `boolean`",
				"validation warn: [line 58:11] generator-validate-consts-defaults - `default` value `1` does not match type `boolean`",
				"validation warn: [line 61:11] generator-validate-consts-defaults - `const` value `test` does not match type `null`",
				"validation warn: [line 64:11] generator-validate-consts-defaults - `default` value `1` does not match type `null`",
				"validation warn: [line 67:11] generator-validate-consts-defaults - `const` value `true` does not match type `integer`",
				"validation warn: [line 67:11] generator-validate-consts-defaults - `const` value `true` does not match type `string`",
				"validation warn: [line 70:11] generator-validate-consts-defaults - `default` value `null` cannot be used with non-nullable type `integer`",
				"validation warn: [line 70:11] generator-validate-consts-defaults - `default` value `null` cannot be used with non-nullable type `string`",
				"validation warn: [line 74:11] generator-validate-consts-defaults - `default` value `test3` not present in `enum` values",
				"validation warn: [line 78:11] generator-validate-consts-defaults - `default` value `[test test2]` does not match type `string`",
				"validation warn: [line 82:11] generator-validate-consts-defaults - `const` value `` not present in `enum` values",
				"validation warn: [line 86:11] generator-validate-consts-defaults - `default` value `3` not present in `enum` values",
				"validation warn: [line 90:11] generator-validate-consts-defaults - `const` value `-1` not present in `enum` values",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateConstsDefaults{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errStrs := make([]string, len(res.GetValidationErrors()))
			for i, err := range res.GetValidationErrors() {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
