package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes operations associated with a Terraform action.
type TerraformActionOperations struct {
	// Ordered invoke operations. Populated by MergeOperationShards.
	Invoke TerraformOperations `json:"-" yaml:"-"`

	// Merged request shard across all invoke operations. Populated by
	// MergeOperationShards.
	InvokeRequestShard *TypeDef `json:"-" yaml:"-"`

	// Merged response shard across all invoke operations. Populated by
	// MergeOperationShards.
	InvokeResponseShard *TypeDef `json:"-" yaml:"-"`

	// Merged request and response shard across all invoke operations.
	// Populated by MergeOperationShards.
	InvokeShard *TypeDef `json:"-" yaml:"-"`

	// Raw invoke operations keyed by order value. Populated by AddOperation.
	invoke map[int]*TerraformOperation
}

// Creates a new Terraform action operations, safely initializing
// underlying fields.
func NewTerraformActionOperations() *TerraformActionOperations {
	return &TerraformActionOperations{
		invoke: make(map[int]*TerraformOperation),
	}
}

func (o *TerraformActionOperations) AddOperation(generationConfig map[string]any, entityOperationConfig extensions.EntityOperationV1Config, op *Operation) error {
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
		if entityOperationType == string(terraform.ActionOperationTypeInvoke) {
			if o.invoke == nil {
				o.invoke = make(map[int]*TerraformOperation)
			}

			if _, ok := o.invoke[order]; ok {
				return fmt.Errorf("duplicate entity operation %s in %s and %s", entityOperationConfig, op.ID, o.invoke[order].APIOperation.ID)
			}

			o.invoke[order] = terraformOp
		}
	}

	return nil
}

// MergeOperationShards computes and stores the merged shards for each operation
// type. This should be called after all operations have been added and any
// per-operation shard mutations (e.g. tagging rules) have been applied.
func (o *TerraformActionOperations) MergeOperationShards() error {
	var err error

	o.Invoke = orderedTerraformOperations(o.invoke)

	if len(o.Invoke) > 0 {
		o.Invoke[len(o.Invoke)-1].skipDataModelRefresh = true
	}

	o.InvokeRequestShard, o.InvokeResponseShard, o.InvokeShard, err = o.Invoke.Shards()
	if err != nil {
		return fmt.Errorf("merging invoke operation shards: %w", err)
	}

	return nil
}

// All returns all operations. For actions this is the invoke operations.
func (o *TerraformActionOperations) All() TerraformOperations {
	return o.Invoke
}

// DataModelRefreshOperations returns operations whose API responses are mapped
// back into the Terraform data model. For actions this is the invoke
// operations.
//
// Read operations are listed first across all resource types to ensure
// pagination-aware RefreshFrom methods take priority during method name
// deduplication.
func (o *TerraformActionOperations) DataModelRefreshOperations() TerraformOperations {
	return o.Invoke
}

// Validate checks that the operations are valid for schema assembly. This
// should be called before AssembleSchemaTypeDef to surface errors early.
func (o *TerraformActionOperations) Validate() error {
	return nil
}
