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

func TestValidation_DuplicateComponentSchemas(t *testing.T) {
	tests := []struct {
		name         string
		schema       string
		wantCount    int
		wantContains []string
	}{
		{
			name: "structurally distinct schemas are not flagged",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Truncation:
      type: object
      properties:
        kind:
          type: string
    Pagination:
      type: object
      properties:
        cursor:
          type: string`,
			wantCount: 0,
		},
		{
			name: "structurally identical named schemas are flagged once per group",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Truncation:
      type: object
      properties:
        kind:
          type: string
        size:
          type: integer
    OpenAIResponsesTruncation:
      type: object
      properties:
        kind:
          type: string
        size:
          type: integer`,
			wantCount:    1,
			wantContains: []string{"structurally identical", "Truncation", "OpenAIResponsesTruncation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.DuplicateComponentSchemas{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))

			got := filterByRule(res.GetValidationErrors(), (&validation.DuplicateComponentSchemas{}).ID())

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
