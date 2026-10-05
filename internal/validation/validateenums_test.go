package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi/openapi/linter/rules"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ValidateEnums_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid enum values",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      enum:
        - unknown_street
        - Unknown Street
      example: "Unknown Street"
      type: string`,
			},
		},
		{
			name: "valid enum with x-speakeasy-enums array",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - unknown_street
        - known_street
      x-speakeasy-enums:
        - UnknownStreet
        - KnownStreet`,
			},
		},
		{
			name: "valid enum with x-speakeasy-enums map",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - unknown_street
        - known_street
      x-speakeasy-enums:
        unknown_street: UnknownStreet
        known_street: KnownStreet`,
			},
		},
		{
			name: "valid enum with x-speakeasy-enums partial map",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - unknown_street
        - known_street
        - other_street
      x-speakeasy-enums:
        unknown_street: UnknownStreet
        known_street: KnownStreet`,
			},
		},
		{
			name: "valid enum with x-speakeasy-enum-descriptions array",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - unknown_street
        - known_street
      x-speakeasy-enum-descriptions:
        - Unknown street description
        - Known street description`,
			},
		},
		{
			name: "valid enum with x-speakeasy-enum-descriptions map",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - unknown_street
        - known_street
      x-speakeasy-enum-descriptions:
        unknown_street: Unknown street description
        known_street: Known street description`,
			},
		},
		{
			name: "valid enum with multiline x-speakeasy-enum-descriptions",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - value1
        - value2
      x-speakeasy-enum-descriptions:
        value1: |
          First line description.
          Second line preserved.
        value2: Another line`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.ValidateEnums{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func Test_ValidateEnums_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "nullable enums found",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enumRef:
                    $ref: '#/components/schemas/enum'
                  enum:
                    type: string
                    enum: [big, small]
                    nullable: true
components:
  schemas:
    enum:
      type: ["string", "null"]
      enum: [big, small]`,
			},
			wantErrs: []string{
				"validation hint: [line 23:21] generator-validate-enums - enum is nullable but does not contain a `null` value",
				"validation hint: [line 29:7] generator-validate-enums - enum is nullable but does not contain a `null` value",
			},
		},
		{
			name: "invalid enum values",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: string
      enum:
        - EAN_13
        - ean13
        - EAN13`,
			},
			wantErrs: []string{
				"validation error: [line 25:11] generator-validate-enums - enum value `EAN13` (`Ean13Upper`) will collide with `EAN_13` (`Ean13Upper`) [line `23`] when normalized, try using `x-speakeasy-enums`",
			},
		},
		{
			name: "duplicate enum values",
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
      operationId: test
      parameters:
        - name: enum
          in: query
          schema:
            type: string
            enum:
              - test1
              - test1
              - test2
              - test3
        - name: intEnum
          in: query
          schema:
            type: integer
            enum:
              - 1
              - 1
              - 2
              - 3
      responses:
        '200':
          description: OK`,
			},
			wantErrs: []string{
				"validation warn: [line 18:17] semantic-duplicated-enum - enum contains a duplicate: `test1`",
				"validation warn: [line 27:17] semantic-duplicated-enum - enum contains a duplicate: `int:1`",
			},
		},
		{
			name: "duplicate enum names from extension",
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
      operationId: test
      parameters:
        - name: enum
          in: query
          schema:
            type: integer
            format: int32
            enum:
              - 1
              - 2
              - 3
            x-speakeasy-enums:
              - test1
              - Test1
              - test2
      responses:
        '200':
          description: OK`,
			},
			wantErrs: []string{
				"validation error: [line 23:17] generator-validate-enums - enum name `Test1` (`Test1`) will collide with `test1` (`Test1`) [line `22`] when normalized, try using `x-speakeasy-enums`",
			},
		},
		{
			name: "enum wrong type",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: number
                    enum: [0.1, 0.2]`,
			},
			wantErrs: []string{
				"validation warn: [line 20:21] generator-validate-enums - only `enum` types of `string` or `integer` supported. Enum type won't be generated and will be treated as base type",
			},
		},
		{
			name: "enum wrong type ref",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    $ref: '#/components/schemas/enum'
components:
  schemas:
    enum:
      type: number
      enum: [0.1, 0.2]`,
			},
			wantErrs: []string{
				"validation warn: [line 24:7] generator-validate-enums - only `enum` types of `string` or `integer` supported. Enum type won't be generated and will be treated as base type",
			},
		},
		{
			name: "enum only 1 type allowed",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: [number, string]
                    enum: [0.1, 0.2]`,
			},
			wantErrs: []string{
				"validation error: [line 20:21] generator-validate-enums - only one type allowed for `enum`",
			},
		},
		{
			name: "enum value not an integer",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: integer
                    enum:
                      - 1
                      - 0.1
                      - hello
                      - 3`,
			},
			wantErrs: []string{
				"validation error: [line 23:25] generator-validate-enums - enum value `0.1` is not an `integer`",
				"validation error: [line 24:25] generator-validate-enums - enum value `hello` is not an `integer`",
			},
		},
		{
			name: "x-speakeasy-enums array length mismatch",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                      - value3
                    x-speakeasy-enums:
                      - Name1
                      - Name2`,
			},
			wantErrs: []string{
				"validation error: [line 26:23] generator-validate-enums - `x-speakeasy-enums` array must be the same length as enum values array",
			},
		},
		{
			name: "x-speakeasy-enum-descriptions array length mismatch",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                      - value3
                    x-speakeasy-enum-descriptions:
                      - Description1
                      - Description2`,
			},
			wantErrs: []string{
				"validation error: [line 26:23] generator-validate-enums - `x-speakeasy-enum-descriptions` array must be the same length as enum values array",
			},
		},
		{
			name: "x-speakeasy-enum-descriptions map with invalid enum value",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                    x-speakeasy-enum-descriptions:
                      value1: Description1
                      value2: Description2
                      value3: Description3`,
			},
			wantErrs: []string{
				"validation warn: [line 27:23] generator-validate-enums - `x-speakeasy-enum-descriptions` map contains key `value3` that does not exist in enum values",
			},
		},
		{
			name: "`x-speakeasy-enum-descriptions` map missing description",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                    x-speakeasy-enum-descriptions:
                      value1: Description1`,
			},
			wantErrs: []string{
				"validation warn: [line 23:25] generator-validate-enums - `x-speakeasy-enum-descriptions` map missing description for enum value `value2`",
			},
		},
		{
			name: "x-speakeasy-enums map with invalid enum value",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                    x-speakeasy-enums:
                      value1: Name1
                      value2: Name2
                      value3: Name3`,
			},
			wantErrs: []string{
				"validation error: [line 25:23] generator-validate-enums - `x-speakeasy-enums` map contains key `value3` that does not exist in enum values",
			},
		},
		{
			name: "x-speakeasy-enums invalid format",
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
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  enum:
                    type: string
                    enum:
                      - value1
                      - value2
                    x-speakeasy-enums: "invalid"`,
			},
			wantErrs: []string{
				"validation error: [line 24:40] generator-validate-enums - `x-speakeasy-enums` must be either an array or a map",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{
				(&validation.ValidateEnums{}).ID(),
				(&rules.DuplicatedEnumRule{}).ID(),
			}))
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

// Test_ValidateEnums_NullableWithCustomNames tests that nullable enums with
// x-speakeasy-enums work correctly (null should be skipped in both arrays)
func Test_ValidateEnums_NullableWithCustomNames(t *testing.T) {
	schema := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  nullableIntEnum:
                    type: integer
                    enum:
                      - 1
                      - 2
                      - 3
                      - null
                    x-speakeasy-enums:
                      - First
                      - Second
                      - Third
                      - null
                  nullableStringEnum:
                    type: string
                    enum:
                      - First
                      - Second
                      - Third
                      - null
                    x-speakeasy-enums:
                      - FirstName
                      - SecondName
                      - ThirdName
                      - null
`

	v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(schema), "", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Should have NO length mismatch errors - null values should be skipped in both enum and x-speakeasy-enums
	for _, err := range errs {
		errStr := err.Error()
		assert.NotContains(t, errStr, "`x-speakeasy-enums` array must be the same length as enum values array",
			"Nullable enums with x-speakeasy-enums should not produce length mismatch errors")
	}
}
