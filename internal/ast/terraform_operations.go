package ast

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

// Describes an ordered collection of Terraform operations for a single
// operation type (e.g. all create operations for a managed resource).
type TerraformOperations []*TerraformOperation

// SDKRequestMethods builds a deduplicated, sorted list of request SDK
// methods (To_ prefix) from the operations. Methods are deduplicated by
// MethodName (first operation wins) and sorted alphabetically.
func (ops TerraformOperations) SDKRequestMethods() []*TerraformSDKMethod {
	return ops.deduplicatedSDKMethods(
		func(op *TerraformOperation) []TerraformSDKMethodTarget { return op.RequestSDKMethodTargets },
		TerraformSDKMethodTarget.SDKRequestMethod,
	)
}

// SDKResponseMethods builds a deduplicated, sorted list of response SDK
// methods (RefreshFrom_ or RefreshFromArrayOf_ prefix) from the operations.
// Methods are deduplicated by MethodName (first operation wins) and sorted
// alphabetically.
//
// The receiver should be DataModelRefreshOperations() which lists Read
// operations first for pagination-aware RefreshFrom method priority.
func (ops TerraformOperations) SDKResponseMethods() []*TerraformSDKMethod {
	return ops.deduplicatedSDKMethods(
		func(op *TerraformOperation) []TerraformSDKMethodTarget { return op.ResponseSDKMethodTargets },
		TerraformSDKMethodTarget.SDKResponseMethod,
	)
}

// deduplicatedSDKMethods collects SDK methods from targets across all
// operations, deduplicating by MethodName (first wins) and sorting
// alphabetically.
func (ops TerraformOperations) deduplicatedSDKMethods(
	getTargets func(*TerraformOperation) []TerraformSDKMethodTarget,
	buildMethod func(TerraformSDKMethodTarget) *TerraformSDKMethod,
) []*TerraformSDKMethod {
	seen := make(map[string]bool)
	var methods []*TerraformSDKMethod

	for _, op := range ops {
		for _, target := range getTargets(op) {
			method := buildMethod(target)

			if seen[method.MethodName] {
				continue
			}

			seen[method.MethodName] = true

			methods = append(methods, method)
		}
	}

	slices.SortFunc(methods, func(a, b *TerraformSDKMethod) int {
		return strings.Compare(a.MethodName, b.MethodName)
	})

	return methods
}

// Shards computes and returns the merged request shard, response shard, and
// combined shard from all operations. Any of the returned values may be nil if
// no operations have the corresponding shard.
//
// The merged shards are computed from scratch each time this method is called,
// reflecting the current state of each operation's individual RequestShard and
// ResponseShard. This allows callers to re-compute after mutating individual
// operation shards (e.g. applying tagging rules).
//
// Alias node merging into SchemaTypeDef is handled separately by each
// resource type's AssembleSchemaTypeDef since the timing and shard sources
// vary per resource type.
func (ops TerraformOperations) Shards() (requestShard, responseShard, shard *TypeDef, err error) {
	for _, op := range ops {
		if op.RequestShard == nil {
			continue
		}

		if requestShard == nil {
			requestShard = op.RequestShard.Clone()
			continue
		}

		if err := requestShard.TerraformMerge(op.RequestShard); err != nil {
			return nil, nil, nil, fmt.Errorf("merging request shard: %w", err)
		}
	}

	for _, op := range ops {
		if op.ResponseShard == nil {
			continue
		}

		if responseShard == nil {
			responseShard = op.ResponseShard.Clone()
			continue
		}

		if err := responseShard.TerraformMerge(op.ResponseShard); err != nil {
			return nil, nil, nil, fmt.Errorf("merging response shard: %w", err)
		}
	}

	switch {
	case requestShard != nil && responseShard != nil:
		shard = requestShard.Clone()

		if err := shard.TerraformMerge(responseShard.Clone()); err != nil {
			return nil, nil, nil, fmt.Errorf("merging combined shard: %w", err)
		}
	case requestShard != nil:
		shard = requestShard.Clone()
	case responseShard != nil:
		shard = responseShard.Clone()
	}

	return requestShard, responseShard, shard, nil
}

