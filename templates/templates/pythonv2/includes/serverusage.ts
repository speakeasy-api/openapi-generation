// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "parameter",
    ServerByName: "server: str",
    ServerByIndex: "server_idx: int",
    ServerByUrl: "server_url: str",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = templateSDKInitFieldName(v.Name, "global");
  const type = templateSimpleType(v.Type, "sdks");

  return `${fieldName}: ${type}`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
