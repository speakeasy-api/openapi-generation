// @ts-ignore
function getProjectIdentifier(): string {
  return sanitizeSDKPackageName(false);
}

// @ts-ignore
function templateReadmeTitle(): string {
  return getProjectIdentifier();
}
unregisterTemplateFunc("templateReadmeTitle");
registerTemplateFunc("templateReadmeTitle", templateReadmeTitle);

// @ts-ignore
function templateSDKOptionName(name: string): string {
  const opt = getOptionNames();
  let reservedOptions: string[] = [
    opt.GlobalServerURL,
    opt.GlobalTemplatedServerURL,
    opt.GlobalClient,
  ];

  if (context.Global.AST.MainSDK.Security) {
    reservedOptions.push(opt.GlobalSecurity);
  }
  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reservedOptions.push(opt.GlobalServer);
    reservedOptions.push("WithTemplatedServer");
  } else {
    reservedOptions.push(opt.GlobalServerIndex);
  }

  let optionName = `With${sanitizeClassName(name)}`;

  if (reservedOptions.includes(optionName)) {
    optionName = `WithGlobal${sanitizeClassName(name)}`;
  }

  return optionName;
}

registerTemplateFunc("templateSDKOptionName", templateSDKOptionName);

// @ts-ignore
function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example, additionalContext);
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

  const secUsage = templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    2,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    example,
    additionalContext,
  );
  return `${sanitizeSDKPackageName(true)}.WithSecurity(${secUsage.trim()}),`;
}

// @ts-ignore
function joinSDKOptions(options: string[]): string {
  if (options.length > 0) {
    options.unshift("");

    return options.join(`\n${"    ".repeat(2)}`) + "\n    ";
  }

  return "";
}

// @ts-ignore
function templateUsageSDKOptions(local: UsageContext): string {
  const options = [];
  let hasSecurity = local.Operation.Security != undefined;
  let hasAbsoluteServerURL =
    context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false;

  const hasOperationServers = local.Operation?.Servers?.Servers?.length > 0;

  const hasAnyOperationServers =
    context.Global.AST.MainSDK.HasAnyOperationServers() || false;

  if (!hasAbsoluteServerURL && !hasOperationServers) {
    const url = getUsageServerUrl(local, undefined, {
      isTest: local.Test && true,
    });
    if (hasAbsoluteServerURL || hasAnyOperationServers) {
      // Constructor takes opts only — use WithServerURL option
      options.push(`${sanitizeSDKPackageName(true)}.WithServerURL(${url}),`);
    } else {
      // Constructor requires positional serverURL — pass raw value
      options.push(`${url},`);
    }
  }

  const ctx: TemplateValueContext = {
    usageContext: local, // TODO see if this can be decomposed into the required fields below
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
  };

  local.Scopes?.forEach((scope) => {
    if (scope.IsGlobal && scope.Feature != "") {
      const feature = scope.Feature.toString();

      switch (feature) {
        case "security":
          if (!hasSecurity) {
            const globalSecurity = getGlobalSecurity(scope.Value, ctx);
            if (globalSecurity) {
              options.push(globalSecurity);
              hasSecurity = true;
            }
          }
          break;
        case "server_url": {
          if (hasAbsoluteServerURL && !hasOperationServers) {
            const url = getUsageServerUrl(local, scope, ctx);
            options.push(
              `${sanitizeSDKPackageName(true)}.WithServerURL(${url}),`,
            );
          }
          break;
        }
        case "server_selection": {
          const sdkName = sanitizeSDKPackageName(true);
          const server: UsageGlobalServer = getUsageGlobalServer(!!scope.Value);

          if (server.ID) {
            options.push(`${sdkName}.WithServer("${server.ID}"),`);
          } else if (server.Index !== undefined) {
            options.push(`${sdkName}.WithServerIndex(${server.Index}),`);
          }

          if (server.Variables) {
            server.Variables.forEach((v: ServerVariable) => {
              options.push(
                `${sdkName}.${templateSDKOptionName(
                  v.Name,
                )}("${getUsageServerVariableValue(v)}"),`,
              );
            });
          }

          break;
        }
        case "retries": {
          options.push(
            `${sanitizeSDKPackageName(true)}.WithRetryConfig(${templateString(
              "usage/retries.stmpl",
              {},
            )}),`,
          );
          break;
        }
        case "http_client": {
          options.push(
            `${sanitizeSDKPackageName(true)}.WithClient(${templateHTTPClient(
              scope.Value,
            )}),`,
          );
          break;
        }
        case "parameter": {
          const parameter = scope.Value as ParameterUsage;

          const value = templateValueAsRequired(
            parameter.field,
            getExampleValue(parameter.example),
            {
              ...ctx,
            },
          );

          if (value === "") {
            return;
          }

          options.push(
            `${sanitizeSDKPackageName(true)}.${templateSDKOptionName(
              parameter.field.Name,
            )}(${value}),`,
          );

          break;
        }
      }
    }
  });

  if (!hasSecurity && opUsesGlobalSecurity(local.Operation)) {
    const globalSecurity = getGlobalUsageSecurity(local);
    if (globalSecurity) {
      options.push(globalSecurity);
    }

    if (context.Global.Config.FlattenGlobalSecurity == false) {
      addModelImport(context.Global.AST.MainSDK.Security.Type.OutputLocation);
    }
  }

  return joinSDKOptions(options);
}

