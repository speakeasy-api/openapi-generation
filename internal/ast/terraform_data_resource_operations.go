package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes operations associated with a Terraform data resource.
type TerraformDataResourceOperations struct {
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

	// Raw read operations keyed by order value. Populated by AddOperation.
	read map[int]*TerraformOperation
}

// Creates a new Terraform data resource operations, safely initializing
// underlying fields.
func NewTerraformDataResourceOperations() *TerraformDataResourceOperations {
	return &TerraformDataResourceOperations{
		read: make(map[int]*TerraformOperation),
	}
}

func (o *TerraformDataResourceOperations) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, op *Operation) error {
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
		if entityOperationType == string(terraform.DataResourceOperationTypeRead) {
			if o.read == nil {
				o.read = make(map[int]*TerraformOperation)
			}

			if _, ok := o.read[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.read[order].APIOperation.ID)
			}

			o.read[order] = terraformOp
		}
	}

	return nil
}

// MergeOperationShards computes and stores the merged shards for each operation
// type. This should be called after all operations have been added and any
// per-operation shard mutations (e.g. tagging rules) have been applied.
func (o *TerraformDataResourceOperations) MergeOperationShards() error {
	var err error

	o.Read = orderedTerraformOperations(o.read)

	o.ReadRequestShard, o.ReadResponseShard, o.ReadShard, err = o.Read.Shards()
	if err != nil {
		return fmt.Errorf("merging read operation shards: %w", err)
	}

	return nil
}

// All returns all operations. For data resources this is the read operations.
func (o *TerraformDataResourceOperations) All() TerraformOperations {
	return o.Read
}

// DataModelRefreshOperations returns operations whose API responses are mapped
// back into the Terraform data model. For data resources this is the read
// operations.
//
// Read operations are listed first across all resource types to ensure
// pagination-aware RefreshFrom methods take priority during method name
// deduplication.
func (o *TerraformDataResourceOperations) DataModelRefreshOperations() TerraformOperations {
	return o.Read
}

// Validate checks that the operations are valid for schema assembly. This
// should be called before AssembleSchemaTypeDef to surface errors early.
func (o *TerraformDataResourceOperations) Validate() error {
	for _, op := range o.Read {
		if op.ResponseShard == nil {
			return fmt.Errorf("could not find or infer x-speakeasy-entity: %s type definition within x-speakeasy-entity-operation: %s", op.EntityName, op.EntityOperation)
		}
	}

	return nil
}
