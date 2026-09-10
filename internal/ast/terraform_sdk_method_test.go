package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformGoTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "CreateThingRequest",
			expected: "CreateThingRequest",
		},
		{
			name:     "name with underscores",
			input:    "my_thing_request",
			expected: "MyThingRequest",
		},
		{
			name:     "name with hyphens",
			input:    "my-thing-request",
			expected: "MyThingRequest",
		},
		{
			name:     "SDK acronym preserved",
			input:    "sdk_response",
			expected: "SDKResponse",
		},
		{
			name:     "ID acronym preserved",
			input:    "thing_id",
			expected: "ThingID",
		},
		{
			name:     "URL acronym preserved",
			input:    "server_url",
			expected: "ServerURL",
		},
		{
			name:     "method prefix with scope dot",
			input:    "To_shared.CreateThingRequest",
			expected: "ToSharedCreateThingRequest",
		},
		{
			name:     "method prefix with operations scope dot",
			input:    "RefreshFrom_operations.GetThingResponse",
			expected: "RefreshFromOperationsGetThingResponse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TerraformGoTypeName(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTerraformSDKMethodName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		td       *TypeDef
		expected string
	}{
		{
			name:   "operations-scoped class gets operations prefix",
			prefix: "To_",
			td: &TypeDef{
				Name:  "CreateThingRequest",
				Type:  DataTypeClass,
				Scope: ScopeOperations,
			},
			// Default scope = operations. td.Scope matches → swap to shared.
			// td.Scope (operations) != swapped scope (shared) → prefix "operations."
			expected: "ToOperationsCreateThingRequest",
		},
		{
			name:   "shared-scoped class gets shared prefix",
			prefix: "RefreshFrom_",
			td: &TypeDef{
				Name:  "Thing",
				Type:  DataTypeClass,
				Scope: ScopeShared,
			},
			// Default scope = operations. td.Scope (shared) != operations → no swap.
			// td.Scope (shared) != scope (operations) → prefix "shared."
			expected: "RefreshFromSharedThing",
		},
		{
			name:   "union in shared scope",
			prefix: "RefreshFrom_",
			td: &TypeDef{
				Name:  "ThingVariant",
				Type:  DataTypeUnion,
				Scope: ScopeShared,
			},
			expected: "RefreshFromSharedThingVariant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := terraformSDKMethodName(tt.prefix, tt.td)

			assert.Equal(t, tt.expected, result)
		})
	}
}
