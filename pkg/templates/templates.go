package templates

import (
	"path"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/templates"
)

func AllSupportedTargets() []types.Target {
	return slices.Concat(
		GetSupportedMCPTargets(),
		GetSupportedSDKTargets(),
		GetSupportedTerraformTargets(),
	)
}

func AllSupportedTemplateNames() []string {
	return slices.Concat(
		GetSupportedMCPTemplateNames(),
		GetSupportedSDKTemplateNames(),
		GetSupportedTerraformTemplateNames(),
	)
}

func IsSupportedTemplateName(templateName string) bool {
	return slices.Contains(AllSupportedTemplateNames(), templateName)
}

// Returns true if the template exists, by checking if the template directory
// can be read.
func templateExists(templateName string) bool {
	templatePath := templateDirectoryPath(templateName)

	_, err := templates.TemplateFS.ReadDir(templatePath)

	return err == nil
}

// Returns true if the template is hidden, by having a "HIDDEN" file in its
// template directory. If SPEAKEASY_DEBUG is set to "true", it will always
// return false, allowing versioned targets to be used in test environments.
func templateIsHidden(templateName string) bool {
	if env.IsDebug() {
		return false
	}

	hiddenFile := templateHiddenFilePath(templateName)

	_, err := templates.TemplateFS.ReadFile(hiddenFile)

	return err == nil
}

// Returns true if the template is sunset, by having a "SUNSET" file in its
// template directory.
func templateIsSunset(templateName string) bool {
	sunsetFile := templateSunsetFilePath(templateName)

	_, err := templates.TemplateFS.ReadFile(sunsetFile)

	return err == nil
}

// Returns all template names in the "templates" directory. It reads the
// directory and filters out entries that are not directories, returning only
// the names of directories that represent templates (minus "common").
func templateNames() []string {
	entries, err := templates.TemplateFS.ReadDir("templates")

	if err != nil {
		return []string{}
	}

	templateNames := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if entry.Name() == "common" {
			continue
		}

		templateNames = append(templateNames, entry.Name())
	}

	return templateNames
}

// Joins the "templates" directory with the given template name to create a
// path to the template directory. The path is explicitly constructed using
// path.Join to ensure it is platform-independent for Goja.
func templateDirectoryPath(templateName string) string {
	return path.Join("templates", templateName)
}

// Joins the template directory with "HIDDEN" file name. The path is explicitly
// constructed using path.Join to ensure it is platform-independent for Goja.
func templateHiddenFilePath(templateName string) string {
	return path.Join(templateDirectoryPath(templateName), "HIDDEN")
}

// Joins the template directory with "SUNSET" file name. The path is explicitly
// constructed using path.Join to ensure it is platform-independent for Goja.
func templateSunsetFilePath(templateName string) string {
	return path.Join(templateDirectoryPath(templateName), "SUNSET")
}
