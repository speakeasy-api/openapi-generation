enum GeneratedDocSource {
  README = "readme",
  SDKDOCS = "sdkdocs",
}

type TemplateTypeMarkdownParams = {
  fieldDef: FieldDef;
  streamable?: boolean;
  docsRoot?: string;
  source?: GeneratedDocSource;
  isErrorType?: boolean;
};

function getSDKReadmeFileName(sdk: SDK): string {
  return `docs/sdks/${sanitizeMarkdownFileName(sdk.Type.Name)}/README.md`;
}

function createMarkdownTable(
  contents: string[][],
  uniformCellWidth: boolean = true, // legacy setting
): string {
  if (contents.length === 0) {
    return "";
  }

  function sanitizeContent(contents: string[][]): string[][] {
    return contents.map((row) =>
      row.map((cell) => cell.replaceAll("|", "\\|").replaceAll("\n", "<br/>")),
    );
  }

  if (!uniformCellWidth) {
    // make sure variable cell widths account for sanitization
    contents = sanitizeContent(contents);
  }

  const headers = contents.shift();
  const minCellWidth = 3;
  const cellWidths: number[] = headers.map((h) =>
    Math.max(minCellWidth, h.length),
  );
  let maxCellCount = headers.length;

  contents.forEach((row) => {
    if (row.length > maxCellCount) {
      maxCellCount = row.length;
    }

    row.forEach((value, i) => {
      if (i > cellWidths.length - 1) {
        cellWidths.push(Math.max(minCellWidth, value.length));
      } else if (value.length > cellWidths[i]) {
        cellWidths[i] = value.length;
      }
    });
  });

  if (uniformCellWidth) {
    // Note: sanitization used to be performed after calculating the max cell width.
    // Behavior is retained as such to prevent excessive diffs in docs/**.md files.
    contents = sanitizeContent(contents);

    const maxCellWidth = Math.max(...cellWidths);
    for (let i = 0; i < cellWidths.length; i++) {
      cellWidths[i] = maxCellWidth;
    }
  }

  let table = "";
  function fillRow(row: string[]) {
    row.forEach((cell, i) => {
      table += `| ${cell.padEnd(cellWidths[i])} `;
    });

    if (row.length < maxCellCount) {
      for (let i = row.length - 1; i < maxCellCount; i++) {
        table += `| ${"".padEnd(cellWidths[i])} `;
      }
    }
    table += "|\n";
  }

  // fill Headers
  fillRow(headers);
  headers.forEach((_header, i) => {
    table += `| ${"-".repeat(cellWidths[i])} `;
  });
  table += "|\n";

  // fill Rows
  for (const row of contents) {
    fillRow(row);
  }

  return table.trim();
}

function templateDocumentationComments(comments: CommentDef): string {
  if (!comments) {
    return "";
  }

  let lines: string[] = [];

  if (comments.Deprecated) {
    let deprecated = `: warning: ** DEPRECATED **: ${comments.DeprecationMessage}.`;

    if (comments.DeprecationReplacement) {
      const replacement = sanitizeDeprecationReplacement(
        comments.DeprecationReplacement,
        "field",
      );

      if (replacement) {
        deprecated += ` Use ${replacement} instead.`;
      }
    }

    lines.push(deprecated);
  }

  if (comments.Summary) {
    lines.push(comments.Summary);
  }

  if (comments.Description) {
    lines.push(comments.Description);
  }

  if (comments.ExternalDocs) {
    lines.push(
      `[${comments.ExternalDocs.Description}](${comments.ExternalDocs.URL})`,
    );
  }

  return lines.join("\n\n");
}

function sDKDocsFileFunc(name: string) {
  name = sanitizeName(name);
  return caser().ToPascal(name);
}

// Returns the Markdown filename associated with a class and sanitized based on source
function classDocsFileName(
  className: string,
  source: GeneratedDocSource,
): string {
  className = sanitizeClassName(className);
  if (source === GeneratedDocSource.SDKDOCS) {
    return `${sDKDocsFileFunc(className)}.md`;
  }
  return `${sanitizeFileName(className)}.md`;
}

