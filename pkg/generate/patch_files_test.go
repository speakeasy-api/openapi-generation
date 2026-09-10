package generate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	internalpatches "github.com/speakeasy-api/openapi-generation/v2/internal/patches"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/patchfiles"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestApplyImplicitPatchFiles_PatchAppliesAccordingToPersistentEditsFlag(t *testing.T) {
	t.Parallel()

	pristine := []byte(`package testsdk

type Status struct{}
`)
	patched := []byte(`package testsdk

type Status struct{}

func (s *Status) String() string {
	return "custom"
}
`)
	cases := []struct {
		name                   string
		persistentEditsEnabled config.PersistentEditsEnabled
		expected               []byte
	}{
		{"persistent edits disabled writes patched to disk", config.PersistentEditsEnabledNever, patched},
		{"persistent edits enabled defers disk write", config.PersistentEditsEnabledTrue, pristine},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			writePatchFixture(t, tempDir, "status.go", pristine, patched)
			require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), pristine, 0o644))

			g := newPatchFilesGenerator(tempDir, &tc.persistentEditsEnabled, nil)
			vf := &merge.VirtualFile{Path: "status.go", Content: pristine, Mode: 0o644}
			g.virtualFiles.Store("status.go", vf)

			err := g.applyImplicitPatchFiles(context.Background())
			require.NoError(t, err)

			assert.Equal(t, string(patched), string(vf.Content))
			content, err := os.ReadFile(filepath.Join(tempDir, "status.go"))
			require.NoError(t, err)
			assert.Equal(t, string(tc.expected), string(content))
		})
	}
}

func TestApplyImplicitPatchFiles_RejectsInvalidPatch(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writePatchFile(t, patchfiles.PathFor(tempDir, "sdk.go"), []byte(`diff --git a/sdk.go b/sdk.go
--- a/sdk.go
+++ b/sdk.go
@@ -1,3 +1,4 @@
 package testsdk

+// sdk custom
 type SDK struct{}
diff --git a/models/components/pet.go b/models/components/pet.go
--- a/models/components/pet.go
+++ b/models/components/pet.go
@@ -1,3 +1,4 @@
 package components

+// pet custom
 type Pet struct{}
`))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	vf := &merge.VirtualFile{Path: "sdk.go", Content: []byte("package testsdk\n\ntype SDK struct{}\n"), Mode: 0o644}
	g.virtualFiles.Store("sdk.go", vf)

	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "patch apply failed for sdk.go")
}

// Verifies the continue-on-error iteration policy: malformed patches don't
// abort the loop so that errors on subsequent VFs can still be surfaced
// (joined via errors.Join) and successful patches still get applied.
func TestApplyImplicitPatchFiles_ContinuesAcrossPerVFErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Multi-file diff with a path the VF doesn't own — triggers non-conflict parse/apply err.
	badPatch := []byte(`diff --git a/sdk.go b/sdk.go
--- a/sdk.go
+++ b/sdk.go
@@ -1,3 +1,4 @@
 package testsdk

+// sdk custom
 type SDK struct{}
diff --git a/models/components/pet.go b/models/components/pet.go
--- a/models/components/pet.go
+++ b/models/components/pet.go
@@ -1,3 +1,4 @@
 package components

+// pet custom
 type Pet struct{}
`)
	writePatchFile(t, patchfiles.PathFor(tempDir, "first.go"), badPatch)
	writePatchFile(t, patchfiles.PathFor(tempDir, "second.go"), badPatch)

	cleanPristine := []byte("package testsdk\n\ntype Clean struct{}\n")
	cleanPatched := []byte("package testsdk\n\ntype Clean struct{}\n\nfunc (c *Clean) String() string {\n\treturn \"patched\"\n}\n")
	writePatchFixture(t, tempDir, "clean.go", cleanPristine, cleanPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "clean.go"), cleanPristine, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.virtualFiles.Store("first.go", &merge.VirtualFile{Path: "first.go", Content: []byte("package testsdk\n\ntype SDK struct{}\n"), Mode: 0o644})
	g.virtualFiles.Store("second.go", &merge.VirtualFile{Path: "second.go", Content: []byte("package testsdk\n\ntype SDK struct{}\n"), Mode: 0o644})
	cleanVF := &merge.VirtualFile{Path: "clean.go", Content: cleanPristine, Mode: 0o644}
	g.virtualFiles.Store("clean.go", cleanVF)

	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "patch apply failed for first.go")
	assert.Contains(t, err.Error(), "patch apply failed for second.go")

	assert.Equal(t, string(cleanPatched), string(cleanVF.Content), "clean patch must still apply to its VF despite sibling failures")
	written, readErr := os.ReadFile(filepath.Join(tempDir, "clean.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(cleanPatched), string(written), "clean patch must still be flushed to disk")
}

// Orphan implicit patch coexisting with applied generic patchset.
// Asserts warnUnusedPatchFiles distinguishes membership in usedPaths
// regardless of source (implicit vs generic) — only the orphan warns.
func TestApplyImplicitPatchFiles_WarnsOnlyForOrphanAlongsideGenericPatch(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	writePatchFixture(t, tempDir, "orphan.go", []byte("package testsdk\n"), []byte("package testsdk\n\nfunc orphan() {}\n"))

	genericPristine := []byte("package testsdk\n\ntype Generic struct{}\n")
	genericPatched := []byte("package testsdk\n\ntype Generic struct{}\n\nfunc (g *Generic) String() string {\n\treturn \"generic\"\n}\n")
	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), buildUnifiedPatchset(t, map[string][]byte{
		"generic.go": genericPristine,
	}, map[string][]byte{
		"generic.go": genericPatched,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "generic.go"), genericPristine, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	genericVF := &merge.VirtualFile{Path: "generic.go", Content: genericPristine, Mode: 0o644}
	g.virtualFiles.Store("generic.go", genericVF)

	logger := &patchFilesMockLogger{}
	ctx := logging.With(context.Background(), logger)

	require.NoError(t, g.applyImplicitPatchFiles(ctx))
	assert.Equal(t, string(genericPatched), string(genericVF.Content))

	require.Len(t, logger.warns, 1)
	assert.Equal(t, "Patch file does not match any generated file and was not applied", logger.warns[0])
}

func TestApplyImplicitPatchFiles_WritesConflictMarkersAndStagesGitConflict(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	enabled := config.PersistentEditsEnabledTrue
	base := []byte(`package testsdk

func StatusLabel() string {
	return "base"
}
`)
	currentGenerated := []byte(`package testsdk

func StatusLabel() string {
	return "user"
}
`)
	desiredPatched := []byte(`package testsdk

func StatusLabel() string {
	return "custom"
}
`)

	writePatchFixture(t, tempDir, "status.go", base, desiredPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), currentGenerated, 0o644))

	mockGit := newPatchFilesMockGit()
	baseHash := mockGit.addObject(base)

	g := newPatchFilesGenerator(tempDir, &enabled, mockGit)
	g.lockFile.TrackedFiles.Set("status.go", lockfile.TrackedFile{PristineGitObject: baseHash})

	vf := &merge.VirtualFile{Path: "status.go", Content: currentGenerated, Mode: 0o644}
	g.virtualFiles.Store("status.go", vf)

	err := g.applyImplicitPatchFiles(context.Background())
	var conflictsErr *merge.ConflictsError
	require.ErrorAs(t, err, &conflictsErr)
	require.Equal(t, []string{"status.go"}, conflictsErr.Files)

	written, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(written), string(vf.Content))
	assert.Contains(t, string(written), "<<<<<<< Current (Your changes)")
	assert.Contains(t, string(written), ">>>>>>> New (Generated by Speakeasy)")

	state, ok := mockGit.conflicts["status.go"]
	require.True(t, ok)
	assert.Equal(t, string(base), string(state.base))
	assert.Equal(t, string(currentGenerated), string(state.ours))
	assert.Equal(t, string(desiredPatched), string(state.theirs))
}

