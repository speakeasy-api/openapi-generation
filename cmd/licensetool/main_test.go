package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_ExitCodes(t *testing.T) {
	t.Setenv("SPEAKEASY_LICENSE_TOKEN", "")
	t.Setenv("SPEAKEASY_GENERATED_LICENSE", "")
	envFile := filepath.Join(t.TempDir(), ".env")

	require.Equal(t, exitUsage, run(nil))
	require.Equal(t, exitUsage, run([]string{"bogus"}))
	require.Equal(t, exitUsage, run([]string{"elect", "mit", "--env-file", envFile}))
	require.Equal(t, exitNothing, run([]string{"status", "--env-file", envFile}))

	require.Equal(t, exitOK, run([]string{"elect", "agpl-3.0-only", "--env-file", envFile}))
	require.Equal(t, exitOK, run([]string{"status", "--env-file", envFile}))

	require.NoError(t, os.WriteFile(envFile, []byte("SPEAKEASY_GENERATED_LICENSE=mit\n"), 0o600))
	require.Equal(t, exitUnsupported, run([]string{"status", "--env-file", envFile}))

	require.NoError(t, os.WriteFile(envFile, []byte("SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only\nSPEAKEASY_LICENSE_TOKEN=a.b.c\n"), 0o600))
	require.Equal(t, exitConflict, run([]string{"status", "--env-file", envFile}))

	require.NoError(t, os.WriteFile(envFile, []byte("SPEAKEASY_LICENSE_TOKEN=a.b.c\n"), 0o600))
	require.Equal(t, exitInvalidToken, run([]string{"status", "--env-file", envFile}))
}

func TestDescribe_IncludesTargets(t *testing.T) {
	result := describe(licensetoken.TokenInfo{
		WorkspaceSlug: "acme",
		OrgSlug:       "acme-org",
		Tier:          "free",
		Targets:       []string{"go"},
	})
	assert.Contains(t, result, "tier free, targets go")

	result = describe(licensetoken.TokenInfo{Targets: []string{"*"}})
	assert.Contains(t, result, "targets *")
}

func TestRunFetch_Target(t *testing.T) {
	t.Setenv("SPEAKEASY_API_KEY", "secret-key")

	for _, testCase := range []struct {
		name string
		body string
		want int
	}{
		{
			name: "allowed response reaches token validation",
			body: `{"generation_allowed":true,"level":"allowed","message":"","license_jwt":"h.p.s"}`,
			want: exitInvalidToken,
		},
		{
			name: "blocked response is not entitled",
			body: `{"generation_allowed":false,"level":"blocked","message":"target is not available","license_jwt":null}`,
			want: exitNotEntitled,
		},
		{
			name: "allowed response without token is not entitled",
			body: `{"generation_allowed":true,"level":"allowed","message":"","license_jwt":null}`,
			want: exitNotEntitled,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/workspace/access", r.URL.Path)
				assert.Equal(t, "go", r.URL.Query().Get("targetType"))
				assert.Equal(t, "secret-key", r.Header.Get("x-api-key"))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(testCase.body))
			}))
			t.Cleanup(server.Close)

			envFile := filepath.Join(t.TempDir(), ".env")
			require.Equal(t, testCase.want, run([]string{"fetch", "--env-file", envFile, "--server", server.URL, "--target", "go"}))
		})
	}
}
