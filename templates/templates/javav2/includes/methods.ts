type JavaParam = {
  Type: string;
  EnhancedType: string; //  Type wrapped with Optional/JsonNullable (if appropriate)
  BigType?: BigType;
  Name: string;
  Optional: boolean;
  Nullable: boolean;
  Default?: AnyValue;
  IncludedInSignature: boolean;
  Util: boolean;
  ParamType: string;
  Comment: string;
  HasItemType: boolean;
  IsFlattened?: boolean;
  Final?: boolean;
  Field?: FieldDef;
  Assignment?: string;
};

function getRequestParameter(
  operation: Operation,
  isAsync: boolean = false,
): JavaParam | null {
  if (!operation.Request) {
    return null;
  }

  let optional = false;
  let nullable = false;
  let defaultValue: AnyValue = undefined;
  // if request body and params present then type is wrapped
  if (operation.Request.RequestBody && !operation.Request.Params) {
    optional = operation.Request.RequestBody.Optional;
    nullable = operation.Request.RequestBody.Nullable;
    defaultValue = operation.Request.RequestBody.Default;
  }
  return {
    Type: sanitizeTypeMandatory(operation.Request.Field.Type, isAsync),
    EnhancedType: sanitizeType(
      operation.Request.Field.Type,
      optional,
      nullable,
      true,
      isAsync,
    ),
    BigType: bigType(operation.Request.Field.Type),
    Name: "request",
    Optional: optional,
    Nullable: nullable,
    Default: defaultValue,
    IncludedInSignature: true,
    Util: false,
    ParamType: operation.Request.Field.Type.Name,
    Comment:
      "The request object containing all the parameters for the API call.",
    HasItemType: operation.Request.Field.Type.ItemType != undefined,
    Field: operation.Request.Field,
  };
}

registerTemplateFunc("getRequestParameter", getRequestParameter);

function parameterForField(
  field: FieldDef,
  isAsync: boolean = false,
): JavaParam {
  return {
    Type: sanitizeTypeMandatory(field.Type, isAsync),
    EnhancedType: sanitizeType(
      field.Type,
      field.Optional,
      field.Nullable,
      true,
      isAsync,
    ),
    BigType: bigType(field.Type),
    Name: sanitizeFieldName(field.Name),
    Optional: field.Optional,
    Nullable: field.Nullable,
    Default: field.Default,
    IncludedInSignature: true,
    Util: false,
    ParamType:
      field.Type.Type.valueOf() === "class" ||
      field.Type.Type.valueOf() === "enum" ||
      field.Type.Type.valueOf() === "union"
        ? field.Type.Name
        : field.Type.Type.valueOf(),
    Comment: field.Type.Comments?.Description ?? "",
    HasItemType: field.Type.ItemType != undefined,
    IsFlattened: true,
    Field: field,
  };
}

registerTemplateFunc("parameterForField", parameterForField);

function parametersForType(
  type: TypeDef,
  includeAdditionalProperties: boolean = true,
  isAsync: boolean = false,
): JavaParam[] {
  const parameters: JavaParam[] = [];
  for (const field of type.Fields) {
    if (field.IsAdditionalProperties && !includeAdditionalProperties) {
      continue;
    }
    if (!!field.Const) {
      continue;
    }
    parameters.push(parameterForField(field, isAsync));
  }

  return parameters;
}

registerTemplateFunc("parametersForType", parametersForType);

function methodParameters(
  operation: Operation,
  scope: Scope,
  isAsync: boolean = false,
): JavaParam[] {
  let parameters: JavaParam[] = [];

  if (operationParametersFlattened(operation)) {
    if (operation.Security) {
      parameters.push({
        Type: sanitizeClass(operation.Security.Type, "", false),
        EnhancedType: sanitizeClass(operation.Security.Type, "", false),
        BigType: bigType(operation.Security.Type),
        Name: "security",
        Optional: false,
        Nullable: false,
        IncludedInSignature: true,
        Util: false,
        ParamType: operation.Security.Type.Name,
        Comment: "The security details to use for authentication.",
        HasItemType: operation.Security.Type.ItemType != undefined,
        Field: operation.Security,
      });
    }
    parameters.push(
      ...parametersForType(operation.Request.Field.Type, true, isAsync),
    );
  } else {
    const requestParam = getRequestParameter(operation, isAsync);
    if (requestParam) {
      parameters.push(requestParam);
    }

    if (operation.Security) {
      parameters.push({
        Type: sanitizeClass(operation.Security.Type, "", false),
        EnhancedType: sanitizeClass(operation.Security.Type, "", false),
        BigType: bigType(operation.Security.Type),
        Name: "security",
        Optional: false,
        Nullable: false,
        IncludedInSignature: true,
        Util: false,
        ParamType: operation.Security.Type.Name,
        Comment: "The security details to use for authentication.",
        HasItemType: operation.Security.Type.ItemType != undefined,
        Field: operation.Security,
      });
    }
  }

  if (operation.Servers) {
    parameters.push({
      Type: "java.lang.String",
      EnhancedType: "java.util.Optional<java.lang.String>",
      BigType: undefined,
      Name: "serverURL",
      Optional: true,
      Nullable: false,
      IncludedInSignature: true,
      Util: false,
      ParamType: "string",
      Comment: "Overrides the server URL.",
      HasItemType: false,
    });
  }

  const optionParameters = supportedMethodOptions(operation);
  if (optionParameters.length > 0) {
    parameters.push(methodOptionsParameter());
    parameters.push(...optionParameters);
  }

  return parameters;
}
registerTemplateFunc("methodParameters", methodParameters);

function getAuxiliaryParameters(params: JavaParam[]): JavaParam[] {
  const optionsParam = methodOptionsParameter();
  return params
    .filter((p) => p.IncludedInSignature)
    .filter((p) => !p.IsFlattened)
    .filter((p) => p.Name != optionsParam.Name);
}
registerTemplateFunc("getAuxiliaryParameters", getAuxiliaryParameters);

enum BigType {
  BigInteger,
  BigIntegerString,
  BigDecimal,
  BigDecimalString,
}

function parameterIsBigInteger(p: JavaParam): boolean {
  return (
    p.BigType &&
    // @ts-expect-error Unsure why tsc is convinced BigType.BigInteger is never a value.
    (p.BigType == BigType.BigInteger ||
      p.BigType == BigType.BigIntegerString) &&
    !p.HasItemType
  );
}

function parameterIsBigDecimal(p: JavaParam): boolean {
  return (
    p.BigType &&
    (p.BigType == BigType.BigDecimal ||
      p.BigType == BigType.BigDecimalString) &&
    !p.HasItemType
  );
}

function hasShapeString(type: TypeDef): boolean {
  const bt = bigType(type);
  const result =
    bt && (bt == BigType.BigDecimalString || bt == BigType.BigIntegerString);
  if (result) {
    return true;
  } else {
    if (type.ItemType) {
      return hasShapeString(type.ItemType);
    } else {
      return false;
    }
  }
}

registerTemplateFunc("hasShapeString", hasShapeString);

function includeTypeReference(field: FieldDef): boolean {
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() == "param") {
      return hasShapeString(field.Type);
    }
  }
  return false;
}

registerTemplateFunc("includeTypeReference", includeTypeReference);

function templateMethodSignatures(operation: Operation, scope: Scope): string {
  const params = methodParameters(operation, scope);

  return templateMethodOverload(params, operation, scope);
}

registerTemplateFunc("templateMethodSignatures", templateMethodSignatures);

function templateRequestBuilderSimpleClassName(operation: Operation): string {
  return sanitizeClassName(operation.ID) + "RequestBuilder";
}

registerTemplateFunc(
  "templateRequestBuilderSimpleClassName",
  templateRequestBuilderSimpleClassName,
);

function templateOperationSimpleClassName(operation: Operation): string {
  return sanitizeClassName(operation.ID);
}

registerTemplateFunc(
  "templateOperationSimpleClassName",
  templateOperationSimpleClassName,
);

function templateRequestBuilderFullClassName(
  operation: Operation,
  isAsync: boolean = false,
): string {
  const ns = getModelNamespace(
    true,
    "operations",
    "",
    `${templateRequestBuilderSimpleClassName(operation)}`,
  );
  if (!isAsync) {
    return ns;
  }
  const parts = ns.split(".");
  const className = parts.pop();
  return [...parts, "async", className].join(".");
}

registerTemplateFunc(
  "templateRequestBuilderFullClassName",
  templateRequestBuilderFullClassName,
);

function templatePojoBuilderClassName(operation: Operation): string | null {
  if (operationParametersFlattened(operation)) {
    return `${sanitizeTypeMandatory(operation.Request.Field.Type)}`;
  }
  return null;
}

registerTemplateFunc(
  "templatePojoBuilderClassName",
  templatePojoBuilderClassName,
);

function useNullableWrapperForParam(
  p: JavaParam | null,
  force: boolean = false,
): boolean {
  if (p?.Nullable) {
    if (force) {
      return true;
    }
    return Boolean(p?.Optional || p?.Default);
  }

  return false;
}
registerTemplateFunc("useNullableWrapperForParam", useNullableWrapperForParam);

