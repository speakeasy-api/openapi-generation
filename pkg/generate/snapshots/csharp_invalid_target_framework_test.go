package snapshots

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapCsharpInvalidTargetFramework asserts the cross-field validation in
// templates/templates/csharp/config.ts (targetFramework <= dotnetVersion).
// Validation runs during generation so ShouldCompile is disabled for speed.
func TestSnapCsharpInvalidTargetFramework(t *testing.T) {
	t.Parallel()

	shouldCompile := false

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: csharpInvalidTargetFrameworkSpec,
		GenYaml: `csharp:
  packageName: Petstore
  dotnetVersion: net6.0
  targetFramework: net8.0
`,
		ShouldCompile: &shouldCompile,
		ExpectErrors:  []string{"Invalid dotnetVersion: net6.0. Cannot be lower than targetFramework: net8.0."},
	})
}

const csharpInvalidTargetFrameworkSpec = `openapi: 3.1.0
info:
  title: Validate targetFramework <= dotnetVersion
  version: 1.0.0
paths:
  /test:
    get:
      operationId: opID
      responses:
        '200':
          description: OK
`
