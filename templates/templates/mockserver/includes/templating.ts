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
    descriptionLines = sanitizeComments(sdkComments.Description).split("\n");
  }

  if (sdkComments.Summary) {
    firstLineParts.push(sanitizeComments(sdkComments.Summary));
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
    lines.push(...descriptionLines);
  }

  if (lines.length == 0) {
    return "";
  }

  lines = lines.map((line) => `// ${line}`);

  return indentLines(lines, indent) + "\n";
}

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  docgroup: string | null,
  title: string | null,
  indent: number,
  type: "method" | "field" | "type" | "const" | "var" | null = null,
  omitExternalDocs: boolean = false,
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

  if (!comments) {
    return cmsComment;
  }

  let lines = [];

  if (comments.ExternalDocs && !omitExternalDocs) {
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

  if (comments.Deprecated) {
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

function templateErrorMessage(errorType: TypeDef, errorObject: string): string {
  const errorMessageField = getErrorMessageField(errorType);

  if (errorMessageField) {
    if (errorMessageField.Optional) {
      return `if ${errorObject}.${sanitizeErrorFieldName(
        errorMessageField.Name,
      )} == nil {
    return "unknown error"
}

return *${errorObject}.${sanitizeErrorFieldName(errorMessageField.Name)}`;
    } else {
      return `return ${errorObject}.${sanitizeErrorFieldName(
        errorMessageField.Name,
      )}`;
    }
  } else {
    addImport("encoding/json");
    return `data, _ := json.Marshal(${errorObject})
return string(data)`;
  }
}

registerTemplateFunc("templateErrorMessage", templateErrorMessage);

/** Returns the template-ready header parameter example values. */
function templateHeaderParameterExampleValues(
  usageContext: UsageContext,
  param: ParamDef,
): string | undefined {
  const paramAnnotation = param.Field.Annotations?.Get(
    "param",
  ) as ParamAnnotation;

  if (paramAnnotation.Style != "simple") {
    return undefined;
  }

  const paramExampleValue = getExampleValue(
    findExampleByName(param.Examples ?? [], usageContext.ExampleName),
  );

  if (
    paramExampleValue === undefined ||
    !includeField(param.Field, paramExampleValue)
  ) {
    return undefined;
  }

  switch (param.Field.Type.Type.toString()) {
    case "array":
    case "set":
      return templateStringSlice([paramExampleValue.join(",")]);
    case "class":
    case "map":
      let values: string[] = [];

      if (paramAnnotation.Explode) {
        for (const key of Object.keys(paramExampleValue)) {
          values.push(`${key}=${paramExampleValue[key]}`);
        }
      } else {
        for (const key of Object.keys(paramExampleValue)) {
          values.push(key, paramExampleValue[key]);
        }
      }

      return templateStringSlice([values.join(",")]);
    default:
      return templateStringSlice([paramExampleValue]);
  }
}

/** Returns the template-ready path parameter example value. */
function templatePathParameterExampleValue(
  usageContext: UsageContext,
  param: ParamDef,
): string | undefined {
  const paramAnnotation = param.Field.Annotations?.Get(
    "param",
  ) as ParamAnnotation;

  if (paramAnnotation.Style != "simple") {
    return undefined;
  }

  const example = findExampleByName(
    param.Examples ?? [],
    usageContext.ExampleName,
  );
  if (!example) {
    return undefined;
  }

  // TODO figure out the right way to handle this
  if (example.Reference) {
    return undefined;
  }

  const paramExampleValue = getExampleValue(example, {
    usageContext: usageContext,
  });

  if (
    paramExampleValue === undefined ||
    !includeField(param.Field, paramExampleValue)
  ) {
    return undefined;
  }

  switch (param.Field.Type.Type.toString()) {
    case "array":
    case "set":
      return paramExampleValue.join(",");
    case "class":
    case "map":
      let result: string[] = [];

      for (const key of Object.keys(paramExampleValue)) {
        if (paramAnnotation.Explode) {
          result.push(`${key}=${paramExampleValue[key]}`);
        } else {
          result.push(`${key},${paramExampleValue[key]}`);
        }
      }

      return result.join(",");
    default:
      return paramExampleValue;
  }
}

/** Returns the template-ready query parameter example values. */
function templateQueryParameterExampleValues(
  usageContext: UsageContext,
  param: ParamDef,
): Record<string, string> | undefined {
  const paramAnnotation = param.Field.Annotations?.Get(
    "param",
  ) as ParamAnnotation;

  if (paramAnnotation.Style != "form") {
    return undefined;
  }

  const paramExampleValue = getExampleValue(
    findExampleByName(param.Examples ?? [], usageContext.ExampleName),
  );

  if (
    paramExampleValue === undefined ||
    !includeField(param.Field, paramExampleValue)
  ) {
    return undefined;
  }

  let result: Record<string, string> = {};

  switch (param.Field.Type.Type.toString()) {
    case "array":
    case "set":
      if (paramAnnotation.Explode) {
        result[param.Field.Name] = templateStringSlice(paramExampleValue);

        return result;
      }

      result[param.Field.Name] = templateStringSlice([
        paramExampleValue.join(","),
      ]);

      return result;
    case "class":
    case "map":
      if (paramAnnotation.Explode) {
        for (const key of Object.keys(paramExampleValue)) {
          result[key] = templateStringSlice([paramExampleValue[key]]);
        }

        return result;
      }

      let values: string[] = [];

      for (const key of Object.keys(paramExampleValue)) {
        values.push(key, paramExampleValue[key]);
      }

      result[param.Field.Name] = templateStringSlice([values.join(",")]);

      return result;
    default:
      result[param.Field.Name] = templateStringSlice([paramExampleValue]);

      return result;
  }
}

// @ts-ignore
function templateSDKOptionName(name: string): string {
  let reservedOptions = [
    "WithServerURL",
    "WithTemplatedServerURL",
    "WithClient",
  ];

  if (context.Global.AST.MainSDK.Security) {
    reservedOptions.push("WithSecurity");
  }
  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reservedOptions.push("WithServer");
    reservedOptions.push("WithTemplatedServer");
  } else {
    reservedOptions.push("WithServerIndex");
  }

  let optionName = `With${sanitizeClassName(name)}`;

  if (reservedOptions.includes(optionName)) {
    optionName = `WithGlobal${sanitizeClassName(name)}`;
  }

  return optionName;
}

registerTemplateFunc("templateSDKOptionName", templateSDKOptionName);

/** Returns a template-ready []string{} containing the given values. */
function templateStringSlice(values: string[]): string {
  let result: string = "[]string{";

  if (!Array.isArray(values)) {
    return result + "}";
  }

  values?.forEach((value, index) => {
    if (index != 0) {
      result = result.concat(", ");
    }

    result = result.concat(`"${value}"`);
  });

  return result.concat("}");
}

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

  let typeName = "";

  if (fieldDef.Type.Discriminator) {
    for (const mapping of fieldDef.Type.Discriminator.Mapping) {
      if (
        mapping.Type.Name === typeToUse.Name &&
        mapping.Type.Type.toString() === typeToUse.Type.toString()
      ) {
        typeName = sanitizeClassName(getDiscriminatorDisplayName(mapping));
        break;
      }
    }
  } else {
    typeName = sanitizeClassName(sanitizeUnionTypeName(typeToUse));
  }

  let value = templateValue(
    typeDefToFieldDef(typeToUse, fieldDef),
    selectedExample,
    false,
    additionalContext,
  );

  if (value != "") {
    value += ",";
  }

  const unionConstructor = [
    `${getAccessNamespace(
      "usage",
      fieldDef.Type.Scope.toString(),
      fieldDef.Type.OutputLocation,
    )}Create${sanitizeClassName(fieldDef.Type.Name)}${typeName}(
    ${value}
  )`,
  ].join("\n");

  // Import the actual output location for namespace types, not just the scope
  addImport(fieldDef.Type.OutputLocation, true);

  if (fieldDef.Optional || fieldDef.Nullable) {
    addImport("types");
    return `types.Pointer(${unionConstructor})`;
  }

  return unionConstructor;
}

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
function templateResponse(response: HandlerResponseContext): string {
  let body: any = undefined;
  if (response.Body) {
    body = getExampleValue(response.Body);
  }

  for (const assertion of response.ResponseBodyAssertions ?? []) {
    if (response.BodyContent) {
      const hasJP = assertion.Path != "" && assertion.Path != "/";

      if (hasJP) {
        if (body === undefined) {
          throw new Error(
            `missing response body example for ${response.UsageContext.Operation.OriginalID}`,
          );
        }

        const assertionValue = getExampleValue(assertion.Value);

        if (assertionValue !== undefined) {
          jsonpointer.set(
            body,
            assertion.Path,
            getExampleValue(assertion.Value),
          );
        }
      } else {
        body = getExampleValue(assertion.Value);
      }
    }
  }

  if (!response.BodySerialization) {
    return "";
  }

  switch (response.BodySerialization) {
    case "json":
      addImport("utils");

      return `${templateResponseBodyValue(response.BodyField, body, {
        usageContext: response.UsageContext,
        operation: response.UsageContext?.Operation,
      })}
respBodyBytes, err := utils.MarshalJSON(respBody, "", true)

if err != nil {
    http.Error(
        w,
        "Unable to encode response body as JSON: "+err.Error(),
        http.StatusInternalServerError,
    )
    return
}`;
    case "string":
      return `respBodyBytes := []byte(${templateStringValue(`${body}`, {
        ...response.BodyField,
        Optional: false,
        Nullable: false,
      })})`;
    case "raw":
      switch (response.BodyField?.Type.Type.toString()) {
        case "bytes":
          return `respBodyBytes := []byte(${templateStringValue(`${body}`, {
            ...response.BodyField,
            Optional: false,
            Nullable: false,
          })})`;
        case "response-stream":
          if (body === undefined) {
            addImport("crypto/rand");
            return `respBodyBytes := make([]byte, 16);
            _, _ = rand.Read(respBodyBytes)`;
          }

          return `respBodyBytes := ${templateValue(
            response.BodyField,
            body,
            false,
            {
              usageContext: response.UsageContext,
              operation: response.UsageContext?.Operation,
              isTest: true,
              isResponse: true,
            },
          )}`;
        default:
          throw new Error(
            `Unsupported response body type: ${response.BodyField?.Type.Type.toString()}`,
          );
      }

    // TODO will probably need to deal with streaming responses here
    default:
      throw new Error(
        `Unsupported response type: ${response.BodySerialization}`,
      );
  }
}
registerTemplateFunc("templateResponse", templateResponse);

// @ts-ignore
function templateFileToStringValue(
  filePath: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isRequest) {
    return sanitizeFieldName(fieldDef.Name); // TODO we will likely need to handle conflicts with field names at some point
  }

  if (additionalContext?.isTest) {
    filePath = `./testdata/${filePath}`;
  }

  addImport("mockserver/internal/handler/values");
  return `values.ReadFileToString("${filePath}")`;
}

// @ts-ignore
function templateStream(
  fieldDef: FieldDef,
  filePath: string,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (filePath) {
    if (additionalContext?.isRequest) {
      return sanitizePrivateFieldName(fieldDef.Name); // TODO we will likely need to handle conflicts with field names at some point
    }

    if (additionalContext?.isTest) {
      filePath = `./testdata/${filePath}`;
    }

    if (additionalContext?.isResponse) {
      addImport("mockserver/internal/handler/values");
      return `values.ReadFileToBytes("${filePath}")`;
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
