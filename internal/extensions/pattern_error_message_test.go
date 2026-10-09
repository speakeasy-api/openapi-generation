package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	openapiExtensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHandlePatternErrorMessageExtension(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, value, want string
		invalid           bool
	}{
		{name: "absent"},
		{name: "message", value: `'Use lowercase letters.'`, want: "Use lowercase letters."},
		{name: "preserves whitespace and literals", value: `"  Use \"café\" at 100%\nTry again.  "`, want: "  Use \"café\" at 100%\nTry again.  "},
		{name: "empty", value: `""`, invalid: true},
		{name: "whitespace", value: `" \t\n\u2003"`, invalid: true},
		{name: "number", value: "42", invalid: true},
		{name: "boolean", value: "true", invalid: true},
		{name: "null", value: "null", invalid: true},
		{name: "array", value: "[message]", invalid: true},
		{name: "object", value: "{message: text}", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			exts := openapiExtensions.New()
			if tc.value != "" {
				var node yaml.Node
				require.NoError(t, yaml.Unmarshal([]byte(tc.value), &node))
				exts.Set(ExtPatternErrorMessage.Name(), node.Content[0])
			}
			got, err := New(types.Target{}).HandlePatternErrorMessageExtension(exts)
			if tc.invalid {
				require.ErrorContains(t, err, "x-speakeasy-pattern-error-message must be a non-empty string")
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestHandlePatternErrorMessageExtensionRewrite(t *testing.T) {
	t.Parallel()
	e := New(types.Target{})
	var rewrites yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("x-speakeasy-pattern-error-message: x-pattern-message"), &rewrites))
	require.NoError(t, e.HandleRewriteExtension(WithRewritesExtensionNode(rewrites.Content[0])))
	exts := openapiExtensions.New()
	exts.Set("x-pattern-message", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "Use letters."})
	got, err := e.HandlePatternErrorMessageExtension(exts)
	require.NoError(t, err)
	assert.Equal(t, "Use letters.", got)
}
