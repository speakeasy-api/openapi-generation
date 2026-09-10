package ast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/speakeasy-api/openapi/yml"
	"gopkg.in/yaml.v3"
)

// ResponseBodyTarget represents a reference to a particular field in a response body of a particular operation in an Arazzo workflow
type ResponseBodyTarget struct {
	// Target is the target FieldDef or TypeDef that the reference is pointing to
	Target any
	// Path is a json-pointer to the field in the response body
	Path string
	// The response body content that the reference is pointing to
	Body *ResponseBodyContent
	// The index of the step and therefore the operation in the workflow that the reference is pointing to
	StepIdx int
}

type InputTarget struct {
	// Target is the target FieldDef or TypeDef that the reference is pointing to
	Target any
	// Path is a json-pointer to the field in the inputs
	Path string
	// The inputs content that the reference is pointing to
	Inputs *FieldDef
}

type OutputTarget struct {
	// Target is the target FieldDef or TypeDef that the reference is pointing to
	Target any
	// Path is a json-pointer to the field in the outputs
	Path string
	// The outputs content that the reference is pointing to
	Outputs *FieldDef
	// The index of the step and therefore the operation in the workflow that the reference is pointing to
	StepIdx int
}

// ExampleReference represents a reference to another field in another operation, it could be a response body, headers etc
type ExampleReference struct {
	// Target can be one of a number of different target types like ResponseBodyTarget, InputTarget or OutputTarget
	Target any
	// The type of the Target to help determine its typing
	Type string
}

// ExampleReplacement represents a replacement for a particular field in an example
type ExampleReplacement struct {
	// Path is a json-pointer to the field in the example
	Path string
	// Value is the replacement for the field
	Value *Example
}

type Example struct {
	name string

	// Description is the description of the example
	Description string

	// Value is the example value if set otherwise it is a reference to another field
	Value *yaml.Node
	// Reference is a reference to another field in another operation if set
	Reference *ExampleReference

	// Replacements are replacements for fields in the example or reference
	Replacements []*ExampleReplacement
}

type Examples []*Example

// Clone creates a deep copy of the Examples
func (e Examples) Clone() Examples {
	if e == nil {
		return nil
	}

	cloned := make(Examples, len(e))

	for i, example := range e {
		cloned[i] = example.Clone()
	}

	return cloned
}

func (e Examples) MarshalYAML() (any, error) {
	type example struct {
		Name        string     `yaml:",omitempty"`
		Value       *yaml.Node `yaml:",omitempty"`
		Description string     `yaml:",omitempty"`
	}

	examples := make([]example, 0, len(e))

	for _, ex := range e {
		examples = append(examples, example{
			Name:        ex.name,
			Value:       ex.Value,
			Description: ex.Description,
		})
	}

	return examples, nil
}

func (e *Examples) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected sequence node, got %v", node.Kind)
	}

	type example struct {
		Name        string
		Value       yaml.Node
		Description string
	}

	examples := []example{}

	if err := node.Decode(&examples); err != nil {
		return err
	}

	for _, ex := range examples {
		*e = append(*e, &Example{
			name:        ex.Name,
			Value:       &ex.Value,
			Description: ex.Description,
		})
	}

	return nil
}

func (e Examples) Match(matchers Matchers) error {
	if matchers.Examples != nil {
		return matchers.Examples(e)
	}

	return nil
}

// Merges the given Examples into this Examples. The algorithm adds data from
// the given Examples to this Examples where it is undefined. Where there is
// a conflicting Example, this Examples's data is preserved. It does not remove
// any data from this Examples.
func (e Examples) Merge(other Examples) {
	if e == nil || other == nil {
		return
	}

	for _, otherExample := range other {
		if e.FindByName(otherExample.Name()) != nil {
			continue
		}

		e = append(e, otherExample)
	}
}

func (e Examples) FindByName(name string) *Example {
	for _, ex := range e {
		if ex.Name() == name {
			return ex
		}
	}

	return nil
}

func NewExample(name, description string, value *yaml.Node) *Example {
	return &Example{
		name:        name,
		Value:       value,
		Description: description,
	}
}

func NewExampleReference(name, description string, reference *ExampleReference) *Example {
	return &Example{
		name:        name,
		Reference:   reference,
		Description: description,
	}
}

// NewExampleFromJSON creates a new Example from a JSON string by converting the JSON to YAML, then
// using the YAML representation as the `value` in the Example.
// Resulting equivalent OpenAPI YAML:
//
//	examples:
//	  <name>:
//	    description: <description>
//	    value:
//	      <json>: <value>
//	      <converted>:
//	        <to>: <yaml>
func NewExampleFromJSON(name, description string, jsonValue string) (*Example, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonValue), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal example JSON: %w", err)
	}

	// Then marshal to YAML bytes
	yamlBytes, err := yaml.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal example into YAML: %w", err)
	}

	// Finally unmarshal YAML bytes into a yaml.Node
	var value yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal example back to YAML: %w", err)
	}

	return NewExample(name, description, value.Content[0]), nil
}

// NewExampleFromString creates a new Example for a single-value field from a string.
// Resulting equivalent OpenAPI YAML:
//
//	examples:
//	  <name>:
//	    description: <description>
//	    value: <exampleValue>
func NewExampleFromString(name, description string, exampleValue string) (*Example, error) {
	value := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: exampleValue,
	}

	return NewExample(name, description, value), nil
}

