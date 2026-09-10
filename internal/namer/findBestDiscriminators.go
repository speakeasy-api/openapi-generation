package namer

import (
	"sort"
)

// TieBreaker is a function type used to break ties between labels.
type TieBreaker func(labelA, labelB string) string

// FindBestDiscriminators finds the smallest subset of labels to
// discriminate between types. For each type it picks the smallest
// subset of its labels such that no other type has all of those labels.
//
// Top down algorithm:
//  1. Start with the name for every type being the concatenation of all the labels.
//  2. Try to remove one label at a time from all the types.
//  3. If removing the label leaves the type name ambiguous, put the label back.
//     A name is considered ambiguous if it is a subset of another name.
//
// Parameters:
//   - types: A list of lists of strings, where each inner list represents
//     a type represented by its labels.
//   - optionalTieBreakerFunc: An optional function that can be used to break
//     ties between labels of equal granularity. Return empty string if no preference.
//
// Returns:
//   - A list of strings, where each string represents a label that discriminates
//     between the types.
func FindBestDiscriminators(types [][]string, optionalTieBreakerFunc ...TieBreaker) [][]string {
	// Set up the tieBreaker function.
	var tieBreaker TieBreaker = func(a, b string) string { return "" }
	if len(optionalTieBreakerFunc) > 0 {
		tieBreaker = optionalTieBreakerFunc[0]
	}

	// Build a coverageMap: for each label, count in how many types it appears
	coverageMap := make(map[string]map[int]bool)
	allLabelsSet := map[string]bool{}
	allLabels := []string{}
	for i, labels := range types {
		seen := map[string]bool{}
		for _, label := range labels {
			if !seen[label] && label != "" {
				seen[label] = true
				if coverageMap[label] == nil {
					coverageMap[label] = make(map[int]bool)
				}
				coverageMap[label][i] = true
				if !allLabelsSet[label] {
					allLabelsSet[label] = true
					allLabels = append(allLabels, label)
				}
			}
		}
	}

	// Sort the labels in descending order of coverage
	sort.SliceStable(allLabels, func(i, j int) bool {
		a, b := allLabels[i], allLabels[j]
		covA, covB := coverageMap[a], coverageMap[b]
		if len(covA) != len(covB) {
			return len(covA) > len(covB)
		}
		res := tieBreaker(a, b)
		switch res {
		case a:
			return false
		case b:
			return true
		}

		return j < i
	})

	labelToIndex := mapLabelsToIndex(allLabels)

	// Start with the initial name for a type being all its labels
	n := len(types)
	initialNames := make([]stringSet, n)
	for i := 0; i < n; i++ {
		initialNames[i] = StringSetFrom(types[i], &labelToIndex, &allLabels)
	}

	// Record the initial ambiguity of all pairs of types
	initialAmbiguity := make([][]bool, n)
	for i := 0; i < n; i++ {
		if initialAmbiguity[i] == nil {
			initialAmbiguity[i] = make([]bool, n)
		}
		for j := i + 1; j < n; j++ {
			if initialAmbiguity[j] == nil {
				initialAmbiguity[j] = make([]bool, n)
			}
			initialAmbiguity[i][j] = initialNames[i].IsSubSet(&initialNames[j])
			initialAmbiguity[j][i] = initialNames[j].IsSubSet(&initialNames[i])
		}
	}

	names := make([]stringSet, n)
	for i := 0; i < n; i++ {
		names[i] = initialNames[i].Clone()
	}

	// Attempt to remove labels one by one until we find the smallest set of labels
	// Removal is accepted only if, after removal type is still uniquely discriminated
	for _, label := range allLabels {
		for i := 0; i < n; i++ {
			newNameI := names[i].Clone()
			newNameI.Clear(label)
			validRemoval := true
			for j := 0; j < n; j++ {
				if i == j {
					continue
				}

				willBeAmbiguous := newNameI.IsSubSet(&initialNames[j])
				startedAmbiguous := initialAmbiguity[i][j]
				if willBeAmbiguous && !startedAmbiguous {
					validRemoval = false
					break
				}

			}

			if validRemoval {
				names[i] = newNameI
			}
		}
	}

	result := make([][]string, n)
	for i := 0; i < n; i++ {
		result[i] = names[i].AsSlice()
	}
	return result
}

func mapLabelsToIndex(labels []string) map[string]int {
	labelToIndex := make(map[string]int)
	for i, label := range labels {
		labelToIndex[label] = i
	}
	return labelToIndex
}
