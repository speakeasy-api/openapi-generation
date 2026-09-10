package merge

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
)

// fullPath joins the OutDir with a relative path.
// If OutDir is empty, returns the path as-is.
func fullPath(ctx *MergeContext, path string) string {
	if ctx.OutDir == "" {
		return path
	}
	return filepath.Join(ctx.OutDir, path)
}

// VirtualFile represents a file to be generated, held in memory before writing.
type VirtualFile struct {
	Path     string      // Relative path (e.g., "pkg/models/user.go")
	Content  []byte      // File content
	Mode     fs.FileMode // File mode (e.g., 0644 or 0755)
	IsBinary bool        // True if binary file (.png, .jar, etc.)
}

// MergeContext contains all the context needed to perform a 3-way merge
type MergeContext struct {
	// Generated files from template rendering
	VirtualFiles map[string]*VirtualFile

	// TrackedFiles contains the files that should be included in the tree hash
	// for no-op detection. This comes from the FileTracker and excludes untracked
	// files like CONTRIBUTING.md and README.md, ensuring stability across runs.
	TrackedFiles config.TrackedFiles

	// Current lockfile state
	LockFile *config.LockFile

	// Interface implementations
	FileSystem filesystem.FileSystem
	Git        Git

	// Enabled
	Enabled bool

	// OutDir is the output directory for this target.
	// All file operations are relative to this directory.
	OutDir string

	// PathRemapping maps computed paths to actual paths on disk.
	// When a user moves a file from path A to path B, and we generate for path A,
	// PathRemapping[A] = B tells us to look for the file at B instead.
	PathRemapping map[string]string
}

// MergeResult contains the result of the merge operation
type MergeResult struct {
	// The tree hash of the generated content (for no-op detection)
	TreeHash string

	// The commit hash (if a commit was created)
	CommitHash string

	// The new generation ID
	GenerationID string

	// Whether this was a no-op (no changes needed)
	IsNoOp bool

	// List of files that had conflicts
	ConflictFiles []string
}

// Merge performs the complete 3-way merge algorithm:
// Step 1: Immediate Blobbing - Write all generated content to Git object database
// Step 2: Determinism Check - Compare tree hash to detect no-op runs
// Step 3: 3-Way Merge Loop - For each file, perform dirty check and merge if needed
// Step 4: Commit & Push - Create commit, update lockfile, push snapshot
func Merge(ctx *MergeContext) (*MergeResult, error) {
	if ctx.Git == nil {
		return nil, errors.New("git interface is nil")
	}
	if len(ctx.VirtualFiles) == 0 {
		return nil, errors.New("no virtual files provided")
	}

	result := &MergeResult{
		ConflictFiles: []string{},
	}

	// ========== STEP 1: Immediate Blobbing ==========
	blobHashes := make(map[string]string) // path -> blob hash
	for path, vFile := range ctx.VirtualFiles {
		hash, err := ctx.Git.WriteObject(vFile.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to blob file %s: %w", path, err)
		}
		blobHashes[path] = hash
	}

	// ========== STEP 2: Determinism Check (No-Op Detection) ==========
	// Build complete tree by merging new blobs with unchanged files from previous generation.
	// Only include files that are in TrackedFiles to ensure stability - untracked files
	// like CONTRIBUTING.md and README.md are excluded since the generator may skip them
	// on subsequent runs when they already exist on disk.
	completeTreeHashes := make(map[string]string, len(blobHashes))
	for path, hash := range blobHashes {
		// Only include files that will be tracked in the lockfile
		if ctx.TrackedFiles != nil {
			if _, isTracked := ctx.TrackedFiles.Get(path); !isTracked {
				continue
			}
		}
		completeTreeHashes[path] = hash
	}

	// Reuse previous hashes for unchanged files still in the intended set.
	if ctx.TrackedFiles != nil {
		for path := range ctx.TrackedFiles.Keys() {
			if _, exists := completeTreeHashes[path]; !exists {
				if tracked, ok := ctx.LockFile.TrackedFiles.Get(path); ok && tracked.PristineGitObject != "" {
					completeTreeHashes[path] = tracked.PristineGitObject
				}
			}
		}
	}

	treeHash, err := ctx.Git.CreateSnapshotTree(completeTreeHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot tree: %w", err)
	}
	result.TreeHash = treeHash

	// Check if this is a no-op run (same tree hash as previous generation)
	if ctx.LockFile.PersistentEdits != nil &&
		ctx.LockFile.PersistentEdits.PristineTreeHash == treeHash {
		result.IsNoOp = true
		// No-op: the lockfile already contains the correct state from the previous run.
		// We don't modify TrackedFiles at all to preserve the complete state.
		// The caller should skip updating the lockfile when IsNoOp is true.
		return result, nil
	}

	// Snapshots stored, if not enabled we'll stop here.
	if !ctx.Enabled {
		for path, vFile := range ctx.VirtualFiles {
			_ = updateTracking(ctx, path, blobHashes[path], vFile.Content)
		}
		return result, nil
	}

	// ========== STEP 3: 3-Way Merge Loop ==========
	for path, vFile := range ctx.VirtualFiles {
		hasConflict, err := mergeFile(ctx, path, vFile, blobHashes[path])
		if err != nil {
			return nil, fmt.Errorf("failed to merge file %s: %w", path, err)
		}
		if hasConflict {
			result.ConflictFiles = append(result.ConflictFiles, path)
		}
	}

	// ========== STEP 4: Commit & Push Snapshot ==========
	// This is handled by the caller who has access to UUID generation and commit messages

	return result, nil
}

