package generate

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	"github.com/speakeasy-api/openapi/openapi"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/internal/buckettypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/instrument"
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/targetconfig"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap/zapcore"

	templatesConfig "github.com/speakeasy-api/openapi-generation/v2/pkg/templates"

	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"gopkg.in/yaml.v3"

	"github.com/posthog/posthog-go"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/cms"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/templates"

	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi-generation/v2/internal/versioning"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"go.uber.org/zap"
)

type (
	GeneratorOptions func(g *Generator)
	ValidateOpts     struct {
		Schema     []byte
		SchemaPath string
		IsRemote   bool
		WorkingDir string
		Target     string
	}
)

func WithFileSystem(fs filesystem.FileSystem) GeneratorOptions {
	return func(g *Generator) {
		g.fs = fs
	}
}

// WithGit injects the Git interface for all Git operations (blobbing, tree creation, commits).
// This is required for the 3-way merge / persistentEdits feature.
func WithGit(git merge.Git) GeneratorOptions {
	return func(g *Generator) {
		g.git = git
	}
}

func WithDontWrite() GeneratorOptions {
	return func(g *Generator) {
		g.dontWrite = true
	}
}

func WithLogger(log logging.Logger) GeneratorOptions {
	return func(g *Generator) {
		g.log = log
	}
}

func WithTracer(tracer trace.Tracer) GeneratorOptions {
	return func(g *Generator) {
		g.tracer = tracer
	}
}

func WithAnalytics(client posthog.Client) GeneratorOptions {
	return func(g *Generator) {
		g.posthog = client
	}
}

func WithCustomerID(id string) GeneratorOptions {
	return func(g *Generator) {
		g.customerID = id
	}
}

func WithSkipVersioning(skipVersioning bool) GeneratorOptions {
	return func(g *Generator) {
		g.skipVersioning = skipVersioning
	}
}

func WithWorkspaceID(id string) GeneratorOptions {
	return func(g *Generator) {
		g.workspaceID = id
	}
}

func WithWorkspaceUri(uri string) GeneratorOptions {
	return func(g *Generator) {
		g.workspaceUri = uri
	}
}

func WithRunLocation(location string) GeneratorOptions {
	return func(g *Generator) {
		g.runLocation = location
	}
}

func WithDebuggingEnabled() GeneratorOptions {
	return func(g *Generator) {
		g.debug = true
	}
}

// WithSplitModelThreshold overrides the default large-model splitting
// thresholds. Models with more than `threshold` types are split into chunks
// of `chunkSize`. This is useful in tests to exercise splitting logic without
// needing hundreds of types.
func WithSplitModelThreshold(threshold, chunkSize int) GeneratorOptions {
	return func(g *Generator) {
		g.splitModelThreshold = threshold
		g.splitModelChunkSize = chunkSize
	}
}

func WithGenVersion(version string) GeneratorOptions {
	return func(g *Generator) {
		g.genVersion = version
	}
}

func WithCLIVersion(version string) GeneratorOptions {
	return func(g *Generator) {
		g.cliVersion = version
	}
}

func WithOutputTests() GeneratorOptions {
	return func(g *Generator) {
		g.outputTests = true
	}
}

func WithOutputTestGroup(testGroup string) GeneratorOptions {
	return func(g *Generator) {
		g.testGroup = testGroup
	}
}

func WithOutputUsageGroup(usageGroup string) GeneratorOptions {
	if usageGroup == "all" {
		return func(g *Generator) {
			g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
				All: true,
			}
		}
	}

	return func(g *Generator) {
		g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
			Namespace: usageGroup,
		}
	}
}

func WithInstallationURL(installationURL string) GeneratorOptions {
	return func(g *Generator) {
		g.installationURL = installationURL
	}
}

func WithPublished(published bool) GeneratorOptions {
	return func(g *Generator) {
		g.published = published
	}
}

func WithRepoDetails(repoURL, subDirectory string) GeneratorOptions {
	return func(g *Generator) {
		g.repoURL = repoURL
		g.repoSubDirectory = subDirectory
	}
}

func WithUsageSnippetArgsByRootExample() GeneratorOptions {
	return func(g *Generator) {
		g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
			Standalone:  true,
			RootExample: true,
		}
	}
}

// WithUsageSnippetArgsByOperationID is used to generate usage snippets for a specific operation ID or list of operation IDs
// We keep the string parameter to maintain backwards compatibility in the CLI
func WithUsageSnippetArgsByOperationID(operationIDs string) GeneratorOptions {
	return func(g *Generator) {
		g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
			Standalone:   true,
			OperationIDs: strings.Split(operationIDs, ","),
		}
	}
}

func WithUsageSnippetArgsByNamespace(namespace string) GeneratorOptions {
	return func(g *Generator) {
		g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
			Standalone: true,
			Namespace:  namespace,
		}
	}
}

func WithUsageSnippetArgsGenerateAll() GeneratorOptions {
	return func(g *Generator) {
		g.generateUsageSnippetArgs = &generateUsageSnippetArgs{
			Standalone: true,
			All:        true,
		}
	}
}

func WithUsageSnippetExampleRequestBody(requestBodyJSON string) GeneratorOptions {
	return func(g *Generator) {
		if g.generateUsageSnippetArgs == nil {
			g.generateUsageSnippetArgs = &generateUsageSnippetArgs{}
		}
		g.generateUsageSnippetArgs.ExampleValues.RequestBodyJSON = requestBodyJSON
	}
}

func WithUsageSnippetExampleParams(paramNameToExampleValue map[string]string) GeneratorOptions {
	return func(g *Generator) {
		if g.generateUsageSnippetArgs == nil {
			g.generateUsageSnippetArgs = &generateUsageSnippetArgs{}
		}
		g.generateUsageSnippetArgs.ExampleValues.Params = paramNameToExampleValue
	}
}

type standaloneReadmeConfig struct {
	FileName    string // the name of the standalone readme file
	HeadersOnly bool   // whether to template section headers only (no content)
}

func WithStandaloneReadme(filename string, headersOnly bool) GeneratorOptions {
	return func(g *Generator) {
		g.standaloneReadmeConfig = &standaloneReadmeConfig{
			FileName:    filename,
			HeadersOnly: headersOnly,
		}
	}
}

func WithProgressUpdates(targetID string, onProgressUpdate func(ProgressUpdate), updateGenSteps, updateFileStatus bool) GeneratorOptions {
	return func(g *Generator) {
		g.progress = Progress{
			TargetID:         targetID,
			OnProgressUpdate: onProgressUpdate,
			UpdateGenSteps:   updateGenSteps,
			UpdateFileStatus: updateFileStatus,
		}
	}
}

func WithValidationRuleset(ruleset string) GeneratorOptions {
	return func(g *Generator) {
		g.validationRuleset = ruleset
	}
}

func WithParseValidOperations() GeneratorOptions {
	return func(g *Generator) {
		g.parseValidOperations = true
	}
}

func WithWarningLoggerDisabled() GeneratorOptions {
	return func(g *Generator) {
		g.warningLoggerDisabled = true
	}
}

// WithCompileEnv sets extra environment variables on compile and lint child
// processes. This is useful for isolating per-invocation tool state (e.g.
// GEM_HOME, BUNDLE_PATH) when multiple generations run in parallel.
func WithCompileEnv(env map[string]string) GeneratorOptions {
	return func(g *Generator) {
		g.compileEnv = env
	}
}

type generateUsageSnippetArgs struct {
	Standalone    bool     // true only in the context of standalone snippet generation
	OperationIDs  []string // the operation IDs to generate usage examples for
	Namespace     string   // the group identifier to generate usage examples for
	RootExample   bool     // if true, only the first main usage example gets generated
	All           bool     // if true, generate all available usage examples
	OutputDir     string   // if not Standalone, holds the testproject output directory
	ExampleValues struct {
		Params          map[string]string // Map from param `name` to the example value to use for that param
		RequestBodyJSON string            // The request body JSON to use for the example
	}
}

