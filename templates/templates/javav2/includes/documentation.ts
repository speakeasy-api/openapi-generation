//@ts-ignore
function templateMethodParametersDocsSDKs(operation: Operation): SDKDocType[] {
  let contents: SDKDocType[] = [];

  if (operationParametersFlattened(operation)) {
    if (operation.Security) {
      addSecurityParamDocumentationForSDKDocs(contents, operation);
    }

    const parameters = operation.Request.Field.Type.Fields;

    for (const parameter of parameters) {
      const description = templateDocumentationComments(parameter.Comments);

      let example = "";
      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
      }

      const useSetterType =
        context.Global.Config.NullFriendlyParameters &&
        context.Global.Config.ShowSetterGetterTypesInDocs;
      const paramTypeMarkdown = useSetterType
        ? templateSetterTypeDoc(parameter, "../..", GeneratedDocSource.SDKDOCS)
        : templateTypeMarkdown({
            fieldDef: parameter,
            source: GeneratedDocSource.SDKDOCS,
          });

      contents.push({
        Name: `${sanitizeFieldName(parameter.Name)}`,
        Type: paramTypeMarkdown,
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
        const scopePath = getScopePath(operation.Request.Field.Type.Scope);
        const link = `../../${scopePath}/${sDKDocsFileFunc(
          sanitizeClassName(operation.Request.Field.Type.Name),
        )}.md`;
        reqType = `[${unqualifyType(
          sanitizeType(operation.Request.Field.Type, false, false, false),
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

  if (operation.Servers) {
    contents.push({
      Name: "serverURL",
      Type: `String`,
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
      addSecurityParamDocumentation(contents, operation);
    }

    const parameters = operation.Request.Field.Type.Fields;

    for (const parameter of parameters) {
      const comments = templateDocumentationComments(parameter.Comments);

      let description = comments || "N/A";

      let example = "";

      if (parameter.Type.Examples.length > 0) {
        example = formatExamples(parameter.Type.Examples);
        validExamplesAdded = true;
      }

      const useSetterType =
        context.Global.Config.NullFriendlyParameters &&
        context.Global.Config.ShowSetterGetterTypesInDocs;
      const paramTypeMarkdown = useSetterType
        ? templateSetterTypeDoc(parameter, "../..", GeneratedDocSource.README)
        : templateTypeMarkdown({ fieldDef: parameter });

      contents.push([
        `\`${sanitizeFieldName(parameter.Name)}\``,
        paramTypeMarkdown,
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
        `[${unqualifyType(
          sanitizeTypeMandatory(operation.Request.Field.Type),
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
      "`serverURL`",
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
function templateMethodResponseDocsSDKs(operation: Operation): SDKDocType[] {
  const scopePath = getScopePath(operation.Response.Type.Scope);
  const link = `../../${scopePath}/${sDKDocsFileFunc(
    sanitizeClassName(operation.Response.Type.Name),
  )}.md`;

  let contents: SDKDocType[] = [
    {
      Name: "",
      Type: `[${unqualifyType(
        sanitizeType(operation.Response.Type, true, false, false),
      )}](${link})`,
      Optional: false,
    },
  ];

  return contents;
}

//@ts-ignore
function templateMethodResponseDocs(operation: Operation): string {
  const scopePath = getScopePath(operation.Response.Type.Scope);
  const link = `../../${scopePath}/${sanitizeFileName(
    sanitizeClassName(operation.Response.Type.Name),
  )}.md`;

  let documentation = `**[${unqualifyType(
    sanitizeType(operation.Response.Type, false, false, false),
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
      const scopePath = getScopePath(fieldDef.Type.Scope);
      const link = `${docsRoot}/${scopePath}/${filename}`;
      return ghMarkdownEscape(
        `[${unqualifyType(
          sanitizeType(
            fieldDef.Type,
            fieldDef.Optional,
            fieldDef.Nullable,
            false,
          ),
        )}](${link})`,
      );
    }
    case "array":
      return ghMarkdownEscape(`List<${itemMarkdown}>`);
    case "map":
      return ghMarkdownEscape(`Map<String, ${itemMarkdown}>`);
    case "date-time": {
      return `[OffsetDateTime](https://docs.oracle.com/javase/8/docs/api/java/time/OffsetDateTime.html)`;
    }
    case "date": {
      return `[LocalDate](https://docs.oracle.com/javase/8/docs/api/java/time/LocalDate.html)`;
    }
    case "response":
      if (useOkHttp()) {
        return ghMarkdownEscape(
          `[HttpResponse<${
            asyncEnabled() ? "?" : "InputStream"
          }>](${docsRoot}/../${getSourceDirectory()}/utils/transport/HttpResponse.java)`,
        );
      }
      return ghMarkdownEscape(
        `[HttpResponse<${
          asyncEnabled() ? "?" : "InputStream"
        }>](https://docs.oracle.com/en/java/javase/11/docs/api/java.net.http/java/net/http/HttpResponse.html)`,
      );
    case "request-stream":
      return useBlob()
        ? `[Blob](${docsRoot}/../${getSourceDirectory()}/utils/Blob.java)`
        : `byte[]`;
    default:
      return ghMarkdownEscape(
        `*${unqualifyType(
          sanitizeType(
            fieldDef.Type,
            fieldDef.Optional,
            fieldDef.Nullable,
            false,
          ),
        )}*`,
      );
  }
}

function ghMarkdownEscape(s: string): string {
  return s.replaceAll("<", "\\<");
}

/**
 * Returns the base (unwrapped) type markdown for a field,
 * without Optional/JsonNullable wrapping.
 */
function templateBaseTypeMarkdown(
  field: FieldDef,
  docsRoot: string,
  source: GeneratedDocSource,
): string {
  const baseFieldDef = typeDefToFieldDef(field.Type, field, false, false);
  return templateTypeMarkdown({
    fieldDef: baseFieldDef,
    streamable: isMultipartFileField(field),
    docsRoot,
    source,
  });
}

/**
 * Returns the setter type for documentation in null-friendly mode.
 * Setters accept the raw unwrapped type; nullable/optional fields are @Nullable.
 */
function templateSetterTypeDoc(
  field: FieldDef,
  docsRoot: string,
  source: GeneratedDocSource,
): string {
  const param = parameterForField(field);
  const variant = getFieldVariant(param);
  const baseMarkdown = templateBaseTypeMarkdown(field, docsRoot, source);

  switch (variant) {
    case "nullable-optional-default":
    case "nullable-optional":
    case "nullable-only":
    case "optional-only":
      return `@Nullable ${baseMarkdown}`;
    case "required":
      return baseMarkdown;
  }
}

/**
 * Returns the getter return type for documentation in null-friendly mode.
 * Shows JsonNullable<T>, Optional<T>, or raw T depending on the field variant.
 */
function templateGetterTypeDoc(
  field: FieldDef,
  type: TypeDef,
  docsRoot: string,
  source: GeneratedDocSource,
): string {
  const param = parameterForField(field);
  const variant = getFieldVariant(param);
  const baseMarkdown = templateBaseTypeMarkdown(field, docsRoot, source);

  // Discriminator field getters are independent of getterStyle
  if (isDiscriminatorFieldName(type, field.Name)) {
    return baseMarkdown;
  }

  const getterStyle = fieldGetterStyle(field);

  if (isRawGetter(getterStyle)) {
    return variant === "required" ? baseMarkdown : `@Nullable ${baseMarkdown}`;
  }
  if (isAlwaysOptionalGetter(getterStyle)) {
    return `Optional\\<${baseMarkdown}>`;
  }

  switch (variant) {
    case "nullable-optional-default":
    case "nullable-optional":
    case "nullable-only":
      return `JsonNullable\\<${baseMarkdown}>`;
    case "optional-only":
      return `Optional\\<${baseMarkdown}>`;
    case "required":
      return baseMarkdown;
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

// Strips fully-qualified class names from generated Java code for documentation
// readability. SDK model FQCNs are collected as referenced types for cross-linking.
function simplifyDocExample(code: string): {
  simplified: string;
  referencedTypes: { className: string; scopePath: string }[];
} {
  const pkgPrefix = templatePackageName();
  const escapedPrefix = pkgPrefix.replace(/\./g, "\\.");
  const referencedTypes: { className: string; scopePath: string }[] = [];
  const seen = new Set<string>();

  // Strip SDK model FQCNs and collect model references
  const sdkRegex = new RegExp(
    escapedPrefix + "\\.([a-z][a-z0-9_.]*)\\.([A-Z][A-Za-z0-9_]*)",
    "g",
  );
  let simplified = code.replace(sdkRegex, (_, scopeDots, clsName) => {
    if (!seen.has(clsName)) {
      seen.add(clsName);
      referencedTypes.push({
        className: clsName,
        scopePath: scopeDots.replace(/\./g, "/"),
      });
    }
    return clsName;
  });

  // Strip remaining JDK FQCNs (java.util.Map, java.time.LocalDate, etc.)
  // Word boundary prevents matching partial words (e.g. Enum.SIXTY_NINE)
  simplified = simplified.replace(
    /\b([a-z][a-z0-9_]*\.)+([A-Z][A-Za-z0-9_]*)/g,
    "$2",
  );

  return { simplified, referencedTypes };
}

// Renders a "Referred Types" section linking to models referenced in a code
// example. Inline for 1-2 refs, bulleted list for 3+.
function renderReferredTypes(
  referencedTypes: { className: string; scopePath: string }[],
  excludeTypeName: string,
): string {
  const related = referencedTypes.filter(
    (r) => r.className !== excludeTypeName,
  );
  if (related.length === 0) return "";

  const links = related.map((r) => {
    const fn = classDocsFileName(r.className, GeneratedDocSource.README);
    return `[${r.className}](../../${r.scopePath}/${fn})`;
  });

  if (links.length <= 2) {
    return `**Referred Types:** ${links.join(", ")}\n\n`;
  }

  let result = "**Referred Types:**\n\n";
  for (const link of links) {
    result += `- ${link}\n`;
  }
  result += "\n";
  return result;
}

// Returns the doc link path for a type, preferring OutputLocation over Scope.
function typeDocLink(t: TypeDef): string {
  const scopePath = t.OutputLocation
    ? sanitizeOutputLocation(t.OutputLocation)
    : getScopePath(t.Scope.toString());
  const filename = classDocsFileName(t.Name, GeneratedDocSource.README);
  return `../../${scopePath}/${filename}`;
}

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
  const importPath = getModelPackage(
    true,
    typeDef.Scope.toString(),
    typeDef.OutputLocation,
    className,
  );

  let code = `import ${importPath};\n\n${className} value = ${className}.${enumNames[0]};`;
  if (isOpen) {
    const exampleValue = isString ? '"custom_value"' : "999";
    code += `\n\n// Open enum: use .of() to create instances from custom ${
      isString ? "string" : "integer"
    } values`;
    code += `\n${className} custom = ${className}.of(${exampleValue});`;
  }

  return code;
}

registerTemplateFunc("templateModelSnippet", templateModelSnippet);

function accessorValueVariableName(accessorMethodName: string): string {
  const base = accessorMethodName.startsWith("as")
    ? accessorMethodName.slice(2)
    : accessorMethodName;
  return `${caser().ToCamel(base)}Value`;
}

function buildDiscriminatedUnionSwitchLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const discriminatorField = sanitizeFieldName(
    type.Discriminator.TypePropertyName,
  );
  const lines: string[] = [`switch (${unionVar}.${discriminatorField}()) {`];

  for (const mapping of type.Discriminator.Mapping) {
    lines.push(`    case ${JSON.stringify(mapping.Name)}:`);
    lines.push(`        // Handle ${mapping.Name} discriminator variant`);
    lines.push("        break;");
  }

  lines.push("    default:");
  lines.push("        // Handle unknown discriminator variant");
  lines.push("}");
  return lines;
}

function buildDiscriminatedUnionInstanceofPatternLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const lines: string[] = [];

  for (let i = 0; i < type.Discriminator.Mapping.length; i++) {
    const mapping = type.Discriminator.Mapping[i];
    const typeName = sanitizeClassName(mapping.Type.Name);
    const variableName = caser().ToCamel(typeName);
    lines.push(
      `${
        i === 0 ? "if" : "} else if"
      } (${unionVar} instanceof ${typeName} ${variableName}) {`,
    );
    lines.push(`    // Handle ${typeName} variant`);
  }

  lines.push("} else {");
  lines.push("    // Handle unknown discriminator variant");
  lines.push("}");
  return lines;
}

function buildDiscriminatedUnionTypeSwitchLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const lines: string[] = [`switch (${unionVar}) {`];

  for (const mapping of type.Discriminator.Mapping) {
    const typeName = sanitizeClassName(mapping.Type.Name);
    const variableName = caser().ToCamel(typeName);
    lines.push(`    case ${typeName} ${variableName} -> {`);
    lines.push(`        // Handle ${typeName} variant`);
    lines.push("    }");
  }

  lines.push("    default -> {");
  lines.push("        // Handle unknown discriminator variant");
  lines.push("    }");
  lines.push("}");
  return lines;
}

function buildNonDiscriminatedUnionAccessorLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const members = getUnionMemberMetadata(type);
  const lines: string[] = [];

  for (let i = 0; i < members.length; i++) {
    const member = members[i];
    const accessorCall = `${unionVar}.${member.accessorMethodName}()`;
    const typeName = javaImport(member.nonPrimitiveType);
    const variableName = accessorValueVariableName(member.accessorMethodName);
    lines.push(
      `${i === 0 ? "if" : "} else if"} (${accessorCall}.isPresent()) {`,
    );
    lines.push(`    ${typeName} ${variableName} = ${accessorCall}.get();`);
    lines.push(`    // Handle ${member.accessorMethodName} variant`);
  }

  if (isForwardCompatibleUnions()) {
    const jsonNodeType = javaImport("com.fasterxml.jackson.databind.JsonNode");
    lines.push(`} else if (${unionVar}.asJson().isPresent()) {`);
    lines.push(`    ${jsonNodeType} raw = ${unionVar}.asJson().get();`);
    lines.push("    // Handle unknown variant fallback");
  }

  lines.push("}");
  return lines;
}

function buildNonDiscriminatedUnionValueLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const members = getUnionMemberMetadata(type);
  const lines: string[] = [
    `${javaImport("java.lang.Object")} raw = ${unionVar}.value();`,
  ];

  for (let i = 0; i < members.length; i++) {
    const member = members[i];
    const typeName = javaImport(member.nonPrimitiveType);
    const rawType = javaImport(member.rawType);
    const variableName = accessorValueVariableName(member.accessorMethodName);
    lines.push(`${i === 0 ? "if" : "} else if"} (raw instanceof ${rawType}) {`);
    lines.push(`    ${typeName} ${variableName} = (${typeName}) raw;`);
    lines.push(`    // Handle ${member.accessorMethodName} variant`);
  }

  lines.push("} else {");
  lines.push("    // Unknown or unsupported variant");
  lines.push("}");
  return lines;
}

function buildNonDiscriminatedUnionPatternLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const members = getUnionMemberMetadata(type);
  const lines: string[] = [
    `${javaImport("java.lang.Object")} raw = ${unionVar}.value();`,
  ];

  for (let i = 0; i < members.length; i++) {
    const member = members[i];
    const rawType = javaImport(member.rawType);
    const variableName = accessorValueVariableName(member.accessorMethodName);
    lines.push(
      `${
        i === 0 ? "if" : "} else if"
      } (raw instanceof ${rawType} ${variableName}) {`,
    );
    lines.push(`    // Handle ${member.accessorMethodName} variant`);
  }

  lines.push("} else {");
  lines.push("    // Unknown or unsupported variant");
  lines.push("}");
  return lines;
}

function buildNonDiscriminatedUnionTypeSwitchLines(
  type: TypeDef,
  unionVar: string,
): string[] {
  const members = getUnionMemberMetadata(type);
  const lines: string[] = [
    `${javaImport("java.lang.Object")} raw = ${unionVar}.value();`,
    "switch (raw) {",
  ];

  for (const member of members) {
    const rawType = javaImport(member.rawType);
    const variableName = accessorValueVariableName(member.accessorMethodName);
    lines.push(`    case ${rawType} ${variableName} -> {`);
    lines.push(`        // Handle ${member.accessorMethodName} variant`);
    lines.push("    }");
  }

  lines.push("    default -> {");
  lines.push("        // Unknown or unsupported variant");
  lines.push("    }");
  lines.push("}");
  return lines;
}

