package validation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi/hashing"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

type DuplicateInlineSchemas struct{}

var _ Rule = (*DuplicateInlineSchemas)(nil)

func (r *DuplicateInlineSchemas) ID() string {
	return "generator-duplicate-inline-schemas"
}

func (r *DuplicateInlineSchemas) Category() string {
	return "validation"
}

func (r *DuplicateInlineSchemas) Summary() string {
	return "Ensure no duplicate inline schemas generate conflicting types."
}

func (r *DuplicateInlineSchemas) HowToFix() string {
	return "Extract identical inline schemas into shared components or change their content/titles so they do not generate duplicate types."
}

func (r *DuplicateInlineSchemas) Description() string {
	return "Detect duplicate inline schemas that would produce multiple generated SDK types. When their generated names collide, models become ambiguous or overwritten."
}

func (r *DuplicateInlineSchemas) Link() string {
	return ""
}

func (r *DuplicateInlineSchemas) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *DuplicateInlineSchemas) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateInlineSchemas) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	schemaNodes := []*openapi.IndexNode[*oas3.JSONSchemaReferenceable]{}

	// Iterate through all inline schemas
	for _, indexNode := range docInfo.Index.InlineSchemas {
		schemaRef := indexNode.Node
		if schemaRef == nil {
			continue
		}

		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Check if this is an object or enum type
		types := schema.GetType()
		isObject := false
		for _, t := range types {
			if string(t) == "object" {
				isObject = true
				break
			}
		}

		enumValues := schema.GetEnum()
		isEnum := len(enumValues) > 0

		// TODO: Handle anyOf, oneOf, allOf
		if !isObject && !isEnum {
			continue
		}

		schemaNodes = append(schemaNodes, indexNode)
	}

	schemasByHash := r.groupSchemasByHash(schemaNodes)

	validationErrors := make([]error, 0, len(schemasByHash))
	for _, schemas := range schemasByHash {
		description := r.getSchemaGroupDescription(schemas)
		if description.Description == "" {
			continue
		}

		lineNumbers := []int{}
		for _, indexNode := range schemas {
			node := indexNode.Node.GetSchema().GetRootNode()
			if node != nil && node.Line > 0 {
				lineNumbers = append(lineNumbers, node.Line)
			}
		}

		// Sort line numbers in ascending order
		sort.Ints(lineNumbers)

		// Deduplicate line numbers
		uniqueLineNumbers := []int{}
		for i, lineNum := range lineNumbers {
			if i == 0 || lineNumbers[i-1] != lineNum {
				uniqueLineNumbers = append(uniqueLineNumbers, lineNum)
			}
		}

		// Skip if all schemas are at the same location (not real duplicates)
		if len(uniqueLineNumbers) < 2 {
			continue
		}

		lines := []string{}
		for _, lineNum := range uniqueLineNumbers {
			lines = append(lines, strconv.Itoa(lineNum))
		}

		message := fmt.Sprintf("`%d` duplicates of `%s` %q at lines [%s]", len(uniqueLineNumbers), description.Type, description.Description, strings.Join(lines, ","))

		first := schemas[0]
		firstNode := first.Node.GetSchema().GetRootNode()

		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            firstNode,
			UnderlyingError: errors.New(message),
		})
	}

	return validationErrors
}

func (r *DuplicateInlineSchemas) groupSchemasByHash(nodes []*openapi.IndexNode[*oas3.JSONSchemaReferenceable]) map[string][]*openapi.IndexNode[*oas3.JSONSchemaReferenceable] {
	schemaNodesByContent := make(map[string][]*openapi.IndexNode[*oas3.JSONSchemaReferenceable])

	for _, indexNode := range nodes {
		schema := indexNode.Node.GetSchema()
		if schema == nil {
			continue
		}

		hash := hashing.Hash(schema.GetRootNode())
		schemaNodesByContent[hash] = append(schemaNodesByContent[hash], indexNode)
	}

	return schemaNodesByContent
}

func (r *DuplicateInlineSchemas) getSchemaGroupDescription(nodes []*openapi.IndexNode[*oas3.JSONSchemaReferenceable]) schemaDescription {
	if len(nodes) < 2 {
		return schemaDescription{}
	}

	first := nodes[0]
	schema := first.Node.GetSchema()
	if schema == nil {
		return schemaDescription{}
	}

	return r.getSchemaDescription(schema)
}

type schemaDescription struct {
	Type        string
	Description string
}

func (r *DuplicateInlineSchemas) getSchemaDescription(schema *oas3.Schema) schemaDescription {
	// Check for title first
	title := schema.GetTitle()
	if title != "" {
		// Max length of title is 64 characters
		t := strings.TrimSpace(title)
		if len(t) > 70 {
			t = t[:70] + "..."
		}
		return schemaDescription{Type: "schema", Description: t}
	}

	// TODO: Handle OneOf, AnyOf, AllOf, ItemType

	// Check for enum
	enumValues := schema.GetEnum()
	if len(enumValues) > 1 {
		// Get the first 3 enum values and append "..." if there are more
		values := []string{}
		for i := 0; i < 3 && i < len(enumValues); i++ {
			// Access the Value field of the yaml.Node to get the actual value string
			v := enumValues[i].Value
			if v == "" {
				v = "''"
			}
			values = append(values, v)
		}

		description := strings.Join(values, ", ")

		if len(enumValues) > 3 {
			description += fmt.Sprintf("... `%d` more", len(enumValues)-3)
		}

		return schemaDescription{
			Type:        "enum",
			Description: description,
		}
	}

	// Check for object properties
	properties := schema.GetProperties()
	if properties == nil {
		return schemaDescription{}
	}

	parts := []string{}
	countProperties := 0

	for propName, propRef := range properties.All() {
		countProperties++
		if len(parts) == 3 {
			continue
		}

		propSchema := propRef.GetSchema()
		if propSchema == nil {
			continue
		}

		types := propSchema.GetType()
		if len(types) > 0 {
			parts = append(parts, fmt.Sprintf("`%s`: `%s`", propName, types[0]))
		}
	}

	if len(parts) < 2 {
		return schemaDescription{}
	}

	remainingProperties := countProperties - len(parts)

	description := strings.Join(parts, ", ")

	if remainingProperties > 0 {
		description += fmt.Sprintf("... `%d` more", remainingProperties)
	}

	return schemaDescription{Type: "object", Description: description}
}
