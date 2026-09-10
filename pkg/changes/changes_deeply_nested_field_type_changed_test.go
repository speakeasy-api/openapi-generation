package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedFields_DeeplyNestedFieldTypeChanged(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /config:
    put:
      operationId: updateConfig
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                system:
                  type: object
                  properties:
                    modules:
                      type: object
                      properties:
                        authentication:
                          type: object
                          properties:
                            providers:
                              type: object
                              properties:
                                oauth:
                                  type: object
                                  properties:
                                    settings:
                                      type: object
                                      properties:
                                        timeout:
                                          type: integer
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /config:
    put:
      operationId: updateConfig
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                system:
                  type: object
                  properties:
                    modules:
                      type: object
                      properties:
                        authentication:
                          type: object
                          properties:
                            providers:
                              type: object
                              properties:
                                oauth:
                                  type: object
                                  properties:
                                    settings:
                                      type: object
                                      properties:
                                        timeout:
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_field_type_changed", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
