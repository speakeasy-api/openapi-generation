package versioning

import (
	"context"
)

// MockServerVersionInfo is a subset of VersionInfo that is only relevant for
// the mockserver target.
type MockServerVersionInfo struct {
	FeatureVersions         map[string]string
	PreviousFeatureVersions map[string]string
}

// ShouldMockServerRegenerate returns true if the mockserver target has relevant
// feature changes that should cause target regeneration.
func ShouldMockServerRegenerate(ctx context.Context, versionInfo MockServerVersionInfo) (bool, error) {
	featureVersionBumpType, err := getFeatureVersionBump(ctx, versionInfo.FeatureVersions, versionInfo.PreviousFeatureVersions)

	if err != nil {
		return false, err
	}

	return featureVersionBumpType != BumpNone, nil
}
