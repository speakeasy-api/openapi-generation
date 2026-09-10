package processrunner

import "strings"

const licenseTokenVariable = "SPEAKEASY_LICENSE_TOKEN="

func childEnvironment(environ []string) []string {
	kept := make([]string, 0, len(environ))
	for _, entry := range environ {
		if strings.HasPrefix(entry, licenseTokenVariable) {
			continue
		}
		kept = append(kept, entry)
	}
	return kept
}
