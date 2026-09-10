package generate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	jsTemplate "github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	"github.com/speakeasy-api/openapi-generation/v2/internal/readme"
	"github.com/speakeasy-api/openapi-generation/v2/internal/template"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	formattemplateerror "github.com/speakeasy-api/openapi-generation/v2/pkg/generate/formatTemplateError"
)

type GlobalContext struct {
	Config                  map[string]any
	AST                     *ast.AST
	EnabledTemplateFeatures map[string]bool
}

type UsageDirs struct {
	AbsoluteOutputDir string // Absolute path of the testproject output directory
	RelativeOutputDir string // Path of the testproject relative to the SDK outDir
	RelativeSourceDir string // Path of the SDK outDir relative to the testproject
}

var validationErrorRegex = regexp.MustCompile(`(?s).*?ValidationError: (.*?)\n\tat .*`)

func (g *Generator) templateSDK(ctx context.Context, a *ast.AST) error {
	if g.watchTemplatesLocation != "" {
		// Execute the inner function once from the embedded FS for consistency
		err := g.templateSDKInner(ctx, a)

		fn := utils.OneManQueue(func() {
			_ = os.RemoveAll(filepath.Join(g.outDir, "__debug__"))
			err = g.templateSDKInner(ctx, a)
		})
		g.watchTemplates(fn)
		return err
	}
	return g.templateSDKInner(ctx, a)
}

func (g *Generator) watchTemplates(fn func()) {
	g.shouldReadTemplatesFromDisk = true
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	templatesLocation := g.watchTemplatesLocation
	fullLocation := filepath.Join(templatesLocation, "templates")
	targetLocation := filepath.Join(templatesLocation, "templates", g.target.Template)
	commonLocation := filepath.Join(templatesLocation, "templates", "common")
	if s, err := os.Stat(targetLocation); err != nil || !s.IsDir() {
		panic("Unable to watch templates, this location is not a directory: " + targetLocation)
	}
	if s, err := os.Stat(commonLocation); err != nil || !s.IsDir() {
		panic("Unable to watch templates, this location is not a directory: " + commonLocation)
	}

	wg := sync.WaitGroup{}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !event.Op.Has(fsnotify.Chmod) {
					g.log.Debug(fmt.Sprintf("watchTemplates event: %s %s", event.Op.String(), strings.ReplaceAll(event.Name, fullLocation, "")))
					wg.Add(1)
					go func() {
						defer wg.Done()
						fn()
						g.log.Info("Press 'c' to stop watching templates and complete generation")
					}()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				g.log.Error("watchTemplates error", zap.Error(err))
			}
		}
	}()

	dirs := utils.RecursivelyListDirectories(targetLocation, commonLocation)
	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			g.log.Error("watchTemplates add path error", zap.Error(err))
		}
	}

	g.log.Debug("Now watching templates under '" + fullLocation + "' (" + strconv.Itoa(len(dirs)) + " sub directories)")
	g.log.Info("Press 'c' to stop watching templates and complete generation")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) == "c" {
			break
		}
	}
	g.log.Debug("Stopping watchTemplates, will now complete generation")
	g.shouldReadTemplatesFromDisk = false
	wg.Wait()
}

func (g *Generator) templateSDKInner(ctx context.Context, a *ast.AST) error {
	ctx, span := g.tracer.Start(ctx, "Generator.templateSDK")
	defer span.End()

	g.log.Debug("templateSDK started")

	start := time.Now()

	// Base Configs
	cfg := g.getBaseTemplateConfigs(g.target, g.subsystem.Config, g.lockFile)

	// Dev Container Configs
	if g.subsystem.Config.Generation.DevContainers != nil && g.subsystem.Config.Generation.DevContainers.Enabled {
		cfg["DevContainerSchemaPath"] = g.subsystem.Config.Generation.DevContainers.SchemaPath
	}

	// Standalone Readme
	if g.standaloneReadmeConfig != nil {
		cfg["Readme"] = map[string]any{
			"Standalone":  true,
			"FileName":    g.standaloneReadmeConfig.FileName,
			"HeadersOnly": g.standaloneReadmeConfig.HeadersOnly,
		}

		return g.templateStandalone(ctx, a, cfg)
	}

	// Standalone Usage / TestProject
	if g.generateUsageSnippetArgs != nil {
		usageContexts, err := readme.OperationsForUsageSnippets(a.MainSDK,
			g.generateUsageSnippetArgs.OperationIDs,
			g.generateUsageSnippetArgs.Namespace,
			g.generateUsageSnippetArgs.RootExample,
			g.generateUsageSnippetArgs.All,
			g.generateUsageSnippetArgs.ExampleValues.RequestBodyJSON,
			g.generateUsageSnippetArgs.ExampleValues.Params,
		)
		if err != nil {
			return err
		}

		usageConfig := map[string]any{
			"UsageContexts": usageContexts,
			"Standalone":    g.generateUsageSnippetArgs.Standalone,
		}

		// Standalone Usage Generation
		if g.generateUsageSnippetArgs.Standalone {
			cfg["Usage"] = usageConfig
			return g.templateStandalone(ctx, a, cfg)
		}

		// Snippet Compilation (testproject)
		testProjectDir := fmt.Sprintf("testprojects/%s/%s", g.target.Template, filepath.Base(g.outDir))
		dirs, err := g.getUsageDirs(testProjectDir)
		if err != nil {
			return err
		}
		g.generateUsageSnippetArgs.OutputDir = dirs.AbsoluteOutputDir
		usageConfig["OutputDir"] = dirs.RelativeOutputDir
		usageConfig["SourceDir"] = dirs.RelativeSourceDir
		cfg["TestProject"] = usageConfig
		cfg["TestGroup"] = g.testGroup
	}

	// TestUsage
	if g.testGroup != "" {
		testUsageDir := fmt.Sprintf("testusages/%s/%s", g.target.Template, g.testGroup)
		dirs, err := g.getUsageDirs(testUsageDir)
		if err != nil {
			return err
		}
		cfg["Usage"] = map[string]any{
			"Standalone": false,
			"OutputDir":  dirs.RelativeOutputDir,
			"SourceDir":  dirs.RelativeSourceDir,
		}
	}

	err := g.templateTarget(ctx, g.target, a, cfg, nil, g.onWriteFile(""), g.onReadFile(""))
	if err != nil {
		if g.debug {
			s := err.Error()
			g.log.Debug("templateSDK error " + formattemplateerror.FormatTemplateErrorString(s))

		}
	}

	// Pre-render standalone usage snippets as a side output when opted in,
	// but only if SDK generation succeeded and we're not already in snippet mode.
	if err == nil && g.renderUsageSnippets && g.generateUsageSnippetArgs == nil {
		if snippetErr := g.renderStandaloneSnippets(ctx, a, cfg); snippetErr != nil {
			g.log.Warn("failed to pre-render usage snippets: " + snippetErr.Error())
		}
	}

	g.log.Debug(fmt.Sprintf("templateSDK completed in '%s'", time.Since(start)))
	return err
}

