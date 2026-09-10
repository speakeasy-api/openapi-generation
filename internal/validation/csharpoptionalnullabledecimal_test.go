package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const csharpScalarDecimalSpec = `
openapi: 3.1.0
info: { "title": "Test", "version": "0.0.1" }
servers: [{ "url": "http://localhost:35123" }]
paths:
  /test:
    post:
      operationId: postTest
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DecimalBody'
      responses:
        '200':
          description: OK
components:
  schemas:
    DecimalBody:
      type: object
      required:
        - requiredDecimal
      properties:
        scalarDecimal:
          type: [number, "null"]
          format: decimal
        decimalStr:
          type: [string, "null"]
          format: decimal
        decimalList:
          type: [array, "null"]
          items:
            type: number
            format: decimal
        plainNum:
          type: [number, "null"]
        requiredDecimal:
          type: [number, "null"]
          format: decimal
`

const csharpInlineScalarDecimalSpec = `
openapi: 3.1.0
info: { "title": "Test", "version": "0.0.1" }
servers: [{ "url": "http://localhost:35123" }]
paths:
  /test:
    post:
      operationId: postTest
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                inlineScalarDecimal:
                  type: [number, "null"]
                  format: decimal
      responses:
        '200':
          description: OK
`

func csharpConfig(presenceAware bool) *config.Configuration {
	return &config.Configuration{
		Languages: map[string]config.LanguageConfig{
			"csharp": {
				Cfg: map[string]any{
					"presenceAwareJsonSerialization": presenceAware,
				},
			},
		},
	}
}

func Test_CsharpOptionalNullableDecimal(t *testing.T) {
	ruleID := (&validation.CsharpOptionalNullableDecimal{}).ID()
	wantWarning := "C#: 'DecimalBody.scalarDecimal' (format: decimal) cannot be both optional and nullable without precision loss: treating as plain nullable 'decimal?' instead. Consider using a string-encoded decimal: 'type: string, format: decimal' to preserve both precision and presence awareness."

	tests := []struct {
		name          string
		target        string
		presenceAware bool
		wantWarn      bool
	}{
		{name: "csharp + presence-aware warns exactly once for scalar decimal", target: "csharp", presenceAware: true, wantWarn: true},
		{name: "csharp + flag off no warning", target: "csharp", presenceAware: false, wantWarn: false},
		{name: "non-csharp target no warning", target: "typescript", presenceAware: true, wantWarn: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			cfg := csharpConfig(tt.presenceAware)
			v, err := validation.NewValidator(cfg, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{ruleID}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(utils.Dedent(csharpScalarDecimalSpec)), "", types.NewTargetFromTemplate(tt.target))
			got := filterByRule(res.GetValidationErrors(), ruleID)

			if tt.wantWarn {
				require.Len(t, got, 1, "expected exactly one warning, got %v", got)
				assert.Contains(t, got[0], wantWarning)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

func Test_CsharpOptionalNullableDecimal_InlineSchema(t *testing.T) {
	ruleID := (&validation.CsharpOptionalNullableDecimal{}).ID()

	t.Setenv("SPEAKEASY_DEBUG", "true")

	cfg := csharpConfig(true)
	v, err := validation.NewValidator(cfg, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{ruleID}))
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(utils.Dedent(csharpInlineScalarDecimalSpec)), "", types.NewTargetFromTemplate("csharp"))
	got := filterByRule(res.GetValidationErrors(), ruleID)

	require.Len(t, got, 1, "expected inline scalar decimal to warn, got %v", got)
	assert.Contains(t, got[0], ".inlineScalarDecimal")
}
