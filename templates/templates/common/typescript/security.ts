function templateHoistedSecurityIndices(operation: Operation): string {
  if (!operation.HoistedSecurityConfig?.Fields) {
    return "[]";
  }

  const indices = Array.from(
    new Set(operation.HoistedSecurityConfig.Fields.map((hf) => hf.Group)),
  );

  return `[${indices.join(", ")}]`;
}
registerTemplateFunc(
  "templateHoistedSecurityIndices",
  templateHoistedSecurityIndices,
);
