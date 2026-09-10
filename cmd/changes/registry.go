package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensebootstrap"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/changes"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	speakeasy "github.com/speakeasy-api/speakeasy-client-sdk-go/v3"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/operations"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/retry"
	"github.com/spf13/cobra"
)

var (
	// Registry-specific flags
	org         string
	workspace   string
	oldRevision string
	newRevision string
	namespace   string
	apiKey      string
	openBrowser bool
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Compare OpenAPI spec changes between registry revisions",
	Long:  `Fetch OpenAPI specifications from the Speakeasy registry and compare changes between revisions.`,
	RunE:  runRegistryComparison,
}

func init() {
	registryCmd.Flags().StringVar(&org, "org", "", "Organization name (optional, defaults to current workspace)")
	registryCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace name (optional, defaults to current workspace)")
	registryCmd.Flags().StringVar(&oldRevision, "old-revision", "", "Old revision to compare (optional, defaults to recent event)")
	registryCmd.Flags().StringVar(&newRevision, "new-revision", "", "New revision to compare (optional, defaults to recent event)")
	registryCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace for both revisions (required when using --old-revision/--new-revision)")
	registryCmd.Flags().StringVar(&lang, "lang", "typescript", "Target language")
	registryCmd.Flags().StringVar(&outputDir, "output-dir", "/tmp/debug-changelog", "Output directory for saving specs and changelog")
	registryCmd.Flags().StringVar(&apiKey, "api-key", "", "Speakeasy API key (optional, can also be set via SPEAKEASY_API_KEY env var)")
	registryCmd.Flags().StringVar(&detailLevel, "detail-level", "compact", "Detail level for changelog output (compact or full)")
	registryCmd.Flags().BoolVar(&openBrowser, "open", false, "Open the HTML changelog in default browser")
}

func runRegistryComparison(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	logger := logging.NewLogger(logging.LevelFromEnv())
	logger.Debug("Starting registry comparison")

	setupStart := time.Now()
	ctx, client, err := setupClient()
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Setup client completed in %v", time.Since(setupStart)))

	workspaceStart := time.Now()
	if err := resolveWorkspace(ctx, client, logger); err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Resolve workspace completed in %v", time.Since(workspaceStart)))

	revisionStart := time.Now()
	if err := resolveRevisions(ctx, client, logger); err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Resolve revisions completed in %v", time.Since(revisionStart)))

	compareStart := time.Now()
	err = compareRevisions(ctx, client, logger)
	logger.Debug(fmt.Sprintf("Compare revisions completed in %v", time.Since(compareStart)))
	logger.Info(fmt.Sprintf("Total execution time: %v", time.Since(startTime)))

	return err
}

type registryContextKey struct{}

type registryContext struct {
	apiKey       string
	workspaceID  string
	httpClient   *http.Client
	preflightURL string
}

func setupClient() (context.Context, *speakeasy.Speakeasy, error) {
	return setupClientWithHTTPClient(http.DefaultClient)
}

func setupClientWithHTTPClient(httpClient *http.Client) (context.Context, *speakeasy.Speakeasy, error) {
	if httpClient == nil {
		return nil, nil, errors.New("registry HTTP client is required")
	}

	key, err := getAPIKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get API key: %w", err)
	}

	serverURL := os.Getenv("SPEAKEASY_SERVER_URL")
	clientOpts := registrySDKOptions(key, httpClient, serverURL)
	client := speakeasy.New(clientOpts...)

	validationClient := speakeasy.New(registrySDKOptions(key, httpClient, validationServerURL(serverURL))...)
	validationResp, err := validationClient.Auth.ValidateAPIKey(context.Background(), operations.WithRetries(retry.Config{
		Strategy: "backoff",
		Backoff: &retry.BackoffStrategy{
			InitialInterval: 1000,
			MaxInterval:     60000,
			Exponent:        1.5,
			MaxElapsedTime:  120000,
		},
		RetryConnectionErrors: true,
	}))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	if validationResp == nil || validationResp.StatusCode != http.StatusOK || validationResp.APIKeyDetails == nil || strings.TrimSpace(validationResp.APIKeyDetails.WorkspaceID) == "" {
		return nil, nil, errors.New("API key validation did not return a workspace identity")
	}

	workspaceID := strings.TrimSpace(validationResp.APIKeyDetails.WorkspaceID)
	ctx := context.WithValue(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), registryContextKey{}, registryContext{
		apiKey:       key,
		workspaceID:  workspaceID,
		httpClient:   httpClient,
		preflightURL: preflightServerURL(serverURL),
	})
	return ctx, client, nil
}

