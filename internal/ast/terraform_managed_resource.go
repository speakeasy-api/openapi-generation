package ast

import (
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes a Terraform managed resource.
type TerraformManagedResource struct {
	// Description for the managed resource. Sourced from
	// x-speakeasy-entity-description configuration, if available.
	Description string `json:"description" yaml:"description"`

	// Mapping of global field names to definitions for all operations. These
	// are added to the entity resource struct type, copied in the Configure()
	// method, made optional in the resource schema, and checked in the resource
	// methods.
	GlobalFields map[string]*FieldDef `json:"globalFields" yaml:"globalFields"`

	// Computed Go type name for the managed resource data model struct, e.g.
	// "ExampleResourceModel". Set by NewTerraformManagedResource.
	GoDataModelTypeName string `json:"goDataModelTypeName" yaml:"goDataModelTypeName"`

	// Computed Go type name for the managed resource private data model struct,
	// e.g. "ExampleResourcePrivateDataModel". Set by NewTerraformManagedResource.
	GoPrivateDataModelTypeName string `json:"goPrivateDataModelTypeName" yaml:"goPrivateDataModelTypeName"`

	// Computed Go type name for the managed resource struct implementation, e.g.
	// "ExampleResource". Set by NewTerraformManagedResource.
	GoTypeName string `json:"goTypeName" yaml:"goTypeName"`

	// Import state TypeDef for the managed resource, derived from the read
	// request shard. Contains only required fields (as determined by
	// FieldDef.IsTerraformImportRequired), sorted alphabetically. Used by
	// Terraform import state template generation to determine the import ID
	// shape. Populated by AssembleSchemaTypeDef.
	ImportStateTypeDef *TypeDef `json:"-" yaml:"-"`

	// Whether the entity requires SDK method options in its generated code.
	// This is true when any operation has a Patch.Style configuration, any
	// update operation has UsePriorState parameters, or any create/update
	// operation has a TerraformWriteOnly extension. Populated by
	// AssembleSchemaTypeDef.
	IncludeSDKMethodOptions bool `json:"includeSDKMethodOptions,omitempty" yaml:"includeSDKMethodOptions,omitempty"`

	// List of HTTP status codes that represent the managed resource is not
	// found in the API during read operations. These codes automatically cause
	// the resource to be removed from the Terraform state rather than return an
	// API error.
	//
	// Defaults to 404, which has historically been used by many APIs to
	// represent this situation. 410 could potentially also be considered in the
	// future. Values can be customized via x-speakeasy-entity-missing-codes
	// extension configuration.
	MissingCodes []int `json:"missingCodes" yaml:"missingCodes"`

	// Name of the managed resource type, such as "examplecloud_thing".
	Name string `json:"name" yaml:"name"`

	// Operation security configuration for the managed resource. This is set
	// when the underlying operations have operation security defined and the
	// enableOperationSecurity generation configuration flag is enabled. Only
	// a single operation security configuration is supported per managed
	// resource, even if multiple operations have differing security
	// configurations. This is intentional to simplify the generated Terraform
	// code and avoid complexity around per-operation security configuration
	// for consumers.
	OperationSecurity *FieldDef `json:"operationSecurity" yaml:"operationSecurity"`

	// All operations associated with the managed resource.
	Operations *TerraformManagedResourceOperations `json:"operations" yaml:"operations"`

	// Mapping of pagination input field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationInputFields map[string]bool `json:"paginationInputFields" yaml:"paginationInputFields"`

	// Mapping of pagination output field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationOutputFields map[string]bool `json:"paginationOutputFields" yaml:"paginationOutputFields"`

	// Fields from the read response shard that have the
	// x-speakeasy-soft-delete-property extension set. These are used to
	// generate soft-delete handling in the resource Read method.
	// Populated by AssembleSchemaTypeDef.
	ReadSoftDeleteProperties Fields `json:"readSoftDeleteProperties,omitempty" yaml:"readSoftDeleteProperties,omitempty"`

	// Terraform schema for the managed resource.
	Schema *TerraformManagedResourceSchema `json:"schema" yaml:"schema"`

	// Merged Terraform resource schema TypeDef across all operations. This
	// represents the combined schema view of the resource after merging all
	// operation shards together.
	SchemaTypeDef *TypeDef `json:"-" yaml:"-"`

	// Warnings collected during schema assembly. These are non-fatal messages
	// (e.g. optional read request field ignoring, update-in-create suggestions)
	// that should be logged after assembly completes.
	SchemaWarnings []string `json:"schemaWarnings,omitempty" yaml:"schemaWarnings,omitempty"`

	// Deduplicated, sorted list of SDK request conversion methods (To_ prefix)
	// pre-computed from all operations. Populated by AssembleSchemaTypeDef.
	SDKRequestMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK response conversion methods
	// (RefreshFrom_ or RefreshFromArrayOf_ prefix) pre-computed from
	// DataModelRefreshOperations. Populated by AssembleSchemaTypeDef.
	SDKResponseMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Server configuration for the managed resource. This is defined when
	// the underlying operations have path or operation server URLs defined.
	// Only a single server configuration is supported per managed resource,
	// even if multiple operations have differing server URLs defined. This is
	// intentional to simplify the generated Terraform code and avoid complexity
	// around per-operation server configuration for consumers.
	Server *TerraformServer `json:"server" yaml:"server"`

	// Resolved full Terraform managed resource type name (e.g.
	// "myprovider_my_resource"), composed from the resolved provider type
	// name and the snake_case form of Name. Used as the value of
	// resp.TypeName in the resource Metadata implementation, in example
	// Terraform configuration files, and in import command examples.
	// Populated by TerraformProvider.AddOrGetManagedResource.
	TerraformTypeName string `json:"terraformTypeName" yaml:"terraformTypeName"`
}

// Creates a new Terraform managed resource, safely initializing underlying fields.
func NewTerraformManagedResource(name string) *TerraformManagedResource {
	goTypeName := TerraformGoTypeName(name) + "Resource"

	return &TerraformManagedResource{
		GoDataModelTypeName:        goTypeName + "Model",
		GoPrivateDataModelTypeName: goTypeName + "PrivateDataModel",
		GoTypeName:                 goTypeName,
		MissingCodes:               []int{http.StatusNotFound},
		Name:                       name,
		Operations:                 NewTerraformManagedResourceOperations(),
		Schema:                     NewTerraformManagedResourceSchema(),
	}
}

// Adds an operation to the managed resource.
func (r *TerraformManagedResource) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, operation *Operation) error {
	if r.Operations == nil {
		r.Operations = NewTerraformManagedResourceOperations()
	}

	if entityOperationConfig.Entity != r.Name {
		return nil
	}

	if err := r.Operations.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
		return err
	}

	r.setDescriptionFromOperation(operation)
	r.setGlobalFieldsFromOperation(operation)
	r.setMissingCodesFromOperation(entityOperationConfig, operation)
	r.setOperationSecurityFromOperation(generationConfig, operation)
	r.setPaginationFieldsFromOperation(operation)
	r.setServerFromOperation(generationConfig, operation)

	if err := r.setSchemaVersionFromOperation(operation); err != nil {
		return err
	}

	return nil
}

