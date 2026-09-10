package validation

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateCompositeSchemas struct{}

var _ Rule = (*ValidateCompositeSchemas)(nil)

func (r *ValidateCompositeSchemas) ID() string {
	return "generator-validate-composite-schemas"
}

func (r *ValidateCompositeSchemas) Category() string {
	return "validation"
}

func (r *ValidateCompositeSchemas) Summary() string {
	return "Validate composite schemas avoid duplicate refs or identical schemas."
}

func (r *ValidateCompositeSchemas) HowToFix() string {
	return "Remove duplicate references or identical schemas from anyOf/oneOf/allOf, and avoid single-item composites that only contain type: null."
}

func (r *ValidateCompositeSchemas) Description() string {
	return "Ensure anyOf/allOf/oneOf don't contain duplicate references or identical inline schemas. Duplicates make schemas ambiguous and can lead to conflicting generated types."
}

func (r *ValidateCompositeSchemas) Link() string {
	return ""
}

func (r *ValidateCompositeSchemas) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateCompositeSchemas) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateCompositeSchemas) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Iterate through all schemas (including inline ones)
	allSchemas := docInfo.Index.GetAllSchemas()
	for _, indexNode := range allSchemas {
		if indexNode == nil || indexNode.Node == nil {
			continue
		}

		schemaRef := indexNode.Node
		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Check anyOf, allOf, oneOf
		validationErrors = append(validationErrors, r.validateCompositeType(schema, schema.GetAnyOf(), "anyOf")...)
		validationErrors = append(validationErrors, r.validateCompositeType(schema, schema.GetAllOf(), "allOf")...)
		validationErrors = append(validationErrors, r.validateCompositeType(schema, schema.GetOneOf(), "oneOf")...)
	}

	return validationErrors
}

func (r *ValidateCompositeSchemas) validateCompositeType(parentSchema *oas3.Schema, compositeSchemas []*oas3.JSONSchemaReferenceable, nodeType string) []error {
	if compositeSchemas == nil {
		return nil
	}

	var validationErrors []error

	// Get the node for the composite type array
	compositeNode := r.getCompositeNode(parentSchema, nodeType)
	if compositeNode == nil {
		return nil
	}

	if len(compositeSchemas) == 0 {
		return nil
	}

	// Check for single item that is only "type: null"
	if len(compositeSchemas) == 1 {
		firstSchema := compositeSchemas[0].GetSchema()
		if firstSchema != nil {
			types := firstSchema.GetType()
			if len(types) == 1 && string(types[0]) == "null" {
				// Get the node for the type field
				typeNode := r.getTypeNode(firstSchema)
				if typeNode == nil {
					typeNode = compositeNode
				}
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        validation.SeverityWarning,
					Node:            typeNode,
					UnderlyingError: fmt.Errorf("`%s` cannot contain `type: null` only", nodeType),
				})
			}
		}
		return validationErrors
	}

	// Check for duplicate references
	refs := make(map[string]int)
	for i, schemaRef := range compositeSchemas {
		if schemaRef.IsReference() {
			refValue := schemaRef.GetReference().String()
			if prevIndex, exists := refs[refValue]; exists {
				// Get the node for the duplicate reference
				refNode := schemaRef.GetRootNode()
				if refNode == nil {
					refNode = compositeNode
				}
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            refNode,
					UnderlyingError: fmt.Errorf("duplicate schema reference found in `%s` (first occurrence at index `%d`, duplicate at index `%d`)", nodeType, prevIndex, i),
				})
			} else {
				refs[refValue] = i
			}
		}
	}

	return validationErrors
}

// getCompositeNode finds the YAML node for the composite type (anyOf/allOf/oneOf)
func (r *ValidateCompositeSchemas) getCompositeNode(schema *oas3.Schema, nodeType string) *yaml.Node {
	rootNode := schema.GetRootNode()
	if rootNode == nil || rootNode.Content == nil {
		return nil
	}

	// Search for the nodeType key in the schema node
	for i := 0; i < len(rootNode.Content)-1; i += 2 {
		keyNode := rootNode.Content[i]
		if keyNode.Value == nodeType {
			return rootNode.Content[i+1]
		}
	}

	return nil
}

// getTypeNode finds the YAML node for the "type" field
func (r *ValidateCompositeSchemas) getTypeNode(schema *oas3.Schema) *yaml.Node {
	rootNode := schema.GetRootNode()
	if rootNode == nil || rootNode.Content == nil {
		return nil
	}

	// Search for the "type" key in the schema node
	for i := 0; i < len(rootNode.Content)-1; i += 2 {
		keyNode := rootNode.Content[i]
		if keyNode.Value == "type" {
			return keyNode
		}
	}

	return nil
}