// No git wired in: patchConflictStages bails, Merge3Way is skipped, and
// gitdiffparser's inline conflict markers are written verbatim.
func TestApplyImplicitPatchFiles_WritesRawConflictMarkersWithoutGitStages(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	enabled := config.PersistentEditsEnabledTrue
	base := []byte(`package testsdk

type Status struct{}
`)
	currentGenerated := []byte(`package testsdk

type Status struct {
	Value string
}
`)
	desiredPatched := []byte(`package testsdk

type Status struct{}

func (s *Status) String() string {
	return "custom"
}
`)

	writePatchFixture(t, tempDir, "status.go", base, desiredPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), currentGenerated, 0o644))

	g := newPatchFilesGenerator(tempDir, &enabled, nil)
	vf := &merge.VirtualFile{Path: "status.go", Content: currentGenerated, Mode: 0o644}
	g.virtualFiles.Store("status.go", vf)

	err := g.applyImplicitPatchFiles(context.Background())
	var conflictsErr *merge.ConflictsError
	require.ErrorAs(t, err, &conflictsErr)
	require.Equal(t, []string{"status.go"}, conflictsErr.Files)

	written, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(written), string(vf.Content))
	assert.Contains(t, string(written), "<<<<<<< Current")
	assert.Contains(t, string(written), ">>>>>>> Incoming patch")
}

// Test applyImplicitPatchFiles (rebases stale patch onto regenerated content) followed by
// a hand-rolled 3-way merge (standing in for handlePatches) to reconcile user disk edits.
// Does not exercise subsystem.Patches.PerformMerge orchestration.
func TestApplyImplicitPatchFiles_RebasesStalePatchAndComposesWithPersistentEdits(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	enabled := config.PersistentEditsEnabledTrue
	base := []byte(`package testsdk

type Status struct{}
`)
	// persistent edit
	userEditedDisk := []byte(`package testsdk

type Status struct{}

type UserField struct{}
`)
	// spec adds `Value` field to Status.
	currentGenerated := []byte(`package testsdk

type Status struct {
	NewValue string
}
`)
	// implicit patch authored against base
	desiredPatched := []byte(`package testsdk

type Status struct{}

func (s *Status) String() string {
	return "custom"
}
`)
	// applyImplicitPatchFiles output: schema + patch merged. User field absent
	// because disk is excluded from this merge — handlePatches reconciles it next.
	expectedPatchApplied := []byte(`package testsdk

type Status struct {
	NewValue string
}

func (s *Status) String() string {
	return "custom"
}
`)
	expectedMerged := []byte(`package testsdk

type Status struct {
	NewValue string
}

func (s *Status) String() string {
	return "custom"
}

type UserField struct{}
`)

	writePatchFixture(t, tempDir, "status.go", base, desiredPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), userEditedDisk, 0o644))

	mockGit := newPatchFilesMockGit()
	baseHash := mockGit.addObject(base)

	g := newPatchFilesGenerator(tempDir, &enabled, mockGit)
	g.lockFile.TrackedFiles.Set("status.go", lockfile.TrackedFile{PristineGitObject: baseHash})

	vf := &merge.VirtualFile{Path: "status.go", Content: currentGenerated, Mode: 0o644}
	g.virtualFiles.Store("status.go", vf)

	logger := &patchFilesMockLogger{}
	ctx := logging.With(context.Background(), logger)

	// Step 1: applyImplicitPatchFiles rebases stale patch onto currentGenerated.
	require.NoError(t, g.applyImplicitPatchFiles(ctx))

	written, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(expectedPatchApplied), string(vf.Content), "VF holds merged schema+patch content")
	assert.Equal(t, string(userEditedDisk), string(written), "applyImplicitPatchFiles must leave disk untouched; handlePatches reconciles user edits later")

	_, gitConflict := mockGit.conflicts["status.go"]
	assert.False(t, gitConflict)

	require.Len(t, logger.warns, 1)
	assert.Equal(t, "Stale patch was rebased onto regenerated content. Please verify result", logger.warns[0])

	// Step 2: stand in for handlePatches' 3-way merge of (pristine, VF, disk).
	final, mergeConflict := merge.Merge3Way(base, vf.Content, userEditedDisk)
	require.False(t, mergeConflict, "expected orthogonal merge: schema/patch in VF, UserField on disk")
	assert.Equal(t, string(expectedMerged), string(final), "schema, patch, and user persistent edit all preserved")
}

