// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

// @ts-ignore
function templateDependencies(): string {
  const defaultDependencies = {
    express: deps.express.version,
    "@modelcontextprotocol/sdk": deps["@modelcontextprotocol/sdk"].version,
    "@stricli/core": deps["@stricli/core"].version,
    zod: deps.zod.version,
    ...(context.Global.Config.EvalProvider
      ? {
          ai: deps.ai.version,
          "@ai-sdk/mcp": deps["@ai-sdk/mcp"].version,
          ...(context.Global.Config.EvalProvider === "openai"
            ? { "@ai-sdk/openai": deps["@ai-sdk/openai"].version }
            : { "@ai-sdk/anthropic": deps["@ai-sdk/anthropic"].version }),
        }
      : {}),
  };

  if (context.Global.Config.CloudflareEnabled) {
    defaultDependencies["agents"] = deps.agents.version;
  }

  if (context.Global.Config.GramEnabled) {
    defaultDependencies["@gram-ai/functions"] =
      deps["@gram-ai/functions"].version;
  }

  if (isFeatureUsed("decimal")) {
    defaultDependencies["decimal.js"] = deps["decimal.js"].version;
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.dependencies,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

function templateDevDependencies(): string {
  const defaultDependencies = {
    bun: deps.bun.version,
    eslint: deps.eslint.version,
    "@eslint/js": deps["@eslint/js"].version,
    globals: deps.globals.version,
    typescript: deps.typescript.version,
    "typescript-eslint": deps["typescript-eslint"].version,
    "@types/bun": deps["@types/bun"].version,
    "@types/express": deps["@types/express"].version,
    "@types/node": deps["@types/node"].version,
    "@anthropic-ai/mcpb": deps["@anthropic-ai/mcpb"].version,
  };

  if (context.Global.Config.CloudflareEnabled) {
    defaultDependencies["@cloudflare/workers-types"] =
      deps["@cloudflare/workers-types"].version;
    defaultDependencies["wrangler"] = deps.wrangler.version;
  }

  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    defaultDependencies["vitest"] = deps.vitest.version;
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.devDependencies,
  );
}
registerTemplateFunc("templateDevDependencies", templateDevDependencies);

function templatePeerDependencies(): string {
  const defaultDependencies: Record<string, string> = {};

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies?.peerDependencies,
  );
}
registerTemplateFunc("templatePeerDependencies", templatePeerDependencies);

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

// @ts-ignore
function templateMCPServerSecurityHeaderFields(): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  const fields = unnestSecurityEnvFields(
    context.Global.AST.MainSDK.Security.Type,
  );
  const securityFields: string[] = [];

  for (const field of fields) {
    const headerName = field.Name.toLowerCase();
    const headerKey = headerName.startsWith("x-")
      ? headerName
      : `x-${headerName}`;
    securityFields.push(`${field.Name}: this.props["${headerKey}"]`);
  }

  return `sdk.security = () => ({\n      ${securityFields.join(
    ",\n      ",
  )}\n    });`;
}
registerTemplateFunc(
  "templateMCPServerSecurityHeaderFields",
  templateMCPServerSecurityHeaderFields,
);
