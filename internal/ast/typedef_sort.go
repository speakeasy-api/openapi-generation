package ast

import (
	"slices"
	"sort"

	"github.com/speakeasy-api/openapi/sequencedmap"
)

// TopologicalSortTypeDefs sorts TypeDefs in topological order based on their
// dependencies. Children are placed before their parents.
//
// hoistCyclicUnions was introduced for pythonv2 and is set to true
// only when conflictResistantModelImportsFeb2026 is enabled. It pre-sorts the
// input so that any union which participates in a cycle with one of its variant
// types (typically a class whose field references back to the union) is visited
// before non-unions. Non-cyclic unions are left in their original input position
// to minimize churn in generated SDKs.
func TopologicalSortTypeDefs(types TypeDefs, hoistCyclicUnions bool) TypeDefs {
	if len(types) == 0 {
		return types
	}
	include := make(map[*TypeDef]bool)
	for _, t := range types {
		include[t] = true
	}
	ret := sequencedmap.New[*TypeDef, bool]()
	visited := make(map[*TypeDef]bool)

	var visit func(*TypeDef)
	visit = func(t *TypeDef) {
		if visited[t] {
			return
		}
		visited[t] = true
		for _, child := range t.Children {
			visit(child)
		}

		// Once all children are placed we can place this type
		if include[t] {
			ret.Set(t, true)
		}
	}

	if hoistCyclicUnions {
		cyclicUnions := collectCyclicUnions(types)
		if len(cyclicUnions) > 0 {
			types = slices.Clone(types)
			sort.SliceStable(types, func(i, j int) bool {
				return cyclicUnions[types[i]] && !cyclicUnions[types[j]]
			})
		}
	}

	for _, t := range types {
		visit(t)
	}

	return slices.Collect(ret.Keys())
}

// collectCyclicUnions returns the set of union TypeDefs that participate in a
// dependency cycle (i.e. a child path leads back to the union itself).
func collectCyclicUnions(types TypeDefs) map[*TypeDef]bool {
	cyclicUnions := make(map[*TypeDef]bool)
	for _, t := range types {
		if isCyclicUnion(t) {
			cyclicUnions[t] = true
		}
	}
	return cyclicUnions
}

// isCyclicUnion reports whether any descendant of t (via Children) is t itself
func isCyclicUnion(t *TypeDef) bool {
	if t.Type != DataTypeUnion {
		return false
	}

	seen := map[*TypeDef]bool{t: true}

	var dfs func(node *TypeDef) bool
	dfs = func(node *TypeDef) bool {
		for _, child := range node.Children {
			if child == t {
				return true
			}
			if seen[child] {
				continue
			}
			seen[child] = true
			if dfs(child) {
				return true
			}
		}
		return false
	}
	return dfs(t)
}
