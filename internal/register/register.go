package register

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

/*
Register

The type register ensures that we have a minimal set of types needed in our SDKs
It helps recognize when we can, and when we cannot de-duplicate types.
The primary method is just:

RegisterType(ctx, typeDef, isInputType)

This register a given typeDef. It might return a different typeDef if it determines it is appropriate.

Extra methods to handle some special cases: in particular some types are expected to be *mutated*
*/
type Register struct {
	// types in the type register are generally considered immutable.
	Types                *sequencedmap.Map[string, *ast.TypeDef]
	TypesUsedInWeakUnion map[string]*ast.TypeDef

	// mutableTypes represents an exception to immutability: during the construction of an operation response shared types can migrate to error types
	// depending on information in the x-speakeasy-errors extensions
	mutableTypes *sequencedmap.Map[string, ast.TypeDefs]

	// duplicateTypes helps to reduce our escape hatch (same registry ID, different typedef) duplications
	duplicateTypes *sequencedmap.Map[string, ast.TypeDefs]
	// truncatedTypes are used to help manage circular references
	truncatedTypes *sequencedmap.Map[string, *ast.TypeDef]

	Config *configuration.Config
}

func New() *Register {
	return &Register{
		Types:                sequencedmap.New[string, *ast.TypeDef](),
		TypesUsedInWeakUnion: make(map[string]*ast.TypeDef),
		mutableTypes:         sequencedmap.New[string, ast.TypeDefs](),
		duplicateTypes:       sequencedmap.New[string, ast.TypeDefs](),
		truncatedTypes:       sequencedmap.New[string, *ast.TypeDef](),
	}
}

func (r *Register) IsTypeRegistered(typedef *ast.TypeDef) bool {
	t2, ok := r.Types.Get(typedef.GetRegistrationID())
	return ok && t2 == typedef
}

func (r *Register) AllTypes() ast.TypeDefs {
	return slices.Collect(r.Types.Values())
}

