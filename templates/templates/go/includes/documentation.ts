// @ts-ignore
function getDocsPath(relativeLocation: string, outputLocation: string) {
  if (outputLocation) {
    relativeLocation += `/${outputLocation}`;
  }

  return relativeLocation;
}

//@ts-ignore
function templateMethodResponseDocsSDKs(operation: Operation): SDKDocType[] {
  const link = `${getDocsPath(
    "../..",
    operation.Response.Type.OutputLocation,
  )}/${sDKDocsFileFunc(sanitizeClassName(operation.Response.Type.Name))}.md`;

  let contents: SDKDocType[] = [
    {
      Name: "",
      Type: `[${sanitizeType(operation.Response.Type, true, "")}](${link})`,
      Optional: false,
    },
    {
      Name: "",
      Type: "error",
      Optional: false,
    },
  ];

  return contents;
}

registerTemplateFunc(
  "templateMethodResponseDocsSDKs",
  templateMethodResponseDocsSDKs,
);

//@ts-ignore
function templateMethodParametersDocsSDKs(operation: Operation): SDKDocType[] {
  let contents: SDKDocType[] = [
    {
      Name: "ctx",
      Type: "[context.Context](https://pkg.go.dev/context#Context)",
      Optional: false,
      Description: "The context to use for the request.",
    },
  ];

  if (operationParametersFlattened(operation)) {
    if (
      operation.Security &&
      !getGoMethodArgumentOption(operation, operation.Security)
    ) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }

    const metadata = getGoOperationOptionMetadata(operation);
    const parameters = (
      metadata.mode === "pointers"
        ? sortMethodParams(operation.Request.Field.Type.Fields)
        : getArguments(operation).request
    ).filter((parameter) => !getGoMethodArgumentOption(operation, parameter));

    for (const parameter of parameters) {
      const description = templateDocumentationComments(parameter.Comments);

      let example = "";
      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
      }

      contents.push({
        Name: `${sanitizePrivateFieldName(parameter.Name)}`,
        Type: templateTypeMarkdown({
          fieldDef: parameter,
          source: GeneratedDocSource.SDKDOCS,
        }),
        Optional: parameter.Optional,
        Description: description != "" ? description : undefined,
        Example: example != "" ? example : undefined,
      });
    }
  } else {
    if (
      operation.Request &&
      !getGoMethodArgumentOption(operation, operation.Request.Field)
    ) {
      let reqType: string;
      if (operation.Request.Field.Type.Name == "") {
        reqType = templateTypeMarkdown({
          fieldDef: operation.Request.Field,
          source: GeneratedDocSource.SDKDOCS,
        });
      } else {
        const link = `${getDocsPath(
          "../..",
          operation.Request.Field.Type.OutputLocation,
        )}/${sDKDocsFileFunc(
          sanitizeClassName(operation.Request.Field.Type.Name),
        )}.md`;
        reqType = `[${sanitizeType(
          operation.Request.Field.Type,
          false,
          "",
        )}](${link})`;
      }
      contents.push({
        Name: "request",
        Type: reqType,
        Optional: false,
        Description: "The request object to use for the request.",
      });
    }

    if (
      operation.Security &&
      !getGoMethodArgumentOption(operation, operation.Security)
    ) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }
  }

  if (shouldDocumentGoMethodOptions(operation)) {
    contents.push({
      Name: "opts",
      Type: `[]${getAccessNamespace(
        "usage",
        "operations",
      )}${getGoMethodOptionType(operation)}`,
      Optional: true,
      Description: getGoMethodOptionsDescription(operation),
    });
  }

  return contents;
}

registerTemplateFunc(
  "templateMethodParametersDocsSDKs",
  templateMethodParametersDocsSDKs,
);

