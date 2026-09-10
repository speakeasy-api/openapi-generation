package extensions

import (
	"testing"

	oasextensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/require"
)

func TestHandleReactHookExtensionQueryKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    *ReactHook
		wantErr string
	}{
		{
			name: "parses includeRequestBody on query hooks",
			raw: `
type: query
queryKey:
  includeRequestBody: true`,
			want: &ReactHook{
				Type:     "query",
				QueryKey: &ReactHookQueryKey{IncludeRequestBody: true},
			},
		},
		{
			name: "allows includeRequestBody with inferred type",
			raw: `
queryKey:
  includeRequestBody: true`,
			want: &ReactHook{
				Type:     "infer",
				QueryKey: &ReactHookQueryKey{IncludeRequestBody: true},
			},
		},
		{
			name: "rejects includeRequestBody on mutation hooks",
			raw: `
type: mutation
queryKey:
  includeRequestBody: true`,
			wantErr: "queryKey.includeRequestBody only applies to query hooks",
		},
		{
			name: "allows queryKey without includeRequestBody on mutation hooks",
			raw: `
type: mutation
queryKey:
  includeRequestBody: false`,
			want: &ReactHook{
				Type:     "mutation",
				QueryKey: &ReactHookQueryKey{IncludeRequestBody: false},
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			operation := &openapi.Operation{
				Extensions: oasextensions.New(),
			}
			operation.Extensions.Set(ExtReactHook.Name(), parseRetryTestNode(t, tt.raw))

			got, err := (&Extensions{}).HandleReactHookExtension(operation)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
