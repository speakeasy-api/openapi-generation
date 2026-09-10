function generateSchemaVersion(resource: TerraformManagedResource): string {
  if (!resource.Schema?.Version) {
    return "";
  }

  return `Version: ${resource.Schema.Version},\n`;
}

registerTemplateFunc("generateSchemaVersion", generateSchemaVersion);

function generateExpectedInterface(resource: TerraformManagedResource): string {
  // TODO: Expected interfaces should be handled per feature, not conditionally.
  if (!resource.Schema?.Version) {
    return "ResourceWithImportState";
  }

  return "ResourceWithUpgradeState";
}
registerTemplateFunc("generateExpectedInterface", generateExpectedInterface);

function boilerplateStateUpgrader(
  entity: TerraformManagedResource,
  stateVersion: number,
): string {
  const filePath = `internal/stateupgraders/${entity.Name.toLowerCase()}_v${stateVersion}.go`;
  addUntrackedPattern(
    `^internal/stateupgraders/${entity.Name.toLowerCase()}_v${stateVersion}\\.go`,
  );

  let existingData = readFile(filePath);
  const funcName = `${entity.Name.toLowerCase()}_StateUpgrader_V${stateVersion}`;

  // If we don't already have a file, create it
  if (!existingData) {
    templateFile(`boilerplate/stateupgrader.go.stmpl`, filePath, {
      FuncName: sanitizeClassName(funcName),
      Resource: entity,
      StateVersion: stateVersion,
    });
  }
  addGenImport(
    `${getRootPackage()}/internal/stateupgraders`,
    false,
    `stateupgraders`,
  );

  return `stateupgraders.${sanitizeClassName(funcName)}`;
}

function generateUpgradeState(resource: TerraformManagedResource): string {
  if (!resource.Schema?.Version) {
    return "";
  }

  let subBuilder = "\n";
  for (let i = 0; i < resource.Schema.Version; i++) {
    subBuilder += `${i}: {StateUpgrader: ${boilerplateStateUpgrader(
      resource,
      i,
    )}},\n`;
  }
  return `
func (r *${sanitizeClassName(
    resource.Name,
  )}Resource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
  	return map[int64]resource.StateUpgrader{${subBuilder}}
}
`;
}
registerTemplateFunc("generateUpgradeState", generateUpgradeState);