// Describes a Terraform operation, which is an API operation that has been
// mapped to a Terraform resource operation. This accounts for configuration,
// such as the x-speakeasy-entity extension, that modifies the operation
// request and response data in preparation for merging all operation data
// into a single Terraform resource schema.
type TerraformOperation struct {
	// Original API operation. Should always remain untouched.
	APIOperation *Operation

	// EntityMissingCodes holds the HTTP status codes that indicate the entity
	// is not found by the API. Set on read and delete operations for managed
	// resources. The semantics vary by operation type: read operations remove
	// the resource from state (soft delete detection), while delete operations
	// treat these codes as success (the resource is already gone).
	EntityMissingCodes []int

	// Name of the entity this operation is associated with.
	EntityName string

	// String representation from the entity operation configuration in the form
	// of Entity#Op[#Order], such as Thing#create.
	EntityOperation string

	// Has409ConflictCode indicates whether the operation's API response spec
	// defines a 409 Conflict response code. When true, create operations
	// produce "resource already exists" error handling.
	Has409ConflictCode bool

	// IncludeOperationSecurity indicates that this operation has API-level
	// security defined and the entity's enableOperationSecurity generation
	// config flag is enabled. When true, the Terraform resource method
	// invocation includes an operation security variable.
	IncludeOperationSecurity bool

	// Options associated with the entity operation configuration.
	Options *extensions.EntityOperationV1Options

	// Operation request shard. Unset if there is no relevant request.
	//
	// A shard represents the portion of the data that is most relevant to the
	// Terraform resource. For example, the x-speakeasy-entity extension
	// is used to indicate which data represents the top level of the Terraform
	// resource schema, potentially bypassing API-defined intermediate
	// properties such as "data" that are undesirable in Terraform
	// configurations. Resource logic is generated to handle the translation
	// between the API-defined schema and the Terraform resource schema.
	RequestShard *TypeDef

	// RequestBodySDKMethod is the SDK method for the operation's root request
	// type. Matches the target in RequestSDKMethodTargets whose TypeDef
	// equals the request field type. Nil when no matching target exists
	// (e.g. non-class/union request types).
	RequestBodySDKMethod *TerraformSDKMethod

	// RequestSDKMethodTargets holds all SDK TypeDefs that need model-to-SDK
	// (To_) conversion methods for this operation's request. Includes
	// TypeDefs on the path from the request type to the entity TypeDef,
	// plus the root request type itself.
	RequestSDKMethodTargets []TerraformSDKMethodTarget

	// ResponseBodyFieldDef is the response body FieldDef for this operation,
	// resolved by calling TerraformBodyFieldDef(entityName) on the API
	// operation response. Populated by setResponseShard during operation
	// construction. Nil when the response has no body field relevant to the
	// entity or when skipDataModelRefresh is true (cleared by setSDKMethodData).
	// A nil check on this field alone is sufficient to determine whether a
	// data model refresh applies.
	ResponseBodyFieldDef *FieldDef

	// SupportsPagination indicates this operation has a Pagination extension
	// and the response body represents a direct entity (not an array
	// extraction). When true, the Terraform resource method includes
	// pagination loop and pre-refresh field reinitialization logic.
	SupportsPagination bool

	// responseBodyHasEntityArray indicates that the response body
	// contains a nested array field with entity-matching items, but the
	// Terraform schema does not include that array field (it was elided
	// during schema assembly). The entity represents a single item from
	// this array. Pagination is not applicable for this operation.
	responseBodyHasEntityArray bool

	// responseBodyIsEntityArray indicates that the response body field
	// type is directly an array or set whose items match the entity. The
	// entity represents a single item from this array, so response
	// mapping must account for the array-to-entity extraction. Pagination
	// is not applicable for this operation.
	responseBodyIsEntityArray bool

	// ResponseBodySDKMethod is the SDK method for the operation's response
	// body type. Matches the target in ResponseSDKMethodTargets whose TypeDef
	// equals ResponseBodyFieldDef.Type. Nil when no matching target or body
	// field exists.
	ResponseBodySDKMethod *TerraformSDKMethod

	// ResponseSDKMethodTargets holds all SDK TypeDefs that need SDK-to-model
	// (RefreshFrom_) conversion methods for this operation's response.
	// Includes the response body type, TypeDefs on the path from the response
	// body to the entity TypeDef, and the entity TypeDef from the full
	// response type. For entity array responses, includes the array wrapper
	// (IsArrayWrapper=true) and item type targets.
	ResponseSDKMethodTargets []TerraformSDKMethodTarget

	// Operation response shard. Unset if there is no relevant response.
	//
	// A shard represents the portion of the data that is most relevant to the
	// Terraform resource. For example, the x-speakeasy-entity extension
	// is used to indicate which data represents the top level of the Terraform
	// resource schema, potentially bypassing API-defined intermediate
	// properties such as "data" that are undesirable in Terraform
	// configurations. Resource logic is generated to handle the translation
	// between the API-defined schema and the Terraform resource schema.
	ResponseShard *TypeDef

	// ServerAttributeName is the Terraform schema attribute name used to look
	// up the per-operation server URL from state. Empty string indicates no
	// server URL support for this operation. The value is derived from
	// entity-level Server configuration during assembly.
	ServerAttributeName string

	// skipDataModelRefresh indicates that this operation's API response should
	// not be mapped back into the Terraform data model during resource method
	// execution. Set during MergeOperationShards for operations where
	// refreshing the data model is not meaningful, such as the final delete
	// operation (resource is being destroyed) or the final invoke operation
	// (result is not persisted to state). When true, setSDKMethodData nils
	// out ResponseBodyFieldDef and ResponseBodySDKMethod so that a nil check
	// on those fields alone determines whether a refresh applies.
	skipDataModelRefresh bool

	// SuccessCodes holds the unique 2xx HTTP status codes defined in the
	// operation's API response spec, used for response validation in
	// Terraform resource methods. Codes containing "X" (e.g. "2XX") are
	// included as-is and matched with a prefix check.
	SuccessCodes []string
}

