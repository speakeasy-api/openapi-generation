# Tracing with OpenTelemetry

The generator program can be run with tracing enabled by passing the flag
`-trace <target>`. The `<target>` argument can be:

- `file://path/to/trace.json`: Writes the trace to a file. This file can be
opened in various trace visualiser tools like Jaeger.

- `grpc`: Push traces to a collector that implements the OpenTelemetry gRPC
protocol e.g. Jaeger

- `stderr`: Writes the trace to directly `STDERR`.

The easiest way to get started is to run the Jaeger all-in-one Docker image:

```sh
docker run --rm --name jaeger -d \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.56
```

The web interface for Jaeger will be at http://localhost:16686.

Once that is running, run the generator like so:

```
go run cmd/generate/main.go -trace grpc -s <spec-path> -o <output-path> -l <lang>
```

You should start to see a trace created as the generator runs.

## Instrumenting code

The `*generate.Generator` type has a `tracer` field that can be used to start
new spans around operations in the generator code base. For example:

```go
func (g *Generator) resolveAST(ctx context.Context, doc *v3.Document, ad *analytics.Data) (*ast.SDK, []error) {
	ctx, span := g.tracer.Start(ctx, "Generator.resolveAST")
	defer span.End()

  // rest of code
}
```
