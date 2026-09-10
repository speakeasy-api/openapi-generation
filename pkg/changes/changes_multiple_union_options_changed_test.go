package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_MultipleUnionOptionsChanged(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                status:
                  oneOf:
                    - type: string
                      enum: [active]
                    - type: string
                      enum: [inactive]
                    - type: string
                      enum: [pending]
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                status:
                  oneOf:
                    - type: string
                      enum: [active]
                    - type: string
                      enum: [suspended]
                    - type: string
                      enum: [deleted]
                    - type: string
                      enum: [archived]
      responses:
        '200':
          description: Success`

	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    "typescript",
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	// Snapshot both compact and full modes

	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff("multiple_union_options_changed", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
