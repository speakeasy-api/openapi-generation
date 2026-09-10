package filetracking

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Ignore encapsulates zero or more .genignore files and provides the
// ShouldIgnore method to check whether a given file matches the .genignore
// loaded files.
type Ignore struct {
	ignores  map[string]*gitignore.GitIgnore
	override map[string]bool
}

// NewIgnore creates an empty Ignore struct. You will need to call
// AddGenIgnoreFile to add .genignore files to it.
func NewIgnore() *Ignore {
	return &Ignore{
		ignores:  map[string]*gitignore.GitIgnore{},
		override: map[string]bool{},
	}
}

func (i *Ignore) HasIgnoredFiles() bool {
	return len(i.ignores) > 0
}

// AddGenIgnoreFile adds a .genignore file to the Ignore struct. The root
// parameter is the root directory of the project, and genFile is the path to
// the .genignore file relative to the root.
func (i *Ignore) AddGenIgnoreFile(root, genFile string) error {
	dirName := filepath.Dir(genFile)

	lines := []string{}
	lr, err := os.Open(filepath.Join(root, genFile))
	if err != nil {
		return err
	}
	defer func() { _ = lr.Close() }()

	scanner := bufio.NewScanner(lr)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	gi := gitignore.CompileIgnoreLines(lines...)
	if err != nil {
		return err
	}

	i.ignores[dirName] = gi

	return nil
}

// NewIgnoreFromRoot creates an Ignore struct by walking the given root
// directory and adding any .genignore files it finds.
func NewIgnoreFromRoot(root string) (*Ignore, error) {
	i := NewIgnore()
	root = filepath.Join(root, ".")

	if _, err := os.Stat(root); err != nil {
		//nolint:nilerr
		return i, nil
	}

	err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return skipInaccessibleDir(info, err)
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if filepath.Base(rel) == ".genignore" {
			err := i.AddGenIgnoreFile(root, rel)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return i, nil
}

// ShouldIgnore returns true if the given path should be ignored, false
// otherwise.
func (i *Ignore) ShouldIgnore(path string) bool {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")
	dirName := "."
	for {
		if gi, ok := i.ignores[dirName]; ok {
			if gi.MatchesPath(path) {
				return !i.shouldIgnoreOverride(path)
			}
		}

		if len(parts) == 0 {
			return false
		}

		dirName = filepath.Join(dirName, parts[0])
		parts = parts[1:]
	}
}

// Add a regex pattern that will not be ignored even if its present in the .genignore file
func (i *Ignore) AddIgnoreOverride(regex string) {
	i.override[regex] = true
}

func (i *Ignore) shouldIgnoreOverride(path string) bool {
	for regex := range i.override {
		if ok, _ := regexp.MatchString(regex, path); ok {
			return true
		}
	}

	return false
}
