package templates

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// The generator compiles declared projections with its own gojq to reject
// them at generation time; the generated CLI must run the same gojq version
// so an expression that passes generation cannot fail at runtime.
func TestCLIGojqPinMatchesGenerator(t *testing.T) {
	goMod, err := os.ReadFile("../go.mod")
	require.NoError(t, err)
	generator := regexp.MustCompile(`github\.com/itchyny/gojq (v[0-9.]+)`).FindSubmatch(goMod)
	require.NotNil(t, generator, "gojq is not a generator dependency")

	config, err := TemplateFS.ReadFile("templates/cli/config.ts")
	require.NoError(t, err)
	cli := regexp.MustCompile(`"github\.com/itchyny/gojq",\s*version: "(v[0-9.]+)"`).FindSubmatch(config)
	require.NotNil(t, cli, "gojq pin not found in templates/cli/config.ts")

	require.Equal(t, string(generator[1]), string(cli[1]), "templates/cli/config.ts must pin the generator's gojq version")
}
