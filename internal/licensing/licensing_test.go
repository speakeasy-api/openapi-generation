package licensing

import (
	"context"
	"testing"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAccountHasFeatureAccess(t *testing.T) {
	t.Run("missing state is rejected", func(t *testing.T) {
		err := ValidateAccountHasFeatureAccess(context.Background(), features.FeatureSDKHooks)

		require.ErrorIs(t, err, ErrMissingGenerationAccess)
	})

	t.Run("direct state permits every licensed feature", func(t *testing.T) {
		ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

		for feature := range featureLicenseConfig {
			require.NoError(t, ValidateAccountHasFeatureAccess(ctx, feature), feature.String())
		}
	})

	t.Run("authenticated free remains restricted", func(t *testing.T) {
		ctx := authenticatedContext(t, shared.AccountTypeFree, nil, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))

		assert.Error(t, ValidateAccountHasFeatureAccess(ctx, features.FeatureSDKHooks))
	})

	t.Run("authenticated add-ons and enforcement start are preserved", func(t *testing.T) {
		beforeEnforcement := authenticatedContext(t, shared.AccountTypeFree, nil, time.Date(2024, time.October, 22, 23, 59, 59, 0, time.UTC))
		afterEnforcement := authenticatedContext(t, shared.AccountTypeFree, nil, time.Date(2024, time.October, 23, 0, 0, 1, 0, time.UTC))
		businessWithTests := authenticatedContext(t, shared.AccountTypeBusiness, []shared.BillingAddOn{shared.BillingAddOnSDKTesting}, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))

		require.NoError(t, ValidateAccountHasFeatureAccess(beforeEnforcement, features.FeatureServerEvents))
		require.Error(t, ValidateAccountHasFeatureAccess(afterEnforcement, features.FeatureServerEvents))
		require.NoError(t, ValidateAccountHasFeatureAccess(businessWithTests, features.FeatureTests))
	})
}

func authenticatedContext(t *testing.T, accountType shared.AccountType, addOns []shared.BillingAddOn, createdAt time.Time) context.Context {
	t.Helper()

	ctx, err := generationaccess.WithAuthenticated(context.Background(), generationaccess.AuthenticatedInfo{
		AccountType:        accountType,
		BillingAddOns:      addOns,
		WorkspaceCreatedAt: &createdAt,
		WorkspaceID:        "workspace-id",
		GeneratedLicense:   generationaccess.GeneratedLicenseCommercial,
	})
	require.NoError(t, err)
	return ctx
}

func TestAccountHasFeatureAccessReturnsFalseForMissingGenerationAccess(t *testing.T) {
	assert.False(t, AccountHasFeatureAccess(context.Background(), features.FeatureSDKHooks))
	assert.ErrorIs(t, ValidateAccountHasFeatureAccess(context.Background(), features.FeatureSDKHooks), ErrMissingGenerationAccess)
}