func (g *Generator) generateStandaloneUsage() bool {
	return g.generateUsageSnippetArgs != nil && g.generateUsageSnippetArgs.Standalone
}

// RenderedUsageSnippets holds pre-rendered standalone usage snippets produced
// as a side effect of main SDK generation, avoiding a separate AST resolution pass.
type RenderedUsageSnippets struct {
	RawOutput string // Concatenated snippet files in the "// Usage snippet provided for ..." format
}

func WithRenderUsageSnippets() GeneratorOptions {
	return func(g *Generator) {
		g.renderUsageSnippets = true
	}
}

func WithVerboseOutput(verbose bool) GeneratorOptions {
	return func(g *Generator) {
		g.verboseOutput = verbose
		g.log = logging.NewLogger(zapcore.DebugLevel)
		g.tracer = otel.Tracer("github.com/speakeasy-api/openapi-generation/v2")
	}
}

func WithForceGeneration() GeneratorOptions {
	return func(g *Generator) {
		g.forceGeneration = true
	}
}

func WithChangelogReleaseNotes(releaseNotes string) GeneratorOptions {
	return func(g *Generator) {
		g.releaseNotes = releaseNotes
	}
}

func WithWatchTemplatesLocation(location string) GeneratorOptions {
	return func(g *Generator) {
		g.watchTemplatesLocation = location
	}
}

func WithDisableMockServer() GeneratorOptions {
	return func(g *Generator) {
		g.disableMockServer = true
	}
}

// Generator is the central tool for handling all things related to SDK
// generation and contains all configuration and options for customizing and
// holding state of the process.
type Generator struct {
	lockFile *config.LockFile // the parsed lock file (usually from gen.lock)

	// target describes the target itself, such as its name, template directory,
	// and maturity status.
	target types.Target

	// targetConfig is the target-to-generator configuration, which is fetched
	// from the target after fetching config.
	targetConfig *targetconfig.Configuration

	validationOnly bool
	openAPIVersion string
	schemas        *schemas.Schemas
	comments       *cms.CommentTracker
	ignore         *filetracking.Ignore
	namer          *namer.Namer
	subsystem      *subsystem.Subsystem
	bucketer       *buckettypes.Bucketer

	// outDir is the main working directory for most of the Generator
	// functionality including loading configurations and the root directory for
	// writing files. Use outSubDir for additional target generation within a
	// Generator.
	outDir string

	// outSubDir is a working directory relative to outDir. It is used for
	// additional target generation file writing and compilation, such as the
	// mockserver target underneath another target.
	outSubDir string

	result                *validation.Result
	warningLogger         *logging.WarningLogger
	warningLoggerDisabled bool

	// Options
	// Architecture: Functional Core / Imperative Shell
	// These interfaces are injected by the CLI to keep Generator pure and WASM-compatible.
	fs        filesystem.FileSystem // abstraction for disk I/O (ReadFile, WriteFile, etc.)
	git       merge.Git             // abstraction for Git operations (blob, tree, commit, push)
	dontWrite bool                  // if true, don't write any files

	// Virtual files collected during generation (when git interface is available)
	virtualFiles SyncMap[string, *merge.VirtualFile]

	// mergeWasNoOp is set by handlePatches when merge detects no changes needed
	mergeWasNoOp bool

	log         logging.Logger
	tracer      trace.Tracer
	posthog     posthog.Client
	customerID  string
	workspaceID string
	// Used to link to the workspace in the readme
	workspaceUri             string
	runLocation              string
	debug                    bool
	docVersion               string
	genVersion               string
	cliVersion               string
	outputTests              bool
	testGroup                string
	installationURL          string
	published                bool
	repoURL                  string
	repoSubDirectory         string
	generateUsageSnippetArgs *generateUsageSnippetArgs
	standaloneReadmeConfig   *standaloneReadmeConfig
	renderUsageSnippets      bool                   // opt-in: render standalone snippets as a side output of SDK generation
	renderedUsageSnippets    *RenderedUsageSnippets // populated after generation if renderUsageSnippets is true
	verboseOutput            bool
	validationRuleset        string
	parseValidOperations     bool
	skipVersioning           bool

	// Enabled when the SPEAKEASY_FORCE_GENERATION environment variable is set
	// to "true" or the WithForceGeneration option is used.
	forceGeneration bool

	// Pass your directory to activate "watch" mode. If changes are detected to the templates directory the SDK will be automatically regenerated.
	// Warning: changes to the spec or to the go code won't live reload
	watchTemplatesLocation      string
	shouldReadTemplatesFromDisk bool

	// Used to keep track of the current generation step and optionally "stream" progress updates to the CLI.
	progress Progress

	// Disables the mock server from starting and overriding the TEST_SERVER_URL env var. The mock server will still be generated.
	disableMockServer bool

	// mockServerSubDir is the outSubDir resolved for the mock server target during this run, if any.
	// Populated by maybeGenerateMockServer so cleanup can scope a directory walk to that subtree.
	mockServerSubDir string
	// mockServerRegenerated indicates whether the mock server actually re-templated this run.
	// When false, the directory walk is skipped — newFiles holds nothing for the mock server,
	// so a walk would mistake the entire prior subtree for orphans.
	mockServerRegenerated bool
	// ReleaseNotes to be shown when releasing the binary
	releaseNotes string

	// Overrides for large-model splitting thresholds (0 = use defaults).
	splitModelThreshold int
	splitModelChunkSize int

	// compileEnv holds extra environment variables to set on compile/lint
	// child processes. This is used by snapshot tests to isolate per-test
	// gem/bundle directories and prevent parallel Ruby compilations from
	// racing on shared system state.
	compileEnv map[string]string

	generatedLicense generationaccess.GeneratedLicense
	licenseResolved  bool
}

