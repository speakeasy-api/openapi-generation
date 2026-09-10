function getRootPackage() {
  let versionParts = context.Global.Config.SDKVersion.split(".");

  let majorVersion = parseInt(versionParts[0], 10);

  if (majorVersion > 1) {
    return `github.com/${context.Global.Config.Author}/terraform-provider-${context.Global.Config.PackageName}/v${majorVersion}`;
  } else {
    return `github.com/${context.Global.Config.Author}/terraform-provider-${context.Global.Config.PackageName}`;
  }
}

registerTemplateFunc("getRootPackage", getRootPackage);

function getSDKPackage() {
  return `${getRootPackage()}/internal/sdk`;
}

registerTemplateFunc("getSDKPackage", getSDKPackage);

function sanitizeStrictCheckNewType(typeDef, scope) {
  let className = "New" + sanitizeClassName(typeDef.Name);

  let prefix = "";
  if (typeDef.Scope.toString() != scope.toString()) {
    prefix = typeDef.Scope.toString() + ".";
  }

  return prefix + className;
}

registerTemplateFunc("sanitizeStrictCheckNewType", sanitizeStrictCheckNewType);

function getPluralizedVarSymbolName(
  symbolManager: Record<string, boolean>,
  name: string,
  suffix: string = "",
): string {
  for (let i = 0; i < 1000; i++) {
    const option = sanitizeVariableName(
      name + suffix + (i > 0 ? String(i) : ""),
    );
    if (!symbolManager[option]) {
      symbolManager[option] = true;
      return option;
    }
  }
  return name;
}

unregisterTemplateFunc("getGolangPackage");
registerTemplateFunc("getGolangPackage", getSDKPackage);
// @ts-ignore
getGolangPackage = getSDKPackage;

unregisterTemplateFunc("sanitizeSDKPackageName");

function customSanitizeSDKPackageName(): string {
  return sanitizeFileName(context.Global.Config.SDKName);
}

registerTemplateFunc("sanitizeSDKPackageName", customSanitizeSDKPackageName);

function templateJSONFieldName(fieldDef: FieldDef): string {
  for (const annotation of fieldDef.Annotations) {
    if (annotation.Type().toString() == "json") {
      return escapeString((annotation as JSONAnnotation).FieldName);
    }
  }
  return "";
}
registerTemplateFunc("templateJSONFieldName", templateJSONFieldName);

// @ts-ignore

function getFieldType(obj, fieldName: string) {
  const found = obj.Fields.find((field) => field.Name === fieldName);
  if (!found) {
    throw new Error(`Could not find field type: ${fieldName}`);
  }

  return found.Type;
}
registerTemplateFunc("getFieldType", getFieldType);

function sanitizeTFStateName(name) {
  name = sanitizeName(name);

  return caser().ToSNAKE(name).replace(/_$/, "").toLowerCase();
}

registerTemplateFunc("sanitizeTFStateName", sanitizeTFStateName);

function sanitizePackageReference(name) {
  name = sanitizeName(name);

  return caser().ToKebab(name).toLowerCase();
}

registerTemplateFunc("sanitizePackageReference", sanitizePackageReference);

function sanitizeLocalName(name: string): string {
  name = sanitizeName(name);

  return name.toLowerCase();
}

registerTemplateFunc("sanitizeLocalName", sanitizeLocalName);

function sanitizeTFUnionTypeName(
  unionTypeDef: TypeDef | undefined,
  associatedTypeDef: TypeDef,
): string {
  // If exists, prefer discriminator mapping name.
  const mapping = unionTypeDef?.Discriminator?.Mapping.find(
    (mapping) =>
      associatedTypeDef.Name == mapping.Type.Name ||
      (associatedTypeDef.OriginalName &&
        associatedTypeDef.OriginalName == mapping.Type.OriginalName),
  );
  if (mapping) {
    return sanitizeClassName(getDiscriminatorDisplayName(mapping));
  }
  if (associatedTypeDef.OriginalName) {
    return sanitizeClassName(associatedTypeDef.OriginalName);
  }
  if (associatedTypeDef.Name) {
    return sanitizeClassName(associatedTypeDef.Name);
  }

  switch (associatedTypeDef.Type.toString()) {
    case "string":
      return "Str";
    case "map":
    case "array":
    case "set":
      return sanitizeClassName(
        associatedTypeDef.Type.toString() +
          "_Of_" +
          sanitizeTFUnionTypeName(
            associatedTypeDef,
            associatedTypeDef.ItemType,
          ),
      );
    default:
      return sanitizeClassName(associatedTypeDef.Type.toString());
  }
}

registerTemplateFunc("sanitizeTFUnionTypeName", sanitizeTFUnionTypeName);
