package generate

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"

	"go.uber.org/zap"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	config "github.com/speakeasy-api/sdk-gen-config"
)

// Replacement for generate.GetCommonConfigFields
func GetGenerationConfigFields(newSDK bool) []config.SDKGenConfigField {
	return config.GetGenerationDefaults(newSDK)
}

func GetLanguageConfigDefaults(target string, newSDK bool) (*config.LanguageConfig, error) {
	// This is a special case and if multiple versions of a language are supported the default templateVersion needs to live in the base template ie `typescript` instead of `typescriptv2`
	t, err := templates.GetTargetFromTargetString(target)
	if err != nil {
		return nil, err
	}

	return getLanguageConfigDefaults(t, newSDK)
}

// LoadConfig loads the configuration, usually from a gen.yaml file, and
// returns it.
func (g *Generator) LoadConfig(ctx context.Context, outDir string, targets ...string) (*config.Config, error) {
	cfg, err := config.Load(outDir, g.getConfigOptions(ctx, true, targets...)...)
	if err != nil {
		return nil, err
	}

	if g.debug {
		for k, v := range cfg.Config.New {
			logging.From(ctx).Debug(fmt.Sprintf("%s target detected as newSDK: %v", k, v))
		}
	}

	if nr := cfg.Config.Generation.NameResolution; nr != "" && !nr.IsValid() {
		return nil, fmt.Errorf("invalid generation.nameResolution %q: must be one of %v", nr, config.NameResolutionModes())
	}

	g.mergeAbsentObjectDefaults(cfg)

	logging.From(ctx).Info("Found config file", zap.String("path", cfg.ConfigPath))
	return cfg, nil
}

// mergeAbsentObjectDefaults fills in missing keys from object-typed defaults
// (e.g. fixFlags) into the loaded config. The sdk-gen-config Load() only does
// a shallow top-level key check, so if a user has fixFlags: {key1: true} and a
// new key2 is added to defaults, key2 would otherwise be missing. Defaults are
// constructed with newSDK=false, so this only adds falsy default values for
// keys the user hasn't explicitly set.
func (g *Generator) mergeAbsentObjectDefaults(cfg *config.Config) {
	for lang, langCfg := range cfg.Config.Languages {
		newSDK := cfg.Config.New[lang]
		defaults, err := g.getLanguageConfigDefaults(lang, newSDK)
		if err != nil {
			continue
		}

		for k, defaultVal := range defaults.Cfg {
			defaultMap, ok := defaultVal.(map[string]any)
			if !ok {
				continue
			}
			resultMap, ok := langCfg.Cfg[k].(map[string]any)
			if !ok {
				continue
			}
			for subKey, subVal := range defaultMap {
				if _, exists := resultMap[subKey]; !exists {
					resultMap[subKey] = subVal
				}
			}
		}
	}
}

func getLanguageConfigDefaults(target types.Target, newSDK bool) (*config.LanguageConfig, error) {
	configFields, err := templates.GetLanguageConfigFields(target, newSDK)
	if err != nil {
		return nil, err
	}

	var versionDefault string
	for _, field := range configFields {
		if field.Name == "version" && field.DefaultValue != nil {
			versionDefault = (*field.DefaultValue).(string)
		}
	}

	langCfg := &config.LanguageConfig{
		Version: versionDefault,
		Cfg:     map[string]any{},
	}

	for _, field := range configFields {
		if field.Name == "version" {
			continue
		}

		if field.DefaultValue != nil {
			langCfg.Cfg[field.Name] = *field.DefaultValue
		}
	}

	return langCfg, nil
}

func (g *Generator) getLanguageConfigDefaults(target string, newSDK bool) (*config.LanguageConfig, error) {
	t := g.target

	return getLanguageConfigDefaults(t, newSDK)
}

func (g *Generator) upgradeConfig(target, template, oldVersion, newVersion string, cfg map[string]any) (map[string]any, error) {
	// Not a language section in the config skip
	if !slices.Contains(templates.AllSupportedTemplateNames(), target) {
		return cfg, nil
	}

	t := types.Target{Target: target, Template: template}

	e, err := executor.New(t, "config.ts")
	if err != nil {
		return nil, err
	}

	defaults, err := getLanguageConfigDefaults(t, false)
	if err != nil {
		return nil, err
	}

	defaultsFlat := map[string]any{}
	defaultsFlat["version"] = defaults.Version
	delete(defaults.Cfg, "version")
	for k, v := range defaults.Cfg {
		defaultsFlat[k] = v
	}

	cfgVal, err := e.Run("upgradeConfig", oldVersion, newVersion, cfg, defaultsFlat)
	if err != nil {
		return nil, err
	}

	return cfgVal.(map[string]any), nil
}

