package ast

import (
	"cmp"
	"slices"
	"strings"
	"sync"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
)

var publicExportAmbientPluralizeClient = pluralize.NewClient()

var publicExportAmbientSingularize = publicExportAmbientCachedSingularize()

type publicExportAmbientRoot struct {
	Group  string
	Name   string
	Target *TypeDef
	Input  bool
}

type publicExportAmbientExport struct {
	Group  string
	Name   string
	Target *TypeDef
	Via    string
}

type publicExportAmbientEdgeKind int

const (
	publicExportAmbientEdgeField publicExportAmbientEdgeKind = iota
	publicExportAmbientEdgeOneOf
)

type publicExportAmbientEdge struct {
	root                  publicExportAmbientRoot
	parent                *TypeDef
	target                *TypeDef
	kind                  publicExportAmbientEdgeKind
	alias                 string
	fieldName             string
	groupParts            []string
	sourceIsParent        bool
	parentIsOperationRoot bool
	parentIsComponentRoot bool
	targetIsExplicitRoot  bool
	targetAliasIsRoot     bool
	parentConstAlias      string
}

type PublicExportAmbientRoot = publicExportAmbientRoot

type PublicExportAmbientExport = publicExportAmbientExport

func InferAmbientPublicExportsForRoots(roots []PublicExportAmbientRoot) []PublicExportAmbientExport {
	return publicExportInferAmbientExports(roots)
}

// publicExportInferAmbientExports infers nested public TypeScript namespaces:
// resource-model roots stay flat, while non-root helper models referenced by
// those roots become nested exports.
func publicExportInferAmbientExports(roots []publicExportAmbientRoot) []publicExportAmbientExport {
	rootTypes := map[string]bool{}
	rootAliases := map[string]bool{}
	for _, root := range roots {
		for _, key := range publicExportAmbientRootKeys(root.Target) {
			rootTypes[key] = true
		}
		if alias := publicExportNormalizedName(root.Name); alias != "" {
			rootAliases[alias] = true
		}
	}

	exportsByPath := map[string]publicExportAmbientExport{}
	for _, root := range roots {
		group := strings.TrimSpace(root.Group)
		name := strings.TrimSpace(root.Name)
		if group == "" || name == "" || root.Target == nil {
			continue
		}

		publicExportInferAmbientChildren(
			root,
			root.Target,
			root.Target.Name,
			append(publicExportGroupParts(group), name),
			rootTypes,
			rootAliases,
			map[*TypeDef]bool{},
			exportsByPath,
		)
	}

	result := make([]publicExportAmbientExport, 0, len(exportsByPath))
	for _, export := range exportsByPath {
		result = append(result, export)
	}
	slices.SortFunc(result, func(a, b publicExportAmbientExport) int {
		return cmp.Compare(a.Group+"."+a.Name, b.Group+"."+b.Name)
	})
	return result
}

