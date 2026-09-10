package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

// Describes a Terraform action.
type TerraformAction struct {
	// Description for the action. Sourced from
	// x-speakeasy-entity-description configuration, if available.
	Description string `json:"description" yaml:"description"`

	// Mapping of global field names to definitions for all operations. These
	// are added to the entity resource struct type, copied in the Configure()
	// method, made optional in the resource schema, and checked in the resource
	// methods.
	GlobalFields map[string]*FieldDef `json:"globalFields" yaml:"globalFields"`

	// Computed Go type name for the action data model struct, e.g.
	// "ExampleActionModel". Set by NewTerraformAction.
	GoDataModelTypeName string `json:"goDataModelTypeName" yaml:"goDataModelTypeName"`

	// Computed Go type name for the action private data model struct, e.g.
	// "ExampleActionPrivateDataModel". Set by NewTerraformAction.
	GoPrivateDataModelTypeName string `json:"goPrivateDataModelTypeName" yaml:"goPrivateDataModelTypeName"`

	// Computed Go type name for the action struct implementation, e.g.
	// "ExampleAction". Set by NewTerraformAction.
	GoTypeName string `json:"goTypeName" yaml:"goTypeName"`

	// Whether the entity requires SDK method options in its generated code.
	// This is true when any operation has a Patch.Style configuration.
	// Populated by AssembleSchemaTypeDef.
	IncludeSDKMethodOptions bool `json:"includeSDKMethodOptions,omitempty" yaml:"includeSDKMethodOptions,omitempty"`

	// Unsanitized name of the action type, such as "Thing".
	Name string `json:"name" yaml:"name"`

	// Operation security configuration for the action. This is set
	// when the underlying operations have operation security defined and the
	// enableOperationSecurity generation configuration flag is enabled. Only
	// a single operation security configuration is supported per action,
	// even if multiple operations have differing security configurations. This
	// is intentional to simplify the generated Terraform code and avoid
	// complexity around per-operation security configuration for consumers.
	OperationSecurity *FieldDef `json:"operationSecurity" yaml:"operationSecurity"`

	// All operations associated with the action.
	Operations *TerraformActionOperations `json:"operations" yaml:"operations"`

	// Mapping of pagination input field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationInputFields map[string]bool `json:"paginationInputFields" yaml:"paginationInputFields"`

	// Mapping of pagination output field names to an unused boolean value
	// that is replaceable for future updates.
	PaginationOutputFields map[string]bool `json:"paginationOutputFields" yaml:"paginationOutputFields"`

	// Merged Terraform resource schema TypeDef across all operations. This
	// represents the combined schema view of the action after merging all
	// operation shards together.
	SchemaTypeDef *TypeDef `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK request conversion methods (To_ prefix)
	// pre-computed from all operations. Populated by AssembleSchemaTypeDef.
	SDKRequestMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Deduplicated, sorted list of SDK response conversion methods
	// (RefreshFrom_ or RefreshFromArrayOf_ prefix) pre-computed from
	// DataModelRefreshOperations. Populated by AssembleSchemaTypeDef.
	SDKResponseMethods []*TerraformSDKMethod `json:"-" yaml:"-"`

	// Server configuration for the action. This is defined when
	// the underlying operations have path or operation server URLs defined.
	// Only a single server configuration is supported per action,
	// even if multiple operations have differing server URLs defined. This is
	// intentional to simplify the generated Terraform code and avoid complexity
	// around per-operation server configuration for consumers.
	Server *TerraformServer `json:"server" yaml:"server"`

	// Resolved full Terraform action type name (e.g.
	// "myprovider_my_action"), composed from the resolved provider type
	// name and the snake_case form of Name. Used as the value of
	// resp.TypeName in the action Metadata implementation and in example
	// Terraform configuration files. Populated by
	// TerraformProvider.AddOrGetAction.
	TerraformTypeName string `json:"terraformTypeName" yaml:"terraformTypeName"`
}

// Creates a new Terraform action, safely initializing underlying fields.
func NewTerraformAction(name string) *TerraformAction {
	goTypeName := TerraformGoTypeName(name) + "Action"

	return &TerraformAction{
		GoDataModelTypeName:        goTypeName + "Model",
		GoPrivateDataModelTypeName: goTypeName + "PrivateDataModel",
		GoTypeName:                 goTypeName,
		Name:                       name,
		Operations:                 NewTerraformActionOperations(),
	}
}

// Adds an operation to the action.
func (r *TerraformAction) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, operation *Operation) error {
	if r.Operations == nil {
		r.Operations = NewTerraformActionOperations()
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

// AssembleSchemaTypeDef builds the merged schema TypeDef for the action by
// combining all invoke operation shards. The result is deep-cloned to break
// shared *TypeDef pointers from OAS $ref resolution, making it safe for
// subsequent path-dependent modifications (extension tagging, propagation).
func (r *TerraformAction) AssembleSchemaTypeDef(excludeEmptyObjectSchemas bool) error {
	if r.Operations == nil {
		return nil
	}

	ops := r.Operations

	if err := ops.Validate(); err != nil {
		return fmt.Errorf("assembling action %s schema: %w", r.Name, err)
	}

	if err := ops.MergeOperationShards(); err != nil {
		return fmt.Errorf("assembling action %s schema: %w", r.Name, err)
	}

	// Compute IncludeSDKMethodOptions from operation conditions.
	r.IncludeSDKMethodOptions = false

	for _, op := range ops.Invoke {
		if op.HasPatchStyle() {
			r.IncludeSDKMethodOptions = true
			break
		}
	}

	// Base TypeDef starts from the merged invoke request shard.
	// If nil, fall back to the first invoke operation's request shard.
	r.SchemaTypeDef = ops.InvokeRequestShard
	if r.SchemaTypeDef != nil {
		r.SchemaTypeDef = r.SchemaTypeDef.Clone()
	} else {
		for _, op := range ops.Invoke {
			if op.RequestShard != nil {
				r.SchemaTypeDef = op.RequestShard.Clone()
				break
			}
		}
	}

	if r.SchemaTypeDef == nil {
		return nil
	}

	// Pre-assembly extension rename so that per-operation tagging below sees
	// the renamed extension rather than the original.
	r.SchemaTypeDef.RenameExtensionRecursive("x-speakeasy-param-readonly", "x-speakeasy-manual-param-readonly")

	// Merge each invoke operation's request shard.
	for _, op := range ops.Invoke {
		if op.RequestShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.RequestShard); err != nil {
				return fmt.Errorf("assembling action %s schema: merging invoke request shard: %w", r.Name, err)
			}
		}
	}

	// Merge alias nodes from the merged invoke request shard.
	if ops.InvokeRequestShard != nil {
		if err := r.SchemaTypeDef.TerraformMergeAliasNodes(r.Name, ops.InvokeRequestShard); err != nil {
			return fmt.Errorf("assembling action %s schema: merging invoke alias nodes: %w", r.Name, err)
		}
	}

	// Merge each invoke operation's response shard.
	for _, op := range ops.Invoke {
		if op.ResponseShard != nil {
			if err := r.SchemaTypeDef.TerraformMerge(op.ResponseShard); err != nil {
				return fmt.Errorf("assembling action %s schema: merging invoke response shard: %w", r.Name, err)
			}
		}
	}

	// Apply per-operation modifications.
	for _, op := range ops.Invoke {
		// Prevent request attributes from being marked readonly.
		r.SchemaTypeDef.SetExtensionOnEquivalentTypes(op.RequestShard, "x-speakeasy-param-readonly", false)

		// Hide response-only fields from schema.
		if op.ResponseShard != nil {
			r.SchemaTypeDef.SetTerraformIgnoreOnExclusiveTypes(op.ResponseShard, op.RequestShard, false, true)
		}
	}

	// DeepClone breaks all shared *TypeDef pointers introduced during
	// assembly. TerraformFinalizeSchema applies shared post-assembly
	// modifications (root extension, conflicts-with, propagation, trim, sort,
	// and extension rename) common to all entity types.
	r.SchemaTypeDef = r.SchemaTypeDef.DeepClone()
	r.SchemaTypeDef.TerraformFinalizeSchema(excludeEmptyObjectSchemas)

	// Cleanup per-operation shards by removing ignored fields.
	for _, op := range ops.Invoke {
		op.CleanupShards()
	}

	// Pre-compute SDK method data per operation. Must run after schema
	// assembly since it cross-references SchemaTypeDef.
	for _, op := range ops.All() {
		if err := op.setSDKMethodData(r.Name, r.SchemaTypeDef); err != nil {
			return fmt.Errorf("assembling action %s schema: %w", r.Name, err)
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
func (r *TerraformAction) SchemaDescription() string {
	if r.Description != "" {
		return r.Description
	}

	return TerraformGoTypeName(r.Name) + " Action"
}

// Returns true if TerraformAction is valid. A valid action must have at least
// one Invoke operation associated with it.
func (r *TerraformAction) isValid() bool {
	if r == nil || r.Operations == nil {
		return false
	}

	return len(r.Operations.invoke) > 0
}

// Sets the action description from the operation. Searches both the request
// and response for the first x-speakeasy-entity-description extension found.
func (r *TerraformAction) setDescriptionFromOperation(operation *Operation) {
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

			if typeDef.Extensions.EntityDescription.TerraformAction == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformAction
			return
		}
	}

	if operation.Response != nil && operation.Response.Type != nil {
		for _, typeDef := range operation.Response.Type.Walk() {
			if typeDef.Extensions == nil || typeDef.Extensions.EntityDescription == nil {
				continue
			}

			if typeDef.Extensions.EntityDescription.TerraformAction == "" {
				continue
			}

			r.Description = typeDef.Extensions.EntityDescription.TerraformAction
			return
		}
	}
}

// Sets the global fields from the operation.
func (r *TerraformAction) setGlobalFieldsFromOperation(operation *Operation) {
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
func (r *TerraformAction) setPaginationFieldsFromOperation(operation *Operation) {
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
func (r *TerraformAction) setOperationSecurityFromOperation(generationConfig map[string]any, operation *Operation) {
	if operation == nil || operation.Security == nil {
		return
	}

	enableOperationSecurity, ok := generationConfig["enableOperationSecurity"].(bool)

	if !ok || !enableOperationSecurity {
		return
	}

	if r.OperationSecurity != nil {
		// NOTE: This assumes that all operations for an action have the same
		// operation security configuration. If that requirement changes, this
		// logic should be updated to combine fields into a single FieldDef so
		// the templating logic for the schema and data model have a complete
		// view of all security fields.
		return
	}

	r.OperationSecurity = operation.Security
}

// Sets the Server from the operation.
func (r *TerraformAction) setServerFromOperation(generationConfig map[string]any, operation *Operation) {
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
