// @ts-ignore
function getRootLocationRelativeToPackage(): string {
  let path = getProjectFilePath();
  path = path.substring(0, path.lastIndexOf("\\"));

  const parts = path.split("\\");

  return "..\\".repeat(parts.length);
}
registerTemplateFunc(
  "getRootLocationRelativeToPackage",
  getRootLocationRelativeToPackage,
);

// Strips <details>/<summary> HTML blocks that NuGet doesn't support.
// Converts them to bold headings with the content expanded.
// @ts-ignore
function stripDetailsBlocks(content: string): string {
  return content.replace(
    /<details[^>]*>\s*<summary>([^<]*)<\/summary>([\s\S]*?)<\/details>/gi,
    (_, title, inner) => `**${title.trim()}**\n\n${inner.trim()}`,
  );
}

// @ts-ignore
function templateNugetReadme() {
  const nugetSections = globallyRegisteredReadmeSections
    .filter(
      (s) => !["toc", "summary", "operations", "installation"].includes(s.ID),
    )
    .map((s) => ({
      ...s,
      // Wrap the Content callback to strip <details> blocks from output
      Content: (sdk: SDK) => stripDetailsBlocks(s.Content(sdk)),
    }));

  createOrEditReadme(
    context.Global.AST.MainSDK,
    `# ${getPackageId()}`,
    nugetSections,
    "readme/nuget.stmpl",
    "NUGET.md",
  );
}

// @ts-ignore
function getLicenseFile(): string {
  const license = listFiles(".").find((file) => /^LICENSE(\..+)?$/.test(file));
  return license || "";
}
registerTemplateFunc("getLicenseFile", getLicenseFile);

// @ts-ignore
function getIconFile(): string {
  const icon = listFiles(".").find((file) =>
    /^icon\.(jpg|jpeg|png)$/.test(file),
  );
  return icon || "";
}
registerTemplateFunc("getIconFile", getIconFile);

// @ts-ignore
function getCopyright(): string {
  const year = new Date().getFullYear();
  return `Copyright (c) ${context.Global.Config.Author} ${year}`;
}

registerTemplateFunc("getCopyright", getCopyright);

// @ts-ignore
function getPackageProjectURL(): string {
  // an invalid URL in the package manifest will cause the NuGet package to get rejected
  const url = context.Global.AST.MainSDK.Comments?.ExternalDocs?.URL;
  if (!url) {
    return "";
  }
  const basicPattern =
    /(?:https?):\/\/(\w+:?\w*)?(\S+)(:\d+)?(\/|\/([\w#!:.?+=&%!\-\/]))?/;
  return basicPattern.test(url) ? url : "";
}

registerTemplateFunc("getPackageProjectURL", getPackageProjectURL);
