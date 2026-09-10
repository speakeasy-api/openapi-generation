package validation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_SuggestDiscriminator(t *testing.T) {
	tests := []struct {
		name         string
		schema       string
		wantCount    int
		wantContains []string
	}{
		{
			name: "object union distinguished by const type suggests a discriminator",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    OutputText:
      type: object
      required: [type]
      properties:
        type:
          type: string
          enum: [text]
        text:
          type: string
    OutputImage:
      type: object
      required: [type]
      properties:
        type:
          type: string
          enum: [image]
        url:
          type: string
    OutputItems:
      oneOf:
        - $ref: '#/components/schemas/OutputText'
        - $ref: '#/components/schemas/OutputImage'`,
			wantCount:    1,
			wantContains: []string{"distinct `type` values", "propertyName: type"},
		},
		{
			name: "union that already has a discriminator is not flagged",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    OutputText:
      type: object
      properties:
        type:
          type: string
          enum: [text]
    OutputImage:
      type: object
      properties:
        type:
          type: string
          enum: [image]
    OutputItems:
      oneOf:
        - $ref: '#/components/schemas/OutputText'
        - $ref: '#/components/schemas/OutputImage'
      discriminator:
        propertyName: type`,
			wantCount: 0,
		},
		{
			name: "multiple qualifying properties pick the lexicographically first deterministically",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Alpha:
      type: object
      required: [kind, zone]
      properties:
        kind:
          type: string
          enum: [a]
        zone:
          type: string
          enum: [x]
    Beta:
      type: object
      required: [kind, zone]
      properties:
        kind:
          type: string
          enum: [b]
        zone:
          type: string
          enum: [y]
    AlphaOrBeta:
      oneOf:
        - $ref: '#/components/schemas/Alpha'
        - $ref: '#/components/schemas/Beta'`,
			wantCount:    1,
			wantContains: []string{"propertyName: kind"},
		},
		{
			name: "empty-string const is treated as a valid discriminator value",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Unset:
      type: object
      required: [kind]
      properties:
        kind:
          const: ""
    Active:
      type: object
      required: [kind]
      properties:
        kind:
          const: active
    KindUnion:
      oneOf:
        - $ref: '#/components/schemas/Unset'
        - $ref: '#/components/schemas/Active'`,
			wantCount:    1,
			wantContains: []string{"propertyName: kind"},
		},
		{
			name: "const property not required on one member is flagged with a suggestion to mark it required",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Dog:
      type: object
      properties:
        kind:
          type: string
          const: dog
        name:
          type: string
    Cat:
      type: object
      required: [kind]
      properties:
        kind:
          type: string
          const: cat
        name:
          type: string
    Pet:
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'`,
			wantCount:    1,
			wantContains: []string{"propertyName: kind", "mark `kind` as required"},
		},
		{
			name: "union without a shared constant property is not flagged",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Cat:
      type: object
      properties:
        meow:
          type: string
    Dog:
      type: object
      properties:
        bark:
          type: string
    Animal:
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'`,
			wantCount: 0,
		},
		{
			name: "scalar union is not flagged",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    StringOrNumber:
      oneOf:
        - type: string
        - type: number`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.SuggestDiscriminator{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))

			got := filterByRule(res.GetValidationErrors(), (&validation.SuggestDiscriminator{}).ID())

			assert.Len(t, got, tt.wantCount, "errors: %v", got)
			for _, frag := range tt.wantContains {
				found := false
				for _, g := range got {
					if strings.Contains(g, frag) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected an error containing %q, got %v", frag, got)
			}
		})
	}
}
