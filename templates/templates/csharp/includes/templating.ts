// @ts-ignore
function templateConstName(name: string, prefix: string) {
  return caser().ToPascal(sanitizeName(`${prefix}${name}`));
}

registerTemplateFunc("templateConstName", templateConstName);

// @ts-ignore
function sanitizeScopeName(name: string) {
  return caser().ToPascal(name);
}

registerTemplateFunc("sanitizeScopeName", sanitizeScopeName);

// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  const lines = [];
  const enumNames = getEnumNames(typeDef);
  const descriptions = typeDef.Enum.Descriptions || {};

  typeDef.Enum.Values.forEach((value, index) => {
    const description = descriptions[`${value}`];
    if (typeof description === "string" && description) {
      lines.push(
        ...templateCommentElement("summary", null, description, 0).split("\n"),
      );
    }

    if (typeDef.Enum.Type.Type.toString() == "string") {
      lines.push(`[JsonProperty("${value.replaceAll(`\\`, `\\\\`)}")]`);
    }

    const assignment =
      typeDef.Enum.Type.Type.toString() != "string" ? ` = ${value}` : "";
    lines.push(`${enumNames[index]}${assignment},`);
  });

  return "\n" + indentLines(lines, 2);
}

registerTemplateFunc("templateEnums", templateEnums);

function templateOpenEnumKnownValues(typeDef: TypeDef): string {
  const enumNames = getEnumNames(typeDef);
  const enumType = sanitizeType(typeDef.Enum.Type, false, "");
  const className = sanitizeClassName(typeDef.Name, false);
  const isStringEnum = typeDef.Enum.Type.Type.toString() === "string";

  function formatValue(value: string): string {
    return isStringEnum ? `"${value}"` : value;
  }

  const staticFields = typeDef.Enum.Values.map(
    (v, i) =>
      `public static readonly ${className} ${
        enumNames[i]
      } = new ${className}(${formatValue(v)});`,
  ).join("\n");

  const knownValues = typeDef.Enum.Values.map(
    (v, i) => `${templateIndent(2)}[${formatValue(v)}] = ${enumNames[i]}`,
  ).join(",\n");

  const knownValuesDict = `
private static readonly Dictionary <${enumType}, ${className}> _knownValues =
    new Dictionary <${enumType}, ${className}> ()
    {
${knownValues}
    };`;

  return `\n${staticFields}\n${knownValuesDict}`;
}

registerTemplateFunc(
  "templateOpenEnumKnownValues",
  templateOpenEnumKnownValues,
);

function getMethodParameters(op: Operation): FieldDef[] {
  const fields: FieldDef[] = [];

  const isFlattened = op.Arguments.Flattening != "none";
  const isLegacyOrdering = context.Global.Config.FlatteningOrder == "";

  if (isFlattened && isLegacyOrdering) {
    fields.push(...sortMethodParams(op.Request.Field.Type.Fields));

    if (op.Security) {
      op.Security.Optional
        ? fields.push(op.Security)
        : fields.unshift(op.Security);
    }
  } else {
    fields.push(...op.Arguments.Sorted);
  }

  // Optional parameters must appear after all required parameters
  fields.sort((a, b) => {
    const canAddDefaultValueForA =
      canFieldHaveDefaultValue(a) || a.Optional || a.Nullable;
    const canAddDefaultValueForB =
      canFieldHaveDefaultValue(b) || b.Optional || b.Nullable;
    // if there is a default for A and not for B then A should be after B
    if (canAddDefaultValueForA && !canAddDefaultValueForB) {
      return 1;
    }
    // if there is a default for B and not for A then B should be after A
    if (!canAddDefaultValueForA && canAddDefaultValueForB) {
      return -1;
    }
    // Keep the original order
    return 0;
  });

  return fields;
}

type MethodArgument = {
  Type: string;
  Name: string;
  Optional: boolean;
  Default: string;
  Description: string;
};

function templateMethodArgument(
  arg: MethodArgument,
  definition: boolean,
): string {
  if (!definition) {
    return arg.Name;
  }

  const dfltAssignment =
    arg.Default !== "" ? ` = ${arg.Default}` : arg.Optional ? " = null" : "";

  return `${arg.Type}${arg.Optional ? "?" : ""} ${arg.Name}${dfltAssignment}`;
}

