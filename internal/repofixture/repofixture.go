package repofixture

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

const ModulePath = "github.com/speakeasy-api/openapi-generation/v2"

var fixtureRoots = [...]string{"zSDKs", "testSDKs"}

func DetectFixtureOutput(outDir string) (bool, error) {
	path, err := filepath.Abs(outDir)
	if err != nil {
		return false, err
	}
	path, err = resolveExistingPrefix(path)
	if err != nil {
		return false, err
	}

	root, err := findCheckoutRoot(path)
	if err != nil || root == "" {
		return false, err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	first := strings.Split(filepath.ToSlash(rel), "/")[0]
	for _, fixtureRoot := range fixtureRoots {
		if first == fixtureRoot {
			return true, nil
		}
	}
	return false, nil
}

func resolveExistingPrefix(path string) (string, error) {
	existing := path
	var pending []string
	for {
		_, err := os.Lstat(existing)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return path, nil
		}
		pending = append([]string{filepath.Base(existing)}, pending...)
		existing = parent
	}
	resolved, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{resolved}, pending...)...), nil
}

func findCheckoutRoot(dir string) (string, error) {
	for {
		matches, err := goModDeclares(filepath.Join(dir, "go.mod"), ModulePath)
		if err != nil {
			return "", err
		}
		if matches {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func goModDeclares(path, modulePath string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return modfile.ModulePath(content) == modulePath, nil
}
