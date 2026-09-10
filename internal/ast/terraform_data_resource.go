package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

// Describes a Terraform data resource.
type TerraformDataResource struct {
	// Description for the data resource. Sourced from
	// x-speakeasy-entity-description configuration, if available.
	Description string `json:"description" yaml:"description"`

	// Mapping of global field names to definitions for all operations. These
	// are added to the entity resource struct type, copied in the Configure()
	// method, made optional in the resource schema, and checked in the resource
	// methods.
	GlobalFields map[string]*FieldDef `json:"globalFields" yaml:"globalFields"`

	// Computed Go type name for the data resource data model struct, e.g.
	// "ExampleDataSourceModel". Set by NewTerraformDataResource.
	GoDataModelTypeName string `json:"goDataModelTypeName" yaml:"goDataModelTypeName"`

	// Computed Go type name for the data resource private data model struct, e.g.
	// "ExampleDataSourcePrivateDataModel". Set by NewTerraformDataResource.
	GoPrivateDataModelTypeName string `json:"goPrivateDataModelTypeName" yaml:"goPrivateDataModelTypeName"`

	// Computed Go type name for the data resource struct implementation, e.g.
	// "ExampleDataSource". Set by NewTerraformDataResource.
	GoTypeName string `json:"goTypeName" yaml:"goTypeName"`

	// Whether the entity requires SDK method options in its generated code.
	// This is true when any operation has a Patch.Style configuration.
	// Populated by AssembleSchemaTypeDef.
	IncludeSDKMethodOptions bool `json:"includeSDKMethodOptions,omitempty" yaml:"includeSDKMethodOptions,omitempty"`

	// Unsanitized name of the data resource type, such as "Thing".
	Name string `json:"name" yaml:"name"`

	// Operation security configuration for the data resource. This is set
	// when the underlying operations have operation security defined and the
	// enableOperationSecurity generation configuration flag is enabled. Only
	// a single operation security configuration is supported per data resource,
	// even if multiple operations have differing security configurations. This
	// is intentional to simplify the generated Terraform code and avoid
	// complexity around per-operation security configuration for consumers.
	OperationSecurity *FieldDef `json:"operationSecurity" yaml:"operationSecurity"`

	// All operations associated with the data resource.
	Operations *TerraformDataResourceOperations `json:"operations" yaml:"operations"`

	// Mapping of pagination input field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationInputFields map[string]bool `json:"paginationInputFields" yaml:"paginationInputFields"`

	// Mapping of pagination output field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationOutputFields map[string]bool `json:"paginationOutputFields" yaml:"paginationOutputFields"`

	// Merged Terraform resource schema TypeDef across all operations. This
	// represents the combined schema view of the data resource after merging
	// all operation shards together.
	SchemaTypeDef *TypeDef `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK request conversion methods (To_ prefix)
	// pre-computed from all operations. Populated by AssembleSchemaTypeDef.
	SDKRequestMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK response conversion methods
	// (RefreshFrom_ or RefreshFromArrayOf_ prefix) pre-computed from
	// DataModelRefreshOperations. Populated by AssembleSchemaTypeDef.
	SDKResponseMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Server configuration for the data resource. This is defined when
	// the underlying operations have path or operation server URLs defined.
	// Only a single server configuration is supported per data resource,
	// even if multiple operations have differing server URLs defined. This is
	// intentional to simplify the generated Terraform code and avoid complexity
	// around per-operation server configuration for consumers.
	Server *TerraformServer `json:"server" yaml:"server"`

	// Resolved full Terraform data source type name (e.g.
	// "myprovider_my_data_source"), composed from the resolved provider type
	// name and the snake_case form of Name. Used as the value of
	// resp.TypeName in the data source Metadata implementation and in
	// example Terraform configuration files. Populated by
	// TerraformProvider.AddOrGetDataResource.
	TerraformTypeName string `json:"terraformTypeName" yaml:"terraformTypeName"`
}

// Creates a new Terraform data resource, safely initializing underlying fields.
func NewTerraformDataResource(name string) *TerraformDataResource {
	goTypeName := TerraformGoTypeName(name) + "DataSource"

	return &TerraformDataResource{
		GoDataModelTypeName:        goTypeName + "Model",
		GoPrivateDataModelTypeName: goTypeName + "PrivateDataModel",
		GoTypeName:                 goTypeName,
		Name:                       name,
		Operations:                 NewTerraformDataResourceOperations(),
	}
}

// Adds an operation to the data resource.
func (r *TerraformDataResource) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, operation *Operation) error {
	if r.Operations == nil {
		r.Operations = NewTerraformDataResourceOperations()
	}

	if operation == nil {
		return nil
	}

	if entityOperationConfig.Entity != r.Name {
		return nil
	}

	if err := r.Operations.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
		return err
	}

	r.setDescriptionFromOperation(operation)
	r.setGlobalFieldsFromOperation(operation)
	r.setOperationSecurityFromOperation(generationConfig, operation)
	r.setPaginationFieldsFromOperation(operation)
	r.setServerFromOperation(generationConfig, operation)

	return nil
}

// AssembleSchemaTypeDef builds the merged schema TypeDef for the data resource
// by combining all read operation shards with pre-processing and per-operation
// extension tagging. The result is deep-cloned to break shared *TypeDef
// pointers from OAS $ref resolution, making it safe for subsequent
// modifications.
//
// The method performs the following steps:
//  1. Merge individual operation shards into combined per-type shards
//  2. Pre-processing: tag readonly/computed extensions and apply field properties
//  3. Override unsound readonly fields using per-op request shards
//  4. Rename pre-processed extensions (readonly → manual-param-readonly)
//  5. Merge per-operation shards and alias nodes
//  6. Per-operation extension tagging and cleanup of non-read fields
//  7. DeepClone to break all shared pointers
func (r *TerraformDataResource) AssembleSchemaTypeDef(excludeEmptyObjectSchemas bool) error {
	if r.Operations == nil {
		return nil
	}

	ops := r.Operations

	if err := ops.Validate(); err != nil {
		return fmt.Errorf("assembling data resource %s schema: %w", r.Name, err)
	}

	if err := ops.MergeOperationShards(); err != nil {
		return fmt.Errorf("assembling data resource %s schema: %w", r.Name, err)
	}

	// Compute IncludeSDKMethodOptions from operation conditions.
	r.IncludeSDKMethodOptions = false

	for _, op := range ops.Read {
		if op.HasPatchStyle() {
			r.IncludeSDKMethodOptions = true
			break
		}
	}

	// Base TypeDef starts from the merged read shard (combined request+response).
	r.SchemaTypeDef = ops.ReadShard
	if r.SchemaTypeDef == nil {
		return nil
	}

	r.SchemaTypeDef = r.SchemaTypeDef.Clone()

	// Pre-process on the clone: tag readonly/computed extensions and apply
	// equivalent field properties. This operates on the clone rather than the
	// original shard to avoid mutating the Go AST data.
	r.SchemaTypeDef.SetExtensionOnExclusiveTypes(ops.ReadResponseShard, ops.ReadRequestShard, "x-speakeasy-param-readonly", true)
	r.SchemaTypeDef.SetExtensionOnEquivalentTypes(ops.ReadResponseShard, "x-speakeasy-param-computed", true)
	r.SchemaTypeDef.TerraformApplyEquivalentFieldProperties(ops.ReadRequestShard)

	// Override unsound readonly fields: any field that appears in a per-op
	// request shard is a user-provided input and should not be marked as
	// manual-readonly after the rename step below. This includes both
	// direct name matches and alias matches via x-speakeasy-match.
	for _, op := range ops.Read {
		if op.RequestShard != nil {
			r.SchemaTypeDef.OverrideExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-param-readonly", false)
			r.SchemaTypeDef.OverrideExtensionOnAliasMatchedFields(op.RequestShard, "x-speakeasy-param-readonly", false)
		}
	}

	// Pre-assembly extension rename so that per-operation tagging below sees
	// the renamed extension rather than the original.
	r.SchemaTypeDef.RenameExtensionRecursive("x-speakeasy-param-readonly", "x-speakeasy-manual-param-readonly")

	// Merge each read operation's request and response shards.
	for _, op := range ops.Read {
		if op.RequestShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.RequestShard); err != nil {
				return fmt.Errorf("assembling data resource %s schema: merging read request shard: %w", r.Name, err)
			}
		}

		if err := r.SchemaTypeDef.TerraformMerge(op.ResponseShard); err != nil {
			return fmt.Errorf("assembling data resource %s schema: merging read response shard: %w", r.Name, err)
		}
	}

	// Merge alias nodes from the merged read request shard.
	if ops.ReadRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, ops.ReadRequestShard); err != nil {
			return fmt.Errorf("assembling data resource %s schema: merging read alias nodes: %w", r.Name, err)
		}
	}

	// Per-operation extension tagging.
	for _, op := range ops.Read {
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-in-get-request", true)
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-in-get", true)
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-param-computed", true)
		r.SchemaTypeDef.SetExtensionOnExclusiveTypes(op.ResponseShard, op.RequestShard, "x-speakeasy-param-readonly", true)
	}

	// Delete fields not in read request/response.
	r.SchemaTypeDef.SetExtensionRecursive("x-speakeasy-ignore", true)
	r.SchemaTypeDef.RemoveExtensionOnEquivalentTypes(ops.ReadRequestShard, "x-speakeasy-ignore")
	r.SchemaTypeDef.RemoveExtensionOnEquivalentTypes(ops.ReadResponseShard, "x-speakeasy-ignore")
	r.SchemaTypeDef.TerraformDeleteIgnored()

	// Clean up soft-delete extension from schema.
	r.SchemaTypeDef.RemoveExtensionRecursive("x-speakeasy-soft-delete-property")

	// DeepClone breaks all shared *TypeDef pointers introduced during
	// assembly. TerraformFinalizeSchema applies shared post-assembly
	// modifications (root extension, conflicts-with, propagation, trim, sort,
	// and extension rename) common to all entity types.
	r.SchemaTypeDef = r.SchemaTypeDef.DeepClone()
	r.SchemaTypeDef.TerraformFinalizeSchema(excludeEmptyObjectSchemas)

	// Promote response filter fields to Optional+Computed so users can provide
	// filter values in data source configurations.
	r.SchemaTypeDef.PromoteResponseFilterFields()

	// Hoist response filter fields from array items to the entity level for
	// the "entity-above-array" pattern where the entity wraps a filtered array.
	// Must run after PromoteResponseFilterFields so it can undo the item-level
	// promotion while keeping the hoisted field as Optional-only.
	r.SchemaTypeDef.HoistArrayItemResponseFilterFields()

	// Validate alias nodes (x-speakeasy-match) in merged request shards
	// against the finalized schema.
	if err := r.SchemaTypeDef.ValidateAliasNodes(ops.ReadRequestShard); err != nil {
		return fmt.Errorf("assembling data resource %s schema: validating alias nodes: %w", r.Name, err)
	}

	// Cleanup per-operation shards by removing ignored fields.
	for _, op := range ops.Read {
		op.CleanupShards()
	}

	// Pre-compute SDK method data per operation. Must run after schema
	// assembly since it cross-references SchemaTypeDef.
	for _, op := range ops.All() {
		if err := op.setSDKMethodData(r.Name, r.SchemaTypeDef); err != nil {
			return fmt.Errorf("assembling data resource %s schema: %w", r.Name, err)
		}
	}

	// Pre-compute deduplicated SDK method lists from the per-operation targets.
	r.SDKRequestMethods = ops.All().SDKRequestMethods()
	r.SDKResponseMethods = ops.DataModelRefreshOperations().SDKResponseMethods()

	return nil
}

// SchemaDescription returns the description string for the Terraform schema.
// If a description is set via x-speakeasy-entity-description, it is returned
// directly. Otherwise, a default description is generated from the entity name.
func (r *TerraformDataResource) SchemaDescription() string {
	if r.Description != "" {
		return r.Description
	}

	return TerraformGoTypeName(r.Name) + " DataSource"
}

// Returns true if TerraformDataResource is valid. A valid data resource must
// have at least one Read operation associated with it.
func (r *TerraformDataResource) isValid() bool {
	if r == nil || r.Operations == nil {
		return false
	}

	return len(r.Operations.read) > 0
}

// Sets the data resource description from the operation. Searches both the
// request and response for the first x-speakeasy-entity-description extension
// found.
func (r *TerraformDataResource) setDescriptionFromOperation(operation *Operation) {
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

			if typeDef.Extensions.EntityDescription.TerraformDataResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformDataResource
			return
		}
	}

	if operation.Response != nil && operation.Response.Type != nil {
		for _, typeDef := range operation.Response.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityDescription == nil {
				continue
			}

			if typeDef.Extensions.EntityDescription.TerraformDataResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformDataResource
			return
		}
	}
}

// Sets the global fields from the operation.
func (r *TerraformDataResource) setGlobalFieldsFromOperation(operation *Operation) {
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

// Sets the pagination fields from the operation.
func (r *TerraformDataResource) setPaginationFieldsFromOperation(operation *Operation) {
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

// Enables OperationSecurity if the operation defines security and the
// enableOperationSecurity generation configuration is enabled.
func (r *TerraformDataResource) setOperationSecurityFromOperation(generationConfig map[string]any, operation *Operation) {
	if operation == nil || operation.Security == nil {
		return
	}

	enableOperationSecurity, ok := generationConfig["enableOperationSecurity"].(bool)

	if !ok || !enableOperationSecurity {
		return
	}

	if r.OperationSecurity != nil {
		// NOTE: This assumes that all operations for a data resource
		// have the same operation security configuration. If that requirement
		// changes, this logic should be updated to combine fields into a single
		// FieldDef so the templating logic for the schema and data model have
		// a complete view of all security fields.
		return
	}

	r.OperationSecurity = operation.Security
}

// Sets the Server from the operation.
func (r *TerraformDataResource) setServerFromOperation(generationConfig map[string]any, operation *Operation) {
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
