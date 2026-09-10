// @ts-ignore
function templateUsageSDKParams(local: UsageContext): string {
  const params = [];
  let hasSecurity = local.Operation.Security != undefined;
  let hasAbsoluteServerURL =
    context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false;

  const ctx: TemplateValueContext = {
    usageContext: local, // TODO see if this can be decomposed into the required fields below
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
  };

  const hasOperationServers = local.Operation?.Servers?.Servers?.length > 0;

  if (!hasAbsoluteServerURL && !hasOperationServers) {
    const url = getUsageServerUrl(local, undefined, ctx);
    params.push(`server_url: ${url},`);
  }

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal && scope.Feature != "") {
        const feature = scope.Feature.toString();
        switch (feature) {
          case "retries": {
            params.push(
              `retry_config: ${templateIndentedString(
                "usage/retries.stmpl",
                {},
                3,
              ).trim()},`,
            );
            break;
          }
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
            if (hasAbsoluteServerURL && !hasOperationServers) {
              params.push(
                `server_url: ${getUsageServerUrl(local, scope, ctx)},`,
              );
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              params.push(`server: "${server.ID}",`);
            } else if (server.Index !== undefined) {
              params.push(`server_idx: ${server.Index},`);
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const fieldName = templateSDKInitFieldName(v.Name, "global");
                params.push(
                  `${fieldName}: "${getUsageServerVariableValue(v)}",`,
                );
              });
            }

            break;
          }
          case "http_client": {
            params.push(`client: ${templateHTTPClient(scope.Value)},`);
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
              `${sanitizeFieldName(parameter.field.Name)}: ${value},`,
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

  if (params.length > 0) {
    return `(${joinSDKParams(params)})`;
  } else {
    return "";
  }
}

registerTemplateFunc("templateUsageSDKParams", templateUsageSDKParams);

// @ts-ignore
function getOptionalUsageMethodParameters(
  local: UsageContext,
  indent: number,
): string[] {
  const params = [];

  if (local.Operation.Security) {
    let securityExample = undefined;

    if (local.Scopes != undefined) {
      for (const scope of local.Scopes) {
        if (scope.Feature == "security") {
          if (scope.Value) {
            if ("example" in scope.Value) {
              securityExample = scope.Value.example;
            } else {
              securityExample = getExampleValue(scope.Value, {
                usageContext: local,
              });
            }
          }
          break;
        }
      }
    }

    const ctx: TemplateValueContext = {
      usageContext: local,
      operation: local.Operation,
      isTest: local.Test && true,
      test: local.Test,
    };

    const secValue = templateSecurityUsage(
      local.Operation.Security.Type,
      0,
      true,
      false,
      undefined,
      securityExample,
      ctx,
    ).trimStart();
    // Strip trailing commas before closing parens in security model constructors
    const cleanedSec = secValue.replace(/,(\n\s*\))/g, "$1");
    params.push(`security: ${cleanedSec}`);
  }

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (!scope.IsGlobal && scope.Feature != "") {
        switch (scope.Feature.toString()) {
          case "retries": {
            return `retries: ${templateString("usage/retries.stmpl", {})},`;
          }
          case "server_url": {
            const url = getUsageServerUrl(local, scope);
            params.push(`server_url: ${url}`);
            break;
          }
          case "content_type": {
            const acceptType = scope.OpFilter.split(";")[0];
            params.push(`accept_header_override: '${acceptType}'`);
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

  if (operationParametersFlattened(operation)) {
    methodParams.push(...getOptionalUsageMethodParameters(local, indent));

    if (operation.Request) {
      for (const field of sortMethodParams(
        operation.Request.Field.Type.Fields,
      )) {
        const fieldExample = getOperationMethodFieldExample(local, field);

        const value = templateValue(field, fieldExample, false, {
          ...ctx,
          isRequest: true,
        });

        if (value == "") {
          continue;
        }

        methodParams.push(
          `${sanitizeMethodArgumentName(field.Name)}: ${value}`,
        );
      }
    }
  } else {
    if (operation.Request) {
      methodParams.push("request: req");
    }

    methodParams.push(...getOptionalUsageMethodParameters(local, indent));
  }

  return methodParams.join(", ");
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function templateHTTPClient(value: HTTPClientUsage): string {
  switch (value.type) {
    case "test":
      return `test_http_client`;
    default:
      throw new Error(`unsupported http client type: ${value.type}`);
  }
}

// Ruby-specific override: sort map keys for deterministic ordering in generated code.
// @ts-ignore
function templateMap(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let mapValues = [];

  // We shouldn't render empty maps for optional field if we are targeting the mock server as it will fail to return an empty map for an optional field
  if (
    Object.keys(example).length == 0 &&
    fieldDef.Optional &&
    additionalContext?.targetingMockServer &&
    !additionalContext?.isResponse
  ) {
    return "";
  }

  for (const key of Object.keys(example).sort()) {
    mapValues.push(
      templateMapValue(
        key,
        templateValue(
          typeDefToFieldDef(
            fieldDef.Type.ItemType,
            fieldDef,
            undefined,
            fieldDef.Type.ContainsNull,
          ),
          example[key],
          false,
          additionalContext,
        ),
      ),
    );
  }

  const lines = [
    `${templateType(fieldDef.Type, additionalContext)}${templateBracket(
      fieldDef.Type,
      true,
      additionalContext,
    )}`,
  ];
  const content = indentLines(processMapValues(mapValues), 1);
  const closingBracket = templateBracket(
    fieldDef.Type,
    false,
    additionalContext,
  );
  if (closingBracketOnSeparateLine()) {
    lines.push(content);
    lines.push(closingBracket);
  } else {
    lines.push(content + closingBracket);
  }
  return lines.join("\n");
}