function useNullableWrapperForField(f: FieldDef): boolean {
  return Boolean(f.Nullable && (f.Optional || f.Default));
}
registerTemplateFunc("useNullableWrapperForField", useNullableWrapperForField);

function templateMethodParams(
  parameters: JavaParam[],
  // include extra parameters such as headers, timeouts
  includeRequestOptions: boolean = true,
  continuationIndent = 8,
  groupLength: number = 3,
): string {
  const filtered = parameters.filter((p) => p.IncludedInSignature);
  if (filtered.length === 0) return "";
  const mapped = filtered.map((p) => {
    const paramType = context.Global.Config.NullFriendlyParameters
      ? javaImport(getConstructorParamType(p))
      : javaImport(p.EnhancedType);

    return `${paramType} ${p.Name}`;
  });
  if (includeRequestOptions) {
    mapped.push(`${javaImportLocal("utils.Headers")} _headers`);
  }
  const grouped = groupArgs(mapped, groupLength);
  if (grouped.length == 1) {
    return grouped[0];
  }
  const indentString = " ".repeat(continuationIndent);

  return "\n" + grouped.map((line) => `${indentString}${line}`).join(`,\n`);
}
registerTemplateFunc("templateMethodParams", templateMethodParams);

function templateMethodCallArgs(
  parameters: JavaParam[] | string[],
  // include extra parameters such as headers, timeouts
  includeRequestOptions: boolean = true,
  continuationIndent = 8,
  groupLength: number = 3,
): string {
  const filtered = parameters.filter((p) => {
    if (typeof p === "string") {
      return true;
    }
    return p.IncludedInSignature;
  });
  if (filtered.length === 0) return "";
  const mapped = filtered.map((p) => {
    if (typeof p === "string") {
      return p;
    }
    return p.Assignment || p.Name;
  });
  if (includeRequestOptions) {
    mapped.push("_headers");
  }
  const grouped = groupArgs(mapped, groupLength);
  if (grouped.length == 1) {
    return grouped[0];
  }
  const indentString = " ".repeat(continuationIndent);

  return "\n" + grouped.map((line) => `${indentString}${line}`).join(`,\n`);
}
registerTemplateFunc("templateMethodCallArgs", templateMethodCallArgs);

function collectRequiredArgs(
  parameters: JavaParam[],
  operation: Operation,
): string[] {
  const args = [];
  for (const p of parameters) {
    if (context.Global.Config.NullFriendlyParameters) {
      if (!isRequiredArg(p)) {
        if (p.Nullable) {
          if (p.Optional) {
            args.push(`${javaImportJsonNullable()}.undefined()`);
            continue;
          }
          if (p.Default) {
            const modelBuilder = javaImportLocal(
              templatePojoBuilderClassName(operation),
            );
            if (modelBuilder) {
              args.push(
                `${javaImportJsonNullable()}.of(${modelBuilder}.${singletonValueGet(
                  p.Name,
                )})`,
              );
            } else {
              args.push(`${javaImportJsonNullable()}.of(${p.Default})`);
            }
            continue;
          }
        }
        args.push("null");
        continue;
      }
      args.push(p.Name);
      continue;
    }

    if (p.Optional && p.Nullable) {
      args.push(`${javaImportJsonNullable()}.undefined()`);
    } else if (p.Optional || p.Nullable) {
      args.push(`${javaImportOptional()}.empty()`);
    } else {
      args.push(p.Name);
    }
  }

  return args;
}
registerTemplateFunc("collectRequiredArgs", collectRequiredArgs);

function templateCommentsForOperation(
  operation: Operation,
  fallbackComment: string,
  params: JavaParam[],
  returnsComment?: string,
): string {
  const comments = operation.Comments || simpleComment(fallbackComment);

  return templateComments(
    comments,
    null,
    1,
    "method",
    false,
    operation,
    params,
    returnsComment,
    false,
    true,
    3,
    getHoistedSecurityRemark(operation),
  );
}
registerTemplateFunc(
  "templateCommentsForOperation",
  templateCommentsForOperation,
);

// @ts-ignore
function templateMethodOverload(
  parameters: JavaParam[],
  operation: Operation,
  scope: Scope,
): string {
  let methodNameBase = sanitizeMethodName(operation);

  const signatureParameters = parameters.filter((p) => p.IncludedInSignature);

  // if no parameters we need to distinguish the service
  // method that returns a RequestBuilder (allowing
  // customization of timeouts, retries etc.) vs the
  // direct call that returns the Response.
  const allParamsMethodName =
    methodNameBase + (signatureParameters.length == 0 ? "Direct" : "");

  // we add a parameter-less method that returns a builder
  const cls = `${getModelNamespace(
    true,
    "operations",
  )}.${templateRequestBuilderSimpleClassName(operation)}`;

  const builderComments = templateComments(
    operation.Comments || simpleComment("Returns a builder to make a request."),
    null,
    1,
    "method",
    false,
    operation,
    [],
    "The call builder",
    false,
    false,
    3,
    getHoistedSecurityRemark(operation),
  );
  let sig = builderComments;
  if (operation.Comments?.Deprecated) {
    sig += `    @${javaImportDeprecated()}\n`;
  }
  sig += `    public ${javaImport(cls)} ${methodNameBase}() {
        return new ${javaImport(cls)}(sdkConfiguration);
    }

`;
  // add a non-builder method that has just the required signature parameters
  const requiredParameters = signatureParameters.filter(isRequiredArg);
  if (requiredParameters.length != signatureParameters.length) {
    const comments = templateComments(
      operation.Comments ||
        simpleComment("Makes a request (required parameters only)."),
      null,
      1,
      "method",
      false,
      operation,
      requiredParameters,
      "The response from the API call",
      true,
      false,
      3,
      getHoistedSecurityRemark(operation),
    );

    // if no parameters we need to distinguish the service
    // method that returns a RequestBuilder (allowing
    // customization of timeouts, retries etc.) vs the
    // direct call that returns the Response.
    const methodName =
      methodNameBase + (requiredParameters.length == 0 ? "Direct" : "");
    sig += comments;

    if (operation.Comments?.Deprecated) {
      sig += `    @${javaImportDeprecated()}\n`;
    }
    // delegate to the primary method filling in all optional fields as empty
    sig += `    public ${javaImport(
      sanitizeClass(operation.Response.Type, scope, false),
    )} ${methodName}(${templateMethodParams(
      requiredParameters,
      false,
      12,
      2,
    )}) {
        return ${allParamsMethodName}(${groupArgs(
          collectRequiredArgs(signatureParameters, operation),
          3,
        ).join(",\n            ")});
    }

`;
  }
  let returns = operation.Response.Type?.Comments?.Description;
  if (!returns) {
    returns = "The response from the API call";
  }
  const allParamsComments = templateComments(
    operation.Comments || simpleComment("Makes a request."),
    null,
    1,
    "method",
    false,
    operation,
    parameters,
    returns,
    true,
    false,
    3,
    getHoistedSecurityRemark(operation),
  );
  sig += allParamsComments;
  if (operation.Comments?.Deprecated) {
    sig += `    @${javaImportDeprecated()}\n`;
  }
  sig += `    public ${javaImport(
    sanitizeClass(operation.Response.Type, scope, false),
  )} ${allParamsMethodName}(${templateMethodParams(
    signatureParameters,
    false,
    12,
    2,
  )}) {`;

  return sig;
}

function simpleComment(description: string): CommentDef {
  return {
    Description: description,
    Summary: "",
  };
}

// @ts-ignore
function getOptionalUsageMethodParameters(
  usageContext: UsageContext,
  indent: number,
): string[] {
  const params = [];

  if (usageContext.Operation.Security) {
    let securityScope = undefined;

    for (const scope of usageContext.Scopes) {
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
        securityExample = getExampleValue(securityScope.Value, {
          usageContext: usageContext,
        });
      }
    }

    if (
      securityExample !== undefined ||
      !usageContext.Operation.Security.Optional
    ) {
      const sec = usageContext.Operation.Security.Type;
      params.push(
        `\n${templateUnflattenedSecurityUsage(
          sec,
          indent + 1,
          true,
          secFieldIndex(securityExample, sec) || 0,
          securityExample,
          { isTest: usageContext.Test && true },
        )}`,
      );
    }
  }

  if (usageContext.Scopes != undefined) {
    usageContext.Scopes.forEach((scope) => {
      if (!scope.IsGlobal && scope.Feature != "") {
        switch (scope.Feature.toString()) {
          case "server_url": {
            const url = getUsageServerUrl(usageContext, scope, {
              isTest: usageContext.Test && true,
            });
            params.push(`\n${templateIndent(indent + 1)}.serverURL\(${url}\)`);
            break;
          }
          case "retries": {
            params.push(
              indentString(
                `\n${templateString("usage/retries.stmpl", {})}`,
                indent + 1,
              ),
            );
          }
        }
      }
    });
  }

  return params;
}

