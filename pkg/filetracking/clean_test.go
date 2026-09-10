package filetracking

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCleanOrphanedFilesByDirWalk_DeletesUntrackedFiles verifies the core
// behavior: a file present on disk but not in the tracked set is removed,
// while a tracked file is left alone.
func TestCleanOrphanedFilesByDirWalk_DeletesUntrackedFiles(t *testing.T) {
	outDir := t.TempDir()
	subDir := "src/__tests__/mockserver"

	tracked := writeFile(t, outDir, subDir, "internal/sdk/models/components/aftertype.go")
	orphan := writeFile(t, outDir, subDir, "internal/sdk/models/components/accessgroup1.go")

	var removed []string
	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{
			filepath.ToSlash(filepath.Join(subDir, "internal/sdk/models/components/aftertype.go")): {},
		},
		nil,
		nil,
		NewIgnore(),
		func(p string) { removed = append(removed, p) },
	)
	require.NoError(t, err)

	assert.FileExists(t, tracked)
	assert.NoFileExists(t, orphan)

	sort.Strings(removed)
	assert.Equal(t, []string{
		filepath.ToSlash(filepath.Join(subDir, "internal/sdk/models/components/accessgroup1.go")),
	}, removed)
}

// TestCleanOrphanedFilesByDirWalk_HonorsGenIgnore verifies that customer
// .genignore'd files are preserved even when not in the tracked set. This is
// the documented escape hatch for hand-edited mock server content.
func TestCleanOrphanedFilesByDirWalk_HonorsGenIgnore(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	custom := writeFile(t, outDir, subDir, "custom_handler.go")

	require.NoError(t, os.WriteFile(filepath.Join(outDir, ".genignore"), []byte("mockserver/custom_handler.go\n"), 0o644))

	ignore, err := NewIgnoreFromRoot(outDir)
	require.NoError(t, err)

	err = CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{},
		nil,
		nil,
		ignore,
		func(string) {},
	)
	require.NoError(t, err)

	assert.FileExists(t, custom)
}

// TestCleanOrphanedFilesByDirWalk_HonorsUntrackedPatterns verifies the regex
// allowlist used by exclusions.ts also keeps matching files in place.
func TestCleanOrphanedFilesByDirWalk_HonorsUntrackedPatterns(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	allowlisted := writeFile(t, outDir, subDir, "internal/hooks/registration.go")

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{},
		nil,
		[]*regexp.Regexp{regexp.MustCompile(`mockserver/internal/hooks/registration\.go$`)},
		NewIgnore(),
		func(string) {},
	)
	require.NoError(t, err)

	assert.FileExists(t, allowlisted)
}

// TestCleanOrphanedFilesByDirWalk_MissingSubDir is a no-op rather than an
// error. Mock server may not be enabled or may not have run yet.
func TestCleanOrphanedFilesByDirWalk_MissingSubDir(t *testing.T) {
	outDir := t.TempDir()

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		"does/not/exist",
		map[string]struct{}{},
		nil,
		nil,
		NewIgnore(),
		func(string) {},
	)
	assert.NoError(t, err)
}

// TestCleanOrphanedFilesByDirWalk_IgnoresOtherSubtrees ensures the walk is
// strictly scoped to subDir; SDK files outside the mock server tree must not
// be touched even when they're absent from the tracked set passed here (the
// caller intentionally only passes the mock server's tracked set).
func TestCleanOrphanedFilesByDirWalk_IgnoresOtherSubtrees(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	sdkFile := writeFile(t, outDir, "src/sdk", "main.go")
	mockOrphan := writeFile(t, outDir, subDir, "orphan.go")

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{},
		nil,
		nil,
		NewIgnore(),
		func(string) {},
	)
	require.NoError(t, err)

	assert.FileExists(t, sdkFile)
	assert.NoFileExists(t, mockOrphan)
}

// TestCleanOrphanedFilesByDirWalk_SkipsListedSubDirs covers the testdata
// regression: the mock server's testdata/ subtree is populated by a direct
// cp.Copy that bypasses FileTracker. Without skipSubDirs the walk would treat
// every fixture as an orphan and delete it.
func TestCleanOrphanedFilesByDirWalk_SkipsListedSubDirs(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	fixture := writeFile(t, outDir, subDir, "testdata/example.file")
	orphanOutsideSkip := writeFile(t, outDir, subDir, "internal/orphan.go")

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{},
		[]string{"testdata"},
		nil,
		NewIgnore(),
		func(string) {},
	)
	require.NoError(t, err)

	assert.FileExists(t, fixture)
	assert.NoFileExists(t, orphanOutsideSkip)
}

