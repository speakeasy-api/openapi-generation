require("go/sanitization.ts");

// CLI framework files that are generated with hardcoded names in the same
// directories as operation command files.  If an operation's sanitized file
// name matches one of these, it must be suffixed to avoid overwriting.
const cliReservedFileNames = [
  "root", // subroot.go.stmpl → root.go (group command registration)
  "configure", // configure.go.stmpl → configure.go
  "whoami", // whoami.go.stmpl → whoami.go
  "masking", // masking.go.stmpl → masking.go
  "version", // version.go.stmpl → version.go
];

/**
 * Sanitize a CLI operation file name.  Wraps the base Go sanitizeFileName
 * and additionally checks CLI-reserved names (root, configure, etc.).
 */
function sanitizeCLIOpFileName(name: string): string {
  let result = sanitizeFileName(name);
  if (cliReservedFileNames.includes(result)) {
    result = `${result}_`;
  }
  return result;
}
registerTemplateFunc("sanitizeCLIOpFileName", sanitizeCLIOpFileName);

// Package names that collide with CLI framework imports or Go conventions.
// "documentation" — Go toolchain ignores 'package documentation' files.
// "config" — collides with internal/config (credential management package).
const cliReservedPkgNames = ["documentation", "config"];

// Sanitize a Go package/directory name for CLI command groups.
function sanitizeCLIPkgName(name: string): string {
  let result = sanitizeFile(name, "").toLowerCase();
  if (cliReservedPkgNames.includes(result)) {
    result += "cmds";
  }
  return result;
}
registerTemplateFunc("sanitizeCLIPkgName", sanitizeCLIPkgName);

/**
 * Sanitize the CLI binary name for use in file paths, directory names,
 * and display. Reads from config and preserves kebab-case (hyphens)
 * which is idiomatic for CLI binary names (e.g., "my-cli", "docker-compose").
 * Only allows lowercase alphanumeric characters and hyphens.
 */
function sanitizeCliName(): string {
  const name = context.Global.Config.CliName || "cli";
  return name.toLowerCase().replace(/[^a-z0-9-]/g, "") || "cli";
}
registerTemplateFunc("sanitizeCliName", sanitizeCliName);

/**
 * Normalize plural acronyms before kebab-case conversion.
 * Converts patterns like "IDs" → "Ids", "URLs" → "Urls" so the caser
 * treats them as single words instead of splitting "IDs" into "I-ds".
 */
function normalizePluralAcronyms(name: string): string {
  return name.replace(/([A-Z]{2,})s(?=[A-Z]|$)/g, (match) => {
    return match[0] + match.slice(1, -1).toLowerCase() + match.slice(-1);
  });
}

function sanitizeCLICommand(name: string): string {
  return caser().ToKebab(sanitizeName(normalizePluralAcronyms(name)));
}
registerTemplateFunc("sanitizeCLICommand", sanitizeCLICommand);

/**
 * Return the source-defined group segment used to name a SubSDK's CLI command.
 * FieldName is the name passed to ast.NewSubSDK: one namespace segment when a
 * hierarchy is materialized, or the whole resolved tag otherwise. Unlike
 * Type.Name, it is not changed when Go type-name collisions are resolved.
 */
function getSDKGroupName(sdk: SDK): string {
  return sdk?.FieldName || "";
}
registerTemplateFunc("getSDKGroupName", getSDKGroupName);

/**
 * Check if removeStutter config is enabled.
 */
function isRemoveStutterEnabled(): boolean {
  const val = context.Global.Config.RemoveStutter;
  // Default to true if not set
  if (val === undefined || val === null) return true;
  return val === true || val === "true";
}

/**
 * Check if an operation has an explicit x-speakeasy-name-override.
 * When set, stutter removal should be skipped to respect the user's intent.
 */
function hasNameOverride(op: Operation): boolean {
  return Boolean(op.Extensions?.MethodNameOverride);
}

