package generate

import (
	"bytes"
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/buckettypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/versioning"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/cms"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	config "github.com/speakeasy-api/sdk-gen-config"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type TestResolveASTInput struct {
	OpenAPIContents  []byte
	OpenAPIPath      string
	Target           string
	AdditionalConfig *map[string]any
}

// TestResolveAST loads a spec, sets up the generator config, and returns the resulting AST.
// This is a test helper function that can be used across different test files.
func TestResolveAST(input TestResolveASTInput) (*Generator, *ast.AST, error) {
	if input.OpenAPIPath == "" {
		input.OpenAPIPath = "test.yaml"
	}

	ctx := context.Background()

	doc, err := parseOpenAPISpec(ctx, input.OpenAPIContents)
	if err != nil {
		return nil, nil, err
	}

	// Set up document info
	docInfo := &document.DocumentInfo{
		Doc:        doc,
		Schema:     input.OpenAPIContents,
		SchemaPath: input.OpenAPIPath,
		IsRemote:   false,
	}

	target := input.Target
	var resolvedTarget types.Target
	if target == "" {
		target = "pythonv2"
		resolvedTarget = types.Target{Target: target, Template: target}
	} else {
		resolvedTarget, err = GetTargetFromTargetString(target)
		if err != nil {
			return nil, nil, err
		}
	}

	// Initialize the generator with proper configuration
	g := &Generator{
		log:       logging.NewLogger(zap.DebugLevel),
		tracer:    otel.Tracer("test"),
		subsystem: subsystem.New(nil),
		target:    resolvedTarget,
		outDir:    "/tmp",
		lockFile:  config.NewLockFile(),
		debug:     true,
	}

	// Load minimal configuration
	cfg, err := config.GetDefaultConfig(true, GetLanguageConfigDefaults, map[string]bool{target: true})
	if err != nil {
		return nil, nil, err
	}

	ctx = logging.With(ctx, g.log)

	// Initialize subsystem
	c := configuration.New(cfg, target)
	initOpts := subsystem.InitOptions{
		Target:    g.target,
		Config:    c,
		OutDir:    g.outDir,
		GenLockId: g.lockFile.ID,
		FS:        g,
		Git:       g.git,
		LockFile:  g.lockFile,
	}
	if target != "pythonv2" {
		initOpts.BaseTemplateConfigs = g.getBaseTemplateConfigs(g.target, c, g.lockFile)
	}
	err = g.subsystem.Init(ctx, initOpts)
	if err != nil {
		return nil, nil, err
	}

	sanitizer, err := g.getSanitizer(ctx, g.target)
	if err != nil {
		return nil, nil, err
	}

	g.subsystem.Sanitizer = sanitizer
	g.bucketer = buckettypes.New(g.subsystem)
	if g.splitModelThreshold > 0 {
		g.bucketer.SetSplitThresholds(g.splitModelThreshold, g.splitModelChunkSize)
	}
	if names := g.getReservedModelFileNames(); len(names) > 0 {
		g.bucketer.SetReservedModelFileNames(names)
	}

	// Initialize required components
	g.namer = namer.New(g.subsystem)
	g.ignore = filetracking.NewIgnore()
	g.comments = cms.New("sdk")
	g.warningLogger = logging.NewWarningLogger(true, func(err error) {})

	// Set required config values
	g.subsystem.Config.Generation.SDKClassName = "PetStore"
	g.subsystem.Config.Generation.DeduplicateErrors = true
	g.subsystem.Config.Generation.NameResolution = config.NameResolutionShortest
	g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025 = true

	if target == "pythonv2" {
		// Default language config for pythonv2
		g.subsystem.Config.Languages[target] = config.LanguageConfig{
			Version: "1.0.0",
			Cfg: map[string]any{
				"packageName":    "petstore",
				"packageVersion": "1.0.0",
				"imports": configuration.ImportConfig{
					Option: configuration.ImportOptionOpenAPI,
					Paths: map[string]string{
						"shared":     "shared",
						"operations": "operations",
						"errors":     "errors",
						"webhooks":   "webhooks",
						"callbacks":  "callbacks",
					},
				},
				"responseFormat":                  "flat",
				"clientServerStatusCodesAsErrors": true,
				"inputModelSuffix":                "Input",
				"outputModelSuffix":               "Output",
			},
		}
	}

	// Override config if additionalConfig is provided
	if input.AdditionalConfig != nil {
		for k, v := range *input.AdditionalConfig {
			g.subsystem.Config.Languages[target].Cfg[k] = v
		}
	}

	result, err := g.LoadAndValidateDoc(context.Background(), docInfo, &analytics.Data{}, "", false)
	if err != nil {
		return nil, nil, err
	}

	for _, err := range result.Result.GetValidationErrors() {
		fmt.Println(err)
	}

	if result.Result.HasFatalErrors() {
		return nil, nil, errors.New("fatal errors")
	}

	ast, err := g.resolveAST(context.Background(), docInfo, versioning.VersionInfo{}, &analytics.Data{})
	if err != nil {
		return nil, nil, err
	}

	return g, ast, nil
}

// parseOpenAPISpec parses an OpenAPI specification from a string into a *openapi.OpenAPI.
func parseOpenAPISpec(ctx context.Context, spec []byte) (*openapi.OpenAPI, error) {
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewBuffer(spec))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return doc, nil
}
