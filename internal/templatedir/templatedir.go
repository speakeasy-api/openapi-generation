package templatedir

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/speakeasy-api/openapi-generation/v2/templates"
	"gopkg.in/yaml.v3"
)

const (
	TemplateRoot   = "templates"
	TemplateCommon = TemplateRoot + "/common"
	ConfigFile     = "template.yaml"
)

type templateYAML struct {
	BaseTemplate          string   `yaml:"baseTemplate"`
	AdditionalSearchPaths []string `yaml:"additionalSearchPaths"`
}

// Config holds the resolved template directory layout for a target.
// Three distinct modes supported, controlled by templates/<target>/template.yaml
//
//	0 Standalone: no template.yaml -> only the primary directory is used.
//
//	1. `baseTemplate: <base>` -> the primary directory is a sparse overlay:
//	  - both paths are included in file generation (*.stmpl) and template resolution (*.ts)
//	  - files present in the primary override base files with the same path
//	  - files absent from the primary fall through to the base directory
//	  For example csharpv2 (primary) is using csharp (base) as a fallback.
//
//	2. `additionalSearchLocation: <auxiliary>` -> the primary owns all generated files
//	  - auxiliary paths are only used for template resolution (*.ts)
//	  - template files (*.stmpl) are never injected into the generated output
//	  For example cli (primary) includes go (auxiliary) for resolution-only purpuses.
type Config struct {
	PrimaryTemplate       string
	BaseTemplate          string
	AdditionalSearchPaths []string
}

// LoadConfig reads an optional template.yaml from the primary template directory
// in the embedded FS and returns the resulting config. Missing file is not an error.
func LoadConfig(primaryTemplate string) (Config, error) {
	cfg := Config{PrimaryTemplate: primaryTemplate}

	primaryDir := path.Join(TemplateRoot, primaryTemplate)
	if _, err := fs.Stat(templates.TemplateFS, primaryDir); err != nil {
		return cfg, fmt.Errorf("template %q not found: %w", primaryTemplate, err)
	}

	data, _ := fs.ReadFile(templates.TemplateFS, path.Join(primaryDir, ConfigFile))
	if data == nil {
		return cfg, nil
	}

	var y templateYAML
	if err := yaml.Unmarshal(data, &y); err != nil {
		return cfg, err
	}

	cfg.BaseTemplate = y.BaseTemplate
	cfg.AdditionalSearchPaths = y.AdditionalSearchPaths

	if cfg.BaseTemplate != "" {
		baseDir := cfg.baseDir()
		if _, err := fs.Stat(templates.TemplateFS, baseDir); err != nil {
			return cfg, fmt.Errorf("base template %q not found: %w", cfg.BaseTemplate, err)
		}
	}

	for _, p := range cfg.AdditionalSearchPaths {
		dir := path.Join(TemplateRoot, p)
		if _, err := fs.Stat(templates.TemplateFS, dir); err != nil {
			return cfg, fmt.Errorf("additionalSearchPath %q not found: %w", p, err)
		}
	}

	return cfg, nil
}

func (c Config) primaryDir() string {
	return path.Join(TemplateRoot, c.PrimaryTemplate)
}

func (c Config) baseDir() string {
	if c.BaseTemplate == "" {
		return ""
	}
	return path.Join(TemplateRoot, c.BaseTemplate)
}

// SearchLocations returns the ordered list of template directories for this
// target. The primary dir (overrides) always comes first, followed by the base
// template dir (if any), then templates/common and "" (full-path fallback).
func (c Config) SearchLocations() []string {
	locs := []string{c.primaryDir()}
	if b := c.baseDir(); b != "" {
		locs = append(locs, b)
	}
	for _, p := range c.AdditionalSearchPaths {
		locs = append(locs, path.Join(TemplateRoot, p))
	}
	locs = append(locs, TemplateCommon, "")
	return locs
}

// DirectoryExists reports whether dir exists under the primary or base template dir.
func (c Config) DirectoryExists(dir string) bool {
	if directoryExists(c.primaryDir(), dir) {
		return true
	}
	if b := c.baseDir(); b != "" {
		return directoryExists(b, dir)
	}
	return false
}

// GetDirectoryFiles returns files under dir, merging primary and base template
// dirs. Base files load first; primary files load last and win on re-registration.
func (c Config) GetDirectoryFiles(dir string) ([]string, error) {
	primaryDir := c.primaryDir()
	baseDir := c.baseDir()

	primaryFiles, primaryErr := getDirectoryFiles(primaryDir, dir)
	if primaryErr != nil {
		if !errors.Is(primaryErr, fs.ErrNotExist) || baseDir == "" {
			return nil, primaryErr
		}
	}

	if baseDir == "" {
		return primaryFiles, nil
	}

	baseFiles, baseErr := getDirectoryFiles(baseDir, dir)
	if baseErr != nil {
		if errors.Is(baseErr, fs.ErrNotExist) {
			if primaryErr != nil {
				return nil, primaryErr
			}
			return primaryFiles, nil
		}
		return nil, baseErr
	}

	seen := make(map[string]struct{}, len(primaryFiles))
	for _, f := range primaryFiles {
		seen[f] = struct{}{}
	}
	merged := make([]string, 0, len(primaryFiles)+len(baseFiles))
	for _, f := range baseFiles {
		if _, ok := seen[f]; !ok {
			merged = append(merged, f)
		}
	}
	return append(merged, primaryFiles...), nil
}

// ResolveFileBase returns the base directory that contains file, trying
// primary first then the base template dir.
func (c Config) ResolveFileBase(file string) string {
	primaryDir := c.primaryDir()
	baseDir := c.baseDir()

	if _, err := fs.Stat(templates.TemplateFS, path.Join(primaryDir, file)); err == nil {
		return primaryDir
	}
	if baseDir != "" {
		return baseDir
	}
	return primaryDir
}

func directoryExists(baseDir, dir string) bool {
	_, err := fs.ReadDir(templates.TemplateFS, path.Join(baseDir, dir))
	return err == nil
}

func getDirectoryFiles(baseDir, dir string) ([]string, error) {
	paths := []string{}

	files, err := fs.ReadDir(templates.TemplateFS, path.Join(baseDir, dir))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			subPaths, err := getDirectoryFiles(baseDir, path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, path.Join(dir, file.Name()))
		}
	}

	return paths, nil
}
