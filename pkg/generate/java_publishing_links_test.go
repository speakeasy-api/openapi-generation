package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const javaPublishingLinksSpec = `openapi: 3.1.0
info:
  title: Acme
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /widgets:
    get:
      operationId: listWidgets
      responses:
        "200":
          description: OK
`

func TestJavaPublishingLinksNormalizeGithubURL(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		githubURL     string
		pomURL        string
		scmURL        string
		scmConnection string
	}{
		"canonical": {
			githubURL:     "github.com/acme/acme-java",
			pomURL:        "https://github.com/acme/acme-java",
			scmURL:        "github.com/acme/acme-java",
			scmConnection: "scm:git:ssh://git@github.com/acme/acme-java.git",
		},
		"scheme, www, .git and trailing slash": {
			githubURL:     "https://www.github.com/acme/acme-java.git/",
			pomURL:        "https://github.com/acme/acme-java",
			scmURL:        "github.com/acme/acme-java",
			scmConnection: "scm:git:ssh://git@github.com/acme/acme-java.git",
		},
		"uppercase scheme, www and .git": {
			githubURL:     "HTTPS://WWW.github.com/acme/acme-java.GIT",
			pomURL:        "https://github.com/acme/acme-java",
			scmURL:        "github.com/acme/acme-java",
			scmConnection: "scm:git:ssh://git@github.com/acme/acme-java.git",
		},
		"non-string value": {
			githubURL:     "123",
			pomURL:        "https://123",
			scmURL:        "123",
			scmConnection: "scm:git:ssh://git@123.git",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			outDir := t.TempDir()
			speakeasyDir := filepath.Join(outDir, ".speakeasy")
			require.NoError(t, os.MkdirAll(speakeasyDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(speakeasyDir, "gen.yaml"), []byte(`configVersion: 2.0.0
generation:
  sdkClassName: Acme
java:
  version: 0.0.1
  groupID: com.acme
  artifactID: acme
  githubURL: `+tc.githubURL+`
`), 0o644))

			g, err := New(WithPublished(true))
			require.NoError(t, err)

			errs := g.Generate(
				generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()),
				[]byte(javaPublishingLinksSpec),
				"openapi.yaml",
				"java",
				outDir,
				false,
				false,
			)
			require.Empty(t, errs)

			publishing, err := os.ReadFile(filepath.Join(outDir, "publishing.gradle"))
			require.NoError(t, err)

			assert.Contains(t, string(publishing), "                    url = '"+tc.pomURL+"'\n")
			assert.Contains(t, string(publishing), "                        url = '"+tc.scmURL+"'\n")
			assert.Contains(t, string(publishing), "                        connection = '"+tc.scmConnection+"'\n")
		})
	}
}
