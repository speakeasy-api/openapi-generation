package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityDescription_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description *EntityDescription
		testFunc    func(t *testing.T, original, cloned *EntityDescription)
	}{
		{
			name:        "nil description",
			description: nil,
			testFunc: func(t *testing.T, original, cloned *EntityDescription) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:        "empty description",
			description: &EntityDescription{},
			testFunc: func(t *testing.T, original, cloned *EntityDescription) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Empty(t, cloned.TerraformDataResource)
				assert.Empty(t, cloned.TerraformEphemeralResource)
				assert.Empty(t, cloned.TerraformManagedResource)
			},
		},
		{
			name: "description with values",
			description: &EntityDescription{
				TerraformDataResource:      "data_resource_desc",
				TerraformEphemeralResource: "ephemeral_resource_desc",
				TerraformManagedResource:   "managed_resource_desc",
			},
			testFunc: func(t *testing.T, original, cloned *EntityDescription) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.TerraformDataResource, cloned.TerraformDataResource)
				assert.Equal(t, original.TerraformEphemeralResource, cloned.TerraformEphemeralResource)
				assert.Equal(t, original.TerraformManagedResource, cloned.TerraformManagedResource)

				// Verify deep copy by modifying original
				original.TerraformDataResource = "modified"
				assert.Equal(t, "data_resource_desc", cloned.TerraformDataResource)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.description.Clone()
			tt.testFunc(t, tt.description, cloned)
		})
	}
}
