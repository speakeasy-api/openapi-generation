// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "parameter",
    ServerByName: "server: string",
    ServerByIndex: "serverIndex: int",
    ServerByUrl: "serverUrl: string",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = sanitizePrivateFieldName(v.Name);

  if (v.Type.Type.toString() == "enum") {
    return `${fieldName}: ${sanitizeClass(v.Type, "", true)}`;
  }

  return `${fieldName}: ${sanitizeType(v.Type)}`;
}
registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
