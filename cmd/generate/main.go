package main

import (
	"flag"
	"os"

	_ "github.com/KimMachineGun/automemlimit"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/runner"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

func main() {
	schema := flag.String("s", "", "path to the schema file")
	outDir := flag.String("o", "", "path to the output directory")
	lang := flag.String("l", "", "language to generate")
	testGroup := flag.String("t", "", "test group to generate")
	usageGroup := flag.String("u", "", "The namespace to generate usage snippets for (or 'all' for all subSDKs)")
	published := flag.Bool("p", false, "is the SDK published")
	installationURL := flag.String("i", "", "installation url for the SDK")
	repoURL := flag.String("r", "", "repo url for the SDK")
	repoSubDirectory := flag.String("b", "", "repo subdirectory for the SDK")
	skipCompile := flag.Bool("skip-compile", false, "skip compiling the SDK")
	skipVersioning := flag.Bool("skip-versioning", false, "skip automatic SDK version bumps")
	traceDest := flag.String("trace", "", "enable OpenTelemetry tracing (allowed values: stderr, grpc, file://<path>)")
	verbose := flag.Bool("v", false, "verbose output")
	watchTemplatesLocation := flag.String("watch-templates-location", "", "watch the given templates directory for changes and regenerate when they change use this or --watch-templates not both")
	watchTemplates := flag.Bool("watch-templates", false, "watch the templates directory you must set environment WATCH_TEMPLATES_LOCATION=... use this or --watch-templates-location not both")
	profile := flag.String("profile", "", "start profiling and save files to the given directory")
	validateIntegrity := flag.Bool("validate-integrity", false, "after generation, validate that gen.lock checksums match actual file contents on disk")
	licenseTokenPath := flag.String("license-token", "", "path to a Speakeasy license token file (a registry-signed JWT); a valid token generates with authenticated commercial access. Defaults to the SPEAKEASY_LICENSE_TOKEN environment variable (the raw token) when unset")
	licenseElection := flag.String("license", "", "license election for generated output: 'agpl-3.0-only' accepts AGPL-3.0-only licensing, 'commercial' asserts that a license token must be present. Defaults to the SPEAKEASY_GENERATED_LICENSE environment variable. Without an election a valid license token is required — output is never silently licensed")
	flag.Parse()

	logger := logging.NewLogger(logging.LevelFromEnv())

	generationContextFactory := runner.ResolveGenerationContextFactory(*licenseElection, *licenseTokenPath, os.Getenv)

	runner.Run(runner.RunConfig{
		SpanName:                 "main",
		TraceDest:                runner.ResolveTraceDest(*traceDest),
		ProfileDir:               *profile,
		ValidateIntegrity:        *validateIntegrity,
		Logger:                   logger,
		GenerationContextFactory: generationContextFactory,
		GenOpts: generate.GenerateOptions{
			Lang:                   *lang,
			SchemaPath:             *schema,
			OutDir:                 *outDir,
			TestGroup:              *testGroup,
			UsageGroup:             *usageGroup,
			Published:              *published,
			InstallationURL:        *installationURL,
			RepoURL:                *repoURL,
			RepoSubDirectory:       *repoSubDirectory,
			SkipCompile:            *skipCompile,
			SkipVersioning:         *skipVersioning,
			Verbose:                *verbose,
			WatchTemplatesLocation: runner.ResolveWatchTemplatesDir(*watchTemplatesLocation, *watchTemplates, logger),
		},
	})
}