function fencedJava(lines: string[]): string {
  return ["```java", ...lines, "```", ""].join("\n");
}

function templateUnionConsumptionDocs(type: TypeDef): string {
  const sections: string[] = ["## Consumption Patterns", ""];

  if (type.Discriminator) {
    sections.push("### Java 11+ (Discriminator Switch)", "");
    sections.push(
      fencedJava(buildDiscriminatedUnionSwitchLines(type, "value")),
    );
    sections.push("### Java 16+ (Instanceof Pattern Matching)", "");
    sections.push(
      fencedJava(buildDiscriminatedUnionInstanceofPatternLines(type, "value")),
    );
    sections.push("### Java 21+ (Type Pattern Switch)", "");
    sections.push(
      fencedJava(buildDiscriminatedUnionTypeSwitchLines(type, "value")),
    );
    return sections.join("\n");
  }

  if (context.Global.Config.GenerateOptionalUnionAccessors) {
    sections.push("### Java 11+ (Accessor Methods)", "");
    sections.push(
      fencedJava(buildNonDiscriminatedUnionAccessorLines(type, "value")),
    );
    return sections.join("\n");
  }

  sections.push("### Java 11+ (Raw Value + Cast)", "");
  sections.push(
    fencedJava(buildNonDiscriminatedUnionValueLines(type, "value")),
  );
  sections.push("### Java 16+ (Instanceof Pattern Matching)", "");
  sections.push(
    fencedJava(buildNonDiscriminatedUnionPatternLines(type, "value")),
  );
  sections.push("### Java 21+ (Type Pattern Switch)", "");
  sections.push(
    fencedJava(buildNonDiscriminatedUnionTypeSwitchLines(type, "value")),
  );

  return sections.join("\n");
}

