function methodOptionsParameter(): JavaParam {
  return {
    Type: `${templatePackageName()}.utils.Options`,
    EnhancedType: `java.util.Optional<${templatePackageName()}.utils.Options>`,
    Name: "options",
    Optional: true,
    Nullable: false,
    IncludedInSignature: true,
    Util: true,
    ParamType: "string",
    Comment: "additional options",
    HasItemType: false,
  };
}

registerTemplateFunc("methodOptionsParameter", methodOptionsParameter);

function supportedMethodOptions(operation: Operation): JavaParam[] {
  const options: JavaParam[] = [];

  if (operation.Extensions.Retries) {
    const retriesType = `${templatePackageName()}.utils.RetryConfig`;
    options.push({
      Type: retriesType,
      EnhancedType: `java.util.Optional<${retriesType}>`,
      Name: "retryConfig",
      Optional: true,
      Nullable: false,
      IncludedInSignature: false,
      Util: false,
      ParamType: "string",
      Comment: "allows the configuration of retry parameters",
      HasItemType: false,
    });
  }

  if (!operationHasParameterNamed(operation, "timeout")) {
    options.push({
      Type: "java.time.Duration",
      EnhancedType: "java.util.Optional<java.time.Duration>",
      Name: "timeout",
      Optional: true,
      Nullable: false,
      IncludedInSignature: false,
      Util: false,
      ParamType: "string",
      Comment: "overrides the request timeout for this call",
      HasItemType: false,
    });
  }

  return options;
}

function operationHasParameterNamed(
  operation: Operation,
  name: string,
): boolean {
  if (!operationParametersFlattened(operation) || !operation.Request?.Field) {
    return false;
  }
  return parametersForType(operation.Request.Field.Type, true, false).some(
    (p) => p.Name === name,
  );
}

registerTemplateFunc("supportedMethodOptions", supportedMethodOptions);

function templateTimeoutResolution(
  operation: Operation,
  optionsExpr: string,
): string {
  const optional = javaImportOptional();
  const extension = operation.Extensions.Timeout
    ? `${optional}.of(${javaImport("java.time.Duration")}.ofMillis(${
        operation.Extensions.Timeout
      }))`
    : "";
  if (!optionsExpr) {
    if (!extension) {
      return "sdkConfiguration.timeout()";
    }
    return isJava8()
      ? `${javaImportLocal(
          "utils.Java8Compat",
        )}.or(sdkConfiguration.timeout(), () -> ${extension})`
      : `sdkConfiguration.timeout().or(() -> ${extension})`;
  }
  const perCall = `${optionsExpr}.flatMap(${javaImportLocal(
    "utils.Options",
  )}::timeout)`;
  if (isJava8()) {
    const compat = javaImportLocal("utils.Java8Compat");
    const withDefault = `${compat}.or(${perCall}, sdkConfiguration::timeout)`;
    return extension
      ? `${compat}.or(${withDefault}, () -> ${extension})`
      : withDefault;
  }
  const withDefault = `${perCall}.or(sdkConfiguration::timeout)`;
  return extension ? `${withDefault}.or(() -> ${extension})` : withDefault;
}

registerTemplateFunc("templateTimeoutResolution", templateTimeoutResolution);

function templateLegacyConstructor(
  operation: Operation,
  isAsync: boolean,
  className: string,
): string {
  if (operation.Extensions.Retries) {
    return "";
  }
  const optionsName = methodOptionsParameter().Name;
  const params = operationParameters(operation, isAsync).filter(
    (p) => p.IncludedInSignature,
  );
  if (!params.some((p) => p.Name == optionsName)) {
    return "";
  }
  const emptyOptions = context.Global.Config.NullFriendlyParameters
    ? "null"
    : `${javaImportOptional()}.empty()`;
  const args = params
    .map((p) => (p.Name == optionsName ? emptyOptions : p.Name))
    .concat(["_headers"]);
  const signature = templateMethodParams(
    params.filter((p) => p.Name != optionsName),
    true,
    16,
    2,
  );
  return `
        public ${className}(${signature}) {
            this(${groupArgs(args, 2).join(",\n                    ")});
        }
`;
}

registerTemplateFunc("templateLegacyConstructor", templateLegacyConstructor);

function optionEnumName(name: string): string {
  return name.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toUpperCase();
}

registerTemplateFunc("optionEnumName", optionEnumName);

function templateRequestBuilderCallOptions(
  options: JavaParam[],
  indent: number = 1,
  inlineReturn?: boolean,
): string {
  if (options.length == 0) {
    return "";
  }
  const optionsParam = methodOptionsParameter();
  const optionsLines: string[] = [];
  if (inlineReturn) {
    optionsLines.push(
      `return ${javaImportOptional()}.of(${javaImport(
        optionsParam.Type,
      )}.builder()`,
    );
  } else {
    optionsLines.push(
      `${javaImport(optionsParam.EnhancedType)} ${
        optionsParam.Name
      } = ${javaImportOptional()}.of(${javaImport(
        optionsParam.Type,
      )}.builder()`,
    );
  }

  options.forEach((p) => {
    optionsLines.push(
      `${templateIndent(indent)}.${sanitizeFieldName(p.Name)}(${p.Name})`,
    );
  });
  optionsLines.push(`${templateIndent(indent)}.build());`);

  return optionsLines.join("\n");
}

registerTemplateFunc(
  "templateRequestBuilderCallOptions",
  templateRequestBuilderCallOptions,
);
