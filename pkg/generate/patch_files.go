package generate

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	gitdiffparser "github.com/speakeasy-api/git-diff-parser"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/patchfiles"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
	"go.uber.org/zap"
)

func (g *Generator) applyImplicitPatchFiles(ctx context.Context) error {
	hasPatchFiles, err := patchfiles.HasAny(g, g.outDir)
	if err != nil {
		return fmt.Errorf("checking patch files: %w", err)
	}
	if !hasPatchFiles {
		return nil
	}

	step := logging.StartStepCtx(ctx, "Applying patch files")
	appliedAny := false
	var applyErrs []error
	usedPaths := map[string]struct{}{}
	conflictPaths := make([]string, 0)

	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		patchData, used, err := patchfiles.Read(g, g.outDir, path)
		if err != nil {
			applyErrs = append(applyErrs, fmt.Errorf("reading patch file for %s: %w", path, err))
			return true
		}
		if !used {
			return true
		}
		operations, err := gitdiffparser.ParsePatchOperations(patchData)
		if err != nil {
			applyErrs = append(applyErrs, fmt.Errorf("parsing patch file for %s: %w", path, err))
			return true
		}
		if patchOperationsMutateFileSet(operations) {
			return true
		}

		appliedAny = true
		usedPaths[normalizePatchLookupPath(path)] = struct{}{}

		currentGenerated := append([]byte(nil), vf.Content...)
		patchedContent, err := gitdiffparser.ApplyFileWithConflicts(vf.Content, patchData)
		if err == nil {
			vf.Content = patchedContent
			g.updateTrackedChecksum(path, patchedContent)

			if g.shouldDeferPatchWrite() {
				return true
			}

			if err := g.WriteFile(filepath.Join(g.outDir, path), patchedContent, vf.Mode); err != nil {
				applyErrs = append(applyErrs, fmt.Errorf("writing patched file %s: %w", path, err))
			}
			return true
		}

		if !stderrors.Is(err, gitdiffparser.ErrPatchConflict) {
			applyErrs = append(applyErrs, fmt.Errorf("patch apply failed for %s: %w", path, err))
			return true
		}

		hasConflict, materializeErr := g.materializePatchConflict(ctx, path, vf, patchData, currentGenerated, patchedContent)
		if materializeErr != nil {
			applyErrs = append(applyErrs, materializeErr)
			return true
		}
		g.updateTrackedChecksum(path, vf.Content)

		if hasConflict {
			conflictPaths = append(conflictPaths, path)
		}
		return true
	})

	if len(applyErrs) > 0 {
		step.Fail()
		return stderrors.Join(applyErrs...)
	}
	genericApplied, err := g.applyImplicitGenericPatchFiles(usedPaths)
	if err != nil {
		step.Fail()
		return err
	}
	appliedAny = appliedAny || genericApplied
	if err := g.warnUnusedPatchFiles(ctx, usedPaths); err != nil {
		step.Fail()
		return err
	}
	if len(conflictPaths) > 0 {
		step.Fail()
		return &merge.ConflictsError{Files: conflictPaths}
	}
	if !appliedAny {
		step.Skip()
		return nil
	}

	step.Succeed()
	return nil
}

