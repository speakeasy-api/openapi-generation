package patchfiles

import (
	"errors"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"slices"
	"strings"
)

const patchesDir = ".speakeasy/patches"

// FS is the minimal filesystem interface needed by the patchfiles package.
type FS interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type PatchFile struct {
	Path string
	Data []byte
}

// PatchesDir returns the absolute path to the patches directory for the given output dir.
func PatchesDir(outDir string) string {
	return filepath.Join(outDir, patchesDir)
}

func HasAny(fsys FS, outDir string) (bool, error) {
	root := PatchesDir(outDir)
	if _, err := fsys.Stat(root); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return hasAnyIn(fsys, root)
}

func hasAnyIn(fsys FS, dir string) (bool, error) {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".patch") {
			return true, nil
		}
		if e.IsDir() {
			found, err := hasAnyIn(fsys, filepath.Join(dir, e.Name()))
			if err != nil {
				return false, err
			}
			if found {
				return true, nil
			}
		}
	}
	return false, nil
}

func PathFor(outDir, generatedPath string) string {
	return filepath.Join(outDir, RelPathFor(generatedPath))
}

// RelPathFor returns the patch path for a generated file relative to the
// output directory, for callers whose file access is already rooted there.
func RelPathFor(generatedPath string) string {
	normalized := filepath.FromSlash(normalizeGeneratedPath(generatedPath))
	return filepath.Join(patchesDir, normalized) + ".patch"
}

func Read(fsys FS, outDir, generatedPath string) ([]byte, bool, error) {
	patchFile := PathFor(outDir, generatedPath)
	data, err := fsys.ReadFile(patchFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

func ExistingPaths(fsys FS, outDir string) ([]string, error) {
	root := PatchesDir(outDir)
	if _, err := fsys.Stat(root); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	if err := collectPatchPaths(fsys, root, root, &paths); err != nil {
		return nil, err
	}
	slices.Sort(paths)
	return paths, nil
}

func collectPatchPaths(fsys FS, base, dir string, paths *[]string) error {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if err := collectPatchPaths(fsys, base, full, paths); err != nil {
				return err
			}
			continue
		}
		if !strings.HasSuffix(e.Name(), ".patch") {
			continue
		}
		rel, err := filepath.Rel(base, full)
		if err != nil {
			return err
		}
		*paths = append(*paths, normalizeGeneratedPath(strings.TrimSuffix(filepath.ToSlash(rel), ".patch")))
	}
	return nil
}

func ReadAll(fsys FS, outDir string) ([]PatchFile, error) {
	existingPaths, err := ExistingPaths(fsys, outDir)
	if err != nil {
		return nil, err
	}
	patches := make([]PatchFile, 0, len(existingPaths))
	for _, p := range existingPaths {
		data, used, err := Read(fsys, outDir, p)
		if err != nil {
			return nil, err
		}
		if !used {
			continue
		}
		patches = append(patches, PatchFile{
			Path: p,
			Data: data,
		})
	}
	return patches, nil
}

func normalizeGeneratedPath(path string) string {
	path = strings.ReplaceAll(path, `\`, `/`)
	return pathpkg.Clean(path)
}