func publicExportInferAmbientChildren(
	root publicExportAmbientRoot,
	parent *TypeDef,
	parentName string,
	groupParts []string,
	rootTypes map[string]bool,
	rootAliases map[string]bool,
	visited map[*TypeDef]bool,
	exportsByPath map[string]publicExportAmbientExport,
) {
	if parent == nil || visited[parent] {
		return
	}

	nextVisited := map[*TypeDef]bool{}
	for typ, value := range visited {
		nextVisited[typ] = value
	}
	nextVisited[parent] = true

	for _, field := range parent.Fields {
		if field == nil {
			continue
		}
		target := publicExportAmbientUnwrapContainer(field.Type)
		if target == nil || !publicExportAmbientCanExposeNested(target) {
			continue
		}
		if publicExportRegistrationID(target) == "" && !publicExportAmbientIsTransparentUnion(target, rootTypes) {
			continue
		}

		alias := publicExportAmbientFieldAlias(field, target)
		if alias == "" {
			continue
		}

		if publicExportAmbientShouldExpandFieldUnion(alias, target, rootTypes) {
			for _, associatedType := range target.AssociatedTypes {
				publicExportInferAmbientAssociated(
					root,
					parentName,
					groupParts,
					parent,
					target,
					associatedType,
					rootTypes,
					rootAliases,
					nextVisited,
					exportsByPath,
				)
			}
			continue
		}
		if !publicExportAmbientCanExposeFieldChild(root, parent, target, alias, field.Name, groupParts, rootTypes) {
			continue
		}

		edge := publicExportAmbientNewEdge(
			root,
			parent,
			parent,
			target,
			publicExportAmbientEdgeField,
			alias,
			field.Name,
			groupParts,
			rootTypes,
			rootAliases,
		)
		publicExportInferAmbientAddExport(
			edge,
			"field "+field.Name,
			rootTypes,
			rootAliases,
			true,
			nextVisited,
			exportsByPath,
		)
	}

	for _, associatedType := range parent.AssociatedTypes {
		publicExportInferAmbientAssociated(
			root,
			parentName,
			groupParts,
			parent,
			parent,
			associatedType,
			rootTypes,
			rootAliases,
			nextVisited,
			exportsByPath,
		)
	}
}

func publicExportInferAmbientAssociated(
	root publicExportAmbientRoot,
	parentName string,
	groupParts []string,
	parent *TypeDef,
	source *TypeDef,
	associatedType *TypeDef,
	rootTypes map[string]bool,
	rootAliases map[string]bool,
	visited map[*TypeDef]bool,
	exportsByPath map[string]publicExportAmbientExport,
) {
	target := publicExportAmbientUnwrapContainer(associatedType)
	if target == nil || !publicExportAmbientCanExposeNested(target) {
		return
	}
	if publicExportAmbientIsTransparentUnion(target, rootTypes) {
		for _, child := range target.AssociatedTypes {
			publicExportInferAmbientAssociated(
				root,
				parentName,
				groupParts,
				parent,
				target,
				child,
				rootTypes,
				rootAliases,
				visited,
				exportsByPath,
			)
		}
		return
	}

	alias := publicExportAmbientAssociatedAlias(parentName, target)
	if alias == "" {
		return
	}
	recurse := publicExportAmbientAliasMatchesTarget(alias, target)

	sourceName := ""
	if source != nil {
		sourceName = source.Name
	}
	edge := publicExportAmbientNewEdge(
		root,
		parent,
		source,
		target,
		publicExportAmbientEdgeOneOf,
		alias,
		"",
		groupParts,
		rootTypes,
		rootAliases,
	)
	publicExportInferAmbientAddExport(
		edge,
		"oneOf "+sourceName,
		rootTypes,
		rootAliases,
		recurse,
		visited,
		exportsByPath,
	)
}

func publicExportInferAmbientAddExport(
	edge publicExportAmbientEdge,
	via string,
	rootTypes map[string]bool,
	rootAliases map[string]bool,
	recurse bool,
	visited map[*TypeDef]bool,
	exportsByPath map[string]publicExportAmbientExport,
) {
	groupParts := edgeGroupParts(edge)
	alias := edge.alias
	target := edge.target
	if target == nil {
		return
	}
	if !publicExportAmbientCanAddExport(edge, groupParts, rootTypes) {
		return
	}

	group := strings.Join(groupParts, ".")
	path := group + "." + alias
	exportsByPath[path] = publicExportAmbientExport{
		Group:  group,
		Name:   alias,
		Target: target,
		Via:    via,
	}

	if recurse {
		publicExportInferAmbientChildren(
			edge.root,
			target,
			target.Name,
			append(slices.Clone(groupParts), alias),
			rootTypes,
			rootAliases,
			visited,
			exportsByPath,
		)
	}
}

