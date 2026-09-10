package ast

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

type PublicExports struct {
	RootChildren PublicExportChildren `yaml:",omitempty"`
	Groups       PublicExportGroups   `yaml:",omitempty"`
}

// HasExplicitExports reports whether any export came from an explicit
// x-speakeasy-exports declaration (as opposed to implicit model-namespace
// auto-exports).
func (p *PublicExports) HasExplicitExports() bool {
	if p == nil {
		return false
	}
	for _, group := range p.Groups {
		for _, export := range group.Exports {
			if !export.Implicit {
				return true
			}
		}
	}
	return false
}

type PublicExportGroups []*PublicExportGroup

type PublicExportGroup struct {
	Group    string               `yaml:",omitempty"`
	Parts    []string             `yaml:",omitempty"`
	Exports  PublicExportTargets  `yaml:",omitempty"`
	Children PublicExportChildren `yaml:",omitempty"`
}

type PublicExportTargets []*PublicExportTarget

type PublicExportTarget struct {
	Name   string   `yaml:",omitempty"`
	Target *TypeDef `yaml:",omitempty"`
	// Input marks exports that alias the target's input-side representation
	// (e.g. the Python TypedDict companion) instead of the model class.
	Input bool `yaml:",omitempty"`
	// Implicit marks exports registered from x-speakeasy-model-namespace
	// tagging rather than an explicit x-speakeasy-exports declaration.
	// Languages only render implicit exports when the SDK opts in via the
	// imports.paths.resources configuration.
	Implicit bool `yaml:",omitempty"`
}

type PublicExportChildren []*PublicExportChild

type PublicExportChild struct {
	Name  string   `yaml:",omitempty"`
	Group string   `yaml:",omitempty"`
	Parts []string `yaml:",omitempty"`
}

type publicExportGroupBuilder struct {
	parts       []string
	targets     map[string]*TypeDef
	inputs      map[string]bool
	implicit    map[string]bool
	conflicting map[string]bool
	children    map[string]*PublicExportChild
}

func newPublicExportGroupBuilder(parts []string) *publicExportGroupBuilder {
	return &publicExportGroupBuilder{
		parts:       slices.Clone(parts),
		targets:     map[string]*TypeDef{},
		inputs:      map[string]bool{},
		implicit:    map[string]bool{},
		conflicting: map[string]bool{},
		children:    map[string]*PublicExportChild{},
	}
}

// BuildPublicExports resolves x-speakeasy-exports into a language-agnostic
// public export tree after type names and model buckets have been finalized.
// Every rendered type tagged with x-speakeasy-model-namespace is implicitly
// exported to the export group matching its namespace, so new models appear
// on the public surface without per-type export annotations. Explicit
// x-speakeasy-exports entries always take precedence over implicit ones.
func BuildPublicExports(ctx context.Context, a *AST) *PublicExports {
	if a == nil || a.BucketedTypes == nil {
		return nil
	}

	rendered := newPublicExportRenderedTypes(a.BucketedTypes)
	groups := map[string]*publicExportGroupBuilder{}

	_ = a.Walk(func(n Node, _ []Node, _ *AST) error {
		return n.Match(Matchers{
			TypeDef: func(t *TypeDef) error {
				if t == nil || t.Extensions == nil || len(t.Extensions.PublicExports) == 0 {
					return nil
				}

				target := publicExportTargetForType(t, rendered)
				if target == nil {
					return nil
				}

				for _, export := range t.Extensions.PublicExports {
					publicExportRegister(ctx, groups, target, export)
				}

				return nil
			},
			Operation: func(o *Operation) error {
				if o == nil || o.Extensions == nil || len(o.Extensions.PublicExports) == 0 {
					return nil
				}
				if o.Request == nil || o.Request.Field == nil || o.Request.Field.Type == nil {
					return nil
				}

				target := publicExportRenderedTarget(o.Request.Field.Type, rendered)
				if target == nil {
					return nil
				}

				for _, export := range o.Extensions.PublicExports {
					publicExportRegister(ctx, groups, target, export)
				}

				return nil
			},
		})
	})

	ambientRoots := publicExportAmbientRootsFromGroups(groups)
	ambientExports := publicExportInferAmbientExports(ambientRoots)
	ambientGroupKeys := publicExportAmbientRootGroupKeysWithExports(ambientRoots, ambientExports)
	publicExportRegisterAmbientExports(ctx, groups, ambientExports)

	for _, models := range a.BucketedTypes.All() {
		for _, types := range models.All() {
			for _, t := range types {
				if publicExportImplicitNamespaceSuppressed(ambientGroupKeys, t) {
					continue
				}
				publicExportRegisterImplicitNamespaceExport(groups, t)
			}
		}
	}

	for key, group := range groups {
		if len(group.targets) == 0 {
			delete(groups, key)
		}
	}
	if len(groups) == 0 {
		return nil
	}

	for _, group := range groups {
		for i := 1; i < len(group.parts); i++ {
			parentParts := group.parts[:i]
			parentKey := publicExportGroupKey(parentParts)
			parent := groups[parentKey]
			if parent == nil {
				parent = newPublicExportGroupBuilder(parentParts)
				groups[parentKey] = parent
			}
		}
	}

	for _, group := range groups {
		if len(group.parts) <= 1 {
			continue
		}

		parentParts := group.parts[:len(group.parts)-1]
		parent := groups[publicExportGroupKey(parentParts)]
		if parent == nil {
			continue
		}
		child := publicExportChild(group.parts)
		parent.children[child.Group] = child
	}

	rootChildrenByGroup := map[string]*PublicExportChild{}
	publicGroups := make(PublicExportGroups, 0, len(groups))
	for _, group := range groups {
		publicGroup := &PublicExportGroup{
			Group: strings.Join(group.parts, "."),
			Parts: slices.Clone(group.parts),
			Exports: publicExportTargets(
				group.targets,
				group.inputs,
				group.implicit,
			),
			Children: publicExportChildren(group.children),
		}
		publicGroups = append(publicGroups, publicGroup)
		if len(group.parts) == 1 {
			child := publicExportChild(group.parts)
			rootChildrenByGroup[child.Group] = child
		}
	}

	slices.SortFunc(publicGroups, func(a, b *PublicExportGroup) int {
		return cmp.Compare(a.Group, b.Group)
	})

	rootChildren := publicExportChildren(rootChildrenByGroup)
	if len(rootChildren) == 0 || len(publicGroups) == 0 {
		return nil
	}

	return &PublicExports{
		RootChildren: rootChildren,
		Groups:       publicGroups,
	}
}

