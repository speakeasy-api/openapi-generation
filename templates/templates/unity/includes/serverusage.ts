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
  const type = sanitizeType(v.Type, false, "sdk");

  return `${fieldName}: ${type}`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
