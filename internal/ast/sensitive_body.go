package ast

import "sort"

// SensitiveBodyGraph locates the values marked sensitive (x-speakeasy-param-sensitive
// or format: password) in the bodies of an operation, by their JSON location,
// so that generated code can mask them when bodies are logged.
//
// Nodes[0] is a sentinel meaning "nothing sensitive below here"; a node
// reference of 0 can be skipped. Roots are the body schemas: one per request
// content type, or one per response status code and content type. Only nodes
// that can reach a sensitive value are kept.
type SensitiveBodyGraph struct {
	Nodes []*SensitiveBodyNode
	Roots []int
}

// SensitiveBodyNode is a schema in a SensitiveBodyGraph.
type SensitiveBodyNode struct {
	// Sensitive masks the whole value. Streams are masked as a whole when any
	// of their items holds a sensitive value.
	Sensitive bool
	// Fields are object properties by wire name (JSON property, form field
	// or multipart part name), sorted by name. A property
	// that is not listed uses Values. When Values is set, declared properties
	// are listed even when they hold nothing sensitive (with Node 0), so that
	// they never fall back to Values.
	Fields []SensitiveBodyField
	// Item is the schema of array items.
	Item int
	// Values is the schema of map values and additional properties.
	Values int
	// Variants are union members; the member a value matches is not known.
	Variants []int
	// Plain reports a union member holding nothing sensitive, so a value
	// that matches none of Variants may still match the schema.
	Plain bool
}

// SensitiveBodyField is an object property of a SensitiveBodyNode.
type SensitiveBodyField struct {
	Name string
	Node int
}

// SensitiveRequestBodyGraph returns the sensitive value locations of the
// operation's request body, or nil when it holds none.
func (o *Operation) SensitiveRequestBodyGraph() *SensitiveBodyGraph {
	b := newSensitiveBodyBuilder()
	b.addRoot(o.requestBodyType())
	return b.build()
}

// SensitiveResponseBodyGraph returns the sensitive value locations of all the
// operation's response bodies, including errors, or nil when they hold none.
func (o *Operation) SensitiveResponseBodyGraph() *SensitiveBodyGraph {
	b := newSensitiveBodyBuilder()
	if o.Response != nil {
		for _, response := range o.Response.Responses {
			for _, content := range response.Content {
				if content != nil && content.Content != nil {
					b.addRoot(content.Content.Type)
				}
			}
		}
	}
	return b.build()
}

// HasSensitiveBodies reports whether the operation's request or response
// bodies hold values marked sensitive.
func (o *Operation) HasSensitiveBodies() bool {
	return o.SensitiveRequestBodyGraph() != nil || o.SensitiveResponseBodyGraph() != nil
}

func (o *Operation) requestBodyType() *TypeDef {
	if o.Request == nil {
		return nil
	}
	if o.Request.Field != nil && o.Request.Field.Type != nil {
		for _, field := range o.Request.Field.Type.Fields {
			if field.Annotations.Has(AnnotationTypeRequest) {
				return field.Type
			}
		}
	}
	if o.Request.IsRequestBody && o.Request.RequestBody != nil {
		return o.Request.RequestBody.Type
	}
	return nil
}

func isSensitiveBodyType(t *TypeDef) bool {
	if t.Format == "password" {
		return true
	}
	if t.Extensions == nil {
		return false
	}
	marker := t.Extensions.All["x-speakeasy-param-sensitive"]
	return marker == true || marker == "true"
}

type sensitiveBodyBuilder struct {
	nodes []*sensitiveBodyBuildNode
	roots []int
}

type sensitiveBodyBuildNode struct {
	sensitive bool
	stream    bool
	fields    []SensitiveBodyField
	item      int
	values    int
	variants  []int
}

func newSensitiveBodyBuilder() *sensitiveBodyBuilder {
	return &sensitiveBodyBuilder{nodes: []*sensitiveBodyBuildNode{{}}}
}

