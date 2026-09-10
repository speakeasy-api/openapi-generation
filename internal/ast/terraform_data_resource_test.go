package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformDataResource_SchemaDescription(t *testing.T) {
	t.Parallel()

	t.Run("returns description when set", func(t *testing.T) {
		t.Parallel()

		dataResource := NewTerraformDataResource("example")
		dataResource.Description = "Custom data source description"

		assert.Equal(t, "Custom data source description", dataResource.SchemaDescription())
	})

	t.Run("returns default when description is empty", func(t *testing.T) {
		t.Parallel()

		dataResource := NewTerraformDataResource("example")

		assert.Equal(t, "Example DataSource", dataResource.SchemaDescription())
	})

	t.Run("sanitizes name in default description", func(t *testing.T) {
		t.Parallel()

		dataResource := NewTerraformDataResource("my_example_thing")

		assert.Equal(t, "MyExampleThing DataSource", dataResource.SchemaDescription())
	})
}
