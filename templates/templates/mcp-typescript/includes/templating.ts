type TypescriptModel = {
  DocGroup: string;
  Name: string;
  Types: TypeDef[];
  Servers?: any;
  OutputLocation: string;
};

function finalizeComments(lines: string[], indent: number): string {
  if (lines.length == 0) {
    return "";
  }

  lines = lines.map((line) => {
    return ` * ${line}`;
  });

  lines.unshift("/**");
  lines.push(" */");

  return "\n" + indentLines(lines, indent);
}

//@ts-ignore
function getModelJobs(types: BucketedTypes, path: string): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];

  let modelsToExport: Record<
    string,
    {
      templatedModels: string[];
      templatedTypes: string[];
    }
  > = {};

  // Pre-add default API error
  modelsToExport[
    getModelsLocation(context.Global.Config.Imports.GetErrorsPath())
  ] = {
    templatedModels: [
      getDefaultErrorFileName(),
      "sdkvalidationerror",
      "httpclienterrors",
    ],
    templatedTypes: [
      getDefaultErrorClassName(),
      "SDKValidationError",
      "HTTPClientError",
    ],
  };

  for (const [outputLocation, models] of sequencedMapEntries(types)) {
    for (const [model, types] of sequencedMapEntries(models)) {
      let modelPath = getModelsLocation(outputLocation);

      if (!(modelPath in modelsToExport)) {
        modelsToExport[modelPath] = {
          templatedModels: [],
          templatedTypes: [],
        };
      }

      let docGroup = path;
      if (modelPath) {
        docGroup += "/" + modelPath;
      }

      let ctx: TypescriptModel = {
        DocGroup: docGroup,
        Name: model,
        Types: types,
        OutputLocation: outputLocation,
      };

      if (
        types.find((t) => t.Scope.toString() == "operations") ||
        types.length == 0
      ) {
        ctx.Servers = context.Global.AST.OperationServers.Get(model)[0];
      }
      if (types.length === 0 && !ctx.Servers) {
        continue;
      }

      let modelName = sanitizeFileName(model);

      modelsToExport[modelPath].templatedModels.push(modelName);

      for (const type of types) {
        let typeName = sanitizeClassName(type.Name);
        modelsToExport[modelPath].templatedTypes.push(`"${typeName}"`);
      }

      jobs.push(
        createTemplateFileJob(
          `modelfile.ts.stmpl`,
          `${path}/${modelPath}/${sanitizeFileName(model)}.ts`,
          ctx,
        ),
      );
    }
  }

  return jobs;
}

function getSDKFunctionsJobs(root: SDK): Job[] {
  const jobs: Job[] = [];

  const queue: Array<SDK> = [root];

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    for (const op of sdk.Operations) {
      const funcFilename = sanitizeFuncFilename(op);
      const funcPath = `src/funcs/${funcFilename}.ts`;

      jobs.push(
        createTemplateFileJob("func.ts.stmpl", funcPath, {
          DocGroup: funcPath,
          Operation: op,
        }),
      );
    }
  }

  return jobs;
}

/**
 * Returns the value for TypeScript.
 */
// @ts-ignore
function templateArrayValue(value: any): string {
  return `${value}`;
}

/**
 * Returns the true or false keyword for TypeScript.
 */
// @ts-ignore
function templateBoolValue(
  value: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return value ? "true" : "false";
}

/**
 * Returns the bracket character for TypeScript, depending on the given TypeDef.
 *
 *   - If TypeDef is an array, returns left square bracket ("[") character for
 *     opening and right square bracket ("]") for closing.
 *   - Otherwise, returns left curly brace ("{") character for opening and right
 *     curly brace ("}") character for closing.
 */
// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  opening: boolean,
  additionalContext?: TemplateValueContext,
): string {
  if (typeDef.Type.toString() === "array") {
    return opening ? "[" : "]";
  }

  return opening ? "{" : "}";
}

// @ts-ignore
function templateBuiltinComment(
  docgroup: string | null,
  title: string,
  description: string,
  indent: number,
): string {
  const resolved = resolveComment(
    "builtin",
    docgroup,
    "",
    title,
    "",
    description,
  );
  return finalizeComments(templateCmsComment(resolved), indent);
}

registerTemplateFunc("templateBuiltinComment", templateBuiltinComment);

/**
 * Returns `new TextEncoder().encode()` for TypeScript.
 */
