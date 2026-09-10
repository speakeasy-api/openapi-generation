package versioning

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/speakeasy-api/versioning-reports/versioning"

	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

var (
	firstVersion       = version.Must(version.NewVersion("0.0.1"))
	v1                 = version.Must(version.NewVersion("1.0.0"))
	BumpOverrideEnvVar = "SPEAKEASY_BUMP_OVERRIDE"
)

type BumpType string

const (
	BumpMajor    BumpType = "major"
	BumpMinor    BumpType = "minor"
	BumpPatch    BumpType = "patch"
	BumpGraduate BumpType = "graduate"
	BumpNone     BumpType = "none"
)

type VersionInfo struct {
	SDKVersion         string
	PreviousSDKVersion string

	DocVersion          string
	PreviousDocVersion  string
	DocChecksum         string
	PreviousDocChecksum string

	FeatureVersions         map[string]string
	AvailableFeatures       map[string]bool
	PreviousFeatureVersions map[string]string

	ConfigChecksum         string
	PreviousConfigChecksum string
}

func GetNewSDKVersion(ctx context.Context, versionInfo VersionInfo, force bool, lang string, skipVersioning bool) (string, versioning.BumpType, error) {
	enrichTelemetryWithVersionInfo(ctx, versionInfo)
	if skipVersioning || (versionInfo.SDKVersion != versionInfo.PreviousSDKVersion) {
		logging.From(ctx).Info("versioning: custom SDK version detected, using custom version", zap.String("previous", versionInfo.PreviousSDKVersion), zap.String("current", versionInfo.SDKVersion))
		bumpType := DetermineBumpType(versionInfo.PreviousSDKVersion, versionInfo.SDKVersion)
		enrichTelemetryWithBump(ctx, shared.GenerateBumpType(bumpType))
		return versionInfo.SDKVersion, bumpType, nil
	}

	bumpMajor := false
	bumpMinor := false
	bumpPatch := false
	bumpGraduate := false

	checkBump := func(bumpType BumpType) {
		switch bumpType {
		case BumpMajor:
			bumpMajor = true
		case BumpMinor:
			bumpMinor = true
		case BumpPatch:
			bumpPatch = true
		case BumpGraduate:
			bumpGraduate = true
		}
	}

	var sdkV *version.Version
	var isPreV1 bool

	if versionInfo.SDKVersion != "" {
		var err error
		sdkV, err = version.NewVersion(versionInfo.SDKVersion)
		if err != nil {
			return "", versioning.BumpNone, fmt.Errorf("error parsing sdk version %s: %w", versionInfo.SDKVersion, err)
		}

		isPreV1 = sdkV.LessThan(v1)
	} else {
		enrichTelemetryWithBump(ctx, shared.GenerateBumpTypeCustom)
		return firstVersion.String(), versioning.BumpCustom, nil
	}

	featureVersionBump, err := getFeatureVersionBump(ctx, versionInfo.FeatureVersions, versionInfo.PreviousFeatureVersions)
	if err != nil {
		return "", versioning.BumpNone, err
	}
	checkBump(featureVersionBump)

	documentVersionBump := getDocumentVersionBump(ctx, versionInfo.DocVersion, versionInfo.PreviousDocVersion, versionInfo.DocChecksum, versionInfo.PreviousDocChecksum)
	checkBump(documentVersionBump)

	configVersionBump := getConfigVersionBump(ctx, versionInfo.ConfigChecksum, versionInfo.PreviousConfigChecksum)
	checkBump(configVersionBump)

	// allow bump type to be overwritten by environment variable, supports label based versioning
	var explicitlyAppliedBump bool
	if envVarBump := envVarBumpType(); envVarBump != BumpNone {
		// clear out existing entries to prioritize the env var override
		// graduate doesn't apply here since it's a pre-release specific type of bump
		if envVarBump == BumpPatch || envVarBump == BumpMinor || envVarBump == BumpMajor {
			bumpMajor = false
			bumpMinor = false
			bumpPatch = false
		}
		checkBump(envVarBump)
		explicitlyAppliedBump = true
	}

	if force {
		logging.From(ctx).Info("versioning: generation forced, bumping patch version")
		bumpPatch = true
	}

	if bumpMajor || bumpMinor || bumpPatch {
		major := sdkV.Segments()[0]
		minor := sdkV.Segments()[1]
		patch := sdkV.Segments()[2]
		preRelease := sdkV.Prerelease()

		// We are assuming breaking changes are okay pre v1
		if bumpMajor && !explicitlyAppliedBump {
			if isPreV1 || lang == "go" {
				if isPreV1 {
					logging.From(ctx).Info("versioning: breaking change detected, but pre v1, bumping minor version, manually bump to major if needed")
				} else {
					logging.From(ctx).Info("versioning: breaking change detected, but only bumping minor version, manually bump to major if needed")
				}
				bumpMajor = false
				bumpMinor = true
			}
		}

		bumpType := versioning.BumpNone
		switch {
		case preRelease != "":
			if bumpGraduate {
				enrichTelemetryWithBump(ctx, shared.GenerateBumpTypeGraduate)
				bumpType = versioning.BumpGraduate
				return fmt.Sprintf("%d.%d.%d", major, minor, patch), bumpType, nil
			}

			preReleaseSplit := strings.Split(preRelease, ".")
			preReleaseCounter := 1
			if len(preReleaseSplit) > 1 {
				if num, err := strconv.Atoi(preReleaseSplit[1]); err == nil {
					num++
					preReleaseCounter = num
				}
			}
			// TODO: Add bump type for pre-release
			enrichTelemetryWithBump(ctx, shared.GenerateBumpTypePrerelease)
			bumpType = versioning.BumpPrerelease
			return fmt.Sprintf("%d.%d.%d-%s.%d", major, minor, patch, preReleaseSplit[0], preReleaseCounter), bumpType, nil
		case bumpMajor:
			logging.From(ctx).Info("versioning: breaking change detected, bumping major version")
			major++
			minor = 0
			patch = 0
			bumpType = versioning.BumpMajor
			enrichTelemetryWithBump(ctx, shared.GenerateBumpTypeMajor)
		case bumpMinor:
			logging.From(ctx).Info("versioning: feature level change detected, bumping minor version")
			minor++
			patch = 0
			bumpType = versioning.BumpMinor
			enrichTelemetryWithBump(ctx, shared.GenerateBumpTypeMinor)
		case bumpPatch:
			logging.From(ctx).Info("versioning: fix level change detected, bumping patch version")
			patch++
			bumpType = versioning.BumpPatch
			enrichTelemetryWithBump(ctx, shared.GenerateBumpTypePatch)
		}

		return fmt.Sprintf("%d.%d.%d", major, minor, patch), bumpType, nil
	}

	return versionInfo.SDKVersion, versioning.BumpNone, nil
}

