// Borrowed from https://github.com/daveshanley/vacuum/blob/main/functions/openapi/path_parameters.go
package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

var strRx = regexp.MustCompile(`[{}?*;]`)
var paramRegex = regexp.MustCompile(`(\{;?\??[\.a-zA-Z0-9_-]+\*?\})`)

type PathParams struct{}

var _ Rule = (*PathParams)(nil)

func (p *PathParams) ID() string {
	return "generator-path-params"
}

func (p *PathParams) Category() string {
	return "validation"
}

func (p *PathParams) Summary() string {
	return "Ensure path parameters match and appear in the path template."
}

func (p *PathParams) HowToFix() string {
	return "Define all path template parameters in the operation or path parameters list, and remove parameters that are not used in the path."
}

func (p *PathParams) Description() string {
	return "Path parameters must be defined on the operation and used in the path template. This prevents mismatches between URI templates and parameter lists."
}

func (p *PathParams) Link() string {
	return ""
}

func (p *PathParams) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (p *PathParams) Versions() []string {
	return nil // Applies to all versions
}

func (p *PathParams) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Check OpenAPI version to know if we should check additionalOperations (3.2+)
	isOpenAPI32OrLater := docInfo.Document.OpenAPI != "" && strings.HasPrefix(docInfo.Document.OpenAPI, "3.2")

	// Track paths for equivalence checking
	seenPaths := make(map[string]string) // normalized path -> original path

	// Iterate through all path items
	for _, pathItemNode := range docInfo.Index.InlinePathItems {
		// Skip webhooks and callbacks - only validate actual paths
		location := pathItemNode.Location.ToJSONPointer().String()
		if strings.HasPrefix(location, "/webhooks/") || strings.Contains(location, "/callbacks/") {
			continue
		}

		pathStr := pathItemNode.Location.ParentKey()
		if pathStr == "" {
			continue
		}

		pathItemRef := pathItemNode.Node
		if pathItemRef == nil {
			continue
		}

		pathItem := pathItemRef.GetObject()
		if pathItem == nil {
			continue
		}

		// Check for equivalent paths
		normalizedPath := paramRegex.ReplaceAllString(pathStr, "%")
		if existingPath, exists := seenPaths[normalizedPath]; exists {
			var node *yaml.Node
			if pathItemRef.GetCore() != nil {
				node = pathItemRef.GetRootNode()
			}
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            p.ID(),
				Severity:        validation.SeverityWarning,
				Node:            node,
				UnderlyingError: fmt.Errorf("paths `%s` and `%s` must not be equivalent, paths must be unique", existingPath, pathStr),
			})
		} else {
			seenPaths[normalizedPath] = pathStr
		}

		// Extract expected path parameters from URL
		expectedParams := make(map[string]bool)
		for _, match := range paramRegex.FindAllString(pathStr, -1) {
			paramName := strRx.ReplaceAllString(match, "")
			expectedParams[paramName] = true
		}

		// Collect path-level parameters
		pathLevelParams := make(map[string]*openapi.Parameter)
		for _, paramRef := range pathItem.GetParameters() {
			param := paramRef.GetObject()
			if param == nil || param.GetIn() != "path" {
				continue
			}

			pathLevelParams[param.GetName()] = param
		}

		// Process each operation in this path
		// PathItem embeds *sequencedmap.Map[HTTPMethod, *Operation]
		for httpMethod, operation := range pathItem.All() {
			if operation == nil {
				continue
			}

			method := strings.ToUpper(httpMethod.String())

			// Collect operation-level path parameters
			opParams := make(map[string]*openapi.Parameter)
			for _, paramRef := range operation.GetParameters() {
				param := paramRef.GetObject()
				if param == nil || param.GetIn() != "path" {
					continue
				}

				opParams[param.GetName()] = param
			}

			// Combine path-level and operation-level parameters
			allDefinedParams := make(map[string]bool)
			for name := range pathLevelParams {
				allDefinedParams[name] = true
			}
			for name := range opParams {
				allDefinedParams[name] = true
			}

			// Get node for error reporting
			opNode := operation.GetRootNode()

			// Check that all expected parameters are defined
			for paramName := range expectedParams {
				if !allDefinedParams[paramName] {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            p.ID(),
						Severity:        p.DefaultSeverity(),
						Node:            opNode,
						UnderlyingError: fmt.Errorf("`%s` must define parameter `%s` as expected by path `%s`", method, paramName, pathStr),
					})
				}
			}

			// Check that all defined parameters are used in path
			for paramName := range allDefinedParams {
				if !expectedParams[paramName] {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            p.ID(),
						Severity:        p.DefaultSeverity(),
						Node:            opNode,
						UnderlyingError: fmt.Errorf("parameter `%s` must be used in path `%s`", paramName, pathStr),
					})
				}
			}
		}

		// OpenAPI 3.2+: Also check additionalOperations (custom HTTP methods)
		if isOpenAPI32OrLater {
			additionalOps := pathItem.GetAdditionalOperations()
			if additionalOps != nil {
				for customMethodName, operation := range additionalOps.All() {
					customMethod := strings.ToUpper(customMethodName)
					if operation == nil {
						continue
					}

					// Collect operation-level path parameters
					opParams := make(map[string]*openapi.Parameter)
					for _, paramRef := range operation.GetParameters() {
						param := paramRef.GetObject()
						if param == nil || param.GetIn() != "path" {
							continue
						}

						opParams[param.GetName()] = param
					}

					// Combine path-level and operation-level parameters
					allDefinedParams := make(map[string]bool)
					for name := range pathLevelParams {
						allDefinedParams[name] = true
					}
					for name := range opParams {
						allDefinedParams[name] = true
					}

					// Get node for error reporting
					opNode := operation.GetRootNode()

					// Check that all expected parameters are defined
					for paramName := range expectedParams {
						if !allDefinedParams[paramName] {
							validationErrors = append(validationErrors, &validation.Error{
								Rule:            p.ID(),
								Severity:        p.DefaultSeverity(),
								Node:            opNode,
								UnderlyingError: fmt.Errorf("`%s` must define parameter `%s` as expected by path `%s`", customMethod, paramName, pathStr),
							})
						}
					}

					// Check that all defined parameters are used in path
					for paramName := range allDefinedParams {
						if !expectedParams[paramName] {
							validationErrors = append(validationErrors, &validation.Error{
								Rule:            p.ID(),
								Severity:        p.DefaultSeverity(),
								Node:            opNode,
								UnderlyingError: fmt.Errorf("parameter `%s` must be used in path `%s`", paramName, pathStr),
							})
						}
					}
				}
			}
		}
	}

	return validationErrors
}
