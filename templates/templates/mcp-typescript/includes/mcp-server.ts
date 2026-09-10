type EvalProviderMeta = {
  Package: string;
  ImportName: string;
  DefaultModel: string;
  ApiKeyEnvVar: string;
  ApiKeyFlag: string;
  ApiKeyFlagBrief: string;
};

const evalProviders: Record<string, EvalProviderMeta> = {
  anthropic: {
    Package: "@ai-sdk/anthropic",
    ImportName: "anthropic",
    DefaultModel: "claude-sonnet-4-20250514",
    ApiKeyEnvVar: "ANTHROPIC_API_KEY",
    ApiKeyFlag: "anthropic-api-key",
    ApiKeyFlagBrief:
      "Anthropic API key (defaults to ANTHROPIC_API_KEY env var)",
  },
  openai: {
    Package: "@ai-sdk/openai",
    ImportName: "openai",
    DefaultModel: "gpt-4o",
    ApiKeyEnvVar: "OPENAI_API_KEY",
    ApiKeyFlag: "openai-api-key",
    ApiKeyFlagBrief: "OpenAI API key (defaults to OPENAI_API_KEY env var)",
  },
};

const defaultMCPResourceMethods = new Set(["get"]);
const defaultMCPToolMethods = new Set([
  "post",
  "put",
  "patch",
  "delete",
  "query",
  "head",
]);
type MCPPrimitive = "tool" | "resource";
const mcpPrimitivePath: Record<MCPPrimitive, string> = {
  resource: "resources",
  tool: "tools",
};

function* mcpOperations(root: SDK): Generator<{
  op: Operation;
  mcpType: MCPPrimitive;
}> {
  const seenNames: Record<MCPPrimitive, Record<string, Operation>> = {
    resource: {},
    tool: {},
  };
  const seenPaths: Record<MCPPrimitive, Record<string, Operation>> = {
    resource: {},
    tool: {},
  };
  const seenOpIDs = new Set<string>();
  const queue: Array<SDK> = [root];

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    for (const op of sdk.Operations) {
      if (op.Extensions.MCP?.Disabled || op.Webhook) {
        continue;
      }
      if (seenOpIDs.has(op.ID)) {
        continue;
      }
      seenOpIDs.add(op.ID);

      const mcpType = getMCPPrimitive(op);
      if (!mcpType) {
        continue;
      }

      const filename = sanitizeFuncFilename(op);
      const name = sanitizeMCPPrimitiveName(op);

      const names = seenNames[mcpType];

      if (names[name]) {
        throw new Error(
          `Duplicate MCP server ${mcpType} names detected for operations "${names[name].ID}" and "${op.ID}": ${name}`,
        );
      }
      names[name] = op;

      const paths = seenPaths[mcpType];

      const dir = `src/mcp-server/${mcpPrimitivePath[mcpType]}`;
      const path = `${dir}/${filename}.ts`;
      if (paths[path]) {
        throw new Error(
          `Duplicate MCP server ${mcpType} file names detected for operations "${paths[path].ID}" and "${op.ID}": ${path}`,
        );
      }
      paths[path] = op;

      yield {
        op,
        mcpType,
      };
    }
  }
}

