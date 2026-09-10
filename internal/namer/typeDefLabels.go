package namer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
)

// labelSource keeps track of where each label came from
// eg. "scope", "data_type", "response_status_code", "response_description", "ref_type", "ref_name", "one_of", "input_output"
type labelSource string

type LabelWithSource struct {
	Label string
	// Sanitized, lowercase, no spaces or underscores - maximized conflict
	LabelNormalized string
	Source          labelSource
}

// TypeDefLabels keeps track of all the labels for a TypeDef
type TypeDefLabels struct {
	// Where each label came from, not all labels are used in the name
	All []LabelWithSource

	// Labels that were selected as discriminators which are concatenated with the original name
	Selected []*LabelWithSource
}

func (n *TypeDefLabels) addLabel(label string, source labelSource) {
	if label == "" {
		return
	}
	l2 := LabelWithSource{Label: label, Source: source, LabelNormalized: normalizeClassName(label)}
	for _, l := range n.All {
		if l.LabelNormalized == l2.LabelNormalized {
			return
		}
	}
	n.All = append(n.All, l2)
}

func (n *TypeDefLabels) deleteLabel(label string) {
	for i, l := range n.All {
		if l.Label == label {
			n.All = append(n.All[:i], n.All[i+1:]...)
			return
		}
	}
}

func (n *TypeDefLabels) isSelected(label string) bool {
	for _, l := range n.Selected {
		if l.Label == label {
			return true
		}
	}
	return false
}

func (n *TypeDefLabels) selectLabel(label string) {
	if label == "" {
		return
	}

	// First try to find the exact label
	foundIdx := slices.IndexFunc(n.All, func(l LabelWithSource) bool {
		return l.Label == label
	})

	if foundIdx == -1 {
		// If not found, try to find the normalized label
		foundIdx = slices.IndexFunc(n.All, func(l LabelWithSource) bool {
			return l.LabelNormalized == normalizeClassName(label)
		})

		if foundIdx == -1 {
			if env.IsDebug() {
				fmt.Printf("ERROR: you must track the source of %q before selecting it, existing labels: %s\n", label, n.allLabelsString())
			}
			return
		}
	}

	if !n.isSelected(label) && label != "" {
		n.Selected = append(n.Selected, &n.All[foundIdx])
	}
}

// You know the label is already selected but you want to select it again
func (n *TypeDefLabels) selectDuplicateLabel(label string, source labelSource) {
	labelWithSource := LabelWithSource{Label: label, Source: source, LabelNormalized: normalizeClassName(label)}
	n.All = append(n.All, labelWithSource)
	n.Selected = append(n.Selected, &labelWithSource)
}

// Merge the selected labels with the existing selected labels
func (n *TypeDefLabels) selectLabels(newLabels []string) {
	for _, label := range newLabels {
		n.selectLabel(label)
	}
}

func (n *TypeDefLabels) availableLabels() []string {
	labels := []string{}
	for _, l := range n.All {
		if n.isSelected(l.Label) {
			continue
		}
		labels = append(labels, l.Label)
	}
	return labels
}

func (n *TypeDefLabels) allLabels() []string {
	labels := make([]string, 0, len(n.All))
	for _, l := range n.All {
		labels = append(labels, l.Label)
	}
	return labels
}

func (n *TypeDefLabels) allLabelsString() string {
	labels := make([]string, 0, len(n.All))
	for _, l := range n.All {
		labels = append(labels, fmt.Sprintf("%s:%s", l.Source, l.Label))
	}
	return strings.Join(labels, " ")
}

func (n *TypeDefLabels) getLabelForSource(source labelSource) string {
	labels := []string{}
	for _, l := range n.All {
		if l.Source == source {
			labels = append(labels, l.Label)
		}
	}
	if len(labels) == 0 {
		return ""
	}
	if len(labels) > 1 && env.IsDebug() {
		panic(fmt.Sprintf("multiple labels found for source %q: %v", source, labels))
	}
	return labels[0]
}

// getOrCollectLabels collects the labels for a TypeDef and keep track of where each label came from
// If the label appears in multiple sources, the sourcePriority is used to determine which source is more important
func (r *Resolver) getOrCollectLabels(t *ast.TypeDef) *TypeDefLabels {
	if _, ok := r.labelsCache[t]; ok {
		return r.labelsCache[t]
	}
	r.labelsCache[t] = r.collectLabels(t)
	return r.labelsCache[t]
}

func (r *Resolver) collectLabels(t *ast.TypeDef) *TypeDefLabels {
	labels := &TypeDefLabels{
		All:      []LabelWithSource{},
		Selected: []*LabelWithSource{},
	}

	// Skipped discriminators
	skippedLabels := map[labelSource]bool{}

	if t.IsComponent {
		refName := t.ContextStack.FindLastFrameOfType(ast.ContextTypeRefName)
		if refName == nil {
			panic(fmt.Sprintf("component type %s is missing its ContextTypeRefName context frame", t.Name))
		}
		if t.Name == refName.DisplayName() {
			skippedLabels = map[labelSource]bool{
				labelSource(ast.ContextTypeConstProperty): true,
			}
		}

	}

	// Always select the original name
	labels.addLabel(t.Name, "original_name")
	labels.selectLabel(t.Name)

	for _, source := range discriminatorPriority {
		switch source {
		case "scope":
			if t.Scope == ast.ScopeErrors || t.Scope == ast.ScopeShared {
				labels.addLabel(string(t.Scope), source)
			}
		case "data_type":
			labels.addLabel(r.dataTypeLabel(t), source)
		default:
			// In reverse order of frames
			for i := len(t.ContextStack) - 1; i >= 0; i-- {
				frame := t.ContextStack[i]
				if ast.ContextType(source) == frame.Type && !frame.Used && !skippedLabels[labelSource(frame.Type)] {
					labels.addLabel(frame.DisplayName(), source)
				}
			}
		}
	}

	// Handle remaining ContextTypes in reverse order of frames
	for i := len(t.ContextStack) - 1; i >= 0; i-- {
		frame := t.ContextStack[i]
		source := labelSource(frame.Type)
		if !frame.Used && !skippedLabels[source] {
			labels.addLabel(frame.DisplayName(), source)
		}
	}
	return labels
}

func (r *Resolver) dataTypeLabel(t *ast.TypeDef) string {
	if isSubSDK(t) {
		return "SDK"
	}

	if t.Type == ast.DataTypeClass {
		return ""
	}

	if t.Type == ast.DataTypeError {
		// TODO: Move language specific logic into language config
		if r.subsystem.Target.Target == "csharp" || r.subsystem.Target.Target == "java" || r.subsystem.Target.Target == "php" || r.subsystem.Target.Target == "unity" {
			return "exception"
		}
		return "error"
	}
	return string(t.Type)
}

func isSubSDK(t *ast.TypeDef) bool {
	// TODO: if we ever introduce classes which are not SDKs in the SDKs scope we could run into those things being suffixed "_SDK"
	return t.Scope == ast.ScopeSDK && t.Type == ast.DataTypeClass
}