function templateUsageBuilderMethodParameters(
  usageContext: UsageContext,
  indent: number,
  finalCall: string,
  reqAssignment: string = "",
): string {
  const operation = usageContext.Operation;
  const methodParams = [];

  const ctx: TemplateValueContext = {
    usageContext: usageContext,
    operation: usageContext.Operation,
    isTest: usageContext.Test && true,
    test: usageContext.Test,
  };

  if (operationParametersFlattened(operation)) {
    methodParams.push(
      ...getOptionalUsageMethodParameters(usageContext, indent),
    );
    if (operation.Request) {
      for (const field of operation.Request.Field.Type.Fields) {
        if (context.Global.Config.NullFriendlyParameters) {
          if (hasConstValue(field)) {
            continue;
          }
        }
        let fieldExample = getOperationMethodFieldExample(usageContext, field);

        const value = templateValue(field, fieldExample, false, ctx);

        if (value == "") {
          continue;
        }

        methodParams.push(
          `\n${templateIndent(indent + 1)}.${sanitizeFieldName(
            field.Name,
          )}(${indentLines([value], indent + 1).trimStart()})`,
        );
      }
    }
  } else {
    if (operation.Request && !omitRequest(operation.Request) && reqAssignment) {
      const requestVarName = getRequestVariableName(usageContext.StepID);
      methodParams.push(
        `\n${templateIndent(indent + 1)}.request(${requestVarName})`,
      );
    }

    methodParams.push(
      ...getOptionalUsageMethodParameters(usageContext, indent),
    );
  }

  if (finalCall) {
    methodParams.push(`\n${templateIndent(indent + 1)}${finalCall}`);
  }

  return methodParams.join("");
}

registerTemplateFunc(
  "templateUsageBuilderMethodParameters",
  templateUsageBuilderMethodParameters,
);

function methodName(operation: Operation, scope: Scope): string {
  const parameters = methodParameters(operation, scope);
  let methodName = sanitizeMethodName(operation);
  if (parameters.length == 0) {
    // if no parameters we need to distinguish the service
    // method that returns a RequestBuilder (allowing
    // customization of timeouts, retries etc.) vs the
    // direct call that returns the Response.
    methodName += "Direct";
  }
  return methodName;
}

function templateRequestBuilderContent(
  op: SDKOperation,
  builderType: string,
): string {
  const sdkConfig = javaImportLocal("SDKConfiguration");

  const parameters = methodParameters(op.Operation, "operations");
  const options = supportedMethodOptions(op.Operation);
  return `${templateBuilderFields(parameters, 1)}
    private final ${sdkConfig} sdkConfiguration;
    private final ${javaImportLocal(
      "utils.Headers",
    )} _headers = new ${javaImportLocal("utils.Headers")}(); 

    public ${builderType}(${sdkConfig} sdkConfiguration) {
        this.sdkConfiguration = sdkConfiguration;
    }${requestBuilderMethods(parameters, builderType)}

${requestBuilderCallMethod(op, options)}${templateRequestDefaultValueSuppliers(
    op.Operation,
    "operations",
  )}`;
}

registerTemplateFunc(
  "templateRequestBuilderContent",
  templateRequestBuilderContent,
);

function parameterHasDefaultValue(p: JavaParam) {
  return p.Default && typeof p.Default.Value !== "undefined";
}

function templateBuilderFields(
  parameters: JavaParam[],
  indent: number,
): string {
  const optionsParamName = methodOptionsParameter().Name;
  const fields = parameters
    .filter((p) => p.Name != optionsParamName)
    .map((p) => {
      const fieldName = sanitizeFieldName(p.Name);
      let defaultValueExpression: String = undefined;
      if (parameterHasDefaultValue(p)) {
        defaultValueExpression = defaultValueJavaExpression(
          fieldName,
          p.Default.Value,
          p.EnhancedType,
        );
      }
      let assignment: string;
      if (p.Optional && p.Nullable) {
        if (defaultValueExpression) {
          assignment = ` = ${defaultValueExpression}`;
        } else {
          assignment = ` = ${javaImportJsonNullable()}.undefined()`;
        }
      } else if (p.Optional || p.Nullable) {
        if (defaultValueExpression) {
          assignment = ` = ${defaultValueExpression}`;
        } else {
          assignment = ` = ${javaImportOptional()}.empty()`;
        }
      } else if (p.Type.startsWith("java.util.Map")) {
        assignment = ` = new ${javaImportHashMap()}<>()`;
      } else {
        assignment = "";
      }
      return `\nprivate ${javaImport(
        toNonPrimitive(p.EnhancedType),
      )} ${fieldName}${assignment};`;
    })
    .join("");
  return indentString(fields, indent);
}

function templateRequestDefaultValueSuppliers(
  operation: Operation,
  scope: Scope,
): string {
  const lines = methodParameters(operation, scope)
    .filter((p) => parameterHasDefaultValue(p))
    .map((p) =>
      modelSingletonValueSupplier(p.Name, p.Default.Value, p.EnhancedType),
    );
  if (lines.length > 0) {
    lines.unshift("");
  }
  return indentLines(lines, 1);
}

function requestBuilderMethods(
  parameters: JavaParam[],
  builderType: string,
): string {
  const optionsParamName = methodOptionsParameter().Name;
  return parameters
    .filter((p) => p.Name != optionsParamName)
    .map((p) => {
      const fieldName = sanitizeFieldName(p.Name);
      let result = "";

      // write convenience builder method so that `value`
      // can be passed instead of only Optional.of(value)
      if (p.Optional && p.Nullable) {
        result += `

    public ${builderType} ${fieldName}(${javaImport(p.Type)} ${fieldName}) {
        ${javaImportUtils()}.checkNotNull(${fieldName}, "${fieldName}");
        this.${fieldName} = ${javaImportJsonNullable()}.of(${fieldName});
        return this;
    }`;
        if (parameterIsBigInteger(p)) {
          result += `
               
    public ${builderType} ${fieldName}(long ${fieldName}) {
        this.${fieldName} = ${javaImportJsonNullable()}.of(${javaImport(
          "java.math.BigInteger",
        )}.valueOf(${fieldName}));
        return this;
    }`;
        } else if (parameterIsBigDecimal(p)) {
          result += `
               
    public ${builderType} ${fieldName}(double ${fieldName}) {
        this.${fieldName} = ${javaImportJsonNullable()}.of(${javaImport(
          "java.math.BigDecimal",
        )}.valueOf(${fieldName}));
        return this;
    }`;
        }
      } else if (p.Optional || p.Nullable) {
        result += `
                
    public ${builderType} ${fieldName}(${javaImport(p.Type)} ${fieldName}) {
        ${javaImportUtils()}.checkNotNull(${fieldName}, "${fieldName}");
        this.${fieldName} = ${javaImportOptional()}.of(${fieldName});
        return this;
    }`;
        if (parameterIsBigInteger(p)) {
          result += `
               
    public ${builderType} ${fieldName}(long ${fieldName}) {
        this.${fieldName} = ${javaImportOptional()}.of(${javaImport(
          "java.math.BigInteger",
        )}.valueOf(${fieldName}));
        return this;
    }`;
        } else if (parameterIsBigDecimal(p)) {
          result += `
               
    public ${builderType} ${fieldName}(double ${fieldName}) {
        this.${fieldName} = ${javaImportOptional()}.of(${javaImport(
          "java.math.BigDecimal",
        )}.valueOf(${fieldName}));
        return this;
    }`;
        }
      } else {
        if (parameterIsBigInteger(p)) {
          result += `
               
    public ${builderType} ${fieldName}(long ${fieldName}) {
        this.${fieldName} = ${javaImport(
          "java.math.BigInteger",
        )}.valueOf(${fieldName});
        return this;
    }`;
        } else if (parameterIsBigDecimal(p)) {
          result += `
               
    public ${builderType} ${fieldName}(double ${fieldName}) {
        this.${fieldName} = ${javaImport(
          "java.math.BigDecimal",
        )}.valueOf(${fieldName});
        return this;
    }`;
        }
      }

      result += `\n
    public ${builderType} ${fieldName}(${javaImport(
      p.EnhancedType,
    )} ${fieldName}) {
        ${javaImportUtils()}.checkNotNull(${fieldName}, "${fieldName}");
        this.${fieldName} = ${fieldName};
        return this;
    }`;

      return result;
    })
    .join("");
}

type SDKOperation = {
  SDK: SDK;
  Operation: Operation;
  IsAsync: boolean;
};

function sdkOperation(
  sdk: SDK,
  operation: Operation,
  isAsync: boolean = false,
): SDKOperation {
  return {
    SDK: sdk,
    Operation: operation,
    IsAsync: isAsync,
  };
}

registerTemplateFunc("sdkOperation", sdkOperation);

