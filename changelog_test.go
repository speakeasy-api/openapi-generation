package openapigeneration

import (
	"os"
	"regexp"
	"testing"
)

func TestGetLatestVersion(t *testing.T) {
	actual := GetLatestVersion()
	if expected := os.Getenv("EXPECTED_RELEASE_VERSION"); expected != "" && actual != expected {
		t.Fatalf("GetLatestVersion() = %q, want release version %q", actual, expected)
	}
	if !regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(actual) {
		t.Fatalf("GetLatestVersion() = %q, want a semantic version", actual)
	}
	t.Logf("GetLatestVersion() = %q", actual)
}
