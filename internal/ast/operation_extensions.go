package ast

import "github.com/speakeasy-api/openapi-generation/v2/internal/extensions"

// OperationExtensions represents the extensions that operate on an operation
type OperationExtensions struct {
	MethodNameOverride        string                                `yaml:",omitempty"`
	UsageExample              *extensions.UsageExampleConfig        `yaml:",omitempty"`
	Retries                   *extensions.Retries                   `yaml:",omitempty"`
	Timeout                   *int64                                `yaml:",omitempty"`
	Pagination                *extensions.Pagination                `yaml:",omitempty"`
	DocsRateLimits            []extensions.RateLimit                `yaml:",omitempty"`
	ReactHook                 *extensions.ReactHook                 `yaml:",omitempty"`
	MCP                       *extensions.MCP                       `yaml:",omitempty"`
	SSEOverload               *extensions.SSEOverloadConfig         `yaml:",omitempty"`
	PublicExports             []extensions.PublicExport             `yaml:",omitempty"`
	GoOptionalMethodArguments *extensions.GoOptionalMethodArguments `yaml:",omitempty"`
	All                       map[string]any                        `yaml:",omitempty"`

	// Describes x-speakeasy-entity-operation extension configuration. The extension
	// is typically configured along with the x-speakeasy-entity extension to build
	// a single model to describe an API entity. That model is used to describe a
	// Terraform data or managed resource currently, but may represent other target
	// entities in the future.
	//
	// This extension accepts the following data types:
	//
	// - String: A value in the form of Entity#OpType[,OpType...][#Order].
	//   Entity is the name of the entity, OpType is the set of entity operation
	//   types (typically entity lifecycle operations such as "create", "read",
	//   "update", and "delete"), and Order is an optional integer greater than 0
	//   that specifies the order of the operation compared to other definitions of
	//   x-speakeasy-entity-operation with the same Entity#OpType.
	// - Array of string: Multiple values of the above string form when a single API
	//   operation is necessary across multiple entities, such as ["Entity#OpType",
	//   "Entity2#OpType"].
	// - Object: Target-specific configuration for the entity operation, where the
	//   properties are target-defined entity sub-types. This is typically used to
	//   disable automatically generated entity sub-types, should the target create
	//   multiple entity sub-types for the same entity operation, or other advanced
	//   configuration of individual entity sub-types. For example, the Terraform
	//   target automatically generates both a data and managed resource for each
	//   Entity#read associated with another Entity#create, so the object form can
	//   be used to disable data resource generation by specifying
	//   {"terraform-datasource": null, "terraform-resource": "Entity#read"}. Each
	//   property value accepts the above string and array of string forms.
	//   The supported properties are:
	//     - "terraform-datasource": Explicit configuration for Terraform target
	//       data resources.
	//     - "terraform-resource": Explicit configuration for Terraform target
	//       managed resources.
	//
	//
	// Only one of the EntityOperation (x-speakeasy-entity-operation) or
	// EntityOperations (x-speakeasy-entity-operations) extensions is valid in a
	// single API operation.
	EntityOperation *extensions.EntityOperationV1 `yaml:",omitempty"`

	// Describes polling configuration, derived from the x-speakeasy-polling
	// extension configuration. This configuration enables targets implementing
	// the operationPolling generator feature to template operation polling
	// logic for consumer opt-in (or in the case of the terraform target,
	// implement via x-speakeasy-entity-operation).
	Polling *Polling `yaml:",omitempty"`

	// Describes HTTP status codes that indicate an entity is missing/deleted.
	// Derived from the x-speakeasy-entity-missing-codes extension. Used by the
	// Terraform target to call RemoveResource() during Read operations when the
	// API returns one of these status codes.
	EntityMissingCodes extensions.EntityMissingCodes `yaml:",omitempty"`
}

func (o *OperationExtensions) Match(matchers Matchers) error {
	if matchers.OperationExtensions != nil {
		return matchers.OperationExtensions(o)
	}

	return nil
}
