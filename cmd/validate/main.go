package main

import (
	"context"
	goerrs "errors"
	"flag"
	"fmt"
	"os"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/instrument"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func main() {
	schema := flag.String("s", "", "path to the schema file")
	ruleset := flag.String("r", validation.RulesetSpeakeasyRecommended, "ruleset to use for validation")
	traceDest := flag.String("trace", "", "enable OpenTelemetry tracing (allowed values: stderr, grpc, file://<path>)")
	flag.Parse()

	_ = os.Setenv("SPEAKEASY_DISABLE_TELEMETRY", "true")
	_ = os.Setenv("SPEAKEASY_DEBUG", "true")

	logger := logging.NewLogger(logging.LevelFromEnv())

	ctx := generationaccess.WithDirect(logging.With(context.Background(), logger))

	tdest := os.Getenv("SPEAKEASY_OTEL_TRACE")
	if traceDest != nil && *traceDest != "" {
		tdest = *traceDest
	}
	if tdest != "" {
		otelShutdown, err := instrument.SetupOTelSDK(ctx, tdest)
		if err != nil {
			logger.Fatal("Failed to set up OpenTelemetry SDK", zap.Error(err))
		}
		defer func() {
			sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := goerrs.Join(err, otelShutdown(sctx))
			if err != nil {
				logger.Fatal("Failed to set shutdown OpenTelemetry SDK", zap.Error(err))
			}
		}()
	}

	tracer := otel.Tracer("github.com/speakeasy-api/openapi-generation/v2")

	var err error
	ctx, span := tracer.Start(ctx, "main")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if err := validate(ctx, *schema, *ruleset); err != nil {
		logger.Fatal(err.Error())
	}
}

func validate(ctx context.Context, schemaPath, ruleset string) error {
	logger := logging.From(ctx)
	fmt.Println("Validating OpenAPI Document")

	start := time.Now()

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	opts := make([]generate.GeneratorOptions, 0, 4)
	opts = append(opts, generate.WithLogger(logger))
	opts = append(opts, generate.WithValidationRuleset(ruleset))
	opts = append(opts, generate.WithRunLocation("cli"))
	opts = append(opts, generate.WithDebuggingEnabled())

	g, err := generate.New(opts...)
	if err != nil {
		return err
	}

	hasWarnings := false
	hasHints := false

	res, err := g.Validate(ctx, schema, schemaPath, false, "")
	if err != nil {
		return err
	}
	vErrs := res.GetValidationErrors()
	hasErrors := false
	for _, err := range vErrs {
		vErr := errors.GetValidationErr(err)
		uErr := errors.GetUnsupportedErr(err)

		switch {
		case vErr != nil:
			switch vErr.Severity {
			case errors.SeverityError:
				hasErrors = true
				logger.Error(err.Error())
			case errors.SeverityWarn:
				hasWarnings = true
				logger.Warn(err.Error())
			default:
				hasHints = true
				logger.Info(err.Error())
			}
		case uErr != nil:
			hasWarnings = true
			logger.Warn(err.Error())
		default:
			hasErrors = true
			logger.Error(err.Error())
		}
	}

	rf, err := os.CreateTemp("", "lint-report-*.html")
	if err == nil {
		defer rf.Close()
		_, _ = rf.Write(res.GenerateReport())
		fmt.Printf("Validation report written to %s\n", rf.Name())
	}

	switch {
	case hasErrors:
		return errors.New("OpenAPI spec invalid ✖")
	case hasWarnings:
		fmt.Println("OpenAPI Document is valid with warnings")
	case hasHints:
		fmt.Println("OpenAPI Document is valid with suggestions")
	default:
		fmt.Println("OpenAPI Document is valid ✔")
	}

	logging.From(ctx).Info(fmt.Sprintf("Validated document in %s", time.Since(start)))

	return nil
}
