/* Open Discriminated Union helpers for forward-compatible union parsing */

// @ts-ignore
function templateOpenDiscriminatedUnion(
  typeDef: TypeDef,
  discriminatorPropertyName: string,
  unionZod: string[],
  discriminatorMappings: DiscriminatorMapping[],
  usageLocation: string,
): string {
  // No-zod: no `.zod` expression is read by any caller, so emit nothing.
  if (isNoZod()) {
    return "";
  }

  // Build the options object mapping discriminator values to schemas
  const optionsEntries = discriminatorMappings
    .map((mapping, idx) => {
      const keyName = mapping.Name;
      // Use simple property syntax for valid identifiers, computed syntax otherwise
      const isValidIdentifier = /^[a-zA-Z_$][a-zA-Z0-9_$]*$/.test(keyName);
      const key = isValidIdentifier ? keyName : JSON.stringify(keyName);
      return isValidIdentifier
        ? `${key}: ${unionZod[idx]}`
        : `[${key}]: ${unionZod[idx]}`;
    })
    .join(",");

  const optionsObject = `{${optionsEntries}}`;

  addTypeImport(
    getDiscriminatedUnionFileName(),
    "discriminatedUnion",
    usageLocation,
  );

  const discrKey = JSON.stringify(discriminatorPropertyName);
  const tsPropertyName = sanitizeFieldName(discriminatorPropertyName);
  const unknownValue = findUniqueUnknownValue(typeDef, discriminatorMappings);

  // Build the options object for the 3rd argument (only include non-default values)
  const needsOutputPropertyName = tsPropertyName !== discriminatorPropertyName;
  const needsUnknownValue =
    unknownValue !== getConstants().defaultUnknownDiscriminatorValue;

  if (needsOutputPropertyName || needsUnknownValue) {
    const optsParts: string[] = [];
    if (needsUnknownValue) {
      optsParts.push(`unknownValue: ${JSON.stringify(unknownValue)}`);
    }
    if (needsOutputPropertyName) {
      optsParts.push(`outputPropertyName: ${JSON.stringify(tsPropertyName)}`);
    }
    return `discriminatedUnion(${discrKey}, ${optionsObject}, { ${optsParts.join(
      ", ",
    )} })`;
  }
  return `discriminatedUnion(${discrKey}, ${optionsObject})`;
}

registerTemplateFunc(
  "templateOpenDiscriminatedUnion",
  templateOpenDiscriminatedUnion,
);

function templateOpenDiscriminatedUnionFunc(): string {
  // No-zod: the discriminatedUnion helper file is not emitted (gated in
  // {{getDiscriminatedUnionFileName}}.ts.stmpl); no caller references it.
  const typesLocation = getTypesLocation();
  const defaultValue = getConstants().defaultUnknownDiscriminatorValue;

  addTypeImport(
    getUnrecognizedFileName(),
    "startCountingUnrecognized",
    typesLocation,
  );
  if (isLaxMode()) {
    addTypeImport(
      getDefaultToZeroValueFileName(),
      "startCountingDefaultToZeroValue",
      typesLocation,
    );
  }

  return `
    const UNKNOWN = Symbol("${defaultValue}");

    export type Unknown<Discriminator extends string, UnknownValue = "${defaultValue}"> = {
      [K in Discriminator]: UnknownValue;
    } & {
      raw: unknown;
      isUnknown: true;
    };

    export function isUnknown<Discriminator extends string>(value: unknown): value is Unknown<Discriminator> {
      return typeof value === "object" && value !== null && UNKNOWN in value;
    }

    /**
     * Forward-compatible discriminated union parser.
     *
     * If the input does not match one of the predefined options, it will be
     * captured and available as \`{ raw: <original input>, [discriminator]: "${defaultValue}", isUnknown: true }\`.
     *
     * @param inputPropertyName - The discriminator property name in the input payload
     * @param options - Map of discriminator values to their corresponding schemas
     * @param opts - Optional configuration object
     * @param opts.unknownValue - The value to use for the discriminator when the input is unknown (default: "${defaultValue}")
     * @param opts.outputPropertyName - Output property name if the sanitized (camelCase) property name differs from inputPropertyName
     */
    export function discriminatedUnion<
      InputDiscriminator extends string,
      TOptions extends Readonly<Record<string, ${z.ZodType()}>>,
      UnknownValue extends string = "${defaultValue}",
      OutputDiscriminator extends string = InputDiscriminator,
    >(
      inputPropertyName: InputDiscriminator,
      options: TOptions,
      opts: { unknownValue?: UnknownValue; outputPropertyName?: OutputDiscriminator } = {},
    ): ${z.ZodType(
      "z.output<TOptions[keyof TOptions]> | Unknown<OutputDiscriminator, UnknownValue>",
      "unknown",
    )} {
      const { unknownValue = "${defaultValue}" as UnknownValue, outputPropertyName } = opts;
      return ${z.transform(
        z.unknown(),
        `(input) => {
          const fallback = Object.defineProperties(
            { raw: input, [outputPropertyName ?? inputPropertyName]: unknownValue, isUnknown: true as const },
            { [UNKNOWN]: { value: true, enumerable: false, configurable: false } },
          );

          const isObject = typeof input === "object" && input !== null;
          if (!isObject) return fallback;

          const discriminator = input[inputPropertyName as keyof typeof input];
          if (typeof discriminator !== "string") return fallback;
          if (!(discriminator in options)) return fallback;

          const schema = options[discriminator];
          if (!schema) return fallback;

          // Start counters before parsing to track nested unrecognized/zeroDefault values
          const unrecognizedCtr = startCountingUnrecognized();
          ${
            isLaxMode()
              ? `const zeroDefaultCtr = startCountingDefaultToZeroValue();`
              : ""
          }

          const result = ${z.safeParse("schema", "input")};
          if (!result.success) {
            // Parse failed - don't propagate any counts from the failed attempt
            unrecognizedCtr.end(0);
            ${isLaxMode() ? `zeroDefaultCtr.end(0);` : ""}
            return fallback;
          }

          // Parse succeeded - propagate the actual counts
          unrecognizedCtr.end();
          ${isLaxMode() ? `zeroDefaultCtr.end();` : ""}

          if (outputPropertyName) {
            (result.data as any)[outputPropertyName] = discriminator;
          }

          return result.data;
        }`,
      )} as any;
    }`;
}

registerTemplateFunc(
  "templateOpenDiscriminatedUnionFunc",
  templateOpenDiscriminatedUnionFunc,
);

function shouldTemplateOpenDiscriminatedUnion(): boolean {
  const allTypes = getAllTypes();
  for (const t of allTypes) {
    if (t.IsUnionOpen) {
      return true;
    }
  }
  return false;
}

registerTemplateFunc(
  "shouldTemplateOpenDiscriminatedUnion",
  shouldTemplateOpenDiscriminatedUnion,
);
