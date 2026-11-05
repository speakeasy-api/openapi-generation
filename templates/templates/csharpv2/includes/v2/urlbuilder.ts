// @ts-ignore
function getParamAnnotation(f: FieldDef): ParamAnnotation | null {
  for (const a of f.Annotations) {
    if (a.Type().toString() === "param") return a as ParamAnnotation;
  }
  return null;
}
registerTemplateFunc("getParamAnnotation", getParamAnnotation);

function collectParamFields(
  op: Operation,
  paramType: "pathParam" | "queryParam" | "header",
): Array<{ field: FieldDef; param: ParamAnnotation }> {
  const out: Array<{ field: FieldDef; param: ParamAnnotation }> = [];
  if (!op.Request) return out;
  for (const f of op.Request.Field.Type.Fields ?? []) {
    const p = getParamAnnotation(f);
    if (p && p.ParamType === paramType) out.push({ field: f, param: p });
  }
  return out;
}

function urlBuilderJsonName(f: FieldDef): string {
  for (const a of f.Annotations) {
    if (a.Type().toString() === "json") {
      return (a as JSONAnnotation).FieldName;
    }
  }
  for (const a of f.Annotations) {
    if (a.Type().toString() === "param") {
      const p = a as ParamAnnotation;
      if (p.Name) return p.Name;
    }
  }
  return f.Name;
}

function urlBuilderObjectFieldsSorted(typeDef: TypeDef): FieldDef[] {
  const fields = [...(typeDef.Fields ?? [])];
  fields.sort((a, b) =>
    urlBuilderJsonName(a).localeCompare(urlBuilderJsonName(b)),
  );
  return fields;
}

function urlBuilderDictValueExpr(propAccess: string, field: FieldDef): string {
  if (field.Type.Type.toString() === "enum") {
    const enumType = field.Type.Enum?.Type?.Type?.toString();
    const supported =
      enumType === "string" ||
      enumType === "int32" ||
      enumType === "int64" ||
      enumType === "integer";
    if (supported) {
      const acc =
        field.Nullable || field.Optional ? `${propAccess}?` : propAccess;
      return field.Type.Enum?.Open ? `${acc}.Value` : `${acc}.Value()`;
    }
  }
  return propAccess;
}

function urlBuilderEmitObjectDict(
  dictVar: string,
  sourceExpr: string,
  objType: TypeDef,
): string[] {
  const lines: string[] = [
    `var ${dictVar} = new System.Collections.Generic.Dictionary<string, object?>();`,
  ];
  for (const f of urlBuilderObjectFieldsSorted(objType)) {
    const wireName = urlBuilderJsonName(f);
    const escaped = escapeString(wireName);
    const propAccess = wrappedFieldAccess(
      `${sourceExpr}.${sanitizeField(f, objType)}`,
      f,
    );
    const valueExpr = urlBuilderDictValueExpr(propAccess, f);
    const assign = `${dictVar}["${escaped}"] = ${valueExpr};`;
    if (canFieldBeNull(f)) {
      lines.push(`if (${propAccess} != null) { ${assign} }`);
    } else {
      lines.push(assign);
    }
  }
  return lines;
}

function urlBuilderIsObjectParam(field: FieldDef): boolean {
  return field.Type.Type.toString() === "class";
}

// A class-based open enum is not a C# enum, so Utilities.ToString cannot stringify it as a
// scalar path/query/header param. Emit its underlying `.Value` (a primitive) at the call
// site. Native (closed) enums are real C# enums and serialize fine as-is.
function urlBuilderScalarParamExpr(
  propAccess: string,
  field: FieldDef,
): string {
  if (field.Type.Type.toString() === "enum" && field.Type.Enum?.Open) {
    const nullSafe = field.Nullable || field.Optional ? "?" : "";
    return `${propAccess}${nullSafe}.Value`;
  }
  return propAccess;
}