// @ts-ignore
function getMethodArguments(operation: Operation): MethodArgument[] {
  const args: MethodArgument[] = [];

  getMethodParameters(operation).forEach((field) => {
    args.push({
      Type: sanitizeType(field.Type, false, ""),
      Name: sanitizeMethodParamName(field.Name),
      Optional: field.Optional || field.Nullable,
      Default: canFieldHaveDefaultValue(field)
        ? templateConstOrDefaultValue(field)
        : "",
      Description: extractParamDescription(
        field.Comments,
        getFieldFallbackDescription(field, ""),
      ),
    });
  });

  if (operation.Servers) {
    args.push({
      Type: "string",
      Name: "serverUrl",
      Optional: true,
      Default: "",
      Description:
        "The server URL to use for this operation. If not provided, the default server URL will be used.",
    });
  }

  if (operation.Extensions.Retries) {
    args.push({
      Type: "RetryConfig",
      Name: "retryConfig",
      Optional: true,
      Default: "",
      Description: "The retry configuration to use for this operation.",
    });
  }

  if (operation.Extensions.Pagination && hasPaginationURL(operation)) {
    args.push({
      Type: "string",
      Name: "urlOverride",
      Optional: true,
      Default: "",
      Description: "The URL to use for the next page of results.",
    });
  }

  if (context.Global.Config.EnableCancellationToken) {
    args.push({
      Type: "CancellationToken",
      Name: "cancellationToken",
      Optional: true,
      Default: "null",
      Description:
        "An optional cancellation token to signal when the operation should be aborted.",
    });
  }

  return args;
}

function templateMethodExceptionComments(op: Operation): string[] {
  const exceptions: Record<string, { codes: string[]; description: string }> =
    {};
  let isDefaultExceptionOverridden = false;
  const defaultExceptionCodes: string[] = [];

  // Collect arguments that should not be null
  const nonNullFieldNames = getMethodParameters(op)
    .filter((f) => !f.Optional && !f.Nullable && canTypeBeNull(f.Type))
    .map((f) => sanitizeMethodParamName(f.Name));

  switch (nonNullFieldNames.length) {
    case 0:
      break;
    case 1:
      exceptions["ArgumentNullException"] = {
        codes: [],
        description: `The required parameter <paramref name="${nonNullFieldNames[0]}"/> is null.`,
      };
      break;
    default:
      const paramrefs = nonNullFieldNames
        .map((name) => `<paramref name="${name}"/>`)
        .join(", ")
        .replace(/, ([^,]*)$/, " or $1");

      exceptions["ArgumentNullException"] = {
        codes: [],
        description: `One of ${paramrefs} is null.`,
      };
      break;
  }

  // Cancellation token (if enabled)
  if (context.Global.Config.EnableCancellationToken) {
    exceptions["OperationCanceledException"] = {
      codes: [],
      description:
        "The operation was aborted via the provided cancellation token.",
    };
  }

  // HttpClient.SendAsync()
  exceptions["HttpRequestException"] = {
    codes: [],
    description: "The HTTP request failed due to network issues.",
  };

  // Deserialization failures
  if (op.Response.Responses.some((subResp) => subResp.Content.length > 0)) {
    exceptions["ResponseValidationException"] = {
      codes: [],
      description: "The response body could not be deserialized.",
    };
  }

  // Collect exceptions from error responses
  op.Response.Responses.filter((r) => r.Error).forEach((subResp) => {
    if (subResp.Content.length === 0) {
      defaultExceptionCodes.push(...subResp.Code);
    } else {
      if ("default" in subResp.Code) {
        isDefaultExceptionOverridden = true;
      }
      subResp.Content.forEach((content) => {
        if (content.Content) {
          let errorType = "";
          if (content.Content.Type.Type.toString() == "union") {
            errorType = sanitizeFieldName(
              sanitizeUnionTypeName(content.Content.Type),
            );
          } else {
            errorType = sanitizeType(
              content.Content.Type,
              false,
              content.Content.Type.OutputLocation,
              sanitizeClassName(op.OwningSDK.Type.Name),
            );
          }

          if (!exceptions[errorType]) {
            exceptions[errorType] = { codes: [], description: "" };
          }
          const entry = exceptions[errorType];
          entry.codes.push(...subResp.Code);

          const errorDesc = content.Content.Comments?.Description || "";
          if (errorDesc != "" && !entry.description) {
            entry.description = sanitizeComments(errorDesc) + " ";
          }
        }
      });
    }
  });

  // Default Exception
  if (!isDefaultExceptionOverridden) {
    exceptions[getDefaultErrorClassName()] = {
      codes: defaultExceptionCodes,
      description: "Default API Exception. ",
    };
  }

  // Format <exception> tags
  const comments: string[] = [];
  for (const e in exceptions) {
    let { codes, description } = exceptions[e];

    function codeText(codes: string[]): string {
      return codes.length === 1
        ? codes[0]
        : `${codes.slice(0, -1).join(", ")} or ${codes[codes.length - 1]}`;
    }

    if (codes.includes("default")) {
      description += `Thrown when the response status code is none of ${codeText(
        getNonErrorStatusCodes(op.Response),
      )}.`;
    } else if (codes.length > 0) {
      description += `Thrown when the API returns a ${codeText(
        sanitizeStatusCodes(codes),
      )} response.`;
    }
    comments.push(templateExceptionComment(e, description, 0));
  }

  return comments;
}

