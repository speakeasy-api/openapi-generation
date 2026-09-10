package changes

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestChanges_DeeplyNestedMethods_DeeplyNestedMethodChanged(t *testing.T) {
	oldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /admin/analytics/reports/revenue/daily/summary:
    get:
      operationId: getDailyRevenueSummary
      x-speakeasy-group: admin.analytics.reports.revenue.daily
      x-speakeasy-name-override: getSummary
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  total:
                    type: number
                  date:
                    type: string`

	newSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /admin/analytics/reports/revenue/daily/summary:
    get:
      operationId: getDailyRevenueSummary
      x-speakeasy-group: admin.analytics.reports.revenue.daily
      x-speakeasy-name-override: getSummary
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  total:
                    type: number
                  date:
                    type: string
                  currency:
                    type: string
                  breakdown:
                    type: object
                    properties:
                      morning:
                        type: number
                      afternoon:
                        type: number
                      evening:
                        type: number`

	lang := "python"
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
	snapshot := formatTestSnapshotWithDiff("deeply_nested_method_changed", oldSpec, newSpec, diff)
	snaps.MatchSnapshot(t, snapshot)
}