function getResponseContentType(
  operation: Operation,
  serializationMethod: string = "",
): TypeDef | null {
  let preferred: TypeDef | null = null;
  let backup: TypeDef | null = null;

  for (const response of operation.Response.Responses) {
    if (response.Error) {
      continue;
    }
    for (const content of response.Content) {
      if (
        serializationMethod &&
        content.SerializationMethod !== serializationMethod
      ) {
        continue;
      }
      if (!backup) {
        backup = content.Content.Type;
      }
      if (!preferred && content.UsageExample) {
        preferred = content.Content.Type;
      }
    }
  }

  return preferred ?? backup;
}

function templateOperationResponseUnionHandling(
  operation: Operation,
  unionExpr: string,
): string {
  const contentType = getResponseContentType(operation, "");
  if (!contentType || contentType.Type.toString() !== "union") {
    return `System.out.println(${unionExpr});`;
  }

  const unionType = javaImportTypeMandatory(contentType);
  const unionVar = "unionValue";
  const lines: string[] = [`${unionType} ${unionVar} = ${unionExpr};`];

  if (contentType.Discriminator) {
    lines.push(...buildDiscriminatedUnionSwitchLines(contentType, unionVar));
  } else if (context.Global.Config.GenerateOptionalUnionAccessors) {
    lines.push(
      ...buildNonDiscriminatedUnionAccessorLines(contentType, unionVar),
    );
  } else {
    lines.push(...buildNonDiscriminatedUnionValueLines(contentType, unionVar));
  }

  return lines.join("\n");
}

