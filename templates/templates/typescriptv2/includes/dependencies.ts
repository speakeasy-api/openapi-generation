type ExportConfig = {
  source: string;
  types: string;
  default: string;
};

// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

function templatePeerDependencies(): string {
  const defaultDependencies = {};

  if (isReactQueryEnabled()) {
    defaultDependencies["@tanstack/react-query"] =
      deps["react-query-peer"].version;
    defaultDependencies["react"] = deps.react.version;
    defaultDependencies["react-dom"] = deps["react-dom"].version;
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.peerDependencies,
  );
}
registerTemplateFunc("templatePeerDependencies", templatePeerDependencies);

function templatePeerDependenciesMeta(): string {
  const defaultDependencies = {};

  if (isReactQueryEnabled()) {
    defaultDependencies["@tanstack/react-query"] = { optional: true };
    defaultDependencies["react"] = { optional: true };
    defaultDependencies["react-dom"] = { optional: true };
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.peerDependenciesMeta,
  );
}
registerTemplateFunc(
  "templatePeerDependenciesMeta",
  templatePeerDependenciesMeta,
);

function templateDevDependencies(): string {
  const defaultDependencies = {
    typescript: deps.typescript.version,
  };

  if (context.Global.Config.UseOxlint) {
    defaultDependencies["oxlint"] = deps.oxlint.version;
  } else {
    defaultDependencies["@eslint/js"] = deps["@eslint/js"].version;
    defaultDependencies["eslint"] = deps.eslint.version;
    defaultDependencies["globals"] = deps.globals.version;
    defaultDependencies["typescript-eslint"] =
      deps["typescript-eslint"].version;
  }

  if (context.Global.Config.UseTsgo) {
    defaultDependencies["@typescript/native-preview"] =
      deps["@typescript/native-preview"].version;
  }

  if (isReactQueryEnabled()) {
    defaultDependencies["@tanstack/react-query"] =
      deps["@tanstack/react-query"].version;
    defaultDependencies["@types/react"] = deps["@types/react"].version;
  }

  if (isMCPServerEnabled()) {
    defaultDependencies["@stricli/core"] = deps["@stricli/core"].version;
    defaultDependencies["@types/express"] = deps["@types/express"].version;
    defaultDependencies["bun"] = deps.bun.version;
    defaultDependencies["bun-types"] = deps["bun-types"].version;
    defaultDependencies["express"] = deps.express.version;
  }

  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    defaultDependencies["vitest"] = deps.vitest.version;
    defaultDependencies["@types/node"] = deps["@types/node"].version;
  }

  if (isFeatureUsed("pagination") && usesJSONPath(context.Global.AST.MainSDK)) {
    const lib = context.Global.Config.Jsonpath || "legacy";
    if (lib === "legacy") {
      defaultDependencies["@types/jsonpath"] = deps["@types/jsonpath"].version;
    }
  }

  if (context.Global.Config.ModuleFormat === "dual") {
    defaultDependencies["tshy"] = deps.tshy.version;
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.devDependencies,
  );
}
registerTemplateFunc("templateDevDependencies", templateDevDependencies);

