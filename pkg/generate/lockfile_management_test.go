package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyManagementInfoVersioningStrategy(t *testing.T) {
	for _, tt := range []struct {
		name          string
		strategy      config.VersioningStrategy
		wantChecksums bool
	}{
		{name: "automatic keeps checksums", strategy: config.VersioningStrategyAutomatic, wantChecksums: true},
		{name: "manual omits checksums", strategy: config.VersioningStrategyManual, wantChecksums: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.GetDefaultConfig(true, GetLanguageConfigDefaults, map[string]bool{"go": true})
			require.NoError(t, err)
			c := configuration.New(cfg, "go")
			c.Generation.VersioningStrategy = tt.strategy

			g := &Generator{
				subsystem:  subsystem.New(nil),
				lockFile:   config.NewLockFile(),
				docVersion: "1.1.0",
				genVersion: "2.500.0",
				cliVersion: "1.400.0",
			}
			g.subsystem.Config = c
			// Simulate a lock file carrying checksums from a previous run to
			// verify manual mode clears them rather than preserving stale values.
			g.lockFile.Management.DocChecksum = "stale-doc"
			g.lockFile.Management.ConfigChecksum = "stale-config"

			g.applyManagementInfo("config-sum", "doc-sum", "1.2.3")

			if tt.wantChecksums {
				assert.Equal(t, "doc-sum", g.lockFile.Management.DocChecksum)
				assert.Equal(t, "config-sum", g.lockFile.Management.ConfigChecksum)
			} else {
				assert.Empty(t, g.lockFile.Management.DocChecksum)
				assert.Empty(t, g.lockFile.Management.ConfigChecksum)
			}
			assert.Equal(t, "1.2.3", g.lockFile.Management.ReleaseVersion)
			assert.Equal(t, "1.1.0", g.lockFile.Management.DocVersion)
			assert.Equal(t, "2.500.0", g.lockFile.Management.GenerationVersion)
			assert.Equal(t, "1.400.0", g.lockFile.Management.SpeakeasyVersion)
		})
	}
}
