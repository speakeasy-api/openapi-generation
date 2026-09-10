function getEventEncoding(typeDef: TypeDef): string {
  // When called from model.go.stmpl, typeDef is the envelope type (e.g. textEvent).
  // Find the "data" field and check its type.
  // When called from union.go.stmpl, typeDef is the variant type (the data type itself).
  for (const field of typeDef.Fields) {
    if (field.Name === "data") {
      return sseDataTypeEncoding(field.Type);
    }
  }

  // No "data" field found — this is a variant type from union discriminator mapping.
  return sseDataTypeEncoding(typeDef);
}

registerTemplateFunc("getEventEncoding", getEventEncoding);
