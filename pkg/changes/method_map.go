package changes

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

// methodInfo holds information about an operation for comparison
type methodInfo struct {
	Operation *ast.Operation
	MethodKey string // The hierarchical SubSDK path like "foo.bar.baz"
}

// buildMethodMap recursively builds an ordered map of operations with their hierarchical paths
func buildMethodMap(sdk *ast.SDK, pathPrefix string) *sequencedmap.Map[string, *methodInfo] {
	operations := sequencedmap.New[string, *methodInfo]()

	if sdk == nil {
		return operations
	}

	// Add operations from current SDK - use raw operation IDs without formatting
	for _, op := range sdk.Operations {
		opName := op.GetID()
		if opName == "" {
			opName = "unknown"
		}

		fullPath := pathPrefix
		if fullPath != "" {
			fullPath += "." + opName
		} else {
			fullPath = opName
		}

		operations.Set(fullPath, &methodInfo{
			Operation: op,
			MethodKey: fullPath,
		})
	}

	// Recursively add operations from sub-SDKs - use raw field names
	for _, subSDK := range sdk.SubSDKs {
		subSDKName := subSDK.FieldName
		if subSDKName == "" {
			subSDKName = "unknown"
		}

		newPrefix := pathPrefix
		if newPrefix != "" {
			newPrefix += "." + subSDKName
		} else {
			newPrefix = subSDKName
		}

		subOps := buildMethodMap(subSDK, newPrefix)
		for key, value := range subOps.All() {
			operations.Set(key, value)
		}
	}

	return operations
}
