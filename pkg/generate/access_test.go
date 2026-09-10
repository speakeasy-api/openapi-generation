package generate

import (
	"context"
	"testing"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withLicenseVerification(t *testing.T, accountType shared.AccountType, targets []string) {
	t.Helper()
	previous := resolveLicenseToken
	resolveLicenseToken = func(ctx context.Context) (context.Context, *licensetoken.Verification, error) {
		createdAt := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
		resolved, err := generationaccess.WithAuthenticated(ctx, generationaccess.AuthenticatedInfo{
			AccountType:        accountType,
			WorkspaceCreatedAt: &createdAt,
			WorkspaceID:        "workspace-id",
			GeneratedLicense:   generationaccess.GeneratedLicenseCommercial,
		})
		return resolved, &licensetoken.Verification{Targets: append([]string(nil), targets...)}, err
	}
	t.Cleanup(func() { resolveLicenseToken = previous })
}

func TestResolveGenerationAccess(t *testing.T) {
	t.Run("missing state is rejected", func(t *testing.T) {
		_, err := resolveGenerationAccess(context.Background(), "go", true)

		require.ErrorIs(t, err, ErrMissingGenerationAccess)
	})

	t.Run("direct state is accepted", func(t *testing.T) {
		ctx, err := resolveGenerationAccess(generationaccess.WithDirect(context.Background()), "go", true)

		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(ctx)
		require.True(t, ok)
		assert.Empty(t, state.GeneratedLicense(), "no license election is ever stamped implicitly")
	})

	t.Run("direct state with an AGPL election carries it", func(t *testing.T) {
		ctx, err := resolveGenerationAccess(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), "go", true)

		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, generationaccess.GeneratedLicenseAGPL, state.GeneratedLicense())
	})

	t.Run("authenticated commercial state without a validated token is rejected", func(t *testing.T) {
		createdAt := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
		ctx, err := generationaccess.WithAuthenticated(context.Background(), generationaccess.AuthenticatedInfo{
			AccountType:        shared.AccountTypeEnterprise,
			WorkspaceCreatedAt: &createdAt,
			WorkspaceID:        "workspace-id",
			GeneratedLicense:   generationaccess.GeneratedLicenseCommercial,
		})
		require.NoError(t, err)

		_, err = resolveGenerationAccess(ctx, "go", true)
		require.ErrorIs(t, err, ErrUnprovenCommercialLicense)
	})

	t.Run("authenticated AGPL state needs no token", func(t *testing.T) {
		createdAt := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
		ctx, err := generationaccess.WithAuthenticated(context.Background(), generationaccess.AuthenticatedInfo{
			AccountType:        shared.AccountTypeFree,
			WorkspaceCreatedAt: &createdAt,
			WorkspaceID:        "workspace-id",
			GeneratedLicense:   generationaccess.GeneratedLicenseAGPL,
		})
		require.NoError(t, err)

		resolved, err := resolveGenerationAccess(ctx, "go", true)
		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(resolved)
		require.True(t, ok)
		assert.Equal(t, generationaccess.GeneratedLicenseAGPL, state.GeneratedLicense())
	})

	t.Run("validated token establishes verified commercial state", func(t *testing.T) {
		token := requireLicenseToken(t)

		resolved, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), token), "go", true)
		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(resolved)
		require.True(t, ok)
		assert.Equal(t, generationaccess.ModeAuthenticated, state.Mode())
		assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
	})

	t.Run("license token is validated inside the generator and cannot be bypassed", func(t *testing.T) {
		ctx := licensetoken.WithToken(generationaccess.WithDirect(context.Background()), []byte("not-a-real-token"))

		_, err := resolveGenerationAccess(ctx, "go", true)
		require.Error(t, err)
		assert.ErrorIs(t, err, licensetoken.ErrInvalidLicenseToken)
	})

	t.Run("named target is covered", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"go"})

		resolved, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "go", true)
		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(resolved)
		require.True(t, ok)
		assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
	})

	t.Run("versioned template name is checked as its target", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"python"})

		_, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "pythonv2", true)
		require.NoError(t, err)

		_, err = resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "typescriptv2", true)
		require.ErrorIs(t, err, ErrLicenseTargetNotCovered)
		assert.Contains(t, err.Error(), `target "typescript", token covers [python]`)
	})

	t.Run("uncovered named target is rejected", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"go"})

		_, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "typescript", true)
		require.ErrorIs(t, err, ErrLicenseTargetNotCovered)
		assert.Contains(t, err.Error(), `target "typescript", token covers [go]`)
	})

	t.Run("wildcard covers any target", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"*"})

		_, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "typescript", true)
		require.NoError(t, err)
	})

	t.Run("free tier establishes commercial generation state for its target", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeFree, []string{"go"})

		resolved, err := resolveGenerationAccess(licensetoken.WithToken(context.Background(), []byte("test-token")), "go", true)
		require.NoError(t, err)
		state, ok := generationaccess.StateFromContext(resolved)
		require.True(t, ok)
		accountType, ok := state.AccountType()
		require.True(t, ok)
		assert.Equal(t, shared.AccountTypeFree, accountType)
		assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
	})
}
