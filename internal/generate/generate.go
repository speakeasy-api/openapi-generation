package generate

import (
	"context"
	"fmt"
	"os"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.opentelemetry.io/otel/trace"
)

type GenerateOptions struct {
	Lang                   string
	SchemaPath             string
	OutDir                 string
	InstallationURL        string
	TestGroup              string
	UsageGroup             string
	Published              bool
	SkipCompile            bool
	RepoURL                string
	RepoSubDirectory       string
	Logger                 logging.Logger
	Tracer                 trace.Tracer
	Verbose                bool
	WatchTemplatesLocation string
	SkipVersioning         bool
}

// Generate generates an SDK for the given language NOTE: intended for internal workflows only
func Generate(ctx context.Context, genOpts GenerateOptions) []error {
	if _, ok := generationaccess.StateFromContext(ctx); !ok {
		ctx = generationaccess.WithDirect(ctx)
	}

	_ = os.Setenv("SPEAKEASY_DISABLE_TELEMETRY", "true")
	_ = os.Setenv("SPEAKEASY_DEBUG", "true")
	startTime := time.Now()

	logging.From(ctx).Info(fmt.Sprintf("Generating SDK for %s...", genOpts.Lang))

	schema, err := os.ReadFile(genOpts.SchemaPath)
	if err != nil {
		return []error{fmt.Errorf("failed to read schema file %s: %w", genOpts.SchemaPath, err)}
	}

	opts := []generate.GeneratorOptions{
		generate.WithLogger(genOpts.Logger),
		generate.WithTracer(genOpts.Tracer),
		generate.WithPublished(genOpts.Published),
		generate.WithInstallationURL(genOpts.InstallationURL),
		generate.WithVerboseOutput(genOpts.Verbose),
		generate.WithWatchTemplatesLocation(genOpts.WatchTemplatesLocation),
	}

	if genOpts.RepoURL != "" || genOpts.RepoSubDirectory != "" {
		opts = append(opts, generate.WithRepoDetails(genOpts.RepoURL, genOpts.RepoSubDirectory))
	}
	opts = append(opts, generate.WithSkipVersioning(genOpts.SkipVersioning))

	opts = append(opts, generate.WithRunLocation("cli"))
	opts = append(opts, generate.WithDebuggingEnabled())

	outputTests := genOpts.TestGroup != ""
	if outputTests {
		opts = append(opts, generate.WithOutputTests())
		opts = append(opts, generate.WithOutputTestGroup(genOpts.TestGroup))
	}

	if genOpts.UsageGroup != "" {
		opts = append(opts, generate.WithOutputUsageGroup(genOpts.UsageGroup))
	}

	g, err := generate.New(opts...)
	if err != nil {
		return []error{err}
	}

	if errs := g.Generate(ctx, schema, genOpts.SchemaPath, genOpts.Lang, genOpts.OutDir, false, !genOpts.SkipCompile); len(errs) > 0 {
		return errs
	}

	endTime := time.Now()

	logging.From(ctx).Info(fmt.Sprintf("Generated SDK for %s in %s", genOpts.Lang, endTime.Sub(startTime).String()))

	if os.Getenv("SPEAKEASY_DEBUG") == "true" {
		logging.From(ctx).Info("Output directory: " + genOpts.OutDir)
	}

	return nil
}