registerTemplateFunc(
  "templateOperationResponseUnionHandling",
  templateOperationResponseUnionHandling,
);

function isOperationResponseUnion(operation: Operation): boolean {
  const contentType = getResponseContentType(operation, "");
  return contentType?.Type.toString() === "union";
}

registerTemplateFunc("isOperationResponseUnion", isOperationResponseUnion);

function hasUnionTypesInSDK(): boolean {
  for (const [, modelsByScope] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, models] of sequencedMapEntries(modelsByScope)) {
      if (models.some((m) => m.Type.toString() === "union")) {
        return true;
      }
    }
  }
  return false;
}

registerTemplateFunc("hasUnionTypesInSDK", hasUnionTypesInSDK);

// @ts-ignore
function templateUnionDocs(type: TypeDef): string {
  if (!context.Global.Config.GenerateUnionDocs) {
    return "";
  }

  let unionDocs = "";
  const className = sanitizeClassName(type.Name);

  if (type.Discriminator) {
    // Quick-reference discriminator table
    const discFieldName = sanitizeFieldName(
      type.Discriminator.TypePropertyName,
    );
    unionDocs += `### Discriminator: \`${discFieldName}\`\n\n`;
    unionDocs += "| Value | Type |\n";
    unionDocs += "| ----- | ---- |\n";

    for (const mapping of type.Discriminator.Mapping) {
      const typeName = sanitizeClassName(mapping.Type.Name);
      const link = typeDocLink(mapping.Type);
      unionDocs += `| \`"${mapping.Name}"\` | [${typeName}](${link}) |\n`;
    }
    unionDocs += "\n";

    // Per-type sections with hydrated examples
    for (const mapping of type.Discriminator.Mapping) {
      const typeName = sanitizeClassName(mapping.Type.Name);
      const link = typeDocLink(mapping.Type);

      unionDocs += `### [\`${typeName}\`](${link})\n\n`;
      unionDocs += `Discriminator value: \`"${mapping.Name}"\`\n\n`;

      seedFaker(typeName);
      const exampleValue = templateModelUsage(
        typeDefToFieldDef(mapping.Type),
        0,
      );
      const rawCode = `${className} value = ${exampleValue};`;
      const { simplified, referencedTypes } = simplifyDocExample(rawCode);

      unionDocs += "```java\n";
      unionDocs += simplified + "\n";
      unionDocs += "```\n\n";

      unionDocs += renderReferredTypes(referencedTypes, typeName);
    }
  } else {
    const methodNames = associatedTypeFactoryMethodNames(type);

    for (const subType of type.AssociatedTypes) {
      const typeName = unqualifyType(sanitizeTypeMandatory(subType));
      const typeStr = subType.Type.toString();

      // Link to model docs for class/enum/union types
      if (typeStr === "class" || typeStr === "enum" || typeStr === "union") {
        const link = typeDocLink(subType);
        unionDocs += `### [\`${typeName}\`](${link})\n\n`;
      } else {
        unionDocs += `### \`${typeName}\`\n\n`;
      }

      // Generate hydrated example value
      seedFaker(typeName);
      const exampleValue = templateModelUsage(typeDefToFieldDef(subType), 0);
      const methodName =
        methodNames.get(sanitizeTypeMandatory(subType)) || "of";

      const rawCode = `${className} value = ${className}.${methodName}(${exampleValue});`;
      const { simplified, referencedTypes } = simplifyDocExample(rawCode);

      unionDocs += "```java\n";
      unionDocs += simplified + "\n";
      unionDocs += "```\n\n";

      unionDocs += renderReferredTypes(referencedTypes, typeName);
    }
  }

  unionDocs += templateUnionConsumptionDocs(type);

  return unionDocs;
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);

