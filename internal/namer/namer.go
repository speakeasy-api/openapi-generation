package namer

import (
	"context"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/openapi"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/references"
)

type Features interface {
	RecordFeatureUsage(ctx context.Context, feature features.Feature)
}
type Namer struct {
	Subsystem *subsystem.Subsystem
}

func New(ss *subsystem.Subsystem) *Namer {
	return &Namer{
		Subsystem: ss,
	}
}

func (n *Namer) GetFieldName(ctx context.Context, schema *oas3.JSONSchema[oas3.Referenceable], defaultName string, parentComponentRef references.Reference) (string, error) {
	if schema == nil {
		return defaultName, nil
	}

	name, _, _, err := n.GetTypeName(ctx, schema, nil, parentComponentRef, false)
	if err != nil {
		return "", err
	}

	if name == "" {
		name = defaultName
	}

	return name, nil
}

// GetTypeName returns the type name and ref type for a schema.
// The skipNestedRefTracking parameter controls whether the libopenapi nested ref bug behavior
// should be skipped. Set to true for array items and additionalProperties where the bug doesn't apply.
func (n *Namer) GetTypeName(ctx context.Context, s *oas3.JSONSchema[oas3.Referenceable], contextStack ast.ContextStack, parentComponentRef references.Reference, skipNestedRefTracking bool) (string, string, ast.ContextStack, error) {
	name, refType := GetRefName(s.GetRef())

	if name == "" && parentComponentRef != "" {
		name, refType = GetRefName(parentComponentRef)
	}

	// Track and handle nested references for the SharedNestedComponentsJan2026 feature flag.
	// When the flag is false, we replicate old libopenapi behavior where multiple schemas
	// referencing the same shared nested schema all resolve to the first referencer's name.
	// Skip this for array items and additionalProperties where the bug doesn't apply.
	if name != "" && !n.Subsystem.Config.Generation.Fixes.SharedNestedComponentsJan2026 && !skipNestedRefTracking {
		name = n.handleNestedReferenceTracking(s, name)
	}

	schema := s.MustGetResolvedSchema()

	if schema.IsBool() {
		return "", "", contextStack, nil
	}

	nameOverride, err := n.Subsystem.Extensions.HandleClassNameExtension(schema.GetSchema().GetExtensions())
	if err != nil {
		return "", "", nil, err
	}
	if nameOverride != nil {
		name = nameOverride.Name
		n.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureNameOverrides)
	}

	if name != "" {
		return name, refType, contextStack, nil
	}

	if schema.GetSchema().GetAnchor() != "" {
		return schema.GetSchema().GetAnchor(), "", contextStack, nil
	}

	if schema.GetSchema().GetTitle() != "" {
		return schema.GetSchema().GetTitle(), "", contextStack, nil
	}

	// We will use the last frame of the context stack and pop that off the stack used for the type
	lastFrame := contextStack.LastFrame()
	if lastFrame != nil {
		name := lastFrame.Identifier

		if n.Subsystem.Config.Generation.Fixes.NameResolutionFeb2025 {
			name = lastFrame.DisplayName()
		}

		if name != "" {
			contextStack.PopLastFrame()
			return name, "", contextStack, nil
		}
	}

	return "", "", contextStack, nil
}

func GetRefName(reference references.Reference) (string, string) {
	if reference == "" {
		return "", ""
	}

	ref := string(reference)

	refParts := []string{ref}
	if strings.Contains(ref, "#") {
		refParts = strings.Split(ref, "#")
		refParts = append([]string{refParts[0]}, "#"+strings.Join(refParts[1:], "#"))
	}
	if len(refParts) == 1 {
		return "", ""
	}

	switch {
	case strings.HasPrefix(refParts[1], "#/paths/"):
		return strings.TrimPrefix(refParts[1], "#/paths/"), "Path"
	default:
		parts := strings.Split(refParts[1], "/")
		return parts[len(parts)-1], strcase.ToGoPascal(parts[len(parts)-2])
	}
}

// handleNestedReferenceTracking looks up the original referencer's name for a shared
// nested reference from the pre-populated registry.
//
// The registry is populated by walking ALL schemas in the document and recording
// when a $ref is encountered that leads to nested references (chain length > 1).
// Only schemas that are FOUND via $ref are recorded.
//
// For example, given paths that $ref to Schema1 and Schema2, where:
//
//	Schema1: $ref: "#/components/schemas/SchemaShared"
//	Schema2: $ref: "#/components/schemas/SchemaShared"
//
// The registry would map SchemaShared -> Schema1 (first $ref encountered wins).
//
// KEY BEHAVIOR (replicating libopenapi bug from main):
// - When accessing Schema2 (alias), the final target (SchemaShared) is looked up in registry
// - SchemaShared maps to Schema1, so Schema2 becomes Schema1
// - When accessing SchemaShared directly, it maps to Schema1, so it becomes Schema1
// - First alias (Schema1) keeps its own name
//
// IMPORTANT: The registry must be pre-populated using openapi.PopulateFromDocument()
// before any schema processing begins.
func (n *Namer) handleNestedReferenceTracking(s *oas3.JSONSchema[oas3.Referenceable], currentName string) string {
	currentRef := string(s.GetRef())
	if currentRef == "" {
		return currentName
	}

	registry := openapi.GlobalNestedRefRegistry()

	// Ensure we have a resolved schema before inspecting the reference chain.
	//
	// Important: some schemas are constructed during processing (e.g. oneOf/allOf merge paths)
	// and may not have a pre-populated resolved schema at this point. Using MustGetResolvedSchema
	// matches the rest of the naming flow and ensures the libopenapi bug emulation also applies
	// to those constructed schemas.
	resolvedSchema := s.MustGetResolvedSchema()
	chain := resolvedSchema.GetReferenceChain()

	resolveToFirstAliasName := func(ref string) string {
		if ref == "" {
			return ""
		}
		if firstAliasRef := registry.GetOriginalRef(ref); firstAliasRef != "" {
			firstAliasName, _ := GetRefName(references.Reference(firstAliasRef))
			return firstAliasName
		}
		return ""
	}

	// Some schemas are constructed during processing (not purely resolved from the original document)
	// and can have an empty reference chain. In libopenapi's buggy behavior, the presence of the
	// registry entry is enough to force the first-alias name.
	if len(chain) == 0 {
		if firstAliasName := resolveToFirstAliasName(currentRef); firstAliasName != "" {
			return firstAliasName
		}
		return currentName
	}

	if len(chain) > 1 {
		// This is an alias with nested references (e.g., Schema2 -> SchemaShared)
		// Check if the FINAL ref in chain (the target) is tracked in registry
		finalRef := string(chain[len(chain)-1].Reference)
		if firstAliasName := resolveToFirstAliasName(finalRef); firstAliasName != "" {
			return firstAliasName
		}
	} else if len(chain) == 1 {
		// This is direct access to a target (e.g., ConfigurationManifest).
		//
		// IMPORTANT: The libopenapi bug behavior depends on the NUMBER of aliases:
		// - If the target has 2+ aliases pointing to it → direct access transforms to first alias
		// - If the target has only 1 alias → direct access stays as target name
		//
		// This is because libopenapi's internal indexing merges the target with the first alias
		// only when there are multiple aliases creating a "collision" in its indexing.
		if registry.HasMultipleAliases(currentRef) {
			if firstAliasName := resolveToFirstAliasName(currentRef); firstAliasName != "" {
				return firstAliasName
			}
		}
	}

	return currentName
}