/**
 * Find the stutter prefix length (in kebab parts) for an operation under a group.
 * Returns the number of operation kebab parts that overlap with the group, or 0
 * if there's no stutter. A return value equal to the total number of parts means
 * the operation name matches the group exactly.
 *
 * Three detection strategies (checked in order):
 *   1. Standard prefix — opKebab starts with groupKebab + "-"
 *   2. Normalized prefix — collapse hyphens and compare
 *      (handles "linked-in" vs "linkedin" where the same word is hyphenated differently)
 *   3. Word suffix — a contiguous word-suffix of the group matches a word-prefix of the op
 *      (handles "file-uploads" group with "uploads-create" operation)
 */
function findStutterPrefixLen(groupKebab: string, opKebab: string): number {
  const opParts = opKebab.split("-");

  // 1. Standard prefix: opKebab starts with groupKebab + "-"
  if (opKebab === groupKebab) return opParts.length;
  if (opKebab.startsWith(groupKebab + "-")) {
    const groupPartCount = groupKebab.split("-").length;
    return groupPartCount;
  }

  // 2. Normalized prefix: collapse hyphens and compare character sequences.
  //    Handles cases like group "linked-in" vs operation "linkedin-get-mention"
  //    where the same word is split differently by the kebab caser.
  const groupNorm = groupKebab.replaceAll("-", "");
  const opNorm = opKebab.replaceAll("-", "");
  if (opNorm.startsWith(groupNorm)) {
    // Walk op parts to find a clean word boundary at groupNorm.length chars
    let consumed = 0;
    for (let i = 0; i < opParts.length; i++) {
      consumed += opParts[i].length;
      if (consumed === groupNorm.length) {
        return i + 1;
      }
      if (consumed > groupNorm.length) break;
    }
  }

  // 3. Word suffix: check if a contiguous suffix of the group's words matches
  //    a prefix of the operation's words. Handles cases like group "file-uploads"
  //    with operation "uploads-create" where the op uses a shortened prefix.
  const groupParts = groupKebab.split("-");
  for (let suffixStart = 1; suffixStart < groupParts.length; suffixStart++) {
    const suffix = groupParts.slice(suffixStart);
    if (suffix.length <= opParts.length) {
      const opPrefix = opParts.slice(0, suffix.length);
      if (suffix.every((s: string, i: number) => s === opPrefix[i])) {
        return suffix.length;
      }
    }
  }

  return 0;
}

/**
 * Find the stutter suffix length (in kebab parts) for an operation under a group.
 * Returns the number of operation kebab parts at the END that overlap with the
 * group name, or 0 if there's no suffix stutter.
 *
 * Handles the common verb+Noun operationId pattern where the noun matches the
 * group name, e.g., group "requests" with operation "list-requests".
 *
 * Two detection strategies (checked in order):
 *   1. Standard suffix — opKebab ends with "-" + groupKebab
 *   2. Normalized suffix — collapse hyphens and compare character sequences
 *      (handles "linked-in" vs "linkedin" where the same word is split differently)
 */
function findStutterSuffixLen(groupKebab: string, opKebab: string): number {
  const opParts = opKebab.split("-");

  // Exact match is handled by findStutterPrefixLen, skip here
  if (opKebab === groupKebab) return 0;

  // Try the group name as-is, and also its singular form (strip trailing "s")
  // to handle the common pattern where the group is plural ("meters") but
  // operation IDs use singular nouns ("create-meter", "get-meter").
  const candidates = [groupKebab];
  const groupParts = groupKebab.split("-");
  const lastPart = groupParts[groupParts.length - 1];
  if (lastPart.endsWith("s") && lastPart.length > 1) {
    const singularLast = lastPart.slice(0, -1);
    const singularGroup = [...groupParts.slice(0, -1), singularLast].join("-");
    candidates.push(singularGroup);
  }

  for (const candidate of candidates) {
    // 1. Standard suffix: opKebab ends with "-" + candidate
    if (opKebab.endsWith("-" + candidate)) {
      const candidatePartCount = candidate.split("-").length;
      return candidatePartCount;
    }

    // 2. Normalized suffix: collapse hyphens and compare character sequences.
    const candidateNorm = candidate.replaceAll("-", "");
    const opNorm = opKebab.replaceAll("-", "");
    if (opNorm.endsWith(candidateNorm)) {
      // Walk op parts from the end to find a clean word boundary
      let consumed = 0;
      for (let i = opParts.length - 1; i >= 0; i--) {
        consumed += opParts[i].length;
        if (consumed === candidateNorm.length) {
          return opParts.length - i;
        }
        if (consumed > candidateNorm.length) break;
      }
    }
  }

  return 0;
}

