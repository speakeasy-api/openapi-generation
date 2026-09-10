package mermaid

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type MermaidAST struct {
	visited map[*ast.TypeDef]bool
	edges   []MermaidEdge
}

type MermaidEdge struct {
	From *ast.TypeDef
	To   *ast.TypeDef
}

func NewMermaidAST() *MermaidAST {
	return &MermaidAST{
		visited: make(map[*ast.TypeDef]bool),
		edges:   make([]MermaidEdge, 0),
	}
}

func (m *MermaidAST) Visit(t *ast.TypeDef) {
	if m.visited[t] {
		return
	}
	m.visited[t] = true
	for _, child := range t.Children {
		m.edges = append(m.edges, MermaidEdge{t, child})
		m.Visit(child)
	}
}

type RenderOpts struct {
	ShouldIncludeEdge func(edge MermaidEdge) bool
}

func (m *MermaidAST) Render(opts ...RenderOpts) string {
	opt := RenderOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	var sb strings.Builder
	sb.WriteString("flowchart LR\n")
	pointerID := func(t *ast.TypeDef) string {
		return fmt.Sprintf("%p", t)
	}

	nameOrType := func(t *ast.TypeDef) string {
		if t.Name == "" {
			return string(t.Type)
		}
		return t.Name
	}

	for _, edge := range m.edges {
		if opt.ShouldIncludeEdge != nil && !opt.ShouldIncludeEdge(edge) {
			continue
		}
		fmt.Fprintf(&sb, "%s[%q] --> %s[%q]\n", pointerID(edge.From), nameOrType(edge.From), pointerID(edge.To), nameOrType(edge.To))
	}
	return sb.String()
}
