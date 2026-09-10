// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "parameter",
    ServerByName: "server: keyof typeof ServerList",
    ServerByIndex: `${getConstants().serverIdx}: number`,
    ServerByUrl: `${getConstants().serverURL}: string`,
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = sanitizeFieldName(v.Name);
  const type = sanitizeType(v.Type, false, "sdk");

  return `${fieldName}: ${type}`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
