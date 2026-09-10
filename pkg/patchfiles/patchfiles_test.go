package patchfiles

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// osFS implements patchfiles.FS using the real filesystem.
type osFS struct{}

func (osFS) Stat(name string) (fs.FileInfo, error)      { return os.Stat(name) }
func (osFS) ReadFile(name string) ([]byte, error)       { return os.ReadFile(name) }
func (osFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }

func TestRead_UsesPatchFileFromDisk(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	patchPath := PathFor(tempDir, "status.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(patchPath), 0o755))
	require.NoError(t, os.WriteFile(patchPath, []byte("diff --git a/status.go b/status.go\n"), 0o644))

	data, used, err := Read(osFS{}, tempDir, "status.go")
	require.NoError(t, err)
	require.True(t, used)
	assert.Equal(t, "diff --git a/status.go b/status.go\n", string(data))
}

func TestRead_NormalizesGeneratedPathLookup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	patchPath := PathFor(tempDir, "models/pet.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(patchPath), 0o755))
	require.NoError(t, os.WriteFile(patchPath, []byte("diff --git a/models/pet.go b/models/pet.go\n"), 0o644))

	data, used, err := Read(osFS{}, tempDir, `models\pet.go`)
	require.NoError(t, err)
	require.True(t, used)
	assert.Equal(t, "diff --git a/models/pet.go b/models/pet.go\n", string(data))
}

func TestExistingPaths_ReturnsNormalizedPaths(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	patchPath := PathFor(tempDir, "models/pet.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(patchPath), 0o755))
	require.NoError(t, os.WriteFile(patchPath, []byte("diff --git a/models/pet.go b/models/pet.go\n"), 0o644))

	paths, err := ExistingPaths(osFS{}, tempDir)
	require.NoError(t, err)
	require.Equal(t, []string{"models/pet.go"}, paths)
}

func TestReadAll_ReturnsNormalizedPatchFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	firstPath := PathFor(tempDir, "models/pet.go")
	secondPath := filepath.Join(tempDir, patchesDir, "delta.patch")
	require.NoError(t, os.MkdirAll(filepath.Dir(firstPath), 0o755))
	require.NoError(t, os.WriteFile(firstPath, []byte("diff --git a/models/pet.go b/models/pet.go\n"), 0o644))
	require.NoError(t, os.WriteFile(secondPath, []byte("diff --git a/status.go b/status.go\n"), 0o644))

	patches, err := ReadAll(osFS{}, tempDir)
	require.NoError(t, err)
	require.Len(t, patches, 2)
	assert.Equal(t, "delta", patches[0].Path)
	assert.Equal(t, "diff --git a/status.go b/status.go\n", string(patches[0].Data))
	assert.Equal(t, "models/pet.go", patches[1].Path)
	assert.Equal(t, "diff --git a/models/pet.go b/models/pet.go\n", string(patches[1].Data))
}

func TestReadAll_ReturnsNilWhenPatchDirDoesNotExist(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	patches, err := ReadAll(osFS{}, tempDir)
	require.NoError(t, err)
	assert.Empty(t, patches)
}

func TestRead_ReturnsNilWhenNoPatchExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	data, used, err := Read(osFS{}, tempDir, "nonexistent.go")
	require.NoError(t, err)
	assert.False(t, used)
	assert.Nil(t, data)
}