// RegisterType registers a type based on its name and context.
func (r *Register) RegisterType(ctx context.Context, typeDef *ast.TypeDef, isInputType bool) *ast.TypeDef {
	if typeDef == nil {
		return nil
	}

	if !typeDef.IsCustomType() {
		return typeDef
	}

	typeDef.FreezeOriginalName()

	// NOTE: there is a big assumption here that we will always see the truncated type before the non truncated type
	// which is true at the moment but if that ever changes the below code won't work

	if typeDef.Truncated {
		id := typeDef.GetRegistrationID()
		if isInputType {
			id += "_Input"
		} else {
			id += "_Output"
		}

		// Certain schemas may contain multiple circular references in them. For
		// any we've already seen make sure we return the one we're tracking in
		// the set so that when the real TypeDef is registered all truncated
		// references are updated.
		existing, ok := r.truncatedTypes.Get(id)
		if ok {
			existing.UsedInUnion = existing.UsedInUnion || typeDef.UsedInUnion
			return existing
		}

		r.truncatedTypes.Set(id, typeDef)
		return typeDef
	} else {
		id := typeDef.GetRegistrationID(ast.WithSkipInputOutputFrame())
		if isInputType {
			id += "_Input"
		} else {
			id += "_Output"
		}
		existing, ok := r.truncatedTypes.Get(id)
		if ok {
			*existing = *typeDef
			existing.UsedInUnion = existing.UsedInUnion || typeDef.UsedInUnion
			existing.DeduplicatedContextStacks = append(existing.DeduplicatedContextStacks, typeDef.ContextStack)
			r.truncatedTypes.Delete(id)
			return r.RegisterType(ctx, existing, isInputType)
		}
	}

	isErrorType := typeDef.ContextStack.HasFrameOfType(ast.ContextTypeResponseError)
	id := typeDef.GetRegistrationID()

	existing, ok := r.Types.Get(id)
	if !ok {
		typeDef.Registered = true
		r.Types.Set(id, typeDef)
		if isErrorType {
			r.mutableTypes.Set(id, append(r.mutableTypes.GetOrZero(id), typeDef))
		}
		r.duplicateTypes.Set(id, append(r.duplicateTypes.GetOrZero(id), typeDef))
		return typeDef
	} else if !isErrorType {
		r.mutableTypes.Delete(id)
	}

	err := existing.IsEqual(typeDef)
	if err != nil {
		switch {
		case errors.Is(err, ast.ErrFieldMismatch):
			fallthrough
		case errors.Is(err, ast.ErrUnionDiscriminatorMismatch):
			fallthrough
		case errors.Is(err, ast.ErrUnionTypeMismatch):
			fallthrough
		case errors.Is(err, ast.ErrEnumMismatch):
			fallthrough
		case errors.Is(err, ast.ErrTypeMismatch):
			// Special case for shared types that are used as errors they get moved to the errors scope
			for _, t := range r.duplicateTypes.GetOrZero(id) {
				if err := t.IsEqual(typeDef); err == nil {
					// Supplement it. This is mostly for debugging purposes.
					supplementErrorContextFrame(t, typeDef)
					return t
				}
			}

			if env.IsDebug() {
				idsDiffer := typeDef.GetRegistrationID() != existing.GetRegistrationID()
				msg := fmt.Sprintf("Found duplicate type for %q - %q: %v", typeDef.GetRegistrationID(), existing.GetRegistrationID(), err)
				if !idsDiffer {
					msg = fmt.Sprintf("Found duplicate type for %q: %v", typeDef.GetRegistrationID(), err)
				}
				logging.From(ctx).Error(msg)
			}

			count := len(r.duplicateTypes.GetOrZero(id))
			r.duplicateTypes.Set(id, append(r.duplicateTypes.GetOrZero(id), typeDef))

			typeDef.ContextStack.AppendWithHumanized(ast.ContextTypeRegisterDuplicate, strconv.Itoa(count+1), "")

			typeDef.Registered = true
			id = typeDef.GetRegistrationID()
			if isErrorType {
				r.mutableTypes.Set(id, append(r.mutableTypes.GetOrZero(id), typeDef))
			}
			r.Types.Set(id, typeDef)
			return typeDef
		case errors.Is(err, ast.ErrScopeMismatch):
			// Special case for shared types that are used as errors they get moved to the errors scope
			if typeDef.Scope == ast.ScopeShared && existing.Scope == ast.ScopeErrors || typeDef.Scope == ast.ScopeErrors && existing.Scope == ast.ScopeShared {
				return existing
			}
			fallthrough
		default:
			panic(fmt.Errorf("type %s already registered with different definition %s: %w", id, existing.GetRegistrationID(), err))
		}
	}

	// Found existing type make sure annotations are merged and return the existing type
	for i, f := range typeDef.Fields {
		for _, a := range f.Annotations {
			if i >= len(existing.Fields) {
				continue
			}

			_, existingAnnotation := existing.Fields[i].Annotations.Find(a)

			if existingAnnotation == nil {
				existing.Fields[i].Annotations = append(existing.Fields[i].Annotations, a)
			}
		}
	}

	existing.UsedInUnion = existing.UsedInUnion || typeDef.UsedInUnion
	existing.DeduplicatedContextStacks = append(existing.DeduplicatedContextStacks, typeDef.ContextStack)

	return existing
}

func (r *Register) UnregisterOrCloneTypeInPreparationForCreatingErrorType(ctx context.Context, typ *ast.TypeDef) *ast.TypeDef {
	id := typ.GetRegistrationID()
	_, canMove := r.mutableTypes.Get(id)

	if canMove {
		// If the type is only referenced in error responses we prefer to move it
		// to avoid having duplicate types in the SDK. Mutable types intends to keep
		// track of types that have only been used in error responses.
		r.UnregisterType(typ)
		return typ
	}
	// This type is already used elsewhere in a non-error response so we must clone it
	tmp := *typ
	typ = &tmp
	return typ
}

// Deprecated: use g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025
func (r *Register) UpdateRegisteredTypeToErrorsScopeLegacy(ctx context.Context, typ *ast.TypeDef, isRootType bool) *ast.TypeDef {
	id := typ.GetRegistrationID()

	if preExisting, ok := r.Types.Get(id); ok && typ.Scope == ast.ScopeErrors {
		return preExisting
	}

	_, isOpenType := r.mutableTypes.Get(id)

	// If this type is already used elsewhere in the SDK we need to make a copy of it
	if isRootType && typ.IsComponent && !isOpenType {
		newType := *typ
		typ = &newType
	} else {
		r.UnregisterType(typ)
	}

	// Drop the duplicate frames: they would create unnecessary "1", "2" suffixes
	// Note: this mutates both the original and the copy
	typ.ContextStack.Filter(func(frame ast.ContextFrame) bool {
		return frame.Type != ast.ContextTypeRegisterDuplicate
	})

	typ.Scope = ast.ScopeErrors
	if isRootType {
		typ.Type = ast.DataTypeError
	}

	return r.RegisterType(ctx, typ, false)
}

