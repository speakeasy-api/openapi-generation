package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformSDKMethodTarget_SDKRequestMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		target             TerraformSDKMethodTarget
		expectedMethodName string
	}{
		{
			name: "operations-scoped class",
			target: TerraformSDKMethodTarget{
				TypeDef: &TypeDef{
					Name:  "CreateThingRequestBody",
					Type:  DataTypeClass,
					Scope: ScopeOperations,
				},
			},
			expectedMethodName: "ToOperationsCreateThingRequestBody",
		},
		{
			name: "shared-scoped class",
			target: TerraformSDKMethodTarget{
				TypeDef: &TypeDef{
					Name:  "ThingInput",
					Type:  DataTypeClass,
					Scope: ScopeShared,
				},
			},
			expectedMethodName: "ToSharedThingInput",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			method := tt.target.SDKRequestMethod()

			assert.Equal(t, tt.expectedMethodName, method.MethodName)
			assert.False(t, method.Target.IsArrayWrapper)
		})
	}
}

func TestTerraformSDKMethodTarget_SDKResponseMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		target             TerraformSDKMethodTarget
		expectedMethodName string
		expectedArray      bool
	}{
		{
			name: "shared-scoped class uses RefreshFrom prefix",
			target: TerraformSDKMethodTarget{
				TypeDef: &TypeDef{
					Name:  "Thing",
					Type:  DataTypeClass,
					Scope: ScopeShared,
				},
			},
			expectedMethodName: "RefreshFromSharedThing",
		},
		{
			name: "operations-scoped class uses RefreshFrom prefix",
			target: TerraformSDKMethodTarget{
				TypeDef: &TypeDef{
					Name:  "GetThingResponse",
					Type:  DataTypeClass,
					Scope: ScopeOperations,
				},
			},
			expectedMethodName: "RefreshFromOperationsGetThingResponse",
		},
		{
			name: "array wrapper uses RefreshFromArrayOf prefix",
			target: TerraformSDKMethodTarget{
				TypeDef: &TypeDef{
					Name:  "ListThingsResponse",
					Type:  DataTypeClass,
					Scope: ScopeOperations,
				},
				IsArrayWrapper: true,
			},
			expectedMethodName: "RefreshFromArrayOfOperationsListThingsResponse",
			expectedArray:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			method := tt.target.SDKResponseMethod()

			assert.Equal(t, tt.expectedMethodName, method.MethodName)
			assert.Equal(t, tt.expectedArray, method.Target.IsArrayWrapper)
		})
	}
}