function getMCPServerJobs(root: SDK): Job[] {
  const jobs: Job[] = [];
  const seenNames: Record<MCPPrimitive, Record<string, Operation>> = {
    resource: {},
    tool: {},
  };
  const scopes = new Set<string>([]);

  // Process all MCP operations using the generator
  for (const { op, mcpType } of mcpOperations(root)) {
    const filename = sanitizeFuncFilename(op);
    const name = sanitizeMCPPrimitiveName(op);

    // Track seen names for later use in server.ts generation
    seenNames[mcpType][name] = op;

    const dir = `src/mcp-server/${mcpPrimitivePath[mcpType]}`;
    const path = `${dir}/${filename}.ts`;

    let templateName: string = mcpType;
    if (mcpType === "resource" && isResourceTemplate(op)) {
      templateName = "resource-template";
    }

    op.Extensions.MCP?.Scopes?.forEach((s) => scopes.add(s));

    // tool.ts.stmpl resource.ts.stmpl resource-template.ts.stmpl prompt.ts.stmpl
    jobs.push(
      createTemplateFileJob(`mcp-server/${templateName}.ts.stmpl`, path, {
        DocGroup: path,
        Operation: op,
        Dir: dir,
        Filename: filename,
      }),
    );
  }

  jobs.push(
    createTemplateFileJob(
      "mcp-server/scopes.ts.stmpl",
      "src/mcp-server/scopes.ts",
      {
        Scopes: Array.from(scopes).sort(),
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/flags.ts.stmpl",
      "src/mcp-server/flags.ts",
      {
        Scopes: Array.from(scopes).sort(),
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/server.ts.stmpl",
      "src/mcp-server/server.ts",
      {
        Tools: Object.values(seenNames.tool),
        Resources: Object.values(seenNames.resource),
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/cli/start/command.ts.stmpl",
      "src/mcp-server/cli/start/command.ts",
      {
        Scopes: Array.from(scopes).sort(),
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/cli/start/impl.ts.stmpl",
      "src/mcp-server/cli/start/impl.ts",
      {
        Scopes: Array.from(scopes).sort(),
        UsageLocation: "mcp-server/cli/start",
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/cli/serve/command.ts.stmpl",
      "src/mcp-server/cli/serve/command.ts",
      {
        Scopes: Array.from(scopes).sort(),
      },
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "mcp-server/cli/serve/impl.ts.stmpl",
      "src/mcp-server/cli/serve/impl.ts",
      {
        Scopes: Array.from(scopes).sort(),
        UsageLocation: "mcp-server/cli/serve",
      },
    ),
  );
  if (context.Global.Config.EvalProvider) {
    const evalProviderKey = context.Global.Config.EvalProvider;
    const evalProvider = evalProviders[evalProviderKey];
    if (!evalProvider) {
      throw new Error(
        `Unknown eval provider: ${evalProviderKey}. Supported: ${Object.keys(
          evalProviders,
        ).join(", ")}`,
      );
    }

    jobs.push(
      createTemplateFileJob(
        "mcp-server/cli/eval/command.ts.stmpl",
        "src/mcp-server/cli/eval/command.ts",
        {
          Scopes: Array.from(scopes).sort(),
          EvalProvider: evalProvider,
        },
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "mcp-server/cli/eval/impl.ts.stmpl",
        "src/mcp-server/cli/eval/impl.ts",
        {
          Scopes: Array.from(scopes).sort(),
          UsageLocation: "mcp-server/cli/eval",
          EvalProvider: evalProvider,
        },
      ),
    );
  }
  jobs.push(
    createTemplateFileJob("manifest.json.stmpl", "manifest.json", {
      SDK: root,
    }),
  );

  // Generate server.json for Anthropic MCP registry if mcpName is configured
  if (context.Global.Config.McpName) {
    jobs.push(
      createTemplateFileJob("server.json.stmpl", "server.json", {
        SDK: root,
      }),
    );
  }

  // Generate gram.ts for Gram deployment if gramEnabled is configured
  if (context.Global.Config.GramEnabled) {
    jobs.push(
      createTemplateFileJob(
        "mcp-server/gram.ts.stmpl",
        "src/mcp-server/gram.ts",
        {
          Scopes: Array.from(scopes).sort(),
          UsageLocation: "mcp-server",
        },
      ),
    );
  }

  return jobs;
}

function getMCPPrimitive(op: Operation): MCPPrimitive | null {
  // Per-operation security is not currently compatible with MCP
  if (op.Security) {
    return null;
  }

  // // Blocked by: https://github.com/modelcontextprotocol/typescript-sdk/issues/149
  // const method = op.Method.toLowerCase();
  // if (defaultMCPResourceMethods.has(method)) {
  //   return "resource";
  // }
  // if (defaultMCPToolMethods.has(method)) {
  //   return "tool";
  // }

  // return null;

  return "tool";
}
registerTemplateFunc("getMCPPrimitive", getMCPPrimitive);

function sanitizeMCPPrimitiveExport(operation: Operation): string {
  const type = getMCPPrimitive(operation);
  return `${type}$${sanitizeFuncName(operation)}`;
}
registerTemplateFunc("sanitizeMCPPrimitiveExport", sanitizeMCPPrimitiveExport);

const MCP_NAME_MAX_LENGTH = 52;
const MCP_NAME_HASH_LENGTH = 3;

function mcpNameHash(str: string): string {
  let hash = 0x811c9dc5; // FNV-1a offset basis
  for (let i = 0; i < str.length; i++) {
    hash ^= str.charCodeAt(i);
    hash = (hash * 0x01000193) | 0;
  }
  return (hash >>> 0)
    .toString(16)
    .padStart(MCP_NAME_HASH_LENGTH, "0")
    .slice(0, MCP_NAME_HASH_LENGTH);
}

function sanitizeMCPToolName(fullName: string, lineNumber: number): string {
  // Strip characters not matching [a-zA-Z0-9_-]
  const sanitized = fullName.replace(/[^a-zA-Z0-9_-]/g, "");

  if (sanitized.length <= MCP_NAME_MAX_LENGTH) {
    return sanitized;
  }

  const hash = mcpNameHash(sanitized);
  const maxPrefixLen = MCP_NAME_MAX_LENGTH - 1 - MCP_NAME_HASH_LENGTH; // 48

  const parts = sanitized.split("-");
  let prefix = "";
  for (const part of parts) {
    const candidate = prefix ? `${prefix}-${part}` : part;
    if (candidate.length > maxPrefixLen) break;
    prefix = candidate;
  }

  if (!prefix) {
    prefix = sanitized.slice(0, maxPrefixLen);
  }

  const truncated = `${prefix}-${hash}`;

  logWarning(
    `MCP tool name "${fullName}" (${fullName.length} chars) exceeds ${MCP_NAME_MAX_LENGTH}-char limit, ` +
      `truncated to "${truncated}". Use x-speakeasy-mcp-name to set a custom tool name.`,
    lineNumber,
  );

  return truncated;
}

function sanitizeMCPPrimitiveName(operation: Operation): string {
  if (operation.Extensions.MCP?.Name) {
    return operation.Extensions.MCP.Name;
  }

  const group = operation.OwningSDK.Group;
  let prefix = "";
  // We only qualify sub-sdk functions, not the root SDKs functions.
  if (group) {
    prefix = caser().ToKebab(group) + "-";
  }

  const name = caser().ToKebab(sanitizeName(operation.GetID()));
  const lineNumber = operation.Location?.Line ?? 0;
  return sanitizeMCPToolName(`${prefix}${name}`, lineNumber);
}
registerTemplateFunc("sanitizeMCPPrimitiveName", sanitizeMCPPrimitiveName);

function sanitizeMCPServerName(sdk?: SDK) {
  const s = sdk ?? context.Global.AST.MainSDK;
  const name = sanitizeName(s.Type.Name);
  return caser().ToPascal(name);
}
registerTemplateFunc("sanitizeMCPServerName", sanitizeMCPServerName);

function addMCPPrimitiveImport(op: Operation, usageLocation: string) {
  const name = sanitizeMCPPrimitiveExport(op);
  const filename = sanitizeFuncFilename(op);
  const mcpType = getMCPPrimitive(op);
  const dir = mcpPrimitivePath[mcpType];
  if (!dir) {
    throw new Error(`${op.ID}: Invalid MCP primitive path for: ${mcpType}`);
  }

  const prefix = getImportPrefix(`mcp-server/${dir}`, usageLocation);
  return addImport(`${prefix}/${filename}.js`, name, typeImport);
}

registerTemplateFunc("addMCPPrimitiveImport", addMCPPrimitiveImport);

function resourceURITemplate(op: Operation): string {
  const rootSDK = context.Global.AST.MainSDK;
  const scheme = caser().ToKebab(rootSDK.Type.Name.toLowerCase());

  const q = op.Request?.Params?.QueryParams?.flatMap((param) => {
    if (param.Hidden) {
      return [];
    }

    return [sanitizeFieldName(param.Field.Name)];
  }).join(",");

  const qTemplate = q ? `{?${q}}` : "";

  return `${scheme}:/${op.Path}${qTemplate}`;
}
registerTemplateFunc("resourceURITemplate", resourceURITemplate);

function isResourceTemplate(op: Operation): boolean {
  const hasPathParam = op.Request?.Params?.PathParams?.length > 0;
  const hasQuery = op.Request?.Params?.QueryParams?.some(
    (param) => !param.Hidden,
  );

  return hasPathParam || hasQuery;
}
registerTemplateFunc("isResourceTemplate", isResourceTemplate);

// Derived from DXT UserConfiguration
// https://github.com/anthropics/dxt/blob/main/MANIFEST.md#user-configuration
type FlagDescriptor = {
  Name: string;
  Type: "string" | "number" | "boolean" | "directory" | "file";
  Title: string;
  Description?: string;
  Optional?: boolean;
  Default?: string;
  Multiple?: boolean;
  Sensitive?: boolean;
  Min?: number;
  Max?: number;
};

function typeDefToFlagType(typeDef: TypeDef): FlagDescriptor["Type"] {
  const typeStr = typeDef.Type.toString() as DataType;
  switch (typeStr) {
    case "integer":
    case "int32":
    case "bigint":
    case "number":
    case "float32":
    case "decimal":
      return "number";
    case "boolean":
      return "boolean";
    default:
      return "string";
  }
}

function gatherServerFlags(servers: Servers | null): Array<FlagDescriptor> {
  const flags: Array<FlagDescriptor> = [];
  const hasServers = servers?.HasAbsoluteURL();

  if (!hasServers) {
    flags.push({
      Name: "server-url",
      Title: "Server URL",
      Description: "The URL of the server to connect to",
      Type: "string",
    });
    return flags;
  }

  // Server selection flags
  if (servers.ServerMap) {
    const serverDescriptions = servers.Servers.map(
      (s) => `${s.ID}: ${s.URL}`,
    ).join(", ");
    flags.push({
      Name: "server",
      Title: "Server",
      Description: `Available servers: [${serverDescriptions}]`,
      Type: "string",
      Optional: true,
      Default: servers.Default || servers.Servers[0]?.ID,
    });
  } else if (servers.Servers && servers.Servers.length > 1) {
    const serverDescriptions = servers.Servers.map((s) => `${s.URL}`).join(
      ", ",
    );
    flags.push({
      Name: "server-index",
      Title: "Server Index",
      Description: `Available servers: [${serverDescriptions}]`,
      Type: "number",
      Optional: true,
      Default: "0",
      Min: 0,
      Max: servers.Servers.length - 1,
    });
  }

  // Server variable flags - with server correlation
  const variables = servers.GetVariables();
  for (const variable of variables) {
    let description = `${variable.Type.Comments?.Description || ""} `;

    // Add server correlation info if multiple servers exist
    if (servers.Servers.length > 1) {
      description += `Templated in ${variable.Server.URL}.`;
    }

    // Add enum options to description if available
    if (variable.Type.Enum && variable.Type.Enum.Values) {
      const options = variable.Type.Enum.Values.join(", ");
      description += ` Options: [${options}].`;
    }

    const flagDescriptor: FlagDescriptor = {
      Name: sanitizeMCPCLIFlag(variable.Name),
      Title: variable.Name,
      Description: description,
      Type: typeDefToFlagType(variable.Type),
      Optional: true,
      Default: variable.Default,
    };

    flags.push(flagDescriptor);
  }

  return flags;
}

function gatherMCPFlags(): Array<FlagDescriptor> {
  const globalSDK = context.Global.AST.MainSDK;
  const security = globalSDK.Security?.Type;
  const globals = globalSDK.Globals;
  const servers = globalSDK.Servers;
  let flags: Array<FlagDescriptor> = [];

  flags.push(...gatherServerFlags(servers));

  if (security) {
    unnestSecurityEnvFields(security).forEach((field) => {
      flags.push({
        Name: sanitizeMCPCLIFlag(field.Name),
        Type: typeDefToFlagType(field.Type),
        Title: sanitizeFieldName(field.Name),
        ...(field.Optional ? { Optional: field.Optional } : {}),
        ...(field.Default ? { Default: field.Default?.Value } : {}),
        ...(field.Comments?.Description
          ? { Description: field.Comments.Description }
          : {}),
        // For string types, mask input and store securely
        Sensitive: field.Type.Type.toString() === "string",
      });
    });
  }
  if (globals) {
    globals.Fields.forEach((field) => {
      if (!field.Type.IsPrimitive()) {
        return;
      }
      flags.push({
        Name: sanitizeMCPCLIFlag(field.Name),
        Type: typeDefToFlagType(field.Type),
        Title: sanitizeFieldName(field.Name),
        Optional: field.Optional,
        ...(field.Default ? { Default: field.Default?.Value } : {}),
        ...(field.Comments?.Description
          ? { Description: field.Comments.Description }
          : {}),
      });
    });
  }

  // exclude advanced flags, get the list of advanced flags from the config
  const advancedFlags = context.Global.Config.AdvancedFlags;
  if (advancedFlags) {
    flags = flags.filter(
      (flag) => !advancedFlags.includes(sanitizeFieldName(flag.Name)),
    );
  }

  return flags;
}
registerTemplateFunc("gatherMCPFlags", gatherMCPFlags);

function generateMCPFlagsUsage(): Array<{ key: string; value: string }> {
  const flags = gatherMCPFlags();
  const vars: Array<{ key: string; value: string }> = [];

  for (const flag of flags) {
    vars.push({ key: `--${flag.Name}`, value: flag.Default || "" });
  }

  return vars;
}
registerTemplateFunc("generateMCPFlagsUsage", generateMCPFlagsUsage);

function sanitizeMCPComments(comment: string): string {
  return sanitizeComments(comment)
    .replaceAll(/\\/g, "\\\\")
    .replaceAll("`", "\\`")
    .replaceAll(/(\$\{.*\})/g, "\\$1");
}

function sanitizeMCPFieldDescription(
  field: FieldDef,
  zodSchema: string,
): string {
  if (zodSchema.includes(".describe(")) {
    return "";
  }
  const summary = field.Comments?.Summary || "";
  const description = field.Comments?.Description || "";
  const combined = [summary, description].filter(Boolean).join("\n\n");
  if (!combined) {
    return "";
  }
  return sanitizeMCPComments(combined);
}
registerTemplateFunc(
  "sanitizeMCPFieldDescription",
  sanitizeMCPFieldDescription,
);

function sanitizeMCPPrimitiveDescription(op: Operation): string {
  if (op.Extensions.MCP?.Description) {
    return sanitizeMCPComments(op.Extensions.MCP.Description);
  }

  const summary = op.Comments?.Summary || "";
  const description = op.Comments?.Description || "";

  return sanitizeMCPComments(
    [summary, description].filter(Boolean).join("\n\n"),
  );
}
registerTemplateFunc(
  "sanitizeMCPPrimitiveDescription",
  sanitizeMCPPrimitiveDescription,
);

function sanitizeMCPCLIFlag(name: string, quoteIfNeeded = false): string {
  let value = caser().ToKebab(sanitizeName(name)).toLowerCase();
  // We want to reserve all short flags so if they appear anywhere in the SDK
  // server variables then we qualify them.
  if (reservedMCPCLIFlags.has(value) || value.length === 1) {
    value = `api-${value}`;
  }

  if (quoteIfNeeded && value.includes("-")) {
    return `"${value}"`;
  }

  return value;
}
registerTemplateFunc("sanitizeMCPCLIFlag", sanitizeMCPCLIFlag);

function hasResponseBody(op: Operation): boolean {
  for (let resp of op.Response.Responses) {
    if (!resp.Error && resp.Content.length) {
      return true;
    }
  }

  return false;
}
registerTemplateFunc("hasResponseBody", hasResponseBody);

function templateMCPSecurityAccess(
  secField: FieldDef,
  usageLocation: string,
  topLevel = true,
  allRequired = false,
): string {
  let result = "";

  let flatAndBasic = topLevel;

  for (const field of secField.Type.Fields) {
    const ann = field.Annotations?.Get("security");
    const fieldName = ann?.Option ? field.Name : originalFieldName(field);

    if (ann?.Option) {
      flatAndBasic = false;
      const res = templateMCPSecurityAccess(field, usageLocation, false, true);
      addInternalImport("primitives", "allRequired", usageLocation);
      result += `${sanitizeFieldName(fieldName)}: allRequired(${res}),\n`;
      continue;
    }

    if (flatAndBasic) {
      flatAndBasic =
        flatAndBasic &&
        ann?.SecType === "http" &&
        ann?.SubType === "basic" &&
        !!ann?.FieldName;
    }

    const [k, v] = templateMCPSecurityFieldAccess(field, allRequired);
    result += `${k}: ${v},\n`;
  }

  // Annoyingly if http basic is the only scheme, we generate global security
  // type with two fields instead of one field which would play nice with global
  // security flattening. As a result, the username and password fields can
  // appear as distinct security schemes unfortunately. When that happens, we
  // need to make sure both fields are explicitly set in this way.
  if (flatAndBasic) {
    addInternalImport("primitives", "allRequired", usageLocation);
    return `allRequired({${result}})`;
  }

  return `{${result}}`;
}
registerTemplateFunc("templateMCPSecurityAccess", templateMCPSecurityAccess);

function templateMCPSecurityFieldAccess(
  secField: FieldDef,
  allRequired = false,
): readonly [key: string, value: string] {
  const fieldName = sanitizeFieldName(originalFieldName(secField));
  const flagName = mcpflag("flags", secField.Name);
  let flagValue = flagName;

  const templateEmpty = (secField: FieldDef): string => {
    if (secField.Default || allRequired) {
      return "";
    }

    switch (secField.Type.Type.toString()) {
      case "string":
        return ` ?? ""`;
      default:
        return ""; // TODO more cases?
    }
  };

  if (secField.Default) {
    flagValue = `${flagValue} ?? ${JSON.stringify(secField.Default.Value)}`;
  }

  flagValue = `${flagValue}${templateEmpty(secField)}`;

  const DEFAULT = [fieldName, flagValue] as const;

  const ann = secField.Annotations?.Get("security");
  if (!ann || !isSecurityAnnotation(ann) || !ann.Scheme) {
    return DEFAULT;
  }

  const type = `${ann.SecType}:${ann.SubType}`;

  switch (type) {
    case "http:basic": {
      if (!ann.FieldName) {
        const user = mcpflag("flags", "username");
        const pass = mcpflag("flags", "password");
        return [
          fieldName,
          `${user} != null && ${pass} != null ? { username: ${user}, password: ${pass} } : void 0`,
        ];
      } else {
        return DEFAULT;
      }
    }
    case "http:custom": {
      if (secField.Type.Fields.length > 0) {
        const fields = secField.Type.Fields.map((f) => {
          const k = sanitizeFieldName(originalFieldName(f));
          const v = mcpflag("flags", f.Name);
          return `${k}: ${v} ?? ""`;
        }).join(", ");

        if (!secField.Optional) {
          return [fieldName, `{ ${fields} }`];
        }

        const fieldCheck = secField.Type.Fields.map(
          (f) => `${mcpflag("flags", f.Name)} != null`,
        ).join(" && ");

        return [fieldName, `${fieldCheck} ? { ${fields} } : void 0`];
      }

      return DEFAULT;
    }
    case "oauth2:client_credentials": {
      if (isClientCredentialsFlattened() || ann.FieldName) {
        return DEFAULT;
      } else {
        const cid = mcpflag("flags", "client-id");
        const cs = mcpflag("flags", "client-secret");
        const t = mcpflag("flags", "token-url");
        const coreNames = new Set([
          "clientid",
          "clientsecret",
          "tokenurl",
          "scopes",
        ]);
        const additionalKVs = (secField.Type?.Fields || [])
          .filter(
            (f: FieldDef) =>
              !coreNames.has(f.Name.replace(/_/g, "").toLowerCase()) &&
              f.Type?.Type.toString() === "string",
          )
          .map(
            (f: FieldDef) =>
              `${sanitizeFieldName(originalFieldName(f))}: ${mcpflag(
                "flags",
                f.Name,
              )} ?? ''`,
          );
        const additionalStr =
          additionalKVs.length > 0 ? `, ${additionalKVs.join(", ")}` : "";
        let v = `{ clientID: ${cid} || '', clientSecret: ${cs} || '', tokenURL: ${t} || ''${additionalStr} }`;
        if (secField.Optional) {
          v = `${cid} != null && ${cs} != null ? ${v} : undefined`;
        }
        return [fieldName, v];
      }
    }
    case "oauth2:password": {
      if (isFeatureUsed("oauth2Password")) {
        const username = mcpflag("flags", "username");
        const password = mcpflag("flags", "password");
        const clientID = mcpflag("flags", "client-id");
        const clientSecret = mcpflag("flags", "client-secret");
        const tokenURL = mcpflag("flags", "token-url");
        const token = mcpflag("flags", "token");
        return [
          fieldName,
          `${username} != null && ${password} != null ` +
            `? { clientID: ${clientID}, clientSecret: ${clientSecret}, username: ${username}, password: ${password}, tokenURL: ${tokenURL} } ` +
            `: ${token} ?? ""`,
        ];
      } else {
        return DEFAULT;
      }
    }
    default:
      return DEFAULT;
  }
}

function mcpflag(flagsVar: string, flag: string): string {
  return sanitizeAccessor(flagsVar, sanitizeMCPCLIFlag(flag));
}

registerTemplateFunc("mcpflag", mcpflag);

/**
 * Generates the flag override assignments for the SSE handler, where security
 * credentials can be overridden via HTTP request headers. Each field type is
 * converted appropriately: strings are passed through, integers are cast via
 * Number(), arrays are split by comma, and enums are type-asserted. If the
 * header is absent or conversion fails, the CLI flag value is used as fallback.
 */
function templateMCPFlagHeaderOverrides(security: TypeDef): string {
  const fields = unnestSecurityEnvFields(security);
  if (fields.length === 0) return "";

  const lines: string[] = [];
  const seen = new Set<string>();

  for (const field of fields) {
    const flagName = sanitizeMCPCLIFlag(field.Name, false);
    if (seen.has(flagName)) continue;
    seen.add(flagName);

    // Express/Node.js normalizes req.headers keys to lowercase, so we must
    // lowercase the field name to match.
    const headerName = originalFieldName(field).toLowerCase();
    const fieldType = field.Type.Type.toString();
    const header = `req.headers["${headerName}"]`;
    const fallback = `cliFlags["${flagName}"]`;

    let headerExpr: string;
    switch (fieldType) {
      case "string":
        headerExpr = `(${header} as string)`;
        break;
      case "integer":
      case "number":
      case "int":
        headerExpr = `(typeof ${header} === "string" ? Number(${header}) : undefined)`;
        break;
      case "array":
        headerExpr = `(typeof ${header} === "string" ? ${header}.split(",") : undefined)`;
        break;
      case "enum":
        headerExpr = `(${header} as string as typeof ${fallback})`;
        break;
      default:
        headerExpr = `(${header} as typeof ${fallback})`;
        break;
    }

    lines.push(`"${flagName}": ${headerExpr} ?? ${fallback},`);
  }

  return lines.join("\n      ");
}

registerTemplateFunc(
  "templateMCPFlagHeaderOverrides",
  templateMCPFlagHeaderOverrides,
);

function sanitizeMCPAnnotations(op: Operation): string {
  const mcp = op.Extensions?.MCP;

  let destructiveHint =
    mcp?.DestructiveHint ?? op.Method.toUpperCase() === "DELETE";
  let readOnlyHint = mcp?.ReadOnlyHint ?? op.Method.toUpperCase() === "GET";

  const title = mcp?.Title ?? mcp?.Name;

  let annotations = {
    title: title ?? "",
    destructiveHint: destructiveHint,
    idempotentHint: mcp?.IdempotentHint ?? false,
    openWorldHint: mcp?.OpenWorldHint ?? false,
    readOnlyHint: readOnlyHint,
  };

  return JSON.stringify(annotations);
}
registerTemplateFunc("sanitizeMCPAnnotations", sanitizeMCPAnnotations);

function isGlobalServerVar(variable: string): boolean {
  for (const v of context?.Global?.AST.MainSDK.Servers.GetVariables()) {
    if (v.Name === variable) {
      return true;
    }
  }
  return false;
}
registerTemplateFunc("isGlobalServerVar", isGlobalServerVar);

function isStrictMCPServer(): boolean {
  return !!context.Global.Config.ValidateResponse;
}
registerTemplateFunc("isStrictMCPServer", isStrictMCPServer);
