package cycles

import (
	"fmt"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/mermaid"
)

// Cycle represents a strongly connected component (SCC) as a set of TypeDef pointers
type Cycle map[*ast.TypeDef]bool

// ToMermaid makes it easy to visualize the cycle, only used for debugging
func (c Cycle) ToMermaid() string {
	m := mermaid.NewMermaidAST()
	for t := range c {
		m.Visit(t)
	}
	return m.Render(mermaid.RenderOpts{
		ShouldIncludeEdge: func(edge mermaid.MermaidEdge) bool {
			return c[edge.From] && c[edge.To]
		},
	})
}

type CycleDetector struct {
	// Cycles found in the AST
	Cycles []Cycle
	// visited only traverse each TypeDef once.
	visited map[*ast.TypeDef]bool
	// lookup Map from TypeDef to the cycle it belongs to if any.
	lookup map[*ast.TypeDef]Cycle
	mu     sync.Mutex
}

// CycleDetector detects cycles otherwise known as strongly connected components (SCCs) using Tarjan's algorithm
// Cycles differ from SCCs in one subtle way: a single node is considered an SCC even if it's not self-referential
// we discard these trivial SCCs are they are not cycles.
func New() *CycleDetector {
	return &CycleDetector{
		visited: make(map[*ast.TypeDef]bool),
		Cycles:  make([]Cycle, 0),
		lookup:  make(map[*ast.TypeDef]Cycle),
	}
}

func (c *CycleDetector) DetectCycles(t ast.TypeDefs) []Cycle {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range t {
		c.traverse(t, nil)
	}
	return c.Cycles
}

// IsPartOfCycle checks if a TypeDef is part of a cycle
func (c *CycleDetector) IsPartOfCycle(t *ast.TypeDef) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.visited[t] {
		c.Traverse(t)
	}
	return c.lookup[t] != nil
}

// AreCircular checks if two TypeDefs are part of the same cycle.
func (c *CycleDetector) AreCircular(t1 *ast.TypeDef, t2 *ast.TypeDef) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.visited[t1] {
		c.Traverse(t1)
	}
	if !c.visited[t2] {
		c.Traverse(t2)
	}
	t1Cycle := c.lookup[t1]
	return t1Cycle != nil && t1Cycle[t2]
}

func (c *CycleDetector) Traverse(t *ast.TypeDef) {
	c.traverse(t, nil)
}

// state maintains the traversal state for Tarjan's algorithm
type state struct {
	onStack  map[*ast.TypeDef]bool
	stack    ast.TypeDefs
	lowLinks map[*ast.TypeDef]int
	id       int // a pre-order counter, used to identify the "root" of the SCC
}

// traverse performs a DFS to detect SCCs using Tarjan's algorithm
// See: https://en.wikipedia.org/wiki/Tarjan's_strongly_connected_components_algorithm
// I also found this video helpful for understanding the algorithm: https://www.youtube.com/watch?v=wUgWX0nc4NY
func (c *CycleDetector) traverse(t *ast.TypeDef, st *state) {
	if c.visited[t] {
		return
	}
	c.visited[t] = true

	if st == nil {
		st = &state{
			onStack:  make(map[*ast.TypeDef]bool),
			stack:    make(ast.TypeDefs, 0),
			lowLinks: make(map[*ast.TypeDef]int),
			id:       0,
		}
	}

	currentIndex := len(st.stack)
	currentID := st.id
	st.stack = append(st.stack, t)
	st.onStack[t] = true
	st.lowLinks[t] = currentID
	st.id++

	for _, child := range t.Children {
		c.traverse(child, st)
		if st.onStack[child] {
			st.lowLinks[t] = min(st.lowLinks[t], st.lowLinks[child])
		}
	}

	if st.lowLinks[t] == currentID {
		cycle := make(ast.TypeDefs, 0, len(st.stack)-currentIndex)
		cycle = append(cycle, st.stack[currentIndex:]...)
		for _, t := range cycle {
			st.onStack[t] = false
		}
		c.addCycle(cycle)
		st.stack = st.stack[:currentIndex]
	}
}

// addCycle records a cycle that has been found
func (c *CycleDetector) addCycle(types ast.TypeDefs) {
	// Construct the cycle set
	cycle := make(Cycle)
	for _, t := range types {
		cycle[t] = true
	}

	if len(cycle) == 1 {
		// Check if the single node has a self-reference
		node := types[0]
		hasSelfRef := false
		for _, child := range node.Children {
			if child == node {
				hasSelfRef = true
				break
			}
		}
		if !hasSelfRef {
			// This is a strongly connected component
			// but not a self-referencing cycle
			return
		}
	}

	if env.DebugCycles() {
		fmt.Println("Cycle:")
		fmt.Println(cycle.ToMermaid())
	}

	c.Cycles = append(c.Cycles, cycle)

	// Index the cycle by all its members
	for _, t := range types {
		c.lookup[t] = cycle
	}
}