// New constructs a new Generator with some default values and then applies all
// the given options and returns it.
//
// It returns an no generator and an error if the given options are invalid. The
// following situations are considered an error:
//
// * Either a WriteFileFunc must be provided or the validation only setting must
// be set
func New(opts ...GeneratorOptions) (*Generator, error) {
	g := &Generator{
		genVersion: "internal",
		cliVersion: "internal",
		comments:   cms.New("sdk"),
		subsystem:  subsystem.New(nil),
	}
	if env.IsForceGeneration() {
		opts = append(opts, WithForceGeneration())
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.log == nil {
		g.log = logging.NewLogger(logging.LevelFromEnv())
	}

	if g.tracer == nil {
		g.tracer = noop.NewTracerProvider().Tracer("github.com/speakeasy-api/openapi-generation/v2")
	}

	return g, nil
}

// createMockServerGenerator returns a copied Generator for the mock server target.
//
// Certain generator components, such as the file tracker, output directory,
// and tracer, remain the same as the original.
func (g *Generator) createMockServerGenerator(ctx context.Context) (*Generator, error) {
	targetName := "mockserver"

	target, err := g.GetTarget(ctx, targetName, g.outDir)
	if err != nil {
		return nil, fmt.Errorf("error getting target %q during clone: %w", targetName, err)
	}

	result := &Generator{
		cliVersion: g.cliVersion,
		comments:   cms.New("sdk"),
		// config intentionally handled later due to getLanguageConfigDefaults g.target.Target checking
		customerID:       g.customerID,
		debug:            g.debug,
		dontWrite:        g.dontWrite,
		forceGeneration:  g.forceGeneration,
		fs:               g.fs,
		genVersion:       g.genVersion,
		generatedLicense: g.generatedLicense,
		licenseResolved:  g.licenseResolved,
		ignore:           g.ignore,
		lockFile:         g.lockFile,
		log:              g.log,
		outDir:           g.outDir,
		posthog:          g.posthog,
		// result intentionally not cloned
		runLocation: g.runLocation,
		target:      target,
		// targetConfig intentionally handled later due to needing config
		tracer:         g.tracer,
		validationOnly: g.validationOnly,
		verboseOutput:  g.verboseOutput,
		workspaceID:    g.workspaceID,
		workspaceUri:   g.workspaceUri,
		testGroup:      g.testGroup,
	}

	cfg, err := config.GetDefaultConfig(true, result.getLanguageConfigDefaults, map[string]bool{targetName: true})
	if err != nil {
		return nil, err
	}

	c := configuration.New(cfg, targetName)
	// Copy the tests config from the original config as it also drives the behaviour of the mock server
	c.Generation.Tests = g.subsystem.Config.Generation.Tests
	// Copy persistent edits config so the file tracker options (SkipTempLockFile) match the parent target
	c.Generation.PersistentEdits = g.subsystem.Config.Generation.PersistentEdits

	// We must keep the same file tracker instance as it contains an in memory
	// list of files that are being written to the file system, if we create a new
	// instance then the in memory list of files will be lost and never written to gen.lock
	ss := subsystem.New(g.subsystem.FileTracker)
	if err := ss.Init(ctx, subsystem.InitOptions{
		Target:              target,
		Config:              c,
		BaseTemplateConfigs: result.getBaseTemplateConfigs(result.target, c, result.lockFile),
		OutDir:              result.outDir,
		GenLockId:           result.lockFile.ID,
		FS:                  g,
		Git:                 g.git,
		LockFile:            result.lockFile,
	}); err != nil {
		return nil, fmt.Errorf("error initializing subsystem for target %q during clone: %w", targetName, err)
	}
	result.subsystem = ss

	sanitizer, err := result.getSanitizer(ctx, target)
	if err != nil {
		return nil, err
	}

	result.subsystem.Sanitizer = sanitizer
	result.bucketer = buckettypes.New(result.subsystem)
	if result.splitModelThreshold > 0 {
		result.bucketer.SetSplitThresholds(result.splitModelThreshold, result.splitModelChunkSize)
	}
	if names := result.getReservedModelFileNames(); len(names) > 0 {
		result.bucketer.SetReservedModelFileNames(names)
	}

	result.namer = namer.New(ss)

	if err := result.getTargetInitialConfiguration(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

// Generate generates the SDK. It takes the content of the schema file and the
// path to that file. This performs generation for the single language templates
// named and writes all files to the given output directory.
//
// It returns a slice of errors that occurred during generation.
// ctx must carry generation access state (generation-context/access) or a license token (licensetoken.WithToken).
func (g *Generator) Generate(ctx context.Context, schema []byte, schemaPath, target, outDir string, isRemote, compileOutput bool) []error {
	ctx, err := g.establishLicense(ctx, target, outDir)
	if err != nil {
		return []error{err}
	}

	g.warningLogger = logging.NewWarningLogger(g.validationOnly, func(err error) {
		var vErr *errors.ValidationError
		if errors.As(err, &vErr) && g.result != nil {
			g.result.Insert(vErr.GetResult())
		}
	})
	if g.warningLoggerDisabled {
		g.warningLogger.DisableLogging()
	}

	ctx = logging.With(ctx, g.log)
	ctx = logging.WithWarningLogger(ctx, g.warningLogger)

	var span trace.Span
	if g.tracer != nil && g.verboseOutput {
		shutdown, err := instrument.SetupOTelSDK(ctx, "stdout")
		if err == nil {
			defer shutdown(ctx) //nolint:errcheck
		}
		ctx, span = g.tracer.Start(ctx, "Generator.Generate")
		defer span.End()
	}

	// * * Setup Environment * * //
	if err := g.Init(ctx, target, outDir); err != nil {
		return []error{err}
	}

	ad := &analytics.Data{}

	// * * Load and Validate Document * * //
	docResult, errs := g.loadAndValidateDocument(ctx, schema, schemaPath, isRemote, ad)
	if len(errs) > 0 || !docResult.runGenerator {
		return errs
	}

	// * * Generate SDK * * //
	sdkResult, errs := g.generateSDK(ctx, docResult.docInfo, docResult.docChecksum, ad)
	if len(errs) > 0 || sdkResult.results == nil {
		return errs
	}

	// * * Generate Mock Server * * //
	mockServerGenerator, errs := g.maybeGenerateMockServer(ctx, sdkResult.results, docResult.docInfo, ad)
	if len(errs) > 0 {
		return errs
	}

	// Mock server files should participate in the same patch/compile pipeline as
	// the parent target, so we absorb their virtual files before patch replay.
	if mockServerGenerator != nil {
		g.absorbVirtualFiles(mockServerGenerator)
	}

	var conflictsErr *merge.ConflictsError
	var finalCompileErrs []error

	// Apply any patch files before cleanup, tracking, compile, or persistent-edits merge.
	if err := g.applyImplicitPatchFiles(ctx); err != nil {
		// Patch-file conflicts still need cleanup and lockfile persistence so the
		// next run sees updated tracked file state. We stop before compile and
		// persistent-edits merge below, then return the conflict after saving.
		if !stderrors.As(err, &conflictsErr) {
			return []error{err}
		}
	}

	// * * Cleanup Orphaned Files * * //
	if errs := g.cleanOrphanedFiles(ctx); len(errs) > 0 {
		return errs
	}

	// * * Save Tracked Files (pre-compile) * * //
	// Save tracked files to gen.lock before compile. This ensures that even if
	// compilation fails or is cancelled, the files are recorded and can be cleaned
	// up on subsequent runs. Checksums are raw (pre-format) at this point.
	if err := g.saveTrackedFilesToLockFile(ctx); err != nil {
		return []error{fmt.Errorf("failed to save tracked files: %w", err)}
	}

	// Determine which compile passes to run (computed outside conflictsErr guard
	// so shouldCompileFinal is available after lockfile save)
	shouldCompilePristine, shouldCompileFinal := compilePlan(
		compileOutput,
		g.subsystem.Config.Generation.PersistentEdits.ShouldCompilePristine(),
		g.subsystem.Patches.IsEnabled(),
	)

	if conflictsErr == nil {

		if shouldCompilePristine {
			// Pristine compile: normalize files to post-format state so checksums are accurate
			if errs := g.compilePristine(ctx, mockServerGenerator); len(errs) > 0 {
				return errs
			}
		}

		// * * Handle custom user code (if persistentEdits enabled) * * //
		if isNoOp, patchErr := g.handlePatches(ctx); patchErr != nil {

			// If base content could not be retrieved, return immediately
			var snapshotRefErr *merge.SnapshotRefError
			if stderrors.As(patchErr, &snapshotRefErr) {
				return []error{patchErr}
			}

			// Check if it's a conflicts error - we still want to save the lockfile
			// so the updated pristine hashes are persisted for the next run
			if !stderrors.As(patchErr, &conflictsErr) {
				return []error{patchErr}
			}
			// Continue with lockfile save, will return the conflicts error at the end
		} else {
			g.mergeWasNoOp = isNoOp
		}

	}

	// Final compile runs after patch application but before the final config/lockfile
	// write is completed. That lets us persist metadata even if compilation fails,
	// while still capturing any successful formatter/linter mutations in gen.lock.
	if conflictsErr == nil && shouldCompileFinal {
		finalCompileErrs = g.compileAll(ctx, mockServerGenerator)
		if len(finalCompileErrs) == 0 {
			if err := g.syncPristineFromCompiledFiles(); err != nil {
				return []error{fmt.Errorf("failed to sync files after final compile: %w", err)}
			}
		}
	}

	// * * Save config * * //
	configChecksum, errs := g.saveConfig(ctx)
	if len(errs) > 0 {
		return errs
	}
	// Use the new checksum if config was saved, otherwise use the one from SDK generation
	if configChecksum == "" {
		configChecksum = sdkResult.configChecksum
	}

	// * * Update LockFile * * //
	if errs := g.updateLockFile(ctx, configChecksum, docResult.docChecksum, sdkResult.results.sdkVersion, conflictsErr); len(errs) > 0 {
		return errs
	}
	if len(finalCompileErrs) > 0 {
		return finalCompileErrs
	}

	g.logStep(ProgressStepDone)
	g.enrichTelemetryEventPostGeneration(ctx)

	return nil
}

// loadAndValidateDocumentResult holds the results of document loading and validation.
type loadAndValidateDocumentResult struct {
	docInfo      *document.DocumentInfo
	docChecksum  string
	runGenerator bool
}

// loadAndValidateDocument loads and validates the OpenAPI document.
// Returns the document info, checksum, and whether generation should proceed.
func (g *Generator) loadAndValidateDocument(ctx context.Context, schema []byte, schemaPath string, isRemote bool, ad *analytics.Data) (result loadAndValidateDocumentResult, errs []error) {
	err := g.withStep(ctx, ProgressStepValidate, func(ctx context.Context) error {
		result.docInfo = &document.DocumentInfo{
			Schema:     schema,
			SchemaPath: schemaPath,
			IsRemote:   isRemote,
		}

		res, err := g.LoadAndValidateDoc(ctx, result.docInfo, ad, g.outDir, false)
		if err != nil {
			errs = []error{err}
			return err
		}
		g.result = res.Result
		if res.Result.HasFatalErrors() {
			errs = res.Result.GetValidationErrors()
			return stderrors.New("validation errors")
		} else {
			vErrs := res.Result.GetValidationErrors()
			for _, vErr := range vErrs {
				logging.From(ctx).Info(vErr.Error())
			}
		}

		result.runGenerator = res.RunGenerator
		if result.runGenerator {
			result.docChecksum = getSchemaChecksum(schema)
		}
		return nil
	})
	if err != nil && len(errs) == 0 {
		errs = []error{err}
	}
	return result, errs
}

// generateSDKResult holds the results of SDK generation.
type generateSDKResult struct {
	results        *generateResults
	configChecksum string
}

// generateSDK generates the SDK code.
func (g *Generator) generateSDK(ctx context.Context, docInfo *document.DocumentInfo, docChecksum string, ad *analytics.Data) (result generateSDKResult, errs []error) {
	err := g.withStep(ctx, ProgressStepGenSDK, func(ctx context.Context) error {
		configOpts := []config.Option{
			config.WithFileSystem(g),
		}

		var err error
		result.configChecksum, err = config.GetConfigChecksum(g.outDir, configOpts...)
		if err != nil && !g.validationOnly && !mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
			logging.LogWarning(ctx, "failed to get config checksum", err)
		}

		if err := g.setupExclusions(ctx); err != nil {
			return err
		}

		result.results, err = g.startGenerate(ctx, "generate", docInfo, g.getVersioningInfo(docInfo.Doc, docChecksum, result.configChecksum), ad)
		return err
	})
	if err != nil {
		errs = []error{err}
	}
	return result, errs
}

// maybeGenerateMockServer generates the mock server if enabled.
func (g *Generator) maybeGenerateMockServer(ctx context.Context, results *generateResults, docInfo *document.DocumentInfo, ad *analytics.Data) (*Generator, []error) {
	if !results.generateMockServer {
		return nil, nil
	}

	var mockServerGenerator *Generator
	var mockServerErrs []error
	err := g.withStep(ctx, ProgressStepGenMockServer, func(ctx context.Context) error {
		var err error
		mockServerGenerator, err = g.createMockServerGenerator(ctx)
		if err != nil {
			return err
		}

		mockServerDirectory, err := g.getMockServerDirectory(ctx, g.target)
		if err != nil {
			return err
		}

		mockServerGenerator.outSubDir = mockServerDirectory
		g.mockServerSubDir = mockServerDirectory

		g.mockServerRegenerated, mockServerErrs = mockServerGenerator.generateMockServer(ctx, docInfo, ad)
		if len(mockServerErrs) > 0 {
			return mockServerErrs[0]
		}

		// Delete the mockserver config as this shouldn't be editable by the end user and always use the defaults
		delete(g.subsystem.Config.Languages, "mockserver")
		return nil
	})
	if err != nil {
		if len(mockServerErrs) > 0 {
			return nil, mockServerErrs
		}
		return nil, []error{err}
	}
	return mockServerGenerator, nil
}

func (g *Generator) saveConfig(ctx context.Context) (configChecksum string, errs []error) {
	if mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
		return "", nil
	}

	configOpts := []config.Option{
		config.WithFileSystem(g),
	}

	g.subsystem.Config.Generation.SyncNameResolution()

	if err := config.SaveConfig(g.outDir, &g.subsystem.Config.Configuration, configOpts...); err != nil {
		return "", []error{err}
	}

	var err error
	configChecksum, err = config.GetConfigChecksum(g.outDir, configOpts...)
	if err != nil && !g.validationOnly {
		logging.LogWarning(ctx, "failed to get config checksum", err)
	}

	return configChecksum, nil
}

// cleanOrphanedFiles removes files that are no longer needed.
func (g *Generator) cleanOrphanedFiles(ctx context.Context) []error {
	var errs []error
	err := g.withStep(ctx, ProgressStepCleanup, func(ctx context.Context) error {
		fileTrackerResult := g.subsystem.FileTracker.GetResult()

		// Build list of previously generated files, using MovedTo paths where files were relocated
		previousGeneratedFiles := make([]string, 0, g.lockFile.TrackedFiles.Len())
		for path := range g.lockFile.TrackedFiles.Keys() {
			if tracked, ok := g.lockFile.TrackedFiles.Get(path); ok && tracked.MovedTo != "" {
				previousGeneratedFiles = append(previousGeneratedFiles, tracked.MovedTo)
			} else {
				previousGeneratedFiles = append(previousGeneratedFiles, path)
			}
		}

		// Also include files from old generatedFiles format (stored in AdditionalProperties)
		// This ensures orphan cleanup works for lockfiles that haven't been migrated yet
		previousGeneratedFiles = append(previousGeneratedFiles, extractLegacyGeneratedFiles(g.lockFile.AdditionalProperties)...)

		previousFiles := slices.Concat(previousGeneratedFiles, fileTrackerResult.PreviousFiles)

		// Build list of newly generated files, including both original and relocated paths
		newFilesSlice := make([]string, 0, fileTrackerResult.NewFiles.Len())
		for path := range fileTrackerResult.NewFiles.Keys() {
			newFilesSlice = append(newFilesSlice, path)
			if tracked, ok := g.lockFile.TrackedFiles.Get(path); ok && tracked.MovedTo != "" {
				newFilesSlice = append(newFilesSlice, tracked.MovedTo)
			}
		}

		onRemovedFile := func(path string) {
			g.sendFileProgressUpdate(path, nil)
		}
		if err := filetracking.CleanOrphanedFiles(ctx, g.outDir, newFilesSlice, previousFiles, g.subsystem.FileTracker.UntrackedPatterns, g.ignore, onRemovedFile); err != nil {
			return fmt.Errorf("failed to clean orphaned files: %w", err)
		}

		// The mock server subtree is fully generator-owned: customers who need to
		// hand-edit it should declare those files in .genignore. Walk the subtree
		// and delete anything not in this run's tracked set so historic orphans —
		// files written by prior runs that fell out of gen.lock for any reason —
		// get removed on the next regeneration. Only safe to run when the mock
		// server actually re-templated this run; otherwise newFiles holds nothing
		// for it and the walk would delete the entire prior subtree.
		if g.mockServerRegenerated && g.mockServerSubDir != "" {
			if err := g.cleanMockServerOrphans(ctx, fileTrackerResult.NewFiles, onRemovedFile); err != nil {
				return fmt.Errorf("failed to clean mock server orphans: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		errs = []error{err}
	}
	return errs
}

// cleanMockServerOrphans walks the mock server subtree and removes any file
// not present in newFiles. The standard gen.lock-anchored cleanup cannot reach
// files that were written but never recorded in gen.lock; this disk-anchored
// pass removes those orphans on the next regeneration.
//
// .genignore and the file tracker's UntrackedPatterns are honored so customers
// can pin custom files under the mock server subtree if they have a reason to.
func (g *Generator) cleanMockServerOrphans(ctx context.Context, newFiles config.TrackedFiles, onRemovedFile func(string)) error {
	tracked := make(map[string]struct{}, newFiles.Len())
	for path := range newFiles.Keys() {
		tracked[filepath.ToSlash(path)] = struct{}{}
	}

	// testdata/ is populated by a direct cp.Copy from the user's test files
	// directory (see generateMockServer) and bypasses FileTracker entirely.
	// The walk would otherwise see every test fixture as an untracked orphan
	// and delete it.
	skipSubDirs := []string{mockServerTestDataSubDir}

	return filetracking.CleanOrphanedFilesByDirWalk(
		ctx,
		g.outDir,
		g.mockServerSubDir,
		tracked,
		skipSubDirs,
		g.subsystem.FileTracker.UntrackedPatterns,
		g.ignore,
		onRemovedFile,
	)
}

// extractLegacyGeneratedFiles extracts file paths from the old generatedFiles format
// stored in AdditionalProperties. This ensures backward compatibility with lockfiles
// that haven't been migrated to the new TrackedFiles format.
func extractLegacyGeneratedFiles(additionalProperties map[string]any) []string {
	if additionalProperties == nil {
		return nil
	}

	generatedFiles, ok := additionalProperties["generatedFiles"]
	if !ok {
		return nil
	}

	// Type is []any because YAML/JSON unmarshalers don't preserve element types
	// when deserializing into map[string]any
	fileList, ok := generatedFiles.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(fileList))
	for _, f := range fileList {
		if filePath, ok := f.(string); ok {
			result = append(result, filePath)
		}
	}

	return result
}

// updateLockFile updates the lock file with new generation metadata.
// Returns any conflicts error that should be returned after lockfile is saved.
// manualVersioningMode reports whether versioningStrategy: manual is set.
// In this mode the lock file is kept merge-conflict free: doc/config
// checksums are omitted and the examples map is sorted.
func (g *Generator) manualVersioningMode() bool {
	return g.subsystem != nil && g.subsystem.Config != nil &&
		g.subsystem.Config.Generation.VersioningStrategy == config.VersioningStrategyManual
}

// applyManagementInfo records generation metadata in the lock file. Under
// versioningStrategy: manual the doc/config checksums have no consumer
// (version bumping is skipped entirely), so they are omitted — this keeps the
// management block stable across regenerations and free of merge conflicts
// between parallel branches.
func (g *Generator) applyManagementInfo(configChecksum, docChecksum, sdkVersion string) {
	if g.manualVersioningMode() {
		configChecksum = ""
		docChecksum = ""
	}
	g.lockFile.Management.ConfigChecksum = configChecksum
	g.lockFile.Management.DocChecksum = docChecksum
	g.lockFile.Management.DocVersion = g.docVersion
	g.lockFile.Management.GenerationVersion = g.genVersion
	g.lockFile.Management.ReleaseVersion = sdkVersion
	g.lockFile.Management.SpeakeasyVersion = g.cliVersion
}

func (g *Generator) updateLockFile(ctx context.Context, configChecksum, docChecksum, sdkVersion string, conflictsErr *merge.ConflictsError) []error {
	var errs []error
	err := g.withStep(ctx, ProgressStepLockFile, func(ctx context.Context) error {
		usedFeatures := g.subsystem.Features.GetUsedFeatures(ctx)
		g.lockFile.Features[g.target.Target] = usedFeatures
		g.applyManagementInfo(configChecksum, docChecksum, sdkVersion)

		// Only update TrackedFiles if the merge was not a no-op.
		// On no-op, the lockfile already contains the complete TrackedFiles from the previous run
		// with all their id and pristine_git_object values. Overwriting with fileTrackerResult.NewFiles
		// would lose that data since FileTracker doesn't populate those fields on incremental runs.
		if !g.mergeWasNoOp {
			fileTrackerResult := g.subsystem.FileTracker.GetResult()
			g.lockFile.TrackedFiles = sequencedmap.From(fileTrackerResult.NewFiles.AllOrdered(sequencedmap.OrderKeyAsc))
		}

		// Sorting examples by operation key spreads concurrent additions across
		// the map instead of clustering them at the tail (spec order), avoiding
		// adjacent-line merge conflicts. Gated to manual versioning mode so
		// other configurations don't get a one-time reordering diff.
		if g.manualVersioningMode() && g.lockFile.Examples != nil {
			g.lockFile.Examples = sequencedmap.From(g.lockFile.Examples.AllOrdered(sequencedmap.OrderKeyAsc))
		}

		g.lockFile.ReleaseNotes = g.releaseNotes

		// Same as the comment above about sdk-gen-config.
		if !mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
			configOpts := []config.Option{
				config.WithFileSystem(g),
			}
			if err := config.SaveLockFile(g.outDir, g.lockFile, configOpts...); err != nil {
				return err
			}
		}
		g.subsystem.FileTracker.Close(ctx)
		return nil
	})
	if err != nil {
		return []error{err}
	}
	// Return conflicts error after lockfile is saved (so pristine hashes are persisted)
	if conflictsErr != nil {
		return []error{conflictsErr}
	}
	return errs
}

// saveTrackedFilesToLockFile writes tracked files to gen.lock without finalizing
// all metadata. This lets us persist files early (before compile) so failures
// still record them for cleanup. May be called multiple times:
//   - Before compile: raw (pre-format) checksums
//   - After compile: post-format checksums
//
// The final updateLockFile call overwrites TrackedFiles with complete state.
//
// IMPORTANT: This function preserves metadata from the previous run's TrackedFiles
// (ID, PristineGitObject) so that the 3-way merge in handlePatches can correctly
// identify the base content for merging user modifications.
func (g *Generator) saveTrackedFilesToLockFile(ctx context.Context) error {
	if mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
		return nil
	}

	fileTrackerResult := g.subsystem.FileTracker.GetResult()
	oldTrackedFiles := g.lockFile.TrackedFiles

	// Copy persistent metadata (ID, PristineGitObject, etc.) from previous run to new files
	for path := range fileTrackerResult.NewFiles.Keys() {
		newTracked, _ := fileTrackerResult.NewFiles.Get(path)
		if oldTracked, ok := oldTrackedFiles.Get(path); ok {
			newTracked.ID = oldTracked.ID
			newTracked.PristineGitObject = oldTracked.PristineGitObject
			newTracked.Deleted = oldTracked.Deleted
			newTracked.AdditionalProperties = oldTracked.AdditionalProperties
			newTracked.MovedTo = oldTracked.MovedTo
			fileTrackerResult.NewFiles.Set(path, newTracked)
		}
	}

	// Preserve entries for files that were moved or deleted by the user
	for oldPath := range oldTrackedFiles.Keys() {
		if _, existsInNew := fileTrackerResult.NewFiles.Get(oldPath); !existsInNew {
			oldTracked, _ := oldTrackedFiles.Get(oldPath)
			if oldTracked.MovedTo != "" || oldTracked.Deleted {
				fileTrackerResult.NewFiles.Set(oldPath, oldTracked)
			}
		}
	}

	g.lockFile.TrackedFiles = sequencedmap.From(fileTrackerResult.NewFiles.AllOrdered(sequencedmap.OrderKeyAsc))

	configOpts := []config.Option{
		config.WithFileSystem(g),
	}
	return config.SaveLockFile(g.outDir, g.lockFile, configOpts...)
}

// compileAll runs all compilation and linting steps if compileOutput is enabled.
func (g *Generator) compileAll(ctx context.Context, mockServerGenerator *Generator) []error {
	// * * Compile SDK * * //
	if err := g.withStep(ctx, ProgressStepCompileSDK, func(ctx context.Context) error {
		return g.compileTarget(ctx)
	}); err != nil {
		return []error{err}
	}

	// * * Compile Testing * * //
	if g.targetConfig.Testing != nil && g.targetConfig.Testing.Compile != nil {
		if err := g.withStep(ctx, ProgressStepCompileTests, func(ctx context.Context) error {
			return g.compileTargetTesting(ctx)
		}); err != nil {
			return []error{err}
		}
	}

	// * * Lint SDK * * //
	if g.targetConfig.Lint != nil {
		if err := g.withStep(ctx, ProgressStepLintSDK, func(ctx context.Context) error {
			return g.lintTarget(ctx)
		}); err != nil {
			return []error{err}
		}
	}

	// * * Compile Usage * * //
	if g.generateUsageSnippetArgs != nil && g.generateUsageSnippetArgs.OutputDir != "" {
		if err := g.withStep(ctx, ProgressStepCompileUsage, func(ctx context.Context) error {
			return g.compileTestProject(ctx)
		}); err != nil {
			return []error{err}
		}
	}

	if mockServerGenerator != nil {
		// * * Compile Mock Server * * //
		if err := g.withStep(ctx, ProgressStepCompileMockServer, func(ctx context.Context) error {
			return mockServerGenerator.compileTarget(ctx)
		}); err != nil {
			return []error{err}
		}

		// * * Lint Mock Server * * //
		if err := g.withStep(ctx, ProgressStepLintMockServer, func(ctx context.Context) error {
			return mockServerGenerator.lintTarget(ctx)
		}); err != nil {
			return []error{err}
		}
	}

	return nil
}

// Validate and ValidateWithOpts validates the schema. It takes the content of the schema file and the
// path to that file. It performs validation for the languages found in the
// default configuration. ctx must contain explicit generation access state from
// generation-context/access; use WithDirect or WithAuthenticated.
//
// It returns a slice of validation errors that are found in the schema.
// Validate method was kept to support backward compatibility with snapshot testing.
func (g *Generator) Validate(ctx context.Context, schema []byte, schemaPath string, isRemote bool, workingDir string) (*validation.Result, error) {
	opts := ValidateOpts{
		Schema:     schema,
		SchemaPath: schemaPath,
		IsRemote:   isRemote,
		WorkingDir: workingDir,
		Target:     "",
	}

	return g.ValidateWithOpts(ctx, opts)
}

func (g *Generator) ValidateWithOpts(ctx context.Context, opts ValidateOpts) (*validation.Result, error) {
	ctx, err := resolveGenerationAccess(ctx, "", false)
	if err != nil {
		return nil, err
	}

	if opts.Target == "" {
		opts.Target = "go"
	}
	g.warningLogger = logging.NewWarningLogger(g.validationOnly, func(err error) {
		var vErr *errors.ValidationError
		if errors.As(err, &vErr) && g.result != nil {
			g.result.Insert(vErr.GetResult())
		}
	})
	if g.warningLoggerDisabled {
		g.warningLogger.DisableLogging()
	}

	ctx = logging.With(ctx, g.log)
	ctx = logging.WithWarningLogger(ctx, g.warningLogger)
	if err := g.Init(ctx, opts.Target, ""); err != nil {
		return nil, err
	}
	g.validationOnly = true

	ad := &analytics.Data{}

	docInfo := &document.DocumentInfo{
		Schema:     opts.Schema,
		SchemaPath: opts.SchemaPath,
		IsRemote:   opts.IsRemote,
	}

	res, err := g.LoadAndValidateDoc(ctx, docInfo, ad, opts.WorkingDir, true)
	if err != nil {
		return nil, err
	}
	g.result = res.Result
	if res.Result.HasFatalErrors() || !res.RunGenerator {
		return res.Result, nil
	}

	docChecksum := getSchemaChecksum(opts.Schema)

	configOpts := []config.Option{
		config.WithFileSystem(g),
	}

	configChecksum, err := config.GetConfigChecksum(g.outDir, configOpts...)
	if err != nil && !g.validationOnly {
		logging.LogWarning(ctx, "failed to get config checksum", err)
	}

	if _, err := g.startGenerate(ctx, "validate", docInfo, g.getVersioningInfo(docInfo.Doc, docChecksum, configChecksum), ad); err != nil {
		res.Result.AddError(err)
	}

	return res.Result, nil
}

// compileTarget will compile the target based on the target compile
// configuration. The working directory is the joined g.outDir and g.outSubDir.
func (g *Generator) compileTarget(ctx context.Context) error {
	if g.targetConfig.Compile == nil || g.targetConfig.Compile.Runner == nil {
		return nil
	}

	var outText string
	var err error

	ctx, span := g.tracer.Start(ctx, "Generator.compileTarget")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	g.applyCompileEnv(g.targetConfig.Compile.Runner)

	argVars := map[string]string{
		"PackageVersion": g.subsystem.Config.Languages[g.target.Target].Version,
		"Usage":          "false",
	}
	workingDir := filepath.Join(g.outDir, g.outSubDir)

	outText, err = g.targetConfig.Compile.Runner.Run(ctx, workingDir, argVars)
	if err != nil {
		if isCancelled(ctx) {
			return err
		}

		if !errors.Is(err, processrunner.ErrDependencyNotFound) && !errors.Is(err, processrunner.ErrDependencyVersion) {
			logging.From(ctx).Warn(outText)
			return fmt.Errorf("compilation failed: %w\n%s", err, outText)
		}

		// If we are running locally or in the action we want to fail if we don't have the right dependencies for compiling
		failOnCompilation := env.IsDebug() || g.runLocation == "action"
		if failOnCompilation {
			logging.From(ctx).Warn("can't compile - " + outText)
			return fmt.Errorf("can't compile: %w\n%s", err, outText)
		}

		logging.From(ctx).Warn("skipping compilation - " + outText)
		logging.From(ctx).Debug(err.Error())

	}
	return nil
}

// compileTargetTesting will compile the target testing based on the testing
// configuration. The working directory is the joined g.outDir and g.outSubDir.
func (g *Generator) compileTargetTesting(ctx context.Context) error {
	if g.targetConfig.Testing == nil || g.targetConfig.Testing.Compile == nil || g.targetConfig.Testing.Compile.Runner == nil {
		return nil
	}

	var outText string
	var err error

	ctx, span := g.tracer.Start(ctx, "Generator.compileTargetTesting")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	g.applyCompileEnv(g.targetConfig.Testing.Compile.Runner)

	argVars := map[string]string{
		"PackageVersion": g.subsystem.Config.Languages[g.target.Target].Version,
		"Usage":          "false",
	}
	workingDir := filepath.Join(g.outDir, g.outSubDir)

	outText, err = g.targetConfig.Testing.Compile.Runner.Run(ctx, workingDir, argVars)
	if err != nil {
		if isCancelled(ctx) {
			return err
		}

		if !errors.Is(err, processrunner.ErrDependencyNotFound) && !errors.Is(err, processrunner.ErrDependencyVersion) {
			logging.From(ctx).Warn(outText)
			return fmt.Errorf("test compilation failed: %w\n%s", err, outText)
		}

		// If we are running locally or in the action we want to fail if we don't have the right dependencies for compiling
		failOnCompilation := env.IsDebug() || g.runLocation == "action"
		if failOnCompilation {
			logging.From(ctx).Warn("can't compile testing - " + outText)
			return fmt.Errorf("can't compile testing: %w\n%s", err, outText)
		}

		logging.From(ctx).Warn("skipping test compilation - " + outText)
		logging.From(ctx).Debug(err.Error())
	}

	return nil
}

func (g *Generator) compileTestProject(ctx context.Context) error {
	if g.targetConfig.Compile == nil || g.targetConfig.Compile.Runner == nil {
		return nil
	}

	var err error
	var outText string

	ctx, span := g.tracer.Start(ctx, "Generator.compileTestProject")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	g.applyCompileEnv(g.targetConfig.Compile.Runner)

	argVars := map[string]string{
		"PackageVersion": g.subsystem.Config.Languages[g.target.Target].Version,
		"Usage":          "true",
	}
	workingDir := g.generateUsageSnippetArgs.OutputDir

	outText, err = g.targetConfig.Compile.Runner.Run(ctx, workingDir, argVars)
	if err != nil {
		logging.From(ctx).Error(fmt.Sprintf("failed to compile usage snippets in %s - %s", workingDir, outText))
		return fmt.Errorf("compilation of usage snippets failed: %w\n%s", err, outText)
	}

	return nil
}

// lintTarget will lint the target based on the target lint configuration. The
// working directory is the joined g.outDir and g.outSubDir.
func (g *Generator) lintTarget(ctx context.Context) error {
	if g.targetConfig.Lint == nil || g.targetConfig.Lint.Runner == nil {
		return nil
	}

	var outText string
	var err error

	ctx, span := g.tracer.Start(ctx, "Generator.lintTarget")
	defer func() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	g.applyCompileEnv(g.targetConfig.Lint.Runner)

	argVars := map[string]string{}
	workingDir := filepath.Join(g.outDir, g.outSubDir)

	outText, err = g.targetConfig.Lint.Runner.Run(ctx, workingDir, argVars)
	if err != nil {
		if isCancelled(ctx) {
			return err
		}

		if !errors.Is(err, processrunner.ErrDependencyNotFound) && !errors.Is(err, processrunner.ErrDependencyVersion) {
			logging.From(ctx).Warn(outText)
			return fmt.Errorf("lint failed: %w\n%s", err, outText)
		}

		// If we are running locally or in the action we want to fail if we don't have the right dependencies for lint
		failOnLint := env.IsDebug() || g.runLocation == "action"
		if failOnLint {
			logging.From(ctx).Warn("can't lint - " + outText)
			return fmt.Errorf("can't lint: %w\n%s", err, outText)
		}

		logging.From(ctx).Warn("skipping lint - " + outText)
		logging.From(ctx).Debug(err.Error())
	}

	return nil
}

// applyCompileEnv merges g.compileEnv into every command in the given runner.
// This is a no-op when compileEnv is empty.
func (g *Generator) applyCompileEnv(r *processrunner.Runner) {
	if len(g.compileEnv) == 0 {
		return
	}
	for i := range r.Commands {
		if r.Commands[i].EnvironmentVariables == nil {
			r.Commands[i].EnvironmentVariables = make(map[string]string, len(g.compileEnv))
		}
		for k, v := range g.compileEnv {
			r.Commands[i].EnvironmentVariables[k] = v
		}
	}
}

func (g *Generator) recordUsage(ctx context.Context, eventType string, errs []error, data *analytics.Data) {
	cfg := analytics.Config{
		PostHog:         g.posthog,
		CustomerID:      g.customerID,
		WorkspaceID:     g.workspaceID,
		RunLocation:     g.runLocation,
		GenVersion:      g.genVersion,
		CLIVersion:      g.cliVersion,
		Language:        g.target.Target,
		Template:        g.target.Template,
		FeatureTracking: map[string]bool{},
		ConfigTracking:  map[string]any{},
		GenIgnoreUsed:   g.ignore.HasIgnoredFiles(),
	}

	for feat := range g.subsystem.Features.GetUsedFeatures(ctx) {
		cfg.FeatureTracking[feat] = true
	}

	for k, v := range g.getConfigUsage() {
		cfg.ConfigTracking[k] = v
	}

	analytics.RecordUsage(ctx, eventType, errs, g.warningLogger.GetWarnings(), data, cfg)
}

func (g *Generator) shouldWrite(ctx context.Context, outFile string) bool {
	if g.ignore.ShouldIgnore(outFile) {
		logging.From(ctx).Warn("Generated file matches .genignore rule, skipping", zap.String("filename", outFile))
		return false
	}
	return true
}

func (g *Generator) Init(ctx context.Context, target, outDir string) error {
	return g.withStep(ctx, ProgressStepSetup, func(ctx context.Context) error {
		return g.initInternal(ctx, target, outDir)
	})
}

func (g *Generator) initInternal(ctx context.Context, target, outDir string) error {
	isTerraformTarget := target == "terraform"
	mcpTargetSupported := CheckMCPTargetNameSupported(target)
	sdkTargetSupported := CheckSDKTargetNameSupported(target)

	if !mode.IsSpeakeasyExecutionContextEmbedded(ctx) && (!isTerraformTarget && !mcpTargetSupported && !sdkTargetSupported) {
		if target == "" {
			return errors.New("language not specified")
		}
		return fmt.Errorf("language not supported: %s", target)
	}

	t, err := g.GetTarget(ctx, target, outDir)
	if err != nil {
		return err
	}

	g.outDir = outDir
	g.target = t
	g.comments.SetTags(target, "sdk") // TODO "terraform", "sdk" is not really appropriate, but...

	var cfg *config.Configuration

	switch {
	case g.outDir != "":
		c, err := g.LoadConfig(ctx, g.outDir, g.target.Target)
		if err != nil {
			return err
		}
		cfg = c.Config
		g.lockFile = c.LockFile
	case mode.IsSpeakeasyExecutionContextEmbedded(ctx):
		// In a WASM environment, we always write a gen.yaml file to the root of the fs
		// so we can read it here

		genYamlBytes, err := g.fs.ReadFile("gen.yaml")
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cfg, err = config.GetDefaultConfig(true, g.getLanguageConfigDefaults, map[string]bool{g.target.Target: true})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err = yaml.Unmarshal(genYamlBytes, &cfg); err != nil {
				return err
			}
		}

		g.lockFile = config.NewLockFile()
	default:
		cfg, err = config.GetDefaultConfig(true, g.getLanguageConfigDefaults, map[string]bool{g.target.Target: true})
		if err != nil {
			return err
		}
		g.lockFile = config.NewLockFile()
	}

	if _, ok := cfg.Languages[g.target.Target]; !ok {
		cfg.Languages[g.target.Target] = config.LanguageConfig{}
	}

	c := configuration.New(cfg, g.target.Target)
	logging.From(ctx).Debug("Resolving configuration for target" + g.target.Target)
	if err = c.Resolve(g.target, g.outDir, g); err != nil {
		logging.From(ctx).Warn(fmt.Sprintf("Failed to resolve config: %v", err))
	}
	logging.From(ctx).Debug("Configuration resolved")

	// Honor versioningStrategy: manual from config
	if c.Generation.VersioningStrategy == config.VersioningStrategyManual {
		g.skipVersioning = true
	}

	if err := g.subsystem.Init(ctx, subsystem.InitOptions{
		Target:              g.target,
		Config:              c,
		BaseTemplateConfigs: g.getBaseTemplateConfigs(g.target, c, g.lockFile),
		OutDir:              g.outDir,
		GenLockId:           g.lockFile.ID,
		FS:                  g,
		Git:                 g.git,
		LockFile:            g.lockFile,
	}); err != nil {
		return err
	}

	g.subsystem.Patches.OnProgress = g.sendPatchesProgressUpdate

	sanitizer, err := g.getSanitizer(ctx, g.target)
	if err != nil {
		return err
	}

	g.subsystem.Sanitizer = sanitizer
	g.bucketer = buckettypes.New(g.subsystem)
	if g.splitModelThreshold > 0 {
		g.bucketer.SetSplitThresholds(g.splitModelThreshold, g.splitModelChunkSize)
	}

	g.namer = namer.New(g.subsystem)

	gi, err := filetracking.NewIgnoreFromRoot(g.outDir)
	if err != nil {
		return err
	}
	g.ignore = gi

	overlay, err := g.getConfigOverlay()
	if err == nil {
		g.subsystem.Config.AddOverlay(overlay)
	}

	if names := g.getReservedModelFileNames(); len(names) > 0 {
		g.bucketer.SetReservedModelFileNames(names)
	}

	if err := g.getTargetInitialConfiguration(ctx); err != nil {
		return err
	}

	if _, exists := g.subsystem.Config.Languages[g.target.Target]; exists {
		if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureConfigurableModuleName) && g.subsystem.Config.Languages[g.target.Target].Cfg["moduleName"] != "" {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureConfigurableModuleName)
		}
	}

	if g.generateStandaloneUsage() {
		never := config.PersistentEditsEnabledNever
		g.subsystem.Config.Generation.PersistentEdits.Enabled = &never
		g.subsystem.Patches.Enabled = &never
	}
	g.enrichTelemetryEventPreGeneration(ctx)

	return nil
}

