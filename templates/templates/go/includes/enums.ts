// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  let enumNames = getEnumNames(typeDef);

  let enums = [];

  const descriptions = typeDef.Enum.Descriptions || {};

  typeDef.Enum.Values.forEach((value, index) => {
    const description = descriptions[`${value}`];
    if (typeof description === "string" && description) {
      enums.push(
        ...formatGoEnumDescriptionComments(enumNames[index], description),
      );
    }

    enums.push(
      `${enumNames[index]} ${sanitizeClassName(
        typeDef.Name,
      )} = ${sanitizeEnumValue(value, typeDef.Enum.Type.Type)}`,
    );
  });

  return enums.join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);

// @ts-ignore
function templateEnumValueNames(typeDef: TypeDef): string {
  return getEnumNames(typeDef)
    .map((name) => `${name},`)
    .join("\n");
}

registerTemplateFunc("templateEnumValueNames", templateEnumValueNames);

type PackageIdentifiers = {
  taken: Set<string>;
  helpers: Map<string, string>;
};

function allocateEnumValuesFuncName(
  enumType: string,
  taken: Set<string>,
): string {
  let name = `${enumType}Values`;
  if (taken.has(name)) {
    name = `${enumType}EnumValues`;
    for (let i = 2; taken.has(name); i++) {
      name = `${enumType}EnumValues${i}`;
    }
  }
  taken.add(name);
  return name;
}

function unionPackageIdentifiers(
  union: TypeDef,
  className: string,
  outputLocation: string,
): string[] {
  const unionType = getUnionTypeName(className, outputLocation);
  const variants: string[] = [];
  const typeConstants: string[] = [];
  if (union.Discriminator) {
    for (const mapping of union.Discriminator.Mapping) {
      variants.push(sanitizeClassName(getDiscriminatorDisplayName(mapping)));
    }
    typeConstants.push(
      ...getUnionEnumNamesFromDiscriminator(
        unionType,
        union.Discriminator.Mapping,
        outputLocation,
      ),
    );
  } else {
    for (const assocType of union.AssociatedTypes) {
      variants.push(sanitizeClassName(sanitizeUnionTypeName(assocType)));
    }
    typeConstants.push(
      ...getUnionEnumNamesFromAssociatedTypes(
        unionType,
        union.AssociatedTypes,
        outputLocation,
      ),
    );
  }
  if (union.IsUnionOpen) {
    variants.push("Unknown");
    typeConstants.push(openUnionUnknownEnumName(unionType));
  }
  const names = [unionType, `New${className}`, ...typeConstants];
  for (const v of variants) {
    names.push(`New${className}${v}`, `Create${className}${v}`);
  }
  return names;
}

// sdkPackageIdentifiers lists the root package declarations an enum helper
// could collide with: the SDK-level server constants and list, the server
// variable and global parameter setters, and the SDK option functions.
function sdkPackageIdentifiers(): string[] {
  const names = ["ServerList", ...Object.values(getOptionNames())];
  const servers = context.Global.AST.MainSDK.Servers;
  if (servers) {
    if (servers.ServerMap) {
      for (const s of servers.Servers) {
        names.push(templateConstName(s.ID, "Server"));
      }
    }
    for (const v of servers.GetVariables()) {
      names.push(templateSDKOptionName(v.Name));
    }
  }
  const globals = context.Global.AST.MainSDK.Globals;
  if (globals) {
    for (const f of globals.Fields) {
      names.push(templateSDKOptionName(f.Name));
    }
  }
  return names;
}

// optionsPackageIdentifiers lists what options.go and method_options.go
// declare in the operations package: the shared option types and functions,
// and the per-operation method argument options.
function optionsPackageIdentifiers(): string[] {
  const names = ["Options", "Option", ...Object.values(getOptionNames())];
  if (!hasGoOptionalMethodArgumentOptions() && !hasGoMethodSpecificOptions()) {
    return names;
  }
  const controls = [
    "ServerURL",
    "TemplatedServerURL",
    "Timeout",
    "SetHeaders",
    "Retries",
    "Polling",
    "AcceptHeaderOverride",
    "URLOverride",
    "SkipDeserialization",
  ];
  for (const m of getGoOperationOptionMetadataList()) {
    if (m.mode === "method-options") {
      names.push(m.optionsType, m.optionType);
      for (const c of controls) {
        names.push(`With${m.prefix}${c}`);
      }
    }
    for (const a of m.optionalArguments) {
      names.push(a.name, a.supportedOptionName);
    }
  }
  return names;
}

