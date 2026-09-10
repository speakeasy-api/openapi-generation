package openapi

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedReferenceRegistry_Track(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()

	// First tracking should succeed
	tracked := registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema1",
	)
	assert.True(t, tracked, "first tracking should return true")

	// Second tracking of the same shared ref should not update
	tracked = registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema2",
	)
	assert.False(t, tracked, "second tracking should return false")

	// Verify the original ref is still Schema1, not Schema2
	originalRef := registry.GetOriginalRef("#/components/schemas/SchemaShared")
	assert.Equal(t, "#/components/schemas/Schema1", originalRef)
}

func TestNestedReferenceRegistry_GetOriginalRef(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()

	// Non-existent ref returns empty string
	ref := registry.GetOriginalRef("#/components/schemas/NotTracked")
	assert.Empty(t, ref)

	// Track a reference
	registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema1",
	)

	// Now it should return the original ref
	ref = registry.GetOriginalRef("#/components/schemas/SchemaShared")
	assert.Equal(t, "#/components/schemas/Schema1", ref)
}

func TestNestedReferenceRegistry_IsTracked(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()

	assert.False(t, registry.IsTracked("#/components/schemas/SchemaShared"))

	registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema1",
	)

	assert.True(t, registry.IsTracked("#/components/schemas/SchemaShared"))
}

func TestNestedReferenceRegistry_Clear(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()

	registry.Track(
		"#/components/schemas/Schema1",
		"#/components/schemas/Ref1",
	)
	registry.Track(
		"#/components/schemas/Schema2",
		"#/components/schemas/Ref2",
	)

	assert.Equal(t, 2, registry.Len())

	registry.Clear()

	assert.Equal(t, 0, registry.Len())
	assert.False(t, registry.IsTracked("#/components/schemas/Schema1"))
	assert.False(t, registry.IsTracked("#/components/schemas/Schema2"))
}

func TestNestedReferenceRegistry_All(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()

	registry.Track(
		"#/components/schemas/Schema1",
		"#/components/schemas/Ref1",
	)
	registry.Track(
		"#/components/schemas/Schema2",
		"#/components/schemas/Ref2",
	)

	all := registry.All()

	assert.Len(t, all, 2)
	assert.Equal(t, "#/components/schemas/Ref1", all["#/components/schemas/Schema1"])
	assert.Equal(t, "#/components/schemas/Ref2", all["#/components/schemas/Schema2"])

	// Modifying the returned map should not affect the registry
	delete(all, "#/components/schemas/Schema1")
	assert.Equal(t, 2, registry.Len())
}

func TestNestedReferenceRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	registry := NewNestedReferenceRegistry()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.Track(
				"#/components/schemas/Shared",
				"#/components/schemas/Ref",
			)
		}()
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.GetOriginalRef("#/components/schemas/Shared")
			registry.IsTracked("#/components/schemas/Shared")
			registry.Len()
			registry.All()
		}()
	}

	wg.Wait()

	// Should only have one entry since all writes were for the same key
	assert.Equal(t, 1, registry.Len())
}

func TestNestedReferenceRegistry_MultipleSharedReferences(t *testing.T) {
	t.Parallel()

	// Simulates the scenario:
	// - Schema1 references SchemaShared
	// - Schema2 also references SchemaShared
	// We should track that SchemaShared was first referenced via Schema1

	registry := NewNestedReferenceRegistry()

	// Schema1 -> SchemaShared discovered first
	tracked := registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema1",
	)
	assert.True(t, tracked, "first discovery should be tracked")

	// Schema2 -> SchemaShared discovered second
	tracked = registry.Track(
		"#/components/schemas/SchemaShared",
		"#/components/schemas/Schema2",
	)
	assert.False(t, tracked, "already tracked, should return false")

	// Verify the tracking
	originalRef := registry.GetOriginalRef("#/components/schemas/SchemaShared")
	require.NotEmpty(t, originalRef, "should have tracked SchemaShared")
	assert.Equal(t, "#/components/schemas/Schema1", originalRef, "Schema1 should be the original ref")
}

func TestNestedReferenceRegistry_ThreeLevelChain(t *testing.T) {
	t.Parallel()

	// Tests tracking with three levels of nesting:
	// Level1 -> Level2 -> Level3
	// The original ref for Level3 would be Level1 (the start of the chain)

	registry := NewNestedReferenceRegistry()

	// Track Level3 with Level1 as the original reference
	tracked := registry.Track(
		"#/components/schemas/Level3",
		"#/components/schemas/Level1",
	)
	assert.True(t, tracked)

	// Verify it's tracked correctly
	assert.Equal(t, 1, registry.Len())

	originalRef := registry.GetOriginalRef("#/components/schemas/Level3")
	assert.Equal(t, "#/components/schemas/Level1", originalRef)
}

func TestGlobalNestedRefRegistry(t *testing.T) {
	// Reset global state before and after test
	ResetGlobalNestedRefRegistry()
	defer ResetGlobalNestedRefRegistry()

	global := GlobalNestedRefRegistry()
	require.NotNil(t, global)

	// Verify it's a working registry
	tracked := global.Track(
		"#/components/schemas/Test",
		"#/components/schemas/Ref",
	)
	assert.True(t, tracked)
	assert.Equal(t, 1, global.Len())

	// Reset should clear it
	ResetGlobalNestedRefRegistry()
	assert.Equal(t, 0, global.Len())
}
