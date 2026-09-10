package licensing

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
)

type FeatureAccess struct {
	// If not set, all tiers are permitted
	permittedTiers []shared.AccountType
	// If not set, no feature flags are required
	requiredAddOn []shared.BillingAddOn
	// If not set, the feature is always permitted
	enforcementStart *time.Time
}

var sseEnforcementStart = time.Date(2024, time.October, 23, 0, 0, 0, 0, time.UTC)

var featureLicenseConfig = map[features.Feature]FeatureAccess{
	features.FeatureSDKHooks: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	features.FeatureOAuth2ClientCredentials: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	features.FeatureOAuth2Password: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	features.FeatureCustomSecuritySchemes: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	features.FeatureServerEvents: {
		permittedTiers:   []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
		enforcementStart: &sseEnforcementStart,
	},
	features.FeatureJsonlResponses: {
		permittedTiers: []shared.AccountType{shared.AccountTypeEnterprise},
	},
	features.FeatureWebhooks: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	features.FeatureWebhookHandlers: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
		requiredAddOn:  []shared.BillingAddOn{shared.BillingAddOnWebhooks},
	},
	features.FeatureReactQueryHooks: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
	},
	// Testing features
	features.FeatureTests: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
		requiredAddOn:  []shared.BillingAddOn{shared.BillingAddOnSDKTesting},
	},
	features.FeatureMockServer: {
		permittedTiers: []shared.AccountType{shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
		requiredAddOn:  []shared.BillingAddOn{shared.BillingAddOnSDKTesting},
	},
	features.FeatureCustomCodeRegions: {
		permittedTiers: []shared.AccountType{shared.AccountTypeScaleUp, shared.AccountTypeBusiness, shared.AccountTypeEnterprise},
		requiredAddOn:  []shared.BillingAddOn{shared.BillingAddOnCustomCodeRegions},
	},
}

func AccountHasFeatureAccess(ctx context.Context, feat features.Feature) bool {
	return ValidateAccountHasFeatureAccess(ctx, feat) == nil
}

var ErrMissingGenerationAccess = errors.New("generation access state is required")

func ValidateAccountHasFeatureAccess(ctx context.Context, feat features.Feature) error {
	state, ok := generationaccess.StateFromContext(ctx)
	if !ok {
		return ErrMissingGenerationAccess
	}
	if state.Mode() == generationaccess.ModeDirect {
		return nil
	}

	currentTier, ok := state.AccountType()
	if !ok {
		return ErrMissingGenerationAccess
	}
	workspaceCreatedAt := state.WorkspaceCreatedAt()
	license, exists := featureLicenseConfig[feat]
	meetsEnforcementStartTime := license.enforcementStart == nil || workspaceCreatedAt == nil || workspaceCreatedAt.After(*license.enforcementStart)

	// Default to all tiers are permitted
	var permittedTiers = []shared.AccountType{
		shared.AccountTypeFree,
		shared.AccountTypeScaleUp,
		shared.AccountTypeBusiness,
		shared.AccountTypeEnterprise,
	}

	// Otherwise, use the specified tiers
	if exists && len(license.permittedTiers) > 0 && meetsEnforcementStartTime {
		permittedTiers = license.permittedTiers
	}

	if !slices.Contains(permittedTiers, currentTier) {
		return fmt.Errorf("feature %s requires at least tier %q - contact us to upgrade", feat.String(), findMinimumTier(permittedTiers))
	}

	for _, addOn := range license.requiredAddOn {
		if !state.HasBillingAddOn(addOn) {
			return fmt.Errorf("feature %s is not enabled - try running speakeasy billing activate --feature %s", string(addOn), string(addOn))
		}
	}

	return nil
}

func findMinimumTier(tiers []shared.AccountType) shared.AccountType {
	orderedTiers := []shared.AccountType{
		shared.AccountTypeFree,
		shared.AccountTypeScaleUp,
		shared.AccountTypeBusiness,
		shared.AccountTypeEnterprise,
	}

	for _, tier := range tiers {
		if slices.Contains(orderedTiers, tier) {
			return tier
		}
	}

	return shared.AccountTypeFree
}
