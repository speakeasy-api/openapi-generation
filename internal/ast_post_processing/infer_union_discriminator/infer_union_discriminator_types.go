package infer_union_discriminator

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// DiscriminatorSpecificity represents how specific the discriminator values are
// Lower values are more specific (better)
type DiscriminatorSpecificity int

const (
	// DiscriminatorSpecificityConst indicates discrimination by const values
	// Example: type: "dog" (const) vs type: "cat" (const)
	DiscriminatorSpecificityConst DiscriminatorSpecificity = 1
	// DiscriminatorSpecificitySingle indicates discrimination by single-value enums or const (mixed)
	// Example: type: enum["dog"] vs type: "cat" (const)
	DiscriminatorSpecificitySingle DiscriminatorSpecificity = 2
	// DiscriminatorSpecificityMulti indicates discrimination by multi-value enums
	// Example: type: enum["dog", "puppy"] vs type: enum["cat", "kitten"]
	DiscriminatorSpecificityMulti DiscriminatorSpecificity = 3
	// DiscriminatorSpecificityType indicates discrimination by primitive type only
	// Example: id: string vs id: number
	DiscriminatorSpecificityType DiscriminatorSpecificity = 4
)

// discriminator contains information about a discriminator property
type discriminator struct {
	PropertyName string
	TypeDefType  ast.DataType
	Specificity  DiscriminatorSpecificity
	Mapping      []*discriminatorMapping
}

// discriminatorMapping represents a mapping from discriminator value(s) to a type
type discriminatorMapping struct {
	Consts []*ast.AnyValue
	Type   *ast.TypeDef
}
