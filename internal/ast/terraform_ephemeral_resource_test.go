package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformEphemeralResource_SchemaDescription(t *testing.T) {
	t.Parallel()

	t.Run("returns description when set", func(t *testing.T) {
		t.Parallel()

		ephemeralResource := NewTerraformEphemeralResource("example")
		ephemeralResource.Description = "Custom ephemeral resource description"

		assert.Equal(t, "Custom ephemeral resource description", ephemeralResource.SchemaDescription())
	})

	t.Run("returns default when description is empty", func(t *testing.T) {
		t.Parallel()

		ephemeralResource := NewTerraformEphemeralResource("example")

		assert.Equal(t, "Example Ephemeral Resource", ephemeralResource.SchemaDescription())
	})

	t.Run("sanitizes name in default description", func(t *testing.T) {
		t.Parallel()

		ephemeralResource := NewTerraformEphemeralResource("my_example_thing")

		assert.Equal(t, "MyExampleThing Ephemeral Resource", ephemeralResource.SchemaDescription())
	})
}
