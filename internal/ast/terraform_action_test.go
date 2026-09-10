package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformAction_SchemaDescription(t *testing.T) {
	t.Parallel()

	t.Run("returns description when set", func(t *testing.T) {
		t.Parallel()

		action := NewTerraformAction("example")
		action.Description = "Custom action description"

		assert.Equal(t, "Custom action description", action.SchemaDescription())
	})

	t.Run("returns default when description is empty", func(t *testing.T) {
		t.Parallel()

		action := NewTerraformAction("example")

		assert.Equal(t, "Example Action", action.SchemaDescription())
	})

	t.Run("sanitizes name in default description", func(t *testing.T) {
		t.Parallel()

		action := NewTerraformAction("my_example_thing")

		assert.Equal(t, "MyExampleThing Action", action.SchemaDescription())
	})
}