// okhttp async arm only: call() body for the request builder. Mirrors the
// SDK-class emission in async-method.java.stmpl — sync handleResponse applied
// via Operations.applyBodyReadAsync; streaming operations wrap the raw
// response body in the blocking EventStream/JsonLStream and complete the
// user-visible future on Operations.streamCompletionExecutor() so blocking
// user continuations (stream iteration) never pin response-processing workers.
function requestBuilderAsyncCallBodyOkHttp(op: SDKOperation): string {
  const reqArg = op.Operation.Request ? "request" : "";
  const opsClass = javaImportLocal("operations.Operations");
  if (!isStreamingOperation(op.Operation)) {
    return `
        return ${opsClass}.relayCancel(${opsClass}.applyBodyReadAsync(operation.doRequest(${reqArg}),
            operation::handleResponse), operation);
    }`;
  }
  const itemType = getStreamingItemType(op.Operation);
  const typeRef = `new ${javaImport(
    "com.fasterxml.jackson.core.type.TypeReference",
  )}<${itemType}>() {
                        }`;
  let streamNew: string;
  if (isSSEOperation(op.Operation)) {
    const sentinel = getSSESentinel(op.Operation);
    const sentinelArg =
      sentinel === "null"
        ? `${javaImportOptional()}.empty()`
        : `${javaImportOptional()}.of(${sentinel})`;
    const dataRequiredArg = getSSEDataRequired(op.Operation)
      ? ""
      : `,\n                        false`;
    streamNew = `new ${javaImportLocal("utils.EventStream")}<${itemType}>(
                        response.rawResponse().body(),
                        ${typeRef},
                        ${javaImportUtils()}.mapper(),
                        ${sentinelArg}${dataRequiredArg})`;
  } else {
    streamNew = `new ${javaImportLocal("utils.JsonLStream")}<${itemType}>(
                        response.rawResponse().body(),
                        ${typeRef},
                        ${javaImportUtils()}.mapper())`;
  }
  return `
        return ${opsClass}.relayCancel(${opsClass}.applyBodyReadAsync(operation.doRequest(${reqArg}),
                operation::handleResponse)
                .thenApplyAsync(response -> ${streamNew}, ${opsClass}.streamCompletionExecutor()), operation);
    }`;
}

function requestBuilderCallMethod(
  op: SDKOperation,
  options: JavaParam[],
): string {
  const responseType = javaImportResponseReturnType(op.Operation, op.IsAsync);

  const code = [];
  const buildRequest = templateBuildRequest(op);
  if (buildRequest) {
    code.push(buildRequest + "\n");
  }

  code.push(`    public ${responseType} call() {`);
  code.push(templateIndent(2) + templateOperationInit(op, true, op.IsAsync));
  const initRequest = templateInitRequest(op);
  if (initRequest) {
    code.push(templateIndent(2) + initRequest);
  }

  if (op.IsAsync) {
    // Check if this is a streaming operation (SSE or JSONL)
    const isStreaming = isStreamingOperation(op.Operation);
    if (useOkHttp()) {
      // okhttp arm: CompletableFuture surface — sync handleResponse applied on
      // the body-read executor; streaming wraps the raw body in the blocking
      // EventStream/JsonLStream. No reactive types on this arm.
      code.push(requestBuilderAsyncCallBodyOkHttp(op));
    } else if (isStreaming) {
      const isSSE = isSSEOperation(op.Operation);

      if (isSSE) {
        const sentinel = getSSESentinel(op.Operation);
        const sseDataRequired = getSSEDataRequired(op.Operation);
        const streamMethod = "forSSE";
        const dataRequiredArg = sseDataRequired
          ? ""
          : `,\n                false`;
        const streamArgs = `,\n                ${javaImportUtils()}.mapper(),\n                ${sentinel}${dataRequiredArg}`;

        if (op.Operation.Request) {
          code.push(`
        return ${javaImportLocal("utils.reactive.EventStream")}.${streamMethod}(
                operation.doRequest(request).thenCompose(operation::handleResponse),
                new ${javaImport(
                  "com.fasterxml.jackson.core.type.TypeReference",
                )}<${templateStreamingTypeRefParam(op.Operation)}>() {
                }${streamArgs});
    }`);
        } else {
          code.push(`
        return ${javaImportLocal("utils.reactive.EventStream")}.${streamMethod}(
                operation.doRequest().thenCompose(operation::handleResponse),
                new ${javaImport(
                  "com.fasterxml.jackson.core.type.TypeReference",
                )}<${templateStreamingTypeRefParam(op.Operation)}>() {
                }${streamArgs});
    }`);
        }
      } else {
        // JSONL operation
        const streamMethod = "forJsonL";
        const streamArgs = `,\n                ${javaImportUtils()}.mapper()`;

        if (op.Operation.Request) {
          code.push(`
        return ${javaImportLocal("utils.reactive.EventStream")}.${streamMethod}(
                operation.doRequest(request).thenCompose(operation::handleResponse),
                new ${javaImport(
                  "com.fasterxml.jackson.core.type.TypeReference",
                )}<${templateStreamingTypeRefParam(op.Operation)}>() {
                }${streamArgs});
    }`);
        } else {
          code.push(`
        return ${javaImportLocal("utils.reactive.EventStream")}.${streamMethod}(
                operation.doRequest().thenCompose(operation::handleResponse),
                new ${javaImport(
                  "com.fasterxml.jackson.core.type.TypeReference",
                )}<${templateStreamingTypeRefParam(op.Operation)}>() {
                }${streamArgs});
    }`);
        }
      }
    } else {
      if (op.Operation.Request) {
        code.push(`
        return operation.doRequest(request)
            .thenCompose(operation::handleResponse);
    }`);
      } else {
        code.push(`
        return operation.doRequest()
            .thenCompose(operation::handleResponse);
    }`);
      }
    }
  } else {
    if (op.Operation.Request) {
      code.push(`
        return operation.handleResponse(operation.doRequest(request));
    }`);
    } else {
      code.push(`
        return operation.handleResponse(operation.doRequest());
    }`);
    }
  }

  if (op.Operation.Response.Type.Extensions?.Pagination) {
    if (op.IsAsync) {
      if (useOkHttp()) {
        // Async pagination methods (okhttp arm): callAsStream,
        // callAsStreamUnwrapped — CompletableFuture over the blocking
        // auto-pager, no reactive Publisher on this arm.
        code.push(templatePaginationCallAsFutureStream(op));
        code.push(callAsFutureStreamUnwrappedCode(op));
      } else {
        // Async pagination methods: callAsPublisher, callAsPublisherUnwrapped
        code.push(templatePaginationCallAsPublisher(op));
        code.push(callAsPublisherUnwrappedCode(op));
      }
    } else {
      // Sync pagination methods: callAsIterable, callAsStream, callAsStreamUnwrapped
      code.push(templatePaginationCallAsIterable(op));
      code.push(`
    /**
     * Returns a stream that performs next page calls till no more pages
     * are returned.
     **/  
    public ${javaImport(
      "java.util.stream.Stream",
    )}<${responseType}> callAsStream() {
        return ${javaImportStatic(
          "utils.Utils.toStream",
          true,
        )}(callAsIterable());
    }`);

      // will add empty string if cannot unwrap stream
      code.push(callAsStreamUnwrappedCode(op.Operation));
    }
  }

  return code.join("\n");
}

function templateOperationInit(
  op: SDKOperation,
  initOptions: boolean = true,
  isAsync: boolean = false,
): string {
  const code = [];
  if (initOptions) {
    const opts = templateRequestBuilderCallOptions(
      supportedMethodOptions(op.Operation),
      3,
    );
    if (opts) {
      code.push(opts);
    }
  }
  const resType = isAsync
    ? javaImportTypeAsync(op.Operation.Response.Type)
    : javaImportTypeMandatory(op.Operation.Response.Type);
  const opHelper = javaImportLocal(
    `operations.${templateOperationSimpleClassName(op.Operation)}`,
  );
  let opParams = operationParameters(op.Operation, isAsync);

  const reqParam = getRequestParameter(op.Operation, isAsync);
  const useConcreteType = hasPaginationURL(op.Operation);
  if (useConcreteType) {
    // Use concrete type for URL-paginated operations so we can access
    // baseUrl() and setUrlOverride() directly without casting.
    const syncOrAsync = isAsync ? "Async" : "Sync";
    code.push(`
        ${opHelper}.${syncOrAsync} operation
              = new ${opHelper}.${syncOrAsync}(${templateMethodCallArgs(
                opParams,
                true,
                36,
              )});`);
  } else if (reqParam) {
    const nullFriendly = context.Global.Config.NullFriendlyParameters;
    if (isAsync) {
      code.push(`
        ${javaImportStatic(
          "operations.Operations.AsyncRequestOperation",
          true,
        )}<${
          nullFriendly
            ? javaImportBoxed(javaImportNullableIfRequired(reqParam, true))
            : javaImportBoxed(reqParam.EnhancedType)
        }, ${javaImportBoxed(resType)}> operation
              = new ${opHelper}.Async(${templateMethodCallArgs(
                opParams,
                true,
                36,
              )});`);
    } else {
      code.push(`
        ${javaImportStatic("operations.Operations.RequestOperation", true)}<${
          nullFriendly
            ? javaImportBoxed(javaImportNullableIfRequired(reqParam, true))
            : javaImportBoxed(reqParam.EnhancedType)
        }, ${javaImportBoxed(resType)}> operation
              = new ${opHelper}.Sync(${templateMethodCallArgs(
                opParams,
                true,
                36,
              )});`);
    }
  } else {
    if (isAsync) {
      code.push(`
        ${javaImportStatic(
          "operations.Operations.AsyncRequestlessOperation",
          true,
        )}<${javaImportBoxed(resType)}> operation
            = new ${opHelper}.Async(${templateMethodCallArgs(
              opParams,
              true,
              36,
            )});`);
    } else {
      code.push(`
        ${javaImportStatic(
          "operations.Operations.RequestlessOperation",
          true,
        )}<${javaImportBoxed(resType)}> operation
            = new ${opHelper}.Sync(${templateMethodCallArgs(
              opParams,
              true,
              36,
            )});`);
    }
  }

  return code.join("\n");
}

