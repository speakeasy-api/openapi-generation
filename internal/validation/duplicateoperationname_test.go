package validation_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DuplicateOperationName_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid operation ids",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                operationId: TestSomething
                responses:
                  '200':
                    description: OK
              post:
                operationId: testSomethingElse
                responses:
                  '200':
                    description: OK
        `),
			},
		},
		{
			name: "operation id don't collide with missing operation id",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                responses:
                  '200':
                    description: OK
              post:
                operationId: postTest
                responses:
                  '200':
                    description: OK
        `),
			},
		},
		{
			name: "global overridden operation ids don't collide in different tags",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          x-speakeasy-name-override:
            - operationId: ^get.*
              methodNameOverride: get
          tags:
            - name: dog
            - name: cat
          paths:
            /dog:
              get:
                operationId: getDog
                tags:
                  - dog
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                operationId: getCat
                tags:
                  - cat
                responses:
                  '200':
                    description: OK
        `),
			},
		},
		{
			name: "local overridden operation ids don't collide in different tags",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          tags:
            - name: dog
            - name: cat
          paths:
            /dog:
              get:
                operationId: getDog
                tags:
                  - dog
                x-speakeasy-name-override: get
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                operationId: getCat
                tags:
                  - cat
                x-speakeasy-name-override: get
                responses:
                  '200':
                    description: OK
        `),
			},
		},
		{
			name: "callback operations with duplicate method names are ignored",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              post:
                operationId: createTest
                responses:
                  '200':
                    description: OK
                callbacks:
                  onEvent:
                    '{$request.body#/callbackUrl}':
                      post:
                        responses:
                          '200':
                            description: OK
                  onOtherEvent:
                    '{$request.body#/callbackUrl}':
                      post:
                        responses:
                          '200':
                            description: OK
        `),
			},
		},
		{
			name: "webhook operations with duplicate method names are ignored",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              post:
                operationId: createTest
                responses:
                  '200':
                    description: OK
          webhooks:
            onEvent:
              post:
                responses:
                  '200':
                    description: OK
            onOtherEvent:
              post:
                responses:
                  '200':
                    description: OK
        `),
			},
		},
		{
			name: "duplicate tags are tolerated",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          tags:
            - name: foo
            - name: bar
          paths:
            /test:
              get:
                operationId: TestSomething
                tags:
                  - foo
                  - foo
                  - bar
                responses:
                  '200':
                    description: OK
        `),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateOperationName{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func Test_DuplicateOperationName_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "plain duplicate operation IDs",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                operationId: TestSomething
                responses:
                  '200':
                    description: OK
              post:
                operationId: TestSomething
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 12:7] generator-duplicate-operation-name - method name `sdk.testSomething()` will collide with operation `get /test` [line `7`], try using `x-speakeasy-name-override`",
				"validation error: [line 12:20] validation-operation-id-unique - the `post` operation at path `/test` contains a duplicate operationId `TestSomething`",
			},
		},
		{
			name: "duplicate operation IDs due to snakecase vs joined up",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                operationId: foo_bar
                responses:
                  '200':
                    description: OK
              post:
                operationId: foobar
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 12:7] generator-duplicate-operation-name - method name `sdk.foobar()` will collide with operation `get /test` [line `7`], try using `x-speakeasy-name-override`",
			},
		},
		{
			name: "duplicate operation ids due to case sensitivity",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                operationId: TestSomething
                responses:
                  '200':
                    description: OK
              post:
                operationId: testSomething
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 12:7] generator-duplicate-operation-name - method name `sdk.testSomething()` will collide with operation `get /test` [line `7`], try using `x-speakeasy-name-override`",
			},
		},
		{
			name: "operation id collides with missing operation id",
			args: args{
				schema: utils.Dedent(`
              openapi: 3.1.0
              info: { "title": "Test", "version": "0.0.1" }
              servers: [{ "url": "http://localhost:35123" }]
              paths:
                /test:
                  get:
                    responses:
                      '200':
                        description: OK
                  post:
                    operationId: getTest
                    responses:
                      '200':
                        description: OK
    `),
			},
			wantErrs: []string{
				"validation error: [line 11:7] generator-duplicate-operation-name - method name `sdk.getTest()` will collide with operation `get /test` [line `7`], try using `x-speakeasy-name-override`",
			},
		},
		{
			name: "different operations without ids collide with each other",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test_something:
              get:
                responses:
                  '200':
                    description: OK
            /test-something:
              get:
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 12:7] generator-duplicate-operation-name - method name `sdk.getTestSomething()` will collide with operation `get /test_something` [line `7`], try adding `operationId`",
			},
		},
		{
			name: "global overridden operation ids collide with each other",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          x-speakeasy-name-override:
            - operationId: ^get.*
              methodNameOverride: get
          paths:
            /dog:
              get:
                operationId: getDog
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                operationId: getCat
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 16:7] generator-duplicate-operation-name - method name `sdk.get()` will collide with operation `get /dog` [line `10`], try using `x-speakeasy-group`",
			},
		},
		{
			name: "local overridden operation ids collide with each other",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /dog:
              get:
                operationId: getDog
                x-speakeasy-name-override: get
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                operationId: getCat
                x-speakeasy-name-override: get
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 15:34] generator-duplicate-operation-name - method name `sdk.get()` will collide with operation `get /dog` [line `8`], try using `x-speakeasy-group`",
			},
		},
		{
			name: "diff opID, same name override, same group",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /dog:
              get:
                operationId: getDog
                x-speakeasy-name-override: get
                x-speakeasy-group: dog
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                operationId: getCat
                x-speakeasy-name-override: get
                x-speakeasy-group: dog
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 16:34] generator-duplicate-operation-name - method name `sdk.dog.get()` will collide with operation `get /dog` [line `8`], try adjusting `x-speakeasy-name-override`",
			},
		},
		{
			name: "no opID, same name override, group matches tag",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          tags:
            - name: dog
          paths:
            /dog:
              get:
                x-speakeasy-name-override: get
                x-speakeasy-group: dog
                responses:
                  '200':
                    description: OK
            /cat:
              get:
                x-speakeasy-name-override: get
                tags:
                  - dog
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 16:34] generator-duplicate-operation-name - method name `sdk.dog.get()` will collide with operation `get /dog` [line `9`], try adding `operationId`",
			},
		},
		{
			name: "test operation IDs with whitespace ",
			args: args{
				schema: utils.Dedent(`
          openapi: 3.1.0
          info: { "title": "Test", "version": "0.0.1" }
          servers: [{ "url": "http://localhost:35123" }]
          paths:
            /test:
              get:
                operationId: " Test Something "
                responses:
                  '200':
                    description: OK
            /test-2:
              get:
                operationId: "Test Something"
                responses:
                  '200':
                    description: OK
        `),
			},
			wantErrs: []string{
				"validation error: [line 13:7] generator-duplicate-operation-name - method name `sdk.testSomething()` will collide with operation `get /test` [line `7`], try using `x-speakeasy-name-override`",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateOperationName{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("typescript"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.Equal(t, tt.wantErrs, errStrs)
		})
	}
}

func Test_DuplicateOperationName_ExternalRefOperation(t *testing.T) {

	// If this is windows this test will fail because of paths used so early exit
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows")
	}

	testCases := []struct {
		name         string
		baseSpec     string
		externalYaml string
		expectError  bool
	}{
		// --- Duplicate name override ---
		{
			name: "duplicate name override",
			baseSpec: utils.Dedent(`
                  openapi: 3.1.0
                  info:
                    title: Test
                    version: 0.0.1
                  servers:
                    - url: http://localhost:35123
                  paths:
                    /a:
                      $ref: "./test_data/external.yaml#/a"
                    /b:
                      $ref: "./test_data/external.yaml#/b"
              `),
			externalYaml: utils.Dedent(`
                  a:
                    get:
                      operationId: op-a
                      x-speakeasy-group: group-a
                      x-speakeasy-name-override: asdf
                      responses:
                        '200':
                          description: OK
                  b:
                    get:
                      operationId: op-b
                      x-speakeasy-group: group-a
                      x-speakeasy-name-override: asdf
                      responses:
                        '200':
                          description: OK
              `),
			expectError: true,
		},
		// --- Duplicate operation ID ---
		{
			name: "duplicate operation ID",
			baseSpec: utils.Dedent(`
                  openapi: 3.1.0
                  info:
                    title: Test
                    version: 0.0.1
                  servers:
                    - url: http://localhost:35123
                  paths:
                    /a:
                      $ref: "./test_data/external.yaml#/a"
                    /b:
                      $ref: "./test_data/external.yaml#/b"
              `),
			externalYaml: utils.Dedent(`
                  a:
                    get:
                      operationId: op-a
                      responses:
                        '200':
                          description: OK
                  b:
                    get:
                      operationId: op-a
                      responses:
                        '200':
                          description: OK
              `),
			expectError: true,
		},
		// --- Unique operation id ---
		{
			name: "unique operation id",
			baseSpec: utils.Dedent(`
                  openapi: 3.1.0
                  info:
                    title: Test
                    version: 0.0.1
                  servers:
                    - url: http://localhost:35123
                  paths:
                    /a:
                      $ref: "./test_data/external.yaml#/a"
                    /b:
                      $ref: "./test_data/external.yaml#/b"
              `),
			externalYaml: utils.Dedent(`
                  a:
                    get:
                      operationId: op-a
                      x-speakeasy-group: group-a
                      x-speakeasy-name-override: asdf1
                      responses:
                        '200':
                          description: OK
                  b:
                    get:
                      operationId: op-b
                      x-speakeasy-group: group-a
                      x-speakeasy-name-override: asdf2
                      responses:
                        '200':
                          description: OK
              `),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.baseSpec = utils.Dedent(tc.baseSpec)
			tc.externalYaml = utils.Dedent(tc.externalYaml)

			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyGeneration,
				validation.WithFilteredRules([]string{(&validation.DuplicateOperationName{}).ID()}),
			)
			require.NoError(t, err)

			// Write the externalYaml to ./test_data/external.yaml
			err = os.MkdirAll("./test_data", 0755)
			require.NoError(t, err)
			err = os.WriteFile("./test_data/external.yaml", []byte(tc.externalYaml), 0644)
			require.NoError(t, err)

			res := validateSpec(v,
				context.Background(),
				[]byte(tc.baseSpec),
				"",
				types.NewTargetFromTemplate("go"),
			)
			errs := res.GetValidationErrors()

			if tc.expectError {
				assert.NotEmpty(t, errs)
				if len(errs) > 0 {
					errStr := errs[0].Error()
					assert.Contains(t, errStr, "duplicate-operation-name")
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
