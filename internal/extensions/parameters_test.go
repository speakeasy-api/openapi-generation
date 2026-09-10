package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHandleAllowEmptyQueryParameterValueExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		paramIn     openapi.ParameterIn
		yamlNode    *yaml.Node
		wantResult  bool
		wantErr     bool
		errContains string
	}{
		{
			name:    "returns true for true value",
			paramIn: "query",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("true"), &node)
				return node.Content[0]
			}(),
			wantResult: true,
		},
		{
			name:    "returns false for false value",
			paramIn: "query",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("false"), &node)
				return node.Content[0]
			}(),
			wantResult: false,
		},
		{
			name:    "returns error for invalid value",
			paramIn: "query",
			yamlNode: func() *yaml.Node {
				var node yaml.Node
				_ = yaml.Unmarshal([]byte("'not a boolean'"), &node)
				return node.Content[0]
			}(),
			wantErr:     true,
			errContains: "failed to unmarshal",
		},
		{
			name:       "returns false when not present",
			paramIn:    "query",
			yamlNode:   nil,
			wantResult: false,
		},
	}

	for i := range tests {
		tt := &tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			param := &openapi.Parameter{
				In:         tt.paramIn,
				Extensions: extensions.New(),
			}

			if tt.yamlNode != nil {
				param.Extensions.Set(ExtAllowEmptyValue.Name(), tt.yamlNode)
			}

			e := &Extensions{}
			result, err := e.HandleAllowEmptyQueryParameterValueExtension(param)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantResult, result)
		})
	}
}
