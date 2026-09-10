type Param = {
  name: string;
  accessor: string;
  fieldDef: FieldDef;
  allowEmptyValue: boolean;
};

type EncodingMethod =
  | "encodeJSON"
  | "encodeMatrix"
  | "encodeLabel"
  | "encodeForm"
  | "encodeSimple"
  | "encodeSpaceDelimited"
  | "encodePipeDelimited"
  | "encodeDeepObject";
type Explode = boolean;
type CharEncoding = "percent" | "percentExceptReserved" | "none";
type EncodingGroup = `${EncodingMethod}:${Explode}:${CharEncoding}`;

// We only support these encodings for query parameters as per OpenAPI 3.x
const supportedQueryMethods = {
  encodeJSON: true,
  encodeForm: true,
  encodeSpaceDelimited: true,
  encodePipeDelimited: true,
  encodeDeepObject: true,
} satisfies Partial<Record<EncodingMethod, true>>;

function generateParamsSpec(options: {
  op: Operation;
  localKey: string;
  globalKey: string;
}) {
  const { op, localKey, globalKey } = options;
  const locals = op.Request?.Field.Type.Fields ?? [];
  const globals = op.Globals?.Fields ?? [];

  const groups: Record<
    "path" | "query" | "header" | "cookie",
    Record<string, Param>
  > = {
    path: {},
    query: {},
    header: {},
    cookie: {},
  };

  const allowEmptyValueSet = new Set(getAllowEmptyValueQueryParamNames(op));
  const findAllowEmptyValue = (fieldDef: FieldDef): boolean => {
    const ann = fieldDef.Annotations?.Get("param");
    if (!ann || !isParamAnnotation(ann)) {
      return false;
    }
    return allowEmptyValueSet.has(ann.Name);
  };

  const getKey = (pt: string) => {
    if (pt === "pathParam") {
      return "path";
    } else if (pt === "queryParam") {
      return "query";
    } else if (pt === "header") {
      return "header";
    } else if (pt === "cookie") {
      return "cookie";
    } else {
      return null;
    }
  };

  locals.forEach((fieldDef) => {
    if (fieldDef.Annotations?.Has("security")) {
      return;
    }
    const ann = fieldDef.Annotations?.Get("param");
    if (!ann || !isParamAnnotation(ann)) {
      return;
    }

    const outboundName = isParameterField(fieldDef)
      ? fieldDef.Name
      : originalFieldName(fieldDef);
    const accessor = sanitizeAccessor(
      localKey,
      outboundName,
      isRequestOptional(op.Request),
    );

    const target = getKey(ann.ParamType);
    if (target == null) {
      return;
    }

    const allowEmptyValue =
      target === "query" ? findAllowEmptyValue(fieldDef) : false;
    groups[target][outboundName] = {
      name: ann.Name,
      accessor,
      fieldDef,
      allowEmptyValue,
    };
  });

  globals.forEach((fieldDef) => {
    if (fieldDef.Annotations?.Has("security")) {
      return;
    }
    const ann = fieldDef.Annotations?.Get("param");
    if (!ann || !isParamAnnotation(ann)) {
      return;
    }

    if (!("ParamType" in ann) || typeof ann.ParamType !== "string") {
      return;
    }

    const accessor = sanitizeAccessor(
      globalKey,
      templateGlobalFieldName(fieldDef.Name),
    );

    const target = getKey(ann.ParamType);
    if (target == null) {
      return;
    }

    const name = fieldDef.Name;
    const existing = groups[target][name];
    if (!existing) {
      const allowEmptyValue =
        target === "query" ? findAllowEmptyValue(fieldDef) : false;
      groups[target][name] = {
        name: ann.Name,
        accessor,
        fieldDef,
        allowEmptyValue,
      };
      return;
    }

    groups[target][name].accessor = `${existing.accessor} ?? ${accessor}`;
  });

  return groups;
}
registerTemplateFunc("generateParamsSpec", generateParamsSpec);