func publicExportAmbientNewEdge(
	root publicExportAmbientRoot,
	parent *TypeDef,
	source *TypeDef,
	target *TypeDef,
	kind publicExportAmbientEdgeKind,
	alias string,
	fieldName string,
	groupParts []string,
	rootTypes map[string]bool,
	rootAliases map[string]bool,
) publicExportAmbientEdge {
	parentIsRoot := publicExportAmbientIsCurrentRoot(root, parent, groupParts)
	parentConstAlias := ""
	if aliases := publicExportAmbientParentConstAliases(parent); len(aliases) > 0 {
		parentConstAlias = aliases[0]
	}

	return publicExportAmbientEdge{
		root:                  root,
		parent:                parent,
		target:                target,
		kind:                  kind,
		alias:                 alias,
		fieldName:             fieldName,
		groupParts:            slices.Clone(groupParts),
		sourceIsParent:        source != nil && parent != nil && source == parent,
		parentIsOperationRoot: parentIsRoot && publicExportAmbientRootIsOperation(root),
		parentIsComponentRoot: parentIsRoot && publicExportAmbientRootIsComponent(root),
		targetIsExplicitRoot:  publicExportAmbientIsRoot(target, rootTypes),
		targetAliasIsRoot:     rootAliases[publicExportNormalizedName(alias)],
		parentConstAlias:      parentConstAlias,
	}
}

func edgeGroupParts(edge publicExportAmbientEdge) []string {
	return edge.groupParts
}

func publicExportAmbientCanAddExport(
	edge publicExportAmbientEdge,
	groupParts []string,
	rootTypes map[string]bool,
) bool {
	if edge.target == nil {
		return false
	}
	if !publicExportAmbientMatchesGroup(groupParts, edge.target) {
		return false
	}
	if publicExportAmbientRootIsOperation(edge.root) {
		return edge.kind == publicExportAmbientEdgeOneOf &&
			edge.parentIsOperationRoot &&
			edge.sourceIsParent &&
			edge.targetIsExplicitRoot &&
			edge.targetAliasIsRoot
	}
	if edge.kind == publicExportAmbientEdgeField &&
		!edge.targetIsExplicitRoot &&
		publicExportAmbientIsRootAlias(edge.alias, rootTypes) {
		return false
	}
	if edge.targetIsExplicitRoot {
		switch edge.kind {
		case publicExportAmbientEdgeOneOf:
			if edge.parentIsComponentRoot {
				return false
			}
			return edge.parentIsOperationRoot
		case publicExportAmbientEdgeField:
			return publicExportAmbientFieldRootTargetOwned(edge)
		default:
			return false
		}
	}
	return true
}

func publicExportAmbientUnwrapContainer(t *TypeDef) *TypeDef {
	for t != nil {
		switch t.Type {
		case DataTypeArray, DataTypeSet, DataTypeMap:
			if t.ItemType == nil {
				return t
			}
			t = t.ItemType
		default:
			return t
		}
	}
	return nil
}

func publicExportAmbientIsTransparentUnion(t *TypeDef, rootTypes map[string]bool) bool {
	return t != nil &&
		t.Type == DataTypeUnion &&
		!publicExportAmbientIsRoot(t, rootTypes) &&
		len(t.Fields) == 0 &&
		len(t.AssociatedTypes) > 0
}

func publicExportAmbientShouldExpandFieldUnion(alias string, t *TypeDef, rootTypes map[string]bool) bool {
	if t == nil || t.Type != DataTypeUnion || len(t.Fields) != 0 || len(t.AssociatedTypes) == 0 {
		return false
	}
	return !publicExportAmbientIsRoot(t, rootTypes) || !publicExportAmbientAliasMatchesTarget(alias, t)
}

func publicExportAmbientIsRoot(t *TypeDef, rootTypes map[string]bool) bool {
	for _, key := range publicExportAmbientRootKeys(t) {
		if rootTypes[key] {
			return true
		}
	}
	return false
}

