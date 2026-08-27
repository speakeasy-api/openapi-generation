package generate

import (
	"context"
	goerrors "errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/changelogs"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast_post_processing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/format"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	internalOpenAPI "github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"

	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/tests"
	testsArazzo "github.com/speakeasy-api/openapi-generation/v2/internal/tests/arazzo"
	"github.com/speakeasy-api/openapi-generation/v2/internal/versioning"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	versioning2 "github.com/speakeasy-api/versioning-reports/versioning"
)

type generateResults struct {
	sdkVersion         string
	generateMockServer bool
}

func (g *Generator) generate(ctx context.Context, docInfo *document.DocumentInfo, versionInfo versioning.VersionInfo, ad *analytics.Data) (*generateResults, error) {
	ast, err := g.resolveAST(ctx, docInfo, versionInfo, ad)
	if err != nil {
		return nil, err
	}

	if g.validationOnly {
		return nil, nil
	}

	enrichTelemetryEventPostAST(ctx, ast.MainSDK)

	// Record usage of dev containers
	if g.subsystem.Config.Generation.DevContainers != nil && g.subsystem.Config.Generation.DevContainers.Enabled {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureDevContainers)
	}

	shouldGenerateMockServer := false

	if g.shouldGenerateMockServer(ctx, ast) {
		shouldGenerateMockServer = true
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMockServer)
	}

	if enableFormatting, ok := g.subsystem.Config.Languages[g.target.Target].Cfg["enableFormatting"].(bool); ok {
		format.SetFormattingEnabled(g.target.Target, enableFormatting)
	}

	enableCustomCodeRegions := g.subsystem.Config.Languages[g.target.Target].Cfg["enableCustomCodeRegions"]
	if flag, ok := enableCustomCodeRegions.(bool); ok && flag {
		if err := licensing.ValidateAccountHasFeatureAccess(ctx, features.FeatureCustomCodeRegions); err != nil {
			return nil, err
		}

		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureCustomCodeRegions)
	}

	if enableFormatting, ok := g.subsystem.Config.Languages[g.target.Target].Cfg["enableFormatting"].(bool); ok {
		format.SetFormattingEnabled(g.target.Target, enableFormatting)
	}

	// All features should be marked as used by now, so we can get the list of used features
	versionInfo.FeatureVersions = g.subsystem.Features.GetUsedFeatures(ctx)
	versionInfo.AvailableFeatures = g.subsystem.Features.GetAvailableFeatures(ctx)
	forceGeneration := g.forceGeneration || versioning2.MustGenerate(ctx)

	// Calculate the new sdk version
	sdkVersion, bumpType, err := versioning.GetNewSDKVersion(ctx, versionInfo, forceGeneration, g.target.Target, g.skipVersioning)
	if err != nil {
		logging.LogWarning(ctx, "failed to get new sdk version", err)
	}

	// Calculate the changelog
	changelog, err := changelogs.GetTemplateChangeLog(g.target.Template, versionInfo.FeatureVersions, versionInfo.PreviousFeatureVersions)
	if err != nil {
		logging.LogWarning(ctx, "failed to get changelog", err)
	}

	// Check if we really need to generate this
	mustGenerate, err := changelogs.MustGenerate(g.target.Template, versionInfo.FeatureVersions, versionInfo.PreviousFeatureVersions)
	if err != nil {
		logging.LogWarning(ctx, "failed to check [force-gen] flag", err)
	}

	if len(changelog) > 0 {
		changelog = fmt.Sprintf("## %s CHANGELOG\n%s\n", strings.ToUpper(g.target.Target), changelog)
	} else {
		changelog = fmt.Sprintf("## %s CHANGELOG\nNo relevant generator changes\n", strings.ToUpper(g.target.Target))
	}

	// We do not want to write a version report for x-codeSamples it will interfere with the generation report
	if g.generateUsageSnippetArgs == nil {
		// Only produce a version report if there was a prior generation.
		// NB: usage snippet generation for x-codeSamples always generates with a "fresh" gen.lock so this also skips that.
		versionReport := versioning2.VersionReport{
			Key:          "GENERATE_" + g.target.Target,
			Priority:     1,
			MustGenerate: mustGenerate,
			BumpType:     bumpType,
			NewVersion:   sdkVersion,
		}
		if len(g.lockFile.Features) > 0 {
			versionReport.PRReport = changelog
		}
		err = versioning2.AddVersionReport(ctx, versionReport)
		if err != nil {
			logging.LogWarning(ctx, "failed to add version report", err)
		}

	}

	// Update the sdk version in the config
	langCfg := g.subsystem.Config.Languages[g.target.Target]
	langCfg.Version = sdkVersion
	g.subsystem.Config.Languages[g.target.Target] = langCfg

	if err := g.templateSDK(ctx, ast); err != nil {
		return nil, err
	}

	return &generateResults{
		sdkVersion:         sdkVersion,
		generateMockServer: shouldGenerateMockServer,
	}, nil
}

