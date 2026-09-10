package openapi

import (
	"context"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
)

// NestedReferenceRegistry tracks the first reference that discovered each shared nested schema.
//
// This registry is populated by walking ALL schemas in the document (not just component schemas).
// When we ENCOUNTER a $ref during the walk and resolve it, if the target is itself a $ref
// (creating a nested reference chain), we record the mapping.
//
// Key insight: We only record schemas that are FOUND via $ref somewhere in the document.
// If a component schema is a $ref but nothing references it, it won't be recorded.
//
// Example 1 - Nested references via paths:
//
//	paths:
//	  /get:
//	    get:
//	      responses:
//	        200:
//	          schema:
//	            $ref: "#/components/schemas/Schema1"  <- We FIND Schema1 via $ref
//	  /post:
//	    post:
//	      responses:
//	        200:
//	          schema:
//	            $ref: "#/components/schemas/Schema2"  <- We FIND Schema2 via $ref
//	components:
//	  schemas:
//	    Schema1:
//	      $ref: "#/components/schemas/SchemaShared"  <- Schema1 IS a $ref (nested)
//	    Schema2:
//	      $ref: "#/components/schemas/SchemaShared"  <- Schema2 IS a $ref (nested)
//	    SchemaShared:
//	      type: object
//
// When we walk and encounter $ref to Schema1, we resolve it and see:
// - Chain: Schema1 -> SchemaShared (nested!)
// - Record: SchemaShared -> Schema1
//
// When we encounter $ref to Schema2, we resolve it and see:
// - Chain: Schema2 -> SchemaShared (nested!)
// - SchemaShared already recorded with Schema1 (first wins)
//
// Example 2 - Direct access (no $ref):
//
//	components:
//	  schemas:
//	    TestSchema:
//	      $ref: "#/components/schemas/StatusEnum"
//	    StatusEnum:
//	      type: string
//	      enum: [active, inactive]
//
// If the code directly accesses TestSchema (not via $ref resolution),
// TestSchema is never FOUND via $ref, so nothing is recorded.
// When we process TestSchema, it should resolve to "StatusEnum".
type NestedReferenceRegistry struct {
	mu sync.RWMutex
	// registry maps the final resolved ref (e.g., "#/components/schemas/SchemaShared")
	// to the first intermediate ref that led to it (e.g., "#/components/schemas/Schema1")
	registry map[string]string
	// aliasCount tracks how many pure $ref aliases point to each target.
	// This is important because the libopenapi bug behavior differs:
	// - 2+ aliases: direct target access transforms to first alias name
	// - 1 alias: direct target access stays as target name
	aliasCount map[string]int
}

// NewNestedReferenceRegistry creates a new NestedReferenceRegistry.
func NewNestedReferenceRegistry() *NestedReferenceRegistry {
	return &NestedReferenceRegistry{
		registry:   make(map[string]string),
		aliasCount: make(map[string]int),
	}
}

// Track records that a shared reference was first discovered via a particular intermediate reference.
// If the shared reference has already been tracked for first alias, the first wins.
// However, alias count is always incremented.
//
// Parameters:
//   - sharedRef: The final resolved reference (e.g., "#/components/schemas/SchemaShared")
//   - intermediateRef: The intermediate reference in the chain (e.g., "#/components/schemas/Schema1")
//
// Returns true if this is the first time tracking this shared reference, false if it was already tracked.
func (r *NestedReferenceRegistry) Track(sharedRef, intermediateRef string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Always increment alias count
	r.aliasCount[sharedRef]++

	if _, exists := r.registry[sharedRef]; exists {
		return false // Already tracked for first alias, first wins
	}

	r.registry[sharedRef] = intermediateRef
	return true
}

// GetOriginalRef returns the first intermediate reference that discovered the given shared reference.
// Returns empty string if the shared reference has not been tracked.
func (r *NestedReferenceRegistry) GetOriginalRef(sharedRef string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry[sharedRef]
}

// IsTracked returns true if the given shared reference has been tracked.
func (r *NestedReferenceRegistry) IsTracked(sharedRef string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.registry[sharedRef]
	return exists
}

// GetAliasCount returns the number of aliases that point to the given target reference.
// Returns 0 if the target has no aliases tracked.
func (r *NestedReferenceRegistry) GetAliasCount(targetRef string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.aliasCount[targetRef]
}

// HasMultipleAliases returns true if the target has 2 or more aliases pointing to it.
// This is important for the libopenapi bug behavior:
// - 2+ aliases: direct target access transforms to first alias name
// - 1 alias: direct target access stays as target name
func (r *NestedReferenceRegistry) HasMultipleAliases(targetRef string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.aliasCount[targetRef] >= 2
}

