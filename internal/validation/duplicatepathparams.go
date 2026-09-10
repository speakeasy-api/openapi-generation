package validation

import (
	"context"
	"fmt"
	"regexp"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type DuplicatePathParams struct{}

var _ Rule = (*DuplicatePathParams)(nil)

func (r *DuplicatePathParams) ID() string {
	return "generator-duplicate-path-params"
}

func (r *DuplicatePathParams) Category() string {
	return "validation"
}

func (r *DuplicatePathParams) Summary() string {
	return "Ensure no duplicated path parameters found."
}

func (r *DuplicatePathParams) HowToFix() string {
	return "Remove repeated path parameters or rename them so each URI template parameter appears only once."
}

func (r *DuplicatePathParams) Description() string {
	return "Check for duplicate path parameters in URI templates. Path parameters must be unique within the path template (e.g., /users/{id}/{id} is invalid)."
}

func (r *DuplicatePathParams) Link() string {
	return ""
}

func (r *DuplicatePathParams) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *DuplicatePathParams) Versions() []string {
	return nil // Applies to all versions
}

var (
	// Regex to extract path parameters from URI template
	// Matches patterns like {userId}, {;userId}, {?userId}, {.userId}, {userId*}
	pathParamExtractRegex = regexp.MustCompile(`(\{;?\??[\.a-zA-Z0-9_-]+\*?\})`)
	// Regex to strip formatting characters from parameter names
	pathParamStripRegex = regexp.MustCompile(`[{}?*;.]`)
)

func (r *DuplicatePathParams) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Use the pre-computed InlinePathItems index (path items from /paths)
	// These are the only ones with path templates in their location parent key
	for _, pathItemNode := range docInfo.Index.InlinePathItems {
		// Extract path template from location (parent key of the path item)
		pathTemplate := pathItemNode.Location.ParentKey()
		if pathTemplate == "" {
			continue
		}

		// Extract all path parameters from the URI template
		pathParams := make(map[string]bool)
		matches := pathParamExtractRegex.FindAllString(pathTemplate, -1)

		for _, match := range matches {
			// Strip off curly brackets and formatting characters to get parameter name
			paramName := pathParamStripRegex.ReplaceAllString(match, "")

			// Check if this parameter name was already used
			if pathParams[paramName] {
				// Get the path item node for error reporting
				var node *yaml.Node
				pathItemRef := pathItemNode.Node
				if pathItemRef != nil && pathItemRef.GetCore() != nil {
					node = pathItemRef.GetRootNode()
				}

				validationErrors = append(validationErrors, &validation.Error{
					Rule:     r.ID(),
					Severity: r.DefaultSeverity(),
					Node:     node,
					UnderlyingError: fmt.Errorf(
						"path `%s` must not use the parameter `%s` multiple times",
						pathTemplate,
						paramName,
					),
				})
				// Only report once per path template
				break
			}

			pathParams[paramName] = true
		}
	}

	return validationErrors
}
