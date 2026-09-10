package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

func TestSnapCLIMajorVersionModulePath(t *testing.T) {
	t.Parallel()

	genYaml := `cli:
  packageName: github.com/example/petstore-cli
  cliName: petstore
  envVarPrefix: PETSTORE
  version: 2.0.0
`

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:    cliReleaseSnapshotSpec,
		GenYaml: genYaml,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			sdkSource, err := os.ReadFile(filepath.Join(tempDir, "internal", "sdk", "sdk.go"))
			require.NoError(t, err)
			require.Contains(t, string(sdkSource), `"speakeasy-sdk/go 2.0.0 internal 0.1.0 github.com/example/petstore-cli/v2/internal/sdk"`)
			require.NotContains(t, string(sdkSource), "internal/sdk/v2")
			require.Contains(t, string(sdkSource), `SDKVersion:        "2.0.0",`)

			versionSource, err := os.ReadFile(filepath.Join(tempDir, "internal", "cli", "version.go"))
			require.NoError(t, err)
			require.Contains(t, string(versionSource), `var Version = "2.0.0"`)
		},
	})
}