func registrySDKOptions(key string, httpClient *http.Client, serverURL string) []speakeasy.SDKOption {
	opts := []speakeasy.SDKOption{
		speakeasy.WithClient(httpClient),
		speakeasy.WithSecurity(shared.Security{APIKey: speakeasy.String(key)}),
	}
	if serverURL != "" {
		opts = append(opts, speakeasy.WithServerURL(serverURL))
	}
	return opts
}

func validationServerURL(serverURL string) string {
	if isAuthServerOverride(serverURL) {
		return serverURL
	}
	return "https://app.speakeasy.com"
}

func preflightServerURL(serverURL string) string {
	if isAuthServerOverride(serverURL) {
		return serverURL
	}
	return "https://app.speakeasy.com"
}

func isAuthServerOverride(serverURL string) bool {
	switch serverURL {
	case "http://localhost:35291", "https://app.speakeasy.com", "https://staging.speakeasy.com", "https://dev.speakeasy.com":
		return true
	default:
		return false
	}
}

func registryStateFromContext(ctx context.Context) (registryContext, error) {
	state, ok := ctx.Value(registryContextKey{}).(registryContext)
	if !ok || state.workspaceID == "" || state.apiKey == "" || state.httpClient == nil || state.preflightURL == "" {
		return registryContext{}, errors.New("registry authentication state is missing")
	}
	return state, nil
}

func resolveWorkspace(ctx context.Context, client *speakeasy.Speakeasy, logger logging.Logger) error {
	if org != "" && workspace != "" {
		return nil
	}

	logger.Debug("Fetching current workspace info")

	state, err := registryStateFromContext(ctx)
	if err != nil {
		return err
	}

	wsResp, err := getWorkspace(ctx, client, state.workspaceID)
	if err != nil {
		return err
	}

	workspace = wsResp.Workspace.Slug
	logger.Debug("Workspace: " + wsResp.Workspace.Name)

	orgResp, err := getOrganization(ctx, client, wsResp.Workspace.OrganizationID)
	if err != nil {
		return err
	}

	org = orgResp.Organization.Slug
	logger.Debug("Organization: " + orgResp.Organization.Name)
	logger.Info(fmt.Sprintf("Using workspace: %s/%s", org, workspace))

	return nil
}

func getWorkspace(ctx context.Context, client *speakeasy.Speakeasy, workspaceID string) (*operations.GetWorkspaceResponse, error) {
	wsReq := operations.GetWorkspaceRequest{
		WorkspaceID: &workspaceID,
	}

	wsResp, err := client.Workspaces.GetByID(ctx, wsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace details: %w", err)
	}

	if wsResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code getting workspace: %d", wsResp.StatusCode)
	}

	if wsResp.Workspace == nil {
		return nil, errors.New("unexpected missing workspace response")
	}

	return wsResp, nil
}

func getOrganization(ctx context.Context, client *speakeasy.Speakeasy, orgID string) (*operations.GetOrganizationResponse, error) {
	orgReq := operations.GetOrganizationRequest{
		OrganizationID: orgID,
	}

	orgResp, err := client.Organizations.Get(ctx, orgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization details: %w", err)
	}

	if orgResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code getting organization: %d", orgResp.StatusCode)
	}

	if orgResp.Organization == nil {
		return nil, errors.New("unexpected missing organization response")
	}

	return orgResp, nil
}

