// @ts-ignore
function templateSDKOptionName(name: string): string {
  let reservedOptions = ["setServerURL", "setTemplatedServerURL", "setClient"];

  if (context.Global.AST.MainSDK.Security) {
    reservedOptions.push("setSecurity");
  }
  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reservedOptions.push("setServer");
    reservedOptions.push("setTemplatedServer");
  } else {
    reservedOptions.push("setServerIndex");
  }

  let optionName = `set${sanitizeClassName(name)}`;

  if (reservedOptions.includes(optionName)) {
    optionName = `setGlobal${sanitizeClassName(name)}`;
  }

  return optionName;
}

registerTemplateFunc("templateSDKOptionName", templateSDKOptionName);

// @ts-ignore
function getModelJobs(types: BucketedTypes, path: string): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];

  for (const [rawOutputLocation, models] of sequencedMapEntries(types)) {
    const outputLocation = sanitizeOutputLocation(rawOutputLocation);
    for (const [model, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        let modelName = t.Name;

        let servers = null;

        if (t.Scope.toString() == "operations" && t.IsRequest()) {
          servers = context.Global.AST.OperationServers.Get(model)[0];
        }

        let ctx: PHPModel = {
          Name: modelName,
          Type: t,
          Servers: servers,
          OutputLocation: rawOutputLocation,
        };

        let outputPath = path;
        if (outputLocation) {
          outputPath += `/${outputLocation}`;
        }

        if (t.Type.toString() == "union") {
          continue;
        }
        if (t.Type.toString() == "error") {
          jobs.push(
            createTemplateFileJob(
              "error.php.stmpl",
              `${outputPath}/${sanitizeFileName(modelName)}Throwable.php`,
              ctx,
            ),
          );
        }
        jobs.push(
          createTemplateFileJob(
            `modelfile.php.stmpl`,
            `${outputPath}/${sanitizeFileName(modelName)}.php`,
            ctx,
          ),
        );
      }
    }
  }

  return jobs;
}

// @ts-ignore
function templateSDKInitFieldName(name: string, suffix: string): string {
  let reserved = ["security", "server_url", "url_params", "client"];

  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reserved.push("server");
  } else {
    reserved.push("server_idx");
  }

  let fieldName = sanitizeFieldName(name);

  if (reserved.includes(name)) {
    fieldName += "_" + suffix;
  }

  return fieldName;
}

registerTemplateFunc("templateSDKInitFieldName", templateSDKInitFieldName);

function escapeNamespace(namespace: string): string {
  return namespace.replaceAll("\\", "\\\\");
}

registerTemplateFunc("escapeNamespace", escapeNamespace);

function templateDefaultValue(
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  if (fieldDef.Optional) {
    return "null";
  }

  switch (fieldDef.Type.Type.toString()) {
    case "union":
      return templateDefaultTypeValue(
        fieldDef.Type.AssociatedTypes[0],
        usageLocation,
      );
    default:
      return templateDefaultTypeValue(fieldDef.Type, usageLocation);
  }
}

function templateDefaultTypeValue(
  typeDef: TypeDef,
  usageLocation: string,
): string {
  switch (typeDef.Type.toString()) {
    case "enum":
      return `${sanitizeClass(
        typeDef,
        usageLocation,
        false,
        Qualification.TYPE,
      )}::${getEnumNames(typeDef)[0]}`;
    case "class":
      return `new ${sanitizeClass(
        typeDef,
        usageLocation,
        false,
        Qualification.TYPE,
      )}()`;
    case "string":
      return `''`;
    case "date":
      return "\\Brick\\DateTime\\LocalDate::now(\\Brick\\DateTime\\TimeZone::utc())";
    case "date-time":
      return "new \\DateTime()";
    case "integer":
    case "int32":
    case "number":
    case "float32":
      return "0";
    case "boolean":
      return "false";
    case "bytes":
      return `''`;
    case "map":
    case "array":
      return "[]";
    case "any":
      return "null";
    case "response":
      return `new \\${getScopeNamespace("utils", true)}\\DefaultResponse()`;
    case "request":
      return `new \\${getScopeNamespace("utils", true)}\\DefaultRequest()`;
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("templateDefaultValue", templateDefaultValue);

//@ts-ignore
function templateSDKBuilderName(name: string): string {
  let reserved = ["setClient", "setSecurity", "setServerUrl", "setServer"];

  let methodName = sanitizeFieldName("set_" + name);

  if (reserved.includes(methodName)) {
    methodName = sanitizeFieldName("setGlobal_" + name);
  }

  return methodName;
}

registerTemplateFunc("templateSDKBuilderName", templateSDKBuilderName);

// @ts-ignore
function joinSDKOptions(options: string[]): string {
  if (options.length > 0) {
    options.unshift("");

    return options.join(`\n${"    ".repeat(1)}`) + "\n    ";
  }

  return "";
}

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

  return templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    0,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    undefined,
    additionalContext,
  );
}

