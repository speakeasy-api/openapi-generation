package templates

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// Returns the supported MCP targets based on the templates directories.
func GetSupportedMCPTargets() []types.Target {
	targets := []types.Target{}

	for _, templateName := range templateNames() {
		if !IsSupportedMCPTemplateName(templateName) {
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

// Returns the supported MCP template names based on the templates directories.
func GetSupportedMCPTemplateNames() []string {
	supportedTemplateNames := []string{}

	for _, templateName := range templateNames() {
		if !IsSupportedMCPTemplateName(templateName) {
			continue
		}

		supportedTemplateNames = append(supportedTemplateNames, templateName)
	}

	return supportedTemplateNames
}

// Returns true if given template name is a supported MCP template.
//
// Checks include:
//   - Whether the template is outside MCP template naming.
//   - Whether the template name exists in the templates directory.
//   - Whether the template is hidden (i.e., does not have a "HIDDEN" file in
//     its directory).
func IsSupportedMCPTemplateName(templateName string) bool {
	if !isMCPTemplateName(templateName) {
		return false
	}

	if !templateExists(templateName) {
		return false
	}

	return !templateIsHidden(templateName)
}

// Returns true if the template name is a valid MCP template name with prefix of
// "mcp-".
func isMCPTemplateName(templateName string) bool {
	return strings.HasPrefix(templateName, "mcp-")
}