// Creates a new TerraformOperation.
func NewTerraformOperation(apiOperation *Operation, entityOperation extensions.EntityOperationV1Config) (*TerraformOperation, error) {
	result := &TerraformOperation{
		APIOperation:    apiOperation,
		EntityName:      entityOperation.Entity,
		EntityOperation: entityOperation.String(),
		Options:         entityOperation.Options,
	}

	result.setResponseCodes()

	if err := result.setRequestShard(); err != nil {
		return result, err
	}

	if err := result.setResponseShard(); err != nil {
		return result, err
	}

	return result, nil
}

// CleanupShards removes ignored fields from the operation's request and
// response shards. This should be called after schema assembly is complete,
// as the shards are used during assembly but need cleanup for code generation.
func (o *TerraformOperation) CleanupShards() {
	if o == nil {
		return
	}

	if o.RequestShard != nil {
		o.RequestShard.TerraformDeleteIgnored()
	}

	if o.ResponseShard != nil {
		o.ResponseShard.TerraformDeleteIgnored()
	}
}

// Returns a deep clone of the TerraformOperation. Note: SDK method targets
// and pre-computed body SDK methods retain Operation pointers to the original,
// not the clone. This is safe because Clone is called before setSDKMethodData
// during entity assembly, so these fields are empty at clone time.
func (o *TerraformOperation) Clone() *TerraformOperation {
	if o == nil {
		return nil
	}

	result := &TerraformOperation{
		APIOperation:               o.APIOperation,
		EntityMissingCodes:         o.EntityMissingCodes,
		EntityName:                 o.EntityName,
		EntityOperation:            o.EntityOperation,
		Has409ConflictCode:         o.Has409ConflictCode,
		IncludeOperationSecurity:   o.IncludeOperationSecurity,
		Options:                    o.Options,
		RequestBodySDKMethod:       o.RequestBodySDKMethod,
		ResponseBodyFieldDef:       o.ResponseBodyFieldDef,
		responseBodyHasEntityArray: o.responseBodyHasEntityArray,
		responseBodyIsEntityArray:  o.responseBodyIsEntityArray,
		ResponseBodySDKMethod:      o.ResponseBodySDKMethod,
		SupportsPagination:         o.SupportsPagination,
		ServerAttributeName:        o.ServerAttributeName,
		skipDataModelRefresh:       o.skipDataModelRefresh,
		SuccessCodes:               o.SuccessCodes,
	}

	if o.RequestShard != nil {
		result.RequestShard = o.RequestShard.Clone()
	}

	if o.ResponseShard != nil {
		result.ResponseShard = o.ResponseShard.Clone()
	}

	if o.RequestSDKMethodTargets != nil {
		result.RequestSDKMethodTargets = slices.Clone(o.RequestSDKMethodTargets)
	}

	if o.ResponseSDKMethodTargets != nil {
		result.ResponseSDKMethodTargets = slices.Clone(o.ResponseSDKMethodTargets)
	}

	return result
}