// AssembleSchemaTypeDef builds the merged Terraform resource schema TypeDef
// from all operation shards. This includes merging individual operation shards,
// pre-processing, merging CREATE/READ/DELETE/UPDATE shards with validation,
// deep-cloning, and extension modifications for computed/readonly/force-new
// semantics.
func (r *TerraformManagedResource) AssembleSchemaTypeDef(excludeEmptyObjectSchemas bool) error {
	if r == nil || r.Operations == nil {
		return nil
	}

	o := r.Operations

	if err := o.Validate(); err != nil {
		return fmt.Errorf("assembling managed resource %s schema: %w", r.Name, err)
	}

	if err := o.MergeOperationShards(r.Name); err != nil {
		return fmt.Errorf("assembling managed resource %s schema: %w", r.Name, err)
	}

	for _, op := range o.Read {
		op.EntityMissingCodes = r.MissingCodes
	}

	for _, op := range o.Delete {
		op.EntityMissingCodes = r.MissingCodes
	}

	r.SchemaWarnings = nil
	r.IncludeSDKMethodOptions = r.computeIncludeSDKMethodOptions()

	// Collect soft-delete properties from the read response shard.
	r.ReadSoftDeleteProperties = collectSoftDeleteProperties(o.ReadResponseShard)

	// Base TypeDef starts from the merged create shard (combined request+response).
	r.SchemaTypeDef = o.CreateShard
	if r.SchemaTypeDef == nil {
		return nil
	}

	// Pre-processing on create shard: tag computed extensions and apply
	// equivalent field properties from the create request shard.
	r.SchemaTypeDef.SetExtensionOnEquivalentTypes(o.CreateResponseShard, "x-speakeasy-param-computed", true)
	r.SchemaTypeDef.TerraformApplyEquivalentFieldProperties(o.CreateRequestShard)

	// Clone to separate from the original merged shard so that assembly
	// mutations below don't contaminate the original shard data used for
	// later comparisons.
	r.SchemaTypeDef = r.SchemaTypeDef.Clone()
	r.SchemaTypeDef.RenameExtensionRecursive("x-speakeasy-param-readonly", "x-speakeasy-manual-param-readonly")

	// Merge CREATE operation shards.
	for _, op := range o.Create {
		if err := r.SchemaTypeDef.TerraformMerge(op.RequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging create request shard: %w", r.Name, err)
		}

		if op.ResponseShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.ResponseShard); err != nil {
				return fmt.Errorf("assembling managed resource %s schema: merging create response shard: %w", r.Name, err)
			}
		}
	}

	if o.CreateRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, o.CreateRequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging create alias nodes: %w", r.Name, err)
		}
	}

	// Validate x-speakeasy-match type compatibility for create operations.
	for _, op := range o.Create {
		if err := r.SchemaTypeDef.ValidateMatchConfigTypes(op.RequestShard, op.EntityOperation); err != nil {
			return err
		}
	}

	// Merge READ operation shards with validation.
	for _, op := range o.Read {
		if o.CreateShard != nil {
			if err := r.SchemaTypeDef.ValidateMatchConfigTypes(op.RequestShard, op.EntityOperation); err != nil {
				return err
			}

			if err := r.validateReadRequestResolvable(op); err != nil {
				return err
			}
		}

		if err := r.SchemaTypeDef.TerraformMerge(op.ResponseShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging read response shard: %w", r.Name, err)
		}
	}

	if o.ReadRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, o.ReadRequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging read alias nodes: %w", r.Name, err)
		}

		r.ImportStateTypeDef = o.ReadRequestShard.Clone()
		r.ImportStateTypeDef.TerraformImportRequiredFields()
	}

	// Merge DELETE operation shards.
	for _, op := range o.Delete {
		if err := r.SchemaTypeDef.TerraformMerge(op.RequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging delete request shard: %w", r.Name, err)
		}
	}

	if o.DeleteRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, o.DeleteRequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging delete alias nodes: %w", r.Name, err)
		}
	}

	// Merge UPDATE operation shards with validation and warnings.
	for _, op := range o.Update {
		if err := r.resolveUpdateRequestShard(op); err != nil {
			return err
		}

		r.warnUpdateFieldsNotInCreate(op)
	}

	if o.UpdateRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, o.UpdateRequestShard); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: merging update alias nodes: %w", r.Name, err)
		}
	}

	// Deep clone to break shared *TypeDef pointers from OAS $ref resolution,
	// making it safe for subsequent path-dependent modifications.
	r.SchemaTypeDef = r.SchemaTypeDef.DeepClone()

	// Apply CREATE extension modifications: mark response-exclusive fields as
	// readonly, all response fields as computed, and override readonly for
	// fields that appear in the request body.
	for _, op := range o.Create {
		// Prevent request attributes from being marked readonly.
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-param-readonly", false)

		if op.ResponseShard != nil {
			// Response-exclusive fields (not in request) are readonly and computed.
			r.SchemaTypeDef.SetExtensionOnExclusiveTypes(op.ResponseShard, op.RequestShard, "x-speakeasy-param-readonly", true)

			// All response fields are computed.
			r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-param-computed", true)

			// Fields in the create request body that were marked readonly by
			// another request should be computed instead, since the user
			// provides them directly.
			if op.APIOperation != nil && op.APIOperation.Request != nil && op.APIOperation.Request.Field != nil && op.APIOperation.Request.Field.Type != nil {
				entityShard := op.APIOperation.Request.Field.Type.FindEntityTypeDef(r.Name)
				if entityShard != nil {
					r.SchemaTypeDef.TerraformWalkEquivalent(entityShard, op.EntityOperation+".req.body", func(_ string, self *TypeDef, _ *TypeDef, _ bool) bool {
						if self.Extensions != nil {
							if val, ok := self.Extensions.Get("x-speakeasy-param-readonly"); ok && val == true {
								self.Extensions.Set("x-speakeasy-param-readonly", false)
								self.Extensions.Set("x-speakeasy-param-computed", true)
							}
						}

						return false
					}, false)
				}
			}
		}
	}

	// Apply READ extension modifications: mark response-exclusive fields as
	// readonly (not in create request) and all response fields as computed.
	if o.CreateShard != nil {
		for _, op := range o.Read {
			r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-in-get-request", true)
			r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-in-get", true)
			r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-param-computed", true)

			// Build merged create request shard across all create operations
			// to determine which response fields are not user-provided.
			var mergedCreateReq *TypeDef
			for i, createOp := range o.Create {
				if createOp.RequestShard == nil {
					continue
				}

				if i == 0 {
					mergedCreateReq = createOp.RequestShard.Clone()
				} else {
					mergedCreateReq.TerraformMerge(createOp.RequestShard) //nolint:errcheck // best-effort merge for comparison
				}
			}

			if mergedCreateReq != nil {
				r.SchemaTypeDef.SetExtensionOnExclusiveTypes(op.ResponseShard, mergedCreateReq, "x-speakeasy-param-readonly", true)
			}
		}
	}

	// Apply UPDATE extension modifications: preserve update wrapped attributes,
	// apply create request optionality, and mark force-new fields.

	// Preserve update wrapped attributes on the schema.
	if o.UpdateShard != nil && o.UpdateShard.Extensions != nil {
		for extensionKey, extensionValue := range o.UpdateShard.Extensions.All {
			if strings.HasPrefix(extensionKey, "x-speakeasy-wrapped-") {
				r.SchemaTypeDef.EnsureExtensions()
				r.SchemaTypeDef.Extensions.Set(extensionKey, extensionValue)
			}
		}
	}

	// Apply optionality from the create request shard.
	if o.CreateRequestShard != nil {
		r.SchemaTypeDef.TerraformApplyEquivalentFieldProperties(o.CreateRequestShard)
	}

	// When no update operation exists, all create request fields require
	// resource replacement since they cannot be changed in-place.
	if o.CreateRequestShard != nil && o.UpdateRequestShard != nil {
		r.SchemaTypeDef.SetExtensionOnExclusiveTypes(o.CreateRequestShard, o.UpdateRequestShard, "x-speakeasy-param-force-new", true)
	} else if o.CreateRequestShard != nil && o.UpdateRequestShard == nil {
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(o.CreateRequestShard, "x-speakeasy-param-force-new", true)
	}

	// Clear readonly on parent containers of nested match paths so they
	// become Optional+Computed instead of Computed-only (readonly).
	for _, shard := range []*TypeDef{o.CreateRequestShard, o.UpdateRequestShard} {
		r.SchemaTypeDef.TerraformClearReadonlyOnMatchPaths(shard)
	}

	// Apply shared finalization (root extension, conflicts-with, propagation,
	// trim, sort, and extension rename).
	r.SchemaTypeDef.TerraformFinalizeSchema(excludeEmptyObjectSchemas)

	// Validate alias nodes (x-speakeasy-match) in merged request shards
	// against the finalized schema.
	if err := r.SchemaTypeDef.ValidateAliasNodes(o.ReadRequestShard, o.UpdateRequestShard, o.DeleteRequestShard); err != nil {
		return fmt.Errorf("assembling managed resource %s schema: validating alias nodes: %w", r.Name, err)
	}

	// Cleanup per-operation shards by removing ignored fields. This must run
	// after assembly since shards are consumed during merge/validation.
	for _, op := range o.All() {
		op.CleanupShards()
	}

	// Pre-compute SDK method data per operation. Must run after schema
	// assembly since it cross-references SchemaTypeDef.
	for _, op := range o.All() {
		if err := op.setSDKMethodData(r.Name, r.SchemaTypeDef); err != nil {
			return fmt.Errorf("assembling managed resource %s schema: %w", r.Name, err)
		}
	}

	// Pre-compute deduplicated SDK method lists from the per-operation targets.
	r.SDKRequestMethods = o.All().SDKRequestMethods()
	r.SDKResponseMethods = o.DataModelRefreshOperations().SDKResponseMethods()

	return nil
}

