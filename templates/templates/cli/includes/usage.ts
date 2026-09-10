// =============================================================================
// Usage/KDL schema generation for generated CLIs
// =============================================================================

interface UsageFlagDef {
  spec: string;
  help?: string;
  longHelp?: string;
  global?: boolean;
  defaultValue?: string | number | boolean;
  env?: string;
  config?: string;
  count?: boolean;
  variadic?: boolean;
  hidden?: boolean;
}

interface UsageArgDef {
  name: string;
  help?: string;
  required?: boolean;
  variadic?: boolean;
}

interface UsageCommandDef {
  name: string;
  help?: string;
  aliases?: string[];
  args?: UsageArgDef[];
  flags?: UsageFlagDef[];
  commands?: UsageCommandDef[];
}

function parseQuotedGoStringLiteral(input: string): string {
  const trimmed = input.trim();
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    try {
      return JSON.parse(trimmed);
    } catch (_err) {
      return trimmed.slice(1, -1);
    }
  }
  if (trimmed.startsWith("`") && trimmed.endsWith("`")) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

// KDL v2 forbids raw control characters, delete, NEL, LS/PS, direction
// control characters, and U+FEFF inside quoted strings; escape them with the
// \u{...} form the dialect defines.
const kdlForbiddenChars =
  // eslint-disable-next-line no-control-regex
  /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f\u0085\u200e\u200f\u2028\u2029\u202a-\u202e\u2066-\u2069\ufeff]/g;

function kdlQuoted(value: string): string {
  return `"${value
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\r")
    .replace(/\t/g, "\\t")
    .replace(
      kdlForbiddenChars,
      (c) => `\\u{${c.codePointAt(0)!.toString(16)}}`,
    )}"`;
}

// Breaks any run of three quotes so it cannot terminate a multi-line string:
// the first two stay literal (legal inside """ ... """), the third is escaped.
function kdlEscapeMultilineLine(line: string): string {
  let out = "";
  let quoteRun = 0;
  for (const c of line) {
    if (c === '"') {
      quoteRun++;
      if (quoteRun === 3) {
        out += '\\"';
        quoteRun = 0;
        continue;
      }
      out += '"';
      continue;
    }
    quoteRun = 0;
    if (c === "\\") {
      out += "\\\\";
    } else if (c === "\r") {
      out += "\\r";
    } else if (c === "\t") {
      out += "\\t";
    } else if (kdlForbiddenChars.test(c)) {
      out += `\\u{${c.codePointAt(0)!.toString(16)}}`;
      kdlForbiddenChars.lastIndex = 0;
    } else {
      out += c;
    }
  }
  return out;
}

// A value containing newlines renders as a KDL v2 multi-line string so help
// text stays readable in --usage output instead of collapsing into one line
// of \n escapes. KDL's dedent rules strip the closing line's whitespace
// prefix from every content line and treat whitespace-only lines as empty,
// so values whose lines carry trailing whitespace or are whitespace-only
// fall back to the lossless single-line escaped form.
function kdlStringAt(value: string, contentPrefix: string): string {
  if (!value.includes("\n")) {
    return kdlQuoted(value);
  }
  const lines = value.split("\n");
  if (lines.some((l) => l !== "" && (/^\s+$/.test(l) || /\s$/.test(l)))) {
    return kdlQuoted(value);
  }
  const body = lines
    .map((l) =>
      l === "" ? "" : `${contentPrefix}${kdlEscapeMultilineLine(l)}`,
    )
    .join("\n");
  return `"""\n${body}\n${contentPrefix}"""`;
}

function kdlScalar(value: string | number | boolean): string {
  if (typeof value === "boolean") {
    return value ? "#true" : "#false";
  }
  if (typeof value === "number") {
    return Number.isFinite(value) ? String(value) : kdlQuoted(String(value));
  }
  return kdlQuoted(value);
}

function kdlScalarAt(
  value: string | number | boolean,
  contentPrefix: string,
): string {
  if (typeof value === "string") {
    return kdlStringAt(value, contentPrefix);
  }
  return kdlScalar(value);
}

function kdlProp(name: string, value: string | number | boolean): string {
  return `${name}=${kdlScalar(value)}`;
}

function kdlPropAt(
  name: string,
  value: string | number | boolean,
  contentPrefix: string,
): string {
  return `${name}=${kdlScalarAt(value, contentPrefix)}`;
}

function pushIf<T>(arr: T[], value: T | "" | undefined | null) {
  if (value !== undefined && value !== null && value !== "") {
    arr.push(value as T);
  }
}

function usageRootAbout(): string {
  return parseQuotedGoStringLiteral(templateRootLong());
}

// Whether the root about text renders as a KDL multi-line string — only
// newline-bearing descriptions do; single-line ones stay ordinary strings.
function templateUsageRootAboutMultiline(): boolean {
  return usageRootAbout().includes("\n");
}
registerTemplateFunc(
  "templateUsageRootAboutMultiline",
  templateUsageRootAboutMultiline,
);

function usageRootHelp(): string {
  return parseQuotedGoStringLiteral(templateRootShort());
}

function usageGroupHelp(sdk: SDK): string {
  return parseQuotedGoStringLiteral(templateGroupShort(sdk));
}

function usageOperationHelp(op: Operation): string {
  return parseQuotedGoStringLiteral(templateCmdShort(op));
}

function usageFlagKindToArgName(
  kind: string,
  flagName: string,
): string | undefined {
  const last = flagName.split(".").pop() || flagName;
  switch (kind) {
    case "FlagKindBool":
      return undefined;
    case "FlagKindFile":
      return "file";
    case "FlagKindFileArray":
      return "file";
    case "FlagKindInt64":
      return last.replace(/-/g, "_") || "value";
    case "FlagKindFloat64":
      return last.replace(/-/g, "_") || "value";
    case "FlagKindDateTime":
      return last.replace(/-/g, "_") || "datetime";
    case "FlagKindDate":
      return last.replace(/-/g, "_") || "date";
    case "FlagKindStringArray":
      return last.replace(/-/g, "_") || "value";
    default:
      return last.replace(/-/g, "_") || "value";
  }
}

function usageFlagSpec(
  flagName: string,
  shorthand?: string,
  argName?: string,
  variadic?: boolean,
): string {
  const parts: string[] = [];
  if (shorthand) {
    parts.push(`-${shorthand}`);
  }
  parts.push(
    `--${flagName}${argName ? ` <${argName}${variadic ? "..." : ""}>` : ""}`,
  );
  return parts.join(" ");
}

function inferKindNameForField(field: FieldDef, kindOverride?: string): string {
  const typeDef = field.Type;
  if (kindOverride) return kindOverride;
  if (isEnumType(typeDef)) {
    return isIntBackedEnum(typeDef) ? "FlagKindIntEnum" : "FlagKindEnum";
  }
  if (isArrayType(typeDef)) {
    const itemTypeStr = typeDef.ItemType?.Type?.toString() || "string";
    if (itemTypeStr === "string" || itemTypeStr === "enum") {
      return "FlagKindStringArray";
    }
    return "FlagKindJSON";
  }
  if (typeDef.Type.toString() === "date-time") {
    return "FlagKindDateTime";
  }
  switch (typeDef.Type.toString()) {
    case "string":
      return "FlagKindString";
    case "date":
      return "FlagKindDate";
    case "boolean":
      return "FlagKindBool";
    case "integer":
    case "int32":
      return "FlagKindInt64";
    case "number":
    case "float32":
      return "FlagKindFloat64";
    default:
      return "FlagKindString";
  }
}

function usageFlagFromField(
  field: FieldDef,
  flagName: string,
  kindOverride?: string,
): UsageFlagDef {
  const kind = inferKindNameForField(field, kindOverride);
  const argName = usageFlagKindToArgName(kind, flagName);
  const out: UsageFlagDef = {
    spec: usageFlagSpec(
      flagName,
      undefined,
      argName,
      kind === "FlagKindStringArray",
    ),
    help: getFlagDescription(field),
  };

  if (field.Default?.Value !== undefined && field.Default?.Value !== null) {
    out.defaultValue = field.Default.Value;
  }

  if (kind === "FlagKindStringArray") {
    out.variadic = true;
  }

  return out;
}