func (g *Generator) applyImplicitGenericPatchFiles(usedPaths map[string]struct{}) (bool, error) {
	patches, err := patchfiles.ReadAll(g, g.outDir)
	if err != nil {
		return false, fmt.Errorf("listing patch files: %w", err)
	}

	exactPaths := map[string]struct{}{}
	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		exactPaths[normalizePatchLookupPath(path)] = struct{}{}
		return true
	})

	appliedAny := false
	for _, patchFile := range patches {
		if _, used := usedPaths[patchFile.Path]; used {
			continue
		}

		operations, err := gitdiffparser.ParsePatchOperations(patchFile.Data)
		if err != nil {
			return false, fmt.Errorf("parsing patch file %s.patch: %w", patchFile.Path, err)
		}
		mutatesFileSet := patchOperationsMutateFileSet(operations)
		if _, exact := exactPaths[patchFile.Path]; exact && !mutatesFileSet {
			continue
		}

		virtualFiles := g.collectVirtualFiles()
		_, currentExact := virtualFiles[patchFile.Path]
		if patchMentionsLookupPath(patchFile.Data, patchFile.Path) && !mutatesFileSet && !currentExact {
			continue
		}

		if err := g.validatePatchOperations(operations, virtualFiles); err != nil {
			return false, fmt.Errorf("invalid patch file %s.patch: %w", patchFile.Path, err)
		}

		patchVirtualFiles := g.collectPatchVirtualFiles(virtualFiles, operations)
		diskBackedPaths := g.collectDiskBackedPatchTargets(virtualFiles, operations)
		appliedTree, err := gitdiffparser.ApplyPatchOperations(virtualFileContents(patchVirtualFiles), operations)
		if err != nil {
			return false, fmt.Errorf("patch apply failed for %s.patch: %w", patchFile.Path, err)
		}

		changed, err := g.writeAppliedVirtualFileTree(appliedTree, patchVirtualFiles, diskBackedPaths, operations)
		if err != nil {
			return false, fmt.Errorf("writing patched files from %s.patch: %w", patchFile.Path, err)
		}
		if changed {
			usedPaths[patchFile.Path] = struct{}{}
			appliedAny = true
		}
	}

	return appliedAny, nil
}

// materializePatchConflict handles a patch that did not apply cleanly to the
// freshly generated content. When a pristine baseline is available it rebases
// the patch via 3-way merge:
//
//	base   = lockfile pristine (the content the patch was likely authored against)
//	ours   = currentGenerated  (regenerated content; may differ from base if spec evolved)
//	theirs = base + patch      (patch's intended end state)
//
// Disk content is intentionally excluded — user disk edits are reconciled by
// the persistent-edits subsystem (handlePatches) later in the pipeline.
//
// Outcome:
//   - auto-merge: VF holds merged content; disk write is deferred when
//     persistent-edits is enabled (handlePatches owns disk reconciliation).
//     A warning is emitted when the patch was rebased onto a different base.
//   - merge conflict: conflict markers are written to disk immediately so
//     the user sees them; git index conflict state is staged when available.
//   - no pristine baseline: falls back to writing the inline conflict markers
//     produced by gitdiffparser.
func (g *Generator) materializePatchConflict(ctx context.Context, path string, vf *merge.VirtualFile, patchData, currentGenerated, conflictContent []byte) (bool, error) {
	fullPath := filepath.Join(g.outDir, path)
	finalContent := conflictContent
	hasConflict := true

	if baseContent, oursContent, theirsContent, ok := g.patchConflictStages(path, currentGenerated, patchData); ok {
		mergedContent, mergeConflict := merge.Merge3Way(baseContent, oursContent, theirsContent)
		finalContent = mergedContent
		hasConflict = mergeConflict

		if !mergeConflict && !bytes.Equal(baseContent, oursContent) {
			logging.From(ctx).Warn(
				"Stale patch was rebased onto regenerated content. Please verify result",
				zap.String("file", path),
			)
		}

		if g.git != nil && hasConflict {
			isExecutable := vf.Mode&0o111 != 0
			if err := g.git.SetConflictState(path, baseContent, oursContent, theirsContent, isExecutable); err != nil {
				logging.From(ctx).Warn(
					"Failed to stage patch conflict in git index",
					zap.String("file", path),
					zap.Error(err),
				)
			}
		}
	}

	vf.Content = append([]byte(nil), finalContent...)
	if !hasConflict && g.shouldDeferPatchWrite() {
		return hasConflict, nil
	}
	if err := g.WriteFile(fullPath, finalContent, vf.Mode); err != nil {
		return false, fmt.Errorf("writing patch conflict markers for %s: %w", path, err)
	}

	return hasConflict, nil
}

