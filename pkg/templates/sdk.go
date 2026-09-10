package templates

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

var SupportedSDKDocsLanguages = []string{"go", "python", "typescript", "csharp", "unity", "java", "curl"}

// Returns the supported SDK targets based on the templates directories.
func GetSupportedSDKTargets() []types.Target {
	targets := []types.Target{}

	for _, templateName := range templateNames() {
		if !IsSupportedSDKTemplateName(templateName) {
			continue
		}

		if templateIsSunset(templateName) {
			// Skip sunset templates
			if env.IsDebug() {
				continue
			}
		}

		target := types.NewTargetFromTemplate(templateName)
		targets = append(targets, target)
	}

	return targets
}

// Returns the supported SDK template names based on the templates directories.
func GetSupportedSDKTemplateNames() []string {
	supportedTemplateNames := []string{}

	for _, templateName := range templateNames() {
		if !IsSupportedSDKTemplateName(templateName) {
			continue
		}

		supportedTemplateNames = append(supportedTemplateNames, templateName)
	}

	return supportedTemplateNames
}

// Returns true if given template name is a supported SDK template.
//
// Checks include:
//   - Whether the template is outside SDK templates, such as "common" (used for
//     shared code), starts with "mcp-" (MCP templates), "terraform" (Terraform
//     template).
//   - Whether the template name exists in the templates directory.
//   - Whether the template is hidden (i.e., does not have a "HIDDEN" file in
//     its directory).
func IsSupportedSDKTemplateName(templateName string) bool {
	if !isSDKTemplateName(templateName) {
		return false
	}

	if !templateExists(templateName) {
		return false
	}

	return !templateIsHidden(templateName)
}

// Returns true if the template name is a valid SDK template name. Checks
// include if the template name is empty, "common" (shared templating code),
// Terraform, or other template groupings such as MCP.
func isSDKTemplateName(templateName string) bool {
	if templateName == "" || templateName == "common" {
		return false
	}

	// MCP templates are not considered SDK templates.
	if isMCPTemplateName(templateName) {
		return false
	}

	// Terraform template is not considered SDK template.
	if templateName == "terraform" {
		return false
	}

	return true
}
