package env

import (
	"os"
	"strings"
)

// Returns [os.Environ] as a mapping of environment variable names to values.
func GetMap() map[string]string {
	environ := os.Environ()
	result := make(map[string]string, len(environ))

	for _, e := range environ {
		index := strings.Index(e, "=")
		result[e[:index]] = e[index+1:]
	}

	return result
}
