package generate

import (
	"context"
	"fmt"
	"runtime"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/targetconfig"
)

// Fetches the initial target-to-generator configuration from the config.ts file
// getGeneratorConfig function and sets it into the Generator targetConfig
// field.
func (g *Generator) getTargetInitialConfiguration(ctx context.Context) error {
	repoURL := g.repoURL

	// Fetch prior repository URL from the lock file if it exists and the URL
	// is not set during this generation. Not all generations pass the URL (e.g.
	// speakeasy run; without -r flag), so ensure any prior value is used.
	if repoURL == "" && g.lockFile != nil && g.lockFile.Management.RepoURL != "" {
		repoURL = g.lockFile.Management.RepoURL
	}

	genConfig := targetconfig.GeneratorInitialConfiguration{
		Env:      env.GetMap(),
		LangCfg:  g.subsystem.Config.Languages[g.target.Target].Cfg,
		OutDir:   g.outDir,
		Platform: runtime.GOOS,
		Publish:  g.published,
		RepoURL:  repoURL,
	}

	targetConfig, err := targetconfig.NewConfiguration(ctx, g.target, genConfig)
	if err != nil {
		return fmt.Errorf("unable to get target initial configuration: %w", err)
	}

	g.targetConfig = targetConfig

	return nil
}
