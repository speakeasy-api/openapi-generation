// @ts-ignore
function addCSharpSecurityParamDocumentation(
  contents: string[][],
  operation: Operation,
) {
  if (!operation.Security) {
    return;
  }
  let description = "The security requirements to use for the request.";

  const comments = templateDocumentationComments(operation.Security.Comments);

  if (comments) {
    description += "\n\n" + comments;
  }

  const link = `../../${getModelsLocation(
    operation.Security.Type.OutputLocation,
  )}/${sanitizeFileName(sanitizeClassName(operation.Security.Type.Name))}.md`;

  contents.push([
    "`security`",
    `[${sanitizeClass(operation.Security.Type, "", false)}](${link})`,
    ":heavy_check_mark:",
    description,
    "",
  ]);
}

//@ts-ignore
function templateMethodParametersDocsSDKs(operation: Operation): SDKDocType[] {
  let contents: SDKDocType[] = [];

  if (operationParametersFlattened(operation)) {
    if (operation.Security) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }

    const parameters = sortMethodParams(operation.Request.Field.Type.Fields);

    for (const parameter of parameters) {
      const description = templateDocumentationComments(parameter.Comments);

      let example = "";
      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
      }

      contents.push({
        Name: `${sanitizeFieldName(parameter.Name)}`,
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
    if (operation.Request) {
      let reqType: string;
      if (operation.Request.Field.Type.Name == "") {
        reqType = templateTypeMarkdown({
          fieldDef: operation.Request.Field,
          source: GeneratedDocSource.SDKDOCS,
        });
      } else {
        const link = `../../${getModelsLocation(
          operation.Request.Field.Type.OutputLocation,
        )}/${sDKDocsFileFunc(
          sanitizeClassName(operation.Request.Field.Type.Name),
        )}.md`;
        reqType = `[${sanitizeType(operation.Request.Field.Type)}](${link})`;
      }
      contents.push({
        Name: "request",
        Type: reqType,
        Optional: false,
        Description: "The request object to use for the request.",
      });
    }

    if (operation.Security) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }
  }

  if (operation.Servers) {
    contents.push({
      Name: "serverURL",
      Type: `string`,
      Optional: true,
      Description: "An optional server URL to use.",
      Example: "http://localhost:8080",
    });
  }

  return contents;
}

//@ts-ignore
function templateMethodParametersDocs(operation: Operation): string {
  let contents: string[][] = [
    ["Parameter", "Type", "Required", "Description", "Example"],
  ];

  let validExamplesAdded = false;

  if (operationParametersFlattened(operation)) {
    if (operation.Security) {
      addCSharpSecurityParamDocumentation(contents, operation);
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
        `\`${sanitizeFieldName(parameter.Name)}\``,
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
      contents.push([
        "`request`",
        templateTypeMarkdown({
          fieldDef: operation.Request.Field,
        }),
        ":heavy_check_mark:",
        "The request object to use for the request.",
        "",
      ]);
    }

    if (operation.Security) {
      addCSharpSecurityParamDocumentation(contents, operation);
    }
  }

  if (operation.Servers) {
    contents.push([
      "`serverURL`",
      `*string*`,
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
function templateMethodResponseDocsSDKs(operation: Operation): SDKDocType[] {
  const link = `../../${getModelsLocation(
    operation.Response.Type.OutputLocation,
  )}/${sDKDocsFileFunc(sanitizeClassName(operation.Response.Type.Name))}.md`;

  let contents: SDKDocType[] = [
    {
      Name: "",
      Type: `[${sanitizeType(operation.Response.Type)}](${link})`,
      Optional: false,
    },
  ];

  return contents;
}

//@ts-ignore
function templateMethodResponseDocs(operation: Operation): string {
  if (!operation.Response.Type) {
    return "";
  }

  const link = `../../${getModelsLocation(
    operation.Response.Type.OutputLocation,
  )}/${sanitizeFileName(sanitizeClassName(operation.Response.Type.Name))}.md`;

  return `**[${sanitizeType(operation.Response.Type)}](${link})**`;
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

  switch (fieldDef.Type.Type.toString()) {
    case "class":
    case "enum":
    case "union": {
      const filename = classDocsFileName(fieldDef.Type.Name, source);
      const modelDir = getModelsLocation(fieldDef.Type.OutputLocation);
      const dir = `${docsRoot}/${modelDir}`.replace(/\/+$/, "");
      const link = `${dir}/${filename}`;

      return `[${sanitizeType(fieldDef.Type)}](${link})`;
    }
    case "array":
      return `List\<${itemMarkdown}>`;
    case "map":
      return `Dictionary\<String, ${itemMarkdown}>`;
    case "date-time": {
      return `[DateTime](https://learn.microsoft.com/en-us/dotnet/api/system.datetime?view=net-5.0)`;
    }
    case "date": {
      if (useNodatime()) {
        return `[LocalDate](https://nodatime.org/3.1.x/api/NodaTime.LocalDate.html)`;
      } else {
        return `[DateOnly](https://learn.microsoft.com/en-us/dotnet/api/system.dateonly?view=net-6.0)`;
      }
    }
    case "response":
      return `[HttpResponseMessage](https://learn.microsoft.com/en-us/dotnet/api/system.net.http.httpresponsemessage?view=net-5.0)`;
    default:
      return `*${sanitizeType(fieldDef.Type)}*`;
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
  const modelsNs = getModelNamespace(typeDef.OutputLocation);

  let code = `using ${modelsNs};\n\nvar value = ${className}.${enumNames[0]};`;
  if (isOpen) {
    const exampleValue = isString ? '"custom_value"' : "999";
    code += `\n\n// Open enum: use .Of() to create instances from custom ${
      isString ? "string" : "integer"
    } values`;
    code += `\nvar custom = ${className}.Of(${exampleValue});`;
  }

  return code;
}

registerTemplateFunc("templateModelSnippet", templateModelSnippet);

// @ts-ignore
function templateOptionalSymbol() {
  return "";
}

// @ts-ignore
function templateUnionDocs(typeDef: TypeDef): string {
  const className = sanitizeClassName(typeDef.Name);

  function templateUnionOption(typeName: string): string {
    const optionName = sanitizeClassName(typeName);
    return `### ${optionName}

\`\`\`csharp
${className}.Create${optionName}(/* values here */);
\`\`\`
`;
  }

  const options = typeDef.Discriminator
    ? typeDef.Discriminator.Mapping.map((m) =>
        templateUnionOption(getDiscriminatorDisplayName(m)),
      )
    : typeDef.AssociatedTypes.map((subType) =>
        templateUnionOption(sanitizeUnionTypeName(subType)),
      );

  return options.join("\n");
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);
