package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_FormattingByLanguage_CsharpPascalCase(t *testing.T) {
	// Common spec for all languages
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /user-profiles/{userId}:
    get:
      operationId: getUserProfile
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
  /admin/system-config:
    post:
      operationId: updateSystemConfig
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /user-profiles/{userId}:
    get:
      operationId: getUserProfile
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
  /admin/system-config:
    post:
      operationId: updateSystemConfig
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                maxRetries:
                  type: integer
      responses:
        '200':
          description: Success
  /analytics/user-events:
    post:
      operationId: trackUserEvents
      responses:
        '201':
          description: Created`

	lang := "csharp"
	oldConfig, newConfig := createTestConfigs(t, testSpecPair{
		oldSpec: spec,
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
	snapshot := formatTestSnapshotWithDiff("csharp_pascal_case", spec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