// Template function to format type content for SDK Docs.
function templateObjectFieldsContentSDKDocs(
  type: TypeDef,
  language: string,
): SDKDocType[] {
  let contents: SDKDocType[] = [];

  for (const field of type.Fields) {
    const description = templateDocumentationComments(field.Comments);
    let example = "";
    if (field.Type.Examples.length > 0) {
      example = formatExamples(field.Type.Examples);
    }

    let markdownType = templateTypeMarkdown({
      fieldDef: field,
      streamable: isMultipartFileField(field),
      docsRoot: "../..",
      source: GeneratedDocSource.SDKDOCS,
    });

    contents.push({
      Name: `${sanitizeField(field)}`,
      Type: markdownType,
      Optional: field.Optional,
      Description: description != "" ? description : undefined,
      Example: example != "" ? example : undefined,
    });
  }

  return contents;
}

//@ts-ignore
function formatExamples(examples: Example[]): string {
  if (examples.length === 0) {
    return "";
  }

  if (examples.length === 1) {
    return examples[0].ToString();
  }

  // Show all examples with labels
  const formattedExamples = examples
    .map((ex, idx) => {
      const exampleName = ex.Name();
      const name =
        exampleName && exampleName !== "" ? exampleName : `Example ${idx + 1}`;
      const value = ex.ToString();
      return `**${name}:** ${value}`;
    })
    .join("<br/>");

  return formattedExamples;
}

registerTemplateFunc("formatExamples", formatExamples);

//@ts-ignore
function templateObjectFieldDocs(type: TypeDef): string {
  let contents: string[][] = [
    ["Field", "Type", "Required", "Description", "Example"],
  ];

  let validExamplesAdded = false;

  for (const field of type.Fields) {
    const comments = templateDocumentationComments(field.Comments);
    const description = comments || "N/A";

    const deprecationDelimiter = field.Comments?.Deprecated ? "~~" : "";
    const fieldName = sanitizeField(field, type);

    let example = "";
    if (field.Type.Examples.length > 0) {
      example = formatExamples(field.Type.Examples);
      validExamplesAdded = true;
    }

    const docsRoot = relativePath(
      getModelsLocation(type.OutputLocation) || "/",
      "/", // We consider the /docs folder the root folder for all doc files
    );

    contents.push([
      `${deprecationDelimiter}\`${fieldName}\`${deprecationDelimiter}`,
      templateTypeMarkdown({
        fieldDef: field,
        streamable: isMultipartFileField(field),
        docsRoot,
        source: GeneratedDocSource.README,
        isErrorType: type.Type.toString() === "error",
      }),
      field.Optional ? ":heavy_minus_sign:" : ":heavy_check_mark:",
      description,
      example,
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

registerTemplateFunc("templateObjectFieldDocs", templateObjectFieldDocs);

function addSecurityParamDocumentation(
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

  let path = "../..";
  if (operation.Security.Type.OutputLocation) {
    path += `/${operation.Security.Type.OutputLocation}`;
  }

  const link = `${path}/${sanitizeFileName(
    sanitizeClassName(operation.Security.Type.Name),
  )}.md`;

  contents.push([
    "`security`",
    `[${sanitizeClass(operation.Security.Type, "", false)}](${link})`,
    ":heavy_check_mark:",
    description,
    "",
  ]);
}

function addSecurityParamDocumentationForSDKDocs(
  contents: SDKDocType[],
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

  let path = "../..";
  if (operation.Security.Type.OutputLocation) {
    path += `/${operation.Security.Type.OutputLocation}`;
  }

  const link = `${path}/${sDKDocsFileFunc(
    sanitizeClassName(operation.Security.Type.Name),
  )}.md`;

  contents.push({
    Name: "security",
    Type: `[${sanitizeClass(operation.Security.Type, "", false)}](${link})`,
    Optional: false,
    Description: description,
  });
}
