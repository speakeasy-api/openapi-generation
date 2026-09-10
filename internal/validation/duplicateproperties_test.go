package validation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DuplicateProperties_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "properties can be renamed to avoid collisions",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    Guid:
      type: string
      format: uuid
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  guid:
                    type: string
                  Guid:
                    x-speakeasy-name-override: guids
                    type: array
                    items:
                      type: string
                  g-uid:
                    x-speakeasy-name-override: uuid
                    $ref: '#/components/schemas/Guid'`,
			},
		},
		{
			name: "sibling name override with namespace does not collide",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    foo_Pet:
      x-speakeasy-name-override: Pet
      x-speakeasy-model-namespace: foo
      type: object
      properties:
        name:
          type: string
    bar_Pet:
      x-speakeasy-name-override: Pet
      x-speakeasy-model-namespace: bar
      type: object
      properties:
        name:
          type: string
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  fooPet:
                    x-speakeasy-name-override: Pet
                    $ref: '#/components/schemas/foo_Pet'
                  barPet:
                    x-speakeasy-name-override: Pet
                    $ref: '#/components/schemas/bar_Pet'`,
			},
		},
		{
			name: "namespaced refs with same name override do not collide",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    pet:
      type: object
      properties:
        name:
          type: string
    foo_Pet:
      x-speakeasy-name-override: Pet
      x-speakeasy-model-namespace: foo
      type: object
      properties:
        name:
          type: string
    bar_Pet:
      x-speakeasy-name-override: Pet
      x-speakeasy-model-namespace: bar
      type: object
      properties:
        name:
          type: string
    NamespaceConflictTest:
      type: object
      properties:
        Pet:
          $ref: '#/components/schemas/pet'
        fooPet:
          $ref: '#/components/schemas/foo_Pet'
        barPet:
          $ref: '#/components/schemas/bar_Pet'`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateProperties{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			assert.Empty(t, errs)
		})
	}
}

func Test_DuplicateProperties_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	type wantErrs struct {
		nameOverrideFeb2026False []string
		nameOverrideFeb2026True  []string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs wantErrs
	}{
		{
			name: "inline name override causes collision",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    NestedConflict:
      type: object
      properties:
        conflict:
          type: string
        renamed:
          x-speakeasy-name-override: conflict
          type: string
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  conflict:
                    type: string
                  renamed:
                    x-speakeasy-name-override: conflict
                    type: string`,
			},
			wantErrs: wantErrs{
				nameOverrideFeb2026False: []string{
					"validation warn: [line 14:9] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `12`] when converted to field name. Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition.",
					"validation warn: [line 30:19] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `28`] when converted to field name. Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition.",
				},
				nameOverrideFeb2026True: []string{
					"validation error: [line 14:9] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `12`] when converted to field name.",
					"validation error: [line 30:19] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `28`] when converted to field name.",
				},
			},
		},
		{
			name: "ref component name override causes collision",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    RenamedSchema:
      x-speakeasy-name-override: conflict
      type: string
    NestedConflict:
      type: object
      properties:
        conflict:
          type: string
        renamed:
          $ref: '#/components/schemas/RenamedSchema'
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  conflict:
                    type: string
                  renamed:
                    $ref: '#/components/schemas/RenamedSchema'
                  nested:
                    $ref: '#/components/schemas/NestedConflict'`,
			},
			wantErrs: wantErrs{
				nameOverrideFeb2026False: []string{
					"validation warn: [line 17:9] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `15`] when converted to field name. Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition.",
					"validation warn: [line 32:19] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `30`] when converted to field name. Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition.",
				},
				nameOverrideFeb2026True: nil, // flag prevents component-level override from bleeding through $ref
			},
		},
		{
			name: "sibling ref name override causes collision",
			args: args{
				schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
components:
  schemas:
    SomeSchema:
      type: string
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  conflict:
                    type: string
                  renamed:
                    x-speakeasy-name-override: conflict
                    $ref: '#/components/schemas/SomeSchema'`,
			},
			wantErrs: wantErrs{
				nameOverrideFeb2026False: []string{
					"validation warn: [line 24:19] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `22`] when converted to field name. Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition.",
				},
				nameOverrideFeb2026True: []string{
					"validation error: [line 24:19] generator-duplicate-properties - `conflict` (`Conflict`) will collide with `conflict` (`Conflict`) [line `22`] when converted to field name.",
				},
			},
		},
		{
			name: "invalid properties",
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
          content:
            application/json:
              schema:
                type: object
                properties:
                  test!something:
                    type: string
                  test^something:
                    type: string
                  '':
                    type: string`,
			},
			wantErrs: func() wantErrs {
				errs := []string{
					"validation error: [line 20:19] generator-duplicate-properties - `test^something` (`TestSomething`) will collide with `test!something` (`TestSomething`) [line `18`] when converted to field name.",
					"validation error: [line 22:19] generator-duplicate-properties - empty property name",
				}
				// no x-speakeasy-name-override so NameOverrideFeb2026 value does not apply
				return wantErrs{nameOverrideFeb2026False: errs, nameOverrideFeb2026True: errs}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, nameOverrideFeb2026 := range []bool{false, true} {
				want := tt.wantErrs.nameOverrideFeb2026False
				if nameOverrideFeb2026 {
					want = tt.wantErrs.nameOverrideFeb2026True
				}
				t.Run(fmt.Sprintf("nameOverrideFeb2026=%v", nameOverrideFeb2026), func(t *testing.T) {
					t.Setenv("SPEAKEASY_DEBUG", "true")
					cfg := &config.Configuration{}
					cfg.Generation.Fixes = &config.Fixes{
						NameOverrideFeb2026: nameOverrideFeb2026,
					}
					v, err := validation.NewValidator(cfg, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.DuplicateProperties{}).ID()}))
					require.NoError(t, err)

					res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
					errs := res.GetValidationErrors()
					errStrs := make([]string, len(errs))
					for i, err := range errs {
						errStrs[i] = err.Error()
					}

					assert.ElementsMatch(t, want, errStrs)
				})
			}
		})
	}
}
