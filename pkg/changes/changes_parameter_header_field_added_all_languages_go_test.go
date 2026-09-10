package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_ParameterChanges_AllLanguages_HeaderFieldAdded_Go(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUser
      parameters:
        - name: X-Request-ID
          in: header
          schema:
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
    get:
      operationId: getUser
      parameters:
        - name: X-Request-ID
          in: header
          schema:
            type: string
        - name: X-API-Key
          in: header
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success`

	lang := "go"
	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: oldSpec,
		newSpec: newSpec,
		lang:    lang,
	})

	diff, err := Changes(t.Context(), oldConfig, newConfig)
	if err != nil {
		t.Fatalf("failed to get changes. error: %s", err.Error())
	}
	// Snapshot both compact and full modes

	// Include specs and output for easier review
	snapshot := formatTestSnapshotWithDiff("header_field_added", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
