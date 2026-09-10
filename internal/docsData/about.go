// The about chunk mostly represents information in the `info` section of the
// spec. There will always be exactly one of these per document.

package docsData

import (
	"encoding/json"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type Server struct {
	URL string `json:"url"`
}

// Note: this is a trimmed down version of the Contact/License types found in
// the high-level model, since we don't want the low model or extensions
type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AboutData struct {
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Description    string   `json:"description"`
	TermsOfService string   `json:"termsOfService"`
	Version        string   `json:"version"`
	Contact        *Contact `json:"contact"`
	License        *License `json:"license"`
	Servers        []Server `json:"servers"`
}

type AboutChunk struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	ChunkData AboutData `json:"chunkData"`
	ChunkType string    `json:"chunkType"`
}

func SerializeAbout(docs *Docs, ast *ast.AST) error {
	chunk := AboutChunk{
		ID:   docs.GenerateID(),
		Slug: "about",
		ChunkData: AboutData{
			Title:          ast.OpenAPIDocument.GetInfo().GetTitle(),
			Summary:        ast.OpenAPIDocument.GetInfo().GetSummary(),
			Description:    ast.OpenAPIDocument.GetInfo().GetDescription(),
			TermsOfService: ast.OpenAPIDocument.GetInfo().GetTermsOfService(),
			Version:        ast.OpenAPIDocument.GetInfo().GetVersion(),
			Servers:        []Server{},
		},
		ChunkType: "about",
	}
	if ast.OpenAPIDocument.GetInfo().GetContact() != nil {
		chunk.ChunkData.Contact = &Contact{
			Name:  ast.OpenAPIDocument.GetInfo().GetContact().GetName(),
			URL:   ast.OpenAPIDocument.GetInfo().GetContact().GetURL(),
			Email: ast.OpenAPIDocument.GetInfo().GetContact().GetEmail(),
		}
	}
	if ast.OpenAPIDocument.GetInfo().GetLicense() != nil {
		chunk.ChunkData.License = &License{
			Name: ast.OpenAPIDocument.GetInfo().GetLicense().GetName(),
			URL:  ast.OpenAPIDocument.GetInfo().GetLicense().GetURL(),
		}
	}
	for _, server := range ast.MainSDK.Servers.Servers {
		chunk.ChunkData.Servers = append(chunk.ChunkData.Servers, Server{URL: server.URL})
	}
	b, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	docs.SaveSerializedChunk(chunk.ID, string(b))
	return nil
}