// @ts-ignore
function templateUsageSDKOptions(local: UsageContext): string {
  const options = [];
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
                options.push(`->setSecurity(
${indentString(globalSecurity, 2)}
    )`);
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            const hasOperationServers =
              local.Operation?.Servers?.Servers?.length > 0;
            if (!hasOperationServers) {
              const url = getUsageServerUrl(local, scope);
              options.push(`->setServerURL(${url})`);
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              options.push(`->setServer('${server.ID}')`);
            } else if (server.Index !== undefined) {
              options.push(`->setServerIndex(${server.Index})`);
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const builderName = templateSDKBuilderName(v.Name);
                options.push(
                  `->${builderName}('${getUsageServerVariableValue(v)}')`,
                );
              });
            }

            break;
          }
          case "retries": {
            options.push(
              `->setRetryConfig(${indentString(
                templateString("usage/retries.stmpl", {}),
                2,
              )}
  )`,
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

            if (value === "") {
              return;
            }

            options.push(
              `->${templateSDKOptionName(parameter.field.Name)}(${value})`,
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
      options.push(`->setSecurity(
${indentString(globalSecurity, 2)}
    )`);
    }
  }

  return joinSDKOptions(options);
}

registerTemplateFunc("templateUsageSDKOptions", templateUsageSDKOptions);

function templateClassComments(
  comments: CommentDef | null,
  className: string,
  pkg: string,
  indent: number,
): string {
  return templateComments(comments, className, [], indent, false, "class");
}

registerTemplateFunc("templateClassComments", templateClassComments);

function needsFieldComments(comments: CommentDef | null, fieldDef: FieldDef) {
  return true;
}
registerTemplateFunc("needsFieldComments", needsFieldComments);

function templateFieldComments(
  comments: CommentDef | null,
  fieldDef: FieldDef,
  usageLocation: string,
  indent: number,
) {
  let fieldName = sanitizeFieldName(fieldDef.Name);

  if (!comments) {
    let summary = "";

    if (
      fieldDef.Type.Type.toString() == "map" ||
      fieldDef.Type.Type.toString() == "array"
    ) {
      summary = `$` + fieldName;
    }

    comments = {
      Summary: summary,
      Description: "",
      Deprecated: false,
    };
  }

  return templateComments(
    comments,
    null,
    [
      `@var ${sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        fieldDef.Nullable,
        usageLocation,
        Qualification.DOCSTRING,
      )} $${fieldName}`,
    ],
    indent,
    true,
    "field",
  );
}

registerTemplateFunc("templateFieldComments", templateFieldComments);

function templateSubSDKFieldComments(
  comments: CommentDef | null,
  fieldName: string,
  fieldType: string,
  indent: number,
) {
  return templateComments(
    comments,
    null,
    [`@var ${fieldType} $${fieldName}`],
    indent,
    true,
    "field",
  );
}

registerTemplateFunc(
  "templateSubSDKFieldComments",
  templateSubSDKFieldComments,
);

// @ts-ignore
function templateMethodArguments(
  operation: Operation,
  usageLocation: string,
  includeType: boolean = true,
): string {
  const args = [];

  const buildArgString = (
    type: string,
    optional: boolean,
    fieldName: string,
  ) => {
    return `${type} $${fieldName}${optional ? " = null" : ""}`;
  };

  // PHP treats both Optional and Nullable fields as having a `= null` default,
  // so we must sort them after required fields to avoid PHP 8.0+ deprecation
  // warnings about required parameters following optional ones.
  const sortedFields = sortMethodFields(operation.Arguments.Sorted);

  sortedFields.forEach((field) => {
    var typeName = sanitizeType(
      field.Type,
      field.Optional,
      field.Nullable,
      usageLocation,
      Qualification.TYPE,
    );

    if (includeType) {
      args.push(
        buildArgString(
          typeName,
          field.Optional || field.Nullable,
          sanitizeMethodParamName(field.Name),
        ),
      );
    } else {
      args.push(`$${sanitizeMethodParamName(field.Name)}`);
    }
  });

  if (operation.Servers) {
    if (includeType) {
      args.push(`?string $serverURL = null`);
    } else {
      args.push(`$serverURL`);
    }
  }
  if (operation.Extensions?.Pagination && hasPaginationURL(operation)) {
    if (includeType) {
      args.push("?string $urlOverride = null");
    } else {
      args.push("$urlOverride");
    }
  }

  if (includeType) {
    addImportInline(getScopeNamespace("utils", true) + "\\Options");
    args.push("?Options $options = null");
  } else {
    args.push("$options");
  }

  return args.join(", ");
}