// @ts-ignore
function templateByteValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `new TextEncoder().encode("${value.replace(/"/g, '\\"')}")`;
}

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  docgroup: string | null,
  title: string | null,
  indent: number,
  type: "method" | "field" | "class" | "enum" | "const" | "var" = null,
  omitExternalDocs: boolean = false,
  additionalNotes: string = "",
): string {
  const resolved = resolveCommentDef(
    "openapi",
    docgroup,
    type ?? "",
    title,
    comments,
  );
  const cmsComment = templateCmsComment(resolved);

  if (!comments && !additionalNotes) {
    return finalizeComments(cmsComment, indent);
  }

  let lines = cmsComment;

  if (comments?.ExternalDocs && !omitExternalDocs) {
    let externalDocsLines = [];

    if (comments.ExternalDocs.Description) {
      externalDocsLines = sanitizeComments(
        comments.ExternalDocs.Description,
      ).split("\n");
    }

    lines.push("");
    lines.push(
      `@see {@link ${comments.ExternalDocs.URL}}${
        externalDocsLines.length > 0 ? " - " + externalDocsLines.shift() : ""
      }`,
    );
    lines.push(...externalDocsLines);
  }

  if (additionalNotes) {
    if (lines.length > 0) {
      lines.push("");
    }

    lines.push(...additionalNotes.split("\n"));
  }

  if (comments?.Deprecated) {
    if (lines.length > 0) {
      lines.push("");
    }

    let deprecated = `@deprecated ${type}: ${comments.DeprecationMessage}.`;

    if (comments.DeprecationReplacement) {
      const replacement = sanitizeDeprecationReplacement(
        comments.DeprecationReplacement,
        type as "method" | "field" | "class",
      );
      if (replacement) {
        deprecated += ` Use ${replacement} instead.`;
      }
    }

    lines.push(deprecated);
  }

  if (lines.length == 0) {
    return "";
  }

  return finalizeComments(lines, indent);
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map(
    (f) => `{@link Security.${sanitizeFieldName(f.Name)}}`,
  );

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set on the \`security\` parameter when initializing the SDK.`
      : `If set, this operation will use ${items} from the global security.`;

  if (fieldList.length === 1) {
    return remark(fieldList[0]);
  }

  const last = fieldList.pop();
  if (fieldList.length === 1) {
    return remark(`either ${fieldList[0]} or ${last}`);
  }

  return remark(`one of ${fieldList.join(", ")}, or ${last}`);
}

// @ts-ignore
function templateCmsComment(comments: SimpleCommentDef): string[] {
  let lines = [];

  let firstLineParts = [];

  let descriptionLines = [];
  if (comments.Description) {
    descriptionLines = comments.Description.split("\n");
  }

  if (comments.Summary) {
    firstLineParts.push(comments.Summary);
  } else if (descriptionLines.length > 0) {
    firstLineParts.push(descriptionLines.shift());
  }

  const firstLine = firstLineParts.join(" - ");

  if (firstLine) {
    lines.push(...sanitizeComments(firstLine).split("\n"));
  }

  if (descriptionLines.length > 0) {
    lines.push("");
    lines.push("@remarks");
    lines.push(...descriptionLines.map(sanitizeComments));
  }

  return lines;
}

// @ts-ignore
function templateConstName(name: string, prefix: string) {
  name = sanitizeName(name);
  return caser().ToPascal(`${prefix}_${name}`);
}

registerTemplateFunc("templateConstName", templateConstName);

/**
 * Returns same as templateStringValue.
 */
// @ts-ignore
function templateDateValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return templateStringValue(`${value}`, fieldDef, additionalContext);
}

/**
 * Returns same as templateStringValue.
 */
// @ts-ignore
function templateDateTimeValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return templateStringValue(`${value}`, fieldDef, additionalContext);
}

// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  const en = typeDef.Enum;
  if (en == null) {
    return "";
  }

  const rawValues = en.Values;
  if (rawValues == null) {
    return "";
  }

  const sep = ": ";

  const enumNames = getEnumNames(typeDef);
  return rawValues
    .map((val, index) => {
      const value = sanitizeEnumValue(val, en.Type.Type);
      return `${enumNames[index]}${sep}${value},`;
    })
    .join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);

// @ts-ignore
function templateErrorStatusCodes(response: ResponseDef): string {
  return `[${response
    .GetErrorStatusCodes()
    .map((s) => `"${s}"`)
    .join(",")}]`;
}

registerTemplateFunc("templateErrorStatusCodes", templateErrorStatusCodes);

/**
 * Returns the sanitized field name, a colon (":") character, and a space (" ")
 * character for TypeScript.
 */
// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
) {
  return `${sanitizeFieldName(originalFieldName(fieldDef))}: `;
}

/**
 * Returns comma (",") character for TypeScript.
 */
// @ts-ignore
function templateFieldDelimiter() {
  return ",";
}

// @ts-ignore
function templateFieldRemaps(fields: FieldDef[]): {
  additionalPropsField: FieldDef;
} {
  let additionalPropsField = undefined;

  for (const field of fields) {
    if (field.IsAdditionalProperties) {
      additionalPropsField = field;
      break;
    }
  }

  return { additionalPropsField };
}

registerTemplateFunc("templateFieldRemaps", templateFieldRemaps);

/**
 * Returns the value for TypeScript, unless decimal, then `new Decimal()`.
 */
// @ts-ignore
function templateFloatValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  switch (fieldDef.Type.Type.toString()) {
    case "decimal":
      addUsageTypeImport("decimal", "Decimal");
      return `new Decimal("${value}")`;
    default:
      return `${value}`;
  }
}

