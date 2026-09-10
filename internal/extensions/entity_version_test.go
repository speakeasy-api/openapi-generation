package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityVersion_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  *EntityVersion
		testFunc func(t *testing.T, original, cloned *EntityVersion)
	}{
		{
			name:    "nil version",
			version: nil,
			testFunc: func(t *testing.T, original, cloned *EntityVersion) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:    "zero version",
			version: &EntityVersion{},
			testFunc: func(t *testing.T, original, cloned *EntityVersion) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, int64(0), cloned.TerraformManagedResource)
			},
		},
		{
			name: "version with value",
			version: &EntityVersion{
				TerraformManagedResource: 42,
			},
			testFunc: func(t *testing.T, original, cloned *EntityVersion) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.TerraformManagedResource, cloned.TerraformManagedResource)

				// Verify deep copy by modifying original
				original.TerraformManagedResource = 999
				assert.Equal(t, int64(42), cloned.TerraformManagedResource)
			},
		},
		{
			name: "negative version",
			version: &EntityVersion{
				TerraformManagedResource: -1,
			},
			testFunc: func(t *testing.T, original, cloned *EntityVersion) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, int64(-1), cloned.TerraformManagedResource)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.version.Clone()
			tt.testFunc(t, tt.version, cloned)
		})
	}
}
