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

func Test_DuplicateInlineSchemas_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "duplicate inline schemas with single valid outlier",
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
      parameters:
        - name: test
          in: query
          description: test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - message
            properties:
              code:
                type: integer
                format: int32
              message:
                type: string
      responses:
        '200':
          description: OK
        '5XX':
          description: test response for slightly different inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - messages
                properties:
                  code:
                    type: integer
                    format: int32
                  messages:
                    type: string
        default:
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 17:13] generator-duplicate-inline-schemas - `2` duplicates of `object` \"`code`: `integer`, `message`: `string`\" at lines [17,50]",
			},
		},
		{
			name: "all duplicate inline schemas",
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
      parameters:
        - name: test
          in: query
          description: test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - message
            properties:
              code:
                type: integer
                format: int32
              message:
                type: string
      responses:
        '200':
          description: OK
        '5XX':
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
        default:
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 17:13] generator-duplicate-inline-schemas - `3` duplicates of `object` \"`code`: `integer`, `message`: `string`\" at lines [17,35,50]",
			},
		},
		{
			name: "multiple sets of duplicate inline schemas",
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
      parameters:
        - name: test
          in: query
          description: test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - message
            properties:
              code:
                type: integer
                format: int32
              message:
                type: string
        - name: test2
          in: query
          description: nested test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - obj
            properties:
              code:
                type: integer
                format: int32
              obj:
                type: object
                required:
                  - field1
                  - field2
                properties:
                  field1:
                    type: string
                  field2:
                    type: array
                    items:
                      type: integer
        - name: test3
          in: query
          description: test param for primitive type inline schema
          required: false
          schema:
            type: string
      responses:
        '200':
          description: nested test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - obj
                properties:
                  code:
                    type: integer
                    format: int32
                  obj:
                    type: object
                    required:
                      - field1
                      - field2
                    properties:
                      field1:
                        type: string
                      field2:
                        type: array
                        items:
                          type: integer
        '3XX':
          description: test response for primitive type inline schema
          content:
            application/json:
              schema:
                type: string
        '4XX':
          description: test response for slightly different inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - obj
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  obj:
                    type: object
                    required:
                      - field1
                      - field2
                    properties:
                      field1:
                        type: string
                      field2:
                        type: array
                        items:
                          type: integer
                  message:
                    type: string
        '5XX':
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
        default:
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 17:13] generator-duplicate-inline-schemas - `3` duplicates of `object` \"`code`: `integer`, `message`: `string`\" at lines [17,123,138]",
				"validation hint: [line 32:13] generator-duplicate-inline-schemas - `2` duplicates of `object` \"`code`: `integer`, `obj`: `object`\" at lines [32,64]",
				"validation hint: [line 41:17] generator-duplicate-inline-schemas - `3` duplicates of `object` \"`field1`: `string`, `field2`: `array`\" at lines [41,73,105]",
			},
		},
		{
			name: "duplicate inline schemas across different operations",
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
      parameters:
        - name: test
          in: query
          description: test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - message
            properties:
              code:
                type: integer
                format: int32
              message:
                type: string
      responses:
        '200':
          description: OK
        default:
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string
  /test2:
    get:
      operationId: TestSomethingElse
      parameters:
        - name: test2
          in: query
          description: test param for duplicate inline schema
          required: false
          schema:
            type: object
            required:
              - code
              - message
            properties:
              code:
                type: integer
                format: int32
              message:
                type: string
      responses:
        '200':
          description: OK
        default:
          description: test response for duplicate inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - code
                  - message
                properties:
                  code:
                    type: integer
                    format: int32
                  message:
                    type: string`,
			},
			wantErrs: []string{
				"validation hint: [line 17:13] generator-duplicate-inline-schemas - `4` duplicates of `object` \"`code`: `integer`, `message`: `string`\" at lines [17,35,54,72]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.DuplicateInlineSchemas{}).ID()}))
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

func Test_DuplicateInlineSchemas_Enums(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "duplicate inline enum schemas",
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
      parameters:
        - name: status1
          in: query
          schema:
            type: integer
            enum: [1, 2, 3]
        - name: status2
          in: query
          schema:
            type: integer
            enum: [1, 2, 3]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: integer
                    enum: [1, 2, 3]`,
			},
			wantErrs: []string{
				"validation hint: [line 15:13] generator-duplicate-inline-schemas - `3` duplicates of `enum` \"1, 2, 3\" at lines [15,20,31]",
			},
		},
		{
			name: "duplicate inline string enum schemas",
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
      parameters:
        - name: color1
          in: query
          schema:
            type: string
            enum: [red, green, blue, yellow]
        - name: color2
          in: query
          schema:
            type: string
            enum: [red, green, blue, yellow]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  color:
                    type: string
                    enum: [red, green, blue, yellow]`,
			},
			wantErrs: []string{
				"validation hint: [line 15:13] generator-duplicate-inline-schemas - `3` duplicates of `enum` \"red, green, blue... `1` more\" at lines [15,20,31]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.DuplicateInlineSchemas{}).ID()}))
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

func Test_DuplicateInlineSchemas_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "no duplicate inline schemas",
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
      parameters:
        - name: test
          in: query
          description: test param for inline schema
          required: false
          schema:
            type: object
            required:
              - test1
              - test2
            properties:
              test1:
                type: integer
              test2:
                type: string
      responses:
        '200':
          description: OK
        default:
          description: test response for slightly different inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - test1
                  - test2
                properties:
                  test1:
                    type: integer
                  test2:
                    type: integer
  /test2:
    get:
      operationId: TestSomethingElse
      parameters:
        - name: test2
          in: query
          description: another test param for inline schema
          required: false
          schema:
            type: object
            required:
              - test1
              - test2
            properties:
              test1:
                type: string
              test2:
                type: integer
      responses:
        '200':
          description: OK
        default:
          description: another test response for slightly different inline schema
          content:
            application/json:
              schema:
                type: object
                required:
                  - test1
                  - test2
                properties:
                  test1:
                    type: string
                  test2:
                    type: integer
                    format: int32`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateInlineSchemas{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}
