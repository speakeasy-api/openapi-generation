type ResolvedTypes = {
  inputType: string;
  importInputTypes: () => string;
  outputType: string;
  importOutputTypes: () => string;
  zod: string;
  importZodTypes: () => string;
};

/**
 * In no-zod mode for a leaf primitive type (date, decimal, bigint, etc.),
 * the public TS type matches the JSON wire shape exactly and no `.zod`
 * fragment is needed — callers no longer read it (matchers dispatch by
 * `enc` alone; func-source ships `payload = input`).
 */
function noZodLeafResolved(wireType: string): ResolvedTypes {
  return {
    inputType: wireType,
    outputType: wireType,
    importInputTypes: () => "",
    importOutputTypes: () => "",
    zod: "",
    importZodTypes: () => "",
  };
}

/** Inbound variant of noZodLeafResolved (no inputType/importInputTypes). */
function noZodInboundLeafResolved(
  wireType: string,
): Omit<ResolvedTypes, "importInputTypes" | "inputType"> {
  return {
    outputType: wireType,
    importOutputTypes: () => "",
    zod: "",
    importZodTypes: () => "",
  };
}

function needsLazyRef(usageLocation: string, a: TypeDef, b: TypeDef) {
  if (!a.IsCustomType() || !b.IsCustomType()) {
    return false;
  }

  if (!a.ResolvedModel || !b.ResolvedModel) {
    return false;
  }

  const aLoc = `${usageLocation}${a.ResolvedModel}`;
  const bLoc = `${getModelsLocation(b.OutputLocation)}${b.ResolvedModel}`;

  if (aLoc === bLoc) {
    return true;
  }

  return areCircular(a, b);
}

//@ts-ignore
function sortUnionMembers(
  union: TypeDef,
): Array<[TypeDef, DiscriminatorMapping?]> {
  if (union.Discriminator == null) {
    // Sort the TypeDefs first, then map to the desired format
    const sortedTypes = sortTypeDefRequiredFieldsDescending(
      union.AssociatedTypes,
    );
    return sortedTypes.map((t): [TypeDef, DiscriminatorMapping?] => [t]);
  } else {
    return union.Discriminator.Mapping.map(
      (m): [TypeDef, DiscriminatorMapping?] => [m.Type, m],
    );
  }
}

function callImportTypesFunc(func: () => string): string {
  assert(func, "func missing for callImportTypesFunc");
  assert(
    typeof func === "function",
    "func is not a function in callImportTypesFunc",
  );
  return func();
}

registerTemplateFunc("callImportTypesFunc", callImportTypesFunc);

function toOutbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  constValue?: unknown;
  defaultValue?: unknown;
  visited: Set<string>;
  optional: boolean;
}): ResolvedTypes {
  const {
    usageLocation,
    typeDef,
    rootTypeDef,
    constValue,
    defaultValue,
    visited,
    optional,
  } = options;

  const rootModel = rootTypeDef.ResolvedModel;
  const typeName = typeDef.Type.toString();
  const zodNS = sanitizeZodRef(typeDef, usageLocation);
  const zodRef = `${zodNS}outboundSchema`;
  const isFormatString = typeDef.Format === "string";
  // Nullable is handled separately, undefined means no const
  const hasConst = constValue != null;
  const hasDefault = defaultValue != null;
  // We can use the const as a default if it's optional or it was marked with default
  const useConstAsDefault =
    hasConst && (isConstFieldsAlwaysOptional() || optional) && hasDefault;

  switch (typeName) {
    case "enum": {
      const enumName = sanitizeClassRef(typeDef, usageLocation);
      const enumTSType = sanitizeTSEnumRef(typeDef, usageLocation);

      if (constValue !== undefined) {
        const literal = enumLiteralFromValue(enumName, typeDef, constValue);
        const zod = useConstAsDefault
          ? z.default(z.literal(literal), literal)
          : z.literal(literal);
        return {
          inputType: literal,
          importInputTypes: () =>
            getEnumFormat(typeDef) === "enum"
              ? addTypeDefImport(typeDef, usageLocation, rootModel)
              : "",
          outputType: literal,
          importOutputTypes: () =>
            getEnumFormat(typeDef) === "enum"
              ? addTypeDefImport(typeDef, usageLocation, rootModel)
              : "",
          zod,
          importZodTypes: () => {
            if (getEnumFormat(typeDef) === "enum") {
              addTypeDefImport(typeDef, usageLocation, rootModel);
            }
            addZodPackageImport();
            return "";
          },
        };
      }

      let zod = zodRef;
      if (defaultValue != null) {
        const literal = enumLiteralFromValue(enumName, typeDef, defaultValue);
        zod = z.default(zod, literal);
      }

      return {
        inputType: enumTSType,
        importInputTypes: () =>
          addEnumTypeImport(typeDef, usageLocation, rootModel),
        outputType: getEnumDataType(typeDef),
        importOutputTypes: () =>
          addEnumTypeImport(typeDef, usageLocation, rootModel),
        zod,
        importZodTypes: () => {
          if (getEnumFormat(typeDef) === "enum" && defaultValue != null) {
            addTypeDefImport(typeDef, usageLocation, rootModel);
          }

          addZodImport(typeDef, usageLocation, rootModel, "outboundSchema");
          return "";
        },
      };
    }
    case "union": {
      // In no-zod, unions emit no $Outbound alias (no discriminating schema
      // means wire shape = TS shape). Reference the plain TS class name.
      const noZodUnionOutbound = isNoZod()
        ? sanitizeClassRef(typeDef, usageLocation) ||
          sanitizeClassName(typeDef.Name)
        : null;

      const regid = typeDef.GetRegistrationID();
      const shouldTruncate = visited.has(regid);
      if (shouldTruncate) {
        return {
          inputType: sanitizeClassName(typeDef.Name),
          importInputTypes: () => "",
          outputType: noZodUnionOutbound ?? `${zodNS}Outbound`,
          importOutputTypes: () => "",
          zod: z.lazy(`() => ${zodRef}`),
          importZodTypes: () => addZodPackageImport(),
        };
      } else {
        visited.add(regid);
      }

      // If we're in a different location to the TypeDef we're resolving then it
      // implies we want to import the union type. Therefore, we need to import
      // a reference to the union instead of building it out. For example, this
      // code path may be hit when an operation takes a request that is a union.
      if (!inSameModel(rootModel, usageLocation, typeDef)) {
        const lazy = needsLazyRef(usageLocation, typeDef, rootTypeDef);
        return {
          inputType: sanitizeClassRef(typeDef, usageLocation),
          importInputTypes: () =>
            addTypeDefImport(typeDef, usageLocation, rootModel),
          outputType: noZodUnionOutbound ?? `${zodNS}Outbound`,
          importOutputTypes: noZodUnionOutbound
            ? () => addTypeDefImport(typeDef, usageLocation, rootModel)
            : () => addZodImport(typeDef, usageLocation, rootModel, "Outbound"),
          zod: lazy ? z.lazy(`() => ${zodRef}`) : zodRef,
          importZodTypes: () => {
            addZodImport(typeDef, usageLocation, rootModel, "outboundSchema");
            if (lazy) {
              addZodPackageImport();
            }
            return "";
          },
        };
      }

      // Unions with a single member should be reduced to resolving the
      // underlying type.
      if (typeDef.AssociatedTypes.length === 1) {
        return resolveOutbound({
          usageLocation,
          typeDef: typeDef.AssociatedTypes[0],
          rootTypeDef,
          optional: false,
          nullable: false,
          streamable: false,
          visited,
        });
      }

      // Iterate over each union member and resolve the runtime type info then
      // it's a matter of concatenating the results to build typescript/zod
      // unions.
      const discr = typeDef.Discriminator?.TypePropertyName;
      const unionInput: string[] = [];
      const unionImportInputTypes: (() => string)[] = [];
      const unionOutput: string[] = [];
      const unionImportOutputTypes: (() => string)[] = [];
      const unionZod: string[] = [];
      const unionImportZodTypes: (() => string)[] = [];
      sortUnionMembers(typeDef).forEach(([at, mapping]) => {
        let {
          inputType,
          importInputTypes,
          outputType,
          importOutputTypes,
          zod,
          importZodTypes,
        } = resolveOutbound({
          usageLocation,
          typeDef: at,
          rootTypeDef,
          optional: false,
          nullable: false,
          streamable: false,
          visited,
        });

        const atName = mapping?.Name;
        if (discr && !atName) {
          throw new Error(
            `union member does not have a discriminator mapping: ${typeDef.Name} > ${at.Name}`,
          );
        }

        // If we're dealing with a discriminated union, we expect union members
        // to be objects. Each object is expected to have a discriminator field
        // the value of that value is explicitly defined in the spec and our ast
        // in the disciminator mappings list. So if these expectations are met
        // then we alter the typings to more explicitly list the discriminator
        // key as having a literal string type.
        if (discr) {
          const dkey = sanitizeKey(discr);
          const dfield = sanitizeFieldName(discr);
          const dacc = sanitizeAccessor("v", dfield);
          let transform = `(v) => ({ ${dkey}: ${dacc} })`;
          if (dfield === discr) {
            transform = "";
          }
          let { literal, importFunc } = resolveDiscriminator(
            usageLocation,
            mapping.Type,
            rootTypeDef,
            discr,
            atName,
          );

          if (!at.DiscriminatorPreApplied) {
            inputType = `(${inputType} & { ${dfield}: ${literal} })`;
            outputType = `(${outputType} & { ${dkey}: ${literal} })`;
            if (at.Type.valueOf() === "error" && isZodV4Mini()) {
              // We can't `.and()` with error classes so we need to use custom instead
              const xacc = sanitizeAccessor("x", dfield);
              zod = z.pipe(
                zod,
                z.custom(
                  outputType,
                  `(x) => ${xacc} === ${literal}`,
                  `'Must be exactly "${literal}"'`,
                ),
              );
            } else {
              zod = z.and(
                zod,
                z.transform(
                  z.object(`{ ${dfield}: ${z.literal(literal)} }`),
                  transform,
                ),
              );
            }
          }

          unionImportInputTypes.push(importFunc);
          unionImportOutputTypes.push(importFunc);
          unionImportZodTypes.push(importFunc);
        }

        unionInput.push(inputType);
        unionImportInputTypes.push(importInputTypes);
        unionOutput.push(outputType);
        unionImportOutputTypes.push(importOutputTypes);
        unionZod.push(zod);
        unionImportZodTypes.push(importZodTypes);
      });

      let zod = useSmartUnion(typeDef)
        ? `smartUnion([${unionZod.join(", ")}])`
        : z.union(unionZod);

      let inputTypeStr = unionInput.join(" | ");
      if (discr && typeDef.IsUnionOpen && !isNoZod()) {
        const discrField = sanitizeFieldName(discr);
        const unknownValue = findUniqueUnknownValue(
          typeDef,
          typeDef.Discriminator.Mapping,
        );
        inputTypeStr += ` | discriminatedUnionTypes.Unknown<"${discrField}"${
          unknownValue === getConstants().defaultUnknownDiscriminatorValue
            ? ""
            : `, "${unknownValue}"`
        }>`;
      }

      return {
        inputType: inputTypeStr,
        importInputTypes: () => {
          unionImportInputTypes.forEach((f) => f());
          if (discr && typeDef.IsUnionOpen && !isNoZod()) {
            addTypeImport(
              getDiscriminatedUnionFileName(),
              "discriminatedUnionTypes",
              usageLocation,
              aliasImport,
            );
          }
          return "";
        },
        outputType: unionOutput.join(" | "),
        importOutputTypes: () => {
          unionImportOutputTypes.forEach((f) => f());
          return "";
        },
        zod,
        importZodTypes: () => {
          addZodPackageImport();
          if (useSmartUnion(typeDef)) {
            addTypeImport(getSmartUnionFileName(), "smartUnion", usageLocation);
          }
          unionImportZodTypes.forEach((f) => f());
          return "";
        },
      };
    }
    case "error":
    case "class": {
      let zod = needsLazyRef(usageLocation, typeDef, rootTypeDef)
        ? z.lazy(`() => ${zodRef}`)
        : zodRef;

      let inputType = sanitizeClassRef(typeDef, usageLocation);
      const className = sanitizeClassName(typeDef.Name);
      if (
        context.RecursiveComputed?.DefiningClassName === className &&
        inputType === className
      ) {
        inputType = `${className}$Model`;
      }

      // In no-zod, no model emits a `$Outbound` wire-shape type — the TS type
      // is the wire shape. Reference the plain TS name in every child.
      const noZodPlainOutbound = isNoZod();
      const outputType = noZodPlainOutbound ? inputType : `${zodNS}Outbound`;
      const importOutputTypes = noZodPlainOutbound
        ? () => addTypeDefImport(typeDef, usageLocation, rootModel)
        : () => addZodImport(typeDef, usageLocation, rootModel, "Outbound");

      return {
        inputType,
        importInputTypes: () =>
          addTypeDefImport(typeDef, usageLocation, rootModel),
        outputType,
        importOutputTypes,
        zod: zod,
        importZodTypes: () => {
          zod.startsWith("z.") && addZodPackageImport();
          addZodImport(typeDef, usageLocation, rootModel, "outboundSchema");
          return "";
        },
      };
    }
    case "string": {
      if (constValue !== undefined) {
        const cv = quote(String(constValue)).replaceAll("{{", `{{"{{"}}`);
        const zod = useConstAsDefault
          ? z.default(z.literal(cv), `${cv} as const`)
          : z.literal(cv);
        return {
          inputType: cv,
          importInputTypes: () => "",
          outputType: cv,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = z.string();
      if (defaultValue != null) {
        const dv = quote(String(defaultValue)).replaceAll("{{", `{{"{{"}}`);
        zod = z.default(zod, dv);
      }

      return {
        inputType: "string",
        importInputTypes: () => "",
        outputType: "string",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "date": {
      // No-zod: ISO date string straight on the wire; no RFCDate runtime class.
      if (isNoZod()) {
        return noZodLeafResolved("string");
      }
      if (isLaxMode()) {
        let dv = defaultValue == null ? "" : `new Date("${defaultValue}")`;
        if (constValue !== undefined) {
          dv = dv ? `new Date("${constValue}")` : "";
        }
        return LaxMode.dateOutbound({ default: dv, usageLocation });
      }

      const inputType = "RFCDate";
      const outputType = "string";
      let zod = z.instanceof("RFCDate");

      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        if (useConstAsDefault) {
          zod = z.default(zod, `new RFCDate(${cv})`);
        }
        zod = z.refine(
          zod,
          `(v) => v.toString() === "${constValue}"`,
          `{message: "Value must be ${constValue}"}`,
        );
        zod = z.transform(zod, `v => v.toString()`);

        return {
          inputType,
          importInputTypes: () =>
            addTypeImport("rfcdate", "RFCDate", usageLocation),
          outputType,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => {
            addZodPackageImport();
            addTypeImport("rfcdate", "RFCDate", usageLocation);
            return "";
          },
        };
      }

      if (defaultValue != null) {
        zod = z.default(zod, `() => new RFCDate("${defaultValue}")`);
      }
      zod = z.transform(zod, "v => v.toString()");

      return {
        inputType,
        importInputTypes: () =>
          addTypeImport("rfcdate", "RFCDate", usageLocation),
        outputType,
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => {
          addZodPackageImport();
          addTypeImport("rfcdate", "RFCDate", usageLocation);
          return "";
        },
      };
    }
    case "date-time": {
      // No-zod: ISO datetime string on the wire; no Date conversion.
      if (isNoZod()) {
        return noZodLeafResolved("string");
      }
      const inputType = "Date";
      const outputType = "string";
      let zod = z.date();

      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        if (useConstAsDefault) {
          zod = z.default(zod, `new Date(${cv})`);
        }
        zod = z.refine(
          zod,
          `(v) => v.getTime() === new Date(${cv}).getTime()`,
          `{message: "Value must be equivalent to ${constValue}"}`,
        );
        zod = z.transform(zod, "v => v.toISOString()");

        return {
          inputType,
          importInputTypes: () => "",
          outputType,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      if (defaultValue != null) {
        zod = z.default(zod, `() => new Date("${defaultValue}")`);
      }
      zod = z.transform(zod, "v => v.toISOString()");

      return {
        inputType,
        importInputTypes: () => "",
        outputType,
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "bigint": {
      // No-zod: bigints are not part of the JSON wire — they ship as either
      // a number (default) or a string (when format=string). No JS bigint
      // conversion happens client-side.
      if (isNoZod()) {
        return noZodLeafResolved(isFormatString ? "string" : "number");
      }
      if (isLaxMode() && !isFormatString) {
        if (constValue !== undefined) {
          let zod = useConstAsDefault
            ? z.default(z.literal(`${constValue}`), `${constValue} as const`)
            : z.literal(`${constValue}`);

          return {
            inputType: `${constValue}`,
            importInputTypes: () => "",
            outputType: `${constValue}`,
            importOutputTypes: () => "",
            zod: zod,
            importZodTypes: () => addZodPackageImport(),
          };
        }

        let zod = z.number();
        if (typeof defaultValue === "number") {
          zod = z.default(zod, `${defaultValue}`);
        }

        return {
          inputType: "number",
          importInputTypes: () => "",
          outputType: "number",
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      if (constValue !== undefined) {
        let zod = z.literal(`BigInt("${constValue}") as ${constValue}n`);
        if (useConstAsDefault) {
          zod = z.default(zod, `BigInt("${constValue}") as ${constValue}n`);
        }
        zod = isFormatString
          ? z.transform(zod, "v => `${v}`")
          : z.transform(zod, "v => Number(v)");
        return {
          inputType: `${constValue}n`,
          importInputTypes: () => "",
          outputType: isFormatString ? "string" : "number",
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = z.bigint();
      if (
        typeof defaultValue === "number" ||
        typeof defaultValue === "string"
      ) {
        zod = z.default(zod, `BigInt("${defaultValue}")`);
      }
      zod = isFormatString
        ? z.transform(zod, "v => `${v}`")
        : z.transform(zod, "v => Number(v)");

      return {
        inputType: "bigint",
        importInputTypes: () => "",
        outputType: isFormatString ? "string" : "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "decimal": {
      // No-zod: no decimal.js Decimal class is created at parse time. Wire
      // shape is `string | number` (number is what JSON has; format=string
      // additionally allows the precise string representation).
      if (isNoZod()) {
        return noZodLeafResolved(isFormatString ? "string | number" : "number");
      }
      const inputType = "Decimal$ | number";
      const outputType = isFormatString ? "string" : "number";
      let zod = z.union([z.instanceof("Decimal$"), z.number()]);
      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        if (useConstAsDefault) {
          zod = z.default(zod, `new Decimal$(${cv})`);
        }
        zod = z.refine(
          zod,
          `(v) => v.toString() === ${cv}`,
          `{message: "Value must be ${constValue}"}`,
        );
        zod = isFormatString
          ? z.transform(zod, "v => `${v}`")
          : z.transform(zod, "v => typeof v === 'number' ? v : v.toNumber()");

        return {
          inputType,
          importInputTypes: () =>
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
          outputType,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => {
            addZodPackageImport();
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
            return "";
          },
        };
      }

      if (
        typeof defaultValue === "number" ||
        typeof defaultValue === "string"
      ) {
        zod = z.default(zod, `() => new Decimal$("${defaultValue}")`);
      }
      zod = isFormatString
        ? z.transform(zod, "v => `${v}`")
        : z.transform(zod, "v => typeof v === 'number' ? v : v.toNumber()");

      return {
        inputType,
        importInputTypes: () =>
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
        outputType,
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => {
          addZodPackageImport();
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
          return "";
        },
      };
    }
    case "integer":
    case "int32": {
      // No-zod: integer-string (format=string) ships as a string on the wire;
      // type the field as `string` so the wire literal matches.
      if (isNoZod()) {
        return noZodLeafResolved(isFormatString ? "string" : "number");
      }
      if (constValue !== undefined) {
        let zod = useConstAsDefault
          ? z.default(z.literal(`${constValue}`), `${constValue} as const`)
          : z.literal(`${constValue}`);
        if (isFormatString) {
          zod = z.transform(zod, "v => `${v}`");
        }

        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          outputType: isFormatString ? `"${constValue}"` : `${constValue}`,
          importOutputTypes: () => "",
          zod: zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = z.int();
      if (typeof defaultValue === "number") {
        zod = z.default(zod, `${defaultValue}`);
      }
      if (isFormatString) {
        zod = z.transform(zod, "v => `${v}`");
      }

      return {
        inputType: "number",
        importInputTypes: () => "",
        outputType: isFormatString ? "string" : "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "number":
    case "float32": {
      // No-zod: number-string (format=string) ships as a string on the wire;
      // type the field as `string` so the wire literal matches.
      if (isNoZod()) {
        return noZodLeafResolved(isFormatString ? "string" : "number");
      }
      if (constValue !== undefined) {
        let zod = useConstAsDefault
          ? z.default(z.literal(`${constValue}`), `${constValue} as const`)
          : z.literal(`${constValue}`);
        if (isFormatString) {
          zod = z.transform(zod, "v => `${v}`");
        }

        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          outputType: isFormatString ? `"${constValue}"` : `${constValue}`,
          importOutputTypes: () => "",
          zod: zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = z.number();
      if (typeof defaultValue === "number") {
        zod = z.default(zod, `${defaultValue}`);
      }
      if (isFormatString) {
        zod = z.transform(zod, "v => `${v}`");
      }

      return {
        inputType: "number",
        importInputTypes: () => "",
        outputType: isFormatString ? "string" : "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "boolean": {
      if (constValue !== undefined) {
        const zod = useConstAsDefault
          ? z.default(z.literal(`${constValue}`), `${constValue} as const`)
          : z.literal(`${constValue}`);
        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          outputType: `${constValue}`,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = z.boolean();
      if (defaultValue === true || defaultValue === false) {
        zod = z.default(zod, `${defaultValue}`);
      }

      return {
        inputType: "boolean",
        importInputTypes: () => "",
        outputType: "boolean",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "bytes":
      return {
        inputType: "Uint8Array | string",
        importInputTypes: () => "",
        outputType: "Uint8Array",
        importOutputTypes: () => "",
        zod: isNoZod() ? "" : "b64$.zodOutbound",
        importZodTypes: isNoZod()
          ? () => ""
          : () =>
              addInternalImport("base64", "b64$", usageLocation, aliasImport),
      };
    case "map":
      const mapType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        nullable: typeDef.ContainsNull,
        optional: false,
        streamable: false,
        visited,
      });

      return {
        inputType: `{ [k: string]: ${mapType.inputType} }`,
        importInputTypes: () => mapType.importInputTypes(),
        outputType: `{ [k: string]: ${mapType.outputType} }`,
        importOutputTypes: () => mapType.importOutputTypes(),
        zod: z.record(mapType.zod),
        importZodTypes: () => {
          mapType.importZodTypes();
          addZodPackageImport();
          return "";
        },
      };
    case "array":
      const arrType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        streamable: false,
        visited,
      });

      let zod = z.array(arrType.zod);
      if (defaultValue != null && typeDef.ItemType.IsPrimitive) {
        zod = z.default(zod, JSON.stringify(defaultValue));
      }

      return {
        inputType: `Array<${arrType.inputType}>`,
        importInputTypes: () => arrType.importInputTypes(),
        outputType: `Array<${arrType.outputType}>`,
        importOutputTypes: () => arrType.importOutputTypes(),
        zod,
        importZodTypes: () => {
          arrType.importZodTypes();
          addZodPackageImport();
          return "";
        },
      };
    case "any":
      return {
        inputType: "any",
        importInputTypes: () => "",
        outputType: "any",
        importOutputTypes: () => "",
        zod: z.any(),
        importZodTypes: () => addZodPackageImport(),
      };
    case "response":
      return {
        inputType: "Response",
        importInputTypes: () => "",
        outputType: "never",
        importOutputTypes: () => "",
        zod: z.transform(
          z.instanceof("Response"),
          `() => { throw new Error("Response cannot be serialized") }`,
        ),
        importZodTypes: () => addZodPackageImport(),
      };
    case "request":
      return {
        inputType: "Request",
        importInputTypes: () => "",
        outputType: "never",
        importOutputTypes: () => "",
        zod: z.transform(
          z.instanceof("Request"),
          `() => { throw new Error("Response cannot be serialized") }`,
        ),
        importZodTypes: () => addZodPackageImport(),
      };
    case "request-stream":
      return {
        inputType:
          "ReadableStream<Uint8Array> | Blob | ArrayBuffer | Uint8Array",
        importInputTypes: () => "",
        outputType:
          "ReadableStream<Uint8Array> | Blob | ArrayBuffer | Uint8Array",
        importOutputTypes: () => "",
        zod: z.union([
          z.instanceof("ReadableStream<Uint8Array>"),
          z.instanceof("Blob"),
          z.instanceof("ArrayBuffer"),
          z.instanceof("Uint8Array"),
        ]),
        importZodTypes: () => addZodPackageImport(),
      };
    case "response-stream":
      return {
        inputType: "ReadableStream<Uint8Array>",
        importInputTypes: () => "",
        outputType: "ReadableStream<Uint8Array>",
        importOutputTypes: () => "",
        zod: z.instanceof("ReadableStream<Uint8Array>"),
        importZodTypes: () => addZodPackageImport(),
      };
    case "event-stream":
      const eventType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        streamable: false,
        visited,
      });

      let eventInputType = eventType.inputType;
      let importFunc = eventType.importInputTypes;

      const dataField = getFlattenedEventStreamField(typeDef.ItemType);
      if (dataField) {
        const dataFieldType = resolveOutbound({
          usageLocation,
          typeDef: dataField.Type,
          rootTypeDef,
          optional: dataField.Optional,
          nullable: dataField.Nullable,
          visited,
        });

        eventInputType = dataFieldType.inputType;
        importFunc = dataFieldType.importInputTypes;
      }

      return {
        inputType: eventStreamTypeRef(eventInputType),
        importInputTypes: () => {
          addEventStreamImport(usageLocation);
          return importFunc();
        },
        outputType: "never",
        importOutputTypes: () => "",
        zod: z.never(),
        importZodTypes: () => addZodPackageImport(),
      };
    case "jsonl":
      const jsonlEventType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        streamable: false,
        visited,
      });
      return {
        inputType: `JsonLStream<${jsonlEventType.inputType}>`,
        importInputTypes: () => {
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          return jsonlEventType.importInputTypes();
        },
        outputType: "never",
        importOutputTypes: () => "",
        zod: z.never(),
        importZodTypes: () => addZodPackageImport(),
      };
    default:
      throw new Error(`Unknown type: ${typeName}`);
  }
}

function resolveOutbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  optional: boolean;
  nullable: boolean;
  streamable?: boolean;
  visited?: Set<string>;
  constValue?: AnyValue;
  defaultValue?: AnyValue;
}): ResolvedTypes {
  const {
    usageLocation,
    rootTypeDef,
    typeDef,
    optional,
    nullable,
    streamable,
    visited,
    constValue,
    defaultValue,
  } = options;

  if (usageLocation == null) {
    throw new Error("usage location cannot be null for: " + typeDef.Name);
  }

  // We can short-circuit type resolution if a field is pinned to null using
  // const in the schema.
  if (constValue?.Value === null) {
    // When constFieldsAlwaysOptional is true (legacy), make all const fields optional
    // When constFieldsAlwaysOptional is false (new default), respect the passed optional parameter
    const shouldMakeOptional = optional || isConstFieldsAlwaysOptional();
    return {
      inputType: shouldMakeOptional ? "null | undefined" : "null",
      importInputTypes: () => "",
      outputType: "null",
      importOutputTypes: () => "",
      zod: shouldMakeOptional
        ? z.default(z.literal("null"), "null")
        : z.literal("null"),
      importZodTypes: () => "",
    };
  }

  let result = toOutbound({
    usageLocation,
    rootTypeDef,
    typeDef,
    constValue: constValue?.Value,
    defaultValue: defaultValue?.Value,
    visited: visited || new Set(),
    optional,
  });

  if (streamable) {
    result.inputType += " | Blob";
    result.outputType += " | Blob";
    result.zod = z.or(result.zod, "blobLikeSchema");
    const origImportZodTypes = result.importZodTypes;
    result.importZodTypes = () => {
      origImportZodTypes();
      addTypeImport("blobs", "blobLikeSchema", usageLocation);
      return "";
    };
  }

  const isNullable = nullable || defaultValue?.Value === null;
  if (isNullable) {
    result.inputType += " | null";
    result.outputType += " | null";
    result.zod = z.nullable(result.zod);
    let importZodTypes = result.importZodTypes;
    result.importZodTypes = () => {
      addZodPackageImport();
      return importZodTypes();
    };
  }
  if (defaultValue?.Value === null) {
    result.zod = z.default(result.zod, "null");
  }

  const hasDefault = defaultValue?.Value !== undefined;

  if (!isConstFieldsAlwaysOptional()) {
    // A type is "response-only" if it IS used in responses but NOT in requests.
    // For response-only types with defaults, the server always provides the value,
    // so don't add | undefined. For request types (or types used in both directions),
    // add | undefined for defaults since the user can omit them.
    const isResponseOnly =
      rootTypeDef &&
      shouldIncludeInboundSchema(rootTypeDef) &&
      !shouldIncludeOutboundSchema(rootTypeDef);

    const shouldAddUndefined =
      isResponseOnly && hasDefault ? false : optional || hasDefault;

    if (shouldAddUndefined) {
      result.inputType += " | undefined";
    }

    if (optional && !hasDefault) {
      result.zod = z.optional(result.zod);
      result.outputType += " | undefined";
    }
  } else {
    // Legacy behavior
    switch (true) {
      // For outbound data, i.e. from the client, we consider const fields optional
      // and fill them in so they become required/set on the way out.
      case constValue?.Value !== undefined:
        result.inputType += " | undefined";
        break;
      case optional && defaultValue?.Value === undefined:
        result.inputType += " | undefined";
        result.outputType += " | undefined";
        result.zod = z.optional(result.zod);
        break;
      case defaultValue?.Value !== undefined:
        result.inputType += " | undefined";
        break;
    }
  }

  return result;
}

registerTemplateFunc("resolveOutbound", resolveOutbound);

function toInbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  visited: Set<string>;
  constValue?: unknown;
  defaultValue?: unknown;
}): Omit<ResolvedTypes, "importInputTypes" | "inputType"> {
  const {
    usageLocation,
    typeDef,
    rootTypeDef,
    constValue,
    defaultValue,
    visited,
  } = options;
  const rootModel = rootTypeDef.ResolvedModel;
  const typeName = typeDef.Type.toString();
  const zodNS = sanitizeZodRef(typeDef, usageLocation);
  const zodRef = `${zodNS}inboundSchema`;
  const isFormatString = typeDef.Format === "string";
  // Nullable is handled separately, undefined means no const
  const hasConst = constValue != null;
  const hasDefault = defaultValue != null;
  // We can use the const as a default if it's optional or it was marked with default
  const useConstAsDefault =
    hasConst && (isConstFieldsAlwaysOptional() || hasDefault);

  switch (typeName) {
    case "enum": {
      const enumName = sanitizeClassRef(typeDef, usageLocation);
      const enumTSType = sanitizeTSEnumRef(typeDef, usageLocation);

      if (constValue !== undefined) {
        const literal = enumLiteralFromValue(enumName, typeDef, constValue);
        let zod = z.literal(literal);
        if (useConstAsDefault) {
          zod = z.default(zod, literal);
        } else if (defaultValue) {
          zod = z.default(
            zod,
            enumLiteralFromValue(enumName, typeDef, defaultValue),
          );
        }

        return {
          outputType: literal,
          importOutputTypes: () =>
            getEnumFormat(typeDef) === "enum"
              ? addTypeDefImport(typeDef, usageLocation, rootModel)
              : "",
          zod,
          importZodTypes: () => {
            if (getEnumFormat(typeDef) === "enum") {
              addTypeDefImport(typeDef, usageLocation, rootModel);
            }
            addZodPackageImport();
            return "";
          },
        };
      }

      let zod = zodRef;
      if (defaultValue != null) {
        const literal = enumLiteralFromValue(enumName, typeDef, defaultValue);
        zod = z.default(zod, literal);
      }

      return {
        outputType: enumTSType,
        importOutputTypes: () =>
          addEnumTypeImport(typeDef, usageLocation, rootModel),
        zod,
        importZodTypes: () => {
          if (getEnumFormat(typeDef) === "enum" && defaultValue != null) {
            addTypeDefImport(typeDef, usageLocation, rootModel);
          }
          addZodImport(typeDef, usageLocation, rootModel, "inboundSchema");
          return "";
        },
      };
    }
    case "union": {
      const regid = typeDef.GetRegistrationID();
      const shouldTruncate = visited.has(regid);
      if (shouldTruncate) {
        return {
          outputType: sanitizeClassName(typeDef.Name),
          importOutputTypes: () => "",
          zod: z.lazy(`() => ${zodRef}`),
          importZodTypes: () => addZodPackageImport(),
        };
      } else {
        visited.add(regid);
      }

      // If we're in a different location to the TypeDef we're resolving then it
      // implies we want to import the union type. Therefore, we need to a
      // reference to the union instead of building it out. For example, this
      // code path may be hit when an operation takes a request that is a union.
      if (!inSameModel(rootModel, usageLocation, typeDef)) {
        const lazy = needsLazyRef(usageLocation, typeDef, rootTypeDef);
        return {
          outputType: sanitizeClassRef(typeDef, usageLocation),
          importOutputTypes: () =>
            addTypeDefImport(typeDef, usageLocation, rootModel),
          zod: lazy ? z.lazy(`() => ${zodRef}`) : zodRef,
          importZodTypes: () => {
            addZodImport(typeDef, usageLocation, rootModel, "inboundSchema");
            if (lazy) {
              addZodPackageImport();
            }
            return "";
          },
        };
      }

      // Unions with a single member should be reduced to resolving the
      // underlying type.
      if (typeDef.AssociatedTypes.length === 1) {
        return resolveInbound({
          usageLocation,
          typeDef: typeDef.AssociatedTypes[0],
          rootTypeDef,
          optional: false,
          nullable: false,
          visited,
        });
      }

      // Iterate over each union member and resolve the runtime type info then
      // it's a matter of concatenating the results to build typescript/zod
      // unions.
      const discr = typeDef.Discriminator?.TypePropertyName;
      const unionOutput: string[] = [];
      const unionImportOutputTypes: (() => string)[] = [];
      const unionZod: string[] = [];
      const unionImportZodTypes: (() => string)[] = [];
      const sortedMappings: DiscriminatorMapping[] = [];
      sortUnionMembers(typeDef).forEach(([at, mapping]) => {
        let { outputType, importOutputTypes, zod, importZodTypes } =
          resolveInbound({
            usageLocation,
            typeDef: at,
            rootTypeDef,
            optional: false,
            nullable: false,
            visited,
          });

        const atName = mapping?.Name;
        if (discr && !atName) {
          throw new Error(
            `Discriminated union member does not have a discriminator mapping: ${typeDef.Name} > ${at.Name}`,
          );
        }

        // If we're dealing with a discriminated union, we expect union members
        // to be objects. Each object is expected to have a discriminator field
        // the value of that value is explicitly defined in the spec and our ast
        // in the disciminator mappings list. So if these expectations are met
        // then we alter the typings to more explicitly list the discriminator
        // key as having a literal string type.
        if (discr) {
          const dkey = sanitizeKey(discr);
          const dfield = sanitizeFieldName(discr);
          const dacc = sanitizeAccessor("v", discr);
          let transform = `(v) => ({ ${dfield}: ${dacc} })`;
          if (dfield === discr) {
            transform = "";
          }

          let { literal, importFunc } = resolveDiscriminator(
            usageLocation,
            mapping.Type,
            rootTypeDef,
            discr,
            atName,
          );

          if (!at.DiscriminatorPreApplied) {
            outputType = `(${outputType} & {${dfield}: ${literal}})`;
            if (at.Type.valueOf() === "error" && isZodV4Mini()) {
              // We can't `.and()` with error classes so we need to use custom instead
              const xacc = sanitizeAccessor("x", dfield);
              zod = z.pipe(
                zod,
                z.custom(
                  outputType,
                  `(x) => ${xacc} === ${literal}`,
                  `'Must be exactly "${literal}"'`,
                ),
              );
            } else {
              zod = z.and(
                zod,
                z.transform(
                  z.object(`{ ${dkey}: ${z.literal(literal)} }`),
                  transform,
                ),
              );
            }
          }

          unionImportOutputTypes.push(importFunc);
          unionImportZodTypes.push(importFunc);
        }

        unionOutput.push(outputType);
        unionImportOutputTypes.push(importOutputTypes);
        unionZod.push(zod);
        unionImportZodTypes.push(importZodTypes);
        if (mapping) {
          sortedMappings.push(mapping);
        }
      });

      let zod = useSmartUnion(typeDef)
        ? `smartUnion([${unionZod.join(", ")}])`
        : z.union(unionZod);
      let outputTypeStr = unionOutput.join(" | ");

      if (discr && typeDef.IsUnionOpen) {
        zod = templateOpenDiscriminatedUnion(
          typeDef,
          discr,
          unionZod,
          sortedMappings,
          usageLocation,
        );
        if (!isNoZod()) {
          const discrField = sanitizeFieldName(discr);
          const unknownValue = findUniqueUnknownValue(typeDef, sortedMappings);
          outputTypeStr += ` | discriminatedUnionTypes.Unknown<"${discrField}"${
            unknownValue === getConstants().defaultUnknownDiscriminatorValue
              ? ""
              : `, "${unknownValue}"`
          }>`;
        }
      }

      return {
        outputType: outputTypeStr,
        importOutputTypes: () => {
          unionImportOutputTypes.forEach((f) => f());
          if (discr && typeDef.IsUnionOpen && !isNoZod()) {
            addTypeImport(
              getDiscriminatedUnionFileName(),
              "discriminatedUnionTypes",
              usageLocation,
              aliasImport,
            );
          }
          return "";
        },
        zod: zod,
        importZodTypes: () => {
          addZodPackageImport();
          if (useSmartUnion(typeDef)) {
            addTypeImport(getSmartUnionFileName(), "smartUnion", usageLocation);
          }
          unionImportZodTypes.forEach((f) => f());
          return "";
        },
      };
    }
    case "error":
    case "class": {
      let zod = needsLazyRef(usageLocation, typeDef, rootTypeDef)
        ? z.lazy(`() => ${zodRef}`)
        : zodRef;
      let outputType = sanitizeClassRef(typeDef, usageLocation);
      const className = sanitizeClassName(typeDef.Name);
      if (
        context.RecursiveComputed?.DefiningClassName === className &&
        outputType === className
      ) {
        outputType = `${className}$Model`;
      }
      return {
        outputType,
        importOutputTypes: () =>
          addTypeDefImport(typeDef, usageLocation, rootModel),
        zod: zod,
        importZodTypes: () => {
          zod.startsWith("z.") && addZodPackageImport();
          addZodImport(typeDef, usageLocation, rootModel, "inboundSchema");
          return "";
        },
      };
    }
    case "string": {
      let dv =
        defaultValue == null
          ? ""
          : `${quote(String(defaultValue)).replaceAll("{{", `{{"{{"}}`)}`;

      if (isLaxMode()) {
        if (constValue !== undefined) {
          const cv = quote(String(constValue)).replaceAll("{{", `{{"{{"}}`);
          return LaxMode.literal({
            const: cv,
            default: dv ? cv : "",
            usageLocation,
          });
        }

        return LaxMode.string({ default: dv, usageLocation });
      }

      if (constValue !== undefined) {
        const cv = quote(String(constValue)).replaceAll("{{", `{{"{{"}}`);
        if (useConstAsDefault) {
          dv = `${cv}`;
        }
        const zod = z.default(z.literal(cv), dv);
        return {
          outputType: cv,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      return {
        outputType: "string",
        importOutputTypes: () => "",
        zod: z.default(z.string(), dv),
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "date": {
      // No-zod: keep the wire ISO date string as-is.
      if (isNoZod()) {
        return noZodInboundLeafResolved("string");
      }
      if (isLaxMode()) {
        let dv = defaultValue == null ? "" : `new Date("${defaultValue}")`;
        if (constValue !== undefined) {
          dv = `new Date("${constValue}")`;
        }
        return LaxMode.date({ default: dv, usageLocation });
      }

      const outputType = "RFCDate";
      let dv = defaultValue == null ? "" : `"${defaultValue}"`;
      if (constValue !== undefined) {
        if (useConstAsDefault) {
          dv = `"${constValue}"`;
        }
        return {
          outputType,
          importOutputTypes: () =>
            addTypeImport("rfcdate", "RFCDate", usageLocation),
          zod: z.transform(
            z.default(z.literal(`"${constValue}"`), dv),
            `v => new RFCDate(v)`,
          ),
          importZodTypes: () => {
            addZodPackageImport();
            addTypeImport("rfcdate", "RFCDate", usageLocation);
            return "";
          },
        };
      }

      return {
        outputType,
        importOutputTypes: () =>
          addTypeImport("rfcdate", "RFCDate", usageLocation),
        zod: z.transform(z.default(z.string(), dv), `v => new RFCDate(v)`),
        importZodTypes: () => {
          addZodPackageImport();
          addTypeImport("rfcdate", "RFCDate", usageLocation);
          return "";
        },
      };
    }
    case "date-time": {
      // No-zod: keep the ISO datetime string from the wire.
      if (isNoZod()) {
        return noZodInboundLeafResolved("string");
      }
      if (isLaxMode()) {
        let dv = defaultValue == null ? "" : `new Date("${defaultValue}")`;
        if (constValue !== undefined) {
          dv = `new Date("${constValue}")`;
        }
        return LaxMode.date({ default: dv, usageLocation });
      }

      const outputType = "Date";
      let dv = defaultValue == null ? "" : `"${defaultValue}"`;
      let zod = z.datetime("{ offset: true }");

      if (constValue !== undefined) {
        if (useConstAsDefault) {
          dv = `"${constValue}"`;
        }
        zod = z.pipe(z.default(`constDateTime("${constValue}")`, dv), zod);
        zod = z.transform(zod, "v => new Date(v)");
        return {
          outputType,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => {
            addTypeImport(
              getConstDateTimeFileName(),
              "constDateTime",
              usageLocation,
            );
            addZodPackageImport();
            return "";
          },
        };
      }

      return {
        outputType,
        importOutputTypes: () => "",
        zod: z.transform(z.default(zod, dv), `v => new Date(v)`),
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "bigint": {
      // No-zod: wire is number (default) or string (format=string). No
      // JS bigint coercion.
      if (isNoZod()) {
        return noZodInboundLeafResolved(isFormatString ? "string" : "number");
      }
      let dv = "";
      if (isLaxMode()) {
        if (isFormatString) {
          dv =
            typeof defaultValue === "number" || typeof defaultValue === "string"
              ? `${defaultValue}n`
              : "";
          if (constValue !== undefined) {
            const cv = `${constValue}n`;
            dv = dv ? cv : "";
            return LaxMode.literalBigInt({
              const: cv,
              default: dv,
              usageLocation,
            });
          }

          return LaxMode.bigint({ default: dv, usageLocation });
        }

        dv = typeof defaultValue === "number" ? `${defaultValue}` : "";
        if (constValue !== undefined) {
          return LaxMode.literal({
            const: `${constValue}`,
            default: dv ? `${constValue}` : "",
            usageLocation,
          });
        }

        return LaxMode.number({ default: dv, usageLocation });
      }

      const outputType = "bigint";

      let zod = isFormatString ? z.string() : z.number();
      if (
        typeof defaultValue === "number" ||
        (isFormatString && typeof defaultValue === "string")
      ) {
        zod = isFormatString
          ? z.default(zod, `"${defaultValue}"`)
          : z.default(zod, `${defaultValue}`);
      } else if (useConstAsDefault) {
        zod = z.default(zod, `${constValue}`);
      }
      zod = z.transform(zod, `v => BigInt(v)`);

      if (constValue !== undefined) {
        zod = z.pipe(
          zod,
          z.literal(`BigInt("${constValue}") as ${constValue}n`),
        );

        return {
          outputType: `BigInt(${constValue})`,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      return {
        outputType,
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "decimal": {
      // No-zod: wire is number (or string when format=string). No Decimal
      // class instances are created.
      if (isNoZod()) {
        return noZodInboundLeafResolved(
          isFormatString ? "string | number" : "number",
        );
      }
      const outputType = "Decimal$";
      let zod = isFormatString ? z.string() : z.number();

      if (constValue !== undefined) {
        zod = isFormatString
          ? z.literal(`"${constValue}"`)
          : z.literal(`${constValue}`);
        if (typeof defaultValue === "number") {
          zod = isFormatString
            ? z.default(zod, `"${defaultValue}" as const`)
            : z.default(zod, `${defaultValue} as const`);
        } else if (useConstAsDefault) {
          zod = z.default(zod, `${constValue} as const`);
        }
        zod = z.transform(zod, `v => new Decimal$(v)`);

        return {
          outputType,
          importOutputTypes: () =>
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
          zod,
          importZodTypes: () => {
            addZodPackageImport();
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
            return "";
          },
        };
      }

      if (
        typeof defaultValue === "number" ||
        typeof defaultValue === "string"
      ) {
        // This is better than using .default() because if the default value is
        // a literal number then it might overflow or be truncated at runtime if
        // interpolated into .default(). Interpolating into a string preserves
        // the literal value to be passed into the Decimal constructor.
        zod = z.transform(zod, `v => new Decimal$(v ?? "${defaultValue}")`);
      } else {
        zod = z.transform(zod, "v => new Decimal$(v)");
      }

      return {
        outputType,
        importOutputTypes: () =>
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
        zod,
        importZodTypes: () => {
          addZodPackageImport();
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
          return "";
        },
      };
    }
    case "integer":
    case "int32": {
      // No-zod: integer-string (format=string) arrives as a string on the
      // wire; type as `string` to match the JSON literal verbatim.
      if (isNoZod()) {
        return noZodInboundLeafResolved(isFormatString ? "string" : "number");
      }
      let dv = "";

      if (isLaxMode()) {
        dv = typeof defaultValue === "number" ? `${defaultValue}` : "";
        if (constValue !== undefined) {
          return LaxMode.literal({
            const: `${constValue}`,
            default: dv ? `${constValue}` : "",
            usageLocation,
          });
        }

        return LaxMode.number({ default: dv, usageLocation });
      }

      if (typeof defaultValue === "number") {
        dv = isFormatString ? `"${defaultValue}"` : `${defaultValue}`;
      }

      if (constValue !== undefined) {
        if (useConstAsDefault) {
          dv ||= `${constValue}`;
        }

        const zod = isFormatString
          ? z.transform(
              z.default(z.literal(`"${constValue}"`), dv),
              `v => parseInt(v, 10)`,
            )
          : z.default(z.literal(String(constValue)), dv);

        return {
          outputType: `${constValue}`,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = isFormatString ? z.string() : z.int();
      zod = z.default(zod, dv);
      if (isFormatString) {
        zod = z.transform(zod, "v => parseInt(v, 10)");
      }

      return {
        outputType: "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "number":
    case "float32": {
      // No-zod: number-string (format=string) arrives as a string on the
      // wire; type as `string` to match the JSON literal verbatim.
      if (isNoZod()) {
        return noZodInboundLeafResolved(isFormatString ? "string" : "number");
      }
      let dv = "";
      if (typeof defaultValue === "number") {
        dv = isFormatString ? `"${defaultValue}"` : `${defaultValue}`;
      }

      if (isLaxMode()) {
        dv = typeof defaultValue === "number" ? `${defaultValue}` : "";
        if (constValue !== undefined) {
          return LaxMode.literal({
            const: `${constValue}`,
            default: dv ? `${constValue}` : "",
            usageLocation,
          });
        }

        return LaxMode.number({ default: dv, usageLocation });
      }

      if (constValue !== undefined) {
        if (useConstAsDefault) {
          dv ||= `${constValue}`;
        }

        const zod = isFormatString
          ? z.transform(
              z.default(z.literal(`"${constValue}"`), dv),
              `v => parseFloat(v)`,
            )
          : z.default(z.literal(String(constValue)), dv);

        return {
          outputType: `${constValue}`,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => addZodPackageImport(),
        };
      }

      let zod = isFormatString ? z.string() : z.number();
      zod = z.default(zod, dv);
      if (isFormatString) {
        zod = z.transform(zod, "v => parseFloat(v)");
      }

      return {
        outputType: "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "boolean": {
      let dv =
        defaultValue === true || defaultValue === false
          ? `${defaultValue}`
          : "";

      if (isLaxMode()) {
        if (constValue !== undefined) {
          return LaxMode.literal({
            const: `${constValue}`,
            default: dv ? `${constValue}` : "",
            usageLocation,
          });
        }

        return LaxMode.boolean({ default: dv, usageLocation });
      }

      if (constValue !== undefined) {
        if (useConstAsDefault) {
          dv ||= `${constValue}`;
        }
        return {
          outputType: `${constValue}`,
          importOutputTypes: () => "",
          zod: z.default(z.literal(`${constValue}`), dv),
          importZodTypes: () => addZodPackageImport(),
        };
      }

      return {
        outputType: "boolean",
        importOutputTypes: () => "",
        zod: z.default(z.boolean(), dv),
        importZodTypes: () => addZodPackageImport(),
      };
    }
    case "bytes":
      return {
        outputType: "Uint8Array",
        importOutputTypes: () => "",
        zod: isNoZod() ? "" : `b64$.zodInbound`,
        importZodTypes: isNoZod()
          ? () => ""
          : () =>
              addInternalImport("base64", "b64$", usageLocation, aliasImport),
      };
    case "map":
      const mapType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        visited,
      });

      return {
        outputType: `{ [k: string]: ${mapType.outputType} }`,
        importOutputTypes: () => mapType.importOutputTypes(),
        zod: z.record(mapType.zod),
        importZodTypes: () => {
          mapType.importZodTypes();
          addZodPackageImport();
          return "";
        },
      };
    case "array":
      const arrType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        visited,
      });

      let zod = z.array(arrType.zod);
      if (defaultValue != null && typeDef.ItemType.IsPrimitive) {
        zod = z.default(zod, JSON.stringify(defaultValue));
      }

      return {
        outputType: `Array<${arrType.outputType}>`,
        importOutputTypes: () => arrType.importOutputTypes(),
        zod,
        importZodTypes: () => {
          arrType.importZodTypes();
          addZodPackageImport();
          return "";
        },
      };
    case "any":
      return {
        outputType: "any",
        importOutputTypes: () => "",
        zod: z.any(),
        importZodTypes: () => addZodPackageImport(),
      };
    case "response":
      return {
        outputType: "Response",
        importOutputTypes: () => "",
        zod: z.instanceof("Response"),
        importZodTypes: () => addZodPackageImport(),
      };
    case "request":
      return {
        outputType: "Request",
        importOutputTypes: () => "",
        zod: z.instanceof("Request"),
        importZodTypes: () => addZodPackageImport(),
      };
    case "request-stream":
      return {
        outputType:
          "ReadableStream<Uint8Array> | Blob | ArrayBuffer | Uint8Array",
        importOutputTypes: () => "",
        zod: z.union([
          z.instanceof("ReadableStream<Uint8Array>"),
          z.instanceof("Blob"),
          z.instanceof("ArrayBuffer"),
          z.instanceof("Uint8Array"),
        ]),
        importZodTypes: () => addZodPackageImport(),
      };
    case "response-stream":
      return {
        outputType: "ReadableStream<Uint8Array>",
        importOutputTypes: () => "",
        zod: z.instanceof("ReadableStream<Uint8Array>"),
        importZodTypes: () => addZodPackageImport(),
      };
    case "event-stream":
      const eventType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        visited,
      });

      let eventOutputType = eventType.outputType;
      let parseExpression = `${eventType.zod}.parse(rawEvent)`;
      let importFunc = eventType.importOutputTypes;

      const dataField = getFlattenedEventStreamField(typeDef.ItemType);
      if (dataField) {
        const dataFieldType = resolveInbound({
          usageLocation,
          typeDef: dataField.Type,
          rootTypeDef,
          optional: dataField.Optional,
          nullable: dataField.Nullable,
          visited,
        });

        eventOutputType = dataFieldType.outputType;
        parseExpression = `${z.parse(eventType.zod, "rawEvent")}?.data`;
        importFunc = dataFieldType.importOutputTypes;
      }

      if (isNoZod()) {
        // No-zod: the matcher dispatches the SSE wrap via
        // `wrapEventStreamResponse(response.body, opts)`. responses.ts
        // gathers sentinel/flattened/dataRequired options from the typeDef
        // and passes them on the matcher; this resolver only surfaces the
        // user-facing configured event stream wrapper type.
        return {
          outputType: eventStreamTypeRef(eventOutputType),
          importOutputTypes: () => {
            addEventStreamImport(usageLocation);
            return importFunc();
          },
          zod: "",
          importZodTypes: () => "",
        };
      }

      const handleSentinel =
        typeDef.EventStreamSentinel !== ""
          ? `if (rawEvent.data === "${typeDef.EventStreamSentinel}") return { done: true, value: undefined };`
          : "";

      const sseDataRequired = isSSEDataRequired(typeDef.ItemType);
      const sseOpts = sseDataRequired ? "" : `, { dataRequired: false }`;

      const transformBody = `stream => {
          return new ${eventStreamClassName()}(stream, rawEvent => {
            ${handleSentinel}
            return { done: false, value: ${parseExpression} };
        }${sseOpts})}`;

      const sseZod = z.transform(
        z.instanceof("ReadableStream<Uint8Array>"),
        transformBody,
        true,
      );

      return {
        outputType: eventStreamTypeRef(eventOutputType),
        importOutputTypes: () => {
          addEventStreamImport(usageLocation);
          return importFunc();
        },
        zod: sseZod,
        importZodTypes: () => {
          eventType.importZodTypes();
          addEventStreamImport(usageLocation);
          addZodPackageImport();
          return "";
        },
      };
    case "jsonl":
      const jsonlEventType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        visited,
      });
      const jsonLDecoder = `decoder(rawEvent) { const schema =  ${jsonlEventType.zod}; return schema.parse(rawEvent); }`;

      return {
        outputType: `JsonLStream<${jsonlEventType.outputType}>`,
        importOutputTypes: () => {
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          return jsonlEventType.importOutputTypes();
        },
        zod: z.transform(
          z.instanceof("ReadableStream<Uint8Array>"),
          `stream => { return new JsonLStream({stream, ${jsonLDecoder}}) }`,
        ),
        importZodTypes: () => {
          jsonlEventType.importZodTypes();
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          addZodPackageImport();
          return "";
        },
      };
    default:
      throw new Error(`Unknown type: ${typeName}`);
  }
}

function resolveInbound(options: {
  usageLocation: string;
  typeDef?: TypeDef;
  rootTypeDef: TypeDef;
  optional: boolean;
  nullable: boolean;
  visited?: Set<string>;
  encoding?: string;
  constValue?: AnyValue;
  defaultValue?: AnyValue;
}): ResolvedTypes & { inputType: "unknown" } {
  const {
    usageLocation,
    rootTypeDef,
    typeDef,
    optional,
    nullable,
    constValue,
    defaultValue,
    visited,
    encoding,
  } = options;

  if (!typeDef) {
    return {
      outputType: "void",
      inputType: "unknown",
      zod: isNoZod() ? "" : z.void(),
      importInputTypes: () => "",
      importOutputTypes: () => "",
      importZodTypes: () => {
        if (isNoZod()) return "";
        return addZodPackageImport();
      },
    };
  }

  const hasConst = constValue?.Value !== undefined;
  const hasDefault = defaultValue?.Value !== undefined;
  const useConstAsDefault = hasConst && hasDefault;
  const isOptional = optional && !hasDefault;

  // We can short-circuit type resolution if a field is pinned to null using
  // const in the schema.
  if (constValue?.Value === null) {
    let outputType = "null";
    outputType += isOptional ? " | undefined" : "";

    let zod = z.literal("null");
    zod = defaultValue ? z.default(zod, "null") : zod;
    zod = isOptional ? z.optional(zod) : zod;

    return {
      inputType: "unknown",
      importInputTypes: () => "",
      outputType,
      importOutputTypes: () => "",
      zod,
      importZodTypes: () => addZodPackageImport(),
    };
  }

  if (usageLocation == null) {
    throw new Error("usage location cannot be null for: " + typeDef.Name);
  }
  let result = {
    ...toInbound({
      usageLocation,
      rootTypeDef,
      typeDef,
      constValue: constValue?.Value,
      defaultValue: defaultValue?.Value,
      visited: visited || new Set(),
    }),
    inputType: "unknown" as const,
    importInputTypes: () => "",
  };

  // If it's optional and nullable, no need for any smart handling
  const laxModeOptionalNullable =
    isLaxMode() && [optional, nullable].filter(Boolean).length < 2;

  const isNullable =
    nullable || defaultValue?.Value === null || constValue?.Value === null;
  if (isNullable) {
    result.outputType += " | null";
    // If the default value is null then each type resolver needs to incorporate
    // that into its zod pipeline.
    if (laxModeOptionalNullable) {
      LaxMode.nullable({ result, usageLocation });
    } else {
      result.zod = z.nullable(result.zod);
      let importZodTypes = result.importZodTypes;
      result.importZodTypes = () => {
        addZodPackageImport();
        return importZodTypes();
      };
    }
  }
  const useConstNullAsDefault = useConstAsDefault && constValue?.Value === null;
  if (defaultValue?.Value === null || useConstNullAsDefault) {
    result.zod = z.default(result.zod, "null");
  }

  if (isOptional) {
    if (laxModeOptionalNullable) {
      LaxMode.optional({ result, usageLocation });
    } else {
      result.zod = z.optional(result.zod);
    }
  }

  if (isOptional || hasDefault) {
    result.outputType += " | undefined";
  }

  if (encoding === "application/json") {
    const undefinedCheck = isOptional
      ? "if (v === undefined) { return undefined; } "
      : "";
    const parseIfStringFn = `(v) => { ${undefinedCheck}if (typeof v !== \"string\") { return v; }`;
    const parseIfStringFnWithCtx = `(v, ctx) => { ${undefinedCheck}if (typeof v !== \"string\") { return v; }`;

    // If this is a mixed union (JSON + string variants), try JSON parse but
    // fall back to raw string instead of failing — the union schema will match
    // the correct variant.
    if (typeDef && isSSEMixedUnion(typeDef)) {
      const transformFn = `${parseIfStringFn}try { return JSON.parse(v); } catch { return v; } }`;
      const rawInput = isOptional ? z.optional(z.unknown()) : z.unknown();
      result.zod = z.pipe(z.transform(rawInput, transformFn), result.zod);
    } else {
      const issueStatement = z.addIssue({
        code: "custom",
        message: "`malformed json: ${err}`",
        input: "v",
      });

      // Allow already-parsed JSON values (e.g. when the same SSE-tagged
      // component is reused in a JSON-body response branch) while still
      // catching malformed JSON strings.
      const transformFn = `${parseIfStringFnWithCtx}try { return JSON.parse(v); } catch (err) { ${issueStatement}; return ${z.NEVER()}; } }`;
      const rawInput = isOptional ? z.optional(z.unknown()) : z.unknown();
      result.zod = z.pipe(z.transform(rawInput, transformFn), result.zod);
    }
  }

  return result;
}

registerTemplateFunc("resolveInbound", resolveInbound);

function enumLiteralFromValue(
  enumName: string,
  typeDef: TypeDef,
  value: unknown,
) {
  if (getEnumFormat(typeDef) === "union") {
    return typeof value === "string" ? quote(value) : `${value}`;
  }

  const idx = typeDef.Enum?.Values.findIndex((v) => v === `${value}`);
  if (idx == null || idx < 0) {
    throw new Error(`Value is not a member of enum "${enumName}": ${value}`);
  }

  const memberNames = getEnumNames(typeDef);
  return `${enumName}.${memberNames[idx]}`;
}

function shouldMakeAnyTypeRequired(): boolean {
  // Only apply this fix when using Zod v4 or v4-mini
  // Zod v3 has different behavior that requires `any` types to be optional
  return isZodV4();
}

function isInputOptional(fieldDef: FieldDef, rootType?: TypeDef) {
  // When using Zod v3, treat `any` types as optional (legacy behavior)
  // When using Zod v4/v4-mini, respect the schema's required status for `any` types
  if (!shouldMakeAnyTypeRequired()) {
    if (
      fieldDef.Type?.Type.toString() === "any" ||
      // Any unions formed with an `any` member are effectively reduced to `any`
      // which means we treat them as optional.
      fieldDef.Type?.AssociatedTypes?.some((at) => at.Type.toString() === "any")
    ) {
      return true;
    }
  }

  const hasConst = fieldDef.Const?.Value !== undefined;
  const hasDefault = fieldDef.Default?.Value !== undefined;

  // When constFieldsAlwaysOptional is true (legacy), make all const fields optional
  // When constFieldsAlwaysOptional is false (new default), respect the required status
  if (hasConst && isConstFieldsAlwaysOptional()) {
    return true;
  }

  // For response-only types, fields with defaults are required (server provides them)
  // For request types (or types used in both), fields with defaults are optional
  if (hasDefault && rootType) {
    const isResponseOnly =
      shouldIncludeInboundSchema(rootType) &&
      !shouldIncludeOutboundSchema(rootType);
    if (isResponseOnly) {
      return false; // Field is required since server always provides the default
    }
  }

  return fieldDef.Optional || hasDefault;
}

registerTemplateFunc("isInputOptional", isInputOptional);

function isErrorFieldOptional(fieldDef: FieldDef) {
  // When using Zod v3, treat `any` types as optional (legacy behavior)
  // When using Zod v4/v4-mini, respect the schema's required status for `any` types
  if (!shouldMakeAnyTypeRequired()) {
    if (
      fieldDef.Type?.Type.toString() === "any" ||
      // Any unions formed with an `any` member are effectively reduced to `any`
      // which means we treat them as optional.
      fieldDef.Type?.AssociatedTypes?.some((at) => at.Type.toString() === "any")
    ) {
      return true;
    }
  }

  const hasDefault = fieldDef.Default?.Value !== undefined;

  return fieldDef.Optional || hasDefault;
}

registerTemplateFunc("isErrorFieldOptional", isErrorFieldOptional);

function isOutputOptional(fieldDef: FieldDef) {
  // When using Zod v3, treat `any` types as optional (legacy behavior)
  // When using Zod v4/v4-mini, respect the schema's required status for `any` types
  if (!shouldMakeAnyTypeRequired()) {
    if (
      fieldDef.Type?.Type.toString() === "any" ||
      // Any unions formed with an `any` member are effectively reduced to `any`
      // which means we treat them as optional.
      fieldDef.Type?.AssociatedTypes?.some((at) => at.Type.toString() === "any")
    ) {
      return true;
    }
  }

  const hasConst = fieldDef.Const?.Value !== undefined;
  // No-zod: defaults are not injected at runtime, so a field with a default
  // can still be absent from the response — treat it as TS-optional. Ignore
  // hasDefault so the field's spec-level Optional status drives the decision.
  const hasDefault = isNoZod() ? false : fieldDef.Default?.Value !== undefined;

  if (isConstFieldsAlwaysOptional()) {
    return fieldDef.Optional && !hasDefault && !hasConst;
  }

  return fieldDef.Optional && !hasDefault;
}

registerTemplateFunc("isOutputOptional", isOutputOptional);

function resolveDiscriminator(
  usageLocation: string,
  typeDef: TypeDef,
  rootTypeDef: TypeDef,
  fieldName: string,
  fieldValue: string,
): { literal: string; type: TypeDef; importFunc: () => string } {
  const type = typeDef.Type.toString();
  if (type !== "class" && type !== "union" && !isUnionOfErrors(rootTypeDef)) {
    throw new Error(
      "Expected discriminator to map over a union of objects but got: " + type,
    );
  }

  let resolvedTypeDef = typeDef;
  if (type === "union") {
    // resolvedTypeDef is used to lookup the discriminator field.
    // the discriminator should be homogenous(present, and all of the same type) across union members.
    // we pick the first union member in such cases.
    resolvedTypeDef = typeDef.AssociatedTypes[0];
  }

  const field = resolvedTypeDef.Fields.find((f) => f.Name === fieldName);
  if (!field) {
    throw new Error(
      `Discriminator field "${fieldName}" not found on "${resolvedTypeDef.OriginalName}"`,
    );
  }

  switch (field.Type?.Type.toString()) {
    case "enum": {
      const enumName = sanitizeClassRef(field.Type, usageLocation);
      return {
        literal: enumLiteralFromValue(enumName, field.Type, fieldValue),
        type: field.Type,
        importFunc: () =>
          getEnumFormat(field.Type) === "enum"
            ? addTypeDefImport(
                field.Type,
                usageLocation,
                rootTypeDef.ResolvedModel,
              )
            : "",
      };
    }
    default: {
      return {
        literal: `"${fieldValue}"`,
        type: field.Type,
        importFunc: () => "",
      };
    }
  }
}

function toEnvSchema(fieldDef: FieldDef, usageLocation: string): string {
  const type = fieldDef.Type.Type.toString();
  const dv = fieldDef.Default?.Value;

  switch (type) {
    case "enum": {
      const types = toInbound({
        typeDef: fieldDef.Type,
        rootTypeDef: context.Global.AST.MainSDK.Type,
        constValue: undefined,
        defaultValue: dv,
        visited: new Set(),
        usageLocation,
      });
      types.importZodTypes();
      return types.zod;
    }
    case "any":
      return injectDefault(z.any(), dv);
    case "string":
      return injectDefault(z.string(), dv);
    case "date":
      return injectDefault(z.coerce.date("{ offset: true }"), dv);
    case "date-time":
      addTypeImport("rfcdate", "RFCDate", usageLocation);
      return z.transform(injectDefault(z.string(), dv), "v => RFCDate(v)");
    case "bigint":
      return injectDefault(z.bigint(), dv);
    case "number":
    case "float32":
      return injectDefault(z.number(), dv);
    case "integer":
    case "int32":
      return injectDefault(z.int(), dv);
    case "boolean": {
      let zod = injectDefault(
        'z.enum(["true", "false"])',
        typeof dv === "undefined" ? undefined : String(dv),
      );
      return z.transform(zod, 'v => v === "true"');
    }
    case "array": {
      const itemType = toEnvSchema(
        typeDefToFieldDef(fieldDef.Type.ItemType, fieldDef),
        usageLocation,
      );
      const zod = injectDefault(z.array(itemType), dv);

      return z.pipe(z.transform(z.string(), "v => v ? v.split(',') : []"), zod);
    }
    default:
      throw new Error(
        `Global parameter has unsupported type: ${fieldDef.Name}: ${type}`,
      );
  }
}

function injectDefault(schema: string, defaultValue: unknown): string {
  if (typeof defaultValue === "undefined") {
    return schema;
  }

  return z.default(schema, JSON.stringify(defaultValue));
}

function resolveEnvSchema(
  fieldDef: FieldDef,
  usageLocation: string,
  options?: {
    /**
     * Useful when we want the schema without worrying about optionality.
     * Specific example of this is the MCP CLI the `parse` callback for a flag
     * is only called when a value is passed in and expects a non-optional
     * return value.
     */
    ignoreOptional?: boolean;
  },
): string {
  let schema = toEnvSchema(fieldDef, usageLocation);
  if (fieldDef.Nullable) {
    schema = z.nullable(schema);
  }
  if (!options?.ignoreOptional && fieldDef.Default == null) {
    schema = z.optional(schema);
  }
  return schema;
}
registerTemplateFunc("resolveEnvSchema", resolveEnvSchema);

function resolveURISchema(fieldDef: FieldDef, usageLocation: string): string {
  const typeDef = fieldDef.Type;
  const type = typeDef.Type.toString() as DataType;
  const isFormatString = typeDef.Format === "string";
  const dv = fieldDef.Default?.Value;

  const inboundTypes = toInbound({
    typeDef: typeDef,
    rootTypeDef: context.Global.AST.MainSDK.Type,
    defaultValue: dv,
    visited: new Set(),
    usageLocation,
  });

  const pass = Symbol("passthrough");
  const unsupported = Symbol("unsupported");

  const resolvers: Record<
    DataType,
    (() => string | typeof pass) | typeof pass | typeof unsupported
  > = {
    number: () => z.optional(z.coerce.number()),
    boolean: () =>
      z.transform(
        z.optional(z.enum('["true", "false"]')),
        "v => v == null ? void 0 : v === 'true'",
      ),
    integer: () => z.coerce.int(),
    int32: () => z.coerce.int(),
    float32: () => z.optional(z.coerce.number()),
    bigint: () => (isFormatString ? pass : z.optional(z.coerce.bigint())),
    decimal: () => (isFormatString ? pass : z.optional(z.coerce.number())),
    map: () =>
      z.transform(
        z.optional(z.coerce.string()),
        "v => v ? JSON.parse(v) : undefined",
      ),
    array: () => {
      const itemType = resolveURISchema(
        typeDefToFieldDef(fieldDef.Type.ItemType, fieldDef),
        usageLocation,
      );

      const item = z.array(itemType);

      return z.pipe(
        z.transform(
          z.optional(z.coerce.string()),
          "v => v ? v.split(',') : []",
        ),
        item,
      );
    },
    class: () =>
      z.transform(
        z.optional(z.coerce.string()),
        "v => v ? JSON.parse(v) : undefined",
      ),

    union: () => {
      if (typeDef.AssociatedTypes.length === 1) {
        return resolveURISchema(
          typeDefToFieldDef(typeDef.AssociatedTypes[0], fieldDef),
          usageLocation,
        );
      }

      const all = typeDef.AssociatedTypes.map((t) =>
        resolveURISchema(typeDefToFieldDef(t, fieldDef), usageLocation),
      );

      return z.union(all);
    },

    string: pass,
    date: pass,
    "date-time": pass,
    any: pass,
    bytes: pass,
    enum: pass,

    "event-stream": unsupported,
    jsonl: unsupported,
    set: unsupported,
    response: unsupported,
    request: unsupported,
    error: unsupported,
    "request-stream": unsupported,
    "response-stream": unsupported,
  };

  const resolver = resolvers[type];
  if (resolver === unsupported) {
    throw new Error(
      `resolveURISchema: ${fieldDef.Name}: unsupported data type: ${type}`,
    );
  }

  if (resolver === pass) {
    inboundTypes.importZodTypes();
    return inboundTypes.zod;
  }

  const firstStage = resolver();
  if (firstStage === pass) {
    inboundTypes.importZodTypes();
    return inboundTypes.zod;
  }

  inboundTypes.importZodTypes();
  addZodPackageImport();

  return z.pipe(firstStage, inboundTypes.zod);
}
registerTemplateFunc("resolveURISchema", resolveURISchema);

// @ts-ignore
function unionMemberIsSelfReferential(
  member: TypeDef,
  rootModel: string,
  usageLocation: string,
): boolean {
  if (member.IsCustomType()) {
    return inSameModel(rootModel, usageLocation, member);
  } else if (member.IsContainer()) {
    return unionMemberIsSelfReferential(
      member.ItemType,
      rootModel,
      usageLocation,
    );
  }
  return false;
}

function isConstFieldsAlwaysOptional(): boolean {
  const value = context.Global.Config.ConstFieldsAlwaysOptional;
  // Done like this because true is the default value
  if (value === false || value === "false") return false;
  return true;
}

registerTemplateFunc(
  "isConstFieldsAlwaysOptional",
  isConstFieldsAlwaysOptional,
);

function isFlatAdditionalProperties(): boolean {
  const value = context.Global.Config.FlatAdditionalProperties;
  return value === true || value === "true";
}

registerTemplateFunc("isFlatAdditionalProperties", isFlatAdditionalProperties);

/**
 * Checks if an additional properties field can be safely flattened.
 * Flattening is only safe when additionalProperties is `true` (any type),
 * not when it has a typed value like `{ type: string }`.
 *
 * When flattening typed additional properties, the type information is lost
 * because index signatures use `unknown`, which would cause TypeScript errors.
 */
function canFlattenAdditionalProperties(field: FieldDef): boolean {
  if (!field || !field.IsAdditionalProperties) {
    return false;
  }

  // Check if the ItemType is "any" - this indicates additionalProperties: true
  // If it has any other type (string, number, object, etc.), we can't safely flatten
  const itemType = field.Type?.ItemType;
  if (!itemType) {
    return true; // No item type means it's effectively `any`
  }

  return itemType.Type?.toString() === "any";
}

registerTemplateFunc(
  "canFlattenAdditionalProperties",
  canFlattenAdditionalProperties,
);

/**
 * Combined check: is flat mode enabled AND can this field be safely flattened?
 */
function shouldFlattenAdditionalProperties(field: FieldDef): boolean {
  return isFlatAdditionalProperties() && canFlattenAdditionalProperties(field);
}

registerTemplateFunc(
  "shouldFlattenAdditionalProperties",
  shouldFlattenAdditionalProperties,
);

/* Isomorphic Zod helpers to support both v3, v4 and v4-mini */
const z = {
  Enum: (typeofEnum: string): string => {
    if (isZodV4Mini()) {
      return `z.ZodMiniEnum<${typeofEnum}>`;
    }
    if (isZodV4()) {
      return `z.ZodEnum<${typeofEnum}>`;
    }
    return `z.ZodNativeEnum<${typeofEnum}>`;
  },
  enum: (enumRef: string): string => {
    if (isZodV4()) {
      return `z.enum(${enumRef})`;
    }
    return `z.nativeEnum(${enumRef})`;
  },
  record: (valueSchema: string): string => {
    if (isZodV4()) {
      return `z.record(z.string(), ${valueSchema})`;
    }
    return `z.record(${valueSchema})`;
  },
  ZodType: (output?: string, input?: string): string => {
    if (isZodV4Mini()) {
      if (!output && !input) return "z.ZodMiniType";
      return `z.ZodMiniType<${output}${input ? `, ${input}` : ""}>`;
    }
    if (isZodV4()) {
      if (!output && !input) return "z.ZodType";
      return `z.ZodType<${output}${input ? `, ${input}` : ""}>`;
    }
    if (!output && !input) return "z.ZodType";
    return `z.ZodType<${output}, z.ZodTypeDef, ${input || "unknown"}>`;
  },
  instanceof: (typeRef: string, newLine: boolean = false): string => {
    if (isZodV4()) {
      // In zodv4, instanceof produces a z.Custom where input type is unknown
      // this screws up with a lot of our z.ZodType<Output, Input> types
      // using z.custom explicitly gives us control over the input type
      // Caused by this change https://zod.dev/v4/changelog?id=zcoerce-updates
      return `z.custom<${typeRef}>(x => x instanceof ${stripGenerics(
        typeRef,
      )})`;
    }
    return `z${newLine ? "\n" : ""}.instanceof(${typeRef})`;
  },
  default: (zod: string, defaultValue: string): string => {
    if (defaultValue === "") {
      return zod;
    }

    if (isZodV4Mini()) {
      return `z._default(${zod}, ${defaultValue})`;
    }
    return zod + `.default(${defaultValue})`;
  },
  optional: (x: string): string => {
    // No-zod: passthrough schemas have no `.optional()` member; optionality
    // is captured at the TypeScript-type level (`| undefined`). The schema
    // itself doesn't need a wrapper at runtime.
    if (isNoZod()) {
      return x;
    }
    if (isZodV4Mini()) {
      return `z.optional(${x})`;
    }
    return x + `.optional()`;
  },
  literal: (val: string): string => {
    return `z.literal(${val})`;
  },
  refine: (x: string, validator: string, options?: string): string => {
    if (isZodV4Mini()) {
      return options
        ? `${x}.check(z.refine(${validator}, ${options}))`
        : `${x}.check(z.refine(${validator}))`;
    }
    return options
      ? `${x}.refine(${validator}, ${options})`
      : `${x}.refine(${validator})`;
  },
  array: (x: string): string => {
    if (isZodV4Mini()) {
      return `z.array(${x})`;
    }
    return `z.array(${x})`;
  },
  datetime: (options: string = ""): string => {
    if (isZodV4()) {
      return `z.iso.datetime(${options})`;
    }
    return `z.string().datetime(${options})`;
  },
  and: (x: string, y: string): string => {
    if (isZodV4Mini()) {
      return `z.intersection(${x}, ${y})`;
    }
    return `${x}.and(${y})`;
  },
  zodIssue: (): string => {
    if (isZodV4Mini()) {
      return "z.core.$ZodIssue";
    }
    return "z.ZodIssue";
  },
  transform: (x: string, fn: string, newLine: boolean = false): string => {
    if (fn === "") {
      return x;
    }

    if (isZodV4Mini()) {
      return `z.pipe(${x}, z.transform(${fn}))`;
    }
    return `${x}${newLine ? "\n" : ""}.transform(${fn})`;
  },
  or: (x: string, y: string, newLine: boolean = false): string => {
    if (isZodV4Mini()) {
      return `z.union([${x}, ${y}])`;
    }
    return `${x}${newLine ? "\n" : ""}.or(${y})`;
  },
  boolean: (): string => {
    return `z.boolean()`;
  },
  catchall: (x: string, schema: string, newLine: boolean = false): string => {
    if (isZodV4Mini()) {
      return `z.catchall(${x}, ${schema})`;
    }
    return `${x}${newLine ? "\n" : ""}.catchall(${schema})`;
  },
  string: (): string => {
    return `z.string()`;
  },
  coerce: {
    boolean: (): string => {
      return `z.coerce.boolean()`;
    },
    string: (): string => {
      return `z.coerce.string()`;
    },
    number: (): string => {
      return `z.coerce.number()`;
    },
    int: (): string => {
      return `z.coerce.int()`;
    },
    bigint: (): string => {
      return `z.coerce.bigint()`;
    },
    date: (options: string = ""): string => {
      return `z.coerce.date(${options})`;
    },
    datetime: (options: string = ""): string => {
      return `z.coerce.datetime(${options})`;
    },
  },
  union: (schemas: string[]): string => {
    return `z.union([${schemas.join(", ")}])`;
  },
  void: (): string => {
    return `z.void()`;
  },
  undefined: (): string => {
    return `z.undefined()`;
  },
  null: (): string => {
    return `z.null()`;
  },
  unknown: (): string => {
    return `z.unknown()`;
  },
  any: (): string => {
    return `z.any()`;
  },
  custom: (x: string, validator: string, errorMessage: string): string => {
    if (isZodV4Mini()) {
      return `z.custom<${x}>(${validator}, ${errorMessage})`;
    }
    return `z.custom<${x}>(${validator}, ${errorMessage})`;
  },
  pipe: (x: string, y: string, newLine: boolean = false): string => {
    if (isZodV4Mini()) {
      return `z.pipe(${x}, ${y})`;
    }
    return `${x}${newLine ? "\n  " : ""}.pipe(${y})`;
  },
  nullable: (x: string): string => {
    if (isZodV4Mini()) {
      return `z.nullable(${x})`;
    }
    return `z.nullable(${x})`;
  },
  parse: (schema: string, value: string): string => {
    if (isZodV4Mini()) {
      return `z.parse(${schema}, ${value})`;
    }
    return `${schema}.parse(${value})`;
  },
  safeParse: (schema: string, value: string): string => {
    if (isZodV4Mini()) {
      return `z.safeParse(${schema}, ${value})`;
    }
    return `${schema}.safeParse(${value})`;
  },
  int: (): string => {
    if (isZodV4()) {
      return `z.int()`;
    }
    return `z.number().int()`;
  },
  number: (): string => {
    return `z.number()`;
  },
  bigint: (): string => {
    return `z.bigint()`;
  },
  date: (): string => {
    return `z.date()`;
  },
  lazy: (fn: string): string => {
    return `z.lazy(${fn})`;
  },
  object: (fields: string, newLine: boolean = false): string => {
    return `${newLine ? "\n" : ""}z.object(${fields})`;
  },
  NEVER: (): string => {
    return `z.NEVER`;
  },
  never: (): string => {
    return `z.never()`;
  },
  addIssue: (params: {
    code: string;
    message?: string;
    input?: string;
    errors?: string;
    expected?: string;
    received?: string;
  }): string => {
    const parts: string[] = [];

    // Add input field for v4 and v4-mini
    if (params.input && isZodV4()) {
      parts.push(`input: ${params.input}`);
    }

    // Add code
    parts.push(`code: "${params.code}"`);

    // Add message if provided
    if (params.message) {
      parts.push(`message: ${params.message}`);
    }

    // Handle errors field
    if (params.errors) {
      if (!isZodV4() && params.code === "invalid_union") {
        parts.push(
          `unionErrors: ${params.errors}.map(issues => new z.ZodError(issues))`,
        );
      } else {
        parts.push(`errors: ${params.errors}`);
      }
    }

    if (params.expected) {
      parts.push(`expected: "${params.expected}"`);
    }

    if (params.received) {
      parts.push(`received: "${params.received}"`);
    }

    const issue = `{${parts.join(", ")}}`;

    if (isZodV4Mini()) {
      return `ctx.issues.push(${issue})`;
    }
    return `ctx.addIssue(${issue})`;
  },
  catch: (x: string, fn: string, newLine: boolean = false): string => {
    if (fn === "") {
      return x;
    }

    if (isZodV4Mini()) {
      return `z.catch(${x}, ${fn})`;
    }
    return `${x}${newLine ? "\n" : ""}.catch(${fn})`;
  },
};

registerTemplateFunc("zodEnumType", z.Enum);
registerTemplateFunc("zodEnum", z.enum);
registerTemplateFunc("zodRecord", z.record);
registerTemplateFunc("zodType", z.ZodType);
registerTemplateFunc("zodInstanceof", z.instanceof);
registerTemplateFunc("zodDefault", z.default);
registerTemplateFunc("zodOptional", z.optional);
registerTemplateFunc("zodLiteral", z.literal);
registerTemplateFunc("zodRefine", z.refine);
registerTemplateFunc("zodArray", z.array);
registerTemplateFunc("zodDatetime", z.datetime);
registerTemplateFunc("zodAnd", z.and);
registerTemplateFunc("zodTransform", z.transform);
registerTemplateFunc("zodOr", z.or);
registerTemplateFunc("zodBoolean", z.boolean);
registerTemplateFunc("zodCatchall", z.catchall);
registerTemplateFunc("zodString", z.string);
registerTemplateFunc("zodUnion", z.union);
registerTemplateFunc("zodUndefined", z.undefined);
registerTemplateFunc("zodNull", z.null);
registerTemplateFunc("zodUnknown", z.unknown);
registerTemplateFunc("zodAny", z.any);
registerTemplateFunc("zodPipe", z.pipe);
registerTemplateFunc("zodNullable", z.nullable);
registerTemplateFunc("zodParse", z.parse);
registerTemplateFunc("zodSafeParse", z.safeParse);
registerTemplateFunc("zodInt", z.int);
registerTemplateFunc("zodNumber", z.number);
registerTemplateFunc("zodBigint", z.bigint);
registerTemplateFunc("zodDate", z.date);
registerTemplateFunc("zodLazy", z.lazy);
registerTemplateFunc("zodObject", z.object);
registerTemplateFunc("zodNEVER", z.NEVER);
registerTemplateFunc("zodNever", z.never);
registerTemplateFunc("zodAddIssue", z.addIssue);

function stripGenerics(typeRef: string): string {
  return typeRef.replace(/<[^>]*>/g, "");
}

function useSmartUnion(typeDef?: TypeDef): boolean {
  // The "populated-fields" union strategy scores union candidates by running
  // them through their zod schemas and picking the best match. Without zod
  // there's no scorer, so the unionStrategy setting is inert and we fall back
  // to declaration-order resolution at the TypeScript-type level.
  if (isNoZod()) return false;
  if (!isUnionStrategyPopulatedFields()) return false;
  if (!typeDef) return true;
  const discriminated = !!typeDef.Discriminator?.TypePropertyName;
  if (discriminated) return false;
  return true;
}
registerTemplateFunc("useSmartUnion", useSmartUnion);
