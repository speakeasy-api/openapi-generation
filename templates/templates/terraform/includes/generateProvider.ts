/** Returns the Provider DataSources method contents with the instantiation
 *  functions of all generated and custom data resources (via gen.yaml
 *  additionalDataSources configuration). */
function generateDataSources(datasources: TerraformEntity[]): string {
  const result: string[] = [];

  for (const datasource of datasources) {
    const datasourceFunc = `New${sanitizeFieldName(datasource.Name)}DataSource`;
    result.push(datasourceFunc);
  }

  const additionalDataSources: {
    importLocation?: string;
    importAlias?: string;
    datasource: string;
  }[] = context.Global.Config["AdditionalDataSources"] || [];
  for (const additionalDataSource of additionalDataSources) {
    if (additionalDataSource.importLocation) {
      addGenImport(
        additionalDataSource.importLocation,
        false,
        additionalDataSource.importAlias,
      );
    }

    result.push(additionalDataSource.datasource);
  }

  return result.map((line) => `${line},\n`).join("");
}
registerTemplateFunc("generateDataSources", generateDataSources);

/** Returns the Provider Resources method contents with the instantiation
 *  functions of all generated and custom managed resources (via gen.yaml
 *  additionalResources configuration). */
function generateResources(resources: TerraformEntity[]): string {
  const result: string[] = [];

  for (const resource of resources) {
    const resourceFunc = `New${sanitizeFieldName(resource.Name)}Resource`;
    result.push(resourceFunc);
  }

  const additionalResources: {
    importLocation?: string;
    importAlias?: string;
    resource: string;
  }[] = context.Global.Config["AdditionalResources"] || [];
  for (const additionalResource of additionalResources) {
    if (additionalResource.importLocation) {
      addGenImport(
        additionalResource.importLocation,
        false,
        additionalResource.importAlias,
      );
    }

    result.push(additionalResource.resource);
  }

  return result.map((line) => `${line},\n`).join("");
}
registerTemplateFunc("generateResources", generateResources);

/** Returns the Provider EphemeralResources method contents with the
 *  instantiation functions of all custom ephemeral resources (via gen.yaml
 *  additionalEphemeralResources configuration). */
function generateEphemeralResources(
  ephemeralResources: TerraformEntity[],
): string {
  const result: string[] = [];

  for (const ephemeralResource of ephemeralResources) {
    const ephemeralResourceFunc = `New${sanitizeFieldName(
      ephemeralResource.Name,
    )}EphemeralResource`;
    result.push(ephemeralResourceFunc);
  }

  const additionalEphemeralResources: {
    importLocation?: string;
    importAlias?: string;
    resource: string;
  }[] = context.Global.Config["AdditionalEphemeralResources"] || [];
  for (const additionalEphemeralResource of additionalEphemeralResources) {
    if (additionalEphemeralResource.importLocation) {
      addGenImport(
        additionalEphemeralResource.importLocation,
        false,
        additionalEphemeralResource.importAlias,
      );
    }

    result.push(additionalEphemeralResource.resource);
  }

  return result
    .sort((a, b) => a.localeCompare(b))
    .map((line) => `${line},\n`)
    .join("");
}

registerTemplateFunc("generateEphemeralResources", generateEphemeralResources);

/** Returns the Provider ListResources method contents with the
 *  instantiation functions of all custom list resources (via gen.yaml
 *  additionalListResources configuration). */
function generateListResources(_listResources: TerraformEntity[]): string {
  const result: string[] = [];

  // TODO: enable once custom list resources are supported
  // for (const listResource of listResources) {
  //   const listResourceFunc = `New${sanitizeFieldName(listResource.Name)}ListResource`;
  //   result.push(listResourceFunc);
  // }

  const additionalListResources: {
    importLocation?: string;
    importAlias?: string;
    resource: string;
  }[] = context.Global.Config["AdditionalListResources"] || [];
  for (const additionalListResource of additionalListResources) {
    if (additionalListResource.importLocation) {
      addGenImport(
        additionalListResource.importLocation,
        false,
        additionalListResource.importAlias,
      );
    }

    result.push(additionalListResource.resource);
  }

  return result.map((line) => `${line},\n`).join("");
}

registerTemplateFunc("generateListResources", generateListResources);

function generateFunctions(_functions: TerraformEntity[]): string {
  const results: string[] = [];

  // TODO: When support is added for generated functions
  // for (const fn of functions) {
  //   const func = `New${sanitizeFieldName(fn.Name)}Function`;
  //   results.push(func);
  // }

  const additionalFunctions: {
    importLocation?: string;
    importAlias?: string;
    function: string;
  }[] = context.Global.Config["AdditionalFunctions"] || [];
  for (const additionalFunction of additionalFunctions) {
    if (additionalFunction.importLocation) {
      addGenImport(
        additionalFunction.importLocation,
        false,
        additionalFunction.importAlias,
      );
    }

    results.push(additionalFunction.function);
  }

  return results.map((line) => `${line},\n`).join("");
}

registerTemplateFunc("generateFunctions", generateFunctions);
/** Returns the Provider Actions method contents with the
 *  instantiation functions of all custom actions (via gen.yaml
 *  additionalActions configuration). */
function generateActions(actions: TerraformEntity[]): string {
  const result: string[] = [];

  for (const action of actions) {
    const actionFunc = `New${sanitizeFieldName(action.Name)}Action`;
    result.push(actionFunc);
  }

  const additionalActions: {
    importLocation?: string;
    importAlias?: string;
    action: string;
  }[] = context.Global.Config["AdditionalActions"] || [];
  for (const additionalAction of additionalActions) {
    if (additionalAction.importLocation) {
      addGenImport(
        additionalAction.importLocation,
        false,
        additionalAction.importAlias,
      );
    }

    result.push(additionalAction.action);
  }

  return result.map((line) => `${line},\n`).join("");
}

registerTemplateFunc("generateActions", generateActions);