func (g *Generator) getConfigOptions(ctx context.Context, upgrade bool, targets ...string) []config.Option {
	opts := []config.Option{
		config.WithFileSystem(g),
		config.WithLanguageDefaultFunc(g.getLanguageConfigDefaults),
		config.WithLanguages(targets...),
		config.WithTransformerFunc(func(c *config.Config) (*config.Config, error) {
			return g.transformConfig(ctx, c)
		}),
		config.WithValidateFunc(func(c config.Config) error {
			return g.validateConfig(ctx, c, targets...)
		}),
	}

	if g.dontWrite {
		opts = append(opts, config.WithDontWrite())
	}

	if upgrade {
		opts = append(opts, config.WithUpgradeFunc(g.upgradeConfig))
	}

	return opts
}

func (g *Generator) transformConfig(ctx context.Context, cfg *config.Config) (*config.Config, error) {
	// Track whether repoURL was explicitly configured (via CLI flag or gen.yaml)
	// so we can derive installationURL from it below.
	repoURLExplicitlySet := false

	// gen.yaml repoURL takes precedence over CLI-passed values (which in CI
	// are derived from GITHUB_REPOSITORY and may point to a source monorepo
	// rather than the published SDK repo).
	if repoURL, ok := cfg.Config.Generation.AdditionalProperties["repoUrl"].(string); ok && repoURL != "" {
		cfg.LockFile.Management.RepoURL = repoURL
		repoURLExplicitlySet = true
	} else if repoURL, ok := cfg.Config.Generation.AdditionalProperties["repoURL"].(string); ok && repoURL != "" {
		cfg.LockFile.Management.RepoURL = repoURL
		repoURLExplicitlySet = true
	} else if g.repoURL != "" {
		cfg.LockFile.Management.RepoURL = g.repoURL
		repoURLExplicitlySet = true
	}

	// TODO this currently doesn't support setting back to false if the sdk-generation-action stops publishing it but that is a minor issue
	if g.published {
		cfg.LockFile.Management.Published = g.published
	}

	// gen.yaml repoSubDirectory takes precedence over CLI-passed values.
	// An explicit empty string in gen.yaml means "no subdirectory" (the
	// published repo has the SDK at root), which differs from the CI-computed
	// value that reflects the source monorepo layout.
	if _, ok := cfg.Config.Generation.AdditionalProperties["repoSubDirectory"]; ok {
		subDir, _ := cfg.Config.Generation.AdditionalProperties["repoSubDirectory"].(string)
		cfg.LockFile.Management.RepoSubDirectory = subDir
	} else if g.repoSubDirectory != "" {
		cfg.LockFile.Management.RepoSubDirectory = g.repoSubDirectory
	}
	if g.installationURL != "" {
		cfg.LockFile.Management.InstallationURL = g.installationURL
	}

	// When repoURL is explicitly configured, derive installationURL from it.
	// This ensures installation commands in the README point to the correct
	// published repository rather than a stale value from gen.lock (which may
	// point to a different source/monorepo where generation runs).
	if repoURLExplicitlySet {
		// Use the gen.yaml repoSubDirectory if present (even if empty), since it
		// describes the subdirectory in the *published* repo. The CLI-computed
		// repoSubDirectory comes from the workflow output path, which reflects the
		// source monorepo layout and may be wrong for the published repo.
		subDir := cfg.LockFile.Management.RepoSubDirectory
		if _, ok := cfg.Config.Generation.AdditionalProperties["repoSubDirectory"]; ok {
			subDir, _ = cfg.Config.Generation.AdditionalProperties["repoSubDirectory"].(string)
		}
		cfg.LockFile.Management.InstallationURL = deriveInstallationURL(
			g.target.Target,
			cfg.LockFile.Management.RepoURL,
			subDir,
		)
	}

	if !licensing.AccountHasFeatureAccess(ctx, features.FeatureOAuth2ClientCredentials) {
		cfg.Config.Generation.Auth.OAuth2ClientCredentialsEnabled = false
	}

	if !licensing.AccountHasFeatureAccess(ctx, features.FeatureOAuth2Password) {
		cfg.Config.Generation.Auth.OAuth2PasswordEnabled = false
	}

	if langCfg, ok := cfg.Config.Languages[g.target.Target]; ok {
		if _, ok := langCfg.Cfg["templateVersion"]; ok {
			templateVersion := strings.TrimPrefix(g.target.Template, g.target.Target)
			if templateVersion != "" {
				langCfg.Cfg["templateVersion"] = templateVersion
				cfg.Config.Languages[g.target.Target] = langCfg
			}
		}

		// Disable MCP server when using Zod v4 for TypeScript targets
		if g.target.Target == "typescript" {
			if zodVersion, ok := langCfg.Cfg["zodVersion"]; ok && (zodVersion == "v4" || zodVersion == "v4-mini") {
				if enableMCP, ok := langCfg.Cfg["enableMCPServer"]; ok && enableMCP == true {
					langCfg.Cfg["enableMCPServer"] = false
					cfg.Config.Languages[g.target.Target] = langCfg
					logging.From(ctx).Warn("MCP server is disabled when using Zod v4")
				}
			}

			// zodVersion: "none" strips zod from the generated SDK. MCP and React
			// Query are whole subsystems that can't compile without zod schemas,
			// so disable them up-front. Other settings such as laxMode and
			// unionStrategy describe behaviour of zod transforms — when there
			// is no zod, the templating layer simply ignores them (they remain
			// in the user's gen.yaml as a no-op).
			if zodVersion, ok := langCfg.Cfg["zodVersion"]; ok && zodVersion == "none" {
				mutated := false
				if enableMCP, ok := langCfg.Cfg["enableMCPServer"]; ok && enableMCP == true {
					langCfg.Cfg["enableMCPServer"] = false
					mutated = true
					logging.From(ctx).Warn("MCP server is disabled when zodVersion is \"none\"")
				}
				// Force preserveModelFieldNames on: no-zod has no $inboundSchema $outboundSchema transforms to translate between TS
				// identifiers and wire keys, so the only safe shape is one where they match. Without this, auto camelCase sanitization
				// and x-speakeasy-name-override would silently ship the wrong JSON keys.
				if preserve, ok := langCfg.Cfg["preserveModelFieldNames"]; !ok || preserve != true {
					langCfg.Cfg["preserveModelFieldNames"] = true
					mutated = true
					logging.From(ctx).Warn(
						"preserveModelFieldNames forced to true when zodVersion is \"none\"",
						zap.String("detail", "TS model field identifiers now match wire keys verbatim (no auto camelCase, x-speakeasy-name-override ignored)"),
					)
				}
				// React Query layer wraps the func-level API and is independent of
				// zod schemas — the funcs already handle no-zod mode internally.
				// Kept enabled when the customer opts in.
				if mutated {
					cfg.Config.Languages[g.target.Target] = langCfg
				}
			}

			// Force constFieldsAlwaysOptional to false when preApplyUnionDiscriminators is enabled
			// This is because discriminators will be applied as `consts` and we need to respect the required-ness of `const` fields to use them as a discriminator
			if applyDiscriminators, ok := langCfg.Cfg["preApplyUnionDiscriminators"]; ok && applyDiscriminators == true {
				if constFieldsOptional, ok := langCfg.Cfg["constFieldsAlwaysOptional"]; ok && constFieldsOptional == true {
					langCfg.Cfg["constFieldsAlwaysOptional"] = false
					cfg.Config.Languages[g.target.Target] = langCfg
					logging.From(ctx).Warn("constFieldsAlwaysOptional is set to false when preApplyUnionDiscriminators is enabled")
				}
			}
		}

		// Check for deprecated dxtManifestOverlay field (for mcp-typescript target)
		if g.target.Target == "mcp-typescript" {
			if _, hasDxtOverlay := langCfg.Cfg["dxtManifestOverlay"]; hasDxtOverlay {
				logging.From(ctx).Warn(
					"Deprecated configuration field",
					zap.String("field", "dxtManifestOverlay"),
					zap.String("message", "'dxtManifestOverlay' is deprecated. Use 'mcpbManifestOverlay' instead."),
					zap.String("target", g.target.Target),
				)
				// rename the field to mcpbManifestOverlay; do not overwrite
				if _, hasMcpbOverlay := langCfg.Cfg["mcpbManifestOverlay"]; !hasMcpbOverlay {
					langCfg.Cfg["mcpbManifestOverlay"] = langCfg.Cfg["dxtManifestOverlay"]
					delete(langCfg.Cfg, "dxtManifestOverlay")
				}
			}
		}

		// Enable unionStrategy: populated-fields when laxMode is enabled
		if g.target.Target == "typescript" {
			if laxMode, ok := langCfg.Cfg["laxMode"]; ok && laxMode == "lax" {
				langCfg.Cfg["unionStrategy"] = "populated-fields"
				cfg.Config.Languages[g.target.Target] = langCfg
			}
		}

		// Migrate unionDeserializationStrategy to unionStrategy for Go and Terraform
		if g.target.Target == "go" || g.target.Target == "terraform" {
			if oldStrategy, ok := langCfg.Cfg["unionDeserializationStrategy"]; ok {
				// Only migrate if unionStrategy is not already set
				if _, hasNew := langCfg.Cfg["unionStrategy"]; !hasNew {
					langCfg.Cfg["unionStrategy"] = oldStrategy
				}
				delete(langCfg.Cfg, "unionDeserializationStrategy")
				cfg.Config.Languages[g.target.Target] = langCfg
			}
		}

		// Migrate enhancedUnionMemberResolution to unionStrategy for Java
		if g.target.Target == "java" {
			if oldResolution, ok := langCfg.Cfg["enhancedUnionMemberResolution"]; ok {
				// Only migrate if unionStrategy is not already set
				if _, hasNew := langCfg.Cfg["unionStrategy"]; !hasNew {
					// Convert boolean to strategy string
					if enhanced, isBool := oldResolution.(bool); isBool {
						if enhanced {
							langCfg.Cfg["unionStrategy"] = "populated-fields"
						} else {
							langCfg.Cfg["unionStrategy"] = "left-to-right"
						}
					}
				}
				delete(langCfg.Cfg, "enhancedUnionMemberResolution")
				cfg.Config.Languages[g.target.Target] = langCfg
			}
		}

		// Migrate openUnions to forwardCompatibleUnionsByDefault for Java
		if g.target.Target == "java" {
			if oldUnions, ok := langCfg.Cfg["openUnions"]; ok {
				// Only migrate if forwardCompatibleUnionsByDefault is not already set
				if _, hasNew := langCfg.Cfg["forwardCompatibleUnionsByDefault"]; !hasNew {
					langCfg.Cfg["forwardCompatibleUnionsByDefault"] = oldUnions
				}
				delete(langCfg.Cfg, "openUnions")
				cfg.Config.Languages[g.target.Target] = langCfg
			}
		}
	}

	return cfg, nil
}

