package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type DuplicateProperties struct {
	sdkConfig *config.Configuration
	target    types.Target
}

var _ Rule = (*DuplicateProperties)(nil)
var _ configAwareRule = (*DuplicateProperties)(nil)
var _ targetAwareRule = (*DuplicateProperties)(nil)

func (r *DuplicateProperties) SetConfig(cfg *config.Configuration) {
	r.sdkConfig = cfg
}

func (r *DuplicateProperties) SetTarget(target types.Target) {
	r.target = target
}

func (r *DuplicateProperties) ID() string {
	return "generator-duplicate-properties"
}

func (r *DuplicateProperties) Category() string {
	return "validation"
}

func (r *DuplicateProperties) Summary() string {
	return "Ensure no duplicate property names collide after naming rules."
}

func (r *DuplicateProperties) HowToFix() string {
	return "Rename properties or apply name overrides so sanitized field names are unique, and remove empty property names."
}

func (r *DuplicateProperties) Description() string {
	return "Property names must be unique and not empty within a schema when converted to field names. This accounts for language-specific sanitization rules (e.g., camelCase in TypeScript, snake_case in Python) to avoid field collisions."
}

func (r *DuplicateProperties) Link() string {
	return ""
}

func (r *DuplicateProperties) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *DuplicateProperties) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateProperties) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
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

	// Handle extension rewrites from document-level extensions (errors intentionally ignored - extension rewriting is optional)
	_ = exts.HandleRewriteExtension(extensions.WithDocumentExtensions(doc.GetExtensions()))

	// Check all schemas in the document
	allSchemas := docInfo.Index.GetAllSchemas()

	for _, schemaNode := range allSchemas {
		schemaRef := schemaNode.Node
		if schemaRef == nil {
			continue
		}

		// Get the actual schema object
		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		// Skip if no properties
		properties := schema.GetProperties()
		if properties == nil || properties.Len() == 0 {
			continue
		}

		// Track unique property names for collision detection
		uniquePropertyNames := []nameReference{}

		// Iterate through all properties
		for propertyName, propertySchema := range properties.All() {
			if propertySchema == nil {
				continue
			}

			actualPropertySchema := propertySchema.GetSchema()
			if actualPropertySchema == nil {
				continue
			}

			// Check for x-speakeasy-ignore extension
			if ignored, _ := exts.HandleIgnoreExtension(actualPropertySchema.GetExtensions()); ignored != nil && *ignored {
				continue
			}

			// Check for empty property name
			if propertyName == "" {
				// Get the empty property key node
				node := schema.GetCore().Properties.GetMapKeyNodeOrRoot(propertyName, actualPropertySchema.GetRootNode())

				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            node,
					UnderlyingError: errors.New("empty property name"),
				})
				continue
			}

			// Get name override if present (e.g., x-speakeasy-name-override)
			resolved := propertySchema.GetResolvedSchema()
			nameOverrideFeb2026 := r.sdkConfig != nil && r.sdkConfig.Generation.Fixes != nil && r.sdkConfig.Generation.Fixes.NameOverrideFeb2026
			overriddenPropertyName, err := exts.GetPropertyName(propertySchema, resolved, propertyName, nameOverrideFeb2026)
			if err != nil {
				// If error getting resolved name, use original property name
				overriddenPropertyName = propertyName
			}
			nameWasOverridden := overriddenPropertyName != propertyName

			// Prefix with model namespace to prevent false positives when x-speakeasy-model-namespace is used
			if resolved != nil && resolved.IsSchema() {
				if ns, err := exts.GetModelNamespace(resolved.GetSchema().GetExtensions()); err == nil && ns != "" {
					overriddenPropertyName = ns + "." + overriddenPropertyName
				}
			}

			// Sanitize the property name for target language
			sanitizedPropertyNameResult := sanitization.GetSanitizedFieldNameResult(overriddenPropertyName)

			// Check for name conflicts
			if sanitizedPropertyName, origPropertyName, orig := findNameConflict(sanitizedPropertyNameResult, uniquePropertyNames); orig != nil {
				// Get the property key node, with property's root node as fallback
				node := schema.GetCore().Properties.GetMapKeyNodeOrRoot(propertyName, actualPropertySchema.GetRootNode())

				// When nameOverrideFeb2026 is false, name collisions can arise from name overrides bleeding through $ref or allOf.
				// In this case, we raise a warning instead of an error to avoid breaking existing workflows.
				nameCollisionSeverity := r.DefaultSeverity()
				nameCollisionTip := ""
				if !nameOverrideFeb2026 && nameWasOverridden {
					nameCollisionSeverity = validation.SeverityWarning
					nameCollisionTip = " Consider enabling fixes.nameOverrideFeb2026 to prevent name overrides from propagating through $ref or allOf composition."
				}

				validationErrors = append(validationErrors, &validation.Error{
					Rule:     r.ID(),
					Severity: nameCollisionSeverity,
					Node:     node,
					UnderlyingError: fmt.Errorf(
						"`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to field name.%s",
						overriddenPropertyName,
						sanitizedPropertyName,
						orig.Name,
						origPropertyName,
						orig.Line,
						nameCollisionTip,
					),
				})
			} else {
				// Add to unique names list - use property key node for accurate line number
				propertyKeyNode := schema.GetCore().Properties.GetMapKeyNodeOrRoot(propertyName, actualPropertySchema.GetRootNode())
				var line int
				if propertyKeyNode != nil {
					line = propertyKeyNode.Line
				}

				uniquePropertyNames = append(uniquePropertyNames, nameReference{
					Name:   overriddenPropertyName,
					Line:   line,
					Result: sanitizedPropertyNameResult,
				})
			}
		}
	}

	return validationErrors
}

// getTarget returns the target for SDK generation
func (r *DuplicateProperties) getTarget() types.Target {
	if r.target.Template != "" {
		return r.target
	}
	// Default to "go" if no target was set
	return types.NewTargetFromTemplate("go")
}
