// @ts-ignore
function getEnumFormat(typeDef: TypeDef): "enum" | "union" {
  const raw = typeDef.Enum.Format || "union";

  switch (raw) {
    case "enum":
      return "enum";
    case "union":
      return "union";
    default:
      throw new Error(`Unknown enum format: ${raw}`);
  }
}
unregisterTemplateFunc("getEnumFormat");
registerTemplateFunc("getEnumFormat", getEnumFormat);

function getEnumDataType(typeDef: TypeDef): "string" | "number" {
  const underlying = typeDef.Enum?.Type.Type.toString();
  switch (underlying) {
    case "string":
      return "string";
    case "int32":
    case "integer":
      return "number";
    default:
      throw new Error(`Unsupported enum type: ${underlying}`);
  }
}
registerTemplateFunc("getEnumDataType", getEnumDataType);

// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  const en = typeDef.Enum;
  if (en == null) {
    return "";
  }

  const rawValues = en.Values;
  if (rawValues == null) {
    return "";
  }

  const sep = getEnumFormat(typeDef) === "enum" ? " = " : ": ";

  const enumNames = getEnumNames(typeDef);
  return rawValues
    .map((val, index) => {
      const value = sanitizeEnumValue(val, en.Type.Type);
      const comment = formatTypeScriptEnumDescription(
        en.Descriptions?.[`${val}`],
      );

      const lines: string[] = [];
      if (comment) {
        lines.push(comment);
      }

      lines.push(`${enumNames[index]}${sep}${value},`);

      return lines.join("\n");
    })
    .join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);

// @ts-ignore
function templateEnumLiteralUnion(typeDef: TypeDef): string {
  const en = typeDef.Enum;
  if (en == null || en.Values == null) {
    return "never";
  }
  return en.Values.map((val) => sanitizeEnumValue(val, en.Type.Type)).join(
    " | ",
  );
}

registerTemplateFunc("templateEnumLiteralUnion", templateEnumLiteralUnion);

function formatTypeScriptEnumDescription(
  description?: string,
): string | undefined {
  if (!description) {
    return undefined;
  }

  const sanitized = sanitizeComments(description).split("\n");
  const docLines = ["/**"];

  sanitized.forEach((line) => {
    const trimmed = line.trim();
    if (trimmed.length === 0) {
      docLines.push(" *");
    } else {
      docLines.push(` * ${trimmed}`);
    }
  });

  docLines.push(" */");

  return docLines.join("\n");
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

// @ts-ignore
function getEnumNames(t: TypeDef): string[] {
  if (t.Enum?.Names.length > 0) {
    if (context.Global.Config.FixEnumNameSanitization === true) {
      return t.Enum.Names.map((n) => sanitizeName(n.trim() || "Unknown"));
    }
    return t.Enum.Names.map((n) => getEnumName(n));
  } else {
    return getEnumNamesFromValues(t.Enum?.Values);
  }
}
