package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

// The README's quick-install one-liners fetch scripts/install.sh|ps1 and
// install binaries produced by the generated goreleaser pipeline. All of
// that exists only when release generation is on, so the installation
// section must follow the same gate: a release-less CLI README advertising
// `curl .../scripts/install.sh | bash` points at a file that is never
// generated and releases that are never built.
func TestSnapCLIInstallDocsFollowReleaseGate(t *testing.T) {
	t.Parallel()

	t.Run("release off", func(t *testing.T) {
		t.Parallel()
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec: cliReleaseSnapshotSpec,
			GenYaml: `cli:
  packageName: github.com/example/petstore-cli
  cliName: petstore
  envVarPrefix: PETSTORE
  generateRelease: false
`,
			// Pure-absence assertion: the install scripts must not be
			// generated at all when release is off.
			ExcludeGlobs: []string{"scripts/install.sh", "scripts/install.ps1"},
			AfterGenerate: func(t *testing.T, tempDir string) {
				t.Helper()
				readme, err := os.ReadFile(filepath.Join(tempDir, "README.md"))
				require.NoError(t, err)
				require.NotContains(t, string(readme), "scripts/install.sh",
					"release-less README must not advertise the quick-install script")
				require.NotContains(t, string(readme), "scripts/install.ps1")
				require.Contains(t, string(readme), "go install",
					"release-less README keeps the go install fallback")
			},
		})
	})

	t.Run("release on", func(t *testing.T) {
		t.Parallel()
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec: cliReleaseSnapshotSpec,
			GenYaml: `cli:
  packageName: github.com/example/petstore-cli
  cliName: petstore
  envVarPrefix: PETSTORE
  generateRelease: true
`,
			AfterGenerate: func(t *testing.T, tempDir string) {
				t.Helper()
				readme, err := os.ReadFile(filepath.Join(tempDir, "README.md"))
				require.NoError(t, err)
				require.Contains(t, string(readme), "scripts/install.sh",
					"release README advertises the quick-install script")
				_, err = os.Stat(filepath.Join(tempDir, "scripts", "install.sh"))
				require.NoError(t, err, "the advertised install script is generated")
			},
		})
	})
}
