package schemas

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

// CreateTestParams creates test parameters for schema handling
func CreateTestParams(schema *oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo) Params {
	return Params{
		ContextStack:        ast.ContextStack{},
		SerializationMethod: ast.SerializationMethodJSON,
		Schema:              schema,
		Scope:               ast.ScopeShared,
		IsRequest:           false,
		Depth:               0,
		MaxDepth:            10,
		Parents:             []string{},
		TypeDefCache:        make(map[string]*ast.TypeDef),
		LoopContext:         []LoopFrame{},
		Nullable:            false,
		CircularReference:   false,
		DocInfo:             docInfo,
	}
}

// CreateTestParamsWithSerialization creates test parameters with a specific serialization method
func CreateTestParamsWithSerialization(schema *oas3.JSONSchema[oas3.Referenceable], serializationMethod ast.SerializationMethod, docInfo *document.DocumentInfo) Params {
	return Params{
		ContextStack:        ast.ContextStack{},
		SerializationMethod: serializationMethod,
		Schema:              schema,
		Encoding:            sequencedmap.New[string, *oas.Encoding](), // Initialize empty encoding map
		Scope:               ast.ScopeShared,
		IsRequest:           false,
		Depth:               0,
		MaxDepth:            10,
		Parents:             []string{},
		TypeDefCache:        make(map[string]*ast.TypeDef),
		LoopContext:         []LoopFrame{},
		Nullable:            false,
		CircularReference:   false,
		DocInfo:             docInfo,
	}
}
