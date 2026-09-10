package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

const patchedRegisterGenYAML = `cli:
  packageName: github.com/example/petstore-cli
  cliName: petstore
  envVarPrefix: PETSTORE
  enableCustomCodeRegions: true
`

// The content the register.go patch installs. Deliberately tiny and stable so
// the snapshot does not chase the scaffold template's prose.
const patchedRegisterContent = `package custom

import "github.com/spf13/cobra"

// Register content is owned by the .speakeasy/patches patch in this test.
func Register(root *cobra.Command) {
	_ = root
}
`

// A patch on internal/cli/custom/register.go pins the file to
// generator+patch ownership: the generated-once scaffold must keep being
// emitted so the patch has its target, even when a register.go already
// exists in the output tree from a previous generation. Without that rule
// the patch applies on a hermetic fresh-tree build but is silently skipped
// on incremental regeneration — same inputs, different outputs.
func TestSnapCLIPatchedRegisterAppliesOnIncrementalRegen(t *testing.T) {
	t.Parallel()

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         cliReleaseSnapshotSpec,
		GenYaml:      patchedRegisterGenYAML,
		IncludeGlobs: []string{"internal/cli/custom/register.go"},
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			registerPath := filepath.Join(tempDir, "internal", "cli", "custom", "register.go")
			pristine, err := os.ReadFile(registerPath)
			require.NoError(t, err)

			writeSnapshotPatch(t, tempDir, "internal/cli/custom/register.go", pristine, []byte(patchedRegisterContent))
		},
		Expected: `--- internal/cli/custom/register.go ---
` + patchedRegisterContent + `

`,
	})
}

// When a hand-edited register.go and a patch both exist, the patch wins
// deterministically: hand-edit preservation applies only to unpatched files.
func TestSnapCLIPatchedRegisterOverridesHandEdit(t *testing.T) {
	t.Parallel()

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         cliReleaseSnapshotSpec,
		GenYaml:      patchedRegisterGenYAML,
		IncludeGlobs: []string{"internal/cli/custom/register.go"},
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			registerPath := filepath.Join(tempDir, "internal", "cli", "custom", "register.go")
			pristine, err := os.ReadFile(registerPath)
			require.NoError(t, err)

			writeSnapshotPatch(t, tempDir, "internal/cli/custom/register.go", pristine, []byte(patchedRegisterContent))

			// A stray in-place edit must not survive once a patch owns the
			// file; keep it compilable so only content decides the outcome.
			handEdited := append([]byte(nil), pristine...)
			handEdited = append(handEdited, []byte("\n// hand edit that the patch must supersede\n")...)
			require.NoError(t, os.WriteFile(registerPath, handEdited, 0o644))
		},
		Expected: `--- internal/cli/custom/register.go ---
` + patchedRegisterContent + `

`,
	})
}
