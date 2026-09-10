package snapshots

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapGoNamespaceCollisionStress stress-tests Go namespace collision detection
// by using x-speakeasy-model-namespace values that match common internal Go SDK
// package names (types, utils, operations, shared, hooks, retry, etc.).
// Each "Widget" model is placed into a namespace that would normally collide with
// an internal import. The generator must detect these collisions and produce unique
// import aliases (e.g., modelstypes "module/models/types") to avoid duplicate
// package identifiers.
func TestSnapGoNamespaceCollisionStress(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Namespace Collision Stress Test
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /ns/utils:
    post:
      operationId: createUtilsWidget
      tags:
        - utils
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/utils_Widget"
  /ns/types:
    post:
      operationId: createTypesWidget
      tags:
        - types
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/types_Widget"
  /ns/operations:
    post:
      operationId: createOperationsWidget
      tags:
        - operations
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/operations_Widget"
  /ns/shared:
    post:
      operationId: createSharedWidget
      tags:
        - shared
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/shared_Widget"
  /ns/hooks:
    post:
      operationId: createHooksWidget
      tags:
        - hooks
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/hooks_Widget"
  /ns/retry:
    post:
      operationId: createRetryWidget
      tags:
        - retry
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/retry_Widget"
  /ns/config:
    post:
      operationId: createConfigWidget
      tags:
        - config
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/config_Widget"
  /ns/apierrors:
    post:
      operationId: createApierrorsWidget
      tags:
        - apierrors
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/apierrors_Widget"
  /ns/safe:
    post:
      operationId: createSafeWidget
      tags:
        - safe
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/safe_Widget"
components:
  schemas:
    utils_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: utils
      type: object
      properties:
        utils:
          type: string
    types_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: types
      type: object
      properties:
        types:
          type: string
    operations_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: operations
      type: object
      properties:
        operations:
          type: string
    shared_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: shared
      type: object
      properties:
        shared:
          type: string
    hooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: hooks
      type: object
      properties:
        hooks:
          type: string
    retry_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: retry
      type: object
      properties:
        retry:
          type: string
    config_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: config
      type: object
      properties:
        config:
          type: string
    apierrors_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: apierrors
      type: object
      properties:
        apierrors:
          type: string
    safe_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: safe
      type: object
      properties:
        safe:
          type: string
`

	genYaml := `go:
  packageName: github.com/example/nscollision
`

	expectedSnapshotFiles := []string{
		"utils.go",
		"models/utils/widget.go",
		"models/types/widget.go",
	}

	expectedSnapshot := `--- models/types/widget.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package types

type Widget struct {
	Types *string ` + "`" + `json:"types,omitzero"` + "`" + `
}

func (w *Widget) GetTypes() *string {
	if w == nil {
		return nil
	}
	return w.Types
}


--- models/utils/widget.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package utils

type Widget struct {
	Utils *string ` + "`" + `json:"utils,omitzero"` + "`" + `
}

func (w *Widget) GetUtils() *string {
	if w == nil {
		return nil
	}
	return w.Utils
}


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