// getOperationBodyFieldUsageFlags lists the KDL flags for the operation's
// registered flag metadata (parameters and expanded body fields) — exactly
// what flagutil.RegisterFlags registers, without the body/schema extras.
// Intent commands reuse it: they register the same metadata.
function getOperationBodyFieldUsageFlags(op: Operation): UsageFlagDef[] {
  if (!op.Request) return [];
  const flags: UsageFlagDef[] = [];
  const globalFlags = getGlobalFlagNames();

  const walkFields = (
    fields: FieldDef[],
    flagPrefix: string,
    parentFields?: FieldDef[],
  ) => {
    for (const field of fields) {
      if (field.Const) continue;

      const flagName = flagPrefix
        ? `${flagPrefix}.${sanitizeFlagNameWithReserved(field.Name)}`
        : sanitizeFlagNameWithReserved(field.Name);

      if (!flagPrefix && globalFlags.has(flagName)) continue;

      if (field.Type.Type.toString() === "union") {
        flags.push({
          spec: usageFlagSpec(
            flagName,
            undefined,
            usageFlagKindToArgName("FlagKindJSON", flagName),
          ),
          help: describeTypeForHelp(field.Type),
        });
        // Mirror flagutil.registerUnionFlags: discriminated unions also
        // register per-variant JSON flags and their expandable leaf flags.
        if (isDiscriminatedUnion(field.Type)) {
          const discriminatorKey = getDiscriminatorKey(field.Type);
          for (const variant of getUnionVariants(field.Type)) {
            const variantFlagName = `${flagName}.${variant.flagName}`;
            const variantTypeName =
              variant.type.Name || sanitizeFieldName(variant.name);
            flags.push({
              spec: usageFlagSpec(
                variantFlagName,
                undefined,
                usageFlagKindToArgName("FlagKindJSON", variantFlagName),
              ),
              help: `${variantTypeName} variant as JSON`,
            });
            if (canExpandVariant(variant.type)) {
              for (const variantField of variant.type.Fields || []) {
                if (variantField.Const) continue;
                if (variantField.Name === discriminatorKey) continue;
                flags.push(
                  usageFlagFromField(
                    variantField,
                    `${variantFlagName}.${sanitizeFlagNameWithReserved(
                      variantField.Name,
                    )}`,
                  ),
                );
              }
            }
          }
        }
        continue;
      }

      if (
        field.Nullable &&
        field.Optional &&
        context.Global.Config.NullableOptionalWrapper
      ) {
        flags.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
        continue;
      }

      if (getInputClassType(field) === "MultipartRequestBody") {
        walkMultipart(field.Type.Fields || [], flags);
        continue;
      }

      if (shouldExpandNestedField(field)) {
        const isBodyWrapper = field.Annotations?.Has("request");
        const expandFlagPrefix = isBodyWrapper
          ? getBodyFlagPrefix(field, parentFields || fields, flagPrefix)
          : flagName;
        walkFields(
          field.Type.Fields || [],
          expandFlagPrefix,
          field.Type.Fields || [],
        );
        continue;
      }

      const classType = getInputClassType(field);
      if (classType === "JSONRequestBody" || classType === "FormRequestBody") {
        flags.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
        continue;
      }

      const typeStr = field.Type?.Type?.toString() || "";
      if (
        typeStr === "any" ||
        typeStr === "map" ||
        typeStr === "bigint" ||
        typeStr === "decimal" ||
        typeStr === "class"
      ) {
        flags.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
        continue;
      }

      flags.push(usageFlagFromField(field, flagName));
    }
  };

  const walkMultipart = (fields: FieldDef[], acc: UsageFlagDef[]) => {
    for (const field of fields) {
      if (field.Const) continue;
      const flagName = sanitizeFlagNameWithReserved(field.Name);
      const fieldTypeStr = field.Type?.Type?.toString() || "";
      const isFileArray =
        fieldTypeStr === "array" &&
        field.Type?.ItemType?.Type?.toString() === "class" &&
        (() => {
          const ann = field.Annotations?.Get("multipartForm");
          return Boolean(ann && isMultipartFormAnnotation(ann) && ann.File);
        })();

      if (isMultipartFileField(field) || isFileArray) {
        const isArray = fieldTypeStr === "array";
        acc.push({
          spec: usageFlagSpec(flagName, undefined, "file", isArray),
          help:
            field.Comments?.Description ||
            (isArray
              ? "Comma-separated paths to files to upload"
              : "Path to file to upload"),
          variadic: isArray,
        });
        continue;
      }

      if (isMultipartJSONField(field)) {
        acc.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
        continue;
      }

      const ft = field.Type?.Type?.toString() || "";
      if (
        ft === "any" ||
        ft === "map" ||
        ft === "bigint" ||
        ft === "decimal" ||
        ft === "union"
      ) {
        acc.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
        continue;
      }
      if (ft === "array") {
        const itemType = field.Type?.ItemType?.Type?.toString() || "";
        if (
          itemType === "class" ||
          itemType === "map" ||
          itemType === "union"
        ) {
          acc.push(usageFlagFromField(field, flagName, "FlagKindJSON"));
          continue;
        }
      }
      acc.push(usageFlagFromField(field, flagName));
    }
  };

  if (op.Request.IsRequestBody) {
    if (op.SerializationMethod?.toString() === "multipart") {
      walkMultipart(op.Request.RequestBody.Type.Fields || [], flags);
    } else if (isRequestBodyExpandable(op)) {
      walkFields(op.Request.RequestBody.Type.Fields || [], "");
    }
  } else {
    walkFields(op.Request.Field.Type.Fields || [], "");
  }

  return flags;
}

// Metadata retained by a route-dispatch intent: path/query/header parameters
// only. All request-body metadata is deliberately replaced by the declared
// variant flags plus the generic --body escape.
function getOperationNonBodyUsageFlags(op: Operation): UsageFlagDef[] {
  if (!op.Request?.Params) return [];
  const flags: UsageFlagDef[] = [];
  const globalFlags = getGlobalFlagNames();
  const params = [
    ...(op.Request.Params.PathParams || []),
    ...(op.Request.Params.QueryParams || []),
    ...(op.Request.Params.HeaderParams || []),
  ];
  for (const param of params) {
    const field = param.Field;
    if (field.Const) continue;
    const name = sanitizeFlagNameWithReserved(field.Name);
    if (globalFlags.has(name)) continue;
    flags.push(usageFlagFromField(field, name));
  }
  return flags;
}

function getOperationFieldUsageFlags(op: Operation): UsageFlagDef[] {
  const flags = getOperationBodyFieldUsageFlags(op);

  if (templateHasBodyFlag(op)) {
    flags.push({
      spec: usageFlagSpec("body", undefined, "body"),
      help: templateBodyFlagDescription(op),
    });
  }

  if (hasBodySchemaForOp(op)) {
    flags.push({
      spec: usageFlagSpec("schema"),
      help: "Print the exact JSON Schema of the request body and exit",
    });
  }

  return flags;
}

// getOperationSecurityUsageFlags lists the operation-level security flags —
// shared by operation commands and the intent commands that reuse their run
// functions (which read security from the invoking command's flags).
function getOperationSecurityUsageFlags(op: Operation): UsageFlagDef[] {
  const flags: UsageFlagDef[] = [];
  const secFlags = new Set<string>();
  if (op.Security?.Type?.Fields) {
    walkSecurityLeafFields(op.Security.Type.Fields || [], (leaf) => {
      const flagName = sanitizeFlagNameWithReserved(leaf.name);
      if (secFlags.has(flagName)) return;
      secFlags.add(flagName);
      flags.push({
        spec: usageFlagSpec(flagName, undefined, leaf.name.replace(/-/g, "_")),
        help: leaf.description,
      });
    });
  }
  return flags;
}

function getOperationUsageFlags(op: Operation): UsageFlagDef[] {
  const flags: UsageFlagDef[] = [];

  for (const f of getOperationFieldUsageFlags(op)) {
    flags.push(f);
  }

  for (const f of getOperationSecurityUsageFlags(op)) {
    flags.push(f);
  }

  const declared = getCLIOperationCtx(op);
  for (const f of declared?.Flags || []) {
    const usage: UsageFlagDef = {
      spec: usageFlagSpec(
        f.Name,
        f.Shorthand || undefined,
        f.Type === "bool" ? undefined : f.Name.replace(/-/g, "_"),
      ),
      help: f.Summary,
    };
    if (f.HasDefault) usage.defaultValue = f.DefaultValue;
    flags.push(usage);
  }

  if (hasPagination(op)) {
    flags.push({
      spec: usageFlagSpec("all", "a"),
      help: "Automatically paginate and fetch all results (streams NDJSON for JSON output)",
    });
    flags.push({
      spec: usageFlagSpec("max-pages", undefined, "max_pages"),
      help: "Maximum number of pages to fetch when using --all (0 = no limit)",
      defaultValue: 0,
    });
  }

  if (hasBinaryResponse(op)) {
    flags.push({
      spec: usageFlagSpec("output-file", undefined, "output_file"),
      help: "Save the response body to a file path (recommended for binary/file responses)",
    });
    flags.push({
      spec: usageFlagSpec("output-b64"),
      help: "Encode binary response as base64 and print to stdout",
    });
  }

  return flags;
}