func (g *Generator) startGenerate(ctx context.Context, runName string, docInfo *document.DocumentInfo, versionInfo versioning.VersionInfo, ad *analytics.Data) (results *generateResults, err error) {
	defer func() {
		if r := recover(); r != nil {
			if panicErr, ok := r.(error); ok {
				err = errors.NewPanicError(panicErr)
			} else {
				err = errors.NewPanicError(fmt.Errorf("%v", r))
			}
		}
	}()

	start := time.Now()
	results, err = g.generate(ctx, docInfo, versionInfo, ad)
	var errs []error
	if err != nil {
		errs = []error{err}
	}

	duration := time.Since(start)
	logging.From(ctx).Debug(fmt.Sprintf("g.generate completed in %s", duration))

	g.recordUsage(ctx, runName, errs, ad)

	return results, err
}

func getSchemaChecksum(schema []byte) string {
	hash := md5.Sum(schema)
	return hex.EncodeToString(hash[:])
}

func (g *Generator) getVersioningInfo(doc *openapi.OpenAPI, docChecksum, configChecksum string) versioning.VersionInfo {
	return versioning.VersionInfo{
		SDKVersion:              g.subsystem.Config.Languages[g.target.Target].Version,
		PreviousSDKVersion:      g.lockFile.Management.ReleaseVersion,
		DocVersion:              doc.Info.Version,
		PreviousDocVersion:      g.lockFile.Management.DocVersion,
		DocChecksum:             docChecksum,
		PreviousDocChecksum:     g.lockFile.Management.DocChecksum,
		PreviousFeatureVersions: g.lockFile.Features[g.target.Target],
		ConfigChecksum:          configChecksum,
		PreviousConfigChecksum:  g.lockFile.Management.ConfigChecksum,
	}
}

