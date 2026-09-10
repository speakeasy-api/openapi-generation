package snapshots

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/stretchr/testify/require"
)

// TestSnapPyHTTPClientLibrary verifies the httpClientLibrary gen.yaml option:
// the default keeps the SDK on httpx, while httpx2 swaps the dependency and
// aliases every import so the rest of the SDK is unchanged.
func TestSnapPyHTTPClientLibrary(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: HTTP Client Library Test
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /things/{id}:
    get:
      operationId: getThing
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
`

	generateSDK := func(t *testing.T, genYaml string) string {
		t.Helper()

		ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

		generator, err := generate.New(generate.WithDebuggingEnabled())
		require.NoError(t, err)

		errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDir, false, false)
		require.Empty(t, errs)

		return tempDir
	}

	readGenerated := func(t *testing.T, dir string, parts ...string) string {
		t.Helper()

		data, err := os.ReadFile(filepath.Join(append([]string{dir}, parts...)...))
		require.NoError(t, err)
		return string(data)
	}

	t.Run("DefaultRemainsHttpx", func(t *testing.T) {
		t.Parallel()

		dir := generateSDK(t, `python:
  packageName: clientlib
  imports:
    option: openapi
    paths:
      callbacks: ""
      errors: ""
      operations: ""
      shared: ""
      webhooks: ""
`)

		pyproject := readGenerated(t, dir, "pyproject.toml")
		require.Contains(t, pyproject, `"httpx >=`)
		require.Contains(t, pyproject, `"httpcore >=`)
		require.NotContains(t, pyproject, "httpx2")

		basesdk := readGenerated(t, dir, "src", "clientlib", "basesdk.py")
		require.Contains(t, basesdk, "import httpx")
		require.NotContains(t, basesdk, "import httpx2")

		values := readGenerated(t, dir, "src", "clientlib", "utils", "values.py")
		require.Contains(t, values, "from httpx import Response")

		readme := readGenerated(t, dir, "README.md")
		require.Contains(t, readme, "### httpx2 (Pydantic's httpx fork)")
		require.Contains(t, readme, "alias_httpx()")
		require.Contains(t, readme, "python.httpClientLibrary: httpx2")
	})

	t.Run("Httpx2SwapsDependencyAndAliasesImports", func(t *testing.T) {
		t.Parallel()

		dir := generateSDK(t, `python:
  packageName: clientlib
  httpClientLibrary: httpx2
  imports:
    option: openapi
    paths:
      callbacks: ""
      errors: ""
      operations: ""
      shared: ""
      webhooks: ""
`)

		pyproject := readGenerated(t, dir, "pyproject.toml")
		require.Contains(t, pyproject, `"httpx2 >=`)
		require.NotContains(t, pyproject, `"httpx >=`)
		require.NotContains(t, pyproject, "httpcore")

		basesdk := readGenerated(t, dir, "src", "clientlib", "basesdk.py")
		require.Contains(t, basesdk, "import httpx2 as httpx")

		httpclient := readGenerated(t, dir, "src", "clientlib", "httpclient.py")
		require.Contains(t, httpclient, "import httpx2 as httpx")

		values := readGenerated(t, dir, "src", "clientlib", "utils", "values.py")
		require.Contains(t, values, "from httpx2 import Response")

		headers := readGenerated(t, dir, "src", "clientlib", "utils", "headers.py")
		require.Contains(t, headers, "from httpx2 import Headers")

		readme := readGenerated(t, dir, "README.md")
		require.Contains(t, readme, "This SDK is generated against [httpx2]")
		// The alias route is for SDKs still on httpx; it would be misleading here.
		require.NotContains(t, readme, "alias_httpx()")
	})
}