func (g *Generator) validateConfig(ctx context.Context, cfg config.Config, targets ...string) error {
	for _, target := range targets {
		langCfg, ok := cfg.Config.Languages[target]
		if !ok {
			continue
		}

		if langCfg.Cfg == nil {
			continue
		}

		t := types.Target{Target: target, Template: target}

		templateVersion, ok := langCfg.Cfg["templateVersion"]
		if ok {
			t.Template = fmt.Sprintf("%s%s", target, templateVersion)
		}

		e, err := executor.New(t, "config.ts")
		if err != nil {
			logging.From(ctx).Debug("Error creating executor", zap.Error(err))
			continue
		}

		validationMessage, err := e.Run("validateConfig", langCfg.Cfg)
		if err != nil {
			logging.From(ctx).Debug("Error validating config", zap.Error(err))
			continue
		}

		if validationMessage != nil && (validationMessage.(string) != "") {
			return errors.NewValidationError("gen.yaml validation error for "+target, nil, errors.New(validationMessage.(string)))
		}
	}

	return nil
}

func (g *Generator) getConfigUsage() map[string]any {
	usage := map[string]any{}

	usage = mergeMaps(usage, GetConfigUsageFromType(g.subsystem.Config.Generation, ""))

	langConfig, ok := g.subsystem.Config.Languages[g.target.Target]
	if ok {
		usage = mergeMaps(usage, GetConfigUsageFromType(langConfig.Cfg, g.target.Target))
	}

	return usage
}