func (g *Generator) patchConflictStages(path string, currentGenerated, patchData []byte) ([]byte, []byte, []byte, bool) {
	if g.git == nil || g.lockFile == nil || g.lockFile.TrackedFiles == nil {
		return nil, nil, nil, false
	}

	trackingPath := g.trackingPathForVirtualFile(path)
	tracked, ok := g.lockFile.TrackedFiles.Get(trackingPath)
	if !ok || tracked.PristineGitObject == "" {
		return nil, nil, nil, false
	}

	if !g.git.HasObject(tracked.PristineGitObject) && g.lockFile.PersistentEdits != nil {
		if err := g.git.FetchSnapshot(g.lockFile.PersistentEdits.GenerationID); err != nil {
			return nil, nil, nil, false
		}
	}
	if !g.git.HasObject(tracked.PristineGitObject) {
		return nil, nil, nil, false
	}

	baseContent, err := g.git.ReadBlob(tracked.PristineGitObject)
	if err != nil {
		return nil, nil, nil, false
	}

	theirsContent, err := gitdiffparser.ApplyFile(baseContent, patchData)
	if err != nil {
		return nil, nil, nil, false
	}

	return baseContent, currentGenerated, theirsContent, true
}

func (g *Generator) shouldDeferPatchWrite() bool {
	return g.subsystem != nil && g.subsystem.Patches != nil && g.subsystem.Patches.IsEnabled()
}

func (g *Generator) updateTrackedChecksum(path string, content []byte) {
	if g.subsystem != nil && g.subsystem.FileTracker != nil {
		g.subsystem.FileTracker.UpdateTrackedFileChecksum(g.trackingPathForVirtualFile(path), content)
	}
}

func (g *Generator) warnUnusedPatchFiles(ctx context.Context, usedPaths map[string]struct{}) error {
	existingPaths, err := patchfiles.ExistingPaths(g, g.outDir)
	if err != nil {
		return fmt.Errorf("listing patch files: %w", err)
	}

	for _, patchPath := range existingPaths {
		if _, ok := usedPaths[patchPath]; ok {
			continue
		}
		logging.From(ctx).Warn("Patch file does not match any generated file and was not applied", zap.String("patch", patchPath))
	}

	return nil
}

func (g *Generator) trackingPathForVirtualFile(path string) string {
	if g.lockFile == nil || g.lockFile.TrackedFiles == nil {
		return path
	}

	for trackedPath := range g.lockFile.TrackedFiles.Keys() {
		tracked, ok := g.lockFile.TrackedFiles.Get(trackedPath)
		if ok && tracked.MovedTo == path {
			return trackedPath
		}
	}

	return path
}

func normalizePatchLookupPath(p string) string {
	return path.Clean(filepath.ToSlash(p))
}

func (g *Generator) collectVirtualFiles() map[string]*merge.VirtualFile {
	files := map[string]*merge.VirtualFile{}
	g.virtualFiles.Range(func(path string, vf *merge.VirtualFile) bool {
		files[path] = vf
		return true
	})
	return files
}

func virtualFileContents(files map[string]*merge.VirtualFile) map[string][]byte {
	tree := make(map[string][]byte, len(files))
	for path, vf := range files {
		tree[path] = vf.Content
	}
	return tree
}

func (g *Generator) writeAppliedVirtualFileTree(appliedTree map[string][]byte, virtualFiles map[string]*merge.VirtualFile, diskBackedPaths map[string]struct{}, operations []gitdiffparser.PatchOperation) (bool, error) {
	appliedTree, err := g.normalizeAppliedPatchTree(appliedTree)
	if err != nil {
		return false, err
	}

	changed := false
	for path, vf := range virtualFiles {
		content, exists := appliedTree[path]
		if !exists {
			continue
		}
		if bytes.Equal(vf.Content, content) {
			continue
		}

		changed = true
		vf.Content = append([]byte(nil), content...)
		g.updateTrackedChecksum(path, content)

		if g.shouldDeferPatchWrite() {
			continue
		}
		if err := g.WriteFile(filepath.Join(g.outDir, path), content, vf.Mode); err != nil {
			return false, err
		}
	}

	for p, content := range appliedTree {
		if _, ok := virtualFiles[p]; ok {
			continue
		}

		mode, err := patchModeForPath(operations, p)
		if err != nil {
			return false, err
		}
		if mode == 0 {
			mode = defaultFileMode
		}
		if _, diskBacked := diskBackedPaths[p]; diskBacked {
			if err := g.WriteFile(filepath.Join(g.outDir, p), content, mode); err != nil {
				return false, err
			}
			changed = true
			continue
		}

		vf := &merge.VirtualFile{
			Path:     p,
			Content:  append([]byte(nil), content...),
			Mode:     mode,
			IsBinary: IsBinaryContent(content),
		}
		g.virtualFiles.Store(p, vf)
		if err := g.trackPatchedCreatedFile(p, content); err != nil {
			return false, err
		}
		changed = true

		if g.shouldDeferPatchWrite() {
			continue
		}
		if err := g.WriteFile(filepath.Join(g.outDir, p), content, vf.Mode); err != nil {
			return false, err
		}
	}

	for path := range virtualFiles {
		if _, ok := appliedTree[path]; ok {
			continue
		}
		g.virtualFiles.Delete(path)
		g.untrackPatchedDeletedFile(path)
		changed = true

		if g.shouldDeferPatchWrite() {
			continue
		}
		if err := g.Remove(filepath.Join(g.outDir, path)); err != nil && !stderrors.Is(err, fs.ErrNotExist) {
			return false, err
		}
	}
	return changed, nil
}

