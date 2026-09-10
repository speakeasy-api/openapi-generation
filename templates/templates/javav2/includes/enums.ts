function templateEnums(typeDef: TypeDef, indent: number = 1): string {
  const members = javaEnumMembers(typeDef);
  const descriptions = typeDef.Enum.Descriptions || {};

  const lines: string[] = [];

  members.forEach((member, idx) => {
    const value = typeDef.Enum.Values[idx];
    const description = descriptions[`${value}`];
    if (description) {
      lines.push(...formatJavaEnumDescriptionComments(description));
    }

    const suffix = idx < members.length - 1 ? "," : "";
    lines.push(`    ${member.Identifier}(${member.Value})${suffix}`);
  });

  return indentLines(lines, Math.max(indent - 1, 0));
}
registerTemplateFunc("templateEnums", templateEnums);

function formatJavaEnumDescriptionComments(description: string): string[] {
  const sanitizedLines = sanitizeComments(description).split("\n");
  const docLines = ["    /**"];

  sanitizedLines.forEach((line) => {
    const trimmed = line.trim();
    if (trimmed.length === 0) {
      docLines.push("     *");
    } else {
      docLines.push(`     * ${trimmed}`);
    }
  });

  docLines.push("     */");
  return docLines;
}

function javaEnumMembers(typeDef: TypeDef): JavaEnumMember[] {
  const enumNames = getEnumNames(typeDef);
  return typeDef.Enum.Values.map((value, index) => {
    return {
      Identifier: enumNames[index],
      Value: sanitizeEnumValue(value, typeDef.Enum.Type.Type),
    };
  });
}
registerTemplateFunc("javaEnumMembers", javaEnumMembers);

type JavaEnumMember = {
  // upper case enum member identifier
  Identifier: string;
  // java representation of value (like a string including double quotes)
  Value: string;
};

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString() as DataType) {
    case "string":
      return `"${value.replaceAll("\\", "\\\\")}"`;
    case "int32":
      return value;
    case "integer":
      return value + "L";
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

function getEnumNames(t: TypeDef): string[] {
  if (t.Enum?.Names.length > 0) {
    return t.Enum.Names.map((n) => getEnumName(n));
  } else {
    return getEnumNamesFromValues(t.Enum?.Values);
  }
}

function getEnumMemberForValue(typeDef: TypeDef, value: any): string {
  const members = javaEnumMembers(typeDef);
  const member = members.find((m) => {
    // Compare the raw value with the enum value
    const rawValue = typeDef.Enum.Values[members.indexOf(m)];
    return rawValue === value;
  });
  return member ? member.Identifier : "";
}
registerTemplateFunc("getEnumMemberForValue", getEnumMemberForValue);
