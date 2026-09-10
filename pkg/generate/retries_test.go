package generate

import (
	"context"
	"net/http"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMethodRetries(t *testing.T) {
	t.Parallel()

	globalRetries := &extensions.Retries{Strategy: "backoff"}
	operationRetries := &extensions.Retries{Strategy: "attempt-count-backoff"}

	tests := []struct {
		name                  string
		target                string
		policy                string
		method                string
		overrideGlobalRetries bool
		operationRetries      *extensions.Retries
		want                  *extensions.Retries
	}{
		{
			name:   "spec policy preserves inherited post retries",
			target: "cli",
			policy: "spec",
			method: http.MethodPost,
			want:   globalRetries,
		},
		{
			name:   "non CLI targets ignore safe methods policy",
			target: "go",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodPost,
			want:   globalRetries,
		},
		{
			name:   "safe methods policy removes inherited post retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodPost,
		},
		{
			name:   "safe methods policy removes inherited put retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodPut,
		},
		{
			name:   "safe methods policy removes inherited patch retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodPatch,
		},
		{
			name:   "safe methods policy removes inherited delete retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodDelete,
		},
		{
			name:   "safe methods policy preserves inherited get retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodGet,
			want:   globalRetries,
		},
		{
			name:   "safe methods policy preserves inherited head retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodHead,
			want:   globalRetries,
		},
		{
			name:   "safe methods policy preserves inherited options retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodOptions,
			want:   globalRetries,
		},
		{
			name:   "safe methods policy preserves inherited trace retries",
			target: "cli",
			policy: cliRetryMethodPolicySafeMethods,
			method: http.MethodTrace,
			want:   globalRetries,
		},
		{
			name:                  "explicit post policy wins",
			target:                "cli",
			policy:                cliRetryMethodPolicySafeMethods,
			method:                http.MethodPost,
			overrideGlobalRetries: true,
			operationRetries:      operationRetries,
			want:                  operationRetries,
		},
		{
			name:                  "explicit post none wins",
			target:                "cli",
			policy:                cliRetryMethodPolicySafeMethods,
			method:                http.MethodPost,
			overrideGlobalRetries: true,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveMethodRetries(tt.target, tt.policy, oas.HTTPMethod(tt.method), globalRetries, tt.operationRetries, tt.overrideGlobalRetries)
			require.Same(t, tt.want, got)
		})
	}
}

func TestInheritedRetriesFilteredByMethod(t *testing.T) {
	t.Parallel()

	assert.True(t, inheritedRetriesFilteredByMethod("cli", cliRetryMethodPolicySafeMethods))
	assert.False(t, inheritedRetriesFilteredByMethod("cli", "spec"))
	assert.False(t, inheritedRetriesFilteredByMethod("cli", ""))
	assert.False(t, inheritedRetriesFilteredByMethod("go", cliRetryMethodPolicySafeMethods))
}

func TestHandlePathsRecordsRetryFeatureAfterMethodFiltering(t *testing.T) {
	t.Parallel()

	const unsafeOnly = `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      responses:
        "200":
          description: ok
`
	const withSafe = unsafeOnly + `    get:
      operationId: listThings
      responses:
        "200":
          description: ok
`
	const explicitOverride = `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      x-speakeasy-retries:
        strategy: backoff
        backoff:
          initialInterval: 500
          maxInterval: 60000
          maxElapsedTime: 3600000
          exponent: 1.5
        statusCodes: [5XX]
        retryConnectionErrors: true
      responses:
        "200":
          description: ok
`

	tests := []struct {
		name   string
		spec   string
		policy string
		want   bool
	}{
		{"safe methods policy with only unsafe operations", unsafeOnly, cliRetryMethodPolicySafeMethods, false},
		{"safe methods policy with a safe operation", withSafe, cliRetryMethodPolicySafeMethods, true},
		{"safe methods policy with an explicit operation policy", explicitOverride, cliRetryMethodPolicySafeMethods, true},
		{"spec policy keeps inherited retries", unsafeOnly, "spec", true},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML: tt.spec,
				Target:      "cli",
			})
			require.NoError(t, err)
			setup.Config.Languages["cli"].Cfg["retryMethodPolicy"] = tt.policy

			g := &Generator{
				subsystem: setup.Subsystem,
				schemas: &schemas.Schemas{
					Config:    setup.Config,
					Target:    setup.Target,
					Subsystem: setup.Subsystem,
					Namer:     setup.Namer,
				},
				target: setup.Target,
				namer:  setup.Namer,
			}
			_, err = g.handlePaths(context.Background(), handlePathsParams{
				Paths:         setup.DocInfo.Doc.Paths,
				GlobalRetries: &extensions.Retries{Strategy: "backoff"},
				AnalyticsData: &analytics.Data{},
				DocInfo:       setup.DocInfo,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, setup.Subsystem.Features.IsFeatureUsed(features.FeatureRetries))
		})
	}
}