registerTemplateFunc("templateMethodArguments", templateMethodArguments);

// @ts-ignore
// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map((f) => `\`${sanitizeFieldName(f.Name)}\``);

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set via \`setSecurity\` on the SDK builder.`
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
function templateMethodComments(
  operation: Operation,
  usageLocation: string,
  indent: number,
  generator: boolean = false,
) {
  let comments = operation.Comments;
  let attributes = [];

  if (!comments) {
    comments = {
      Summary: "",
      Description: "",
      Deprecated: false,
    };
  }

  if (!comments.Summary) {
    comments.Summary = sanitizeMethodName(operation);
  }

  const buildDocString = (type: string, fieldName: string) => {
    return `@param  ${type}  $${fieldName}`;
  };
  const sortedDocFields = sortMethodFields(operation.Arguments.Sorted);
  sortedDocFields.forEach((field) => {
    var typeName = sanitizeType(
      field.Type,
      field.Optional,
      field.Nullable,
      usageLocation,
      Qualification.DOCSTRING,
    );

    attributes.push(
      buildDocString(typeName, sanitizeMethodParamName(field.Name)),
    );
  });

  if (operation.Servers) {
    attributes.push(`@param  ?string  $serverURL`);
  }
  if (operation.Response.Type) {
    if (generator) {
      attributes.push(
        `@return \\Generator<${sanitizeType(
          operation.Response.Type,
          false,
          false,
          usageLocation,
          Qualification.DOCSTRING,
        )}>`,
      );
    } else {
      attributes.push(
        `@return ${sanitizeType(
          operation.Response.Type,
          false,
          false,
          usageLocation,
          Qualification.DOCSTRING,
        )}`,
      );
    }
  } else {
    attributes.push(`@return void`);
  }
  attributes.push(
    `@throws \\${getErrorsNamespace(true)}\\${getDefaultErrorClassName()}`,
  );
  return templateComments(
    comments,
    "",
    attributes,
    indent,
    false,
    "method",
    getHoistedSecurityRemark(operation),
  );
}

registerTemplateFunc("templateMethodComments", templateMethodComments);

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  title: string | null,
  attributes: string[] | null,
  indent: number,
  omitDocs: boolean = false,
  type: "method" | "field" | "class" | "enum" | null,
  additionalNotes: string = "",
): string {
  if (comments || additionalNotes) {
    let lines = [];

    let firstLineParts = [];

    if (title) {
      firstLineParts.push(title);
    }

    let descriptionLines = [];
    if (comments?.Description) {
      descriptionLines = comments.Description.split("\n");
    }

    if (comments?.Summary) {
      firstLineParts.push(comments.Summary);
    } else if (descriptionLines.length > 0) {
      firstLineParts.push(descriptionLines.shift());
    }

    const firstLine = firstLineParts.join(" - ");

    if (firstLine) {
      lines.push(...sanitizeComments(firstLine).split("\n"));
    }

    let description = false;

    if (descriptionLines.length > 0) {
      lines.push("");
      lines.push(...descriptionLines.map(sanitizeComments));
      description = true;
    }

    if (attributes == null) {
      attributes = [];
    }

    if (comments?.ExternalDocs && !omitDocs) {
      if (!description) {
        lines.push("");
      }
      let externalDocs = comments.ExternalDocs.URL;
      if (comments.ExternalDocs.Description) {
        externalDocs += ` - ${comments.ExternalDocs.Description}`;
      }

      lines.push(...sanitizeComments(externalDocs).split("\n"));
      attributes.push(`@see ${comments.ExternalDocs.URL}`);
    }

    if (additionalNotes) {
      if (lines.length > 0) {
        lines.push("");
      }
      lines.push(...additionalNotes.split("\n"));
    }

    if (comments?.Deprecated) {
      let deprecated = `@deprecated  ${type}: ${comments.DeprecationMessage}.`;

      if (comments.DeprecationReplacement) {
        const replacement = sanitizeDeprecationReplacement(
          comments.DeprecationReplacement,
          (type ?? "field") as "field" | "method" | "class",
        );
        if (replacement) {
          deprecated += ` Use ${replacement} instead.`; // TODO use @see to link eventually
        }
      }

      attributes.push(deprecated);
    }

    if (lines.length == 0 && attributes?.length == 0) {
      return "";
    }

    if (attributes?.length > 0) {
      lines.push("");
      lines.push(...attributes);
    }

    if (lines.length == 1) {
      return indentLines([`/** ${lines[0]} */`], indent);
    }

    lines = lines.map((line, index) => {
      if (line.length > 0) {
        return ` * ${line}`;
      }
      return " *";
    });

    lines.unshift("/**");
    lines.push(" */");

    return indentLines(lines, indent);
  }

  return "";
}