// HasPatchStyle returns true if the operation has a Patch.Style configuration
// set. This is one of the conditions that enables IncludeSDKMethodOptions on a
// Terraform entity.
func (o *TerraformOperation) HasPatchStyle() bool {
	return o != nil && o.Options != nil && o.Options.Patch != nil && o.Options.Patch.Style != ""
}

// HasTerraformWriteOnly returns true if the operation's API request body
// contains any TypeDef with the TerraformWriteOnly extension set. This is
// checked for create/update operations as one of the conditions that enables
// IncludeSDKMethodOptions.
func (o *TerraformOperation) HasTerraformWriteOnly() bool {
	if o == nil || o.APIOperation == nil || o.APIOperation.Request == nil || o.APIOperation.Request.RequestBody == nil {
		return false
	}

	requestBodyType := o.APIOperation.Request.RequestBody.Type
	if requestBodyType == nil {
		return false
	}

	found := false

	requestBodyType.TerraformWalkTypes("", func(_ string, typedef *TypeDef, _ bool) {
		if found {
			return
		}

		if typedef != nil && typedef.Extensions != nil &&
			typedef.Extensions.TerraformWriteOnly != nil && *typedef.Extensions.TerraformWriteOnly {
			found = true
		}
	}, false)

	return found
}

// HasUsePriorStateParameters returns true if the operation's API request has
// any parameters (path, query, or header) with the MatchConfig.UsePriorState
// flag set. This is checked for update operations as one of the conditions
// that enables IncludeSDKMethodOptions.
func (o *TerraformOperation) HasUsePriorStateParameters() bool {
	if o == nil || o.APIOperation == nil || o.APIOperation.Request == nil || o.APIOperation.Request.Params == nil {
		return false
	}

	params := o.APIOperation.Request.Params

	for _, param := range slices.Concat(params.PathParams, params.QueryParams, params.HeaderParams) {
		if param.HasMatchConfigUsePriorState() {
			return true
		}
	}

	return false
}

// setIncludeOperationSecurity sets IncludeOperationSecurity to true if the
// operation has API-level security defined and the enableOperationSecurity
// generation config flag is enabled.
func (o *TerraformOperation) setIncludeOperationSecurity(generationConfig map[string]any) {
	if o == nil || o.APIOperation == nil || o.APIOperation.Security == nil {
		return
	}

	enableOperationSecurity, ok := generationConfig["enableOperationSecurity"].(bool)

	if !ok || !enableOperationSecurity {
		return
	}

	o.IncludeOperationSecurity = true
}