/**
 * Returns the sanitized global field name for TypeScript.
 */
// @ts-ignore
function templateGlobalFieldName(name: string): string {
  let reserved = ["security", "defaultClient", "serverURL"];

  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reserved.push("server");
  } else {
    reserved.push("serverIdx");
  }

  let fieldName = sanitizeFieldName(name);

  if (reserved.includes(name)) {
    fieldName += "Global";
  }

  return fieldName;
}

registerTemplateFunc("templateGlobalFieldName", templateGlobalFieldName);

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    addTestImport("./files.js", "filesToString");
    filePath = `.speakeasy/testfiles/${filePath}`;
    return `await filesToString("${filePath}")`;
  }

  return `"" // Populate with string from file, for example ${filePath}`;
}

/**
 * Returns given indent count times two spaces ("  ") characters for TypeScript.
 */
// @ts-ignore
function templateIndent(indent: number): string {
  return "  ".repeat(indent);
}

/**
 * Returns the value for TypeScript, unless bigint, then wrapped with BigInt().
 */
// @ts-ignore
function templateIntValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  switch (fieldDef.Type.Type.toString()) {
    case "bigint":
      return `BigInt("${value}")`;
    default:
      return `${value}`;
  }
}

/**
 * Returns the quoted key, a colon (":") and space (" ") character, and the
 * value for TypeScript.
 */
// @ts-ignore
function templateMapValue(key: any, value: any): string {
  return `"${key}": ${value}`;
}

function templateMethodOptions(op: Operation) {
  let out: string[] = [];

  const dedupedAcceptTypes = deduplicateAcceptTypes(op.GetAcceptTypes());
  if (dedupedAcceptTypes.length > 1) {
    out.push(`acceptHeaderOverride?: ${sanitizeAcceptEnumName(op)};`);
  }

  return out.length
    ? `RequestOptions & { ${out.join(" ")} }`
    : "RequestOptions";
}

registerTemplateFunc("templateMethodOptions", templateMethodOptions);

/**
 * Returns the null keyword for TypeScript.
 */
// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "null";
}

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "null";
  }

  return JSON.stringify(scopes);
}

registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

/**
 * Returns an empty string for TypeScript.
 */
// @ts-ignore
function templateOptionalSymbol() {
  return "";
}

// @ts-ignore
function templateServerID(op: Operation, id: string, relativeTo = ""): string {
  const model = getOperationModelName(op);
  const outputLocation = getModelsLocation(getOperationsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);
  const idconst = templateConstName(id, `${sanitizeFieldName(model)}Server`);

  addImport(`${prefix}/${sanitizeFileName(model)}.js`, idconst, typeImport);

  return idconst;
}

registerTemplateFunc("templateServerID", templateServerID);

// @ts-ignore
function templateServerListName(op: Operation, relativeTo = ""): string {
  const model = getOperationModelName(op);
  const outputLocation = getModelsLocation(getOperationsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);
  const listName = `${templateConstName(model, "")}ServerList`;

  addImport(`${prefix}/${sanitizeFileName(model)}.js`, listName, typeImport);

  return listName;
}

registerTemplateFunc("templateServerListName", templateServerListName);

/**
 * Returns the quoted and sanitized string value for TypeScript.
 */
// @ts-ignore
function templateStringValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (value.includes(".env[")) {
    return value;
  }

  return quote(value).replaceAll("{{", `{{"{{"}}`);
}

/**
 * Returns an empty string for TypeScript.
 */
// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

function mcpFlags() {
  const flags = generateMCPFlagsUsage();
  return flags.flatMap((flag) => [flag.key, flag.value]);
}

function mcpFlagsString() {
  return mcpFlags().join(" ");
}

type ConfigLocation = "readme" | "landingpage";

function mcpStartCommand(
  location: ConfigLocation = "readme",
  executor: Executor = "manual",
) {
  if (location === "readme" && executor === "manual") {
    return ["node", "./bin/mcp-server.js", "start", ...mcpFlags()];
  }
  return ["npx", context.Global.Config.PackageName, "start", ...mcpFlags()];
}

function templateMcpStartCommandString(
  location: ConfigLocation = "readme",
  executor: string = "npx",
) {
  const executorEnum = executor === "manual" ? "manual" : "npx";
  return mcpStartCommand(location, executorEnum).join(" ");
}
registerTemplateFunc(
  "templateMcpStartCommandString",
  templateMcpStartCommandString,
);

function mcpName() {
  return sanitizeClassName(context.Global.AST.MainSDK.Type.Name);
}

