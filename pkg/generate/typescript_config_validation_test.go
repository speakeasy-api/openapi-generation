package generate

import (
	"context"
	"testing"

	pkgerrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigTypeScriptFieldValidation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		target string
		cfg    map[string]any
		errMsg string
	}{
		"usageSDKInitImports must be an array": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2", "usageSDKInitImports": "acme"},
			errMsg: "usageSDKInitImports must be an array",
		},
		"usageSDKInitImports items must be objects": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{"acme"}},
			errMsg: "usageSDKInitImports[0] must be an object",
		},
		"usageSDKInitImports items must not be null": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{nil}},
			errMsg: "usageSDKInitImports[0] must be an object",
		},
		"usageSDKInitImports items require package": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"import": "Acme"},
			}},
			errMsg: "usageSDKInitImports[0].package must be a non-empty string",
		},
		"usageSDKInitImports items reject empty package": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "", "import": "Acme"},
			}},
			errMsg: "usageSDKInitImports[0].package must be a non-empty string",
		},
		"usageSDKInitImports items reject non-string package": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": 1, "import": "Acme"},
			}},
			errMsg: "usageSDKInitImports[0].package must be a non-empty string",
		},
		"usageSDKInitImports items require import": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk"},
			}},
			errMsg: "usageSDKInitImports[0].import must be a non-empty string",
		},
		"usageSDKInitImports items reject empty import": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk", "Import": ""},
			}},
			errMsg: "usageSDKInitImports[0].import must be a non-empty string",
		},
		"usageSDKInitImports items reject non-string import": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk", "import": 1},
			}},
			errMsg: "usageSDKInitImports[0].import must be a non-empty string",
		},
		"usageSDKInitImports items reject unknown type": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk", "import": "Acme"},
				map[string]any{"package": "@acme/sdk", "import": "Acme", "type": "defaultImport"},
			}},
			errMsg: "usageSDKInitImports[1].type must be one of: typeImport, packageImport, aliasImport",
		},
		"usageSDKInitImports items reject unknown capitalized type": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"Package": "@acme/sdk", "Import": "Acme", "Type": "defaultImport"},
			}},
			errMsg: "usageSDKInitImports[0].type must be one of: typeImport, packageImport, aliasImport",
		},
		"advancedFlags must be an array": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": "verbose"},
			errMsg: "advancedFlags must be an array",
		},
		"advancedFlags items must be strings": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": []any{1}},
			errMsg: "advancedFlags[0] must be a string",
		},
		"advancedFlags items must be alphanumeric": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": []any{"verbose", "dry-run"}},
			errMsg: "advancedFlags[1] must contain only letters or numbers",
		},
		"advancedFlags items must not be empty": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": []any{""}},
			errMsg: "advancedFlags[0] must contain only letters or numbers",
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := g.validateConfig(context.Background(), newLanguageTestConfig(tc.target, tc.cfg), tc.target)
			require.Error(t, err)

			var validationErr *pkgerrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "validation error: gen.yaml validation error for "+tc.target+": "+tc.errMsg, err.Error())
		})
	}
}

func TestValidateConfigTypeScriptFieldValidationSuccess(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		target string
		cfg    map[string]any
	}{
		"usageSDKInitImports absent": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2"},
		},
		"usageSDKInitImports null": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2", "usageSDKInitImports": nil},
		},
		"usageSDKInitImports empty": {
			target: "typescript",
			cfg:    map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{}},
		},
		"usageSDKInitImports valid": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk", "import": "Acme"},
				map[string]any{"package": "@acme/sdk", "import": "AcmeOptions", "type": "packageImport"},
				map[string]any{"package": "@acme/sdk", "import": "AcmeTypes", "type": "typeImport"},
				map[string]any{"package": "@acme/sdk", "import": "acme", "type": "aliasImport"},
			}},
		},
		"usageSDKInitImports capitalized keys": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"Package": "@acme/sdk", "Import": "Acme", "Type": "packageImport"},
			}},
		},
		"usageSDKInitImports mixed-case keys": {
			target: "typescript",
			cfg: map[string]any{"templateVersion": "v2", "usageSDKInitImports": []any{
				map[string]any{"package": "@acme/sdk", "Import": "Acme", "type": "packageImport"},
				map[string]any{"package": "", "Package": "@acme/sdk", "import": "AcmeOptions"},
			}},
		},
		"advancedFlags absent": {
			target: "mcp-typescript",
			cfg:    map[string]any{},
		},
		"advancedFlags null": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": nil},
		},
		"advancedFlags valid": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": []any{"verbose", "loneQueryParam", "option2"}},
		},
		"advancedFlags empty": {
			target: "mcp-typescript",
			cfg:    map[string]any{"advancedFlags": []any{}},
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, g.validateConfig(context.Background(), newLanguageTestConfig(tc.target, tc.cfg), tc.target))
		})
	}
}