func (g *Generator) validatePatchOperations(operations []gitdiffparser.PatchOperation, virtualFiles map[string]*merge.VirtualFile) error {
	for _, op := range operations {
		if op.IsBinary || op.Type == gitdiffparser.PatchOperationTypeBinary {
			return fmt.Errorf("binary patches are not supported for %q", firstNonEmpty(op.TargetPath, op.SourcePath))
		}
		if op.Type == gitdiffparser.PatchOperationTypeModeChange {
			return fmt.Errorf("mode-only patches are not supported for %q", firstNonEmpty(op.TargetPath, op.SourcePath))
		}
		if err := validatePatchOperationModes(op); err != nil {
			return err
		}
		for _, p := range []string{op.SourcePath, op.TargetPath} {
			if p == "" {
				continue
			}
			normalizedPath, err := g.validatePatchOutputPath(p)
			if err != nil {
				return err
			}
			if isProtectedPatchPath(normalizedPath) {
				return fmt.Errorf("patch references protected path %q", normalizedPath)
			}
		}
		if op.MutatesFileSet() {
			if err := g.validateFileSetMutation(op, virtualFiles); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) validateFileSetMutation(op gitdiffparser.PatchOperation, virtualFiles map[string]*merge.VirtualFile) error {
	switch op.Type {
	case gitdiffparser.PatchOperationTypeCreate, gitdiffparser.PatchOperationTypeCopy, gitdiffparser.PatchOperationTypeRename:
		if op.TargetPath != "" {
			_, exists := virtualFiles[op.TargetPath]
			if !exists && g.diskFileExists(op.TargetPath) && !g.isTrackedGeneratedFile(op.TargetPath) && !g.canCreateOverUntrackedFile(op) {
				return fmt.Errorf("patch would write %q, but a non-generated file already exists there; remove or rename that file before applying this patch", op.TargetPath)
			}
		}
	}

	if g.shouldDeferPatchWrite() {
		switch op.Type {
		case gitdiffparser.PatchOperationTypeDelete, gitdiffparser.PatchOperationTypeRename:
			if op.SourcePath != "" && g.isUserModifiedTrackedFile(op.SourcePath) {
				return fmt.Errorf("patch would remove user-modified file %q", op.SourcePath)
			}
		}
	}

	return nil
}

func (g *Generator) collectPatchVirtualFiles(virtualFiles map[string]*merge.VirtualFile, operations []gitdiffparser.PatchOperation) map[string]*merge.VirtualFile {
	files := map[string]*merge.VirtualFile{}
	for _, op := range operations {
		for _, p := range []string{op.SourcePath, op.TargetPath} {
			if p == "" {
				continue
			}
			if op.Type == gitdiffparser.PatchOperationTypeCreate && p == op.TargetPath {
				continue
			}
			if vf, ok := virtualFiles[p]; ok {
				files[p] = vf
			}
		}
	}
	return files
}

func (g *Generator) collectDiskBackedPatchTargets(virtualFiles map[string]*merge.VirtualFile, operations []gitdiffparser.PatchOperation) map[string]struct{} {
	paths := map[string]struct{}{}
	for _, op := range operations {
		if op.Type != gitdiffparser.PatchOperationTypeCreate || op.TargetPath == "" || !g.isUntrackedPatternFile(op.TargetPath) {
			continue
		}
		if _, exists := virtualFiles[op.TargetPath]; exists {
			continue
		}

		// Some target-managed files, like hook registration files, are generated
		// once and then intentionally untracked. A create patch for one of those
		// paths is an explicit overwrite so repeat generations stay idempotent
		// after the patched output has been committed.
		normalizedPath, err := g.validatePatchOutputPath(op.TargetPath)
		if err != nil {
			continue
		}
		paths[normalizedPath] = struct{}{}
	}
	return paths
}

func (g *Generator) canCreateOverUntrackedFile(op gitdiffparser.PatchOperation) bool {
	return op.Type == gitdiffparser.PatchOperationTypeCreate && g.isUntrackedPatternFile(op.TargetPath)
}

func (g *Generator) isUntrackedPatternFile(p string) bool {
	if g.subsystem == nil || g.subsystem.FileTracker == nil {
		return false
	}

	normalizedPath, err := g.validatePatchOutputPath(p)
	if err != nil {
		return false
	}
	return g.subsystem.FileTracker.IsUntrackedPatternFile(normalizedPath)
}

func (g *Generator) trackPatchedCreatedFile(path string, content []byte) error {
	if g.subsystem == nil || g.subsystem.FileTracker == nil {
		return nil
	}
	return g.subsystem.FileTracker.TrackFileImmediate(g.trackingPathForVirtualFile(path), content)
}

func (g *Generator) untrackPatchedDeletedFile(path string) {
	if g.subsystem != nil && g.subsystem.FileTracker != nil {
		g.subsystem.FileTracker.UntrackFile(g.trackingPathForVirtualFile(path))
	}
}

func (g *Generator) diskFileExists(path string) bool {
	_, err := g.Stat(filepath.Join(g.outDir, path))
	return err == nil
}

func (g *Generator) isTrackedGeneratedFile(path string) bool {
	if g.lockFile == nil || g.lockFile.TrackedFiles == nil {
		return false
	}

	tracked, ok := g.lockFile.TrackedFiles.Get(g.trackingPathForVirtualFile(path))
	if !ok || tracked.Deleted {
		return false
	}
	return tracked.MovedTo == "" || tracked.MovedTo == path
}

func (g *Generator) isUserModifiedTrackedFile(path string) bool {
	if g.lockFile == nil || g.lockFile.TrackedFiles == nil {
		return false
	}

	trackingPath := g.trackingPathForVirtualFile(path)
	tracked, ok := g.lockFile.TrackedFiles.Get(trackingPath)
	if !ok || tracked.LastWriteChecksum == "" {
		return false
	}

	diskPath := path
	if tracked.MovedTo != "" {
		diskPath = tracked.MovedTo
	}

	content, err := g.ReadFile(filepath.Join(g.outDir, diskPath))
	if err != nil {
		return false
	}

	checksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(content))
	if err != nil {
		return false
	}
	return tracked.LastWriteChecksum != "sha1:"+checksum
}

func (g *Generator) normalizeAppliedPatchTree(tree map[string][]byte) (map[string][]byte, error) {
	normalized := make(map[string][]byte, len(tree))
	for p, content := range tree {
		normalizedPath, err := g.validatePatchOutputPath(p)
		if err != nil {
			return nil, err
		}
		if _, ok := normalized[normalizedPath]; ok {
			return nil, fmt.Errorf("patch contains multiple file entries that resolve to generated file %q; each patch file entry must target a distinct generated file", normalizedPath)
		}
		normalized[normalizedPath] = content
	}
	return normalized, nil
}

func (g *Generator) validatePatchOutputPath(p string) (string, error) {
	if strings.ContainsRune(p, 0) {
		return "", fmt.Errorf("patch path %q contains a NUL byte; patch files can only target generated files inside the output directory", p)
	}
	if isAbsolutePatchPath(p) {
		return "", fmt.Errorf("patch path %q is absolute; patch files must use paths relative to the SDK output directory", p)
	}

	normalizedPath := normalizePatchLookupPath(p)
	if isAbsolutePatchPath(normalizedPath) {
		return "", fmt.Errorf("patch path %q is absolute; patch files must use paths relative to the SDK output directory", p)
	}

	absOutDir, err := filepath.Abs(g.outDir)
	if err != nil {
		return "", fmt.Errorf("resolving output directory for patch path %q: %w", p, err)
	}
	absTarget, err := filepath.Abs(filepath.Join(absOutDir, filepath.FromSlash(normalizedPath)))
	if err != nil {
		return "", fmt.Errorf("resolving patch path %q: %w", p, err)
	}
	rel, err := filepath.Rel(absOutDir, absTarget)
	if err != nil {
		return "", fmt.Errorf("checking patch path %q relative to output directory: %w", p, err)
	}
	if rel == "." {
		return "", fmt.Errorf("patch path %q resolves to the output directory; patch files must target a generated file inside the output directory", p)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("patch path %q resolves outside the output directory (%s); patch files can only target generated files inside %s", p, absTarget, absOutDir)
	}

	return filepath.ToSlash(rel), nil
}

func isAbsolutePatchPath(p string) bool {
	p = strings.ReplaceAll(p, `\`, `/`)
	return path.IsAbs(p) || filepath.IsAbs(p) || hasWindowsVolumeName(p)
}

func hasWindowsVolumeName(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	return (p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')
}

func isProtectedPatchPath(p string) bool {
	p = normalizePatchLookupPath(p)
	return strings.HasPrefix(p, ".speakeasy/") ||
		p == ".speakeasy" ||
		strings.HasPrefix(p, ".git/") ||
		p == ".git"
}

func patchMentionsLookupPath(patchData []byte, lookupPath string) bool {
	needle := normalizePatchLookupPath(lookupPath)
	matched := false
	_, _, _ = gitdiffparser.SignificantChange(string(patchData), func(fd *gitdiffparser.FileDiff, _ *gitdiffparser.ContentChange) (bool, string) {
		fromPath := strings.TrimPrefix(fd.FromFile, "a/")
		toPath := strings.TrimPrefix(fd.ToFile, "b/")
		if normalizePatchLookupPath(fromPath) == needle || normalizePatchLookupPath(toPath) == needle {
			matched = true
			return true, ""
		}
		return false, ""
	})

	return matched
}

func patchOperationsMutateFileSet(operations []gitdiffparser.PatchOperation) bool {
	for _, op := range operations {
		if op.MutatesFileSet() {
			return true
		}
	}
	return false
}

func patchModeForPath(operations []gitdiffparser.PatchOperation, path string) (fs.FileMode, error) {
	path = normalizePatchLookupPath(path)
	for i := len(operations) - 1; i >= 0; i-- {
		op := operations[i]
		if normalizePatchLookupPath(op.TargetPath) != path {
			continue
		}
		mode, ok, err := patchMode(op.NewMode)
		if err != nil {
			return 0, err
		}
		if ok {
			return mode, nil
		}
	}
	return 0, nil
}

func validatePatchOperationModes(op gitdiffparser.PatchOperation) error {
	path := firstNonEmpty(op.TargetPath, op.SourcePath)
	for _, mode := range []string{op.OldMode, op.NewMode, op.IndexMode} {
		if _, _, err := patchMode(mode); err != nil {
			return fmt.Errorf("%w for %q", err, path)
		}
	}
	return nil
}

func patchMode(mode string) (fs.FileMode, bool, error) {
	if mode == "" {
		return 0, false, nil
	}

	originalMode := mode
	if len(mode) > 3 {
		if !strings.HasPrefix(mode, "100") {
			return 0, false, fmt.Errorf("unsupported git mode %q: only regular files are currently supported (modes 100644/100755)", originalMode)
		}
		mode = mode[len(mode)-3:]
	}
	if mode != "644" && mode != "755" {
		return 0, false, fmt.Errorf("unsupported git mode %q: only regular files are currently supported (modes 100644/100755)", originalMode)
	}

	perm, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, false, err
	}
	return fs.FileMode(perm), true, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
