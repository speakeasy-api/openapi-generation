// @ts-ignore
function genImports(): string {
  return `{{ templateImports .RecursiveComputed.Imports }}`;
}
registerTemplateFunc("genImports", genImports);

// @ts-ignore
function templateImports(imports: string[]): string {
  if (!imports || imports.length == 0) {
    return "";
  }

  return `import(\n\t${imports.join("\n\t")}\n)`;
}
registerTemplateFunc("templateImports", templateImports);

/**
 * Represents a Go import with an optional alias for use with addGenImport().
 */
type GoImport = {
  Path: string;
  Alias?: string;
};

// @ts-ignore
function addGenImport(
  imp: string,
  local: boolean = false,
  alias: string = "",
): string {
  if (local) {
    if (!imp) {
      return "";
    }

    imp = `${getSDKPackage()}/${imp}`;
  }
  imp = `"${imp}"`;
  if (alias) {
    imp = `${alias} ${imp}`;
  }

  // Temporary check while we migrate to new imports
  if (context.RecursiveComputed) {
    let imports = context.RecursiveComputed.Imports;
    if (!imports) {
      imports = [];
    }

    if (
      !imports.find((i) => {
        return i === imp;
      })
    ) {
      imports.push(imp);
    }

    context.RecursiveComputed.Imports = imports;
  }

  return ""; // Still needs to return a string if used in templates
}
registerTemplateFunc("addGenImport", addGenImport);
