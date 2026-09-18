package namer

import (
	"context"
	"fmt"
	"math"
	"path"
	"slices"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/register"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type Resolver struct {
	enumNamer   enumNamer
	subsystem   *subsystem.Subsystem
	logger      logging.Logger
	labelsCache map[*ast.TypeDef]*TypeDefLabels
}

func NewResolver(subsystem *subsystem.Subsystem, logger logging.Logger) (*Resolver, error) {
	return &Resolver{
		enumNamer:   newEnumNamer(subsystem.Target.Target),
		subsystem:   subsystem,
		logger:      logger,
		labelsCache: make(map[*ast.TypeDef]*TypeDefLabels),
	}, nil
}

// resolveEnums should be set to false for first run to ensure we don't resolve enums before types are resolved
func (r *Resolver) ResolveNames(ctx context.Context, register *register.Register, cfg configuration.ImportConfig, maintainOriginalOrder, resolveEnums bool) {
	registeredTypes := register.Types

	if !maintainOriginalOrder {
		// We sort the keys alphabetically to ensure we always resolve in the same order (for consistency)
		registeredTypes = sequencedmap.From(registeredTypes.AllOrdered(sequencedmap.OrderKeyAsc))
	}

	for t := range registeredTypes.Values() {
		t.FreezeOriginalName()
		if t.Name == "" {
			r.handleTypeWithNoName(ctx, t)
		}

		if t.IsStructurallyDeduplicated() {
			preProcessStructurallyDeduplicatedTypes(t)
		}

		r.ensureNamingConventions(ctx, t)
	}

	var buckets *sequencedmap.Map[string, *sequencedmap.Map[string, ast.TypeDefs]]

	switch cfg.Option {
	case configuration.ImportOptionOpenAPI:
		buckets = r.getOpenAPIBuckets(registeredTypes, cfg, resolveEnums)
	default:
		panic(fmt.Sprintf("unknown import option %s", cfg.Option))
	}

	for bucket := range buckets.Values() {
		r.resolveBucket(ctx, bucket, resolveEnums)
	}

	if !resolveEnums {
		r.ResolveNames(ctx, register, cfg, maintainOriginalOrder, true)

		// Sanity check that we didn't miss any types
		for t := range registeredTypes.Values() {
			if t.Name == "" {
				logging.From(ctx).Error(fmt.Sprintf("type %q has no name", t.GetRegistrationID()))
				if env.IsDebug() {
					panic(fmt.Sprintf("type %q has no name", t.GetRegistrationID()))
				}
			}
		}
	}

	r.fixBuiltInErrorNameConflicts(cfg, buckets)
}

func (r *Resolver) getOpenAPIBuckets(registeredTypes *sequencedmap.Map[string, *ast.TypeDef], cfg configuration.ImportConfig, resolveEnums bool) *sequencedmap.Map[string, *sequencedmap.Map[string, ast.TypeDefs]] {
	buckets := sequencedmap.New[string, *sequencedmap.Map[string, ast.TypeDefs]]()

	if cfg.Paths == nil {
		cfg.Paths = map[string]string{}
	}

	for t := range registeredTypes.Values() {
		bucketName := getBucketPathForType(cfg, t)
		t.OutputLocation = bucketName

		bucket, ok := buckets.Get(bucketName)
		if !ok {
			bucket = sequencedmap.New[string, ast.TypeDefs]()
		}

		// Rename the type with any frames that must be applied then create sub buckets based on the type name
		if !resolveEnums {
			renameWithMustFrames(t, r.subsystem.Config.Generation.GetNameResolution())
		}

		if resolveEnums && r.enumNamer != nil && t.Type == ast.DataTypeEnum {
			// Enums in Go (maybe other future languages) are not namespaced so can conflict with type names
			enumNames := r.enumNamer.EnumNames(t)
			t.CachedEnumNames = enumNames
			for _, enumName := range t.CachedEnumNames {
				types, _ := bucket.Get(enumName)
				bucket.Set(enumName, append(types, t))
			}
		}

		name := normalizeClassName(t.Name)

		types, _ := bucket.Get(name)
		bucket.Set(name, append(types, t))

		buckets.Set(bucketName, bucket)
	}

	return buckets
}

