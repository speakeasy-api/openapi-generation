package templates_test

import (
	"regexp"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRubyModuleValidationRegex(t *testing.T) {
	t.Parallel()

	fields, err := templates.GetLanguageConfigFields(types.NewTargetFromTemplate("ruby"), false)
	require.NoError(t, err)

	var pattern string
	for _, field := range fields {
		if field.Name == "module" {
			require.NotNil(t, field.ValidationRegex)
			pattern = *field.ValidationRegex
		}
	}
	require.NotEmpty(t, pattern)

	regex := regexp.MustCompile(pattern)

	testCases := map[string]bool{
		"OpenApiSdk":  true,
		"AcmeRubySDK": true,
		"OpenAPI":     true,
		"SDK":         true,
		"Acme2":       true,
		"openApiSdk":  false,
		"Open_Api":    false,
		"Open::Api":   false,
		"Open-Api":    false,
		"2Acme":       false,
		"":            false,
	}

	for name, valid := range testCases {
		assert.Equal(t, valid, regex.MatchString(name), "module %q", name)
	}
}
