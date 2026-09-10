package filetracking

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIgnore_ShouldIgnore(t *testing.T) {
	expect := map[string]bool{
		".genignore":           false,
		"bar.txt":              false,
		"foo.txt":              true,
		"baz/.genignore":       false,
		"baz/fred.txt":         true,
		"baz/quux.txt":         true,
		"baz/qux.txt":          false,
		"baz/corge":            true,
		"baz/corge/.genignore": true,
		"baz/corge/grault.txt": true,
	}
	ignore, err := NewIgnoreFromRoot("testdata")
	require.NoError(t, err)
	require.NotNil(t, ignore)

	err = filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel("testdata", path)

		if err != nil {
			return err
		}

		assert.Equalf(t, expect[rel], ignore.ShouldIgnore(rel), "ShouldIgnore %s", rel)

		return nil
	})

	require.NoError(t, err)
}

// TestIgnore_SkipsInaccessibleDir ensures an unreadable directory under the
// root does not abort the walk with a permission error.
func TestIgnore_SkipsInaccessibleDir(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission bits not enforced")
	}

	root := t.TempDir()
	unreadable := filepath.Join(root, "unreadable")
	require.NoError(t, os.Mkdir(unreadable, 0o000))
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o700) })

	ignore, err := NewIgnoreFromRoot(root)
	require.NoError(t, err)
	require.NotNil(t, ignore)
}

// TestIgnore_ShouldIgnore_OSPaths tests that OS-native paths (which use
// backslashes on Windows) are correctly normalized by ShouldIgnore to match
// gitignore patterns (which always use forward slashes).
func TestIgnore_ShouldIgnore_OSPaths(t *testing.T) {
	ignore, err := NewIgnoreFromRoot("testdata")
	require.NoError(t, err)
	require.NotNil(t, ignore)

	// filepath.FromSlash converts "/" to the OS-native separator.
	// On Windows: "baz/fred.txt" -> "baz\fred.txt"
	// On Unix: "baz/fred.txt" -> "baz/fred.txt" (unchanged)
	// This simulates paths produced by filepath.Join on each OS.
	p := filepath.FromSlash

	testCases := []struct {
		path     string
		expected bool
	}{
		{p("baz/fred.txt"), true},         // ignored by baz/.genignore
		{p("baz/qux.txt"), false},         // not ignored
		{p("baz/corge"), true},            // ignored directory
		{p("baz/corge/.genignore"), true}, // inside ignored directory
		{p("baz/corge/grault.txt"), true}, // inside ignored directory
	}

	for _, tc := range testCases {
		result := ignore.ShouldIgnore(tc.path)
		assert.Equalf(t, tc.expected, result, "ShouldIgnore(%q)", tc.path)
	}
}
