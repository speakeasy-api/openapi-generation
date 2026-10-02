package snapshots

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

const cliStringConstraintsSnapshotSpec = `openapi: 3.1.0
info:
  title: Items CLI
  version: 0.1.0
paths:
  /items/{itemId}:
    get:
      operationId: getItem
      tags: [items]
      parameters:
        - name: itemId
          in: path
          required: true
          schema:
            type: string
            pattern: "^[a-z0-9-]+$"
            maxLength: 8
        - name: patternOnly
          in: query
          schema:
            type: string
            pattern: "^[a-z]+$"
        - name: maxOnly
          in: query
          schema:
            type: string
            maxLength: 3
        - name: maxZero
          in: query
          schema:
            type: string
            maxLength: 0
        - name: plain
          in: query
          schema:
            type: string
        - $ref: "#/components/parameters/RefCode"
        - name: escaped
          in: query
          schema:
            type: string
            pattern: '^"\\[a-z]{2}é$'
        - name: lookahead
          in: query
          schema:
            type: string
            pattern: "^(?!x)[a-z]+$"
            maxLength: 3
        - name: backreference
          in: query
          schema:
            type: string
            pattern: '^(a)\1$'
        - name: hugeRepeat
          in: query
          schema:
            type: string
            pattern: "^a{1001}$"
        - name: whitespace
          in: query
          schema:
            type: string
            pattern: '^[a-z\s]+$'
        - name: posixClass
          in: query
          schema:
            type: string
            pattern: "^[[:alpha:]]+$"
        - name: unicodeProperty
          in: query
          schema:
            type: string
            pattern: '^\p{L}+$'
        - name: dotRepeat
          in: query
          schema:
            type: string
            pattern: "^.{2}$"
        - name: alternation
          in: query
          schema:
            type: string
            pattern: '^(?:ab|cd)+\b[\w.-]*\x41?$'
        - name: belEscape
          in: query
          schema:
            type: string
            pattern: '^\a$'
        - name: loneBrace
          in: query
          schema:
            type: string
            pattern: "^a{$"
        - name: leadingBracketClass
          in: query
          schema:
            type: string
            pattern: "^[]a]$"
        - name: zeroBound
          in: query
          schema:
            type: string
            pattern: "^a{0,1}$"
        - name: leadingZeroBound
          in: query
          schema:
            type: string
            pattern: "^a{01}$"
        - name: leadingZeroUpperBound
          in: query
          schema:
            type: string
            pattern: "^a{1,02}$"
      responses:
        '200':
          description: ok
components:
  parameters:
    RefCode:
      name: code
      in: query
      schema:
        $ref: "#/components/schemas/Code"
  schemas:
    Code:
      type: string
      pattern: "^[A-Z]{3}$"
      maxLength: 3
`

// flagConstraint is the pattern/maxLength a generated FlagMeta literal
// declares, decoded from the Go source.
type flagConstraint struct {
	Pattern      *string
	HasMaxLength bool
	MaxLength    *string
}

func parseFlagConstraints(t *testing.T, path string) map[string]flagConstraint {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	require.NoError(t, err)

	constraints := map[string]flagConstraint{}
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		var name string
		var c flagConstraint
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			if ident, ok := kv.Value.(*ast.Ident); ok && key.Name == "HasMaxLength" {
				c.HasMaxLength = ident.Name == "true"
				continue
			}
			value, ok := kv.Value.(*ast.BasicLit)
			if !ok {
				continue
			}
			switch key.Name {
			case "FlagName":
				name, err = strconv.Unquote(value.Value)
				require.NoError(t, err)
			case "Pattern":
				pattern, err := strconv.Unquote(value.Value)
				require.NoError(t, err)
				c.Pattern = &pattern
			case "MaxLength":
				maxLength := value.Value
				c.MaxLength = &maxLength
			}
		}
		if name != "" {
			constraints[name] = c
		}
		return true
	})
	return constraints
}

// String flags carry the schema's pattern and maxLength so the CLI can reject
// non-conforming values before sending. Patterns Go's regexp would match
// differently from ECMA-262, or cannot compile, are left to the server.
func TestSnapCLIStringConstraints(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: cliStringConstraintsSnapshotSpec,
		GenYaml: `cli:
  packageName: github.com/example/items-cli
  cliName: items
  envVarPrefix: ITEMS
`,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()
			got := parseFlagConstraints(t, filepath.Join(tempDir, "internal", "cli", "items", "getitem.go"))
			require.Equal(t, map[string]flagConstraint{
				"item-id":          {Pattern: ptr("^[a-z0-9-]+$"), HasMaxLength: true, MaxLength: ptr("8")},
				"pattern-only":     {Pattern: ptr("^[a-z]+$")},
				"max-only":         {HasMaxLength: true, MaxLength: ptr("3")},
				"max-zero":         {HasMaxLength: true, MaxLength: ptr("0")},
				"plain":            {},
				"code":             {Pattern: ptr("^[A-Z]{3}$"), HasMaxLength: true, MaxLength: ptr("3")},
				"escaped":          {Pattern: ptr(`^"\\[a-z]{2}é$`)},
				"dot-repeat":       {Pattern: ptr("^.{2}$")},
				"alternation":      {Pattern: ptr(`^(?:ab|cd)+\b[\w.-]*\x41?$`)},
				"lookahead":        {HasMaxLength: true, MaxLength: ptr("3")},
				"backreference":    {},
				"huge-repeat":      {},
				"whitespace":       {},
				"posix-class":      {},
				"unicode-property": {},
				"bel-escape":       {},
				"lone-brace":       {},
				// ECMA-262 reads [] as an empty class; Go reads ] as a member.
				"leading-bracket-class": {},
				"zero-bound":            {Pattern: ptr("^a{0,1}$")},
				// Go reads a quantifier bound with a leading zero as literal text.
				"leading-zero-bound":       {},
				"leading-zero-upper-bound": {},
			}, got)
		},
	})
}
