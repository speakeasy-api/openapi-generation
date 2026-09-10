package licensebootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envTemplate = "# Local license election (see ./zero)\n#SPEAKEASY_LICENSE_TOKEN=\n\n# Open source\n#SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n"

func TestSetEnvValue(t *testing.T) {
	t.Run("replaces the commented template line in place", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".env")
		require.NoError(t, os.WriteFile(path, []byte(envTemplate), 0o600))

		require.NoError(t, SetEnvValue(path, "SPEAKEASY_LICENSE_TOKEN", "a.b.c"))

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "# Local license election (see ./zero)\nSPEAKEASY_LICENSE_TOKEN=a.b.c\n\n# Open source\n#SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n", string(content))
	})

	t.Run("replaces an active line", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".env")
		require.NoError(t, os.WriteFile(path, []byte("SPEAKEASY_LICENSE_TOKEN=old\nOTHER=1\n"), 0o600))

		require.NoError(t, SetEnvValue(path, "SPEAKEASY_LICENSE_TOKEN", "new"))

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "SPEAKEASY_LICENSE_TOKEN=new\nOTHER=1\n", string(content))
	})

	t.Run("collapses duplicate active lines", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".env")
		require.NoError(t, os.WriteFile(path, []byte("SPEAKEASY_LICENSE_TOKEN=old\nOTHER=1\nSPEAKEASY_LICENSE_TOKEN=older\n"), 0o600))

		require.NoError(t, SetEnvValue(path, "SPEAKEASY_LICENSE_TOKEN", "new"))

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "SPEAKEASY_LICENSE_TOKEN=new\nOTHER=1\n", string(content))
	})

	t.Run("appends when absent and creates the file with mode 0600", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".env")

		require.NoError(t, SetEnvValue(path, "SPEAKEASY_GENERATED_LICENSE", "agpl-3.0-only"))

		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n", string(content))
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
}

func TestWriteElection(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(path, []byte(envTemplate), 0o600))

	require.NoError(t, WriteElection(path, "SPEAKEASY_LICENSE_TOKEN", "a.b.c", "SPEAKEASY_GENERATED_LICENSE"))
	require.NoError(t, WriteElection(path, "SPEAKEASY_GENERATED_LICENSE", "agpl-3.0-only", "SPEAKEASY_LICENSE_TOKEN"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "# Local license election (see ./zero)\n\n# Open source\nSPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n", string(content))
}

func TestUnsetEnvValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(path, []byte("#SPEAKEASY_LICENSE_TOKEN=\nSPEAKEASY_LICENSE_TOKEN=a.b.c\nOTHER=1\n"), 0o600))

	require.NoError(t, UnsetEnvValue(path, "SPEAKEASY_LICENSE_TOKEN"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "#SPEAKEASY_LICENSE_TOKEN=\nOTHER=1\n", string(content))

	require.NoError(t, UnsetEnvValue(filepath.Join(t.TempDir(), "missing"), "X"), "missing file is a no-op")
}

func TestEnvFileMustBeRegular(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "elsewhere")
	require.NoError(t, os.WriteFile(target, []byte("SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n"), 0o600))
	link := filepath.Join(dir, ".env")
	require.NoError(t, os.Symlink(target, link))

	_, err := ReadEnvValue(link, "SPEAKEASY_GENERATED_LICENSE")
	require.ErrorContains(t, err, "not a regular file")
	require.ErrorContains(t, SetEnvValue(link, "SPEAKEASY_LICENSE_TOKEN", "a.b.c"), "not a regular file")
	require.ErrorContains(t, UnsetEnvValue(link, "SPEAKEASY_GENERATED_LICENSE"), "not a regular file")

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\n", string(content), "link target untouched")
}

func TestReadEnvValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(path, []byte("#SPEAKEASY_LICENSE_TOKEN=commented\nSPEAKEASY_GENERATED_LICENSE = agpl-3.0-only\nQUOTED=\"x.y.z\"\n"), 0o600))

	value, err := ReadEnvValue(path, "SPEAKEASY_LICENSE_TOKEN")
	require.NoError(t, err)
	assert.Empty(t, value, "commented lines are not values")

	value, err = ReadEnvValue(path, "SPEAKEASY_GENERATED_LICENSE")
	require.NoError(t, err)
	assert.Equal(t, "agpl-3.0-only", value, "surrounding whitespace is trimmed")

	value, err = ReadEnvValue(path, "QUOTED")
	require.NoError(t, err)
	assert.Equal(t, "x.y.z", value, "matching double quotes are stripped")

	value, err = ReadEnvValue(filepath.Join(t.TempDir(), "missing"), "X")
	require.NoError(t, err)
	assert.Empty(t, value)

	dup := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(dup, []byte("K=first\nK=last\n"), 0o600))
	value, err = ReadEnvValue(dup, "K")
	require.NoError(t, err)
	assert.Equal(t, "last", value, "the last active line wins, as in mise")
}
