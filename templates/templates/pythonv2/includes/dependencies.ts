// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

// @ts-ignore
function templateDependencies(): string {
  const defaultDependencies: Record<string, string> = {
    pydantic: deps.pydantic.version,
  };

  // httpcore is httpx's transport and is pinned only to hold a CVE floor;
  // httpx2 depends on its own httpcore2 fork so the pin does not apply.
  if (pythonHttpClientLibrary() === "httpx2") {
    defaultDependencies["httpx2"] = deps.httpx2.version;
  } else {
    defaultDependencies["httpx"] = deps.httpx.version;
    defaultDependencies["httpcore"] = deps.httpcore.version;
  }

  if (isFeatureUsed("pagination")) {
    defaultDependencies["jsonpath-python"] = deps["jsonpath-python"].version;
  }

  return indentLines(
    renderPythonProjectDependencies(
      defaultDependencies,
      context.Global.Config.AdditionalDependencies?.main,
    ),
    0,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

function templateAdditionalDependencies(): string {
  const devDependencies: Record<string, string> = {
    pylint: deps.pylint.version,
    mypy: deps.mypy.version,
    pyright: deps.pyright.version,
  };

  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    devDependencies["pytest"] = deps.pytest.version;
    devDependencies["pytest-xdist"] = deps["pytest-xdist"].version;
    devDependencies["pytest-asyncio"] = deps["pytest-asyncio"].version;

    if (
      context.Global.Config.PytestTimeout &&
      context.Global.Config.PytestTimeout > 0
    ) {
      devDependencies["pytest-timeout"] = deps["pytest-timeout"].version;
    }
  }

  const packageManager = context.Global.Config.PackageManager || "uv";

  if (packageManager === "poetry") {
    let lines = renderPythonDependencies(
      devDependencies,
      context.Global.Config.AdditionalDependencies?.dev ?? {},
    );

    for (const group of Object.keys(
      context.Global.Config.AdditionalDependencies ?? {},
    )) {
      if (group == "main" || group == "dev") {
        continue;
      }

      lines.push("");

      lines.push(`[tool.poetry.group.${group}.dependencies]`);
      lines = lines.concat(
        renderPythonDependencies(
          context.Global.Config.AdditionalDependencies?.[group] ?? {},
          {},
        ),
      );
    }
    return indentLines(lines, 0);
  } else {
    let lines: string[] = [];

    // Add dev dependencies
    const devDeps = renderPythonProjectDependencies(
      devDependencies,
      context.Global.Config.AdditionalDependencies?.dev ?? {},
    );
    lines.push("dev = [");
    lines = lines.concat(devDeps);
    lines.push("]");

    // Add additional dependency groups
    for (const group of Object.keys(
      context.Global.Config.AdditionalDependencies ?? {},
    )) {
      if (group == "main" || group == "dev") {
        continue;
      }

      lines.push("");
      const groupDeps = renderPythonProjectDependencies(
        context.Global.Config.AdditionalDependencies?.[group] ?? {},
        {},
      );
      lines.push(`${group} = [`);
      lines = lines.concat(groupDeps);
      lines.push("]");
    }
    return indentLines(lines, 0);
  }
}
registerTemplateFunc(
  "templateAdditionalDependencies",
  templateAdditionalDependencies,
);

function templateOptionalDependencies(): string {
  const optionalDeps = context.Global.Config.OptionalDependencies ?? {};
  const extrasNames = Object.keys(optionalDeps).sort();

  if (extrasNames.length === 0) {
    return "";
  }

  let lines: string[] = [];
  lines.push("[project.optional-dependencies]");

  extrasNames.forEach((extrasName) => {
    const deps = renderPythonProjectDependencies(
      optionalDeps[extrasName] ?? {},
      {},
    );

    if (deps.length === 0) {
      lines.push(`${extrasName} = []`);
    } else {
      lines.push(`${extrasName} = [`);
      lines = lines.concat(deps);
      lines.push("]");
    }
  });

  return "\n" + indentLines(lines, 0) + "\n";
}
registerTemplateFunc(
  "templateOptionalDependencies",
  templateOptionalDependencies,
);

/** Converts a dependency specification to a PEP 508 compatible string. This is
 *  a compatibility function for projects that may have been previously using
 *  poetry dependency specifiers which are not compatible with PEP 508, such as
 *  carat (^) and tilde (~) character prefixes. Additional information can be
 *  found at https://python-poetry.org/docs/dependency-specification/ and
 *  https://peps.python.org/pep-0508/ */
function convertDependencySpecificationToPEP508(original: string): string {
  if (original.startsWith("^")) {
    return convertCaratDependencySpecificationToPEP508(original);
  }

  if (original.startsWith("~") && !original.startsWith("~=")) {
    return convertTildeDependencySpecificationToPEP508(original);
  }

  return original;
}

/** Converts a carat (^) character prefix dependency specification to a PEP 508
 *  compatible string ((>=X.Y.Z <A.B.C)). */
function convertCaratDependencySpecificationToPEP508(original: string): string {
  const parts = original.substring(1).split(".");
  const major = parseInt(parts[0]);
  const minor = parseInt(parts[1] ?? "0");
  const patch = parseInt(parts[2] ?? "0");

  // ^X.Y.Z or ^0
  if (major >= 1 || !parts[1]) {
    return `(>=${major}.${minor}.${patch},<${major + 1}.0.0)`;
  }

  // ^0.Y
  if (!parts[2]) {
    return `(>=${major}.${minor}.${patch},<0.${minor + 1}.0)`;
  }

  // ^0.Y.Z
  if (minor >= 1) {
    return `(>=${major}.${minor}.${patch},<0.${minor + 1}.0)`;
  }

  // ^0.0.Z
  return `(>=${major}.${minor}.${patch},<0.0.${patch + 1})`;
}

/** Converts a tilde (~) character prefix dependency specification to a PEP 508
 *  compatible string, such as (>=X.Y.Z <A.B.C). Do not pass tilde equal (~=)
 *  prefixes as they have a different meaning. */
function convertTildeDependencySpecificationToPEP508(original: string): string {
  const parts = original.substring(1).split(".");
  const major = parseInt(parts[0]);
  const minor = parseInt(parts[1] ?? "0");
  const patch = parseInt(parts[2] ?? "0");

  // ~X
  if (!parts[1]) {
    return `(>=${major}.${minor}.${patch},<${major + 1}.0.0)`;
  }

  // ~X.Y or ~X.Y.Z
  return `(>=${major}.${minor}.${patch},<${major}.${minor + 1}.0)`;
}

/** Renders the project.dependencies field in pyproject.toml. Entries must
 *  conform to PEP 508 strings. Additional information can be found at
 *  https://python-poetry.org/docs/dependency-specification/ and
 *  https://peps.python.org/pep-0508/ */
function renderPythonProjectDependencies(
  defaultDependencies: Record<string, string>,
  additionalDependencies: Record<string, string>,
): string[] {
  const dependencies = deepMerge(
    {},
    defaultDependencies,
    additionalDependencies,
  );

  let lines: string[] = [];

  for (const dependency of Object.keys(dependencies).sort()) {
    const dependencySpecifier = convertDependencySpecificationToPEP508(
      dependencies[dependency],
    );

    lines.push(`    "${dependency} ${dependencySpecifier}",`);
  }

  return lines;
}

function renderPythonDependencies(
  defaultDependencies: Record<string, string>,
  additionalDependencies: Record<string, string>,
): string[] {
  const dependencies = deepMerge(
    {},
    defaultDependencies,
    additionalDependencies,
  );

  let lines = [];

  for (const dependency of Object.keys(dependencies).sort()) {
    lines.push(`${dependency} = "${dependencies[dependency]}"`);
  }

  return lines;
}

function isObject(item: any) {
  return item && typeof item === "object" && !Array.isArray(item);
}

function deepMerge(target: any, ...sources: any[]): any {
  if (!sources.length) return target;
  const source = sources.shift();

  if (isObject(target) && isObject(source)) {
    for (const key in source) {
      if (isObject(source[key])) {
        if (!target[key]) Object.assign(target, { [key]: {} });
        deepMerge(target[key], source[key]);
      } else {
        Object.assign(target, { [key]: source[key] });
      }
    }
  }
  return deepMerge(target, ...sources);
}