registerTemplateFunc("templateUsageSDKOptions", templateUsageSDKOptions);

// @ts-ignore
function templateCmsComment(
  sdkComments: SimpleCommentDef,
  title: string,
  indent: number,
  keepTitle: boolean = false,
): string {
  let lines = [];

  let firstLineParts: string[] = [];

  if (title !== null) {
    if (/[.]/.test(title)) {
      title = title.replace(/^[^.]*[.]/, "");
    }

    if (title && !/\[/.test(title)) {
      firstLineParts.push(title);
    }
  }

  if (
    (!keepTitle || title === null) &&
    !sdkComments.Description &&
    !sdkComments.Summary
  ) {
    return "";
  }

  let descriptionLines = [];
  if (sdkComments.Description) {
    descriptionLines = sdkComments.Description.split("\n");
  }

  if (sdkComments.Summary) {
    firstLineParts.push(sdkComments.Summary);
  } else if (descriptionLines.length > 0) {
    firstLineParts.push(descriptionLines.shift());
  }

  let firstLine: string;
  if (firstLineParts.length > 1) {
    const words = firstLineParts[1].split(" ");
    // If the first couple words are the title, we don't need to add the title.
    if (
      (words.length > 0 && words[0] === title) ||
      (words.length > 1 && words[1] === title)
    ) {
      firstLine = firstLineParts[1];
    } else {
      firstLine = firstLineParts.join(" - ");
    }
  } else if (firstLineParts.length == 1) {
    firstLine = firstLineParts[0];
  }

  if (firstLine) {
    lines.push(...sanitizeComments(firstLine).split("\n"));
  }

  if (descriptionLines.length > 0) {
    lines.push(...descriptionLines.map(sanitizeComments));
  }

  if (lines.length == 0) {
    return "";
  }

  lines = lines.map((line) => `// ${line}`);

  return indentLines(lines, indent) + "\n";
}

// When op-level security was hoisted and it is a strict subset of global security,
// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map(
    (f) => `[Security.${sanitizeFieldName(f.Name)}]`,
  );

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set via [WithSecurity].`
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
function templateComments(
  comments: CommentDef | null,
  docgroup: string,
  title: string | null,
  indent: number,
  type: "method" | "field" | "type" | "const" | "var" | null = null,
  omitExternalDocs: boolean = false,
  additionalNotes: string = "",
): string {
  const keepTitle =
    comments &&
    (comments.Deprecated || (comments.ExternalDocs && !omitExternalDocs));
  const resolved = resolveCommentDef(
    "openapi",
    docgroup,
    type ?? "",
    title,
    comments,
  );
  const cmsTitle =
    type === "field" || type === "const" || type === "var" ? null : title;
  const cmsComment = templateCmsComment(resolved, cmsTitle, indent, keepTitle);

  if (!comments && !additionalNotes) {
    return cmsComment;
  }

  let lines = [];

  if (comments?.ExternalDocs && !omitExternalDocs) {
    if (cmsComment) {
      lines.push("");
    }

    let externalDocs = comments.ExternalDocs.URL;
    if (comments.ExternalDocs.Description) {
      externalDocs += ` - ${sanitizeComments(
        comments.ExternalDocs.Description,
      )}`;
    }

    lines.push(...sanitizeComments(externalDocs).split("\n"));
  }

  if (additionalNotes) {
    if (lines.length > 0 || cmsComment) {
      lines.push("");
    }

    lines.push(...additionalNotes.split("\n"));
  }

  if (comments?.Deprecated) {
    if (lines.length > 0 || cmsComment) {
      lines.push("");
    }

    let deprecated = `Deprecated: ${comments.DeprecationMessage}.`;

    if (comments.DeprecationReplacement) {
      const replacement = sanitizeDeprecationReplacement(
        comments.DeprecationReplacement,
        type as any,
      );
      if (replacement) {
        deprecated += ` Use ${replacement} instead.`;
      }
    }

    lines.push(...deprecated.split("\n"));
  }

  if (lines.length == 0) {
    return cmsComment;
  }

  lines = lines.map((line) => `// ${line}`);

  return cmsComment + indentLines(lines, indent) + "\n";
}

