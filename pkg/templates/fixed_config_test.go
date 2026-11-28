package templates

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFixedConfigWithoutHook(t *testing.T) {
	t.Parallel()

	for _, template := range []string{"go", "csharp", "typescriptv2"} {
		t.Run(template, func(t *testing.T) {
			t.Parallel()

			fixed, err := GetFixedConfig(types.NewTargetFromTemplate(template))
			require.NoError(t, err)
			assert.Nil(t, fixed)
		})
	}
}