/**
 * Determine the stutter relationship between a group name and an operation ID.
 * Returns:
 *   "exact" — the operation's kebab name matches the group's kebab name exactly
 *   "prefix" — the operation's kebab name starts with the group's kebab name + "-"
 *   "suffix" — the operation's kebab name ends with "-" + the group's kebab name
 *   "none" — no stutter
 */
function getStutterKind(
  groupName: string,
  operationID: string,
): "exact" | "prefix" | "suffix" | "none" {
  if (!isRemoveStutterEnabled()) return "none";
  if (!groupName) return "none";

  const groupKebab = sanitizeCLICommand(groupName);
  const opKebab = sanitizeCLICommand(operationID);
  const opParts = opKebab.split("-");
  const prefixLen = findStutterPrefixLen(groupKebab, opKebab);

  if (prefixLen > 0) {
    if (prefixLen >= opParts.length) return "exact";
    return "prefix";
  }

  const suffixLen = findStutterSuffixLen(groupKebab, opKebab);
  if (suffixLen > 0 && suffixLen < opParts.length) {
    return "suffix";
  }

  return "none";
}

/**
 * Get the de-stuttered command name for an operation.
 * For prefix matches, strips the group prefix from the operation name.
 * For suffix matches, strips the group suffix from the operation name.
 * For exact matches, returns the original name (caller handles promotion).
 * For no stutter, returns the original name.
 */
function getDeStutteredCommandName(
  groupName: string,
  operationID: string,
): string {
  if (!isRemoveStutterEnabled() || !groupName) {
    return sanitizeCLICommand(operationID);
  }

  const groupKebab = sanitizeCLICommand(groupName);
  const opKebab = sanitizeCLICommand(operationID);
  const opParts = opKebab.split("-");
  const prefixLen = findStutterPrefixLen(groupKebab, opKebab);

  if (prefixLen > 0 && prefixLen < opParts.length) {
    return opParts.slice(prefixLen).join("-");
  }

  const suffixLen = findStutterSuffixLen(groupKebab, opKebab);
  if (suffixLen > 0 && suffixLen < opParts.length) {
    return opParts.slice(0, opParts.length - suffixLen).join("-");
  }

  return opKebab;
}
registerTemplateFunc("getDeStutteredCommandName", getDeStutteredCommandName);
registerTemplateFunc("getStutterKind", getStutterKind);

/**
 * Sanitize command name to PascalCase with Cmd suffix.
 * e.g., "createUser" -> "CreateUserCmd"
 */
function sanitizeCommandName(name: string): string {
  return caser().ToPascal(sanitizeName(name) + "Cmd");
}
registerTemplateFunc("sanitizeCommandName", sanitizeCommandName);

/**
 * Generate init function name for a command.
 * e.g., "CreateUserCmd" -> "initCreateUserCmd"
 */
function sanitizeInitFuncName(name: string, privateFunc = true): string {
  const pascalName = caser().ToPascal(sanitizeName(name));
  return (privateFunc ? "init" : "Init") + pascalName;
}
registerTemplateFunc("sanitizeInitFuncName", sanitizeInitFuncName);

/**
 * Generate run function name for a command.
 * e.g., "CreateUserCmd" -> "runCreateUserCmd"
 */
function sanitizeRunFuncName(cmdName: string): string {
  return "run" + cmdName;
}
registerTemplateFunc("sanitizeRunFuncName", sanitizeRunFuncName);