registerTemplateFunc("templateComments", templateComments);

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
    title + " " + description,
  );
  return templateCmsComment(resolved, title, indent);
}

registerTemplateFunc("templateBuiltinComment", templateBuiltinComment);

function getAlternateParameterDeclaration(
  field: FieldDef,
  value: string,
): string {
  const fieldName = sanitizePrivateFieldName(field.Name);

  switch (field.Type.Type.toString()) {
    case "request-stream":
      if (value.includes("os.")) {
        return `${fieldName}, fileErr := `;
      }
    case "string":
    case "bytes":
      if (value.includes("os.")) {
        return `${fieldName}, fileErr := `;
      }
  }

  return "";
}

function addAdditionalParameterCode(
  usageContext: UsageContext,
  field: FieldDef,
) {
  switch (field.Type.Type.toString()) {
    case "request-stream":
      if (usageContext.Test) {
        return `\nrequire.NoError(t, fileErr)\n`;
      } else {
        return `\nif fileErr != nil {
${templateIndent(1)}panic(fileErr)
}\n`;
      }
    default:
      return "";
  }
}

// @ts-ignore
function getOptionalUsageMethodParameters(local: UsageContext): string[] {
  let optionalParams = [];

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal) {
        return;
      }

      switch (scope.Feature.toString()) {
        case "server_url": {
          addImport("operations");
          const url = getUsageServerUrl(local, scope, {
            isTest: local.Test && true,
          });
          const optionName = getGoCommonControlOptionName(
            local.Operation,
            "ServerURL",
          );
          optionalParams.push(
            `${getAccessNamespace("usage", "operations")}${optionName}(${url})`,
          );
          break;
        }
        case "content_type": {
          addImport("operations");
          const acceptHeaderOption = sanitizeAcceptEnumKey(scope.OpFilter);
          const operationAccessNamespace = getAccessNamespace(
            "usage",
            "operations",
          );
          const optionName = getGoCommonControlOptionName(
            local.Operation,
            "AcceptHeaderOverride",
          );
          optionalParams.push(
            `${operationAccessNamespace}${optionName}(${operationAccessNamespace}${acceptHeaderOption})`,
          );
          break;
        }
      }
    });
  }

  return optionalParams;
}

const localDeclarationRegex = /\{-\{-\{([^:]+):([\s\S]*?)\}-\}-\}/g;