func publicExportAmbientIsRootAlias(alias string, rootTypes map[string]bool) bool {
	return rootTypes["name:"+publicExportNormalizedName(alias)]
}

func publicExportAmbientRootKeys(t *TypeDef) []string {
	if t == nil {
		return nil
	}
	var keys []string
	if id := publicExportRegistrationID(t); id != "" {
		keys = append(keys, "id:"+id)
	}
	if t.Name != "" {
		keys = append(keys, "name:"+publicExportNormalizedName(t.Name))
	}
	if t.OriginalName != "" {
		keys = append(keys, "name:"+publicExportNormalizedName(t.OriginalName))
	}
	return keys
}

func publicExportAmbientCanExposeNested(t *TypeDef) bool {
	if !publicExportCanExpose(t) {
		return false
	}
	if t.Comments != nil && t.Comments.Deprecated {
		return false
	}
	if publicExportNormalizedName(t.Name) == "anonymous" || publicExportNormalizedName(t.OriginalName) == "anonymous" {
		return false
	}
	if t.Type == DataTypeClass && len(t.Fields) == 0 {
		return false
	}
	return t.Type != DataTypeEnum
}

func publicExportAmbientMatchesGroup(groupParts []string, target *TypeDef) bool {
	if len(groupParts) == 0 || target == nil || target.Extensions == nil || target.Extensions.ModelNamespace == nil {
		return true
	}
	namespaceParts := publicExportGroupParts(*target.Extensions.ModelNamespace)
	return len(namespaceParts) <= len(groupParts) &&
		publicExportGroupKey(namespaceParts) == publicExportGroupKey(groupParts[:len(namespaceParts)])
}

func publicExportAmbientIsCurrentRoot(root publicExportAmbientRoot, parent *TypeDef, groupParts []string) bool {
	if parent == nil || root.Target == nil || len(groupParts) != len(publicExportGroupParts(root.Group))+1 {
		return false
	}
	if parent == root.Target {
		return true
	}

	parentID := publicExportRegistrationID(parent)
	rootID := publicExportRegistrationID(root.Target)
	return parentID != "" && parentID == rootID
}

func publicExportAmbientRootIsOperation(root publicExportAmbientRoot) bool {
	return root.Input ||
		(root.Target != nil && root.Target.ContextStack.FindLastFrameOfType(ContextTypeOperation) != nil)
}

func publicExportAmbientRootIsComponent(root publicExportAmbientRoot) bool {
	return !publicExportAmbientRootIsOperation(root)
}

func publicExportAmbientFieldRootTargetOwned(edge publicExportAmbientEdge) bool {
	if edge.parent == nil || edge.target == nil || edge.alias == "" {
		return false
	}

	aliases := publicExportAmbientParentConstAliases(edge.parent)
	if edge.parentConstAlias != "" {
		aliases = append([]string{edge.parentConstAlias}, aliases...)
	}

	seen := map[string]bool{}
	for _, parentAlias := range aliases {
		normalizedParentAlias := publicExportNormalizedName(parentAlias)
		normalizedFieldAlias := publicExportNormalizedName(edge.alias)
		if normalizedParentAlias == "" || normalizedFieldAlias == "" {
			continue
		}

		key := normalizedParentAlias + "." + normalizedFieldAlias
		if seen[key] {
			continue
		}
		seen[key] = true

		for _, targetName := range []string{edge.target.Name, edge.target.OriginalName} {
			normalizedTargetName := publicExportNormalizedName(targetName)
			if normalizedTargetName == "" {
				continue
			}
			if normalizedTargetName == normalizedParentAlias+normalizedFieldAlias {
				return true
			}
			if normalizedTargetName == normalizedParentAlias && strings.HasSuffix(normalizedParentAlias, normalizedFieldAlias) {
				return true
			}
		}
	}

	return false
}

