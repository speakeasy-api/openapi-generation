package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformManagedResource_SchemaDescription(t *testing.T) {
	t.Parallel()

	t.Run("returns description when set", func(t *testing.T) {
		t.Parallel()

		managedResource := NewTerraformManagedResource("example")
		managedResource.Description = "Custom resource description"

		assert.Equal(t, "Custom resource description", managedResource.SchemaDescription())
	})

	t.Run("returns default when description is empty", func(t *testing.T) {
		t.Parallel()

		managedResource := NewTerraformManagedResource("example")

		assert.Equal(t, "Example Resource", managedResource.SchemaDescription())
	})

	t.Run("sanitizes name in default description", func(t *testing.T) {
		t.Parallel()

		managedResource := NewTerraformManagedResource("my_example_thing")

		assert.Equal(t, "MyExampleThing Resource", managedResource.SchemaDescription())
	})
}

func TestTerraformManagedResource_SchemaVersion(t *testing.T) {
	t.Parallel()

	t.Run("returns 0 when schema is nil", func(t *testing.T) {
		t.Parallel()

		managedResource := &TerraformManagedResource{}

		assert.Equal(t, int64(0), managedResource.SchemaVersion())
	})

	t.Run("returns 0 for default schema", func(t *testing.T) {
		t.Parallel()

		managedResource := NewTerraformManagedResource("example")

		assert.Equal(t, int64(0), managedResource.SchemaVersion())
	})

	t.Run("returns configured version", func(t *testing.T) {
		t.Parallel()

		managedResource := NewTerraformManagedResource("example")
		managedResource.Schema.Version = 3

		assert.Equal(t, int64(3), managedResource.SchemaVersion())
	})
}
