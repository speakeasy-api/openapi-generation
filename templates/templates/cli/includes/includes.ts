require("common/includes.ts");

require("go/annotations.ts");
require("includes/utils.ts");
require("includes/dependencies.ts");
require("includes/sanitization.ts");
require("includes/flags.ts");
require("includes/descriptions.ts");
require("includes/templating.ts");
require("includes/imports.ts");
require("includes/unions.ts");
require("includes/metadata.ts");
require("includes/intents.ts");
require("includes/errors.ts");
require("includes/readme-examples.ts");
require("includes/readme-commands.ts");
require("includes/security.ts");
require("includes/usage.ts");
require("includes/serverusage.ts");
require("includes/unimplemented.ts");
require("includes/release.ts");

// CLI-specific globals table including the --flag name column
function templateCLIGlobalsTable(): string {
  const headers: string[] = ["Flag", "Type", "Description"];
  if (context.Global.Config.EnvVarPrefix) {
    headers.push("Environment");
  }

  const contents: string[][] = [headers];

  context.Global.AST.MainSDK.Globals.Fields.forEach((field: FieldDef) => {
    const flagName = sanitizeFlagNameWithReserved(field.Name);
    const row: string[] = [
      `\`--${flagName}\``,
      templateGlobalType(field),
      templateGlobalDescription(field),
    ];

    if (context.Global.Config.EnvVarPrefix) {
      row.push(templateGlobalEnvVars(field));
    }

    contents.push(row);
  });

  return createMarkdownTable(contents, false);
}
registerTemplateFunc("templateCLIGlobalsTable", templateCLIGlobalsTable);

// A realistic value for the first global parameter: its declared default,
// then its example, then the first enum member, then a type-shaped literal.
function readmeGlobalExampleValue(field: FieldDef): string {
  if (field.Default?.Value !== undefined && field.Default?.Value !== null) {
    return String(field.Default.Value);
  }
  // @ts-ignore — Example is a dynamic Go proxy field not in TS type defs
  if (field.Example?.Value !== undefined && field.Example?.Value !== null) {
    // @ts-ignore
    return String(field.Example.Value);
  }
  const typeDef = field.Type;
  if (
    typeDef?.Type?.toString() === "enum" &&
    typeDef.Enum?.Values?.length > 0
  ) {
    return String(typeDef.Enum.Values[0]);
  }
  switch (typeDef?.Type?.toString()) {
    case "boolean":
      return "true";
    case "integer":
    case "int32":
      return "1";
    case "number":
    case "float32":
      return "1.5";
  }
  return "value";
}

