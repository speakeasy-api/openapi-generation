// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "builder method",
    ServerByName: "setServer(string $serverName)",
    ServerByIndex: "setServerIndex(int $serverIdx)",
    ServerByUrl: "setServerUrl(string $serverUrl)",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = templateSDKInitFieldName(v.Name, "server");
  const builderName = templateSDKBuilderName(v.Name);
  const type = sanitizeType(v.Type, false, false, "sdk", Qualification.USAGE);

  return `${builderName}(${type} ${fieldName})`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