func (r *Resolver) resolveBucket(ctx context.Context, bucket *sequencedmap.Map[string, ast.TypeDefs], resolveEnums bool) {
	foundBucketWithDupes := false

	for b, types := range bucket.All() {
		if len(types) == 1 {
			continue
		}

		if r.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
			_ = r.RenameTypesWithDuplicateNames(ctx, types)
			continue
		}

		// --- Below code is for legacy name resolution Dec 2023 ---
		foundBucketWithDupes = true
		type typeToCheck struct {
			t     *ast.TypeDef
			moved bool
		}
		typesToCheck := make([]*typeToCheck, len(types))

		for i, t := range types {
			typesToCheck[i] = &typeToCheck{
				t: t,
			}
		}

		otherIdx := 0
		currIdx := 0

		for currIdx < len(typesToCheck) {

			if otherIdx >= len(typesToCheck) {
				currIdx++
				otherIdx = 0
				continue
			}

			toCheck := typesToCheck[currIdx]
			other := typesToCheck[otherIdx]

			if toCheck == other {
				otherIdx++
				continue
			}
			if toCheck.moved {
				currIdx++
				otherIdx = 0
				continue
			}
			if other.moved {
				otherIdx++
				continue
			}

			if r.renameConflictingTypesDec2023(ctx, toCheck.t, other.t, resolveEnums) {
				name := normalizeClassName(toCheck.t.Name)
				types, _ := bucket.Get(name)
				bucket.Set(name, append(types, toCheck.t))
				toCheck.moved = true
				currIdx++
				otherIdx = 0
			} else {
				name := normalizeClassName(other.t.Name)
				types, _ := bucket.Get(name)
				bucket.Set(name, append(types, other.t))
				other.moved = true
				otherIdx++
			}
		}

		typesToKeep := ast.TypeDefs{}
		for _, t := range typesToCheck {
			if !t.moved {
				typesToKeep = append(typesToKeep, t.t)
			}
		}

		// Refill current sub bucket with types that were not renamed
		bucket.Set(b, typesToKeep)
	}

	if !foundBucketWithDupes {
		return
	}

	// Go over buckets again to make sure we have resolved all dupes
	r.resolveBucket(ctx, bucket, resolveEnums)
}

func (r *Resolver) renameConflictingTypesDec2023(ctx context.Context, toCheck *ast.TypeDef, other *ast.TypeDef, resolveEnums bool) bool {
	// below we check the type of model and resolve in this order of precedence:
	// 1. Main SDK types
	// 2. Shared components (errors treated as shared)
	// 3. SubSDKs
	// 4. Shared models that are not components
	// 5. Request/Response types
	// 6. Operations (and similar)
	switch {
	// Enums in Go (maybe other future languages) are not namespaced so can conflict with type names
	case resolveEnums && r.isEnum(toCheck, other):
		return r.compareEnums(ctx, toCheck, other)
	case resolveEnums && r.isEnum(other, toCheck):
		return r.compareEnums(ctx, other, toCheck)
	// Main SDK Types
	case isMainSDKType(toCheck):
		return r.compareMainSDKs(ctx, toCheck, other)
	case isMainSDKType(other):
		return !r.compareMainSDKs(ctx, other, toCheck)
	// Shared Components (errors treated as shared)
	case isSharedComponent(toCheck):
		return r.compareSharedComponent(ctx, toCheck, other)
	case isSharedComponent(other):
		return !r.compareSharedComponent(ctx, other, toCheck)
	// SubSDKs
	case toCheck.Scope == ast.ScopeSDK:
		return r.compareSubSDK(ctx, toCheck, other)
	case other.Scope == ast.ScopeSDK:
		return !r.compareSubSDK(ctx, other, toCheck)
	// Shared Models that are not components
	case toCheck.Scope == ast.ScopeShared:
		return r.compareSharedModel(ctx, toCheck, other)
	case other.Scope == ast.ScopeShared:
		return !r.compareSharedModel(ctx, other, toCheck)
	// Request/Response types
	case isRequestOrResponseType(toCheck):
		return r.compareRequestOrResponseType(ctx, toCheck, other)
	case isRequestOrResponseType(other):
		return !r.compareRequestOrResponseType(ctx, other, toCheck)
	// Operations (and similar)
	default:
		return r.compareOperationsModels(ctx, toCheck, other)
	}
}

func (r *Resolver) compareEnums(ctx context.Context, e *ast.TypeDef, other *ast.TypeDef) bool {
	switch {
	case r.isEnum(e, other):
		r.namespaceName(ctx, other, e)
		return false
	default:
		r.namespaceName(ctx, e, other)
		return true
	}
}

