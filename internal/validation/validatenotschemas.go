package validation

import (
	"context"
	"errors"

	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
)

// ValidateNotSchemas flags schemas that use the JSON Schema `not` keyword.
// The SDK generator does not support `not`; any operation whose request or
// response schemas reach such a schema is silently skipped at generation time.
// This rule surfaces the unsupported construct during linting so it can be
// caught before methods disappear from a published SDK. It ships only in the
// opt-in speakeasy-unsupported-constructs ruleset.
type ValidateNotSchemas struct{}

var _ Rule = (*ValidateNotSchemas)(nil)

func (r *ValidateNotSchemas) ID() string {
	return "generator-validate-not-schemas"
}

func (r *ValidateNotSchemas) Category() string {
	return "validation"
}

func (r *ValidateNotSchemas) Summary() string {
	return "Schemas must not use the unsupported `not` keyword."
}

func (r *ValidateNotSchemas) HowToFix() string {
	return "Remove the `not` keyword from the schema, or remodel the constraint using supported constructs such as `oneOf`, `anyOf`, `allOf`, or more specific property/type definitions. Lower the rule severity to warn in lint.yaml to generate anyway, accepting that every operation reaching this schema is dropped from the SDK."
}

func (r *ValidateNotSchemas) Description() string {
	return "The JSON Schema `not` keyword is unsupported by the SDK generator; any operation reaching a schema that uses it is silently skipped during generation."
}

func (r *ValidateNotSchemas) Link() string {
	return ""
}

func (r *ValidateNotSchemas) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateNotSchemas) Versions() []string {
	return nil
}

func (r *ValidateNotSchemas) Run(_ context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], _ *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	for _, indexNode := range docInfo.Index.GetAllSchemas() {
		if indexNode == nil || indexNode.Node == nil {
			continue
		}

		schema := indexNode.Node.GetSchema()
		if schema == nil || schema.GetNot() == nil {
			continue
		}

		node := schema.GetCore().Not.GetKeyNodeOrRoot(schema.GetRootNode())
		validationErrors = append(validationErrors, &validation.Error{
			Rule:     r.ID(),
			Severity: r.DefaultSeverity(),
			Node:     node,
			UnderlyingError: errors.New(
				"schema uses `not`, which the SDK generator does not support — every operation reaching this schema will be silently dropped from the SDK",
			),
		})
	}

	return validationErrors
}