registerTemplateFunc("templateOperationInit", templateOperationInit);

// For backward compatibility, provide templateAsyncOperationInit as a wrapper
function templateAsyncOperationInit(
  op: SDKOperation,
  initOptions: boolean = true,
): string {
  return templateOperationInit(op, initOptions, true);
}

registerTemplateFunc("templateAsyncOperationInit", templateAsyncOperationInit);

function templateInitRequest(op: SDKOperation, indent: number = 2): string {
  if (!operationParametersFlattened(op.Operation)) {
    return "";
  }
  const reqType = importedClass(op.Operation.Request.Field.Type);
  return `${reqType} request = buildRequest();`;
}

function templateBuildRequest(op: SDKOperation, indent: number = 2): string {
  if (!operationParametersFlattened(op.Operation)) {
    return "";
  }
  const reqType = importedClass(op.Operation.Request.Field.Type);
  const code = [
    `
    private ${reqType} buildRequest() {`,
  ];
  const defaults = templateSetRequestDefaults(op.Operation, 2);
  if (defaults) {
    code.push(`${templateIndent(indent)}${defaults}`);
  }
  const params = parametersForType(op.Operation.Request.Field.Type)
    .map((p) => sanitizeFieldName(p.Name))
    .join(`,\n${templateIndent(indent + 1)}`);
  code.push(`
        ${importedClass(
          op.Operation.Request.Field.Type,
        )} request = new ${importedClass(
          op.Operation.Request.Field.Type,
        )}(${params});`);
  code.push(`
        return request;
    }`);

  return code.join(`\n`);
}

function templatePaginationCallAsIterable(op: SDKOperation): string {
  const resType = javaImportTypeMandatory(op.Operation.Response.Type);
  const comments = `
    /**
    * Returns an iterable that performs next page calls till no more pages
    * are returned.
    *
    * <p>The returned iterable can be used in a for-each loop:
    * <pre><code>
    * for (${resType} page : builder.callAsIterable()) {
    *     // Process each page
    * }
    * </code></pre>
    * 
    * @return An iterable that can be used to iterate through all pages
    */`;
  const code = [
    `
    public ${javaImport("java.lang.Iterable")}<${resType}> callAsIterable() {`,
  ];
  code.push(templateOperationInit(op));
  const initRequest = templateInitRequest(op);
  if (initRequest) {
    code.push(initRequest);
  }
  if (op.Operation.Request?.Field?.Nullable) {
    // a nullable body may be absent; materialize an empty one to read
    // pagination defaults from
    const reqType = importedClass(op.Operation.Request.Field.Type);
    code.push(
      `${reqType} pagedRequest = request.orElse(null) == null ? new ${reqType}() : request.get();`,
    );
  }
  const [tracker, pageFetcher] = templatePaginatorArgs(op.Operation, false);
  code.push(`${javaImport("java.util.Iterator")}<${javaHttpResponseType(
    "java.io.InputStream",
  )}> iterator = new ${javaImportPaginationEntity("Paginator")}<>(
            request,
            ${tracker},
            ${pageFetcher});

        return () -> ${javaImportStatic(
          "utils.Utils.transform",
          true,
        )}(iterator, operation::handleResponse);
    }`);

  return comments + code.join("\n        ");
}

function asyncPaginatorSetupLines(
  op: SDKOperation,
  nullFriendlyBuilder: boolean = false,
): string[] {
  // Emits the `request` and `operation` local declarations shared by the
  // async pagination method non-null-friendly bodies (callAsPublisher /
  // callAsPublisherUnwrapped on the reactive arm, callAsStream /
  // callAsStreamUnwrapped on the okhttp arm).
  const operation = op.Operation;
  const lines: string[] = [];
  const reqParam = getRequestParameter(operation);
  if (operation.Request) {
    if (nullFriendlyBuilder) {
      lines.push(
        `        ${javaImportNullableIfRequired(
          reqParam!,
          true,
        )} request = this._buildRequest();`,
      );
      if (reqParam?.Nullable) {
        // a nullable body may be absent; materialize an empty one to read
        // pagination defaults from
        const reqType = javaImportType(operation.Request.Field.Type);
        lines.push(
          `        ${reqType} pagedRequest = request.orElse(null) == null ? new ${reqType}() : request.get();`,
        );
      }
    } else if (operationParametersFlattened(operation)) {
      lines.push(
        `        ${importedClass(
          operation.Request.Field.Type,
        )} request = this.buildRequest();`,
      );
    } else if (reqParam?.Nullable) {
      // a nullable body may be absent; materialize an empty one to read
      // pagination defaults from
      const reqType = javaImportType(operation.Request.Field.Type);
      lines.push(
        `        ${javaImportOptional()}<? extends ${reqType}> request = this.request;`,
      );
      lines.push(
        `        ${reqType} pagedRequest = request.orElse(null) == null ? new ${reqType}() : request.get();`,
      );
    } else {
      lines.push(
        `        ${javaImportType(
          operation.Request.Field.Type,
        )} request = this.request;`,
      );
    }
  }
  if (nullFriendlyBuilder) {
    if (supportedMethodOptions(operation).length > 0) {
      lines.push(
        `        ${javaImport(
          methodOptionsParameter().Type,
        )} options = optionsBuilder.build();`,
      );
    }
    lines.push(`${templateIndent(2)}${templateOperationInit(op, false, true)}`);
  } else {
    lines.push(`${templateIndent(2)}${templateOperationInit(op, true, true)}`);
  }
  return lines;
}

function templatePaginationCallAsPublisher(op: SDKOperation): string {
  const resType = javaImportTypeAsync(op.Operation.Response.Type);
  const comments = `
    /**
     * Returns a Publisher that performs next page calls till no more pages
     * are returned.
     *
     * <p>The returned publisher can be used with reactive frameworks:
     * <pre><code>
     * Publisher&lt;${resType}&gt; publisher = builder.callAsPublisher();
     * publisher.subscribe(new Subscriber&lt;${resType}&gt;() {
     *     // Handle onNext, onError, onComplete
     * });
     * </code></pre>
     *
     * @return A Publisher that emits pages asynchronously
     */`;
  const code = [
    `
    public ${javaImportReactivePublisher()}<${resType}> callAsPublisher() {`,
    ...asyncPaginatorSetupLines(op),
  ];
  const [tracker, pageFetcher] = templatePaginatorArgs(op.Operation, true);
  code.push(`
        ${javaImportFlowPublisher()}<${javaImport(
          "java.net.http.HttpResponse",
        )}<${javaImportLocal(
          "utils.Blob",
        )}>> asyncPaginator = new ${javaImportPaginationEntity(
          "AsyncPaginator",
        )}<>(
            request,
            ${tracker},
            ${pageFetcher});

        Flow.Publisher<${resType}> flowPublisher = ${javaImportStatic(
          "utils.reactive.ReactiveUtils.mapAsync",
          true,
        )}(asyncPaginator, operation::handleResponse);

        // Convert Flow.Publisher to Reactive Streams Publisher at the last stage
        return ${javaImport(
          "org.reactivestreams.FlowAdapters",
        )}.toPublisher(flowPublisher);
    }`);

  return comments + code.join("\n");
}

function callAsPublisherUnwrappedCode(op: SDKOperation): string {
  const operation = op.Operation;
  const output = getUnwrapStreamMetadata(operation, true);
  if (!output) {
    return "";
  }

  const { itemType, streamOps } = output;

  const [tracker, pageFetcher] = templatePaginatorArgs(operation, true);

  return `
    /**
     * Returns a flat Publisher of ${itemType} items. Subsequent pagination
     * calls are transparently handled by the SDK until no more pages are returned.
     **/
    public ${javaImportReactivePublisher()}<${itemType}> callAsPublisherUnwrapped() {
${asyncPaginatorSetupLines(op).join("\n")}

        // Get the internal Flow publisher first, then convert to Reactive Streams at the end
        ${javaImportFlowPublisher()}<${javaImportTypeAsync(
          operation.Response.Type,
        )}> internalPublisher = ${javaImportStatic(
          "utils.reactive.ReactiveUtils.mapAsync",
          true,
        )}(
            new ${javaImportPaginationEntity("AsyncPaginator")}<>(
                request,
                ${tracker},
                ${pageFetcher}),
            operation::handleResponse);

        Flow.Publisher<${javaImport(
          "java.util.List",
        )}<${itemType}>> items = ${javaImportStatic(
          "utils.reactive.ReactiveUtils.map",
          true,
        )}(internalPublisher,
                ${streamOps.join("\n                ")}
        );

        Flow.Publisher<${itemType}> flattenedFlow = ${javaImportStatic(
          "utils.reactive.ReactiveUtils.flatten",
          true,
        )}(items);

        // Convert Flow.Publisher to Reactive Streams Publisher at the last stage
        return ${javaImport(
          "org.reactivestreams.FlowAdapters",
        )}.toPublisher(flattenedFlow);
    }`;
}