func (g *Generator) GenerateAST(ctx context.Context, docInfo *document.DocumentInfo, ad *analytics.Data) (*ast.AST, error) {
	start := time.Now()
	g.log.Debug("starting generateAST")
	ctx, span := g.tracer.Start(ctx, "Generator.generateAST")
	defer span.End()

	if err := g.subsystem.Extensions.HandleRewriteExtension(extensions.WithDocumentExtensions(docInfo.Doc.GetExtensions())); err != nil {
		return nil, err
	}

	g.schemas = &schemas.Schemas{
		Config:    g.subsystem.Config,
		Target:    g.target,
		Subsystem: g.subsystem,
		Namer:     g.namer,
	}

	g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureCore)

	if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureEnvVarSecurityUsage) {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureEnvVarSecurityUsage)
	}

	// PopulateFromDocument is only needed for replicating old libopenapi behavior
	if !g.subsystem.Config.Generation.Fixes.SharedNestedComponentsJan2026 {
		// Reset the registry before populating to avoid stale data from previous generations
		internalOpenAPI.ResetGlobalNestedRefRegistry()
		internalOpenAPI.PopulateFromDocument(ctx, docInfo)
	}

	mainSDKTypeDef := ast.NewSDKTypeDef(g.subsystem.Config.Generation.SDKClassName, ast.ContextStack{})
	mainSDKTypeDef = g.subsystem.Register.RegisterType(ctx, mainSDKTypeDef, false)
	mainSDK := ast.NewMainSDK(mainSDKTypeDef)
	mainSDK.OutputTests = g.outputTests
	mainSDK.TestGroup = g.testGroup

	a := &ast.AST{
		MainSDK:         mainSDK,
		OpenAPIDocument: docInfo.Doc,
		Components:      g.subsystem.Register.Types,
		Webhooks:        sequencedmap.New[string, []ast.Operation](),
	}

	// The CLI command manifest and body-schema catalog are decoded for the
	// cli target only: no other target consumes them, and a manifest
	// authoring error must never break SDK generation from the same document.
	if g.target.Target == "cli" {
		cliCommands, cliCommandsErr := g.subsystem.Extensions.HandleCLICommandsExtension(ctx, docInfo)
		if cliCommandsErr != nil {
			return nil, cliCommandsErr
		}
		a.CLICommands = cliCommands
		if cliCommands != nil && g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureCLICommands) {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureCLICommands)
		}

		cliErrors, cliErrorsErr := g.subsystem.Extensions.HandleCLIErrorsExtension(ctx, docInfo)
		if cliErrorsErr != nil {
			return nil, cliErrorsErr
		}
		a.CLIErrors = cliErrors

		cliBodySchemas, cliBodySchemasErr := g.subsystem.Extensions.CollectCLIBodySchemas(ctx, docInfo)
		if cliBodySchemasErr != nil {
			return nil, cliBodySchemasErr
		}
		a.CLIBodySchemas = cliBodySchemas
	}

	_, err := g.handleDocumentComments(ctx, docInfo.Doc, a)
	if err != nil {
		return nil, err
	}

	globalSecurity, securitySchemes, err := g.handleGlobalSecurity(ctx, docInfo, a)
	if err != nil {
		return nil, err
	}

	globalServers, err := g.handleGlobalServers(ctx, docInfo.Doc.Servers, a)
	if err != nil {
		return nil, err
	}

	if globalServers != nil {
		ad.ServerURL = globalServers.GetDefaultURL(false)

		if ad.ServerURL == "" && len(globalServers.Servers) > 0 {
			ad.ServerURL = globalServers.Servers[0].URL
		}

		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalServerURLs)
	}

	var timeout *int64
	if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureOperationTimeout) {
		timeout, err = g.subsystem.Extensions.HandleGlobalTimeoutExtension(docInfo.Doc)
		if err != nil {
			return nil, err
		}

		if timeout != nil {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureOperationTimeout)
		}
	}

	defaultEnabledRetries := g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureDefaultEnabledRetries)
	if defaultEnabledRetries {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureDefaultEnabledRetries)
	}

	globalRetries, err := g.subsystem.Extensions.HandleGlobalRetryExtension(docInfo.Doc, defaultEnabledRetries)
	if err != nil {
		return nil, err
	}

	if globalRetries != nil && !inheritedRetriesFilteredByMethod(g.target.Target, g.retryMethodPolicy()) {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureRetries)
	}

	globalNameOverrides, err := g.subsystem.Extensions.HandleGlobalNameOverrideExtensions(docInfo.Doc)
	if err != nil {
		return nil, err
	}
	if len(globalNameOverrides) > 0 {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureNameOverrides)
	}

	globals, err := g.handleGlobals(ctx, docInfo.Doc, a, globalNameOverrides, docInfo)
	if err != nil {
		return nil, err
	}

	if globals != nil {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobals)
	}

	globalMaxMethodParams, err := g.subsystem.Extensions.GetMaxMethodParams(docInfo.Doc.GetExtensions())
	if err != nil {
		return nil, err
	}
	globalFlattenRequest, err := g.subsystem.Extensions.GetFlattenRequest(docInfo.Doc.GetExtensions())
	if err != nil {
		return nil, err
	}

	hpp := handlePathsParams{
		Paths:               docInfo.Doc.GetPaths(),
		SecuritySchemes:     securitySchemes,
		GlobalNameOverrides: globalNameOverrides,
		GlobalTimeout:       timeout,
		GlobalRetries:       globalRetries,
		GlobalSecurity:      globalSecurity,
		AnalyticsData:       ad,
		Globals:             globals,
		MaxMethodParams:     globalMaxMethodParams,
		FlattenRequest:      globalFlattenRequest,
		DocInfo:             docInfo,
	}

	operations, err := g.handlePaths(ctx, hpp)
	if err != nil {
		return nil, err
	}

	webhookOperations, err := g.handleWebhooks(ctx, docInfo, hpp)
	if err != nil {
		return nil, err
	}
	a.Webhooks = webhookOperations

	err = g.handleComponents(ctx, docInfo, a)
	if err != nil {
		return nil, err
	}

	subSDKMap := map[string]*ast.SDK{} // quick access for subSDK by id

	// Build OpenAPI 3.2 tag hierarchy maps for parent resolution and kind filtering
	tagPathMap := internalOpenAPI.BuildTagPathMap(docInfo.Doc.Tags)
	nonNavTags := internalOpenAPI.BuildNonNavTagSet(docInfo.Doc.Tags)

	// Track operations added from non-nav tags to avoid duplicates when an operation
	// has multiple non-nav tags (e.g., both "badge" and "audience" kind tags).
	nonNavAdded := map[string]bool{}

	for tag, oo := range operations.All() {
		if tag == "" {
			a.MainSDK.AddOperations(oo...)
			continue
		}

		// Skip non-nav tags (badge, audience, etc.) — these don't create SubSDKs
		if nonNavTags[tag] {
			var deduped []ast.Operation
			for _, op := range oo {
				if !nonNavAdded[op.ID] {
					nonNavAdded[op.ID] = true
					deduped = append(deduped, op)
				}
			}
			if len(deduped) > 0 {
				a.MainSDK.AddOperations(deduped...)
			}
			continue
		}

		var tagDef *openapi.Tag
		for _, t := range docInfo.Doc.Tags {
			if t.Name == tag {
				tagDef = t
				break
			}
		}

		comments, err := g.handleTagComments(ctx, tagDef)
		if err != nil {
			return nil, err
		}

		// Resolve tag parent hierarchy (OpenAPI 3.2) to dot-notation path
		resolvedTag := tag
		if path, ok := tagPathMap[tag]; ok {
			resolvedTag = path
		}

		if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureMultiLevelTagging) && (strings.Contains(resolvedTag, ".") && strings.Repeat(".", len(resolvedTag)) != resolvedTag) {
			// Traverse through multi-level dot path while creating subSDKs
			currentSDK := a.MainSDK
			namespaces := strings.Split(resolvedTag, ".")

			for i, namespace := range namespaces {
				group := strings.Join(namespaces[:i+1], ".")
				sdk, isNew := g.getOrCreateSubSDK(ctx, namespace, group, *currentSDK.ChildContextStack(), subSDKMap)

				if isNew {
					currentSDK.SubSDKs = append(currentSDK.SubSDKs, sdk)

					// Look up tag descriptions for parent SubSDKs so intermediate groups
					// (e.g., "federation" in "federation.entities") get meaningful descriptions
					// instead of "Operations for X".
					// TODO: This is currently gated on CLI only. Review whether enriching
					// intermediate SubSDK descriptions from OpenAPI tags would be beneficial
					// for all targets (Go SDK, TypeScript SDK, etc.) as well.
					if g.target.Target == "cli" && i < len(namespaces)-1 {
						for _, t := range docInfo.Doc.Tags {
							if t.Name == namespace {
								parentComments, err := g.handleTagComments(ctx, t)
								if err != nil {
									return nil, err
								}
								if parentComments != nil && !parentComments.IsEmpty() {
									sdk.Comments = parentComments
								}
								break
							}
						}
					}
				}

				currentSDK = sdk
			}

			currentSDK.AddOperations(oo...)
			// Only set comments if we have them, to avoid overwriting existing comments with empty
			// when the same SubSDK is referenced by multiple tags (e.g., "Search" and "Reports.Search")
			if !comments.IsEmpty() {
				currentSDK.Comments = comments
			}
		} else {
			sdk, isNew := g.getOrCreateSubSDK(ctx, resolvedTag, resolvedTag, *a.MainSDK.ChildContextStack(), subSDKMap)

			if isNew {
				a.MainSDK.SubSDKs = append(a.MainSDK.SubSDKs, sdk)
			}

			sdk.AddOperations(oo...)
			// Only set comments if we have them, to avoid overwriting existing comments with empty
			// when the same SubSDK is referenced by multiple tags
			if !comments.IsEmpty() {
				sdk.Comments = comments
			}
		}
	}

	if !a.MainSDK.HasOperations() {
		return nil, errors.NewValidationError("no valid methods or webhooks found", nil, nil)
	}

	if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureTagBasedOrdering) && len(docInfo.Doc.Tags) > 0 {
		if langCfg, configExists := g.subsystem.Config.Languages[g.target.Target]; configExists {
			if configuration.LanguageConfigHasMaintainTagBasedOrdering(langCfg) {
				a.MainSDK.SortSubSDKsByTags(docInfo.Doc.Tags)
			}
		}
	}

	startPostProcessing := time.Now()
	g.subsystem.Cycles.DetectCycles(g.subsystem.Register.AllTypes())
	ast_post_processing.MarkUsedInRequest(a)
	ast_post_processing.MarkUsedInResponse(a)
	ast_post_processing.MarkUsedInWebhook(a)
	ast_post_processing.MarkUsedInCallback(a)
	ast_post_processing.MarkUsedInSecurity(a)
	// Use merged config (base + overlay) so overlay-only settings like
	// forwardCompatibleEnumsByDefault are visible to post-processing.
	mergedLangCfg := g.subsystem.Config.GetLanguageConfig(g.target.Target).Cfg
	ast_post_processing.InferUnionDiscriminators(ctx, g.subsystem.Register.AllTypes(), docInfo.SchemaPath, mergedLangCfg)
	ast_post_processing.PreApplyUnionDiscriminators(ctx, a, mergedLangCfg)
	ast_post_processing.MarkUnionsOpen(ctx, slices.Collect(g.subsystem.Register.Types.Values()), mergedLangCfg)
	ast_post_processing.PropagateIncludeExtension(g.subsystem.Register.AllTypes())
	ast_post_processing.MarkOpenEnums(g.subsystem.Register.AllTypes(), mergedLangCfg)
	g.log.Debug(fmt.Sprintf("ast_post_processing completed in '%s'", time.Since(startPostProcessing)))

	elapsed := time.Since(start)
	g.log.Debug(fmt.Sprintf("generateAST completed in '%s'", elapsed))

	// Sanity check all types have been registered properly
	_ = a.Walk(func(n ast.Node, parents []ast.Node, a *ast.AST) error {
		t, ok := n.(*ast.TypeDef)
		if !ok {
			return nil
		}
		if !t.IsCustomType() || t.ContextStack == nil {
			return nil
		}

		name := t.Name
		if t.OriginalName != "" {
			name = t.OriginalName
		}
		t2, ok := g.subsystem.Register.Types.Get(t.GetRegistrationID())
		if !ok && g.warningLogger != nil {
			g.warningLogger.LogWarning(ctx, "type not registered", fmt.Errorf("%q not registered", name))
		} else if t2 != t && g.warningLogger != nil {
			g.warningLogger.LogWarning(ctx, "type not registered properly", fmt.Errorf("%q not registered properly", name))
		}

		return nil
	})

	a.UsedTypes = make(map[string]bool)
	_ = a.Walk(func(n ast.Node, parents []ast.Node, a *ast.AST) error {
		return n.Match(ast.Matchers{
			TypeDef: func(t *ast.TypeDef) error {
				a.UsedTypes[string(t.Type)] = true
				return nil
			},
		})
	})

	if !g.validationOnly {
		if err := g.resolveNames(ctx, a); err != nil {
			return nil, err
		}

		bucketedTypes, operationServers := g.bucketer.BucketTypes(ctx, a)
		a.BucketedTypes = bucketedTypes
		a.OperationServers = operationServers
		a.PublicExports = ast.BuildPublicExports(ctx, a)
		// Only the languages that render an export surface record the
		// feature, and implicit model-namespace exports only count once the
		// SDK opts in via the imports.paths.resources configuration.
		if a.PublicExports != nil && (g.target.Target == "python" || g.target.Target == "typescript") {
			importCfg, _ := g.subsystem.Config.GetLanguageConfigValue("imports").(configuration.ImportConfig)
			if a.PublicExports.HasExplicitExports() || importCfg.GetResourcesPath() != "" {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeaturePublicExports)
			}
		}
		if g.target.Target == "typescript" {
			if methodSignature, _ := g.subsystem.Config.GetLanguageConfigValue("methodSignature").(string); methodSignature == "params-object" {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMethodSignatures)
			}
		}

		if err := g.writeAST(ctx, a); err != nil {
			return nil, err
		}
	}

	// Make sure we activate any relevant features if certain extensions are used
	_ = a.Walk(func(node ast.Node, nodes []ast.Node, a *ast.AST) error {
		return node.Match(ast.Matchers{
			TypeDef: func(t *ast.TypeDef) error {
				hasJQ := (t.Extensions.TransformFromAPI != nil && t.Extensions.TransformFromAPI.Type == extensions.Jq) ||
					(t.Extensions.TransformToAPI != nil && t.Extensions.TransformToAPI.Type == extensions.Jq)
				if hasJQ && g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureTransformJQ) {
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureTransformJQ)
				}
				return nil
			},
		})
	})
	if err := g.generateTerraformAST(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to prepare Terraform generation: %w", err)
	}

	return a, nil
}

