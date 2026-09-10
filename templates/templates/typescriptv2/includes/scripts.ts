function templateScripts(): string {
  const defaultScripts: Record<string, string> = {};

  // Add test scripts if tests are enabled
  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    defaultScripts["test"] =
      "vitest run src --reporter=junit --outputFile=.speakeasy/reports/tests.xml --reporter=default";
    defaultScripts["check"] = "npm run test && npm run lint";
  }

  // Add lint script
  if (context.Global.Config.UseOxlint) {
    defaultScripts["lint"] =
      "oxlint --max-warnings=0 --deny-warnings --no-error-on-unmatched-pattern src";
  } else {
    defaultScripts["lint"] = "eslint --cache --max-warnings=0 src";
  }

  // Add MCP build script if enabled
  if (isMCPServerEnabled()) {
    defaultScripts["build:mcp"] = "bun src/mcp-server/build.mts";
  }

  // Add build script based on module format
  let buildCommand: string;
  if (context.Global.Config.ModuleFormat === "dual") {
    buildCommand = "tshy";
  } else {
    buildCommand =
      context.Global.Config.UseTsgo &&
      context.Global.Config.ModuleFormat !== "commonjs"
        ? "tsgo"
        : "tsc";
  }

  // Prepend MCP build if enabled
  if (isMCPServerEnabled()) {
    buildCommand = `npm run build:mcp && ${buildCommand}`;
  }

  defaultScripts["build"] = buildCommand;
  defaultScripts["prepublishOnly"] = "npm run build";

  return renderScripts(defaultScripts, context.Global.Config.AdditionalScripts);
}
registerTemplateFunc("templateScripts", templateScripts);

// @ts-ignore
function renderScripts(
  defaultScripts: Record<string, string>,
  additionalScripts: Record<string, string>,
): string {
  const scripts = {
    ...defaultScripts,
    ...additionalScripts,
  };

  let renderedScripts = "";

  for (const scriptName of Object.keys(scripts)) {
    const command = JSON.stringify(scripts[scriptName]);
    renderedScripts += `"${scriptName}": ${command},\n`;
  }
  renderedScripts = renderedScripts.replace(/,\n$/, "");

  return renderedScripts;
}
