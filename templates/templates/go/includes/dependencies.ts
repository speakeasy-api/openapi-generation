// Load config.ts to make getTemplateDependencies available globally

// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

function getDefaultDependencies(): Record<string, string> {
  const defaultDependencies = {
    "github.com/spyzhov/ajson": deps["github.com/spyzhov/ajson"].version,
  };
  if (includeDecimal()) {
    defaultDependencies["github.com/ericlagergren/decimal"] =
      deps["github.com/ericlagergren/decimal"].version;
  }
  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    defaultDependencies["github.com/stretchr/testify"] =
      deps["github.com/stretchr/testify"].version;
  }
  if (hasClientCredentials() || hasOAuth2PasswordFlow()) {
    defaultDependencies["golang.org/x/sync"] =
      deps["golang.org/x/sync"].version;
  }

  if (isFeatureUsed("transformJq")) {
    defaultDependencies["github.com/itchyny/gojq"] =
      deps["github.com/itchyny/gojq"].version;
  }
  return defaultDependencies;
}

// @ts-ignore
function templateDependencies(): string {
  return renderDependencies(
    getDefaultDependencies(),
    context.Global.Config.AdditionalDependencies,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

// @ts-ignore
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
