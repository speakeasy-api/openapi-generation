package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_ArrayAdded(t *testing.T) {
	// Test case: Array field added to existing response
	oldSpec := `openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get all users
      operationId: getUsers
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string`

	newSpec := `openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get all users
      operationId: getUsers
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                  tags:
                    type: array
                    items:
                      type: string
                  permissions:
                    type: array
                    items:
                      type: object
                      properties:
                        resource:
                          type: string
                        actions:
                          type: array
                          items:
                            type: string`

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
	snapshot := formatTestSnapshotWithDiff(t.Name(), oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