func resolveRevisions(ctx context.Context, client *speakeasy.Speakeasy, logger logging.Logger) error {
	// If user provides revisions, they must also provide namespace
	if oldRevision != "" || newRevision != "" {
		if oldRevision == "" || newRevision == "" {
			return errors.New("both --old-revision and --new-revision must be provided together")
		}
		if namespace == "" {
			return errors.New("--namespace is required when using --old-revision and --new-revision")
		}
		logger.Debug(fmt.Sprintf("Using user-provided revisions: old=%s, new=%s, namespace=%s", oldRevision, newRevision, namespace))
		return nil
	}

	// Auto-detect revisions from events
	logger.Debug("Searching for recent GENERATE events")

	revs, err := findRevisionsFromEvents(ctx, client, org, workspace)
	if err != nil {
		return fmt.Errorf("failed to find revisions from events: %w", err)
	}

	oldRevision = revs.oldRevision
	newRevision = revs.newRevision
	namespace = revs.namespace

	logger.Debug(fmt.Sprintf("Using revisions: old=%s, new=%s, namespace=%s", oldRevision, newRevision, namespace))
	return nil
}

func compareRevisions(ctx context.Context, client *speakeasy.Speakeasy, logger logging.Logger) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	logger.Debug("Downloading old spec: revision=" + oldRevision)
	oldSpecPath, err := downloadSpec(ctx, client, oldRevision, filepath.Join(outputDir, "old"), logger)
	if err != nil {
		return fmt.Errorf("failed to download old spec: %w", err)
	}
	logger.Debug("Old spec downloaded to: " + oldSpecPath)

	logger.Debug("Downloading new spec: revision=" + newRevision)
	newSpecPath, err := downloadSpec(ctx, client, newRevision, filepath.Join(outputDir, "new"), logger)
	if err != nil {
		return fmt.Errorf("failed to download new spec: %w", err)
	}
	logger.Debug("New spec downloaded to: " + newSpecPath)

	return generateDiff(ctx, oldSpecPath, newSpecPath, logger)
}

func downloadSpec(ctx context.Context, client *speakeasy.Speakeasy, revision, outputPath string, logger logging.Logger) (string, error) {
	// Use the single namespace for both revisions
	return downloadAndExtractSpec(ctx, client, org, workspace, namespace, revision, outputPath, logger)
}

func generateDiff(ctx context.Context, oldSpecPath, newSpecPath string, logger logging.Logger) error {
	startTime := time.Now()

	if err := validateSpecFiles(oldSpecPath, newSpecPath, logger); err != nil {
		return err
	}

	if err := copySpecsToOutput(oldSpecPath, newSpecPath, logger); err != nil {
		return err
	}

	diff, err := computeDiff(ctx, oldSpecPath, newSpecPath, logger)
	if err != nil {
		return fmt.Errorf("failed to compute diff: %w", err)
	}

	if err := saveChangelog(diff); err != nil {
		return err
	}

	if err := printDiff(diff); err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Total diff generation took %v", time.Since(startTime)))
	return nil
}

func validateSpecFiles(oldSpecPath, newSpecPath string, logger logging.Logger) error {
	validateStart := time.Now()

	if err := validateSingleSpecFile(oldSpecPath, "old"); err != nil {
		return err
	}

	if err := validateSingleSpecFile(newSpecPath, "new"); err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("File validation took %v", time.Since(validateStart)))
	return nil
}

func validateSingleSpecFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s spec file not found: %w", label, err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("%s spec file is empty", label)
	}
	return nil
}

func copySpecsToOutput(oldSpecPath, newSpecPath string, logger logging.Logger) error {
	copyStart := time.Now()
	err := copySpecs(oldSpecPath, newSpecPath)
	logger.Debug(fmt.Sprintf("Copying specs took %v", time.Since(copyStart)))
	return err
}

