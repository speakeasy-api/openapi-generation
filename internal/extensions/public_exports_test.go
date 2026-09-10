package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	oasextensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHandlePublicExportsExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expected    []PublicExport
		wantErr     bool
		errContains string
	}{
		{
			name:     "no extension",
			yaml:     `type: object`,
			expected: nil,
		},
		{
			name: "list",
			yaml: `
type: object
x-speakeasy-exports:
  - group: chat.completions
    name: chat_completion
  - group: responses
    name: response`,
			expected: []PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
				{Group: "responses", Name: "response"},
			},
		},
		{
			name: "single object",
			yaml: `
type: object
x-speakeasy-exports:
  group: chat.completions
  name: chat_completion`,
			expected: []PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
			},
		},
		{
			name: "input flag",
			yaml: `
type: object
x-speakeasy-exports:
  - group: webhooks
    name: webhook_ping_params
    representation: input
  - group: webhooks
    name: webhook_ping_request`,
			expected: []PublicExport{
				{Group: "webhooks", Name: "webhook_ping_params", Representation: PublicExportRepresentationInput},
				{Group: "webhooks", Name: "webhook_ping_request"},
			},
		},
		{
			name: "trims and deduplicates",
			yaml: `
type: object
x-speakeasy-exports:
  - group: " chat.completions "
    name: " chat_completion "
  - group: chat.completions
    name: chat_completion`,
			expected: []PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
			},
		},
		{
			name: "missing group",
			yaml: `
type: object
x-speakeasy-exports:
  - name: chat_completion`,
			wantErr:     true,
			errContains: "x-speakeasy-exports.group is required",
		},
		{
			name: "input only is not empty",
			yaml: `
type: object
x-speakeasy-exports:
  - representation: input`,
			wantErr:     true,
			errContains: "x-speakeasy-exports.group is required",
		},
		{
			name: "missing name",
			yaml: `
type: object
x-speakeasy-exports:
  - group: chat.completions`,
			wantErr:     true,
			errContains: "x-speakeasy-exports.name is required",
		},
		{
			name: "invalid shape",
			yaml: `
type: object
x-speakeasy-exports: true`,
			wantErr:     true,
			errContains: "failed to unmarshal x-speakeasy-exports",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			extensions := parseTestExtensions(t, tt.yaml)
			e := New(types.Target{Target: "typescript"})
			result, err := e.HandlePublicExportsExtension(extensions)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func parseTestExtensions(t *testing.T, raw string) *oasextensions.Extensions {
	t.Helper()

	var schemaNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(raw), &schemaNode))

	extensions := oasextensions.New()
	if schemaNode.Kind != yaml.DocumentNode || len(schemaNode.Content) == 0 {
		return extensions
	}

	mappingNode := schemaNode.Content[0]
	if mappingNode.Kind != yaml.MappingNode {
		return extensions
	}

	for i := 0; i < len(mappingNode.Content); i += 2 {
		key := mappingNode.Content[i].Value
		val := mappingNode.Content[i+1]
		extensions.Set(key, val)
	}

	return extensions
}