function getRootUsageFlags(): UsageFlagDef[] {
  const flags: UsageFlagDef[] = [
    {
      spec: usageFlagSpec("usage"),
      help: "Print the CLI Usage schema in KDL format",
      global: true,
    },
  ];

  // --help-global is deliberately root-local: it documents the inherited
  // runtime flags once without implying that subcommands accept it.
  if (helpStyle() === "compact") {
    flags.push({
      spec: usageFlagSpec("help-global"),
      help: "Print global flags shared by every command",
    });
  }

  flags.push(
    {
      spec: usageFlagSpec("output-format", "o", "format"),
      help: "Specify the output format. Options: pretty, json, yaml, table, toon.",
      global: true,
      defaultValue: "pretty",
      config: "output_format",
    },
    {
      spec: usageFlagSpec("color", undefined, "color"),
      help: "Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars.",
      global: true,
      defaultValue: defaultColorMode(),
    },
    {
      spec: usageFlagSpec("jq", "q", "jq"),
      help: "Filter and transform output using a jq expression (e.g., '.name', '.items[] | .id')",
      global: true,
    },
    {
      spec: usageFlagSpec("raw-output"),
      help: "Write --jq string results as raw text instead of JSON strings (like jq -r); non-string results stay JSON",
      global: true,
      defaultValue: jqRawOutputDefault(),
    },
    {
      spec: usageFlagSpec("server-url", undefined, "url"),
      help: "Override the default server URL",
      global: true,
    },
  );

  // Server selection mirrors runtime registration: "none" omits the flag
  // entirely, while "hidden" keeps it (it still parses) marked hide.
  const serverVisibility = serverSelectionFlag();
  if (serverVisibility !== "none") {
    flags.push({
      spec: usageFlagSpec("server", undefined, "server"),
      help: "Select a server by index (for indexed servers) or name (for named servers)",
      global: true,
      hidden: serverVisibility === "hidden",
    });
  }

  flags.push(
    {
      spec: usageFlagSpec("header", "H", "header", true),
      help: 'Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.',
      global: true,
      variadic: true,
    },
    {
      spec: usageFlagSpec("include-headers"),
      help: "Include HTTP response headers in the output",
      global: true,
      defaultValue: false,
    },
    {
      spec: usageFlagSpec("timeout", undefined, "duration"),
      help: "HTTP request timeout (e.g., 30s, 5m, 100ms)",
      global: true,
      config: "timeout",
    },
    ...(isAnyInteractiveEnabled()
      ? [
          {
            spec: usageFlagSpec("interactive"),
            help: interactiveFlagHelp(),
            global: true,
            defaultValue: interactiveByDefault(),
          },
        ]
      : []),
    ...(isAnyInteractiveEnabled()
      ? [
          {
            spec: usageFlagSpec("no-interactive"),
            help: "Disable all interactive features (auto-prompting, explorer auto-launch, TUI forms)",
            global: true,
            defaultValue: false,
          },
        ]
      : []),
    {
      spec: usageFlagSpec("dry-run"),
      help: dryRunFlagHelp(),
      global: true,
      defaultValue: false,
    },
    {
      spec: usageFlagSpec("debug", "d"),
      help: "Log request and response diagnostics to stderr",
      global: true,
      defaultValue: false,
    },
    {
      spec: usageFlagSpec("agent-mode"),
      help: agentEnvironmentDetection()
        ? "Enable structured errors and default TOON output for AI coding agents. Automatically enabled when a known agent environment is detected (CLAUDECODE, CURSOR_AGENT, etc.). Use --agent-mode=false to disable."
        : "Enable structured errors and default TOON output for AI coding agents.",
      global: true,
      defaultValue: false,
    },
  );

  // Retry flags mirror runtime registration: "none" omits them entirely,
  // "hidden" keeps them (they still parse) marked hide.
  const retryVisibility = retryFlagsVisibility();
  if (isFeatureUsed("retries") && retryVisibility !== "none") {
    const retryHidden = retryVisibility === "hidden";
    const retryConfigHelp = hasAttemptCountRetries()
      ? 'Full retry config as JSON. Schema: {"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":10000,"exponent":1.5,"maxElapsedTime":30000},"retryConnectionErrors":false}. Use strategy "attempt-count-backoff" with maxRetries for attempt-count retries. Times are in milliseconds.'
      : 'Full retry config as JSON. Schema: {"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":10000,"exponent":1.5,"maxElapsedTime":30000},"retryConnectionErrors":false}. Times are in milliseconds.';
    flags.push(
      {
        spec: usageFlagSpec("no-retries"),
        help: "Disable automatic retries (default: retries enabled with exponential backoff)",
        global: true,
        defaultValue: false,
        config: "no_retries",
        hidden: retryHidden,
      },
      {
        spec: usageFlagSpec("retry-max-elapsed-time", undefined, "duration"),
        help: "Maximum total time for retries (e.g., 30s, 5m). Default: 30s",
        global: true,
        config: "retry_max_elapsed_time",
        hidden: retryHidden,
      },
      {
        spec: usageFlagSpec("retry-connection-errors"),
        help: "Retry on connection errors (EOF, reset, etc.)",
        global: true,
        defaultValue: false,
        config: "retry_connection_errors",
        hidden: retryHidden,
      },
      {
        spec: usageFlagSpec("retry-config", undefined, "json"),
        help: retryConfigHelp,
        global: true,
        config: "retry_config",
        hidden: retryHidden,
      },
    );
  }

  const servers = context.Global.AST.MainSDK.Servers;
  if (servers) {
    for (const v of servers.GetVariables()) {
      flags.push({
        spec: usageFlagSpec(
          sanitizeFlagName(v.Name),
          undefined,
          sanitizeFlagName(v.Name).replace(/-/g, "_"),
        ),
        help: `Server template variable: ${v.Name}`,
        global: true,
      });
    }
  }

  for (const field of getCLISecurityFields()) {
    const env = context.Global.Config.EnvVarPrefix
      ? `${context.Global.Config.EnvVarPrefix.toUpperCase()}_${
          field.envVarSuffix
        }`
      : field.envVarSuffix;
    flags.push({
      spec: usageFlagSpec(
        field.flagName,
        undefined,
        field.flagName.replace(/-/g, "_"),
        field.isArray,
      ),
      help: field.description,
      global: true,
      env,
      config: `security.${field.configKey}`,
      variadic: field.isArray,
    });
  }

  if (hasGlobals()) {
    for (const field of context.Global.AST.MainSDK.Globals.Fields) {
      const flagName = sanitizeFlagNameWithReserved(field.Name);
      const argName = usageFlagKindToArgName(
        inferKindNameForField(field),
        flagName,
      );
      flags.push({
        spec: usageFlagSpec(flagName, undefined, argName),
        help:
          field.Comments?.Summary?.trim() ||
          field.Comments?.Description?.trim()?.split(/[.\n]/)[0]?.trim() ||
          `Global ${sanitizeFlagName(field.Name)} parameter`,
        global: true,
        env: templateGlobalEnvVars(field),
        config: `globals.${toConfigKey(field.Name)}`,
        defaultValue:
          field.Default?.Value !== undefined && field.Default?.Value !== null
            ? field.Default.Value
            : undefined,
      });
    }
  }

  return flags;
}

function getOperationCommandAliases(op: Operation): string[] {
  const group = op.OwningSDK?.Group;
  const siblingNames: string[] = [];
  if (group) {
    const owning = op.OwningSDK;
    if (owning) {
      const owningGroupName = getSDKGroupName(owning);
      for (const childSDK of owning.SubSDKs || []) {
        siblingNames.push(
          getDeStutteredCommandName(owningGroupName, getSDKGroupName(childSDK)),
        );
      }
      for (const candidate of owning.Operations || []) {
        const stutterKind = hasNameOverride(candidate)
          ? "none"
          : getStutterKind(owningGroupName, candidate.GetID());
        if (stutterKind !== "exact") {
          siblingNames.push(
            stutterKind === "prefix" || stutterKind === "suffix"
              ? getDeStutteredCommandName(owningGroupName, candidate.GetID())
              : sanitizeCLICommand(candidate.GetID()),
          );
        }
      }
    }
  } else {
    const sdk = context.Global.AST.MainSDK;
    for (const subSDK of sdk.SubSDKs)
      siblingNames.push(sanitizeCLICommand(getSDKGroupName(subSDK)));
    for (const candidate of sdk.Operations) {
      const stutterKind = getStutterKind("", candidate.GetID());
      if (stutterKind !== "exact") {
        siblingNames.push(sanitizeCLICommand(candidate.GetID()));
      }
    }
  }

  const aliasMap = computeCommandAliases(siblingNames);
  const cmdPath = getCLICommandPath(op);
  const name = cmdPath[cmdPath.length - 1] || sanitizeCLICommand(op.GetID());
  return aliasMap.get(name) || [];
}

function buildOperationUsageCommand(op: Operation): UsageCommandDef | null {
  if (isOperationOverridden(op)) return null;
  const path = getCLICommandPath(op);
  if (path.length === 0) return null;
  return {
    name: path[path.length - 1],
    help: usageOperationHelp(op),
    aliases: getOperationCommandAliases(op),
    flags: getOperationUsageFlags(op),
    commands: [],
  };
}

function getSiblingCommandNamesForSDK(
  sdk: SDK,
  parentGroupName: string,
): string[] {
  const siblingNames: string[] = [];
  for (const childSDK of sdk.SubSDKs || []) {
    siblingNames.push(
      getDeStutteredCommandName(parentGroupName, getSDKGroupName(childSDK)),
    );
  }
  // Overridden operations stay in the sibling set: runtime alias allocation
  // (main.ts) counts every generated operation, and the usage tree must hand
  // out the same aliases. Only their command nodes are omitted.
  for (const op of sdk.Operations || []) {
    const stutterKind = hasNameOverride(op)
      ? "none"
      : getStutterKind(parentGroupName, op.GetID());
    if (stutterKind !== "exact") {
      siblingNames.push(
        stutterKind === "prefix" || stutterKind === "suffix"
          ? getDeStutteredCommandName(parentGroupName, op.GetID())
          : sanitizeCLICommand(op.GetID()),
      );
    }
  }
  return siblingNames;
}