func computeDiff(ctx context.Context, oldSpecPath, newSpecPath string, logger logging.Logger) (changes.SDKDiff, error) {
	oldOptions := generate.GenerateOptions{
		Lang:       lang,
		SchemaPath: oldSpecPath,
		OutDir:     outputDir,
		Logger:     logger,
		Verbose:    false,
	}

	newOptions := generate.GenerateOptions{
		Lang:       lang,
		SchemaPath: newSpecPath,
		OutDir:     outputDir,
		Logger:     logger,
		Verbose:    false,
	}

	logger.Debug("Generating diff...")
	fmt.Fprintf(os.Stderr, "Generating diff...\n")

	diffStart := time.Now()
	diff, err := changes.Changes(ctx, oldOptions, newOptions)
	if err != nil {
		return changes.SDKDiff{}, fmt.Errorf("failed to get changes. error: %s", err.Error())
	}
	logger.Debug(fmt.Sprintf("Diff generation took %v", time.Since(diffStart)))

	return diff, nil
}

func copySpecs(oldSpecPath, newSpecPath string) error {
	oldSpecContent, err := os.ReadFile(oldSpecPath)
	if err != nil {
		return fmt.Errorf("error reading old spec: %w", err)
	}

	newSpecContent, err := os.ReadFile(newSpecPath)
	if err != nil {
		return fmt.Errorf("error reading new spec: %w", err)
	}

	if err := os.WriteFile(filepath.Join(outputDir, "old.openapi.yaml"), oldSpecContent, 0o644); err != nil {
		return fmt.Errorf("error writing old spec: %w", err)
	}

	if err := os.WriteFile(filepath.Join(outputDir, "new.openapi.yaml"), newSpecContent, 0o644); err != nil {
		return fmt.Errorf("error writing new spec: %w", err)
	}

	return nil
}

func saveChangelog(diff changes.SDKDiff) error {
	logger := logging.NewLogger(logging.LevelFromEnv())
	saveStart := time.Now()

	// Save markdown changelog
	changelogContent := changes.ToMarkdown(diff, changes.DetailLevel(detailLevel))
	if err := os.WriteFile(filepath.Join(outputDir, "changelog.txt"), []byte(changelogContent), 0o644); err != nil {
		return err
	}

	// Save HTML changelog
	htmlContent := changes.ToHTML(diff)
	htmlPath := filepath.Join(outputDir, "changelog.html")
	if err := os.WriteFile(htmlPath, htmlContent, 0o644); err != nil {
		return err
	}

	// Open in browser if requested
	if openBrowser {
		if err := openInBrowser(htmlPath); err != nil {
			logger.Warn(fmt.Sprintf("Failed to open browser: %v", err))
		}
	}

	logger.Debug(fmt.Sprintf("Saving changelog took %v", time.Since(saveStart)))
	return nil
}

func printDiff(diff changes.SDKDiff) error {
	logger := logging.NewLogger(logging.LevelFromEnv())
	printStart := time.Now()

	fmt.Printf("Changes between revision %s and %s:\n\n", oldRevision, newRevision)

	changelogContent := changes.ToMarkdown(diff, changes.DetailLevel(detailLevel))
	fmt.Println(changelogContent)

	fmt.Fprintf(os.Stderr, "\nFiles saved to: %s\n", outputDir)

	logger.Debug(fmt.Sprintf("Printing diff took %v", time.Since(printStart)))
	return nil
}

