// @ts-ignore
function getProjectIdentifier(): string {
  return caser()
    .ToKebab(context.Global.Config?.PackageName ?? "unknown")
    .toLowerCase();
}

function getSpeakeasyTrackingLink(): string {
  return `https://www.speakeasy.com/?utm_source=${getProjectIdentifier()}&utm_campaign=${
    context.Global.Config.Language
  }`;
}
registerTemplateFunc("getSpeakeasyTrackingLink", getSpeakeasyTrackingLink);

// @ts-ignore
function templateReadmeTitle(): string {
  return context.Global.Config.PackageName;
}
registerTemplateFunc("templateReadmeTitle", templateReadmeTitle);

// @ts-ignore
function templateReadmeForeword(): string {
  const apiName = templateReadmeTitle();
  const language = context.Global.Config.Language;

  if (language.startsWith("mcp-")) {
    return `Model Context Protocol (MCP) Server for the *${apiName}* API.`;
  } else if (language === "terraform") {
    return `Terraform Provider for the *${apiName}* API.`;
  } else if (language === "cli") {
    return `Command-line interface for the *${apiName}* API.`;
  }

  return `Developer-friendly & type-safe ${caser().ToPascal(
    language,
  )} SDK specifically catered to leverage *${apiName}* API.`;
}
registerTemplateFunc("templateReadmeForeword", templateReadmeForeword);

function getLicenseUrl(): string {
  if (context.Global.Config?.GeneratedLicense === "agpl") {
    return "https://www.gnu.org/licenses/agpl-3.0.html";
  }

  return (
    context.Global.Config?.License?.url ?? "https://opensource.org/licenses/MIT"
  );
}
registerTemplateFunc("getLicenseUrl", getLicenseUrl);

function getLicenseShortName(): string {
  if (context.Global.Config?.GeneratedLicense === "agpl") {
    return "AGPL-3.0-only";
  }

  return context.Global.Config?.License?.shortName ?? "MIT";
}
registerTemplateFunc("getLicenseShortName", getLicenseShortName);

function getLicenseBadgeShortName(): string {
  return getLicenseShortName().replaceAll("-", "--");
}
registerTemplateFunc("getLicenseBadgeShortName", getLicenseBadgeShortName);
