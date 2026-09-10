package types_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_TargetFromTemplate(t *testing.T) {
	noSuffix := "typescript"
	withSuffix := "typescriptv2"

	// The resolved target should be the same for suffix vs no suffix
	assert.Equal(t, "typescript", types.NewTargetFromTemplate(noSuffix).Target)
	assert.Equal(t, "typescript", types.NewTargetFromTemplate(withSuffix).Target)

	// The template directory will be different for suffix vs no suffix
	assert.Equal(t, "typescript", types.NewTargetFromTemplate(noSuffix).Template)
	assert.Equal(t, "typescriptv2", types.NewTargetFromTemplate(withSuffix).Template)
}
