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

    const parameters = sortMethodFields(operation.Request.Field.Type.Fields);

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
      let path = "../..";
      if (operation.Request.Field.Type.OutputLocation) {
        path += "/" + operation.Request.Field.Type.OutputLocation;
      }

      const link = `${path}/${sanitizeFileName(
        sanitizeClassName(operation.Request.Field.Type.Name),
      )}.md`;

      contents.push([
        "`$request`",
        `[${sanitizeType(
          operation.Request.Field.Type,
          false,
          false,
          "",
          Qualification.USAGE,
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
      "`$serverURL`",
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
function templateMethodResponseDocs(operation: Operation): string {
  if (!operation.Response.Type) {
    return "";
  }
  let path = "../..";
  if (operation.Response.Type.OutputLocation) {
    path += "/" + operation.Response.Type.OutputLocation;
  }

  const link = `${path}/${sanitizeFileName(
    sanitizeClassName(operation.Response.Type.Name),
  )}.md`;

  let documentation = `**[${sanitizeType(
    operation.Response.Type,
    true,
    true,
    "",
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
      const relativeDocsPath =
        fieldDef.Type.OutputLocation == ""
          ? docsRoot
          : getTypeRelativeDocsPath(fieldDef.Type.OutputLocation);
      const filename = classDocsFileName(fieldDef.Type.Name, source);
      const link = `${relativeDocsPath}/${filename}`;
      return `[${sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        false,
        "",
      )}](${link})`;
    }
    case "array":
      return `array<${itemMarkdown}>`;
    case "map":
      return `array<string, ${itemMarkdown}>`;
    case "date":
    case "date-time": {
      return `[\\DateTime](https://www.php.net/manual/en/class.datetime.php)`;
    }
    case "response":
      return `[\\Psr\\Http\\Message\\ResponseInterface](https://www.php-fig.org/psr/psr-7/#33-psrhttpmessageresponseinterface)`;
    default:
      return `*${sanitizeType(fieldDef.Type, fieldDef.Optional, false, "")}*`;
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
function templateUnionDocs(type: TypeDef): string {
  let unionDocs = "";

  let types: TypeDef[] = type.AssociatedTypes;
  if (type.Discriminator) {
    types = type.Discriminator.Mapping.map((m) => m.Type);
  }

  for (const subType of types) {
    const typename = sanitizeType(
      subType,
      false,
      false,
      "documentation", // not "" to ensure full namespacing
      Qualification.TYPE,
    );

    unionDocs += `### \`${typename}\`\n\n`;
    unionDocs += "```php\n";
    unionDocs += `/**\n`;
    unionDocs += `* @var ${sanitizeType(
      subType,
      false,
      false,
      "",
      Qualification.DOCSTRING,
    )}\n`;
    unionDocs += `*/\n`;
    unionDocs += `${typename} $value = /* values here */\n`;
    unionDocs += "```\n\n";
  }

  return unionDocs;
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);
