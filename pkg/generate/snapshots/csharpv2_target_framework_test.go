package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

// TestSnapCsharpV2TargetFramework asserts that templateVersion v2 emits a
// net* targetFramework below dotnetVersion as the SDK TargetFramework, with
// the LangVersion of dotnetVersion, and targets dotnetVersion alone without a
// LangVersion when targetFramework is unset.
func TestSnapCsharpV2TargetFramework(t *testing.T) {
	t.Parallel()

	shouldCompile := false

	cases := []struct {
		name                string
		targetFramework     string
		expectedFramework   string
		expectedLangVersion string
	}{
		{
			name:                "below dotnetVersion",
			targetFramework:     "\n  targetFramework: net8.0",
			expectedFramework:   "net8.0",
			expectedLangVersion: "14.0",
		},
		{
			name:              "unset",
			expectedFramework: "net10.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec: csharpV2TargetFrameworkSpec,
				GenYaml: `csharp:
  templateVersion: v2
  packageName: Petstore
  dotnetVersion: net10.0` + tc.targetFramework + `
`,
				ShouldCompile: &shouldCompile,
				AfterGenerate: func(t *testing.T, tempDir string) {
					t.Helper()
					csproj, err := os.ReadFile(filepath.Join(tempDir, "src", "Petstore", "Petstore.csproj"))
					require.NoError(t, err)
					require.Contains(t, string(csproj), "<TargetFramework>"+tc.expectedFramework+"</TargetFramework>")
					if tc.expectedLangVersion == "" {
						require.NotContains(t, string(csproj), "<LangVersion>")
					} else {
						require.Contains(t, string(csproj), "<LangVersion>"+tc.expectedLangVersion+"</LangVersion>")
					}
					require.Contains(t, string(csproj), "<IsAotCompatible>true</IsAotCompatible>")
					require.NotContains(t, string(csproj), "<TargetFrameworks>")
				},
			})
		})
	}
}

// TestSnapCsharpV2InvalidTargetFramework asserts that templateVersion v2
// rejects a targetFramework below net8.0 or above dotnetVersion.
func TestSnapCsharpV2InvalidTargetFramework(t *testing.T) {
	t.Parallel()

	shouldCompile := false

	cases := []struct {
		name            string
		dotnetVersion   string
		targetFramework string
		expectedError   string
	}{
		{
			name:            "below net8.0",
			dotnetVersion:   "net10.0",
			targetFramework: "net6.0",
			expectedError:   "targetFramework: net6.0 incompatible with templateVersion: v2. netstandard2.0, net8.0, net10.0 or dotnetVersion required.",
		},
		{
			name:            "above dotnetVersion",
			dotnetVersion:   "net8.0",
			targetFramework: "net10.0",
			expectedError:   "Invalid dotnetVersion: net8.0. Cannot be lower than targetFramework: net10.0.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec: csharpV2TargetFrameworkSpec,
				GenYaml: `csharp:
  templateVersion: v2
  packageName: Petstore
  dotnetVersion: ` + tc.dotnetVersion + `
  targetFramework: ` + tc.targetFramework + `
`,
				ShouldCompile: &shouldCompile,
				ExpectErrors:  []string{tc.expectedError},
			})
		})
	}
}

const csharpV2TargetFrameworkSpec = `openapi: 3.1.0
info:
  title: Validate templateVersion v2 targetFramework
  version: 1.0.0
paths:
  /test:
    get:
      operationId: opID
      responses:
        '200':
          description: OK
`
