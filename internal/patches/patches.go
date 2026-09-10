package patches

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
)

// ProgressCallback is the function type for reporting merge progress.
// This matches the signature used by the generator's OnProgressUpdate.
type ProgressCallback func(stepID string, merge *MergeProgress)

// MergeProgress contains merge-specific progress information (conflicts, errors).
type MergeProgress struct {
	ConflictFile string // Set when a file conflict is detected
	PushError    error  // Set when pushing to remote fails (soft failure)
}

// Subsystem handles 3-way merge with persistent edits
type Subsystem struct {
	Enabled     *config.PersistentEditsEnabled
	config      *configuration.Config
	fileSystem  filesystem.FileSystem
	git         merge.Git
	lockFile    *config.LockFile
	fileTracker *filetracking.Tracker
	outDir      string

	// OnProgress is called to report merge phases and events.
	// Injected by the Generator to enable progress reporting.
	OnProgress ProgressCallback

	// scannedIDs maps UUID -> current path on disk (from scanning @generated-id headers)
	// This is populated lazily on first call to GetActualPath
	scannedIDs map[string]string
	scanDone   bool

	// pathRemapping maps computed path -> actual path (where user moved the file)
	// Built from scannedIDs + lockFile.TrackedFiles
	pathRemapping map[string]string
}

// NewSubsystem creates a new patches subsystem with the given configuration.
func NewSubsystem(
	cfg *configuration.Config,
	fs filesystem.FileSystem,
	git merge.Git,
	lockFile *config.LockFile,
	fileTracker *filetracking.Tracker,
	outDir string,
	enabled *config.PersistentEditsEnabled,
) *Subsystem {
	return &Subsystem{
		config:      cfg,
		fileSystem:  fs,
		git:         git,
		lockFile:    lockFile,
		fileTracker: fileTracker,
		outDir:      outDir,
		Enabled:     enabled,
	}
}

// IsEnabled returns true if persistent edits are explicitly enabled ("true")
func (p *Subsystem) IsEnabled() bool {
	return p.Enabled != nil && *p.Enabled == config.PersistentEditsEnabledTrue
}

// IsNever returns true if persistent edits are explicitly disabled ("never")
func (p *Subsystem) IsNever() bool {
	return p.Enabled != nil && *p.Enabled == config.PersistentEditsEnabledNever
}

// ensureScanned performs the scan for @generated-id headers if not already done.
// This builds the pathRemapping map that maps computed paths to actual paths.
func (p *Subsystem) ensureScanned() {
	if p.scanDone {
		return
	}
	p.scanDone = true
	p.pathRemapping = make(map[string]string)

	if p.fileSystem == nil {
		return
	}

	// Scan for current UUID -> path mappings on disk
	scannedIDs, err := p.fileSystem.ScanForGeneratedIDs()
	if err != nil {
		// Log but don't fail - scanning is best-effort
		return
	}
	p.scannedIDs = scannedIDs

	// Build path remapping by comparing scanned IDs with computed IDs.
	// If a file at path B has an embedded ID that equals ComputeFileID(A),
	// it means the file was moved from A to B.
	//
	// We build a reverse map: computedID -> actualPath (from scan)
	// Then for each tracked path, check if its computed ID is at a different location.
	idToActualPath := make(map[string]string)
	for id, path := range scannedIDs {
		idToActualPath[id] = path
	}

	if p.lockFile == nil {
		return
	}

	// For each tracked file, check if its computed ID is now at a different location
	for computedPath := range p.lockFile.TrackedFiles.Keys() {
		expectedID := ComputeFileID(computedPath)
		if actualPath, found := idToActualPath[expectedID]; found {
			if actualPath != computedPath {
				// File was moved from computedPath to actualPath
				p.pathRemapping[computedPath] = actualPath
			}
		}
	}
}

// GetActualPath returns the actual path for a file, accounting for user moves.
// If the user moved a file from path A to path B, and we're trying to write to A,
// this returns B so we respect the user's file organization.
func (p *Subsystem) GetActualPath(computedPath string) string {
	p.ensureScanned()
	if actualPath, found := p.pathRemapping[computedPath]; found {
		return actualPath
	}
	return computedPath
}

// ComputeFileID returns a deterministic short ID (12 hex chars) for a file path.
// This is the default ID for a file when no embedded ID exists.
// CRITICAL: Path is normalized to forward slashes to ensure consistent IDs
// across Windows/Linux/macOS.
//
// The ID is the first 6 bytes (12 hex chars) of SHA1(path), providing:
// - 48 bits of entropy (~0.00000001% collision probability at 10k files)
// - Readable format in file headers: @generated-id: a1b2c3d4e5f6
// - Deterministic: same path always produces same ID (no lockfile churn)
func ComputeFileID(path string) string {
	normalized := filepath.ToSlash(path)
	hash := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:6]) // 12 hex chars = 48 bits
}