// mergeFile performs the merge decision for a single file.
// Returns true if there was a conflict.
func mergeFile(ctx *MergeContext, path string, vFile *VirtualFile, newBlobHash string) (bool, error) {
	tracked, hasTracked := ctx.LockFile.TrackedFiles.Get(path)

	// If not found directly, check if this is a moved file destination
	// (the path might be the NEW location after user moved the file)
	trackingPath := path
	if !hasTracked {
		// Search for an entry where MovedTo points to this path
		for origPath := range ctx.LockFile.TrackedFiles.Keys() {
			if t, ok := ctx.LockFile.TrackedFiles.Get(origPath); ok && t.MovedTo == path {
				tracked = t
				hasTracked = true
				trackingPath = origPath // Use original path for tracking
				break
			}
		}
	}

	// Handle user-deleted files: if the user deleted a file, respect that and don't regenerate it
	if hasTracked && tracked.Deleted {
		// Update tracking to keep the blob hash but preserve deleted state
		tracked.PristineGitObject = newBlobHash
		ctx.LockFile.TrackedFiles.Set(trackingPath, tracked)
		return false, nil
	}

	// The pathInOutDir is the actual path on disk (which may be remapped by GetActualPath)
	// The trackingPath is the original computed path used as key in lockfile
	pathInOutDir := path
	if hasTracked && tracked.MovedTo != "" && tracked.MovedTo != path {
		// This case handles when we're processing the original path
		// and need to redirect to the new location
		pathInOutDir = tracked.MovedTo
	}

	// Case 1: New file (not in lockfile) - write directly
	if !hasTracked {
		return false, writeFile(ctx, pathInOutDir, trackingPath, vFile, newBlobHash, vFile.Content)
	}

	// Case 2: File exists in lockfile, check if user modified it
	diskContent, err := ctx.FileSystem.ReadFile(fullPath(ctx, pathInOutDir))
	if err != nil {
		// File doesn't exist on disk - write directly
		return false, writeFile(ctx, pathInOutDir, trackingPath, vFile, newBlobHash, vFile.Content)
	}

	// Compute hash of disk content
	diskChecksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(diskContent))
	if err != nil {
		return false, fmt.Errorf("failed to hash disk content: %w", err)
	}
	diskChecksum = "sha1:" + diskChecksum

	// If disk matches LastWriteChecksum, user hasn't modified it - write new version directly
	if tracked.LastWriteChecksum != "" && diskChecksum == tracked.LastWriteChecksum {
		return false, writeFile(ctx, pathInOutDir, trackingPath, vFile, newBlobHash, vFile.Content)
	}

	// Case 3: File is dirty (user modified it), need 3-way merge
	if tracked.PristineGitObject == "" {
		// No base available for merge - overwrite with new content
		return false, writeFile(ctx, pathInOutDir, trackingPath, vFile, newBlobHash, vFile.Content)
	}

	// Binary file check
	if vFile.IsBinary {
		// Keep user version for binary files, but update tracking
		return false, updateTracking(ctx, trackingPath, newBlobHash, diskContent)
	}

	// Retrieve base content from Git
	if !ctx.Git.HasObject(tracked.PristineGitObject) {
		if ctx.LockFile.PersistentEdits != nil {
			if err := ctx.Git.FetchSnapshot(ctx.LockFile.PersistentEdits.GenerationID); err != nil {
				return false, &SnapshotRefError{
					GenerationID: ctx.LockFile.PersistentEdits.GenerationID,
					BlobHash:     tracked.PristineGitObject,
					Cause:        err,
				}
			}

			if !ctx.Git.HasObject(tracked.PristineGitObject) {
				// Blob still missing after fetch attempts, snapshot is corrupt
				return false, &SnapshotRefError{
					GenerationID: ctx.LockFile.PersistentEdits.GenerationID,
					BlobHash:     tracked.PristineGitObject,
					Cause:        fmt.Errorf("blob %s still not found after fetching snapshot", tracked.PristineGitObject),
				}
			}
		}
	}

	var baseContent []byte
	if tracked.PristineGitObject != "" && ctx.Git.HasObject(tracked.PristineGitObject) {
		baseContent, err = ctx.Git.ReadBlob(tracked.PristineGitObject)
		if err != nil {
			baseContent = []byte{}
		}
	} else {
		baseContent = []byte{}
	}

	// Perform 3-way text merge
	mergedContent, hasConflict := merge3Way(baseContent, diskContent, vFile.Content)

	// Write merged content
	if err := ctx.FileSystem.WriteFile(fullPath(ctx, pathInOutDir), mergedContent, vFile.Mode); err != nil {
		return hasConflict, fmt.Errorf("failed to write merged file: %w", err)
	}

	// If there was a conflict, set up git's index for conflict resolution
	if hasConflict {
		isExecutable := vFile.Mode&0o111 != 0 // Check if any execute bit is set
		// Pass relative path - git_adapter.SetConflictState will prepend baseDir
		if err := ctx.Git.SetConflictState(pathInOutDir, baseContent, diskContent, vFile.Content, isExecutable); err != nil {
			// Log but don't fail - the conflict markers in the file are still useful
			// The git conflict state is supplementary for tools like mergetool
			_ = err
		}
	}

	return hasConflict, updateTracking(ctx, trackingPath, newBlobHash, mergedContent)
}