// @ts-ignore
function templateUsageMethodParameters(
  local: UsageContext,
  indent: number,
): {
  MethodParams: string;
  LocalDeclarations: string;
} {
  const operation = local.Operation;

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
    isRequest: true,
  };

  let methodParams = [];

  const args = getArguments(operation);
  for (const field of args.merged) {
    const argumentOption = getGoMethodArgumentOption(operation, field);
    if (isSecurityClassField(field)) {
      let securityScope = undefined;

      for (const scope of local.Scopes) {
        if (scope.Feature == "security") {
          securityScope = scope;
          break;
        }
      }

      let securityExample = undefined;
      if (securityScope?.Value) {
        if ("example" in securityScope.Value) {
          securityExample = securityScope.Value.example;
        } else {
          securityExample = getExampleValue(securityScope.Value, ctx);
        }
      }

      const excludeField = securityExample === undefined && field.Optional;

      if (excludeField) {
        if (!argumentOption) {
          methodParams.push("nil");
        }
      } else {
        let secUsage = templateSecurityUsage(
          local.Operation.Security.Type,
          0,
          true,
          false,
          undefined,
          securityExample,
          ctx,
        );
        if (argumentOption) {
          addImport("operations");
          methodParams.push(
            `${getAccessNamespace("usage", "operations")}${
              argumentOption.name
            }(${secUsage})`,
          );
        } else {
          if (local.Operation.Security.Optional) {
            secUsage = `${sanitizeSDKPackageName(true)}.Pointer(${secUsage})`;
          }
          methodParams.push(secUsage);
        }
      }
    } else {
      const fieldExample = getOperationMethodFieldExample(local, field);

      if (argumentOption) {
        if (fieldExample === undefined) {
          continue;
        }
        const value = templateValue(
          {
            ...field,
            Optional: false,
            Nullable: false,
          } as FieldDef,
          fieldExample,
          false,
          ctx,
        );
        if (value == "") {
          continue;
        }
        addImport("operations");
        const optionValue = field.Nullable
          ? fieldExample === null
            ? "nil"
            : `${sanitizeSDKPackageName(true)}.Pointer(${value})`
          : value;
        methodParams.push(
          `${getAccessNamespace("usage", "operations")}${
            argumentOption.name
          }(${optionValue})`,
        );
      } else {
        const value = templateValue(field, fieldExample, false, ctx);

        if (value == "") {
          continue;
        }

        methodParams.push(value);
      }
    }
  }

  const optionals = getOptionalUsageMethodParameters(local);

  methodParams = methodParams.concat(optionals);

  let result = indentLines([methodParams.join(", ")], indent).trimStart();

  if (!result) {
    return {
      MethodParams: "",
      LocalDeclarations: "",
    };
  }

  result = `, ${result}`;

  let localDecs = [];

  // Replace patterns like {-{-{someFieldName: some code}-}-} with just the field name
  result = result.replace(
    localDeclarationRegex,
    (match, fieldName, content) => {
      localDecs.push({
        fieldName: fieldName.trim(),
        content: content.trim(),
      });

      return fieldName.trim();
    },
  );

  // Get unique declarations and replace escaped newlines and spaces with actual newlines and spaces
  let localDeclarations = "";

  if (localDecs.length > 0) {
    localDeclarations =
      "\n" +
      [
        ...new Set(
          localDecs.map((match) =>
            match.content.replaceAll("\\n", "\n").replaceAll("\\s", " "),
          ),
        ),
      ].join("\n\n") +
      "\n";
  }

  return {
    MethodParams: result,
    LocalDeclarations: localDeclarations,
  };
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

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
      return "types.Date{}";
    case "date-time":
      return "time.Time{}";
    case "float32":
      return explicitNumberTypes ? "float32(0)" : "0.0";
    case "integer":
      return explicitNumberTypes ? "int64(0)" : "0";
    case "int32":
      return "0";
    case "bigint":
      return "big.NewInt(0)";
    case "decimal":
      return `new(decimal.Big).SetFloat64(0.0)`;
    case "number":
      return explicitNumberTypes ? "float64(0)" : "0.0";
    case "boolean":
      return "false";
    case "any":
      return "nil";
    case "enum":
      return `${sanitizeType(typeDef, false, scope)}(${templateZero(
        typeDef.Enum.Type,
        false,
        scope,
      )})`;
    case "request":
    case "request-stream":
    case "response":
    case "response-stream":
      return "nil";
    case "event-stream":
      return `&stream.EventStream[${sanitizeType(
        typeDef.ItemType,
        false,
        scope,
      )}]{}`;
    case "jsonl":
      return `&jsonl.JsonLStream[${sanitizeType(
        typeDef.ItemType,
        false,
        scope,
      )}]{}`;
  }
  return sanitizeType(typeDef, false, scope) + "{}";
}
registerTemplateFunc("templateZero", templateZero);

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "nil";
  }

  return `[]string{${scopes.map((s) => `"${s}"`).join(",")}}`;
}
registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

// @ts-ignore
function templateGlobalParamAccess(param: FieldDef): string {
  return `s.sdkConfiguration.Globals.${sanitizeFieldName(param.Name)}`;
}
registerTemplateFunc("templateGlobalParamAccess", templateGlobalParamAccess);