func publicExportRegister(
	ctx context.Context,
	groups map[string]*publicExportGroupBuilder,
	target *TypeDef,
	export extensions.PublicExport,
) {
	name := strings.TrimSpace(export.Name)
	parts := publicExportGroupParts(export.Group)
	if name == "" || len(parts) == 0 {
		return
	}

	key := publicExportGroupKey(parts)
	group := groups[key]
	if group == nil {
		group = newPublicExportGroupBuilder(parts)
		groups[key] = group
	}

	input := export.Input()
	existing := group.targets[name]
	if existing != nil && publicExportRegistrationID(existing) != publicExportRegistrationID(target) {
		// Input exports alias the input companion, so when a schema splits into
		// input/output flavors the request-side flavor carries the companion
		// being requested.
		if input {
			if preferred := publicExportPreferredRequestTarget(existing, target); preferred != nil {
				publicExportSetPreferredTarget(group, name, existing, target, preferred, input)
				return
			}
		}
		if preferred := publicExportPreferredNameMatchedTarget(name, existing, target); preferred != nil {
			publicExportSetPreferredTarget(group, name, existing, target, preferred, input)
			return
		}
		preferred := publicExportPreferredTarget(existing, target)
		if preferred != nil {
			publicExportSetPreferredTarget(group, name, existing, target, preferred, input)
			return
		}
		delete(group.targets, name)
		group.conflicting[name] = true
		logging.From(ctx).Warn(
			"x-speakeasy-exports: conflicting types registered for the same alias; the export is dropped",
			zap.String("group", strings.Join(group.parts, ".")),
			zap.String("name", name),
		)
		return
	}
	if !group.conflicting[name] {
		group.targets[name] = target
		group.inputs[name] = group.inputs[name] || input
	}
}

func publicExportRegisterAmbientExports(
	ctx context.Context,
	groups map[string]*publicExportGroupBuilder,
	ambientExports []publicExportAmbientExport,
) {
	for _, ambientExport := range ambientExports {
		publicExportRegister(ctx, groups, ambientExport.Target, extensions.PublicExport{
			Group: ambientExport.Group,
			Name:  ambientExport.Name,
		})
	}
}

func publicExportAmbientRootGroupKeysWithExports(
	roots []publicExportAmbientRoot,
	exports []publicExportAmbientExport,
) map[string]bool {
	keys := map[string]bool{}
	for _, export := range exports {
		for _, root := range roots {
			if export.Group != root.Group && !strings.HasPrefix(export.Group, root.Group+".") {
				continue
			}
			keys[publicExportGroupKey(publicExportGroupParts(root.Group))] = true
		}
	}
	return keys
}