func supplementErrorContextFrame(target *ast.TypeDef, source *ast.TypeDef) {
	if target == nil || source == nil {
		return
	}
	targetContextFrame := target.ContextStack.FindLastFrameOfType(ast.ContextTypeResponseError)
	sourceContextFrame := source.ContextStack.FindLastFrameOfType(ast.ContextTypeResponseError)
	if targetContextFrame == nil || sourceContextFrame == nil {
		return
	}

	// Create a set to store unique codes from both target and source
	codeSet := make(map[string]struct{})

	// Extract codes from the target context frame
	targetCodes := strings.Split(targetContextFrame.Identifier, ",")
	for _, code := range targetCodes {
		code = strings.TrimSpace(code)
		if code != "" {
			codeSet[code] = struct{}{}
		}
	}

	// Extract codes from the source context frame
	sourceCodes := strings.Split(sourceContextFrame.Identifier, ",")
	for _, code := range sourceCodes {
		code = strings.TrimSpace(code)
		if code != "" {
			codeSet[code] = struct{}{}
		}
	}

	// Collect the merged codes and sort them alphabetically
	mergedCodes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		mergedCodes = append(mergedCodes, code)
	}
	sort.Strings(mergedCodes)

	// Update the target context frame with the merged codes
	targetContextFrame.Identifier = strings.Join(mergedCodes, ",")
}

func (r *Register) UnregisterType(typeDef *ast.TypeDef) {
	if typeDef == nil {
		return
	}

	if !typeDef.IsCustomType() {
		return
	}

	id := typeDef.GetRegistrationID()
	r.Types.Delete(id)

	idWithoutFrame := typeDef.GetRegistrationID(ast.WithSkipDuplicateFrame())
	r.duplicateTypes.Set(idWithoutFrame, slices.DeleteFunc(r.duplicateTypes.GetOrZero(idWithoutFrame), func(t *ast.TypeDef) bool {
		return t.GetRegistrationID() == id
	}))
	r.mutableTypes.Delete(id)
}

// Notice: this won't differentiate between types with the same name in different contexts, is that a problem?
func (r *Register) RegisterTypeUsedInWeakUnion(typeDef *ast.TypeDef) {
	if typeDef == nil {
		return
	}

	if !typeDef.IsCustomType() {
		return
	}

	id := typeDef.GetRegistrationID()

	r.TypesUsedInWeakUnion[id] = typeDef
}

func (r *Register) CloseType(typeDef *ast.TypeDef) {
	id := typeDef.GetRegistrationID()
	r.mutableTypes.Delete(id)
}

func (r *Register) CloseTypes(resp *ast.Response) error {
	for id := range r.mutableTypes.All() {
		r.mutableTypes.Delete(id)
	}
	// These checks are for sanity; if we have errors here, we have an issue with creating types that are not registered.
	if resp == nil {
		return nil
	}
	for _, types := range resp.Responses {
		for _, content := range types.Content {
			if content.Content == nil {
				continue
			}
			contentType := content.Content.Type

			if contentType.Type == ast.DataTypeClass || contentType.Type == ast.DataTypeError {
				id := contentType.GetRegistrationID()
				_, ok := r.Types.Get(id) // See below // found, ok := r.Types.Get(id)
				if !ok {
					return fmt.Errorf("unhandled type %s", id)
				}
				// This check needs follow-up before it can safely enforce type identity.
				// if found != content.Content.Type {
				// 	return fmt.Errorf("Unhandled (mismatched) type %s", id)
				// }
			}

			var err error
			for _, t := range contentType.Walk() {
				if t.Type == ast.DataTypeClass || t.Type == ast.DataTypeError {
					id := t.GetRegistrationID()
					_, ok := r.Types.Get(id) // See below // found, ok := r.Types.Get(id)
					if !ok {
						err = fmt.Errorf("unknown type %s", id)
						return err
					}
					// This check needs follow-up before it can safely enforce type identity.
					// if found != t {
					// 	err = fmt.Errorf("unhandled (mismatched) type %s", id)
					// 	return false
					// }
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