// okhttp async arm only: CompletableFuture<Stream<...>> pagination surface.
// The first page is fetched asynchronously (sendAsync); consuming the stream
// pages through the remaining responses with the blocking Paginator on the
// consuming thread. No reactive Publisher is emitted on this arm. The
// user-visible future completes on Operations.streamCompletionExecutor():
// consumers commonly block in forEach inside continuations chained without
// *Async, which would pin response-completion workers (or an okhttp per-host
// call slot) and deadlock against the page fetches they wait on.
function templatePaginationCallAsFutureStream(
  op: SDKOperation,
  nullFriendlyBuilder: boolean = false,
): string {
  const resType = javaImportTypeAsync(op.Operation.Response.Type);
  const comments = `
    /**
     * Returns a CompletableFuture that completes with a stream performing
     * next page calls till no more pages are returned.
     *
     * <p>The first page is fetched asynchronously; consuming the returned
     * stream performs the subsequent page calls on the consuming thread:
     * <pre><code>
     * builder.callAsStream()
     *     .thenAccept(pages -&gt; pages.forEach(page -&gt; {
     *         // Process each page
     *     }));
     * </code></pre>
     *
     * @return A CompletableFuture that completes with the page stream once
     *         the first response arrives
     */`;
  const code = [
    `
    public ${javaImport("java.util.concurrent.CompletableFuture")}<${javaImport(
      "java.util.stream.Stream",
    )}<${resType}>> callAsStream() {`,
    ...asyncPaginatorSetupLines(op, nullFriendlyBuilder),
  ];
  const [tracker, pageFetcher] = templatePaginatorArgs(op.Operation, true);
  code.push(`
        ${javaImport("java.util.concurrent.CompletableFuture")}<${javaImport(
          "java.util.stream.Stream",
        )}<${javaHttpResponseType(
          "java.io.InputStream",
        )}>> pages = new ${javaImportPaginationEntity("AsyncPaginator")}<>(
            request,
            ${tracker},
            ${pageFetcher}).callAsStream();

        return ${javaImportLocal(
          "operations.Operations",
        )}.relayCancel(pages.thenApplyAsync(
            stream -> stream.map(operation::handleResponse),
            ${javaImportLocal(
              "operations.Operations",
            )}.streamCompletionExecutor()), operation);
    }`);

  return comments + code.join("\n");
}
registerTemplateFunc(
  "templatePaginationCallAsFutureStream",
  templatePaginationCallAsFutureStream,
);

// okhttp async arm only: flat-item analogue of callAsStreamUnwrapped.
function callAsFutureStreamUnwrappedCode(op: SDKOperation): string {
  const output = getUnwrapStreamMetadata(op.Operation);
  if (!output) {
    return "";
  }

  const { itemType, streamOps } = output;

  return `
    /**
     * Returns a CompletableFuture that completes with a flat stream of
     * ${itemType} items. Subsequent pagination calls are transparently
     * handled by the SDK until no more pages are returned.
     **/
    public ${javaImport("java.util.concurrent.CompletableFuture")}<${javaImport(
      "java.util.stream.Stream",
    )}<${itemType}>> callAsStreamUnwrapped() {
        return callAsStream()
            .thenApply(stream -> stream
${indentLines(streamOps, 5)});
    }`;
}
registerTemplateFunc(
  "callAsFutureStreamUnwrappedCode",
  callAsFutureStreamUnwrappedCode,
);

function paginationStreamCall(operation: Operation): string {
  // we don't want to affect the imports stack because we are not using the callAsStreamUnwrappedCode code
  javaImportsStack.push(new Imports("dummy", () => false, false));
  const result =
    callAsStreamUnwrappedCode(operation).length > 0
      ? ".callAsStreamUnwrapped()"
      : ".callAsStream()";
  javaImportsStack.pop();
  return result;
}
registerTemplateFunc("paginationStreamCall", paginationStreamCall);

interface PaginationMetadata {
  pageType: string;
  responseField: FieldDef;
  errorType: string;
  itemType?: string;
  itemFieldPath?: {
    name: string;
    optional: boolean;
    nullable: boolean;
  }[];
}

function getPaginationMetadata(
  operation: Operation,
  isAsync: boolean = false,
): PaginationMetadata | null {
  const nonStandardFields = operation.Response.Type.Fields.filter(
    (f: FieldDef) => !f.IsResponseMetadata,
  );

  const paginationConfig = operation.Response.Type.Extensions?.Pagination;
  if (!paginationConfig) {
    return null;
  }
  const outputs = paginationConfig.Outputs ?? {};
  const numPagesMentionedInOutputs = Object.values(outputs).some((v) =>
    v.toString().includes("numPages"),
  );

  let itemType = undefined;
  let itemFieldPath = undefined;
  // Find array fields in response structure
  for (const field of nonStandardFields) {
    if (field.Type.Type.toString() === "array") {
      // Direct array field
      itemType = javaImport(
        toNonPrimitive(sanitizeTypeMandatory(field.Type.ItemType)),
      );
      itemFieldPath = [
        {
          name: field.Name,
          optional: field.Optional,
          nullable: field.Nullable,
        },
      ];
      break;
    } else if (field.Type.Type.toString() === "class" && field.Type.Fields) {
      // Look for array fields in nested class
      for (const nestedField of field.Type.Fields) {
        if (nestedField.Type.Type.toString() === "array") {
          const fieldName = nestedField.Name;
          const isRelevantField =
            Object.values(outputs).some((v) =>
              v.toString().includes(fieldName),
            ) || numPagesMentionedInOutputs;

          if (isRelevantField) {
            itemType = javaImport(
              toNonPrimitive(sanitizeTypeMandatory(nestedField.Type.ItemType)),
            );
            itemFieldPath = [
              {
                name: field.Name,
                optional: field.Optional,
                nullable: field.Nullable,
              },
              {
                name: nestedField.Name,
                optional: nestedField.Optional,
                nullable: nestedField.Nullable,
              },
            ];
            break;
          }
        }
      }
    }
  }

  // find the response type based on the OAS ContentItem with the key "application/json"
  const responseField = operation.Response.Responses.find(
    (r) => !r.Error,
  )?.Content.find((c) => c.SerializationMethod === "json")?.Content;

  return {
    pageType: isAsync
      ? javaImportTypeAsync(operation.Response.Type)
      : fullClassName(operation.Response.Type),
    responseField,
    itemType,
    errorType: getDefaultErrorClassName(),
    itemFieldPath,
  };
}

registerTemplateFunc("getPaginationMetadata", getPaginationMetadata);

function templateRequestWrapperInitialization(
  request: RequestDef,
  indent: number = 2,
  useGetter?: boolean,
): string {
  const reqType = request.Field.Type;
  const requestType = importedClass(reqType);
  const fields = reqType.Fields.map(
    (f) =>
      `${templateIndent(indent + 1)}.${sanitizeFieldName(f.Name)}(${
        useGetter ? `${sanitizeFieldName(f.Name)}()` : sanitizeFieldName(f.Name)
      })`,
  ).join("\n");

  const lines = [
    `${templateIndent(indent)}${requestType} request = ${requestType}`,
    `${templateIndent(indent + 1)}.builder()`,
    fields,
    `${templateIndent(indent + 1)}.build();`,
  ];

  return lines.join("\n");
}

registerTemplateFunc(
  "templateRequestWrapperInitialization",
  templateRequestWrapperInitialization,
);

function derivePaginationInputParamType(
  type: TypeDef,
  paginationInput: PaginationInputs,
): string {
  const input = type.Fields.find((f) => f.Name == paginationInput.Name);
  if (input) {
    return `${javaImport(toNonPrimitive(sanitizeType(input.Type)))}`;
  }
  // Search inside nested body fields (e.g., when cursor is in a request body sub-object)
  for (const field of type.Fields) {
    if (field.Type?.Fields) {
      const nested = field.Type.Fields.find(
        (f) => f.Name == paginationInput.Name,
      );
      if (nested) {
        return `${javaImport(toNonPrimitive(sanitizeType(nested.Type)))}`;
      }
    }
  }
  return "";
}

interface UnwrapStreamMetadata {
  errorType: string;
  itemType: string;
  streamOps: string[];
}

