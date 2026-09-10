// @ts-ignore
function templateRequestParams(
  usageContext: UsageContext,
  fieldDef: FieldDef,
  indent: number,
  parameterFlattening: boolean,
): string {
  const ctx: TemplateValueContext = {
    usageContext: usageContext,
    operation: usageContext.Operation,
    isTest: usageContext.Test && true,
    test: usageContext.Test,
    isUsage: true,
    isRequest: true,
  };

  if (parameterFlattening) {
    let fieldInstantiations = [];

    for (const field of sortMethodFields(fieldDef.Type.Fields)) {
      if (field.Type.Type.toString() == "class") {
        const fieldExample = getOperationMethodFieldExample(
          usageContext,
          field,
        );

        const value = templateValue(field, fieldExample, false, ctx);
        if (value == "") {
          continue;
        }

        fieldInstantiations.push(
          `$${sanitizeFieldName(field.Name)} = ${value};`,
        );
      }
    }

    if (fieldInstantiations.length == 0) {
      return "";
    }
    return indentLines(fieldInstantiations, indent);
  } else {
    const fieldExample = getOperationMethodFieldExample(usageContext, fieldDef);

    const value = templateValue(fieldDef, fieldExample, false, ctx);
    if (value == "") {
      return "";
    }

    return indentLines([`$request = ${value};`], indent);
  }
}

registerTemplateFunc("templateRequestParams", templateRequestParams);

// @ts-ignore
function templateUnion(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let { type: typeToUse, example: selectedExample } = selectExampleUnionType(
    fieldDef,
    example,
    undefined,
    additionalContext,
  );

  addTypeImportInline(typeToUse);
  return templateValue(
    typeDefToFieldDef(typeToUse, fieldDef),
    selectedExample,
    false,
    additionalContext,
  );
}

// @ts-ignore
function getOptionalUsageMethodParameters(
  local: UsageContext,
  indent: number,
): string[] {
  const params = [];

  if (local.Operation.Security) {
    params.push("security: $requestSecurity");
  }

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (!scope.IsGlobal && scope.Feature != "") {
        switch (scope.Feature.toString()) {
          case "server_url": {
            const url = getUsageServerUrl(local, scope);
            params.push(`${url}`);
            break;
          }
        }
      }
    });
  }

  return params;
}

// @ts-ignore
function joinParams(params: string[], indent): string {
  switch (params.length) {
    case 0:
      return "";
    case 1:
      return indentString(params[0], indent + 1);
    default:
      const indentedParams = params.map((p) => indentString(p, indent + 1));
      return `${indentedParams.join(",\n")}`;
  }
}

// @ts-ignore
function templateUsageMethodParameters(
  local: UsageContext,
  indent: number = 0,
  retriesParam: string = "",
): string {
  const operation = local.Operation;
  let methodParams = [];

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
    isUsage: true,
    isRequest: true,
  };

  if (operationParametersFlattened(operation)) {
    methodParams.push(...getOptionalUsageMethodParameters(local, indent));

    if (operation.Request) {
      for (const field of sortMethodFields(
        operation.Request.Field.Type.Fields,
      )) {
        if (field.Type.Type.toString() == "class") {
          methodParams.push(
            `${sanitizeFieldName(field.Name)}: $${sanitizeFieldName(
              field.Name,
            )}`,
          );
        } else {
          const fieldExample = getOperationMethodFieldExample(local, field);

          const value = templateValue(field, fieldExample, false, ctx);
          if (value == "") {
            continue;
          }

          methodParams.push(`${sanitizeFieldName(field.Name)}: ${value}`);
        }
      }
    }
  } else {
    if (operation.Request) {
      methodParams.push("request: $request");
    }

    methodParams.push(...getOptionalUsageMethodParameters(local, indent));
  }

  if (retriesParam) {
    methodParams.push(
      `options: Utils\\Options->builder()->setRetryConfig(${indentString(
        retriesParam,
        1,
      )})->build()`,
    );
  }
  if (methodParams.length > 1 && operationParametersFlattened(operation)) {
    return joinParams(methodParams, indent) + "\n";
  }
  return joinParams(methodParams, indent);
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function templateSDKUsage(outputLocation: string): string {
  const parts = context.Global.Config.Namespace.split("\\");

  if (context.Global.Config.Namespace == outputLocation) {
    return sanitizeFileName(context.Global.AST.MainSDK.Type.Name);
  }

  return `${parts[parts.length - 1]}\\${sanitizeFileName(
    context.Global.AST.MainSDK.Type.Name,
  )}`;
}

registerTemplateFunc("templateSDKUsage", templateSDKUsage);