// Mirror of RebasesStalePatchAndComposesWithPersistentEdits but with PE disabled.
// Asserts the rebased content is flushed to disk immediately (no deferred write
// since handlePatches is not in the pipeline to reconcile later).
func TestApplyImplicitPatchFiles_StaleRebaseFlushesToDiskWhenPersistentEditsDisabled(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	disabled := config.PersistentEditsEnabledNever
	base := []byte(`package testsdk

type Status struct{}
`)
	currentGenerated := []byte(`package testsdk

type Status struct {
	NewValue string
}
`)
	desiredPatched := []byte(`package testsdk

type Status struct{}

func (s *Status) String() string {
	return "custom"
}
`)
	expectedPatchApplied := []byte(`package testsdk

type Status struct {
	NewValue string
}

func (s *Status) String() string {
	return "custom"
}
`)

	writePatchFixture(t, tempDir, "status.go", base, desiredPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), currentGenerated, 0o644))

	mockGit := newPatchFilesMockGit()
	baseHash := mockGit.addObject(base)

	g := newPatchFilesGenerator(tempDir, &disabled, mockGit)
	g.lockFile.TrackedFiles.Set("status.go", lockfile.TrackedFile{PristineGitObject: baseHash})

	vf := &merge.VirtualFile{Path: "status.go", Content: currentGenerated, Mode: 0o644}
	g.virtualFiles.Store("status.go", vf)

	logger := &patchFilesMockLogger{}
	ctx := logging.With(context.Background(), logger)

	require.NoError(t, g.applyImplicitPatchFiles(ctx))

	written, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(expectedPatchApplied), string(written), "rebased content must be flushed to disk when PE disabled")
	assert.Equal(t, string(expectedPatchApplied), string(vf.Content))

	_, gitConflict := mockGit.conflicts["status.go"]
	assert.False(t, gitConflict)

	require.Len(t, logger.warns, 1)
	assert.Equal(t, "Stale patch was rebased onto regenerated content. Please verify result", logger.warns[0])
}

func TestApplyImplicitPatchFiles_MixedApplyAndConflictAcrossFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanPristine := []byte(`package testsdk

type Clean struct{}
`)
	cleanPatched := []byte(`package testsdk

type Clean struct{}

func (c *Clean) String() string {
	return "patched"
}
`)
	writePatchFixture(t, tempDir, "clean.go", cleanPristine, cleanPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "clean.go"), cleanPristine, 0o644))

	conflictBase := []byte(`package testsdk

func StatusLabel() string {
	return "base"
}
`)
	conflictCurrent := []byte(`package testsdk

func StatusLabel() string {
	return "user"
}
`)
	conflictPatched := []byte(`package testsdk

func StatusLabel() string {
	return "custom"
}
`)
	writePatchFixture(t, tempDir, "status.go", conflictBase, conflictPatched)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), conflictCurrent, 0o644))

	mockGit := newPatchFilesMockGit()
	baseHash := mockGit.addObject(conflictBase)

	g := newPatchFilesGenerator(tempDir, nil, mockGit)
	g.lockFile.TrackedFiles.Set("status.go", lockfile.TrackedFile{PristineGitObject: baseHash})

	cleanVF := &merge.VirtualFile{Path: "clean.go", Content: cleanPristine, Mode: 0o644}
	conflictVF := &merge.VirtualFile{Path: "status.go", Content: conflictCurrent, Mode: 0o644}
	g.virtualFiles.Store("clean.go", cleanVF)
	g.virtualFiles.Store("status.go", conflictVF)

	err := g.applyImplicitPatchFiles(context.Background())
	var conflictsErr *merge.ConflictsError
	require.ErrorAs(t, err, &conflictsErr)
	require.Equal(t, []string{"status.go"}, conflictsErr.Files)

	cleanWritten, readErr := os.ReadFile(filepath.Join(tempDir, "clean.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(cleanPatched), string(cleanVF.Content))
	assert.Equal(t, string(cleanWritten), string(cleanVF.Content))

	conflictWritten, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(conflictWritten), string(conflictVF.Content))
	assert.Contains(t, string(conflictWritten), "<<<<<<< Current (Your changes)")
	assert.Contains(t, string(conflictWritten), ">>>>>>> New (Generated by Speakeasy)")
}

func TestApplyImplicitPatchFiles_AppliesGenericGitDiffAcrossMultipleFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanPristine := []byte(`package testsdk

type Clean struct{}
`)
	cleanPatched := []byte(`package testsdk

type Clean struct{}

func (c *Clean) String() string {
	return "patched"
}
`)
	statusPristine := []byte(`package testsdk

type Status struct{}
`)
	statusPatched := []byte(`package testsdk

type Status struct{}

func (s *Status) String() string {
	return "patched"
}
`)

	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), buildUnifiedPatchset(t, map[string][]byte{
		"clean.go":  cleanPristine,
		"status.go": statusPristine,
	}, map[string][]byte{
		"clean.go":  cleanPatched,
		"status.go": statusPatched,
	}))

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "clean.go"), cleanPristine, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "status.go"), statusPristine, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	cleanVF := &merge.VirtualFile{Path: "clean.go", Content: cleanPristine, Mode: 0o644}
	statusVF := &merge.VirtualFile{Path: "status.go", Content: statusPristine, Mode: 0o644}
	g.virtualFiles.Store("clean.go", cleanVF)
	g.virtualFiles.Store("status.go", statusVF)

	err := g.applyImplicitPatchFiles(context.Background())
	require.NoError(t, err)

	cleanWritten, readErr := os.ReadFile(filepath.Join(tempDir, "clean.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(cleanPatched), string(cleanVF.Content))
	assert.Equal(t, string(cleanPatched), string(cleanWritten))

	statusWritten, readErr := os.ReadFile(filepath.Join(tempDir, "status.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(statusPatched), string(statusVF.Content))
	assert.Equal(t, string(statusPatched), string(statusWritten))
}

func TestApplyImplicitPatchFiles_CreatesUntrackedDiskFileOverExistingFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	enabled := config.PersistentEditsEnabledTrue

	registrationPath := "src/hooks/registration.ts"
	existing := []byte(`import { Hooks } from "./types.js";

export function initHooks(hooks: Hooks) {
    hooks.registerAfterSuccessHook(new FixResponseHook());
}
`)
	created := []byte(`import { Hooks } from "./types.js";
import { IdempotencyHook } from "./idempotency.js";

export function initHooks(hooks: Hooks) {
    hooks.registerBeforeRequestHook(new IdempotencyHook());
    hooks.registerAfterSuccessHook(new FixResponseHook());
}
`)

	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), []byte(`diff --git a/src/hooks/registration.ts b/src/hooks/registration.ts
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/src/hooks/registration.ts
@@ -0,0 +1,7 @@
+import { Hooks } from "./types.js";
+import { IdempotencyHook } from "./idempotency.js";
+
+export function initHooks(hooks: Hooks) {
+    hooks.registerBeforeRequestHook(new IdempotencyHook());
+    hooks.registerAfterSuccessHook(new FixResponseHook());
+}
`))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "src", "hooks"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, registrationPath), existing, 0o644))

	g := newPatchFilesGenerator(tempDir, &enabled, nil)
	g.subsystem.FileTracker = filetracking.NewTracker(context.Background(), filetracking.Options{
		OutDir:    tempDir,
		GenLockId: "test",
		FS:        g,
	})
	g.subsystem.FileTracker.AddUntrackedPattern(regexp.MustCompile(`^src/hooks/registration\.ts`))

	err := g.applyImplicitPatchFiles(context.Background())
	require.NoError(t, err)

	written, readErr := os.ReadFile(filepath.Join(tempDir, registrationPath))
	require.NoError(t, readErr)
	assert.Equal(t, string(created), string(written))

	_, inVirtualFiles := g.virtualFiles.Load(registrationPath)
	assert.False(t, inVirtualFiles)
	_, tracked := g.subsystem.FileTracker.GetResult().NewFiles.Get(registrationPath)
	assert.False(t, tracked)
}