// computeIncludeSDKMethodOptions returns true if any operation requires SDK
// method options in generated code. This is triggered by patch style
// configuration, prior state parameters on update operations, or write-only
// extensions on create/update operations.
func (r *TerraformManagedResource) computeIncludeSDKMethodOptions() bool {
	o := r.Operations

	for _, op := range o.All() {
		if op.HasPatchStyle() {
			return true
		}
	}

	for _, op := range o.Create {
		if op.HasTerraformWriteOnly() {
			return true
		}
	}

	for _, op := range o.Update {
		if op.HasUsePriorStateParameters() || op.HasTerraformWriteOnly() {
			return true
		}
	}

	return false
}

// SchemaDescription returns the description string for the Terraform schema.
// If a description is set via x-speakeasy-entity-description, it is returned
// directly. Otherwise, a default description is generated from the entity name.
func (r *TerraformManagedResource) SchemaDescription() string {
	if r.Description != "" {
		return r.Description
	}

	return TerraformGoTypeName(r.Name) + " Resource"
}

// SchemaVersion returns the schema version for the managed resource. Returns 0
// if no schema is configured.
func (r *TerraformManagedResource) SchemaVersion() int64 {
	if r.Schema == nil {
		return 0
	}

	return r.Schema.Version
}

// Returns true if TerraformManagedResource is valid. A valid managed resource
// must have at least one Create operation associated with it.
func (r *TerraformManagedResource) isValid() bool {
	if r == nil || r.Operations == nil {
		return false
	}

	return len(r.Operations.create) > 0
}

