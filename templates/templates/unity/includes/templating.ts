// @ts-ignore
function templateConstName(name: string, prefix: string) {
  return caser().ToPascal(`${prefix}${name}`);
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
    if (description) {
      lines.push(...formatCSharpEnumDescriptionComments(description));
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

function formatCSharpEnumDescriptionComments(description: string): string[] {
  const summaryLines: string[] = ["/// <summary>"];

  description
    .split(/\r?\n/)
    .map((segment) => segment.trim())
    .filter((segment) => segment.length > 0)
    .map((segment) => sanitizeComments(segment).replace(/<br\/>/g, ""))
    .forEach((segment) => {
      summaryLines.push(`/// ${segment}`);
    });

  summaryLines.push("/// </summary>");
  return summaryLines;
}

// @ts-ignore
function templateMethodArguments(
  operation: Operation,
  usageLocation: string,
): string {
  const paramFlattening = operationParametersFlattened(operation);
  const args = [];

  const buildArgString = (
    type: string,
    optional: boolean,
    fieldName: string,
  ) => {
    return `${type}${optional ? "?" : ""} ${fieldName}${
      optional ? " = null" : ""
    }`;
  };

  if (paramFlattening) {
    sortMethodParams(operation.Request.Field.Type.Fields).forEach((field) => {
      var typeName = sanitizeType(field.Type, field.Optional, usageLocation);

      args.push(
        buildArgString(
          typeName,
          field.Optional || field.Nullable,
          sanitizeMethodParamName(field.Name),
        ),
      );
    });
  } else {
    if (operation.Request) {
      var typeName = sanitizeType(
        operation.Request.Field.Type,
        !operation.Request.IsRequestBodyRequired,
        usageLocation,
      );

      args.push(
        buildArgString(
          typeName,
          isRequestOptional(operation.Request),
          "request",
        ),
      );
    }
  }

  if (operation.Security) {
    args.unshift(
      buildArgString(
        sanitizeType(operation.Security.Type, false, usageLocation),
        false,
        "security",
      ),
    );
  }

  if (operation.Servers) {
    args.push(`string? serverUrl = null`);
  }

  return args.join(", ");
}

registerTemplateFunc("templateMethodArguments", templateMethodArguments);

// @ts-ignore
function joinParams(params: string[]): string {
  switch (params.length) {
    case 0:
      return "";
    case 1:
      return params[0];
    default:
      return `\n${"    ".repeat(1)}` + params.join(`,\n${"    ".repeat(1)}`);
  }
}

// @ts-ignore
function templateUsageSDKParams(local: UsageContext): string {
  const params = [];
  let hasSecurity = local.Operation.Security != undefined;

  const ctx: TemplateValueContext = {
    usageContext: local, // TODO see if this can be decomposed into the required fields below
    operation: local.Operation,
    isTest: local.Test && true,
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
                params.push(globalSecurity);
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            const hasOperationServers =
              local.Operation?.Servers?.Servers?.length > 0;
            if (!hasOperationServers) {
              params.push(`serverUrl: ${getUsageServerUrl(local, scope)}`);
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              params.push(`server: "${server.ID}"`);
            } else if (server.Index !== undefined) {
              params.push(`serverIndex: ${server.Index}`);
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const fieldName = sanitizePrivateFieldName(v.Name);
                params.push(
                  `${fieldName}: "${getUsageServerVariableValue(v)}"`,
                );
              });
            }

            break;
          }
          case "parameter": {
            const parameter = scope.Value as ParameterUsage;

            const value = templateValueAsRequired(
              parameter.field,
              getExampleValue(parameter.example),
              ctx,
            );

            if (value === "") {
              return;
            }

            params.push(
              `${sanitizePrivateFieldName(parameter.field.Name)}: ${value}`,
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
      params.push(globalSecurity);
    }
  }

  return joinParams(params);
}

registerTemplateFunc("templateUsageSDKParams", templateUsageSDKParams);