// addRoot walks a body schema along the edges given by TypeDef.Children.
// Custom types are walked once per root, by identity and by registration ID
// (which also matches the truncated copies that break recursive references),
// so that recursive schemas terminate.
func (b *sensitiveBodyBuilder) addRoot(t *TypeDef) {
	if !containsSensitiveBodyType(t) {
		return
	}
	walkedTypes := map[*TypeDef]int{}
	walkedIDs := map[string]int{}
	var walk func(t *TypeDef) int
	walk = func(t *TypeDef) int {
		if t == nil {
			return 0
		}
		if id, ok := walkedTypes[t]; ok {
			return id
		}
		typeID := ""
		if t.IsCustomType() && t.ContextStack != nil {
			typeID = t.GetRegistrationID() + " name:" + t.Name
			if id, ok := walkedIDs[typeID]; ok {
				return id
			}
		}
		id := len(b.nodes)
		node := &sensitiveBodyBuildNode{}
		b.nodes = append(b.nodes, node)
		walkedTypes[t] = id
		if typeID != "" {
			walkedIDs[typeID] = id
		}
		if isSensitiveBodyType(t) {
			node.sensitive = true
			return id
		}
		for field, child := range t.Children {
			switch {
			case field == nil:
				switch t.Type {
				case DataTypeArray, DataTypeSet:
					node.item = walk(child)
				case DataTypeMap:
					node.values = walk(child)
				case DataTypeEventStream, DataTypeJsonL:
					node.stream = true
					node.item = walk(child)
				case DataTypeUnion:
					node.variants = append(node.variants, walk(child))
				}
			case t.Type != DataTypeClass && t.Type != DataTypeError:
			case field.IsAdditionalProperties:
				if child != nil {
					node.values = walk(child.ItemType)
				}
			default:
				node.fields = append(node.fields, SensitiveBodyField{Name: sensitiveBodyFieldName(field)})
				if field.Const == nil {
					node.fields[len(node.fields)-1].Node = walk(child)
				}
			}
		}
		return id
	}
	b.roots = append(b.roots, walk(t))
}

func sensitiveBodyFieldName(field *FieldDef) string {
	if field.OriginalName != "" {
		return field.OriginalName
	}
	return field.Name
}

// containsSensitiveBodyType reports whether t or any type below it is marked
// sensitive, so that schemas without sensitive values are not walked further.
func containsSensitiveBodyType(t *TypeDef) bool {
	for _, typ := range t.Walk() {
		if isSensitiveBodyType(typ) {
			return true
		}
	}
	return false
}

// build prunes the nodes that cannot reach a sensitive value and renumbers
// the rest.
func (b *sensitiveBodyBuilder) build() *SensitiveBodyGraph {
	reachable := make([]bool, len(b.nodes))
	for i, node := range b.nodes {
		reachable[i] = node.sensitive
	}
	for changed := true; changed; {
		changed = false
		for i, node := range b.nodes {
			if reachable[i] {
				continue
			}
			children := append([]int{node.item, node.values}, node.variants...)
			for _, field := range node.fields {
				children = append(children, field.Node)
			}
			for _, child := range children {
				if child != 0 && reachable[child] {
					reachable[i] = true
					changed = true
					break
				}
			}
		}
	}

	remap := make([]int, len(b.nodes))
	graph := &SensitiveBodyGraph{Nodes: []*SensitiveBodyNode{{}}}
	for i := 1; i < len(b.nodes); i++ {
		if reachable[i] {
			remap[i] = len(graph.Nodes)
			graph.Nodes = append(graph.Nodes, &SensitiveBodyNode{})
		}
	}
	for _, root := range b.roots {
		if root != 0 && reachable[root] {
			graph.Roots = append(graph.Roots, remap[root])
		}
	}
	if len(graph.Roots) == 0 {
		return nil
	}

	for i, node := range b.nodes {
		if i == 0 || !reachable[i] {
			continue
		}
		out := graph.Nodes[remap[i]]
		if node.sensitive || node.stream {
			out.Sensitive = true
			continue
		}
		out.Values = remap[node.values]
		for _, field := range node.fields {
			if out.Values != 0 || remap[field.Node] != 0 {
				out.Fields = append(out.Fields, SensitiveBodyField{Name: field.Name, Node: remap[field.Node]})
			}
		}
		sort.SliceStable(out.Fields, func(a, b int) bool { return out.Fields[a].Name < out.Fields[b].Name })
		out.Item = remap[node.item]
		for _, variant := range node.variants {
			if remap[variant] != 0 {
				out.Variants = append(out.Variants, remap[variant])
			} else {
				out.Plain = true
			}
		}
	}
	return graph
}
