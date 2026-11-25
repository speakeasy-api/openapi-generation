// SDK-side serializers (form/multipart/url) read request fields directly, so a
// wrapped field must be unwrapped to its inner T? via .ValueOrDefault.
// @ts-ignore
function wrappedFieldAccess(base: string, field: FieldDef): string {
  return needsOptionalNullableWrapper(field) ? `${base}.ValueOrDefault` : base;
}

let serverVariableTypeNames: Set<string> | undefined;

// Server variable enums render into SDK.cs, under the SDK namespace.
function isServerVariableEnum(typeDef: TypeDef): boolean {
  if (!serverVariableTypeNames) {
    const names = new Set<string>();
    const visit = (sdk: SDK) => {
      for (const v of sdk.Servers?.GetVariables() ?? []) {
        if (v.Type?.Name) names.add(v.Type.Name);
      }
      (sdk.SubSDKs ?? []).forEach(visit);
    };
    visit(context.Global.AST.MainSDK);
    serverVariableTypeNames = names;
  }
  return serverVariableTypeNames.has(typeDef.Name);
}

function getEnumQualifiedName(typeDef: TypeDef): string {
  const namespace = isServerVariableEnum(typeDef)
    ? getSDKNamespace()
    : getModelNamespace(typeDef.OutputLocation, true);
  return `global::${namespace}.${sanitizeEnumName(typeDef.Name)}`;
}

registerTemplateFunc("getEnumQualifiedName", getEnumQualifiedName);

// Input of enum-extension.cs.stmpl, in the shape of its OAuth2 scopes caller.
function getEnumExtensionData(typeDef: TypeDef): object {
  return {
    Name: sanitizeEnumName(typeDef.Name),
    QualifiedName: getEnumQualifiedName(typeDef),
    Scopes: typeDef.Enum.Values.map((value) => ({ Name: `${value}` })),
    DedupNames: getEnumNames(typeDef),
    Int: typeDef.Enum.Type.Type.toString() !== "string",
  };
}

registerTemplateFunc("getEnumExtensionData", getEnumExtensionData);

function templateStringLiteral(value: string): string {
  return `"${value.replaceAll(`\\`, `\\\\`).replaceAll(`"`, `\\"`)}"`;
}

registerTemplateFunc("templateStringLiteral", templateStringLiteral);

// Parenthesized so the C# formatter keeps a negative constant on one line.
function templateIntPattern(value: string): string {
  return value.startsWith("-") ? `(${value})` : value;
}

registerTemplateFunc("templateIntPattern", templateIntPattern);