// isASTSerializationEnabled controls whether the AST is serialized to YAML in gen.lock.
// Currently disabled, but left as a quick way to output the AST for debugging by
// changing the return value.
func (g *Generator) isASTSerializationEnabled() bool {
	return false
}

//nolint:unused // TODO: Determine whether we want to support this generator functionality still.
func (g *Generator) loadAST(ctx context.Context, versionInfo versioning.VersionInfo) (*ast.AST, error) {
	if g.target.Target == "mockserver" || // TODO should we be storing and loading a version for the mock server as well?
		versionInfo.ConfigChecksum != versionInfo.PreviousConfigChecksum ||
		versionInfo.DocChecksum != versionInfo.PreviousDocChecksum ||
		g.lockFile.AdditionalProperties["astversion"] != ast.ASTVersion ||
		g.lockFile.AdditionalProperties["ast"] == "" ||
		!g.isASTSerializationEnabled() {
		return nil, nil
	}

	// TODO if we release this to the public we should compress and obfuscate the ast
	// // base 64 decode the ast
	// compressedData, err := base64.StdEncoding.DecodeString(g.lockFile.AdditionalProperties["ast"].(string))
	// if err != nil {
	// 	return nil, err
	// }

	// // gzip decode the ast
	// r, err := gzip.NewReader(bytes.NewReader(compressedData))
	// if err != nil {
	// 	return nil, err
	// }
	// defer r.Close()

	// data, err := io.ReadAll(r)
	// if err != nil {
	// 	return nil, err
	// }

	a := ast.NewAST()
	if err := yaml.Unmarshal([]byte(g.lockFile.AdditionalProperties["ast"].(string)), a); err != nil {
		return nil, err
	}
	a.Resolve()

	g.subsystem.Register.Types = a.Components
	g.subsystem.Features.ReinstateFeatureUsage(ctx, a.UsedFeatures)

	g.log.Debug("finished loading ast from gen.lock")

	return a, nil
}

