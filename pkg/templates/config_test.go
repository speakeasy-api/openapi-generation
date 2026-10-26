package templates_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RegExp.source emits "/" as the Unicode escape below, which the CLI rewrites before compiling the pattern with Go's regexp.
const escapedSlash = "\x5cu002f"

func TestConfigFieldValidationRegexes(t *testing.T) {
	t.Parallel()

	for _, name := range templates.GetAvailableTemplates() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, newSDK := range []bool{false, true} {
				fields, err := templates.GetLanguageConfigFields(types.NewTargetFromTemplate(name), newSDK)
				require.NoError(t, err)

				for _, field := range fields {
					if field.ValidationRegex == nil {
						continue
					}

					pattern := strings.ReplaceAll(*field.ValidationRegex, escapedSlash, "/")
					regex, err := regexp.Compile(pattern)
					require.NoError(t, err, "field %q (newSDK=%v)", field.Name, newSDK)

					if field.DefaultValue == nil {
						continue
					}

					defaultValue := fmt.Sprintf("%v", *field.DefaultValue)
					assert.True(t, regex.MatchString(defaultValue),
						"field %q (newSDK=%v): default %q does not match %s", field.Name, newSDK, defaultValue, pattern)
				}
			}
		})
	}
}