// GetOrCreateID returns the ID for a file path.
// If the file already has an embedded ID (from scanning), use that for stability.
// Otherwise, compute a deterministic ID from the path using SHA1.
// This means:
// - New files get deterministic IDs (no storage needed, no churn)
// - Files with embedded IDs keep those IDs (enables move detection)
// - Supports both legacy UUIDs and new short IDs (12 hex chars)
//
// The path parameter may be absolute or relative. For lookup, we convert to
// outDir-relative (matching ScanForGeneratedIDs output). For computation,
// we use repo-relative paths to ensure unique IDs across targets.
func (p *Subsystem) GetOrCreateID(path string) string {
	p.ensureScanned()

	// Convert to outDir-relative path for lookup in scannedIDs.
	// ScanForGeneratedIDs returns paths relative to outDir (e.g., "sdk.go").
	lookupPath := path
	if p.outDir != "" {
		if rel, err := filepath.Rel(p.outDir, path); err == nil && !strings.HasPrefix(rel, "..") {
			lookupPath = filepath.ToSlash(rel)
		}
	}

	// Check if file already has an embedded ID on disk - use it for stability
	// This is important for move detection: the embedded ID stays constant
	// even when the file is moved to a new path.
	for id, scannedPath := range p.scannedIDs {
		if filepath.ToSlash(scannedPath) == lookupPath {
			return id
		}
	}

	// No embedded ID found - compute deterministic ID from path.
	// Use repo-relative path to ensure:
	// 1. Stable IDs regardless of where repo is checked out (no absolute paths)
	// 2. Unique IDs across targets (packages/sub/sdk.go != sdk.go)
	computePath := p.toRepoRelativePath(path)
	return ComputeFileID(computePath)
}

// toRepoRelativePath converts an absolute path to a repo-relative path.
// This ensures IDs are stable across different machines/checkouts while
// remaining unique across different targets in the same repository.
func (p *Subsystem) toRepoRelativePath(path string) string {
	// If already relative, use as-is
	if !filepath.IsAbs(path) {
		return path
	}

	// Use the git repository root for consistent path resolution
	if p.git != nil {
		if repoRoot := p.git.RepoRoot(); repoRoot != "" {
			if rel, err := filepath.Rel(repoRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
				return rel
			}
		}
	}

	// Fallback: use the path as-is (will still work, just less portable)
	return path
}

