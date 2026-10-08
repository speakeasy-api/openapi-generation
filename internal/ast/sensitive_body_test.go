package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sensitiveString() *TypeDef {
	return &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{All: map[string]any{"x-speakeasy-param-sensitive": true}}}
}

func plainString() *TypeDef {
	return &TypeDef{Type: DataTypeString}
}

func class(name string, fields ...*FieldDef) *TypeDef {
	return &TypeDef{Type: DataTypeClass, Name: name, Fields: fields}
}

func field(name string, typ *TypeDef) *FieldDef {
	return &FieldDef{Name: name, OriginalName: name, Type: typ}
}

func responseOperation(types ...*TypeDef) *Operation {
	response := &SubResponse{Code: []string{"200"}}
	for _, typ := range types {
		response.Content = append(response.Content, &ResponseBodyContent{Content: &FieldDef{Type: typ}})
	}
	return &Operation{BaseOperation: BaseOperation{Response: &Response{Responses: SubResponses{response}}}}
}

// sensitiveChild follows a JSON property from the nodes, as a masker would.
func sensitiveChild(g *SensitiveBodyGraph, nodes []int, name string) []int {
	var children []int
	for _, id := range sensitiveVariants(g, nodes) {
		node := g.Nodes[id]
		child := node.Values
		for _, f := range node.Fields {
			if f.Name == name {
				child = f.Node
			}
		}
		if child != 0 {
			children = append(children, child)
		}
	}
	return children
}

func sensitiveItems(g *SensitiveBodyGraph, nodes []int) []int {
	var items []int
	for _, id := range sensitiveVariants(g, nodes) {
		if item := g.Nodes[id].Item; item != 0 {
			items = append(items, item)
		}
	}
	return items
}

func sensitiveVariants(g *SensitiveBodyGraph, nodes []int) []int {
	var out []int
	seen := map[int]bool{}
	var visit func(int)
	visit = func(id int) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, id)
		for _, v := range g.Nodes[id].Variants {
			visit(v)
		}
	}
	for _, id := range nodes {
		visit(id)
	}
	return out
}

func isSensitiveAt(g *SensitiveBodyGraph, nodes []int) bool {
	for _, id := range sensitiveVariants(g, nodes) {
		if g.Nodes[id].Sensitive {
			return true
		}
	}
	return false
}

func TestSensitiveResponseBodyGraph(t *testing.T) {
	t.Parallel()

	tree := class("Tree", field("label", plainString()), field("secret", sensitiveString()))
	tree.Fields = append(tree.Fields, field("children", &TypeDef{Type: DataTypeArray, ItemType: tree}))
	labels := class("Labels", field("visible", plainString()), &FieldDef{Name: "AdditionalProperties", IsAdditionalProperties: true, Type: &TypeDef{Type: DataTypeMap, ItemType: sensitiveString()}})
	password := &TypeDef{Type: DataTypeString, Format: "password"}
	payload := class("Payload",
		field("name", plainString()),
		field("token", sensitiveString()),
		&FieldDef{Name: "Password", OriginalName: "pass_word", Type: password},
		field("nested", class("Nested", field("label", plainString()), field("secret", sensitiveString()))),
		field("tokens", &TypeDef{Type: DataTypeArray, ItemType: sensitiveString()}),
		field("labels", labels),
		field("credential", &TypeDef{Type: DataTypeUnion, Name: "Credential", AssociatedTypes: TypeDefs{
			class("APIKey", field("key", sensitiveString())),
			class("Basic", field("user", plainString())),
		}}),
		field("tree", tree),
		field("unrelated", class("Unrelated", field("secret", plainString()))),
	)
	errorType := &TypeDef{Type: DataTypeError, Name: "Error", Fields: Fields{field("message", plainString()), field("api_key", sensitiveString())}}

	g := responseOperation(payload, errorType).SensitiveResponseBodyGraph()
	require.NotNil(t, g)
	require.Len(t, g.Roots, 2)
	root := g.Roots

	at := func(nodes []int, path ...string) []int {
		for _, step := range path {
			if step == "[]" {
				nodes = sensitiveItems(g, nodes)
			} else {
				nodes = sensitiveChild(g, nodes, step)
			}
		}
		return nodes
	}

	for _, path := range [][]string{
		{"token"},
		{"pass_word"},
		{"nested", "secret"},
		{"tokens", "[]"},
		{"labels", "undeclared"},
		{"credential", "key"},
		{"tree", "secret"},
		{"tree", "children", "[]", "children", "[]", "secret"},
		{"api_key"},
	} {
		assert.True(t, isSensitiveAt(g, at(root, path...)), "%v should be sensitive", path)
	}
	for _, path := range [][]string{
		{"name"},
		{"Password"},
		{"nested", "label"},
		{"labels", "visible"},
		{"credential", "user"},
		{"tree", "label"},
		{"unrelated", "secret"},
		{"message"},
	} {
		assert.False(t, isSensitiveAt(g, at(root, path...)), "%v should not be sensitive", path)
	}
}

