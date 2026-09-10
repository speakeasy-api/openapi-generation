package security

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// Returns true if the OAuth2 Client Credentials flow is enabled by checking:
//   - If the account has access to the oauth2ClientCredentials feature.
//   - If the target supports the oauth2ClientCredentials feature.
//   - If the generation configuration has auth.oAuth2ClientCredentialsEnabled
//     enabled.
func IsOAuth2ClientCredentialsEnabled(ctx context.Context, subsystem *subsystem.Subsystem) bool {
	feature := features.FeatureOAuth2ClientCredentials
	accessErr := licensing.ValidateAccountHasFeatureAccess(ctx, feature)
	featureSupported := subsystem.Features.IsFeatureSupported(ctx, feature)
	enabled := subsystem.Config.Generation.Auth.OAuth2ClientCredentialsEnabled

	if !featureSupported && enabled && accessErr == nil {
		logging.LogWarning(ctx, "OAuth2 client credentials flow support is enabled but not available in this target.", errors.NewUnsupportedError("OAuth2 client credentials support is not available in this target", nil))
	}
	if featureSupported && enabled && accessErr != nil {
		msg := "OAuth2 client credentials flow support is enabled but " + accessErr.Error()
		logging.LogWarning(ctx, msg, errors.NewUnsupportedError(msg, nil))
	}

	return accessErr == nil && featureSupported && enabled
}

// Returns true if the OAuth2 Resource Owner Password flow is enabled by checking:
// - If the account has access to the oauth2Password feature.
// - If the target supports the oauth2Password feature.
// - If the generation configuration has auth.oAuth2PasswordEnabled enabled.
func IsOAuth2PasswordEnabled(ctx context.Context, subsystem *subsystem.Subsystem) bool {
	feature := features.FeatureOAuth2Password
	accessErr := licensing.ValidateAccountHasFeatureAccess(ctx, feature)
	featureSupported := subsystem.Features.IsFeatureSupported(ctx, feature)
	enabled := subsystem.Config.Generation.Auth.OAuth2PasswordEnabled

	if !featureSupported && enabled && accessErr == nil {
		logging.LogWarning(
			ctx,
			"OAuth2 password flow support is enabled but not available in this target.",
			errors.NewUnsupportedError("OAuth2 password flow support is not available in this target", nil),
		)
	}
	if featureSupported && enabled && accessErr != nil {
		msg := "OAuth2 password flow support is enabled but " + accessErr.Error()
		logging.LogWarning(ctx, msg, errors.NewUnsupportedError(msg, nil))
	}

	return accessErr == nil && featureSupported && enabled
}

func isCustomSecuritySchemesAvailable(ctx context.Context, subsystem *subsystem.Subsystem) bool {
	hooksSupported := subsystem.Features.IsFeatureSupported(ctx, features.FeatureSDKHooks)
	customSecuritySupported := subsystem.Features.IsFeatureSupported(ctx, features.FeatureCustomSecuritySchemes)

	if !hooksSupported || !customSecuritySupported {
		return false
	}

	customSecurityAccountAccess := licensing.AccountHasFeatureAccess(ctx, features.FeatureCustomSecuritySchemes)
	hooksAccountAccess := licensing.AccountHasFeatureAccess(ctx, features.FeatureSDKHooks)

	return customSecurityAccountAccess && hooksAccountAccess
}