func publicExportAmbientRootsFromGroups(groups map[string]*publicExportGroupBuilder) []publicExportAmbientRoot {
	var roots []publicExportAmbientRoot
	for _, group := range groups {
		if group == nil {
			continue
		}
		groupName := strings.Join(group.parts, ".")
		for name, target := range group.targets {
			if name == "" || target == nil || group.conflicting[name] {
				continue
			}
			if !publicExportAmbientCanInferFromRootName(name) {
				continue
			}
			roots = append(roots, publicExportAmbientRoot{
				Group:  groupName,
				Name:   name,
				Target: target,
				Input:  group.inputs[name],
			})
		}
	}
	slices.SortFunc(roots, func(a, b publicExportAmbientRoot) int {
		return cmp.Compare(a.Group+"."+a.Name, b.Group+"."+b.Name)
	})
	return roots
}

type publicExportRenderedTypes struct {
	byPointer        map[*TypeDef]*TypeDef
	byRegistrationID map[string]*TypeDef
}

func newPublicExportRenderedTypes(bucketedTypes BucketedTypes) publicExportRenderedTypes {
	rendered := publicExportRenderedTypes{
		byPointer:        map[*TypeDef]*TypeDef{},
		byRegistrationID: map[string]*TypeDef{},
	}

	for _, models := range bucketedTypes.All() {
		for _, types := range models.All() {
			for _, t := range types {
				if !publicExportCanExpose(t) {
					continue
				}

				rendered.byPointer[t] = t
				if id := publicExportRegistrationID(t); id != "" {
					rendered.byRegistrationID[id] = t
				}
			}
		}
	}

	return rendered
}

func publicExportTargetForType(t *TypeDef, rendered publicExportRenderedTypes) *TypeDef {
	if target := publicExportRenderedTarget(t, rendered); target != nil {
		return target
	}

	candidates := map[string]*TypeDef{}
	publicExportCollectDescendantTargets(t, rendered, candidates, map[*TypeDef]bool{})
	if len(candidates) != 1 {
		return nil
	}
	for _, candidate := range candidates {
		return candidate
	}
	return nil
}

func publicExportCollectDescendantTargets(
	t *TypeDef,
	rendered publicExportRenderedTypes,
	candidates map[string]*TypeDef,
	visited map[*TypeDef]bool,
) {
	if t == nil || visited[t] {
		return
	}
	visited[t] = true

	if target := publicExportRenderedTarget(t, rendered); target != nil {
		candidates[publicExportRegistrationID(target)] = target
		return
	}

	if t.ItemType != nil {
		publicExportCollectDescendantTargets(t.ItemType, rendered, candidates, visited)
	}
	for _, field := range t.Fields {
		if field == nil {
			continue
		}
		publicExportCollectDescendantTargets(field.Type, rendered, candidates, visited)
	}
	for _, associatedType := range t.AssociatedTypes {
		publicExportCollectDescendantTargets(associatedType, rendered, candidates, visited)
	}
}

func publicExportRenderedTarget(t *TypeDef, rendered publicExportRenderedTypes) *TypeDef {
	if !publicExportCanExpose(t) {
		return nil
	}
	if target := rendered.byPointer[t]; target != nil {
		return target
	}
	if id := publicExportRegistrationID(t); id != "" {
		return rendered.byRegistrationID[id]
	}
	return nil
}

func publicExportCanExpose(t *TypeDef) bool {
	if t == nil {
		return false
	}

	switch t.Type {
	case DataTypeClass, DataTypeUnion, DataTypeEnum, DataTypeError:
		return true
	default:
		return false
	}
}

func publicExportRegistrationID(t *TypeDef) string {
	if t == nil || !t.IsCustomType() || t.ContextStack == nil {
		return ""
	}
	return t.GetRegistrationID()
}

func publicExportNormalizedName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "_", "")
}

func publicExportPreferredNameMatchedTarget(name string, a, b *TypeDef) *TypeDef {
	expectedName := publicExportNormalizedName(name)
	aMatches := publicExportNormalizedName(a.Name) == expectedName
	bMatches := publicExportNormalizedName(b.Name) == expectedName
	switch {
	case aMatches && !bMatches:
		return a
	case bMatches && !aMatches:
		return b
	default:
		return nil
	}
}

func publicExportPreferredRequestTarget(a, b *TypeDef) *TypeDef {
	switch {
	case a.UsedInRequest && !b.UsedInRequest:
		return a
	case b.UsedInRequest && !a.UsedInRequest:
		return b
	default:
		return nil
	}
}

