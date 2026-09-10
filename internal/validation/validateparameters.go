package validation

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

type ValidateParameters struct {
	sdkConfig *config.Configuration
	target    types.Target
}

var _ Rule = (*ValidateParameters)(nil)
var _ configAwareRule = (*ValidateParameters)(nil)
var _ targetAwareRule = (*ValidateParameters)(nil)

func (r *ValidateParameters) SetConfig(cfg *config.Configuration) {
	r.sdkConfig = cfg
}

func (r *ValidateParameters) SetTarget(target types.Target) {
	r.target = target
}

func (r *ValidateParameters) ID() string {
	return "generator-validate-parameters"
}

func (r *ValidateParameters) Category() string {
	return "validation"
}

func (r *ValidateParameters) Summary() string {
	return "Validate parameters use unique names after naming rules and valid serialization."
}

func (r *ValidateParameters) HowToFix() string {
	return "Ensure parameters have unique names after naming rules, include required 'name' and 'in' fields, and use valid style/serialization for their location."
}

func (r *ValidateParameters) Description() string {
	return "Validate parameters are unique once converted to field names and have a non-empty name. Also enforce correct `style` values for their location and serialization."
}

func (r *ValidateParameters) Link() string {
	return ""
}

func (r *ValidateParameters) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateParameters) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateParameters) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Get target from SDK config
	target := r.getTarget()

	// Initialize extensions based on the target context
	exts := extensions.New(target)

	doc := docInfo.Document
	if doc == nil {
		return nil
	}

	// Handle extension rewrites (errors intentionally ignored)
	_ = exts.HandleRewriteExtension(extensions.WithDocumentExtensions(doc.GetExtensions()))

	// Check OpenAPI version to know if we should check additionalOperations (3.2+)
	isOpenAPI32OrLater := doc.OpenAPI != "" && strings.HasPrefix(doc.OpenAPI, "3.2")

	// Track unique parameter names per operation per 'in' value
	// Key: "path:method", Value: map of "in" -> []nameReference
	uniqueParameterNames := make(map[string]map[string][]nameReference)

	// Iterate through all operations
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		// Check if operation should be ignored
		if ignored, _ := exts.HandleIgnoreExtension(operation.GetExtensions()); ignored != nil && *ignored {
			continue
		}

		// Get path and method from index node location
		httpMethod, httpPath := openapi.ExtractMethodAndPath(opIndexNode.Location)
		operationKey := fmt.Sprintf("%s:%s", httpPath, httpMethod)

		operationParamNames, ok := uniqueParameterNames[operationKey]
		if !ok {
			operationParamNames = make(map[string][]nameReference)
			uniqueParameterNames[operationKey] = operationParamNames
		}

		// Validate parameters for this operation
		parameters := operation.GetParameters()
		for _, paramRef := range parameters {
			param := paramRef.GetObject()
			if param == nil {
				continue
			}

			paramNode := paramRef.GetRootNode()
			if paramNode == nil {
				continue
			}

			// Check if parameter should be ignored
			if ignored, _ := exts.HandleIgnoreExtension(param.GetExtensions()); ignored != nil && *ignored {
				continue
			}

			// Validate 'in' property
			inVal, err := r.validateIn(param, paramNode, exts)
			if err != nil {
				validationErrors = append(validationErrors, err)
			}

			// Validate parameter name uniqueness
			paramNames, ok := operationParamNames[inVal]
			if !ok {
				paramNames = []nameReference{}
			}

			uniqueNames, err := r.validateNames(param, paramNode, paramNames, sanitization.GetSanitizedFieldNameResult, exts)
			if err != nil {
				validationErrors = append(validationErrors, err)
			} else {
				operationParamNames[inVal] = uniqueNames
			}

			// Validate content
			if err := r.validateContent(param, paramNode); err != nil {
				validationErrors = append(validationErrors, err)
			}

			// Validate allow-empty-value extension
			if err := r.validateAllowEmptyValue(param, paramNode, inVal, exts); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}

	// OpenAPI 3.2+: Also check additionalOperations (custom HTTP methods)
	if isOpenAPI32OrLater {
		for _, pathItemNode := range docInfo.Index.InlinePathItems {
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

			additionalOps := pathItem.GetAdditionalOperations()
			if additionalOps != nil {
				for customMethodName, operation := range additionalOps.All() {
					if operation == nil {
						continue
					}

					customMethod := strings.ToUpper(customMethodName)

					// Check if operation should be ignored
					if ignored, _ := exts.HandleIgnoreExtension(operation.GetExtensions()); ignored != nil && *ignored {
						continue
					}

					operationKey := fmt.Sprintf("%s:%s", pathStr, customMethod)

					operationParamNames, ok := uniqueParameterNames[operationKey]
					if !ok {
						operationParamNames = make(map[string][]nameReference)
						uniqueParameterNames[operationKey] = operationParamNames
					}

					// Validate parameters for this custom operation
					parameters := operation.GetParameters()
					for _, paramRef := range parameters {
						param := paramRef.GetObject()
						if param == nil {
							continue
						}

						paramNode := paramRef.GetRootNode()
						if paramNode == nil {
							continue
						}

						// Check if parameter should be ignored
						if ignored, _ := exts.HandleIgnoreExtension(param.GetExtensions()); ignored != nil && *ignored {
							continue
						}

						// Validate 'in' property
						inVal, err := r.validateIn(param, paramNode, exts)
						if err != nil {
							validationErrors = append(validationErrors, err)
						}

						// Validate parameter name uniqueness
						paramNames, ok := operationParamNames[inVal]
						if !ok {
							paramNames = []nameReference{}
						}

						uniqueNames, err := r.validateNames(param, paramNode, paramNames, sanitization.GetSanitizedFieldNameResult, exts)
						if err != nil {
							validationErrors = append(validationErrors, err)
						} else {
							operationParamNames[inVal] = uniqueNames
						}

						// Validate content
						if err := r.validateContent(param, paramNode); err != nil {
							validationErrors = append(validationErrors, err)
						}

						// Validate allow-empty-value extension
						if err := r.validateAllowEmptyValue(param, paramNode, inVal, exts); err != nil {
							validationErrors = append(validationErrors, err)
						}
					}
				}
			}
		}
	}

	return validationErrors
}

