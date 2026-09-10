package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformOperation_Clone(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil operation", func(t *testing.T) {
		t.Parallel()

		var op *TerraformOperation

		assert.Nil(t, op.Clone())
	})

	t.Run("copies skipDataModelRefresh", func(t *testing.T) {
		t.Parallel()

		op := &TerraformOperation{
			EntityName:           "Thing",
			EntityOperation:      "Thing#delete",
			skipDataModelRefresh: true,
		}

		cloned := op.Clone()

		assert.True(t, cloned.skipDataModelRefresh)
	})

	t.Run("preserves false skipDataModelRefresh", func(t *testing.T) {
		t.Parallel()

		op := &TerraformOperation{
			EntityName:           "Thing",
			EntityOperation:      "Thing#read",
			skipDataModelRefresh: false,
		}

		cloned := op.Clone()

		assert.False(t, cloned.skipDataModelRefresh)
	})

	t.Run("clone is independent from original", func(t *testing.T) {
		t.Parallel()

		op := &TerraformOperation{
			EntityName:           "Thing",
			EntityOperation:      "Thing#delete",
			skipDataModelRefresh: true,
		}

		cloned := op.Clone()
		cloned.skipDataModelRefresh = false

		assert.True(t, op.skipDataModelRefresh, "original should not be affected by mutation of clone")
	})
}

func TestTerraformOperations_SDKRequestMethods(t *testing.T) {
	t.Parallel()

	requestTypeA := &TypeDef{
		Name:  "CreateThingRequest",
		Type:  DataTypeClass,
		Scope: ScopeOperations,
	}
	requestTypeB := &TypeDef{
		Name:  "UpdateThingRequest",
		Type:  DataTypeClass,
		Scope: ScopeOperations,
	}

	tests := []struct {
		name                string
		ops                 TerraformOperations
		expectedMethodNames []string
	}{
		{
			name:                "nil operations produces empty list",
			ops:                 nil,
			expectedMethodNames: nil,
		},
		{
			name: "single operation with single target",
			ops: TerraformOperations{
				&TerraformOperation{
					RequestSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: requestTypeA},
					},
				},
			},
			expectedMethodNames: []string{
				"ToOperationsCreateThingRequest",
			},
		},
		{
			name: "multiple operations with different targets sorted alphabetically",
			ops: TerraformOperations{
				&TerraformOperation{
					RequestSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: requestTypeB},
					},
				},
				&TerraformOperation{
					RequestSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: requestTypeA},
					},
				},
			},
			expectedMethodNames: []string{
				"ToOperationsCreateThingRequest",
				"ToOperationsUpdateThingRequest",
			},
		},
		{
			name: "duplicate targets across operations are deduplicated",
			ops: TerraformOperations{
				&TerraformOperation{
					RequestSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: requestTypeA},
					},
				},
				&TerraformOperation{
					RequestSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: requestTypeA},
					},
				},
			},
			expectedMethodNames: []string{
				"ToOperationsCreateThingRequest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			methods := tt.ops.SDKRequestMethods()

			var names []string
			for _, m := range methods {
				names = append(names, m.MethodName)
			}

			assert.Equal(t, tt.expectedMethodNames, names)
		})
	}
}

func TestTerraformOperations_SDKRequestMethods_Operation(t *testing.T) {
	t.Parallel()

	requestType := &TypeDef{
		Name:  "CreateThingRequest",
		Type:  DataTypeClass,
		Scope: ScopeOperations,
	}

	createOp := &TerraformOperation{
		EntityOperation: "Thing#create",
	}
	createOp.RequestSDKMethodTargets = []TerraformSDKMethodTarget{
		{TypeDef: requestType, Operation: createOp},
	}

	updateOp := &TerraformOperation{
		EntityOperation: "Thing#update",
	}
	updateOp.RequestSDKMethodTargets = []TerraformSDKMethodTarget{
		{TypeDef: requestType, Operation: updateOp},
	}

	ops := TerraformOperations{createOp, updateOp}
	methods := ops.SDKRequestMethods()

	require.Len(t, methods, 1)
	assert.Same(t, createOp, methods[0].Operation, "first operation should win dedup")
}