// @ts-ignore
function templateBuildURL(op: Operation): string {
  const path = collectParamFields(op, "pathParam");
  const query = collectParamFields(op, "queryParam");
  const utilsNs = getScopeNamespace("utils", true);
  const ub = `global::${utilsNs}.URLBuilder`;
  const reqAccess = isRequestOptional(op.Request) ? "request?." : "request.";

  const opPath = op.Path.split("#", 1)[0];
  const lines: string[] = [`var urlString = baseUrl + "${opPath}";`];

  for (const { field, param } of path) {
    const propAccess = `${reqAccess}${sanitizeField(
      field,
      op.Request.Field.Type,
    )}`;
    const style = param.Style || "simple";
    const templateReplacement = (d: string) =>
      `urlString = ${ub}.ReplacePathParam(urlString, "${param.Name}", ${d}, "${style}", ${param.Explode});`;
    if (param.Serialization === "json") {
      lines.push(
        `urlString = ${ub}.ReplacePathParamJson(urlString, "${
          param.Name
        }", ${propAccess}, ${jsonTypeInfoRef(field.Type)});`,
      );
    } else if (urlBuilderIsObjectParam(field)) {
      const dictVar = "__paramDict";
      const dict = indentString(
        urlBuilderEmitObjectDict(dictVar, propAccess, field.Type).join("\n"),
        1,
      );
      lines.push(
        `if (${propAccess} != null) {
${dict}
    ${templateReplacement(dictVar)}
} else {
    ${templateReplacement("null")}
}`,
      );
    } else {
      lines.push(
        templateReplacement(urlBuilderScalarParamExpr(propAccess, field)),
      );
    }
  }

  if (query.length > 0) {
    const allowEmptyByField = new Map<string, boolean>();
    for (const qp of op.Request?.Params?.QueryParams ?? []) {
      if (qp.Field?.Name)
        allowEmptyByField.set(qp.Field.Name, qp.AllowEmptyValue);
    }
    lines.push(`var __qs = new System.Collections.Generic.List<string>();`);
    for (const { field, param } of query) {
      const propAccess = `${reqAccess}${sanitizeField(
        field,
        op.Request.Field.Type,
      )}`;
      const style = param.Style || "form";
      const allowEmpty = allowEmptyByField.get(field.Name) ? "true" : "false";
      const templateAppend = (d: string) =>
        `${ub}.AppendQueryParam(__qs, "${param.Name}", ${d}, "${style}", ${param.Explode}, ${allowEmpty});`;
      if (param.Serialization === "json") {
        lines.push(
          `${ub}.AppendQueryParamJson(__qs, "${
            param.Name
          }", ${propAccess}, ${jsonTypeInfoRef(field.Type)});`,
        );
      } else if (urlBuilderIsObjectParam(field)) {
        const dictVar = "__paramDict";
        const dict = indentString(
          urlBuilderEmitObjectDict(dictVar, propAccess, field.Type).join("\n"),
          1,
        );
        lines.push(
          `if (${propAccess} != null) {
${dict}
    ${templateAppend(dictVar)}
}`,
        );
      } else {
        lines.push(
          templateAppend(urlBuilderScalarParamExpr(propAccess, field)),
        );
      }
    }
    lines.push(`urlString = ${ub}.Finalize(urlString, __qs);`);
  }

  return lines.join("\n            ");
}

unregisterTemplateFunc("templateBuildURL");
registerTemplateFunc("templateBuildURL", templateBuildURL);

// @ts-ignore
function templateApplyHeaders(op: Operation): string {
  const headers = collectParamFields(op, "header");
  if (headers.length === 0) return "";
  const utilsNs = getScopeNamespace("utils", true);
  const ub = `global::${utilsNs}.URLBuilder`;
  const reqAccess = isRequestOptional(op.Request) ? "request?." : "request.";

  const lines: string[] = [];
  for (const { field, param } of headers) {
    const propAccess = `${reqAccess}${sanitizeField(
      field,
      op.Request.Field.Type,
    )}`;
    const style = param.Style || "simple";
    const templateAppend = (d: string) =>
      `${ub}.AppendHeader(httpRequest, "${param.Name}", ${d}, "${style}", ${param.Explode});`;
    if (urlBuilderIsObjectParam(field)) {
      const dictVar = "__hdrDict";
      const dict = indentString(
        urlBuilderEmitObjectDict(dictVar, propAccess, field.Type).join("\n"),
        1,
      );
      lines.push(
        `if (${propAccess} != null) {
${dict}
    ${templateAppend(dictVar)}
}`,
      );
    } else {
      lines.push(templateAppend(urlBuilderScalarParamExpr(propAccess, field)));
    }
  }
  return lines.join("\n            ");
}

unregisterTemplateFunc("templateApplyHeaders");
registerTemplateFunc("templateApplyHeaders", templateApplyHeaders);
