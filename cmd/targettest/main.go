package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap/zapcore"
)

var logger = logging.NewLogger(zapcore.DebugLevel)

func main() {
	outDir := flag.String("o", "", "path to existing generated target output directory")
	target := flag.String("t", "", "target name")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	if *outDir == "" {
		logger.Fatal("existing target output directory (-o) must be provided")
	}

	if *target == "" {
		logger.Fatal("target name (-t) must be provided")
	}

	// TODO: Some generator output, such as processrunner stdout, is tied to the
	// debug environment variable rather than generator flags.
	// Reference: internal issue reference
	_ = os.Setenv("SPEAKEASY_DEBUG", "true")
	_ = os.Setenv("SPEAKEASY_DISABLE_TELEMETRY", "true")

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())

	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithRunLocation("cli"),
	}

	if *verbose {
		opts = append(opts, generate.WithVerboseOutput(true))
	}

	g, err := generate.New(opts...)
	if err != nil {
		logger.Fatal("unable to create generator: " + err.Error())
	}

	fmt.Printf("Running target %s testing in %s ...\n", *target, *outDir)

	if err := g.RunTargetTesting(ctx, *target, *outDir); err != nil {
		logger.Fatal(fmt.Sprintf("error running target %s testing in %s: %s", *target, *outDir, err.Error()))
	}
}
