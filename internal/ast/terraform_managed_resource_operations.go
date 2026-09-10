package ast

import (
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes operations associated with a Terraform managed resource.
type TerraformManagedResourceOperations struct {
	// Ordered create operations. Populated by MergeOperationShards.
	Create TerraformOperations `json:"-" yaml:"-"`

	// CreateNeedsReadAfter indicates whether the read operation should be
	// invoked after create operations because the read response shard
	// contains entity fields not present in the create response shard.
	// Populated by MergeOperationShards.
	CreateNeedsReadAfter bool `json:"-" yaml:"-"`

	// Merged request shard across all create operations. Populated by
	// MergeOperationShards.
	CreateRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all create operations. Populated by
	// MergeOperationShards.
	CreateResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all create operations.
	// Populated by MergeOperationShards.
	CreateShard *TypeDef `json:"-" yaml:"-"`

	// Ordered delete operations. Populated by MergeOperationShards.
	Delete TerraformOperations `json:"-" yaml:"-"`

	// Merged request shard across all delete operations. Populated by
	// MergeOperationShards.
	DeleteRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all delete operations. Populated by
	// MergeOperationShards.
	DeleteResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all delete operations.
	// Populated by MergeOperationShards.
	DeleteShard *TypeDef `json:"-" yaml:"-"`

	// Ordered read operations. Populated by MergeOperationShards.
	Read TerraformOperations `json:"-" yaml:"-"`

	// Merged request shard across all read operations. Populated by
	// MergeOperationShards.
	ReadRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all read operations. Populated by
	// MergeOperationShards.
	ReadResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all read operations.
	// Populated by MergeOperationShards.
	ReadShard *TypeDef `json:"-" yaml:"-"`

	// Ordered update operations. Populated by MergeOperationShards.
	Update TerraformOperations `json:"-" yaml:"-"`

	// UpdateNeedsReadAfter indicates whether the read operation should be
	// invoked after update operations because the read response shard
	// contains entity fields not present in the update response shard.
	// Populated by MergeOperationShards.
	UpdateNeedsReadAfter bool `json:"-" yaml:"-"`

	// Merged request shard across all update operations. Populated by
	// MergeOperationShards.
	UpdateRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all update operations. Populated by
	// MergeOperationShards.
	UpdateResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all update operations.
	// Populated by MergeOperationShards.
	UpdateShard *TypeDef `json:"-" yaml:"-"`

	// Raw create operations keyed by order value. Populated by AddOperation.
	create map[int]*TerraformOperation

	// Raw delete operations keyed by order value. Populated by AddOperation.
	delete_ map[int]*TerraformOperation

	// Raw read operations keyed by order value. Populated by AddOperation.
	read map[int]*TerraformOperation

	// Raw update operations keyed by order value. Populated by AddOperation.
	update map[int]*TerraformOperation
}

// Creates a new Terraform managed resource operations, safely initializing
// underlying fields.
func NewTerraformManagedResourceOperations() *TerraformManagedResourceOperations {
	return &TerraformManagedResourceOperations{
		create:  make(map[int]*TerraformOperation),
		delete_: make(map[int]*TerraformOperation),
		read:    make(map[int]*TerraformOperation),
		update:  make(map[int]*TerraformOperation),
	}
}

func (o *TerraformManagedResourceOperations) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, op *Operation) error {
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
		switch entityOperationType {
		case string(terraform.ManagedResourceOperationTypeCreate):
			if o.create == nil {
				o.create = make(map[int]*TerraformOperation)
			}

			if _, ok := o.create[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.create[order].APIOperation.ID)
			}

			o.create[order] = terraformOp
		case string(terraform.ManagedResourceOperationTypeDelete):
			if o.delete_ == nil {
				o.delete_ = make(map[int]*TerraformOperation)
			}

			if _, ok := o.delete_[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.delete_[order].APIOperation.ID)
			}

			o.delete_[order] = terraformOp
		case string(terraform.ManagedResourceOperationTypeRead):
			if o.read == nil {
				o.read = make(map[int]*TerraformOperation)
			}

			if _, ok := o.read[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.read[order].APIOperation.ID)
			}

			o.read[order] = terraformOp
		case string(terraform.ManagedResourceOperationTypeUpdate):
			if o.update == nil {
				o.update = make(map[int]*TerraformOperation)
			}

			if _, ok := o.update[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.update[order].APIOperation.ID)
			}

			o.update[order] = terraformOp
		}
	}

	return nil
}

