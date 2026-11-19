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
    const valueExpr = scalarValueExpr(propAccess, f);
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
    const allowReserved = param.AllowReserved ? ", true" : "";
    const templateReplacement = (d: string) =>
      `urlString = ${ub}.ReplacePathParam(urlString, "${param.Name}", ${d}, "${style}", ${param.Explode}${allowReserved});`;
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
      lines.push(templateReplacement(scalarValueExpr(propAccess, field)));
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
      const allowReserved = param.AllowReserved ? ", true" : "";
      const templateAppend = (d: string) =>
        `${ub}.AppendQueryParam(__qs, "${param.Name}", ${d}, "${style}", ${param.Explode}, ${allowEmpty}${allowReserved});`;
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
        lines.push(templateAppend(scalarValueExpr(propAccess, field)));
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
      lines.push(templateAppend(scalarValueExpr(propAccess, field)));
    }
  }
  return lines.join("\n            ");
}

unregisterTemplateFunc("templateApplyHeaders");
registerTemplateFunc("templateApplyHeaders", templateApplyHeaders);