// @ts-ignore
function templateErrorMessage(errorType: TypeDef, errorObject: string): string {
  const errorMessageField = getErrorMessageField(errorType);

  if (errorMessageField) {
    // Use sanitizeErrorFieldName only when the type is in the errors scope,
    // where fields are rendered with that sanitizer. Types that remain in the
    // components/shared scope use sanitizeFieldName instead (e.g. when a union
    // error variant references a component type that couldn't be moved).
    const fieldName = sanitizeScopedFieldName(
      errorMessageField.Name,
      errorType,
    );

    if (errorMessageField.Optional) {
      return `if ${errorObject}.${fieldName} == nil {
    return "unknown error"
}

return *${errorObject}.${fieldName}`;
    } else {
      return `return ${errorObject}.${fieldName}`;
    }
  } else {
    addImport("encoding/json");
    return `data, _ := json.Marshal(${errorObject})
return string(data)`;
  }
}
registerTemplateFunc("templateErrorMessage", templateErrorMessage);

// @ts-ignore
function templateStream(
  fieldDef: FieldDef,
  filePath: string,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (filePath) {
    if (additionalContext?.isRequest) {
      return templateLocalFileDeclaration(
        fieldDef,
        filePath,
        example,
        additionalContext,
      );
    }

    if (additionalContext?.isTest) {
      filePath = `../.speakeasy/testfiles/${filePath}`;
    }

    if (additionalContext?.isResponse) {
      return `readFileToBytes("${filePath}")`;
    }

    addImport("os");
    return `os.Open("${filePath}")`;
  } else {
    if (typeof example !== "string") {
      example = JSON.stringify(example);
    }
    const byteValue = templateByteValue(example, fieldDef, additionalContext);

    if (additionalContext?.isResponse) {
      return byteValue;
    } else {
      addImport("bytes");
      return `bytes.NewBuffer(${byteValue})`;
    }
  }
}

// Will template a special placeholder for a local file read declaration that will be pulled out and formatted later
function templateLocalFileDeclaration(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
  templateFieldAccess?: (field) => string,
): string {
  // fieldName will be the file name before the extension formatted as a valid Go identifier
  const fieldName = sanitizePrivateFieldName(
    filePath.split("/").pop().split(".")[0],
  );

  const declaration = `${fieldName}, fileErr := `;

  const value = templateValue(fieldDef, example, false, {
    ...additionalContext,
    isRequest: false,
  });

  let errorHandlingCode = "";

  if (additionalContext?.isTest) {
    errorHandlingCode = `\\nrequire.NoError(t, fileErr)`;
  } else {
    errorHandlingCode = `\\nif fileErr != nil {\\n${templateIndent(
      1,
    ).replaceAll(" ", "\\s")}panic(fileErr)\\n}`;
  }

  return templateSpecialLocalDeclarationValue(
    templateFieldAccess ? templateFieldAccess(fieldName) : fieldName,
    `${declaration}${value}${errorHandlingCode}`,
  );
}

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    filePath = `../.speakeasy/testfiles/${filePath}`;

    return `readFileToBytes("${filePath}")`;
  }

  if (additionalContext?.isRequest) {
    return templateLocalFileDeclaration(
      fieldDef,
      filePath,
      example,
      additionalContext,
    );
  }

  addImport("os");
  return `os.ReadFile("${filePath}")`;
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    filePath = `../.speakeasy/testfiles/${filePath}`;

    return `readFileToString("${filePath}")`;
  }

  if (additionalContext?.isRequest) {
    return templateLocalFileDeclaration(
      fieldDef,
      filePath,
      example,
      additionalContext,
      (fieldName) => `string(${fieldName})`,
    );
  }

  addImport("os");
  return `os.ReadFile("${filePath}")`;
}

// @ts-ignore
function templateReadmeGlobalValue(fieldDef: FieldDef): string {
  return templateValueAsRequired(fieldDef);
}
unregisterTemplateFunc("templateReadmeGlobalValue");
registerTemplateFunc("templateReadmeGlobalValue", templateReadmeGlobalValue);

// @ts-ignore
function templateHTTPClient(value: HTTPClientUsage): string {
  switch (value.type) {
    case "test":
      return `testHTTPClient`;
    default:
      throw new Error(`unsupported http client type: ${value.type}`);
  }
}

