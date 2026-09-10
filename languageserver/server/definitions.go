package server

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"gopkg.in/yaml.v3"
)

//go:embed definitions.yaml
var defintionsYaml string

//go:embed openapi-definitions.yaml
var openAPIDefinitionsYaml string

var (
	ErrNodeNotFound         = errors.New("unable to find node at position")
	ErrNoDocumentationFound = errors.New("no documentation found")
)

type definition struct {
	node          *yaml.Node
	documentation string
}

func lookupDefinition(doc *Document, params *protocol.HoverParams) (*definition, error) {
	// delegate the heavy lifting to libopenapi
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Content), &root); err != nil {
		return nil, fmt.Errorf("unable to unmarshall doc.Content into yaml, %w", err)
	}

	line := int(params.Position.Line)
	character := int(params.Position.Character)
	node, path := findNodeAtPosition(&root, line, character, []string{})
	if node == nil {
		return nil, ErrNodeNotFound
	}

	// Get enhanced documentation using openapi index if available
	documentation := getEnhancedDocumentation(node, path, doc.Index)
	if documentation == "" {
		return nil, ErrNoDocumentationFound
	}
	return &definition{
		node:          node,
		documentation: documentation,
	}, nil
}

// getEnhancedDocumentation gets documentation for OpenAPI properties
func getEnhancedDocumentation(_ *yaml.Node, path []string, _ *openapi.Index) string {
	if len(path) == 0 {
		return ""
	}

	property := path[len(path)-1]

	// TODO: Add enhanced context using our Index once we implement it
	// For now, just provide basic documentation

	// Handle Speakeasy extensions
	if strings.HasPrefix(property, "x-speakeasy-") {
		doc, err := getSpeakeasyExtensionDoc(property, path)
		if err != nil {
			return doc
		}
		return doc
	}

	// Handle standard OpenAPI properties
	return getOpenAPIPropertyDoc(property, path)
}

func getPropertyContext(path []string) string {
	if len(path) < 2 {
		return ""
	}
	return path[len(path)-2]
}

func getOpenAPIPropertyDoc(property string, path []string) string {
	// Try to get documentation from YAML first
	doc, err := getOpenAPIPropertyDocFromYAML(property, path)
	if err == nil && doc != "" {
		return doc
	}

	// Fallback for HTTP methods that have dynamic formatting
	switch property {
	case "get", "post", "put", "delete", "options", "head", "patch", "trace":
		return fmt.Sprintf("**HTTP %s Operation**\n\nDefines the %s operation for this path.\n\n**Type:** Operation Object", strings.ToUpper(property), strings.ToUpper(property))
	default:
		return ""
	}
}