func (g *Generator) writeAST(ctx context.Context, a *ast.AST) error {
	if g.target.Target != "mockserver" && g.isASTSerializationEnabled() {
		a.UsedFeatures = g.subsystem.Features.GetFeatureUsage(ctx)

		out, err := yaml.Marshal(a)
		if err != nil {
			return err
		}

		// TODO if we release this to the public we should compress and obfuscate the ast
		// // gzip compress the ast
		// var b bytes.Buffer
		// w := gzip.NewWriter(&b)
		// _, err = w.Write(out)
		// if err != nil {
		// 	return err
		// }
		// if err := w.Close(); err != nil {
		// 	return err
		// }

		// // base 64 encode the ast
		// g.lockFile.AST = base64.StdEncoding.EncodeToString(b.Bytes())

		if g.lockFile.AdditionalProperties == nil {
			g.lockFile.AdditionalProperties = map[string]any{}
		}
		g.lockFile.AdditionalProperties["ast"] = string(out)
		g.lockFile.AdditionalProperties["astversion"] = ast.ASTVersion

		g.log.Debug("finished writing ast to gen.lock")
	}

	return nil
}

func (g *Generator) resolveAST(ctx context.Context, docInfo *document.DocumentInfo, _ versioning.VersionInfo, ad *analytics.Data) (*ast.AST, error) {
	g.docVersion = docInfo.Doc.Info.Version

	// TODO: Disabled at least until target features additions are handled.
	// internal issue reference
	// a, err := g.loadAST(ctx, versionInfo)
	// if err != nil {
	// 	return nil, err
	// }

	a, err := g.GenerateAST(ctx, docInfo, ad)
	if err != nil {
		return nil, err
	}

	if !g.validationOnly {
		if err := g.writeAST(ctx, a); err != nil {
			return nil, err
		}
	}

	if !g.validationOnly {
		// TODO make sure the examples are only precalculated when the ast is generated
		if err := g.preCalculateExamples(ctx, a); err != nil {
			return nil, err
		}

		// Temporary check until the tests code is using the virtual filesystem correctly but they will require a larger refactor
		// reason being this code is writing files using os instead of the file system and so files get create during usage snippet generation
		if !g.generateStandaloneUsage() && !mode.IsSpeakeasyExecutionContextEmbedded(ctx) {
			if err := g.resolveTests(ctx, docInfo, a); err != nil {
				return nil, err
			}
		}
	}

	return a, nil
}

