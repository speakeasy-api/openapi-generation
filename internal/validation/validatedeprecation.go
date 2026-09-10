package validation

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

type ValidateDeprecation struct{}

var _ Rule = (*ValidateDeprecation)(nil)

func (r *ValidateDeprecation) ID() string {
	return "generator-validate-deprecation"
}

func (r *ValidateDeprecation) Category() string {
	return "validation"
}

func (r *ValidateDeprecation) Summary() string {
	return "Validate deprecation extensions and replacement references."
}

func (r *ValidateDeprecation) HowToFix() string {
	return "Set deprecated: true when using deprecation extensions and ensure replacement operation IDs or parameter names exist."
}

func (r *ValidateDeprecation) Description() string {
	return fmt.Sprintf("Ensure correct usage of %s and %s extensions. Deprecations should provide a replacement and consistent messaging.", extensions.ExtDeprecationReplacement.Name(), extensions.ExtDeprecationMessage.Name())
}

func (r *ValidateDeprecation) Link() string {
	return ""
}

func (r *ValidateDeprecation) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateDeprecation) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateDeprecation) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	const (
		deprecationReplacementExt = "x-speakeasy-deprecation-replacement"
		deprecationMessageExt     = "x-speakeasy-deprecation-message"
	)

	// Iterate over all operations
	for _, opNode := range docInfo.Index.Operations {
		operation := opNode.Node
		if operation == nil {
			continue
		}

		// Check if deprecated is set to true in the YAML
		// Use the core model to check if the field is actually present
		deprecatedNode := operation.GetCore().Deprecated
		deprecatedIsNotSet := !deprecatedNode.Present || deprecatedNode.Value == nil || !*deprecatedNode.Value

		// Get operation extensions
		opExtensions := operation.GetExtensions()

		// Check for deprecation-replacement extension
		if replacementVal, hasReplacement := opExtensions.Get(deprecationReplacementExt); hasReplacement {
			if deprecatedIsNotSet {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:     r.ID(),
					Severity: r.DefaultSeverity(),
					Node:     operation.GetRootNode(),
					UnderlyingError: fmt.Errorf(
						"operation must have a `deprecated` property set to `true` to use `%s`", extensions.ExtDeprecationReplacement.Name(),
					),
				})
			} else {
				// Verify the replacement operation exists
				var replacementOperationID string
				if err := replacementVal.Decode(&replacementOperationID); err == nil && replacementOperationID != "" {
					if !operationIDExists(docInfo, replacementOperationID) {
						validationErrors = append(validationErrors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            operation.GetRootNode(),
							UnderlyingError: fmt.Errorf("`%s` not found, operation with id `%s` doesn't exist", extensions.ExtDeprecationReplacement.Name(), replacementOperationID),
						})
					}
				}
			}
		}

		// Check for deprecation-message extension
		if _, hasMessage := opExtensions.Get(deprecationMessageExt); hasMessage && deprecatedIsNotSet {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:     r.ID(),
				Severity: r.DefaultSeverity(),
				Node:     operation.GetRootNode(),
				UnderlyingError: fmt.Errorf(
					"operation must have a `deprecated` property set to `true` to use `%s`", extensions.ExtDeprecationMessage.Name(),
				),
			})
		}

		// Check parameters for deprecation extensions
		parameters := operation.GetParameters()
		for _, paramRef := range parameters {
			param := paramRef.GetObject()
			if param == nil {
				continue
			}

			// Check if parameter is deprecated in the YAML
			// Use the core model to check if the field is actually present
			paramDeprecatedNode := param.GetCore().Deprecated
			paramDeprecatedIsNotSet := !paramDeprecatedNode.Present || paramDeprecatedNode.Value == nil || !*paramDeprecatedNode.Value

			paramExtensions := param.GetExtensions()

			// Check for deprecation-replacement extension on parameter
			if replacementVal, hasReplacement := paramExtensions.Get(deprecationReplacementExt); hasReplacement {
				if paramDeprecatedIsNotSet {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:     r.ID(),
						Severity: r.DefaultSeverity(),
						Node:     paramRef.GetRootNode(),
						UnderlyingError: fmt.Errorf(
							"parameter must have a `deprecated` property set to `true` to use `%s`", extensions.ExtDeprecationReplacement.Name(),
						),
					})
				} else {
					// Verify the replacement parameter exists
					var replacementParameterName string
					if err := replacementVal.Decode(&replacementParameterName); err == nil && replacementParameterName != "" {
						if !parameterNameExists(parameters, replacementParameterName) {
							validationErrors = append(validationErrors, &validation.Error{
								Rule:            r.ID(),
								Severity:        r.DefaultSeverity(),
								Node:            paramRef.GetRootNode(),
								UnderlyingError: fmt.Errorf("`%s` not found, parameter with name `%s` doesn't exist", extensions.ExtDeprecationReplacement.Name(), replacementParameterName),
							})
						}
					}
				}
			}

			// Check for deprecation-message extension on parameter
			if _, hasMessage := paramExtensions.Get(deprecationMessageExt); hasMessage && paramDeprecatedIsNotSet {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:     r.ID(),
					Severity: r.DefaultSeverity(),
					Node:     paramRef.GetRootNode(),
					UnderlyingError: fmt.Errorf(
						"parameter must have a `deprecated` property set to `true` to use `%s`", extensions.ExtDeprecationMessage.Name(),
					),
				})
			}
		}
	}

	return validationErrors
}

// operationIDExists checks if an operation with the given ID exists in the document
func operationIDExists(docInfo *linter.DocumentInfo[*openapi.OpenAPI], operationID string) bool {
	for _, opNode := range docInfo.Index.Operations {
		operation := opNode.Node
		if operation != nil && operation.GetOperationID() == operationID {
			return true
		}
	}
	return false
}

// parameterNameExists checks if a parameter with the given name exists in the parameter list
func parameterNameExists(parameters []*openapi.ReferencedParameter, name string) bool {
	for _, paramRef := range parameters {
		param := paramRef.GetObject()
		if param != nil && param.GetName() == name {
			return true
		}
	}
	return false
}
