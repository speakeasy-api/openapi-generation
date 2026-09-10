package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_Breaking_ResponseRequiredFieldRemoved(t *testing.T) {
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
                required:
                  - id
                  - name
                properties:
                  id:
                    type: string
                  name:
                    type: string`

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
                required:
                  - id
                properties:
                  id:
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
	snapshot := formatTestSnapshotWithDiff("response_required_field_removed", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
