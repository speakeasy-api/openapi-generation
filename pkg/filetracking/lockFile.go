package filetracking

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
)

type lockFile struct {
	file        filesystem.File
	path        string
	isDestroyed bool

	// We use our filesystem abstraction so that the filetracking/lockfile
	// behaviour will still work (even if it might be noop) in environments
	// such as WASM where syscalls are not supported.
	fs filesystem.FileSystem
}

func newLockFile(opts Options) (*lockFile, []string, error) {
	// Avoid conflicts with generators running in different directories or concurrent targets
	tempDir := filepath.Join(opts.OutDir, ".speakeasy")
	lockPath := filepath.Join(tempDir, fmt.Sprintf("generated-files-%s.lock", opts.GenLockId))

	// Defer creating the file until we need to write to it, pathways such as linting could leave behind dangling lock files
	file, err := opts.FS.OpenFile(lockPath, os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		if os.IsNotExist(err) {
			return &lockFile{path: lockPath, fs: opts.FS}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	return &lockFile{file: file, path: lockPath, fs: opts.FS}, files, nil
}

func (lf *lockFile) WriteLine(line string) error {
	if lf.isDestroyed {
		return fmt.Errorf("file tracking lock file %s has already been destroyed", lf.path)
	}

	// Defer creating the file until we need to write to it, pathways such as linting could leave behind dangling lock files
	if lf.file == nil {
		tempDir := filepath.Dir(lf.path)
		if err := lf.fs.MkdirAll(tempDir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		file, err := lf.fs.OpenFile(lf.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("failed to create or open lock file: %w", err)
		}
		lf.file = file
	}

	if _, err := fmt.Fprintln(lf.file, line); err != nil {
		return fmt.Errorf("failed to write to lock file %s: %w", lf.path, err)
	}
	return nil
}

func (lf *lockFile) Destroy() error {
	if lf.file == nil {
		return nil
	}

	if lf.isDestroyed {
		return nil
	}

	var errs []error

	if err := lf.file.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing file: %w", err))
	}

	if err := lf.fs.Remove(lf.path); err != nil {
		errs = append(errs, fmt.Errorf("removing file: %w", err))
	}

	lf.isDestroyed = true
	lf.file = nil

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