//@ts-ignore
function templateMethodParametersDocs(operation: Operation): string {
  let contents: string[][] = [
    ["Parameter", "Type", "Required", "Description", "Example"],
    [
      "`ctx`",
      "[context.Context](https://pkg.go.dev/context#Context)",
      ":heavy_check_mark:",
      "The context to use for the request.",
      "",
    ],
  ];

  let validExamplesAdded = false;

  if (operationParametersFlattened(operation)) {
    if (
      operation.Security &&
      !getGoMethodArgumentOption(operation, operation.Security)
    ) {
      addSecurityParamDocumentation(contents, operation);
    }

    const metadata = getGoOperationOptionMetadata(operation);
    const parameters = (
      metadata.mode === "pointers"
        ? sortMethodParams(operation.Request.Field.Type.Fields)
        : getArguments(operation).request
    ).filter((parameter) => !getGoMethodArgumentOption(operation, parameter));

    for (const parameter of parameters) {
      const comments = templateDocumentationComments(parameter.Comments);

      let description = comments || "N/A";

      let example = "";

      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
        validExamplesAdded = true;
      }

      contents.push([
        `\`${sanitizePrivateFieldName(parameter.Name)}\``,
        templateTypeMarkdown({
          fieldDef: parameter,
        }),
        parameter.Optional ? ":heavy_minus_sign:" : ":heavy_check_mark:",
        description,
        example,
      ]);
    }
  } else {
    if (
      operation.Request &&
      !getGoMethodArgumentOption(operation, operation.Request.Field)
    ) {
      const link = `${getDocsPath(
        "../..",
        operation.Request.Field.Type.OutputLocation,
      )}/${sanitizeFileName(
        sanitizeClassName(operation.Request.Field.Type.Name),
      )}.md`;

      contents.push([
        "`request`",
        `[${sanitizeType(operation.Request.Field.Type, false, "")}](${link})`,
        ":heavy_check_mark:",
        "The request object to use for the request.",
        "",
      ]);
    }

    if (
      operation.Security &&
      !getGoMethodArgumentOption(operation, operation.Security)
    ) {
      addSecurityParamDocumentation(contents, operation);
    }
  }

  if (shouldDocumentGoMethodOptions(operation)) {
    const metadata = getGoOperationOptionMetadata(operation);
    const optionType = `${getAccessNamespace(
      "usage",
      "operations",
    )}${getGoMethodOptionType(operation)}`;
    let optionTypeDocs = `\`[]${optionType}\``;
    if (metadata.mode !== "method-options") {
      const opsLocation = context.Global.Config.Imports.GetOperationsPath();
      const link = `${getDocsPath("../..", opsLocation)}/option.md`;
      optionTypeDocs = `[][${optionType}](${link})`;
    }

    contents.push([
      "`opts`",
      optionTypeDocs,
      ":heavy_minus_sign:",
      getGoMethodOptionsDescription(operation),
      "",
    ]);
  }

  if (!validExamplesAdded) {
    contents = contents.map((row) => {
      row.pop();
      return row;
    });
  }

  return createMarkdownTable(contents);
}

registerTemplateFunc(
  "templateMethodParametersDocs",
  templateMethodParametersDocs,
);

//@ts-ignore
function templateMethodResponseDocs(operation: Operation): string {
  if (!operation.Response.Type) {
    return "**error**";
  }

  const link = `${getDocsPath(
    "../..",
    operation.Response.Type.OutputLocation,
  )}/${sanitizeFileName(sanitizeClassName(operation.Response.Type.Name))}.md`;

  let documentation = `**[${sanitizeType(
    operation.Response.Type,
    true,
    "",
  )}](${link}), error**`;

  return documentation;
}

registerTemplateFunc("templateMethodResponseDocs", templateMethodResponseDocs);

// @ts-ignore
function listCustomTypes(): string[] {
  const customTypes = [];
  if ("date" in context.Global.AST.UsedTypes) {
    customTypes.push("date");
  }

  return customTypes;
}

registerTemplateFunc("listCustomTypes", listCustomTypes);