// compareMainSDKs will compare two main SDK types and return true if the first type was renamed otherwise
// it will return false signifying the second type was renamed
func (r *Resolver) compareMainSDKs(ctx context.Context, sdk *ast.TypeDef, other *ast.TypeDef) bool {
	switch {
	case isMainSDKType(other):
		panic("shouldn't be possible")
	// Anything else is a lower precedence
	default:
		r.namespaceName(ctx, other, sdk)
		return false
	}
}

func (r *Resolver) compareSharedComponent(ctx context.Context, sc *ast.TypeDef, other *ast.TypeDef) bool {
	switch {
	case isSharedComponent(other):
		switch {
		case sc.Scope == ast.ScopeErrors && other.Scope == ast.ScopeShared:
			r.namespaceName(ctx, sc, other)
			return true
		case other.Scope == ast.ScopeErrors && sc.Scope == ast.ScopeShared:
			r.namespaceName(ctx, other, sc)
			return false
		default:
			switch compareReferenceTypes(sc, other) {
			case -1:
				r.namespaceName(ctx, other, sc)
				return false
			case 1:
				r.namespaceName(ctx, sc, other)
				return true
			default:
				// If we have two matching components then the tie break will be which has the shortest context stack as its "higher level"
				switch compareContextStacks(sc, other) {
				case 1:
					r.namespaceName(ctx, sc, other)
					return true
				default:
					r.namespaceName(ctx, other, sc)
					return false
				}
			}
		}
	// Anything else is a lower precedence
	default:
		r.namespaceName(ctx, other, sc)
		return false
	}
}

func (r *Resolver) compareSubSDK(ctx context.Context, sub *ast.TypeDef, other *ast.TypeDef) bool {
	switch other.Scope {
	case ast.ScopeSDK:
		switch compareGroups(sub, other) {
		case 1:
			r.namespaceName(ctx, sub, other)
			return true
		default:
			r.namespaceName(ctx, other, sub)
			return false
		}
	// Anything else is a lower precedence
	default:
		r.namespaceName(ctx, other, sub)
		return false
	}
}

func (r *Resolver) compareSharedModel(ctx context.Context, s *ast.TypeDef, other *ast.TypeDef) bool {
	switch other.Scope {
	case ast.ScopeShared:
		switch compareContextStacks(s, other) {
		case 1:
			r.namespaceName(ctx, s, other)
			return true
		default:
			r.namespaceName(ctx, other, s)
			return false
		}
	// Anything else is a lower precedence
	default:
		r.namespaceName(ctx, other, s)
		return false
	}
}

func (r *Resolver) compareRequestOrResponseType(ctx context.Context, rt *ast.TypeDef, other *ast.TypeDef) bool {
	switch {
	case isRequestOrResponseType(other):
		switch compareOperationTags(rt, other) {
		case 1:
			r.namespaceName(ctx, rt, other)
			return true
		default:
			r.namespaceName(ctx, other, rt)
			return false
		}
	// Anything else is a lower precedence
	default:
		r.namespaceName(ctx, other, rt)
		return false
	}
}

func (r *Resolver) compareOperationsModels(ctx context.Context, o *ast.TypeDef, other *ast.TypeDef) bool {
	// For operation models we will try to resolve using this order of precedence:
	// 1. Request body types have higher precedence than response types due to more likely needing to refer to the type to instantiate input
	//    a. Request bodies will have higher priority then parameters as they are more likely to be meaningful objects
	// 2. Parameters
	// 3. Response types
	switch {
	case o.ContextStack.HasFrameOfType(ast.ContextTypeRequestBody):
		r.namespaceName(ctx, other, o)
		return false
	case other.ContextStack.HasFrameOfType(ast.ContextTypeRequestBody):
		r.namespaceName(ctx, o, other)
		return true
	case o.ContextStack.HasFrameOfType(ast.ContextTypeParameter):
		r.namespaceName(ctx, other, o)
		return false
	case other.ContextStack.HasFrameOfType(ast.ContextTypeParameter):
		r.namespaceName(ctx, o, other)
		return true
	}

	switch compareContextStacks(o, other) {
	case 1:
		r.namespaceName(ctx, o, other)
		return true
	default:
		r.namespaceName(ctx, other, o)
		return false
	}
}

// Catches classes that only differ by case otherwise they will conflict when converted to file names in certain cases
func normalizeClassName(name string) string {
	return strings.ToLower(strcase.ToPascal(sanitization.SanitizeName(name)))
}

