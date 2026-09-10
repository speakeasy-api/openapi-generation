package versioning

import (
	"context"
	"github.com/speakeasy-api/versioning-reports/versioning"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNewSDKVersion_Success(t *testing.T) {
	type args struct {
		versionInfo VersionInfo
		force       bool
		lang        string
	}
	tests := []struct {
		name         string
		args         args
		wantVersion  string
		wantBumpType versioning.BumpType
	}{
		{
			name: "returns the SDKVersion if it is different from the previous SDKVersion",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "1.0.0",
					PreviousSDKVersion: "0.0.1",
				},
			},
			wantVersion:  "1.0.0",
			wantBumpType: versioning.BumpMajor,
		},
		{
			name: "bumps patch version if a feature patch version is bumped",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					FeatureVersions: map[string]string{
						"feature1": "1.0.2",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.1",
					},
				},
			},
			wantVersion:  "0.0.2",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "bumps minor version if a feature minor version is bumped",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					FeatureVersions: map[string]string{
						"feature1": "1.1.0",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.0",
					},
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps major version if a feature major version is bumped",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "1.0.1",
					PreviousSDKVersion: "1.0.1",
					FeatureVersions: map[string]string{
						"feature1": "2.0.0",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.0",
					},
				},
			},
			wantVersion:  "2.0.0",
			wantBumpType: versioning.BumpMajor,
		},
		{
			name: "doesn't bump major version if a feature major version is bumped but the sdk version is pre v1",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					FeatureVersions: map[string]string{
						"feature1": "2.0.0",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.0",
					},
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps minor version if a feature is removed",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					FeatureVersions: map[string]string{
						"feature1": "1.0.0",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.0",
						"feature2": "1.0.0",
					},
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps minor version if a feature is added",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					FeatureVersions: map[string]string{
						"feature1": "1.0.0",
						"feature2": "1.0.0",
					},
					PreviousFeatureVersions: map[string]string{
						"feature1": "1.0.0",
					},
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps patch if patch version is bumped in the document",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					DocVersion:         "1.0.1",
					PreviousDocVersion: "1.0.0",
				},
			},
			wantVersion:  "0.0.2",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "bumps minor if minor version is bumped in the document",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					DocVersion:         "1.1.0",
					PreviousDocVersion: "1.0.0",
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps major if major version is bumped in the document",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "1.0.0",
					PreviousSDKVersion: "1.0.0",
					DocVersion:         "2.0.0",
					PreviousDocVersion: "1.0.0",
				},
			},
			wantVersion:  "2.0.0",
			wantBumpType: versioning.BumpMajor,
		},
		{
			name: "doesn't bump major version if a document major version is bumped but the sdk version is pre v1",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					DocVersion:         "2.0.0",
					PreviousDocVersion: "1.0.0",
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps patch if only a document checksum change is detected",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:          "0.0.1",
					PreviousSDKVersion:  "0.0.1",
					DocVersion:          "1.0.0",
					PreviousDocVersion:  "1.0.0",
					DocChecksum:         "123",
					PreviousDocChecksum: "456",
				},
			},
			wantVersion:  "0.0.2",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "bumps patch if only a document version isn't a semver and checksum change is detected",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:          "0.0.1",
					PreviousSDKVersion:  "0.0.1",
					DocVersion:          "version 2",
					PreviousDocVersion:  "version 1",
					DocChecksum:         "123",
					PreviousDocChecksum: "456",
				},
			},
			wantVersion:  "0.0.2",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "bumps patch if only a config checksum change is detected",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:             "1.0.0",
					PreviousSDKVersion:     "1.0.0",
					ConfigChecksum:         "123",
					PreviousConfigChecksum: "456",
				},
			},
			wantVersion:  "1.0.1",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "doesn't bump major if a major bump is detected but the sdk version is pre v1",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
					PreviousDocVersion: "0.1.0",
					DocVersion:         "1.0.0",
				},
			},
			wantVersion:  "0.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "doesn't bump major if lang is go and a breaking change is detected",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "1.0.0",
					PreviousSDKVersion: "1.0.0",
					PreviousDocVersion: "1.0.0",
					DocVersion:         "2.0.0",
				},
				lang: "go",
			},
			wantVersion:  "1.1.0",
			wantBumpType: versioning.BumpMinor,
		},
		{
			name: "bumps patch if force is set",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
				},
				force: true,
			},
			wantVersion:  "0.0.2",
			wantBumpType: versioning.BumpPatch,
		},
		{
			name: "if no version is set starts from 0.0.0 and bumps from there",
			args: args{
				versionInfo: VersionInfo{},
				force:       true,
			},
			wantVersion:  "0.0.1",
			wantBumpType: versioning.BumpCustom,
		},
		{
			name: "nothing is bumped if no changes are detected",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "0.0.1",
					PreviousSDKVersion: "0.0.1",
				},
			},
			wantVersion:  "0.0.1",
			wantBumpType: versioning.BumpNone,
		},
		{
			name: "if this is the first release we use the first version",
			args: args{
				versionInfo: VersionInfo{},
			},
			wantVersion:  "0.0.1",
			wantBumpType: versioning.BumpCustom,
		},
		{
			name: "custom bump when versions are incomparable",
			args: args{
				versionInfo: VersionInfo{
					SDKVersion:         "1.0.0.1",
					PreviousSDKVersion: "1.0.0",
				},
			},
			wantVersion:  "1.0.0.1",
			wantBumpType: versioning.BumpCustom,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newVersion, bumpType, err := GetNewSDKVersion(context.Background(), tt.args.versionInfo, tt.args.force, tt.args.lang, false)
			require.NoError(t, err)
			assert.Equal(t, tt.wantVersion, newVersion)
			assert.Equal(t, tt.wantBumpType, bumpType)
		})
	}
	t.Run("returns the PreviousSDKVersion when skipVersioning is true", func(t *testing.T) {
		versionInfo := VersionInfo{
			SDKVersion:         "1.2.3",
			PreviousSDKVersion: "1.2.3",
			DocVersion:         "1.0.0",
			PreviousDocVersion: "1.0.0",
		}
		newVersion, bumpType, err := GetNewSDKVersion(context.Background(), versionInfo, false, "", true)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", newVersion)
		assert.Equal(t, versioning.BumpNone, bumpType)
	})
}

