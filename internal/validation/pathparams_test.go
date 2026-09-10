package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_ValidatePathParams(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "Parameter defined in $ref of path item object but not defined under parameters field of path item object should error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /pizza/{type}/{topping}:
              parameters:
                - name: type
                  in: path
                  required: true
              get:
                operationId: get_pizza
                responses:
                  "200":
                    description: Successful Response
                    content:
                      application/json:
                        schema: {}`),
			},
			wantErrs: []string{
				"validation error: [line 13:7] generator-path-params - `GET` must define parameter `topping` as expected by path `/pizza/{type}/{topping}`",
			},
		},
		{
			name: "Parameter defined in $ref of path item object but not defined under parameters field of path item should error when using different http methods",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /pizza/{type}/{topping}:
              parameters:
                - name: type
                  in: path
                  required: true
              get:
                parameters:
                  - name: topping
                    in: path
                    required: true
                operationId: get_pizza
                responses:
                  "200":
                    description: Success
              post:
                operationId: make_pizza
                responses:
                  "200":
                    description: Success`)},
			wantErrs: []string{
				"validation error: [line 22:7] generator-path-params - `POST` must define parameter `topping` as expected by path `/pizza/{type}/{topping}`",
			},
		},
		{
			name: "path-params paths should not be equivalent, paths must be unique",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /pizza/{cake}/{limes}:
              parameters:
                - name: cake
                  in: path
                  required: true
              get:
                parameters:
                  - name: limes
                    in: path
                    required: true
                responses:
                  "200":
                    description: Success
            /pizza/{minty}/{tape}:
              parameters:
                - name: minty
                  in: path
                  required: true
              get:
                parameters:
                  - name: tape
                    in: path
                    required: true
                responses:
                  "200":
                    description: Success`)},
			wantErrs: []string{
				"validation warn: [line 21:5] generator-path-params - paths `/pizza/{cake}/{limes}` and `/pizza/{minty}/{tape}` must not be equivalent, paths must be unique",
			},
		},
		{
			name: "Parameter declared but not used in path should error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /pizza/{cake}/{cake}:
              parameters:
                - name: cake
                  in: path
                  required: true
              get:
                parameters:
                  - name: limes
                    in: path
                    required: true
                responses:
                  "200":
                    description: Success
            /pizza/{minty}/:
              parameters:
                - name: minty
                  in: path
                  required: true
              get:
                responses:
                  "200":
                    description: Success
        `)},
			wantErrs: []string{
				"validation error: [line 13:7] generator-path-params - parameter `limes` must be used in path `/pizza/{cake}/{cake}`",
			},
		},
		{
			name: "period (.) in param should not result in error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /pizza/{cake}/{cake.id}:
              parameters:
                - name: cake
                  in: path
                  required: true
              get:
                parameters:
                  - name: cake.id
                    in: path
                    required: true
                responses:
                  "200":
                    description: Success`)},
			wantErrs: []string{},
		},
		{
			name: "Valid openapi spec when some path parameters are defined at path level and some are defined at operation level",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /musical/{melody}/{pizza}/:
              parameters:
                - name: melody
                  in: path
                  required: true
              get:
                parameters:
                  - in: path
                    name: pizza
                    required: true
                responses:
                  "200":
                    description: Success`)},
			wantErrs: []string{},
		},
		{
			name: "Valid openapi spec when all parameters are defined at path level and none are defined at operation level",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /musical/{melody}/:
              parameters:
                - name: melody
                  in: path
                  required: true
              get:
                operationId: getSomething
                tags:
                  - tag1
                responses:
                  '200':
                    description: Ok`)},
			wantErrs: []string{},
		},
		{
			name: "A missing param definition at operation level should be flaged",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /musical/{melody}/{pizza}/{cake}:
              parameters:
                - name: melody
                  in: path
                  required: true
              get:
                parameters:
                  - in: path
                    name: pizza
                    required: true
                responses:
                  "200":
                    description: Success
          `)},
			wantErrs: []string{"validation error: [line 13:7] generator-path-params - `GET` must define parameter `cake` as expected by path `/musical/{melody}/{pizza}/{cake}`"},
		},
		{
			name: "A missing param at operation level should be flagged when multiple paths are present",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /musical/{melody}/{pizza}/{cake}:
              parameters:
                - name: melody
                  in: path
                  required: true
              get:
                parameters:
                  - in: path
                    name: pizza
                    required: true
                responses:
                  "200":
                    description: Success
            /dogs/{chicken}/{ember}:
              get:
                parameters:
                  - in: path
                    name: ember
                    required: true
                  - $ref: '#/components/parameters/chicken'
                responses:
                  "200":
                    description: Success
              post:
                parameters:
                  - in: path
                    name: ember
                    required: true
                  - $ref: '#/components/parameters/chicken'
                responses:
                  "200":
                    description: Success
          components:
            parameters:
              chicken:
                in: path
                required: true
                name: chicken`)},
			wantErrs: []string{"validation error: [line 13:7] generator-path-params - `GET` must define parameter `cake` as expected by path `/musical/{melody}/{pizza}/{cake}`"},
		},
		{
			name: "Parameter not defined should error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info:
            title: FastAPI
            version: 0.1.0
            description: stop complaining
          paths:
            /update/{something}:
              post:
                operationId: postSomething
                tags:
                  - tag1
                responses:
                  '200':
                    description: Post ok
              get:
                operationId: getSomething
                summary: get something
                tags:
                  - tag1
                responses:
                  '200':
                    description: Get Ok
          components:
            securitySchemes:
              basicAuth:
                type: http
                scheme: basic
          `)},
			wantErrs: []string{
				"validation error: [line 16:7] generator-path-params - `GET` must define parameter `something` as expected by path `/update/{something}`",
				"validation error: [line 9:7] generator-path-params - `POST` must define parameter `something` as expected by path `/update/{something}`",
			},
		},
		{
			name: "OpenAPI 3.2: Valid spec with additionalOperations (custom HTTP method)",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.2.0
          info:
            title: Test API
            version: 1.0.0
          paths:
            /items/{itemId}:
              parameters:
                - name: itemId
                  in: path
                  required: true
              additionalOperations:
                search:
                  operationId: searchItems
                  responses:
                    "200":
                      description: Success
          `)},
			wantErrs: []string{},
		},
		{
			name: "OpenAPI 3.2: Missing param in additionalOperations should error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.2.0
          info:
            title: Test API
            version: 1.0.0
          paths:
            /items/{itemId}/{categoryId}:
              parameters:
                - name: itemId
                  in: path
                  required: true
              additionalOperations:
                search:
                  operationId: searchItems
                  responses:
                    "200":
                      description: Success
          `)},
			wantErrs: []string{
				"validation error: [line 13:9] generator-path-params - `SEARCH` must define parameter `categoryId` as expected by path `/items/{itemId}/{categoryId}`",
			},
		},
		{
			name: "OpenAPI 3.2: Unused param in additionalOperations should error",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.2.0
          info:
            title: Test API
            version: 1.0.0
          paths:
            /items/{itemId}:
              parameters:
                - name: itemId
                  in: path
                  required: true
              additionalOperations:
                search:
                  parameters:
                    - name: unusedParam
                      in: path
                      required: true
                  operationId: searchItems
                  responses:
                    "200":
                      description: Success
          `)},
			wantErrs: []string{
				"validation error: [line 13:9] generator-path-params - parameter `unusedParam` must be used in path `/items/{itemId}`",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.PathParams{}).ID()}))
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

// TestValidation_PathParams_IgnoresWebhooksAndCallbacks verifies that webhooks and callbacks
// are not validated by the PathParams rule (they should only validate actual paths)
func TestValidation_PathParams_IgnoresWebhooksAndCallbacks(t *testing.T) {
	spec := utils.Dedent(`
    openapi: 3.1.0
    info:
      title: Test API
      version: 1.0.0
    servers:
      - url: http://localhost:8080
    paths:
      /users/{userId}:
        parameters:
          - name: userId
            in: path
            required: true
            schema:
              type: string
        get:
          operationId: getUser
          responses:
            '200':
              description: OK
              content:
                application/json:
                  schema:
                    type: object
          callbacks:
            onData:
              # Callback with same name as webhook - should not cause duplicate error
              '{$request.body#/callbackUrl}':
                post:
                  operationId: onDataCallback
                  requestBody:
                    content:
                      application/json:
                        schema:
                          type: object
                  responses:
                    '200':
                      description: OK
    webhooks:
      # Webhook with same name as callback - should not cause duplicate error
      onData:
        post:
          operationId: onDataWebhook
          requestBody:
            content:
              application/json:
                schema:
                  type: object
          responses:
            '200':
              description: OK
  `)

	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyGeneration,
		validation.WithFilteredRules([]string{(&validation.PathParams{}).ID()}),
	)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(spec), "", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Should have NO errors - webhooks and callbacks should be ignored
	assert.Empty(t, errs, "PathParams rule should not validate webhooks or callbacks")
}
