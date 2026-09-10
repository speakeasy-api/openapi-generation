// For a high-level overview of what the docsData package does, see README.md

package docsData

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type AnnotatedOperation struct {
	Operation *ast.Operation
	Tags      []string
}

// Operations are nested by tag. We can have indefinitely nested tags if they're
// written in `foo.bar.baz` format, so we need to recursively traverse the SDK
// tree to get all operations.
func getOperationsList(sdk *ast.SDK, parentTags []string) []*AnnotatedOperation {
	operations := make([]*AnnotatedOperation, 0, len(sdk.Operations))
	if sdk.FieldName != "" {
		parentTags = append(parentTags, sdk.FieldName)
	}
	for _, operation := range sdk.Operations {
		operations = append(operations, &AnnotatedOperation{
			Operation: operation,
			Tags:      parentTags,
		})
	}
	for _, subSDK := range sdk.SubSDKs {
		operations = append(operations, getOperationsList(subSDK, parentTags)...)
	}
	return operations
}

// Creates docs data from the given AST. This function returns chunks in a
// serialized format that can be written to a JSONL file, sent across an IPC
// connection, sent over a network connection, etc.
func NewDocsData(ast *ast.AST) ([]*string, error) {
	// First, create the docs data structure for registering chunks
	docs := &Docs{
		slugMap:   map[string]string{},
		chunkList: []*string{},
	}

	// Next, create the about chunk
	err := SerializeAbout(docs, ast)
	if err != nil {
		return nil, err
	}

	// Next, create the global security chunk
	if ast.MainSDK.Security != nil {
		_, err = SerializeSecurity(docs, ast.MainSDK.Security.Type, true)
		if err != nil {
			return nil, err
		}
	}

	// Next, get the tag descriptions. Tag descriptions aren't propogated to SDKs,
	// so we have to get them directly from the OpenAPI document
	tagToDescription := map[string]string{}
	for _, tag := range ast.OpenAPIDocument.Tags {
		tagName := strings.Join(strings.Split(tag.Name, "."), "/")
		tagToDescription[tagName] = tag.GetDescription()
	}

	// Next, register the operation chunks
	operations := getOperationsList(ast.MainSDK, nil)
	tagToOperationIds := map[string][]string{}
	for _, operation := range operations {
		operationId, err := SerializeOperation(docs, operation.Operation, "")
		if err != nil {
			return nil, err
		}
		tagName := strings.Join(operation.Tags, "/")
		tagToOperationIds[tagName] = append(tagToOperationIds[tagName], operationId)
	}

	// Next, register the tag chunks
	for tag, operationIds := range tagToOperationIds {
		var description *string
		if desc, exists := tagToDescription[tag]; exists && desc != "" {
			description = &desc
		}
		err := SerializeTag(docs, tag, description, operationIds)
		if err != nil {
			return nil, err
		}
	}

	// Now we've finished and are ready to return the serialized data
	return docs.chunkList, nil
}