func GetConfigUsageFromType(s any, parent string) map[string]any {
	usage := map[string]any{}

	if s == nil {
		return usage
	}

	v := reflect.ValueOf(s)

	if v.IsZero() {
		return usage
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return usage
		}

		v = v.Elem()
	}

	t := v.Type()

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields - they can't be accessed via reflection
			if !field.IsExported() {
				continue
			}

			val := v.Field(i)

			keyParts := []string{}
			if parent != "" {
				keyParts = append(keyParts, parent)
			}
			keyParts = append(keyParts, casing.New().ToSnake(field.Name))

			key := strings.Join(keyParts, "_")

			usage = mergeMaps(usage, GetConfigUsageFromType(val.Interface(), key))
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			keyName := casing.New().ToSnake(fmt.Sprintf("%v", k.Interface()))
			keyParts := []string{}
			if parent != "" {
				keyParts = append(keyParts, parent)
			}
			keyParts = append(keyParts, keyName)

			key := strings.Join(keyParts, "_")

			usage = mergeMaps(usage, GetConfigUsageFromType(v.MapIndex(k).Interface(), key))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			keyParts := []string{}
			if parent != "" {
				keyParts = append(keyParts, parent)
			}
			keyParts = append(keyParts, strconv.Itoa(i))

			key := strings.Join(keyParts, "_")

			usage = mergeMaps(usage, GetConfigUsageFromType(v.Index(i).Interface(), key))
		}
	default:
		if parent == "" {
			panic("parent is empty")
		}

		usage[parent] = s
	}

	return usage
}

