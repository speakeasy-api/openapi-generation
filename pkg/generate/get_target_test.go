package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		genYAML      string
		wantTarget   string
		wantTemplate string
	}{
		{name: "csharp without gen.yaml", target: "csharp", wantTarget: "csharp", wantTemplate: "csharp"},
		{name: "csharp templateVersion v2", target: "csharp", genYAML: "csharp:\n  templateVersion: v2\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "csharpv2", target: "csharpv2", genYAML: "csharp:\n  templateVersion: v2\n", wantTarget: "csharp", wantTemplate: "csharpv2"},
		{name: "typescript without gen.yaml", target: "typescript", wantTarget: "typescript", wantTemplate: "typescriptv2"},
		{name: "typescriptv2", target: "typescriptv2", genYAML: "typescript:\n  templateVersion: v2\n", wantTarget: "typescript", wantTemplate: "typescriptv2"},
		{name: "javav2", target: "javav2", wantTarget: "java", wantTemplate: "javav2"},
		{name: "pythonv2", target: "pythonv2", wantTarget: "python", wantTemplate: "pythonv2"},
		{name: "go templateVersion v2 without go v2 template", target: "go", genYAML: "go:\n  templateVersion: v2\n", wantTarget: "go", wantTemplate: "go"},
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
			})
		}
	}
}
