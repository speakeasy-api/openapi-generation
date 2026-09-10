package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	err := os.WriteFile(path, []byte(contents), 0o644)
	require.NoError(t, err)
}

func TestJoinSpecificationsSupportsPartialJoinedSpecs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "source.yaml")
	joinPath := filepath.Join(tempDir, "extra.yaml")

	writeTestFile(t, sourcePath, `openapi: 3.1.0
info:
  title: Main API
  version: 1.0.0
paths:
  /foo:
    get:
      operationId: getFoo
      responses:
        "200":
          description: ok
`)

	writeTestFile(t, joinPath, `paths:
  /bar:
    get:
      operationId: getBar
      responses:
        "200":
          description: ok
components:
  schemas:
    ExtraThing:
      type: object
      properties:
        name:
          type: string
`)

	node, err := joinSpecifications(ctx, sourcePath, []string{joinPath}, joinConflictCounter)
	require.NoError(t, err)

	rendered, err := encodeYAMLDocument(node)
	require.NoError(t, err)

	renderedText := string(rendered)
	assert.Contains(t, renderedText, "/foo:")
	assert.Contains(t, renderedText, "/bar:")
	assert.Contains(t, renderedText, "ExtraThing:")
}

func TestJoinSpecificationsPreservesMatchingRootSecurity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "source.yaml")
	joinPath := filepath.Join(tempDir, "extra.yaml")

	writeTestFile(t, sourcePath, `openapi: 3.1.0
info:
  title: Main API
  version: 1.0.0
security:
  - bearerAuth: []
  - {}
paths:
  /foo:
    get:
      operationId: getFoo
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`)

	writeTestFile(t, joinPath, `security:
  - bearerAuth: []
  - {}
paths:
  /bar:
    get:
      operationId: getBar
      responses:
        "200":
          description: ok
`)

	node, err := joinSpecifications(ctx, sourcePath, []string{joinPath}, joinConflictCounter)
	require.NoError(t, err)

	rendered, err := encodeYAMLDocument(node)
	require.NoError(t, err)

	renderedText := string(rendered)
	assert.Contains(t, renderedText, "security:")
	assert.Contains(t, renderedText, "bearerAuth: []")
	assert.Contains(t, renderedText, "/bar:")
	assert.NotContains(t, renderedText, "/bar:\n    get:\n      security:")
}

func TestJoinSpecificationsPreservesExplicitOperationSecurityOverrides(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "source.yaml")
	joinPath := filepath.Join(tempDir, "extra.yaml")

	writeTestFile(t, sourcePath, `openapi: 3.1.0
info:
  title: Main API
  version: 1.0.0
security:
  - bearerAuth: []
paths:
  /foo:
    get:
      operationId: getFoo
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`)

	writeTestFile(t, joinPath, `security:
  - bearerAuth: []
paths:
  /bar:
    get:
      operationId: getBar
      security: []
      responses:
        "200":
          description: ok
`)

	node, err := joinSpecifications(ctx, sourcePath, []string{joinPath}, joinConflictCounter)
	require.NoError(t, err)

	rendered, err := encodeYAMLDocument(node)
	require.NoError(t, err)

	renderedText := string(rendered)
	assert.Contains(t, renderedText, "/bar:")
	assert.Contains(t, renderedText, "security: []")
	assert.Equal(t, 1, strings.Count(renderedText, "security: []"))
}

func TestApplyOverlaysAfterJoin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "source.yaml")
	joinPath := filepath.Join(tempDir, "extra.yaml")
	overlayPath := filepath.Join(tempDir, "overlay.yaml")

	writeTestFile(t, sourcePath, `openapi: 3.1.0
info:
  title: Main API
  version: 1.0.0
paths:
  /foo:
    get:
      operationId: getFoo
      responses:
        "200":
          description: ok
`)

	writeTestFile(t, joinPath, `paths:
  /bar:
    get:
      operationId: getBar
      responses:
        "200":
          description: ok
`)

	writeTestFile(t, overlayPath, `overlay: 1.0.0
info:
  title: Test Overlay
  version: 1.0.0
actions:
  - target: $.info
    update:
      title: Overlayed API
  - target: $.paths["/bar"]
    remove: true
`)

	node, err := joinSpecifications(ctx, sourcePath, []string{joinPath}, joinConflictCounter)
	require.NoError(t, err)

	err = applyOverlays(node, []string{overlayPath}, true)
	require.NoError(t, err)

	rendered, err := encodeYAMLDocument(node)
	require.NoError(t, err)

	renderedText := string(rendered)
	assert.Contains(t, renderedText, "title: Overlayed API")
	assert.Contains(t, renderedText, "/foo:")
	assert.NotContains(t, renderedText, "/bar:")
}