// Sets the request shard for the TerraformOperation based on the entity
// configuration.
func (o *TerraformOperation) setRequestShard() error {
	if o.APIOperation.Request == nil || o.APIOperation.Request.Field == nil || o.APIOperation.Request.Field.Type == nil {
		return nil
	}

	// TODO: Clone entity TypeDef before tagging. Currently, this mutates the
	// original TypeDef in APIOperation, which is not desirable.
	entityTypeDef := o.APIOperation.Request.Field.Type.FindEntityTypeDef(o.EntityName)
	if entityTypeDef != nil {
		entityTypeDef.SetExtensionRecursive("x-speakeasy-terraform-in-entity", true)
	}
	o.APIOperation.Request.Field.Type.SetExtensionIfAbsentRecursive("x-speakeasy-terraform-in-entity", false)

	requestShard := o.APIOperation.Request.Field.Type.Clone()

	if err := requestShard.TerraformHoistByEntityName(o.EntityName); err != nil {
		return err
	}

	o.RequestShard = requestShard

	return nil
}

// setResponseBodyFields computes responseBodyIsEntityArray and
// responseBodyHasEntityArray from the already-populated
// ResponseBodyFieldDef and the assembled entity schema. Called internally
// by setSDKMethodData before response target computation.
func (o *TerraformOperation) setResponseBodyFields(entityName string, schemaTypeDef *TypeDef) {
	if o == nil || o.ResponseBodyFieldDef == nil || o.ResponseBodyFieldDef.Type == nil {
		return
	}

	fieldType := o.ResponseBodyFieldDef.Type

	// Direct array: the response body field type itself is an array whose
	// items match the entity. Example: API returns [{"id": "1"}, ...] and
	// the entity represents one item from the array.
	if (fieldType.Type == DataTypeArray || fieldType.Type == DataTypeSet) &&
		fieldType.ItemType != nil &&
		fieldType.ItemType.FindEntityTypeDef(entityName) != nil {
		o.responseBodyIsEntityArray = true

		return
	}

	// Nested hidden array: the response body contains a nested array field
	// with entity-matching items, but the Terraform schema does not include
	// that array field. Example: API returns {"data": [{"id": "1"}, ...]}
	// where the schema exposes "id" directly, not the "data" array.
	for _, field := range fieldType.Fields {
		if field.Type == nil {
			continue
		}

		if field.Type.Type != DataTypeArray && field.Type.Type != DataTypeSet {
			continue
		}

		if field.Type.ItemType == nil || field.Type.ItemType.FindEntityTypeDef(entityName) == nil {
			continue
		}

		// Found a response array field with entity items. If the entity
		// schema has a matching array field, the array is exposed in the
		// schema and this is not a hidden array case.
		if schemaTypeDef != nil {
			hasMatchingSchemaField := false

			for _, schemaField := range schemaTypeDef.Fields {
				if schemaField.Type != nil &&
					(schemaField.Type.Type == DataTypeArray || schemaField.Type.Type == DataTypeSet) &&
					schemaField.Name == field.Name {
					hasMatchingSchemaField = true

					break
				}
			}

			if hasMatchingSchemaField {
				continue
			}
		}

		o.responseBodyHasEntityArray = true

		return
	}
}

// setResponseCodes computes and sets SuccessCodes and Has409ConflictCode from
// the API operation's response spec.
func (o *TerraformOperation) setResponseCodes() {
	if o.APIOperation == nil || o.APIOperation.Response == nil {
		return
	}

	seen := make(map[string]bool)

	for _, resp := range o.APIOperation.Response.Responses {
		for _, code := range resp.Code {
			if seen[code] {
				continue
			}

			seen[code] = true

			if strings.HasPrefix(code, "2") {
				o.SuccessCodes = append(o.SuccessCodes, code)
			}

			if code == "409" {
				o.Has409ConflictCode = true
			}
		}
	}
}

