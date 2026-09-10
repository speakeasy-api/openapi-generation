package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGoOptionalMethodArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expected    *GoOptionalMethodArguments
		errContains string
	}{
		{
			name: "unset",
			yaml: "type: object",
		},
		{
			name:     "pointers",
			yaml:     "x-speakeasy-go-optional-method-arguments: pointers",
			expected: goOptionalMethodArgumentsPtr(GoOptionalMethodArgumentsPointers),
		},
		{
			name:     "shared options",
			yaml:     "x-speakeasy-go-optional-method-arguments: shared-options",
			expected: goOptionalMethodArgumentsPtr(GoOptionalMethodArgumentsSharedOptions),
		},
		{
			name:     "method options",
			yaml:     "x-speakeasy-go-optional-method-arguments: method-options",
			expected: goOptionalMethodArgumentsPtr(GoOptionalMethodArgumentsMethodOptions),
		},
		{
			name:        "invalid value",
			yaml:        "x-speakeasy-go-optional-method-arguments: options",
			errContains: "value is not allowed",
		},
		{
			name:        "invalid type",
			yaml:        "x-speakeasy-go-optional-method-arguments: [pointers]",
			errContains: "failed to unmarshal x-speakeasy-go-optional-method-arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exts := parseTestExtensions(t, tt.yaml)
			e := New(types.Target{Target: "go"})
			result, err := e.GetGoOptionalMethodArguments(exts)

			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetGoOptionalMethodArgumentsRewrittenExtensionName(t *testing.T) {
	t.Parallel()

	e := New(types.Target{Target: "go"})
	rewrites := parseTestExtensions(t, `
x-speakeasy-extension-rewrite:
  x-speakeasy-go-optional-method-arguments: x-customer-go-optional-method-arguments`)
	require.NoError(t, e.HandleRewriteExtension(WithDocumentExtensions(rewrites)))

	exts := parseTestExtensions(t, "x-customer-go-optional-method-arguments: method-options")
	result, err := e.GetGoOptionalMethodArguments(exts)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, GoOptionalMethodArgumentsMethodOptions, *result)
}

func goOptionalMethodArgumentsPtr(value GoOptionalMethodArguments) *GoOptionalMethodArguments {
	return &value
}
