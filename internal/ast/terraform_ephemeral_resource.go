package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

// Describes a Terraform ephemeral resource.
type TerraformEphemeralResource struct {
	// Description for the ephemeral resource. Sourced from
	// x-speakeasy-entity-description configuration, if available.
	Description string `json:"description" yaml:"description"`

	// Mapping of global field names to definitions for all operations. These
	// are added to the entity resource struct type, copied in the Configure()
	// method, made optional in the resource schema, and checked in the resource
	// methods.
	GlobalFields map[string]*FieldDef `json:"globalFields" yaml:"globalFields"`

	// Computed Go type name for the ephemeral resource data model struct, e.g.
	// "ExampleEphemeralResourceModel". Set by NewTerraformEphemeralResource.
	GoDataModelTypeName string `json:"goDataModelTypeName" yaml:"goDataModelTypeName"`

	// Computed Go type name for the ephemeral resource private data model struct,
	// e.g. "ExampleEphemeralResourcePrivateDataModel". Set by
	// NewTerraformEphemeralResource.
	GoPrivateDataModelTypeName string `json:"goPrivateDataModelTypeName" yaml:"goPrivateDataModelTypeName"`

	// Computed Go type name for the ephemeral resource struct implementation, e.g.
	// "ExampleEphemeralResource". Set by NewTerraformEphemeralResource.
	GoTypeName string `json:"goTypeName" yaml:"goTypeName"`

	// Whether the entity requires SDK method options in its generated code.
	// This is always false for ephemeral resources since their operations
	// (open/close) are not checked for SDK method option conditions.
	IncludeSDKMethodOptions bool `json:"includeSDKMethodOptions,omitempty" yaml:"includeSDKMethodOptions,omitempty"`

	// Unsanitized name of the ephemeral resource type, such as "Thing".
	Name string `json:"name" yaml:"name"`

	// Operation security configuration for the ephemeral resource. This is set
	// when the underlying operations have operation security defined and the
	// enableOperationSecurity generation configuration flag is enabled. Only
	// a single operation security configuration is supported per ephemeral resource,
	// even if multiple operations have differing security configurations. This
	// is intentional to simplify the generated Terraform code and avoid
	// complexity around per-operation security configuration for consumers.
	OperationSecurity *FieldDef `json:"operationSecurity" yaml:"operationSecurity"`

	// All operations associated with the ephemeral resource.
	Operations *TerraformEphemeralResourceOperations `json:"operations" yaml:"operations"`

	// Mapping of pagination input field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationInputFields map[string]bool `json:"paginationInputFields" yaml:"paginationInputFields"`

	// Mapping of pagination output field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationOutputFields map[string]bool `json:"paginationOutputFields" yaml:"paginationOutputFields"`

	// Merged Terraform resource schema TypeDef across all operations. This
	// represents the combined schema view of the ephemeral resource after
	// merging all operation shards together.
	SchemaTypeDef *TypeDef `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK request conversion methods (To_ prefix)
	// pre-computed from all operations. Populated by AssembleSchemaTypeDef.
	SDKRequestMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK response conversion methods
	// (RefreshFrom_ or RefreshFromArrayOf_ prefix) pre-computed from
	// DataModelRefreshOperations. Populated by AssembleSchemaTypeDef.
	SDKResponseMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Server configuration for the ephemeral resource. This is defined when
	// the underlying operations have path or operation server URLs defined.
	// Only a single server configuration is supported per ephemeral resource,
	// even if multiple operations have differing server URLs defined. This is
	// intentional to simplify the generated Terraform code and avoid complexity
	// around per-operation server configuration for consumers.
	Server *TerraformServer `json:"server" yaml:"server"`

	// Resolved full Terraform ephemeral resource type name (e.g.
	// "myprovider_my_ephemeral_resource"), composed from the resolved
	// provider type name and the snake_case form of Name. Used as the value
	// of resp.TypeName in the ephemeral resource Metadata implementation and
	// in example Terraform configuration files. Populated by
	// TerraformProvider.AddOrGetEphemeralResource.
	TerraformTypeName string `json:"terraformTypeName" yaml:"terraformTypeName"`
}

// Creates a new Terraform ephemeral resource, safely initializing underlying fields.
func NewTerraformEphemeralResource(name string) *TerraformEphemeralResource {
	goTypeName := TerraformGoTypeName(name) + "EphemeralResource"

	return &TerraformEphemeralResource{
		GoDataModelTypeName:        goTypeName + "Model",
		GoPrivateDataModelTypeName: goTypeName + "PrivateDataModel",
		GoTypeName:                 goTypeName,
		Name:                       name,
		Operations:                 NewTerraformEphemeralResourceOperations(),
	}
}

// Adds an operation to the ephemeral resource.
func (r *TerraformEphemeralResource) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, operation *Operation) error {
	if r.Operations == nil {
		r.Operations = NewTerraformEphemeralResourceOperations()
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

// AssembleSchemaTypeDef builds the merged schema TypeDef for the ephemeral
// resource by combining all open operation shards and close alias nodes with
// pre-processing and per-operation extension tagging. The result is deep-cloned
// to break shared *TypeDef pointers from OAS $ref resolution, making it safe
// for subsequent modifications.
//
// The method performs the following steps:
//  1. Merge individual operation shards into combined per-type shards
//  2. Pre-processing: tag readonly/computed extensions and apply field properties
//  3. Rename pre-processed extensions (readonly → manual-param-readonly)
//  4. Merge per-operation shards and alias nodes (open + close)
//  5. Per-operation extension tagging and request body readonly override
//  6. DeepClone to break all shared pointers
func (r *TerraformEphemeralResource) AssembleSchemaTypeDef(excludeEmptyObjectSchemas bool) error {
	if r.Operations == nil {
		return nil
	}

	ops := r.Operations

	if err := ops.Validate(); err != nil {
		return fmt.Errorf("assembling ephemeral resource %s schema: %w", r.Name, err)
	}

	if err := ops.MergeOperationShards(); err != nil {
		return fmt.Errorf("assembling ephemeral resource %s schema: %w", r.Name, err)
	}

	// Base TypeDef starts from the merged open shard (combined request+response).
	// If nil, fall back to the first open operation's request shard.
	r.SchemaTypeDef = ops.OpenShard
	if r.SchemaTypeDef != nil {
		r.SchemaTypeDef = r.SchemaTypeDef.Clone()
	} else {
		for _, op := range ops.Open {
			if op.RequestShard != nil {
				r.SchemaTypeDef = op.RequestShard.Clone()
				break
			}
		}
	}

	if r.SchemaTypeDef == nil {
		return nil
	}

	// Pre-process on the clone: tag readonly/computed extensions and apply
	// equivalent field properties. This operates on the clone rather than the
	// original shard to avoid mutating the Go AST data.
	r.SchemaTypeDef.SetExtensionOnExclusiveTypes(ops.OpenResponseShard, ops.OpenRequestShard, "x-speakeasy-param-readonly", true)
	r.SchemaTypeDef.SetExtensionOnEquivalentTypes(ops.OpenResponseShard, "x-speakeasy-param-computed", true)
	r.SchemaTypeDef.TerraformApplyEquivalentFieldProperties(ops.OpenRequestShard)

	// Pre-assembly extension rename so that per-operation tagging below sees
	// the renamed extension rather than the original.
	r.SchemaTypeDef.RenameExtensionRecursive("x-speakeasy-param-readonly", "x-speakeasy-manual-param-readonly")

	// Merge each open operation's request and response shards.
	for _, op := range ops.Open {
		if op.RequestShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.RequestShard); err != nil {
				return fmt.Errorf("assembling ephemeral resource %s schema: merging open request shard: %w", r.Name, err)
			}
		}

		if op.ResponseShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.ResponseShard); err != nil {
				return fmt.Errorf("assembling ephemeral resource %s schema: merging open response shard: %w", r.Name, err)
			}
		}
	}

	// Merge alias nodes from the merged open request shard.
	if ops.OpenRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, ops.OpenRequestShard); err != nil {
			return fmt.Errorf("assembling ephemeral resource %s schema: merging open alias nodes: %w", r.Name, err)
		}
	}

	// Merge alias nodes from the merged close request shard.
	if ops.CloseRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, ops.CloseRequestShard); err != nil {
			return fmt.Errorf("assembling ephemeral resource %s schema: merging close alias nodes: %w", r.Name, err)
		}
	}

	// Per-operation extension tagging.
	for _, op := range ops.Open {
		// Prevent request attributes from being marked readonly.
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-param-readonly", false)

		if op.ResponseShard != nil {
			// Response-exclusive fields must be readonly.
			r.SchemaTypeDef.SetExtensionOnExclusiveTypes(op.ResponseShard, op.RequestShard, "x-speakeasy-param-readonly", true)
			// All response fields must be computed.
			r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.ResponseShard, "x-speakeasy-param-computed", true)
		}

		// Override readonly for fields in the request body, since the
		// user provides them directly. Only the readonly override is
		// needed; computed is already set above for response fields.
		if op.APIOperation != nil && op.APIOperation.Request != nil && op.APIOperation.Request.Field != nil && op.APIOperation.Request.Field.Type != nil {
			entityShard := op.APIOperation.Request.Field.Type.FindEntityTypeDef(r.Name)
			if entityShard != nil {
				r.SchemaTypeDef.OverrideExtensionOnEquivalentTypes(entityShard, "x-speakeasy-param-readonly", false)
			}
		}
	}

	// Clean up soft-delete extension from schema.
	r.SchemaTypeDef.RemoveExtensionRecursive("x-speakeasy-soft-delete-property")

	// DeepClone breaks all shared *TypeDef pointers introduced during
	// assembly. TerraformFinalizeSchema applies shared post-assembly
	// modifications (root extension, conflicts-with, propagation, trim, sort,
	// and extension rename) common to all entity types.
	r.SchemaTypeDef = r.SchemaTypeDef.DeepClone()
	r.SchemaTypeDef.TerraformFinalizeSchema(excludeEmptyObjectSchemas)

	// Cleanup per-operation shards by removing ignored fields.
	for _, op := range ops.Open {
		op.CleanupShards()
	}
	for _, op := range ops.Close {
		op.CleanupShards()
	}

	// Pre-compute SDK method data per operation. Must run after schema
	// assembly since it cross-references SchemaTypeDef.
	for _, op := range ops.All() {
		if err := op.setSDKMethodData(r.Name, r.SchemaTypeDef); err != nil {
			return fmt.Errorf("assembling ephemeral resource %s schema: %w", r.Name, err)
		}
	}

	// Pre-compute deduplicated SDK method lists from the per-operation targets.
	r.SDKRequestMethods = ops.All().SDKRequestMethods()
	r.SDKResponseMethods = ops.DataModelRefreshOperations().SDKResponseMethods()

	return nil
}

// HasCloseOperations returns whether the ephemeral resource has close
// operations defined.
func (r *TerraformEphemeralResource) HasCloseOperations() bool {
	return len(r.Operations.Close) > 0
}

// SchemaDescription returns the description string for the Terraform schema.
// If a description is set via x-speakeasy-entity-description, it is returned
// directly. Otherwise, a default description is generated from the entity name.
func (r *TerraformEphemeralResource) SchemaDescription() string {
	if r.Description != "" {
		return r.Description
	}

	return TerraformGoTypeName(r.Name) + " Ephemeral Resource"
}

// Returns true if TerraformEphemeralResource is valid. A valid ephemeral resource
// must have at least one Open operation associated with it.
func (r *TerraformEphemeralResource) isValid() bool {
	if r == nil || r.Operations == nil {
		return false
	}

	return len(r.Operations.open) > 0
}

// Sets the ephemeral resource description from the operation. Searches both the
// request and response for the first x-speakeasy-entity-description extension
// found.
func (r *TerraformEphemeralResource) setDescriptionFromOperation(operation *Operation) {
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

			if typeDef.Extensions.EntityDescription.TerraformEphemeralResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformEphemeralResource
			return
		}
	}

	if operation.Response != nil && operation.Response.Type != nil {
		for _, typeDef := range operation.Response.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityDescription == nil {
				continue
			}

			if typeDef.Extensions.EntityDescription.TerraformEphemeralResource == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformEphemeralResource
			return
		}
	}
}

// Sets the global fields from the operation.
func (r *TerraformEphemeralResource) setGlobalFieldsFromOperation(operation *Operation) {
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
func (r *TerraformEphemeralResource) setPaginationFieldsFromOperation(operation *Operation) {
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
func (r *TerraformEphemeralResource) setOperationSecurityFromOperation(generationConfig map[string]any, operation *Operation) {
	if operation == nil || operation.Security == nil {
		return
	}

	enableOperationSecurity, ok := generationConfig["enableOperationSecurity"].(bool)

	if !ok || !enableOperationSecurity {
		return
	}

	if r.OperationSecurity != nil {
		// NOTE: This assumes that all operations for an ephemeral resource
		// have the same operation security configuration. If that requirement
		// changes, this logic should be updated to combine fields into a single
		// FieldDef so the templating logic for the schema and data model have
		// a complete view of all security fields.
		return
	}

	r.OperationSecurity = operation.Security
}

// Sets the Server from the operation.
func (r *TerraformEphemeralResource) setServerFromOperation(generationConfig map[string]any, operation *Operation) {
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
		// Reference: internal issue reference
		AttributeName: "server_url",
		URL:           server.URL,
	}

	if server.Comments != nil && server.Comments.Description != "" {
		r.Server.Description = server.Comments.Description
	}
}