func publicExportAmbientCanExposeFieldChild(
	root publicExportAmbientRoot,
	parent *TypeDef,
	target *TypeDef,
	alias string,
	fieldName string,
	groupParts []string,
	rootTypes map[string]bool,
) bool {
	if publicExportAmbientIsCurrentRoot(root, parent, groupParts) || target == nil || publicExportAmbientIsRoot(target, rootTypes) {
		return true
	}
	if publicExportAmbientFieldChildOwnedByParent(parent, target, alias, fieldName) {
		return true
	}
	return publicExportAmbientIsLeafValueObject(target)
}

func publicExportAmbientFieldChildOwnedByParent(parent *TypeDef, target *TypeDef, alias string, fieldName string) bool {
	if parent == nil || target == nil {
		return false
	}

	parentNames := []string{parent.Name, parent.OriginalName}
	fieldAliases := []string{alias, casing.New().ToPascal(fieldName)}
	for _, parentName := range parentNames {
		normalizedParentName := publicExportNormalizedName(parentName)
		if normalizedParentName == "" {
			continue
		}
		for _, fieldAlias := range fieldAliases {
			normalizedFieldAlias := publicExportNormalizedName(fieldAlias)
			if normalizedFieldAlias == "" {
				continue
			}
			if publicExportAmbientTargetNameMatches(target, normalizedParentName+normalizedFieldAlias) {
				return true
			}
		}
	}
	return false
}

func publicExportAmbientTargetNameMatches(target *TypeDef, normalizedName string) bool {
	if target == nil || normalizedName == "" {
		return false
	}
	for _, targetName := range []string{target.Name, target.OriginalName} {
		if publicExportNormalizedName(targetName) == normalizedName {
			return true
		}
	}
	return false
}

func publicExportAmbientIsLeafValueObject(t *TypeDef) bool {
	if t == nil || t.Type != DataTypeClass || len(t.Fields) == 0 {
		return false
	}
	for _, field := range t.Fields {
		if field == nil {
			continue
		}
		if publicExportAmbientContainsUntypedMap(field.Type) {
			return false
		}
		child := publicExportAmbientUnwrapContainer(field.Type)
		if publicExportCanExpose(child) {
			return false
		}
	}
	return true
}

func publicExportAmbientContainsUntypedMap(t *TypeDef) bool {
	for t != nil {
		switch t.Type {
		case DataTypeArray, DataTypeSet:
			t = t.ItemType
		case DataTypeMap:
			return t.ItemType == nil || t.ItemType.Type == DataTypeAny
		default:
			return false
		}
	}
	return false
}

func publicExportAmbientParentConstAliases(parent *TypeDef) []string {
	if parent == nil {
		return nil
	}

	seen := map[string]bool{}
	var aliases []string
	addAlias := func(alias string) {
		alias = casing.New().ToPascal(strings.TrimSpace(alias))
		normalized := publicExportNormalizedName(alias)
		if normalized == "" || seen[normalized] {
			return
		}
		seen[normalized] = true
		aliases = append(aliases, alias)
	}

	for _, frame := range parent.ContextStack {
		if frame.Type != ContextTypeConstProperty {
			continue
		}
		addAlias(frame.DisplayName())
		addAlias(frame.Identifier)
	}

	for _, field := range parent.Fields {
		if !publicExportAmbientConstFieldIdentifiesParent(parent, field) {
			continue
		}
		if value, ok := field.Const.Value.(string); ok {
			addAlias(value)
		}
	}

	return aliases
}

func publicExportAmbientConstFieldIdentifiesParent(parent *TypeDef, field *FieldDef) bool {
	if parent == nil || field == nil || field.Const == nil {
		return false
	}
	if parent.Discriminator != nil && parent.Discriminator.TypePropertyName != "" {
		return publicExportNormalizedName(field.Name) == publicExportNormalizedName(parent.Discriminator.TypePropertyName)
	}

	matches := 0
	for _, candidate := range parent.Fields {
		if candidate == nil || candidate.Const == nil {
			continue
		}
		if _, ok := candidate.Const.Value.(string); ok {
			matches++
		}
	}
	_, fieldIsStringConst := field.Const.Value.(string)
	return fieldIsStringConst && matches == 1
}

