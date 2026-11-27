package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		genYAML      string
		wantTarget   string
		wantTemplate string
		wantCfg      map[string]any
	}{
		{name: "csharp without gen.yaml", target: "csharp", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharp without templateVersion", target: "csharp", genYAML: "csharp:\n  packageName: SDK\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharp templateVersion v1", target: "csharp", genYAML: "csharp:\n  templateVersion: v1\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharp templateVersion v2", target: "csharp", genYAML: "csharp:\n  templateVersion: v2\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharp templateVersion v1 dotnetVersion net6.0", target: "csharp", genYAML: "csharp:\n  templateVersion: v1\n  dotnetVersion: net6.0\n  packageName: Acme.Sdk\n  enableFormatting: false\n", wantTarget: "csharp", wantTemplate: "csharpv2", wantCfg: map[string]any{"packageName": "Acme.Sdk", "enableFormatting": false}},
		{name: "csharp dotnetVersion net6.0", target: "csharp", genYAML: "csharp:\n  dotnetVersion: net6.0\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharp templateVersion v2 dotnetVersion net10.0", target: "csharp", genYAML: "csharp:\n  templateVersion: v2\n  dotnetVersion: net10.0\n  enableFormatting: true\n", wantTarget: "csharp", wantTemplate: "csharpv2", wantCfg: map[string]any{"enableFormatting": true}},
		{name: "csharpv2 without gen.yaml", target: "csharpv2", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharpv2", target: "csharpv2", genYAML: "csharp:\n  templateVersion: v2\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharpv2 dotnetVersion net10.0", target: "csharpv2", genYAML: "csharp:\n  templateVersion: v2\n  dotnetVersion: net10.0\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "typescript without gen.yaml", target: "typescript", wantTarget: "typescript", wantTemplate: "typescriptv2"},
		{name: "typescriptv2", target: "typescriptv2", genYAML: "typescript:\n  templateVersion: v2\n", wantTarget: "typescript", wantTemplate: "typescriptv2", wantCfg: map[string]any{"templateVersion": "v2"}},
		{name: "javav2", target: "javav2", wantTarget: "java", wantTemplate: "javav2"},
		{name: "pythonv2", target: "pythonv2", wantTarget: "python", wantTemplate: "pythonv2"},
		{name: "go templateVersion v2 without go v2 template", target: "go", genYAML: "go:\n  templateVersion: v2\n", wantTarget: "go", wantTemplate: "go", wantCfg: map[string]any{"templateVersion": "v2"}},
	}

	for _, debug := range []string{"true", "false"} {
		for _, tt := range tests {
			t.Run(tt.name+"/debug="+debug, func(t *testing.T) {
				t.Setenv("SPEAKEASY_DEBUG", debug)

				outDir := t.TempDir()
				if tt.genYAML != "" {
					require.NoError(t, os.MkdirAll(filepath.Join(outDir, ".speakeasy"), 0o755))
					genYAML := "configVersion: 2.0.0\ngeneration:\n  sdkClassName: SDK\n" + tt.genYAML
					require.NoError(t, os.WriteFile(filepath.Join(outDir, ".speakeasy", "gen.yaml"), []byte(genYAML), 0o644))
				}

				g, err := New()
				require.NoError(t, err)

				got, err := g.GetTarget(context.Background(), tt.target, outDir)
				require.NoError(t, err)
				assert.Equal(t, tt.wantTarget, got.Target)
				assert.Equal(t, tt.wantTemplate, got.Template)

				g.target = got
				loaded, err := g.LoadConfig(context.Background(), outDir, got.Target)
				require.NoError(t, err)
				effective := loaded.Config.Languages[got.Target].Cfg

				contents, err := os.ReadFile(loaded.ConfigPath)
				require.NoError(t, err)
				var persisted map[string]any
				require.NoError(t, yaml.Unmarshal(contents, &persisted))
				persistedCfg, ok := persisted[got.Target].(map[string]any)
				require.True(t, ok)

				for _, cfg := range []map[string]any{effective, persistedCfg} {
					for k, v := range tt.wantCfg {
						assert.Equal(t, v, cfg[k], k)
					}
					if got.Target == "csharp" {
						assert.Equal(t, "v2", cfg["templateVersion"])
						assert.Equal(t, "net8.0", cfg["dotnetVersion"])
					} else {
						assert.NotContains(t, cfg, "dotnetVersion")
					}
				}
			})
		}
	}
}