func mergeMaps(m1, m2 map[string]any) map[string]any {
	for k, v := range m2 {
		m1[k] = v
	}

	return m1
}

func (g *Generator) getConfigOverlay() (map[string]any, error) {
	e, err := executor.New(g.target, "config.ts")
	if err != nil {
		return nil, err
	}

	defVal, err := e.Run("getConfigOverlay")
	if err != nil {
		return nil, err
	}

	return defVal.(map[string]any), nil
}

// deriveInstallationURL constructs a language-appropriate installation URL from
// a repository URL and optional subdirectory. Each language ecosystem has its
// own convention for installing packages from git repositories.
func deriveInstallationURL(lang, repoURL, subDirectory string) string {
	subDirectory = strings.TrimSpace(subDirectory)
	subDirectory = strings.TrimSuffix(subDirectory, "/")
	if subDirectory == "." {
		subDirectory = ""
	}

	switch lang {
	case "python":
		// Python uses: git+URL.git#subdirectory=DIR
		// The installation template already adds "git+" prefix and ".git" suffix,
		// so we just need to handle the subdirectory fragment.
		if !strings.HasSuffix(repoURL, ".git") {
			repoURL += ".git"
		}
		if subDirectory == "" {
			return repoURL
		}
		return repoURL + "#subdirectory=" + subDirectory

	case "typescript":
		// TypeScript installation URL is the bare URL. Subdirectory handling
		// is done per-package-manager in the installation template, since each
		// package manager uses a different format (npm: ?subdir=, pnpm: #path:, etc).
		return strings.TrimSuffix(repoURL, ".git")

	case "ruby":
		// Ruby uses: URL -d DIR
		base := strings.TrimSuffix(repoURL, ".git")
		if subDirectory == "" {
			return base
		}
		return base + " -d " + subDirectory

	case "php":
		// PHP uses the URL in composer.json repository config.
		// Note: Composer VCS requires composer.json at the repo root,
		// so subdirectory is not included in the URL (it won't work).
		if !strings.HasSuffix(repoURL, ".git") {
			repoURL += ".git"
		}
		return repoURL

	default:
		return strings.TrimSuffix(repoURL, ".git")
	}
}

// getReservedModelFileNames reads the reservedModelFileNames config value
// (set via getConfigOverlay in language templates) and returns it as []string.
func (g *Generator) getReservedModelFileNames() []string {
	val := g.subsystem.Config.GetLanguageConfigValue("reservedModelFileNames")
	if val == nil {
		return nil
	}
	items, ok := val.([]interface{})
	if !ok {
		return nil
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			names = append(names, s)
		}
	}
	return names
}
