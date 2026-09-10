package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedFields_DeeplyNestedFieldAdded(t *testing.T) {
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
                    personal:
                      type: object
                      properties:
                        details:
                          type: object
                          properties:
                            contact:
                              type: object
                              properties:
                                primary:
                                  type: object
                                  properties:
                                    email:
                                      type: string
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
                profile:
                  type: object
                  properties:
                    personal:
                      type: object
                      properties:
                        details:
                          type: object
                          properties:
                            contact:
                              type: object
                              properties:
                                primary:
                                  type: object
                                  properties:
                                    email:
                                      type: string
                                    phone:
                                      type: string
                                    address:
                                      type: object
                                      properties:
                                        street:
                                          type: string
                                        city:
                                          type: string
                                        country:
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
	// Snapshot both compact and full modes

	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff("deeply_nested_field_added", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
