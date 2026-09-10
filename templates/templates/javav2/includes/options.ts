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

  return options;
}

registerTemplateFunc("supportedMethodOptions", supportedMethodOptions);

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