function getGroupAliases(
  parentSDK: SDK,
  parentGroupName: string,
): Map<string, string[]> {
  return computeCommandAliases(
    getSiblingCommandNamesForSDK(parentSDK, parentGroupName),
  );
}

function buildSubSDKUsageCommand(
  sdk: SDK,
  parentGroupName: string = "",
  parentSDK: SDK = context.Global.AST.MainSDK,
): UsageCommandDef {
  const aliasMap = getGroupAliases(parentSDK, parentGroupName);
  const sdkGroupName = getSDKGroupName(sdk);
  const cmdName = getDeStutteredCommandName(parentGroupName, sdkGroupName);
  const cmd: UsageCommandDef = {
    name: cmdName,
    help: usageGroupHelp(sdk),
    aliases: aliasMap.get(cmdName) || [],
    flags: [],
    commands: [],
  };

  for (const op of sdk.Operations) {
    if (isOperationOverridden(op)) continue;
    if (
      !hasNameOverride(op) &&
      getStutterKind(sdkGroupName, op.GetID()) === "exact"
    ) {
      cmd.help = usageOperationHelp(op);
      cmd.flags = getOperationUsageFlags(op);
    } else {
      const child = buildOperationUsageCommand(op);
      if (child) cmd.commands!.push(child);
    }
  }

  for (const sub of sdk.SubSDKs) {
    cmd.commands!.push(buildSubSDKUsageCommand(sub, sdkGroupName, sdk));
  }

  return cmd;
}

// buildIntentUsageCommand mirrors the runtime flag surface an intent command
// registers: the backing operation's flag metadata (when present), the
// operation's security flags, the body flag it reads, --schema, and the
// declared intent and artifact output flags — in registration order.
function buildIntentUsageCommand(intent: IntentCmdCtx): UsageCommandDef {
  const flags: UsageFlagDef[] = [];
  const found = findIntentOperation(intent.OpID);

  if (intent.HasMeta && found) {
    flags.push(
      ...(intent.Dispatch
        ? getOperationNonBodyUsageFlags(found.op)
        : getOperationBodyFieldUsageFlags(found.op)),
    );
  }
  if (found && intent.SecurityFlags) {
    flags.push(...getOperationSecurityUsageFlags(found.op));
  }
  if (intent.BodyFlag) {
    flags.push({
      spec: usageFlagSpec(intent.BodyFlag, undefined, "body"),
      help: `Request body as JSON (advanced; replaces intent arguments). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF.${
        intent.HasSchema ? " Use --schema to print the exact JSON Schema." : ""
      }`,
    });
  }
  if (intent.HasSchema) {
    flags.push({
      spec: usageFlagSpec("schema"),
      help: "Print the exact JSON Schema of the request body and exit",
    });
  }
  for (const f of intent.Flags) {
    const argName =
      f.Type === "bool" ? undefined : (f.Name || "value").replace(/-/g, "_");
    flags.push({
      spec: usageFlagSpec(f.Name || "", f.Shorthand || undefined, argName),
      help: f.Summary,
    });
  }
  if (intent.ArtifactJSON) {
    flags.push({
      spec: usageFlagSpec("out", undefined, "path"),
      help: `Write the ${intent.ArtifactKind} to this file (or into this directory). Default: ./${intent.ArtifactDefaultPath}`,
    });
    flags.push({
      spec: usageFlagSpec("raw-response"),
      help: `Print the raw API response instead of writing the ${intent.ArtifactKind} to a file`,
    });
  }
  if (intent.AsyncJSON) {
    flags.push({
      spec: usageFlagSpec("async"),
      help: "Return the operation handle without waiting for a terminal response",
    });
    flags.push({
      spec: usageFlagSpec("poll-interval", undefined, "duration"),
      help: "Override the initial polling interval (positive Go duration, for example 500ms or 2s)",
    });
    flags.push({
      spec: usageFlagSpec("poll-timeout", undefined, "duration"),
      help: "Override the overall polling deadline (positive Go duration, at least the effective poll interval)",
    });
  }

  return {
    name: firstUseWord(intent.Use),
    help: intent.Summary,
    args: intent.Args.map((arg) => ({
      name: arg.Name || "input",
      help: arg.Summary,
      required: Boolean(arg.Required && !arg.PresetCovered),
      variadic: Boolean(arg.Variadic),
    })),
    flags,
    commands: [],
  };
}

// findUsageCommandByPath walks the usage tree by successive child names
// (the KDL analogue of the runtime findCommandByPath).
function findUsageCommandByPath(
  root: UsageCommandDef,
  path: string[],
): UsageCommandDef | null {
  let current: UsageCommandDef | undefined = root;
  for (const segment of path) {
    current = (current?.commands || []).find((c) => c.name === segment);
    if (!current) return null;
  }
  return current || null;
}

// usageCommandOwnsName mirrors the runtime findOwningCommand: a command owns
// a name via its primary name or any alias (both participate in dispatch).
function usageCommandOwnsName(
  commands: UsageCommandDef[],
  name: string,
): boolean {
  return commands.some(
    (c) => c.name === name || (c.aliases || []).includes(name),
  );
}

// insertIntentUsageCommands mounts declared commands (x-speakeasy-cli-commands)
// into the usage tree, mirroring runtime mounting: bound intents and planned
// placeholders attach under their ParentPath, in manifest order (so declared
// parents exist before their declared children). Planned placeholders yield
// when a real command already owns the name (by name or alias) — exactly
// like the runtime.
function insertIntentUsageCommands(
  root: UsageCommandDef,
  manifest: IntentManifestCtx,
) {
  const boundByPath = new Map<string, IntentCmdCtx>();
  for (const b of manifest.Bound) {
    boundByPath.set([...b.ParentPath, firstUseWord(b.Use)].join(" "), b);
  }
  const plannedByPath = new Map<string, PlannedCmdCtx>();
  for (const p of manifest.Planned) {
    plannedByPath.set([...p.ParentPath, p.Name].join(" "), p);
  }

  for (const entry of manifest.Order) {
    const key = [...entry.ParentPath, entry.Name].join(" ");
    const parent = findUsageCommandByPath(root, entry.ParentPath);
    if (!parent) continue; // runtime init fails loudly on missing parents

    const bound = boundByPath.get(key);
    if (bound) {
      parent.commands = parent.commands || [];
      if (usageCommandOwnsName(parent.commands, entry.Name)) {
        throw new Error(
          `x-speakeasy-cli-commands: command "${key}" collides with an existing usage command; use override: true only to replace the exact generated operation command`,
        );
      }
      parent.commands.push(buildIntentUsageCommand(bound));
      continue;
    }

    const planned = plannedByPath.get(key);
    if (planned) {
      parent.commands = parent.commands || [];
      if (usageCommandOwnsName(parent.commands, planned.Name)) continue;
      parent.commands.push({
        name: planned.Name,
        help: `${planned.Summary}${
          planned.Tagline ? ` ${planned.Tagline}` : ""
        }`.trim(),
        commands: [],
      });
    }
    // Group tags categorize existing commands; no usage entry of their own.
  }
}

// applyDeclaredUsageOrder mirrors the runtime applyDeclaredCommandOrder pass:
// parents that host declared commands render them first in manifest order and
// sort the rest alphabetically; every other parent is entirely alphabetical.
// This keeps the full KDL tree in the same order as runtime help.
function applyDeclaredUsageOrder(
  root: UsageCommandDef,
  manifest: IntentManifestCtx,
) {
  const namesByParent = new Map<UsageCommandDef, string[]>();
  for (const entry of manifest.Order) {
    // Runtime falls back to the root when the parent path does not resolve.
    const parent = findUsageCommandByPath(root, entry.ParentPath) || root;
    const names = namesByParent.get(parent) || [];
    names.push(entry.Name);
    namesByParent.set(parent, names);
  }

  const reorder = (parent: UsageCommandDef) => {
    const children = parent.commands || [];
    const taken = new Set<UsageCommandDef>();
    const reordered: UsageCommandDef[] = [];
    for (const name of namesByParent.get(parent) || []) {
      const child = children.find((c) => c.name === name && !taken.has(c));
      if (child) {
        taken.add(child);
        reordered.push(child);
      }
    }
    const rest = children
      .filter((c) => !taken.has(c))
      .sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0));
    parent.commands = [...reordered, ...rest];
    for (const child of parent.commands) reorder(child);
  };
  reorder(root);
}