func (r *Resolver) namespaceName(ctx context.Context, t *ast.TypeDef, other *ast.TypeDef) {
	if t.OriginalName == "" {
		t.OriginalName = t.Name
	}

	switch {
	// Main SDK Types
	case isMainSDKType(t):
		panic("shouldn't be possible")
	// Shared Components (errors treated as shared)
	case isSharedComponent(t):
		r.renameSharedComponent(ctx, t, other)
	// SubSDKs
	case t.Scope == ast.ScopeSDK:
		r.renameSDK(ctx, t, other)
	// Shared Models that are not components
	case t.Scope == ast.ScopeShared:
		r.renameSharedModel(ctx, t, other)
	// Request/Response types
	case isRequestOrResponseType(t):
		renameRequestOrResponseType(ctx, t, other)
	// Operations (and similar)
	default:
		r.renameOperationModel(ctx, t, other)
	}

	// We need to regenerate the enum names cache for enums that are renamed
	if t.Type == ast.DataTypeEnum && len(t.CachedEnumNames) > 0 {
		t.CachedEnumNames = r.enumNamer.EnumNames(t)
	}
}

func (r *Resolver) renameSharedComponent(ctx context.Context, t *ast.TypeDef, other *ast.TypeDef) {
	name := normalizeClassName(t.Name)

	switch {
	case isMainSDKType(other):
		// If we conflict with an SDK start with a suffix
		if t.Scope == ast.ScopeShared && !strings.HasSuffix(strings.ToLower(name), "model") {
			t.Name += "_model"
			return
		}

		if t.Scope != ast.ScopeErrors {
			break
		}

		fallthrough
	case t.Scope == ast.ScopeErrors && other.Scope == ast.ScopeShared:
		// If we conflict with a shared model start with an Error suffix
		if !strings.HasSuffix(strings.ToLower(name), "error") {
			t.Name += "_error"
			return
		}
	}

	if renameWithInputOutput(t) {
		return
	}

	// If its between two shared components we will use the ref type as a namespace
	refTypeFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeRefType)
	otherRefTypeFrame := other.ContextStack.FindLastFrameOfType(ast.ContextTypeRefType)

	if refTypeFrame.Identifier != otherRefTypeFrame.Identifier && !refTypeFrame.Used {
		refType := refTypeFrame.Identifier
		// depluralize
		if refType == "requestBodies" {
			refType = "requestBody"
		} else {
			refType = strings.TrimSuffix(refType, "s")
		}

		t.Name = refType + "_" + t.Name
		refTypeFrame.Used = true
		return
	}

	incrementName(ctx, t)
}

func (r *Resolver) renameSDK(ctx context.Context, t *ast.TypeDef, other *ast.TypeDef) {
	switch {
	// Main SDK Types
	case isMainSDKType(other):
		fallthrough
	// Shared Components
	case isSharedComponent(other):
		// Add an SDK suffix first if it conflicts with anything
		if !strings.HasSuffix(strings.ToLower(t.Name), "sdk") {
			t.Name += "_SDK"
			return
		}
	}

	r.renameWithContextStack(ctx, t)
}

func (r *Resolver) renameSharedModel(ctx context.Context, t *ast.TypeDef, other *ast.TypeDef) {
	name := normalizeClassName(t.Name)

	// If we conflict with an SDK start with a Model suffix
	if other.Scope == ast.ScopeSDK {
		if !strings.HasSuffix(strings.ToLower(name), "model") {
			t.Name += "_model"
			return
		}
	}

	if r.subsystem.Config.Generation.NameResolutionAtLeastOrdered() {
		// Special case to rename a oneOf subtype with the oneOf name first
		if t.ContextStack.HasFrameOfType(ast.ContextTypeOneOf) {
			oneOfFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeOneOf)
			if oneOfFrame != nil && !oneOfFrame.Used {
				t.Name = oneOfFrame.Identifier + "_" + t.Name
				oneOfFrame.Used = true
				return
			}
		}
	}

	// then start with appending the parent ref name
	refNameFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeRefName)
	if refNameFrame != nil && !refNameFrame.Used {
		t.Name = refNameFrame.Identifier + "_" + t.Name
		refNameFrame.Used = true
		return
	}

	r.renameWithContextStack(ctx, t)
}

