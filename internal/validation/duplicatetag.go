package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

type DuplicateTag struct{}

var _ Rule = (*DuplicateTag)(nil)

func (r *DuplicateTag) ID() string {
	return "generator-duplicate-tag"
}

func (r *DuplicateTag) Category() string {
	return "validation"
}

func (r *DuplicateTag) Summary() string {
	return "Ensure no duplicated tags found."
}

func (r *DuplicateTag) HowToFix() string {
	return "Rename tags so their sanitized names are unique and remove duplicate tags within operations."
}

func (r *DuplicateTag) Description() string {
	return "Tag names must be unique when converted to class, field, or file names. Collisions can merge unrelated operations into the same SDK group."
}

func (r *DuplicateTag) Link() string {
	return ""
}

func (r *DuplicateTag) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *DuplicateTag) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateTag) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Check global tags section
	uniqueGlobalTagNames := []nameReference{}
	for _, tagIndexNode := range docInfo.Index.Tags {
		tag := tagIndexNode.Node
		if tag == nil {
			continue
		}

		name := tag.GetName()
		if name == "" {
			continue
		}

		// Use the tag's root node for error reporting
		nameNode := tag.GetRootNode()

		sanitizedTagNameResult := sanitization.GetSanitizedClassNameResult(name)

		if sanitizedTagName, origTagName, orig := findNameConflict(sanitizedTagNameResult, uniqueGlobalTagNames); orig != nil {
			message := fmt.Sprintf("`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to class/field name", name, sanitizedTagName, orig.Name, origTagName, orig.Line)

			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            nameNode,
				UnderlyingError: errors.New(message),
			})
		} else {
			uniqueGlobalTagNames = append(uniqueGlobalTagNames, nameReference{
				Name:   name,
				Line:   nameNode.Line,
				Result: sanitizedTagNameResult,
			})
		}
	}

	// Check tags used in operations
	uniqueOperationTagNames := []nameReference{}
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		opTags := operation.GetTags()
		if len(opTags) == 0 {
			continue
		}

		opCore := operation.GetCore()
		opRootNode := operation.GetRootNode()
		tagsInOp := map[string]bool{}

		for i, tagName := range opTags {
			// Get the specific tag node from the tags array
			tagNode := opCore.Tags.GetSliceValueNodeOrRoot(i, opRootNode)

			// Check for duplicate tags within the same operation
			if _, ok := tagsInOp[tagName]; ok {
				message := fmt.Sprintf("duplicate tag `%s` in operation", tagName)

				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        validation.SeverityWarning,
					Node:            tagNode,
					UnderlyingError: errors.New(message),
				})
			}
			tagsInOp[tagName] = true

			// Check for tag name collisions across operations
			sanitizedTagNameResult := sanitization.GetSanitizedClassNameResult(tagName)

			if sanitizedTagName, origTagName, orig := findNameConflict(sanitizedTagNameResult, uniqueOperationTagNames); orig != nil {
				if orig.Name != tagName {
					message := fmt.Sprintf("`%s` (`%s`) will collide with `%s` (`%s`) [line `%d`] when converted to class/field name", tagName, sanitizedTagName, orig.Name, origTagName, orig.Line)

					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            tagNode,
						UnderlyingError: errors.New(message),
					})
				}
			} else {
				uniqueOperationTagNames = append(uniqueOperationTagNames, nameReference{
					Name:   tagName,
					Line:   tagNode.Line,
					Result: sanitizedTagNameResult,
				})
			}
		}
	}

	return validationErrors
}
