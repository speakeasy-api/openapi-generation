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

func TestValidation_ValidateContentType_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "request body is valid multipart with explicit object type",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: object
              properties:
                test:
                  type: string
                  format: binary
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
		{
			name: "request body is valid multipart with implicit object type (properties)",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema:
              properties:
                test:
                  type: string
                  format: binary
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
		{
			name: "request body is valid multipart with implicit object type (additionalProperties)",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema:
              additionalProperties: true
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
		{
			name: "request body is valid multipart with implicit object type (patternProperties)",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema:
              patternProperties:
                "^S_":
                  type: string
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
		{
			name: "no request body",
			args: args{
				schema: `
openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
		{
			name: "request body is not multipart",
			args: args{
				schema: `
openapi: 3.1.0
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
              properties:
                test:
                  type: string
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request`[1:],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{
					Generation: config.Generation{},
				},
				validation.RulesetSpeakeasyGeneration,
				validation.WithFilteredRules([]string{(&validation.ValidateContentType{}).ID()}),
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			assert.Empty(t, res.GetValidationErrors())
		})
	}
}

func TestValidation_ValidateContentType_Error(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "request body is not an explicit object type",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: string
              format: binary
      responses:
        '200':
          description: OK`[1:],
			},
			wantErrs: []string{
				"validation error: [line 13:13] generator-validate-content-type - multipart schema must be an `object`",
			},
		},
		{
			name: "request body is not an implicitly an object type",
			args: args{
				schema: `
openapi: 3.1.0
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
          multipart/form-data:
            schema: {}
      responses:
        '200':
          description: OK`[1:],
			},
			wantErrs: []string{
				"validation error: [line 13:13] generator-validate-content-type - multipart schema must be an `object`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateContentType{}).ID()}))
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

func TestValidation_ValidateMultipartRequestBody(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "multipart request body with valid object schema",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: object
              properties:
                test:
                  type: string
              required:
                - test
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{},
		},
		{
			name: "multipart request body with no object schema",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: string
              format: binary
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string`,
			},
			wantErrs: []string{
				"validation error: [line 13:13] generator-validate-content-type - multipart schema must be an `object`",
			},
		},
		{
			name: "invalid reference schema",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              $ref: "#/components/schemas/test"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string
components:
  schemas:
    test:
      type: string`,
			},
			wantErrs: []string{
				"validation error: [line 13:13] generator-validate-content-type - multipart schema must be an `object`",
			},
		},
		{
			name: "valid reference schema",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              $ref: "#/components/schemas/test"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string
components:
  schemas:
    test:
      type: object
      properties:
        test:
          type: string`,
			},
			wantErrs: []string{},
		},
		{
			name: "valid allOf schema with inferred object type",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              allOf:
                - type: object
                  properties:
                    name:
                      type: string
                - type: object
                  properties:
                    file:
                      type: string
                      format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{},
		},
		{
			name: "valid array schema with object items",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: array
              items:
                type: object
                properties:
                  file:
                    type: string
                    format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{},
		},
		{
			name: "invalid array schema with non-object items",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: array
              items:
                type: string
                format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{
				"validation error: [line 16:17] generator-validate-content-type - multipart schema must be an `object`",
			},
		},
		{
			name: "valid allOf schema with inferred object variants",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              allOf:
                - properties:
                    name:
                      type: string
                - properties:
                    file:
                      type: string
                      format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{},
		},
		{
			name: "valid array with object items schema with inferred object variants",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: array
              items:
                allOf:
                  - properties:
                      file:
                        type: string
                        format: binary
                  - properties:
                      name:
                        type: string
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{},
		},
		{
			name: "is valid with empty object",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              type: object
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{},
		},
		{
			name: "is invalid with oneOf",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              oneOf:
                - type: object
                  properties:
                    name:
                      type: string
                - type: object
                  properties:
                    file:
                      type: string
                      format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{
				"validation warn: [line 13:13] generator-validate-content-type - multipart schema must not contain `oneOf`",
			},
		},
		{
			name: "is invalid with anyOf",
			args: args{
				schema: `openapi: 3.1.0
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
          multipart/form-data:
            schema:
              anyOf:
                - type: object
                  properties:
                    name:
                      type: string
                - type: object
                  properties:
                    file:
                      type: string
                      format: binary
      responses:
        "200":
          description: OK`,
			},
			wantErrs: []string{
				"validation warn: [line 13:13] generator-validate-content-type - multipart schema must not contain `anyOf`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateContentType{}).ID()}))
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