// setSDKMethodData computes response body fields, request/response SDK
// method targets, and pre-computed body SDK methods. It calls
// setResponseBodyFields internally so callers do not need to worry about
// ordering. The entityName identifies the entity for array detection, and
// schemaTypeDef is the entity's assembled SchemaTypeDef used for
// compatibility filtering and array field cross-referencing.
//
// Returns an error if the operation has a request body but no matching SDK
// method target can be found (indicating an assembly or configuration bug).
func (o *TerraformOperation) setSDKMethodData(entityName string, schemaTypeDef *TypeDef) error {
	o.setResponseBodyFields(entityName, schemaTypeDef)
	o.setRequestSDKMethodTargets()
	o.setResponseSDKMethodTargets(schemaTypeDef)
	o.setRequestBodySDKMethod()
	o.setResponseBodySDKMethod()

	// Validate that every operation with a request body has a computed
	// RequestBodySDKMethod. This catches assembly issues early rather than
	// surfacing them as unexpected nil values at render time.
	if o.APIOperation != nil && o.APIOperation.Request != nil &&
		o.APIOperation.Request.Field != nil && o.APIOperation.Request.Field.Type != nil &&
		o.RequestBodySDKMethod == nil {
		return fmt.Errorf("no request body SDK method target for operation %s", o.APIOperation.ID)
	}

	o.SupportsPagination = o.APIOperation != nil &&
		o.APIOperation.Extensions != nil &&
		o.APIOperation.Extensions.Pagination != nil &&
		!o.responseBodyIsEntityArray &&
		!o.responseBodyHasEntityArray

	// When skipDataModelRefresh is set, the response should not be mapped
	// back into the data model. Nil out the response body fields so that
	// a nil check on ResponseBodyFieldDef alone determines whether a
	// refresh applies.
	if o.skipDataModelRefresh {
		o.ResponseBodyFieldDef = nil
		o.ResponseBodySDKMethod = nil
	}

	return nil
}

// setRequestSDKMethodTargets computes RequestSDKMethodTargets from the API
// operation request type. Targets include all TypeDefs on the path from the
// request type to the entity TypeDef, plus the root request type itself.
func (o *TerraformOperation) setRequestSDKMethodTargets() {
	if o.APIOperation == nil || o.APIOperation.Request == nil ||
		o.APIOperation.Request.Field == nil || o.APIOperation.Request.Field.Type == nil {
		return
	}

	requestType := o.APIOperation.Request.Field.Type

	// Collect class/union TypeDefs on the path from the request type to the
	// entity. FindEntitySDKMethodTargets traverses through array ItemTypes,
	// so this must run even when requestType itself is not class/union.
	// Non-class/union targets are skipped because only class-like TypeDefs
	// have meaningful SDK conversion methods.
	for _, target := range requestType.FindEntitySDKMethodTargets(o.EntityName, true) {
		if target.TypeDef.Type != DataTypeClass && target.TypeDef.Type != DataTypeUnion {
			continue
		}

		target.Operation = o
		o.RequestSDKMethodTargets = append(o.RequestSDKMethodTargets, target)
	}

	// Include the root request type as a target when it is a class or union.
	// Non-class/union root types (e.g. arrays, primitives) are not included
	// as their items are already handled by the path traversal above.
	if requestType.Type != DataTypeClass && requestType.Type != DataTypeUnion {
		return
	}

	o.RequestSDKMethodTargets = append(o.RequestSDKMethodTargets, TerraformSDKMethodTarget{
		TypeDef:   requestType,
		Operation: o,
		Optional:  o.APIOperation.Request.Field.Optional || o.APIOperation.Request.Field.Nullable,
	})
}

