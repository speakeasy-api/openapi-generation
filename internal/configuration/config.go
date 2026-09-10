package configuration

import (
	"maps"

	"github.com/mitchellh/mapstructure"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type Config struct {
	config.Configuration
	target  string
	overlay map[string]any
}

func New(cfg *config.Configuration, target string) *Config {
	for t, langCfg := range cfg.Languages {
		transformImportsConfig(langCfg.Cfg)
		cfg.Languages[t] = langCfg
	}

	return &Config{
		Configuration: *cfg,
		target:        target,
	}
}

// AddOverlay adds and overlay to be applied to the language configuration
// note: this will be applied to each language configuration in the config currently
func (c *Config) AddOverlay(overlay map[string]any) {
	transformImportsConfig(overlay)
	c.overlay = overlay
}

func (c *Config) MaintainOpenAPIOrder() bool {
	return c.Generation.MaintainOpenAPIOrder
}

// Documentation output modes for `generation.documentation` in gen.yaml.
const (
	// DocumentationStandard emits per-operation `.md` docs unchanged. Default.
	DocumentationStandard = "standard"
	// DocumentationMintlify converts `docs/**/*.md` into Mintlify-flavored MDX.
	DocumentationMintlify = "mintlify"
	// DocumentationNone disables per-operation documentation generation entirely.
	DocumentationNone = "none"
)

// Documentation returns the documentation output mode for the SDK.
// Reads `generation.documentation` from gen.yaml via the AdditionalProperties
// catch-all; defaults to DocumentationStandard when unset or non-string.
// Recognized values: "standard", "mintlify", "none".
func (c *Config) Documentation() string {
	v, ok := c.Generation.AdditionalProperties["documentation"].(string)
	if !ok || v == "" {
		return DocumentationStandard
	}
	return v
}

// IsMintlify reports whether documentation should be emitted as Mintlify MDX.
func (c *Config) IsMintlify() bool {
	return c.Documentation() == DocumentationMintlify
}

// DocsDisabled reports whether per-operation documentation generation is
// turned off entirely (`generation.documentation: none`). When true, the
// generator skips writing `docs/**/*.md` files rather than emitting and then
// deleting them — a first-class opt-out.
func (c *Config) DocsDisabled() bool {
	return c.Documentation() == DocumentationNone
}

func (c *Config) GetSequencedMapIterationOrder() sequencedmap.OrderType {
	if c.MaintainOpenAPIOrder() {
		return sequencedmap.OrderAdded
	}

	return sequencedmap.OrderKeyAsc
}

func (c *Config) HoistGlobalSecurity() bool {
	if c.Generation.Auth == nil {
		return false
	}
	return c.Generation.Auth.HoistGlobalSecurity
}

// Returns an overlaid copy of the language configuration
func (c *Config) GetLanguageConfig(target string) *config.LanguageConfig {
	if _, ok := c.Languages[target]; !ok {
		return nil
	}

	langCfg := c.Languages[target]

	cfgCopy := map[string]any{}
	maps.Copy(cfgCopy, langCfg.Cfg)

	transformImportsConfig(cfgCopy)

	maps.Copy(cfgCopy, c.overlay)

	langCfg.Cfg = cfgCopy

	return &langCfg
}

func (c *Config) GetLanguageConfigValue(key string) any {
	if c.overlay != nil {
		if val, ok := c.overlay[key]; ok {
			return val
		}
	}

	if c.Languages[c.target].Cfg == nil {
		return nil
	}

	if val, ok := c.Languages[c.target].Cfg[key]; ok {
		return val
	}

	return nil
}

func transformImportsConfig(cfg map[string]any) {
	importsAny := cfg["imports"]

	if importsAny != nil {
		var importCfg ImportConfig
		if err := mapstructure.Decode(importsAny, &importCfg); err != nil {
			panic(err)
		}

		cfg["imports"] = importCfg
	}
}

// Resolve applies template-specific configuration resolution to the configuration
func (c *Config) Resolve(target types.Target, outDir string, fs filesystem.FileSystem) error {
	return templates.ResolveConfig(&c.Configuration, target, outDir, fs)
}
