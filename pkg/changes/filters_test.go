package changes

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
)

func TestSDKDiff_Filter(t *testing.T) {
	diff := SDKDiff{
		Changes: []MethodDiff{
			{MethodKey: "a", Type: MethodAdded},
			{MethodKey: "b", Type: MethodDeleted},
			{MethodKey: "c", Type: MethodAdded},
		},
	}

	filtered := diff.Filter(func(m MethodDiff) bool {
		return m.Type == MethodAdded
	})

	assert.Len(t, filtered.Changes, 2)
	assert.Equal(t, "a", filtered.Changes[0].MethodKey)
	assert.Equal(t, "c", filtered.Changes[1].MethodKey)
}

func TestSDKDiff_Filter_Empty(t *testing.T) {
	diff := SDKDiff{
		Changes: []MethodDiff{
			{MethodKey: "a", Type: MethodAdded},
		},
	}

	filtered := diff.Filter(func(m MethodDiff) bool {
		return false
	})

	assert.Nil(t, filtered.Changes)
}

func TestIsOperationMCPEnabled_NilOperation(t *testing.T) {
	m := MethodDiff{Operation: nil}
	assert.True(t, IsOperationMCPEnabled(m))
}

func TestIsOperationMCPEnabled_NoExtensions(t *testing.T) {
	m := MethodDiff{Operation: &ast.Operation{}}
	assert.True(t, IsOperationMCPEnabled(m))
}

func TestIsOperationMCPEnabled_NoMCPExtension(t *testing.T) {
	m := MethodDiff{Operation: &ast.Operation{
		Extensions: &ast.OperationExtensions{},
	}}
	assert.True(t, IsOperationMCPEnabled(m))
}

func TestIsOperationMCPEnabled_MCPEnabled(t *testing.T) {
	m := MethodDiff{Operation: &ast.Operation{
		Extensions: &ast.OperationExtensions{
			MCP: &extensions.MCP{Disabled: false},
		},
	}}
	assert.True(t, IsOperationMCPEnabled(m))
}

func TestIsOperationMCPEnabled_MCPDisabled(t *testing.T) {
	m := MethodDiff{Operation: &ast.Operation{
		Extensions: &ast.OperationExtensions{
			MCP: &extensions.MCP{Disabled: true},
		},
	}}
	assert.False(t, IsOperationMCPEnabled(m))
}
