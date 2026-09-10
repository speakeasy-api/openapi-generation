package licensebootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAPIKey(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".speakeasy"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".speakeasy", "config.yaml"), []byte("speakeasy_api_key: from-config\nspeakeasy_workspace_id: ws\n"), 0o600))

	t.Run("environment wins", func(t *testing.T) {
		key, err := ResolveAPIKey(func(name string) string {
			if name == "SPEAKEASY_API_KEY" {
				return "from-env"
			}
			return ""
		}, home)
		require.NoError(t, err)
		assert.Equal(t, "from-env", key)
	})

	t.Run("falls back to the CLI config", func(t *testing.T) {
		key, err := ResolveAPIKey(func(string) string { return "" }, home)
		require.NoError(t, err)
		assert.Equal(t, "from-config", key)
	})

	t.Run("no key anywhere", func(t *testing.T) {
		_, err := ResolveAPIKey(func(string) string { return "" }, t.TempDir())
		require.ErrorIs(t, err, ErrNoAPIKey)
	})

	t.Run("config without the key", func(t *testing.T) {
		emptyHome := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(emptyHome, ".speakeasy"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(emptyHome, ".speakeasy", "config.yaml"), []byte("speakeasy_workspace_id: ws\n"), 0o600))
		_, err := ResolveAPIKey(func(string) string { return "" }, emptyHome)
		require.ErrorIs(t, err, ErrNoAPIKey)
	})
}

func TestResolveServerURL(t *testing.T) {
	noEnv := func(string) string { return "" }
	withEnv := func(name string) string {
		if name == "SPEAKEASY_SERVER_URL" {
			return "https://env.example/"
		}
		return ""
	}

	assert.Equal(t, DefaultServerURL, ResolveServerURL("", noEnv))
	assert.Equal(t, "https://env.example", ResolveServerURL("", withEnv), "trailing slash trimmed")
	assert.Equal(t, "https://flag.example", ResolveServerURL("https://flag.example", withEnv), "flag wins over the environment")
}
