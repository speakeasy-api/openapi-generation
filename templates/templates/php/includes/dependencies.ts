// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

function templateAdditionalDependencies(
  dependencies: Record<string, string>,
): string[] {
  return Object.entries(dependencies).map(
    ([key, value]) => `"${key.replaceAll("\\", "\\\\")}": "${value}"`,
  );
}

//@ts-ignore
function templateAutoload(): string {
  const autoloads = [
    `"${escapeNamespace(context.Global.Config.Namespace)}\\\\": "src/"`,
  ];

  if (context.Global.Config.AdditionalDependencies?.["autoload"]) {
    autoloads.push(
      ...templateAdditionalDependencies(
        context.Global.Config.AdditionalDependencies["autoload"],
      ),
    );
  }

  return autoloads.join(",\n");
}
registerTemplateFunc("templateAutoload", templateAutoload);

//@ts-ignore
function templateAutoloadDev(): string {
  const autoloads: string[] = [];

  if (context.Global.AST.MainSDK.OutputTests) {
    autoloads.push(
      `"${escapeNamespace(
        context.Global.Config.Namespace,
      )}\\\\Tests\\\\": "Tests/"`,
    );
  }

  if (context.Global.Config.AdditionalDependencies?.["autoload-dev"]) {
    autoloads.push(
      ...templateAdditionalDependencies(
        context.Global.Config.AdditionalDependencies["autoload-dev"],
      ),
    );
  }

  return autoloads.join(",\n");
}
registerTemplateFunc("templateAutoloadDev", templateAutoloadDev);

//@ts-ignore
function templateDependencies(): string {
  const depsList = [
    `"php": "${deps.php.version}"`,
    `"galbar/jsonpath": "${deps["galbar/jsonpath"].version}"`,
    `"guzzlehttp/guzzle": "${deps["guzzlehttp/guzzle"].version}"`,
    `"speakeasy/serializer": "${deps["speakeasy/serializer"].version}"`,
    `"brick/date-time": "${deps["brick/date-time"].version}"`,
    `"phpdocumentor/type-resolver": "${deps["phpdocumentor/type-resolver"].version}"`,
    `"brick/math": "${deps["brick/math"].version}"`,
  ];

  if (context.Global.Config.AdditionalDependencies?.["require"]) {
    depsList.push(
      ...templateAdditionalDependencies(
        context.Global.Config.AdditionalDependencies["require"],
      ),
    );
  }

  return depsList.join(",\n");
}
registerTemplateFunc("templateDependencies", templateDependencies);

//@ts-ignore
function templateDevDependencies(): string {
  const devDepsList = [
    `"laravel/pint": "${deps["laravel/pint"].version}"`,
    `"phpstan/phpstan": "${deps["phpstan/phpstan"].version}"`,
    `"phpunit/phpunit": "${deps["phpunit/phpunit"].version}"`,
    `"roave/security-advisories": "${deps["roave/security-advisories"].version}"`,
  ];
  if (context.Global.Config.LaravelServiceProvider.enabled) {
    devDepsList.push(
      `"orchestra/testbench": "${deps["orchestra/testbench"].version}"`,
    );
  }

  if (context.Global.Config.AdditionalDependencies?.["require-dev"]) {
    devDepsList.push(
      ...templateAdditionalDependencies(
        context.Global.Config.AdditionalDependencies["require-dev"],
      ),
    );
  }

  return devDepsList.join(",\n");
}
registerTemplateFunc("templateDevDependencies", templateDevDependencies);