function addHeadersToConfigDefault(config: any, headerEntries: any[]) {
  if (headerEntries.length === 0) return config;

  // Initialize headers object
  config.mcpServers[mcpName()].headers = {};

  // Process headers
  const byHeaderName = new Map<string, RemoteServerHeaders>();
  for (const option of headerEntries) {
    for (const header of option) {
      byHeaderName.set(header.headerName, header);
    }
  }

  for (const [headerName, header] of byHeaderName.entries()) {
    const envVar = `${sanitizeEnvVarName(
      context.Global.Config.PackageName,
    )}_${caser().ToSNAKE(header.fieldName)}`;
    config.mcpServers[mcpName()].headers[headerName] = `\${${envVar}}`;
  }
  return config;
}

function addHeadersToConfigVSCode(config: any, headerEntries: any[]) {
  if (headerEntries.length === 0) return config;

  // Initialize headers object
  config.servers[mcpName()].headers = {};

  // Process headers
  const byHeaderName = new Map<string, RemoteServerHeaders>();
  for (const option of headerEntries) {
    for (const header of option) {
      byHeaderName.set(header.headerName, header);
    }
  }

  for (const [headerName, header] of byHeaderName.entries()) {
    const envVar = `${sanitizeEnvVarName(
      context.Global.Config.PackageName,
    )}_${caser().ToSNAKE(header.fieldName)}`;
    config.servers[mcpName()].headers[headerName] = `\${env:${envVar}}`;
  }
  return config;
}

function mcpRemoteConfigCommand(
  endpointType: string = "sse",
  clientType: string = "default",
  url?: string,
) {
  const remoteURL =
    url ||
    context.Global.Config.CloudflareURL ||
    "https://example-cloudflare-worker.com";

  const headers: string[] = [];
  const flags = gatherMCPFlags();
  for (const flag of flags) {
    let envVar = `${caser().ToSNAKE(flag.Name)}`;
    headers.push("--header", `${flag.Name}:\${${envVar}}`);
  }

  return [
    "npx",
    "-y",
    "mcp-remote@0.1.25",
    `${remoteURL}/${endpointType}`,
    ...headers,
  ];
}

// Define executor type
type Executor = "manual" | "npx";

// ============================================================================
// IDE Configuration (for Cursor, VS Code, Windsurf config files)
// ============================================================================

type IDEConfigFormat = "json" | "js" | "toml";

interface MCPIDEConfigOptions {
  location?: ConfigLocation;
  format?: IDEConfigFormat;
  endpoint?: string;
  executor?: Executor;
}

// Internal helper for generating config object (used by installation URL functions)
function mcpIDEConfigObject(
  options: {
    location?: ConfigLocation;
    endpoint?: string;
    executor?: Executor;
  } = {},
) {
  const { location = "readme", endpoint = "sse", executor = "npx" } = options;

  if (context.Global.Config.CloudflareEnabled && location === "landingpage") {
    const command = mcpRemoteConfigCommand(endpoint);
    return { command: command[0], args: command.slice(1) };
  }

  const command = mcpStartCommand(location, executor);
  return { command: command[0], args: command.slice(1) };
}

function templateMcpIDEConfig(options: MCPIDEConfigOptions = {}) {
  const {
    location = "readme",
    format = "json",
    endpoint = "sse",
    executor = "npx",
  } = options;

  // JS format: JavaScript literal with ${o} placeholder for landing page
  if (format === "js") {
    const command = mcpRemoteConfigCommand(endpoint);
    const args = command.slice(1); // Exclude command

    // Replace the URL (index 2) with the template placeholder
    if (location === "landingpage") {
      args[2] = `\`\${o}/${endpoint}\``;
    }

    const argsJson = args
      .map((arg, i) => (i === 2 ? `${arg}` : `"${arg}"`))
      .join(",\n");

    // Extract env vars from --header args (format: "HeaderName:${ENV_VAR}")
    const envVars: string[] = [];
    for (let i = 0; i < args.length; i++) {
      if (args[i] === "--header" && args[i + 1]) {
        const match = args[i + 1].match(/\$\{([^}]+)\}/);
        if (match) {
          envVars.push(match[1]);
        }
      }
    }
    const envJson = envVars
      .map((envVar) => `"${envVar}": "YOUR_VALUE_HERE"`)
      .join(",");
    return `{
  "command": "${command[0]}",
  "args": [
    ${argsJson}
  ],
  "env": {${envJson}}
}`;
  }

  // TOML format: for Codex CLI configuration
  if (format === "toml") {
    const serverName = mcpName();
    const httpHeaders = mcpHttpHeadersTOML();
    const remoteURL =
      location === "landingpage"
        ? `\${o}/${endpoint}`
        : context.Global.Config.CloudflareEnabled
        ? `${
            context.Global.Config.CloudflareURL ||
            "https://example-cloudflare-worker.com"
          }/${endpoint}`
        : `https://your-server-url/${endpoint}`;

    return `[mcp_servers.${serverName}]
url = "${remoteURL}"
${httpHeaders}`;
  }

  // JSON format: stringified JSON for config files
  return JSON.stringify(
    mcpIDEConfigObject({ location, endpoint, executor }),
    null,
    2,
  );
}
registerTemplateFunc("templateMcpIDEConfig", templateMcpIDEConfig);