func TestDetermineBumpType(t *testing.T) {
	tests := []struct {
		name       string
		oldVersion string
		newVersion string
		expected   versioning.BumpType
	}{
		{
			name:       "Major bump",
			oldVersion: "1.0.0",
			newVersion: "2.0.0",
			expected:   versioning.BumpMajor,
		},
		{
			name:       "Minor bump",
			oldVersion: "1.0.0",
			newVersion: "1.1.0",
			expected:   versioning.BumpMinor,
		},
		{
			name:       "Patch bump",
			oldVersion: "1.0.0",
			newVersion: "1.0.1",
			expected:   versioning.BumpPatch,
		},
		{
			name:       "Pre-release bump",
			oldVersion: "1.0.0",
			newVersion: "1.0.1-alpha.1",
			expected:   versioning.BumpPrerelease,
		},
		{
			name:       "Graduate from pre-release",
			oldVersion: "1.0.0-beta.2",
			newVersion: "1.0.0",
			expected:   versioning.BumpGraduate,
		},
		{
			name:       "Custom bump - new version not greater",
			oldVersion: "1.0.0",
			newVersion: "1.0.0",
			expected:   versioning.BumpNone,
		},
		{
			name:       "Custom bump - new version lower",
			oldVersion: "1.0.0",
			newVersion: "0.9.0",
			expected:   versioning.BumpCustom,
		},
		{
			name:       "Custom bump - invalid old version",
			oldVersion: "invalid",
			newVersion: "1.0.0",
			expected:   versioning.BumpCustom,
		},
		{
			name:       "Custom bump - invalid new version",
			oldVersion: "1.0.0",
			newVersion: "invalid",
			expected:   versioning.BumpCustom,
		},
		{
			name:       "Custom bump - different segment count",
			oldVersion: "1.0.0",
			newVersion: "1.0.0.1",
			expected:   versioning.BumpCustom,
		},
		{
			name:       "Pre-release to pre-release",
			oldVersion: "1.0.0-alpha.1",
			newVersion: "1.0.0-alpha.2",
			expected:   versioning.BumpPrerelease,
		},
		{
			name:       "Major bump with pre-release",
			oldVersion: "1.0.0",
			newVersion: "2.0.0-alpha.1",
			expected:   versioning.BumpPrerelease,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineBumpType(tt.oldVersion, tt.newVersion)
			assert.Equal(t, tt.expected, result, "DetermineBumpType(%s, %s) = %v, want %v", tt.oldVersion, tt.newVersion, result, tt.expected)
		})
	}
}