function templateMethodDefinition(
  op: Operation,
  owningClassName: string,
  indent: number,
  isDeclaration: boolean = false,
): string {
  const methodName = sanitizeMethodName(op, true);

  let resultTypeName = "";
  let task = "Task";
  if (op.Response.Type) {
    resultTypeName = sanitizeTypeName(op.Response.Type, "", owningClassName);
    task += `<${resultTypeName}>`;
  }

  const methodAccess = `public ${isDeclaration ? "" : "async "}`;
  const definition = `${methodAccess} ${task} ${methodName}(`;

  const args: MethodArgument[] = getMethodArguments(op);
  const params: string[] = args.map((a) => templateMethodArgument(a, true));
  const returns = templateReturnsComment(resultTypeName, op.Response.Type, 0);
  const exceptions = templateMethodExceptionComments(op);
  const comments = indentLines(
    [
      "",
      ...args.map((a) => templateParamComment(a, 0)),
      ...returns.split("\n"),
      ...exceptions,
      ...(op.Comments?.Deprecated
        ? [templateDeprecationAnnotation(op.Comments, 0, "method", false)]
        : []),
    ],
    indent,
  );

  const flatLength =
    templateIndent(indent).length +
    definition.length +
    params.join(", ").length;
  if (flatLength > 120) {
    const multiLineParams = params.map(
      (a, i) => `${templateIndent(1)}${a}${i < params.length - 1 ? "," : ""}`,
    );

    return `${comments}
${indentLines([definition, ...multiLineParams, ")"], indent)}`;
  }

  return `${comments}
${templateIndent(indent)}${definition}${params.join(", ")})`;
}

registerTemplateFunc("templateMethodDefinition", templateMethodDefinition);

function canFieldHaveDefaultValue(field: FieldDef): boolean {
  return (
    field.Default &&
    !["date", "date-time", "bytes", "biginteger"].includes(
      field.Type.Type.toString(),
    )
  );
}

function canTypeBeNull(typeDef: TypeDef): boolean {
  return ![
    "date", // LocalDate or DateOnly
    "date-time", // DateTime
    "integer", // long
    "int32", // int
    "bigint", // BigInteger
    "number", // double
    "float32", // float
    "decimal", // decimal
    "boolean", // bool
    "enum", // enum
  ].includes(typeDef.Type.toString());
}

function canFieldBeNull(field: FieldDef): boolean {
  return field.Optional || field.Nullable || canTypeBeNull(field.Type);
}

registerTemplateFunc("canFieldBeNull", canFieldBeNull);

// @ts-ignore
function templateMethodArgumentsNullCheck(
  operation: Operation,
  indent: number,
): string {
  const lines = getMethodParameters(operation)
    .filter((f) => !f.Optional && !f.Nullable && canTypeBeNull(f.Type))
    .map((f) => {
      const fieldName = sanitizeMethodParamName(f.Name);
      return `if (${fieldName} == null) throw new ArgumentNullException(nameof(${fieldName}));`;
    });

  if (lines.length > 0) {
    return "\n" + indentLines(lines, indent);
  }

  return "";
}

registerTemplateFunc(
  "templateMethodArgumentsNullCheck",
  templateMethodArgumentsNullCheck,
);

// @ts-ignore
function joinParams(params: string[]): string {
  switch (params.length) {
    case 0:
      return "";
    case 1:
      return params[0];
    default:
      const indentedParams = params.map((p) => indentString(p, 1));
      return `\n${indentedParams.join(",\n")}`;
  }
}

