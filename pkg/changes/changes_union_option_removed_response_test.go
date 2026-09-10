package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_UnionOptionRemovedResponse(t *testing.T) {
	// Removing a discriminated union option from a response is always breaking because
	// SDK code that handles that option will no longer work.
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: getPet
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  pet:
                    oneOf:
                      - $ref: '#/components/schemas/Dog'
                      - $ref: '#/components/schemas/Cat'
                      - $ref: '#/components/schemas/Bird'
                    discriminator:
                      propertyName: kind
                      mapping:
                        dog: '#/components/schemas/Dog'
                        cat: '#/components/schemas/Cat'
                        bird: '#/components/schemas/Bird'
components:
  schemas:
    Dog:
      type: object
      required: [kind, bark]
      properties:
        kind:
          type: string
        bark:
          type: boolean
    Cat:
      type: object
      required: [kind, meow]
      properties:
        kind:
          type: string
        meow:
          type: boolean
    Bird:
      type: object
      required: [kind, chirp]
      properties:
        kind:
          type: string
        chirp:
          type: boolean`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: getPet
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  pet:
                    oneOf:
                      - $ref: '#/components/schemas/Dog'
                      - $ref: '#/components/schemas/Cat'
                    discriminator:
                      propertyName: kind
                      mapping:
                        dog: '#/components/schemas/Dog'
                        cat: '#/components/schemas/Cat'
components:
  schemas:
    Dog:
      type: object
      required: [kind, bark]
      properties:
        kind:
          type: string
        bark:
          type: boolean
    Cat:
      type: object
      required: [kind, meow]
      properties:
        kind:
          type: string
        meow:
          type: boolean`

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

	snapshot := formatTestSnapshotWithDiff("union_option_removed_response", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
