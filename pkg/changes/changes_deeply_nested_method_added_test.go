package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedMethods_DeeplyNestedMethodAdded(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
x-speakeasy-globals:
  parameters:
    - name: orgId
      in: path
      required: true
      schema:
        type: string
paths:
  /organizations/{orgId}/projects/{projectId}/environments/{envId}/deployments:
    get:
      operationId: listDeployments
      x-speakeasy-group: organizations.projects.environments.deployments
      x-speakeasy-name-override: list
      parameters:
        - name: orgId
          in: path
          required: true
          schema:
            type: string
        - name: projectId
          in: path
          required: true
          schema:
            type: string
        - name: envId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
x-speakeasy-globals:
  parameters:
    - name: orgId
      in: path
      required: true
      schema:
        type: string
paths:
  /organizations/{orgId}/projects/{projectId}/environments/{envId}/deployments:
    get:
      operationId: listDeployments
      x-speakeasy-group: organizations.projects.environments.deployments
      x-speakeasy-name-override: list
      parameters:
        - name: orgId
          in: path
          required: true
          schema:
            type: string
        - name: projectId
          in: path
          required: true
          schema:
            type: string
        - name: envId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
  /organizations/{orgId}/projects/{projectId}/environments/{envId}/deployments/{deploymentId}/logs:
    get:
      operationId: getDeploymentLogs
      x-speakeasy-group: organizations.projects.environments.deployments.logs
      x-speakeasy-name-override: get
      parameters:
        - name: orgId
          in: path
          required: true
          schema:
            type: string
        - name: projectId
          in: path
          required: true
          schema:
            type: string
        - name: envId
          in: path
          required: true
          schema:
            type: string
        - name: deploymentId
          in: path
          required: true
          schema:
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_method_added", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