// @ts-ignore
function templateUsageSDKParams(
  local: UsageContext,
  useBuilderInit: boolean = false,
  indent: number = 0,
): string {
  const params = [];
  let hasSecurity = local.Operation.Security != undefined;

  const ctx: TemplateValueContext = {
    isTest: local.Test && true,
    usageContext: local,
    operation: local.Operation,
    test: local.Test,
  };

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal && scope.Feature != "") {
        const feature = scope.Feature.toString();

        switch (feature) {
          case "security":
            if (!hasSecurity) {
              const globalSecurity = getGlobalSecurity(scope.Value, ctx);
              if (globalSecurity) {
                params.push(
                  useBuilderInit
                    ? `.With${globalSecurity.fieldName}(${globalSecurity.value})`
                    : `${globalSecurity.fieldName}: ${globalSecurity.value}`,
                );
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            const hasOperationServers =
              local.Operation?.Servers?.Servers?.length > 0;
            if (!hasOperationServers) {
              const serverUrl = getUsageServerUrl(local, scope);
              params.push(
                useBuilderInit
                  ? `.WithServerUrl(${serverUrl})`
                  : `serverUrl: ${serverUrl}`,
              );
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              const serverName = `SDKConfig.Server.${sanitizeFieldName(
                server.ID,
              )}`;
              params.push(
                useBuilderInit
                  ? `.WithServer(${serverName})`
                  : `server: ${serverName}`,
              );
            } else if (server.Index !== undefined) {
              params.push(
                useBuilderInit
                  ? `.WithServerIndex(${server.Index})`
                  : `serverIndex: ${server.Index}`,
              );
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const fieldName = sanitizePrivateFieldName(v.Name);
                const exampleValue = `"${getUsageServerVariableValue(v)}"`;
                params.push(
                  useBuilderInit
                    ? `.With${sanitizeFieldName(v.Name)}(${exampleValue})`
                    : `${fieldName}: ${exampleValue}`,
                );
              });
            }

            break;
          }
          case "retries": {
            const retries = templateString("usage/retries.stmpl", {});
            params.push(
              useBuilderInit ? `.WithRetryConfig(${retries})` : retries,
            );
            break;
          }
          case "parameter": {
            const parameter = scope.Value as ParameterUsage;
            const value = templateValueAsRequired(
              parameter.field,
              getExampleValue(parameter.example),
              ctx,
            );

            if (value == "") {
              return;
            }
            const fieldName = sanitizePrivateFieldName(parameter.field.Name);

            params.push(
              useBuilderInit
                ? `.With${sanitizeFieldName(parameter.field.Name)}(${value})`
                : `${fieldName}: ${value}`,
            );

            break;
          }
          case "http_client": {
            params.push(
              useBuilderInit
                ? `.WithClient(${templateHTTPClient(scope.Value)})`
                : `client: ${templateHTTPClient(scope.Value)}`,
            );
            break;
          }
        }
      }
    });
  }

  if (!hasSecurity && opUsesGlobalSecurity(local.Operation)) {
    const globalSecurity = getGlobalUsageSecurity(local);
    if (globalSecurity) {
      params.push(
        useBuilderInit
          ? `.With${sanitizeFieldName(globalSecurity.fieldName)}(${
              globalSecurity.value
            })`
          : `${globalSecurity.fieldName}: ${globalSecurity.value}`,
      );
    }
  }

  if (useBuilderInit) {
    return params.map((p) => indentString(p, indent)).join("\n");
  }

  return joinParams(params) + (params.length > 1 ? "\n" : "");
}

registerTemplateFunc("templateUsageSDKParams", templateUsageSDKParams);

// @ts-ignore
function templateHTTPClient(value: HTTPClientUsage): string {
  switch (value.type) {
    case "test":
      return `testHttpClient`;
    default:
      throw new Error(`unsupported http client type: ${value.type}`);
  }
}

// @ts-ignore
function getOptionalUsageMethodParameters(
  local: UsageContext,
  indent: number,
): string[] {
  const params = [];

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (!scope.IsGlobal && scope.Feature != "") {
        switch (scope.Feature.toString()) {
          case "server_url": {
            const url = getUsageServerUrl(local, scope);
            params.push(`serverUrl: ${url}`);
            break;
          }
        }
      }
    });
  }

  return params;
}

// @ts-ignore
function templateUsageMethodParameters(
  local: UsageContext,
  indent: number,
  extraParam: string = "",
): string {
  const operation = local.Operation;

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
    isUsage: true,
  };

  const methodParams: string[] = [];
  if (extraParam) {
    methodParams.push(extraParam);
  }

  if (operation.Security) {
    let securityExample = undefined;
    for (const scope of local.Scopes) {
      if (scope.Feature == "security") {
        if (scope.Value) {
          if ("example" in scope.Value) {
            securityExample = scope.Value.example;
          } else {
            securityExample = getExampleValue(scope.Value, ctx);
          }
        }
        break;
      }
    }
    methodParams.push(
      `security: ${templateSecurityUsage(
        operation.Security.Type,
        indent,
        true,
        false,
        undefined,
        securityExample,
        ctx,
      ).trimStart()}`,
    );
  }

  if (operationParametersFlattened(operation)) {
    methodParams.push(...getOptionalUsageMethodParameters(local, indent));

    if (operation.Request) {
      for (const field of sortMethodParams(
        operation.Request.Field.Type.Fields,
      )) {
        const fieldExample = getOperationMethodFieldExample(local, field);

        const value = templateValue(field, fieldExample, false, ctx);

        if (value == "") {
          continue;
        }

        methodParams.push(`${sanitizeMethodParamName(field.Name)}: ${value}`);
      }
    }
  } else {
    const optionalParams = getOptionalUsageMethodParameters(local, indent);

    if (operation.Request) {
      const reqVar = getRequestVariableName(local.StepID);
      if (methodParams.length + optionalParams.length == 0) {
        methodParams.push(reqVar);
      } else {
        methodParams.push(
          `${sanitizeMethodParamName(operation.Request.Field.Name)}: ${reqVar}`,
        );
      }
    }

    methodParams.push(...optionalParams);
  }

  if (methodParams.length > 1) {
    return joinParams(methodParams) + "\n";
  }
  return joinParams(methodParams);
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${sanitizeField(fieldDef, parent)} = `;
}

// @ts-ignore
function templateMethodArgumentDeclaration(fieldDef: FieldDef): string {
  return `${sanitizeMethodParamName(fieldDef.Name)}: `;
}

// @ts-ignore
function templateAccessOperator(fieldDef: FieldDef): string {
  const optional = fieldDef.Optional || fieldDef.Nullable;
  return optional ? "?." : ".";
}

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "null";
}

// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

// @ts-ignore
function templateIndent(indent: number): string {
  return "    ".repeat(indent);
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  const type = sanitizeType(typeDef, false, "", "", additionalContext?.isUsage);
  return `new ${type}() `;
}

// @ts-ignore
function templateIntValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "integer";
  }

  switch (t) {
    case "bigint":
      addUsageImport("System.Numerics");
      return `BigInteger.Parse("${val}")`;
    default:
      return `${val}`;
  }
}

// @ts-ignore
function templateBoolValue(
  val: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return val ? "true" : "false";
}

// @ts-ignore
function templateByteValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  addUsageImport("System");
  return `System.Text.Encoding.UTF8.GetBytes("${val.replace(/"/g, '\\"')}")`;
}