func renameRequestOrResponseType(ctx context.Context, t *ast.TypeDef, _ *ast.TypeDef) {
	// Request/Response types don't really have any other context to be renamed on
	// so we will just start adding numbers
	incrementName(ctx, t)
}

func (r *Resolver) renameOperationModel(ctx context.Context, t *ast.TypeDef, _ *ast.TypeDef) {
	// Special case to rename a parameter using the parameter type first
	if t.ContextStack.HasFrameOfType(ast.ContextTypeParameter) {
		parameterFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeParameter)
		if parameterFrame != nil && !parameterFrame.Used {
			t.Name = parameterFrame.Identifier + "_" + t.Name
			parameterFrame.Used = true
			return
		}
	}

	if r.subsystem.Config.Generation.NameResolutionAtLeastOrdered() {
		// Special case to rename a oneOf subtype with the oneOf name first
		if t.ContextStack.HasFrameOfType(ast.ContextTypeOneOf) {
			oneOfFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeOneOf)
			if oneOfFrame != nil && !oneOfFrame.Used {
				t.Name = oneOfFrame.Identifier + "_" + t.Name
				oneOfFrame.Used = true
				return
			}
		}
	}

	// Next we will use the operation name it will always be the first prefix
	operationFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeOperation)
	if operationFrame != nil && !operationFrame.Used {
		t.Name = operationFrame.Identifier + "_" + t.Name
		operationFrame.Used = true
		return
	}

	r.renameWithContextStack(ctx, t)
}

// Shorter context stack wins otherwise the less used context frames wins
func compareContextStacks(a *ast.TypeDef, b *ast.TypeDef) int {
	if len(a.ContextStack) < len(b.ContextStack) {
		return -1
	} else if len(a.ContextStack) > len(b.ContextStack) {
		return 1
	}

	usedContextFramesA := 0
	for _, frame := range a.ContextStack {
		if frame.Used {
			usedContextFramesA++
		}
	}

	usedContextFramesB := 0
	for _, frame := range b.ContextStack {
		if frame.Used {
			usedContextFramesB++
		}
	}

	if usedContextFramesA < usedContextFramesB {
		return -1
	}

	if usedContextFramesA > usedContextFramesB {
		return 1
	}

	return 0
}

// The shorter group stack wins
func compareGroups(a *ast.TypeDef, b *ast.TypeDef) int {
	groupsA := a.ContextStack.GetGroups()
	groupsB := b.ContextStack.GetGroups()

	if len(groupsA) < len(groupsB) {
		return -1
	} else if len(groupsA) > len(groupsB) {
		return 1
	}

	return 0
}

func compareReferenceTypes(a *ast.TypeDef, b *ast.TypeDef) int {
	aRefType := ""
	for _, frame := range a.ContextStack {
		if frame.Type == ast.ContextTypeRefType {
			aRefType = frame.Identifier
			break
		}
	}

	bRefType := ""
	for _, frame := range b.ContextStack {
		if frame.Type == ast.ContextTypeRefType {
			bRefType = frame.Identifier
			break
		}
	}

	if aRefType == bRefType {
		return 0
	}

	switch {
	case aRefType == "schemas":
		return -1
	case bRefType == "schemas":
		return 1
	case aRefType == "parameters":
		return -1
	case bRefType == "parameters":
		return 1
	case aRefType == "requestBodies":
		return -1
	case bRefType == "requestBodies":
		return 1
	case aRefType == "responses":
		return -1
	case bRefType == "responses":
		return 1
	case aRefType == "headers":
		return -1
	case bRefType == "headers":
		return 1
	}

	// Any other ref types are unlikely to be used for renaming
	return 0
}

// The shorter number of tags or the shortest tag wins
func compareOperationTags(a *ast.TypeDef, b *ast.TypeDef) int {
	aTags := []string{}
	for _, frame := range a.ContextStack {
		if frame.Type == ast.ContextTypeOperationTag {
			aTags = append(aTags, frame.Identifier)
		}
	}

	bTags := []string{}
	for _, frame := range b.ContextStack {
		if frame.Type == ast.ContextTypeOperationTag {
			bTags = append(bTags, frame.Identifier)
		}
	}

	if len(aTags) < len(bTags) {
		return -1
	}

	if len(aTags) > len(bTags) {
		return 1
	}

	shortestATag := math.MaxInt32
	for _, tag := range aTags {
		l := len(strings.Split(tag, "."))
		if l < shortestATag {
			shortestATag = l
		}
	}

	shortestBTag := math.MaxInt32
	for _, tag := range bTags {
		l := len(strings.Split(tag, "."))
		if l < shortestBTag {
			shortestBTag = l
		}
	}

	if shortestATag < shortestBTag {
		return -1
	}

	if shortestATag > shortestBTag {
		return 1
	}

	return 0
}