func (e *Example) Name() string {
	if e == nil {
		return ""
	}

	return e.name
}

func (e *Example) ToString() string {
	if e.Value == nil {
		return ""
	}

	if e.isMapOrSequence() {
		return e.ToJSON()
	}

	var v any
	if err := e.Value.Decode(&v); err != nil {
		panic(err)
	}

	// A yaml `!!null` scalar decodes to a nil `any`. `fmt.Sprintf("%v", nil)`
	// would print the literal `<nil>`, which both leaks Go-style output and
	// breaks MDX renderers that parse `<nil>` as a JSX tag. Return "null" to
	// preserve the user's intent (they wrote `example: null` deliberately,
	// usually to signal that null is a valid value for a nullable field).
	// Mirrors `ToJSON()`'s "null" for the same nil case.
	if v == nil {
		return "null"
	}

	// If the decoded value is a slice or map, convert to JSON to avoid
	// Go-style output like [a b] or map[k:v]. This handles cases where
	// the yaml.Node Kind is not correctly set (e.g., Kind is 0).
	// We use json.Marshal on the decoded value directly because ToJSON()
	// relies on yaml.Node.Kind which may be incorrect in this code path.
	kind := reflect.TypeOf(v).Kind()
	if kind == reflect.Slice || kind == reflect.Map {
		j, err := json.MarshalIndent(v, "", "")
		if err != nil {
			panic(err)
		}
		return strings.TrimSpace(string(j))
	}

	return fmt.Sprintf("%v", v)
}

func (e *Example) ToJSON() string {
	if e.Value == nil {
		return "null"
	}

	// TODO: we should replace the below code with json.YAMLToJSON in the future which will format things better and wrap the examples in the docs with <pre><code></code></pre> for better output
	// But just retaining original behavior as we switch to the new openapi parser for now
	out, err := YAMLToJSONCompatibleGoType(e.Value)
	if err != nil {
		panic(err)
	}

	j, err := json.MarshalIndent(out, "", "")
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(j))
}

func (e *Example) Clone() *Example {
	cp := NewExample(e.name, e.Description, nil)
	if e.Value != nil {
		v := *e.Value
		cp.Value = &v
	}

	return cp
}

func (e *Example) isMapOrSequence() bool {
	if e.Value == nil {
		return false
	}

	kind := e.Value.Kind

	// Handle DocumentNode wrapping (common when YAML is unmarshaled)
	if kind == yaml.DocumentNode && len(e.Value.Content) > 0 {
		kind = e.Value.Content[0].Kind
	}

	return kind == yaml.MappingNode || kind == yaml.SequenceNode
}

func (e Examples) AppendExample(example *Example) Examples {
	if e == nil {
		return Examples{example}
	}

	foundIdx := slices.IndexFunc(e, func(e *Example) bool {
		return e.Name() == example.Name()
	})
	if foundIdx == -1 {
		return append(e, example)
	}

	e[foundIdx] = example

	return e
}

func YAMLToJSONCompatibleGoType(node *yaml.Node) (any, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return YAMLToJSONCompatibleGoType(node.Content[0])
	case yaml.SequenceNode:
		return handleSequenceNode(node)
	case yaml.MappingNode:
		return handleMappingNode(node)
	case yaml.ScalarNode:
		return handleScalarNode(node)
	case yaml.AliasNode:
		return YAMLToJSONCompatibleGoType(node.Alias)
	default:
		return nil, fmt.Errorf("unknown node kind: %s", yml.NodeKindToString(node.Kind))
	}
}

func handleMappingNode(node *yaml.Node) (any, error) {
	v := sequencedmap.New[string, any]()
	for i, n := range node.Content {
		if i%2 == 0 {
			continue
		}
		keyNode := node.Content[i-1]
		kv, err := YAMLToJSONCompatibleGoType(keyNode)
		if err != nil {
			return nil, err
		}

		// Handle nil key value to prevent panic in reflect.TypeOf
		if kv == nil {
			kv = "null"
		}

		if reflect.TypeOf(kv).Kind() != reflect.String {
			keyData, err := json.Marshal(kv)
			if err != nil {
				return nil, err
			}
			kv = string(keyData)
		}

		keyStr := fmt.Sprintf("%v", kv)

		// Handle YAML merge key (<<)
		if keyStr == "<<" {
			vv, err := YAMLToJSONCompatibleGoType(n)
			if err != nil {
				return nil, err
			}

			// Merge the values from the referenced map
			if mergeMap, ok := vv.(*sequencedmap.Map[string, any]); ok {
				for mergeKey, mergeValue := range mergeMap.All() {
					// Only set if the key doesn't already exist (merge keys have lower priority)
					if !v.Has(mergeKey) {
						v.Set(mergeKey, mergeValue)
					}
				}
			}
			continue
		}

		vv, err := YAMLToJSONCompatibleGoType(n)
		if err != nil {
			return nil, err
		}

		v.Set(keyStr, vv)
	}

	return v, nil
}

func handleSequenceNode(node *yaml.Node) (any, error) {
	var s []yaml.Node

	if err := node.Decode(&s); err != nil {
		return nil, err
	}

	v := make([]any, len(s))
	for i, n := range s {
		vv, err := YAMLToJSONCompatibleGoType(&n)
		if err != nil {
			return nil, err
		}

		v[i] = vv
	}

	return v, nil
}

func handleScalarNode(node *yaml.Node) (any, error) {
	var v any

	if err := node.Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
}