func TestSensitiveBodyGraphStreamsAndAbsence(t *testing.T) {
	t.Parallel()

	stream := &TypeDef{Type: DataTypeJsonL, ItemType: class("Event", field("token", sensitiveString()))}
	g := responseOperation(stream).SensitiveResponseBodyGraph()
	require.NotNil(t, g)
	assert.True(t, isSensitiveAt(g, g.Roots), "a stream holding sensitive values is sensitive as a whole")

	plain := class("Plain", field("secret", plainString()))
	assert.Nil(t, responseOperation(plain).SensitiveResponseBodyGraph())
	assert.Nil(t, responseOperation(&TypeDef{Type: DataTypeEventStream, ItemType: plain}).SensitiveResponseBodyGraph())
	assert.False(t, responseOperation(plain).HasSensitiveBodies())
}

func TestSensitiveRequestBodyGraph(t *testing.T) {
	t.Parallel()

	body := class("Body", field("name", plainString()), field("secret", sensitiveString()))
	direct := &Operation{BaseOperation: BaseOperation{Request: &Request{IsRequestBody: true, RequestBody: &FieldDef{Type: body}}}}
	g := direct.SensitiveRequestBodyGraph()
	require.NotNil(t, g)
	assert.True(t, isSensitiveAt(g, sensitiveChild(g, g.Roots, "secret")))
	assert.True(t, direct.HasSensitiveBodies())

	wrapped := &Operation{BaseOperation: BaseOperation{Request: &Request{Field: &FieldDef{Type: class("Request",
		field("id", plainString()),
		&FieldDef{Name: "Body", Type: body, Annotations: Annotations{&RequestAnnotation{}}},
	)}}}}
	g = wrapped.SensitiveRequestBodyGraph()
	require.NotNil(t, g)
	assert.True(t, isSensitiveAt(g, sensitiveChild(g, g.Roots, "secret")))

	assert.Nil(t, (&Operation{}).SensitiveRequestBodyGraph())
}

func TestSensitiveBodyGraphPlainUnionVariants(t *testing.T) {
	t.Parallel()

	union := func(name string, variants ...*TypeDef) *TypeDef {
		return &TypeDef{Type: DataTypeUnion, Name: name, AssociatedTypes: variants}
	}
	recursivePlain := union("RecursivePlain", plainString())
	recursivePlain.AssociatedTypes = append(recursivePlain.AssociatedTypes, recursivePlain)
	body := class("Body",
		field("mixed", union("Mixed", &TypeDef{Type: DataTypeArray, ItemType: sensitiveString()}, class("Label", field("label", plainString())))),
		field("text", union("Text", class("APIKey", field("key", sensitiveString())), union("Scalars", plainString(), &TypeDef{Type: DataTypeInteger}))),
		field("flag", union("Flag", class("APIKey", field("key", sensitiveString())), &TypeDef{Type: DataTypeBoolean})),
		field("loop", union("Loop", class("APIKey", field("key", sensitiveString())), recursivePlain)),
		field("sensitive", union("Sensitive", class("APIKey", field("key", sensitiveString())), sensitiveString())),
	)
	g := responseOperation(body).SensitiveResponseBodyGraph()
	require.NotNil(t, g)

	mixed := sensitiveChild(g, g.Roots, "mixed")
	require.Len(t, mixed, 1)
	assert.True(t, g.Nodes[mixed[0]].PlainObject, "a plain object member is recorded")
	assert.False(t, g.Nodes[mixed[0]].PlainArray || g.Nodes[mixed[0]].PlainString || g.Nodes[mixed[0]].PlainNumber || g.Nodes[mixed[0]].PlainBoolean)
	assert.Len(t, g.Nodes[mixed[0]].Variants, 1)

	text := sensitiveChild(g, g.Roots, "text")
	require.Len(t, text, 1)
	assert.True(t, g.Nodes[text[0]].PlainString && g.Nodes[text[0]].PlainNumber, "members of a plain nested union are recorded")
	assert.False(t, g.Nodes[text[0]].PlainObject || g.Nodes[text[0]].PlainArray || g.Nodes[text[0]].PlainBoolean)

	flag := sensitiveChild(g, g.Roots, "flag")
	require.Len(t, flag, 1)
	assert.True(t, g.Nodes[flag[0]].PlainBoolean)
	assert.False(t, g.Nodes[flag[0]].PlainString || g.Nodes[flag[0]].PlainNumber, "a boolean member does not accept text or numbers")

	loop := sensitiveChild(g, g.Roots, "loop")
	require.Len(t, loop, 1)
	assert.True(t, g.Nodes[loop[0]].PlainString, "a recursive plain union is followed once")

	sensitive := sensitiveChild(g, g.Roots, "sensitive")
	require.Len(t, sensitive, 1)
	node := g.Nodes[sensitive[0]]
	assert.False(t, node.PlainObject || node.PlainArray || node.PlainString || node.PlainNumber || node.PlainBoolean, "a union whose members all hold sensitive values has no plain member")
}