// setResponseSDKMethodTargets computes ResponseSDKMethodTargets from the API
// operation response type. Targets include the response body type, TypeDefs on
// the path from the response body to the entity, and the entity TypeDef from
// the full response type. For entity array responses, includes array wrapper
// and item type targets.
//
// Targets are filtered by isTerraformDataModelCompatible to exclude TypeDefs
// where the SDK-to-model conversion would produce no output. The schemaTypeDef
// is the entity's SchemaTypeDef used for compatibility checking.
func (o *TerraformOperation) setResponseSDKMethodTargets(schemaTypeDef *TypeDef) {
	// Cache compatibility results keyed by sdkType pointer. The schemaTypeDef
	// and entityName are constant across all calls within this method, so the
	// sdkType pointer is the only varying input. isTerraformDataModelCompatible
	// is read-only and deterministic, making pointer-keyed caching safe. This
	// avoids redundant recursive tree traversals when the same TypeDef appears
	// in multiple path entries (common with shared $ref types).
	compatibleCache := make(map[*TypeDef]bool)
	isCompatible := func(sdkType *TypeDef) bool {
		if result, ok := compatibleCache[sdkType]; ok {
			return result
		}

		result := schemaTypeDef.isTerraformDataModelCompatible(o.EntityName, sdkType)
		compatibleCache[sdkType] = result

		return result
	}

	// Response body targets.
	if o.ResponseBodyFieldDef != nil && o.ResponseBodyFieldDef.Type != nil {
		bodyType := o.ResponseBodyFieldDef.Type

		if o.responseBodyIsEntityArray {
			// Array wrapper target with RefreshFromArrayOf_ prefix.
			if isCompatible(bodyType) {
				o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, TerraformSDKMethodTarget{
					TypeDef:        bodyType,
					Operation:      o,
					Optional:       o.ResponseBodyFieldDef.Optional,
					IsArrayWrapper: true,
				})
			}

			if bodyType.ItemType != nil {
				// Item type target.
				if isCompatible(bodyType.ItemType) {
					o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, TerraformSDKMethodTarget{
						TypeDef:   bodyType.ItemType,
						Operation: o,
						Optional:  true,
					})
				}

				// Path entries from item type to entity.
				path := bodyType.ItemType.FindEntitySDKMethodTargets(o.EntityName, true)

				for _, target := range path {
					if isCompatible(target.TypeDef) {
						target.Operation = o
						o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, target)
					}
				}
			}
		} else {
			// Root response body type target.
			if isCompatible(bodyType) {
				o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, TerraformSDKMethodTarget{
					TypeDef:   bodyType,
					Operation: o,
					Optional:  o.ResponseBodyFieldDef.Optional,
				})
			}

			// Path entries from body type to entity.
			path := bodyType.FindEntitySDKMethodTargets(o.EntityName, true)

			for _, target := range path {
				if isCompatible(target.TypeDef) {
					target.Operation = o
					o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, target)
				}
			}
		}
	}

	// Entity TypeDef from full response type. This catch-all ensures the
	// entity TypeDef always gets a RefreshFrom_ method even if the response
	// body traversal above already found it (deduplication by method name
	// during SDKResponseMethods handles overlap).
	if o.APIOperation != nil && o.APIOperation.Response != nil && o.APIOperation.Response.Type != nil {
		path := o.APIOperation.Response.Type.FindEntitySDKMethodTargets(o.EntityName, true)

		if len(path) > 0 && isCompatible(path[0].TypeDef) {
			// Only the entity TypeDef itself (first entry in path).
			target := path[0]
			target.Operation = o
			o.ResponseSDKMethodTargets = append(o.ResponseSDKMethodTargets, target)
		}
	}
}

// setRequestBodySDKMethod finds the target in RequestSDKMethodTargets whose
// TypeDef matches the root request field type and pre-computes its SDK method.
// For array/set request types, the root type itself is not in the targets
// (only class/union types are added), so this falls back to matching the
// ItemType which represents the item-level To_ method.
func (o *TerraformOperation) setRequestBodySDKMethod() {
	if o.APIOperation == nil || o.APIOperation.Request == nil ||
		o.APIOperation.Request.Field == nil || o.APIOperation.Request.Field.Type == nil {
		return
	}

	requestType := o.APIOperation.Request.Field.Type

	for _, target := range o.RequestSDKMethodTargets {
		if target.TypeDef == requestType {
			o.RequestBodySDKMethod = target.SDKRequestMethod()

			return
		}
	}

	// For array/set request types, the item class is the relevant target.
	if requestType.ItemType != nil {
		for _, target := range o.RequestSDKMethodTargets {
			if target.TypeDef == requestType.ItemType {
				o.RequestBodySDKMethod = target.SDKRequestMethod()

				return
			}
		}
	}
}