// @ts-ignore
function getGlobalSecurity(options?: SecurityUsageContext | Example): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example);
    }
  }

  if (
    index === undefined &&
    example === undefined &&
    context.Global.AST.MainSDK.Security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme"
  ) {
    return "";
  }

  let securityFieldName = "security";
  const security = context.Global.AST.MainSDK.Security;

  if (
    security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    securityFieldName = sanitizeSecurityFieldName(security.Type.Fields[0].Name);
  }

  return `${securityFieldName}: ${templateSecurityUsage(
    security.Type,
    1,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
  ).trim()}`;
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
): string {
  const operation = local.Operation;
  const methodParams = [];

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
  };

  if (operation.Security) {
    methodParams.push(
      `security: ${templateUnflattenedSecurityUsage(
        local.Operation.Security.Type,
        indent + 1,
        true,
        0,
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

        methodParams.push(`${sanitizePrivateFieldName(field.Name)}: ${value}`);
      }
    }
  } else {
    if (operation.Request) {
      methodParams.push("req");
    }

    methodParams.push(...getOptionalUsageMethodParameters(local, indent));
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
  return `${sanitizeFieldName(fieldDef.Name)} = `;
}

// @ts-ignore
function templateMethodArgumentDeclaration(fieldDef: FieldDef): string {
  return `${sanitizePrivateFieldName(fieldDef.Name)}: `;
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
  let type = sanitizeType(typeDef, false, "");
  type = type.replaceAll("Map<", "Dictionary<");
  return `new ${type}() `;
}

// @ts-ignore
function templateIntValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${val}`;
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
  return `System.Text.Encoding.UTF8.GetBytes("${val.replace(/"/g, '\\"')}")`;
}

// @ts-ignore
function templateDateValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `DateOnly.FromDateTime(System.DateTime.Parse("${val}"))`;
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `System.DateTime.Parse("${val}")`;
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
  if (t.Enum?.Names.length > 0) {
    return t.Enum.Names.map((n) => getEnumName(n));
  } else {
    return getEnumNamesFromValues(t.Enum?.Values);
  }
}

// @ts-ignore
function templateEnumValue(fieldDef: FieldDef, idx: number): string {
  const enumNames = getEnumNames(fieldDef.Type);
  let prefix = "";
  if (doesTypeConflict(fieldDef.Type)) {
    prefix = `${getModelNamespace(fieldDef.Type.OutputLocation, false)}.`;
  }

  return `${prefix}${sanitizeClassName(fieldDef.Type.Name, false)}.${
    enumNames[idx]
  }`;
}

// @ts-ignore
function templateStringValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const str = `"${val.replaceAll(/"/g, '\\"').replaceAll("{{", `{{"{{"}}`)}"`;
  return str.includes("\n") ? `@${str}` : str;
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
  return `{ "${key}", ${val} }`;
}

// @ts-ignore
function templateArrayValue(val: any): string {
  return val;
}

