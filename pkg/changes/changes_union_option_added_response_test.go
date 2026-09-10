package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_UnionOptionAddedResponse(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUser
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  a:
                    type: object
                    properties:
                      b:
                        type: object
                        properties:
                          c:
                            oneOf:
                              - type: string
                              - type: integer`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUser
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  a:
                    type: object
                    properties:
                      b:
                        type: object
                        properties:
                          c:
                            oneOf:
                              - type: string
                              - type: integer
                              - type: boolean`

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
	snapshot := formatTestSnapshotWithDiff("union_option_added_response", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