func (g *Generator) newFileLogger(filename string) *logging.ZapLogger {
	fullPath := filepath.Join(g.outDir, ".speakeasy", "logs", filename)
	return logging.NewFileLogger(zapcore.DebugLevel, fullPath)
}

func (g *Generator) resolveNames(ctx context.Context, ast *ast.AST) error {
	start := time.Now()
	g.log.Debug("starting resolveNames")
	ctx, span := g.tracer.Start(ctx, "Generator.resolveNames")
	defer span.End()

	r, err := namer.NewResolver(g.subsystem, g.log)
	if err != nil {
		return err
	}

	namingLogger := g.newFileLogger("naming.log")
	ctx = logging.With(ctx, namingLogger)

	importCfg := g.subsystem.Config.GetLanguageConfigValue("imports").(configuration.ImportConfig)
	r.ResolveNames(ctx, g.subsystem.Register, importCfg, g.subsystem.Config.MaintainOpenAPIOrder(), false)

	strBuilder := strings.Builder{}
	strBuilder.WriteString("\n")
	strBuilder.WriteString("\n")
	names := collectTypeNameInfo(ast)
	for _, name := range names {
		strBuilder.WriteString(name.NameWithDepth)
		strBuilder.WriteString("\n")
	}
	namingLogger.Debug(strBuilder.String())

	g.log.Debug(fmt.Sprintf("resolveNames completed in '%s'", time.Since(start)))

	return nil
}

