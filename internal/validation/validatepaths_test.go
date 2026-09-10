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

func Test_ValidatePaths_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "simple path with unencoded characters",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /^test/something{":
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 9:5] generator-validate-paths - path contains unencoded characters [ `^ { \"` ]"},
		},
		{
			name: "complex path with unencoded characters",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /^test/something"?value=}#more{:
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 9:5] generator-validate-paths - path contains unencoded characters [ `^ \" } {` ]"},
		},
		{
			name: "simple path with template contains unencoded characters",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /^test/{something}":
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 9:5] generator-validate-paths - path contains unencoded characters [ `^ \"` ]"},
		},
		{
			name: "complex path with templating contains unencoded characters",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /^test/{something_something}"?value=}#more{:
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{"validation error: [line 9:5] generator-validate-paths - path contains unencoded characters [ `^ \" } {` ]"},
		},
		{
			name: "path with extensions should not fail",
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
      responses:
        '200':
          description: OK
  x-internal: true
  x-custom-data:
    something: value
`,
			},
			wantErrs: []string{},
		},
		{
			name: "path with valid characters should not fail",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test/valid-path_123~something:
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{},
		},
		{
			name: "path with percent encoding should not fail",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test/path%20with%20spaces:
    get:
      responses:
        '200':
          description: OK
`,
			},
			wantErrs: []string{},
		},
		{
			name: "referenced path item with invalid characters should fail",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /^test/invalid{:
    $ref: '#/components/pathItems/TestPath'
components:
  pathItems:
    TestPath:
      get:
        responses:
          '200':
            description: OK
`,
			},
			wantErrs: []string{"validation error: [line 13:7] generator-validate-paths - path contains unencoded characters [ `^ {` ]"},
		},
		{
			name: "referenced path item with valid path should not fail",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test/valid:
    $ref: '#/components/pathItems/TestPath'
components:
  pathItems:
    TestPath:
      get:
        responses:
          '200':
            description: OK
`,
			},
			wantErrs: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidatePaths{}).ID()}))
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