func (r *Resolver) isEnum(t, other *ast.TypeDef) bool {
	if r.enumNamer == nil {
		return false
	}

	if t.Type != ast.DataTypeEnum {
		return false
	}

	for _, enumName := range t.CachedEnumNames {
		if enumName == normalizeClassName(other.Name) {
			return true
		}
	}

	return false
}

func isMainSDKType(t *ast.TypeDef) bool {
	return t.Scope == ast.ScopeSDK && !t.ContextStack.HasFrameOfType(ast.ContextTypeMainSDK)
}

func isSharedComponent(t *ast.TypeDef) bool {
	return isSharedModel(t) && t.IsComponent
}

func isSharedModel(t *ast.TypeDef) bool {
	// Errors will be treated as shared models
	return t.Scope == ast.ScopeShared || t.Scope == ast.ScopeErrors
}

func isRequestOrResponseType(t *ast.TypeDef) bool {
	// There is only one frame in the context stack for these types
	if len(t.ContextStack) != 1 {
		return false
	}

	// If the only frame is a request/response type then we are one of these types
	if !t.ContextStack.HasFrameOfType(ast.ContextTypeRequestResponse) {
		return false
	}

	return true
}

func splitPrefix(name, originalName string) (string, string) {
	// find the index of the original name and split it into the prefix and the remainder
	splitIdx := strings.LastIndex(name, originalName)
	if splitIdx == -1 {
		return "", name
	}

	return name[:splitIdx], name[splitIdx:]
}

func (r *Resolver) renameWithContextStack(ctx context.Context, t *ast.TypeDef) {
	if renameWithInputOutput(t) {
		return
	}

	// start consuming context frames
	// we are building up the prefix so we need to break off the current index and append to that
	for i := range t.ContextStack {
		frame := &t.ContextStack[i]

		// We don't want to rename with the ref type or if the frame is already used
		if frame.Used || (r.subsystem.Config.Generation.NameResolutionAtLeastOrdered() && frame.Type == ast.ContextTypeRefType) || frame.Type == ast.ContextTypeRegisterDuplicate {
			continue
		}

		prefix, name := splitPrefix(t.Name, t.OriginalName)
		t.Name = prefix + frame.Identifier + "_" + name
		frame.Used = true
		return
	}

	incrementName(ctx, t)
}

func renameWithInputOutput(t *ast.TypeDef) bool {
	if !t.ContextStack.HasFrameOfType(ast.ContextTypeInputOutput) {
		return false
	}

	inputOutputFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeInputOutput)
	if inputOutputFrame.Used {
		return false
	}

	t.Name = t.Name + "_" + inputOutputFrame.Identifier
	inputOutputFrame.Used = true

	return true
}

func incrementName(ctx context.Context, t *ast.TypeDef) {
	if env.IsDebug() {
		logging.From(ctx).Debug(fmt.Sprintf("Incrementing name for %q", t.Name))
	}

	// If we get here we have no context frames left to use, so we will just start adding numbers
	t.Name = utils.IncrementName(t.Name)
}

func renameWithMustFrames(t *ast.TypeDef, mode config.NameResolutionMode) {
	if mode.AtLeast(config.NameResolutionShortest) {
		return
	}
	for i := len(t.ContextStack) - 1; i >= 0; i-- {
		frame := &t.ContextStack[i]

		if !frame.MustUse {
			continue
		}

		prefix, name := splitPrefix(t.Name, t.OriginalName)
		t.Name = prefix + frame.Identifier + "_" + name
		frame.Used = true
	}
}

