package terraform

import (
	"regexp"
)

var (
	// NOTE: Space characters are intentionally allowed since customers were
	// using them prior to this regular expression validation.
	validResourceNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9 _-]+$`)
)

// Returns true if the given name is valid for a resource name.
func IsResourceNameValid(name string) bool {
	return validResourceNameRegex.MatchString(name)
}
