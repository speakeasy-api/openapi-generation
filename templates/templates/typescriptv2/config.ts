// @ts-ignore
function getCompileDependencies(
  _config: GeneratorInitialConfiguration,
): RunnerCommandDependencies {
  return [
    {
      command: "node",
      version: {
        args: ["--version"],
        regex: `(?m).*?(\\d+\\.\\d+\\.\\d+).*?`,
        minVersion: "18.0.0",
      },
      installDocumentation: `Install Node.js by following the instructions at https://nodejs.org/en/download/.`,
    },
  ];
}

/** Returns the install commands needed to set up node_modules.
 *  Used by both getCompileCommands and getTestingConfiguration. */
function getInstallCommands(langConfig: any): RunnerCommands {
  if (langConfig.compileCommand != null) {
    if (!Array.isArray(langConfig.compileCommand)) {
      throw new Error(
        "compileCommand must be an array of the form [COMMAND, ...ARGS]",
      );
    }

    const [command, ...args] = langConfig.compileCommand;
    if (typeof command !== "string") {
      throw new Error("compileCommand[0] must be a string");
    }

    return [{ command, args }];
  }

  return [
    {
      command: "npm",
      args: [
        "install",
        "--ignore-scripts",
        langConfig.enableMCPServer ? "--prefer-dedupe" : "",
      ],
    },
    {
      command: "npm",
      args: ["rebuild", "bun", "esbuild"],
    },
  ];
}

// @ts-ignore
function getCompileCommands(
  config: GeneratorInitialConfiguration,
): RunnerCommands {
  const langConfig = config.LangCfg;
  const commands = [...getInstallCommands(langConfig)];

  // Custom compileCommand handles everything — no additional steps needed.
  if (langConfig.compileCommand != null) {
    return commands;
  }

  commands.push(
    {
      command: "npm",
      args: ["run", "lint"],
      environmentVariables: {
        NODE_OPTIONS: "--max-old-space-size=6144",
      },
    },
    {
      command: "npm",
      args: ["run", "build"],
      environmentVariables: {
        NODE_OPTIONS: "--max-old-space-size=6144",
      },
    },
  );

  // Only compile examples if generateExamples is enabled
  if (langConfig.generateExamples !== false) {
    commands.push({
      command: "npm",
      args: ["--prefix", "examples", "install", "--ignore-scripts"],
    });
  }

  return commands;
}

