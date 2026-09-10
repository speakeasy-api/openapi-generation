package docsData

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
)

func loadSpecAndBuildAST() (*ast.AST, error) {
	file, err := os.Open("testdata/petstore.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gen, err := generate.New(generate.WithDontWrite())
	if err != nil {
		return nil, err
	}

	schemaStringBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Convert js.Value args to expected types
	docInfo := document.DocumentInfo{
		Schema:     schemaStringBytes,
		SchemaPath: "foo.json",
		IsRemote:   false,
	}

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	// init
	err = gen.Init(ctx, "go", "")
	if err != nil {
		return nil, err
	}

	ad := &analytics.Data{}

	if _, err = gen.LoadAndValidateDoc(ctx, &docInfo, ad, "", true); err != nil {
		return nil, err
	}

	ast, err := gen.GenerateAST(ctx, &docInfo, ad)
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func Test_NewDocsData(t *testing.T) {
	ast, err := loadSpecAndBuildAST()
	if err != nil {
		t.Fatal(err)
	}

	docsData, err := NewDocsData(ast)
	if err != nil {
		t.Fatal(err)
	}

	// Log the chunks for debugging
	for _, chunk := range docsData {
		fmt.Println(*chunk)
	}

	// Verify the number of chunks
	if len(docsData) != 19 {
		t.Errorf("Expected 19 chunks, got %d", len(docsData))
	}

	// Parse the JSON chunks and count by type
	type MinimalChunk struct {
		ChunkType string `json:"chunkType"`
	}

	aboutCount := 0
	schemaCount := 0
	operationCount := 0
	tagCount := 0
	securityCount := 0
	globalSecurityCount := 0

	// Do some very basic testing. In time, we'll add more complex specs with
	// more in-depth tests.
	for _, chunk := range docsData {
		var parsedChunk MinimalChunk
		if err := json.Unmarshal([]byte(*chunk), &parsedChunk); err != nil {
			t.Errorf("Failed to parse chunk: %v", err)
			continue
		}

		switch parsedChunk.ChunkType {
		case "about":
			aboutCount++
		case "schema":
			schemaCount++
		case "operation":
			operationCount++
		case "tag":
			tagCount++
		case "security":
			securityCount++
		case "globalSecurity":
			globalSecurityCount++
		default:
			t.Errorf("Unexpected chunk type: %s", parsedChunk.ChunkType)
		}
	}

	// Verify counts by type
	if aboutCount != 1 {
		t.Errorf("Expected 1 'about' chunk, got %d", aboutCount)
	}
	if schemaCount != 10 {
		t.Errorf("Expected 10 'schema' chunks, got %d", schemaCount)
	}
	if operationCount != 3 {
		t.Errorf("Expected 3 'operation' chunks, got %d", operationCount)
	}
	if tagCount != 3 {
		t.Errorf("Expected 3 'tag' chunks, got %d", tagCount)
	}
	if securityCount != 1 {
		t.Errorf("Expected 1 'security' chunk, got %d", securityCount)
	}
	if globalSecurityCount != 1 {
		t.Errorf("Expected 1 'globalSecurity' chunk, got %d", globalSecurityCount)
	}
}