// resolveUpdateRequestShard merges the update operation's request shard into the
// schema, validating that all required fields are resolvable from the existing
// state. Optional parameters not available are warned and marked as ignored.
// Required parameters not available cause an error. On success, r.SchemaTypeDef
// is updated to the expanded version that includes the update request fields.
func (r *TerraformManagedResource) resolveUpdateRequestShard(op *TerraformOperation) error {
	if op.RequestShard == nil {
		return nil
	}

	expanded := r.SchemaTypeDef.Clone()

	if err := expanded.TerraformMerge(op.RequestShard); err != nil {
		return fmt.Errorf("assembling managed resource %s schema: update request shard resolution merge: %w", r.Name, err)
	}

	expanded.SetExtensionOnExclusiveTypes(op.RequestShard, r.SchemaTypeDef, "not-available", true)

	var errors []string

	expanded.TerraformWalkEquivalent(op.RequestShard, op.EntityOperation+".request", func(hierarchy string, self *TypeDef, other *TypeDef, optional bool) bool {
		if self.Extensions == nil || self.Extensions.All == nil {
			return false
		}

		if !self.Extensions.Has("not-available") {
			return false
		}

		if !optional {
			errors = append(errors, hierarchy)

			return false
		}

		r.SchemaWarnings = append(r.SchemaWarnings, fmt.Sprintf("Unexpected attribute %s: marking as ignored as this isn't available in the create/get requests", hierarchy))

		if self.Extensions.TerraformIgnore != nil {
			self.Extensions.TerraformIgnore.DataModel = true
		} else {
			self.Extensions.TerraformIgnore = &extensions.TerraformIgnore{
				DataModel: true,
				Schema:    true,
			}
		}

		other.EnsureExtensions()

		ignoreTrue := true
		other.Extensions.Ignore = &ignoreTrue
		self.Extensions.Remove("not-available")

		return true
	}, false)

	if len(errors) > 0 {
		return fmt.Errorf("found parameters in %s request that are not defined in create or get operations:\n%s\n\nConsider using x-speakeasy-match, x-speakeasy-name-override, x-speakeasy-ignore, const, input/output or other extensions to align the create and get operations such that the GET can be invoked from the create request/response",
			op.EntityOperation, "  "+strings.Join(errors, "\n  "))
	}

	expanded.TerraformDeleteIgnored()

	r.SchemaTypeDef = expanded

	return nil
}