func (r *ValidateParameters) validateNames(param *openapi.Parameter, paramNode *yaml.Node, uniqueParameterNames []nameReference, getResultFunc func(string) sanitization.Result, exts *extensions.Extensions) ([]nameReference, error) {
	name := param.GetName()
	if name == "" {
		return nil, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            paramNode,
			UnderlyingError: errors.New("parameter must have a non-empty `name` property"),
		}
	}

	paramCore := param.GetCore()
	if paramCore == nil {
		return uniqueParameterNames, nil
	}

	// Get the name node for error reporting
	nameNode := paramCore.Name.GetKeyNodeOrRoot(paramNode)

	// Check for name override extension
	if nameOverride, hasOverride, _ := exts.HandleOperationParameterNameExtension(param); hasOverride && nameOverride != nil {
		name = nameOverride.Name
		if nameOverride.Node != nil {
			nameNode = nameOverride.Node
		}
		sanitizedParamNameResult := getResultFunc(name)
		if sanitizedParamName, origSanitizedParamName, orig := findNameConflict(sanitizedParamNameResult, uniqueParameterNames); orig != nil {
			return nil, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            nameNode,
				UnderlyingError: fmt.Errorf("`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to field name", name, sanitizedParamName, orig.Name, origSanitizedParamName, orig.Line),
			}
		} else {
			uniqueParameterNames = append(uniqueParameterNames, nameReference{
				Name:   name,
				Line:   nameNode.Line,
				Result: sanitizedParamNameResult,
			})
		}
	} else {
		sanitizedParamNameResult := getResultFunc(name)
		if _, _, orig := findNameConflict(sanitizedParamNameResult, uniqueParameterNames); orig == nil {
			uniqueParameterNames = append(uniqueParameterNames, nameReference{
				Name:   name,
				Line:   nameNode.Line,
				Result: sanitizedParamNameResult,
			})
		}
	}

	return uniqueParameterNames, nil
}

func (r *ValidateParameters) validateIn(param *openapi.Parameter, paramNode *yaml.Node, _ *extensions.Extensions) (string, error) {
	inVal := param.GetIn()
	if inVal == "" {
		return "", &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            paramNode,
			UnderlyingError: errors.New("parameter must have an `in` property"),
		}
	}

	inValStr := string(inVal)

	// Check header-specific validations
	if inVal == "header" {
		// Check for reserved header names
		if err := r.validateHeader(param, paramNode); err != nil {
			return inValStr, err
		}

		// Style validation for header parameters is handled by built-in validation-allowed-values rule
	}

	return inValStr, nil
}

func (r *ValidateParameters) validateContent(param *openapi.Parameter, paramNode *yaml.Node) error {
	content := param.GetContent()
	if content == nil {
		return nil
	}

	// Only one pair of content-type to schema is allowed
	if content.Len() > 1 {
		paramCore := param.GetCore()
		var contentNode *yaml.Node
		if paramCore != nil {
			contentNode = paramCore.Content.GetKeyNodeOrRoot(paramNode)
		} else {
			contentNode = paramNode
		}
		return &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            contentNode,
			UnderlyingError: errors.New("parameters with multiple content types are not valid"),
		}
	}

	return nil
}

func (r *ValidateParameters) validateHeader(param *openapi.Parameter, paramNode *yaml.Node) error {
	name := param.GetName()
	if name == "" {
		return nil
	}

	headerName := strings.ToLower(name)

	if slices.Contains([]string{"accept", "content-type", "authorization"}, headerName) {
		paramCore := param.GetCore()
		var nameNode *yaml.Node
		if paramCore != nil {
			nameNode = paramCore.Name.GetKeyNodeOrRoot(paramNode)
		} else {
			nameNode = paramNode
		}
		return &validation.Error{
			Rule:            r.ID(),
			Severity:        validation.SeverityWarning,
			Node:            nameNode,
			UnderlyingError: fmt.Errorf("header name `%s` is reserved and will be ignored", name),
		}
	}

	return nil
}

func (r *ValidateParameters) validateAllowEmptyValue(param *openapi.Parameter, paramNode *yaml.Node, inVal string, exts *extensions.Extensions) error {
	// Check if the x-speakeasy-allow-empty-value extension is present
	paramExtensions := param.GetExtensions()
	if paramExtensions == nil || paramExtensions.Len() == 0 {
		return nil
	}

	extName := exts.GetResolvedName(extensions.ExtAllowEmptyValue)
	if ext, ok := paramExtensions.Get(extName); ok && ext != nil {
		if inVal != "query" {
			return &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            paramNode,
				UnderlyingError: fmt.Errorf("parameter with `in` == `%s` cannot use `%s` extension, only query parameters are supported", inVal, extensions.ExtAllowEmptyValue.Name()),
			}
		}
	}

	return nil
}

// getTarget returns the target for SDK generation
func (r *ValidateParameters) getTarget() types.Target {
	if r.target.Template != "" {
		return r.target
	}
	// Default to "go" if no target was set
	return types.NewTargetFromTemplate("go")
}
