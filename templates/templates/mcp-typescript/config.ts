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

// @ts-ignore
function getCompileCommands(
  config: GeneratorInitialConfiguration,
): RunnerCommands {
  const commands = [
    {
      command: "npm",
      args: ["install", "--ignore-scripts"],
    },
    {
      command: "npm",
      args: ["rebuild", "bun"],
    },
    {
      command: "npm",
      args: ["run", "lint"],
      environmentVariables: {
        NODE_OPTIONS: "--max-old-space-size=6144",
      },
    },
  ];

  // Generate Cloudflare types before build if enabled
  if (config.cloudflareEnabled) {
    commands.push({
      command: "npm",
      args: ["run", "types"],
    });
  }

  commands.push({
    command: "npm",
    args: ["run", "build"],
    environmentVariables: {
      NODE_OPTIONS: "--max-old-space-size=6144",
    },
  });

  return commands;
}

/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    express: {
      name: "express",
      version: "^5.1.0",
      cpe: "cpe:2.3:a:expressjs:express:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    "@modelcontextprotocol/sdk": {
      name: "@modelcontextprotocol/sdk",
      version: "1.26.0",
      cpe: "cpe:2.3:a:anthropic:model_context_protocol_sdk:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    "@stricli/core": {
      name: "@stricli/core",
      version: "^1.1.2",
      cpe: "cpe:2.3:a:stricli:core:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    zod: {
      name: "zod",
      version: "^4.0.0",
      cpe: "cpe:2.3:a:colinhacks:zod:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    ai: {
      name: "ai",
      version: "^6.0.77",
      cpe: "cpe:2.3:a:vercel:ai:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    "@ai-sdk/mcp": {
      name: "@ai-sdk/mcp",
      version: "^1.0.19",
      cpe: "cpe:2.3:a:vercel:ai-sdk-mcp:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    "@ai-sdk/anthropic": {
      name: "@ai-sdk/anthropic",
      version: "^3.0.38",
      cpe: "cpe:2.3:a:anthropic:ai-sdk-anthropic:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    "@ai-sdk/openai": {
      name: "@ai-sdk/openai",
      version: "^3.0.26",
      cpe: "cpe:2.3:a:openai:ai-sdk-openai:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
    },
    agents: {
      name: "agents",
      version: "0.3.10",
      cpe: "cpe:2.3:a:cloudflare:agents:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "cloudflare",
    },
    "@gram-ai/functions": {
      name: "@gram-ai/functions",
      version: "^0.12.1",
      cpe: "cpe:2.3:a:gram:functions:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "gram",
    },
    "decimal.js": {
      name: "decimal.js",
      version: "^10.5.0",
      cpe: "cpe:2.3:a:mikemcl:decimal.js:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "runtime",
      condition: "decimal",
    },
    // Dev dependencies
    bun: {
      name: "bun",
      version: "^1.2.12",
      cpe: "cpe:2.3:a:oven:bun:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    eslint: {
      name: "eslint",
      version: "^9.26.0",
      cpe: "cpe:2.3:a:eslint:eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@eslint/js": {
      name: "@eslint/js",
      version: "^9.26.0",
      cpe: "cpe:2.3:a:eslint:eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    globals: {
      name: "globals",
      version: "^16.0.0",
      cpe: "cpe:2.3:a:sindresorhus:globals:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    typescript: {
      name: "typescript",
      version: "~5.8.3",
      cpe: "cpe:2.3:a:microsoft:typescript:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "typescript-eslint": {
      name: "typescript-eslint",
      version: "^8.31.1",
      cpe: "cpe:2.3:a:typescript-eslint:typescript-eslint:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@types/bun": {
      name: "@types/bun",
      version: "^1.2.13",
      cpe: "cpe:2.3:a:definitelytyped:types-bun:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@types/express": {
      name: "@types/express",
      version: "^5.0.1",
      cpe: "cpe:2.3:a:definitelytyped:types-express:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@types/node": {
      name: "@types/node",
      version: "^18.19.3",
      cpe: "cpe:2.3:a:definitelytyped:types-node:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@anthropic-ai/mcpb": {
      name: "@anthropic-ai/mcpb",
      version: "^1.2.0",
      cpe: "cpe:2.3:a:anthropic:mcpb:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
    },
    "@cloudflare/workers-types": {
      name: "@cloudflare/workers-types",
      version: "^4.20250712.0",
      cpe: "cpe:2.3:a:cloudflare:workers-types:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "cloudflare",
    },
    wrangler: {
      name: "wrangler",
      version: "^4.59.1",
      cpe: "cpe:2.3:a:cloudflare:wrangler:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "cloudflare",
    },
    vitest: {
      name: "vitest",
      version: "^3.2.6",
      cpe: "cpe:2.3:a:vitest:vitest:*:*:*:*:*:node.js:*:*",
      ecosystem: "npm",
      category: "dev",
      condition: "tests",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

// @ts-ignore
function upgradeConfig(_oldVersion, _newVersion, cfg, _defaults) {
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
    envVarPrefix: {
      Name: "envVarPrefix",
      Required: false,
      Description:
        "The environment variable prefix for security and global env variable overrides. If empty these overrides will not be possible",
    },
    cloudflareEnabled: {
      Name: "cloudflareEnabled",
      Required: false,
      DefaultValue: false,
      Description:
        "Enable Cloudflare Workers support for hosting the MCP server on Cloudflare's edge network",
    },
    gramEnabled: {
      Name: "gramEnabled",
      Required: false,
      DefaultValue: false,
      Description:
        "Enable Gram deployment support for hosting the MCP server on Gram",
    },
    cloudflareURL: {
      Name: "cloudflareURL",
      Required: false,
      Description:
        "The Cloudflare Worker URL for the MCP server (e.g., https://my-mcp-server.my-account.workers.dev). This is used for installation documentation.",
      ValidationRegex: /^https?:\/\/.+$/.source,
      ValidationMessage:
        "Must be a valid URL starting with http:// or https://",
    },
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "mcp",
      Description:
        "The npm package name. https://docs.npmjs.com/package-name-guidelines.",
      ValidationRegex: /^[\w\d._@\/-]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    version: {
      Name: "version",
      Required: true,
      DefaultValue: "0.0.1",
      Description: "The current version of the MCP Server",
      ValidationRegex: /^[\w\d._-]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    dxtManifestOverlay: {
      Name: "dxtManifestOverlay",
      Required: false,
      Description:
        "Overlay for the DXT manifest. DEPRECATED: Use mcpbManifestOverlay instead.",
    },
    mcpbManifestOverlay: {
      Name: "mcpbManifestOverlay",
      Required: false,
      Description:
        "Overlay for the MCPB manifest. Can be a deeply nested YAML object. " +
        "See: https://github.com/anthropics/mcpb/blob/main/MANIFEST.md " +
        "for supported fields.",
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
    mcpName: {
      Name: "mcpName",
      Required: false,
      Description:
        "The unique identifier for the MCP server in reverse-DNS format (e.g., 'io.github.username/server-name'). " +
        "This is required for publishing to the Anthropic MCP registry. " +
        "See: https://github.com/modelcontextprotocol/registry/blob/main/docs/guides/publishing/publish-server.md",
      ValidationRegex: /^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*\/[a-z][a-z0-9\-]*$/
        .source,
      ValidationMessage:
        "Must be in reverse-DNS format (e.g., 'io.github.username/server-name')",
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
    formStringArrayEncodeMode: {
      Name: "formStringArrayEncodeMode",
      Required: false,
      DefaultValue: newSDK ? "array" : "encoded-string",
      Description:
        "When in 'array' mode, string arrays will be encoded as ['string-one', 'string-two'], while in 'encoded-string' mode (the legacy behaviour) the string arrays will be encoded as 'string-one,string-two'.  Only use 'encoded-string' mode for compatibility.",
      ValidationRegex: /^(array|encoded-string)$/.source,
      ValidationMessage: '"array" or "encoded-string" only',
    },
    evalProvider: {
      Name: "evalProvider",
      Required: false,
      Description:
        "The LLM provider for the eval command. When set, enables the eval CLI subcommand and installs the corresponding AI SDK package.",
      ValidationRegex: /^(anthropic|openai)$/.source,
      ValidationMessage: '"anthropic" or "openai" only',
    },
    advancedFlags: {
      Name: "advancedFlags",
      Required: false,
      Description:
        "A list of flags that are considered advanced and should not be exposed to the user.",
      DefaultValue: [],
      ValidationFunc: (value: any): string | null => {
        if (!Array.isArray(value)) {
          return "advancedFlags must be an array";
        }
        const validPattern = /^[a-zA-Z0-9]+$/;
        for (let i = 0; i < value.length; i++) {
          const item = value[i];
          if (typeof item !== "string") {
            return `advancedFlags[${i}] must be a string`;
          }
          if (!validPattern.test(item)) {
            return `advancedFlags[${i}] must contain only letters or numbers`;
          }
        }
        return null;
      },
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
    validateResponse: {
      Name: "validateResponse",
      Required: false,
      DefaultValue: newSDK ? false : true,
      Description:
        "Determines how strictly tool/resource results are validated. When set to `true`, HTTP responses are validated using generated Zod schemas before being returned as tool results. When set to `false`, no structured deserialization and validation is done on HTTP response data, instead the raw data is passed through as MCP tool results.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: '"true" or "false" only',
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
    defaultErrorName: "APIError",
    imports: {
      option: "openapi",
      paths: {
        callbacks: "models",
        errors: "models/errors",
        operations: "models",
        shared: "models",
        webhooks: "models",
      },
    },
    reservedModelFileNames: ["sdkvalidationerror", "httpclienterrors"],
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
  const compileCommands = getCompileCommands(genConfig);

  return {
    runner: {
      commands: [
        // Reference: internal issue reference
        ...compileCommands,
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