//@ts-ignore
function templateTypeMarkdown({
  fieldDef,
  docsRoot = "../..",
  source = GeneratedDocSource.README,
  isErrorType = false,
}: TemplateTypeMarkdownParams): string {
  // Error types don't use OptionalNullable wrapper (see error.go.stmpl),
  // they use plain pointer types instead.
  const nullableAndOptional =
    !isErrorType &&
    fieldDef.Nullable &&
    fieldDef.Optional &&
    context.Global.Config.NullableOptionalWrapper;

  const effectiveOptional = nullableAndOptional
    ? false
    : fieldDef.Optional || fieldDef.Nullable;

  let itemMarkdown = "";
  if (fieldDef.Type.ItemType != undefined) {
    itemMarkdown = templateTypeMarkdown({
      fieldDef: typeDefToFieldDef(
        fieldDef.Type.ItemType,
        undefined,
        undefined,
        fieldDef.Type.ContainsNull,
      ),
      docsRoot: docsRoot,
      source: source,
    });
  }

  let result: string;

  switch (fieldDef.Type.Type.toString()) {
    case "class":
    case "enum":
    case "union": {
      const filename = classDocsFileName(fieldDef.Type.Name, source);
      const dir = `${docsRoot}/${getModelsLocation(
        fieldDef.Type.OutputLocation,
      )}`.replace(/\/+$/, "");
      const link = `${dir}/${filename}`;

      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](${link})`;
      break;
    }
    case "array":
    case "set":
      result = `[]${itemMarkdown}`;
      break;
    case "map":
      result = `map[string]${itemMarkdown}`;
      break;
    case "date-time": {
      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](https://pkg.go.dev/time#Time)`;
      break;
    }
    case "date": {
      if (source === GeneratedDocSource.SDKDOCS) {
        result = sanitizeType(fieldDef.Type, effectiveOptional, "");
        break;
      }

      const link = `${docsRoot}/types/date.md`;

      templateFile("docs/types/date.stmpl", "docs/types/date.md", {});

      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](${link})`;
      break;
    }
    case "bigint": {
      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](https://pkg.go.dev/math/big#Int)`;
      break;
    }
    case "decimal": {
      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](https://pkg.go.dev/github.com/ericlagergren/decimal#Big)`;
      break;
    }
    case "request":
      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](https://pkg.go.dev/net/http#Request)`;
      break;
    case "response":
      result = `[${sanitizeType(
        fieldDef.Type,
        effectiveOptional,
        "",
      )}](https://pkg.go.dev/net/http#Response)`;
      break;
    default:
      result = `\`${sanitizeType(fieldDef.Type, effectiveOptional, "")}\``;
      break;
  }

  if (nullableAndOptional) {
    result = `optionalnullable.OptionalNullable[${result}]`;
  }

  return result;
}

// @ts-ignore
function templateEnumDocs(type: TypeDef): string {
  const contents: string[][] = [["Name", "Value"]];
  let enumNames = getEnumNames(type);

  type.Enum.Values.forEach((value, index) => {
    contents.push(["`" + enumNames[index] + "`", value]);
  });

  return createMarkdownTable(contents);
}

registerTemplateFunc("templateEnumDocs", templateEnumDocs);

// @ts-ignore
function templateModelSnippet(typeDef: TypeDef): string {
  if (typeDef.Type.toString() !== "enum") {
    return "";
  }

  const enumNames = getEnumNames(typeDef);
  if (enumNames.length === 0) {
    return "";
  }

  const isOpen = typeDef.Enum?.Open;
  const isString = typeDef.Enum?.Type.Type.toString() === "string";
  const namespace =
    sanitizeModelPackageName(typeDef.OutputLocation, true) + ".";
  const className = sanitizeClassName(typeDef.Name);
  const modelsLocation = getModelsLocation(typeDef.OutputLocation);
  const importPath =
    getRootModulePath() + (modelsLocation ? "/" + modelsLocation : "");

  let code = `import (\n\t"${importPath}"\n)\n\nvalue := ${namespace}${enumNames[0]}`;
  if (isOpen) {
    const exampleValue = isString ? '"custom_value"' : "999";
    code += `\n\n// Open enum: custom values can be created with a direct type cast`;
    code += `\ncustom := ${namespace}${className}(${exampleValue})`;
  }

  return code;
}