// @ts-ignore
const commentSummaryStart = "/// <summary>";
// @ts-ignore
const commentSummaryEnd = "/// </summary>";
// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  indent: number,
  type: "method" | "field" | "class" | "enum" = null,
  addLeadingNewLine: boolean = true,
  addTrailingNewLine: boolean = false,
): string {
  if (!comments) {
    return addTrailingNewLine ? "\n" : "";
  }

  let lines = [];

  let firstLineParts = [];

  let descriptionLines = [];
  if (comments.Description) {
    descriptionLines = sanitizeComments(comments.Description).split("\n");
  }

  if (comments.Summary) {
    firstLineParts.push(sanitizeComments(comments.Summary));
  } else if (descriptionLines.length > 0) {
    firstLineParts.push(descriptionLines.shift());
  }

  const firstLine = firstLineParts.join(" - ");

  if (firstLine) {
    lines.push(...firstLine.split("\n"));
  }

  if (descriptionLines.length > 0) {
    lines.push("");
    lines.push("<remarks>");
    lines.push(...descriptionLines);
    lines.push("</remarks>");
  }

  if (comments.ExternalDocs) {
    let externalDocsLines = [];

    if (comments.ExternalDocs.Description) {
      externalDocsLines = sanitizeComments(
        comments.ExternalDocs.Description,
      ).split("\n");
    }

    lines.push("");
    lines.push(
      `<see>${comments.ExternalDocs.URL}}${
        externalDocsLines.length > 0 ? " - " + externalDocsLines.shift() : ""
      }</see>`,
    );
    lines.push(...externalDocsLines);
  }

  if (lines.length == 0) {
    return addTrailingNewLine ? "\n" : "";
  }

  lines = lines.map((line) => {
    return `/// ${line}`;
  });

  lines.unshift(commentSummaryStart);
  lines.push(commentSummaryEnd);

  return (addLeadingNewLine ? "\n" : "") + "\n" + indentLines(lines, indent);
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function templateSDKConstructor(sdk: SDK): string {
  var parts: string[] = [];

  if (sdk.Security) {
    const secParam = getSecurityParam(sdk);
    parts.push(`${secParam.Type}? ${secParam.Name} = null`);
    parts.push(`Func<${secParam.Type}>? ${secParam.Source} = null`);
  }

  if (sdk.Globals) {
    for (const parameter of sdk.Globals.Fields) {
      parts.push(
        `${sanitizeType(parameter.Type, false, "")}? ${sanitizePrivateFieldName(
          parameter.Name,
        )} = null`,
      );
    }
  }

  if (sdk.Servers) {
    if (sdk.Servers.ServerMap) {
      parts.push("SDKConfig.Server? server = null");
    } else {
      parts.push("int? serverIndex = null");
    }
    const serverVariables = sdk.Servers.GetVariables();
    for (const variable of serverVariables) {
      if (variable.Type.Type == "enum") {
        parts.push(
          `${sanitizeClassName(variable.Type.Name)}? ${sanitizePrivateFieldName(
            variable.Name,
          )} = null`,
        );
      } else {
        parts.push(
          `${sanitizeType(variable.Type, false)}?  ${sanitizePrivateFieldName(
            variable.Name,
          )} = null`,
        );
      }
    }
  }

  parts.push("string? serverUrl = null");
  parts.push("Dictionary<string, string>? urlParams = null");

  parts.push("ISpeakeasyHttpClient? client = null");

  return parts.join(", ");
}

registerTemplateFunc("templateSDKConstructor", templateSDKConstructor);

// @ts-ignore
function templateLanguageName(lang: string): string {
  return "csharp";
}
unregisterTemplateFunc("templateLanguageName");
registerTemplateFunc("templateLanguageName", templateLanguageName);

function templateRequestMethod(method: string): string {
  method = method.toUpperCase();

  // Unity for some reason doesn't have built in constants for some HTTP methods so we need to workaround that
  switch (method) {
    case "PATCH":
    case "DELETE":
    case "TRACE":
    case "OPTIONS":
      return `"${method}"`;
    default:
      return `UnityWebRequest.kHttpVerb${method}`;
  }
}
registerTemplateFunc("templateRequestMethod", templateRequestMethod);

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
      return "DateTime.Today";
    case "date-time":
      return "DateTime.Now";
    case "float32":
      return "0.0f";
    case "integer":
      return "0L";
    case "int32":
      return "0";
    case "bigint":
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
      return `${sanitizeType(typeDef, false, scope)}(${templateZero(
        typeDef.Enum.Type,
        false,
        scope,
      )})`;
    case "response":
      return "null";
  }
  return `new ${sanitizeType(typeDef, false, scope)}()`;
}

registerTemplateFunc("templateZero", templateZero);

// @ts-ignore
function templateAccessOperator(fieldDef: FieldDef): string {
  const optional = fieldDef.Optional || fieldDef.Nullable;
  return optional ? "?." : ".";
}

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
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  return `File.ReadAllBytes("${filePath}")`;
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  return `File.ReadAllText("${filePath}")`;
}
