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

const patternExamplePool =
  "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-_.~ ";
const patternLiteralEscapes: Record<string, string> = {
  t: "\t",
  n: "\n",
  r: "\r",
  f: "\f",
  v: "\v",
};

/**
 * A deterministic value matching a portable pattern (see isPortablePattern),
 * or undefined when none could be built. Alternation takes the first branch,
 * unbounded repetition produces three items, and character sets draw from a
 * small printable pool.
 */
function exampleForPortablePattern(pattern: string): string | undefined {
  let i = 0;
  let picks = 0;

  const classEnd = (start: number): number => {
    let j = start + 1;
    if (pattern[j] === "^") j++;
    while (j < pattern.length && pattern[j] !== "]") {
      j += pattern[j] === "\\" ? (pattern[j + 1] === "x" ? 4 : 2) : 1;
    }
    return j;
  };

  const parseAtom = (): (() => string) | undefined => {
    const c = pattern[i];
    if (c === "^" || c === "$") {
      i++;
      return () => "";
    }
    if (c === "(") {
      i += pattern[i + 1] === "?" ? 3 : 1;
      const inner = parseAlternation();
      if (inner === undefined || pattern[i] !== ")") return undefined;
      i++;
      return () => inner;
    }
    let set: string;
    if (c === "[") {
      const end = classEnd(i);
      set = pattern.slice(i, end + 1);
      i = end + 1;
    } else if (c === "\\") {
      const next = pattern[i + 1] ?? "";
      if (next === "x") {
        const ch = String.fromCharCode(
          parseInt(pattern.slice(i + 2, i + 4), 16),
        );
        i += 4;
        return () => ch;
      }
      i += 2;
      if (next === "b" || next === "B") return () => "";
      const literal = patternLiteralEscapes[next];
      if (literal !== undefined) return () => literal;
      if (!"dDwW".includes(next)) return () => next;
      set = `\\${next}`;
    } else if (c === ".") {
      set = ".";
      i++;
    } else {
      i++;
      return () => c;
    }
    const setRE = new RegExp(`^${set}$`, "u");
    const choices = patternExamplePool.split("").filter((ch) => setRE.test(ch));
    if (choices.length === 0) return undefined;
    return () => choices[picks++ % choices.length]!;
  };

  const parseQuantifier = (): [number, number] => {
    let range: [number, number];
    const c = pattern[i];
    if (c === "*") {
      range = [0, Infinity];
      i++;
    } else if (c === "+") {
      range = [1, Infinity];
      i++;
    } else if (c === "?") {
      range = [0, 1];
      i++;
    } else if (c === "{") {
      const bound = /^\{(\d+)(,(\d*))?\}/.exec(pattern.slice(i));
      if (!bound) return [1, 1];
      const min = Number(bound[1]);
      const max =
        bound[2] === undefined
          ? min
          : bound[3] === ""
          ? Infinity
          : Number(bound[3]);
      range = [min, max];
      i += bound[0].length;
    } else {
      return [1, 1];
    }
    if (pattern[i] === "?") i++;
    return range;
  };

  const parseSequence = (): string | undefined => {
    let out = "";
    while (i < pattern.length && pattern[i] !== "|" && pattern[i] !== ")") {
      const atom = parseAtom();
      if (!atom) return undefined;
      const [min, max] = parseQuantifier();
      const count = Math.min(max, Math.max(min, 3));
      for (let n = 0; n < count; n++) out += atom();
    }
    return out;
  };

  function parseAlternation(): string | undefined {
    const first = parseSequence();
    while (first !== undefined && pattern[i] === "|") {
      i++;
      if (parseSequence() === undefined) return undefined;
    }
    return first;
  }

  let example: string | undefined;
  try {
    example = parseAlternation();
  } catch {
    return undefined;
  }
  if (example === undefined || i !== pattern.length) return undefined;
  return new RegExp(pattern, "u").test(example) ? example : undefined;
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
