// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "builder method",
    ServerByName: ".server(AvailableServers server)",
    ServerByIndex: ".serverIndex(int serverIdx)",
    ServerByUrl: ".serverURL(String serverUrl)",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  const fieldName = sanitizeFieldName(v.Name);
  const builderName = templateSDKBuilderName(v.Name);
  const type = unqualifyType(sanitizeType(v.Type, false, false));

  return `${builderName}(${type} ${fieldName})`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);

// @ts-ignore
function getUsageServerVariableValue(v: ServerVariable): string {
  if (v.Type.Enum) {
    const names = getEnumNamesFromValues(v.Type.Enum.Values);
    const name = names[names.length - 1];
    // server enums have an unusual location in SDK.Builder
    // TODO this should really be in SDK but is breaking change
    // so will consider first

    // default location is in base package so we will insert SDK.Builder
    const className =
      templatePackageName() + ".SDK.Builder." + sanitizeClassName(v.Type.Name);
    return `${javaImport(className)}.${name}`;
  }

  if (isNumeric(v.Default)) {
    return `"${fakeInteger(v.Name.toLowerCase()).toString()}"`;
  }

  return `"${fakeString(v.Name.toLowerCase())}"`;
}