func (g *Generator) skip(ctx context.Context, skippedEntity errors.SkippedEntity, err error) bool {
	if errors.IsUnsupported(err) {
		var unsupportedErr *errors.UnsupportedError
		if goerrors.As(err, &unsupportedErr) {
			skippedErr := errors.NewSkippedError(skippedEntity, unsupportedErr)
			logging.LogWarning(ctx, "skipping: "+skippedEntity.Name, skippedErr)
			return true
		}
	}

	return false
}

func enrichTelemetryEventPostAST(ctx context.Context, ast *ast.SDK) {
	event := generationtelemetry.EventFromContext(ctx)
	if event == nil {
		return
	}

	event.GenerateNumberOfOperationsUsed = new(int64)
	*event.GenerateNumberOfOperationsUsed = ast.CountUniqueOperations()
}

// getOrCreateSubSDK: Fetch an existing SubSDK if it exists or create one for multi-level tagging traversal
func (g *Generator) getOrCreateSubSDK(ctx context.Context, name, group string, contextStack ast.ContextStack, subSDKMap map[string]*ast.SDK) (*ast.SDK, bool) {
	nameResolutionFeb2025 := g.subsystem.Config.Generation.NameResolutionAtLeastShortest()

	sdkTypeDef := ast.NewSDKTypeDef(name, contextStack)
	sdkTypeDef = g.subsystem.Register.RegisterType(ctx, sdkTypeDef, false)
	newSDK := ast.NewSubSDK(sdkTypeDef, name, group)
	newSDKID := newSDK.ID(nameResolutionFeb2025)

	if sdk, ok := subSDKMap[newSDKID]; ok {
		if nameResolutionFeb2025 && sdkTypeDef != sdk.Type {
			// Prevent causing rename collisions - in the case the tags/sdk names differ by casing
			g.subsystem.Register.UnregisterType(sdkTypeDef)
		}
		return sdk, false
	}

	subSDKMap[newSDKID] = newSDK
	return newSDK, true
}