func publicExportSetPreferredTarget(
	group *publicExportGroupBuilder,
	name string,
	existing *TypeDef,
	target *TypeDef,
	preferred *TypeDef,
	targetInput bool,
) {
	group.targets[name] = preferred
	if publicExportRegistrationID(preferred) == publicExportRegistrationID(existing) {
		group.inputs[name] = group.inputs[name] || targetInput
		return
	}
	if publicExportRegistrationID(preferred) == publicExportRegistrationID(target) {
		group.inputs[name] = targetInput
	}
}

func publicExportPreferredTarget(a, b *TypeDef) *TypeDef {
	// Cloned inline schemas can carry the source extension into more specific
	// context stacks. Prefer the less-specific target for that propagation case.
	aScore := publicExportTargetSpecificity(a)
	bScore := publicExportTargetSpecificity(b)
	switch {
	case aScore < bScore:
		return a
	case bScore < aScore:
		return b
	default:
		return nil
	}
}

func publicExportTargetSpecificity(t *TypeDef) int {
	if t == nil || t.ContextStack == nil {
		return 1 << 30
	}
	return len(t.ContextStack)
}

func publicExportGroupParts(group string) []string {
	rawParts := strings.Split(group, ".")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}
	return parts
}

func publicExportGroupKey(parts []string) string {
	return strings.Join(parts, "\x00")
}

func publicExportTargets(
	targets map[string]*TypeDef,
	inputs map[string]bool,
	implicit map[string]bool,
) PublicExportTargets {
	if len(targets) == 0 {
		return nil
	}

	result := make(PublicExportTargets, 0, len(targets))
	for name, target := range targets {
		result = append(result, &PublicExportTarget{
			Name:     name,
			Target:   target,
			Input:    inputs[name],
			Implicit: implicit[name],
		})
	}
	slices.SortFunc(result, func(a, b *PublicExportTarget) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return result
}

func publicExportChildren(children map[string]*PublicExportChild) PublicExportChildren {
	if len(children) == 0 {
		return nil
	}

	result := make(PublicExportChildren, 0, len(children))
	for _, child := range children {
		result = append(result, child)
	}
	slices.SortFunc(result, func(a, b *PublicExportChild) int {
		return cmp.Compare(a.Group, b.Group)
	})
	return result
}

func publicExportChild(parts []string) *PublicExportChild {
	return &PublicExportChild{
		Name:  parts[len(parts)-1],
		Group: strings.Join(parts, "."),
		Parts: slices.Clone(parts),
	}
}

// publicExportRegisterImplicitNamespaceExport registers a rendered type under
// its model namespace group using its own type name. Implicit exports only
// fill vacant names: explicit x-speakeasy-exports declarations and their
// conflict resolution always take precedence.
func publicExportRegisterImplicitNamespaceExport(
	groups map[string]*publicExportGroupBuilder,
	t *TypeDef,
) {
	if !publicExportCanExpose(t) || t.Name == "" || t.Scope == ScopeUtils {
		return
	}
	if t.Extensions == nil || t.Extensions.ModelNamespace == nil {
		return
	}

	namespace := strings.TrimSpace(*t.Extensions.ModelNamespace)
	parts := publicExportGroupParts(namespace)
	if len(parts) == 0 {
		return
	}

	key := publicExportGroupKey(parts)
	group := groups[key]
	if group == nil {
		group = newPublicExportGroupBuilder(parts)
		groups[key] = group
	}

	// Vacancy is checked on normalized names: explicit aliases like
	// "environment" and the implicit type name "Environment" sanitize to the
	// same identifier in generated code. Two implicit candidates with the
	// same name resolve through the regular target preference so duplicate
	// rendered names stay order-independent.
	normalized := publicExportNormalizedName(t.Name)
	for name := range group.targets {
		if publicExportNormalizedName(name) != normalized {
			continue
		}
		if name == t.Name && group.implicit[name] {
			existing := group.targets[name]
			if publicExportRegistrationID(existing) != publicExportRegistrationID(t) {
				if preferred := publicExportPreferredTarget(existing, t); preferred != nil {
					group.targets[name] = preferred
				}
			}
		}
		return
	}
	for name := range group.conflicting {
		if publicExportNormalizedName(name) == normalized {
			return
		}
	}
	group.targets[t.Name] = t
	group.implicit[t.Name] = true
}

func publicExportImplicitNamespaceSuppressed(
	suppressedImplicitGroups map[string]bool,
	t *TypeDef,
) bool {
	if t == nil || t.Extensions == nil || t.Extensions.ModelNamespace == nil {
		return false
	}
	parts := publicExportGroupParts(strings.TrimSpace(*t.Extensions.ModelNamespace))
	if len(parts) == 0 {
		return false
	}
	return suppressedImplicitGroups[publicExportGroupKey(parts)]
}
