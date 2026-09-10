//@ts-ignore
function templateMethodResponseDocsSDKs(operation: Operation): SDKDocType[] {
  const link = `../../${getModelsLocation(
    operation.Response.Type.OutputLocation,
  )}/${sDKDocsFileFunc(sanitizeClassName(operation.Response.Type.Name))}.md`;

  let contents: SDKDocType[] = [
    {
      Name: "",
      Type: `Promise\\<[${sanitizeType(
        operation.Response.Type,
        true,
        "",
      )}](${link})\\>`,
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
          streamable: isMultipartFileField(parameter),
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
          streamable: operation.Request.RequestBody
            ? isMultipartFileField(operation.Request.RequestBody)
            : false,
        });
      } else {
        const link = `../../${getModelsLocation(
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

    if (operation.Security) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }
  }

  contents.push({
    Name: "options",
    Type: `RequestOptions`,
    Optional: true,
    Description: "Options for making HTTP requests.",
  });

  contents.push({
    Name: "options.fetchOptions",
    Type: "[RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)",
    Optional: true,
    Description:
      "Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All `Request` options, except `method` and `body`, are allowed.",
  });

  if (operation.Extensions.Retries) {
    contents.push({
      Name: "options.retries",
      Type: "[RetryConfig](../../lib/utils/retryconfig.md)",
      Optional: true,
      Description:
        "Enables retrying HTTP requests under certain failure conditions.",
    });
  }

  if (operation.Servers) {
    contents.push({
      Name: "options.serverURL",
      Type: `*string*`,
      Optional: true,
      Description: "An optional server URL to use.",
    });
  }

  return contents;
}

registerTemplateFunc(
  "templateMethodParametersDocsSDKs",
  templateMethodParametersDocsSDKs,
);

let hasTemplatedRetryConfig = false;

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
        `\`${sanitizeFieldName(parameter.Name)}\``,
        templateTypeMarkdown({
          fieldDef: parameter,
          streamable: isMultipartFileField(parameter),
        }),
        parameter.Optional ? ":heavy_minus_sign:" : ":heavy_check_mark:",
        description,
        example,
      ]);
    }
  } else {
    if (operation.Request) {
      const link = `../../${getModelsLocation(
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

    if (operation.Security) {
      addSecurityParamDocumentation(contents, operation);
    }
  }

  contents.push([
    "`options`",
    `RequestOptions`,
    ":heavy_minus_sign:",
    "Used to set various options for making HTTP requests.",
    "",
  ]);

  contents.push([
    "`options.fetchOptions`",
    `[RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)`,
    ":heavy_minus_sign:",
    "Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All `Request` options, except `method` and `body`, are allowed.",
    "",
  ]);

  if (operation.Extensions.Retries) {
    if (!hasTemplatedRetryConfig) {
      hasTemplatedRetryConfig = true;
      templateFile(
        "readme/retryconfig.stmpl",
        "docs/lib/utils/retryconfig.md",
        null,
      );
    }

    contents.push([
      "`options.retries`",
      "[RetryConfig](../../lib/utils/retryconfig.md)",
      ":heavy_minus_sign:",
      "Enables retrying HTTP requests under certain failure conditions.",
      "",
    ]);
  }

  if (operation.Servers) {
    contents.push([
      "`options.serverURL`",
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
    return "**Promise\\<void\\>**";
  }

  const dir = getModelsLocation(operation.Response.Type.OutputLocation);
  const name = sanitizeFileName(
    sanitizeClassName(operation.Response.Type.Name),
  );

  const link = `../../${dir}/${name}.md`;

  let documentation = `**Promise\\<[${sanitizeType(
    operation.Response.Type,
    true,
    "",
  )}](${link})\\>**`;

  return documentation;
}

registerTemplateFunc("templateMethodResponseDocs", templateMethodResponseDocs);

const streamableTypes = [
  "[File](https://developer.mozilla.org/en-US/docs/Web/API/File)",
  "[Blob](https://developer.mozilla.org/en-US/docs/Web/API/Blob)",
].join(" | ");

let hasTemplatedRFCDate = false;
// @ts-ignore
function templateTypeMarkdown({
  fieldDef,
  streamable = false,
  docsRoot = "../..",
  source = GeneratedDocSource.README,
}: TemplateTypeMarkdownParams): string {
  // If the field has a const value, display it as a literal type
  if (fieldDef.Const?.Value !== undefined) {
    const constValue = fieldDef.Const.Value;
    if (typeof constValue === "string") {
      return `*"${constValue}"*`;
    }
    return `*${constValue}*`;
  }

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
    case "enum": {
      const filename = classDocsFileName(fieldDef.Type.Name, source);
      const dir = `${docsRoot}/${getModelsLocation(
        fieldDef.Type.OutputLocation,
      )}`.replace(/\/+$/, "");
      const link = `${dir}/${filename}`;

      const val = `[${sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        "",
      )}](${link})`;
      return streamable ? `${streamableTypes} | ${val}` : val;
    }
    case "array":
      return `${itemMarkdown}[]`;
    case "map":
      return `Record<string, ${itemMarkdown}>`;
    case "date-time": {
      return `[Date](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date)`;
    }
    case "date": {
      if (source === GeneratedDocSource.SDKDOCS) {
        return sanitizeType(fieldDef.Type, fieldDef.Optional, "");
      }

      const link = `${docsRoot}/types/rfcdate.md`;

      if (!hasTemplatedRFCDate) {
        hasTemplatedRFCDate = true;
        templateFile("readme/rfcdate.stmpl", "docs/types/rfcdate.md", null);
      }

      return `[${sanitizeType(fieldDef.Type, fieldDef.Optional, "")}](${link})`;
    }
    case "response":
      return `[Response](https://developer.mozilla.org/en-US/docs/Web/API/Response)`;
    default:
      return `*${sanitizeType(fieldDef.Type, fieldDef.Optional, "")}*`;
  }
}

// @ts-ignore
function templateEnumDocs(type: TypeDef): string {
  const enumFormat = getEnumFormat(type);
  const isOpen = type.Enum?.Open;
  const isString = type.Enum?.Type.Type.toString() === "string";
  const tsType = isString ? "string" : "number";
  const fallbackType = `Unrecognized<${tsType}>`;

  if (enumFormat === "union") {
    const members = type.Enum?.Values.map((v) => {
      return isString ? `"${v}"` : `${v}`;
    });

    if (isOpen) {
      members.push(fallbackType);
    }

    let output = `\`\`\`typescript\n${members.join(" | ")}\n\`\`\``;

    return output;
  } else {
    const contents: string[][] = [["Name", "Value"]];

    let enumNames = getEnumNames(type);

    type.Enum.Values.forEach((value, index) => {
      contents.push(["`" + enumNames[index] + "`", value]);
    });

    if (type.Enum?.Open) {
      contents.push(["-", `\`${fallbackType}\``]);
    }

    return createMarkdownTable(contents);
  }
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
    const typename = sanitizeType(subType, false);
    seedFaker(typename);
    const example = templateModelUsage(typeDefToFieldDef(subType), 0);

    unionDocs += `### \`${typename}\`\n\n`;

    // Special case for EventStream types - just show type declaration
    if (subType.Type && subType.Type.valueOf() === "event-stream") {
      continue;
    }

    unionDocs += "```typescript\n";
    unionDocs +=
      formatUsageSnippetOutput(
        `const value: ${typename} = ${example};`,
      ).trimEnd() + `\n`;
    unionDocs += "```\n\n";
  }

  return unionDocs;
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);
