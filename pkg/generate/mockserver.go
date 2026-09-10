package generate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/tests"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/versioning"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

// mockServerTestDataSubDir is the subdirectory under the mock server outSubDir
// where test fixture files are deposited via direct cp.Copy from the user's
// test files directory (see generateMockServer). It bypasses FileTracker, so
// any disk-anchored cleanup of the mock server subtree must skip it.
const mockServerTestDataSubDir = "testdata"

// generateMockServer performs a separate AST resolution and target generation
// solely for the testing mock server. This is to prevent issues caused by the
// original AST resolution, which sets the OutputLocation bucketing based on
// the original target imports configuration.
//
// This is intended to be ran on the cloned mock server Generator.
func (g *Generator) generateMockServer(ctx context.Context, docInfo *document.DocumentInfo, ad *analytics.Data) (regenerated bool, errs []error) {
	defer func() {
		if r := recover(); r != nil {
			if panicErr, ok := r.(error); ok {
				errs = []error{
					errors.NewPanicError(panicErr),
				}
			} else {
				errs = []error{
					errors.NewPanicError(fmt.Errorf("%v", r)),
				}
			}
		}
	}()
	ctx, span := g.tracer.Start(ctx, "Generator.generateMockServer")
	defer span.End()

	// Create a new doc to avoid any unintended document modification.
	validationResult, err := g.LoadAndValidateDoc(ctx, docInfo, ad, g.outDir, false)
	if err != nil {
		return false, []error{err}
	}

	g.result = validationResult.Result

	if validationResult.Result.HasFatalErrors() {
		return false, validationResult.Result.GetValidationErrors()
	}

	if !validationResult.RunGenerator {
		return false, nil
	}

	if err := g.setupExclusions(ctx); err != nil {
		return false, []error{err}
	}

	ast, err := g.resolveAST(ctx, &document.DocumentInfo{
		Doc:        docInfo.Doc,
		SchemaPath: docInfo.SchemaPath,
		Schema:     docInfo.Schema,
		IsRemote:   docInfo.IsRemote,
	}, versioning.VersionInfo{}, ad)
	if err != nil {
		return false, []error{err}
	}

	shouldRegenerate, err := g.shouldRegenerateMockServer(ctx)
	if err != nil {
		return false, []error{err}
	}

	if !shouldRegenerate {
		return false, nil
	}

	cfg := g.getBaseTemplateConfigs(g.target, g.subsystem.Config, g.lockFile)

	speakeasyTestFilesDir := tests.GetTestFilesDir(g.outDir)
	mockServerTestDataDir := filepath.Join(g.outDir, g.outSubDir, mockServerTestDataSubDir)

	if _, err := os.Stat(speakeasyTestFilesDir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, []error{err}
		}
	} else {
		// TODO this won't work if we ever actually use virtual filesystems and we will likely need our own copy function
		if err := cp.Copy(speakeasyTestFilesDir, mockServerTestDataDir); err != nil {
			return false, []error{err}
		}
	}

	if err := g.templateTarget(ctx, g.target, ast, cfg, nil, g.onWriteFile(g.outSubDir), g.onReadFile(g.outSubDir)); err != nil {
		return false, []error{err}
	}

	return true, nil
}

// getMockServerDirectory gets the mock server directory for the target based
// on the test directory. This requires the target to implement a mockserver.ts
// file with a getMockServerDirectory function.
//
// This is safe to run from any Generator.
func (g *Generator) getMockServerDirectory(ctx context.Context, originalTarget types.Target) (string, error) {
	engine := g.getExecutor(ctx, originalTarget, g.onWriteFile(""), g.onReadFile(""))

	if err := engine.Init(ctx, GlobalContext{}); err != nil {
		return "", err
	}

	if err := engine.RunScript(ctx, "mockserver.ts"); err != nil {
		return "", err
	}

	mockServerDirectoryRaw, err := engine.RunFunction(ctx, "getMockServerDirectory")
	if err != nil {
		return "", err
	}

	if mockServerDirectoryRaw == nil || mockServerDirectoryRaw.Export() == nil {
		return "", fmt.Errorf("missing mockserver directory for target: %s", originalTarget.Target)
	}

	return mockServerDirectoryRaw.Export().(string), nil
}

// getMockServerVersioningInfo returns the version information.
//
// This is intended to be ran on the cloned mock server Generator.
func (g *Generator) getMockServerVersioningInfo(ctx context.Context) versioning.MockServerVersionInfo {
	return versioning.MockServerVersionInfo{
		FeatureVersions:         g.subsystem.Features.GetUsedFeatures(ctx),
		PreviousFeatureVersions: g.lockFile.Features["mockserver"],
	}
}

// shouldGenerateMockServer returns true if the target has the feature enabled,
// the account has feature access, and the generation configuration enables mock
// server generation.
//
// This is intended to be ran on the original (not mock server) Generator.
func (g *Generator) shouldGenerateMockServer(ctx context.Context, a *ast.AST) bool {
	if env.DebugDisableMockServer() {
		return false
	}

	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureMockServer) {
		return false
	}

	if !licensing.AccountHasFeatureAccess(ctx, features.FeatureMockServer) {
		return false
	}

	// Respect explicit generation configuration.
	if g.subsystem.Config.Generation.MockServer != nil {
		return !g.subsystem.Config.Generation.MockServer.Disabled
	}

	mockServerNeeded := false
	if g.subsystem.Config.Generation.Tests.GenerateTests {
		if a.Tests != nil {
			for _, group := range a.Tests.TestGroups {
				for _, test := range group.Tests {
					if test.UsingMockServer {
						mockServerNeeded = true
						break
					}
				}
			}
		}
	}

	// Need to mark the mock server as needed for our test sdks (like primary, secondary etc) as they don't get automatically generated with mock server urls injected
	// as we add them manually in the arazzo doc so `UsingMockServer` doesn't get set.
	// This was a concession made early on to avoid the need to add x-speakeasy-test-server extensions to all the tests in these sdks.
	if !mockServerNeeded && g.testGroup != "" && g.testGroup != "review" {
		mockServerNeeded = true
	}

	// Only generate mock server if tests exist that need the mock server
	return mockServerNeeded
}

// shouldRegenerateMockServer returns true if the any of the following
// conditions are met:
//   - Generator forceGeneration option is enabled.
//   - Any mockserver target features have been bumped.
//
// This is intended to be ran on the cloned mock server Generator.
func (g *Generator) shouldRegenerateMockServer(ctx context.Context) (bool, error) {
	if g.forceGeneration {
		return true, nil
	}

	shouldRegenerate, err := versioning.ShouldMockServerRegenerate(ctx, g.getMockServerVersioningInfo(ctx))
	if err != nil {
		return false, err
	}

	return shouldRegenerate, nil
}