// writeFile writes a file and updates tracking
// diskPath is where to write the file on disk, trackingPath is the key for lockfile tracking
func writeFile(ctx *MergeContext, outDirPath, trackingPath string, vFile *VirtualFile, newBlobHash string, content []byte) error {
	if err := ctx.FileSystem.WriteFile(fullPath(ctx, outDirPath), content, vFile.Mode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return updateTracking(ctx, trackingPath, newBlobHash, content)
}

// updateTracking updates the lockfile tracking for a file
func updateTracking(ctx *MergeContext, path string, newBlobHash string, writtenContent []byte) error {
	checksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(writtenContent))
	if err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	tracked, exists := ctx.LockFile.TrackedFiles.Get(path)
	if !exists {
		tracked = lockfile.TrackedFile{}
	}

	// Update with new pristine blob hash and write checksum
	tracked.PristineGitObject = newBlobHash
	tracked.LastWriteChecksum = "sha1:" + checksum

	ctx.LockFile.TrackedFiles.Set(path, tracked)

	return nil
}

// merge3Way performs a 3-way merge of line-based text files using diffmatchpatch.
// It compares 'current' and 'newContent' against 'base'.
// Returns the merged content and a boolean indicating if conflicts were found.
func merge3Way(base, current, newContent []byte) ([]byte, bool) {
	// 1. Convert all inputs to line-based rune strings.
	// This allows us to use diffmatchpatch (which is char-based) to diff lines.
	baseStr := string(base)
	currStr := string(current)
	newStr := string(newContent)

	baseRunes, currRunes, newRunes, lineMap := linesToCharsMux(baseStr, currStr, newStr)

	// 2. Compute diffs relative to base
	dmp := diffmatchpatch.New()
	// We pass checkLines=false because we have already linearized the text.
	diffsBaseToCurr := dmp.DiffMain(baseRunes, currRunes, false)
	diffsBaseToNew := dmp.DiffMain(baseRunes, newRunes, false)

	// 3. Convert diffs to list of modifications
	modsCurr := diffsToModifications(diffsBaseToCurr)
	modsNew := diffsToModifications(diffsBaseToNew)

	// 4. Merge the modifications
	return mergeModifications(baseRunes, modsCurr, modsNew, lineMap)
}

