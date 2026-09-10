package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// ValidateCasing flags component schema names that are identical once case is ignored. Such names
// produce distinct types in case-sensitive languages but collide in case-insensitive contexts
// (e.g. C# type lookups and case-insensitive filesystems), where one model silently overwrites the other.
type ValidateCasing struct{}

var _ Rule = (*ValidateCasing)(nil)

func (r *ValidateCasing) ID() string {
	return "generator-validate-casing"
}

func (r *ValidateCasing) Category() string {
	return "collision"
}

func (r *ValidateCasing) Summary() string {
	return "Ensure schema names do not differ only by casing."
}

func (r *ValidateCasing) HowToFix() string {
	return "Rename or merge schemas whose names differ only by casing so each name is distinct when case is ignored."
}

func (r *ValidateCasing) Description() string {
	return "Schema names that differ only by casing generate distinct types in case-sensitive languages but collide in case-insensitive contexts such as C# type resolution and case-insensitive filesystems, where one generated model overwrites the other."
}

func (r *ValidateCasing) Link() string {
	return ""
}

func (r *ValidateCasing) DefaultSeverity() validation.Severity {
	return validation.SeverityWarning
}

func (r *ValidateCasing) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateCasing) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	components := docInfo.Document.GetComponents()
	if components == nil {
		return nil
	}

	schemas := components.GetSchemas()
	if schemas == nil {
		return nil
	}

	componentsCore := components.GetCore()
	componentsRoot := components.GetRootNode()

	var validationErrors []error
	firstByLower := map[string]string{}

	for schemaName := range schemas.All() {
		lower := strings.ToLower(schemaName)

		first, seen := firstByLower[lower]
		if !seen {
			firstByLower[lower] = schemaName
			continue
		}
		if first == schemaName {
			continue
		}

		node := componentsCore.Schemas.GetMapKeyNodeOrRoot(schemaName, componentsRoot)
		validationErrors = append(validationErrors, &validation.Error{
			Rule:     r.ID(),
			Severity: r.DefaultSeverity(),
			Node:     node,
			UnderlyingError: fmt.Errorf(
				"schema name `%s` differs from `%s` only by casing and will collide in case-insensitive languages (e.g. C#) and filesystems",
				schemaName,
				first,
			),
		})
	}

	return validationErrors
}
