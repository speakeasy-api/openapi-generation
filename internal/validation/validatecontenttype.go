package validation

import (
	"context"
	"errors"

	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateContentType struct{}

var _ Rule = (*ValidateContentType)(nil)

func (r *ValidateContentType) ID() string {
	return "generator-validate-content-type"
}

func (r *ValidateContentType) Category() string {
	return "validation"
}

func (r *ValidateContentType) Summary() string {
	return "Validate multipart schemas are objects for form encoding."
}

func (r *ValidateContentType) HowToFix() string {
	return "Use object (or array of object) schemas for multipart/form-data and avoid oneOf/anyOf in multipart schemas."
}

func (r *ValidateContentType) Description() string {
	return "Ensure multipart/form-data request schemas are objects as required by OpenAPI. Non-object schemas break form encoding and SDK generation."
}

func (r *ValidateContentType) Link() string {
	return ""
}

func (r *ValidateContentType) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateContentType) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateContentType) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Use the Index.MediaTypes which contains all media types from request bodies and responses
	for _, mediaTypeNode := range docInfo.Index.MediaTypes {
		if mediaTypeNode == nil || mediaTypeNode.Node == nil {
			continue
		}

		mediaType := mediaTypeNode.Node

		// Get the content type key
		contentTypeKey := mediaTypeNode.Location.ParentKey()
		if contentTypeKey == "" {
			continue
		}

		// Only validate multipart content types
		if !contenttypes.IsMultipart(contentTypeKey) {
			continue
		}

		schemaRef := mediaType.GetSchema()
		if schemaRef == nil {
			continue
		}

		// Use GetResolvedSchema to resolve $ref references
		resolvedSchemaRef := schemaRef.GetResolvedSchema()
		if resolvedSchemaRef == nil {
			continue
		}

		schema := resolvedSchemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Get the Schema property key node for reporting on empty schemas
		schemaKeyNode := mediaType.GetPropertyNode("Schema")

		// Validate the schema recursively, using the schema key node for error reporting
		errors := r.validateMultipartSchema(schema, schemaKeyNode)
		validationErrors = append(validationErrors, errors...)
	}

	return validationErrors
}

func (r *ValidateContentType) validateMultipartSchema(schema *oas3.Schema, parentKeyNode *yaml.Node) []error {
	if schema == nil {
		return nil
	}

	var validationErrors []error

	// Get the schema node for error reporting
	// Use schema's root node, or parentKeyNode as fallback
	schemaNode := schema.GetRootNode()
	if schemaNode == nil && parentKeyNode != nil {
		schemaNode = parentKeyNode
	}

	// Check for oneOf or anyOf - these should be warnings
	// Use parentKeyNode for consistency with old implementation
	errorNode := parentKeyNode
	if errorNode == nil {
		errorNode = schemaNode
	}

	if len(schema.GetOneOf()) > 0 {
		return []error{
			&validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityWarning,
				Node:            errorNode,
				UnderlyingError: errors.New("multipart schema must not contain `oneOf`"),
			},
		}
	}

	if len(schema.GetAnyOf()) > 0 {
		return []error{
			&validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityWarning,
				Node:            errorNode,
				UnderlyingError: errors.New("multipart schema must not contain `anyOf`"),
			},
		}
	}

	// Handle allOf - recursively validate each sub-schema
	allOfSchemas := schema.GetAllOf()
	if len(allOfSchemas) > 0 {
		for _, subSchemaRef := range allOfSchemas {
			// Resolve the sub-schema reference
			resolvedSubSchemaRef := subSchemaRef.GetResolvedSchema()
			if resolvedSubSchemaRef == nil {
				continue
			}
			subSchema := resolvedSubSchemaRef.GetSchema()
			if subSchema == nil {
				continue
			}

			validationErrors = append(validationErrors, r.validateMultipartSchema(subSchema, nil)...)
		}
		return validationErrors
	}

	// Get schema types
	types := schema.GetType()

	// Check for array type with object items
	if len(types) > 0 && string(types[0]) == "array" {
		items := schema.GetItems()
		if items != nil {
			// Resolve the items reference
			resolvedItems := items.GetResolvedSchema()
			if resolvedItems == nil {
				return nil
			}
			itemsSchema := resolvedItems.GetSchema()
			if itemsSchema != nil {
				return r.validateMultipartSchema(itemsSchema, nil)
			}
		}
	}

	// Check for explicit object type
	if len(types) > 0 {
		if string(types[0]) == "object" {
			return nil // Valid
		}
		// Invalid type for multipart - report on schema key node for consistency
		node := parentKeyNode
		if node == nil {
			node = schemaNode
		}
		return []error{
			&validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            node,
				UnderlyingError: errors.New("multipart schema must be an `object`"),
			},
		}
	}

	// Check for implicit object type (properties, additionalProperties, patternProperties)
	if schema.GetProperties() != nil && schema.GetProperties().Len() > 0 {
		return nil // Has properties, implicitly an object
	}

	if schema.GetAdditionalProperties() != nil {
		return nil // Has additionalProperties, implicitly an object
	}

	if schema.GetPatternProperties() != nil && schema.GetPatternProperties().Len() > 0 {
		return nil // Has patternProperties, implicitly an object
	}

	// No type and no implicit object indicators - this is invalid
	// Report on the schema key node if available (for empty schemas)
	node := parentKeyNode
	if node == nil {
		node = schemaNode
	}
	return []error{
		&validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            node,
			UnderlyingError: errors.New("multipart schema must be an `object`"),
		},
	}
}