func TestTerraformOperations_SDKResponseMethods(t *testing.T) {
	t.Parallel()

	entityType := &TypeDef{
		Name:  "Thing",
		Type:  DataTypeClass,
		Scope: ScopeShared,
	}
	responseType := &TypeDef{
		Name:  "GetThingResponse",
		Type:  DataTypeClass,
		Scope: ScopeOperations,
	}
	arrayWrapperType := &TypeDef{
		Name:  "ListThingsResponse",
		Type:  DataTypeClass,
		Scope: ScopeOperations,
	}

	tests := []struct {
		name                string
		ops                 TerraformOperations
		expectedMethodNames []string
		expectedArrayFlags  []bool
	}{
		{
			name:                "nil operations produces empty list",
			ops:                 nil,
			expectedMethodNames: nil,
			expectedArrayFlags:  nil,
		},
		{
			name: "response targets with entity and response types",
			ops: TerraformOperations{
				&TerraformOperation{
					ResponseSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: responseType},
						{TypeDef: entityType},
					},
				},
			},
			expectedMethodNames: []string{
				"RefreshFromOperationsGetThingResponse",
				"RefreshFromSharedThing",
			},
			expectedArrayFlags: []bool{false, false},
		},
		{
			name: "array wrapper target uses RefreshFromArrayOf prefix",
			ops: TerraformOperations{
				&TerraformOperation{
					ResponseSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: arrayWrapperType, IsArrayWrapper: true},
						{TypeDef: entityType},
					},
				},
			},
			expectedMethodNames: []string{
				"RefreshFromArrayOfOperationsListThingsResponse",
				"RefreshFromSharedThing",
			},
			expectedArrayFlags: []bool{true, false},
		},
		{
			name: "duplicate response targets across operations are deduplicated with first wins",
			ops: TerraformOperations{
				&TerraformOperation{
					ResponseSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: entityType, Optional: true},
					},
				},
				&TerraformOperation{
					ResponseSDKMethodTargets: []TerraformSDKMethodTarget{
						{TypeDef: entityType, Optional: false},
					},
				},
			},
			expectedMethodNames: []string{
				"RefreshFromSharedThing",
			},
			expectedArrayFlags: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			methods := tt.ops.SDKResponseMethods()

			var names []string
			var arrayFlags []bool

			for _, m := range methods {
				names = append(names, m.MethodName)
				arrayFlags = append(arrayFlags, m.Target.IsArrayWrapper)
			}

			assert.Equal(t, tt.expectedMethodNames, names)
			assert.Equal(t, tt.expectedArrayFlags, arrayFlags)
		})
	}
}

func TestTerraformOperations_SDKResponseMethods_Operation(t *testing.T) {
	t.Parallel()

	entityType := &TypeDef{
		Name:  "Thing",
		Type:  DataTypeClass,
		Scope: ScopeShared,
	}

	readOp := &TerraformOperation{
		EntityOperation: "Thing#read",
	}
	readOp.ResponseSDKMethodTargets = []TerraformSDKMethodTarget{
		{TypeDef: entityType, Operation: readOp},
	}

	createOp := &TerraformOperation{
		EntityOperation: "Thing#create",
	}
	createOp.ResponseSDKMethodTargets = []TerraformSDKMethodTarget{
		{TypeDef: entityType, Operation: createOp},
	}

	ops := TerraformOperations{readOp, createOp}
	methods := ops.SDKResponseMethods()

	require.Len(t, methods, 1)
	assert.Same(t, readOp, methods[0].Operation, "first operation should win dedup")
}