// TestCleanOrphanedFilesByDirWalk_PrunesEmptyDirs verifies that directories
// left empty after the file pass are removed (bottom-up), but the root subDir
// itself and any skipped directory are preserved.
func TestCleanOrphanedFilesByDirWalk_PrunesEmptyDirs(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	// Sole occupant of a deep subtree; deletion should empty all the parents.
	deepOrphan := writeFile(t, outDir, subDir, "internal/foo/bar/orphan.go")
	// Sibling subtree with a tracked file — should survive intact.
	tracked := writeFile(t, outDir, subDir, "internal/keep/keeper.go")
	// Empty skipped dir — must not be removed even though it has no contents.
	skippedRoot := filepath.Join(outDir, subDir, "testdata")
	require.NoError(t, os.MkdirAll(skippedRoot, 0o755))

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{
			filepath.ToSlash(filepath.Join(subDir, "internal/keep/keeper.go")): {},
		},
		[]string{"testdata"},
		nil,
		NewIgnore(),
		func(string) {},
	)
	require.NoError(t, err)

	assert.NoFileExists(t, deepOrphan)
	assert.NoDirExists(t, filepath.Join(outDir, subDir, "internal/foo/bar"))
	assert.NoDirExists(t, filepath.Join(outDir, subDir, "internal/foo"))
	assert.FileExists(t, tracked)
	assert.DirExists(t, filepath.Join(outDir, subDir, "internal/keep"))
	// Skipped dir preserved despite being empty.
	assert.DirExists(t, skippedRoot)
	// Root subDir itself is preserved.
	assert.DirExists(t, filepath.Join(outDir, subDir))
}

// TestCleanOrphanedFilesByDirWalk_NilIgnore verifies the function handles a
// nil *Ignore without panicking. Some callers may have no .genignore loaded.
func TestCleanOrphanedFilesByDirWalk_NilIgnore(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	orphan := writeFile(t, outDir, subDir, "orphan.go")

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{},
		nil,
		nil,
		nil, // ignore
		func(string) {},
	)
	require.NoError(t, err)

	assert.NoFileExists(t, orphan)
}

// TestCleanOrphanedFilesByDirWalk_ContextCancelled verifies the walk aborts
// promptly when the supplied context is cancelled before invocation.
func TestCleanOrphanedFilesByDirWalk_ContextCancelled(t *testing.T) {
	outDir := t.TempDir()
	subDir := "mockserver"

	orphan := writeFile(t, outDir, subDir, "orphan.go")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CleanOrphanedFilesByDirWalk(
		ctx,
		outDir,
		subDir,
		map[string]struct{}{},
		nil,
		nil,
		NewIgnore(),
		func(string) {},
	)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)

	// Orphan must not have been deleted — cancellation should preempt the walk.
	assert.FileExists(t, orphan)
}

// TestCleanOrphanedFilesByDirWalk_SkipsInaccessibleDir verifies an unreadable
// directory within the walked tree does not abort cleanup with a permission
// error. Tracked files elsewhere in the tree are still preserved.
func TestCleanOrphanedFilesByDirWalk_SkipsInaccessibleDir(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission bits not enforced")
	}

	outDir := t.TempDir()
	subDir := "mockserver"

	tracked := writeFile(t, outDir, subDir, "keep.go")

	unreadable := filepath.Join(outDir, subDir, "locked")
	require.NoError(t, os.Mkdir(unreadable, 0o000))
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o700) })

	err := CleanOrphanedFilesByDirWalk(
		context.Background(),
		outDir,
		subDir,
		map[string]struct{}{
			filepath.ToSlash(filepath.Join(subDir, "keep.go")): {},
		},
		nil,
		nil,
		NewIgnore(),
		func(string) {},
	)
	require.NoError(t, err)
	assert.FileExists(t, tracked)
}

