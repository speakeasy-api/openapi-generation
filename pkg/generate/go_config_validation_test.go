package generate

import (
	"context"
	"testing"

	pkgerrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigGoRetractions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		retractions any
		errMsg      string
	}{
		"not an array": {
			retractions: "v1.0.0",
			errMsg:      "retractions must be an array",
		},
		"item without version": {
			retractions: []any{map[string]any{"comment": "Published accidentally"}},
			errMsg:      "retractions[0].version must follow semantic versioning format starting with 'v' (e.g., 'v1.0.0', 'v1.2.3-beta.1')",
		},
		"version without v prefix": {
			retractions: []any{
				map[string]any{"version": "v1.0.0"},
				map[string]any{"version": "1.0.1"},
			},
			errMsg: "retractions[1].version must follow semantic versioning format starting with 'v' (e.g., 'v1.0.0', 'v1.2.3-beta.1')",
		},
		"item is a string": {
			retractions: []any{"v1.0.0"},
			errMsg:      "retractions[0].version must follow semantic versioning format starting with 'v' (e.g., 'v1.0.0', 'v1.2.3-beta.1')",
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := g.validateConfig(context.Background(), newLanguageTestConfig("go", map[string]any{"retractions": tc.retractions}), "go")
			require.Error(t, err)

			var validationErr *pkgerrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "validation error: gen.yaml validation error for go: "+tc.errMsg, err.Error())
		})
	}
}

func TestValidateConfigGoRetractionsSuccess(t *testing.T) {
	t.Parallel()

	testCases := map[string]any{
		"absent": nil,
		"empty":  []any{},
		"valid": []any{
			map[string]any{"version": "v1.0.0", "comment": "Published accidentally"},
			map[string]any{"version": "v1.2.3-beta.1"},
			map[string]any{"version": "v2.0.0+build.5"},
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, retractions := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := map[string]any{}
			if retractions != nil {
				cfg["retractions"] = retractions
			}
			assert.NoError(t, g.validateConfig(context.Background(), newLanguageTestConfig("go", cfg), "go"))
		})
	}
}
