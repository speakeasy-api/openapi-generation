package namer

import (
	"slices"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestRenameTypesWithDuplicateNames(t *testing.T) {
	tests := []struct {
		name     string
		typeDefs ast.TypeDefs
		expected []string
		only     bool
	}{
		{
			name: "Example 1",
			typeDefs: ast.TypeDefs{
				{
					Name:         "User",
					OriginalName: "User",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeRefType, Identifier: "User", Used: false},
						{Type: ast.ContextTypeResponseStatusCode, Identifier: "BadRequest", Used: false},
					},
					Scope: ast.ScopeShared,
					Type:  ast.DataTypeClass,
				},
				{
					Name:         "User",
					OriginalName: "User",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeRefType, Identifier: "User", Used: false},
						{Type: ast.ContextTypeResponseStatusCode, Identifier: "BadRequest", Used: false},
					},
					Scope: ast.ScopeErrors,
					Type:  ast.DataTypeClass,
				},
			},
			expected: []string{"User", "errors_User"},
		},
		{
			name: "Example 2",
			typeDefs: ast.TypeDefs{
				{
					Name:         "User",
					OriginalName: "User",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeOperation, Identifier: "getUser", Used: false},
						{Type: ast.ContextTypeRequestResponse, Identifier: "request", Used: false},
					},
					Scope: ast.ScopeOperations,
				},
				{
					Name:         "User",
					OriginalName: "User",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeOperation, Identifier: "updateUser", Used: false},
						{Type: ast.ContextTypeRequestResponse, Identifier: "response", Used: false},
					},
					Scope: ast.ScopeOperations,
				},
			},
			expected: []string{"getUser_User", "updateUser_User"},
		},
		{
			name: "Example 3",
			typeDefs: ast.TypeDefs{
				{
					Name:         "SingleType",
					OriginalName: "SingleType",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeRefType, Identifier: "Ref", Used: true},
					},
					Scope: ast.ScopeShared,
				},
			},
			expected: []string{"SingleType"},
		},
		{
			name: "Example 4",
			typeDefs: ast.TypeDefs{
				{
					Name:         "X",
					OriginalName: "X",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeProperty, Identifier: "a", Used: false},
					},
				},
				{
					Name:         "X",
					OriginalName: "X",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeProperty, Identifier: "b", Used: false},
					},
				},
			},
			expected: []string{"a_X", "b_X"},
		},
		{
			name: "Example 5",
			typeDefs: ast.TypeDefs{
				{
					Name:         "X",
					OriginalName: "X",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeProperty, Identifier: "a", Used: false},
						{Type: ast.ContextTypeProperty, Identifier: "b", Used: false},
					},
				},
				{
					Name:         "X",
					OriginalName: "X",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeProperty, Identifier: "b", Used: false},
						{Type: ast.ContextTypeProperty, Identifier: "c", Used: false},
					},
				},
				{
					Name:         "X",
					OriginalName: "X",
					ContextStack: ast.ContextStack{
						{Type: ast.ContextTypeProperty, Identifier: "c", Used: false},
						{Type: ast.ContextTypeProperty, Identifier: "d", Used: false},
					},
				},
			},
			expected: []string{"a_X", "b_c_X", "d_X"},
		},
	}

	hasOnly := false
	for _, tt := range tests {
		if tt.only {
			hasOnly = true
		}
	}

	subsystem := &subsystem.Subsystem{
		Config: &configuration.Config{},
	}

	resolver, err := NewResolver(subsystem, nil)
	require.NoError(t, err)

	for _, tt := range tests {
		if hasOnly && !tt.only {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx = logging.With(ctx, logging.NewLogger(zapcore.DebugLevel))
			err := resolver.RenameTypesWithDuplicateNames(ctx, tt.typeDefs)
			require.NoError(t, err)

			names := make([]string, len(tt.typeDefs))
			for i, td := range tt.typeDefs {
				names[i] = td.Name
			}

			assert.Equal(t, tt.expected, names)
		})
	}
}

func TestCollectLabels(t *testing.T) {
	mockTypeDef := &ast.TypeDef{
		ContextStack: ast.ContextStack{
			{Type: ast.ContextTypeRefType, Identifier: "User", Used: false},
			{Type: ast.ContextTypeOperationTag, Identifier: "Admin", Used: false},
			{Type: ast.ContextTypeOperation, Identifier: "getUsers", Used: false},
			{Type: ast.ContextTypeRequestResponse, Identifier: "request", Used: false},
		},
		Scope: ast.ScopeShared,
		Type:  ast.DataTypeEnum,
	}

	subsystem := &subsystem.Subsystem{
		Config: &configuration.Config{},
	}

	resolver, err := NewResolver(subsystem, nil)
	require.NoError(t, err)

	result := resolver.getOrCollectLabels(mockTypeDef)

	expectedLabels := []string{
		"shared",   // From Scope
		"enum",     // From DataType
		"getUsers", // From Operation
		"request",  // From RequestResponse
		"Admin",    // From OperationTag context
		"User",     // From RefType context
	}

	assert.ElementsMatch(t, expectedLabels, result.allLabels())
	i := slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "User"
	})
	assert.NotEqual(t, -1, i)
	l := result.All[i]
	assert.Equal(t, labelSource(ast.ContextTypeRefType), l.Source)
	i = slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "Admin"
	})
	assert.NotEqual(t, -1, i)
	l = result.All[i]
	assert.Equal(t, labelSource(ast.ContextTypeOperationTag), l.Source)
	i = slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "getUsers"
	})
	assert.NotEqual(t, -1, i)
	l = result.All[i]
	assert.Equal(t, labelSource(ast.ContextTypeOperation), l.Source)
	i = slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "request"
	})
	assert.NotEqual(t, -1, i)
	l = result.All[i]
	assert.Equal(t, labelSource(ast.ContextTypeRequestResponse), l.Source)
	i = slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "shared"
	})
	assert.NotEqual(t, -1, i)
	l = result.All[i]
	assert.Equal(t, labelSource("scope"), l.Source)
	i = slices.IndexFunc(result.All, func(l LabelWithSource) bool {
		return l.Label == "enum"
	})
	assert.NotEqual(t, -1, i)
	l = result.All[i]
	assert.Equal(t, labelSource("data_type"), l.Source)
}
