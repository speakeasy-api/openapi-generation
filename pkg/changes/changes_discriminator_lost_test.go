package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DiscriminatorLost(t *testing.T) {
	// Old spec: union with discriminator
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /messages:
    get:
      operationId: getMessage
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  content:
                    oneOf:
                      - $ref: '#/components/schemas/TextMessage'
                      - $ref: '#/components/schemas/ImageMessage'
                    discriminator:
                      propertyName: type
                      mapping:
                        text: '#/components/schemas/TextMessage'
                        image: '#/components/schemas/ImageMessage'
components:
  schemas:
    TextMessage:
      type: object
      properties:
        type:
          type: string
        text:
          type: string
    ImageMessage:
      type: object
      properties:
        type:
          type: string
        url:
          type: string`

	// New spec: union without discriminator
	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /messages:
    get:
      operationId: getMessage
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  content:
                    oneOf:
                      - $ref: '#/components/schemas/TextMessage'
                      - $ref: '#/components/schemas/ImageMessage'
components:
  schemas:
    TextMessage:
      type: object
      properties:
        text:
          type: string
    ImageMessage:
      type: object
      properties:
        url:
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

	snapshot := formatTestSnapshotWithDiff("discriminator_lost", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