// Sets the managed resource description from the operation. Searches both the
// the request and response for the first x-speakeasy-entity-description
// extension found.
func (r *TerraformManagedResource) setDescriptionFromOperation(operation *Operation) {
	if operation == nil {
		return
	}

	if r.Description != "" {
		return
	}

	if operation.Request != nil && operation.Request.Field != nil && operation.Request.Field.Type != nil {
		for _, typeDef := range operation.Request.Field.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityDescription == nil {
				continue
			}

			if typeDef.Extensions.EntityDescription.TerraformManagedResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformManagedResource
			return
		}
	}

	if operation.Response != nil && operation.Response.Type != nil {
		for _, typeDef := range operation.Response.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityDescription == nil {
				continue
			}

			if typeDef.Extensions.EntityDescription.TerraformManagedResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformManagedResource
			return
		}
	}
}

// Sets the global fields from the operation.
func (r *TerraformManagedResource) setGlobalFieldsFromOperation(operation *Operation) {
	if operation == nil || operation.Globals == nil || len(operation.Globals.Fields) == 0 {
		return
	}

	if r.GlobalFields == nil {
		r.GlobalFields = make(map[string]*FieldDef, len(operation.Globals.Fields))
	}

	for _, field := range operation.Globals.Fields {
		r.GlobalFields[SanitizeFieldName(field.Name)] = field
	}
}