func publicExportAmbientCanInferFromRootName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	return name == casing.New().ToPascal(name)
}

func publicExportAmbientAliasMatchesTarget(alias string, target *TypeDef) bool {
	if target == nil {
		return false
	}
	normalizedAlias := publicExportNormalizedName(alias)
	return normalizedAlias != "" &&
		(normalizedAlias == publicExportNormalizedName(target.Name) ||
			normalizedAlias == publicExportNormalizedName(target.OriginalName))
}

func publicExportAmbientFieldAlias(field *FieldDef, target *TypeDef) string {
	if field == nil {
		return ""
	}

	name := field.OriginalName
	if name == "" {
		name = field.Name
	}
	if field.Type != nil && (field.Type.Type == DataTypeArray || field.Type.Type == DataTypeSet) {
		name = publicExportAmbientSingularName(name)
	}

	alias := casing.New().ToPascal(name)
	if target != nil && publicExportNormalizedName(alias) == publicExportNormalizedName(target.Name) {
		return target.Name
	}
	return alias
}

func publicExportAmbientAssociatedAlias(parentName string, target *TypeDef) string {
	name := ""
	if target != nil {
		if discriminant := publicExportAmbientDiscriminantAlias(parentName, target); discriminant != "" {
			return discriminant
		}
		name = target.Name
		if name == "" {
			name = target.OriginalName
		}
	}
	if name == "" {
		return ""
	}

	trimmedName := publicExportAmbientTrimParentSuffix(parentName, name)
	if trimmedName != "" {
		return trimmedName
	}
	return casing.New().ToPascal(name)
}

func publicExportAmbientDiscriminantAlias(parentName string, target *TypeDef) string {
	if target == nil {
		return ""
	}
	constAlias := ""
	for _, field := range target.Fields {
		if !publicExportAmbientConstFieldIdentifiesParent(target, field) {
			continue
		}
		if value, ok := field.Const.Value.(string); ok {
			constAlias = casing.New().ToPascal(value)
			break
		}
	}
	if constAlias == "" {
		return ""
	}

	for _, candidate := range []string{target.Name, target.OriginalName} {
		if candidate == "" {
			continue
		}
		if publicExportNormalizedName(candidate) == publicExportNormalizedName(constAlias) {
			return candidate
		}
		trimmed := publicExportAmbientTrimParentSuffix(parentName, candidate)
		if publicExportNormalizedName(trimmed) == publicExportNormalizedName(constAlias) {
			return trimmed
		}
	}
	return constAlias
}

func publicExportAmbientTrimParentSuffix(parentName string, name string) string {
	for _, suffix := range publicExportAmbientPascalSuffixes(parentName) {
		if suffix == "" || suffix == name || !strings.HasSuffix(name, suffix) {
			continue
		}
		trimmed := strings.TrimSuffix(name, suffix)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func publicExportAmbientPascalSuffixes(name string) []string {
	var suffixes []string
	for i, r := range name {
		if i == 0 || r < 'A' || r > 'Z' {
			continue
		}
		suffixes = append(suffixes, name[i:])
	}
	slices.SortFunc(suffixes, func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	return suffixes
}

func publicExportAmbientSingularName(name string) string {
	if name == "data" {
		return name
	}
	return publicExportAmbientSingularize(name)
}

func publicExportAmbientCachedSingularize() func(string) string {
	mu := sync.Mutex{}
	cache := map[string]string{}
	return func(s string) string {
		mu.Lock()
		defer mu.Unlock()
		if cached, ok := cache[s]; ok {
			return cached
		}
		singularized := publicExportAmbientPluralizeClient.Singular(s)
		cache[s] = singularized
		return singularized
	}
}
