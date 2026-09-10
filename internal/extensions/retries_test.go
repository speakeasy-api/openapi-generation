package extensions

import (
	"testing"

	oasextensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHandleOperationRetryExtensionAttemptCountMaxRetries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		wantRetries *int
		wantErr     string
	}{
		{
			name: "allows zero max retries",
			raw: `
strategy: attempt-count-backoff
maxRetries: 0
statusCodes:
  - 5XX`,
			wantRetries: pointerTo(0),
		},
		{
			name: "rejects negative max retries",
			raw: `
strategy: attempt-count-backoff
maxRetries: -1
statusCodes:
  - 5XX`,
			wantErr: "x-speakeasy-retries.maxRetries must be greater than or equal to 0",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			operation := &openapi.Operation{
				Extensions: oasextensions.New(),
			}
			operation.Extensions.Set(ExtRetries.Name(), parseRetryTestNode(t, tt.raw))

			got, ok, err := (&Extensions{}).HandleOperationRetryExtension(operation)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.False(t, ok)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.True(t, ok)
			require.NotNil(t, got)
			require.Equal(t, tt.wantRetries, got.MaxRetries)
		})
	}
}

func TestHandleOperationRetryExtensionNoRetryOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     *string
		wantSet bool
		wantNil bool
	}{
		{
			name:    "absent extension does not override",
			wantSet: false,
			wantNil: true,
		},
		{
			name:    "empty extension overrides with none",
			raw:     pointerTo("{}"),
			wantSet: true,
			wantNil: true,
		},
		{
			name:    "disabled extension overrides with none",
			raw:     pointerTo("disabled: true"),
			wantSet: true,
			wantNil: true,
		},
		{
			name:    "none strategy overrides without status codes",
			raw:     pointerTo("strategy: none"),
			wantSet: true,
			wantNil: true,
		},
		{
			name: "none strategy tolerates stale merged fields",
			raw: pointerTo(`
strategy: none
maxRetries: -1
backoff:
  initialInterval: 10
statusCodes:
  - 503
retryConnectionErrors: true`),
			wantSet: true,
			wantNil: true,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			operation := &openapi.Operation{Extensions: oasextensions.New()}
			if tt.raw != nil {
				operation.Extensions.Set(ExtRetries.Name(), parseRetryTestNode(t, *tt.raw))
			}

			got, overridden, err := (&Extensions{}).HandleOperationRetryExtension(operation)
			require.NoError(t, err)
			require.Equal(t, tt.wantSet, overridden)
			if tt.wantNil {
				require.Nil(t, got)
			}
		})
	}
}

func TestHandleGlobalRetryExtensionNoRetrySuppressesDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "disabled", raw: "disabled: true"},
		{name: "none", raw: "strategy: none"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := &openapi.OpenAPI{Extensions: oasextensions.New()}
			doc.Extensions.Set(ExtRetries.Name(), parseRetryTestNode(t, tt.raw))

			got, err := (&Extensions{}).HandleGlobalRetryExtension(doc, true)
			require.NoError(t, err)
			require.Nil(t, got)
		})
	}
}

func TestHandleRetryExtensionValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{
			name:    "backoff still requires status codes",
			raw:     "strategy: backoff",
			wantErr: "x-speakeasy-retries.statusCodes is required",
		},
		{
			name: "invalid strategy lists none",
			raw: `
strategy: immediate
statusCodes:
  - 503`,
			wantErr: "x-speakeasy-retries.strategy must be one of backoff, attempt-count-backoff, or none",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			operation := &openapi.Operation{Extensions: oasextensions.New()}
			operation.Extensions.Set(ExtRetries.Name(), parseRetryTestNode(t, tt.raw))

			got, overridden, err := (&Extensions{}).HandleOperationRetryExtension(operation)
			require.ErrorContains(t, err, tt.wantErr)
			require.False(t, overridden)
			require.Nil(t, got)
		})
	}
}

func parseRetryTestNode(t *testing.T, raw string) *yaml.Node {
	t.Helper()

	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(raw), &node))
	require.NotEmpty(t, node.Content)

	return node.Content[0]
}

func pointerTo[T any](value T) *T {
	return &value
}