// If the entity operation is read and x-speakeasy-entity-missing-codes is
// configured, set MissingCodes to that value, otherwise it will be the default.
// Removes all MissingCodes as errors from the Operation.
func (r *TerraformManagedResource) setMissingCodesFromOperation(entityOperationConfig extensions.EntityOperationV1Config, operation *Operation) {
	isDeleteOp := slices.Contains(entityOperationConfig.OperationTypes, string(terraform.ManagedResourceOperationTypeDelete))
	isReadOp := slices.Contains(entityOperationConfig.OperationTypes, string(terraform.ManagedResourceOperationTypeRead))

	if !isDeleteOp && !isReadOp {
		return
	}

	if operation == nil || operation.Response == nil {
		return
	}

	// Defensive default value. Extension can always override, including no codes.
	if len(r.MissingCodes) == 0 {
		r.MissingCodes = []int{http.StatusNotFound}
	}

	if operation.Extensions != nil && operation.Extensions.EntityMissingCodes != nil {
		r.MissingCodes = operation.Extensions.EntityMissingCodes
	}

	for _, code := range r.MissingCodes {
		subResponse := operation.Response.Responses.FindByCode(strconv.Itoa(code))

		if subResponse == nil {
			// Intentionally create non-error SubResponse to ensure underlying
			// SDK errors are not generated for this status code.
			subResponse = &SubResponse{
				Code:  []string{strconv.Itoa(code)},
				Error: false,
			}
			operation.Response.Responses = append(operation.Response.Responses, subResponse)

			continue
		}

		subResponse.Error = false
	}
}

// Sets the pagination fields from the operation.
func (r *TerraformManagedResource) setPaginationFieldsFromOperation(operation *Operation) {
	if operation == nil || operation.Extensions == nil || operation.Extensions.Pagination == nil {
		return
	}

	pagination := operation.Extensions.Pagination

	if r.PaginationInputFields == nil {
		r.PaginationInputFields = make(map[string]bool)
	}

	if r.PaginationOutputFields == nil {
		r.PaginationOutputFields = make(map[string]bool)
	}

	for _, input := range pagination.Inputs {
		// Preserve cursor pagination input fields since they may be
		// useful for Terraform consumers. e.g. "since"
		if input.Type == extensions.PaginationInputTypeCursor {
			continue
		}

		r.PaginationInputFields[SanitizeFieldName(input.Name)] = true
	}

	if pagination.Outputs.NextURL != "" {
		r.PaginationOutputFields[sanitizePaginationOutputName(pagination.Outputs.NextURL)] = true
	}

	if pagination.Outputs.NumPages != "" {
		r.PaginationOutputFields[sanitizePaginationOutputName(pagination.Outputs.NumPages)] = true
	}
}