// @ts-ignore
function templateDateValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (useNodatime()) {
    addUsageImport("NodaTime");
    return `LocalDate.FromDateTime(System.DateTime.Parse("${val}"))`;
  } else {
    addUsageImport("System");
    return `DateOnly.Parse("${val}")`;
  }
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  addUsageImport("System");
  return `System.DateTime.Parse("${val}").ToUniversalTime()`;
}

// @ts-ignore
function templateFloatValue(
  value: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "number";
  }

  switch (t) {
    case "float32":
      return `${value}F`;
    case "decimal":
      return `${value}M`;
    case "number":
      return `${value}D`;
  }
}

// @ts-ignore
function getEnumNames(t: TypeDef): string[] {
  let names;
  if (t.Enum?.Names.length > 0) {
    names = t.Enum.Names.map((n) => getEnumName(n));
  } else {
    names = getEnumNamesFromValues(t.Enum?.Values);
  }

  // Open enums are classes, where a member sharing the class name is CS0542;
  // closed enums stay untouched since enum members may legally shadow the type.
  if (t.Enum?.Open) {
    const className = sanitizeClassName(t.Name, false);
    const taken = new Set(names);
    names = names.map((n) => {
      if (n !== className) {
        return n;
      }
      let candidate = `${n}Value`;
      let suffix = 1;
      while (taken.has(candidate)) {
        candidate = `${n}Value${suffix++}`;
      }
      taken.add(candidate);
      return candidate;
    });
  }

  return names;
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
): string {
  const definition =
    !!(fieldDef.Const || fieldDef.Default) && !additionalContext?.isUsage;

  const enumType = sanitizeClass(
    fieldDef.Type,
    "",
    definition,
    additionalContext?.isUsage,
  );

  return `${enumType}.${getEnumNames(fieldDef.Type)[idx]}`;
}

