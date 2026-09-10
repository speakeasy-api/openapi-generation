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

function getMCPServerJobs(root: SDK): Job[] {
  const jobs: Job[] = [];

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

  const scopes = new Set<string>([]);

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    const ops = sdk.Operations.filter((op) => !op.Webhook);
    for (const op of ops) {
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

      const filename = sanitizeMCPPrimitiveFilename(op);
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
  }

  jobs.push(getTemplateAuxiliaryFilesJob("mcp-server/auxiliary"));
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
      },
    ),
  );

  return jobs;
}

function isMCPServerEnabled(): boolean {
  return isFeatureUsed("mcpServer");
}
registerTemplateFunc("isMCPServerEnabled", isMCPServerEnabled);

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
  return `${prefix}${name}`;
}
registerTemplateFunc("sanitizeMCPPrimitiveName", sanitizeMCPPrimitiveName);

function sanitizeMCPPrimitiveFilename(operation: Operation): string {
  return sanitizeFuncFilename(operation);
}
registerTemplateFunc(
  "sanitizeMCPPrimitiveFilename",
  sanitizeMCPPrimitiveFilename,
);

function sanitizeMCPServerName(sdk?: SDK) {
  const s = sdk ?? context.Global.AST.MainSDK;
  const name = sanitizeName(s.Type.Name);
  return caser().ToPascal(name);
}
registerTemplateFunc("sanitizeMCPServerName", sanitizeMCPServerName);

function addMCPPrimitiveImport(op: Operation, usageLocation: string) {
  const name = sanitizeMCPPrimitiveExport(op);
  const filename = sanitizeMCPPrimitiveFilename(op);
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

function generateMCPFlagsUsage(): Array<{ key: string; value: string }> {
  const globalSDK = context.Global.AST.MainSDK;
  const security = globalSDK.Security?.Type;
  const globals = globalSDK.Globals;
  const hasServers = globalSDK.Servers?.HasAbsoluteURL();
  const vars: Array<{ key: string; value: string }> = [];

  if (!hasServers) {
    vars.push({ key: `--${sanitizeMCPCLIFlag("server-url")}`, value: "..." });
  }

  if (security) {
    unnestSecurityEnvFields(security).forEach((field) => {
      vars.push({ key: `--${sanitizeMCPCLIFlag(field.Name)}`, value: "..." });
    });
  }
  if (globals) {
    globals.Fields.forEach((field) => {
      if (!field.Type.IsPrimitive()) {
        return;
      }
      vars.push({ key: `--${sanitizeMCPCLIFlag(field.Name)}`, value: "..." });
    });
  }

  return vars;
}
registerTemplateFunc("generateMCPFlagsUsage", generateMCPFlagsUsage);

function sanitizeMCPComments(comment: string): string {
  return sanitizeComments(comment)
    .replaceAll("`", "\\`")
    .replaceAll(/\\+\{/g, "{")
    .replaceAll(/\\+\}/g, "}")
    .replaceAll(/(\$\{.*\})/g, "\\$1");
}

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

const reservedMCPCLIFlags = new Set([
  "auth",
  "env",
  "h",
  "help",
  "host-name",
  "host",
  "hostname",
  "log-level",
  "port",
  "prompt",
  "prompts",
  "resource",
  "resources",
  "scope",
  "token",
  "tool",
  "tools",
  "transport",
  "v",
  "version",
]);

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

function isResponseEnveloped(op: Operation): boolean {
  const responseFormat = getResponseFormat();

  if (responseFormat !== "flat") {
    throw new Error(`Unsupported response format: ${responseFormat}`);
  }

  if (op.Extensions.Pagination) {
    return true;
  }

  for (let resp of op.Response.Responses) {
    if (resp.Headers) {
      return true;
    }
  }
}
registerTemplateFunc("isResponseEnveloped", isResponseEnveloped);

function hasResponseBody(op: Operation): boolean {
  const responseFormat = getResponseFormat();

  if (responseFormat !== "flat") {
    throw new Error(`Unsupported response format: ${responseFormat}`);
  }

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
  topLevel = true,
  allRequired = false,
): string {
  let result = "";

  let flatAndBasic = topLevel;

  for (let field of secField.Type.Fields) {
    const fieldName = sanitizeFieldName(field.Name);
    const ann = field.Annotations?.Get("security");
    if (ann?.Option) {
      flatAndBasic = false;
      const res = templateMCPSecurityAccess(field, false, true);
      addImport("../../../lib/primitives.js", "allRequired");
      result += `${fieldName}: allRequired(${res}),\n`;
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
    addImport("../../../lib/primitives.js", "allRequired");
    return `allRequired({${result}})`;
  }

  return `{${result}}`;
}
registerTemplateFunc("templateMCPSecurityAccess", templateMCPSecurityAccess);

function templateMCPSecurityAccessInline(secField: FieldDef): string {
  const obj = templateMCPSecurityAccess(secField);
  if (obj.startsWith("{") && obj.endsWith("}")) {
    return obj.slice(1, -1);
  }
  return `...${obj}`;
}
registerTemplateFunc(
  "templateMCPSecurityAccessInline",
  templateMCPSecurityAccessInline,
);

function templateMCPSecurityFieldAccess(
  secField: FieldDef,
  allRequired = false,
): readonly [key: string, value: string] {
  const fieldName = sanitizeFieldName(secField.Name);
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
      if (!ann.FieldName || secField.Type.Fields.length > 0) {
        const fields = secField.Type.Fields.map((f) => {
          const k = sanitizeFieldName(f.Name);
          const v = mcpflag("flags", f.Name);
          return `${k}: ${v}`;
        }).join(", ");

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
        return [
          fieldName,
          `${cid} != null && ${cs} != null ? { clientID: ${cid}, clientSecret: ${cs}, tokenURL: ${t} } : void 0`,
        ];
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
            `: ${token}`,
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
