package changes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChanges_PersistentEditsEnabled_WithoutGit(t *testing.T) {
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      responses:
        '200':
          description: Success`

	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: spec,
		newSpec: spec,
		lang:    "typescript",
	})

	// Write gen.yaml with persistentEdits enabled into both output dirs.
	// resolveAST creates a generator without Git, so this would previously
	// cause initInternal to hard-error.
	genYAML := `configVersion: 2.0.0
generation:
  sdkClassName: SDK
  persistentEdits:
    enabled: "true"
typescript:
  version: 0.0.1
  author: test
  clientServerStatusCodesAsErrors: true
  flattenGlobalSecurity: true
  maxMethodParams: 4
  packageName: test-sdk
`
	for _, dir := range []string{oldConfig.OutDir, newConfig.OutDir} {
		speakeasyDir := filepath.Join(dir, ".speakeasy")
		if err := os.MkdirAll(speakeasyDir, 0o755); err != nil {
			t.Fatalf("failed to create .speakeasy dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(speakeasyDir, "gen.yaml"), []byte(genYAML), 0o644); err != nil {
			t.Fatalf("failed to write gen.yaml: %v", err)
		}
	}

	_, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("Changes() returned unexpected error: %v", err)
	}
}
