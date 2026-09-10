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

func Test_MissingExamples_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "missing examples in parameters, request body, responses, and components",
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
          description: test param for object missing example
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
          description: test param for string missing example
          required: false
          schema:
            type: string
        - name: test3
          in: query
          description: test param for integer contains example in schema node
          required: false
          schema:
            type: integer
            example: 1
        - name: ref
          in: query
          description: test param for ref missing example
          required: false
          schema:
            $ref: '#/components/schemas/TestObj'
        - name: ref2
          in: query
          description: test param for ref containing example
          required: false
          schema:
            $ref: '#/components/schemas/TestObj'
          example:
            field: test
      requestBody:
        description: test request body; application/json missing example and application/xml contains multiple examples
        content:
          application/json:
            schema:
              type: object
              required:
                - field
              properties:
                field:
                  type: string
          application/xml:
            schema:
              type: object
              required:
                - field2
              properties:
                field2:
                  type: string
            examples:
              test1:
                value:
                  field2: test1
              test2:
                value:
                  field2: test2
          text/plain:
            schema:
              $ref: '#/components/schemas/TestObj'
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/TestObj'
          text/html:
            schema:
              $ref: '#/components/schemas/TestObj'
      responses:
        '1XX':
          description: test response for ref missing example
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TestObj'
        '200':
          description: test response for string contains example
          content:
            text/plain:
              schema:
                type: string
              example: test
        '3XX':
          description: test response for ref containing example
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TestObj'
        '5XX':
          description: test response for object missing example
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
          description: test response for array of objects missing example
          content:
            application/json:
              schema:
                type: array
                items:
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
components:
  schemas:
    TestObj:
      type: object
      required:
        - field
      properties:
        field:
          type: string
    TestObjWithExample:
      type: object
      required:
        - field
      properties:
        field:
          type: string
      example:
        field: test`,
			},
			wantErrs: []string{
				"validation hint: [line 17:13] generator-missing-examples - missing example for parameter. Consider adding an example",
				"validation hint: [line 32:13] generator-missing-examples - missing example for parameter. Consider adding an example",
				"validation hint: [line 45:13] generator-missing-examples - missing example for parameter. Consider adding an example",
				"validation hint: [line 59:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
				"validation hint: [line 82:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
				"validation hint: [line 85:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
				"validation hint: [line 88:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
				"validation hint: [line 95:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
				"validation hint: [line 108:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
				"validation hint: [line 114:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
				"validation hint: [line 129:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
				"validation hint: [line 144:7] generator-missing-examples - missing example for component. Consider adding an example",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.MissingExamples{}).ID()}))
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
