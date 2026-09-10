package openapi

import "strings"

type Version string

const (
	Version2  Version = "2"
	Version3  Version = "3"
	Version31 Version = "3.1"
)

func DetermineOpenAPIVersion(version string) Version {
	if strings.HasPrefix(version, "2") {
		return Version2
	}
	if strings.HasPrefix(version, "3.1") {
		return Version31
	}
	return Version3
}
