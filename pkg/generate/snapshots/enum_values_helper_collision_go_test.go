package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

// An enum's Values helper shares its package with declarations emitted outside
// the model buckets: the SDK-level server constants in sdk.go and the method
// argument option functions in method_options.go. The helper must move aside
// so the existing declaration keeps its name and the package still compiles.
func TestSnapGoEnumValuesHelperCollision(t *testing.T) {
	t.Parallel()

	t.Run("sdk server constant", func(t *testing.T) {
		t.Parallel()
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec: `openapi: 3.0.3
info:
  title: Enum server naming
  version: 1.0.0
servers:
  - url: https://{region}.example.com
    x-speakeasy-server-id: region-values
    variables:
      region:
        default: us
        enum: [us, eu]
paths:
  /status:
    get:
      operationId: getStatus
      responses:
        "204":
          description: No content
`,
			GenYaml: `go:
  packageName: github.com/example/test
`,
			AfterGenerate: func(t *testing.T, tempDir string) {
				t.Helper()
				sdk, err := os.ReadFile(filepath.Join(tempDir, "sdk.go"))
				require.NoError(t, err)
				require.Contains(t, string(sdk), `ServerRegionValues string = "region-values"`)
				require.Contains(t, string(sdk), "func ServerRegionEnumValues() []ServerRegion {")
			},
		})
	})

	t.Run("method argument option", func(t *testing.T) {
		t.Parallel()
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec: `openapi: 3.0.3
info:
  title: Enum option naming
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /status:
    get:
      operationId: get
      parameters:
        - name: values
          in: query
          required: false
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/WithGet"
components:
  schemas:
    WithGet:
      type: string
      enum: [active, inactive]
`,
			GenYaml: `go:
  packageName: github.com/example/test
  optionalMethodArguments: shared-options
  methodArguments: require-security-and-request
  imports:
    option: openapi
    paths:
      callbacks: models/callbacks
      errors: models/apierrors
      operations: models/operations
      shared: models/operations
      webhooks: models/webhooks
`,
			AfterGenerate: func(t *testing.T, tempDir string) {
				t.Helper()
				options, err := os.ReadFile(filepath.Join(tempDir, "models/operations/method_options.go"))
				require.NoError(t, err)
				require.Contains(t, string(options), "func WithGetValues(values string) Option {")
				enum, err := os.ReadFile(filepath.Join(tempDir, "models/operations/withget.go"))
				require.NoError(t, err)
				require.Contains(t, string(enum), "func WithGetEnumValues() []WithGet {")
			},
		})
	})
}