function templateMcpDynamicModeIDEConfig(
  options: { scope?: string } = {},
): string {
  const name = mcpName();
  const packageName = context.Global.Config.PackageName;

  const args: string[] = [packageName, "start", "--mode", "dynamic"];
  if (options.scope) {
    args.push("--scope", options.scope);
  }

  const lines = [
    `{`,
    `  "mcpServers": {`,
    `    "${name}": {`,
    `      "command": "npx",`,
    `      "args": [${args.map((a) => `"${a}"`).join(", ")}],`,
    `      // ... other server arguments`,
    `    }`,
    `  }`,
    `}`,
  ];

  return lines.join("\n");
}
registerTemplateFunc(
  "templateMcpDynamicModeIDEConfig",
  templateMcpDynamicModeIDEConfig,
);

function mcpHttpHeadersTOML(
  options: { escapeForTemplateLiteral?: boolean } = {},
) {
  const flags = gatherMCPFlags();
  if (flags.length === 0) {
    return "http_headers = { }";
  }

  const headers: string[] = [];
  for (const flag of flags) {
    const envVar = `${caser().ToSNAKE(flag.Name)}`;
    const placeholder = `YOUR_${envVar}`;
    headers.push(`"${flag.Name}" = "${placeholder}"`);
  }

  return `http_headers = { ${headers.join(", ")} }`;
}

function mcpRemoteConfigJS() {
  const globalSDK = context.Global.AST.MainSDK;
  const headerEntries = getRemoteServerHeaders(globalSDK.Security?.Type);

  let js = `{
    "mcpServers": {
      "${mcpName()}": {
        "type": "sse",
        "url": \`\${o}/sse\``;

  if (headerEntries.length > 0) {
    js += `,
        "headers": {`;

    const byHeaderName = new Map<string, RemoteServerHeaders>();
    for (const option of headerEntries) {
      for (const header of option) {
        byHeaderName.set(header.headerName, header);
      }
    }

    const headers: string[] = [];
    for (const [headerName, header] of byHeaderName.entries()) {
      const envVar = `${sanitizeEnvVarName(
        context.Global.Config.PackageName,
      )}_${caser().ToSNAKE(header.fieldName)}`;
      headers.push(`
          "${headerName}": "$" + "{" + "${envVar}" + "}"`);
    }

    js += headers.join(",");
    js += `
        }`;
  }

  js += `
      }
    }
  }`;

  return js;
}

function templateMcpCursorInstallationURL() {
  const configString = base64Encode(
    JSON.stringify(
      mcpIDEConfigObject({ location: "landingpage", executor: "npx" }),
    ),
  );
  const name = mcpName();
  return `cursor://anysphere.cursor-deeplink/mcp/install?name=${name}&config=${configString}`;
}
registerTemplateFunc(
  "templateMcpCursorInstallationURL",
  templateMcpCursorInstallationURL,
);

function templateMcpCursorInstallationButton() {
  const url = templateMcpCursorInstallationURL();
  return `[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](${url})`;
}
registerTemplateFunc(
  "templateMcpCursorInstallationButton",
  templateMcpCursorInstallationButton,
);

function templateMcpVSCodeInstallationURL() {
  const configString = base64Encode(
    JSON.stringify(
      mcpIDEConfigObject({ location: "landingpage", executor: "npx" }),
    ),
  );
  const name = mcpName();
  return `vscode://ms-vscode.vscode-mcp/install?name=${name}&config=${configString}`;
}
registerTemplateFunc(
  "templateMcpVSCodeInstallationURL",
  templateMcpVSCodeInstallationURL,
);

function templateMcpVSCodeInstallationButton() {
  const url = templateMcpVSCodeInstallationURL();
  const name = mcpName();
  return `[![Install in VS Code](https://img.shields.io/badge/VS_Code-VS_Code?style=flat-square&label=Install%20${encodeURIComponent(
    name,
  )}%20MCP&color=0098FF)](${url})`;
}
registerTemplateFunc(
  "templateMcpVSCodeInstallationButton",
  templateMcpVSCodeInstallationButton,
);

// ============================================================================
// CLI Commands (for Claude Code CLI, Gemini CLI)
// ============================================================================

type CLIClient = "claude" | "gemini";
type CLICommandFormat = "string" | "array" | "html";

interface MCPCLICommandOptions {
  client: CLIClient;
  location?: ConfigLocation;
  format?: CLICommandFormat;
  cloudflareURL?: string;
}

