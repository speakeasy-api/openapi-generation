package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi/extensions"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHandleTransformExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		yamlNode   *yaml.Node
		wantConfig string
		wantErr    bool
	}{
		{
			name: "accepts structured YAML format with simple expression",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("jq: '.[]'"), &node)
				return node.Content[0]
			}(),
			wantConfig: ".[]",
			wantErr:    false,
		},
		{
			name: "accepts structured YAML format with complex expression",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("jq: '. + {id: .data.result[0].id}'"), &node)
				return node.Content[0]
			}(),
			wantConfig: ". + { id: .data.result[0].id }",
			wantErr:    false,
		},
		{
			name: "rejects mapping without jq field",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("other: '.[]'"), &node)
				return node.Content[0]
			}(),
			wantErr: true,
		},
		{
			name: "rejects empty jq expression",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("jq: ''"), &node)
				return node.Content[0]
			}(),
			wantErr: true,
		},
		{
			name: "rejects scalar string format",
			yamlNode: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "jq .[]",
			},
			wantErr: true,
		},
		{
			name: "rejects invalid jq expression",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("jq: '.[['"), &node)
				return node.Content[0]
			}(),
			wantErr: true,
		},
	}

	for i := range tests {
		tt := &tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exts := extensions.New()
			exts.Set(ExtTransformToAPI.Name(), tt.yamlNode)

			e := &Extensions{}
			cfg, err := e.HandleTransformExtension(exts, ExtTransformToAPI)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			require.Equal(t, Jq, cfg.Type)
			require.Equal(t, tt.wantConfig, cfg.Config)
		})
	}
}