func writeFile(t *testing.T, outDir, subDir, rel string) string {
	t.Helper()
	abs := filepath.Join(outDir, subDir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0o755))
	require.NoError(t, os.WriteFile(abs, []byte("// stub"), 0o644))
	return abs
}

// TestCleanOrphanedFiles_MatchesLegacyLeadingSlashPaths verifies that lockfile
// entries recorded with a leading separator (the pre-trackedFiles format kept
// template output paths verbatim) are matched against the current run's files
// instead of being treated as orphans.
func TestCleanOrphanedFiles_MatchesLegacyLeadingSlashPaths(t *testing.T) {
	outDir := t.TempDir()

	kept := writeFile(t, outDir, "", "models/pet.go")
	orphan := writeFile(t, outDir, "", "models/old.go")

	var removed []string
	err := CleanOrphanedFiles(
		context.Background(),
		outDir,
		[]string{"models/pet.go"},
		[]string{"/models/pet.go", "/models/old.go"},
		nil,
		NewIgnore(),
		func(p string) { removed = append(removed, p) },
	)
	require.NoError(t, err)

	assert.FileExists(t, kept)
	assert.NoFileExists(t, orphan)
	assert.Equal(t, []string{"models/old.go"}, removed)
}

func TestNormalizePath(t *testing.T) {
	cases := map[string]string{
		"models/pet.go":           "models/pet.go",
		"/models/pet.go":          "models/pet.go",
		"//models/pet.go":         "models/pet.go",
		"./models/pet.go":         "models/pet.go",
		"models//pet.go":          "models/pet.go",
		"models/./pet.go":         "models/pet.go",
		"models\\pet.go":          "models/pet.go",
		"\\models\\pet.go":        "models/pet.go",
		"./":                      "",
		"":                        "",
		"models/../shared/pet.go": "shared/pet.go",
	}
	for in, want := range cases {
		assert.Equal(t, want, NormalizePath(in), "NormalizePath(%q)", in)
	}
}

func TestCleanOrphanedFiles_MatchesLegacyPathSpellings(t *testing.T) {
	outDir := t.TempDir()

	kept := []string{
		writeFile(t, outDir, "", "models/a.go"),
		writeFile(t, outDir, "", "models/b.go"),
		writeFile(t, outDir, "", "models/c.go"),
		writeFile(t, outDir, "", "models/d.go"),
	}
	orphan := writeFile(t, outDir, "", "models/old.go")

	var removed []string
	err := CleanOrphanedFiles(
		context.Background(),
		outDir,
		[]string{"models/a.go", "models/b.go", "models/c.go", "models/d.go"},
		[]string{"./models/a.go", "models//b.go", "\\models\\c.go", "/./models/d.go", "models/old.go"},
		nil,
		NewIgnore(),
		func(p string) { removed = append(removed, p) },
	)
	require.NoError(t, err)

	for _, f := range kept {
		assert.FileExists(t, f)
	}
	assert.NoFileExists(t, orphan)
	assert.Equal(t, []string{"models/old.go"}, removed)
}

func TestCleanOrphanedFiles_LegacyLeadingSlashEntryHonorsIgnoreAndUntracked(t *testing.T) {
	outDir := t.TempDir()

	ignored := writeFile(t, outDir, "", "custom/keep.go")
	untracked := writeFile(t, outDir, "", "vendor/dep.go")
	orphan := writeFile(t, outDir, "", "models/old.go")

	require.NoError(t, os.WriteFile(filepath.Join(outDir, ".genignore"), []byte("custom/keep.go\n"), 0o644))
	ignore, err := NewIgnoreFromRoot(outDir)
	require.NoError(t, err)

	var removed []string
	err = CleanOrphanedFiles(
		context.Background(),
		outDir,
		nil,
		[]string{"/custom/keep.go", "/vendor/dep.go", "/models/old.go"},
		[]*regexp.Regexp{regexp.MustCompile(`^vendor/`)},
		ignore,
		func(p string) { removed = append(removed, p) },
	)
	require.NoError(t, err)

	assert.FileExists(t, ignored, "ignore rules must see the normalized path")
	assert.FileExists(t, untracked, "untracked patterns must see the normalized path")
	assert.NoFileExists(t, orphan)
	assert.Equal(t, []string{"models/old.go"}, removed)
}
