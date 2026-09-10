package validation

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi/hashing"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

// DuplicateComponentSchemas flags named component schemas that are structurally identical. Each one
// generates its own SDK type, bloating the surface area with redundant models that could be a single
// component shared via $ref.
type DuplicateComponentSchemas struct{}

var _ Rule = (*DuplicateComponentSchemas)(nil)

func (r *DuplicateComponentSchemas) ID() string {
	return "generator-duplicate-component-schemas"
}

func (r *DuplicateComponentSchemas) Category() string {
	return "collision"
}

func (r *DuplicateComponentSchemas) Summary() string {
	return "Flag structurally identical component schemas that generate redundant types."
}

func (r *DuplicateComponentSchemas) HowToFix() string {
	return "Consolidate structurally identical component schemas into a single component and reference it with `$ref`."
}

func (r *DuplicateComponentSchemas) Description() string {
	return "Named component schemas that are structurally identical each generate a separate SDK type. Merging them into one component referenced via `$ref` reduces the SDK surface area and gives a single, clearly named type."
}

func (r *DuplicateComponentSchemas) Link() string {
	return ""
}

func (r *DuplicateComponentSchemas) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *DuplicateComponentSchemas) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateComponentSchemas) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
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

	type entry struct {
		name string
		node *yaml.Node
	}

	byHash := map[string][]entry{}
	order := []string{}

	for schemaName, schemaRef := range schemas.All() {
		if schemaRef == nil {
			continue
		}

		// Only compare concrete schemas; a component that is itself a bare $ref is an alias, not a duplicate.
		schema := schemaRef.GetSchema()
		if schema == nil {
			continue
		}

		rootNode := schema.GetRootNode()
		if rootNode == nil {
			continue
		}

		hash := hashing.Hash(rootNode)
		if _, seen := byHash[hash]; !seen {
			order = append(order, hash)
		}
		byHash[hash] = append(byHash[hash], entry{
			name: schemaName,
			node: componentsCore.Schemas.GetMapKeyNodeOrRoot(schemaName, componentsRoot),
		})
	}

	var validationErrors []error

	for _, hash := range order {
		group := byHash[hash]
		if len(group) < 2 {
			continue
		}

		sort.Slice(group, func(i, j int) bool {
			return group[i].name < group[j].name
		})

		others := group[1:]
		quoted := make([]string, len(others))
		for i, e := range others {
			quoted[i] = fmt.Sprintf("`%s`", e.name)
		}

		validationErrors = append(validationErrors, &validation.Error{
			Rule:     r.ID(),
			Severity: r.DefaultSeverity(),
			Node:     group[0].node,
			UnderlyingError: fmt.Errorf(
				"schema `%s` is structurally identical to %s; consolidate them into a single component referenced via `$ref` to avoid generating redundant types",
				group[0].name,
				strings.Join(quoted, ", "),
			),
		})
	}

	return validationErrors
}
