package namer

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// nameFormat determines how the final name parts will be concatenated to form the final name
var nameFormat = []labelSource{
	labelSource("scope"),
	labelSource(ast.ContextTypeGroup),
	labelSource(ast.ContextTypeOperationTag),
	labelSource(ast.ContextTypeOperation),
	labelSource(ast.ContextTypeRefName),
	labelSource(ast.ContextTypeOneOf),
	labelSource(ast.ContextTypeResponseStatusCode),
	labelSource("..."),           // Any other label source
	labelSource("original_name"), // The original name before any renaming
	labelSource(ast.ContextTypeConstProperty),
	labelSource(ast.ContextTypeRefType),
	labelSource(ast.ContextTypeRequestResponse),
	labelSource(ast.ContextTypeInputOutput),
	labelSource(ast.ContextTypeRequestBody),
	labelSource(ast.ContextTypeResponseBody),
	labelSource("data_type"),
	labelSource("duplicate_count"),
}

// discriminatorPriority is used for tie breaking which label should be used in the name,
// this is for the edge case where the same label appears in multiple sources
var discriminatorPriority = []labelSource{
	labelSource("scope"),
	labelSource(ast.ContextTypeOperation),
	labelSource(ast.ContextTypeGroup),
	labelSource(ast.ContextTypeOperationTag),
	labelSource(ast.ContextTypeRequestResponse),
	labelSource("data_type"),
	labelSource(ast.ContextTypeResponseStatusCode),
	labelSource(ast.ContextTypeRefType),
	labelSource(ast.ContextTypeRefName),
	labelSource(ast.ContextTypeOneOf),
	labelSource(ast.ContextTypeProperty),
	labelSource(ast.ContextTypeInputOutput),
	labelSource("..."),                        // Any other label source
	labelSource(ast.ContextTypeConstProperty), // Const properties are often a good label but can result in breaking changes
}

func (r *Resolver) RenameTypesWithDuplicateNames(ctx context.Context, typeDefs ast.TypeDefs) error {
	if len(typeDefs) == 0 {
		return nil
	}

	logging.From(ctx).Debug(fmt.Sprintf("\n\n--- Renaming %d types with name %q ---", len(typeDefs), typeDefs[0].Name))

	// Collect the labels for each TypeDef and keep track of where each label came from
	for _, t := range typeDefs {
		r.getOrCollectLabels(t)
	}

	// Track the priority of each label globally - used to break ties between labels
	allLabelsToBestSource := make(map[string]labelSource)
	for _, t := range typeDefs {
		typeDefLabels := r.getOrCollectLabels(t)
		for _, l := range typeDefLabels.All {
			existingSource, exists := allLabelsToBestSource[l.Label]
			if !exists || higherPriority(l.Source, existingSource, discriminatorPriority) {
				allLabelsToBestSource[l.Label] = l.Source
			}
		}
	}

	// Define tieBreaker, if multiple labels are similarly powerful, use the sourcePriority to break the tie
	tieBreaker := func(a, b string) string {
		aSource := allLabelsToBestSource[a]
		bSource := allLabelsToBestSource[b]
		aIdx := slices.Index(discriminatorPriority, aSource)
		bIdx := slices.Index(discriminatorPriority, bSource)
		restIdx := slices.Index(discriminatorPriority, labelSource("..."))
		// If the label is not in the sourcePriority, then it is a "..." label
		if aIdx == -1 {
			aIdx = restIdx
		}
		if bIdx == -1 {
			bIdx = restIdx
		}

		if aIdx < bIdx {
			return a
		} else if aIdx > bIdx {
			return b
		}

		return "" // Indicates no preference
	}

	// Collect the labels for the discriminator algorithm
	types := make([][]string, len(typeDefs))
	for i, t := range typeDefs {
		types[i] = r.getOrCollectLabels(t).availableLabels()
	}
	discriminators := FindBestDiscriminators(types, tieBreaker)
	discriminators = postProcessDiscriminators(discriminators)

	// Apply the chosen discriminators to the TypeDefs to build the new names
	for i, t := range typeDefs {
		// Add the discriminators to the selected labels
		if len(discriminators[i]) > 0 {
			r.getOrCollectLabels(t).selectLabels(discriminators[i])
			r.renameTypeDef(ctx, t, "discriminated")
		}
		if t.Name == "" {
			// Edge case 1: All names are empty and all labels are the same
			r.handleTypeWithNoName(ctx, t)
		}
	}

	r.renameConflictingChildren(ctx, typeDefs)
	// Last resort a subset of the types have identical labels
	r.incrementDuplicateNames(ctx, typeDefs)

	return nil
}

