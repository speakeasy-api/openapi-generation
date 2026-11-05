// @ts-ignore
/**
 * Wraps `expr` in double braces: {{expr}}.
 *
 * Intended for use inside Roslyn source generator `.cs.stmpl` files that
 * contain C# `$$"""..."""` raw interpolated string literals. In those strings,
 * `{{expr}}` is the C# interpolation syntax (double-dollar prefix requires
 * doubled braces). Because Go templates also use `{{ }}` as delimiters, a
 * bare `{{expr}}` inside a `.stmpl` would be parsed as a Go template
 * directive. Using `interp` avoids that conflict while keeping the template
 * source readable.
 *
 * Example:
 *   `.stmpl` source:  `namespace {{interp "namespaceName"}};`
 *   Generated output: `namespace {{namespaceName}};`   ← valid C# $$-interpolation
 */
function interp(expr: string): string {
  return `{{${expr}}}`;
}

registerTemplateFunc("interp", interp);