func TestApplyImplicitPatchFiles_CreatesFileFromExactPatchPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	createdContent := []byte("package testsdk\n\nfunc Custom() {}\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "custom.go"), []byte(`diff --git a/custom.go b/custom.go
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/custom.go
@@ -0,0 +1,3 @@
+package testsdk
+
+func Custom() {}
`))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.subsystem.FileTracker = filetracking.NewTracker(context.Background(), filetracking.Options{
		OutDir:    tempDir,
		GenLockId: "test",
		FS:        g,
	})

	require.NoError(t, g.applyImplicitPatchFiles(context.Background()))

	vf, ok := g.virtualFiles.Load("custom.go")
	require.True(t, ok)
	assert.Equal(t, string(createdContent), string(vf.Content))
	assert.Equal(t, fs.FileMode(defaultFileMode), vf.Mode)

	written, err := os.ReadFile(filepath.Join(tempDir, "custom.go"))
	require.NoError(t, err)
	assert.Equal(t, string(createdContent), string(written))

	expectedChecksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(createdContent))
	require.NoError(t, err)
	trackedFile, tracked := g.subsystem.FileTracker.GetResult().NewFiles.Get("custom.go")
	require.True(t, tracked)
	assert.Equal(t, "sha1:"+expectedChecksum, trackedFile.LastWriteChecksum)
}

func TestApplyImplicitPatchFiles_RejectsUnsupportedSymlinkMode(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), []byte(`diff --git a/current.yaml b/current.yaml
new file mode 120000
index 0000000..b3e7f4a
--- /dev/null
+++ b/current.yaml
@@ -0,0 +1 @@
+../config/prod.yaml
\ No newline at end of file
`))

	g := newPatchFilesGenerator(tempDir, nil, nil)

	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported git mode "120000"`)
	assert.Contains(t, err.Error(), "only regular files are currently supported")

	_, ok := g.virtualFiles.Load("current.yaml")
	assert.False(t, ok)
	_, statErr := os.Stat(filepath.Join(tempDir, "current.yaml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestApplyImplicitPatchFiles_DeletesFileFromExactPatchPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	content := []byte("package testsdk\n\ntype DeleteMe struct{}\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "delete_me.go"), []byte(`diff --git a/delete_me.go b/delete_me.go
deleted file mode 100644
index 1111111..0000000
--- a/delete_me.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package testsdk
-
-type DeleteMe struct{}
`))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "delete_me.go"), content, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.subsystem.FileTracker = filetracking.NewTracker(context.Background(), filetracking.Options{
		OutDir:    tempDir,
		GenLockId: "test",
		FS:        g,
	})
	g.subsystem.FileTracker.TrackFile("delete_me.go", content)
	g.virtualFiles.Store("delete_me.go", &merge.VirtualFile{Path: "delete_me.go", Content: content, Mode: 0o644})

	require.NoError(t, g.applyImplicitPatchFiles(context.Background()))

	_, ok := g.virtualFiles.Load("delete_me.go")
	assert.False(t, ok)
	_, statErr := os.Stat(filepath.Join(tempDir, "delete_me.go"))
	assert.True(t, os.IsNotExist(statErr))

	tracked := g.subsystem.FileTracker.GetResult().NewFiles
	_, trackedFile := tracked.Get("delete_me.go")
	assert.False(t, trackedFile)
}

