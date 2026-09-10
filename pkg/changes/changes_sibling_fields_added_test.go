package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_SiblingFieldsAdded(t *testing.T) {
	// Test that multiple fields added at the same level all appear in full mode
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
                profile:
                  type: object
                  properties:
                    name:
                      type: string
      responses:
        '200':
          description: Success`

	// New spec: multiple sibling fields added at the same level
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
                profile:
                  type: object
                  properties:
                    name:
                      type: string
                    email:
                      type: string
                    phone:
                      type: string
                    address:
                      type: string
      responses:
        '200':
          description: Success`

	lang := "typescript"
	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    lang,
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	diff.NewSubsystem.Target.Target = lang

	snapshot := formatTestSnapshotWithDiff("sibling_fields_added", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
