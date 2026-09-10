// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "option",
    ServerByName: "WithServer(server string)",
    ServerByIndex: "WithServerIndex(serverIndex int)",
    ServerByUrl: "WithServerURL(serverURL string)",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = sanitizePrivateFieldName(v.Name);
  const type = sanitizeType(v.Type, false, "");

  return `${templateSDKOptionName(v.Name)}(${fieldName} ${type})`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