func (g *Generator) resolveTests(ctx context.Context, docInfo *document.DocumentInfo, a *ast.AST) error {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureTests) {
		return nil
	}

	ctx, span := g.tracer.Start(ctx, "Generator.resolveTests")
	defer span.End()

	start := time.Now()
	g.log.Debug("starting resolveTests")

	arazzoDocInfo, err := testsArazzo.ManageArazzoDoc(ctx, testsArazzo.ManageArazzoDocOptions{
		OutDir:             g.outDir,
		OpenAPIDocPath:     docInfo.SchemaPath,
		AST:                a,
		FileSystem:         g.fs,
		LockFile:           g.lockFile,
		IsInternalTestSpec: g.testGroup != "" && g.testGroup != "review",
		LoadOnly:           g.target.Target == "mockserver",
		Config:             g.subsystem.Config,
	})
	if err != nil {
		return err
	}

	testGroups := tests.PopulateTests(ctx, tests.PopulateTestsOptions{
		AST:                   a,
		DocInfo:               arazzoDocInfo,
		Subsystem:             g.subsystem,
		Target:                g.target,
		InternalTestGroupName: g.testGroup,
		OutDir:                g.outDir,
	})

	if len(testGroups) > 0 {
		a.Tests = &ast.Tests{
			TestGroups:          testGroups,
			GenerateExampleFile: tests.GenerateExampleFileIfNeeded(tests.GetTestFilesDir(g.outDir)),
		}
	}

	g.log.Debug(fmt.Sprintf("resolveTests completed in '%s'", time.Since(start)))

	return nil
}