// Override common templateObjectFieldDocs for Java
unregisterTemplateFunc("templateObjectFieldDocs");

// @ts-ignore
function templateObjectFieldDocs(type: TypeDef): string {
  const docsRoot = relativePath(
    getModelsLocation(type.OutputLocation) || "/",
    "/",
  );

  const useSetterGetterDocs =
    context.Global.Config.NullFriendlyParameters &&
    context.Global.Config.ShowSetterGetterTypesInDocs;

  if (!useSetterGetterDocs) {
    // Use the standard single-Type column format (same as common)
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

      contents.push([
        `${deprecationDelimiter}\`${fieldName}\`${deprecationDelimiter}`,
        templateTypeMarkdown({
          fieldDef: field,
          streamable: isMultipartFileField(field),
          docsRoot,
          source: GeneratedDocSource.README,
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

  // NullFriendlyParameters mode: show Setter Type and Getter Type columns
  let contents: string[][] = [
    [
      "Field",
      "Setter Type",
      "Getter Type",
      "Required",
      "Description",
      "Example",
    ],
  ];

  let validExamplesAdded = false;

  for (const field of type.Fields) {
    const comments = templateDocumentationComments(field.Comments);
    let description = comments || "N/A";
    const deprecationDelimiter = field.Comments?.Deprecated ? "~~" : "";
    const fieldName = sanitizeField(field, type);

    if (field.Nullable && field.Optional) {
      const triStateNote =
        "Supports tri-state: omit setter to leave unset, pass null to send explicit JSON null.";
      if (description === "N/A") {
        description = triStateNote;
      } else {
        description = `${description}<br/>${triStateNote}`;
      }
    }

    let example = "";
    if (field.Type.Examples.length > 0) {
      example = formatExamples(field.Type.Examples);
      validExamplesAdded = true;
    }

    contents.push([
      `${deprecationDelimiter}\`${fieldName}\`${deprecationDelimiter}`,
      templateSetterTypeDoc(field, docsRoot, GeneratedDocSource.README),
      templateGetterTypeDoc(field, type, docsRoot, GeneratedDocSource.README),
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