func getAPIKey() (string, error) {
	// Priority: flag > env > config file
	if apiKey != "" {
		return apiKey, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	key, err := licensebootstrap.ResolveAPIKey(os.Getenv, homeDir)
	if err != nil {
		return "", errors.New("no API key found - set SPEAKEASY_API_KEY env var or use --api-key flag")
	}
	return key, nil
}

type revisionPair struct {
	oldRevision string
	newRevision string
	namespace   string
}

func findRevisionsFromEvents(ctx context.Context, client *speakeasy.Speakeasy, org, workspace string) (*revisionPair, error) {
	startTime := time.Now()
	logger := logging.NewLogger(logging.LevelFromEnv())

	logger.Debug(fmt.Sprintf("Searching events for workspace: %s (org: %s)", workspace, org))

	workspaceID, err := getWorkspaceIDForSearch(ctx, logger)
	if err != nil {
		return nil, err
	}

	allEvents, err := getAllEvents(ctx, client, workspaceID, logger)
	if err != nil {
		return nil, err
	}

	if len(allEvents) == 0 {
		return nil, errors.New("no events found for workspace")
	}

	logger.Debug(fmt.Sprintf("Found %d total events", len(allEvents)))

	// Try to find diff event first
	if pair := findDiffEvent(allEvents, org, workspace, logger); pair != nil {
		logger.Info(fmt.Sprintf("Found suitable diff event after %v", time.Since(startTime)))
		return pair, nil
	}

	return nil, fmt.Errorf("no GENERATE events with revision information found in workspace %s/%s. Please specify --old-revision and --new-revision manually", org, workspace)
}

func getWorkspaceIDForSearch(ctx context.Context, logger logging.Logger) (string, error) {
	state, err := registryStateFromContext(ctx)
	if err != nil {
		return "", err
	}
	logger.Debug("Using workspace ID for search: " + state.workspaceID)
	return state.workspaceID, nil
}

func getAllEvents(ctx context.Context, client *speakeasy.Speakeasy, workspaceID string, logger logging.Logger) ([]shared.CliEvent, error) {
	// First try target_generate events
	searchStart := time.Now()
	targetGenerate := shared.InteractionTypeTargetGenerate
	allEvents, err := searchEvents(ctx, client, workspaceID, &targetGenerate, logger)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("Initial event search took %v, found %d events", time.Since(searchStart), len(allEvents)))

	// Check if we have enough suitable events
	if countSuitableEvents(allEvents) >= 2 {
		return allEvents, nil
	}

	// Fall back to searching all event types
	logger.Debug("Not enough target_generate events found, searching all event types")
	fallbackStart := time.Now()
	allEvents, err = searchEvents(ctx, client, workspaceID, nil, logger)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("Fallback event search took %v, found %d events", time.Since(fallbackStart), len(allEvents)))
	logger.Debug(fmt.Sprintf("After searching all events, found %d suitable events", countSuitableEvents(allEvents)))

	return allEvents, nil
}

func countSuitableEvents(events []shared.CliEvent) int {
	count := 0
	for _, event := range events {
		if event.SourceRevisionDigest != nil && *event.SourceRevisionDigest != "" &&
			event.SourceNamespaceName != nil && *event.SourceNamespaceName != "" {
			count++
		}
	}
	return count
}

func findDiffEvent(events []shared.CliEvent, _, workspace string, logger logging.Logger) *revisionPair {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	// First try to find GitHub events with diff information
	for _, event := range events {
		if !event.CreatedAt.After(threeMonthsAgo) {
			continue
		}

		if !isGitHubEventWithDifferentBlobs(event) {
			continue
		}

		// Extract namespace names and ensure they're the same
		oldNamespaceName := workspace
		newNamespaceName := workspace

		if event.OpenapiDiffBaseSourceNamespaceName != nil {
			oldNamespaceName = extractNamespaceName(*event.OpenapiDiffBaseSourceNamespaceName)
		}
		if event.SourceNamespaceName != nil {
			newNamespaceName = extractNamespaceName(*event.SourceNamespaceName)
		}

		// Skip if namespaces are different
		if oldNamespaceName != newNamespaceName {
			logger.Debug(fmt.Sprintf("Skipping event with different namespaces: %s vs %s", oldNamespaceName, newNamespaceName))
			continue
		}

		logger.Debug(fmt.Sprintf("Found event with different blob digests: %s != %s",
			*event.OpenapiDiffBaseSourceBlobDigest, *event.SourceBlobDigest))

		return &revisionPair{
			oldRevision: *event.OpenapiDiffBaseSourceRevisionDigest,
			newRevision: *event.SourceRevisionDigest,
			namespace:   oldNamespaceName,
		}
	}

	// If no GitHub diff events found, try to find any two events from same namespace with different revisions
	logger.Debug("No GitHub diff events found, looking for any two events with different revisions in same namespace")

	// Group events by namespace
	eventsByNamespace := make(map[string][]shared.CliEvent)
	for _, event := range events {
		if !hasValidRevisionInfo(event) {
			continue
		}

		namespaceName := extractNamespaceName(*event.SourceNamespaceName)
		eventsByNamespace[namespaceName] = append(eventsByNamespace[namespaceName], event)
	}

	// Find a namespace with at least 2 different revisions
	for namespaceName, namespaceEvents := range eventsByNamespace {
		var differentRevisions []shared.CliEvent
		seenDigests := make(map[string]bool)

		for _, event := range namespaceEvents {
			digest := *event.SourceRevisionDigest
			if !seenDigests[digest] {
				seenDigests[digest] = true
				differentRevisions = append(differentRevisions, event)

				if len(differentRevisions) >= 2 {
					logger.Debug("Found 2 different revisions in namespace " + namespaceName)
					return &revisionPair{
						oldRevision: *differentRevisions[1].SourceRevisionDigest,
						newRevision: *differentRevisions[0].SourceRevisionDigest,
						namespace:   namespaceName,
					}
				}
			}
		}
	}

	return nil
}