// Sets the managed resource schema version from the operation. Searches both
// the request and response for the first x-speakeasy-entity-version extension
// found.
func (r *TerraformManagedResource) setSchemaVersionFromOperation(operation *Operation) error {
	if operation == nil {
		return nil
	}

	if r.Schema == nil {
		r.Schema = NewTerraformManagedResourceSchema()
	}

	if r.Schema.Version != 0 {
		return nil
	}

	if operation.Request != nil && operation.Request.Field != nil && operation.Request.Field.Type != nil {
		for _, typeDef := range operation.Request.Field.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityVersion == nil {
				continue
			}

			if typeDef.Extensions.EntityVersion.TerraformManagedResource == 0 {
				continue
			}

			return r.Schema.SetVersion(typeDef.Extensions.EntityVersion.TerraformManagedResource)
		}
	}

	if operation.Response != nil && operation.Response.Type != nil {
		for _, typeDef := range operation.Response.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityVersion == nil {
				continue
			}

			if typeDef.Extensions.EntityVersion.TerraformManagedResource == 0 {
				continue
			}

			return r.Schema.SetVersion(typeDef.Extensions.EntityVersion.TerraformManagedResource)
		}
	}

	return nil
}

// Enables OperationSecurity if the operation defines security and the
// enableOperationSecurity generation configuration is enabled.
func (r *TerraformManagedResource) setOperationSecurityFromOperation(generationConfig map[string]any, operation *Operation) {
	if operation == nil || operation.Security == nil {
		return
	}

	enableOperationSecurity, ok := generationConfig["enableOperationSecurity"].(bool)

	if !ok || !enableOperationSecurity {
		return
	}

	if r.OperationSecurity != nil {
		// NOTE: This assumes that all operations for a managed resource
		// have the same operation security configuration. If that requirement
		// changes, this logic should be updated to combine fields into a single
		// FieldDef so the templating logic for the schema and data model have
		// a complete view of all security fields.
		return
	}

	r.OperationSecurity = operation.Security
}

// Sets the Server from the operation.
func (r *TerraformManagedResource) setServerFromOperation(generationConfig map[string]any, operation *Operation) {
	if operation == nil || operation.Servers == nil || len(operation.Servers.Servers) == 0 {
		return
	}

	enableOperationServers, ok := generationConfig["enableOperationServers"].(bool)

	if !ok || !enableOperationServers {
		return
	}

	if r.Server != nil {
		return
	}

	server := operation.Servers.Servers[0]

	r.Server = &TerraformServer{
		// TODO: In the future, this likely needs to be configurable. For now,
		// this is hardcoded to match provider-level server URL configuration.
		AttributeName: "server_url",
		URL:           server.URL,
	}

	if server.Comments != nil && server.Comments.Description != "" {
		r.Server.Description = server.Comments.Description
	}
}

// validateReadRequestResolvable checks that every parameter in the read request
// is resolvable from the state (i.e., was present in the create operation).
// Optional parameters not in create are warned and marked as ignored. Required
// parameters not in create cause an error.
func (r *TerraformManagedResource) validateReadRequestResolvable(op *TerraformOperation) error {
	if op.RequestShard == nil {
		return nil
	}

	expanded := r.SchemaTypeDef.Clone()

	if err := expanded.TerraformMerge(op.RequestShard); err != nil {
		return fmt.Errorf("assembling managed resource %s schema: read request resolvable validation merge: %w", r.Name, err)
	}

	expanded.SetExtensionOnExclusiveTypes(op.RequestShard, r.Operations.CreateShard, "not-in-create", true)

	var errors []string

	expanded.TerraformWalkEquivalent(op.RequestShard, op.EntityOperation+".request", func(hierarchy string, self *TypeDef, _ *TypeDef, optional bool) bool {
		if self.Extensions == nil || self.Extensions.All == nil {
			return false
		}

		if !self.Extensions.Has("not-in-create") {
			return false
		}

		if self.Extensions.MatchConfig != nil {
			return false
		}

		if optional {
			r.SchemaWarnings = append(r.SchemaWarnings, hierarchy+" ignored as it is not in the create operation, and is optional")
			ignoreTrue := true
			self.Extensions.Ignore = &ignoreTrue

			return true
		}

		errors = append(errors, hierarchy)

		return false
	}, false)

	if len(errors) > 0 {
		return fmt.Errorf("found parameters in %s request that are required, but not defined in create operation: \n%s.\n\nConsider using x-speakeasy-match, x-speakeasy-name-override, x-speakeasy-ignore extensions to align the create and get operations such that the GET can be invoked from the create request/response",
			op.EntityOperation, "  "+strings.Join(errors, "\n  "))
	}

	return nil
}