registerTemplateFunc("templateComments", templateComments);

function sortAllNullableOptionalToBackParams(params: FieldDef[]): FieldDef[] {
  params = params.slice();
  return params.sort((a, b) => {
    if (a.Optional || a.Nullable) {
      if (!b.Optional || !b.Nullable) {
        return 1;
      }
    }

    if (b.Optional || b.Nullable) {
      if (!a.Optional || !a.Nullable) {
        return -1;
      }
    }

    return 0;
  });
}
function sortAllConstAndDefaultFieldsToBackParams(
  params: FieldDef[],
): FieldDef[] {
  return params.sort((a, b) => {
    if (a.Const || a.Default) {
      if (!b.Const || !b.Default) {
        return 1;
      }
    }

    if (b.Const || b.Default) {
      if (!a.Const || !a.Default) {
        return -1;
      }
    }

    return 0;
  });
}

function sortClassFields(fields: FieldDef[]): FieldDef[] {
  return sortAllConstAndDefaultFieldsToBackParams(
    sortAllNullableOptionalToBackParams([...fields]),
  );
}

registerTemplateFunc("sortFields", sortClassFields);

function sortMethodFields(fields: FieldDef[]): FieldDef[] {
  return [...fields].sort((a, b) => {
    const aOpt = a.Optional || a.Nullable;
    const bOpt = b.Optional || b.Nullable;
    if (aOpt && !bOpt) return 1;
    if (!aOpt && bOpt) return -1;
    return 0;
  });
}

function sortOperations(operations: Operation[]): Operation[] {
  operations = operations.slice();
  return operations.sort((a, b) => {
    if (a.ID < b.ID) {
      return -1;
    }
    if (a.ID > b.ID) {
      return 1;
    }
    return 0;
  });
}

registerTemplateFunc("sortOperations", sortOperations);

function templateModelConstructorArgs(modelType: TypeDef) {
  let headerParam = "";
  let httpMetaParam = "";
  const params = [];

  let fields = sortClassFields(modelType.Fields);

  for (const field of fields) {
    if (field.Name.toLowerCase() == "headers") {
      headerParam = `${sanitizeType(
        field.Type,
        true,
        field.Nullable,
        modelType.OutputLocation,
        Qualification.TYPE,
      )} $${sanitizeFieldName(field.Name)}`;
      if (field.Type.Type.valueOf() in ["array", "map"]) {
        headerParam += ` = ${templateDefaultValue(
          field,
          modelType.OutputLocation,
        )}`;
      } else {
        headerParam += " = null";
      }
    } else if (field.Name.toLowerCase() == "httpmeta") {
      httpMetaParam = `${sanitizeType(
        field.Type,
        field.Optional,
        field.Nullable,
        modelType.OutputLocation,
        Qualification.TYPE,
      )} $${sanitizeFieldName(field.Name)}`;
    } else {
      if (field.Const || field.Default) {
        params.push(
          `${sanitizeType(
            field.Type,
            field.Optional,
            field.Nullable,
            modelType.OutputLocation,
            Qualification.TYPE,
          )} $${sanitizeFieldName(field.Name)} = ${templateConstOrDefaultValue(
            field,
          )}`,
        );
      } else if (field.Optional || field.Nullable) {
        params.push(
          `${sanitizeType(
            field.Type,
            field.Optional,
            field.Nullable,
            modelType.OutputLocation,
            Qualification.TYPE,
          )} $${sanitizeFieldName(field.Name)} = null`,
        );
      } else {
        params.push(
          `${sanitizeType(
            field.Type,
            field.Optional,
            field.Nullable,
            modelType.OutputLocation,
            Qualification.TYPE,
          )} $${sanitizeFieldName(field.Name)}`,
        );
      }
    }
  }

  if (httpMetaParam.length > 0) {
    params.unshift(httpMetaParam);
  }

  if (headerParam.length > 0) {
    params.push(headerParam);
  }

  return params.join(", ");
}