// Deprecated: Use CheckTargetNameSupported or CheckSDKTargetNameSupported
// instead.
func CheckLanguageSupported(lang string) bool {
	return CheckSDKTargetNameSupported(lang) || lang == "terraform"
}

// Returns true if the target name is a supported MCP target name. Target names
// for MCP always include a "mcp-" prefix, such as "mcp-typescript".
// Version-suffixed template names return false, such as a theoretical
// "mcp-typescriptv2".
func CheckMCPTargetNameSupported(targetName string) bool {
	return templates.IsSupportedMCPTemplateName(targetName)
}

// Returns true if the target name is a supported SDK target name. Target names
// are equivalent to language names for SDKs and do not include any
// version-suffixed template names. For example, "typescript" will return true
// while "typescriptv2" will return false.
func CheckSDKTargetNameSupported(targetName string) bool {
	return templates.IsSupportedSDKTemplateName(targetName)
}

// Returns true if the target name is a supported target name, including SDK,
// MCP, and terraform. Target names for SDKs are equivalent to language names
// and MCP always includes a "mcp-" prefix. Version-suffixed template names are
// not included, such as "typescriptv2" or a theoretical "mcp-typescriptv2".
func CheckTargetNameSupported(targetName string) bool {
	return templates.IsSupportedTemplateName(targetName)
}

// Deprecated: Use GetSupportedTargetNames or GetSupportedSDKTargetNames
// instead.
func GetSupportedLanguages() []string {
	return slices.Concat(
		GetSupportedSDKTargetNames(),
		GetSupportedTerraformTargetNames(),
	)
}

// Returns all supported target names, including SDK, MCP, and terraform. Target
// names for SDKs are equivalent to language names and MCP always includes a
// "mcp-" prefix. Version-suffixed template names are not included, such as
// "typescripv2" or a theoretical "mcp-typescriptv2".
func GetSupportedTargetNames() []string {
	return templates.AllSupportedTemplateNames()
}

// Returns all supported MCP target names. Target names for MCP always include
// a "mcp-" prefix, such as "mcp-typescript". Version-suffixed template names
// are not included, such as a theoretical "mcp-typescriptv2".
func GetSupportedMCPTargetNames() []string {
	return templates.GetSupportedMCPTemplateNames()
}

// Returns all supported SDK target names. Target names for SDKs are equivalent
// language names and do not include any version-suffixed template names.
// For example, "typescript" will return true while "typescriptv2" will return
// false.
func GetSupportedSDKTargetNames() []string {
	return templates.GetSupportedSDKTemplateNames()
}

// Returns all supported Terraform target names. Version-suffixed template names
// are not included, such as a theoretical "terraformv2".
func GetSupportedTerraformTargetNames() []string {
	return templates.GetSupportedTerraformTemplateNames()
}

func GetTargetFromTargetString(target string) (types.Target, error) {
	return templates.GetTargetFromTargetString(target)
}

func GetLanguageConfigFields(target types.Target, newSDK bool) ([]config.SDKGenConfigField, error) {
	return templates.GetLanguageConfigFields(target, newSDK)
}

// Deprecated: Use GetSupportedSDKTargets instead.
func GetSupportedTargets() []types.Target {
	return slices.Concat(
		GetSupportedSDKTargets(),
		GetSupportedTerraformTargets(),
	)
}

// Returns all supported MCP targets.
func GetSupportedMCPTargets() []types.Target {
	return templates.GetSupportedMCPTargets()
}

// Returns all supported SDK targets.
func GetSupportedSDKTargets() []types.Target {
	return templates.GetSupportedSDKTargets()
}

// Returns all supported Terraform targets.
func GetSupportedTerraformTargets() []types.Target {
	return templates.GetSupportedTerraformTargets()
}

// Returns the target maturity level based on the target name. Returns an empty
// string if the target name is not found.
func GetTargetNameMaturity(targetName string) string {
	for _, target := range templates.AllSupportedTargets() {
		if targetName == target.Target {
			return string(target.Maturity)
		}
	}

	return ""
}
