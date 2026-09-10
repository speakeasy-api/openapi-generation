package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedFields_DeeplyNestedArrayAndMapChanges(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /data:
    post:
      operationId: processData
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                datasets:
                  type: object
                  properties:
                    collections:
                      type: array
                      items:
                        type: object
                        properties:
                          metadata:
                            type: object
                            properties:
                              tags:
                                type: object
                                additionalProperties:
                                  type: array
                                  items:
                                    type: object
                                    properties:
                                      value:
                                        type: string
                                      priority:
                                        type: integer
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /data:
    post:
      operationId: processData
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                datasets:
                  type: object
                  properties:
                    collections:
                      type: array
                      items:
                        type: object
                        properties:
                          metadata:
                            type: object
                            properties:
                              tags:
                                type: object
                                additionalProperties:
                                  type: array
                                  items:
                                    type: object
                                    properties:
                                      value:
                                        type: string
                                      priority:
                                        type: number
                                      category:
                                        type: string
      responses:
        '200':
          description: Success`

	lang := "java"
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_array_and_map_changes", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
