package filetracking

import (
	"bytes"
	"context"
	stderrors "errors"
	"io/fs"
	"regexp"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
	"go.uber.org/zap"
)

type OnWriteFileFunc func(filename string, data []byte, perm fs.FileMode, checkExisting bool) error

// skipInaccessibleDir handles expected WalkDir races and permission errors
// while preserving other filesystem failures.
func skipInaccessibleDir(d fs.DirEntry, err error) error {
	if stderrors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if stderrors.Is(err, fs.ErrPermission) && d != nil && d.IsDir() {
		return fs.SkipDir
	}
	return err
}

// File patterns that should be untracked for all targets.
// The internal/sdk handling is for terraform inter-template generation of
// the go target.
var defaultUntrackedFiles = []*regexp.Regexp{
	regexp.MustCompile(`^(internal/sdk/)?CONTRIBUTING\.md`),
	regexp.MustCompile(`^README\.md`),
	regexp.MustCompile(`^(internal/sdk/)?\.gitignore`),
	regexp.MustCompile(`^comments\.yaml"`),
	regexp.MustCompile(`^../../testprojects/[^ ]+/.*`),
	regexp.MustCompile(`^../../testusages/[^ ]+/.*`),
}

type AddHeaderPattern struct {
	Pattern   *regexp.Regexp
	DontMatch bool
}

// fileWithContent pairs a file path with its content for integrity computation
type fileWithContent struct {
	path string
	data []byte
	// untrack removes the path from the tracked set instead of adding it.
	// Routed through the same channel as track events so a queued TrackFile
	// for the same path cannot land after the removal.
	untrack bool
}

// Tracker keeps track of all the files created either by generating file content
// by executing a template or by copying a file. Tracked files are then written to
// a file called files.gen in the root of the output directory.
type Tracker struct {
	// List of file paths to track. These are automatically deleted at the
	// beginning of subsequent target generations unless explicitly ignored with
	// .genignore files or added to UntrackedPatterns.
	filesChan chan fileWithContent
	wg        sync.WaitGroup // ensures background goroutine completes before GetResult returns

	newFiles   config.TrackedFiles
	newFilesMu sync.Mutex

	// Manages cleanup of orphaned files from interrupted runs or crashes.
	//
	// Two-phase tracking:
	// 1. gen.lock - Final generatedFiles list from last SUCCESSFUL run. Cleanup uses:
	//    delete(previousFiles - newFiles)
	// 2. Tracker.lockFile (tmp) - Tracks IN-PROGRESS writes. Updated BEFORE each file creation.
	//
	// On startup: If tmp Tracker.lockFile exists → merge with gen.lock's previousFiles
	// to find ALL orphans (gen.lock incomplete = use tmp records).
	// On success: Write gen.lock + delete tmp Tracker.lockFile.
	// On failure: tmp Tracker.lockFile remains → next run handles cleanup.
	lockFile      *lockFile
	previousFiles []string

	// List of file path regular expressions that should not be tracked.
	// Automatically instantiated with files such as README.md, .gitignore,
	// documentation theme files, and others in [NewTracker]. Targets can
	// dynamically add to these via AddUntrackedPattern.
	UntrackedPatterns   []*regexp.Regexp
	untrackedPatternsMu sync.Mutex

	// List of file path regular expressions that should not have a header.
	AddHeaderPatterns []*AddHeaderPattern

	isClosed bool

	Options Options
}

type Options struct {
	OutDir    string
	GenLockId string
	FS        filesystem.FileSystem
	// SkipTempLockFile skips writing to the temporary lock file during generation.
	// When persistent edits is enabled, files are collected in memory and written
	// atomically at the end, so the incremental crash-recovery lock file is unnecessary.
	// We still read any existing lock file for orphan cleanup from previous crashes.
	SkipTempLockFile bool
	// SkipChecksums skips computing and recording last_write_checksum for tracked
	// files. When persistent edits is explicitly disabled (enabled: never) nothing
	// consumes the checksums, and omitting them keeps gen.lock stable across
	// regenerations so parallel branches don't merge-conflict on checksum churn.
	SkipChecksums bool
}

// NewTracker creates a new Tracker.
func NewTracker(ctx context.Context, opts Options) *Tracker {
	lockFile, previousFiles, err := newLockFile(opts)
	if err != nil {
		logging.From(ctx).Error("Failed to parse lock file", zap.Error(err))
	}

	filesChan := make(chan fileWithContent)
	newFiles := sequencedmap.New[string, config.TrackedFile]()

	t := &Tracker{
		UntrackedPatterns: defaultUntrackedFiles,
		filesChan:         filesChan,
		lockFile:          lockFile,
		previousFiles:     previousFiles,
		newFiles:          newFiles,
		Options:           opts,
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for file := range filesChan {
			if file.untrack {
				t.newFilesMu.Lock()
				t.newFiles.Delete(file.path)
				t.newFilesMu.Unlock()
				continue
			}
			checksum := ""
			if !opts.SkipChecksums {
				var err error
				checksum, err = lockfile.HashNormalizedSHA1(bytes.NewReader(file.data))
				if err != nil {
					logging.From(ctx).Error("Failed to compute checksum hash", zap.String("file", file.path), zap.Error(err))
					checksum = ""
				} else {
					checksum = "sha1:" + checksum
				}
			}

			t.newFilesMu.Lock()
			t.newFiles.Set(file.path, config.TrackedFile{LastWriteChecksum: checksum})
			t.newFilesMu.Unlock()

			// Skip writing to lock file when persistent edits is enabled (atomic writes)
			if !opts.SkipTempLockFile {
				if err := t.lockFile.WriteLine(file.path); err != nil {
					logging.From(ctx).Error("Failed to write to file tracker lock file", zap.Error(err))
				}
			}
		}
	}()

	return t
}

