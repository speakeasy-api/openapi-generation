// The operation chunk represents a single operation, e.g. `POST /some/path`.
// All data, such as request bodies, responses, etc. are stored in separate
// schema chunks.

package docsData

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type TopLevelExample struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Security struct {
	ContentChunkID string `json:"contentChunkId"`
}

type Parameter struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	Deprecated   bool   `json:"deprecated"`
	In           string `json:"in"`
	FieldChunkID string `json:"fieldChunkId"`
}

type Response struct {
	Description    string            `json:"description"`
	ContentType    string            `json:"contentType"`
	ContentChunkID string            `json:"contentChunkId"`
	Examples       []TopLevelExample `json:"examples"`
}

type RequestBody struct {
	Description    string            `json:"description"`
	Required       bool              `json:"required"`
	ContentChunkID string            `json:"contentChunkId"`
	Examples       []TopLevelExample `json:"examples"`
}

type OperationData struct {
	OperationID string                `json:"operationId"`
	Path        string                `json:"path"`
	Method      string                `json:"method"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Parameters  []Parameter           `json:"parameters"`
	RequestBody *RequestBody          `json:"requestBody"`
	Responses   map[string][]Response `json:"responses"`
	Tag         string                `json:"tag"`
	Security    *Security             `json:"security"`
	Deprecated  bool                  `json:"deprecated"`
	CodeSamples map[string]string     `json:"codeSamples"`
}

type OperationChunk struct {
	ID        string        `json:"id"`
	Slug      string        `json:"slug"`
	ChunkData OperationData `json:"chunkData"`
	ChunkType string        `json:"chunkType"`
}

type XCodeSample struct {
	Lang   string `json:"lang"`
	Label  string `json:"label"`
	Source string `json:"source"`
}

func SerializeOperation(docs *Docs, operation *ast.Operation, tag string) (string, error) {
	var chunk OperationChunk

	chunk.ID = docs.GenerateID()
	chunk.Slug = fmt.Sprintf("operation%s/%s", operation.Path, operation.Method)

	var summary, description string
	if operation.Comments != nil {
		summary = operation.Comments.Summary
		description = operation.Comments.Description
	}

	chunk.ChunkData = OperationData{
		OperationID: operation.OriginalID,
		Path:        operation.Path,
		Method:      operation.Method,
		Summary:     summary,
		Description: description,
		Parameters:  []Parameter{},
		Responses:   map[string][]Response{},
		Tag:         tag,
		Security:    nil,
		Deprecated:  operation.Comments != nil && operation.Comments.Deprecated,
		CodeSamples: make(map[string]string),
	}
	chunk.ChunkType = "operation"

	// Add code samples, if present
	if operation.Extensions != nil && operation.Extensions.All != nil && operation.Extensions.All["x-codeSamples"] != nil {
		if codeSamples, ok := operation.Extensions.All["x-codeSamples"].([]any); ok {
			// Process the array of code samples
			for _, codeSample := range codeSamples {
				if codeSampleMap, ok := codeSample.(map[string]any); ok {
					if lang, ok := codeSampleMap["lang"].(string); ok {
						if source, ok := codeSampleMap["source"].(string); ok {
							chunk.ChunkData.CodeSamples[lang] = source
						}
					}
				}
			}
		}
	}

	// Add and serialize the security schemas, if present
	if operation.Security != nil {
		securityChunkId, err := SerializeSecurity(docs, operation.Security.Type, false)
		if err != nil {
			return "", err
		}
		chunk.ChunkData.Security = &Security{
			ContentChunkID: *securityChunkId,
		}
	}

	// Add and serialize the parameters, if any
	if operation.Arguments.ParamFields != nil {
		for _, param := range operation.Arguments.ParamFields {
			fieldChunkId, err := SerializeSchema(docs, param.Type, param.Nullable, GetSerializedDefault(param))
			if err != nil {
				return "", err
			}
			var description string
			if param.Comments != nil {
				description = param.Comments.Description
			}
			if len(param.Annotations) != 1 || param.Annotations[0].Type() != "param" {
				return "", fmt.Errorf("expected 1 param annotation, got %d", len(param.Annotations))
			}
			annotation := param.Annotations[0].(*ast.ParamAnnotation)
			chunk.ChunkData.Parameters = append(chunk.ChunkData.Parameters, Parameter{
				Name:         param.Name,
				Description:  description,
				Required:     !param.Optional,
				In:           strings.ReplaceAll(annotation.ParamType, "Param", ""),
				FieldChunkID: *fieldChunkId,
				Deprecated:   param.Comments != nil && param.Comments.Deprecated,
			})
		}
	}

	// Add and serialize the request body, if it exists
	if operation.Request != nil && operation.Request.RequestBody != nil {
		contentChunkId, err := SerializeSchema(docs, operation.Request.RequestBody.Type, operation.Request.RequestBody.Nullable, GetSerializedDefault(operation.Request.RequestBody))
		if err != nil {
			return "", err
		}
		var description string
		if operation.Request.RequestBody.Comments != nil {
			description = operation.Request.RequestBody.Comments.Description
		}
		examples := []TopLevelExample{}
		if operation.Request.Examples != nil {
			for _, example := range operation.Request.Examples {
				exampleJSON := example.ToJSON()
				examples = append(examples, TopLevelExample{
					Name:  example.Name(),
					Value: exampleJSON,
				})
			}
		}
		chunk.ChunkData.RequestBody = &RequestBody{
			Description:    description,
			Required:       !operation.Request.RequestBody.Optional,
			ContentChunkID: *contentChunkId,
			Examples:       examples,
		}
	}

	for _, response := range operation.Response.Responses {
		responses := make([]Response, 0)

		// TODO: We don't set a response for 4XX/5XX error codes that are explicitly
		// listed in the spec currently as a response, meaning we don't have a
		// description. We should update the generator to pass this through.
		for _, content := range response.Content {
			contentChunkId, err := SerializeSchema(docs, content.Content.Type, content.Content.Nullable, GetSerializedDefault(content.Content))
			if err != nil {
				return "", err
			}
			var description string
			if content.Content.Comments != nil {
				description = content.Content.Comments.Description
			}
			examples := []TopLevelExample{}
			if content.Examples != nil {
				for _, example := range content.Examples {
					exampleJSON := example.ToJSON()
					examples = append(examples, TopLevelExample{
						Name:  example.Name(),
						Value: exampleJSON,
					})
				}
			}
			responses = append(responses, Response{
				Description:    description,
				ContentType:    content.ContentType,
				ContentChunkID: *contentChunkId,
				Examples:       examples,
			})
		}
		// Codes has length > 1 if we have a default 4XX and 5XX response code as
		// well as explicitly listed codes. We only want the explicitly listed code
		// though, which always comes first
		chunk.ChunkData.Responses[response.Code[0]] = responses
	}

	b, err := json.Marshal(chunk)
	if err != nil {
		return "", err
	}
	docs.SaveSerializedChunk(chunk.ID, string(b))
	return chunk.ID, nil
}
