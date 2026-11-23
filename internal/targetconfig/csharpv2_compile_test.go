//go:build dotnet

package targetconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Requires a .NET SDK matching dotnetVersion; run with -tags dotnet.
func TestCsharpv2CompileWithoutFormattingPreservesSourceAndTrimDiagnostics(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	fixtureDir := filepath.Join("testdata", "csharpv2-compile")

	entries, err := os.ReadDir(fixtureDir)
	require.NoError(t, err)

	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(fixtureDir, entry.Name()))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, entry.Name()), data, 0o644))
	}

	source, err := os.ReadFile(filepath.Join(dir, "ReflectionJson.cs"))
	require.NoError(t, err)

	targetConfig, err := NewConfiguration(ctx, types.NewTargetFromTemplate("csharpv2"), GeneratorInitialConfiguration{
		LangCfg: map[string]any{
			"templateVersion":  "v2",
			"dotnetVersion":    "net10.0",
			"enableFormatting": false,
		},
		OutDir: dir,
	})
	require.NoError(t, err)
	require.NotNil(t, targetConfig.Compile)

	runner := targetConfig.Compile.Runner
	argVars := map[string]string{"PackageVersion": "0.0.1", "Usage": "false"}

	output, err := runner.CheckDependencies(ctx, dir, argVars)
	require.NoError(t, err, output)

	stdout, stderr, err := runner.RunCommands(ctx, dir, argVars)
	require.NoError(t, err, "%s\n%s", stdout, stderr)

	formatted, err := os.ReadFile(filepath.Join(dir, "ReflectionJson.cs"))
	require.NoError(t, err)
	assert.Equal(t, string(source), string(formatted))

	assert.Contains(t, stdout, "warning IL2026")
	assert.Contains(t, stdout, "warning IL3050")
}