func (g *Generator) GetTarget(ctx context.Context, target, outDir string) (types.Target, error) {
	// If we are in the WASM environment, we trust that the target / template is valid as we control it the inputs.
	// We want to avoid calling config.GetTemplateVersion as it will return an error due to it making
	// an unsupported filesystem call internally (GetWorkspace calls os.Getwd)
	if mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
		return types.NewTargetFromTemplate(target), nil
	}

	// Allow targeting a specific template version for debugging or WASM environment
	if os.Getenv("SPEAKEASY_DEBUG") == "true" || mode.IsSpeakeasyExecutionContextValidation(ctx) {
		if _, isSunset := templatesConfig.GetMinimumTargetVersion()[target]; !isSunset {
			return types.NewTargetFromTemplate(target), nil
		}
	}

	templateVersion, err := config.GetTemplateVersion(outDir, target, g.getConfigOptions(ctx, false, target)...)
	if err != nil {
		return types.Target{}, err
	}

	if minimumVersion, isSunset := templatesConfig.GetMinimumTargetVersion()[target]; isSunset {
		if v, err := version.NewVersion(templateVersion); err == nil {
			if v.LessThan(version.Must(version.NewVersion(minimumVersion))) {
				logging.LogWarning(ctx, "Invalid templateVersion", errors.NewUnsupportedError(fmt.Sprintf("target '%s' with templateVersion '%s' is sunset. Target will attempt to be upgraded to '%s'", target, templateVersion, minimumVersion), nil))
				templateVersion = minimumVersion
			}
		}
	}

	g.enrichTemplateVersion(ctx, templateVersion)
	switch templateVersion {
	case "v1":
		templateVersion = ""
	case "":
		return GetTargetFromTargetString(target)
	}

	template := fmt.Sprintf("%s%s", target, templateVersion)

	if _, err := templates.TemplateFS.ReadDir(path.Join("templates", template)); err != nil {
		template = target
	}

	return types.Target{
		Target:   target,
		Template: template,
	}, nil
}

