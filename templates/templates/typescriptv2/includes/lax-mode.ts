// @ts-ignore
function isLaxMode(): boolean {
  // Lax-mode coercions are implemented as zod transforms. Without zod there
  // is nothing to coerce — the user's gen.yaml `laxMode` setting is inert.
  if (isNoZod()) {
    return false;
  }
  const laxMode = context.Global.Config.LaxMode;
  return laxMode === "lax" || laxMode === true;
}
registerTemplateFunc("isLaxMode", isLaxMode);

// LaxMode namespace for generating type-safe primitives
namespace LaxMode {
  function importName(usageLocation: string): string {
    let aliasName = "types";
    if ("funcs" === usageLocation) aliasName = "types$";
    return aliasName;
  }

  function types(usageLocation: string) {
    const t = importName(usageLocation);
    return {
      literal: (literal: string) => `${t}.literal(${literal})`,
      literalBigInt: (literal: string) => `${t}.literalBigInt(${literal})`,
      string: () => `${t}.string()`,
      boolean: () => `${t}.boolean()`,
      number: () => `${t}.number()`,
      date: () => `${t}.date()`,
      bigint: () => `${t}.bigint()`,
      nullable: (zod: string) => `${t}.nullable(${zod})`,
      optional: (zod: string) => `${t}.optional(${zod})`,
    };
  }

  function addSmartPrimitivesImport(usageLocation: string) {
    return addTypeImport(
      "primitives",
      importName(usageLocation),
      usageLocation,
      "aliasImport",
    );
  }

  function base(usageLocation: string) {
    return {
      importOutputTypes: () => "",
      importZodTypes: () => addSmartPrimitivesImport(usageLocation),
    };
  }

  export function literal(opts: {
    const: string;
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: opts.const,
      zod: z.default(
        types(opts.usageLocation).literal(opts.const),
        opts.default,
      ),
    };
  }

  export function literalBigInt(opts: {
    const: string;
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: opts.const,
      zod: z.default(
        types(opts.usageLocation).literalBigInt(opts.const),
        opts.default,
      ),
    };
  }

  export function string(opts: {
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: "string",
      zod: z.default(types(opts.usageLocation).string(), opts.default),
    };
  }

  export function boolean(opts: {
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: "boolean",
      zod: z.default(types(opts.usageLocation).boolean(), opts.default),
    };
  }

  export function number(opts: {
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: "number",
      zod: z.default(types(opts.usageLocation).number(), opts.default),
    };
  }

  export function date(opts: {
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: "Date",
      zod: z.default(types(opts.usageLocation).date(), opts.default),
    };
  }

  export function dateOutbound(opts: {
    default: string;
    usageLocation: string;
  }): ResolvedTypes {
    const zod = z.transform(
      z.default(z.date(), opts.default),
      `v => v.toISOString().slice(0, "YYYY-MM-DD".length)`,
    );
    return {
      inputType: "Date",
      importInputTypes: () => "",
      outputType: "string",
      importOutputTypes: () => "",
      zod,
      importZodTypes: () => addZodPackageImport(),
    };
  }

  export function bigint(opts: {
    default: string;
    usageLocation: string;
  }): Omit<ResolvedTypes, "inputType" | "importInputTypes"> {
    return {
      ...base(opts.usageLocation),
      outputType: "bigint",
      zod: z.default(types(opts.usageLocation).bigint(), opts.default),
    };
  }

  export function nullable(opts: {
    result: Omit<ResolvedTypes, "inputType" | "importInputTypes">;
    usageLocation: string;
  }): void {
    opts.result.zod = types(opts.usageLocation).nullable(opts.result.zod);
    const originalImportZodTypes = opts.result.importZodTypes;
    opts.result.importZodTypes = () => {
      addSmartPrimitivesImport(opts.usageLocation);
      return originalImportZodTypes();
    };
  }

  export function optional(opts: {
    result: Omit<ResolvedTypes, "inputType" | "importInputTypes">;
    usageLocation: string;
  }): void {
    const originalImportZodTypes = opts.result.importZodTypes;
    opts.result.zod = types(opts.usageLocation).optional(opts.result.zod);
    opts.result.importZodTypes = () => {
      addSmartPrimitivesImport(opts.usageLocation);
      return originalImportZodTypes();
    };
  }
}
