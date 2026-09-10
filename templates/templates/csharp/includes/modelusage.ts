// @ts-ignore
function templateUnion(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let { type: typeToUse, example: selectedExample } = selectExampleUnionType(
    fieldDef,
    example,
    undefined,
    additionalContext,
  );

  let typeName: string;
  if (fieldDef.Type.Discriminator) {
    const map = fieldDef.Type.Discriminator.Mapping.find(
      (m) => m.Type === typeToUse,
    );
    if (map === undefined) {
      throw new Error(
        `Invalid discriminated union type selected for field "${fieldDef.Name}"`,
      );
    }
    typeName = map.Name;
  } else {
    typeName = sanitizeUnionTypeName(typeToUse);
  }

  const unionClass = sanitizeClass(
    fieldDef.Type,
    "",
    false,
    additionalContext?.isUsage,
  );

  return [
    `${unionClass}.Create${sanitizeClassName(typeName)}(`,
    indentLines(
      [
        templateValue(
          typeDefToFieldDef(typeToUse, fieldDef),
          selectedExample,
          false,
          additionalContext,
        ),
      ],
      1,
    ),
    ")",
  ].join("\n");
}

// @ts-ignore
function getEventStreamField(typeDef: TypeDef): FieldDef | null {
  for (const field of typeDef.Fields) {
    if (field.Type?.Type == "event-stream") {
      return field;
    }
  }

  return null;
}

registerTemplateFunc("getEventStreamField", getEventStreamField);

// @ts-ignore
function getEventStreamSiblingField(operation: Operation): FieldDef | null {
  // Find the SubResponse that contains the event stream
  const subResponse = operation.Response.Responses.find((subResponse) =>
    subResponse.Content.some(
      (content) => content.SerializationMethod === "eventstream",
    ),
  );

  if (!subResponse) return null;

  for (const content of subResponse.Content) {
    if (content.SerializationMethod !== "eventstream" && content.Content) {
      return content.Content;
    }
  }

  return null;
}

registerTemplateFunc("getEventStreamSiblingField", getEventStreamSiblingField);