function templateMcpCLICommand(options: MCPCLICommandOptions) {
  const {
    client,
    location = "readme",
    format = "string",
    cloudflareURL = context.Global.Config.CloudflareURL,
  } = options;

  // HTML format: uses ${o} placeholder for dynamic URL (for landing page)
  if (format === "html") {
    const headers = getRemoteServerHeaders(
      context.Global.AST.MainSDK.Security?.Type,
    );
    const headerArgs = headers.flatMap((header) =>
      header.map((h) => `--header "${h.headerName}: ..."`),
    );
    return [
      client,
      "mcp",
      "add",
      "--transport",
      "sse",
      mcpName(),
      `\${o}/sse`,
      ...headerArgs,
    ].join(" ");
  }

  let commandArray: string[];

  // Remote configuration when Cloudflare is enabled and for landing page
  if (context.Global.Config.CloudflareEnabled && location === "landingpage") {
    commandArray = [
      client,
      "mcp",
      "add",
      "--transport",
      "sse",
      mcpName(),
      `${cloudflareURL}/sse`,
    ];
  } else {
    // Local configuration
    commandArray = [
      client,
      "mcp",
      "add",
      mcpName(),
      "--",
      "npx",
      "-y",
      context.Global.Config.PackageName,
      "start",
      ...mcpFlags(),
    ];
  }

  return format === "array" ? commandArray : commandArray.join(" ");
}
registerTemplateFunc("templateMcpCLICommand", templateMcpCLICommand);

function getMCPBUserConfig() {
  const flags = gatherMCPFlags();
  if (flags.length === 0) {
    return null;
  }

  const userConfig = {};
  for (const flag of flags) {
    userConfig[caser().ToSnake(flag.Name)] = {
      type: flag.Type,
      title: flag.Title,
      description:
        flag.Description || `The ${flag.Title} to use for the request`,
      ...(flag.Optional ? { required: false } : { required: true }),
      ...(flag.Default ? { default: flag.Default } : {}),
      ...(flag.Sensitive ? { sensitive: flag.Sensitive } : {}),
      ...(flag.Min ? { min: flag.Min } : {}),
      ...(flag.Max ? { max: flag.Max } : {}),
    };
  }
  return userConfig;
}

function getMCPBTools(): MCPBTool[] | null {
  const root = context.Global.AST.MainSDK;
  const tools: MCPBTool[] = [];

  // Collect all tool operations using the mcpOperations generator
  for (const { op, mcpType } of mcpOperations(root)) {
    if (mcpType === "tool") {
      const name = sanitizeMCPPrimitiveName(op);
      const description = sanitizeMCPPrimitiveDescription(op);

      tools.push({
        name,
        description,
      });
    }
  }

  // If no tools found, return null
  if (tools.length === 0) {
    return null;
  }

  return tools;
}

// New function that returns the MCP config args as an array
function getMCPBConfigArgs(): string[] {
  const args = ["start"];
  const flags = gatherMCPFlags();

  function templateFlag(flag: FlagDescriptor) {
    // https://github.com/anthropics/dxt/blob/main/MANIFEST.md#variable-substitution-in-user-configuration
    // eg: bearer-auth -> "${user_config.bearer_auth}"
    return `\${user_config.${caser().ToSnake(flag.Name)}}`;
  }

  for (const flag of flags) {
    args.push(`--${flag.Name}`, templateFlag(flag));
  }

  return args;
}

