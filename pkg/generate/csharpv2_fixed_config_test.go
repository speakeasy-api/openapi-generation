package generate

import (
	"bytes"
	"context"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var csharpV2FixedConfig = map[string]any{
	"unionStrategy":                  "populated-fields",
	"optionalNullableScope":          "all",
	"presenceAwareJsonSerialization": true,
	"partialTypes":                   true,
	"unionImplicitConversions":       true,
	"enableFormatting":               true,
}

var csharpLegacyConfig = map[string]any{
	"unionStrategy":                  "left-to-right",
	"optionalNullableScope":          "writable-models",
	"presenceAwareJsonSerialization": false,
	"partialTypes":                   false,
	"unionImplicitConversions":       false,
	"enableFormatting":               false,
}

func TestCsharpV2FixedConfigKeysAreNotConfigFields(t *testing.T) {
	t.Parallel()

	target := types.NewTargetFromTemplate("csharpv2")
	fixed, err := templates.GetFixedConfig(target)
	require.NoError(t, err)
	assert.Equal(t, csharpV2FixedConfig, fixed)

	for _, newSDK := range []bool{true, false} {
		fields, err := templates.GetLanguageConfigFields(target, newSDK)
		require.NoError(t, err)
		for _, field := range fields {
			assert.NotContains(t, csharpV2FixedConfig, field.Name, "newSDK=%t", newSDK)
		}

		defaults, err := GetLanguageConfigDefaults("csharpv2", newSDK)
		require.NoError(t, err)
		for key := range csharpV2FixedConfig {
			assert.NotContains(t, defaults.Cfg, key, "newSDK=%t", newSDK)
		}
	}
}

// Not parallel: captures the standard logger that receives console.warn from config.ts.
func TestCsharpFixedConfigInitAndSave(t *testing.T) {
	tests := []struct {
		name            string
		templateVersion string
		seed            map[string]any
		wantResolved    map[string]any
		wantWarnings    []string
		wantSaved       map[string]any
	}{
		{
			name:            "v2 replaces legacy values with a warning and does not save them",
			templateVersion: "v2",
			seed:            csharpLegacyConfig,
			wantResolved:    csharpV2FixedConfig,
			wantWarnings:    slices.Sorted(maps.Keys(csharpV2FixedConfig)),
			wantSaved:       map[string]any{},
		},
		{
			name:            "v2 drops fixed values silently",
			templateVersion: "v2",
			seed:            csharpV2FixedConfig,
			wantResolved:    csharpV2FixedConfig,
			wantSaved:       map[string]any{},
		},
		{
			name:            "v2 treats empty values as unset",
			templateVersion: "v2",
			seed: map[string]any{
				"unionStrategy":         nil,
				"optionalNullableScope": "",
			},
			wantResolved: csharpV2FixedConfig,
			wantSaved:    map[string]any{},
		},
		{
			name:            "v1 is forced to v2, replaces legacy values with a warning and does not save them",
			templateVersion: "v1",
			seed:            csharpLegacyConfig,
			wantResolved:    csharpV2FixedConfig,
			wantWarnings:    slices.Sorted(maps.Keys(csharpV2FixedConfig)),
			wantSaved:       map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()
			genYAMLPath := filepath.Join(outDir, ".speakeasy", "gen.yaml")
			writeCsharpGenYAML(t, genYAMLPath, tt.templateVersion, tt.seed)

			var logs bytes.Buffer
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(os.Stderr) })

			ctx, err := resolveGenerationAccess(generationaccess.WithDirect(context.Background()), "csharp", true)
			require.NoError(t, err)
			g, err := New()
			require.NoError(t, err)
			require.NoError(t, g.Init(ctx, "csharp", outDir))

			resolved := g.subsystem.Config.Languages["csharp"].Cfg
			for key, value := range tt.wantResolved {
				assert.Equal(t, value, resolved[key], key)
			}

			var warned []string
			for line := range strings.Lines(logs.String()) {
				if _, warning, ok := strings.Cut(line, "csharp: "); ok && strings.Contains(warning, "is not supported with templateVersion: v2") {
					key, _, _ := strings.Cut(warning, " ")
					warned = append(warned, key)
				}
			}
			slices.Sort(warned)
			assert.Equal(t, tt.wantWarnings, warned)

			_, errs := g.saveConfig(ctx)
			require.Empty(t, errs)

			data, err := os.ReadFile(genYAMLPath)
			require.NoError(t, err)
			var saved config.Configuration
			require.NoError(t, yaml.Unmarshal(data, &saved))
			savedFixed := map[string]any{}
			for key := range csharpV2FixedConfig {
				if value, ok := saved.Languages["csharp"].Cfg[key]; ok {
					savedFixed[key] = value
				}
			}
			assert.Equal(t, tt.wantSaved, savedFixed)
			assert.Equal(t, "v2", saved.Languages["csharp"].Cfg["templateVersion"])
			assert.Equal(t, "net8.0", saved.Languages["csharp"].Cfg["dotnetVersion"])
		})
	}
}

func writeCsharpGenYAML(t *testing.T, path, templateVersion string, seed map[string]any) {
	t.Helper()

	langCfg := map[string]any{
		"version":         "0.0.1",
		"packageName":     "Acme.Widgets",
		"templateVersion": templateVersion,
	}
	maps.Copy(langCfg, seed)
	data, err := yaml.Marshal(map[string]any{
		"configVersion": config.Version,
		"generation":    map[string]any{"sdkClassName": "Widgets"},
		"csharp":        langCfg,
	})
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))
}