func isGitHubEventWithDifferentBlobs(event shared.CliEvent) bool {
	return event.ContinuousIntegrationEnvironment != nil &&
		*event.ContinuousIntegrationEnvironment == "github" &&
		event.OpenapiDiffBaseSourceRevisionDigest != nil &&
		event.SourceRevisionDigest != nil &&
		event.OpenapiDiffBaseSourceBlobDigest != nil &&
		event.SourceBlobDigest != nil &&
		*event.OpenapiDiffBaseSourceBlobDigest != *event.SourceBlobDigest
}

func hasValidRevisionInfo(event shared.CliEvent) bool {
	return event.SourceRevisionDigest != nil && *event.SourceRevisionDigest != "" &&
		event.SourceNamespaceName != nil && *event.SourceNamespaceName != ""
}

func searchEvents(ctx context.Context, client *speakeasy.Speakeasy, workspaceID string, interactionType *shared.InteractionType, logger logging.Logger) ([]shared.CliEvent, error) {
	searchStart := time.Now()
	request := operations.SearchWorkspaceEventsRequest{
		WorkspaceID: &workspaceID,
		Limit:       speakeasy.Int64(10000), // Get as many events as possible
	}

	if interactionType != nil {
		request.InteractionType = interactionType
	}

	resp, err := client.Events.Search(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	logger.Debug(fmt.Sprintf("Event API call took %v", time.Since(searchStart)))

	logger.Debug(fmt.Sprintf("Found %d events", len(resp.CliEventBatch)))
	return resp.CliEventBatch, nil
}

func extractNamespaceName(fullNamespace string) string {
	// If namespace is already just a name, return it
	if !strings.Contains(fullNamespace, "/") {
		return fullNamespace
	}
	// Extract the last part after the last slash
	parts := strings.Split(fullNamespace, "/")
	return parts[len(parts)-1]
}

func downloadAndExtractSpec(ctx context.Context, _ *speakeasy.Speakeasy, org, workspace, namespace, revision, outputDir string, logger logging.Logger) (string, error) {
	startTime := time.Now()
	logger.Debug(fmt.Sprintf("downloadAndExtractSpec: org=%s, workspace=%s, namespace=%s, revision=%s", org, workspace, namespace, revision))

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	authClient, err := getAuthenticatedClient(ctx, namespace, logger)
	if err != nil {
		return "", err
	}

	layerDigest, err := getLayerDigest(ctx, authClient, org, workspace, namespace, revision, logger)
	if err != nil {
		return "", err
	}

	content, err := downloadBlob(ctx, authClient, org, workspace, namespace, layerDigest, logger)
	if err != nil {
		return "", err
	}

	specPath, err := extractOpenAPIFromZip(content, outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract OpenAPI spec: %w", err)
	}

	logger.Debug(fmt.Sprintf("Total download and extract took %v", time.Since(startTime)))
	return specPath, nil
}

func getAuthenticatedClient(ctx context.Context, namespace string, logger logging.Logger) (*speakeasy.Speakeasy, error) {
	preflightStart := time.Now()
	preflightResp, err := callPreflight(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get preflight token: %w", err)
	}
	logger.Debug(fmt.Sprintf("Preflight request took %v", time.Since(preflightStart)))

	state, err := registryStateFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authClient := speakeasy.New(
		speakeasy.WithClient(state.httpClient),
		speakeasy.WithSecurity(shared.Security{
			Bearer: speakeasy.String(preflightResp.AuthToken),
		}),
	)

	return authClient, nil
}

func getLayerDigest(ctx context.Context, client *speakeasy.Speakeasy, org, workspace, namespace, revision string, logger logging.Logger) (string, error) {
	manifestStart := time.Now()
	manifestRequest := operations.GetManifestRequest{
		OrganizationSlug:  org,
		WorkspaceSlug:     workspace,
		NamespaceName:     namespace,
		RevisionReference: revision,
	}

	manifestResp, err := client.Artifacts.GetManifest(ctx, manifestRequest)
	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %w", err)
	}
	logger.Debug(fmt.Sprintf("Get manifest took %v", time.Since(manifestStart)))

	if manifestResp.Manifest == nil || len(manifestResp.Manifest.Layers) == 0 {
		return "", errors.New("no layers found in manifest")
	}

	for _, layer := range manifestResp.Manifest.Layers {
		if layer.Digest != nil && *layer.Digest != "" {
			logger.Debug("Using layer digest: " + *layer.Digest)
			return *layer.Digest, nil
		}
	}

	return "", errors.New("no digest found in manifest layers")
}

