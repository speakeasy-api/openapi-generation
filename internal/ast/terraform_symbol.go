package ast

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// maxSymbolNameCandidates is the maximum number of numeric suffix candidates
// tried when resolving symbol name collisions (e.g., Type, Type1, Type2, ...).
const maxSymbolNameCandidates = 1000

// TerraformSymbolName derives the candidate symbol name for a TypeDef by
// sanitizing the TypeDef name via TerraformGoTypeName and stripping Input/Output
// suffixes when the TypeDef is marked as an input or output type respectively.
func TerraformSymbolName(t *TypeDef) string {
	name := TerraformGoTypeName(t.Name)

	if t.Input && strings.HasSuffix(name, "Input") {
		return strings.TrimSuffix(name, "Input")
	}

	if t.Output && strings.HasSuffix(name, "Output") {
		return strings.TrimSuffix(name, "Output")
	}

	return name
}

// symbolManager tracks assigned symbol names and assigns unique deduplicated
// symbols to complex types (class and union). Structurally identical types may
// share a symbol to reduce generated code duplication.
type symbolManager struct {
	// assignedSymbols maps assigned symbol names to their TypeDef pointers.
	// Reserved names are seeded with nil values to prevent collisions.
	assignedSymbols map[string]*TypeDef

	// candidateNames maps TypeDef pointers to their TerraformSymbolName result,
	// avoiding redundant string processing for shared type references.
	candidateNames map[*TypeDef]string

	enableTypeDeduplication bool

	// symbolsByCandidate indexes assigned symbol names by their candidate name
	// (TerraformSymbolName result) for O(k) non-dedup lookup, where k is
	// the number of types sharing the same candidate name (typically 1).
	// Only populated and consulted in non-dedup mode; in dedup mode, symbol
	// reuse checks all assigned symbols by structural equality regardless of
	// candidate name, so the index is not needed for lookup. New allocations
	// update this index unconditionally so it stays consistent with
	// assignedSymbols for any future non-dedup queries.
	symbolsByCandidate map[string][]string
}

// newSymbolManager creates a symbolManager with the given initial symbol
// assignments. Any non-nil TypeDefs in assignedSymbols are indexed into
// symbolsByCandidate for efficient non-dedup lookup. Nil-valued entries (e.g.
// reserved name placeholders) are skipped during indexing.
func newSymbolManager(assignedSymbols map[string]*TypeDef, enableTypeDeduplication bool) *symbolManager {
	sm := &symbolManager{
		assignedSymbols:         assignedSymbols,
		candidateNames:          make(map[*TypeDef]string),
		enableTypeDeduplication: enableTypeDeduplication,
		symbolsByCandidate:      make(map[string][]string),
	}

	for assignedName, td := range assignedSymbols {
		if td == nil {
			continue
		}

		candidateName := TerraformSymbolName(td)
		sm.candidateNames[td] = candidateName
		sm.symbolsByCandidate[candidateName] = append(sm.symbolsByCandidate[candidateName], assignedName)
	}

	return sm
}

// assignSymbol assigns a unique symbol name to the given TypeDef via its
// Extensions. If the type already has a symbol, this is a no-op. Otherwise,
// the method searches for a structurally identical type to reuse, or allocates
// a new name with numeric suffix for collisions.
func (sm *symbolManager) assignSymbol(fieldName string, t *TypeDef) error {
	t.EnsureExtensions()

	if _, ok := t.Extensions.Get("Symbol"); ok {
		return nil
	}

	name := sm.cachedSymbolName(t)
	if name == "" {
		name = TerraformGoTypeName(fieldName)
	}

	if sm.enableTypeDeduplication {
		// Dedup mode: check ALL assigned symbols for structural equality.
		// Keys must be sorted for deterministic symbol assignment, since Go map
		// iteration order is random unlike JavaScript's insertion-order iteration.
		for _, existingName := range slices.Sorted(maps.Keys(sm.assignedSymbols)) {
			existingTypeDef := sm.assignedSymbols[existingName]
			if existingTypeDef == nil {
				continue
			}

			if existingTypeDef.IsTerraformSymbolEqual(t) {
				t.Extensions.Set("Symbol", existingName)

				return nil
			}
		}
	} else {
		// Non-dedup mode: only check symbols with the same candidate name,
		// reducing O(n) iteration to O(k) where k is the number of types
		// sharing the same candidate name (typically 1).
		for _, assignedName := range sm.symbolsByCandidate[name] {
			existingTypeDef := sm.assignedSymbols[assignedName]
			if existingTypeDef.IsTerraformSymbolEqual(t) {
				t.Extensions.Set("Symbol", assignedName)

				return nil
			}
		}
	}

	// Allocate the next available name with numeric suffix for collisions.
	for i := range maxSymbolNameCandidates {
		option := name
		if i > 0 {
			option = name + strconv.Itoa(i)
		}

		if _, taken := sm.assignedSymbols[option]; !taken {
			sm.assignedSymbols[option] = t
			sm.symbolsByCandidate[name] = append(sm.symbolsByCandidate[name], option)
			t.Extensions.Set("Symbol", option)

			return nil
		}
	}

	return fmt.Errorf("failed to assign symbol for type %q: exhausted %d name candidates", name, maxSymbolNameCandidates)
}

