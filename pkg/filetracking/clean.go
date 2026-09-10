package filetracking

import (
	"context"
	stderrors "errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// CleanOrphanedFiles removes any previously generated files that are no longer
// being generated in a subsequent generation run. These orphaned files cannot
// stay behind because they can accidentally break builds or augment an SDK's
// behavior or public interface in unexpected ways.
func CleanOrphanedFiles(
	ctx context.Context,
	outDir string,
	newFiles []string,
	previousFiles []string,
	untrackedPatterns []*regexp.Regexp,
	ignore *Ignore,
	onRemovedFile func(string),
) error {
	logger := logging.From(ctx)

	lookup := make(map[string]struct{}, len(newFiles))

	for _, file := range newFiles {
		lookup[NormalizePath(file)] = struct{}{}
	}

	for _, file := range previousFiles {
		if err := ctx.Err(); err != nil {
			// abort cleanup if context is cancelled
			return err
		}

		file = NormalizePath(file)
		if _, found := lookup[file]; !found {
			ignored := ignore.ShouldIgnore(file)
			untracked := slices.ContainsFunc(untrackedPatterns, func(p *regexp.Regexp) bool {
				return p.MatchString(file)
			})

			if ignored || untracked {
				logger.Debug("Skipping removal of ignored or untracked file: " + file)
				continue
			}

			logger.Debug("Removing orphaned file: " + file)

			if err := os.Remove(filepath.Join(outDir, file)); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("%s: %w", file, err)
				}
			}
			onRemovedFile(file)
		}
	}

	return nil
}

// NormalizePath converts a template output path or lockfile entry to the form
// the tracker records: forward slashes, cleaned, relative to the output
// directory. Older lockfiles kept template paths verbatim ("/models/pet.go",
// "./models/pet.go", "models//pet.go", "models\\pet.go"); all of them must
// match the same tracked file or cleanup removes it right after generation.
func NormalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = path.Clean(p)
	p = strings.TrimLeft(p, "/")
	if p == "." {
		return ""
	}
	return p
}

// CleanOrphanedFilesByDirWalk removes any file under subDir (relative to outDir)
// that is not present in trackedFiles, ignoring matches against .genignore or
// untrackedPatterns. Paths in trackedFiles and the keys passed to ignore must
// be relative to outDir using forward slashes.
//
// skipSubDirs lists directory paths relative to subDir that should not be
// traversed at all. Useful for subtrees populated by direct file copies that
// bypass the file tracker, where every entry would otherwise look like an
// orphan.
//
// Unlike CleanOrphanedFiles, this is anchored to disk state rather than a prior
// gen.lock, so it removes files that were never recorded in gen.lock for any
// reason. Callers should only invoke this for subtrees that are fully owned by
// the generator (e.g. generated test infrastructure).
func CleanOrphanedFilesByDirWalk(
	ctx context.Context,
	outDir string,
	subDir string,
	trackedFiles map[string]struct{},
	skipSubDirs []string,
	untrackedPatterns []*regexp.Regexp,
	ignore *Ignore,
	onRemovedFile func(string),
) error {
	logger := logging.From(ctx)

	root := filepath.Join(outDir, subDir)
	if _, err := os.Stat(root); err != nil {
		if stderrors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	skipLookup := make(map[string]struct{}, len(skipSubDirs))
	for _, s := range skipSubDirs {
		skipLookup[filepath.ToSlash(filepath.Join(subDir, s))] = struct{}{}
	}

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return skipInaccessibleDir(d, err)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		rel, err := filepath.Rel(outDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if _, skip := skipLookup[rel]; skip {
				return filepath.SkipDir
			}
			return nil
		}

		if _, ok := trackedFiles[rel]; ok {
			return nil
		}

		if ignore != nil && ignore.ShouldIgnore(rel) {
			logger.Debug("Skipping removal of ignored or untracked file: " + rel)
			return nil
		}

		untracked := slices.ContainsFunc(untrackedPatterns, func(p *regexp.Regexp) bool {
			return p.MatchString(rel)
		})
		if untracked {
			logger.Debug("Skipping removal of ignored or untracked file: " + rel)
			return nil
		}

		logger.Debug("Removing orphaned file: " + rel)

		if removeErr := os.Remove(path); removeErr != nil && !stderrors.Is(removeErr, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", rel, removeErr)
		}
		onRemovedFile(rel)
		return nil
	}); err != nil {
		return err
	}

	return removeEmptyDirs(ctx, outDir, root, skipLookup)
}

// removeEmptyDirs walks the tree under root bottom-up and removes any directory
// that is empty after the file pass. The root itself and any directory listed
// in skipLookup (and their contents) are preserved — those subtrees were also
// skipped by the file pass, so we mustn't touch their structure here either.
//
// skipLookup keys are forward-slash paths relative to outDir, matching the
// rel form computed for each visited path so the lookups line up.
func removeEmptyDirs(ctx context.Context, outDir, root string, skipLookup map[string]struct{}) error {
	dirs := []string{}
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return skipInaccessibleDir(d, err)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if !d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(outDir, path)
		if err != nil {
			return err
		}
		if _, skip := skipLookup[filepath.ToSlash(rel)]; skip {
			return filepath.SkipDir
		}
		dirs = append(dirs, path)
		return nil
	}); err != nil {
		return err
	}

	// Bottom-up so a parent gets a chance to be empty after its children are removed.
	for i := len(dirs) - 1; i >= 0; i-- {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		dir := dirs[i]
		if dir == root {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if stderrors.Is(err, fs.ErrNotExist) {
				continue
			}
			// leave inaccessible dirs untouched instead of aborting
			if stderrors.Is(err, fs.ErrPermission) {
				continue
			}
			return err
		}
		if len(entries) > 0 {
			continue
		}
		if removeErr := os.Remove(dir); removeErr != nil && !stderrors.Is(removeErr, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", dir, removeErr)
		}
	}
	return nil
}
