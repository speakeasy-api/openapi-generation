// Package runner provides the shared generation pipeline used by cmd/generate and cmd/regen.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/dop251/goja"
	generationaccess "github.com/speakeasy-api/generation-context/access"
	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/instrument"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RunConfig holds everything the shared pipeline needs. Each cmd populates this
// after its own flag-parsing / inference phase.
type RunConfig struct {
	SpanName                 string                          // OTel span name ("main" for generate, "regen" for regen)
	GenOpts                  generate.GenerateOptions        // passed to generate.Generate()
	TraceDest                string                          // merged OTel destination (caller merges flag + env)
	ProfileDir               string                          // profiling output dir (empty = disabled)
	ValidateIntegrity        bool                            // run post-generation checksum check
	Logger                   *logging.ZapLogger              // concrete logger (needed for Fatal)
	GenerationContextFactory func() (context.Context, error) // explicit access context; direct when nil
}

// Run executes the shared generation pipeline: profiling, OTel, tracing,
// signal handling, telemetry wrapper, integrity validation.
func Run(cfg RunConfig) {
	logger := cfg.Logger

	// --- Profiling ---
	profiling := false
	if cfg.ProfileDir != "" {
		if err := os.MkdirAll(cfg.ProfileDir, os.ModePerm); err != nil {
			logger.Fatal(err.Error())
		}

		gpf, err := os.Create(filepath.Join(cfg.ProfileDir, "goja.pprof"))
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer gpf.Close()

		ppf, err := os.Create(filepath.Join(cfg.ProfileDir, "cpu.pprof"))
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer ppf.Close()

		_ = goja.StartProfile(gpf)
		defer goja.StopProfile()
		_ = pprof.StartCPUProfile(ppf)
		defer pprof.StopCPUProfile()

		profiling = true
	}

	// --- Context setup ---
	ctx := generationaccess.WithDirect(context.Background())
	if cfg.GenerationContextFactory != nil {
		var err error
		ctx, err = cfg.GenerationContextFactory()
		if err != nil {
			logger.Fatal("Failed to configure generation access", zap.Error(err))
		}
	}

	// --- OTel ---
	if cfg.TraceDest != "" {
		otelShutdown, err := instrument.SetupOTelSDK(ctx, cfg.TraceDest)
		if err != nil {
			logger.Fatal("Failed to set up OpenTelemetry SDK", zap.Error(err))
		}
		defer func() {
			sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := otelShutdown(sctx); err != nil {
				logger.Fatal("Failed to shutdown OpenTelemetry SDK", zap.Error(err))
			}
		}()
	}

	// --- Tracer + span ---
	tracer := otel.Tracer("github.com/speakeasy-api/openapi-generation/v2")

	var genErr error
	ctx, span := tracer.Start(ctx, cfg.SpanName, trace.WithAttributes(
		attribute.String("language", cfg.GenOpts.Lang),
	))
	defer func() {
		span.RecordError(genErr)
		if genErr != nil {
			span.SetStatus(codes.Error, genErr.Error())
		}
		span.End()
	}()

	// --- Signal handling ---
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(2)
	}()

	// --- Inject logger + tracer into GenOpts ---
	cfg.GenOpts.Logger = logger
	cfg.GenOpts.Tracer = tracer

	// --- Generate with telemetry wrapper ---
	lang := cfg.GenOpts.Lang
	if genErr = generationtelemetry.Run(ctx, shared.InteractionTypeTargetGenerate, func(ctx context.Context, event *shared.CliEvent) error {
		event.GenerateTargetName = &lang

		if errs := generate.Generate(ctx, cfg.GenOpts); len(errs) > 0 {
			for _, e := range errs {
				logger.Error(e.Error())
			}
			logger.Fatal("Failed to generate SDK to " + cfg.GenOpts.OutDir)
		}

		return nil
	}); genErr != nil {
		logger.Fatal("Failed to generate SDK", zap.Error(genErr))
	}

	// --- Integrity validation ---
	if cfg.ValidateIntegrity {
		if err := runIntegrityValidation(logger, cfg.GenOpts.OutDir); err != nil {
			logger.Fatal(err.Error())
		}
		logger.Info("Integrity validation passed: all checksums match")
	}

	// --- Memory profile ---
	if profiling {
		mpf, err := os.Create(filepath.Join(cfg.ProfileDir, "mem.pprof"))
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer mpf.Close()
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(mpf, 0); err != nil {
			logger.Fatal(err.Error())
		}
	}
}