// MergeOperationShards computes and stores the merged shards for each operation
// type. This should be called after all operations have been added and any
// per-operation shard mutations (e.g. tagging rules) have been applied.
//
// The entityName is used to compute CreateNeedsReadAfter and
// UpdateNeedsReadAfter by checking whether the read shard is a structural
// subset of the create/update shards.
func (o *TerraformManagedResourceOperations) MergeOperationShards(entityName string) error {
	var err error

	o.Create = orderedTerraformOperations(o.create)
	o.Delete = orderedTerraformOperations(o.delete_)
	o.Read = orderedTerraformOperations(o.read)
	o.Update = orderedTerraformOperations(o.update)

	if len(o.Delete) > 0 {
		o.Delete[len(o.Delete)-1].skipDataModelRefresh = true
	}

	o.CreateRequestShard, o.CreateResponseShard, o.CreateShard, err = o.Create.Shards()
	if err != nil {
		return fmt.Errorf("merging create operation shards: %w", err)
	}

	o.DeleteRequestShard, o.DeleteResponseShard, o.DeleteShard, err = o.Delete.Shards()
	if err != nil {
		return fmt.Errorf("merging delete operation shards: %w", err)
	}

	o.ReadRequestShard, o.ReadResponseShard, o.ReadShard, err = o.Read.Shards()
	if err != nil {
		return fmt.Errorf("merging read operation shards: %w", err)
	}

	o.UpdateRequestShard, o.UpdateResponseShard, o.UpdateShard, err = o.Update.Shards()
	if err != nil {
		return fmt.Errorf("merging update operation shards: %w", err)
	}

	// Compute whether read operations need to follow create/update operations
	// to capture fields not present in those operation's response shards.
	if o.ReadShard != nil {
		if o.CreateShard != nil {
			o.CreateNeedsReadAfter = !o.ReadShard.IsTerraformSubsetOfEntity(entityName, o.CreateShard)
		}

		if o.UpdateShard != nil {
			o.UpdateNeedsReadAfter = !o.ReadShard.IsTerraformSubsetOfEntity(entityName, o.UpdateShard)
		}
	}

	return nil
}

// All returns all operations. For managed resources this is the create, read,
// update, and delete operations.
func (o *TerraformManagedResourceOperations) All() TerraformOperations {
	return slices.Concat(o.Create, o.Read, o.Update, o.Delete)
}

// DataModelRefreshOperations returns operations whose API responses are mapped
// back into the Terraform data model. For managed resources this is the read,
// create, and update operations. Delete operations are excluded because their
// response types are covered by read operations.
//
// Read operations are listed first to ensure pagination-aware RefreshFrom
// methods take priority during method name deduplication.
func (o *TerraformManagedResourceOperations) DataModelRefreshOperations() TerraformOperations {
	return slices.Concat(o.Read, o.Create, o.Update)
}

// Validate checks that the operations are valid for schema assembly. This
// should be called before AssembleSchemaTypeDef to surface errors early.
func (o *TerraformManagedResourceOperations) Validate() error {
	for _, op := range o.Read {
		if op.ResponseShard == nil {
			return fmt.Errorf("could not find or infer x-speakeasy-entity: %s type definition within x-speakeasy-entity-operation: %s", op.EntityName, op.EntityOperation)
		}
	}

	return nil
}