func DetermineBumpType(oldVersion string, newVersion string) versioning.BumpType {
	oldV, err := version.NewVersion(oldVersion)
	if err != nil {
		return versioning.BumpCustom
	}

	newV, err := version.NewVersion(newVersion)
	if err != nil {
		return versioning.BumpCustom
	}
	if newV.Equal(oldV) {
		return versioning.BumpNone
	}
	if !newV.GreaterThan(oldV) {
		return versioning.BumpCustom
	}

	// If the new version is a pre-release, it's a pre-release bump
	if len(newV.Prerelease()) > 0 {
		return versioning.BumpPrerelease
	}
	// If we're migrating from a pre-release to a normal version, it's a graduate version
	if len(newV.Prerelease()) == 0 && len(oldV.Prerelease()) > 0 {
		return versioning.BumpGraduate
	}
	if len(newV.Segments()) != len(oldV.Segments()) || len(newV.Segments()) != 3 {
		return versioning.BumpCustom
	}
	if newV.Segments()[0] > oldV.Segments()[0] {
		return versioning.BumpMajor
	}
	if newV.Segments()[1] > oldV.Segments()[1] {
		return versioning.BumpMinor
	}
	if newV.Segments()[2] > oldV.Segments()[2] {
		return versioning.BumpPatch
	}

	return versioning.BumpCustom
}

func envVarBumpType() BumpType {
	bumpType := os.Getenv(BumpOverrideEnvVar)
	switch bumpType {
	case "major":
		return BumpMajor
	case "minor":
		return BumpMinor
	case "patch":
		return BumpPatch
	case "graduate":
		return BumpGraduate
	default:
		return BumpNone
	}
}

func enrichTelemetryWithBump(ctx context.Context, bumpType shared.GenerateBumpType) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}

	event.GenerateBumpType = &bumpType
}

func enrichTelemetryWithVersionInfo(ctx context.Context, info VersionInfo) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}

	if info.PreviousConfigChecksum != "" {
		event.GenerateConfigPreChecksum = new(string)
		*event.GenerateConfigPreChecksum = info.PreviousConfigChecksum
	}

	if info.ConfigChecksum != "" {
		event.GenerateConfigPostChecksum = new(string)
		*event.GenerateConfigPostChecksum = info.ConfigChecksum
	}

	if info.PreviousDocChecksum != "" {
		event.GenerateGenLockPreDocChecksum = new(string)
		*event.GenerateGenLockPreDocChecksum = info.PreviousDocChecksum
	}

	if info.DocChecksum != "" {
		event.ManagementDocChecksum = new(string)
		*event.ManagementDocChecksum = info.DocChecksum
	}

	if info.DocVersion != "" {
		event.ManagementDocVersion = new(string)
		*event.ManagementDocVersion = info.DocVersion
	}

	if info.PreviousDocVersion != "" {
		event.GenerateGenLockPreDocVersion = new(string)
		*event.GenerateGenLockPreDocVersion = info.PreviousDocVersion
	}

	rawFeatures, err := yaml.Marshal(info.FeatureVersions)
	if err == nil {
		event.GenerateGenLockPostFeatures = new(string)
		*event.GenerateGenLockPostFeatures = string(rawFeatures)
	}

	availableFeatures, err := yaml.Marshal(info.AvailableFeatures)
	if err == nil {
		event.GenerateEligibleFeatures = new(string)
		*event.GenerateEligibleFeatures = string(availableFeatures)
	}

	preFeatures, err := yaml.Marshal(info.PreviousFeatureVersions)
	if err == nil {
		event.GenerateGenLockPreFeatures = new(string)
		*event.GenerateGenLockPreFeatures = string(preFeatures)
	}

	if info.PreviousSDKVersion != "" {
		event.GenerateGenLockPreVersion = new(string)
		*event.GenerateGenLockPreVersion = info.PreviousSDKVersion
	}
}

