package validation_test

import (
	"context"
	"os"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Validate_References(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")

	data, err := os.ReadFile("testdata/references/openapi.yaml")
	require.NoError(t, err)
	require.NotNil(t, data)

	v, err := validation.NewValidator(&config.Configuration{
		Generation: config.Generation{},
	}, validation.RulesetSpeakeasyRecommended)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), data, "testdata/references/openapi.yaml", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	filteredErrs := []error{}
	for _, e := range errs {
		var ve *errors.ValidationError
		if errors.As(e, &ve) && ve.Severity == errors.SeverityHint {
			// remove hints
			continue
		}

		filteredErrs = append(filteredErrs, e)
	}
	assert.Empty(t, filteredErrs)
}