function sanitizeFlagName(name: string): string {
  // CLI flags are always lowercase. The caser preserves acronyms like "ID" →
  // "account-ID" which is wrong for flags; force lowercase to get "account-id".
  return caser()
    .ToKebab(sanitizeName(normalizePluralAcronyms(name)))
    .toLowerCase();
}
registerTemplateFunc("sanitizeFlagName", sanitizeFlagName);

/**
 * Escape a string for use in Go double-quoted strings without changing its
 * runtime value. Hex-escaping template delimiters keeps recurse from parsing
 * literal {{/}}; hex-escaping control characters keeps them out of emitted Go
 * source, where a raw NUL is a compile error and others are invisible.
 */
registerTemplateFunc("escapeGoString", (s: string) => escapeGoString(s));

function escapeGoString(s: string): string {
  return s
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\r")
    .replace(/\t/g, "\\t")
    .replace(/\{\{/g, "\\x7b\\x7b")
    .replace(/\}\}/g, "\\x7d\\x7d")
    .replace(
      /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/g,
      (c) => `\\x${c.charCodeAt(0).toString(16).padStart(2, "0")}`,
    );
}

/**
 * Render a complete Go string literal without changing its runtime value.
 */
function goStringLiteral(s: string): string {
  return `"${escapeGoString(s)}"`;
}
registerTemplateFunc("goStringLiteral", (s: string) => goStringLiteral(s));

/**
 * Generate an initial-based abbreviation for a kebab-case command name.
 * E.g., "bank-accounts" → "ba", "sub-group" → "sg".
 * Returns "" for single-word names (nothing to abbreviate).
 */
function generateAbbreviation(kebabName: string): string {
  const parts = kebabName.split("-");
  if (parts.length <= 1) return "";
  return parts.map((p) => p[0] || "").join("");
}

/**
 * Compute aliases for a set of sibling command names.
 * Returns a Map<commandName, aliases[]> where each command may get
 * an initial-based abbreviation (e.g., "bank-accounts" → ["ba"]).
 *
 * Conflict resolution: commands are sorted alphabetically for deterministic
 * first-wins ordering. If two commands share the same abbreviation, the first
 * gets it and the second tries progressively longer abbreviations
 * (prefix + more chars of last part). If all expansions conflict or equal
 * the original name, no alias is assigned.
 */
function computeCommandAliases(commandNames: string[]): Map<string, string[]> {
  const result = new Map<string, string[]>();
  const sorted = [...commandNames].sort();

  // Track claimed abbreviations
  const claimed = new Set<string>();
  // Also reserve the original command names to avoid aliases that match a sibling
  for (const name of sorted) {
    claimed.add(name);
  }

  for (const name of sorted) {
    const parts = name.split("-");
    if (parts.length <= 1) continue; // Single-word: no alias

    const baseAbbrev = generateAbbreviation(name);
    if (!baseAbbrev) continue;

    if (!claimed.has(baseAbbrev)) {
      claimed.add(baseAbbrev);
      result.set(name, [baseAbbrev]);
      continue;
    }

    // Conflict: try expanding the last part progressively
    const prefix = parts
      .slice(0, -1)
      .map((p) => p[0] || "")
      .join("");
    const lastPart = parts[parts.length - 1];
    let found = false;
    for (let len = 2; len <= lastPart.length; len++) {
      const candidate = prefix + lastPart.slice(0, len);
      if (candidate === name) break; // No point if alias equals original
      if (!claimed.has(candidate)) {
        claimed.add(candidate);
        result.set(name, [candidate]);
        found = true;
        break;
      }
    }
    // If no expansion worked, skip alias for this command
    if (!found) continue;
  }

  return result;
}

// @ts-ignore
function getGolangPackage(): string {
  let versionParts = context.Global.Config.SDKVersion.split(".");

  let majorVersion = parseInt(versionParts[0], 10);

  if (majorVersion > 1) {
    return `${context.Global.Config.PackageName}/v${majorVersion}`;
  } else {
    return `${context.Global.Config.PackageName}`;
  }
}

registerTemplateFunc("getGolangPackage", getGolangPackage);