func (r *Resolver) ensureNamingConventions(ctx context.Context, t *ast.TypeDef) {
	if !r.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
		return
	}

	isError := t.Type == ast.DataTypeError

	// Prepend with OperationID to inline request/response bodies
	if (t.IsInlineRequestBody || t.IsInlineResponseBody) && !isError {
		typeDefLabels := r.getOrCollectLabels(t)
		originalName := typeDefLabels.getLabelForSource(labelSource("original_name"))
		isUnnamedRequestBody := t.IsInlineRequestBody && (normalizeClassName(originalName) == "request" || normalizeClassName(originalName) == "requestbody")
		isUnnamedResponseBody := t.IsInlineResponseBody && (normalizeClassName(originalName) == "response" || normalizeClassName(originalName) == "responsebody")
		if (isUnnamedRequestBody || isUnnamedResponseBody) && !isError {
			opID := typeDefLabels.getLabelForSource(labelSource(ast.ContextTypeOperation))
			typeDefLabels.selectLabel(opID)
			r.renameTypeDef(ctx, t, "prepend_op_id")
		}
	}

	// Always suffix errors with "error"
	if isError && !r.subsystem.Config.Generation.SkipErrorSuffix {
		typeDefLabels := r.getOrCollectLabels(t)
		originalNameLower := strings.ToLower(t.Name)
		if !strings.HasPrefix(originalNameLower, "err") &&
			!strings.HasSuffix(originalNameLower, "err") &&
			!strings.Contains(originalNameLower, "error") &&
			!strings.Contains(originalNameLower, "exception") {
			errorSuffix := typeDefLabels.getLabelForSource(labelSource("data_type"))
			typeDefLabels.selectLabel(errorSuffix)
			r.renameTypeDef(ctx, t, "append_error")
		}
	}
}

func preProcessStructurallyDeduplicatedTypes(t *ast.TypeDef) {
	// Computed maps for each deduplicated context stack
	maps := make([]map[string]bool, len(t.DeduplicatedContextStacks))
	for i, stack := range t.DeduplicatedContextStacks {
		m := make(map[string]bool, len(stack))
		for _, frame := range stack {
			key := string(frame.Type) + "#" + frame.Identifier
			m[key] = true
		}
		maps[i] = m
	}

	// Intersect the context stacks using the computed maps
	// this is so we assign a name which is applicable to all the deduplicated schemas
	// this means that often things like the OperationID will not be an available identifier
	intersected := ast.ContextStack{}
	for _, frame := range t.ContextStack {
		key := string(frame.Type) + "#" + frame.Identifier
		presentOnAll := true
		for _, m := range maps {
			if !m[key] {
				presentOnAll = false
				break
			}
		}
		if presentOnAll {
			intersected = append(intersected, frame)
		}
	}

	t.ContextStack = intersected
}

// Returns the preferred and alternative error suffixes based on the target language.
func (r *Resolver) getErrorSuffixes() (string, string) {
	if slices.Contains([]string{"csharp", "java", "javav2", "php"}, r.subsystem.Target.Target) {
		return "Exception", "Error"
	}

	return "Error", "Exception"
}

// Resolves a conflict in the error name by trying the following consecutively:
// - insert the error type ("Base" or "Default"), if it is not already present in the original name
// - prepend the sdkClassName, provided it is not already present in the original error name
// - append the error type to the original error name anyway, as the last resort
func (r *Resolver) resolveErrorConflict(
	originalErrorName string, // the error name to resolve the conflict for
	baseOrDefault string, // the error type ("Base" or "Default") to be inserted or appended
) string {
	sdkName := r.subsystem.Config.Generation.SDKClassName

	if !strings.Contains(originalErrorName, baseOrDefault) {
		suffix1, suffix2 := r.getErrorSuffixes()
		for _, suffix := range []string{suffix1, suffix2} {
			if newErrorName, err := utils.InsertBefore(originalErrorName, suffix, baseOrDefault); err == nil {
				return newErrorName
			}
		}
	} else if !strings.Contains(originalErrorName, sdkName) {
		return sdkName + originalErrorName
	}

	return originalErrorName + baseOrDefault
}

