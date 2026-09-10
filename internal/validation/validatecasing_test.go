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

func TestValidation_ValidateCasing(t *testing.T) {
	tests := []struct {
		name         string
		schema       string
		wantCount    int
		wantContains []string
	}{
		{
			name: "distinct names are not flagged",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    Pet:
      type: object
    PetStore:
      type: object`,
			wantCount: 0,
		},
		{
			name: "names differing only by acronym casing collide",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    ExampleBase64PDFSource:
      type: object
    ExampleBase64PdfSource:
      type: object`,
			wantCount:    1,
			wantContains: []string{"`ExampleBase64PdfSource`", "`ExampleBase64PDFSource`", "only by casing"},
		},
		{
			name: "three-way casing collision reports each later name",
			schema: `
openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
paths: {}
components:
  schemas:
    User:
      type: object
    user:
      type: object
    USER:
      type: object`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended, validation.WithFilteredRules([]string{(&validation.ValidateCasing{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("csharp"))

			got := filterByRule(res.GetValidationErrors(), (&validation.ValidateCasing{}).ID())

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
