package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedUnion(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /events:
    post:
      operationId: createEvent
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                event:
                  type: object
                  properties:
                    data:
                      type: object
                      properties:
                        payload:
                          type: object
                          properties:
                            content:
                              type: object
                              properties:
                                body:
                                  oneOf:
                                    - type: string
                                    - type: object
                                      properties:
                                        text:
                                          type: string
                                        format:
                                          type: string
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /events:
    post:
      operationId: createEvent
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                event:
                  type: object
                  properties:
                    data:
                      type: object
                      properties:
                        payload:
                          type: object
                          properties:
                            content:
                              type: object
                              properties:
                                body:
                                  oneOf:
                                    - type: string
                                    - type: object
                                      properties:
                                        text:
                                          type: string
                                        format:
                                          type: string
                                    - type: array
                                      items:
                                        type: string
      responses:
        '200':
          description: Success`

	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    "csharp",
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	diff.NewSubsystem.Target.Target = "csharp"
	// Snapshot both compact and full modes

	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff("deeply_nested_union_changes", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
