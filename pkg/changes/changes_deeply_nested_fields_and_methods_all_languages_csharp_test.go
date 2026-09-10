package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedFieldsAndMethods_AllLanguages_Csharp(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/complex:
    post:
      tags:
        - inner
      operationId: complexOperation
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                level1:
                  type: object
                  properties:
                    level2:
                      type: object
                      properties:
                        level3:
                          type: object
                          properties:
                            field1:
                              type: string
                            level4:
                              type: object
                              properties:
                                level5:
                                  type: object
                                  properties:
                                    field2:
                                      type: integer
                                    field3:
                                      type: boolean
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/complex:
    post:
      tags:
        - inner
      operationId: complexOperation
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                level1:
                  type: object
                  properties:
                    level2:
                      type: object
                      properties:
                        level3:
                          type: object
                          properties:
                            field1:
                              type: integer
                            field1a:
                              type: string
                            level4:
                              type: object
                              properties:
                                level5:
                                  type: object
                                  properties:
                                    field2:
                                      type: integer
                                    field3:
                                      type: string
                                    field4:
                                      type: object
                                      properties:
                                        nested:
                                          type: string
      responses:
        '200':
          description: Success`

	lang := "csharp"
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
	snapshot := formatTestSnapshotWithDiff("multiple_deeply_nested_changes", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
