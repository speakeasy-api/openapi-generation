package ast

import (
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes operations associated with a Terraform ephemeral resource.
type TerraformEphemeralResourceOperations struct {
	// Ordered close operations. Populated by MergeOperationShards.
	Close TerraformOperations `json:"-" yaml:"-"`

	// Merged request shard across all close operations. Populated by
	// MergeOperationShards.
	CloseRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all close operations. Populated by
	// MergeOperationShards.
	CloseResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all close operations.
	// Populated by MergeOperationShards.
	CloseShard *TypeDef `json:"-" yaml:"-"`

	// Ordered open operations. Populated by MergeOperationShards.
	Open TerraformOperations `json:"-" yaml:"-"`

	// Merged request shard across all open operations. Populated by
	// MergeOperationShards.
	OpenRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all open operations. Populated by
	// MergeOperationShards.
	OpenResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all open operations.
	// Populated by MergeOperationShards.
	OpenShard *TypeDef `json:"-" yaml:"-"`

	// Raw close operations keyed by order value. Populated by AddOperation.
	close_ map[int]*TerraformOperation

	// Raw open operations keyed by order value. Populated by AddOperation.
	open map[int]*TerraformOperation
}

// Creates a new Terraform ephemeral resource operations, safely initializing
// underlying fields.
func NewTerraformEphemeralResourceOperations() *TerraformEphemeralResourceOperations {
	return &TerraformEphemeralResourceOperations{
		close_: make(map[int]*TerraformOperation),
		open:   make(map[int]*TerraformOperation),
	}
}

func (o *TerraformEphemeralResourceOperations) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, op *Operation) error {
	if op.Request != nil {
		typeDef := op.Request.FindEntityTypeDef(entityOperationConfig.Entity)

		if typeDef == nil && op.Request.RequestBody != nil {
			typeDef = op.Request.RequestBody.Type
		}

		if typeDef != nil {
			switch typeDef.Type {
			case DataTypeClass, DataTypeUnion:
				break
			default:
				return fmt.Errorf("%s extension is only supported on object or union types, got %s type for entity operation %s", extensions.ExtEntity.Name(), typeDef.Type, entityOperationConfig)
			}

			if typeDef.Extensions.Entity == nil {
				typeDef.Extensions.Entity = extensions.NewEntity()
			}

			typeDef.Extensions.Entity.AddNames(entityOperationConfig.Entity)
		}
	}

	if op.Response != nil {
		fieldDef := op.Response.TerraformBodyFieldDef(entityOperationConfig.Entity)

		if fieldDef != nil && fieldDef.FindEntityFieldDef(entityOperationConfig.Entity) == nil {
			if fieldDef.Type.Extensions.Entity == nil {
				fieldDef.Type.Extensions.Entity = extensions.NewEntity()
			}

			fieldDef.Type.Extensions.Entity.AddNames(entityOperationConfig.Entity)
		}
	}

	// TODO: Coalesce above logic into NewTerraformOperation and only operate on
	// cloned TypeDef.
	terraformOp, err := NewTerraformOperation(op, entityOperationConfig)

	if err != nil {
		return fmt.Errorf("failed to create TerraformOperation for entity operation %s: %w", entityOperationConfig, err)
	}

	terraformOp.setIncludeOperationSecurity(generationConfig)
	terraformOp.setServerAttributeName(generationConfig)

	var order int

	if entityOperationConfig.Order != nil {
		order = *entityOperationConfig.Order
	}

	for _, entityOperationType := range entityOperationConfig.OperationTypes {
		if entityOperationType == string(terraform.EphemeralResourceOperationTypeClose) {
			if o.close_ == nil {
				o.close_ = make(map[int]*TerraformOperation)
			}

			if _, ok := o.close_[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.close_[order].APIOperation.ID)
			}

			o.close_[order] = terraformOp
		}

		if entityOperationType == string(terraform.EphemeralResourceOperationTypeOpen) {
			if o.open == nil {
				o.open = make(map[int]*TerraformOperation)
			}

			if _, ok := o.open[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.open[order].APIOperation.ID)
			}

			o.open[order] = terraformOp
		}
	}

	return nil
}

// MergeOperationShards computes and stores the merged shards for each operation
// type. This should be called after all operations have been added and any
// per-operation shard mutations (e.g. tagging rules) have been applied.
func (o *TerraformEphemeralResourceOperations) MergeOperationShards() error {
	var err error

	o.Close = orderedTerraformOperations(o.close_)
	o.Open = orderedTerraformOperations(o.open)

	for i := range o.Close {
		o.Close[i].skipDataModelRefresh = true
	}

	o.CloseRequestShard, o.CloseResponseShard, o.CloseShard, err = o.Close.Shards()
	if err != nil {
		return fmt.Errorf("merging close operation shards: %w", err)
	}

	o.OpenRequestShard, o.OpenResponseShard, o.OpenShard, err = o.Open.Shards()
	if err != nil {
		return fmt.Errorf("merging open operation shards: %w", err)
	}

	return nil
}

// All returns all operations. For ephemeral resources this is the open and
// close operations.
func (o *TerraformEphemeralResourceOperations) All() TerraformOperations {
	return slices.Concat(o.Open, o.Close)
}

// DataModelRefreshOperations returns operations whose API responses are mapped
// back into the Terraform data model. For ephemeral resources this is the open
// operations. Close operations are excluded because their responses do not
// carry entity data for the data model.
//
// Read operations are listed first across all resource types to ensure
// pagination-aware RefreshFrom methods take priority during method name
// deduplication.
func (o *TerraformEphemeralResourceOperations) DataModelRefreshOperations() TerraformOperations {
	return o.Open
}

// Validate checks that the operations are valid for schema assembly. This
// should be called before AssembleSchemaTypeDef to surface errors early.
func (o *TerraformEphemeralResourceOperations) Validate() error {
	return nil
}