function callAsStreamUnwrappedCode(operation: Operation): string {
  const output = getUnwrapStreamMetadata(operation);
  if (!output) {
    return "";
  }

  const { itemType, streamOps } = output;

  return `
    /**
     * Returns a stream that performs next page calls till no more pages
     * are returned. The elements of the stream are the list elements of
     * each response. 
     **/         
    public ${javaImport(
      "java.util.stream.Stream",
    )}<${itemType}> callAsStreamUnwrapped() {
        return callAsStream()
${indentLines(streamOps, 4)};
    }`;
}

function getUnwrapStreamMetadata(
  operation: Operation,
  isAsync: boolean = false,
): UnwrapStreamMetadata | null {
  // now we attempt to add a callAsStreamUnwrapped method
  // we only do so if we find a non-standard response field which itself
  // has exactly one array field for which either one of the output
  // values contains that field name or one output value contains
  // `numPages`. These checks should make it very unlikely that
  // we get the streaming field wrong.
  const paginationMetadata = getPaginationMetadata(operation, isAsync);
  if (!paginationMetadata || !paginationMetadata.itemType) {
    return null;
  }

  const { itemType, errorType, itemFieldPath } = paginationMetadata;

  // Whether the getter's return can be absent (needs stream-unwrapping)
  // under the configured getterStyle. Both optional and nullable fields can
  // yield an absent/null value.
  const canBeAbsent = (field: { optional: boolean; nullable: boolean }) =>
    isAlwaysOptionalGetter() || field.optional || field.nullable;
  // Converts a getter call into a Stream of its value(s).
  const streamOf = (
    expr: string,
    field: { optional: boolean; nullable: boolean },
  ) => {
    // Stream.ofNullable is Java 9+; Java8Compat.streamOfNullable replaces it.
    const streamOfNullable = (arg: string) =>
      isJava8()
        ? `${javaImportLocal("utils.Java8Compat")}.streamOfNullable(${arg})`
        : `${javaImport("java.util.stream.Stream")}.ofNullable(${arg})`;
    // Under "raw", optional and nullable getters both return @Nullable T
    if (isRawGetter() && (field.optional || field.nullable)) {
      return streamOfNullable(expr);
    }
    // Under presence-aware style a nullable getter returns JsonNullable<T>
    if (!isAlwaysOptionalGetter() && field.nullable) {
      return streamOfNullable(`${expr}.orElse(null)`);
    }
    // Optional.stream() is Java 9+; Java8Compat.stream(Optional) replaces it.
    if (isJava8() && (field.optional || isAlwaysOptionalGetter())) {
      return `${javaImportLocal("utils.Java8Compat")}.stream(${expr})`;
    }
    return `${expr}.stream()`;
  };

  const streamOps = [];
  if (isAsync) {
    //  template the lambda, give it a page and returns a list of items
    // process first itemFieldPath
    const [firstField, ...restFields] = itemFieldPath;
    if (canBeAbsent(firstField)) {
      streamOps.push(
        `page -> ${streamOf(
          `page.${sanitizeFieldName(firstField.name)}()`,
          firstField,
        )}`,
      );
    } else {
      streamOps.push(`page -> page.${sanitizeFieldName(firstField.name)}()`);
    }
    for (const field of restFields) {
      streamOps.push(
        `.flatMap(r -> ${streamOf(
          `r.${sanitizeFieldName(field.name)}()`,
          field,
        )})`,
      );
      if (canBeAbsent(field)) {
        streamOps.push(
          `.flatMap(${javaImport("java.util.Collection")}::stream)`,
        );
      }
    }
    streamOps.push(
      `.collect(${javaImport("java.util.stream.Collectors")}.toList())`,
    );
  } else {
    for (const field of itemFieldPath) {
      streamOps.push(
        `.flatMap(r -> ${streamOf(
          `r.${sanitizeFieldName(field.name)}()`,
          field,
        )})`,
      );
    }
    if (canBeAbsent(itemFieldPath[itemFieldPath.length - 1])) {
      streamOps.push(`.flatMap(${javaImport("java.util.Collection")}::stream)`);
    }
  }

  return {
    errorType,
    itemType,
    streamOps,
  };
}

registerTemplateFunc("getUnwrapStreamMetadata", getUnwrapStreamMetadata);

function templateSetRequestDefaults(
  operation: Operation,
  indent: number = 2,
): string {
  const lines = methodParameters(operation, "operations")
    .filter((p) => parameterHasDefaultValue(p))
    .flatMap((p) => templateSetDefault(sanitizeFieldName(p.Name)));

  return lines.join(`\n${templateIndent(indent)}`);
}

function nonConstFields(fields: FieldDef[]): FieldDef[] {
  return fields.filter((f) => !f.Const);
}

registerTemplateFunc("nonConstFields", nonConstFields);

// @ts-ignore
function templateStatusCodes(response: SubResponse): string {
  return response.Code.map((x) => `, "${x}"`) //
    .join("");
}
registerTemplateFunc("templateStatusCodes", templateStatusCodes);

function sortMethodParamsIntoCopy(params: FieldDef[]): FieldDef[] {
  // performs a shallow copy
  const copy = [...params];
  return sortMethodParams(copy);
}

registerTemplateFunc("sortMethodParamsIntoCopy", sortMethodParamsIntoCopy);

function operationParameters(
  operation: Operation,
  isAsync: boolean = false,
): JavaParam[] {
  const params: JavaParam[] = [
    {
      Type: `${templatePackageName()}.SDKConfiguration`,
      EnhancedType: `${templatePackageName()}.SDKConfiguration`,
      Name: "sdkConfiguration",
      Optional: false,
      Nullable: false,
      IncludedInSignature: true,
      Util: false,
      ParamType: "SDKConfiguration",
      Comment: "The SDK configuration",
      HasItemType: false,
      Final: true,
    },
  ];

  params.push(
    ...methodParameters(operation, operation.Scope)
      .filter((p) => p.IncludedInSignature)
      .filter((p) => !p.IsFlattened)
      .filter((p) => p.Name != "request"),
  );

  // Only add retryScheduler for async operations when retries are actually supported
  if (isAsync && operation.Extensions.Retries) {
    params.push({
      Type: "java.util.concurrent.ScheduledExecutorService",
      EnhancedType: "java.util.concurrent.ScheduledExecutorService",
      Name: "retryScheduler",
      Optional: false,
      Nullable: true,
      IncludedInSignature: true,
      Util: false,
      ParamType: "ScheduledExecutorService",
      Comment: "The scheduler for retrying requests",
      HasItemType: false,
      Final: true,
      Assignment: "sdkConfiguration.retryScheduler()",
    });
  }

  return params;
}

registerTemplateFunc("operationParameters", operationParameters);