function buildRootUsageCommand(): UsageCommandDef {
  const intentManifest = collectIntentManifest();
  const root: UsageCommandDef = {
    name: sanitizeCliName(),
    help: usageRootHelp(),
    flags: getRootUsageFlags(),
    commands: [],
  };

  for (const op of context.Global.AST.MainSDK.Operations) {
    const child = buildOperationUsageCommand(op);
    if (child) root.commands!.push(child);
  }
  for (const sdk of context.Global.AST.MainSDK.SubSDKs) {
    root.commands!.push(
      buildSubSDKUsageCommand(sdk, "", context.Global.AST.MainSDK),
    );
  }

  root.commands!.push({
    name: "configure",
    help: templateConfigureShort(),
    commands: [],
  });

  if (hasConfigurableSettings()) {
    root.commands!.push({
      name: "whoami",
      help: templateWhoamiShort(),
      commands: [],
    });
  }

  root.commands!.push({
    name: "version",
    help: "Print the CLI version",
    commands: [],
  });

  // Built-in auth group (auth.go.stmpl). At runtime the credential
  // subcommands merge into a spec-owned "auth" group (a spec operation
  // keeps any name it already claims); mirror exactly that here.
  if (isInteractiveAuthEnabled() && hasGlobalSecurity()) {
    const credentialSubcommands = [
      {
        name: "login",
        help: templateAuthLoginShort(),
        commands: [],
      },
      {
        name: "whoami",
        help: templateWhoamiShort(),
        commands: [],
      },
      {
        name: "logout",
        help: "Clear all stored authentication credentials",
        commands: [],
      },
    ];
    const owner = root.commands!.find(
      (c) => c.name === "auth" || (c.aliases || []).includes("auth"),
    );
    if (owner) {
      const taken = new Set<string>();
      for (const child of owner.commands || []) {
        taken.add(child.name);
        for (const alias of child.aliases || []) {
          taken.add(alias);
        }
      }
      owner.commands = owner.commands || [];
      for (const sub of credentialSubcommands) {
        if (!taken.has(sub.name)) {
          owner.commands.push(sub);
        }
      }
    } else {
      root.commands!.push({
        name: "auth",
        help: "Manage authentication credentials",
        commands: credentialSubcommands,
      });
    }
  }

  if (
    isInteractiveModeEnabled() &&
    !usageCommandOwnsName(root.commands!, "explore")
  ) {
    root.commands!.push({
      name: "explore",
      help: "Interactively browse and run commands",
      commands: [],
    });
  }

  // Catalog commands (x-speakeasy-cli-catalog, catalog.go.stmpl). A catalog
  // whose exact name is a generated group becomes that group's bare
  // invocation and lends it its summary; one that only matches an alias is
  // unreachable (cobra dispatches to the earlier-registered group) and is
  // skipped; otherwise it is a new leaf command.
  for (const catalog of collectCliCatalogs()) {
    const existing = root.commands!.find((c) => c.name === catalog.Command);
    if (existing) {
      existing.help = catalog.Summary;
      continue;
    }
    if (usageCommandOwnsName(root.commands!, catalog.Command)) continue;
    root.commands!.push({
      name: catalog.Command,
      help: catalog.Summary,
      commands: [],
    });
  }

  insertIntentUsageCommands(root, intentManifest);

  // Cobra's default help and completion commands (root.go.stmpl creates them
  // before usage.Intercept so they honor --usage too). Cobra skips each when
  // an existing command already owns the name or alias. Add them before the
  // final order pass so KDL ordering matches the assembled runtime tree.
  if (!usageCommandOwnsName(root.commands!, "help")) {
    root.commands!.push({
      name: "help",
      help: "Help about any command",
      commands: [],
    });
  }
  if (!usageCommandOwnsName(root.commands!, "completion")) {
    root.commands!.push({
      name: "completion",
      help: "Generate the autocompletion script for the specified shell",
      commands: ["bash", "zsh", "fish", "powershell"].map((shell) => ({
        name: shell,
        help: `Generate the autocompletion script for ${shell}`,
        commands: [],
      })),
    });
  }

  applyDeclaredUsageOrder(root, intentManifest);

  return root;
}

function renderUsageFlag(flag: UsageFlagDef, indent: number): string[] {
  const prefix = "  ".repeat(indent + 1);
  const props: string[] = [];
  pushIf(props, flag.help ? kdlPropAt("help", flag.help, prefix) : "");
  pushIf(
    props,
    flag.longHelp ? kdlPropAt("long_help", flag.longHelp, prefix) : "",
  );
  pushIf(props, flag.global ? kdlProp("global", true) : "");
  pushIf(props, flag.env ? kdlProp("env", flag.env) : "");
  pushIf(props, flag.config ? kdlProp("config", flag.config) : "");
  if (flag.defaultValue !== undefined) {
    pushIf(props, kdlPropAt("default", flag.defaultValue, prefix));
  }
  pushIf(props, flag.count ? kdlProp("count", true) : "");
  pushIf(props, flag.variadic ? kdlProp("var", true) : "");
  pushIf(props, flag.hidden ? kdlProp("hide", true) : "");
  return [
    `${"  ".repeat(indent)}flag ${kdlQuoted(flag.spec)}${
      props.length ? ` ${props.join(" ")}` : ""
    }`,
  ];
}

function renderUsageCommand(cmd: UsageCommandDef, indent: number): string[] {
  const lines: string[] = [];
  const childLines: string[] = [];
  for (const alias of cmd.aliases || []) {
    childLines.push(`${"  ".repeat(indent + 1)}alias ${kdlQuoted(alias)}`);
  }
  for (const arg of cmd.args || []) {
    const props: string[] = [];
    pushIf(
      props,
      arg.help ? kdlPropAt("help", arg.help, "  ".repeat(indent + 2)) : "",
    );
    pushIf(props, arg.required ? kdlProp("required", true) : "");
    pushIf(props, arg.variadic ? kdlProp("var", true) : "");
    childLines.push(
      `${"  ".repeat(indent + 1)}arg ${kdlQuoted(arg.name)}${
        props.length ? ` ${props.join(" ")}` : ""
      }`,
    );
  }
  for (const flag of cmd.flags || []) {
    childLines.push(...renderUsageFlag(flag, indent + 1));
  }
  for (const child of cmd.commands || []) {
    childLines.push(...renderUsageCommand(child, indent + 1));
  }

  const props: string[] = [];
  pushIf(
    props,
    cmd.help ? kdlPropAt("help", cmd.help, "  ".repeat(indent + 1)) : "",
  );
  if (childLines.length > 0) {
    lines.push(
      `${"  ".repeat(indent)}cmd ${kdlQuoted(cmd.name)}${
        props.length ? ` ${props.join(" ")}` : ""
      } {`,
    );
    lines.push(...childLines);
    lines.push(`${"  ".repeat(indent)}}`);
    return lines;
  }

  lines.push(
    `${"  ".repeat(indent)}cmd ${kdlQuoted(cmd.name)}${
      props.length ? ` ${props.join(" ")}` : ""
    }`,
  );
  return lines;
}

function collectCommandPathSchemas(
  cmd: UsageCommandDef,
  path: string[],
  out: Map<string, string>,
  includeRootMetadata: boolean,
) {
  const lines: string[] = [];
  if (includeRootMetadata) {
    lines.push(`name ${kdlQuoted(getHumanCliName())}`);
    lines.push(`bin ${kdlQuoted(sanitizeCliName())}`);
    lines.push(`about ${kdlStringAt(usageRootAbout(), "  ")}`);
    lines.push(`version ${kdlQuoted(context.Global.Config.SDKVersion)}`);
    lines.push(`config {`);
    lines.push(
      `  file ${kdlQuoted(`~/.config/${sanitizeCliName()}/config.yaml`)}`,
    );
    lines.push(`}`);
    for (const flag of cmd.flags || []) {
      lines.push(...renderUsageFlag(flag, 0));
    }
    for (const child of cmd.commands || []) {
      lines.push(...renderUsageCommand(child, 0));
    }
  } else {
    lines.push(...renderUsageCommand(cmd, 0));
  }
  const schema = lines.join("\n") + "\n";
  out.set(path.join(" "), schema);
  for (const alias of cmd.aliases || []) {
    out.set([...path.slice(0, -1), alias].join(" "), schema);
  }
  for (const child of cmd.commands || []) {
    collectCommandPathSchemas(child, [...path, child.name], out, false);
  }
}

function goTemplateSafeStringExpr(s: string): string {
  const parts = s.split(/(\{\{|\}\})/g).filter((part) => part.length > 0);
  if (parts.length === 0) {
    return goStringLiteral("");
  }

  return parts
    .map((part) => {
      if (part === "{{") {
        return "string([]byte{'{', '{'})";
      }
      if (part === "}}") {
        return "string([]byte{'}', '}'})";
      }
      return goStringLiteral(part);
    })
    .join(" + ");
}

interface UsageAliasPathTestCase {
  path: string[];
  canonicalName: string;
  alias: string;
}

interface UsageCommandPathTestCase {
  path: string[];
  commandName: string;
}

function findUsageAliasPathTestCase(
  cmd: UsageCommandDef,
  path: string[] = [],
  includeCurrentInPath: boolean = true,
): UsageAliasPathTestCase | null {
  const nextPath = includeCurrentInPath ? [...path, cmd.name] : path;
  for (const alias of cmd.aliases || []) {
    return {
      path: [...path, alias],
      canonicalName: cmd.name,
      alias,
    };
  }
  for (const child of cmd.commands || []) {
    const found = findUsageAliasPathTestCase(child, nextPath, true);
    if (found) return found;
  }
  return null;
}

function getUsageAliasPathTestCase(): UsageAliasPathTestCase | null {
  const root = buildRootUsageCommand();
  return findUsageAliasPathTestCase(root, [], false);
}