// PerformMerge performs the complete 3-way merge, commit, push, and lockfile update.
// Progress is reported via the logger's step tracking system.
// Returns (isNoOp, error) where isNoOp is true if no changes were needed.
// Returns merge.ConflictsError if merge conflicts are detected (files are still written with conflict markers).
//
// When called within a WithStepCtx block, substeps will influence the parent's outcome:
// if all substeps skip, the parent step will also skip.
func (p *Subsystem) PerformMerge(ctx context.Context, virtualFiles map[string]*merge.VirtualFile) (bool, error) {
	// Early exit if persistent edits is explicitly disabled ("never")
	if p.IsNever() {
		logging.StartStepCtx(ctx, "Skipped").Skip()
		return false, nil
	}

	// Validate dependencies - skip gracefully if not available (unless enabled, then error)
	if p.git == nil {
		logging.StartStepCtx(ctx, "Skipped").Skip()
		if p.IsEnabled() {
			return false, errors.New("persistent edits is enabled but this directory is not a git repository. Either initialize git (`git init`) or disable persistent edits in gen.yaml")
		}
		return false, nil // Skip gracefully when not enabled
	}
	if p.fileSystem == nil {
		logging.StartStepCtx(ctx, "Skipped").Skip()
		if p.IsEnabled() {
			return false, errors.New("persistent edits: fileSystem interface is nil (internal error)")
		}
		return false, nil
	}
	if p.lockFile == nil {
		logging.StartStepCtx(ctx, "Skipped").Skip()
		if p.IsEnabled() {
			return false, errors.New("persistent edits: lockFile is nil (internal error)")
		}
		return false, nil
	}

	// Skip if no files to merge
	if len(virtualFiles) == 0 {
		logging.StartStepCtx(ctx, "Skipped").Skip()
		return false, nil
	}

	// Create generation snapshot
	snapshotStep := logging.StartStepCtx(ctx, "Creating generation snapshot")

	// Get tracked files from FileTracker - this excludes untracked files like
	// CONTRIBUTING.md, README.md, etc. We use this to filter the tree hash so it's
	// stable across runs (untracked files may or may not be regenerated).
	var trackedFiles config.TrackedFiles
	if p.fileTracker != nil {
		trackedFiles = p.fileTracker.GetResult().NewFiles
	}

	// Prepare merge context
	mergeCtx := &merge.MergeContext{
		VirtualFiles: virtualFiles,
		TrackedFiles: trackedFiles,
		LockFile:     p.lockFile,
		FileSystem:   p.fileSystem,
		Git:          p.git,
		OutDir:       p.outDir,
		Enabled:      p.IsEnabled(),
	}
	snapshotStep.Succeed()

	// Phase 3: Perform the 3-way merge
	var mergeStep logging.Step
	if p.IsEnabled() {
		mergeStep = logging.StartStepCtx(ctx, "Merging custom edits")
	} else {
		mergeStep = logging.StartStepCtx(ctx, "Merging custom edits")
		mergeStep.Skip()
	}
	result, err := merge.Merge(mergeCtx)
	if err != nil {
		if mergeStep != nil {
			mergeStep.Fail()
		}
		return false, err
	}

	// Handle no-op case (generated tree matches previous pristine tree)
	if result.IsNoOp {
		if mergeStep != nil {
			mergeStep.Succeed()
		}
		return true, nil
	}

	// Mark merge step as successful before continuing
	if mergeStep != nil {
		mergeStep.Succeed()
	}

	// Collect all conflicts and emit events for each
	conflictPaths := result.ConflictFiles
	for _, path := range conflictPaths {
		p.OnProgress("", &MergeProgress{ConflictFile: path})
	}

	// Create commit
	parentHash := ""
	if p.lockFile.PersistentEdits != nil {
		parentHash = p.lockFile.PersistentEdits.PristineCommitHash
	}
	// We MUST validate the parent commit exists in this repo (and fetch it if not). If it's ever lost, it's no biggie, but we should find another appropriate parent that IS committed
	// to try and keep delta compression working.
	// This might happen if network communication is lost part way through generation.
	if parentHash != "" && !p.git.HasObject(parentHash) {
		// Parent doesn't exist locally - try to fetch it
		if p.lockFile.PersistentEdits != nil && p.lockFile.PersistentEdits.GenerationID != "" && p.IsEnabled() {
			genID := p.lockFile.PersistentEdits.GenerationID
			refName := "refs/speakeasy/gen/" + genID

			// Track fetch progress in workflow UI
			gitFetchStep := logging.StartStepCtx(ctx, "git fetch origin "+refName)
			fetchErr := p.git.FetchSnapshot(genID)
			if fetchErr != nil {
				gitFetchStep.Fail()
			} else {
				gitFetchStep.Succeed()
			}
		}
		// Check again after fetch attempt
		if !p.git.HasObject(parentHash) {
			// Parent is unreachable - create an orphan commit instead.
			// This loses history linkage but ensures the commit can be pushed.
			parentHash = ""
		}
	}

	commitMessage := "Generate SDK: " + p.lockFile.ID
	commitHash, err := p.git.CommitSnapshot(result.TreeHash, parentHash, commitMessage)
	if err != nil {
		return false, err
	}

	// Generate new UUID for this generation
	generationID := uuid.NewString()
	result.GenerationID = generationID

	// Phase 4: Push snapshot to remote
	// SOFT-FAIL: If push fails, warn but continue - user has working code locally
	if p.IsEnabled() {
		refName := "refs/speakeasy/gen/" + generationID

		// Track push progress in workflow UI
		pushStep := logging.StartStepCtx(ctx, "git push origin "+refName)
		pushErr := p.git.PushSnapshot(commitHash, generationID)
		if pushErr != nil {
			pushStep.Fail()
			// Report the push failure but don't return error
			// The user's primary goal (getting generated code) succeeded
			p.OnProgress("", &MergeProgress{PushError: pushErr})
			// Continue with lockfile update - next run will attempt push again
		} else {
			pushStep.Succeed()
		}
	}

	// Phase 5: Apply merged files (update lockfile state)
	applyStep := logging.StartStepCtx(ctx, "Applying merged files")
	defer applyStep.Succeed()

	// Update lockfile with new persistent edits state
	if p.lockFile.PersistentEdits == nil {
		p.lockFile.PersistentEdits = &lockfile.PersistentEdits{}
	}
	p.lockFile.PersistentEdits.GenerationID = generationID
	p.lockFile.PersistentEdits.PristineCommitHash = commitHash
	p.lockFile.PersistentEdits.PristineTreeHash = result.TreeHash

	// Populate file tracker with IDs and pristine hashes for generated files.
	if p.fileTracker != nil {
		for path := range virtualFiles {
			if tracked, ok := p.lockFile.TrackedFiles.Get(path); ok {
				id := tracked.ID
				if id == "" {
					id = ComputeFileID(path)
				}
				p.fileTracker.UpdateTrackedFile(path, id, tracked.PristineGitObject)
			} else {
				id := ComputeFileID(path)
				if updatedTracked, ok := p.lockFile.TrackedFiles.Get(path); ok {
					p.fileTracker.UpdateTrackedFile(path, id, updatedTracked.PristineGitObject)
				}
			}
		}
	}

	// Return error AFTER applying files if conflicts exist
	// This ensures conflict markers are written to disk before exiting
	if len(conflictPaths) > 0 {
		return false, &merge.ConflictsError{Files: conflictPaths}
	}

	return false, nil
}
