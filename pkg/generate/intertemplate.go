package generate

import (
	"context"

	"github.com/dop251/goja"
	"github.com/ettle/strcase"
	"github.com/speakeasy-api/easytemplate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	config "github.com/speakeasy-api/sdk-gen-config"
)

func (g *Generator) ExecuteInterTemplateFunction(ctx context.Context, call easytemplate.CallContext, target types.Target, a *ast.AST, method string, args ...any) goja.Value {
	executor := g.getExecutor(ctx, target, g.onWriteFile(""), g.onReadFile(""))
	cfg := g.getBaseTemplateConfigs(target, g.subsystem.Config, g.lockFile)

	// We override configs like SDKName and InstallationURL to be language specific.
	langCfg, ok := g.subsystem.Config.Languages[target.Target]
	if ok {
		for k, v := range langCfg.Cfg {
			cfg[strcase.ToGoPascal(k)] = v
		}
	}

	if err := executor.Init(ctx, GlobalContext{
		Config: cfg,
		AST:    a,
	}); err != nil {
		panic(err)
	}

	if err := executor.RunScript(ctx, "standalone.ts"); err != nil {
		panic(err)
	}

	value, err := executor.RunFunction(ctx, method, args...)
	if err != nil {
		panic(err)
	}

	return value
}

func (g *Generator) ExecuteInterTemplateTarget(ctx context.Context, target string, outDir string, a *ast.AST, cfg map[string]any, enabledFeatures map[string]bool) error {
	if !g.licenseResolved {
		var err error
		if ctx, err = g.establishLicense(ctx, target, outDir); err != nil {
			return err
		}
	}
	t, err := g.GetTarget(ctx, target, g.outDir)
	if err != nil {
		return err
	}

	// Clear test-related flags when tests are disabled for the child target to prevent
	// test-related auxiliary files from being generated (e.g., Go test helpers)
	if testsEnabled, ok := enabledFeatures["tests"]; ok && !testsEnabled {
		a.MainSDK.TestGroup = ""
		a.MainSDK.OutputTests = false
	}

	if target == "go" && g.target.Target == "terraform" {
		// go does not support default arrays..
		_ = a.Walk(func(n ast.Node, parents []ast.Node, a *ast.AST) error {
			return n.Match(ast.Matchers{
				FieldDef: func(f *ast.FieldDef) error {
					if (f.Type.Type == ast.DataTypeArray || f.Type.Type == ast.DataTypeClass) && f.Default != nil {
						f.Default = nil
						if f.Optional && f.Nullable {
							f.Optional = false
						}
					}
					return nil
				},
			})
		})
	}

	return g.templateTarget(ctx, t, a, cfg, enabledFeatures, g.onWriteFile(outDir), g.onReadFile(outDir))
}

func (g *Generator) GetTargetDefaultTemplateConfig(ctx context.Context, target types.Target) (map[string]any, error) {
	cfg, err := config.GetDefaultConfig(true, func(_ string, newSDK bool) (*config.LanguageConfig, error) {
		return getLanguageConfigDefaults(target, newSDK)
	}, map[string]bool{target.Target: true})
	if err != nil {
		return nil, err
	}

	c := configuration.New(cfg, target.Target)

	overlay, err := g.getConfigOverlay()
	if err == nil {
		c.AddOverlay(overlay)
	}

	return g.getBaseTemplateConfigs(target, c, config.NewLockFile()), nil
}
