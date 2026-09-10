package generate

import (
	"context"
	stderrors "errors"
	"fmt"
	"path/filepath"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
)

// handlePatches orchestrates the 3-way merge using the patches subsystem.
// This is called after template rendering completes (virtualFiles are populated).
// Returns (isNoOp, error) where isNoOp is true if no changes were needed.
//
// The merge is wrapped in a WithStepCtx so that if all substeps skip,
// the "Custom Code" parent step also shows as skipped.
func (g *Generator) handlePatches(ctx context.Context) (bool, error) {
	// Convert sync.Map to regular map for PerformMerge
	files := make(map[string]*merge.VirtualFile)
	g.virtualFiles.Range(func(k string, v *merge.VirtualFile) bool {
		files[k] = v
		return true
	})

	var isNoOp bool
	var mergeErr error
	err := logging.WithStepCtx(ctx, "Custom Code", func(ctx context.Context) error {
		isNoOp, mergeErr = g.subsystem.Patches.PerformMerge(ctx, files)
		return mergeErr
	})

	return isNoOp, err
}

// absorbVirtualFiles copies all virtual files from the source generator into
// the receiver's virtualFiles map. This allows mock server files to participate
// in the parent's persistent edits pipeline (snapshot, flush, compile sync, merge).
func (g *Generator) absorbVirtualFiles(source *Generator) {
	source.virtualFiles.Range(func(key string, value *merge.VirtualFile) bool {
		g.virtualFiles.Store(key, value)
		return true
	})
}

// snapshotUserFiles captures the user's current files from disk into memory.
// This snapshot lets us temporarily overwrite files for compilation, then restore
// the user's working copy so 3-way merges compare against the correct baseline.
// Returns a map of relative path -> file content.
func (g *Generator) snapshotUserFiles() map[string][]byte {
	theirsContent := make(map[string][]byte)

	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		fullPath := filepath.Join(g.outDir, path)
		content, err := g.ReadFile(fullPath)
		if err == nil {
			theirsContent[path] = content
		}
		// If file doesn't exist, that's fine - no "Theirs" to preserve
		return true
	})

	return theirsContent
}

// flushGeneratedFilesToDisk writes generated files directly to disk without merge.
// This is a transient step to give compilers/formatters real files to operate on;
// we restore the user's files afterward from the snapshot.
func (g *Generator) flushGeneratedFilesToDisk() error {
	var firstErr error

	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		fullPath := filepath.Join(g.outDir, path)

		if err := g.WriteFile(fullPath, vf.Content, vf.Mode); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
		return true
	})

	return firstErr
}

// syncPristineFromCompiledFiles reads files back from disk after compilation
// and updates the virtual files with post-compile content. This ensures pristine
// checksums and future patch calculations match what actually exists on disk
// after formatters/linters run, not the raw generator output.
func (g *Generator) syncPristineFromCompiledFiles() error {
	var firstErr error

	// Build reverse lookup: actual path -> original tracking path
	// This handles files where user moved them (MovedTo is set)
	actualToTrackingPath := make(map[string]string)
	for origPath := range g.lockFile.TrackedFiles.Keys() {
		if tracked, ok := g.lockFile.TrackedFiles.Get(origPath); ok && tracked.MovedTo != "" {
			actualToTrackingPath[tracked.MovedTo] = origPath
		}
	}

	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		fullPath := filepath.Join(g.outDir, path)

		content, err := g.ReadFile(fullPath)
		if err != nil {
			if firstErr == nil {
				if _, statErr := g.Stat(fullPath); statErr != nil {
					return true // file deleted by compile, skip
				}
				firstErr = err
			}
			return true
		}

		// Update virtual file with post-compile content
		vf.Content = content

		// Determine the tracking path (original path if file was moved, otherwise same as path)
		trackingPath := path
		if origPath, isMoved := actualToTrackingPath[path]; isMoved {
			trackingPath = origPath
		}

		// Update tracked file checksum to match post-compile content
		g.subsystem.FileTracker.UpdateTrackedFileChecksum(trackingPath, content)
		return true
	})

	return firstErr
}

// compilePristine runs the pristine compile pass: flush virtual files to disk,
// invoke the compile pipeline so formatters/linters can mutate them, then sync
// the post-compile content back into virtual files and tracked checksums.
//
// When patches are enabled, the user's working copy is snapshotted first and
// restored after compile so handlePatches sees the correct baseline for 3-way
// merges. On any failure mid-way, the snapshot is restored before returning.
//
// Returns []error to preserve multi-error fan-out from compileAll.
func (g *Generator) compilePristine(ctx context.Context, mockServerGenerator *Generator) []error {
	var errs []error
	_ = logging.WithStepCtx(ctx, "Compile Pristine", func(ctx context.Context) error {
		var userSnapshot map[string][]byte
		if g.subsystem.Patches.IsEnabled() {
			userSnapshot = g.snapshotUserFiles()
		}

		if err := g.flushGeneratedFilesToDisk(); err != nil {
			errs = []error{fmt.Errorf("failed to flush generated files: %w", err)}
			return err
		}

		if errs = g.compileAll(ctx, mockServerGenerator); len(errs) > 0 {
			g.tryRestoreUserFiles(ctx, userSnapshot)
			return stderrors.Join(errs...)
		}

		if err := g.syncPristineFromCompiledFiles(); err != nil {
			g.tryRestoreUserFiles(ctx, userSnapshot)
			errs = []error{fmt.Errorf("failed to sync pristine from compiled files: %w", err)}
			return err
		}

		if err := g.saveTrackedFilesToLockFile(ctx); err != nil {
			g.tryRestoreUserFiles(ctx, userSnapshot)
			errs = []error{fmt.Errorf("failed to save tracked files after compile: %w", err)}
			return err
		}

		if userSnapshot != nil {
			if err := g.restoreUserFilesFromSnapshot(userSnapshot); err != nil {
				errs = []error{fmt.Errorf("failed to restore user files: %w", err)}
				return err
			}
		}

		return nil
	})
	return errs
}

// tryRestoreUserFiles best-effort restore on the failure path. Errors are
// logged but not surfaced — the original failure takes precedence.
func (g *Generator) tryRestoreUserFiles(ctx context.Context, snapshot map[string][]byte) {
	if snapshot == nil {
		return
	}
	if err := g.restoreUserFilesFromSnapshot(snapshot); err != nil {
		logging.From(ctx).Warn(fmt.Sprintf("failed to restore user files after compile failure: %v", err))
	}
}

// restoreUserFilesFromSnapshot writes the previously snapshotted user files back
// to disk. This reverses the temporary overwrite we did for compilation so that
// handlePatches sees the user's actual working copy when performing 3-way merges.
func (g *Generator) restoreUserFilesFromSnapshot(snapshot map[string][]byte) error {
	var firstErr error

	for path, content := range snapshot {
		fullPath := filepath.Join(g.outDir, path)
		if err := g.WriteFile(fullPath, content, defaultFileMode); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
