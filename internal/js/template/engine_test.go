package template

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
)

func TestNullishString(t *testing.T) {
	vm := goja.New()
	tests := []struct {
		name  string
		value goja.Value
		want  string
	}{
		{name: "null", value: goja.Null(), want: ""},
		{name: "undefined", value: goja.Undefined(), want: ""},
		{name: "empty string", value: vm.ToValue(""), want: ""},
		{name: "string", value: vm.ToValue("comment-key"), want: "comment-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, nullishString(tt.value))
		})
	}
}

func TestGetSymbolCasingOverrides(t *testing.T) {
	tests := []struct {
		name string
		cfg  map[string]any
		want []string
	}{
		{
			name: "reads customCasings",
			cfg: map[string]any{
				"customCasings": map[string]any{
					"api": map[string]any{"initialism": true},
					"datakit": map[string]any{
						"pascal": "DataKit",
					},
					"exampleai": map[string]any{
						"pascal": "ExampleAI",
					},
				},
			},
			want: []string{"API", "DataKit", "ExampleAI"},
		},
		{
			name: "returns independent overrides for each config",
			cfg: map[string]any{
				"customCasings": map[string]any{
					"mcp": map[string]any{"initialism": true},
				},
			},
			want: []string{"MCP"},
		},
		{
			name: "returns empty without customCasings",
			cfg:  map[string]any{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := configuration.New(&config.Configuration{
				Languages: map[string]config.LanguageConfig{
					"typescript": {
						Cfg: tt.cfg,
					},
				},
			}, "typescript")

			got := getSymbolCasingOverrides(Config{
				Subsystem: &subsystem.Subsystem{
					Config: cfg,
				},
			})

			assert.Equal(t, tt.want, got)
		})
	}
}
