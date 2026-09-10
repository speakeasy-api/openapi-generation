package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedMethods_DeeplyNestedWithDifferentLanguages(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /cloud/regions/{regionId}/clusters/{clusterId}/nodes/{nodeId}/metrics:
    get:
      operationId: getNodeMetrics
      x-speakeasy-group: cloud.regions.clusters.nodes
      x-speakeasy-name-override: getMetrics
      parameters:
        - name: regionId
          in: path
          required: true
          schema:
            type: string
        - name: clusterId
          in: path
          required: true
          schema:
            type: string
        - name: nodeId
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
paths:
  /cloud/regions/{regionId}/clusters/{clusterId}/nodes/{nodeId}/metrics:
    get:
      operationId: getNodeMetrics
      x-speakeasy-group: cloud.regions.clusters.nodes
      x-speakeasy-name-override: getMetrics
      parameters:
        - name: regionId
          in: path
          required: true
          schema:
            type: string
        - name: clusterId
          in: path
          required: true
          schema:
            type: string
        - name: nodeId
          in: path
          required: true
          schema:
            type: string
        - name: period
          in: query
          required: true
          schema:
            type: string
            enum: [1h, 24h, 7d, 30d]
      responses:
        '200':
          description: Success`

	lang := "java"
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_with_different_languages", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