func postProcessDiscriminators(allDiscriminators [][]string) [][]string {
	// --- Avoid renaming the union type if possible ---
	// If exactly one of the discriminators is ["union"] and all other discriminators have at least one label -> remove Union
	indexUnion := -1
	countUnion := 0
	othersHaveLabel := true
	for i, discriminators := range allDiscriminators {
		if len(discriminators) == 1 && discriminators[0] == "union" {
			indexUnion = i
			countUnion++
		} else {
			othersHaveLabel = othersHaveLabel && len(discriminators) > 0
		}
	}
	if countUnion != 1 || indexUnion == -1 || !othersHaveLabel {
		return allDiscriminators
	}

	allDiscriminators[indexUnion] = []string{}

	return allDiscriminators
}

func (r *Resolver) renameTypeDef(ctx context.Context, t *ast.TypeDef, reason string) {
	typeDefLabels := r.getOrCollectLabels(t)

	// Group the selected labels by source - map[labelSource]Set[label]
	selectedBySource := map[labelSource][]string{}
	for _, l := range typeDefLabels.Selected {
		source := l.Source

		// Group all labels that are not explicitly ordered in the
		// nameFormat into the "..." source
		if !slices.Contains(nameFormat, source) {
			source = "..."
		}

		labels := selectedBySource[source]
		labels = append(labels, l.Label)
		selectedBySource[source] = labels
	}

	var parts []string
	for _, source := range nameFormat {
		labels := selectedBySource[source]
		if len(labels) == 0 {
			continue
		}

		if source == labelSource("scope") {
			for _, label := range labels {
				// Never rename components with "shared_"
				if label != string(ast.ScopeShared) {
					parts = append(parts, label)
				}
			}
			continue
		}

		for i, frame := range t.ContextStack {
			// Mark the label as used in the context stack
			idx := slices.Index(labels, frame.DisplayName())
			isRightSource := frame.Type == ast.ContextType(source) || source == "..."
			if idx != -1 && isRightSource {
				parts = append(parts, frame.DisplayName())
				t.ContextStack[i].Used = true
				labels = slices.Delete(labels, idx, idx+1)
			}
		}

		// Add any remaining labels which weren't from the context stack
		parts = append(parts, labels...)
	}

	oldName := t.Name
	t.Name = strings.Join(parts, "_")
	isComponent := ""
	if t.IsComponent {
		isComponent = " component"
	}
	if oldName != t.Name {
		logging.From(ctx).Debug(fmt.Sprintf("%s: Renamed%s to %q \n\t labels: %q \n\t registrationID: %q", reason, isComponent, t.Name, typeDefLabels.allLabelsString(), t.GetRegistrationID()))
	} else {
		logging.From(ctx).Debug(fmt.Sprintf("%s: N/A rename%s %q \n\t labels: %q \n\t registrationID: %q", reason, isComponent, t.Name, typeDefLabels.allLabelsString(), t.GetRegistrationID()))
	}
}

func higherPriority(a, b labelSource, priorityOrder []labelSource) bool {
	aIdx := slices.Index(priorityOrder, a)
	bIdx := slices.Index(priorityOrder, b)
	return aIdx < bIdx
}

func (r *Resolver) renameConflictingChildren(ctx context.Context, typeDefs ast.TypeDefs) {
	groupedByNormalizedName := make(map[string]ast.TypeDefs)
	for _, t := range typeDefs {
		groupedByNormalizedName[normalizeClassName(t.Name)] = append(groupedByNormalizedName[normalizeClassName(t.Name)], t)
	}

	// Iterate groups in sorted key order: map iteration order is randomized,
	// which reorders naming.log entries between otherwise-identical runs
	for _, key := range slices.Sorted(maps.Keys(groupedByNormalizedName)) {
		conflicts := groupedByNormalizedName[key]
		if len(conflicts) == 1 {
			continue
		}

		visited := make(map[*ast.TypeDef]bool)
		// Types are normally registered in leaf to root order, so reversing
		// allows us to visit the root types first
		// Potential issue: If leaf nodes are unregistered and re-registered
		reversed := make(ast.TypeDefs, len(conflicts))
		for i, t := range conflicts {
			reversed[len(conflicts)-1-i] = t
		}
		for _, t := range reversed {
			r.renameConflictingChildrenInner(ctx, conflicts, visited, t, []string{})
		}
	}
}