// @ts-ignore
function templateExampleReferenceValue(
  exampleReferenceValue: ExampleReferenceValue,
  targetField?: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const isTargetPointer = targetField.Optional || targetField.Nullable;
  const isSourcePointer =
    exampleReferenceValue.source.Optional ||
    exampleReferenceValue.source.Nullable;

  let value = exampleReferenceValue.path;

  if (isTargetPointer && !isSourcePointer) {
    addSDKPackageImport(true);

    value = `${sanitizeSDKPackageName(true)}.Pointer(${value})`;
  } else if (!isTargetPointer && isSourcePointer) {
    value = `*${value}`;
  }

  if (!exampleReferenceValue.replacements.length) {
    return value;
  }

  const fieldName = sanitizePrivateFieldName(targetField.Name);

  const declaration = getFieldDeclartion(targetField, fieldName);

  let additionalCode = "";

  for (const replacement of exampleReferenceValue.replacements) {
    const path = templateJSONPointerPath(
      targetField,
      fieldName,
      replacement.Path,
    );

    const val = templateValue(
      path.target,
      getExampleValue(replacement.Value, additionalContext),
      false,
      additionalContext,
    );

    if (val === "") {
      continue;
    }

    additionalCode += `\\n${path.path} = ${val}`;
  }

  return templateSpecialLocalDeclarationValue(
    fieldName,
    `${declaration}${value}${additionalCode}`,
  );
}

// @ts-ignore
function templateOmittedValue(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Optional || fieldDef.Nullable) {
    return "nil";
  }

  return "";
}

function getFieldDeclartion(field: FieldDef, fieldName: string): string {
  let declaration = `${fieldName} := `;
  const needsType =
    field.Optional ||
    field.Nullable ||
    !["class", "map", "array", "set"].includes(field.Type.Type.toString());

  if (needsType) {
    const fieldType = sanitizeType(
      field.Type,
      field.Optional || field.Nullable,
      "usage", // forces type declaration to be fully qualified
    );

    declaration = `var ${fieldName} ${fieldType} = `;
  }

  return declaration;
}

/**
 * Creates a special templating placeholder for local variable declarations that need to be
 * extracted and processed separately from the main template output.
 *
 * The {-{- syntax is a custom templating mechanism (not Go templates) used specifically
 * for handling local variable declarations in Go SDK generation. This syntax allows us to:
 *
 * 1. Embed variable declarations within template expressions
 * 2. Extract them later using localDeclarationRegex (line 466)
 * 3. Process them separately to generate proper Go code structure
 *
 * The syntax {-{-{fieldName:declarationCode}-}-} gets processed by templateUsageMethodParameters
 * where the regex extracts the fieldName and declarationCode, then replaces the placeholder
 * with just the fieldName while collecting the declarations for separate output.
 *
 * This is necessary because Go requires variable declarations to appear before their usage,
 * but our templating system generates usage expressions that may need to declare variables
 * inline (e.g., file reading operations that need error handling).
 *
 * @param field - The variable name to use in the final output
 * @param value - The complete declaration code including assignment and error handling
 * @returns A specially formatted string that will be processed by the template system
 */
function templateSpecialLocalDeclarationValue(
  field: string,
  value: string,
): string {
  return `{-{-{${field}:${value}}-}-}`;
}

// @ts-ignore
function templateFieldValueInit(
  field: FieldDef,
  valueVar: string,
  templateOptional = true,
): string {
  const nullableAndOptional =
    field.Nullable &&
    field.Optional &&
    context.Global.Config.NullableOptionalWrapper;

  if (!nullableAndOptional) {
    const isPointer = field.Optional || field.Nullable;
    return `${isPointer && templateOptional ? "&" : ""}${valueVar}`;
  }

  addImport("optionalnullable");
  return `optionalnullable.From(&${valueVar})`;
}
registerTemplateFunc("templateFieldValueInit", templateFieldValueInit);

// @ts-ignore
function templateValueWrapper(
  fieldDef: FieldDef,
  value: string,
  additionalContext?: TemplateValueContext,
): string {
  const nullableAndOptional =
    fieldDef.Nullable &&
    fieldDef.Optional &&
    context.Global.Config.NullableOptionalWrapper;

  if (!nullableAndOptional) {
    return value;
  }

  addImport("optionalnullable");

  if (value === "nil") {
    return `optionalnullable.From[${sanitizeType(
      fieldDef.Type,
      false,
      "",
    )}](nil)`;
  }

  if (
    value.startsWith(`${sanitizeSDKPackageName(true)}.Pointer`) ||
    value.startsWith("&")
  ) {
    return `optionalnullable.From(${value})`;
  }

  addSDKPackageImport(true);
  return `optionalnullable.From(${sanitizeSDKPackageName(
    true,
  )}.Pointer(${value}))`;
}
