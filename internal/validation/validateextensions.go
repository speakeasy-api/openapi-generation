package validation

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type ValidateExtensions struct {
	docInfo   *document.DocumentInfo
	sdkConfig *config.Configuration
	target    types.Target
}

var _ Rule = (*ValidateExtensions)(nil)
var _ docInfoAwareRule = (*ValidateExtensions)(nil)
var _ configAwareRule = (*ValidateExtensions)(nil)
var _ targetAwareRule = (*ValidateExtensions)(nil)

func (r *ValidateExtensions) SetDocInfo(docInfo *document.DocumentInfo) {
	r.docInfo = docInfo
}

func (r *ValidateExtensions) SetConfig(cfg *config.Configuration) {
	r.sdkConfig = cfg
}

func (r *ValidateExtensions) SetTarget(target types.Target) {
	r.target = target
}

func (r *ValidateExtensions) ID() string {
	return "generator-validate-extensions"
}

func (r *ValidateExtensions) Category() string {
	return "validation"
}

func (r *ValidateExtensions) Summary() string {
	return "Validate extension usage for " + extensions.ExtGlobals.Name() + " parameters."
}

func (r *ValidateExtensions) HowToFix() string {
	return "Define x-speakeasy-globals parameters with unique names, valid schemas (primitive types or enums), and avoid collisions with server variables."
}

func (r *ValidateExtensions) Description() string {
	return "Validate " + extensions.ExtGlobals.Name() + " extension usage by enforcing unique parameter names, avoiding collisions with server variables, and only allowing primitive types. This keeps global parameters stable and generator-safe."
}

func (r *ValidateExtensions) Link() string {
	return ""
}

func (r *ValidateExtensions) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateExtensions) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateExtensions) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	validationErrors := make([]error, 0, 1)
	validationErrors = append(validationErrors, r.validateGlobalsExtension(ctx, docInfo, config)...)

	return validationErrors
}