func (r *Resolver) renameConflictingChildrenInner(ctx context.Context, conflicts ast.TypeDefs, visited map[*ast.TypeDef]bool, t *ast.TypeDef, path []string) {
	if visited[t] {
		return
	}
	visited[t] = true
	labels := r.getOrCollectLabels(t)
	if len(path) > 0 {
		for _, p := range path {
			labels.selectDuplicateLabel(p, labelSource("conflict_field"))
		}
		r.renameTypeDef(ctx, t, "conflict_field")
	}
	for _, f := range t.Fields {
		if visited[f.Type] {
			continue // Ignore circular references
		}
		if slices.Contains(conflicts, f.Type) {
			fieldName := f.Name
			path := append(path, fieldName)
			r.renameConflictingChildrenInner(ctx, conflicts, visited, f.Type, path)
		}
	}
}

func (r *Resolver) incrementDuplicateNames(ctx context.Context, typeDefs ast.TypeDefs) {
	groupedByNormalizedName := make(map[string]ast.TypeDefs)
	for _, t := range typeDefs {
		groupedByNormalizedName[normalizeClassName(t.Name)] = append(groupedByNormalizedName[normalizeClassName(t.Name)], t)
	}

	for _, key := range slices.Sorted(maps.Keys(groupedByNormalizedName)) {
		typeDefs := groupedByNormalizedName[key]
		if len(typeDefs) == 1 {
			continue
		}

		componentCount := 0
		for _, t := range typeDefs {
			if t.IsComponent {
				componentCount++
			}
		}

		increment := 1
		for _, t := range typeDefs {
			labels := r.getOrCollectLabels(t)
			oldCount := labels.getLabelForSource(labelSource("duplicate_count"))
			labels.deleteLabel(oldCount)

			if componentCount == 1 && t.IsComponent {
				continue
			}

			newCount := strconv.Itoa(increment)
			increment++
			labels.addLabel(newCount, labelSource("duplicate_count"))
			labels.selectLabel(newCount)

			r.renameTypeDef(ctx, t, "incremented")
		}
	}
}

func (r *Resolver) handleTypeWithNoName(ctx context.Context, t *ast.TypeDef) {
	labels := r.getOrCollectLabels(t)

	// These are good initial names
	preferredTypes := map[ast.ContextType]bool{
		ast.ContextTypeRefType:  true,
		ast.ContextTypeRefName:  true,
		ast.ContextTypeProperty: true,
		ast.ContextTypeOneOf:    true,
	}

	selected := false

	// First pass: search in reverse order for frames with the preferred types.
	for i := len(t.ContextStack) - 1; i >= 0; i-- {
		frame := t.ContextStack[i]
		if !frame.Used {
			label := frame.DisplayName()
			if label == "" {
				continue // Skip empty labels.
			}
			// Only consider frames with one of the preferred types.
			if !preferredTypes[frame.Type] || (strings.ToLower(label) == "responsebody" || strings.ToLower(label) == "requestbody" || strings.ToLower(label) == "response" || strings.ToLower(label) == "request") {
				continue
			}
			labels.selectLabel(label)
			// If the label isn’t a number, we’re satisfied.
			if _, err := strconv.Atoi(label); err != nil {
				selected = true
				break
			}
		}
	}

	// Second pass: if no preferred label was selected, use all available frames
	if !selected {
		for i := len(t.ContextStack) - 1; i >= 0; i-- {
			frame := t.ContextStack[i]
			if !frame.Used {
				label := frame.DisplayName()
				if label == "" {
					continue
				}
				labels.selectLabel(label)
				// Break if the label is not a number.
				if _, err := strconv.Atoi(label); err != nil {
					break
				}
			}
		}
	}

	r.renameTypeDef(ctx, t, "add_name")

	if t.Name == "" {
		t.Name = "_" // TODO: This is a worry-some case which should never happen
		logging.From(ctx).Error(fmt.Sprintf("Failed to generate a name for %q \t %s", t.Name, labels.allLabelsString()))
	}
}