// AddUntrackedPattern appends a compiled regex to the untracked patterns list.
// It is safe for concurrent use.
func (t *Tracker) AddUntrackedPattern(pattern *regexp.Regexp) {
	t.untrackedPatternsMu.Lock()
	defer t.untrackedPatternsMu.Unlock()
	t.UntrackedPatterns = append(t.UntrackedPatterns, pattern)
}

// IsUntrackedPatternFile returns true if the file matches any regular
// expressions in the Tracker UntrackedPatterns.
func (t *Tracker) IsUntrackedPatternFile(file string) bool {
	t.untrackedPatternsMu.Lock()
	defer t.untrackedPatternsMu.Unlock()
	for _, pattern := range t.UntrackedPatterns {
		if pattern.MatchString(file) {
			return true
		}
	}

	return false
}

// AddHeaderToFile returns true if the file requires a header.
func (t *Tracker) AddHeaderToFile(file string) bool {
	addHeader := false

	for _, nhp := range t.AddHeaderPatterns {
		if nhp.Pattern.MatchString(file) {
			addHeader = !nhp.DontMatch
		}
	}

	return addHeader
}

// TrackFile is called for each generated file that needs to be tracked in the
// generatedFiles section in the gen.lock file. Some special file names are ignored.
func (t *Tracker) TrackFile(file string, data []byte) {
	if t.IsUntrackedPatternFile(file) {
		return
	}

	if t.isClosed {
		panic("File tracker has already been closed")
	}

	if t.Options.SkipChecksums {
		// The consumer goroutine skips hashing; avoid holding file content.
		data = nil
	}

	t.filesChan <- fileWithContent{path: file, data: data}
}

// TrackFileImmediate records a generated file outside the normal template
// tracking channel.
func (t *Tracker) TrackFileImmediate(file string, data []byte) error {
	if t.IsUntrackedPatternFile(file) {
		return nil
	}

	checksum := ""
	if !t.Options.SkipChecksums {
		hash, err := lockfile.HashNormalizedSHA1(bytes.NewReader(data))
		if err != nil {
			return err
		}
		checksum = "sha1:" + hash
	}

	t.newFilesMu.Lock()
	t.newFiles.Set(file, config.TrackedFile{LastWriteChecksum: checksum})
	t.newFilesMu.Unlock()

	if !t.Options.SkipTempLockFile {
		if err := t.lockFile.WriteLine(file); err != nil {
			return err
		}
	}

	return nil
}

type FileTrackerResult struct {
	NewFiles      config.TrackedFiles
	PreviousFiles []string
}

func (t *Tracker) GetResult() FileTrackerResult {
	if t.filesChan != nil {
		close(t.filesChan)
		t.filesChan = nil
	}
	// Wait for the background goroutine to finish processing all files
	t.wg.Wait()
	return FileTrackerResult{NewFiles: t.newFiles, PreviousFiles: t.previousFiles}
}

func (t *Tracker) Close(ctx context.Context) {
	if t.isClosed {
		return
	}
	err := t.lockFile.Destroy()
	if err != nil {
		logging.From(ctx).Error("Failed to cleanup file tracking lock file", zap.Error(err))
	}
	t.isClosed = true
}

// UpdateTrackedFile updates a tracked file entry with ID and PristineGitObject metadata.
// This is used by the patches subsystem after the 3-way merge completes.
func (t *Tracker) UpdateTrackedFile(path string, id, pristineGitObject string) {
	t.newFilesMu.Lock()
	defer t.newFilesMu.Unlock()

	if existing, ok := t.newFiles.Get(path); ok {
		existing.ID = id
		existing.PristineGitObject = pristineGitObject
		t.newFiles.Set(path, existing)
	} else if !t.IsUntrackedPatternFile(path) {
		// File not yet tracked - create entry with the metadata
		t.newFiles.Set(path, config.TrackedFile{
			ID:                id,
			PristineGitObject: pristineGitObject,
		})
	}
}

// UpdateTrackedFileChecksum updates the checksum for a tracked file.
// This is used after compilation modifies files to ensure gen.lock
// checksums match the post-compile file content.
func (t *Tracker) UpdateTrackedFileChecksum(path string, data []byte) {
	if t.IsUntrackedPatternFile(path) || t.Options.SkipChecksums {
		return
	}

	checksum, err := lockfile.HashNormalizedSHA1(bytes.NewReader(data))
	if err != nil {
		return
	}
	checksum = "sha1:" + checksum

	t.newFilesMu.Lock()
	defer t.newFilesMu.Unlock()

	if existing, ok := t.newFiles.Get(path); ok {
		existing.LastWriteChecksum = checksum
		t.newFiles.Set(path, existing)
	} else {
		t.newFiles.Set(path, config.TrackedFile{LastWriteChecksum: checksum})
	}
}

// UntrackFile removes a generated file from the current run's tracked set.
// This is used when a patch file deletes a virtual file after templates have
// already reported it to the tracker. The removal goes through the tracking
// channel so it is ordered after any in-flight TrackFile for the same path;
// once the channel is drained (GetResult) it falls back to a direct delete.
func (t *Tracker) UntrackFile(path string) {
	if t.IsUntrackedPatternFile(path) {
		return
	}

	if t.filesChan != nil {
		t.filesChan <- fileWithContent{path: path, untrack: true}
		return
	}

	t.newFilesMu.Lock()
	defer t.newFilesMu.Unlock()

	t.newFiles.Delete(path)
}
