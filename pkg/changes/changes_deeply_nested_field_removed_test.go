package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedFields_DeeplyNestedFieldRemoved(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /products:
    get:
      operationId: getProduct
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  catalog:
                    type: object
                    properties:
                      inventory:
                        type: object
                        properties:
                          warehouse:
                            type: object
                            properties:
                              locations:
                                type: object
                                properties:
                                  primary:
                                    type: object
                                    properties:
                                      shelf:
                                        type: object
                                        properties:
                                          row:
                                            type: string
                                          column:
                                            type: integer
                                          level:
                                            type: integer`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /products:
    get:
      operationId: getProduct
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  catalog:
                    type: object
                    properties:
                      inventory:
                        type: object
                        properties:
                          warehouse:
                            type: object
                            properties:
                              locations:
                                type: object
                                properties:
                                  primary:
                                    type: object
                                    properties:
                                      shelf:
                                        type: object
                                        properties:
                                          row:
                                            type: string`

	lang := "python"
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_field_removed", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
