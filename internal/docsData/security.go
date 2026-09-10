// The global security chunk represents a security that applies to all
// operations. We show this information separately on the Security page so that
// we don't clutter up individual operations

package docsData

import (
	"encoding/json"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type SecurityEntryData struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	In          string `json:"in"`
	// TODO: Scheme, BearerFormat, Flows, OpenIDConnectURL
}

type SecurityData struct {
	ID      string              `json:"id"`
	Entries []SecurityEntryData `json:"entries"`
}

type SecurityChunk struct {
	ID        string       `json:"id"`
	Slug      string       `json:"slug"`
	ChunkData SecurityData `json:"chunkData"`
	ChunkType string       `json:"chunkType"`
}

func SerializeSecurity(docs *Docs, security *ast.TypeDef, isGlobal bool) (*string, error) {
	chunkType := "security"
	if isGlobal {
		chunkType = "globalSecurity"
	}
	chunk := SecurityChunk{
		ID:   docs.GenerateID(),
		Slug: "globalSecurity",
		ChunkData: SecurityData{
			Entries: []SecurityEntryData{},
		},
		ChunkType: chunkType,
	}

	// Each field in the security type is a security entry
	for _, field := range security.Fields {
		// Check if field has a security annotation
		securityAnnotation := field.Annotations.Get(ast.AnnotationTypeSecurity)
		if securityAnnotation == nil {
			return nil, fmt.Errorf("field %s missing security annotation", field.Name)
		}

		description := ""
		if field.Comments != nil {
			// TODO: this doesn't seem to be wired up
			description = field.Comments.Description
		}

		annotation := securityAnnotation.(*ast.SecurityAnnotation)
		chunk.ChunkData.Entries = append(chunk.ChunkData.Entries, SecurityEntryData{
			Type:        annotation.SecType,
			Description: description,
			Name:        field.Name,
			In:          annotation.SubType,
		})
	}

	b, err := json.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	docs.SaveSerializedChunk(chunk.ID, string(b))
	return &chunk.ID, nil
}
