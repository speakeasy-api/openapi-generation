package generate

import (
	"context"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
)

func (g *Generator) getExecutor(ctx context.Context, target types.Target, writeFileFunc template.WriteFileFunc, readFileFunc template.ReadFileFunc) *template.Engine {
	return template.New(ctx, target, g.getTemplateConfig(ctx, target, writeFileFunc, readFileFunc))
}

func (g *Generator) getTemplateConfig(_ context.Context, target types.Target, writeFileFunc template.WriteFileFunc, readFileFunc template.ReadFileFunc) template.Config {
	templateDir := "templates/" + target.Template

	additionalSearchLocations := []string{}

	// TODO we need to pull this from the template
	if target.Target == "cli" {
		additionalSearchLocations = append(additionalSearchLocations, "templates/go")
	}

	watchTemplatesLocation := ""
	if g.shouldReadTemplatesFromDisk {
		watchTemplatesLocation = g.watchTemplatesLocation
		g.log.Debug("Watching templates")
	}

	var importsConfig configuration.ImportConfig

	if ic, ok := g.subsystem.Config.GetLanguageConfigValue("imports").(configuration.ImportConfig); ok {
		importsConfig = ic
	}

	return template.Config{
		TemplateDir:               templateDir,
		AdditionalSearchLocations: additionalSearchLocations,
		Subsystem:                 g.subsystem,
		Bucketer:                  g.bucketer,
		Debug:                     g.debug,
		WriteFileFunc:             writeFileFunc,
		ReadFileFunc:              readFileFunc,
		FileTracker:               g.subsystem.FileTracker,
		Ignore:                    g.ignore,
		Comments:                  g.comments,
		Targeter:                  g,
		Interopter:                g,
		OutDir:                    g.outDir, // TODO this may be an issue
		ImportConfig:              importsConfig,
		WatchTemplatesLocation:    watchTemplatesLocation,
	}
}

func (g *Generator) getBaseTemplateConfigs(target types.Target, cfg *configuration.Config, lockFile *config.LockFile) map[string]any {
	templateCfg := map[string]any{}

	langCfg := cfg.GetLanguageConfig(target.Target)
	if langCfg != nil {
		templateCfg["SDKVersion"] = langCfg.Version
		for k, v := range langCfg.Cfg {
			templateCfg[strcase.ToGoPascal(k)] = v
		}
	}

	templateCfg["Language"] = target.Target
	templateCfg["SDKName"] = cfg.Generation.SDKClassName
	templateCfg["DocVersion"] = g.docVersion
	templateCfg["GenVersion"] = g.genVersion
	templateCfg["CLIVersion"] = g.cliVersion
	templateCfg["Fixes"] = cfg.Generation.Fixes
	templateCfg["WorkspaceUri"] = g.workspaceUri
	templateCfg["Published"] = lockFile.Management.Published
	templateCfg["InstallationURL"] = lockFile.Management.InstallationURL
	templateCfg["RepoSubDirectory"] = lockFile.Management.RepoSubDirectory
	templateCfg["RepoURL"] = lockFile.Management.RepoURL

	if moduleName, ok := templateCfg["ModuleName"]; !ok || moduleName == "" {
		templateCfg["ModuleName"] = templateCfg["PackageName"]
	}
	optionalPropertyRendering := config.OptionalPropertyRenderingOptionWithExample
	if cfg.Generation.UsageSnippets != nil {
		optionalPropertyRendering = cfg.Generation.UsageSnippets.OptionalPropertyRendering
	}
	templateCfg["OptionalPropertyRendering"] = string(optionalPropertyRendering)
	sdkInitStyle := config.SDKInitStyleConstructor
	if cfg.Generation.UsageSnippets != nil {
		sdkInitStyle = cfg.Generation.UsageSnippets.SDKInitStyle
	}
	if cfg.Generation.UsageSnippets != nil {
		if cfg.Generation.UsageSnippets.ServerToShowInSnippets != "" {
			templateCfg["ServerToShowInSnippets"] = cfg.Generation.UsageSnippets.ServerToShowInSnippets
		}
	}
	templateCfg["SDKInitStyle"] = string(sdkInitStyle)
	templateCfg["MaintainOpenAPIOrder"] = cfg.MaintainOpenAPIOrder()
	templateCfg["SDKHooksConfigAccess"] = cfg.Generation.SDKHooksConfigAccess
	templateCfg["Documentation"] = cfg.Documentation()
	templateCfg["GeneratedLicense"] = string(g.generatedLicense)

	return templateCfg
}
