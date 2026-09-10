package generate

import (
	"os"
	"strings"
	"testing"
)

func requireLicenseToken(t *testing.T) []byte {
	t.Helper()
	token := strings.TrimSpace(os.Getenv("SPEAKEASY_LICENSE_TOKEN"))
	if token == "" {
		t.Skip("SPEAKEASY_LICENSE_TOKEN not set; run ./zero to obtain a license token (commercial-path test skipped)")
	}
	return []byte(token)
}
