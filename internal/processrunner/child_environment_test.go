package processrunner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChildEnvironmentStripsLicenseToken(t *testing.T) {
	got := childEnvironment([]string{"PATH=/bin", "SPEAKEASY_LICENSE_TOKEN=a.b.c", "SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only"})
	assert.Equal(t, []string{"PATH=/bin", "SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only"}, got)
}

func TestRun_StripsLicenseTokenFromCommandEnvironment(t *testing.T) {
	t.Setenv("SPEAKEASY_LICENSE_TOKEN", "parent")
	cmd := Command{
		Command:              "sh",
		Args:                 []string{"-c", `echo "${SPEAKEASY_LICENSE_TOKEN:-absent}"`},
		EnvironmentVariables: map[string]string{"SPEAKEASY_LICENSE_TOKEN": "override"},
	}

	stdout, _, err := cmd.run(testContext(), t.TempDir(), nil)
	require.NoError(t, err)
	assert.Equal(t, "absent", strings.TrimSpace(stdout))
}