func downloadBlob(ctx context.Context, client *speakeasy.Speakeasy, org, workspace, namespace, digest string, logger logging.Logger) ([]byte, error) {
	blobStart := time.Now()
	blobRequest := operations.GetBlobRequest{
		OrganizationSlug: org,
		WorkspaceSlug:    workspace,
		NamespaceName:    namespace,
		Digest:           digest,
	}

	blobResp, err := client.Artifacts.GetBlob(ctx, blobRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}
	defer blobResp.Blob.Close()
	logger.Debug(fmt.Sprintf("Get blob took %v", time.Since(blobStart)))

	content, err := io.ReadAll(blobResp.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	logger.Debug(fmt.Sprintf("Downloaded blob size: %d bytes", len(content)))
	return content, nil
}

func callPreflight(ctx context.Context, namespace string) (*PreflightResponse, error) {
	state, err := registryStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Create preflight request
	preflightURL := state.preflightURL + "/v1/artifacts/preflight"

	reqBody := map[string]string{
		"namespace_name": namespace,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, preflightURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", state.apiKey)
	req.Header.Set("x-workspace-id", state.workspaceID)

	resp, err := state.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("preflight request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var preflightResp PreflightResponse
	if err := json.NewDecoder(resp.Body).Decode(&preflightResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &preflightResp, nil
}

type PreflightResponse struct {
	AuthToken   string `json:"auth_token"`
	NamespaceID string `json:"namespace_id"`
}

func extractOpenAPIFromZip(zipContent []byte, outputDir string) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}

	for _, file := range reader.File {
		if !isOpenAPIFile(file.Name) {
			continue
		}

		return extractAndWriteFile(file, outputDir)
	}

	return "", errors.New("no OpenAPI spec found in zip file")
}

func isOpenAPIFile(filename string) bool {
	if !strings.Contains(filename, "openapi") {
		return false
	}

	return strings.HasSuffix(filename, ".yaml") ||
		strings.HasSuffix(filename, ".yml") ||
		strings.HasSuffix(filename, ".json")
}

func extractAndWriteFile(file *zip.File, outputDir string) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	extension := ".yaml"
	if strings.HasSuffix(file.Name, ".json") {
		extension = ".json"
	}

	specPath := filepath.Join(outputDir, "openapi"+extension)
	if err := os.WriteFile(specPath, content, 0o644); err != nil {
		return "", fmt.Errorf("failed to write spec file: %w", err)
	}

	return specPath, nil
}

func openInBrowser(path string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{path}
	case "linux":
		cmd = "xdg-open"
		args = []string{path}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", path}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
