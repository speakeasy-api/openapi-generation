function registerDocumentation(providerContext: ProviderContext) {
  const TERRAFORM_README_SECTIONS: AvailableReadmeSections = {
    summary: {
      Header: "Summary",
      Weight: 0,
      Feature: "core",
    },
    toc: {
      Header: "Table of Contents",
      Weight: 1,
      Feature: "core",
    },
    installation: {
      Header: "Installation",
      Weight: 2,
      Feature: "core",
    },
    security: {
      Header: "Authentication",
      Weight: 3,
      Feature: "core",
    },
    operations: {
      Header: "Available Resources and Data Sources",
      Weight: 20,
      Feature: "core",
      LegacyHeader: "SDK Available Operations",
    },
    usage: {
      Header: "Testing the provider locally",
      Weight: 100,
      Feature: "core",
    },
  };

  registerReadmeSection(
    "summary",
    () => context.Global.AST.MainSDK.Comments != undefined,
    (_sdk: SDK) => templateString("readme/summary.stmpl", {}),
    TERRAFORM_README_SECTIONS,
  );

  registerReadmeSection(
    "toc",
    () => true,
    (_sdk: SDK) => "", // generated separately based on final README content
    TERRAFORM_README_SECTIONS,
  );

  registerReadmeSection(
    "installation",
    () => true,
    (_sdk: SDK) => {
      const providerTFContent =
        readFile("examples/provider/provider.tf") ||
        templateString("examples/provider.tf.stmpl", providerContext);
      return `To install this provider, copy and paste this code into your Terraform configuration. Then, run \`terraform init\`.\n\n\`\`\`hcl\n${providerTFContent.trim()}\n\`\`\``;
    },
    TERRAFORM_README_SECTIONS,
  );

  registerReadmeSection(
    "security",
    () => hasProviderSecurityAttributes(providerContext),
    (_sdk: SDK) => templateReadmeSecurity(providerContext),
    TERRAFORM_README_SECTIONS,
  );

  registerReadmeSection(
    "operations",
    () => true,
    (_sdk: SDK) => templateReadmeResources(providerContext),
    TERRAFORM_README_SECTIONS,
  );

  registerReadmeSection(
    "usage",
    () => true,
    (_sdk: SDK) => templateString("readme/usage.stmpl", {}),
    TERRAFORM_README_SECTIONS,
  );
}

function templateReadmeSecurity(providerContext: ProviderContext): string {
  const providerSecurityAttributes = getProviderSecurityAttributes(
    providerContext.AST.MainSDK.Security?.Type,
  );
  const hasEnvironmentVariables = Object.keys(providerSecurityAttributes).some(
    (attributeName) => {
      return getEnvironmentVariable(attributeName);
    },
  );

  const result: string[] = [];

  if (hasEnvironmentVariables) {
    result.push(
      "This provider supports authentication configuration via environment variables and provider configuration.",
      "",
      "The configuration precedence is:",
      "",
      "- Provider configuration",
      "- Environment variables",
    );
  } else {
    result.push(
      "This provider supports authentication configuration via provider configuration.",
    );
  }

  result.push(
    "",
    "Available configuration:",
    "",
    "| Provider Attribute | Description |",
    "|---|---|",
  );

  Object.keys(providerSecurityAttributes)
    .sort()
    .forEach((attributeName) => {
      const attribute = providerSecurityAttributes[attributeName];

      result.push(
        `| \`${attributeName}\` | ${attribute.markdownDescription} |`,
      );
    });

  result.push("");

  return result.join("\n");
}

type ReadmeEntity = { Name: string; TerraformTypeName: string };

function templateReadmeResourceSection(
  header: string,
  docsSubdir: string,
  entities: ReadmeEntity[],
): string {
  if (entities.length === 0) {
    return "";
  }

  let section = `### ${header}\n\n`;
  for (const entity of entities) {
    section += `* [${
      entity.TerraformTypeName
    }](docs/${docsSubdir}/${sanitizeTFStateName(entity.Name)}.md)\n`;
  }
  section += `\n`;
  return section;
}

function templateReadmeResources(providerContext: ProviderContext): string {
  return (
    templateReadmeResourceSection(
      "Managed Resources",
      "resources",
      providerContext.Resources,
    ) +
    templateReadmeResourceSection(
      "Data Sources",
      "data-sources",
      providerContext.DataSources,
    ) +
    templateReadmeResourceSection(
      "Ephemeral Resources",
      "ephemeral-resources",
      providerContext.EphemeralResources,
    ) +
    templateReadmeResourceSection("Actions", "actions", providerContext.Actions)
  );
}