function genSingleParamEncoder(
  field: FieldDef,
  valueStr: string,
  usageLocation: string = "funcs",
) {
  const ann = field.Annotations.Get("param");
  if (ann == null || !isParamAnnotation(ann)) {
    throw new Error(`Field "${field.Name}" does not have a "param" annotation`);
  }

  const fname = ann.Name;
  const type = ann.ParamType;

  if (type === "queryParam") {
    throw new Error("internal error: query params should use bulk encoders");
  }

  const explode = ann.Explode;
  let charEncoding: CharEncoding = "none";
  if (type === "pathParam") {
    charEncoding = ann.AllowReserved ? "percentExceptReserved" : "percent";
  }

  let style = ann.Serialization || ann.Style;
  if (!style) {
    if (type === "header") {
      style = "simple";
    } else if (type === "pathParam") {
      style = "simple";
    }
  }

  switch (style) {
    case "json":
      addInternalImport("encodings", "encodeJSON", usageLocation, typeImport);
      const jsonExplode = type !== "queryParam";
      return `encodeJSON("${fname}", ${valueStr}, { explode: ${jsonExplode}, charEncoding: "${charEncoding}" })`;
    case "matrix":
      addInternalImport("encodings", "encodeMatrix", usageLocation, typeImport);
      return `encodeMatrix("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "label":
      addInternalImport("encodings", "encodeLabel", usageLocation, typeImport);
      return `encodeLabel("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "form":
      addInternalImport("encodings", "encodeForm", usageLocation, typeImport);
      return `encodeForm("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "simple":
      addInternalImport("encodings", "encodeSimple", usageLocation, typeImport);
      return `encodeSimple("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "spaceDelimited":
      addInternalImport(
        "encodings",
        "encodeSpaceDelimited",
        usageLocation,
        typeImport,
      );
      return `encodeSpaceDelimited("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "pipeDelimited":
      addInternalImport(
        "encodings",
        "encodePipeDelimited",
        usageLocation,
        typeImport,
      );
      return `encodePipeDelimited("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    case "deepObject":
      addInternalImport(
        "encodings",
        "encodeDeepObject",
        usageLocation,
        typeImport,
      );
      return `encodeDeepObject("${fname}", ${valueStr}, { charEncoding: "${charEncoding}" })`;
    default:
      throw new Error(
        `${field.OriginalName}: unrecognized parameter encoding style: ${style}`,
      );
  }
}
registerTemplateFunc("genSingleParamEncoder", genSingleParamEncoder);

function genFormBodyEncoder(
  field: FieldDef,
  valueStr: string,
  usageLocation: string = "funcs",
) {
  const ann = field.Annotations.Get("form");
  if (ann == null || !isFormAnnotation(ann)) {
    throw new Error(`Field "${field.Name}" does not have a "form" annotation`);
  }

  const fname = originalFieldName(field);
  const style = ann.JSON ? "json" : "form";
  const explode = ann.Explode;
  const charEncoding = "percent";

  switch (style) {
    case "json":
      addInternalImport("encodings", "encodeJSON", usageLocation, typeImport);
      return `encodeJSON("${fname}", ${valueStr}, { explode: false, charEncoding: "${charEncoding}" })`;
    case "form":
      addInternalImport(
        "encodings",
        "encodeBodyForm",
        usageLocation,
        typeImport,
      );
      return `encodeBodyForm("${fname}", ${valueStr}, { explode: ${explode}, charEncoding: "${charEncoding}" })`;
    default:
      throw new Error(`Unrecognized request encoding style: ${style}`);
  }
}
registerTemplateFunc("genFormBodyEncoder", genFormBodyEncoder);

function generateFileProcessingCode(
  formAccessor: string,
  fieldName: string,
  fileVar: string,
) {
  return [
    `if (isBlobLike(${fileVar})) {`,
    `  const file = ${fileVar};`,
    `  const blob = await normalizeBlob(file);`,
    `  const name = "name" in file ? (file.name as string) : undefined;`,
    `  appendForm(${formAccessor}, "${fieldName}", blob, name)`,
    `} else if (isReadableStream(${fileVar}.content)) {`,
    `  const buffer = await readableStreamToArrayBuffer(${fileVar}.content)`,
    `  const contentType = getContentTypeFromFileName(${fileVar}.fileName) || 'application/octet-stream';`,
    `  appendForm(${formAccessor}, "${fieldName}", bytesToBlob(buffer, contentType), ${fileVar}.fileName)`,
    `} else {`,
    `  const contentType = getContentTypeFromFileName(${fileVar}.fileName) || 'application/octet-stream';`,
    `  appendForm(${formAccessor}, "${fieldName}", bytesToBlob(${fileVar}.content, contentType), ${fileVar}.fileName)`,
    `}`,
  ].join("\n");
}

function genMultiPartEncoder(
  field: FieldDef,
  formAccessor: string,
  valueStr: string,
  usageLocation: string = "funcs",
) {
  const ann = field.Annotations.Get("multipartForm");
  if (ann == null || !isMultipartFormAnnotation(ann)) {
    throw new Error(
      `Field "${field.Name}" does not have a "multipart" annotation`,
    );
  }

  const fname = originalFieldName(field);
  switch (true) {
    case ann.JSON:
      if (ann.FieldType.Type.toString() === "union") {
        const { isItPrimitive, isItObject } = findTypeDetails(ann.FieldType);
        // If the type is an object then we add it as json string to the multipart/form-data
        if (isItObject && isItPrimitive) {
          addInternalImport(
            "encodings",
            "appendForm",
            usageLocation,
            typeImport,
          );
          addInternalImport(
            "encodings",
            "encodeJSON",
            usageLocation,
            typeImport,
          );
          const jsonified = `encodeJSON("${fname}", ${valueStr}, { explode: true })`;
          return `
            if (typeof ${valueStr} === "object"){
            appendForm(${formAccessor}, "${fname}", ${jsonified})
            }
            else {
            appendForm(${formAccessor}, "${fname}", ${valueStr})
            }
            `;
        } else if (isItObject && !isItPrimitive) {
          addInternalImport(
            "encodings",
            "appendForm",
            usageLocation,
            typeImport,
          );
          addInternalImport(
            "encodings",
            "encodeJSON",
            usageLocation,
            typeImport,
          );
          const jsonified = `encodeJSON("${fname}", ${valueStr}, { explode: true })`;
          return `
            appendForm(${formAccessor}, "${fname}", ${jsonified})
            `;
        } else {
          addInternalImport(
            "encodings",
            "appendForm",
            usageLocation,
            typeImport,
          );
          return `
            appendForm(${formAccessor}, "${fname}", ${valueStr})
            `;
        }
      } else {
        addInternalImport("encodings", "appendForm", usageLocation, typeImport);
        addInternalImport("encodings", "encodeJSON", usageLocation, typeImport);
        const jsonified = `encodeJSON("${fname}", ${valueStr}, { explode: true })`;
        return `appendForm(${formAccessor}, "${fname}", ${jsonified})`;
      }
    case ann.File:
      addInternalImport("encodings", "appendForm", usageLocation, typeImport);
      addInternalImport(
        "encodings",
        "normalizeBlob",
        usageLocation,
        typeImport,
      );
      addTypeImport("blobs", "isBlobLike", usageLocation);
      addTypeImport("streams", "isReadableStream", usageLocation);
      addInternalImport("files", "readableStreamToArrayBuffer", usageLocation);
      addInternalImport("files", "getContentTypeFromFileName", usageLocation);
      addInternalImport("files", "bytesToBlob", usageLocation);

      // Check if this is an array of files at template time
      if (field.Type.Type.toString() === "array") {
        // Use legacy format (with []) if multipartArrayFormat is 'legacy', otherwise use standard format
        const arrayFieldName =
          context.Global.Config.MultipartArrayFormat === "legacy"
            ? `${fname}[]`
            : fname;
        // valueStr may contain optional chaining (e.g. "obj?.files"), but we need
        // a non-optional base for the `for...of` loop. Strip `?.` → `.` and fall
        // back to an empty array so the loop is simply skipped when the value is
        // nullish.  Example: "obj?.files" → "obj.files ?? []"
        return [
          `for (const fileItem of ${valueStr.replace(/\?\./g, ".")} ?? []) {`,
          `  ${generateFileProcessingCode(
            formAccessor,
            arrayFieldName,
            "fileItem",
          ).replace(/\n/g, "\n  ")}`,
          `}`,
        ].join("\n");
      } else {
        return generateFileProcessingCode(formAccessor, fname, valueStr);
      }
    default:
      addInternalImport("encodings", "appendForm", usageLocation, typeImport);
      return `appendForm(${formAccessor}, "${fname}", ${valueStr})`;
  }
}
registerTemplateFunc("genMultiPartEncoder", genMultiPartEncoder);

function findTypeDetails(type: TypeDef): {
  isItPrimitive: boolean;
  isItObject: boolean;
} {
  if (type.Type.toString() === "union") {
    let isPrimitive = false;
    let isObject = false;
    // Go over all the union types and inner types to find if any of them is primitive or object
    for (const item_val of type.AssociatedTypes) {
      const { isItPrimitive: anyItemPrimitive, isItObject: anyItemObject } =
        findTypeDetails(item_val);
      isPrimitive = isPrimitive || anyItemPrimitive;
      isObject = isObject || anyItemObject;
    }
    return { isItPrimitive: isPrimitive, isItObject: isObject };
  } else if (
    type.Type.toString() === "string" ||
    type.Type.toString() === "integer" ||
    type.Type.toString() === "int32" ||
    type.Type.toString() === "bigint" ||
    type.Type.toString() === "number" ||
    type.Type.toString() === "float32" ||
    type.Type.toString() === "decimal" ||
    type.Type.toString() === "boolean" ||
    type.Type.toString() === "date" ||
    type.Type.toString() === "date-time" ||
    type.Type.toString() === "enum"
  ) {
    return { isItPrimitive: true, isItObject: false };
  } else {
    return { isItPrimitive: false, isItObject: true };
  }
}

function encodingGroupForField(ann: ParamAnnotation): EncodingGroup | null {
  const type = ann.ParamType;

  const explode = ann.Explode;
  let charEncoding: CharEncoding = "none";
  if (type === "queryParam" || type === "pathParam") {
    charEncoding = ann.AllowReserved ? "percentExceptReserved" : "percent";
  }

  let style = ann.Serialization || ann.Style;
  if (!style) {
    if (type === "header") {
      style = "simple";
    } else if (type === "queryParam") {
      style = "form";
    } else if (type === "pathParam") {
      style = "simple";
    }
  }

  switch (style) {
    case "json":
      return `encodeJSON:${explode}:${charEncoding}`;
    case "matrix":
      return `encodeMatrix:${explode}:${charEncoding}`;
    case "label":
      return `encodeLabel:${explode}:${charEncoding}`;
    case "form":
      return `encodeForm:${explode}:${charEncoding}`;
    case "simple":
      return `encodeSimple:${explode}:${charEncoding}`;
    case "spaceDelimited":
      return `encodeSpaceDelimited:${explode}:${charEncoding}`;
    case "pipeDelimited":
      return `encodePipeDelimited:${explode}:${charEncoding}`;
    case "deepObject":
      return `encodeDeepObject:${explode}:${charEncoding}`;
    default:
      return null;
  }
}

function genQueryParamEncoder(
  params: Record<string, Param>,
  usageLocation: string = "funcs",
) {
  const groupings: Partial<Record<EncodingGroup, Omit<Param, "fieldDef">[]>> =
    {};

  for (const [_, { accessor, fieldDef, allowEmptyValue }] of Object.entries(
    params,
  )) {
    const ann = fieldDef.Annotations.Get("param");
    if (ann == null || !isParamAnnotation(ann)) {
      throw new Error(
        `Field "${fieldDef.OriginalName}" does not have a "param" annotation`,
      );
    }

    const style = ann.Serialization || ann.Style;
    const groupKey = encodingGroupForField(ann);
    if (groupKey == null) {
      throw new Error(
        `${fieldDef.OriginalName}: unrecognized parameter encoding style: ${style}`,
      );
    }

    if (!supportedQueryMethods[groupKey.split(":", 1)[0]]) {
      throw new Error(
        `${fieldDef.OriginalName}: encoding style not allowed for queries: ${style}`,
      );
    }

    let group = (groupings[groupKey] ||= []);
    group.push({ name: ann.Name, accessor, allowEmptyValue });
  }

  const code: string[] = [];
  const sortedGroups = Object.entries(groupings).sort(kvComparator);
  for (const [groupKey, fields] of sortedGroups) {
    if (!fields.length) {
      continue;
    }

    const sortedFields = fields.sort((a, b) => a.name.localeCompare(b.name));

    const mapping =
      "{\n" +
      sortedFields
        .map(({ name, accessor }) => `    "${name}": ${accessor}`)
        .join(",\n") +
      "\n}";

    const [m, explode, charEncoding] = groupKey.split(":");
    const method = `${m}Query`;
    addInternalImport("encodings", method, usageLocation, typeImport);

    let opts: string[] = [];
    if (explode === "false") {
      opts.push("explode: false");
    }
    if (charEncoding !== "percent") {
      opts.push(`charEncoding: "${charEncoding}"`);
    }

    const allowEmptyFields = sortedFields
      .filter((p) => p.allowEmptyValue)
      .map((p) => `"${p.name}"`);
    if (allowEmptyFields.length > 0) {
      opts.push(`allowEmptyValue: [${allowEmptyFields.join(", ")}]`);
    }

    const options = opts.length ? `, {${opts.join(", ")}}` : "";

    code.push(`  ${method}(${mapping}${options})`);
  }

  if (!code.length) {
    return "";
  }

  if (code.length === 1) {
    return code[0];
  }

  addInternalImport("encodings", `queryJoin`, usageLocation, typeImport);

  return `queryJoin(${code.join(", ")})`;
}
registerTemplateFunc("genQueryParamEncoder", genQueryParamEncoder);

function kvComparator<T extends [key: string, value: unknown]>(a: T, b: T) {
  return a[0].localeCompare(b[0]);
}