func TestApplyImplicitPatchFiles_RejectsProtectedCreateWithoutPartialApply(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sdkOriginal := []byte("package testsdk\n\ntype SDK struct{}\n")
	patchData := []byte("diff --git a/sdk.go b/sdk.go\n" +
		"--- a/sdk.go\n" +
		"+++ b/sdk.go\n" +
		"@@ -1,3 +1,4 @@\n" +
		" package testsdk\n" +
		" \n" +
		"+// patched\n" +
		" type SDK struct{}\n" +
		"diff --git a/.speakeasy/gen.yaml b/.speakeasy/gen.yaml\n" +
		"new file mode 100644\n" +
		"index 0000000..1111111\n" +
		"--- /dev/null\n" +
		"+++ b/.speakeasy/gen.yaml\n" +
		"@@ -0,0 +1,2 @@\n" +
		"+go:\n" +
		"+  packageName: example\n")
	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), patchData)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "sdk.go"), sdkOriginal, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	vf := &merge.VirtualFile{Path: "sdk.go", Content: sdkOriginal, Mode: 0o644}
	g.virtualFiles.Store("sdk.go", vf)

	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "protected path")

	assert.Equal(t, string(sdkOriginal), string(vf.Content))
	written, readErr := os.ReadFile(filepath.Join(tempDir, "sdk.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(sdkOriginal), string(written))
	_, statErr := os.Stat(filepath.Join(tempDir, ".speakeasy", "gen.yaml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestApplyImplicitPatchFiles_RejectsPatchPathOutsideOutputDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	outsideName := filepath.Base(tempDir) + "-outside.txt"
	outsidePath := filepath.Join(filepath.Dir(tempDir), outsideName)
	t.Cleanup(func() { _ = os.Remove(outsidePath) })

	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), []byte(`diff --git a/../`+outsideName+` b/../`+outsideName+`
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/../`+outsideName+`
@@ -0,0 +1 @@
+outside
`))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolves outside the output directory")

	_, statErr := os.Stat(outsidePath)
	assert.True(t, os.IsNotExist(statErr))
}

func TestApplyImplicitPatchFiles_RejectsDeleteOfUserModifiedPersistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	enabled := config.PersistentEditsEnabledTrue
	pristine := []byte("package testsdk\n\ntype DeleteMe struct{}\n")
	userModified := []byte("package testsdk\n\ntype DeleteMe struct{}\n\nfunc Custom() {}\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "delete_me.go"), []byte(`diff --git a/delete_me.go b/delete_me.go
deleted file mode 100644
index 1111111..0000000
--- a/delete_me.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package testsdk
-
-type DeleteMe struct{}
`))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "delete_me.go"), userModified, 0o644))

	g := newPatchFilesGenerator(tempDir, &enabled, nil)
	checksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(pristine))
	require.NoError(t, err)
	g.lockFile.TrackedFiles.Set("delete_me.go", lockfile.TrackedFile{LastWriteChecksum: "sha1:" + checksum})
	g.virtualFiles.Store("delete_me.go", &merge.VirtualFile{Path: "delete_me.go", Content: pristine, Mode: 0o644})

	err = g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "would remove user-modified file")

	_, ok := g.virtualFiles.Load("delete_me.go")
	assert.True(t, ok)
	written, readErr := os.ReadFile(filepath.Join(tempDir, "delete_me.go"))
	require.NoError(t, readErr)
	assert.Equal(t, string(userModified), string(written))
}

func TestApplyImplicitPatchFiles_RejectsCreateOverExistingUntrackedFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	existing := []byte("customer file\n")
	created := []byte("generated file\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "custom.txt"), []byte(`diff --git a/custom.txt b/custom.txt
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/custom.txt
@@ -0,0 +1 @@
+generated file
`))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "custom.txt"), existing, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-generated file already exists")

	_, ok := g.virtualFiles.Load("custom.txt")
	assert.False(t, ok)
	written, readErr := os.ReadFile(filepath.Join(tempDir, "custom.txt"))
	require.NoError(t, readErr)
	assert.Equal(t, string(existing), string(written))
	assert.NotEqual(t, string(created), string(written))
}

func TestApplyImplicitPatchFiles_CreatesFileOverTrackedExistingFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	existing := []byte("old generated file\n")
	created := []byte("new generated file\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "custom.txt"), []byte(`diff --git a/custom.txt b/custom.txt
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/custom.txt
@@ -0,0 +1 @@
+new generated file
`))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "custom.txt"), existing, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.lockFile.TrackedFiles.Set("custom.txt", lockfile.TrackedFile{LastWriteChecksum: "sha1:old"})
	g.subsystem.FileTracker = filetracking.NewTracker(context.Background(), filetracking.Options{
		OutDir:    tempDir,
		GenLockId: "test",
		FS:        g,
	})

	require.NoError(t, g.applyImplicitPatchFiles(context.Background()))

	vf, ok := g.virtualFiles.Load("custom.txt")
	require.True(t, ok)
	assert.Equal(t, string(created), string(vf.Content))

	written, err := os.ReadFile(filepath.Join(tempDir, "custom.txt"))
	require.NoError(t, err)
	assert.Equal(t, string(created), string(written))

	expectedChecksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(created))
	require.NoError(t, err)
	trackedFile, tracked := g.subsystem.FileTracker.GetResult().NewFiles.Get("custom.txt")
	require.True(t, tracked)
	assert.Equal(t, "sha1:"+expectedChecksum, trackedFile.LastWriteChecksum)
}

func TestApplyImplicitPatchFiles_CreatesFileOverCurrentVirtualFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	existing := []byte("old generated file\n")
	created := []byte("new generated file\n")
	writePatchFile(t, patchfiles.PathFor(tempDir, "custom.txt"), []byte(`diff --git a/custom.txt b/custom.txt
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/custom.txt
@@ -0,0 +1 @@
+new generated file
`))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "custom.txt"), existing, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.virtualFiles.Store("custom.txt", &merge.VirtualFile{
		Path:    "custom.txt",
		Content: existing,
		Mode:    0o644,
	})

	require.NoError(t, g.applyImplicitPatchFiles(context.Background()))

	vf, ok := g.virtualFiles.Load("custom.txt")
	require.True(t, ok)
	assert.Equal(t, string(created), string(vf.Content))

	written, err := os.ReadFile(filepath.Join(tempDir, "custom.txt"))
	require.NoError(t, err)
	assert.Equal(t, string(created), string(written))
}

func TestValidatePatchOutputPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	g := newPatchFilesGenerator(tempDir, nil, nil)

	cases := []struct {
		name        string
		path        string
		want        string
		errContains string
	}{
		{
			name: "relative child path",
			path: "models/pet.go",
			want: "models/pet.go",
		},
		{
			name: "dot segments stay inside output directory",
			path: "models/../pet.go",
			want: "pet.go",
		},
		{
			name:        "absolute unix path",
			path:        "/tmp/pet.go",
			errContains: "is absolute",
		},
		{
			name:        "absolute windows path",
			path:        `C:\tmp\pet.go`,
			errContains: "is absolute",
		},
		{
			name:        "parent traversal leaves output directory",
			path:        "../pet.go",
			errContains: "resolves outside the output directory",
		},
		{
			name:        "target is output directory",
			path:        ".",
			errContains: "resolves to the output directory",
		},
		{
			name:        "nul byte",
			path:        "pet\x00.go",
			errContains: "contains a NUL byte",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := g.validatePatchOutputPath(tc.path)
			if tc.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeAppliedPatchTree_RejectsDuplicateResolvedOutputPath(t *testing.T) {
	t.Parallel()

	g := newPatchFilesGenerator(t.TempDir(), nil, nil)
	_, err := g.normalizeAppliedPatchTree(map[string][]byte{
		"models/../pet.go": []byte("package testsdk\n"),
		"pet.go":           []byte("package testsdk\n"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `patch contains multiple file entries that resolve to generated file "pet.go"`)
	assert.Contains(t, err.Error(), "each patch file entry must target a distinct generated file")
}

func TestPatchMentionsLookupPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		patch  string
		lookup string
		want   bool
	}{
		{
			name:   "standard a/b prefix matches",
			patch:  "diff --git a/foo.go b/foo.go\n",
			lookup: "foo.go",
			want:   true,
		},
		{
			name:   "standard a/b prefix non-match",
			patch:  "diff --git a/foo.go b/foo.go\n",
			lookup: "bar.go",
			want:   false,
		},
		{
			name:   "rename header matches from path",
			patch:  "diff --git a/old.go b/new.go\nrename from old.go\nrename to new.go\n",
			lookup: "old.go",
			want:   true,
		},
		{
			name:   "rename header matches to path",
			patch:  "diff --git a/old.go b/new.go\nrename from old.go\nrename to new.go\n",
			lookup: "new.go",
			want:   true,
		},
		{
			name:   "multi-file patch matches second header",
			patch:  "diff --git a/first.go b/first.go\n@@ -0,0 +1 @@\n+x\ndiff --git a/second.go b/second.go\n@@ -0,0 +1 @@\n+y\n",
			lookup: "second.go",
			want:   true,
		},
		{
			name:   "binary diff header matches",
			patch:  "diff --git a/foo.bin b/foo.bin\nBinary files a/foo.bin and b/foo.bin differ\n",
			lookup: "foo.bin",
			want:   true,
		},
		{
			name:   "quoted path with space matches",
			patch:  "diff --git \"a/foo bar.go\" \"b/foo bar.go\"\n",
			lookup: "foo bar.go",
			want:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := patchMentionsLookupPath([]byte(tc.patch), tc.lookup)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestApplyImplicitPatchFiles_RejectsGenericGitDiffWithPrefixedPaths(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanPristine := []byte(`package testsdk

type Clean struct{}
`)
	cleanPatched := []byte(`package testsdk

type Clean struct{}

func (c *Clean) String() string {
	return "patched"
}
`)

	patchData := buildUnifiedPatchset(t, map[string][]byte{
		"clean.go": cleanPristine,
	}, map[string][]byte{
		"clean.go": cleanPatched,
	})
	patchData = bytes.ReplaceAll(patchData, []byte("a/clean.go"), []byte("a/sdk/clean.go"))
	patchData = bytes.ReplaceAll(patchData, []byte("b/clean.go"), []byte("b/sdk/clean.go"))

	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), patchData)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "clean.go"), cleanPristine, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	cleanVF := &merge.VirtualFile{Path: "clean.go", Content: cleanPristine, Mode: 0o644}
	g.virtualFiles.Store("clean.go", cleanVF)

	err := g.applyImplicitPatchFiles(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "patch apply failed for delta.patch")
	assert.Contains(t, err.Error(), `cannot modify missing file "sdk/clean.go"`)
}

func TestApplyImplicitPatchFiles_AppliesGenericPatchWithMultipleFileSetMutations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	contentA := []byte("content a\n")
	contentB := []byte("content b\n")
	contentE := []byte("content e\n")

	patchData := []byte(`diff --git a/a.go b/a.go
deleted file mode 100644
index 1111111..0000000
--- a/a.go
+++ /dev/null
@@ -1 +0,0 @@
-content a
diff --git a/b.go b/b.go
deleted file mode 100644
index 2222222..0000000
--- a/b.go
+++ /dev/null
@@ -1 +0,0 @@
-content b
diff --git a/c.go b/c.go
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/c.go
@@ -0,0 +1 @@
+content c
diff --git a/d.go b/d.go
new file mode 100644
index 0000000..4444444
--- /dev/null
+++ b/d.go
@@ -0,0 +1 @@
+content d
diff --git a/e.go b/f.go
similarity index 100%
rename from e.go
rename to f.go
`)

	writePatchFile(t, filepath.Join(patchfiles.PatchesDir(tempDir), "delta.patch"), patchData)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "a.go"), contentA, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "b.go"), contentB, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "e.go"), contentE, 0o644))

	g := newPatchFilesGenerator(tempDir, nil, nil)
	g.virtualFiles.Store("a.go", &merge.VirtualFile{Path: "a.go", Content: contentA, Mode: 0o644})
	g.virtualFiles.Store("b.go", &merge.VirtualFile{Path: "b.go", Content: contentB, Mode: 0o644})
	g.virtualFiles.Store("e.go", &merge.VirtualFile{Path: "e.go", Content: contentE, Mode: 0o644})

	err := g.applyImplicitPatchFiles(context.Background())
	require.NoError(t, err)

	for _, p := range []string{"a.go", "b.go", "e.go"} {
		_, ok := g.virtualFiles.Load(p)
		assert.False(t, ok, "patch must remove virtual file %s", p)
		_, statErr := os.Stat(filepath.Join(tempDir, p))
		assert.True(t, os.IsNotExist(statErr), "patch must remove %s from disk", p)
	}

	expected := map[string]string{
		"c.go": "content c\n",
		"d.go": "content d\n",
		"f.go": "content e\n",
	}
	for p, expectedContent := range expected {
		vf, ok := g.virtualFiles.Load(p)
		require.True(t, ok, "patch must create virtual file %s", p)
		assert.Equal(t, expectedContent, string(vf.Content))

		written, readErr := os.ReadFile(filepath.Join(tempDir, p))
		require.NoError(t, readErr)
		assert.Equal(t, expectedContent, string(written))
		_, statErr := os.Stat(filepath.Join(tempDir, p))
		assert.NoError(t, statErr, "patch must create %s on disk", p)
	}
}

func TestGenerate_PatchFileConflictsStillPersistLockFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(testPatchFilesSnapshotGenYAML), 0o644))

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	generator, err := New(WithDebuggingEnabled())
	require.NoError(t, err)
	require.Empty(t, generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, false))

	listPetsPath := filepath.Join(tempDir, "pkg", "models", "operations", "listpets.go")
	listPetsOriginal, err := os.ReadFile(listPetsPath)
	require.NoError(t, err)

	listPetsPatched := append([]byte(nil), listPetsOriginal...)
	listPetsPatched = append(listPetsPatched, []byte(`
func (l *ListPetsResponse) PetCount() int {
	if l == nil {
		return 0
	}
	return len(l.Classes)
}
`)...)
	writePatchFixture(t, tempDir, "pkg/models/operations/listpets.go", listPetsOriginal, listPetsPatched)

	petStaleBase := []byte(`// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package shared

type Pet struct {
	ID   int64  ` + "`" + `json:"id"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}

func (p *Pet) GetID() int64 {
	if p == nil {
		return 0
	}
	return p.ID
}

func (p *Pet) GetName() string {
	if p == nil {
		return ""
	}
	return "pet:" + p.Name
}
`)
	petPatched := []byte(`// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package shared

type Pet struct {
	ID   int64  ` + "`" + `json:"id"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}

func (p *Pet) GetID() int64 {
	if p == nil {
		return 0
	}
	return p.ID
}

func (p *Pet) GetName() string {
	if p == nil {
		return ""
	}
	return "animal:" + p.Name
}
`)
	writePatchFixture(t, tempDir, "pkg/models/shared/pet.go", petStaleBase, petPatched)

	cfgBefore, err := config.Load(tempDir)
	require.NoError(t, err)
	trackedBefore, ok := cfgBefore.LockFile.TrackedFiles.Get("pkg/models/operations/listpets.go")
	require.True(t, ok)

	generator, err = New(WithDebuggingEnabled())
	require.NoError(t, err)
	errs := generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, false)
	require.Len(t, errs, 1)

	var conflictsErr *merge.ConflictsError
	require.ErrorAs(t, errs[0], &conflictsErr)
	require.Equal(t, []string{"pkg/models/shared/pet.go"}, conflictsErr.Files)

	cfgAfter, err := config.Load(tempDir)
	require.NoError(t, err)
	trackedAfter, ok := cfgAfter.LockFile.TrackedFiles.Get("pkg/models/operations/listpets.go")
	require.True(t, ok)

	expectedChecksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(listPetsPatched))
	require.NoError(t, err)
	assert.NotEqual(t, trackedBefore.LastWriteChecksum, trackedAfter.LastWriteChecksum)
	assert.Equal(t, "sha1:"+expectedChecksum, trackedAfter.LastWriteChecksum)
}

func TestGenerate_FinalCompileFailureStillPersistsConfigAndLockFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(testPatchFilesSnapshotGenYAMLPersistentEdits), 0o644))

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	mockGit := newPatchFilesMockGit()
	mockGit.repoRoot = tempDir

	generator, err := New(WithDebuggingEnabled(), WithGit(mockGit))
	require.NoError(t, err)
	require.Empty(t, generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, false))

	listPetsPath := filepath.Join(tempDir, "pkg", "models", "operations", "listpets.go")
	listPetsOriginal, err := os.ReadFile(listPetsPath)
	require.NoError(t, err)

	listPetsPatched := append([]byte(nil), listPetsOriginal...)
	listPetsPatched = append(listPetsPatched, []byte(`
func (l *ListPetsResponse) BrokenCount() int {
	if l == nil {
		return 0
	}
	return len(
}
`)...)
	writePatchFixture(t, tempDir, "pkg/models/operations/listpets.go", listPetsOriginal, listPetsPatched)

	cfgBefore, err := config.Load(tempDir)
	require.NoError(t, err)
	cfgBefore.LockFile.Management.DocVersion = "stale-doc-version"
	cfgBefore.LockFile.Management.ReleaseVersion = "stale-release-version"
	cfgBefore.LockFile.Management.GenerationVersion = "stale-generation-version"
	require.NoError(t, config.SaveLockFile(tempDir, cfgBefore.LockFile))

	genYAMLWithMarker := `# compile-failure-marker
go:
  packageName: github.com/example/petstore
generation:
  persistentEdits:
    enabled: true
    compilePristine: false
    patchFiles: true
`
	genYAMLPath := filepath.Join(tempDir, ".speakeasy", "gen.yaml")
	require.NoError(t, os.WriteFile(genYAMLPath, []byte(genYAMLWithMarker), 0o644))

	generator, err = New(WithDebuggingEnabled(), WithGit(mockGit))
	require.NoError(t, err)
	errs := generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, true)
	require.NotEmpty(t, errs)

	errText := errs[0].Error()
	assert.True(t,
		strings.Contains(errText, "compilation") || strings.Contains(errText, "syntax error") || strings.Contains(errText, "lint"),
		"expected compile-related failure, got %q", errText,
	)

	genYAMLAfter, err := os.ReadFile(genYAMLPath)
	require.NoError(t, err)
	assert.NotContains(t, string(genYAMLAfter), "compile-failure-marker")

	genLockAfter, err := os.ReadFile(filepath.Join(tempDir, ".speakeasy", "gen.lock"))
	require.NoError(t, err)
	assert.Contains(t, string(genLockAfter), "docVersion: 1.0.0")
	assert.Contains(t, string(genLockAfter), "releaseVersion:")
	assert.Contains(t, string(genLockAfter), "generationVersion:")
	assert.NotContains(t, string(genLockAfter), "stale-release-version")
	assert.NotContains(t, string(genLockAfter), "stale-generation-version")
}

func newPatchFilesGenerator(outDir string, enabled *config.PersistentEditsEnabled, git merge.Git) *Generator {
	return &Generator{
		outDir: outDir,
		git:    git,
		lockFile: &config.LockFile{
			TrackedFiles: config.NewLockFile().TrackedFiles,
		},
		subsystem: &subsystem.Subsystem{
			FileTracker: nil,
			Patches: &internalpatches.Subsystem{
				Enabled: enabled,
			},
		},
	}
}

func writePatchFixture(t *testing.T, outDir, path string, pristine, patched []byte) {
	t.Helper()
	writePatchFile(t, patchfiles.PathFor(outDir, path), buildUnifiedPatch(t, path, pristine, patched))
}

// writePatchFile writes raw patch bytes to absPath, creating parent dirs as needed.
// absPath should be obtained via patchfiles.PathFor (exact-path patch) or
// filepath.Join(patchfiles.PatchesDir(outDir), name) for generic patches.
func writePatchFile(t *testing.T, absPath string, data []byte) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))
	require.NoError(t, os.WriteFile(absPath, data, 0o644))
}

func buildUnifiedPatch(t *testing.T, path string, pristine, patched []byte) []byte {
	t.Helper()

	repoDir := t.TempDir()
	require.NoError(t, exec.Command("git", "init", "-q", repoDir).Run())

	fullPath := filepath.Join(repoDir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, pristine, 0o644))
	runGit(t, repoDir, "add", "--", path)

	require.NoError(t, os.WriteFile(fullPath, patched, 0o644))
	diff := runGitAllowingExit1(t, repoDir, "diff", "--", path)
	require.NotEmpty(t, diff)

	return diff
}

func buildUnifiedPatchset(t *testing.T, pristine, patched map[string][]byte) []byte {
	t.Helper()

	repoDir := t.TempDir()
	require.NoError(t, exec.Command("git", "init", "-q", repoDir).Run())

	paths := make([]string, 0, len(pristine))
	for path, content := range pristine {
		fullPath := filepath.Join(repoDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, content, 0o644))
		paths = append(paths, path)
	}
	sort.Strings(paths)
	runGit(t, repoDir, "add", "--", ".")

	for path, content := range patched {
		fullPath := filepath.Join(repoDir, path)
		require.NoError(t, os.WriteFile(fullPath, content, 0o644))
	}

	diff := runGitAllowingExit1(t, repoDir, append([]string{"diff", "--"}, paths...)...)
	require.NotEmpty(t, diff)
	return diff
}

func runGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	return output
}

func runGitAllowingExit1(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output
	}

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode(), string(output))
	return output
}

type patchFilesMockLogger struct {
	warns []string
	steps []*patchFilesMockStep
}

func (m *patchFilesMockLogger) Debug(msg string, fields ...zapcore.Field)   {}
func (m *patchFilesMockLogger) Info(msg string, fields ...zapcore.Field)    {}
func (m *patchFilesMockLogger) Error(msg string, fields ...zapcore.Field)   {}
func (m *patchFilesMockLogger) Github(msg string)                           {}
func (m *patchFilesMockLogger) With(fields ...zapcore.Field) logging.Logger { return m }
func (m *patchFilesMockLogger) Scope(name string) logging.Logger            { return m }
func (m *patchFilesMockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.warns = append(m.warns, msg)
}

func (m *patchFilesMockLogger) StartStep(msg string) logging.Step {
	step := &patchFilesMockStep{}
	m.steps = append(m.steps, step)
	return step
}

type patchFilesMockStep struct{}

func (m *patchFilesMockStep) Succeed() {}
func (m *patchFilesMockStep) Fail()    {}
func (m *patchFilesMockStep) Skip()    {}

type patchFilesMockGit struct {
	objects   map[string][]byte
	conflicts map[string]patchFilesConflictState
	repoRoot  string
}

type patchFilesConflictState struct {
	base   []byte
	ours   []byte
	theirs []byte
}

func newPatchFilesMockGit() *patchFilesMockGit {
	return &patchFilesMockGit{
		objects:   map[string][]byte{},
		conflicts: map[string]patchFilesConflictState{},
	}
}

func (m *patchFilesMockGit) addObject(content []byte) string {
	sum := sha1.Sum(content)
	hash := hex.EncodeToString(sum[:])
	m.objects[hash] = append([]byte(nil), content...)
	return hash
}

func (m *patchFilesMockGit) HasObject(hash string) bool {
	_, ok := m.objects[hash]
	return ok
}

func (m *patchFilesMockGit) ReadBlob(hash string) ([]byte, error) {
	content, ok := m.objects[hash]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), content...), nil
}

func (m *patchFilesMockGit) FetchSnapshot(uuid string) error {
	return nil
}

func (m *patchFilesMockGit) RepoRoot() string {
	return m.repoRoot
}

func (m *patchFilesMockGit) WriteObject(content []byte) (string, error) {
	return m.addObject(content), nil
}

func (m *patchFilesMockGit) CreateSnapshotTree(fileHashes map[string]string) (string, error) {
	return "", nil
}

func (m *patchFilesMockGit) CommitSnapshot(treeHash, parentHash, message string) (string, error) {
	return "", nil
}

func (m *patchFilesMockGit) PushSnapshot(commitHash, uuid string) error {
	return nil
}

func (m *patchFilesMockGit) SetConflictState(path string, base, ours, theirs []byte, isExecutable bool) error {
	m.conflicts[path] = patchFilesConflictState{
		base:   append([]byte(nil), base...),
		ours:   append([]byte(nil), ours...),
		theirs: append([]byte(nil), theirs...),
	}
	return nil
}

const testPatchFilesSnapshotSpec = `openapi: 3.1.0
info:
  title: Pet API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
`

const testPatchFilesSnapshotGenYAML = `go:
  packageName: github.com/example/petstore
`

const testPatchFilesSnapshotGenYAMLPersistentEdits = `go:
  packageName: github.com/example/petstore
generation:
  persistentEdits:
    enabled: true
    compilePristine: false
`

func TestGenerate_SkipCompileKeepsFilesListedInLegacyGeneratedFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(testPatchFilesSnapshotGenYAML), 0o644))

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	generator, err := New(WithDebuggingEnabled())
	require.NoError(t, err)
	require.Empty(t, generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, false))

	petPath := filepath.Join(tempDir, "pkg", "models", "shared", "pet.go")
	require.FileExists(t, petPath)

	genLockPath := filepath.Join(tempDir, ".speakeasy", "gen.lock")
	genLock, err := os.ReadFile(genLockPath)
	require.NoError(t, err)
	genLock = append(genLock, []byte("generatedFiles:\n  - /pkg/models/shared/pet.go\n  - /pkg/models/shared/stale.go\n")...)
	require.NoError(t, os.WriteFile(genLockPath, genLock, 0o644))

	stalePath := filepath.Join(tempDir, "pkg", "models", "shared", "stale.go")
	require.NoError(t, os.WriteFile(stalePath, []byte("package shared\n"), 0o644))

	generator, err = New(WithDebuggingEnabled())
	require.NoError(t, err)
	require.Empty(t, generator.Generate(ctx, []byte(testPatchFilesSnapshotSpec), "test-schema.yaml", "go", tempDir, false, false))

	assert.FileExists(t, petPath)
	assert.NoFileExists(t, stalePath)
}
