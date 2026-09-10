// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "parameter",
    ServerByName: "server (Symbol)",
    ServerByIndex: "server_idx (Integer)",
    ServerByUrl: "server_url (String)",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = templateSDKInitFieldName(v.Name, "global");
  const type = sanitizeType(v.Type, false, false, "sdk");

  return `${fieldName} (${type})`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
