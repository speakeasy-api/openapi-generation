function getMultipartFormAnnotation(
  field: FieldDef,
): MultipartFormAnnotation | null {
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() === "multipartForm")
      return annotation as MultipartFormAnnotation;
  }
  return null;
}

function hasMultipartFields(typeDef: TypeDef | undefined): boolean {
  return Boolean(
    typeDef?.Fields?.some((field) => getMultipartFormAnnotation(field)),
  );
}

function multipartWireName(
  field: FieldDef,
  annotation: MultipartFormAnnotation | null,
): string {
  return escapeString(
    annotation?.Name || getJsonAnnotationName(field) || field.Name,
  );
}

function resolveFileShape(typeDef: TypeDef): {
  fileNameProp: string;
  contentProp: string;
} {
  let fileNameProp = "";
  let contentProp = "";

  for (const field of typeDef.Fields ?? []) {
    const annotation = getMultipartFormAnnotation(field);
    if (!annotation) continue;
    if (annotation.Content) {
      contentProp = sanitizeField(field, typeDef);
    } else {
      fileNameProp = sanitizeField(field, typeDef);
    }
  }

  if (!fileNameProp || !contentProp) {
    throw new Error(
      `Multipart file type '${typeDef.Name}' must expose a content (byte[]) property and a filename property`,
    );
  }

  return { fileNameProp, contentProp };
}

function multipartArrayFieldName(name: string): string {
  const fmt = (context.Global.Config as any)?.MultipartArrayFormat;
  return fmt === "legacy" ? `${name}[]` : name;
}

function emitMultipartField(
  formVar: string,
  sourceExpr: string,
  field: FieldDef,
  parentType: TypeDef,
): string[] {
  const annotation = getMultipartFormAnnotation(field);
  if (!annotation) return [];

  const utilsNs = getScopeNamespace("utils", true);
  const ms = `global::${utilsNs}.MultipartSerializer`;
  const prop = wrappedFieldAccess(
    `${sourceExpr}.${sanitizeField(field, parentType)}`,
    field,
  );
  const name = multipartWireName(field, annotation);

  if (annotation.File) {
    if (field.Type.Type.toString() === "array") {
      const itemType = field.Type.ItemType;
      const { fileNameProp, contentProp } = resolveFileShape(itemType);
      const arrayName = multipartArrayFieldName(name);
      return [
        `if (${prop} != null)
{
    foreach (var __mpFile in ${prop})
    {
        if (__mpFile == null) continue;
        ${ms}.AddFile(${formVar}, "${arrayName}", __mpFile.${fileNameProp}, __mpFile.${contentProp});
    }
}`,
      ];
    }
    const { fileNameProp, contentProp } = resolveFileShape(field.Type);
    return [
      `if (${prop} != null)
{
    ${ms}.AddFile(${formVar}, "${name}", ${prop}.${fileNameProp}, ${prop}.${contentProp});
}`,
    ];
  }

  if (annotation.JSON) {
    return [
      `${ms}.AddJson(${formVar}, "${name}", ${prop}, ${jsonTypeInfoRef(
        field.Type,
      )});`,
    ];
  }

  switch (field.Type.Type.toString()) {
    case "array":
      return [`${ms}.AddListScalar(${formVar}, "${name}", ${prop});`];
    default:
      return [
        `${ms}.AddScalar(${formVar}, "${name}", ${scalarValueExpr(
          prop,
          field,
        )});`,
      ];
  }
}

function emitMultipartFields(
  formVar: string,
  sourceExpr: string,
  fields: FieldDef[],
  parentType: TypeDef,
): string[] {
  const lines: string[] = [];
  for (const field of fields) {
    if (!getMultipartFormAnnotation(field)) continue;
    lines.push(...emitMultipartField(formVar, sourceExpr, field, parentType));
  }
  return lines;
}

// @ts-ignore
function templateApplyMultipartBody(op: Operation): string {
  if (!op.Request?.RequestBody) return "";

  const request = op.Request;
  const bodyField = request.RequestBody;
  const formVar = "__mp";
  const requestType = request.Field?.Type;
  const utilsNs = getScopeNamespace("utils", true);
  const ms = `global::${utilsNs}.MultipartSerializer`;

  let sourceExpr = "request";
  let sourceType = bodyField.Type;

  if (requestType && hasMultipartFields(requestType)) {
    sourceType = requestType;
  } else if (request.Field && request.Field.Name !== bodyField.Name) {
    sourceExpr = `request.${sanitizeField(bodyField, requestType)}`;
  }

  const lines: string[] = [`if (request == null)`, `{`];

  if (!bodyField.Nullable && !bodyField.Optional) {
    lines.push(
      `    throw new System.ArgumentNullException(nameof(request), "request body is required");`,
    );
  } else {
    lines.push(`    httpRequest.Content = null;`);
  }

  lines.push(`}`, `else`, `{`, `    var ${formVar} = ${ms}.Create();`);

  const bodyLines =
    sourceType && hasMultipartFields(sourceType)
      ? emitMultipartFields(
          formVar,
          sourceExpr,
          sourceType.Fields ?? [],
          sourceType,
        )
      : [];

  lines.push(...bodyLines.map((line) => `    ${line}`));
  lines.push(`    httpRequest.Content = ${formVar};`, `}`);

  return lines.join("\n            ");
}

registerTemplateFunc("templateApplyMultipartBody", templateApplyMultipartBody);