func (r *ValidateExtensions) validateGlobalsExtension(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], _ *linter.RuleConfig) []error {
	// Location stack index for parent Server of a ServerVariable
	// Based on debugging: Server is at index 1 in the location stack
	const serverParentIdx = 1

	var validationErrors []error

	doc := docInfo.Document

	// Get target from SDK config
	target := r.getTarget()
	exts := extensions.New(target)

	// Handle extension rewrites from document-level extensions (errors intentionally ignored - extension rewriting is optional)
	_ = exts.HandleRewriteExtension(extensions.WithDocumentExtensions(doc.GetExtensions()))

	// Try to get the globals extension
	globals, err := exts.HandleGlobalsExtension(ctx, doc)
	if err != nil {
		// If there's an error parsing globals, return it directly as it's already a validation error
		validationErrors = append(validationErrors, err)
		return validationErrors
	}
	if globals == nil {
		// No globals extension present, nothing to validate
		return validationErrors
	}

	parameters := globals.Parameters
	if len(parameters) == 0 {
		return validationErrors
	}

	uniqueParameterNames := []nameReference{}
	globalServerVariableNames := map[string]nameReference{}

	// Use Index to efficiently collect all server variables
	for _, varNode := range docInfo.Index.ServerVariables {
		if varNode == nil || varNode.Node == nil {
			continue
		}

		variable := varNode.Node
		variableName := varNode.Location.ParentKey()

		if variableName == "" {
			continue
		}

		sanitizedFieldName := sanitization.SanitizeFieldName(variableName)

		if _, ok := globalServerVariableNames[sanitizedFieldName]; !ok {
			// Get the parent Server from the location stack
			var parentServer *openapi.Server
			serverCtx := varNode.Location[serverParentIdx]
			matcher := openapi.Matcher{
				Server: func(s *openapi.Server) error {
					parentServer = s
					return nil
				},
			}
			_ = serverCtx.ParentMatchFunc(matcher)

			if parentServer != nil {
				// Get the key node for this variable from the parent server's variables map
				node := parentServer.GetCore().Variables.GetMapKeyNodeOrRoot(variableName, variable.GetRootNode())
				globalServerVariableNames[sanitizedFieldName] = nameReference{
					Name:   variableName,
					Line:   node.Line,
					Result: sanitization.GetSanitizedFieldNameResult(variableName),
				}
			}
		}
	}

	// Validate each parameter in globals
	for _, paramRef := range parameters {
		if paramRef == nil {
			continue
		}

		// Skip references - globals should define parameters inline
		if paramRef.IsReference() {
			continue
		}

		param := paramRef.GetObject()
		if param == nil {
			continue
		}

		// Get parameter node for error reporting
		paramNode := paramRef.GetRootNode()

		// Check if parameter has schema
		paramSchema := param.GetSchema()
		if paramSchema == nil {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            paramNode,
				UnderlyingError: errors.New("`x-speakeasy-globals` parameter must have `schema`"),
			})
			continue
		}

		// Get the schema to check type
		// Resolve the schema reference to get the actual schema (handles $ref)
		resolutionValidationErrs, err := paramSchema.Resolve(ctx, r.docInfo.GetResolutionOptions(ctx))
		if err != nil {
			// Capture resolution error as validation error
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            paramSchema.GetRootNode(),
				UnderlyingError: fmt.Errorf("failed to resolve schema reference: %w", err),
			})
			continue
		}
		// Add any validation errors from resolution
		validationErrors = append(validationErrors, resolutionValidationErrs...)

		// Get the resolved schema
		resolvedSchema := paramSchema.GetResolvedSchema()
		schema := resolvedSchema.GetSchema()
		if schema != nil {
			types := schema.GetType()
			typ := ""
			if len(types) > 0 {
				typ = string(types[0])
			}

			// Check if it's a primitive type
			// Also allow enums (which may have type or may be inferred from enum values)
			isPrimitive := slices.Contains([]string{"string", "number", "integer", "boolean"}, typ)
			isEnum := len(schema.GetEnum()) > 0

			if !isPrimitive && !isEnum {
				// Get schema node for better error location
				schemaNode := paramSchema.GetRootNode()
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            schemaNode,
					UnderlyingError: fmt.Errorf("only primitive types are allowed in `x-speakeasy-globals` parameters: found `%s`", typ),
				})
			}
		}

		// Check if parameter has name
		paramName := param.GetName()
		if paramName == "" {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            paramNode,
				UnderlyingError: errors.New("`x-speakeasy-globals` parameter must have a `name`"),
			})
			continue
		}

		// Check for name collisions among parameters
		sanitizedParamNameResult := sanitization.GetSanitizedFieldNameResult(paramName)

		if sanitizedParamName, origSanitizedParamName, orig := findNameConflict(sanitizedParamNameResult, uniqueParameterNames); orig != nil {
			if !schemasEqual(resolvedSchema, orig.Schema) {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            paramNode,
					UnderlyingError: fmt.Errorf("`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`], global parameters must be unique", paramName, sanitizedParamName, orig.Name, origSanitizedParamName, orig.Line),
				})
			}
		} else {
			uniqueParameterNames = append(uniqueParameterNames, nameReference{
				Name:   paramName,
				Line:   paramNode.Line,
				Result: sanitizedParamNameResult,
				In:     param.GetIn(),
				Schema: schema,
			})
		}

		// Check for collisions with server variables
		for _, serverVariableRef := range globalServerVariableNames {
			sanitizedParamName, origSanitizedParamName := sanitizedParamNameResult.GetConflicting(serverVariableRef.Result)

			if sanitizedParamName != "" && origSanitizedParamName != "" {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            paramNode,
					UnderlyingError: fmt.Errorf("global parameter `%s` (`%s`) will collide with server variable `%s` (`%s`) [line `%d`], global parameters must be unique", paramName, sanitizedParamName, serverVariableRef.Name, origSanitizedParamName, serverVariableRef.Line),
				})
			}
		}
	}

	return validationErrors
}

func schemasEqual(left *oas3.JSONSchema[oas3.Concrete], right *oas3.Schema) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	if left.IsSchema() {
		return left.GetSchema().IsEqual(right)
	}

	return false
}

// getTarget returns the target for SDK generation
func (r *ValidateExtensions) getTarget() types.Target {
	if r.target.Template != "" {
		return r.target
	}
	// Default to "go" if no target was set
	return types.NewTargetFromTemplate("go")
}