registerTemplateFunc(
  "templateModelConstructorArgs",
  templateModelConstructorArgs,
);

function templateModelConstructorComments(modelType: TypeDef) {
  const params = [];
  for (const field of sortAllNullableOptionalToBackParams(modelType.Fields)) {
    params.push(
      `@param  ${sanitizeType(
        field.Type,
        field.Optional,
        field.Nullable,
        modelType.OutputLocation,
        Qualification.DOCSTRING,
      )}  $${sanitizeFieldName(field.Name)}`,
    );
  }
  if (params.length == 0) {
    return "";
  }
  return `/**
 * ${params.join("\n * ")}
 * @phpstan-pure
 */`;
}

registerTemplateFunc(
  "templateModelConstructorComments",
  templateModelConstructorComments,
);

// @ts-ignore
function templateStatusCodeCheck(
  varName: string,
  statusCodes: string[],
  inverted: boolean = false,
): string {
  const $templatedCodes = statusCodes.map((c) => `'${c}'`);
  return `${
    inverted ? "! " : ""
  }Utils\\Utils::matchStatusCodes(${varName}, [${$templatedCodes.join(", ")}])`;
}
registerTemplateFunc("templateStatusCodeCheck", templateStatusCodeCheck);

function templateModelField(field: FieldDef, outputLocation: string): string {
  let res: string = "";

  let fieldName = sanitizeFieldName(field.Name);
  let optional = fieldName == "httpMeta" ? true : field.Optional;

  let annotations: string = templateAnnotations(
    field,
    outputLocation,
    optional,
  );
  if (annotations !== undefined) {
    if (annotations.trim() !== "") {
      res = `${annotations}\n`;
    }
  }
  if (field.Optional) {
    res += `public ${sanitizeType(
      field.Type,
      optional,
      field.Nullable,
      outputLocation,
      Qualification.TYPE,
    )} $${fieldName} = ${templateDefaultValue(field, outputLocation)};`;
  } else {
    res += `public ${sanitizeType(
      field.Type,
      optional,
      field.Nullable,
      outputLocation,
      Qualification.TYPE,
    )} $${fieldName};`;
  }
  return res;
}
registerTemplateFunc("templateModelField", templateModelField);

function templateClientCall(clientAccess: string, operation: Operation) {
  let base = `${clientAccess}($httpRequest, $httpOptions)`;
  if (operation.Extensions?.Retries) {
    addImportInline(getScopeNamespace("retry", true) + "\\RetryUtils");
    const usesAttemptCountRetries =
      String(operation.Extensions.Retries.Strategy ?? "") ===
      "attempt-count-backoff";
    if (usesAttemptCountRetries) {
      const clientVar = clientAccess.split("->")[0];
      const useVars = ["$httpRequest", "$httpOptions"];
      if (clientVar !== "$this") {
        useVars.push(clientVar);
      }
      base = `RetryUtils::retryWrapper(function (int $attempt) use (${useVars.join(
        ", ",
      )}) {
                return ${clientAccess}($httpRequest, $httpOptions);
            }, $retryConfig, $retryCodes)`;
    } else {
      base = `RetryUtils::retryWrapper(fn () => ${clientAccess}($httpRequest, $httpOptions), $retryConfig, $retryCodes)`;
    }
  }

  return base;
}
registerTemplateFunc("templateClientCall", templateClientCall);

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    filePath = `../.speakeasy/testfiles/${filePath}`;
  }

  return `file_get_contents('${filePath}');`;
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
  }

  return `file_get_contents('${filePath}');`;
}

registerTemplateFunc("templateFileToStringValue", templateFileToStringValue);
