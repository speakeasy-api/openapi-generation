//go:build !js || !wasm

package instrument

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type exporterCleanupFunc func(context.Context) error

func exporterFromURI(ctx context.Context, idOrURI string) (trace.SpanExporter, exporterCleanupFunc, error) {
	switch {
	case idOrURI == "":
		exp, err := stdouttrace.New(stdouttrace.WithWriter(io.Discard))
		return exp, nil, err
	case idOrURI == "stderr":
		exp, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
		return exp, nil, err
	case idOrURI == "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithWriter(os.Stdout))
		return exp, nil, err
	case strings.HasPrefix(idOrURI, "file://"):
		path := strings.TrimSpace(idOrURI[len("file://"):])
		if path == "" {
			return nil, nil, fmt.Errorf("%s: empty file path", idOrURI)
		}
		f, err := os.Create(path)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: failed to open file: %w", idOrURI, err)
		}
		exp, err := stdouttrace.New(stdouttrace.WithWriter(f))
		return exp, func(context.Context) error { return f.Close() }, err
	case idOrURI == "grpc":
		opts := []otlptracegrpc.Option{
			// TODO: Migrate off deprecated option
			otlptracegrpc.WithDialOption(),
			otlptracegrpc.WithInsecure(),
		}
		traceClient := otlptracegrpc.NewClient(opts...)
		exp, err := otlptrace.New(ctx, traceClient)
		return exp, nil, err
	default:
		return nil, nil, fmt.Errorf("%s: unsupported exporter id or uri", idOrURI)
	}
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, uri string) (shutdown func(context.Context) error, err error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	traceExp, cleanup, err := exporterFromURI(dialCtx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// Specify the service name displayed on the backend of Managed Service for OpenTelemetry.
			semconv.ServiceNameKey.String("openapi-generation"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	bsp := trace.NewBatchSpanProcessor(traceExp)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	return func(ctx context.Context) error {
		if cleanup != nil {
			defer cleanup(ctx) //nolint:errcheck
		}

		if err := bsp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
		if err := traceExp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}

		return nil
	}, nil
}
