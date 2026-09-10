package changes

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// isOperationDeprecated checks if an operation is marked as deprecated
func isOperationDeprecated(operation *ast.Operation) bool {
	return operation.Comments != nil && operation.Comments.Deprecated
}

// DiffASTs compares two ASTs and returns the raw differences
func DiffASTs(options DiffOptions) SDKDiff {
	oldAST := options.OldAST
	newAST := options.NewAST
	var methodChanges []MethodDiff

	// Handle nil ASTs gracefully
	if oldAST == nil || newAST == nil {
		return SDKDiff{
			Changes:      methodChanges,
			OldAST:       oldAST,
			NewAST:       newAST,
			OldSubsystem: options.OldSubsystem,
			NewSubsystem: options.NewSubsystem,
		}
	}

	// Build method maps without any formatting
	oldMethods := buildMethodMap(oldAST.MainSDK, "")
	newMethods := buildMethodMap(newAST.MainSDK, "")

	// Find added operations (iterate in order)
	for subSDKPath, opInfo := range newMethods.All() {
		if _, exists := oldMethods.Get(subSDKPath); !exists {
			methodChanges = append(methodChanges, MethodDiff{
				Type:        MethodAdded,
				Operation:   opInfo.Operation,
				MethodParts: strings.Split(subSDKPath, "."),
				MethodKey:   subSDKPath,
			})
		}
	}

	// Find deleted operations (iterate in order)
	for subSDKPath, opInfo := range oldMethods.All() {
		if _, exists := newMethods.Get(subSDKPath); !exists {
			methodChanges = append(methodChanges, MethodDiff{
				Type:        MethodDeleted,
				Operation:   opInfo.Operation,
				MethodParts: strings.Split(subSDKPath, "."),
				MethodKey:   subSDKPath,
			})
		}
	}

	// Find modified operations (iterate in order)
	for subSDKPath, aOpInfo := range oldMethods.All() {
		if bOpInfo, exists := newMethods.Get(subSDKPath); exists {
			// Check for deprecation changes
			wasDeprecated := isOperationDeprecated(aOpInfo.Operation)
			isDeprecated := isOperationDeprecated(bOpInfo.Operation)

			if !wasDeprecated && isDeprecated {
				// Operation became deprecated
				methodChanges = append(methodChanges, MethodDiff{
					Type:        MethodDeprecated,
					Operation:   bOpInfo.Operation,
					MethodParts: strings.Split(subSDKPath, "."),
					MethodKey:   subSDKPath,
				})
			}

			// Check for other modifications (arguments/response changes)
			if methodDiff := diffMethod(aOpInfo, bOpInfo); methodDiff.HasChanges {
				// Determine the change type based on what actually changed
				var changeType ChangeType
				hasResponseChanges := !methodDiff.SuccessResponseResult.Equal || !methodDiff.ErrorResponseResult.Equal
				switch {
				case !methodDiff.ArgumentsResult.Equal && hasResponseChanges:
					changeType = ArgumentsAndResponseChanged
				case !methodDiff.ArgumentsResult.Equal:
					changeType = ArgumentsChanged
				case hasResponseChanges:
					changeType = ResponseChanged
				}

				methodChanges = append(methodChanges, MethodDiff{
					Type:                changeType,
					Operation:           bOpInfo.Operation,
					MethodParts:         strings.Split(subSDKPath, "."),
					MethodKey:           subSDKPath,
					ArgumentsDiff:       methodDiff.ArgumentsResult,
					SuccessResponseDiff: methodDiff.SuccessResponseResult,
					ErrorResponseDiff:   methodDiff.ErrorResponseResult,
				})
			}
		}
	}

	return SDKDiff{
		Changes:      methodChanges,
		OldAST:       oldAST,
		NewAST:       newAST,
		OldSubsystem: options.OldSubsystem,
		NewSubsystem: options.NewSubsystem,
	}
}