func (g *Generator) templateStandalone(ctx context.Context, a *ast.AST, cfg map[string]any) error {
	e := g.getExecutor(ctx, g.target, g.onWriteFile(""), g.onReadFile(""))
	defer func() { _ = e.Close() }()

	if err := e.Init(ctx, GlobalContext{
		Config: cfg,
		AST:    a,
	}); err != nil {
		return err
	}

	if err := e.RunScript(ctx, "standalone.ts"); err != nil {
		return parseScriptError(err)
	}

	return nil
}

// renderStandaloneSnippets reuses the already-resolved AST to render standalone
// usage snippets into a buffer, avoiding a separate generation pass.
func (g *Generator) renderStandaloneSnippets(ctx context.Context, a *ast.AST, cfg map[string]any) error {
	usageContexts, err := readme.OperationsForUsageSnippets(a.MainSDK, nil, "", false, true, "", nil)
	if err != nil {
		return err
	}

	// Build a standalone config overlaying the base template config
	snippetCfg := make(map[string]any, len(cfg)+1)
	for k, v := range cfg {
		snippetCfg[k] = v
	}
	snippetCfg["Usage"] = map[string]any{
		"UsageContexts": usageContexts,
		"Standalone":    true,
	}

	// Buffer-capturing write function: only collect files containing usage snippets
	var buf bytes.Buffer
	captureWrite := func(_ context.Context, _ string, data []byte, _ fs.FileMode, _ bool) error {
		if strings.Contains(string(data), "Usage snippet provided") {
			buf.Write(data)
		}
		return nil
	}

	e := g.getExecutor(ctx, g.target, captureWrite, g.onReadFile(""))
	defer func() { _ = e.Close() }()
	if err := e.Init(ctx, GlobalContext{
		Config: snippetCfg,
		AST:    a,
	}); err != nil {
		return err
	}

	if err := e.RunScript(ctx, "standalone.ts"); err != nil {
		return parseScriptError(err)
	}

	g.renderedUsageSnippets = &RenderedUsageSnippets{
		RawOutput: buf.String(),
	}
	return nil
}

func (g *Generator) getUsageDirs(usageOutputDir string) (*UsageDirs, error) {
	absOutDir, err := filepath.Abs(usageOutputDir)
	if err != nil {
		return nil, err
	}

	genOutDir, err := filepath.Abs(g.outDir)
	if err != nil {
		return nil, err
	}

	relSrcDir, err := filepath.Rel(absOutDir, genOutDir)
	if err != nil {
		return nil, err
	}

	relOutDir, err := filepath.Rel(genOutDir, absOutDir)
	if err != nil {
		return nil, err
	}

	return &UsageDirs{
		AbsoluteOutputDir: absOutDir,
		RelativeOutputDir: relOutDir,
		RelativeSourceDir: relSrcDir,
	}, nil
}

func (g *Generator) templateTarget(ctx context.Context, target types.Target, a *ast.AST, cfg map[string]any, enabledFeatures map[string]bool, writeFileFunc jsTemplate.WriteFileFunc, readFileFunc jsTemplate.ReadFileFunc) error {
	ctx, span := g.tracer.Start(ctx, "Generator.templateTarget", trace.WithAttributes(attribute.String("target", target.Target)))
	defer span.End()

	tmplCfg := g.getTemplateConfig(ctx, target, writeFileFunc, readFileFunc)
	tmplCfg.DebugGojaPort = env.DebugGojaPort()

	t := template.New(ctx, target, template.Config{
		Config:          tmplCfg,
		TemplateConfig:  cfg,
		EnabledFeatures: enabledFeatures,
		AST:             a,
	})

	return t.Execute(ctx)
}

func parseScriptError(err error) error {
	// A bit crap we need to parse the error message but its what we have for now
	if strings.Contains(err.Error(), "ValidationError: ") {
		matches := validationErrorRegex.FindAllStringSubmatch(err.Error(), -1)
		if len(matches) > 0 {
			return errors.NewValidationError(matches[0][1], nil, nil)
		}
	}

	return errors.ErrGeneration.Wrap(err)
}
