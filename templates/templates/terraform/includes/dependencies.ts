// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

function templateDependencies(): string {
  const defaultDependencies = {
    "github.com/hashicorp/go-uuid":
      deps["github.com/hashicorp/go-uuid"].version, // debug logging
    "github.com/hashicorp/terraform-plugin-docs":
      deps["github.com/hashicorp/terraform-plugin-docs"].version,
    "github.com/hashicorp/terraform-plugin-framework-validators":
      deps["github.com/hashicorp/terraform-plugin-framework-validators"]
        .version,
    "github.com/hashicorp/terraform-plugin-framework":
      deps["github.com/hashicorp/terraform-plugin-framework"].version,
    "github.com/hashicorp/terraform-plugin-go":
      deps["github.com/hashicorp/terraform-plugin-go"].version,
    "github.com/hashicorp/terraform-plugin-log":
      deps["github.com/hashicorp/terraform-plugin-log"].version,
    // Transitive security floors for modules pulled in by the Terraform dependencies.
    "golang.org/x/crypto": deps["golang.org/x/crypto"].version,
    "golang.org/x/net": deps["golang.org/x/net"].version,
    "google.golang.org/grpc": deps["google.golang.org/grpc"].version,
  };

  if (isFeatureUsed("decimal")) {
    defaultDependencies["github.com/ericlagergren/decimal"] =
      deps["github.com/ericlagergren/decimal"].version;
  }

  if (isFeatureUsed("pagination")) {
    defaultDependencies["github.com/spyzhov/ajson"] =
      deps["github.com/spyzhov/ajson"].version;
  }

  if (isFeatureUsed("transformJq")) {
    defaultDependencies["github.com/itchyny/gojq"] =
      deps["github.com/itchyny/gojq"].version;
  }

  if (
    isFeatureUsed("typeOverrides") ||
    isFeatureUsed("contentMediaTypeApplicationJSON")
  ) {
    // jsontypes.Normalized is used for "any" types and contentMediaType application/json.
    defaultDependencies[
      "github.com/hashicorp/terraform-plugin-framework-jsontypes"
    ] =
      deps["github.com/hashicorp/terraform-plugin-framework-jsontypes"].version;
  }

  if (isFeatureUsed("formatBinary")) {
    // base64types.Standard is used for format: binary fields.
    defaultDependencies[
      "github.com/speakeasy-api/terraform-plugin-framework-base64types"
    ] =
      deps[
        "github.com/speakeasy-api/terraform-plugin-framework-base64types"
      ].version;
  }

  return renderDependencies(
    defaultDependencies,
    context.Global.Config.AdditionalDependencies,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

function renderDependencies(
  defaultDependencies: Record<string, string>,
  additionalDependencies: Record<string, string>,
): string {
  const dependencies = {
    ...defaultDependencies,
    ...additionalDependencies,
  };

  let renderedDependencies = "";

  for (const dependency of Object.keys(dependencies).sort()) {
    renderedDependencies += `${dependency} ${dependencies[dependency]}\n`;
  }
  renderedDependencies = renderedDependencies.replace(/\n$/, "");

  return renderedDependencies;
}