// setResponseBodySDKMethod finds the target in ResponseSDKMethodTargets whose
// TypeDef matches the response body field type and pre-computes its SDK method.
func (o *TerraformOperation) setResponseBodySDKMethod() {
	if o.ResponseBodyFieldDef == nil || o.ResponseBodyFieldDef.Type == nil {
		return
	}

	if target := o.FindResponseSDKMethodTarget(o.ResponseBodyFieldDef.Type); target != nil {
		o.ResponseBodySDKMethod = target.SDKResponseMethod()
	}
}

// FindResponseSDKMethodTarget returns the target in ResponseSDKMethodTargets
// whose TypeDef pointer matches td. Returns nil if no match is found,
// indicating the TypeDef has no compatible RefreshFrom method.
//
// The returned pointer references a slice element and is valid only while the
// TerraformOperation remains alive. Callers should use it immediately rather
// than storing it for later access.
func (o *TerraformOperation) FindResponseSDKMethodTarget(td *TypeDef) *TerraformSDKMethodTarget {
	for i := range o.ResponseSDKMethodTargets {
		if o.ResponseSDKMethodTargets[i].TypeDef == td {
			return &o.ResponseSDKMethodTargets[i]
		}
	}

	return nil
}

// setResponseShard resolves the response body FieldDef and derives the
// response shard from it. The FieldDef is stored as ResponseBodyFieldDef
// for later use by setResponseBodyFields and code generation.
func (o *TerraformOperation) setResponseShard() error {
	if o.APIOperation.Response == nil || o.APIOperation.Response.Type == nil {
		return nil
	}

	o.ResponseBodyFieldDef = o.APIOperation.Response.TerraformBodyFieldDef(o.EntityName)

	if o.ResponseBodyFieldDef == nil || o.ResponseBodyFieldDef.Type == nil {
		return nil
	}

	// TODO: Clone entity TypeDef before tagging. Currently, this mutates the
	// original TypeDef in APIOperation, which is not desirable.
	entityTypeDef := o.ResponseBodyFieldDef.Type.FindEntityTypeDef(o.EntityName)
	if entityTypeDef != nil {
		entityTypeDef.SetExtensionRecursive("x-speakeasy-terraform-in-entity", true)
	}
	o.ResponseBodyFieldDef.Type.SetExtensionIfAbsentRecursive("x-speakeasy-terraform-in-entity", false)

	responseShard := o.ResponseBodyFieldDef.Type.Clone()

	if err := responseShard.TerraformHoistByEntityName(o.EntityName); err != nil {
		return err
	}

	if err := responseShard.TerraformHoistAssociatedTypesCommonFields(nil); err != nil {
		return err
	}

	responseShard = responseShard.TerraformHandleWrappedAttribute(o.EntityName, o.EntityOperation)

	o.ResponseShard = responseShard

	return nil
}

// setServerAttributeName sets ServerAttributeName if the operation has
// API-level servers defined and the enableOperationServers generation config
// flag is enabled.
func (o *TerraformOperation) setServerAttributeName(generationConfig map[string]any) {
	if o == nil || o.APIOperation == nil || o.APIOperation.Servers == nil {
		return
	}

	enableOperationServers, ok := generationConfig["enableOperationServers"].(bool)

	if !ok || !enableOperationServers {
		return
	}

	// TODO: Support configurable server URL attribute renaming.
	// internal issue reference
	o.ServerAttributeName = "server_url"
}

// Returns a string representation of the TerraformOperation.
func (o TerraformOperation) String() string {
	return o.EntityOperation
}

// Returns the ordered operations from the given operations map.
func orderedTerraformOperations(ops map[int]*TerraformOperation) TerraformOperations {
	keys := slices.Sorted(maps.Keys(ops))

	result := make(TerraformOperations, 0, len(keys))

	for _, key := range keys {
		result = append(result, ops[key])
	}

	return result
}

// Returns a best effort sanitized pagination output property name, which has
// any leading "$." removed and then the remaining is sanitized as a field name.
// There may be use cases where the JSON Path contains further path information
// that would need to be removed, but this is not currently handled until there
// is a need.
func sanitizePaginationOutputName(paginationOutputProperty string) string {
	result := strings.ReplaceAll(paginationOutputProperty, "$.", "")

	return SanitizeFieldName(result)
}
