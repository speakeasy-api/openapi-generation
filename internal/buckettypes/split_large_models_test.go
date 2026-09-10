package buckettypes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSanitizer is a test double that lowercases filenames.
type mockSanitizer struct{}

func (m mockSanitizer) SanitizeFileName(name string) string {
	return strings.ToLower(name)
}

func (m mockSanitizer) GetEnumNames(_ *ast.TypeDef) []string {
	return nil
}

func makeTypeDefs(n int, prefix string) ast.TypeDefs {
	defs := make(ast.TypeDefs, n)
	for i := range defs {
		defs[i] = &ast.TypeDef{
			Name: fmt.Sprintf("%s_Type%d", prefix, i),
			Type: ast.DataTypeClass,
		}
	}
	return defs
}

func TestSplitLargeModels_BelowThreshold(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	defs := makeTypeDefs(100, "Small")
	models.Set("SmallModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	resultDefs, ok := resultModels.Get("SmallModel")
	require.True(t, ok)
	assert.Len(t, resultDefs, 100)
}

func TestSplitLargeModels_AtThreshold(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	defs := makeTypeDefs(SplitModelTypeThreshold, "Exact")
	models.Set("ExactModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	resultDefs, ok := resultModels.Get("ExactModel")
	require.True(t, ok)
	assert.Len(t, resultDefs, SplitModelTypeThreshold)
}

func TestSplitLargeModels_AboveThreshold(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	count := SplitModelTypeThreshold + 1
	defs := makeTypeDefs(count, "Big")
	models.Set("BigModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	// Original model should not exist
	_, ok = resultModels.Get("BigModel")
	assert.False(t, ok)

	// Should have 2 chunks: 150 + 51
	totalTypes := 0
	chunkCount := 0
	for _, chunkDefs := range resultModels.All() {
		chunkCount++
		totalTypes += len(chunkDefs)
	}
	assert.Equal(t, 2, chunkCount)
	assert.Equal(t, count, totalTypes)
}

func TestSplitLargeModels_ChunkNaming(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()

	// 301 types → 3 chunks of 150, 150, 1
	// Without children, topo sort preserves order, so last type in each chunk
	// is at index 149, 299, 300
	defs := makeTypeDefs(301, "T")
	models.Set("OriginalModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	// Check chunk names match last type in each chunk
	expectedNames := []string{"T_Type149", "T_Type299", "T_Type300"}
	i := 0
	for name := range resultModels.All() {
		require.Less(t, i, len(expectedNames))
		assert.Equal(t, expectedNames[i], name)
		i++
	}
	assert.Equal(t, len(expectedNames), i)
}

func TestSplitLargeModels_NameCollision(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()

	defs := makeTypeDefs(301, "T")
	models.Set("OriginalModel", defs)

	// Pre-existing model that collides with what would be the first chunk's name
	collisionDefs := makeTypeDefs(1, "Collision")
	models.Set("T_Type149", collisionDefs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	// The original "T_Type149" model should still exist
	_, ok = resultModels.Get("T_Type149")
	assert.True(t, ok)

	// The colliding chunk should be renamed with suffix
	_, ok = resultModels.Get("T_Type149_1")
	assert.True(t, ok)
}

func TestSplitLargeModels_ResolvedModel(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	defs := makeTypeDefs(301, "R")
	models.Set("BigModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	for modelName, chunkDefs := range resultModels.All() {
		for _, td := range chunkDefs {
			assert.Equal(t, modelName, td.ResolvedModel,
				"Type %s should have ResolvedModel=%s", td.Name, modelName)
		}
	}
}

func TestSplitLargeModels_OperationServers(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	defs := makeTypeDefs(301, "S")
	models.Set("SplitOp", defs)
	types.Set("operations", models)

	servers := &ast.Servers{
		Servers: []*ast.Server{{URL: "https://example.com"}},
	}
	opServers := sequencedmap.New[string, *ast.Servers]()
	opServers.Set("SplitOp", servers)

	result := b.SplitLargeModels(types, opServers)

	// Old key should be removed
	_, ok := opServers.Get("SplitOp")
	assert.False(t, ok)

	// First chunk should have the servers
	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	var firstName string
	for name := range resultModels.All() {
		firstName = name
		break
	}

	s, ok := opServers.Get(firstName)
	require.True(t, ok)
	assert.Equal(t, servers, s)
}

func TestSplitLargeModels_OriginalNameReusable(t *testing.T) {
	b := &Bucketer{}
	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()

	// Create 201 types where the last type (after topo sort) has the same name
	// as the original model. This simulates the real-world case where the root
	// type (e.g. RestCollectorConf) shares its name with the model.
	defs := makeTypeDefs(SplitModelTypeThreshold+1, "T")
	defs[len(defs)-1].Name = "MyModel" // last type matches model name
	models.Set("MyModel", defs)
	types.Set("operations", models)

	opServers := sequencedmap.New[string, *ast.Servers]()

	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("operations")
	require.True(t, ok)

	// The last chunk should be named "MyModel" (not "MyModel_1") since the
	// original model name is freed when splitting.
	_, ok = resultModels.Get("MyModel")
	assert.True(t, ok, "Last chunk should reuse the original model name")

	// Should NOT have MyModel_1
	_, ok = resultModels.Get("MyModel_1")
	assert.False(t, ok, "Should not need collision suffix for original model name")
}

func TestSplitLargeModels_ReservedFileName(t *testing.T) {
	ss := &subsystem.Subsystem{
		Sanitizer: mockSanitizer{},
	}
	b := &Bucketer{
		subsystem:              ss,
		reservedModelFileNames: map[string]bool{"__init__": true},
	}

	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()

	// Create types where the last type's name would sanitize to a reserved name.
	defs := makeTypeDefs(SplitModelTypeThreshold+1, "T")
	defs[len(defs)-1].Name = "__init__"
	models.Set("BigModel", defs)
	types.Set("shared", models)

	opServers := sequencedmap.New[string, *ast.Servers]()
	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("shared")
	require.True(t, ok)

	// The chunk whose last type is "__init__" should NOT use that name.
	_, ok = resultModels.Get("__init__")
	assert.False(t, ok, "Reserved name should not be used as chunk name")

	// It should be suffixed instead.
	_, ok = resultModels.Get("__init___1")
	assert.True(t, ok, "Chunk should be renamed with _1 suffix")
}

func TestSplitLargeModels_NoReservedNames(t *testing.T) {
	// When no reserved names are set, behavior is unchanged (no panics, etc.)
	b := &Bucketer{}

	types := ast.NewBucketedTypes()
	models := sequencedmap.New[string, ast.TypeDefs]()
	defs := makeTypeDefs(SplitModelTypeThreshold+1, "T")
	defs[len(defs)-1].Name = "__init__"
	models.Set("BigModel", defs)
	types.Set("shared", models)

	opServers := sequencedmap.New[string, *ast.Servers]()
	result := b.SplitLargeModels(types, opServers)

	resultModels, ok := result.Get("shared")
	require.True(t, ok)

	// Without reserved names, __init__ is fine to use
	_, ok = resultModels.Get("__init__")
	assert.True(t, ok, "__init__ should be usable when no reserved names are set")
}

func TestChunkTypeDefs(t *testing.T) {
	defs := makeTypeDefs(10, "C")

	chunks := chunkTypeDefs(defs, 3)
	assert.Len(t, chunks, 4)
	assert.Len(t, chunks[0], 3)
	assert.Len(t, chunks[1], 3)
	assert.Len(t, chunks[2], 3)
	assert.Len(t, chunks[3], 1)
}

func TestChunkTypeDefs_ExactMultiple(t *testing.T) {
	defs := makeTypeDefs(9, "C")

	chunks := chunkTypeDefs(defs, 3)
	assert.Len(t, chunks, 3)
	for _, chunk := range chunks {
		assert.Len(t, chunk, 3)
	}
}