// ResolveWatchTemplatesDir resolves the template-watching directory from flags
// and environment variables.  Returns the resolved directory path, or "" if
// template watching is not enabled.
func ResolveWatchTemplatesDir(flagLocation string, flagWatch bool, logger *logging.ZapLogger) string {
	var location string
	enabled := false

	switch {
	case flagLocation != "":
		location = flagLocation
		enabled = true
	case flagWatch:
		location = env.WatchTemplatesLocation()
		if location == "" {
			logger.Fatal("--watch-templates requires WATCH_TEMPLATES_LOCATION environment variable to be set")
		}
		enabled = true
	case env.WatchTemplatesEnabled():
		envLocation := env.WatchTemplatesLocation()
		if envLocation != "" {
			location = envLocation
			enabled = true
		}
	}

	if flagLocation != "" && flagWatch {
		logger.Fatal("Use --watch-templates OR --watch-templates-location not both")
	}

	if !enabled {
		return ""
	}

	if location == "" {
		logger.Fatal("WATCH_TEMPLATES_LOCATION environment variable is not set")
	}

	if !strings.HasSuffix(location, "/templates") && !strings.HasSuffix(location, "/templates/") {
		logger.Fatal("Unexpected WATCH_TEMPLATES_LOCATION must end with /templates - eg /path/to/openapi-generation/templates/")
	}

	if _, err := os.Stat(location); os.IsNotExist(err) {
		logger.Fatal("Unexpected WATCH_TEMPLATES_LOCATION the directory '" + location + "' does not exist")
	}

	return location
}

// ResolveTraceDest merges the trace destination from the environment and a flag override.
func ResolveTraceDest(flagValue string) string {
	dest := os.Getenv("SPEAKEASY_OTEL_TRACE")
	if flagValue != "" {
		dest = flagValue
	}
	return dest
}

// runIntegrityValidation validates that all tracked files in gen.lock have
// checksums matching their actual content on disk.
func runIntegrityValidation(logger *logging.ZapLogger, outDir string) error {
	const checksumPrefix = "sha1:"

	genLockPath := filepath.Join(outDir, ".speakeasy", "gen.lock")
	data, err := os.ReadFile(genLockPath)
	if err != nil {
		return fmt.Errorf("failed to read gen.lock: %w", err)
	}

	lf, err := lockfile.Load(data)
	if err != nil {
		return fmt.Errorf("failed to parse gen.lock: %w", err)
	}

	if lf.TrackedFiles == nil {
		logger.Info("No tracked files found in gen.lock")
		return nil
	}

	var mismatches []string
	for relPath, tracked := range lf.TrackedFiles.All() {
		rawChecksum := tracked.LastWriteChecksum
		if rawChecksum == "" {
			continue
		}

		if !strings.HasPrefix(rawChecksum, checksumPrefix) {
			mismatches = append(mismatches, fmt.Sprintf("%s: unsupported checksum format %q (expected %s<hex>)", relPath, rawChecksum, checksumPrefix))
			continue
		}

		fullPath := filepath.Join(outDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				mismatches = append(mismatches, relPath+": file missing from disk")
			} else {
				mismatches = append(mismatches, fmt.Sprintf("%s: read error: %v", relPath, err))
			}
			continue
		}

		actualChecksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(content))
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: checksum error: %v", relPath, err))
			continue
		}

		expectedChecksum := strings.TrimPrefix(rawChecksum, checksumPrefix)
		if actualChecksum != expectedChecksum {
			mismatches = append(mismatches, fmt.Sprintf("%s: checksum mismatch (expected %s%s, got %s%s)", relPath, checksumPrefix, expectedChecksum, checksumPrefix, actualChecksum))
		}
	}

	if len(mismatches) > 0 {
		for _, m := range mismatches {
			logger.Error("Integrity check failed: " + m)
		}
		return fmt.Errorf("integrity validation failed: %d file(s) have checksum mismatches", len(mismatches))
	}

	return nil
}