/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    zod: {
      name: "zod",
      version: "^3.25.0 || ^4.0.0",
      cpe: "cpe:2.3:a:colinhacks:zod:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    zodV4Mini: {
      name: "zod",
      version: "^3.25.65 || ^4.0.0",
      cpe: "cpe:2.3:a:colinhacks:zod:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "zodV4Mini",
    },
    "@modelcontextprotocol/sdk": {
      name: "@modelcontextprotocol/sdk",
      version: "^1.26.0",
      cpe: "cpe:2.3:a:anthropic:model_context_protocol_sdk:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "mcpServer",
    },
    "decimal.js": {
      name: "decimal.js",
      version: "^10.4.3",
      cpe: "cpe:2.3:a:mikemcl:decimal.js:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "decimal",
    },
    "jsonpath-rfc9535": {
      name: "jsonpath-rfc9535",
      version: "1.1.0",
      cpe: "cpe:2.3:a:jsonpath-rfc9535:jsonpath-rfc9535:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "paginationRfc9535",
    },
    jsonpath: {
      name: "jsonpath",
      version: "^1.2.1",
      cpe: "cpe:2.3:a:dchester:jsonpath:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "paginationLegacy",
    },
    // Dev dependencies
    oxlint: {
      name: "oxlint",
      version: "^1.60.0",
      cpe: "cpe:2.3:a:oxc-project:oxlint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "oxlint",
    },
    "@eslint/js": {
      name: "@eslint/js",
      version: "^9.26.0",
      cpe: "cpe:2.3:a:eslint:eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "eslint",
    },
    eslint: {
      name: "eslint",
      version: "^9.26.0",
      cpe: "cpe:2.3:a:eslint:eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "eslint",
    },
    globals: {
      name: "globals",
      version: "^15.14.0",
      cpe: "cpe:2.3:a:sindresorhus:globals:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "eslint",
    },
    "typescript-eslint": {
      name: "typescript-eslint",
      version: "^8.26.0",
      cpe: "cpe:2.3:a:typescript-eslint:typescript-eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "eslint",
    },
    typescript: {
      name: "typescript",
      version: "~5.8.3",
      cpe: "cpe:2.3:a:microsoft:typescript:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@typescript/native-preview": {
      name: "@typescript/native-preview",
      version: "7.0.0-dev.20260302.1",
      cpe: "cpe:2.3:a:microsoft:typescript:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "tsgo",
    },
    "@tanstack/react-query": {
      name: "@tanstack/react-query",
      version: "^5.61.4",
      cpe: "cpe:2.3:a:tanstack:react-query:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "reactQuery",
    },
    "@types/react": {
      name: "@types/react",
      version: "^18.3.12",
      cpe: "cpe:2.3:a:definitelytyped:types-react:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "reactQuery",
    },
    "@stricli/core": {
      name: "@stricli/core",
      version: "^1.1.1",
      cpe: "cpe:2.3:a:stricli:core:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "mcpServer",
    },
    "@types/express": {
      name: "@types/express",
      version: "^4.17.21",
      cpe: "cpe:2.3:a:definitelytyped:types-express:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "mcpServer",
    },
    bun: {
      name: "bun",
      version: "1.2.17",
      cpe: "cpe:2.3:a:oven:bun:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "mcpServer",
    },
    "bun-types": {
      name: "bun-types",
      version: "1.2.17",
      cpe: "cpe:2.3:a:oven:bun-types:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "mcpServer",
    },
    express: {
      name: "express",
      version: "^4.21.2",
      cpe: "cpe:2.3:a:expressjs:express:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "mcpServer",
    },
    vitest: {
      name: "vitest",
      version: "^3.2.6",
      cpe: "cpe:2.3:a:vitest:vitest:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "tests",
    },
    "@types/node": {
      name: "@types/node",
      version: "^18.19.3",
      cpe: "cpe:2.3:a:definitelytyped:types-node:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "tests",
    },
    "@types/jsonpath": {
      name: "@types/jsonpath",
      version: "^0.2.4",
      cpe: "cpe:2.3:a:definitelytyped:types-jsonpath:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "paginationLegacy",
    },
    tshy: {
      name: "tshy",
      version: "^3.3.2",
      cpe: "cpe:2.3:a:isaacs:tshy:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "moduleFormatDual",
    },
    // Peer dependencies
    "react-query-peer": {
      name: "@tanstack/react-query",
      version: "^5",
      cpe: "cpe:2.3:a:tanstack:react-query:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "peer",
      condition: "reactQuery",
    },
    react: {
      name: "react",
      version: "^18 || ^19",
      cpe: "cpe:2.3:a:facebook:react:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "peer",
      condition: "reactQuery",
    },
    "react-dom": {
      name: "react-dom",
      version: "^18 || ^19",
      cpe: "cpe:2.3:a:facebook:react-dom:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "peer",
      condition: "reactQuery",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  if (oldVersion == "") {
    let upgraded = defaults;

    if (cfg.version) {
      upgraded.version = cfg.version;
    }

    if (cfg.packagename) {
      upgraded.packageName = cfg.packagename;
    }

    if (cfg.author) {
      upgraded.author = cfg.author;
    }

    return upgraded;
  }

  return cfg;
}

/** Returns the configuration fields available for customers. This data is
 * fetched early in the generation process so values can be passed back to
 * getGeneratorConfig. */
// @ts-ignore
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields {
  return {
    ...commonFields,
    version: {
      Name: "version",
      Required: true,
      DefaultValue: "0.0.1",
      Description: "The current version of the SDK",
      ValidationRegex: /^[\w\d._-]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    additionalPackageJSON: {
      Name: "additionalPackageJSON",
      Required: false,
      DefaultValue: {},
      Description: "additional key/values for package.json",
    },
    templateVersion: {
      Name: "templateVersion",
      Required: false,
      DefaultValue: "v2",
      Description: "The template version to use",
      ValidationRegex: /^v\d+$/.source,
      ValidationMessage:
        "Template version must be in the format v1, v2, and so on",
    },
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "openapi",
      Description:
        "The npm package name. https://docs.npmjs.com/package-name-guidelines.",
      ValidationRegex: /^[\w\d._@\/-]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    author: {
      Name: "author",
      Required: true,
      DefaultValue: "Speakeasy",
      Description:
        "The name of the author of the published package. https://docs.npmjs.com/cli/v9/configuring-npm/package-json#people-fields-author-contributors",
    },
    maxMethodParams: {
      Name: "maxMethodParams",
      Required: false,
      DefaultValue: 0,
      Description:
        "The maximum number of parameters a method can have before the resulting SDK endpoint is no longer 'flattened' and an input object is created instead. 0 will use input objects always. https://www.speakeasy.com/docs/customize-sdks/methods",
      ValidationRegex: /^\d+$/.source,
      ValidationMessage: "Numbers only",
    },
    flatteningOrder: {
      Name: "flatteningOrder",
      Required: false,
      DefaultValue: "parameters-first",
      Description:
        "When flattening parameters and body fields, determines the ordering of generated method arguments.",
      ValidationRegex: /^(parameters-first|body-first)$/.source,
      ValidationMessage: "parameters-first or body-first only",
    },
    clientServerStatusCodesAsErrors: {
      Name: "clientServerStatusCodesAsErrors",
      Required: false,
      DefaultValue: newSDK,
      Description: "Whether to treat 4xx and 5xx status codes as errors.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    flattenGlobalSecurity: {
      Name: "flattenGlobalSecurity",
      Required: false,
      DefaultValue: true,
      Description:
        "Flatten the global security configuration if there is only a single option in the spec",
    },
    imports: {
      Name: "imports",
      Required: false,
      DefaultValue: {
        option: "openapi",
        paths: {
          shared: newSDK ? "models" : "sdk/models/shared",
          operations: newSDK ? "models/operations" : "sdk/models/operations",
          errors: newSDK ? "models/errors" : "sdk/models/errors",
          callbacks: newSDK ? "models/callbacks" : "sdk/models/callbacks",
          webhooks: newSDK ? "models/webhooks" : "sdk/models/webhooks",
        },
      },
      Description: "Configuration for model import structure",
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {
        dependencies: {},
        devDependencies: {},
        peerDependencies: {},
      },
      Description:
        "Specify additional dependencies to include in the generated package.json",
    },
    additionalScripts: {
      Name: "additionalScripts",
      Required: false,
      DefaultValue: {},
      Description:
        "Specify additional scripts to include in the generated package.json",
    },
    responseFormat: {
      Name: "responseFormat",
      Required: false,
      DefaultValue: "flat",
      Description:
        "Determines the shape of the response envelope that is returned from SDK methods",
      ValidationRegex: /^(envelope|envelope-http|flat)$/.source,
      ValidationMessage: '"envelope-http", "envelope" or "flat" only',
    },
    eventStreamClassName: {
      Name: "eventStreamClassName",
      Required: false,
      DefaultValue: "EventStream",
      Description:
        "Controls the generated TypeScript server-sent event stream wrapper class name.",
      ValidationRegex: /^[A-Za-z_$][A-Za-z0-9_$]*$/.source,
      ValidationMessage: "Must be a valid TypeScript identifier",
    },
    enumFormat: {
      Name: "enumFormat",
      Required: false,
      DefaultValue: "union",
      Description: "Determines the format to express enums in TypeScript",
      ValidationRegex: /^(union|enum)$/.source,
      ValidationMessage: '"union" or "enum" only',
    },
    methodSignature: {
      Name: "methodSignature",
      Required: false,
      Description:
        "Controls the generated SDK class method shape. 'positional' (default) flattens arguments positionally; 'params-object' keeps required path parameters positional and gathers all remaining arguments into a generated <Operation>Params object, with streaming overload variants for SSE operations.",
      ValidationRegex: /^(positional|params-object)$/.source,
      ValidationMessage: '"positional" or "params-object" only',
    },
    methodArguments: {
      Name: "methodArguments",
      Required: false,
      DefaultValue: "infer-optional-args",
      Description:
        "If infer-optional-args, if all the parameters and the request body is optional, the argument to the function will be optional",
      ValidationRegex: /^(infer-optional-args|require-security-and-request)$/
        .source,
      ValidationMessage:
        '"infer-optional-args" or "require-security-and-request" only',
    },
    useIndexModules: {
      Name: "useIndexModules",
      Required: false,
      DefaultValue: true,
      Description:
        "Determine whether or not index modules (index.ts) are generated",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    moduleFormat: {
      Name: "moduleFormat",
      Required: false,
      DefaultValue: newSDK ? "esm" : "dual",
      Description: "Specifies the module format to use when compiling the SDK.",
      ValidationRegex: /^(commonjs|esm|dual)$/.source,
      ValidationMessage: '"commonjs", "esm" or "dual" only',
    },
    envVarPrefix: {
      Name: "envVarPrefix",
      Required: false,
      Description:
        "The environment variable prefix for security and global env variable overrides. If empty these overrides will not be possible",
    },
    compileCommand: {
      Name: "compileCommand",
      Required: false,
      Description:
        "The command to use for compiling the SDK. This must be an array where the first element is the command and the rest are arguments.",
    },
    defaultErrorName: {
      Name: "defaultErrorName",
      Required: false,
      // The default is handled in `fixBuiltInErrorNameConflicts()`
      DefaultValue: newSDK ? "" : "SDKError",
      Description:
        "The name of the fallback error class if no more specific error class is matched",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage:
        "Must start with a capital letter and contain only letters and numbers",
    },
    baseErrorName: {
      Name: "baseErrorName",
      Required: false,
      // The default is handled in `fixBuiltInErrorNameConflicts()`
      DefaultValue: "",
      Description:
        "The name of the base error class used for HTTP error responses",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage:
        "Must start with a capital letter and contain only letters and numbers",
    },
    enableReactQuery: {
      Name: "enableReactQuery",
      Required: false,
      DefaultValue: false,
      Description: "Generate React hooks using TanStack Query.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableMCPServer: {
      Name: "enableMCPServer",
      Required: false,
      DefaultValue: false,
      Description: "Generate a Model Context Protocol server.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableCustomCodeRegions: {
      Name: "enableCustomCodeRegions",
      Required: false,
      DefaultValue: false,
      Description: "Allow custom code to be inserted into the generated SDK.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    jsonpath: {
      Name: "jsonpath",
      Required: false,
      DefaultValue: newSDK ? "rfc9535" : "legacy",
      Description:
        'Sets the JSONPath implementation to use. "legacy" (deprecated) precedes the introduction of the "rfc9535" specification. The latter option should be preferred.',
      ValidationRegex: /^(legacy|rfc9535)$/.source,
      ValidationMessage: "legacy or rfc9535 only",
    },
    constFieldsAlwaysOptional: {
      Name: "constFieldsAlwaysOptional",
      Required: false,
      DefaultValue: !newSDK,
      Description:
        "Whether const fields should be treated as optional in TypeScript types and schemas regardless of OpenAPI spec requirements. When true (legacy behavior), all const fields are optional. When false (new behavior), const fields respect the OpenAPI spec's required array.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    privateIdentifierPrefix: {
      Name: "privateIdentifierPrefix",
      Required: false,
      DefaultValue: "#",
      Description:
        "Prefix used for internal class fields and methods in generated SDK code. Defaults to '#' (ECMAScript hard-private fields, e.g. `#hooks`).",
      ValidationRegex: /^(#|_|__|\$|)$/.source,
      ValidationMessage: "must be one of: #, _, __, $, or empty string",
    },
    multipartArrayFormat: {
      Name: "multipartArrayFormat",
      Required: false,
      DefaultValue: newSDK ? "standard" : "legacy",
      Description:
        "Format for array field names in multipart form data. 'legacy' appends '[]' to array field names (e.g., 'files[]'), which was the previous behavior. 'standard' uses the field name as-is without any suffix. Set to 'standard' for correct multipart/form-data handling.",
      ValidationRegex: /^(legacy|standard)$/.source,
      ValidationMessage: "legacy or standard only",
    },
    modelPropertyCasing: {
      Name: "modelPropertyCasing",
      Required: false,
      DefaultValue: "camel",
      Description:
        "Property naming convention to use. 'snake' converts property names to snake_case format, 'camel' converts to camelCase format.",
      ValidationRegex: /^(camel|snake)$/.source,
      ValidationMessage: "camel or snake only",
    },
    preserveModelFieldNames: {
      Name: "preserveModelFieldNames",
      Required: false,
      DefaultValue: false,
      Description:
        "When true, preserves the original casing of model property names from the OpenAPI spec. Synthetic fields (like additionalProperties, clientID, etc.) will still respect modelPropertyCasing.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    generateExamples: {
      Name: "generateExamples",
      Required: false,
      DefaultValue: true,
      Description:
        "Whether to generate example files in an examples directory demonstrating SDK usage.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    usageSDKInit: {
      Name: "usageSDKInit",
      Required: false,
      Description:
        "The SDK initialization code to use in usage examples. eg `new Petstore({})`",
    },
    usageSDKInitImports: {
      Name: "usageSDKInitImports",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of imports to add when usageSDKInit is configured. Each import should have 'package', 'import', and optionally 'type' fields.",
      ValidationFunc: (value: any): string | null => {
        if (!Array.isArray(value)) {
          return "usageSDKInitImports must be an array";
        }
        for (let i = 0; i < value.length; i++) {
          const item = value[i];
          if (typeof item !== "object" || item === null) {
            return `usageSDKInitImports[${i}] must be an object`;
          }
          if (!item.package || typeof item.package !== "string") {
            return `usageSDKInitImports[${i}].package must be a non-empty string`;
          }
          if (!item.import || typeof item.import !== "string") {
            return `usageSDKInitImports[${i}].import must be a non-empty string`;
          }
          if (
            item.type &&
            !["typeImport", "packageImport", "aliasImport"].includes(item.type)
          ) {
            return `usageSDKInitImports[${i}].type must be one of: typeImport, packageImport, aliasImport`;
          }
        }
        return null;
      },
    },
    sseFlatResponse: {
      Name: "sseFlatResponse",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether to flatten SSE (Server-Sent Events) responses by extracting the 'data' field from wrapper models, providing direct access to the event data instead of the wrapper object",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    acceptHeaderEnum: {
      Name: "acceptHeaderEnum",
      Required: false,
      DefaultValue: newSDK ? false : true,
      Description:
        "Whether to generate TypeScript enums for controlling the return content type of SDK methods when multiple accept types are available",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    exportZodModelNamespace: {
      Name: "exportZodModelNamespace",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether to export the deprecated $ namespace containing inboundSchema and outboundSchema aliases",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    zodVersion: {
      Name: "zodVersion",
      Required: false,
      DefaultValue: newSDK ? "v4-mini" : "v3",
      Description:
        'The version of Zod to use for schema validation. Options: v3, v4, v4-mini, or "none" to omit zod entirely (no runtime validation or transforms — JSON.parse/JSON.stringify only).',
      ValidationRegex: /^(v3|v4|v4-mini|none)$/.source,
      ValidationMessage: 'v3, v4, v4-mini, or "none" only',
    },
    alwaysIncludeInboundAndOutbound: {
      Name: "alwaysIncludeInboundAndOutbound",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether to always include both inbound and outbound schemas for all types regardless of usage",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: newSDK ? "populated-fields" : "left-to-right",
      Description:
        "Strategy for deserializing union types. 'left-to-right' tries each type in order and returns the first valid match. 'populated-fields' tries all types and returns the one with the most matching fields (including optional fields).",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    laxMode: {
      Name: "laxMode",
      Required: false,
      DefaultValue: newSDK ? "lax" : "strict",
      Description:
        'When set to \'lax\', required fields will be coerced to their zero value. eg a missing required string will fallback to "". Lax mode also applies other coercions eg boolean schemas will accept the string "true". Lax mode only applies to deserialization of responses. When laxMode is enabled, unionStrategy is automatically set to "populated-fields".',
      ValidationRegex: /^(lax|strict)$/.source,
      ValidationMessage: '"lax" or "strict" only',
    },
    forwardCompatibleEnumsByDefault: {
      Name: "forwardCompatibleEnumsByDefault",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "When true, any enum which is used on a response will be automatically open/forward compatible - i.e. unknown values will be tolerated. Single value enums won't be automatically opened. Individual enums can be controlled with x-speakeasy-unknown-values: allow/disallow.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    preApplyUnionDiscriminators: {
      Name: "preApplyUnionDiscriminators",
      Required: false,
      DefaultValue: true,
      Description:
        "When true, discriminator values will be applied as const fields onto union member types when those types are consistently used with the same discriminator value. This simplifies the type structure by pre-applying discriminators where possible. It also enables the use of `z.discriminatedUnion()`. Enabling this option will coerce `constFieldsAlwaysOptional` to false.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    forwardCompatibleUnionsByDefault: {
      Name: "forwardCompatibleUnionsByDefault",
      Required: false,
      DefaultValue: newSDK ? "tagged-only" : "false",
      Description:
        "Controls forward compatibility for discriminated unions in responses. 'tagged-only' makes only discriminated unions open to unknown values. 'false' disables forward compatibility. Unknown discriminator values are returned as an object with the discriminator field set to a special 'UNKNOWN' string and a raw field containing the original input. Individual unions can be controlled with x-speakeasy-unknown-values: allow/disallow.",
      ValidationRegex: /^(tagged-only|false)$/.source,
      ValidationMessage: "tagged-only or false only",
    },
    inferUnionDiscriminators: {
      Name: "inferUnionDiscriminators",
      Required: false,
      DefaultValue: true,
      Description:
        "Infer union discriminators for oneOfs missing explicit OpenAPI discriminator mapping",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    formStringArrayEncodeMode: {
      Name: "formStringArrayEncodeMode",
      Required: false,
      DefaultValue: newSDK ? "array" : "encoded-string",
      Description:
        "When in 'array' mode, string arrays will be encoded as ['string-one', 'string-two'], while in 'encoded-string' mode (the legacy behaviour) the string arrays will be encoded as 'string-one,string-two'.  Only use 'encoded-string' mode for compatibility.",
      ValidationRegex: /^(array|encoded-string)$/.source,
      ValidationMessage: '"array" or "encoded-string" only',
    },
    flatAdditionalProperties: {
      Name: "flatAdditionalProperties",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "When enabled, models with `additionalProperties: true` use an index signature instead of a nested additionalProperties field. Extra properties are accessed directly on the object. Only applies to untyped additionalProperties (i.e. `additionalProperties: true`). Typed additionalProperties (e.g. `additionalProperties: { type: string }`) retain the nested field structure to preserve type information.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    legacyFileNaming: {
      Name: "legacyFileNaming",
      Required: false,
      DefaultValue: !newSDK,
      Description:
        "When true, uses legacy file naming (camelCase/lowercase). When false, uses kebab-case (e.g., simple-object.ts). New SDKs default to false (kebab-case).",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    useOxlint: {
      Name: "useOxlint",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "When true, uses oxlint for linting (faster). When false, uses eslint.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    useTsgo: {
      Name: "useTsgo",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "When true, uses tsgo (native TypeScript compiler) for faster builds. When false, uses tsc.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    requestExtras: {
      Name: "requestExtras",
      Required: false,
      DefaultValue: false,
      Description:
        "When true, SDK method request options accept an extraQuery map whose entries are set on the request URL's query string, replacing operation parameters of the same name (null values are skipped, array values are exploded into repeated keys and plain object values are JSON-encoded), and an extraBody map whose entries are merged into JSON object request bodies, replacing fields of the same name.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    apiPromiseHelpers: {
      Name: "apiPromiseHelpers",
      Required: false,
      DefaultValue: false,
      Description:
        "When true, the generated APIPromise<T> exposes `.withResponse()` and `.asResponse()` helpers for accessing the raw Response object alongside the parsed body.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    fixEnumNameSanitization: {
      Name: "fixEnumNameSanitization",
      Required: false,
      DefaultValue: false,
      Description:
        "When true, preserves custom enum names provided via x-speakeasy-enums without applying casing transformations, with illegal character sanitization still applied.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
  };
}

/** Represents the initial target implementation configuration presented to
 * the generator from the target. Customer configuration keys and values from
 * getConfigFields are fetched prior so those values are available. */
// @ts-ignore
function getGeneratorConfig(
  genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  const result: TargetInitialConfiguration = {
    compile: getCompileConfiguration(genConfig),
    testing: getTestingConfiguration(genConfig),
  };

  return result;
}

// @ts-ignore
function getCompileConfiguration(
  genConfig: GeneratorInitialConfiguration,
): CompileConfiguration {
  return {
    runner: {
      commands: getCompileCommands(genConfig),
      dependencies: {
        commands: getCompileDependencies(genConfig),
      },
    },
  };
}

// @ts-ignore
function getTestingConfiguration(
  genConfig: GeneratorInitialConfiguration,
): TestingConfiguration {
  const installCommands = getInstallCommands(genConfig.LangCfg);

  return {
    runner: {
      commands: [
        ...installCommands,
        {
          command: "npm",
          args: ["run", "test"],
          environmentVariables: {
            NODE_OPTIONS: "--max-old-space-size=6144",
          },
        },
      ],
      dependencies: {
        commands: getCompileDependencies(genConfig),
      },
    },
  };
}

/**
 * Hardcoded configuration overlay for the target. This is used to configure
 * required settings for the target that are intentionally not exposed.
 */
// @ts-ignore
function getConfigOverlay(): Record<string, any> {
  return {
    reservedModelFileNames: ["index", "sdkvalidationerror", "httpclienterrors"],
  };
}
