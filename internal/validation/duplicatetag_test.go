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

func Test_DuplicateTag_Errors(t *testing.T) {
	type args struct {
		schema       string
		sdkClassName string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "duplicate tags in global tag definition",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
tags:
  - name: TestSomething
  - name: testSomething
paths:
  /test:
    get:
      responses:
        '200':
          description: OK`,
				sdkClassName: "SDK",
			},
			wantErrs: []string{
				"validation error: [line 9:5] generator-duplicate-tag - `testSomething` (`TestSomething`) will collide with `TestSomething` (`TestSomething`) [line `8`] when converted to class/field name",
			},
		},
		{
			name: "duplicate identical tags in single operation",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    post:
      tags:
        - testSomething
        - testSomething
      responses:
        '200':
          description: OK`,
				sdkClassName: "SDK",
			},
			wantErrs: []string{
				"validation warn: [line 12:11] generator-duplicate-tag - duplicate tag `testSomething` in operation",
			},
		},
		{
			name: "duplicate  tags in single operation",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    post:
      tags:
        - TestSomething
        - testSomething
      responses:
        '200':
          description: OK`,
				sdkClassName: "SDK",
			},
			wantErrs: []string{
				"validation error: [line 12:11] generator-duplicate-tag - `testSomething` (`TestSomething`) will collide with `TestSomething` (`TestSomething`) [line `11`] when converted to class/field name",
			},
		},
		{
			name: "duplicate tags in per operations list",
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
      tags:
        - TestSomething
      responses:
        '200':
          description: OK
    post:
      tags:
        - testSomething
      responses:
        '200':
          description: OK`,
				sdkClassName: "SDK",
			},
			wantErrs: []string{
				"validation error: [line 17:11] generator-duplicate-tag - `testSomething` (`TestSomething`) will collide with `TestSomething` (`TestSomething`) [line `11`] when converted to class/field name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{
				Generation: config.Generation{
					SDKClassName: tt.args.sdkClassName,
				},
			}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateTag{}).ID()}))
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