function templateMCPBManifest() {
  const config = context.Global.Config;
  const sdk = context.Local.SDK;

  // Build the base manifest object
  let manifest: MCPBManifestSchema = {
    manifest_version: "0.3",
    name: config.PackageName,
    version: config.SDKVersion,
    description: sdk.Comments?.Summary || "",
    long_description: sdk.Comments?.Description || "",
    author: {
      name: config.Author,
    },
    server: {
      type: "node",
      entry_point: "./bin/mcp-server.js",
      mcp_config: {
        command: "node",
        args: ["${__dirname}/bin/mcp-server.js", ...getMCPBConfigArgs()],
      },
    },
  };

  // Add display_name (human-friendly name for UI display)
  if (sdk.Comments?.Summary) {
    manifest.display_name = sdk.Comments.Summary;
  } else if (config.PackageName) {
    // Fallback to a formatted package name
    manifest.display_name = config.PackageName.split(/[-_]/)
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  }

  // Add screenshots array (look for screenshot files)
  const screenshotFiles = listFiles(".").filter((file) =>
    ["png", "jpg", "jpeg", "gif", "webp"].some((ext) =>
      file.toLowerCase().endsWith(ext),
    ),
  );
  if (screenshotFiles.length > 0) {
    manifest.screenshots = screenshotFiles;
  }

  // Add prompts array (currently empty as MCP servers typically don't provide prompts)
  // This could be extended in the future if prompt functionality is added
  manifest.prompts = [];

  // Add keywords (derive from package name and description)
  const keywords = new Set<string>();

  // Add words from package name
  if (config.PackageName) {
    config.PackageName.split(/[-_]/).forEach((word) => {
      if (word.length > 2) {
        keywords.add(word.toLowerCase());
      }
    });
  }

  // Add words from description
  if (sdk.Comments?.Description) {
    const descWords =
      sdk.Comments.Description.toLowerCase().match(/\b[a-z]{3,}\b/g);
    if (descWords) {
      descWords.slice(0, 5).forEach((word) => keywords.add(word));
    }
  }

  manifest.keywords = Array.from(keywords);

  // Add user_config if it exists
  const userConfig = getMCPBUserConfig();
  if (userConfig) {
    manifest.user_config = userConfig;
  }

  // Add icon if a .png exists in the root dir
  const icon = listFiles(".").find((file) => file.endsWith(".png"));
  if (icon) {
    manifest.icon = icon;
  }

  // Apply overlay if defined
  const overlay = context.Global.Config.McpbManifestOverlay;
  // Recursively sort the overlay object by keys (deterministically, deeply)
  function sortObjectRecursively(obj: any): any {
    if (Array.isArray(obj)) {
      return obj.map(sortObjectRecursively);
    } else if (obj && typeof obj === "object" && obj.constructor === Object) {
      const sortedKeys = Object.keys(obj).sort();
      const result: any = {};
      for (const key of sortedKeys) {
        result[key] = sortObjectRecursively(obj[key]);
      }
      return result;
    }
    return obj;
  }

  if (overlay) {
    // Deep merge overlay into manifest, but ensure deterministic key order in the result.
    function deepMergeDeterministic(target: any, source: any): any {
      // Collect all keys from both target and source, sort them for deterministic order
      const keys = Array.from(
        new Set([...Object.keys(target), ...Object.keys(source)]),
      ).sort();

      const result: any = Array.isArray(target) ? [] : {};

      for (const key of keys) {
        const targetValue = target[key];
        const sourceValue = source[key];

        if (sourceValue === undefined) {
          // Only in target
          result[key] = targetValue;
        } else if (targetValue === undefined) {
          // Only in source
          result[key] = sourceValue;
        } else if (sourceValue === null) {
          // Explicitly set null
          result[key] = null;
        } else if (Array.isArray(sourceValue)) {
          // Arrays: replace entirely
          result[key] = [...sourceValue];
        } else if (
          sourceValue &&
          typeof sourceValue === "object" &&
          sourceValue.constructor === Object &&
          targetValue &&
          typeof targetValue === "object" &&
          targetValue.constructor === Object
        ) {
          // Both are plain objects: merge recursively
          result[key] = deepMergeDeterministic(targetValue, sourceValue);
        } else {
          // Primitives, functions, class instances: replace entirely
          result[key] = sourceValue;
        }
      }
      return result;
    }
    manifest = deepMergeDeterministic(manifest, sortObjectRecursively(overlay));
  }

  // Return the complete JSON manifest in deterministic order
  return JSON.stringify(manifest, null, 2);
}
registerTemplateFunc("templateMCPBManifest", templateMCPBManifest);

// Generate server.json conforming to the official MCP Registry schema
// https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json
function templateMCPServerJSON() {
  const config = context.Global.Config;
  const sdk = context.Local.SDK;

  const serverJSON: any = {
    $schema:
      "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
    name: config.McpName || config.PackageName,
    version: config.SDKVersion,
    description: (sdk.Comments?.Summary || "").slice(0, 100),
  };

  // Human-readable title
  if (sdk.Comments?.Summary) {
    serverJSON.title = sdk.Comments.Summary.slice(0, 100);
  } else if (config.PackageName) {
    serverJSON.title = config.PackageName.split(/[-_]/)
      .map((word: string) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ")
      .slice(0, 100);
  }

  // Repository metadata (required fields: url, source)
  if (config.RepoURL) {
    const repoURL = config.RepoURL.replace(/\.git$/, "");
    serverJSON.repository = {
      url: repoURL,
      source: repoURL.includes("github.com")
        ? "github"
        : repoURL.includes("gitlab.com")
        ? "gitlab"
        : "custom",
    };
    if (config.RepoSubDirectory) {
      serverJSON.repository.subfolder = config.RepoSubDirectory;
    }
  }

  // Build npm package entry
  const npmPackage: any = {
    registryType: "npm",
    identifier: config.PackageName,
    version: config.SDKVersion,
    registryBaseUrl: "https://registry.npmjs.org",
    runtimeHint: "npx",
    transport: { type: "stdio" },
  };

  // Map CLI flags to environment variables and package arguments
  const flags = gatherMCPFlags();
  const packageArgs: any[] = [
    { type: "positional", value: "start", valueHint: "start_command" },
  ];

  if (flags.length > 0) {
    const envVars: any[] = [];

    for (const flag of flags) {
      const envName = flag.Name.replace(/-/g, "_").toUpperCase();
      envVars.push({
        name: envName,
        description: flag.Description || `The ${flag.Title} for the server`,
        ...(flag.Optional ? {} : { isRequired: true }),
        ...(flag.Sensitive ? { isSecret: true } : {}),
      });
      packageArgs.push({
        type: "named",
        name: `--${flag.Name}`,
        value: `{${envName}}`,
      });
    }

    npmPackage.environmentVariables = envVars;
  }

  npmPackage.packageArguments = packageArgs;
  serverJSON.packages = [npmPackage];

  return JSON.stringify(serverJSON, null, 2);
}
registerTemplateFunc("templateMCPServerJSON", templateMCPServerJSON);

function templateMcpBundleReleaseURL(): string {
  if (!context.Global.Config.RepoURL) {
    return "./mcp-server.mcpb";
  }
  const repoURL = context.Global.Config.RepoURL.replace(/\.git$/, "");
  const version = context.Global.Config.SDKVersion;
  return `${repoURL}/releases/download/v${version}/mcp-server.mcpb`;
}
registerTemplateFunc(
  "templateMcpBundleReleaseURL",
  templateMcpBundleReleaseURL,
);

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    addTestImport("./files.js", "fileToBase64");
    filePath = `.speakeasy/testfiles/${filePath}`;

    return `await fileToBase64("${filePath}")`;
  }

  return `new Uint8Array(/* Populate with bytes from file, for example ${filePath} */)`;
}