func (g *Generator) enrichTelemetryEventPreGeneration(ctx context.Context) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}
	rawCfg, err := yaml.Marshal(g.subsystem.Config.Configuration)
	if err == nil {
		event.GenerateConfigPreRaw = new(string)
		*event.GenerateConfigPreRaw = string(rawCfg)
	}

	event.GeneratePublished = new(bool)
	*event.GeneratePublished = g.published

	event.GenerateRepoURL = new(string)
	*event.GenerateRepoURL = g.repoURL

	event.GenerateTarget = new(string)
	*event.GenerateTarget = g.target.Target

	event.GenerateOutputTests = new(bool)
	*event.GenerateOutputTests = g.outputTests

	event.GenerateVersion = new(string)
	*event.GenerateVersion = g.genVersion

	event.GenerateGenLockID = new(string)
	*event.GenerateGenLockID = g.lockFile.ID

	event.GenerateConfigPreVersion = new(string)
	*event.GenerateConfigPreVersion = g.subsystem.Config.Languages[g.target.Target].Version
}

func (g *Generator) enrichTelemetryEventPostGeneration(ctx context.Context) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}

	rawCfg, err := yaml.Marshal(g.subsystem.Config.Configuration)
	if err == nil {
		event.GenerateConfigPostRaw = new(string)
		*event.GenerateConfigPostRaw = string(rawCfg)
	}

	event.GenerateConfigPostChecksum = new(string)
	*event.GenerateConfigPostChecksum = g.lockFile.Management.ConfigChecksum

	event.GenerateConfigPostVersion = new(string)
	*event.GenerateConfigPostVersion = g.subsystem.Config.Languages[g.target.Target].Version

	event.GenerateVersion = new(string)
	*event.GenerateVersion = g.lockFile.Management.ReleaseVersion
}

func (g *Generator) enrichTemplateVersion(ctx context.Context, version string) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}

	event.GenerateTargetVersion = new(string)
	*event.GenerateTargetVersion = version
}

func (g *Generator) GetWarnings() []error {
	return g.warningLogger.GetWarnings()
}

// GetSubsystem returns the generator's subsystem
func (g *Generator) GetSubsystem() *subsystem.Subsystem {
	return g.subsystem
}

// GetRenderedUsageSnippets returns pre-rendered standalone usage snippets if
// WithRenderUsageSnippets was enabled and generation succeeded. Returns nil otherwise.
func (g *Generator) GetRenderedUsageSnippets() *RenderedUsageSnippets {
	return g.renderedUsageSnippets
}
