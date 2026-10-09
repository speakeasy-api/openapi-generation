package validation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	generrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_PatternErrorMessage(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, schema, extensionRewrite, want string
		severity                             generrors.Severity
	}{
		{name: "valid", schema: `pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: Use lowercase letters.`},
		{name: "empty pattern", schema: `pattern: ''
      x-speakeasy-pattern-error-message: Use lowercase letters.`},
		{name: "absent pattern", schema: `x-speakeasy-pattern-error-message: Use lowercase letters.`, want: "has no pattern on this schema", severity: generrors.SeverityWarn},
		{name: "empty message", schema: `pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: ''`, want: "must be a non-empty string", severity: generrors.SeverityError},
		{name: "whitespace message", schema: `pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: '  '`, want: "must be a non-empty string", severity: generrors.SeverityError},
		{name: "nonstring message", schema: `pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: 42`, want: "must be a non-empty string", severity: generrors.SeverityError},
		{name: "invalid without pattern", schema: `x-speakeasy-pattern-error-message: false`, want: "must be a non-empty string", severity: generrors.SeverityError},
		{name: "rewritten extension", extensionRewrite: "x-speakeasy-extension-rewrite:\n  x-speakeasy-pattern-error-message: x-pattern-message", schema: `pattern: '^[a-z]+$'
      x-pattern-message: null`, want: "x-pattern-message must be a non-empty string", severity: generrors.SeverityError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			spec := fmt.Sprintf(`openapi: 3.1.0
info: {title: Pattern messages, version: 1.0.0}
paths: {}
%s
components:
  schemas:
    Code:
      type: string
      %s
`, tc.extensionRewrite, tc.schema)
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateExtensions{}).ID()}))
			require.NoError(t, err)
			result := validateSpec(v, context.Background(), []byte(spec), "", types.NewTargetFromTemplate("cli"))
			errs := result.GetValidationErrors()
			if tc.want == "" {
				assert.Empty(t, errs)
				return
			}
			require.Len(t, errs, 1)
			assert.Contains(t, errs[0].Error(), tc.want)
			var diagnostic *generrors.ValidationError
			require.ErrorAs(t, errs[0], &diagnostic)
			assert.Equal(t, tc.severity, diagnostic.Severity)
			assert.Positive(t, diagnostic.Node.Line)
		})
	}
}
