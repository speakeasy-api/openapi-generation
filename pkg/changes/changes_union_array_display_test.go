package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_UnionArrayDisplay(t *testing.T) {
	// Test that array types in unions display as Array<ItemType>
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /messages:
    post:
      operationId: sendMessage
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                content:
                  oneOf:
                    - type: string
                    - type: array
                      items:
                        $ref: '#/components/schemas/ContentChunk'
      responses:
        '200':
          description: Success
components:
  schemas:
    ContentChunk:
      type: object
      properties:
        text:
          type: string`

	// New spec: array item type changed
	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /messages:
    post:
      operationId: sendMessage
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                content:
                  oneOf:
                    - type: string
                    - type: array
                      items:
                        $ref: '#/components/schemas/ContentChunk'
      responses:
        '200':
          description: Success
components:
  schemas:
    ContentChunk:
      type: object
      properties:
        text:
          type: string
        format:
          type: string`

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

	snapshot := formatTestSnapshotWithDiff("union_array_display", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