function templateTrackerStrategy(operation: Operation): string[] {
  const paginationConfig = operation.Response.Type.Extensions?.Pagination;
  if (!paginationConfig) {
    return [];
  }

  // a nullable body arrives wrapped, so inputs are read off pagedRequest
  const requestIsNullable = Boolean(operation.Request?.Field?.Nullable);
  const requestVar = requestIsNullable ? "pagedRequest" : "request";

  const isNullableInput = (input: PaginationInputs) => {
    const field = findPaginationFieldDeep(
      operation.Request.Field.Type,
      input.Name,
    );
    return field ? field.Nullable : false;
  };

  // Unwraps a pagination input getter to a raw value with fallback,
  // accounting for the getter's shape under the configured getterStyle.
  const unwrapWithFallback = (
    input: PaginationInputs,
    value: string,
    fallback: string,
  ) => {
    if (isRawGetter()) {
      return `${javaImportOptional()}.ofNullable(${value}).orElse(${fallback})`;
    }
    if (isAlwaysOptionalGetter()) {
      return `${value}.orElse(${fallback})`;
    }
    // For JsonNullable<T> (required+nullable), .orElse(fallback) returns null
    // when the field is JsonNullable.of(null) because isPresent==true. Wrap in
    // Optional.ofNullable so both Optional<T> and JsonNullable<T> fall back.
    return `${javaImportOptional()}.ofNullable(${value}.orElse(null)).orElse(${fallback})`;
  };

  const getterCanBeAbsent = (input: PaginationInputs) =>
    isAlwaysOptionalGetter() || input.Optional || isNullableInput(input);

  // Normalizes a getter expression to a nullable raw reference: raw getters
  // already return null when absent, wrapped getters need unwrapping.
  const getterToNullable = (value: string) =>
    isRawGetter() ? value : `${value}.orElse(null)`;

  // The body wrapper field the input is nested in, when the request is a
  // wrapper and the input is not directly on it.
  const paginationBodyWrapper = (
    input: PaginationInputs,
  ): FieldDef | undefined => {
    const wrappedField = findWrappedRequestField(operation);
    if (!wrappedField) return undefined;
    if (
      operation.Request.Field.Type.Fields.some((f) => f.Name === input.Name)
    ) {
      return undefined;
    }
    return wrappedField;
  };

  const getInputValue = (input: PaginationInputs, fallback: string): string => {
    const body = paginationBodyWrapper(input);
    if (body && (body.Optional || body.Nullable)) {
      const bodyValue = `${requestVar}.${sanitizeFieldName(body.Name)}()`;
      const leaf = getterCanBeAbsent(input)
        ? `${javaImportOptional()}.ofNullable(${getterToNullable(
            `v.${input.Name}()`,
          )})`
        : `${javaImportOptional()}.ofNullable(v.${input.Name}())`;
      return `${javaImportOptional()}.ofNullable(${getterToNullable(
        bodyValue,
      )}).flatMap(v -> ${leaf}).orElse(${fallback})`;
    }
    let value = `${requestVar}.${input.Name}()`;
    if (body) {
      const unwrap = isAlwaysOptionalGetter() ? ".get()" : "";
      value = `${requestVar}.${sanitizeFieldName(body.Name)}()${unwrap}.${
        input.Name
      }()`;
    }
    if (!getterCanBeAbsent(input)) {
      return value;
    }
    return unwrapWithFallback(input, value, fallback);
  };

  const getInitialValue = (input: PaginationInputs) => {
    const defaults = getPaginationDefaults(operation);
    const fallback = defaults[input.Type.toString()] ?? 0;
    return getInputValue(input, `${fallback}L`);
  };

  const getMinItems = (input?: PaginationInputs) => {
    if (!input) return "1L";
    return getInputValue(input, "1L");
  };

  /**
   * Returns an inline expression that applies the pagination input to a request.
   * Uses standardised parameter names (req, pos) so callers can embed directly
   * into a pageFetcher lambda without post-processing.
   */
  const getInlineModifier = (input: PaginationInputs): string => {
    const withMethod = `with${sanitizeClassName(input.Name)}`;

    // Input is in a nested body field — drill into it
    const wrappedField = findWrappedRequestField(operation);
    if (
      wrappedField &&
      !operation.Request.Field.Type.Fields.some((f) => f.Name === input.Name)
    ) {
      const bodyGetter = sanitizeFieldName(wrappedField.Name);
      const withBody = `with${sanitizeClassName(wrappedField.Name)}`;
      if (wrappedField.Optional || wrappedField.Nullable) {
        // an omitted or null body must be materialized so the advanced
        // position lands on the next page's request
        const bodyClass = sanitizeTypeBase(wrappedField.Type, true);
        return `req.${withBody}(${javaImportOptional()}.ofNullable(${getterToNullable(
          `req.${bodyGetter}()`,
        )}).map(v -> v.${withMethod}(pos)).orElse(${bodyClass}.builder().${sanitizeFieldName(
          input.Name,
        )}(pos).build()))`;
      }
      const unwrap = isAlwaysOptionalGetter() ? ".get()" : "";
      return `req.${withBody}(req.${bodyGetter}()${unwrap}.${withMethod}(pos))`;
    }

    if (requestIsNullable) {
      // re-wrap to match the builder (JsonNullable vs Optional); orElse on
      // JsonNullable yields null even when present-but-null
      const wrapper = context.Global.Config.NullFriendlyParameters
        ? javaImportJsonNullable()
        : javaImportOptional();
      return `${wrapper}.of(${javaImportOptional()}.ofNullable(req.orElse(null)).orElse(${requestVar}).${withMethod}(pos))`;
    }

    return `req.${withMethod}(pos)`;
  };

  // Handle URL-based pagination
  if (hasPaginationURL(operation)) {
    return [
      `new ${javaImportPaginationEntity("URLTracker")}("${sanitizeJsonPath(
        paginationConfig.Outputs.NextURL,
      )}", operation.baseUrl()),`,
      // requestModifier is unused for URL pagination (pageFetcher handles it)
      `(req, url) -> req`,
    ];
  }

  // Handle page-based pagination
  const pageInput = getPaginationInput(paginationConfig, "page");
  if (pageInput) {
    const inputType = derivePaginationInputParamType(
      operation.Request.Field.Type,
      pageInput,
    );
    const initialPage = getInitialValue(pageInput);

    const tracker = paginationConfig.Outputs.Results
      ? `new ${javaImportPaginationEntity("ResultsPageTracker")}<>(
                "${sanitizeJsonPath(
                  paginationConfig.Outputs.Results,
                )}.length()",
                ${inputType}.class,
                ${initialPage},
                ${getMinItems(getPaginationInput(paginationConfig, "limit"))}),`
      : `new ${javaImportPaginationEntity("PageTracker")}<>(
                "${sanitizeJsonPath(paginationConfig.Outputs.NumPages)}",
                ${inputType}.class,
                ${initialPage}),`;

    return [tracker, getInlineModifier(pageInput)];
  }

  // Handle cursor-based pagination
  const cursorInput = getPaginationInput(paginationConfig, "cursor");
  if (cursorInput) {
    const paramType = derivePaginationInputParamType(
      operation.Request.Field.Type,
      cursorInput,
    );
    return [
      `new ${javaImportPaginationEntity("CursorTracker")}<>("${sanitizeJsonPath(
        paginationConfig.Outputs.NextCursor,
      )}", ${javaImport(paramType)}.class),`,
      getInlineModifier(cursorInput),
    ];
  }

  // Handle offset-based pagination
  const offsetInput = getPaginationInput(paginationConfig, "offset");
  if (offsetInput) {
    const paramType = derivePaginationInputParamType(
      operation.Request.Field.Type,
      offsetInput,
    );
    const initialOffset = getInitialValue(offsetInput);
    const minItems = getMinItems(getPaginationInput(paginationConfig, "limit"));

    return [
      `new ${javaImportPaginationEntity("OffsetTracker")}<>(
              "${sanitizeJsonPath(paginationConfig.Outputs.Results)}.length()",
              ${paramType}.class,
              ${initialOffset},
              ${minItems}),`,
      getInlineModifier(offsetInput),
    ];
  }

  return [];
}

registerTemplateFunc("templateTrackerStrategy", templateTrackerStrategy);

// Registered template functions for .stmpl templates to access paginator args
function templatePaginatorTracker(operation: Operation): string {
  return templatePaginatorArgs(operation, false)[0];
}
registerTemplateFunc("templatePaginatorTracker", templatePaginatorTracker);

function templatePaginatorFetcher(
  operation: Operation,
  isAsync: boolean = false,
): string {
  return templatePaginatorArgs(operation, isAsync)[1];
}
registerTemplateFunc("templatePaginatorFetcher", templatePaginatorFetcher);

/**
 * Returns [trackerInit, pageFetcherLambda] for use with the unified Paginator constructor.
 * The pageFetcher is a BiFunction<ReqT, ProgressParamT, Response> that handles both
 * request modification and fetching in a single step.
 */
function templatePaginatorArgs(
  operation: Operation,
  isAsync: boolean,
): [string, string] {
  const [rawTracker, requestModifier] = templateTrackerStrategy(operation);
  // Strip trailing comma from tracker — templateTrackerStrategy adds it for
  // the old 4-arg constructor, but we handle arg separation ourselves now.
  const tracker = rawTracker.replace(/,\s*$/, "");

  const doRequestExpr = (reqVar: string) =>
    isAsync
      ? `operation.doRequest(${reqVar})`
      : `${javaImportStatic(
          "utils.Exceptions.unchecked",
          true,
        )}(() -> operation.doRequest(${reqVar})).get()`;

  // For URL-based pagination, pass the URL as a parameter to the doRequest overload.
  // No mutation — the URL flows through as a pure function argument.
  if (hasPaginationURL(operation)) {
    const doRequestWithUrl = (reqVar: string, urlVar: string) =>
      isAsync
        ? `operation.doRequest(${reqVar}, ${urlVar})`
        : `${javaImportStatic(
            "utils.Exceptions.unchecked",
            true,
          )}(() -> operation.doRequest(${reqVar}, ${urlVar})).get()`;
    return [tracker, `(req, url) -> ${doRequestWithUrl("req", "url")}`];
  }

  // For cursor/page/offset pagination, compose the inline modifier with doRequest.
  const inlineModifier = requestModifier.trim();
  // `var` is Java10+ and infers the request type; In Java 8 we inline the ternary
  // to avoid having to spell out the exact type.
  if (isJava8()) {
    return [
      tracker,
      `(req, pos) -> ${doRequestExpr(
        `(pos == null ? req : ${inlineModifier})`,
      )}`,
    ];
  }
  return [
    tracker,
    `(req, pos) -> {
                var modifiedReq = pos == null ? req : ${inlineModifier};
                return ${doRequestExpr("modifiedReq")};
            }`,
  ];
}

function getResponseField(responses: SubResponse[]): FieldDef | null {
  const response = responses.find((r) => !r.Error);
  if (!response) {
    return null;
  }
  const content = response.Content.find(
    (c) => c.SerializationMethod === "json",
  );
  if (!content) {
    return null;
  }
  return content.Content;
}

registerTemplateFunc("getResponseField", getResponseField);

function setAdditionalProperties(reqType: TypeDef): string {
  const additionalProperties = reqType.Fields?.find(
    (f) => f.IsAdditionalProperties,
  );
  if (!additionalProperties) {
    return "";
  }
  return indentLines(
    [`.additionalProperties(${additionalProperties.Name})`],
    2,
  );
}

registerTemplateFunc("setAdditionalProperties", setAdditionalProperties);