// @ts-ignore
function templateDependencies(): string {
  const defaultDependencies: Record<string, string> = {};

  // In no-zod mode the generated SDK relies only on JSON.parse / JSON.stringify
  // for (de)serialization. Omit the zod runtime dependency entirely.
  if (!isNoZod()) {
    defaultDependencies["zod"] = deps.zod.version;
  }

  if (isMCPServerEnabled()) {
    defaultDependencies["@modelcontextprotocol/sdk"] =
      deps["@modelcontextprotocol/sdk"].version;
  }

  if (isZodV4Mini()) {
    // z.catchall() was introduced in zod v3.25.65 to zod mini
    defaultDependencies["zod"] = deps.zodV4Mini.version;
  }
  // No-zod doesn't construct Decimal instances at runtime, so the decimal.js
  // runtime dependency isn't required even when the spec uses x-decimal types.
  if (isFeatureUsed("decimal") && !isNoZod()) {
    defaultDependencies["decimal.js"] = deps["decimal.js"].version;
  }
  if (isFeatureUsed("pagination") && usesJSONPath(context.Global.AST.MainSDK)) {
    const lib = context.Global.Config.Jsonpath;
    if (lib === "rfc9535") {
      defaultDependencies["jsonpath-rfc9535"] =
        deps["jsonpath-rfc9535"].version;
    } else {
      defaultDependencies["jsonpath"] = deps.jsonpath.version;
    }
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.dependencies,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

// @ts-ignore
function renderDependencies(
  defaultDependencies: Record<string, unknown>,
  additionalDependencies: Record<string, unknown>,
): string {
  const dependencies = {
    ...defaultDependencies,
    ...additionalDependencies,
  };

  let renderedDependencies = "";

  for (const dependency of Object.keys(dependencies).sort()) {
    const value = JSON.stringify(dependencies[dependency]);
    renderedDependencies += `"${dependency}": ${value},\n`;
  }
  renderedDependencies = renderedDependencies.replace(/,\n$/, "");

  return renderedDependencies;
}

function templateAdditionalPackageJSON(): string {
  const inp = context.Global.Config.AdditionalPackageJSON || {};
  if (typeof inp !== "object") {
    throw new Error("additionalPackageJSON must be an object");
  }

  let values = JSON.parse(JSON.stringify(inp));

  if (context.Global.Config.GeneratedLicense === "agpl") {
    delete values.license;
  }

  // sort values keys alphabetically to ensure consistent ordering as AdditionalPackageJSON is backed by a go map and so order changes
  values = Object.fromEntries(
    Object.entries(values).sort((a, b) => a[0].localeCompare(b[0])),
  );

  const moduleFormat = context.Global.Config.ModuleFormat;
  if (moduleFormat === "esm" || moduleFormat === "dual") {
    values.type = "module";
  }

  if (isMCPServerEnabled()) {
    values.bin = values.bin || {};
    values.bin["mcp"] = "bin/mcp-server.js";
  }

  const userDefinedExports =
    "exports" in values &&
    typeof values.exports === "object" &&
    values.exports != null
      ? values.exports
      : {};
  if (Object.keys(userDefinedExports).length) {
    console.warn(
      "🟡 package.json exports are managed by the generator. Ignoring user-defined exports in gen.yaml.",
    );
  }

  if (moduleFormat === "dual") {
    values.tshy = {
      ...(context.Global.Config.UseTsgo ? { compiler: "tsgo" } : {}),
      sourceDialects: [`${context.Global.Config.PackageName}/source`],
      exports: {
        ".": "./src/index.ts",
        "./package.json": "./package.json",
        ...toTshy(makeBarrelExports()),
        "./*.js": "./src/*.ts",
        "./*": "./src/*.ts",
      },
    };
    delete values.exports;
  } else if (moduleFormat === "esm") {
    values.main = "./esm/index.js";
    values.exports = {
      ".": {
        source: "./src/index.ts",
        types: "./esm/index.d.ts",
        default: "./esm/index.js",
      },
      "./package.json": "./package.json",
      ...makeBarrelExports(),
      "./*.js": {
        source: "./src/*.ts",
        types: "./esm/*.d.ts",
        default: "./esm/*.js",
      },
      "./*": {
        source: "./src/*.ts",
        types: "./esm/*.d.ts",
        default: "./esm/*.js",
      },
    };
  } else {
    values.main = "./index.js";
  }

  const serialized = JSON.stringify(values, null, 2).slice(1, -1).trimEnd();

  return serialized ? `${serialized},` : "";
}
registerTemplateFunc(
  "templateAdditionalPackageJSON",
  templateAdditionalPackageJSON,
);

function makeBarrelExports(): Record<string, ExportConfig> {
  if (!context.Global.Config.UseIndexModules) {
    // The public export surface is a single module independent of index
    // barrels, so its mapping applies in either mode.
    return makeResourcesExport({});
  }

  // We need to find out which scopes were used because we should only add
  // barrel exports for them in package.json. If we add exports to unused scopes
  // then certain linters will fail us such as Are the types wrong?
  // (https://arethetypeswrong.github.io/)
  const scopes = new Set<string>();
  for (const [, bucket] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, types] of sequencedMapEntries(bucket)) {
      for (const type of types) {
        scopes.add(type.Scope.toString());
      }
    }
  }

  const paths = new Set([
    sanitizePath(getTypesLocation()),
    // The errors path will always exist because we add the built-in
    // DefaultError and SDKValidationError classes there.
    sanitizePath(getErrorsLocation()),
    scopes.has("shared") ? sanitizePath(getSharedLocation()) : "",
    scopes.has("operations") ? sanitizePath(getOperationsLocation()) : "",
    scopes.has("webhooks") ? sanitizePath(getWebhooksLocation()) : "",
    scopes.has("callbacks") ? sanitizePath(getCallbacksLocation()) : "",
  ]);

  if (isReactQueryEnabled()) {
    paths.add("react-query");
  }

  const out: Record<string, ExportConfig> = {};
  for (const item of paths) {
    if (item === "") {
      continue;
    }

    out[`./${item}`] = {
      source: `./src/${item}/index.ts`,
      types: `./esm/${item}/index.d.ts`,
      default: `./esm/${item}/index.js`,
    };
  }

  return makeResourcesExport(out);
}

// The public export surface is a single module rather than a directory
// index.
function makeResourcesExport(
  out: Record<string, ExportConfig>,
): Record<string, ExportConfig> {
  if (hasPublicExports()) {
    const resourcesModule = tsResourcesModulePath();
    out[`./${resourcesModule}`] = {
      source: `./src/${resourcesModule}.ts`,
      types: `./esm/${resourcesModule}.d.ts`,
      default: `./esm/${resourcesModule}.js`,
    };
  }
  return out;
}

function toTshy(
  exports: Record<string, { source: string }>,
): Record<string, string> {
  const out: Record<string, string> = {};

  for (const key of Object.keys(exports)) {
    out[key] = exports[key].source;
  }

  return out;
}

function sanitizePath(path: string) {
  return path
    .split("/")
    .filter((s) => s && s !== ".")
    .join("/");
}

function usesJSONPath(sdk: SDK): boolean {
  for (const op of sdk.Operations) {
    if (
      op.Extensions.Pagination &&
      !op.Extensions.Pagination.Outputs.CanUseDotNotation
    ) {
      return true;
    }
  }

  for (const sub of sdk.SubSDKs) {
    if (usesJSONPath(sub)) {
      return true;
    }
  }

  return false;
}

function usesDlv(sdk: SDK): boolean {
  for (const op of sdk.Operations) {
    if (
      op.Extensions.Pagination &&
      op.Extensions.Pagination.Outputs.CanUseDotNotation
    ) {
      return true;
    }
  }

  for (const sub of sdk.SubSDKs) {
    if (usesDlv(sub)) {
      return true;
    }
  }

  return false;
}
registerTemplateFunc("usesDlv", usesDlv);