type yamlDefinition struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
	Required    *bool  `yaml:"required,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Default     string `yaml:"default,omitempty"`
	Details     string `yaml:"details,omitempty"`
	Examples    string `yaml:"examples,omitempty"`
	ValidValues string `yaml:"validValues,omitempty"`
}

func getSpeakeasyExtensionDoc(property string, _ []string) (string, error) {
	var docs map[string]yamlDefinition
	err := yaml.Unmarshal([]byte(defintionsYaml), &docs)
	if err != nil {
		return "", err
	}

	if doc, exists := docs[property]; exists {
		if doc.Example != "" {
			return fmt.Sprintf("**%s**\n\n%s\n\n**Example**:\n\n%s", doc.Name, doc.Description, doc.Example), nil
		}
		return fmt.Sprintf("**%s**\n\n%s.", doc.Name, doc.Description), nil

	}

	// Fallback for unknown Speakeasy extensions
	return fmt.Sprintf("**Speakeasy Extension: %s**\n\nCustom Speakeasy extension for SDK generation.", property), nil
}

func getOpenAPIPropertyDocFromYAML(property string, path []string) (string, error) {
	var docs map[string]yamlDefinition
	err := yaml.Unmarshal([]byte(openAPIDefinitionsYaml), &docs)
	if err != nil {
		return "", err
	}

	// Handle context-specific variations (e.g., "title" in "info" context vs general "title")
	context := getPropertyContext(path)
	contextSpecificKey := ""
	if context != "" {
		contextSpecificKey = context + "." + property
	}

	// Try context-specific first, then general
	var definition yamlDefinition
	var exists bool

	if contextSpecificKey != "" {
		definition, exists = docs[contextSpecificKey]
	}

	if !exists {
		definition, exists = docs[property]
	}

	if !exists {
		return "", nil
	}

	// Build the documentation string
	var parts []string

	// Add the main header with name
	parts = append(parts, fmt.Sprintf("**%s**", definition.Name))

	// Add description
	if definition.Description != "" {
		parts = append(parts, "", definition.Description+".")
	}

	// Add required field if specified
	if definition.Required != nil {
		if *definition.Required {
			parts = append(parts, "", "**Required:** Yes")
		} else {
			parts = append(parts, "", "**Required:** No")
		}
	}

	// Add type if specified
	if definition.Type != "" {
		parts = append(parts, "**Type:** "+definition.Type)
	}

	// Add valid values if specified
	if definition.ValidValues != "" {
		parts = append(parts, "**Valid Values:** "+definition.ValidValues)
	}

	// Add default if specified
	if definition.Default != "" {
		parts = append(parts, "**Default:** "+definition.Default)
	}

	// Add example if specified
	if definition.Example != "" {
		parts = append(parts, "**Example:** "+definition.Example)
	}

	// Add examples (multiple) if specified
	if definition.Examples != "" {
		parts = append(parts, "**Examples:** "+definition.Examples)
	}

	// Add additional details if specified
	if definition.Details != "" {
		parts = append(parts, "", definition.Details)
	}

	return strings.Join(parts, "\n"), nil
}

// getLibOpenAPIContext extracts additional context from openapi index
func getLibOpenAPIContext(node *yaml.Node, path []string, idx *openapi.Index) string {
	var contextInfo []string

	// Only proceed if index is not nil
	if idx != nil {
		// TODO: Node origin tracking not yet implemented in our Index
		// This would show the absolute location of a node in the file

		// Check if this node is part of a reference by checking all references
		allRefs := idx.GetAllReferences()
		for _, indexNode := range allRefs {
			if indexNode != nil && indexNode.Node != nil && indexNode.Node.GetReference().String() == node.Value {
				contextInfo = append(contextInfo, fmt.Sprintf("**Reference:** `%s`", node.Value))
				contextInfo = append(contextInfo, "**Status:** Resolved")
				break
			}
		}

		// Check if this is a schema component definition
		if len(path) >= 3 && path[0] == "components" && path[1] == "schemas" {
			schemaName := path[2]
			if idx.ComponentSchemas != nil {
				for _, indexNode := range idx.ComponentSchemas {
					if indexNode != nil && len(indexNode.Location) > 0 {
						lastLocation := indexNode.Location[len(indexNode.Location)-1]
						if pointer.Value(lastLocation.ParentKey) == schemaName {
							contextInfo = append(contextInfo, fmt.Sprintf("**Schema Component:** `%s`", schemaName))
							contextInfo = append(contextInfo, "**Schema Status:** Resolved")
							break
						}
					}
				}
			}
		}

		// Add path information for OpenAPI context
		if len(path) >= 2 {
			contextInfo = append(contextInfo, fmt.Sprintf("**Path:** `%s`", strings.Join(path, " -> ")))
		}

		// Check for circular references
		if circularRefCount := idx.GetValidCircularRefCount(); circularRefCount > 0 {
			contextInfo = append(contextInfo, fmt.Sprintf("**Info:** Document contains %d circular references", circularRefCount))
		}
	} else if len(path) >= 2 {
		// Even without index, add path information for context
		contextInfo = append(contextInfo, fmt.Sprintf("**Path:** `%s`", strings.Join(path, " -> ")))
	}

	// Return formatted context information
	if len(contextInfo) > 0 {
		return strings.Join(contextInfo, "\n\n")
	}
	return ""
}

func findNodeAtPosition(node *yaml.Node, line, character int, path []string) (*yaml.Node, []string) {
	if node == nil {
		return nil, nil
	}

	// Handle DocumentNode by traversing into its content
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			return findNodeAtPosition(node.Content[0], line, character, path)
		}
		return nil, nil
	}

	yamlLine := node.Line - 1
	yamlCol := node.Column - 1

	if yamlLine == line && yamlCol <= character && character <= yamlCol+len(node.Value) {
		return node, path
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := append(path, keyNode.Value)

			if result, resultPath := findNodeAtPosition(keyNode, line, character, newPath); result != nil {
				return result, resultPath
			}

			if result, resultPath := findNodeAtPosition(valueNode, line, character, newPath); result != nil {
				return result, resultPath
			}
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := append(path, fmt.Sprintf("[%d]", i))
			if result, resultPath := findNodeAtPosition(child, line, character, newPath); result != nil {
				return result, resultPath
			}
		}
	}

	return nil, nil
}

// isRenameableSymbol determines if a symbol at the given path can be renamed
func isRenameableSymbol(node *yaml.Node, path []string, doc *Document) bool {
	if node == nil || len(path) == 0 {
		return false
	}

	// Only allow renaming of scalar values (not keys or complex structures)
	if node.Kind != yaml.ScalarNode {
		return false
	}

	// Get the property name (last element in path)
	property := path[len(path)-1]

	// Check if this is a context where renaming makes sense
	if !isRenameableContext(path) {
		return false
	}

	// Don't allow renaming of OpenAPI reserved properties (context-aware, version-specific)
	if doc == nil || doc.ReservedChecker == nil {
		return false // Be conservative and disallow rename
	}

	// Get OpenAPI version from parsed document
	var version string
	if doc.ParsedDoc != nil {
		version = doc.ParsedDoc.OpenAPI
	}
	if doc.ReservedChecker.IsOASReservedPropertyName(property, path, version) {
		return false
	}

	// Don't allow renaming HTTP methods or status codes
	httpMethods := []string{"get", "post", "put", "delete", "patch", "head", "options", "trace"}
	if slices.Contains(httpMethods, property) {
		return false
	}

	// Don't allow renaming if it looks like a status code
	if len(property) == 3 && property[0] >= '1' && property[0] <= '5' {
		return false
	}

	// Don't allow renaming standard MIME types
	if isStandardMimeType(node.Value) {
		return false
	}

	// Don't allow renaming JSON references directly - users should rename the definitions instead
	// Unfortunately, this means we can't rename references like "#/components/schemas/User" directly
	// but the amount of work to support this is a fair bit more than handling it from the definition side
	if strings.HasPrefix(node.Value, "#/") {
		return false
	}

	// Don't allow renaming HTTP references
	if strings.HasPrefix(node.Value, "http") {
		return false
	}

	// Allow renaming for most other cases (schema names, parameter names, etc.)
	return true
}

// isRenameableContext determines if the context/path allows for renaming
func isRenameableContext(path []string) bool {
	if len(path) == 0 {
		return false
	}

	// Check for renameable contexts in OpenAPI
	for i, segment := range path {
		switch segment {
		case "schemas", "parameters", "responses", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks":
			// These are component sections where the next segment would be a user-defined name
			if i+1 < len(path) {
				return true
			}
		case "properties":
			// Property names within schemas are renameable
			return true
		case "allOf", "oneOf", "anyOf", "not":
			// These are schema composition keywords, content depends on context
			continue
		}
	}

	// Check if we're in a path parameter context
	if len(path) >= 2 && path[0] == "paths" {
		// Path parameter names in the path template are renameable
		if strings.Contains(path[1], "{") && strings.Contains(path[1], "}") {
			return false // Don't rename path templates directly
		}
	}

	// Check for tag names
	if len(path) >= 2 && path[len(path)-2] == "tags" && path[len(path)-1] == "name" {
		return true
	}

	// Check for inline path parameter names
	if len(path) >= 3 && path[len(path)-1] == "name" {
		// Look for patterns like [..., "parameters", "[0]", "name"]
		for i, segment := range path {
			if segment == "parameters" && i+2 < len(path) {
				// Check if the next segment is an array index like "[0]"
				if strings.HasPrefix(path[i+1], "[") && strings.HasSuffix(path[i+1], "]") {
					// And the one after that is "name"
					if path[i+2] == "name" {
						return true
					}
				}
			}
		}
	}

	return false
}

// isStandardMimeType checks if a value is a standard MIME type that shouldn't be renamed
func isStandardMimeType(value string) bool {
	standardMimeTypes := []string{
		"application/json", "application/xml", "application/x-www-form-urlencoded",
		"multipart/form-data", "text/plain", "text/html", "text/csv", "text/xml",
		"application/octet-stream", "application/pdf", "image/png", "image/jpeg",
		"image/gif", "image/svg+xml", "application/javascript", "text/css",
	}

	return slices.Contains(standardMimeTypes, value)
}

// findReferences finds all references to a symbol with the given value in the document
func findReferences(root *yaml.Node, symbolValue string, originalPath []string) []*yaml.Node {
	var references []*yaml.Node
	findReferencesRecursive(root, symbolValue, originalPath, []string{}, &references)
	return references
}

// findReferencesRecursive recursively searches for references to a symbol
func findReferencesRecursive(node *yaml.Node, symbolValue string, originalPath []string, currentPath []string, references *[]*yaml.Node) {
	if node == nil {
		return
	}

	// Handle DocumentNode by traversing into its content
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			findReferencesRecursive(node.Content[0], symbolValue, originalPath, currentPath, references)
		}
		return
	}

	// Check if this node matches the symbol we're looking for
	if node.Kind == yaml.ScalarNode {
		// Direct match for symbol value
		if node.Value == symbolValue && isMatchingSymbolContext(currentPath, originalPath) {
			*references = append(*references, node)
		}

		// Check for JSON references (e.g., "#/components/schemas/User" or "#/components/parameters/userId")
		if isJSONReference(node.Value) {
			if isComponentReference(originalPath) {
				refSchemaName := extractComponentNameFromRef(node.Value)
				if refSchemaName == symbolValue {
					*references = append(*references, node)
				}
			}
		}
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := append(currentPath, keyNode.Value)

			// Check both key and value nodes
			findReferencesRecursive(keyNode, symbolValue, originalPath, newPath, references)
			findReferencesRecursive(valueNode, symbolValue, originalPath, newPath, references)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := append(currentPath, fmt.Sprintf("[%d]", i))
			findReferencesRecursive(child, symbolValue, originalPath, newPath, references)
		}
	}
}

// isJSONReference checks if a string is a JSON reference (starts with #/)
func isJSONReference(value string) bool {
	return strings.HasPrefix(value, "#/")
}

// extractComponentNameFromRef extracts the schema name from a JSON reference
func extractComponentNameFromRef(ref string) string {
	// Handle references like "#/components/schemas/User"
	if strings.HasPrefix(ref, "#/components/") {
		parts := strings.Split(ref, "/")
		if len(parts) >= 4 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// extractNameFromUserInput extracts just the name part from user input that might be a full reference path
// For example: "#/components/schemas/Pet" -> "Pet", "Pet" -> "Pet"
func extractNameFromUserInput(input string) string {
	// If the input is a full component ref, extract just the name
	if strings.HasPrefix(input, "#/components") {
		return extractComponentNameFromRef(input)
	}
	// Otherwise, assume it's just the name itself
	return input
}

// isMatchingSymbolContext determines if two paths represent the same type of symbol
func isMatchingSymbolContext(path1, path2 []string) bool {
	if len(path1) == 0 || len(path2) == 0 {
		return false
	}

	// Get the property names (last elements)
	prop1 := path1[len(path1)-1]
	prop2 := path2[len(path2)-1]

	// For schema names, check if both are schema contexts
	if containsSegment(path1, "schemas") && containsSegment(path2, "schemas") {
		// Both are schemas, match if they have the same name
		return prop1 == prop2
	}

	// For parameter names, check if both are parameter contexts
	if containsSegment(path1, "parameters") && containsSegment(path2, "parameters") {
		// Both are parameters, match if they have the same name
		return prop1 == prop2
	}

	// For property names, match based on property context
	if containsSegment(path1, "properties") && containsSegment(path2, "properties") {
		return prop1 == prop2
	}

	// For tag names
	if containsSegment(path1, "tags") && containsSegment(path2, "tags") && prop1 == "name" && prop2 == "name" {
		return true
	}

	// For simple case, if the property names match, consider them the same context
	return prop1 == prop2
}

// isComponentReference checks if the path represents a schema reference or definition
func isComponentReference(path []string) bool {
	return containsSegment(path, "schemas") ||
		(len(path) >= 2 && path[len(path)-2] == "schemas")
}

// isInlinePathParameter checks if the path represents an inline path parameter definition
func isInlinePathParameter(path []string) bool {
	if len(path) < 3 || path[len(path)-1] != "name" {
		return false
	}

	// Look for patterns like [..., "parameters", "[0]", "name"]
	for i, segment := range path {
		if segment == "parameters" && i+2 < len(path) {
			// Check if the next segment is an array index like "[0]"
			if strings.HasPrefix(path[i+1], "[") && strings.HasSuffix(path[i+1], "]") {
				// And the one after that is "name"
				if path[i+2] == "name" {
					return true
				}
			}
		}
	}
	return false
}

// getPathFromInlineParameter extracts the path template from an inline parameter path
func getPathFromInlineParameter(path []string) string {
	// Find the path segment - should be the second element after "paths"
	for i, segment := range path {
		if segment == "paths" && i+1 < len(path) {
			return path[i+1]
		}
	}
	return ""
}

// containsSegment checks if a path contains a specific segment
func containsSegment(path []string, segment string) bool {
	for _, s := range path {
		if s == segment {
			return true
		}
	}
	return false
}