// warnUpdateFieldsNotInCreate warns about fields that are in the update request
// but not in the create request & response. These may indicate the update
// operation should be marked as a second create step.
func (r *TerraformManagedResource) warnUpdateFieldsNotInCreate(op *TerraformOperation) {
	if op.RequestShard == nil || r.SchemaTypeDef == nil {
		return
	}

	ops := r.Operations

	duplicate := r.SchemaTypeDef.Clone()
	duplicate.SetExtensionOnExclusiveTypes(op.RequestShard, ops.CreateRequestShard, "not-in-create-request", true)
	duplicate.SetExtensionOnEquivalentTypes(op.RequestShard, "in-update-request", true)

	if ops.DeleteRequestShard != nil {
		duplicate.SetExtensionOnEquivalentTypes(ops.DeleteRequestShard, "in-delete-request", true)
	}

	if ops.ReadRequestShard != nil {
		duplicate.SetExtensionOnEquivalentTypes(ops.ReadRequestShard, "in-get-request", true)
	}

	duplicate.SetExtensionOnExclusiveTypes(op.RequestShard, ops.CreateResponseShard, "not-in-create-response", true)
	duplicate.SetExtensionOnExclusiveTypes(op.ResponseShard, op.RequestShard, "in-update-response-only", true)

	errorAttributes := make(map[string]struct{})

	duplicate.TerraformWalkTypes(SanitizeFieldName(r.Name), func(name string, typedef *TypeDef, _ bool) {
		if typedef.Extensions == nil || typedef.Extensions.All == nil {
			return
		}

		ext := typedef.Extensions.All

		if ext["not-in-create-request"] != nil && ext["not-in-create-request"] != false &&
			ext["in-update-request"] != nil && ext["in-update-request"] != false &&
			(ext["in-delete-request"] == nil || ext["in-delete-request"] == false) &&
			(ext["in-get-request"] == nil || ext["in-get-request"] == false) &&
			(ext["in-annotate-request"] == nil || ext["in-annotate-request"] == false) &&
			(ext["x-speakeasy-manual-param-readonly"] == nil || ext["x-speakeasy-manual-param-readonly"] == false) &&
			(ext["in-update-response-only"] == nil || ext["in-update-response-only"] == false) &&
			(ext["not-in-create-response"] == nil || ext["not-in-create-response"] == false) {
			errorAttributes[name] = struct{}{}
		}
	}, false)

	if len(errorAttributes) == 0 {
		return
	}

	// Sort by length then lexicographic.
	names := make([]string, 0, len(errorAttributes))
	for name := range errorAttributes {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		if len(names[i]) != len(names[j]) {
			return len(names[i]) < len(names[j])
		}

		return names[i] < names[j]
	})

	var msg strings.Builder
	msg.WriteString("may want to invoke update request additionally due to these attributes being available in the update request, but not the create request:\n")

	for _, name := range names {
		msg.WriteString("  " + name + "\n")
	}

	r.SchemaWarnings = append(r.SchemaWarnings,
		fmt.Sprintf("%s#update %sMark the object as:\nx-speakeasy-entity-operation:\n  - %s#create#2\n  - %s#update\n if these attributes are mutable\nElse mark the types with \"x-speakeasy-param-readonly: true\" to make this message go away\n",
			r.Name, msg.String(), r.Name, r.Name))
}