func TestTerraformOperation_setResponseSDKMethodTargets(t *testing.T) {
	t.Parallel()

	entityTypeDef := &TypeDef{
		Name: "Thing",
		Type: DataTypeClass,
		Extensions: &TypeDefExtensions{
			Entity: &extensions.Entity{Names: []string{"Thing"}},
		},
		Fields: []*FieldDef{
			{Name: "id", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
			{Name: "name", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
		},
	}

	schemaTypeDef := &TypeDef{
		Name:       "ThingSchema",
		Type:       DataTypeClass,
		Extensions: &TypeDefExtensions{},
		Fields: []*FieldDef{
			{Name: "id", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
			{Name: "name", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
		},
	}

	wrapperType := &TypeDef{
		Name:       "GetThingResponse",
		Type:       DataTypeClass,
		Extensions: &TypeDefExtensions{},
		Fields: []*FieldDef{
			{Name: "data", Type: entityTypeDef},
		},
	}

	incompatibleType := &TypeDef{
		Name: "Unrelated",
		Type: DataTypeClass,
		Extensions: &TypeDefExtensions{
			TerraformIgnore: &extensions.TerraformIgnore{
				DataModel: true,
			},
		},
	}

	arrayType := &TypeDef{
		Name:       "ThingList",
		Type:       DataTypeArray,
		Extensions: &TypeDefExtensions{},
		ItemType:   entityTypeDef,
	}

	sharedType := &TypeDef{
		Name:       "SharedWrapper",
		Type:       DataTypeClass,
		Extensions: &TypeDefExtensions{},
		Fields: []*FieldDef{
			{Name: "id", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
			{Name: "data", Type: entityTypeDef},
		},
	}

	tests := []struct {
		name               string
		op                 *TerraformOperation
		expectEmpty        bool
		expectTypeDefs     []*TypeDef // TypeDef pointers expected in targets
		expectArrayWrapper *TypeDef   // if non-nil, expect an IsArrayWrapper target with this TypeDef
	}{
		{
			name: "non-array response body includes compatible targets",
			op: &TerraformOperation{
				EntityName: "Thing",
				APIOperation: &Operation{
					BaseOperation: BaseOperation{Response: &Response{
						Type: wrapperType,
					}},
				},
				ResponseBodyFieldDef: &FieldDef{
					Name: "body",
					Type: wrapperType,
				},
			},
			expectTypeDefs: []*TypeDef{wrapperType, entityTypeDef},
		},
		{
			name: "filters out incompatible targets",
			op: &TerraformOperation{
				EntityName: "Thing",
				APIOperation: &Operation{
					BaseOperation: BaseOperation{Response: &Response{
						Type: incompatibleType,
					}},
				},
				ResponseBodyFieldDef: &FieldDef{
					Name: "body",
					Type: incompatibleType,
				},
			},
			expectEmpty: true,
		},
		{
			name: "entity array response includes array wrapper target",
			op: &TerraformOperation{
				EntityName:                "Thing",
				responseBodyIsEntityArray: true,
				APIOperation: &Operation{
					BaseOperation: BaseOperation{Response: &Response{
						Type: arrayType,
					}},
				},
				ResponseBodyFieldDef: &FieldDef{
					Name: "body",
					Type: arrayType,
				},
			},
			expectArrayWrapper: arrayType,
		},
		{
			name: "shared TypeDef pointers produce same result via cache",
			op: &TerraformOperation{
				EntityName: "Thing",
				APIOperation: &Operation{
					BaseOperation: BaseOperation{Response: &Response{
						Type: sharedType,
					}},
				},
				ResponseBodyFieldDef: &FieldDef{
					Name: "body",
					Type: sharedType,
				},
			},
			expectTypeDefs: []*TypeDef{sharedType},
		},
		{
			name: "nil response body with nil APIOperation response",
			op: &TerraformOperation{
				EntityName: "Thing",
			},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.op.setResponseSDKMethodTargets(schemaTypeDef)
			targets := tt.op.ResponseSDKMethodTargets

			if tt.expectEmpty {
				assert.Empty(t, targets)

				return
			}

			require.NotEmpty(t, targets)

			for _, expectedTypeDef := range tt.expectTypeDefs {
				var found bool

				for _, target := range targets {
					if target.TypeDef == expectedTypeDef {
						found = true

						break
					}
				}

				assert.True(t, found, "expected TypeDef %q to be included in targets", expectedTypeDef.Name)
			}

			if tt.expectArrayWrapper != nil {
				var found bool

				for _, target := range targets {
					if target.IsArrayWrapper {
						found = true
						assert.Equal(t, tt.expectArrayWrapper, target.TypeDef)
					}
				}

				assert.True(t, found, "expected array wrapper target")
			}
		})
	}
}
