function sanitizeMarkdownFileName(name: string): string {
  name = sanitizeFile(name, "");
  return name.toLowerCase();
}

registerTemplateFunc("sanitizeMarkdownFileName", sanitizeMarkdownFileName);

function sanitizeMarkdownSlug(name: string): string {
  name = name.replaceAll(" ", "-");
  return name.toLowerCase();
}

registerTemplateFunc("sanitizeMarkdownSlug", sanitizeMarkdownSlug);

function sanitizeMarkdownTableContent(content: string): string {
  return content.replaceAll("|", "\\|").replaceAll(/\r\n|\n/g, "<br/>");
}

registerTemplateFunc(
  "sanitizeMarkdownTableContent",
  sanitizeMarkdownTableContent,
);

function parseOrgAndRepo(url: string): string {
  const githubPrefix = "github.com/";
  const index = url.indexOf(githubPrefix);

  if (index !== -1) {
    // Extract substring from the end of "github.com/"
    return url.substring(index + githubPrefix.length);
  }

  return "";
}

registerTemplateFunc("parseOrgAndRepo", parseOrgAndRepo);

// @ts-ignore
function sanitizeField(field: FieldDef, _parent?: TypeDef): string {
  return sanitizeFieldName(field.Name);
}

// @ts-ignore
function sanitizeClassName(name: string): string {
  return sanitizeName(name);
}

// @ts-ignore
function sanitizeFileName(name: string): string {
  return sanitizeFile(name, "");
}

// Sanitizes each path segment of an output location to follow the target language's
// package/folder naming conventions. Must be overridden in each language template.
// @ts-ignore
function sanitizeOutputLocation(_outputLocation: string): string {
  throw new Error("sanitizeOutputLocation not implemented by target");
}

// @ts-ignore
function sanitizeClass(
  _typeDef: TypeDef,
  _usageLocation: string,
  _definition: boolean,
): string {
  throw new Error("sanitizeClass not implemented by target");
}

// @ts-ignore
function sanitizeType(
  _typeDef: TypeDef,
  _optional?: boolean,
  _usageLocation?: string,
  _addImports?: boolean,
): string {
  throw new Error("sanitizeType not implemented by target");
}

// Deduplicates OAuth2 scope enum names by extracting scope names and passing
// them through the language-specific getEnumNamesFromValues function.
// Each language defines its own getEnumNamesFromValues (loaded after common),
// which is resolved at call time.
// @ts-ignore
function getDeduplicatedScopeEnumNames(scopes: OAuth2Scope[]): string[] {
  const names = scopes.map((s) => s.Name);
  // @ts-ignore - getEnumNamesFromValues is defined in each language's sanitization.ts
  return getEnumNamesFromValues(names);
}
registerTemplateFunc(
  "getDeduplicatedScopeEnumNames",
  getDeduplicatedScopeEnumNames,
);

// @ts-ignore
function sanitizeStatusCodes(statusCodes: string[]): string[] {
  if (statusCodes.includes("default")) {
    return ["default"];
  }

  const codeSet = new Set(statusCodes);
  const has4XX = codeSet.has("4XX");
  const has5XX = codeSet.has("5XX");

  return [...codeSet]
    .filter(
      (code) =>
        !(has4XX && code.match(/^4\d\d$/)) &&
        !(has5XX && code.match(/^5\d\d$/)),
    )
    .sort();
}