func getFeatureVersionBump(ctx context.Context, current, previous map[string]string) (BumpType, error) {
	bumpMajor := false
	bumpMinor := false
	bumpPatch := false

	previous = copyMap(previous)

	for feature, currentVersion := range current {
		previousVersion, ok := previous[feature]
		if !ok {
			// New feature
			bumpMinor = true
			continue
		}

		delete(previous, feature)

		if currentVersion == previousVersion {
			continue
		}

		currentV, err := version.NewVersion(currentVersion)
		if err != nil {
			return BumpNone, fmt.Errorf("failed to parse current feature version %s for feature %s: %w", currentVersion, feature, err)
		}

		previousV, err := version.NewVersion(previousVersion)
		if err != nil {
			return BumpNone, fmt.Errorf("failed to parse previous feature version %s for feature %s: %w", previousVersion, feature, err)
		}

		if currentV.GreaterThan(previousV) {
			var bumpType BumpType
			switch {
			case currentV.Segments()[0] > previousV.Segments()[0]:
				bumpMajor = true
				bumpType = BumpMajor
			case currentV.Segments()[1] > previousV.Segments()[1]:
				bumpMinor = true
				bumpType = BumpMinor
			default:
				bumpPatch = true
				bumpType = BumpPatch
			}

			logging.From(ctx).Info(fmt.Sprintf("versioning: feature version changed detected, bumping %s version", bumpType), zap.String("feature", feature), zap.String("previous", previousVersion), zap.String("current", currentVersion))
		}
	}

	if len(previous) > 0 {
		// Features removed so we will bump minor for now, maybe breaking though?
		bumpMinor = true
		logging.From(ctx).Info("versioning: feature removal detected, bumping minor version")
	}

	if bumpMajor {
		return BumpMajor, nil
	}

	if bumpMinor {
		return BumpMinor, nil
	}

	if bumpPatch {
		return BumpPatch, nil
	}

	return BumpNone, nil
}

func getDocumentVersionBump(ctx context.Context, currentVersion, previousVersion, currentChecksum, previousChecksum string) BumpType {
	bumpMajor := false
	bumpMinor := false
	bumpPatch := false
	documentVersionHandled := false

	if currentVersion != previousVersion {
		currentV, currentErr := version.NewVersion(currentVersion)
		previousV, previousErr := version.NewVersion(previousVersion)

		// We only deal with semver versions otherwise we just deal with the checksum
		if currentErr == nil && previousErr == nil {
			if currentV.GreaterThan(previousV) {
				switch {
				case currentV.Segments()[0] > previousV.Segments()[0]:
					bumpMajor = true
				case currentV.Segments()[1] > previousV.Segments()[1]:
					bumpMinor = true
				default:
					bumpPatch = true
				}

				documentVersionHandled = true
			}
		} else {
			logging.From(ctx).Info("versioning: document doesn't use semver, falling back to checksum", zap.String("previous", previousVersion), zap.String("current", currentVersion))
		}
	}

	if !documentVersionHandled && currentChecksum != previousChecksum {
		logging.From(ctx).Info("versioning: OpenAPI document checksum changed detected, bumping patch version", zap.String("previousChecksum", previousChecksum), zap.String("currentChecksum", currentChecksum))
		logging.From(ctx).Warn("versioning: OpenAPI document checksum changed but version did not", zap.String("previousVersion", previousVersion), zap.String("currentVersion", currentVersion))
		bumpPatch = true
	}

	if bumpMajor {
		return BumpMajor
	}

	if bumpMinor {
		return BumpMinor
	}

	if bumpPatch {
		return BumpPatch
	}

	return BumpNone
}

func getConfigVersionBump(ctx context.Context, currentChecksum, previousChecksum string) BumpType {
	if currentChecksum != previousChecksum {
		logging.From(ctx).Warn("versioning: config checksum changed detected, consider bumping major version if an interface change is noticed. Bumping patch", zap.String("previousChecksum", previousChecksum), zap.String("currentChecksum", currentChecksum))
		return BumpPatch
	}

	return BumpNone
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	c := make(map[K]V, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}
