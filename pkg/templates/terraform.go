package templates

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// Returns the supported Terraform targets based on the templates directories.
func GetSupportedTerraformTargets() []types.Target {
	targets := []types.Target{}

	for _, templateName := range templateNames() {
		if !IsSupportedTerraformTemplateName(templateName) {
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

// Returns the supported Terraform template names based on the templates directories.
func GetSupportedTerraformTemplateNames() []string {
	supportedTemplateNames := []string{}

	for _, templateName := range templateNames() {
		if !IsSupportedTerraformTemplateName(templateName) {
			continue
		}

		supportedTemplateNames = append(supportedTemplateNames, templateName)
	}

	return supportedTemplateNames
}

// Returns true if given template name is a supported Terraform template.
//
// Checks include:
//   - Whether the template is outside Terraform template naming.
//   - Whether the template name exists in the templates directory.
//   - Whether the template is hidden (i.e., does not have a "HIDDEN" file in
//     its directory).
func IsSupportedTerraformTemplateName(templateName string) bool {
	if !isTerraformTemplateName(templateName) {
		return false
	}

	if !templateExists(templateName) {
		return false
	}

	return !templateIsHidden(templateName)
}

// Returns true if the template name is a valid Terraform template name with prefix of
// "Terraform-".
func isTerraformTemplateName(templateName string) bool {
	return templateName == "terraform"
}
