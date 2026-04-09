package templatedir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDirectoryFilesFallsBackToBaseTemplate(t *testing.T) {
	cfg, err := LoadConfig("csharpv2")
	require.NoError(t, err)

	files, err := cfg.GetDirectoryFiles("testproject")
	require.NoError(t, err)
	require.Contains(t, files, "testproject/Program.cs.stmpl")
	require.Contains(t, files, "testproject/testproject.csproj.stmpl")
}

func TestGetDirectoryFilesMergesPrimaryAndBaseTemplate(t *testing.T) {
	cfg, err := LoadConfig("csharpv2")
	require.NoError(t, err)

	files, err := cfg.GetDirectoryFiles("includes")
	require.NoError(t, err)
	require.Contains(t, files, "includes/dependencies-aot.ts")
	require.Contains(t, files, "includes/tests.ts")
}
