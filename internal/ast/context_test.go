package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextFrame_Clone(t *testing.T) {
	t.Parallel()

	identifierForNaming := "humanized"
	tests := []struct {
		name string
		cf   ContextFrame
		test func(t *testing.T, original, cloned ContextFrame)
	}{
		{
			name: "simple ContextFrame",
			cf: ContextFrame{
				Type:       ContextTypeOperation,
				Identifier: "operationId",
				Used:       true,
				MustUse:    false,
			},
			test: func(t *testing.T, original, cloned ContextFrame) {
				t.Helper()
				assert.Equal(t, original.Type, cloned.Type)
				assert.Equal(t, original.Identifier, cloned.Identifier)
				assert.Equal(t, original.Used, cloned.Used)
				assert.Equal(t, original.MustUse, cloned.MustUse)
				assert.Nil(t, cloned.IdentifierForNaming)
			},
		},
		{
			name: "ContextFrame with IdentifierForNaming",
			cf: ContextFrame{
				Type:                ContextTypeOperation,
				Identifier:          "operation_id",
				IdentifierForNaming: &identifierForNaming,
				Used:                false,
				MustUse:             true,
			},
			test: func(t *testing.T, original, cloned ContextFrame) {
				t.Helper()
				assert.Equal(t, original.Type, cloned.Type)
				assert.Equal(t, original.Identifier, cloned.Identifier)
				assert.Equal(t, original.Used, cloned.Used)
				assert.Equal(t, original.MustUse, cloned.MustUse)
				require.NotNil(t, cloned.IdentifierForNaming)
				assert.NotSame(t, original.IdentifierForNaming, cloned.IdentifierForNaming)
				assert.Equal(t, *original.IdentifierForNaming, *cloned.IdentifierForNaming)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.cf.Clone()
			tt.test(t, tt.cf, cloned)
		})
	}
}

func TestContextStack_Clone(t *testing.T) {
	t.Parallel()

	identifierForNaming := "humanized"
	tests := []struct {
		name  string
		stack ContextStack
		test  func(t *testing.T, original, cloned ContextStack)
	}{
		{
			name:  "nil ContextStack",
			stack: nil,
			test: func(t *testing.T, original, cloned ContextStack) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:  "empty ContextStack",
			stack: ContextStack{},
			test: func(t *testing.T, original, cloned ContextStack) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				assert.Empty(t, cloned)
			},
		},
		{
			name: "ContextStack with single frame",
			stack: ContextStack{
				{
					Type:       ContextTypeOperation,
					Identifier: "op1",
					Used:       true,
				},
			},
			test: func(t *testing.T, original, cloned ContextStack) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)
				assert.Equal(t, original[0].Type, cloned[0].Type)
				assert.Equal(t, original[0].Identifier, cloned[0].Identifier)
				assert.Equal(t, original[0].Used, cloned[0].Used)
			},
		},
		{
			name: "ContextStack with multiple frames",
			stack: ContextStack{
				{
					Type:       ContextTypeOperation,
					Identifier: "op1",
					Used:       true,
				},
				{
					Type:                ContextTypeRequestBody,
					Identifier:          "requestBody",
					IdentifierForNaming: &identifierForNaming,
					MustUse:             true,
				},
				{
					Type:       ContextTypeProperty,
					Identifier: "prop1",
				},
			},
			test: func(t *testing.T, original, cloned ContextStack) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 3)

				for i := range original {
					assert.Equal(t, original[i].Type, cloned[i].Type)
					assert.Equal(t, original[i].Identifier, cloned[i].Identifier)
					assert.Equal(t, original[i].Used, cloned[i].Used)
					assert.Equal(t, original[i].MustUse, cloned[i].MustUse)

					if original[i].IdentifierForNaming != nil {
						require.NotNil(t, cloned[i].IdentifierForNaming)
						assert.NotSame(t, original[i].IdentifierForNaming, cloned[i].IdentifierForNaming)
						assert.Equal(t, *original[i].IdentifierForNaming, *cloned[i].IdentifierForNaming)
					} else {
						assert.Nil(t, cloned[i].IdentifierForNaming)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.stack.Clone()
			tt.test(t, tt.stack, cloned)
		})
	}
}

func TestContextStacks_Clone(t *testing.T) {
	t.Parallel()

	identifierForNaming := "humanized"
	tests := []struct {
		name   string
		stacks ContextStacks
		test   func(t *testing.T, original, cloned ContextStacks)
	}{
		{
			name:   "nil ContextStacks",
			stacks: nil,
			test: func(t *testing.T, original, cloned ContextStacks) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:   "empty ContextStacks",
			stacks: ContextStacks{},
			test: func(t *testing.T, original, cloned ContextStacks) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				assert.Empty(t, cloned)
			},
		},
		{
			name: "ContextStacks with single stack",
			stacks: ContextStacks{
				{
					{
						Type:       ContextTypeOperation,
						Identifier: "op1",
					},
				},
			},
			test: func(t *testing.T, original, cloned ContextStacks) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)
				require.Len(t, cloned[0], 1)
				assert.Equal(t, original[0][0].Type, cloned[0][0].Type)
				assert.Equal(t, original[0][0].Identifier, cloned[0][0].Identifier)
			},
		},
		{
			name: "ContextStacks with multiple stacks",
			stacks: ContextStacks{
				{
					{
						Type:       ContextTypeOperation,
						Identifier: "op1",
						Used:       true,
					},
					{
						Type:       ContextTypeRequestBody,
						Identifier: "body1",
					},
				},
				{
					{
						Type:                ContextTypeResponseBody,
						Identifier:          "resp1",
						IdentifierForNaming: &identifierForNaming,
					},
				},
				{
					{
						Type:       ContextTypeProperty,
						Identifier: "prop1",
						MustUse:    true,
					},
					{
						Type:       ContextTypeProperty,
						Identifier: "prop2",
					},
				},
			},
			test: func(t *testing.T, original, cloned ContextStacks) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 3)

				for i, stack := range original {
					require.Len(t, cloned[i], len(stack))
					assert.NotSame(t, &stack, &cloned[i])

					for j, frame := range stack {
						assert.Equal(t, frame.Type, cloned[i][j].Type)
						assert.Equal(t, frame.Identifier, cloned[i][j].Identifier)
						assert.Equal(t, frame.Used, cloned[i][j].Used)
						assert.Equal(t, frame.MustUse, cloned[i][j].MustUse)

						if frame.IdentifierForNaming != nil {
							require.NotNil(t, cloned[i][j].IdentifierForNaming)
							assert.NotSame(t, frame.IdentifierForNaming, cloned[i][j].IdentifierForNaming)
							assert.Equal(t, *frame.IdentifierForNaming, *cloned[i][j].IdentifierForNaming)
						} else {
							assert.Nil(t, cloned[i][j].IdentifierForNaming)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.stacks.Clone()
			tt.test(t, tt.stacks, cloned)
		})
	}
}
