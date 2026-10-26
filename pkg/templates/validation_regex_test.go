package templates

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// goja's RegExp.source emits an unescaped "/" as the Unicode escape below, which the CLI rewrites before compiling the pattern with Go's regexp.
const escapedSlash = `\u002f`

type validationField struct {
	config.SDKGenConfigField `mapstructure:",squash"`
	ValidationExamples       struct {
		Valid   []string
		Invalid []string
	}
	// Declared on some fields in config.ts but read by neither the generator nor the CLI.
	ValidationFunc any
}

func TestValidationRegexes(t *testing.T) {
	t.Parallel()

	t.Run("common", func(t *testing.T) {
		t.Parallel()

		e, err := executor.New(types.NewTargetFromTemplate("common"), "common/config.ts")
		require.NoError(t, err)

		for _, newSDK := range []bool{false, true} {
			rawFields, err := e.Run("getCommonConfigFields", newSDK)
			require.NoError(t, err)

			checkValidationFields(t, decodeValidationFields(t, rawFields), newSDK)
		}
	})

	for _, name := range GetAvailableTemplates() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			target := types.NewTargetFromTemplate(name)
			e, err := executor.New(target, "config.ts")
			require.NoError(t, err)

			for _, newSDK := range []bool{false, true} {
				_, err := GetLanguageConfigFields(target, newSDK)
				require.NoError(t, err)

				commonFields, err := getCommonConfigFields(newSDK)
				require.NoError(t, err)

				rawFields, err := e.Run("getConfigFields", commonFields, newSDK)
				require.NoError(t, err)

				checkValidationFields(t, decodeValidationFields(t, rawFields), newSDK)
			}
		})
	}
}

func decodeValidationFields(t *testing.T, rawFields any) map[string]validationField {
	t.Helper()

	rawFieldMap, ok := rawFields.(map[string]any)
	require.True(t, ok, "config fields are %T", rawFields)

	fields := map[string]validationField{}
	for key, rawField := range rawFieldMap {
		// Common fields passed through unchanged arrive as Go structs and are checked by the "common" subtest.
		if _, ok := rawField.(map[string]any); !ok {
			continue
		}

		var field validationField
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{ErrorUnused: true, Result: &field})
		require.NoError(t, err)
		require.NoError(t, decoder.Decode(rawField), "field %q", key)
		fields[key] = field
	}

	return fields
}

func checkValidationFields(t *testing.T, fields map[string]validationField, newSDK bool) {
	t.Helper()

	for key, field := range fields {
		if field.ValidationRegex == nil {
			continue
		}

		pattern := strings.ReplaceAll(*field.ValidationRegex, escapedSlash, "/")
		regex, err := regexp.Compile(pattern)
		require.NoError(t, err, "field %q (newSDK=%v)", key, newSDK)

		if field.DefaultValue != nil {
			defaultValue := fmt.Sprintf("%v", *field.DefaultValue)
			assert.True(t, regex.MatchString(defaultValue),
				"field %q (newSDK=%v): default %q does not match %s", key, newSDK, defaultValue, pattern)
		}

		for _, value := range field.ValidationExamples.Valid {
			assert.True(t, regex.MatchString(value),
				"field %q (newSDK=%v): valid example %q does not match %s", key, newSDK, value, pattern)
		}

		for _, value := range field.ValidationExamples.Invalid {
			assert.False(t, regex.MatchString(value),
				"field %q (newSDK=%v): invalid example %q matches %s", key, newSDK, value, pattern)
		}
	}
}
