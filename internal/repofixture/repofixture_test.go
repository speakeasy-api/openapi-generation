package repofixture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeGoMod(t *testing.T, dir, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644))
}

func detect(t *testing.T, dir string) bool {
	t.Helper()
	ok, err := DetectFixtureOutput(dir)
	require.NoError(t, err)
	return ok
}

func TestDetectFixtureOutput(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "openapi-generation")
	writeGoMod(t, repo, "module "+ModulePath+"\n\ngo 1.25\n")

	t.Run("tracked review fixture directory", func(t *testing.T) {
		assert.True(t, detect(t, filepath.Join(repo, "zSDKs", "sdk-go")))
	})

	t.Run("test SDK directory that does not exist yet", func(t *testing.T) {
		assert.True(t, detect(t, filepath.Join(repo, "testSDKs", "sdk-go-primary", "missing")))
	})

	t.Run("nested module inside a fixture root is not a boundary", func(t *testing.T) {
		nested := filepath.Join(repo, "zSDKs", "sdk-go")
		writeGoMod(t, nested, "module github.com/example/generated-sdk\n")
		assert.True(t, detect(t, nested))
	})

	t.Run("relative path is resolved against the working directory", func(t *testing.T) {
		t.Chdir(repo)
		assert.True(t, detect(t, filepath.Join("zSDKs", "sdk-go")))
	})

	t.Run("go.mod with comments and a quoted module path", func(t *testing.T) {
		quoted := filepath.Join(root, "quoted-checkout")
		writeGoMod(t, quoted, "// generated\nmodule \""+ModulePath+"\" // root\n\ngo 1.25\n")
		assert.True(t, detect(t, filepath.Join(quoted, "testSDKs", "sdk-go-primary")))
	})

	t.Run("checkout root itself is not a fixture", func(t *testing.T) {
		assert.False(t, detect(t, repo))
	})

	t.Run("directories in the checkout outside the fixture roots are stamped", func(t *testing.T) {
		assert.False(t, detect(t, filepath.Join(repo, "pkg", "generate", "out")))
		assert.False(t, detect(t, filepath.Join(repo, "zSDKs-archive", "sdk-go")))
		assert.False(t, detect(t, filepath.Join(repo, "docs", "testSDKs", "sdk-go")))
	})

	t.Run("symlink inside a fixture root pointing outside the checkout", func(t *testing.T) {
		outside := filepath.Join(root, "elsewhere")
		require.NoError(t, os.MkdirAll(outside, 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(repo, "zSDKs"), 0o755))
		require.NoError(t, os.Symlink(outside, filepath.Join(repo, "zSDKs", "link")))
		assert.False(t, detect(t, filepath.Join(repo, "zSDKs", "link", "sdk")))
	})

	t.Run("outside the repository", func(t *testing.T) {
		outside := filepath.Join(root, "customer-sdk")
		writeGoMod(t, outside, "module github.com/example/customer-sdk\n")
		assert.False(t, detect(t, outside))
	})

	t.Run("no go.mod anywhere above", func(t *testing.T) {
		assert.False(t, detect(t, filepath.Join(root, "plain", "dir")))
	})
}
