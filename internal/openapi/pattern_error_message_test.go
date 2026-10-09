package openapi

import (
	"context"
	"testing"

	genextensions "github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMergePatternErrorMessage(t *testing.T) {
	t.Parallel()
	for _, rewritten := range []bool{false, true} {
		for _, deep := range []bool{false, true} {
			for _, tc := range []struct {
				name                              string
				pattern                           *string
				message, wantPattern, wantMessage string
			}{
				{name: "equal pattern retains message", pattern: pointer.From("^[a-z]+$"), wantPattern: "^[a-z]+$", wantMessage: "Use letters."},
				{name: "changed pattern clears message", pattern: pointer.From("^[0-9]+$"), wantPattern: "^[0-9]+$"},
				{name: "changed pattern with own message", pattern: pointer.From("^[0-9]+$"), message: "Use digits.", wantPattern: "^[0-9]+$", wantMessage: "Use digits."},
				{name: "equal pattern with own message", pattern: pointer.From("^[a-z]+$"), message: "Use lowercase letters.", wantPattern: "^[a-z]+$", wantMessage: "Use lowercase letters."},
				{name: "inherited pattern with own message", message: "Use lowercase letters.", wantPattern: "^[a-z]+$", wantMessage: "Use lowercase letters."},
				{name: "empty pattern clears message", pattern: pointer.From(""), wantPattern: ""},
			} {
				name := tc.name
				if rewritten {
					name += "/rewritten"
				}
				if deep {
					name += "/deep"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					e := genextensions.New(types.Target{})
					name := genextensions.ExtPatternErrorMessage.Name()
					if rewritten {
						name = "x-pattern-message"
						var rewrite yaml.Node
						require.NoError(t, yaml.Unmarshal([]byte("x-speakeasy-pattern-error-message: "+name), &rewrite))
						require.NoError(t, e.HandleRewriteExtension(genextensions.WithRewritesExtensionNode(rewrite.Content[0])))
					}
					base := &oas3.Schema{Pattern: pointer.From("^[a-z]+$"), Extensions: extensions.New()}
					base.Extensions.Set(name, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "Use letters."})
					overriding := &oas3.Schema{Pattern: tc.pattern, Extensions: extensions.New()}
					if tc.message != "" {
						overriding.Extensions.Set(name, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: tc.message})
					}
					if deep {
						require.NoError(t, deepMergeInto(context.Background(), base, overriding, true, e, nil))
					} else {
						require.NoError(t, Merge(context.Background(), base, overriding, true, true, e, config.AllOfMergeStrategyShallowMerge, nil))
					}
					assert.Equal(t, tc.wantPattern, base.GetPattern())
					got, err := e.HandlePatternErrorMessageExtension(base.GetExtensions())
					require.NoError(t, err)
					assert.Equal(t, tc.wantMessage, got)
				})
			}
		}
	}
}

func TestDeepMergePatternErrorMessageNilOverride(t *testing.T) {
	t.Parallel()
	e := genextensions.New(types.Target{})
	require.NoError(t, deepMergeInto(context.Background(), nil, nil, true, e, nil))
	base := &oas3.Schema{Pattern: pointer.From("^[a-z]+$"), Extensions: extensions.New()}
	base.Extensions.Set(genextensions.ExtPatternErrorMessage.Name(), &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "Use letters."})
	require.NoError(t, deepMergeInto(context.Background(), base, nil, true, e, nil))
	assert.Equal(t, "^[a-z]+$", base.GetPattern())
	message, err := e.HandlePatternErrorMessageExtension(base.GetExtensions())
	require.NoError(t, err)
	assert.Equal(t, "Use letters.", message)
}

func TestDeepMergePatternErrorMessageNilSchema(t *testing.T) {
	t.Parallel()
	e := genextensions.New(types.Target{})
	schema := oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{Pattern: pointer.From("^[a-z]+$")})
	got, err := deepMergeJSONSchema(context.Background(), nil, schema, true, e, nil)
	require.NoError(t, err)
	assert.Same(t, schema, got)
	got, err = deepMergeJSONSchema(context.Background(), schema, nil, true, e, nil)
	require.NoError(t, err)
	assert.Same(t, schema, got)
}
