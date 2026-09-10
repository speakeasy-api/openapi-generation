package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
	"github.com/stretchr/testify/require"
)

func noEnv(string) string { return "" }

func envOf(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func requireTokenAttached(t *testing.T, ctx context.Context) {
	t.Helper()
	_, _, err := licensetoken.ResolveGenerationAccess(ctx)
	require.ErrorIs(t, err, licensetoken.ErrInvalidLicenseToken)

	_, ok := generationaccess.StateFromContext(ctx)
	require.False(t, ok, "factory must not establish access state itself")
}

func TestResolveGenerationContextFactory_NoElectionNoToken(t *testing.T) {
	require.Nil(t, ResolveGenerationContextFactory("", "", noEnv),
		"a nil factory leaves the runner's bare direct context, which the generator refuses")
}

func TestResolveGenerationContextFactory_AGPLElection(t *testing.T) {
	for name, factory := range map[string]func() (context.Context, error){
		"via flag": ResolveGenerationContextFactory("agpl-3.0-only", "", noEnv),
		"via env":  ResolveGenerationContextFactory("", "", envOf(map[string]string{"SPEAKEASY_GENERATED_LICENSE": "agpl-3.0-only"})),
	} {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, factory)
			ctx, err := factory()
			require.NoError(t, err)
			state, ok := generationaccess.StateFromContext(ctx)
			require.True(t, ok)
			require.Equal(t, generationaccess.ModeDirect, state.Mode())
			require.Equal(t, generationaccess.GeneratedLicenseAGPL, state.GeneratedLicense())
		})
	}
}

func TestResolveGenerationContextFactory_EmptyTokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "license.jwt")
	require.NoError(t, os.WriteFile(path, []byte("  \n"), 0o600))

	_, err := ResolveGenerationContextFactory("", path, noEnv)()
	require.ErrorContains(t, err, "is empty")
}

func TestResolveGenerationContextFactory_TokenFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "license.jwt")
	require.NoError(t, os.WriteFile(path, []byte("a.b.c\n"), 0o600))

	ctx, err := ResolveGenerationContextFactory("", path, noEnv)()
	require.NoError(t, err)
	requireTokenAttached(t, ctx)

	_, err = ResolveGenerationContextFactory("", filepath.Join(t.TempDir(), "missing.jwt"), noEnv)()
	require.Error(t, err)
}

func TestResolveGenerationContextFactory_TokenFromEnv(t *testing.T) {
	env := envOf(map[string]string{"SPEAKEASY_LICENSE_TOKEN": "a.b.c"})

	ctx, err := ResolveGenerationContextFactory("", "", env)()
	require.NoError(t, err)
	requireTokenAttached(t, ctx)
}

func TestResolveGenerationContextFactory_CommercialAssertion(t *testing.T) {
	env := envOf(map[string]string{"SPEAKEASY_LICENSE_TOKEN": "a.b.c"})

	ctx, err := ResolveGenerationContextFactory("commercial", "", env)()
	require.NoError(t, err)
	requireTokenAttached(t, ctx)

	_, err = ResolveGenerationContextFactory("commercial", "", noEnv)()
	require.ErrorContains(t, err, "requires a license token")
}

func TestResolveGenerationContextFactory_Rejections(t *testing.T) {
	_, err := ResolveGenerationContextFactory("agpl-3.0-only", "some-path.jwt", noEnv)()
	require.ErrorContains(t, err, "conflicts with a supplied license token")

	_, err = ResolveGenerationContextFactory("agpl-3.0-only", "", envOf(map[string]string{"SPEAKEASY_LICENSE_TOKEN": "a.b.c"}))()
	require.ErrorContains(t, err, "conflicts with a supplied license token")

	_, err = ResolveGenerationContextFactory("MIT", "", noEnv)()
	require.ErrorContains(t, err, `unknown --license value "MIT"`)
}
