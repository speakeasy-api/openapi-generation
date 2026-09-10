package ast

import (
	"slices"
)

// Collection of SDK, such as sub SDKs.
type SDKs []*SDK

// Deletes the SDK with the given field name, if found.
func (s *SDKs) DeleteByFieldName(fieldName string) {
	for index, sdk := range *s {
		if sdk.FieldName == fieldName {
			*s = slices.Delete(*s, index, index+1)

			return
		}
	}
}

// Returns the SDK with the given field name, or nil if not found.
func (s SDKs) GetByFieldName(fieldName string) *SDK {
	for _, sdk := range s {
		if sdk.FieldName == fieldName {
			return sdk
		}
	}

	return nil
}

// Returns true if any SDK has operations.
func (s SDKs) HasOperations() bool {
	for _, sdk := range s {
		if sdk.HasOperations() {
			return true
		}
	}

	return false
}