// @ts-ignore
function templateStringValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const multiline = `${val}`.includes("\n");

  if (multiline && additionalContext?.isTest) {
    // For test context, avoid verbatim strings (@"...") because template
    // indentation becomes part of the string literal. Use regular escaped
    // strings with explicit \n instead.
    val = val.replaceAll(/\\/g, "\\\\");
    val = val.replaceAll(/"/g, '\\"');
    val = val.replaceAll(/{{/g, `{{"{{"}}`);
    val = val.replaceAll(/\r/g, "\\r");
    val = val.replaceAll(/\n/g, "\\n");
    return `"${val}"`;
  }

  // In multi-line strings, we need to escape the double quotes with double quotes
  if (multiline) {
    val = val.replaceAll(/"/g, '""');
  } else {
    // For regular strings, escape backslashes first, then quotes
    val = val.replaceAll(/\\/g, "\\\\");
    val = val.replaceAll(/"/g, '\\"');
  }

  if (additionalContext?.isUsage || additionalContext?.isTest) {
    val = val.replaceAll(/{{/g, `{{"{{"}}`);
  }

  val = val.replaceAll(/\r/g, "\\r");

  return multiline ? `@"${val}"` : `"${val}"`;
}

// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  opening: boolean,
  additionalContext?: TemplateValueContext,
): string {
  return opening ? "{" : "}";
}

// @ts-ignore
function templateMapValue(key: string, val: any): string {
  addUsageImport("System.Collections.Generic");
  return `{ "${key}", ${val} }`;
}

// @ts-ignore
function templateArrayValue(val: any): string {
  addUsageImport("System.Collections.Generic");
  return val;
}

function templateCommentElement(
  tag: "summary" | "remarks" | "param" | "exception" | "returns",
  attribute: string | null,
  description: string,
  indent: number,
): string {
  const descriptionLines = sanitizeComments(description)
    .split("\n")
    .filter((l) => l.trim() !== "");

  let attr = "";
  if (attribute) {
    attr = ` ${sanitizeXMLAttribute(attribute)}`;
  }

  if (["summary", "remarks"].includes(tag) || descriptionLines.length > 1) {
    return indentLines(
      [
        `/// <${tag}${attr}>`,
        ...descriptionLines.map((line) => `/// ${line}`),
        `/// </${tag}>`,
      ],
      indent,
    );
  }

  return `${templateIndent(indent)}/// <${tag}${attr}>${
    descriptionLines[0] || "Description not available."
  }</${tag}>`;
}

// @ts-ignore
function templateParamComment(arg: MethodArgument, indent: number): string {
  return templateCommentElement(
    "param",
    `name="${arg.Name}"`,
    arg.Description,
    indent,
  );
}

// @ts-ignore
function templateReturnsComment(
  returnTypeName: string,
  typeDef: TypeDef | undefined,
  indent: number,
): string {
  const getDescription = () => {
    if (!returnTypeName || !typeDef) {
      return "A task that represents the asynchronous operation.";
    }

    if (typeDef.Comments?.Description) {
      return typeDef.Comments.Description;
    }

    if (getResponseFormat() !== "flat") {
      return `An awaitable task that returns a <see cref="${returnTypeName}"/> response envelope when completed.`;
    }

    switch (typeDef.Type.toString()) {
      case "class":
      case "enum":
      case "union":
        return `An awaitable task that returns a <see cref="${returnTypeName}"/> object when completed.`;
      default:
        return `An awaitable task that returns a ${returnTypeName} when completed.`;
    }
  };

  return templateCommentElement("returns", null, getDescription(), indent);
}

// @ts-ignore
function templateExceptionComment(
  exceptionType: string,
  description: string,
  indent: number,
): string {
  return templateCommentElement(
    "exception",
    `cref="${exceptionType}"`,
    description,
    indent,
  );
}

// @ts-ignore
function getMethodComments(op: Operation): CommentDef | null | undefined {
  const securityRemark = getHoistedSecurityRemark(op);
  if (!securityRemark) {
    return op.Comments;
  }

  const description = op.Comments?.Description
    ? `${op.Comments.Description}
<para>${securityRemark}</para>`
    : securityRemark;

  return {
    ...op.Comments,
    Description: description,
  } as CommentDef;
}

registerTemplateFunc("getMethodComments", getMethodComments);

// @ts-ignore
function templateComments(
  comments: CommentDef | null | undefined,
  indent: number,
  _type: "method" | "field" | "class",
  addLeadingNewLine: boolean = true,
): string {
  let docstring: string = addLeadingNewLine ? "\n" : "";

  // Note only the <summary> element shows in IntelliSense tooltips
  let summary = "";
  function appendToSummary(content: string) {
    summary += (summary ? "\n" : "") + content;
  }

  if (comments) {
    if (comments.Summary) {
      summary = comments.Summary;
    } else if (comments.Description) {
      summary = comments.Description;
    }

    if (comments.ExternalDocs) {
      if (comments.ExternalDocs.URL != "") {
        appendToSummary(
          `<see href="${comments.ExternalDocs.URL}">${
            comments.ExternalDocs.Description || comments.ExternalDocs.URL
          }</see>`,
        );
      } else if (comments.ExternalDocs.Description != "") {
        appendToSummary(comments.ExternalDocs.Description);
      }
    }
  }

  if (!summary) {
    return docstring;
  }

  // <summary> element
  docstring += `
${templateCommentElement("summary", null, summary, indent)}`;

  // <remarks> element is only added if both Summary and Description are provided.
  if (comments?.Summary && comments?.Description) {
    docstring += `
${templateCommentElement("remarks", null, comments.Description, indent)}`;
  }

  return docstring;
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function getSDKArguments(sdk: SDK): MethodArgument[] {
  const params: MethodArgument[] = [];

  // Security parameters
  if (sdk.Security) {
    const secParam = getSecurityParam(sdk);
    params.push({
      Type: secParam.Type,
      Name: secParam.Name,
      Optional: true,
      Default: "null",
      Description:
        "The security configuration to use for API requests. If provided, this will be used as a static security configuration.",
    });
    params.push({
      Type: `Func<${secParam.Type}>`,
      Name: secParam.Source,
      Optional: true,
      Default: "null",
      Description:
        "A function that returns the security configuration dynamically. This takes precedence over the static security parameter if both are provided.",
    });
  }

  // Global parameters
  if (sdk.Globals) {
    for (const parameter of sdk.Globals.Fields) {
      params.push({
        Type: sanitizeType(parameter.Type, false, ""),
        Name: sanitizeMethodParamName(parameter.Name),
        Optional: true,
        Default: "null",
        Description: extractParamDescription(
          parameter.Comments,
          `Global parameter for ${parameter.Name}.`,
        ),
      });
    }
  }

  // Server selection
  if (sdk.Servers) {
    if (sdk.Servers.ServerMap) {
      params.push({
        Type: "SDKConfig.Server",
        Name: "server",
        Optional: true,
        Default: "null",
        Description: "The server to use from the predefined server list.",
      });
    } else {
      params.push({
        Type: "int",
        Name: "serverIndex",
        Optional: true,
        Default: "null",
        Description:
          "The index of the server to use from the predefined server list. Must be between 0 and the length of the server list. Defaults to 0 if not specified.",
      });
    }
  }

  // Server variables
  for (const variable of sdk.Servers.GetVariables()) {
    params.push({
      Type: `${sanitizeType(variable.Type)}`,
      Name: sanitizeMethodParamName(variable.Name),
      Optional: true,
      Default: "null",
      Description: extractParamDescription(
        variable.Server?.Comments,
        `Server variable for ${variable.Name}. This will replace the {${variable.Name}} placeholder in server URLs.`,
      ),
    });
  }

  params.push(
    ...[
      {
        Type: "string",
        Name: "serverUrl",
        Optional: true,
        Default: "null",
        Description:
          "A custom server URL to use instead of the predefined server list. If provided with urlParams, the URL will be templated with the provided parameters.",
      },
      {
        Type: "Dictionary<string, string>",
        Name: "urlParams",
        Optional: true,
        Default: "null",
        Description:
          "A dictionary of parameters to use for templating the serverUrl. Only used when serverUrl is provided.",
      },
      {
        Type: `I${getDefaultHttpClient()}`,
        Name: "client",
        Optional: true,
        Default: "null",
        Description: `A custom HTTP client implementation to use for making API requests. If not provided, the default ${getDefaultHttpClient()} will be used.`,
      },
      {
        Type: "RetryConfig",
        Name: "retryConfig",
        Optional: true,
        Default: "null",
        Description:
          "Configuration for retry behavior when API requests fail. Defines retry strategies, backoff policies, and maximum retry attempts.",
      },
    ],
  );

  return params;
}

// @ts-ignore
function templateSDKConstructor(sdk: SDK): string {
  let output = "\n\n";

  const sdkName = sanitizeSDKName(sdk.Type.Name);
  const args = getSDKArguments(sdk);
  const params = args
    .map(
      (arg) => `
${templateIndent(3)}${templateMethodArgument(arg, true)}`,
    )
    .join(",");

  output += templateCommentElement(
    "summary",
    null,
    "Initializes a new instance of the SDK with optional configuration parameters.",
    2,
  );

  output += "\n" + args.map((p) => templateParamComment(p, 2)).join("\n");

  if (sdk.Servers && !sdk.Servers.ServerMap) {
    output +=
      "\n" +
      templateExceptionComment(
        "ArgumentOutOfRangeException",
        `Invalid value provided for <paramref name="serverIndex"/>: must be between 0 (inclusive) and ${sdk.Servers.Servers.length} (exclusive).`,
        2,
      );
  }

  if (sdk.Security && !sdk.Security.Optional) {
    const secParam = getSecurityParam(sdk);
    output +=
      "\n" +
      templateExceptionComment(
        "ArgumentException",
        `None of <paramref name="${secParam.Name}"/> and <paramref name="${secParam.Source}"/> were provided.`,
        2,
      );
  }

  output += `
        public ${sdkName}(${params}
        )`;

  return output;
}
registerTemplateFunc("templateSDKConstructor", templateSDKConstructor);

// Helper function for extracting description from comments with fallback logic
function extractParamDescription(
  comments: CommentDef | null | undefined,
  fallbackDescription: string = "",
): string {
  // Priority 1: Use Summary if available
  if (comments?.Summary && comments.Summary.trim() != "") {
    return comments?.Summary;
  }

  // Priority 2: Use Description if available and no Summary
  if (comments?.Description && comments.Description.trim() != "") {
    return comments.Description;
  }

  // Priority 3: Use fallback if provided
  return fallbackDescription;
}

// @ts-ignore
function getFieldFallbackDescription(
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  switch (fieldDef.Type.Type.toString()) {
    case "class":
    case "enum":
      const className = sanitizeType(
        fieldDef.Type,
        false,
        usageLocation,
        "",
        fieldDef.Type.OutputLocation != usageLocation,
      );

      return `A <see cref="${className}"/> parameter.`;
  }

  return "";
}

// @ts-ignore
function getGlobalParametersToSet(request?: RequestDef): string[] {
  let params = [];

  if (
    !context.Global.AST.MainSDK.Globals ||
    context.Global.AST.MainSDK.Globals.Fields.length == 0 ||
    !request ||
    request.Field.Type.Fields.length == 0
  ) {
    return params;
  }

  for (const parameter of context.Global.AST.MainSDK.Globals.Fields) {
    for (const field of sortMethodParams(request.Field.Type.Fields)) {
      if (
        parameter.Name == field.Name &&
        parameter.Type.Type == field.Type.Type
      ) {
        params.push(sanitizeFieldName(field.Name));
      }
    }
  }

  return params;
}

registerTemplateFunc("getGlobalParametersToSet", getGlobalParametersToSet);

//@ts-ignore
function getDataTypeFormat(request: RequestDef): string {
  if (request.Field.Type.Format != "") {
    return request.Field.Type.Format;
  }

  switch (request.Field.Type.Type.toString()) {
    case "array":
    case "map":
      return request.Field.Type.ItemType.Format;
    default:
      return "";
  }
}

registerTemplateFunc("getDataTypeFormat", getDataTypeFormat);

// @ts-ignore
function isPrimitive(type: TypeDef): boolean {
  const primitives = [
    "boolean",
    "integer",
    "int32",
    "number",
    "float32",
    "boolean",
    "string",
  ];
  if (primitives.includes(type.Type.toString())) {
    return true;
  }
  return false;
}

registerTemplateFunc("isPrimitive", isPrimitive);

// @ts-ignore
function primitiveConvertFunction(type: TypeDef): string {
  switch (type.Type.toString()) {
    case "integer":
      return "ToInt64";
    case "int32":
      return "ToInt32";
    case "number":
      return "ToDouble";
    case "float32":
      return "ToSingle";
    case "boolean":
      return "ToBoolean";
  }
}

registerTemplateFunc("primitiveConvertFunction", primitiveConvertFunction);

// @ts-ignore
function templateZero(
  typeDef: TypeDef,
  optional: boolean,
  scope: string,
  explicitNumberTypes = false,
): string {
  if (optional) {
    return "nil";
  }

  switch (typeDef.Type.toString()) {
    case "string":
      return '""';
    case "date":
      if (useNodatime()) {
        return "DateTime.Today";
      } else {
        addUsageImport("System");
        return "DateOnly.FromDateTime(DateTime.Today)";
      }
    case "date-time":
      return "DateTime.Now";
    case "float32":
      return "0.0f";
    case "integer":
      return "0L";
    case "int32":
      return "0";
    case "bigint":
      addUsageImport("System.Numerics");
      return 'BigInteger.Parse("0")';
    case "decimal":
      return `0.0M`;
    case "number":
      return "0.0f";
    case "boolean":
      return "false";
    case "any":
      return "null";
    case "enum":
      return `${sanitizeType(
        typeDef,
        false,
        getScopePath(scope),
      )}(${templateZero(typeDef.Enum.Type, false, scope)})`;
    case "response":
      return "null";
  }
  return `new ${sanitizeType(typeDef, false, getScopePath(scope))}()`;
}

registerTemplateFunc("templateZero", templateZero);

// @ts-ignore
function templateDefaultValue(fieldDef: FieldDef): string {
  let defaultValue = "";

  if (fieldDef.Const || fieldDef.Default) {
    defaultValue = `${templateConstOrDefaultValue(fieldDef)}`;
  } else if (!fieldDef.Optional && !fieldDef.Nullable) {
    defaultValue = fieldDef.IsResponseHeaders
      ? "new Dictionary<string, List<string>>()"
      : "default!";
  } else if (fieldDef.Optional && fieldDef.Nullable) {
    defaultValue = "null";
  }
  return defaultValue != "" ? ` = ${defaultValue};` : "";
}
registerTemplateFunc("templateDefaultValue", templateDefaultValue);

// @ts-ignore
function getProjectIdentifier(): string {
  return caser()
    .ToKebab(context.Global.Config.PackageName.split(".").join("-"))
    .toLowerCase();
}

// @ts-ignore
function templateReadmeTitle(): string {
  return getPackageId();
}
unregisterTemplateFunc("templateReadmeTitle");
registerTemplateFunc("templateReadmeTitle", templateReadmeTitle);

// @ts-ignore
function templateReadmeFieldName(fieldDef: FieldDef): string {
  return sanitizePrivateFieldName(fieldDef.Name);
}
unregisterTemplateFunc("templateReadmeFieldName");
registerTemplateFunc("templateReadmeFieldName", templateReadmeFieldName);

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest || additionalContext?.usageContext?.Test) {
    // Tests run from Tests/bin/Debug/net8.0, need to go up to project root for .speakeasy/testfiles/
    filePath = `../../../../.speakeasy/testfiles/${filePath}`;
  }
  return `System.IO.File.ReadAllBytes("${filePath}")`;
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest || additionalContext?.usageContext?.Test) {
    // Tests run from Tests/bin/Debug/net8.0, need to go up to project root for .speakeasy/testfiles/
    filePath = `../../../../.speakeasy/testfiles/${filePath}`;
  }
  return `System.IO.File.ReadAllText("${filePath}")`;
}

// @ts-ignore
function templateAllowEmptyValue(op: Operation): string {
  const paramNames = getAllowEmptyValueQueryParamNames(op);

  if (paramNames.length === 0) {
    return "null";
  }

  return `new List<string> { ${paramNames
    .map((name) => `"${name}"`)
    .join(", ")} }`;
}

registerTemplateFunc("templateAllowEmptyValue", templateAllowEmptyValue);
