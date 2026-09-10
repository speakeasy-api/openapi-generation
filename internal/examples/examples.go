package examples

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

type Examples struct{}

func (e *Examples) NewOperationExamplesMap() *sequencedmap.Map[string, config.OperationExamples] {
	return sequencedmap.New[string, config.OperationExamples]()
}

func (e *Examples) NewExamplesMap() config.Examples {
	return sequencedmap.New[string, *sequencedmap.Map[string, config.OperationExamples]]()
}

func (e *Examples) NewParameterExamples() *config.ParameterExamples {
	return &config.ParameterExamples{}
}

func (e *Examples) NewResponsesExamplesMap() *sequencedmap.Map[string, *sequencedmap.Map[string, yaml.Node]] {
	return sequencedmap.New[string, *sequencedmap.Map[string, yaml.Node]]()
}

func (e *Examples) NewYamlNodeMap() *sequencedmap.Map[string, yaml.Node] {
	return sequencedmap.New[string, yaml.Node]()
}

func (e *Examples) NewExample(name string, value yaml.Node) *ast.Example {
	return ast.NewExample(name, "", &value)
}

func (e *Examples) YamlNodeFromString(s string) yaml.Node {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(s), &node); err != nil {
		panic(err)
	}

	if node.Kind == yaml.DocumentNode {
		node = *node.Content[0]
	}

	return node
}

func (e *Examples) GetDefaultExampleName(op *ast.Operation) string {
	return GetDefaultExampleName(op)
}

func (e *Examples) IsDefaultExample(name string) bool {
	return IsDefaultExample(name)
}

func GetDefaultExampleName(op *ast.Operation) string {
	return "speakeasy-default-" + casing.New().ToKebab(op.ID)
}

func IsDefaultExample(name string) bool {
	return strings.HasPrefix(name, "speakeasy-default-")
}
