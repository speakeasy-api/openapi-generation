package generate

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/templates"
	"github.com/stretchr/testify/require"
)

func selectedGojqVersion(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "list", "-m", "-f", "{{.Version}}", "github.com/itchyny/gojq")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return strings.TrimSpace(string(out))
}

// The generator compiles declared projections with its own gojq to reject
// them at generation time; the generated CLI must run the same gojq version
// so an expression that passes generation cannot fail at runtime.
func TestCLIGojqPinMatchesGenerator(t *testing.T) {
	generator := selectedGojqVersion(t, "../..")
	generatedCLI := selectedGojqVersion(t, "../../zSDKs/sdk-cli")

	config, err := templates.TemplateFS.ReadFile("templates/cli/config.ts")
	require.NoError(t, err)
	cli := regexp.MustCompile(`"github\.com/itchyny/gojq",\s*version: "([^"]+)"`).FindSubmatch(config)
	require.NotNil(t, cli, "gojq pin not found in templates/cli/config.ts")

	require.Equal(t, generator, string(cli[1]), "templates/cli/config.ts must pin the generator's selected gojq version")
	require.Equal(t, generator, generatedCLI, "the generated CLI fixture must select the generator's gojq version")
}
