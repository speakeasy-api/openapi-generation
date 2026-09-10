// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

/** Templates runtime and development dependencies into the gemspec. */
//@ts-ignore
function templateDependencies(
  additionalDependencies: Record<string, Record<string, string>>,
): string {
  const defaultRuntimeDependencies: Record<string, string | string[]> = {
    base64: [">= 0.2.0", "< 1.0"],
    faraday: deps.faraday.version,
    "faraday-multipart": deps["faraday-multipart"].version,
    "faraday-retry": deps["faraday-retry"].version,
  };

  if (isFeatureUsed("pagination")) {
    defaultRuntimeDependencies["janeway-jsonpath"] =
      deps["janeway-jsonpath"].version;
  }

  if (context.Global.Config.TypingStrategy === "sorbet") {
    defaultRuntimeDependencies["sorbet-runtime"] =
      deps["sorbet-runtime"].version;
  }

  const runtimeDependencies: Record<string, string | string[]> = {
    ...defaultRuntimeDependencies,
    ...additionalDependencies.runtime,
  };

  const developmentDependencies: Record<string, string | string[]> = {
    minitest: deps.minitest.version,
    "minitest-focus": deps["minitest-focus"].version,
    "minitest-reporters": deps["minitest-reporters"].version,
    rubocop: deps.rubocop.version,
    "rubocop-minitest": deps["rubocop-minitest"].version,
    rake: deps.rake.version,
    ...additionalDependencies.development,
  };

  if (context.Global.Config.TypingStrategy === "sorbet") {
    developmentDependencies["irb"] = deps.irb.version;
    developmentDependencies["sorbet"] = deps.sorbet.version;
    developmentDependencies["tapioca"] = deps.tapioca.version;
    developmentDependencies["tsort"] = deps.tsort.version;
  }
  const result: string[] = [];

  for (const dependency of Object.keys(runtimeDependencies).sort()) {
    const version = runtimeDependencies[dependency];

    if (version === "") {
      result.push(`s.add_dependency('${dependency}')`);
      continue;
    }

    if (Array.isArray(version)) {
      result.push(
        `s.add_dependency('${dependency}', ${version
          .map((v) => `'${v}'`)
          .join(", ")})`,
      );
    } else {
      result.push(`s.add_dependency('${dependency}', '${version}')`);
    }
  }

  for (const dependency of Object.keys(developmentDependencies).sort()) {
    const version = developmentDependencies[dependency];

    if (version === "") {
      result.push(`s.add_development_dependency('${dependency}')`);
      continue;
    }

    if (Array.isArray(version)) {
      result.push(
        `s.add_development_dependency('${dependency}', ${version
          .map((v) => `'${v}'`)
          .join(", ")})`,
      );
    } else {
      result.push(
        `s.add_development_dependency('${dependency}', '${version}')`,
      );
    }
  }

  return result.join("\n  ");
}

registerTemplateFunc("templateDependencies", templateDependencies);