function findUsageSubgroupPathTestCase(
  cmd: UsageCommandDef,
  path: string[] = [],
  includeCurrentInPath: boolean = true,
): UsageCommandPathTestCase | null {
  const nextPath = includeCurrentInPath ? [...path, cmd.name] : path;
  if ((cmd.commands || []).length > 0 && path.length > 0) {
    return {
      path: nextPath,
      commandName: cmd.name,
    };
  }
  for (const child of cmd.commands || []) {
    const found = findUsageSubgroupPathTestCase(child, nextPath, true);
    if (found) return found;
  }
  return null;
}

function getUsageSubgroupPathTestCase(): UsageCommandPathTestCase | null {
  const root = buildRootUsageCommand();
  return findUsageSubgroupPathTestCase(root, [], false);
}

function templateUsageAliasPathTestEnabled(): boolean {
  return getUsageAliasPathTestCase() !== null;
}
registerTemplateFunc(
  "templateUsageAliasPathTestEnabled",
  templateUsageAliasPathTestEnabled,
);

function templateUsageAliasPathArgs(): string {
  const testCase = getUsageAliasPathTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc("templateUsageAliasPathArgs", templateUsageAliasPathArgs);

function templateUsageAliasPathCanonicalName(): string {
  const testCase = getUsageAliasPathTestCase();
  return testCase?.canonicalName || "";
}
registerTemplateFunc(
  "templateUsageAliasPathCanonicalName",
  templateUsageAliasPathCanonicalName,
);

function templateUsageAliasPathAlias(): string {
  const testCase = getUsageAliasPathTestCase();
  return testCase?.alias || "";
}
registerTemplateFunc(
  "templateUsageAliasPathAlias",
  templateUsageAliasPathAlias,
);

function templateUsageSubgroupPathTestEnabled(): boolean {
  return getUsageSubgroupPathTestCase() !== null;
}
registerTemplateFunc(
  "templateUsageSubgroupPathTestEnabled",
  templateUsageSubgroupPathTestEnabled,
);

function templateUsageSubgroupPathArgs(): string {
  const testCase = getUsageSubgroupPathTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templateUsageSubgroupPathArgs",
  templateUsageSubgroupPathArgs,
);

function templateUsageSubgroupPathCommandName(): string {
  const testCase = getUsageSubgroupPathTestCase();
  return testCase?.commandName || "";
}
registerTemplateFunc(
  "templateUsageSubgroupPathCommandName",
  templateUsageSubgroupPathCommandName,
);

// =============================================================================
// Discriminated-union usage test case (drives usage_test.go assertions)
// =============================================================================

interface DiscriminatedUnionUsageTestCase {
  path: string[];
  unionFlag: string;
  firstVariantFlag: string;
}

function getDiscriminatedUnionUsageTestCase(): DiscriminatedUnionUsageTestCase | null {
  const findInOperation = (
    op: Operation,
  ): DiscriminatedUnionUsageTestCase | null => {
    if (!op.Request) return null;
    const globalFlags = getGlobalFlagNames();
    const walkFields = (
      fields: FieldDef[],
      flagPrefix: string,
      parentFields?: FieldDef[],
    ): { unionFlag: string; firstVariantFlag: string } | null => {
      for (const field of fields) {
        if (field.Const) continue;
        const flagName = flagPrefix
          ? `${flagPrefix}.${sanitizeFlagNameWithReserved(field.Name)}`
          : sanitizeFlagNameWithReserved(field.Name);
        if (!flagPrefix && globalFlags.has(flagName)) continue;

        if (field.Type.Type.toString() === "union") {
          if (!isDiscriminatedUnion(field.Type)) continue;
          const firstVariant = getUnionVariants(field.Type)[0];
          if (!firstVariant) continue;
          return {
            unionFlag: flagName,
            firstVariantFlag: `${flagName}.${firstVariant.flagName}`,
          };
        }
        if (
          field.Nullable &&
          field.Optional &&
          context.Global.Config.NullableOptionalWrapper
        ) {
          continue;
        }
        if (getInputClassType(field) === "MultipartRequestBody") continue;
        if (shouldExpandNestedField(field)) {
          const isBodyWrapper = field.Annotations?.Has("request");
          const expandFlagPrefix = isBodyWrapper
            ? getBodyFlagPrefix(field, parentFields || fields, flagPrefix)
            : flagName;
          const found = walkFields(
            field.Type.Fields || [],
            expandFlagPrefix,
            field.Type.Fields || [],
          );
          if (found) return found;
        }
      }
      return null;
    };

    let found: { unionFlag: string; firstVariantFlag: string } | null = null;
    if (op.Request.IsRequestBody) {
      if (
        op.SerializationMethod?.toString() !== "multipart" &&
        isRequestBodyExpandable(op)
      ) {
        found = walkFields(op.Request.RequestBody.Type.Fields || [], "");
      }
    } else {
      found = walkFields(op.Request.Field.Type.Fields || [], "");
    }
    return found ? { path: getCLICommandPath(op), ...found } : null;
  };

  const walkSDK = (sdk: SDK): DiscriminatedUnionUsageTestCase | null => {
    for (const op of sdk.Operations || []) {
      const found = findInOperation(op);
      if (found) return found;
    }
    for (const child of sdk.SubSDKs || []) {
      const found = walkSDK(child);
      if (found) return found;
    }
    return null;
  };
  return walkSDK(context.Global.AST.MainSDK);
}

function templateDiscriminatedUnionUsageTestEnabled(): boolean {
  return getDiscriminatedUnionUsageTestCase() !== null;
}
registerTemplateFunc(
  "templateDiscriminatedUnionUsageTestEnabled",
  templateDiscriminatedUnionUsageTestEnabled,
);

function templateDiscriminatedUnionUsagePathArgs(): string {
  const testCase = getDiscriminatedUnionUsageTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templateDiscriminatedUnionUsagePathArgs",
  templateDiscriminatedUnionUsagePathArgs,
);

function templateDiscriminatedUnionUsageFlag(): string {
  return getDiscriminatedUnionUsageTestCase()?.unionFlag || "";
}
registerTemplateFunc(
  "templateDiscriminatedUnionUsageFlag",
  templateDiscriminatedUnionUsageFlag,
);

function templateDiscriminatedUnionUsageFirstVariantFlag(): string {
  return getDiscriminatedUnionUsageTestCase()?.firstVariantFlag || "";
}
registerTemplateFunc(
  "templateDiscriminatedUnionUsageFirstVariantFlag",
  templateDiscriminatedUnionUsageFirstVariantFlag,
);

// =============================================================================
// Intent usage test cases (drive the generated usage_test.go assertions)
// =============================================================================

interface IntentUsageTestCase {
  path: string[];
  name: string;
  hasArgs: boolean;
  hasSchema: boolean;
}

// Prefers a bound intent with a positional argument so the generated tests
// can also exercise the with-args invocation.
function getIntentUsageTestCase(): IntentUsageTestCase | null {
  const manifest = collectIntentManifest();
  if (manifest.Bound.length === 0) return null;
  const chosen =
    manifest.Bound.find((b) => b.Args.length > 0) || manifest.Bound[0];
  const name = firstUseWord(chosen.Use);
  return {
    path: [...chosen.ParentPath, name],
    name,
    hasArgs: chosen.Args.length > 0,
    hasSchema: chosen.HasSchema,
  };
}

function templateIntentUsageTestEnabled(): boolean {
  return getIntentUsageTestCase() !== null;
}
registerTemplateFunc(
  "templateIntentUsageTestEnabled",
  templateIntentUsageTestEnabled,
);

function templateIntentUsagePathArgs(): string {
  const testCase = getIntentUsageTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templateIntentUsagePathArgs",
  templateIntentUsagePathArgs,
);

function templateIntentUsageCmdName(): string {
  return getIntentUsageTestCase()?.name || "";
}
registerTemplateFunc("templateIntentUsageCmdName", templateIntentUsageCmdName);

function templateIntentUsageArgsTestEnabled(): boolean {
  return Boolean(getIntentUsageTestCase()?.hasArgs);
}
registerTemplateFunc(
  "templateIntentUsageArgsTestEnabled",
  templateIntentUsageArgsTestEnabled,
);

function templateIntentUsageSchemaPrecedenceTestEnabled(): boolean {
  return Boolean(getIntentUsageTestCase()?.hasSchema);
}
registerTemplateFunc(
  "templateIntentUsageSchemaPrecedenceTestEnabled",
  templateIntentUsageSchemaPrecedenceTestEnabled,
);

// Bound intents whose required input has a backing operation flag: --body
// combined with that flag explicitly empty must be rejected.
interface IntentEmptyBackingFlagTestCase {
  name: string;
  path: string[];
  bodyFlag: string;
  backingFlag: string;
  bodyJSON: string;
}

function getIntentEmptyBackingFlagTestCases(): IntentEmptyBackingFlagTestCase[] {
  const cases: IntentEmptyBackingFlagTestCase[] = [];
  for (const cmd of collectIntentManifest().Bound) {
    if (!cmd.BodyFlag) continue;
    const input = [...cmd.Args, ...cmd.Flags].find(
      (i) =>
        i.Required &&
        !i.PresetCovered &&
        i.BackingFlag &&
        (!i.Type || i.Type === "string"),
    );
    if (!input) continue;
    cases.push({
      name: firstUseWord(cmd.Use),
      path: [...cmd.ParentPath, firstUseWord(cmd.Use)],
      bodyFlag: cmd.BodyFlag,
      backingFlag: input.BackingFlag || "",
      bodyJSON: JSON.stringify({ [input.Key]: "example-value" }),
    });
  }
  return cases;
}

function templateIntentEmptyBackingFlagTestEnabled(): boolean {
  return getIntentEmptyBackingFlagTestCases().length > 0;
}
registerTemplateFunc(
  "templateIntentEmptyBackingFlagTestEnabled",
  templateIntentEmptyBackingFlagTestEnabled,
);

function templateIntentEmptyBackingFlagTestCases(): string {
  return getIntentEmptyBackingFlagTestCases()
    .map((c) => {
      const args = [
        ...c.path,
        `--${c.bodyFlag}`,
        c.bodyJSON,
        `--${c.backingFlag}`,
        "",
      ]
        .map(goStringLiteral)
        .join(", ");
      return `{name: ${goStringLiteral(
        c.name,
      )}, args: []string{${args}}, backingFlag: ${goStringLiteral(
        c.backingFlag,
      )}},`;
    })
    .join("\n\t\t");
}
registerTemplateFunc(
  "templateIntentEmptyBackingFlagTestCases",
  templateIntentEmptyBackingFlagTestCases,
);

function templateArtifactIntentUsageTestEnabled(): boolean {
  return getArtifactTestCommand() !== null;
}
registerTemplateFunc(
  "templateArtifactIntentUsageTestEnabled",
  templateArtifactIntentUsageTestEnabled,
);

function templateArtifactIntentUsagePathArgs(): string {
  const cmd = getArtifactTestCommand();
  if (!cmd) return "";
  return [...cmd.ParentPath, firstUseWord(cmd.Use)]
    .map((part) => goStringLiteral(part))
    .join(", ");
}
registerTemplateFunc(
  "templateArtifactIntentUsagePathArgs",
  templateArtifactIntentUsagePathArgs,
);

// The SSE wire protocol carries each event's JSON payload in its `data` field
// (see the SDK's ServerEvent decoder), so a declared SSE projection is only
// exercisable by the generated mock stream server when its pointer is rooted
// at the schema property mapped to that wire field. This is SSE framing, not
// a spec-authored name.
const SSE_DATA_ENVELOPE = "data";

// A declared stream toggle: a bool flag whose bound body key the command's
// preset pins to true — whatever the author named it. The generated stream
// tests use it to prove flag/preset/body-file merge semantics on the wire.
function intentStreamToggle(
  flags: IntentBodyEntry[],
  preset: Record<string, any>,
): IntentBodyEntry | null {
  for (const f of flags) {
    if ((f.Type || "string") !== "bool") continue;
    if (preset[f.Key] === true) return f;
  }
  return null;
}

// The streamed-projection test needs a bound intent that declares
// output.stream.select over an SSE operation, whose pointer starts at the SSE
// data envelope, whose declared surface carries a stream toggle (a preset-
// pinned bool flag, selected by declaration rather than by name), and whose
// required inputs are satisfiable from a single positional argument (presets
// cover any required flags).
interface IntentStreamTestCase {
  path: string[];
  hasArgs: boolean;
  bodyFlag: string;
  envelopeToken: string; // the SSE data envelope member (first pointer token)
  wireTokens: string[]; // pointer tokens below the SSE data envelope
  sentinel: string;
  promptArg: string;
  promptKey: string;
  toggleFlag: string; // declared stream-toggle flag name
  toggleKey: string; // body key the toggle binds
}

function getIntentStreamTestCase(): IntentStreamTestCase | null {
  const manifest = collectIntentManifest();
  for (const b of manifest.Bound) {
    if (!b.StreamSelect || b.StreamKind !== "sse") continue;
    if (b.PromptFlags.length > 0) continue;
    // The test supplies exactly one positional; an operation whose own
    // parameters (path/query/header) are required cannot be exercised.
    if (b.OpHasRequiredParams) continue;
    let preset: Record<string, any> = {};
    try {
      preset = b.PresetJSON ? JSON.parse(b.PresetJSON) : {};
    } catch (_err) {
      continue;
    }
    const toggle = intentStreamToggle(b.Flags, preset);
    if (!toggle || b.Args.length === 0) continue;
    const tokens = b.StreamSelect.replace(/^\//, "")
      .split("/")
      .map((t) => t.replace(/~1/g, "/").replace(/~0/g, "~"));
    if (tokens.length < 2 || tokens[0] !== SSE_DATA_ENVELOPE) continue;
    if (tokens.slice(1).some((t) => /^[0-9]+$/.test(t))) continue;
    return {
      path: [...b.ParentPath, firstUseWord(b.Use)],
      hasArgs: b.Args.length > 0,
      bodyFlag: b.BodyFlag,
      envelopeToken: tokens[0],
      wireTokens: tokens.slice(1),
      sentinel: b.StreamSentinel,
      promptArg: b.Args[0].Name || "input",
      promptKey: b.Args[0].Key,
      toggleFlag: toggle.Name || toggle.Key,
      toggleKey: toggle.Key,
    };
  }
  return null;
}

function templateIntentStreamTestEnabled(): boolean {
  return getIntentStreamTestCase() !== null;
}
registerTemplateFunc(
  "templateIntentStreamTestEnabled",
  templateIntentStreamTestEnabled,
);

function templateIntentStreamPathArgs(): string {
  const testCase = getIntentStreamTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templateIntentStreamPathArgs",
  templateIntentStreamPathArgs,
);

function templateIntentStreamHasArgs(): boolean {
  return Boolean(getIntentStreamTestCase()?.hasArgs);
}
registerTemplateFunc(
  "templateIntentStreamHasArgs",
  templateIntentStreamHasArgs,
);

function templateIntentStreamWireTokens(): string {
  const testCase = getIntentStreamTestCase();
  if (!testCase) return "";
  return testCase.wireTokens.map((t) => goStringLiteral(t)).join(", ");
}
registerTemplateFunc(
  "templateIntentStreamWireTokens",
  templateIntentStreamWireTokens,
);

function templateIntentStreamSentinel(): string {
  return goStringLiteral(getIntentStreamTestCase()?.sentinel || "");
}
registerTemplateFunc(
  "templateIntentStreamSentinel",
  templateIntentStreamSentinel,
);

function templateIntentStreamPromptArg(): string {
  return goStringLiteral(getIntentStreamTestCase()?.promptArg || "input");
}
registerTemplateFunc(
  "templateIntentStreamPromptArg",
  templateIntentStreamPromptArg,
);

function templateIntentStreamBodyFlag(): string {
  return getIntentStreamTestCase()?.bodyFlag || "body";
}
registerTemplateFunc(
  "templateIntentStreamBodyFlag",
  templateIntentStreamBodyFlag,
);

function templateIntentStreamPromptKey(): string {
  return goStringLiteral(getIntentStreamTestCase()?.promptKey || "prompt");
}
registerTemplateFunc(
  "templateIntentStreamPromptKey",
  templateIntentStreamPromptKey,
);

// Raw flag name of the declared stream toggle (rendered inside "--…" args).
function templateIntentStreamToggleFlag(): string {
  return getIntentStreamTestCase()?.toggleFlag || "stream";
}
registerTemplateFunc(
  "templateIntentStreamToggleFlag",
  templateIntentStreamToggleFlag,
);

// Go string literal of the body key the stream toggle binds.
function templateIntentStreamToggleKey(): string {
  return goStringLiteral(getIntentStreamTestCase()?.toggleKey || "stream");
}
registerTemplateFunc(
  "templateIntentStreamToggleKey",
  templateIntentStreamToggleKey,
);

// Raw SSE data-envelope member (rendered inside jq selectors).
function templateIntentStreamEnvelopeToken(): string {
  return getIntentStreamTestCase()?.envelopeToken || SSE_DATA_ENVELOPE;
}
registerTemplateFunc(
  "templateIntentStreamEnvelopeToken",
  templateIntentStreamEnvelopeToken,
);

// Go string literal of the SSE data-envelope member.
function templateIntentStreamEnvelopeKey(): string {
  return goStringLiteral(
    getIntentStreamTestCase()?.envelopeToken || SSE_DATA_ENVELOPE,
  );
}
registerTemplateFunc(
  "templateIntentStreamEnvelopeKey",
  templateIntentStreamEnvelopeKey,
);

interface OperationStreamTestCase {
  path: string[];
  bodyFlag: string;
  bodyJSON: string;
  envelopeToken: string; // the SSE data envelope member (first pointer token)
  wireTokens: string[];
  sentinel: string;
  toggleFlag: string; // declared stream-toggle flag name
  toggleKey: string; // body key the toggle binds
  toggleSummary: string; // declared stream-toggle help summary
  promptableStreamFlag: boolean;
}

// The operation-level stream toggle: a declared bool flag bound to a
// top-level body key whose effective default is on — an explicit
// `default: true` or a schema-supplied default (the generated test asserts
// the default is materialized on the wire). Selected by declaration shape,
// not by a flag literally named "stream".
function operationStreamToggle(declared: any): any | null {
  for (const f of declared.Flags || []) {
    if ((f.Type || "string") !== "bool") continue;
    if (!intentPointerKey(f.Bind?.Pointer || "")) continue;
    if (f.Default === true || f.DefaultFrom === "schema") return f;
  }
  return null;
}

// True when the operation's request body (including any union variant)
// declares wire property `key` with a schema default of true. Only that
// SDK-schema default is materialized on the wire when the toggle flag is not
// passed, which TestOperationStream_DefaultProjectsIncrementally asserts.
function operationBodySchemaDefaultTrue(op: Operation, key: string): boolean {
  const bodyType = op.Request?.RequestBody?.Type;
  if (!bodyType) return false;
  const candidates: TypeDef[] =
    bodyType.Type.toString() === "union"
      ? bodyType.AssociatedTypes || []
      : [bodyType];
  for (const candidate of candidates) {
    for (const field of candidate.Fields || []) {
      if (field.OriginalName !== key && field.Name !== key) continue;
      const value = field.Default?.Value;
      if (value === true || (value != null && String(value) === "true")) {
        return true;
      }
    }
  }
  return false;
}

function getOperationStreamTestCase(): OperationStreamTestCase | null {
  const manifest = context.Global.AST.CLICommands as any;
  for (const declared of manifest?.Operations || []) {
    if (!declared.Output?.Stream) continue;
    const toggle = operationStreamToggle(declared);
    if (!toggle) continue;
    const toggleKey = intentPointerKey(toggle.Bind?.Pointer || "");
    const found = findIntentOperation(declared.OperationID);
    if (!found || intentStreamKind(found.op).kind !== "sse") continue;
    if (!toggleKey || !operationBodySchemaDefaultTrue(found.op, toggleKey)) {
      continue;
    }
    // The generated test drives the operation through its single whole-JSON
    // body flag (required-body error, duplicate-source rule, whole-body
    // prompting); an expandable body has per-field flags instead and does not
    // satisfy those premises.
    if (isRequestBodyExpandable(found.op)) continue;
    const ctx = getCLIOperationCtx(found.op);
    if (!ctx?.CanonicalBodyFlag) continue;
    const bound = collectIntentManifest().Bound.find(
      (b) =>
        (b.OpID === found.op.OriginalID || b.OpID === found.op.GetID()) &&
        b.Args.length > 0,
    );
    if (!bound) continue;
    const body: Record<string, any> = {};
    body[bound.Args[0].Key] = "hello from operation";
    const tokens = declared.Output.Stream.Pointer.replace(/^\//, "")
      .split("/")
      .map((t: string) => t.replace(/~1/g, "/").replace(/~0/g, "~"));
    if (tokens[0] !== SSE_DATA_ENVELOPE) continue;
    return {
      path: getCLICommandPath(found.op),
      bodyFlag: ctx.CanonicalBodyFlag,
      bodyJSON: JSON.stringify(body),
      envelopeToken: tokens[0],
      wireTokens: tokens.slice(1),
      sentinel: intentStreamKind(found.op).sentinel,
      toggleFlag: toggle.Name || toggleKey,
      toggleKey,
      toggleSummary: toggle.Summary || "",
      promptableStreamFlag: toggle.DefaultFrom !== "schema",
    };
  }
  return null;
}

function templateOperationStreamTestEnabled(): boolean {
  return (
    getOperationStreamTestCase() !== null && getIntentStreamTestCase() !== null
  );
}
registerTemplateFunc(
  "templateOperationStreamTestEnabled",
  templateOperationStreamTestEnabled,
);

function templateOperationStreamPromptTestEnabled(): boolean {
  return Boolean(getOperationStreamTestCase()?.promptableStreamFlag);
}
registerTemplateFunc(
  "templateOperationStreamPromptTestEnabled",
  templateOperationStreamPromptTestEnabled,
);

function templateOperationStreamPathArgs(): string {
  return (getOperationStreamTestCase()?.path || [])
    .map(goStringLiteral)
    .join(", ");
}
registerTemplateFunc(
  "templateOperationStreamPathArgs",
  templateOperationStreamPathArgs,
);

function templateOperationStreamBodyFlag(): string {
  return getOperationStreamTestCase()?.bodyFlag || "body";
}
registerTemplateFunc(
  "templateOperationStreamBodyFlag",
  templateOperationStreamBodyFlag,
);

function templateOperationStreamBodyJSON(): string {
  return goStringLiteral(getOperationStreamTestCase()?.bodyJSON || "{}");
}
registerTemplateFunc(
  "templateOperationStreamBodyJSON",
  templateOperationStreamBodyJSON,
);

function templateOperationStreamWireTokens(): string {
  return (getOperationStreamTestCase()?.wireTokens || [])
    .map(goStringLiteral)
    .join(", ");
}
registerTemplateFunc(
  "templateOperationStreamWireTokens",
  templateOperationStreamWireTokens,
);

function templateOperationStreamSentinel(): string {
  return goStringLiteral(getOperationStreamTestCase()?.sentinel || "");
}
registerTemplateFunc(
  "templateOperationStreamSentinel",
  templateOperationStreamSentinel,
);

// Raw flag name of the operation's declared stream toggle (rendered inside
// "--…" args and flag-referencing assertions).
function templateOperationStreamToggleFlag(): string {
  return getOperationStreamTestCase()?.toggleFlag || "stream";
}
registerTemplateFunc(
  "templateOperationStreamToggleFlag",
  templateOperationStreamToggleFlag,
);

// Go string literal of the body key the operation stream toggle binds.
function templateOperationStreamToggleKey(): string {
  return goStringLiteral(getOperationStreamTestCase()?.toggleKey || "stream");
}
registerTemplateFunc(
  "templateOperationStreamToggleKey",
  templateOperationStreamToggleKey,
);

// Go string literal of the declared stream toggle's help summary ("" when the
// declaration has none; the usage assertion is then vacuous by design).
function templateOperationStreamToggleSummary(): string {
  return goStringLiteral(getOperationStreamTestCase()?.toggleSummary || "");
}
registerTemplateFunc(
  "templateOperationStreamToggleSummary",
  templateOperationStreamToggleSummary,
);

// =============================================================================
// Built-in / catalog usage test cases (drive the generated usage_test.go)
// =============================================================================

// The built-in auth group is only reachable when no generated group owns
// the name (see buildRootUsageCommand).
function templateAuthUsageTestEnabled(): boolean {
  if (!isInteractiveAuthEnabled() || !hasGlobalSecurity()) return false;
  const rootNames = context.Global.AST.MainSDK.SubSDKs.map((sdk: SDK) =>
    sanitizeCLICommand(getSDKGroupName(sdk)),
  );
  return !rootNames.includes("auth");
}
registerTemplateFunc(
  "templateAuthUsageTestEnabled",
  templateAuthUsageTestEnabled,
);

function getCatalogUsageTestCase(): CliCatalog | null {
  const catalogs = collectCliCatalogs();
  return catalogs.length > 0 ? catalogs[0] : null;
}

function templateCatalogUsageTestEnabled(): boolean {
  return getCatalogUsageTestCase() !== null;
}
registerTemplateFunc(
  "templateCatalogUsageTestEnabled",
  templateCatalogUsageTestEnabled,
);

function templateCatalogUsageCmdName(): string {
  return getCatalogUsageTestCase()?.Command || "";
}
registerTemplateFunc(
  "templateCatalogUsageCmdName",
  templateCatalogUsageCmdName,
);

function templateCatalogUsageFirstValue(): string {
  const catalog = getCatalogUsageTestCase();
  return catalog ? escapeGoString(catalog.Values[0].Value) : "";
}
registerTemplateFunc(
  "templateCatalogUsageFirstValue",
  templateCatalogUsageFirstValue,
);

function getPlannedUsageTestCase(): { path: string[]; name: string } | null {
  const manifest = collectIntentManifest();
  const planned = manifest.Planned[0];
  if (!planned) return null;
  return { path: [...planned.ParentPath, planned.Name], name: planned.Name };
}

function templatePlannedUsageTestEnabled(): boolean {
  return getPlannedUsageTestCase() !== null;
}
registerTemplateFunc(
  "templatePlannedUsageTestEnabled",
  templatePlannedUsageTestEnabled,
);

function templatePlannedUsagePathArgs(): string {
  const testCase = getPlannedUsageTestCase();
  if (!testCase) return "";
  return testCase.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templatePlannedUsagePathArgs",
  templatePlannedUsagePathArgs,
);

function templatePlannedUsageCmdName(): string {
  return getPlannedUsageTestCase()?.name || "";
}
registerTemplateFunc(
  "templatePlannedUsageCmdName",
  templatePlannedUsageCmdName,
);

function templateUsageSchemaMap(): string {
  const root = buildRootUsageCommand();
  const schemas = new Map<string, string>();
  collectCommandPathSchemas(root, [], schemas, true);
  const entries = Array.from(schemas.entries()).map(
    ([k, v]) => `${goStringLiteral(k)}: ${goTemplateSafeStringExpr(v)}`,
  );
  return `var usageSchemas = map[string]string{\n\t${entries.join(
    ",\n\t",
  )},\n}`;
}
registerTemplateFunc("templateUsageSchemaMap", templateUsageSchemaMap);