// cachedSymbolName returns TerraformSymbolName(t), computing and caching the
// result on first call to avoid redundant string processing for shared types.
func (sm *symbolManager) cachedSymbolName(t *TypeDef) string {
	if cached, ok := sm.candidateNames[t]; ok {
		return cached
	}

	name := TerraformSymbolName(t)
	sm.candidateNames[t] = name

	return name
}

// annotateTerraformSymbols walks a single entity SchemaTypeDef, processing
// AssociatedTypes first, then Fields.
func (t *TypeDef) annotateTerraformSymbols(sm *symbolManager) error {
	for _, associatedTypeDef := range t.AssociatedTypes {
		unionName := TerraformGoTypeName(t.AssociatedTypeName(associatedTypeDef))

		if err := sm.annotateRecursive(associatedTypeDef, unionName); err != nil {
			return err
		}
	}

	for _, fieldDef := range t.Fields {
		if fieldDef.HasMatchConfigPath() {
			continue
		}

		if err := sm.annotateRecursive(fieldDef.Type, fieldDef.Name); err != nil {
			return err
		}
	}

	return nil
}

// annotateRecursive walks a TypeDef tree depth-first and assigns symbols to
// complex types (class, union).
func (sm *symbolManager) annotateRecursive(t *TypeDef, fieldName string) error {
	if t == nil {
		return nil
	}

	switch t.Type {
	case DataTypeMap:
		// Maps with custom type config are treated as primitives (no recursion).
		if t.Extensions != nil && t.Extensions.Has("x-speakeasy-terraform-custom-type") {
			return nil
		}

		return sm.annotateRecursive(t.ItemType, "")
	case DataTypeArray, DataTypeSet:
		// Arrays/sets with custom type config are treated as primitives.
		if t.Extensions != nil && t.Extensions.Has("x-speakeasy-terraform-custom-type") {
			return nil
		}

		if t.ItemType == nil {
			return nil
		}

		if t.ItemType.IsTerraformPrimitiveType() {
			return nil
		}

		if t.ItemType.Type == DataTypeClass && len(t.ItemType.Fields) > 0 {
			// Array/set of class: assign symbol on ItemType, recurse into fields.
			if err := sm.assignSymbol(fieldName, t.ItemType); err != nil {
				return err
			}

			return sm.annotateClassFields(t.ItemType)
		}

		if len(t.ItemType.AssociatedTypes) > 0 {
			// Array/set of union: assign symbol on ItemType, recurse into
			// associated types and hoisted fields.
			if err := sm.assignSymbol(fieldName, t.ItemType); err != nil {
				return err
			}

			return sm.annotateUnionMembers(t.ItemType)
		}

		// Catch-all: recurse into ItemType for nested collections (e.g.,
		// array of map of class) and other non-primitive item types.
		return sm.annotateRecursive(t.ItemType, fieldName)
	}

	if t.Type == DataTypeClass {
		if err := sm.assignSymbol(fieldName, t); err != nil {
			return err
		}

		return sm.annotateClassFields(t)
	}

	if len(t.AssociatedTypes) > 0 {
		if err := sm.assignSymbol(fieldName, t); err != nil {
			return err
		}

		return sm.annotateUnionMembers(t)
	}

	return nil
}

// annotateClassFields recurses into each field of a class TypeDef, skipping
// fields with path-only match config aliases.
func (sm *symbolManager) annotateClassFields(t *TypeDef) error {
	for _, field := range t.Fields {
		if field.HasMatchConfigPath() {
			continue
		}

		if err := sm.annotateRecursive(field.Type, field.Name); err != nil {
			return err
		}
	}

	return nil
}

// annotateUnionMembers recurses into each associated type of a union TypeDef,
// then into hoisted fields (fields merged onto the union type).
func (sm *symbolManager) annotateUnionMembers(t *TypeDef) error {
	for _, subtype := range t.AssociatedTypes {
		unionName := TerraformGoTypeName(t.AssociatedTypeName(subtype))

		if err := sm.annotateRecursive(subtype, unionName); err != nil {
			return err
		}
	}

	// Hoisted fields (fields that were merged to the parent union type).
	return sm.annotateClassFields(t)
}
