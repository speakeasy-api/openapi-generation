//@ts-ignore
function templateMethodParametersDocs(operation: Operation): string {
  let contents: string[][] = [
    ["Parameter", "Type", "Required", "Description", "Example"],
  ];

  let validExamplesAdded = false;

  if (operationParametersFlattened(operation)) {
    if (operation.Security) {
      addSecurityParamDocumentation(contents, operation);
    }

    const parameters = sortMethodParams(operation.Request.Field.Type.Fields);

    for (const parameter of parameters) {
      const comments = templateDocumentationComments(parameter.Comments);

      let description = comments || "N/A";

      let example = "";

      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
        validExamplesAdded = true;
      }

      contents.push([
        `\`${sanitizeMethodArgumentName(parameter.Name)}\``,
        templateTypeMarkdown({
          fieldDef: parameter,
        }),
        parameter.Optional ? ":heavy_minus_sign:" : ":heavy_check_mark:",
        description,
        example,
      ]);
    }
  } else {
    if (operation.Request) {
      const link = `../../models/${
        operation.Request.Field.Type.Scope
      }/${sanitizeFileName(
        sanitizeClassName(operation.Request.Field.Type.Name),
      )}.md`;

      contents.push([
        "`request`",
        `[${sanitizeType(
          operation.Request.Field.Type,
          false,
          false,
          operation.Scope,
        )}](${link})`,
        ":heavy_check_mark:",
        "The request object to use for the request.",
        "",
      ]);
    }

    if (operation.Security) {
      addSecurityParamDocumentation(contents, operation);
    }
  }

  if (operation.Servers) {
    contents.push([
      "`server_url`",
      `*String*`,
      ":heavy_minus_sign:",
      "An optional server URL to use.",
      "http://localhost:8080",
    ]);
  }

  if (!validExamplesAdded) {
    contents = contents.map((row) => {
      row.pop();
      return row;
    });
  }

  if (contents.length == 1) {
    return "";
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
    return "";
  }
  const link = `../../models/operations/${sanitizeFileName(
    sanitizeClassName(operation.Response.Type.Name),
  )}.md`;

  let documentation = `**[${sanitizeType(
    operation.Response.Type,
    true,
    false,
    operation.Scope,
  )}](${link})**`;

  return documentation;
}

registerTemplateFunc("templateMethodResponseDocs", templateMethodResponseDocs);

// @ts-ignore
function templateTypeMarkdown({
  fieldDef,
  docsRoot = "../..",
  source = GeneratedDocSource.README,
}: TemplateTypeMarkdownParams): string {
  let itemMarkdown = "";
  if (fieldDef.Type.ItemType != undefined) {
    itemMarkdown = templateTypeMarkdown({
      fieldDef: typeDefToFieldDef(fieldDef.Type.ItemType),
      docsRoot: docsRoot,
      source: source,
    });
  }

  switch (fieldDef.Type.Type.toString()) {
    case "class":
    case "enum":
    case "union": {
      const filename = classDocsFileName(fieldDef.Type.Name, source);
      const link = `${docsRoot}/models/${fieldDef.Type.Scope}/${filename}`;
      return `[${sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        fieldDef.Nullable,
        "",
      )}](${link})`;
    }
    case "array":
      return sorbetEnabled()
        ? `T::Array<${itemMarkdown}>`
        : `Crystalline::Array<${itemMarkdown}>`;
    case "map":
      return sorbetEnabled()
        ? `T::Hash[Symbol, ${itemMarkdown}]`
        : `Crystalline::Hash[Symbol, ${itemMarkdown}]`;
    case "date-time": {
      return `[Date](https://ruby-doc.org/stdlib-2.6.1/libdoc/date/rdoc/Date.html)`;
    }
    case "date": {
      return `[DateTime](https://ruby-doc.org/stdlib-2.6.1/libdoc/date/rdoc/DateTime.html)`;
    }
    case "response":
      return `[Faraday::Response](https://www.rubydoc.info/gems/faraday/Faraday/Response)`;
    default:
      return `*${sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        fieldDef.Nullable,
        "",
      )}*`;
  }
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
  const className = sanitizeClassName(typeDef.Name);
  const packageName = context.Global.Config.PackageName;

  let code = `require "${packageName}"\n\nvalue = ${className}::${enumNames[0]}`;
  if (isOpen) {
    const exampleValue = isString ? '"custom_value"' : "999";
    code += `\n\n# Open enum: use .deserialize() to create instances from custom ${
      isString ? "string" : "integer"
    } values`;
    code += `\ncustom = ${className}.deserialize(${exampleValue})`;
  }

  return code;
}

registerTemplateFunc("templateModelSnippet", templateModelSnippet);

// @ts-ignore
function templateUnionDocs(type: TypeDef): string {
  return "";
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);
