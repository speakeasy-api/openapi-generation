package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedMethods_MultipleDeeplyNestedChanges(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/v1/users/profile/settings/privacy:
    get:
      operationId: getPrivacySettings
      x-speakeasy-group: api.v1.users.profile.settings
      x-speakeasy-name-override: getPrivacy
      responses:
        '200':
          description: Success
  /api/v1/users/profile/settings/notifications:
    get:
      operationId: getNotificationSettings
      x-speakeasy-group: api.v1.users.profile.settings
      x-speakeasy-name-override: getNotifications
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/v1/users/profile/settings/privacy:
    get:
      operationId: getPrivacySettings
      x-speakeasy-group: api.v1.users.profile.settings
      x-speakeasy-name-override: getPrivacy
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                includeDefaults:
                  type: boolean
      responses:
        '200':
          description: Success
  /api/v1/users/profile/settings/notifications:
    get:
      operationId: getNotificationSettings
      x-speakeasy-group: api.v1.users.profile.settings
      x-speakeasy-name-override: getNotifications
      deprecated: true
      responses:
        '200':
          description: Success
  /api/v1/users/profile/settings/appearance:
    post:
      operationId: updateAppearanceSettings
      x-speakeasy-group: api.v1.users.profile.settings
      x-speakeasy-name-override: updateAppearance
      responses:
        '201':
          description: Created`

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
	snapshot := formatTestSnapshotWithDiff("multiple_deeply_nested_changes", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
