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

func Test_DuplicateModelNamespace_Success(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "no model namespaces",
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
components:
  schemas:
    Foo:
      type: object
    Bar:
      type: object
`,
		},
		{
			name: "distinct model namespaces",
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
components:
  schemas:
    Foo:
      x-speakeasy-model-namespace: alpha
      type: object
    Bar:
      x-speakeasy-model-namespace: beta
      type: object
`,
		},
		{
			name: "same model namespace on multiple schemas is fine",
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
components:
  schemas:
    Foo:
      x-speakeasy-model-namespace: shared_ns
      type: object
    Bar:
      x-speakeasy-model-namespace: shared_ns
      type: object
`,
		},
		{
			name: "nested namespace does not collide with flat namespace",
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
components:
  schemas:
    Foo:
      x-speakeasy-model-namespace: foo/bar
      type: object
    Bar:
      x-speakeasy-model-namespace: foobar
      type: object
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateModelNamespace{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func Test_DuplicateModelNamespace_Errors(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		wantErrs []string
	}{
		{
			name: "colliding namespaces due to case sensitivity",
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
components:
  schemas:
    Foo:
      x-speakeasy-model-namespace: federation_lookup
      type: object
    Bar:
      x-speakeasy-model-namespace: federationLookup
      type: object
`,
			wantErrs: []string{
				"validation error: [line 18:5] generator-duplicate-model-namespace - model namespace `federationLookup` (`federationlookup`) will collide with `federation_lookup` (`federationlookup`) [line `15`] when converted to folder/package name",
			},
		},
		{
			name: "colliding namespaces due to snake_case vs camelCase",
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
components:
  schemas:
    Foo:
      x-speakeasy-model-namespace: my_models
      type: object
    Bar:
      x-speakeasy-model-namespace: MyModels
      type: object
`,
			wantErrs: []string{
				"validation error: [line 18:5] generator-duplicate-model-namespace - model namespace `MyModels` (`mymodels`) will collide with `my_models` (`mymodels`) [line `15`] when converted to folder/package name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateModelNamespace{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