// Merge3Way performs the same 3-way text merge used by the persistent-edits
// subsystem and returns conflict-marked content when overlaps are detected.
func Merge3Way(base, current, newContent []byte) ([]byte, bool) {
	return merge3Way(base, current, newContent)
}

// modification represents a change relative to the base string.
type modification struct {
	startBase int    // Start index in base (inclusive)
	endBase   int    // End index in base (exclusive)
	text      string // The replacement text (in runes)
}

// diffsToModifications flattens a Diff list into atomic modifications (replacements).
// It combines adjacent Delete and Insert operations into a single Replace operation
// to simplify conflict detection.
func diffsToModifications(diffs []diffmatchpatch.Diff) []modification {
	var mods []modification
	baseIdx := 0

	for i := 0; i < len(diffs); i++ {
		d := diffs[i]
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			baseIdx += utf8.RuneCountInString(d.Text)
		case diffmatchpatch.DiffDelete:
			// Check if next is insert (Replacement)
			endBase := baseIdx + utf8.RuneCountInString(d.Text)
			text := ""
			if i+1 < len(diffs) && diffs[i+1].Type == diffmatchpatch.DiffInsert {
				text = diffs[i+1].Text
				i++ // Consume next
			}
			mods = append(mods, modification{
				startBase: baseIdx,
				endBase:   endBase,
				text:      text,
			})
			baseIdx = endBase
		case diffmatchpatch.DiffInsert:
			mods = append(mods, modification{
				startBase: baseIdx,
				endBase:   baseIdx, // Insert has 0 width in base
				text:      d.Text,
			})
		}
	}
	return mods
}