func (r *Resolver) fixBuiltInErrorNameConflicts(cfg configuration.ImportConfig, buckets *sequencedmap.Map[string, *sequencedmap.Map[string, ast.TypeDefs]]) {
	if langCfg, ok := r.subsystem.Config.Languages[r.subsystem.Target.Target]; ok {
		// Read error names using overlay-aware accessor so that static overlay
		// values (e.g. from CLI's getConfigOverlay) are picked up without
		// needing to be written into gen.yaml.
		baseErrorName := ""
		if v, ok := r.subsystem.Config.GetLanguageConfigValue("baseErrorName").(string); ok {
			baseErrorName = v
		}
		defaultErrorName := ""
		if v, ok := r.subsystem.Config.GetLanguageConfigValue("defaultErrorName").(string); ok {
			defaultErrorName = v
		}
		originalBaseErrorName := baseErrorName
		originalDefaultErrorName := defaultErrorName

		errorSuffix, _ := r.getErrorSuffixes()

		if baseErrorName == "" {
			// Default to `PetStoreError/Exception`
			baseErrorName = r.subsystem.Config.Generation.SDKClassName + errorSuffix
		}
		if defaultErrorName == "" {
			// Default to `PetStoreDefaultError/Exception`
			defaultErrorName = r.subsystem.Config.Generation.SDKClassName + "Default" + errorSuffix
		}
		errorsBucketName := getBucketPathFromScope(cfg, ast.ScopeErrors)
		errorsBucket, ok := buckets.Get(errorsBucketName)
		if !ok {
			errorsBucket = sequencedmap.New[string, ast.TypeDefs]()
		}

		// The base error name conflicts with a user error
		if _, ok := errorsBucket.Get(strings.ToLower(baseErrorName)); ok {
			baseErrorName = r.resolveErrorConflict(baseErrorName, "Base")
		}

		// The default error name conflicts with a user error
		if _, ok := errorsBucket.Get(strings.ToLower(defaultErrorName)); ok {
			defaultErrorName = r.resolveErrorConflict(defaultErrorName, "Default")
		}

		// The base and fallback error names conflict
		if baseErrorName == defaultErrorName {
			baseErrorName = r.resolveErrorConflict(baseErrorName, "Base")
		}

		if originalBaseErrorName != baseErrorName {
			baseErrorName = strings.ReplaceAll(strcase.ToGoPascal(baseErrorName), "Sdk", "SDK")
		}

		if originalDefaultErrorName != defaultErrorName {
			defaultErrorName = strings.ReplaceAll(strcase.ToGoPascal(defaultErrorName), "Sdk", "SDK")
		}

		// Only write to Cfg when the value changed from what was already
		// resolvable (via overlay or Cfg). This avoids persisting static
		// overlay values into gen.yaml unnecessarily.
		if baseErrorName != originalBaseErrorName {
			langCfg.Cfg["baseErrorName"] = baseErrorName
		}
		if defaultErrorName != originalDefaultErrorName {
			langCfg.Cfg["defaultErrorName"] = defaultErrorName
		}
		r.subsystem.Config.Languages[r.subsystem.Target.Target] = langCfg
	}
}

func getBucketPathFromScope(cfg configuration.ImportConfig, scope ast.Scope) string {
	switch scope {
	case ast.ScopeShared, ast.ScopeOperations, ast.ScopeErrors, ast.ScopeWebhooks, ast.ScopeCallbacks:
		return cfg.GetImportPath(scope)
	case ast.ScopeSDK:
		return "" // The root of the SDK
	case ast.ScopeGlobals:
		return "globals"
	default:
		panic(fmt.Sprintf("unexpected scope %s", scope))
	}
}

// getBucketPathForType returns the output bucket path for a type, taking into account
// model namespace extensions. The namespace is resolved from either the type's own
// x-speakeasy-model-namespace extension or inherited from the parent via context stack.
func getBucketPathForType(cfg configuration.ImportConfig, t *ast.TypeDef) string {
	bucketPath := getBucketPathFromScope(cfg, t.Scope)

	namespace := getModelNamespace(t)
	if namespace == "" {
		return bucketPath
	}

	// Replace the last segment of a scope path with the custom namespace,
	// i.e. make the custom module a peer of the original directory.
	//
	//	"models/shared" + "foo" → "models/foo"
	//	"sdk/models/shared" + "foo" → "sdk/models/foo"
	//	"models" + "foo" → "models/foo"
	//	"" + "foo" → "foo"
	parent := path.Dir(bucketPath)
	if parent == "." && bucketPath != "" {
		parent = bucketPath
	}
	return path.Join(parent, namespace)
}

// getModelNamespace returns the model namespace for a type, checking the type's own
// extension first, then falling back to an inherited namespace from the context stack.
func getModelNamespace(t *ast.TypeDef) string {
	if t.Extensions != nil && t.Extensions.ModelNamespace != nil && *t.Extensions.ModelNamespace != "" {
		return *t.Extensions.ModelNamespace
	}

	if namespaceFrame := t.ContextStack.FindLastFrameOfType(ast.ContextTypeModelNamespace); namespaceFrame != nil {
		return namespaceFrame.Identifier
	}

	return ""
}
