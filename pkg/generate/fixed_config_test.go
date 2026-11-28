package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSaveConfigOmitsFixedConfigKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		fixedConfigKeys []string
		wantSaved       map[string]any
	}{
		{
			name:            "fixed keys are not saved",
			fixedConfigKeys: []string{"compactOutput", "widgetMode"},
			wantSaved:       map[string]any{"packageName": "acme"},
		},
		{
			name:            "no fixed keys saves every key",
			fixedConfigKeys: nil,
			wantSaved: map[string]any{
				"packageName":   "acme",
				"widgetMode":    "strict",
				"compactOutput": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outDir := t.TempDir()
			genYAMLPath := filepath.Join(outDir, ".speakeasy", "gen.yaml")
			require.NoError(t, os.MkdirAll(filepath.Dir(genYAMLPath), 0o755))
			require.NoError(t, os.WriteFile(genYAMLPath, []byte("configVersion: "+config.Version+"\n"), 0o644))

			target := types.NewTargetFromTemplate("go")
			cfg := &config.Configuration{
				ConfigVersion: config.Version,
				Languages: map[string]config.LanguageConfig{
					target.Target: {
						Version: "1.0.0",
						Cfg: map[string]any{
							"packageName":   "acme",
							"widgetMode":    "strict",
							"compactOutput": true,
						},
					},
				},
			}
			g := &Generator{
				outDir:          outDir,
				target:          target,
				subsystem:       &subsystem.Subsystem{Config: configuration.New(cfg, target.Target)},
				fixedConfigKeys: tt.fixedConfigKeys,
			}

			_, errs := g.saveConfig(context.Background())
			require.Empty(t, errs)

			data, err := os.ReadFile(genYAMLPath)
			require.NoError(t, err)
			var saved config.Configuration
			require.NoError(t, yaml.Unmarshal(data, &saved))

			assert.Equal(t, tt.wantSaved, saved.Languages[target.Target].Cfg)
			assert.Equal(t, "1.0.0", saved.Languages[target.Target].Version)
			assert.Equal(t, map[string]any{
				"packageName":   "acme",
				"widgetMode":    "strict",
				"compactOutput": true,
			}, g.subsystem.Config.Languages[target.Target].Cfg)
		})
	}
}
