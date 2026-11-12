// Escapes with the same meaning in Go's regexp and ECMA-262 Unicode mode.
const portableEscapes = new Set("dDwWbBtnrfv^$\\.*+?()[]{}|/".split(""));
const portableClassEscapes = new Set("dDwWtnrfv^$\\.*+?()[]{}|/-".split(""));

/**
 * Consumes a `\` escape at pattern[i] and returns the index after it, or -1
 * when the escape is outside the portable set.
 */
function portableEscapeEnd(
  pattern: string,
  i: number,
  allowed: Set<string>,
): number {
  const next = pattern[i + 1];
  if (next === "x") {
    return /^[0-9A-Fa-f]{2}$/.test(pattern.slice(i + 2, i + 4)) ? i + 4 : -1;
  }
  return next !== undefined && allowed.has(next) ? i + 2 : -1;
}

/**
 * Whether a schema pattern means the same thing in Go's regexp and in ECMA-262
 * with the "u" flag (the JSON Schema dialect), so a generated client can
 * enforce it without rejecting a value the API accepts. A pattern qualifies
 * when it stays within the interoperable subset JSON Schema recommends —
 * literals, `.`, character classes and ranges, `\d \w \b` and their negations,
 * control and `\xHH` escapes, anchors, `(...)`/`(?:...)` groups, alternation,
 * and `* + ? {n,m}` quantifiers (greedy or lazy) — and Go compiles it. That
 * excludes, for example, `\s` (ASCII-only in Go), `\p{...}`, lookarounds,
 * backreferences, and POSIX classes. Go's `.` also matches `\r` and U+2028, so
 * it may accept a value ECMA-262 rejects; the server still decides those.
 */
function isPortablePattern(pattern: string): boolean {
  if (!isRE2Regex(pattern)) return false;
  let i = 0;
  while (i < pattern.length) {
    const c = pattern[i];
    if (c === "\\") {
      i = portableEscapeEnd(pattern, i, portableEscapes);
      if (i < 0) return false;
    } else if (c === "(") {
      if (pattern[i + 1] === "?" && pattern[i + 2] !== ":") return false;
      i += pattern[i + 1] === "?" ? 3 : 1;
    } else if (c === "{") {
      // Go reads a bound with a leading zero ({01}) as literal text.
      const quantifier = /^\{(0|[1-9]\d*)(,(0|[1-9]\d*)?)?\}/.exec(
        pattern.slice(i),
      );
      if (!quantifier) return false;
      i += quantifier[0].length;
    } else if (c === "}" || c === "]") {
      return false;
    } else if (c === "[") {
      i++;
      if (pattern[i] === "^") i++;
      if (pattern[i] === "]") return false;
      while (i < pattern.length && pattern[i] !== "]") {
        if (pattern[i] === "[") return false;
        if (pattern[i] === "\\") {
          i = portableEscapeEnd(pattern, i, portableClassEscapes);
          if (i < 0) return false;
        } else {
          i++;
        }
      }
      if (i >= pattern.length) return false;
      i++;
    } else {
      i++;
    }
  }
  return true;
}

/**
 * The schema pattern of a string type when it is portable, otherwise "".
 * Validations.Pattern is pointer-valued in the AST, so it is coerced to a JS
 * string first.
 */
function portableStringPattern(typeDef: TypeDef): string {
  const raw = typeDef.Validations?.Pattern;
  const pattern = raw == null ? "" : String(raw);
  return pattern && isPortablePattern(pattern) ? pattern : "";
}
