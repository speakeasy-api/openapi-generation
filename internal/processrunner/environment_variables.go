package processrunner

import (
	"strings"
)

// Converts environment variables in [os.Environ] form to a mapping of
// environment names to values.
func environMap(environ []string) map[string]string {
	result := make(map[string]string, len(environ))

	for _, e := range environ {
		lastIdx := strings.Index(e, "=")
		result[e[:lastIdx]] = e[lastIdx+1:]
	}

	return result
}

// Converts environment variables in mapping of environment names to values form
// to [os.Environ] form.
func environSlice(envVars map[string]string) []string {
	result := make([]string, 0, len(envVars))

	for name, value := range envVars {
		result = append(result, name+"="+value)
	}

	return result
}