function packageIdentifiers(outputLocation: string): PackageIdentifiers {
  if (!context.GlobalComputed) {
    context.GlobalComputed = {};
  }
  if (!context.GlobalComputed._packageIdentifiers) {
    const byLocation = new Map<string, PackageIdentifiers>();
    const enumsByLocation = new Map<string, string[]>();
    const at = (location: string): PackageIdentifiers => {
      let ids = byLocation.get(location);
      if (!ids) {
        ids = { taken: new Set(), helpers: new Map() };
        byLocation.set(location, ids);
        enumsByLocation.set(location, []);
      }
      return ids;
    };
    for (const [location, models] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      const ids = at(location ?? "");
      const enums = enumsByLocation.get(location ?? "")!;
      for (const [model, types] of sequencedMapEntries(models)) {
        if (
          types.length == 0 ||
          types.some((t) => t.Scope.toString() == "operations")
        ) {
          const servers = context.Global.AST.OperationServers.Get(model)[0];
          if (servers) {
            const modelName = sanitizeFieldName(model);
            ids.taken.add(`${modelName}ServerList`);
            if (servers.ServerMap) {
              for (const s of servers.Servers) {
                ids.taken.add(templateConstName(s.ID, `${modelName}Server`));
              }
            }
          }
        }
        for (const t of types) {
          const name = sanitizeClass(t, location, true, false);
          ids.taken.add(name);
          if (t.Type.toString() == "union") {
            unionPackageIdentifiers(t, name, location).forEach((n) =>
              ids.taken.add(n),
            );
          }
          if (t.Type.toString() == "enum" && t.Enum) {
            getEnumNames(t).forEach((c) => ids.taken.add(c));
            enums.push(name);
          }
        }
      }
    }
    sdkPackageIdentifiers().forEach((n) => at("").taken.add(n));
    const opsLocation = context.Global.Config.Imports.GetOperationsPath() ?? "";
    optionsPackageIdentifiers().forEach((n) => at(opsLocation).taken.add(n));
    for (const [location, ids] of byLocation) {
      for (const name of enumsByLocation.get(location)!.sort()) {
        ids.helpers.set(name, allocateEnumValuesFuncName(name, ids.taken));
      }
    }
    context.GlobalComputed._packageIdentifiers = byLocation;
  }
  let ids = context.GlobalComputed._packageIdentifiers.get(
    outputLocation ?? "",
  );
  if (!ids) {
    ids = { taken: new Set(), helpers: new Map() };
    context.GlobalComputed._packageIdentifiers.set(outputLocation ?? "", ids);
  }
  return ids;
}

// @ts-ignore
function templateEnumValuesFuncName(
  typeDef: TypeDef,
  enumType: string,
): string {
  const ids = packageIdentifiers(typeDef.OutputLocation);
  let name = ids.helpers.get(enumType);
  if (!name) {
    getEnumNames(typeDef).forEach((c) => ids.taken.add(c));
    name = allocateEnumValuesFuncName(enumType, ids.taken);
    ids.helpers.set(enumType, name);
  }
  return name;
}

registerTemplateFunc("templateEnumValuesFuncName", templateEnumValuesFuncName);

function formatGoEnumDescriptionComments(
  enumName: string,
  description: string,
): string[] {
  const sanitized = sanitizeComments(description).split("\n");

  return sanitized.map((line, index) => {
    const trimmed = line.trim();

    if (index === 0) {
      const suffix = trimmed.length > 0 ? ` ${trimmed}` : "";
      return `// ${enumName}${suffix}`;
    }

    return trimmed.length > 0 ? `// ${trimmed}` : "//";
  });
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value.replaceAll(`\\`, `\\\\`).replaceAll(`"`, `\\"`)}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

registerTemplateFunc("sanitizeEnumValue", sanitizeEnumValue);
