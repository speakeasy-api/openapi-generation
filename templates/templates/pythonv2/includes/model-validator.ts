// Model validator files declare Pydantic @model_validator methods on
// generated models. The leading import block is parsed into
// (from, symbol) pairs that flow through ``addImport`` so they merge
// with generator-emitted imports; the body is spliced into the owning
// class body at 4-space indent.
//
// Convention: a file at .speakeasy/addons/model_validator/<modelPath>/<file>.py
// targets the generated model file at <sourcePath>/<modelPath>/<file>.py.
// Feature activates automatically when matching files are present.
//
// Supported import forms:
//
//   - ``from X import a, b, c``  (single line)
//   - ``from X import (a, b, c)``  (paren-spanning, any number of lines)
//   - ``from X import a, \\\n    b``  (trailing-backslash continuation)
//   - ``from X import a as b``  (alias passes through verbatim)
//   - ``from . import x``  (relative)
//   - ``import x`` / ``import x.y.z``  (bare, no alias)
//
// Unsupported (rejected with a warning, splice skipped for the file):
//
//   - ``import x as y``  (bare module alias). The generator's import
//     collector has no alias slot on the bare-import emission path, so
//     this form cannot be merged without a larger refactor. Workaround:
//     use ``from . import x as y`` or restructure the call site.
//
// Author-visible limits (intentionally not validated):
//
//   - Logical-line grouping uses a naive ( / [ / { depth counter that
//     does not skip string literals or comments. A stray unmatched
//     bracket inside a string or comment on an import line will swallow
//     subsequent physical lines as continuation.
//   - Leading whitespace on a ``from`` / ``import`` line is ignored
//     when classifying lines.
//   - A leading module docstring (or any non-import non-comment
//     non-blank first line) is treated as the start of the body. The
//     entire file then becomes body and no imports are extracted.

const MODEL_VALIDATOR_ROOT = ".speakeasy/addons/model_validator";

interface ModelValidatorImport {
  From: string;
  Symbols: string[];
}

interface ParsedModelValidator {
  Imports: ModelValidatorImport[];
  Body: string;
}

function modelValidatorPath(modelPath: string, fileName: string): string {
  return `${MODEL_VALIDATOR_ROOT}/${modelPath}/${fileName}.py`;
}

function readModelValidator(path: string): ParsedModelValidator | null {
  const raw = readFile(path);
  if (!raw) {
    return null;
  }
  return parseModelValidator(path, raw);
}

function parseModelValidator(
  path: string,
  raw: string,
): ParsedModelValidator | null {
  const physical = raw.split("\n");
  const imports: ModelValidatorImport[] = [];
  let bodyStartIdx = physical.length;
  let i = 0;

  while (i < physical.length) {
    const trimmed = physical[i].trim();

    if (trimmed === "" || trimmed.startsWith("#")) {
      i++;
      continue;
    }

    if (/^(from|import)\s/.test(trimmed)) {
      const end = consumeLogicalLine(physical, i);
      const logical = joinLogicalLine(physical.slice(i, end));
      const lineNumber = i + 1;
      const parsed = parseImportStatement(path, lineNumber, logical);
      if (parsed === false) {
        return null;
      }
      if (parsed !== null) {
        imports.push(parsed);
      }
      i = end;
      continue;
    }

    bodyStartIdx = i;
    break;
  }

  const body = physical
    .slice(bodyStartIdx)
    .join("\n")
    .replace(/^\n+|\s+$/g, "");

  return { Imports: imports, Body: body };
}

// Returns the exclusive index at which the logical line starting at
// ``start`` ends. Handles paren-spanning multi-line imports and
// trailing-backslash continuations.
function consumeLogicalLine(physical: string[], start: number): number {
  let depth = 0;
  for (let i = start; i < physical.length; i++) {
    const line = physical[i];
    for (const c of line) {
      if (c === "(" || c === "[" || c === "{") depth++;
      else if (c === ")" || c === "]" || c === "}") depth--;
    }
    const trimmedRight = line.replace(/\s+$/, "");
    const hasBackslash = trimmedRight.endsWith("\\");
    if (depth <= 0 && !hasBackslash) {
      return i + 1;
    }
  }
  return physical.length;
}

// Collapses physical lines into a single logical string by:
//   - dropping trailing backslash continuations
//   - replacing each physical-line newline with a single space
function joinLogicalLine(lines: string[]): string {
  return lines
    .map((line) =>
      line
        .replace(/\s*\\\s*$/, "")
        .replace(/\s+/g, " ")
        .trim(),
    )
    .filter((s) => s.length > 0)
    .join(" ");
}

// Parses one logical import statement. Returns:
//   - ModelValidatorImport on success
//   - null if the line is empty / a comment / has no symbols
//   - false if the statement is unsupported (caller skips the file)
function parseImportStatement(
  path: string,
  lineNumber: number,
  logical: string,
): ModelValidatorImport | null | false {
  // ``from X import a, b`` / ``from X import (a, b)``
  const fromMatch = logical.match(
    /^from\s+(\.+[\w\.]*|[\w\.]+)\s+import\s+(.+?)\s*(?:#.*)?$/,
  );
  if (fromMatch) {
    let symbolsRaw = fromMatch[2].trim();
    if (symbolsRaw.startsWith("(") && symbolsRaw.endsWith(")")) {
      symbolsRaw = symbolsRaw.slice(1, -1);
    }
    const symbols = symbolsRaw
      .split(",")
      .map((s) => s.replace(/\s+/g, " ").trim())
      .filter((s) => s.length > 0);
    if (symbols.length === 0) {
      return null;
    }
    return { From: fromMatch[1], Symbols: symbols };
  }

  // ``import x`` / ``import x.y.z``. ``import x as y`` is rejected.
  const importMatch = logical.match(/^import\s+([\w\.]+)\s*(?:#.*)?$/);
  if (importMatch) {
    return { From: importMatch[1], Symbols: [""] };
  }

  if (/^import\s+[\w\.]+\s+as\s+\w+/.test(logical)) {
    logWarning(
      `${path}: bare 'import X as Y' is not supported; rewrite as ` +
        `'from . import X as Y' or restructure (line ${lineNumber}: ${logical}). ` +
        `Skipping model_validator splice for this file.`,
      lineNumber,
    );
    return false;
  }

  return null;
}
