package ast

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi/openapi"
)

// SDK represents the start of an SDKs Syntax tree.
type SDK struct {
	FieldName       string         `yaml:",omitempty"` // The name of the field used to access the SDK
	Type            *TypeDef       `yaml:",omitempty"` // The type definition for the SDK class
	Group           string         `yaml:",omitempty"` // The group that the SDK belongs to
	Servers         *Servers       `yaml:",omitempty"` // The defined list of servers the SDK has available to make requests to
	Comments        *Comment       `yaml:",omitempty"` // The documentation for the SDK
	SubSDKs         SDKs           `yaml:",omitempty"` // The list of sub SDKs that are available, containing operations scoped to a particular tag
	Security        *FieldDef      `yaml:",omitempty"` // The security definition for the SDK (if any and generally only populated for the main SDK)
	SecurityConfig  SecurityConfig `yaml:",omitempty"` // The security options for the SDK
	Operations      []*Operation   `yaml:",omitempty"` // The list of operations that are available to this SDK
	AdditionalTypes []*TypeDef     `yaml:",omitempty"` // A list of additional types that are available to the SDK, these are populated from components in the OpenAPI document that use the x-speakeasy-include extension and aren't already present elsewhere in the tree (generally only populated for the main SDK)
	OutputTests     bool           `yaml:",omitempty"` // Whether or not to output tests for the SDK
	TestGroup       string         `yaml:",omitempty"` // Which group of tests to generate (should match the spec file)
	Globals         *TypeDef       `yaml:",omitempty"` // The global variables that are available to the SDK
}

// Returns a new main SDK.
func NewMainSDK(typeDef *TypeDef) *SDK {
	return &SDK{
		Type: typeDef,
	}
}

// Returns a new sub SDK.
func NewSubSDK(typeDef *TypeDef, name string, group string) *SDK {
	return &SDK{
		FieldName: name,
		Group:     group,
		Type:      typeDef,
	}
}

// Returns the context stack for the SDK, including the SDK itself.
func (s *SDK) ChildContextStack() *ContextStack {
	result := s.Type.ContextStack.Clone()

	if len(result) == 0 {
		result.AppendMainSDK(s.Type.Name)
	} else {
		result.Append(ContextTypeGroup, s.Type.Name)
	}

	return &result
}

func (s *SDK) Match(matchers Matchers) error {
	if matchers.SDK != nil {
		return matchers.SDK(s)
	}

	return nil
}

// Find an operation by its ID in the SDK or its sub SDKs.
//
// Note: This function is used by templating.
func (s *SDK) FindOperation(operationID string) *Operation {
	for _, op := range s.Operations {
		if op.ID == operationID {
			return op
		}
	}

	for _, subSDK := range s.SubSDKs {
		op := subSDK.FindOperation(operationID)
		if op != nil {
			return op
		}
	}

	return nil
}

func (s *SDK) CountUniqueOperations() int64 {
	seenOperations := make(map[string]bool)
	return s.countUniqueOperations(seenOperations)
}

func (s *SDK) countUniqueOperations(seenOperations map[string]bool) int64 {
	numberOfOperations := int64(0)
	for _, op := range s.Operations {
		// Use OriginalID to avoid counting operation variants (like _xml, _raw) as separate operations
		// When an operation has multiple request body content types, variants are generated with different IDs,
		// but they all share the same OriginalID from the OpenAPI spec
		operationKey := op.OriginalID
		if operationKey == "" {
			// Fallback to ID if OriginalID is not set (shouldn't happen in normal flow)
			operationKey = op.ID
		}

		if !seenOperations[operationKey] {
			seenOperations[operationKey] = true
			numberOfOperations++
		}
	}

	for _, sdk := range s.SubSDKs {
		numberOfOperations += sdk.countUniqueOperations(seenOperations)
	}

	return numberOfOperations
}

// Adds the given operation(s) to the SDK operations, setting the OwningSDK.
func (s *SDK) AddOperations(operations ...Operation) {
	for _, operation := range operations {
		operation.OwningSDK = s

		s.Operations = append(s.Operations, &operation)
	}
}

// Returns true if the SDK has or its sub SDKs have underlying operations.
func (s *SDK) HasOperations() bool {
	if len(s.Operations) > 0 {
		return true
	}

	return s.SubSDKs.HasOperations()
}

// HasAnyOperationServers returns true if any operation in the SDK or its sub SDKs
// has operation-level servers defined. Used by templates to determine whether
// server_url should be optional in the constructor (when no global servers exist
// but some operations define their own).
func (s *SDK) HasAnyOperationServers() bool {
	for op := range s.WalkOperations() {
		if op.Servers != nil && len(op.Servers.Servers) > 0 {
			return true
		}
	}
	return false
}

func (s *SDK) WalkOperations() iter.Seq[*Operation] {
	return func(yield func(operation *Operation) bool) {
		if s == nil {
			return
		}
		var walk func(*SDK) bool
		walk = func(sdk *SDK) bool {
			if sdk == nil {
				return true
			}
			for _, op := range sdk.Operations {
				if op != nil && !yield(op) {
					return false
				}
			}
			for _, child := range sdk.SubSDKs {
				if !walk(child) {
					return false
				}
			}
			return true
		}
		walk(s)
	}
}

// Returns the identifier for the SDK.
func (s *SDK) ID(nameResolutionFeb2025 bool) string {
	// Preserve compatibility when an API uses distinct display and normalized
	// `x-speakeasy-groups` tag values. Using a typedef registration ID would be a
	// breaking change, so this uses a slightly modified ID instead.
	var id strings.Builder

	id.WriteString("scope:")
	id.WriteString(string(s.Type.Scope))
	id.WriteString(" ")

	if s.Type.ContextStack == nil {
		panic(fmt.Errorf("type %s has no context stack", s.Type.Name))
	}

	for _, frame := range s.Type.ContextStack {
		id.WriteString(string(frame.Type))
		id.WriteString(":")
		if frame.Type == ContextTypeGroup && nameResolutionFeb2025 {
			id.WriteString(strcase.ToPascal(sanitization.SanitizeName(frame.Identifier)))
		} else {
			id.WriteString(frame.Identifier)
		}
		id.WriteString(" ")
	}

	id.WriteString("originalName:")
	name := s.Type.Name
	if s.Type.OriginalName != "" {
		name = s.Type.OriginalName
	}

	name = sanitization.SanitizeName(name)
	name = strcase.ToGoPascal(name)
	id.WriteString(name)

	return id.String()
}

// Sorts sub SDKs by the ordering of the OpenAPI document tags. Generally called
// after checking the maintainTagBasedOrdering configuration.
func (s *SDK) SortSubSDKsByTags(tags []*openapi.Tag) {
	originalSubSDKs := slices.Clone(s.SubSDKs)
	sortedSubSDKs := make([]*SDK, 0, len(s.SubSDKs))

	for _, tag := range tags {
		sdk := originalSubSDKs.GetByFieldName(tag.Name)

		if sdk == nil {
			continue
		}

		sortedSubSDKs = append(sortedSubSDKs, sdk)
		originalSubSDKs.DeleteByFieldName(tag.Name)
	}

	// add any non listed tags for compatibility
	for _, sdk := range originalSubSDKs {
		sortedSubSDKs = append(sortedSubSDKs, sdk)
	}

	s.SubSDKs = sortedSubSDKs
}