// mergeModifications iterates through two sorted lists of modifications and merges them.
func mergeModifications(base string, modsA, modsB []modification, lineMap []string) ([]byte, bool) {
	var buf bytes.Buffer
	hasConflict := false

	// Indices for the two modification lists
	idxA, idxB := 0, 0
	// Index in the base string that we have processed up to
	baseIdx := 0
	baseLen := utf8.RuneCountInString(base)
	_ = baseLen // unused but kept for clarity

	// Helper to resolve runes to original strings
	runesToString := func(rStr string) string {
		var sb strings.Builder
		for _, r := range rStr {
			// Ensure rune is within bounds of lineMap
			if int(r) < len(lineMap) {
				sb.WriteString(lineMap[r])
			}
		}
		return sb.String()
	}

	for idxA < len(modsA) || idxB < len(modsB) {
		var modA *modification
		var modB *modification

		if idxA < len(modsA) {
			modA = &modsA[idxA]
		}
		if idxB < len(modsB) {
			modB = &modsB[idxB]
		}

		// Determine which modification comes next
		var nextMod *modification
		var isA bool

		switch {
		case modA != nil && modB != nil:
			if modA.startBase < modB.startBase {
				nextMod = modA
				isA = true
			} else {
				nextMod = modB
				isA = false
			}
		case modA != nil:
			nextMod = modA
			isA = true
		default:
			nextMod = modB
			isA = false
		}

		// Process clean text before this modification
		if nextMod.startBase > baseIdx {
			// Append base text from baseIdx to nextMod.startBase
			runes := []rune(base)
			if baseIdx < len(runes) {
				end := nextMod.startBase
				if end > len(runes) {
					end = len(runes)
				}
				buf.WriteString(runesToString(string(runes[baseIdx:end])))
			}
		}

		// Check for overlap/conflict
		var conflictA, conflictB *modification

		if modA != nil && modB != nil {
			// Check overlap ranges
			startA, endA := modA.startBase, modA.endBase
			startB, endB := modB.startBase, modB.endBase

			// Two intervals overlap if startA < endB && startB < endA
			// Also insert/insert conflict: if startA == startB (and lengths 0), they collide
			overlap := (startA < endB && startB < endA) || (startA == startB)

			if overlap {
				conflictA = modA
				conflictB = modB
			}
		}

		if conflictA != nil {
			// We have a conflict (or potential identical change)

			// 1. Check if changes are identical (Auto-merge)
			if conflictA.startBase == conflictB.startBase &&
				conflictA.endBase == conflictB.endBase &&
				conflictA.text == conflictB.text {
				// Identical change. Apply it once.
				buf.WriteString(runesToString(conflictA.text))
				baseIdx = conflictA.endBase
				idxA++
				idxB++
				continue
			}

			// 2. Real Conflict
			hasConflict = true

			// Determine the range in base that covers both modifications
			maxEnd := conflictA.endBase
			if conflictB.endBase > maxEnd {
				maxEnd = conflictB.endBase
			}

			buf.WriteString("<<<<<<< Current (Your changes)\n")
			buf.WriteString(runesToString(conflictA.text))
			buf.WriteString("=======\n")
			buf.WriteString(runesToString(conflictB.text))
			buf.WriteString(">>>>>>> New (Generated by Speakeasy)\n")

			baseIdx = maxEnd
			idxA++
			idxB++
		} else {
			// No overlap, apply the next mod
			if isA {
				buf.WriteString(runesToString(nextMod.text))
				baseIdx = nextMod.endBase
				idxA++
			} else {
				buf.WriteString(runesToString(nextMod.text))
				baseIdx = nextMod.endBase
				idxB++
			}
		}
	}

	// Append remaining base text
	runes := []rune(base)
	if baseIdx < len(runes) {
		buf.WriteString(runesToString(string(runes[baseIdx:])))
	}

	return buf.Bytes(), hasConflict
}

// linesToCharsMux takes three strings and converts them to rune strings based on unique lines.
// It returns the encoded strings and the slice mapping runes back to original lines.
// This is a 3-way extension of diffmatchpatch's DiffLinesToChars.
func linesToCharsMux(text1, text2, text3 string) (string, string, string, []string) {
	lineArray := []string{}
	lineHash := make(map[string]rune)

	// Helper to process a text
	process := func(text string) string {
		var chars strings.Builder
		// SplitAfter keeps the newline characters, ensuring reconstruction is perfect
		lines := splitAfter(text)
		for _, line := range lines {
			if r, ok := lineHash[line]; ok {
				chars.WriteRune(r)
			} else {
				// Create new ID
				r := rune(len(lineArray))
				lineArray = append(lineArray, line)
				lineHash[line] = r
				chars.WriteRune(r)
			}
		}
		return chars.String()
	}

	out1 := process(text1)
	out2 := process(text2)
	out3 := process(text3)
	return out1, out2, out3, lineArray
}

// splitAfter is a robust line splitter that behaves like strings.SplitAfter
// but ensures we don't get an empty string at the end if the file ends with newline.
func splitAfter(text string) []string {
	if text == "" {
		return nil
	}
	// strings.SplitAfter leaves an empty string at end if text ends with sep
	lines := strings.SplitAfter(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}