registerTemplateFunc("templateModelSnippet", templateModelSnippet);

function templateUnionDocsEntry(
  type: TypeDef,
  varName: string,
  namespace: string,
  className: string,
  subType: TypeDef,
  mapping: DiscriminatorMapping | null = null,
): string {
  const ctorPrefix = unionConstructorPrefix();
  const memberName = mapping
    ? getDiscriminatorDisplayName(mapping)
    : sanitizeUnionTypeName(subType);

  let ctorSuffix = sanitizeClassName(memberName);
  let value = `${sanitizeType(subType, false, "")}{/* values here */}`;

  if (isGenericUnion(type)) {
    ctorSuffix = "";
    const compositeKinds = new Set(["class", "array", "map", "set"]);
    if (!compositeKinds.has(subType.Type.toString())) {
      // The generic constructor infers T from the argument, so scalar members need an
      // explicit typed literal (e.g. `int64(0)`, not `0`) to satisfy the constraint.
      value = templateZero(subType, false, "", true);
    }
  }

  return `### ${sanitizeClassName(
    subType.Name,
  )}\n\n\`\`\`go\n${varName} := ${namespace}${ctorPrefix}${className}${ctorSuffix}(${value})\n\`\`\`\n\n`;
}

// @ts-ignore
function templateUnionDocs(type: TypeDef): string {
  const className = sanitizeClassName(type.Name);
  const varName = caser().ToCamel(className);
  const namespace = getAccessNamespace("usage", type.Scope.toString());

  let unionDocs = "";
  if (type.Discriminator) {
    type.Discriminator.Mapping.forEach((mapping) => {
      unionDocs += templateUnionDocsEntry(
        type,
        varName,
        namespace,
        className,
        mapping.Type,
        mapping,
      );
    });
  } else {
    type.AssociatedTypes.forEach((subType) => {
      unionDocs += templateUnionDocsEntry(
        type,
        varName,
        namespace,
        className,
        subType,
      );
    });
  }

  // Union Discrimination section
  const switchSnippet = buildUnionSwitchSnippet(type, varName, namespace);

  unionDocs += `## Union Discrimination\n\n`;
  unionDocs += `Use the \`Type\` field to determine which variant is active, then access the corresponding field:\n\n`;
  unionDocs += "```go\n";
  unionDocs += switchSnippet;
  unionDocs += "```\n";

  return unionDocs;
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);

