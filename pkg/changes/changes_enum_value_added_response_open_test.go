package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_EnumValueAddedResponseOpen(t *testing.T) {
	// When an enum has x-speakeasy-unknown-values: allow (Open=true),
	// adding a new enum value to a response is NOT breaking because
	// the SDK already handles unknown values gracefully.
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /orders/{id}:
    get:
      operationId: getOrder
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    x-speakeasy-unknown-values: allow
                    enum:
                      - pending
                      - processing`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /orders/{id}:
    get:
      operationId: getOrder
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    x-speakeasy-unknown-values: allow
                    enum:
                      - pending
                      - processing
                      - shipped`

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

	snapshot := formatTestSnapshotWithDiff("enum_value_added_response_open", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