// Quote a value for a POSIX shell. Common flag-friendly characters stay bare;
// everything else is single-quoted so $, backticks and double quotes remain
// literal. A single quote inside the value is emitted as: '\''.
function readmeShellValue(value: string): string {
  if (/^[A-Za-z0-9._@/:-]+$/.test(value)) return value;
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

// Generate an env var example for the README (uses first global param)
function templateGlobalEnvVarExample(): string {
  const globals = context.Global.AST.MainSDK.Globals;
  if (!globals || !globals.Fields || globals.Fields.length === 0) return "";
  const field = globals.Fields[0];
  const envVar = templateGlobalEnvVars(field);
  return `${envVar}=${readmeShellValue(readmeGlobalExampleValue(field))}`;
}
registerTemplateFunc(
  "templateGlobalEnvVarExample",
  templateGlobalEnvVarExample,
);

// Generate a flag example for the README (uses first global param)
function templateGlobalFlagExample(): string {
  const globals = context.Global.AST.MainSDK.Globals;
  if (!globals || !globals.Fields || globals.Fields.length === 0)
    return '--param "value"';
  const field = globals.Fields[0];
  const flagName = sanitizeFlagNameWithReserved(field.Name);
  const value = readmeGlobalExampleValue(field);
  switch (field.Type?.Type?.toString()) {
    case "boolean":
      return `--${flagName}=${value}`;
    case "integer":
    case "int32":
    case "number":
    case "float32":
      return `--${flagName} ${value}`;
    default:
      return `--${flagName} ${readmeShellValue(value)}`;
  }
}
registerTemplateFunc("templateGlobalFlagExample", templateGlobalFlagExample);

// Register README sections individually (operations links to Cobra-generated docs)
registerReadmeSection(
  "summary",
  () => context.Global.AST.MainSDK.Comments != undefined,
  (_sdk: SDK) => templateString("readme/summary.stmpl", {}),
);
registerReadmeSection(
  "toc",
  () => true,
  (_sdk: SDK) => "",
);
registerReadmeSection(
  "installation",
  () => true,
  (_sdk: SDK) => templateString("readme/installation.stmpl", {}),
  {
    installation: {
      Header: "CLI Installation",
      Weight: 2,
      Feature: "core",
    },
  },
);
// CLI Example Usage: declared intent commands lead (they are the curated
// entry points), followed by the spec's usage examples.
function templateCLIUsageSection(): string {
  let out = "";
  const manifest = collectIntentManifest();
  const bound = manifest.Bound.filter(
    (i) => readmeIntentExamples(i).length > 0,
  );
  if (bound.length > 0) {
    out += "### Quick start\n\n```bash\n";
    let first = true;
    for (const intent of bound.slice(0, 3)) {
      const summary = readmeIntentExampleSummaries(intent)[0] || "";
      if (!first) out += "\n";
      first = false;
      if (summary) out += `# ${summary}\n`;
      out += `${readmeIntentExamples(intent)[0]}\n`;
    }
    out += "```\n\n";
  }
  // The spec's usage examples, minus operations whose generated command an
  // override replaced: their example would invoke a removed flag surface.
  out += templateUsageSectionWithContexts(
    collectAllUsageExamples().filter(
      (c: UsageContext) => !c.Operation || !isOperationOverridden(c.Operation),
    ),
    "### Example",
  );
  return out;
}
registerReadmeSection(
  "usage",
  () => true,
  (_sdk: SDK) => templateCLIUsageSection(),
  {
    usage: {
      Header: "CLI Example Usage",
      Weight: 5,
      Feature: "core",
    },
  },
);
registerReadmeSection(
  "stdinpiping" as any,
  () => true,
  (_sdk: SDK) =>
    templateString("readme/stdinpiping.stmpl", readmeExampleContext()),
  {
    stdinpiping: {
      Header: "Request Body Input",
      Weight: 20,
      Feature: "core",
    },
  } as any,
);
// Commands: mirrors root --help (declared categories, intents, planned
// placeholders, then the untagged operation tree). Keeps the historical
// "operations" block ID so existing READMEs update in place.
registerReadmeSection("operations", () => true, templateCLICommandsSection, {
  operations: {
    Header: "Commands",
    Weight: 15,
    Feature: "core",
  },
});
// For AI agents: the discovery ladder (--help → --usage → --schema →
// --dry-run → machine-readable output / agent-mode envelope). Placed right
// after the first usage example so agents reading top-down meet it early.
registerReadmeSection(
  "agents" as any,
  () => true,
  (_sdk: SDK) => templateString("readme/agents.stmpl", readmeExampleContext()),
  {
    agents: {
      Header: "For AI agents",
      Weight: 6,
      Feature: "core",
    },
  } as any,
);
registerReadmeSection(
  "dev-containers",
  () => context.Global.Config.DevContainerSchemaPath != undefined,
  (_sdk: SDK) => templateString("readme/dev-containers.stmpl", {}),
);
const cliGlobalsInfo = prepareGlobalParametersSection();
function templateGlobalExampleCommand(): string {
  return readmeExampleContext().ShowcaseArgs;
}
registerTemplateFunc(
  "templateGlobalExampleCommand",
  templateGlobalExampleCommand,
);
registerReadmeSection(
  "global-parameters",
  () => cliGlobalsInfo.NeedsGlobalParameters,
  (_sdk: SDK) => templateString("readme/globals.stmpl", cliGlobalsInfo),
  {
    "global-parameters": {
      Header: "Configuration",
      Weight: 8,
      Feature: "globals",
    },
  },
);

// Pagination: CLI-specific registration (the shared section renders the
// template without the README example context).
registerReadmeSection(
  "pagination",
  () => anyOperationHasPagination(),
  (_sdk: SDK) =>
    templateString("readme/pagination.stmpl", readmeExampleContext()),
);
require("readme/sections/errors.ts");
// Register security (Authentication) with CLI-specific weight — placed right after usage, before configuration
registerReadmeSection(
  "security",
  sdkHasGlobalSecurity,
  (_sdk: SDK) =>
    templateString("readme/security.stmpl", readmeExampleContext()),
  {
    security: {
      Header: "Authentication",
      Weight: 7,
      Feature: "globalSecurity",
    },
  },
);
// Register server selection with CLI-specific weight — placed after available commands and request body input
registerReadmeSection(
  "server",
  () => context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false,
  (_sdk: SDK) =>
    templateString("readme/server.stmpl", {
      ...getServerReadmeContext(),
      Example: readmeExampleContext(),
    }),
  {
    server: {
      Header: "Server Selection",
      Weight: 25,
      Feature: "serverIDs",
    },
  },
);

registerReadmeSection(
  "retries",
  () => isFeatureUsed("retries"),
  (_sdk: SDK) => templateString("readme/retries.stmpl", readmeExampleContext()),
);

registerReadmeSection(
  "eventstreaming" as any,
  () => isFeatureUsed("serverEvents"),
  (_sdk: SDK) =>
    templateString("readme/eventstreaming.stmpl", readmeExampleContext()),
  {
    eventstreaming: {
      Header: "Server-Sent Event Streaming",
      Weight: 55,
      Feature: "serverEvents",
    },
  } as any,
);

registerReadmeSection(
  "jsonlstreaming" as any,
  () => isFeatureUsed("jsonlResponses"),
  (_sdk: SDK) =>
    templateString("readme/jsonlstreaming.stmpl", readmeExampleContext()),
  {
    jsonlstreaming: {
      Header: "JSONL Streaming",
      Weight: 56,
      Feature: "jsonlResponses",
    },
  } as any,
);

registerReadmeSection(
  "fileuploads" as any,
  () => isFeatureUsed("uploads" as any),
  (_sdk: SDK) =>
    templateString("readme/fileuploads.stmpl", readmeExampleContext()),
  {
    fileuploads: {
      Header: "File Uploads",
      Weight: 57,
      Feature: "uploads" as any,
    },
  } as any,
);

// Output formats — applies to every command, so place early in the doc
registerReadmeSection(
  "output-formats" as any,
  () => true,
  (_sdk: SDK) => templateString("readme/output.stmpl", readmeExampleContext()),
  {
    "output-formats": {
      Header: "Output Formats",
      Weight: 28,
      Feature: "core",
    },
  } as any,
);

// Debug/diagnostics goes last — it's advanced/developer-facing content
registerReadmeSection(
  "diagnostics" as any,
  () => true,
  (_sdk: SDK) =>
    templateString("readme/diagnostics.stmpl", readmeExampleContext()),
  {
    diagnostics: {
      Header: "Diagnostics",
      Weight: 9999,
      Feature: "core",
    },
  } as any,
);

// Shell completion placed right after installation (weight 3)
registerReadmeSection(
  "completion" as any,
  () => true,
  (_sdk: SDK) => templateString("readme/completion.stmpl", {}),
  {
    completion: {
      Header: "Shell Completion",
      Weight: 3,
      Feature: "core",
    },
  } as any,
);