// Builds a Go switch snippet for discriminating a union type.
// `accessPath` is the expression to access the union value (e.g. "myVar" or "res.Field").
// `namespace` is the package qualifier for enum constants (e.g. "shared.").
function buildUnionSwitchSnippet(
  unionTypeDef: TypeDef,
  accessPath: string,
  namespace: string,
  indent: string = "\t",
): string {
  const className = sanitizeClassName(unionTypeDef.Name);
  const unionType = getUnionTypeName(className, unionTypeDef.OutputLocation);
  const cases: string[] = [];

  if (unionTypeDef.Discriminator) {
    const enumNames = getUnionEnumNamesFromDiscriminator(
      unionType,
      unionTypeDef.Discriminator.Mapping,
      unionTypeDef.OutputLocation,
    );
    for (let i = 0; i < unionTypeDef.Discriminator.Mapping.length; i++) {
      const mapping = unionTypeDef.Discriminator.Mapping[i];
      const name = sanitizeUnionTypeName(mapping.Type);
      const memberField = sanitizeScopedFieldName(name, unionTypeDef);
      cases.push(
        `case ${namespace}${enumNames[i]}:\n${indent}${indent}// ${accessPath}.${memberField} is populated`,
      );
    }
  } else {
    const enumNames = getUnionEnumNamesFromAssociatedTypes(
      unionType,
      unionTypeDef.AssociatedTypes,
      unionTypeDef.OutputLocation,
    );
    for (let i = 0; i < unionTypeDef.AssociatedTypes.length; i++) {
      const name = sanitizeUnionTypeName(unionTypeDef.AssociatedTypes[i]);
      const memberField = sanitizeScopedFieldName(name, unionTypeDef);
      cases.push(
        `case ${namespace}${enumNames[i]}:\n${indent}${indent}// ${accessPath}.${memberField} is populated`,
      );
    }
  }

  if (unionTypeDef.IsUnionOpen) {
    cases.push(
      `default:\n${indent}${indent}// Unknown type - use ${accessPath}.GetUnknownRaw() for raw JSON`,
    );
  }

  let snippet = `switch ${accessPath}.Type {\n`;
  for (const c of cases) {
    snippet += `${indent}${c}\n`;
  }
  snippet += "}\n";

  return snippet;
}

// Finds the union TypeDef in a response and returns it along with its access path,
// or null if the response doesn't contain a union.
function findResponseUnion(
  operation: Operation,
  stepID: string,
): { unionType: TypeDef; accessPath: string; namespace: string } | null {
  const responseFormat = getResponseFormat();
  const resVar = getResponseVariableName(stepID);

  // Find the first non-error, non-stream response content
  let contentType: TypeDef | null = null;
  let fieldName = "";

  for (const response of operation.Response.Responses) {
    if (response.Error) continue;
    for (const content of response.Content) {
      if (
        content.SerializationMethod === "eventstream" ||
        content.SerializationMethod === "jsonl"
      ) {
        continue;
      }
      if (content.UsageExample || !contentType) {
        contentType = content.Content.Type;
        fieldName = content.Content.Name;
      }
    }
  }

  if (!contentType) return null;

  // Build base access path
  let accessPath: string;
  if (responseFormat !== "flat" && fieldName) {
    accessPath = `${resVar}.${sanitizeFieldName(fieldName)}`;
  } else {
    accessPath = resVar;
  }

  // Case 1: response field IS the union directly
  if (contentType.Type.toString() === "union") {
    return {
      unionType: contentType,
      accessPath,
      namespace: getAccessNamespace("usage", contentType.Scope.toString()),
    };
  }

  // Case 2: response field is a class wrapping a union
  if (contentType.Type.toString() === "class") {
    for (const field of contentType.Fields) {
      if (field.Type.Type.toString() === "union") {
        return {
          unionType: field.Type,
          accessPath: `${accessPath}.${sanitizeFieldName(field.Name)}`,
          namespace: getAccessNamespace("usage", field.Type.Scope.toString()),
        };
      }
    }
  }

  return null;
}

// @ts-ignore
function responseHasUnionType(operation: Operation): boolean {
  return findResponseUnion(operation, "") !== null;
}
registerTemplateFunc("responseHasUnionType", responseHasUnionType);

// @ts-ignore
function templateResponseUnionDiscrimination(
  operation: Operation,
  stepID: string,
): string {
  const result = findResponseUnion(operation, stepID);
  if (!result) return "";

  addImport(result.unionType.Scope.toString());

  const indent = "    ";
  const prefix = "        ";
  const switchSnippet = buildUnionSwitchSnippet(
    result.unionType,
    result.accessPath,
    result.namespace,
    indent,
  );

  // Indent the whole snippet to match the usage template's nesting level
  return switchSnippet
    .split("\n")
    .map((line) => (line ? `${prefix}${line}` : line))
    .join("\n");
}
registerTemplateFunc(
  "templateResponseUnionDiscrimination",
  templateResponseUnionDiscrimination,
);