// Clear removes all tracked references. This is useful for resetting state between documents.
func (r *NestedReferenceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry = make(map[string]string)
	r.aliasCount = make(map[string]int)
}

// Len returns the number of tracked shared references.
func (r *NestedReferenceRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.registry)
}

// All returns a copy of all tracked references.
// The returned map is safe to modify without affecting the registry.
func (r *NestedReferenceRegistry) All() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]string, len(r.registry))
	for k, v := range r.registry {
		result[k] = v
	}
	return result
}

// Global registry instance.
var globalNestedRefRegistry = NewNestedReferenceRegistry()

// GlobalNestedRefRegistry returns the global nested reference registry.
func GlobalNestedRefRegistry() *NestedReferenceRegistry {
	return globalNestedRefRegistry
}

// ResetGlobalNestedRefRegistry clears the global registry.
// This should be called between processing different OpenAPI documents.
func ResetGlobalNestedRefRegistry() {
	globalNestedRefRegistry.Clear()
}

// PopulateFromDocument walks through the OpenAPI document's component schemas
// and identifies PURE ALIAS schemas (schemas that are just a $ref to another schema).
//
// The libopenapi bug specifically affects pure alias schemas like:
//
//	Schema1:
//	  $ref: "#/components/schemas/SchemaShared"  # Schema1 IS a pure $ref alias
//	Schema2:
//	  $ref: "#/components/schemas/SchemaShared"  # Schema2 IS a pure $ref alias
//
// This is DIFFERENT from schemas that CONTAIN $refs in their properties:
//
//	WebhookRequestCreated:
//	  type: object       # This IS an object, NOT a pure $ref alias
//	  properties:
//	    data:
//	      $ref: ...      # Contains a $ref, but the schema itself is NOT an alias
//
// Only pure alias schemas are tracked in the registry.
func PopulateFromDocument(ctx context.Context, docInfo *document.DocumentInfo) {
	if docInfo == nil || docInfo.Doc == nil {
		return
	}

	doc := docInfo.Doc
	registry := GlobalNestedRefRegistry()

	// Only look at component schemas - that's where the alias schemas live
	if doc.Components == nil || doc.Components.Schemas == nil {
		return
	}

	// First pass: identify which component schemas are pure $ref aliases
	// and what they point to
	aliasMap := make(map[string]string) // aliasRef -> targetRef

	for schemaName, schemaRef := range doc.Components.Schemas.All() {
		// Check if this component schema IS a $ref (a pure alias)
		if !schemaRef.IsReference() {
			continue
		}

		// This schema is a pure alias - it's just a $ref to another schema
		aliasRefStr := "#/components/schemas/" + schemaName
		targetRefStr := string(schemaRef.GetRef())

		// Only track if target is also a component schema reference
		if !isComponentSchemaRef(targetRefStr) {
			continue
		}

		aliasMap[aliasRefStr] = targetRefStr
	}

	// Second pass: walk through all schemas to find which alias schemas are actually
	// referenced via $ref somewhere in the document (paths, webhooks, etc.)
	// The first alias schema we encounter that references a given target wins.
	for item := range openapi.Walk(ctx, doc) {
		_ = item.Match(openapi.Matcher{
			Schema: func(schema *oas3.JSONSchema[oas3.Referenceable]) error {
				if !schema.IsReference() {
					return nil
				}

				currentRef := string(schema.GetRef())

				// Check if this $ref points to an alias schema
				targetOfAlias, isAlias := aliasMap[currentRef]
				if !isAlias {
					return nil
				}

				// This $ref points to an alias schema.
				// Track the alias's target with the alias as the first referencer.
				// Track recursively in case of multi-level aliasing (A -> B -> C)
				for {
					registry.Track(targetOfAlias, currentRef)

					// Check if the target is also an alias (multi-level)
					nextTarget, isNextAlias := aliasMap[targetOfAlias]
					if !isNextAlias {
						break
					}
					targetOfAlias = nextTarget
				}

				return nil
			},
		})
	}
}

// isComponentSchemaRef checks if a reference string points to a component schema.
func isComponentSchemaRef(ref string) bool {
	return len(ref) > 0 && (ref[0] == '#' || !containsExternalRef(ref)) &&
		containsComponentSchemaPath(ref)
}

func containsExternalRef(ref string) bool {
	// External refs contain a file path before the #
	for i := 0; i < len(ref); i++ {
		if ref[i] == '#' {
			return i > 0 // Has content before #
		}
		if ref[i] == '/' || ref[i] == '.' {
			return true // Likely a file path
		}
	}
	return false
}

func containsComponentSchemaPath(ref string) bool {
	const prefix = "#/components/schemas/"
	return len(ref) >= len(prefix) && ref[:len(prefix)] == prefix
}