registerTemplateFunc(
  "templateFileToByteArrayValue",
  templateFileToByteArrayValue,
);

// @ts-ignore
function isConstFieldsAlwaysOptional(): boolean {
  const value = context.Global.Config.ConstFieldsAlwaysOptional;
  // Done like this because true is the default value
  if (value === false || value === "false") return false;
  return true;
}

function templateUsageMethodParameters(
  local: UsageContext,
  indent: number,
): string {
  const operation = local.Operation;
  const args = operation.Arguments;
  const methodParams = [];

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
    exampleName: local.ExampleName,
    scope: "usage",
    shouldTemplateConstValue: (field, additionalContext) => {
      if (additionalContext?.withinUnion) {
        return true;
      }

      if (isConstFieldsAlwaysOptional()) return false;
      return !field.Optional;
    },
  };

  let lastIncludedFieldIndex = -1;
  const examples = [];
  for (const [i, field] of args.Sorted.entries()) {
    const fieldExample = getOperationMethodFieldExample(local, field);
    examples[i] = fieldExample;
    if (includeField(field, fieldExample, false, ctx)) {
      lastIncludedFieldIndex = i;
    }
  }

  for (const [i, field] of args.Sorted.entries()) {
    let value = "";

    if (i > lastIncludedFieldIndex) {
      continue;
    }

    const fieldExample = examples[i];

    // Check if this is a security field and use test security value if available
    if (isSecurityClassField(field)) {
      let securityScope = undefined;

      if (local.Scopes) {
        for (const scope of local.Scopes) {
          if (scope.Feature == "security") {
            securityScope = scope;
            break;
          }
        }
      }

      let securityExample = undefined;
      if (securityScope?.Value) {
        if ("example" in securityScope.Value) {
          securityExample = securityScope.Value.example;
        } else {
          securityExample = getExampleValue(securityScope.Value, {
            usageContext: local,
          });
        }
      }

      if (securityExample !== undefined) {
        value = templateSecurityUsage(
          operation.Security.Type,
          indent,
          true,
          false,
          undefined,
          securityExample,
          ctx,
        ).trimStart();
      } else {
        value = templateValue(field, fieldExample, false, ctx);
      }
    } else {
      value = templateValue(field, fieldExample, false, ctx);
    }

    if (value === "") {
      continue;
    }
    methodParams.push(value);
  }

  const options = getOptionalUsageMethodParameters(local);
  if (options) {
    methodParams.push(options);
  }

  return methodParams.length > 0 ? ", " + methodParams.join(", ") : "";
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function getOptionalUsageMethodParameters(local: UsageContext): string {
  if (!local.Scopes) {
    return "";
  }

  const options = local.Scopes.map((scope) => {
    if (scope.IsGlobal) {
      return "";
    }

    switch (scope.Feature.toString()) {
      case "server_url": {
        return `serverURL: ${getUsageServerUrl(local, scope, {
          isTest: local.Test && true,
        })},`;
      }
    }
  }).filter(Boolean);

  return options.length > 0 ? `{\n${indentLines(options, 1)}\n}` : "";
}

function templateSDKQualifier(usageContext: UsageContext): string {
  if (usageContext.ContextIndex > 0) {
    return "";
  }

  if (!usageContext.Test) {
    return "const ";
  }

  const count = usageContext.Test.Steps.filter(
    (s) => s.Type == "operation" && !s.UsageContext.SkipSDKInstantiation,
  ).length;

  return count == 1 ? "const " : "let ";
}
registerTemplateFunc("templateSDKQualifier", templateSDKQualifier);

// @ts-ignore
function templateExampleReferenceValue(
  exampleReferenceValue: ExampleReferenceValue,
  targetField?: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const isAnyPartOfFieldOptional =
    exampleReferenceValue.parents.some((p) => {
      return p.Optional;
    }) || exampleReferenceValue.source.Optional;

  if (
    additionalContext?.isTest &&
    isAnyPartOfFieldOptional &&
    !targetField.Optional
  ) {
    addTestImport("./assertions.js", "assertDefined");
    return `assertDefined(${exampleReferenceValue.path})`;
  }

  return exampleReferenceValue.path;
}
