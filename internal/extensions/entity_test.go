package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntity_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entity   *Entity
		testFunc func(t *testing.T, original, cloned *Entity)
	}{
		{
			name:   "nil entity",
			entity: nil,
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:   "empty entity",
			entity: &Entity{},
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Nil(t, cloned.Names)
			},
		},
		{
			name: "entity with nil names",
			entity: &Entity{
				Names: nil,
			},
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Nil(t, cloned.Names)
			},
		},
		{
			name: "entity with empty names slice",
			entity: &Entity{
				Names: []string{},
			},
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Names)
				assert.Empty(t, cloned.Names)
				if len(original.Names) > 0 && len(cloned.Names) > 0 {
					assert.NotSame(t, &original.Names[0], &cloned.Names[0])
				}
			},
		},
		{
			name: "entity with single name",
			entity: &Entity{
				Names: []string{"user"},
			},
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Names, cloned.Names)
				for i := range original.Names {
					assert.NotSame(t, &original.Names[i], &cloned.Names[i])
				}

				// Verify deep copy by modifying original
				original.Names[0] = "modified"
				assert.Equal(t, []string{"user"}, cloned.Names)
			},
		},
		{
			name: "entity with multiple names",
			entity: &Entity{
				Names: []string{"user", "account", "profile"},
			},
			testFunc: func(t *testing.T, original, cloned *Entity) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Names, cloned.Names)
				for i := range original.Names {
					assert.NotSame(t, &original.Names[i], &cloned.Names[i])
				}

				// Verify deep copy by modifying original
				original.Names = append(original.Names, "new")
				assert.Equal(t, []string{"user", "account", "profile"}, cloned.Names)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.entity.Clone()
			tt.testFunc(t, tt.entity, cloned)
		})
	}
}
