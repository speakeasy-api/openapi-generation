/**
 * Describes a Terraform action for templating.
 */
type ActionContext = {
  /**
   * Action description.
   */
  Action: TerraformAction;

  /**
   * Provider context for the action, which includes various information
   * about the provider SDK, Configure data, and other relevant details.
   */
  ProviderContext: ProviderResourceContext;
};

/**
 * Describes a Terraform data resource for templating.
 */
type DataResourceContext = {
  /**
   * Data resource description.
   *
   * NOTE: This naming convention preserves accessors in various templating
   * files. If/when this should be updated to DataResource or similar for
   * consistency, a broader rename effort should be coordinated.
   */
  DataSource: TerraformDataResource;

  /**
   * Provider context for the data resource, which includes various information
   * about the provider SDK, Configure data, and other relevant details.
   */
  ProviderContext: ProviderResourceContext;
};

/**
 * Describes a Terraform ephemeral resource for templating.
 */
type EphemeralResourceContext = {
  /**
   * Ephemeral resource description.
   */
  EphemeralResource: TerraformEphemeralResource;

  /**
   * Provider context for the ephemeral resource, which includes various
   * information about the provider SDK, Configure data, and other relevant
   * details.
   */
  ProviderContext: ProviderResourceContext;
};

/**
 * Describes a Terraform managed resource for templating.
 */
type ManagedResourceContext = {
  /**
   * Managed resource description.
   *
   * NOTE: This naming convention preserves accessors in various templating
   * files. If/when this should be updated to ManagedResource or similar for
   * consistency and disambiguation with other "resource", a broader rename
   * effort should be coordinated.
   */
  Resource: TerraformManagedResource;

  /**
   * Provider context for the managed resource, which includes various
   * information about the provider SDK, Configure data, and other relevant
   * details.
   */
  ProviderContext: ProviderResourceContext;
};

/**
 * Templates action Configure() method provider data handling.
 */
function templateActionConfigureProviderData(
  resourceContext: ActionContext,
): string {
  const result: string[] = [`r.client = providerData.SDKClient`];

  const globalFields = resourceContext.Action.GlobalFields;

  Object.keys(globalFields).forEach((fieldName) => {
    result.push(`r.${fieldName} = providerData.${fieldName}`);
  });

  return result.sort().join("\n");
}

registerTemplateFunc(
  "templateActionConfigureProviderData",
  templateActionConfigureProviderData,
);

/**
 * Templates action operation (Invoke method) globals logic.
 */
function templateActionOperationGlobals(
  resourceContext: ActionContext,
): string {
  const globalFields = resourceContext.Action.GlobalFields;
  const result: string[] = [];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      result.push(
        ``,
        `if (data.${fieldName}.IsNull() || data.${fieldName}.IsUnknown()) && !r.${fieldName}.IsUnknown() {`,
        `data.${fieldName} = r.${fieldName}`,
        `}`,
      );
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateActionOperationGlobals",
  templateActionOperationGlobals,
);

/**
 * Templates action struct fields.
 */
function templateActionStructFields(resourceContext: ActionContext): string {
  const globalFields = resourceContext.Action.GlobalFields;
  const sdkType = resourceContext.ProviderContext.SDKType;
  const result: string[] = [
    `// Provider configured SDK client.`,
    `client ${sdkType}`,
  ];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      const fieldDef = globalFields[fieldName];
      const description = sanitizeComments(fieldDef.Comments?.Description)
        .split("\n")
        .map((line) => `// ${line}`)
        .join("\n");
      const structField =
        FrameworkTypeFromFieldDef(fieldDef).templateDataModelStructField(
          fieldName,
        );

      result.push(``, description, structField);
    });

  return result.join("\n");
}

registerTemplateFunc("templateActionStructFields", templateActionStructFields);

/**
 * Templates data resource Configure() method provider data handling.
 */
function templateDataResourceConfigureProviderData(
  resourceContext: DataResourceContext,
): string {
  const result: string[] = [`r.client = providerData.SDKClient`];

  const globalFields = resourceContext.DataSource.GlobalFields;

  Object.keys(globalFields).forEach((fieldName) => {
    result.push(`r.${fieldName} = providerData.${fieldName}`);
  });

  return result.sort().join("\n");
}

registerTemplateFunc(
  "templateDataResourceConfigureProviderData",
  templateDataResourceConfigureProviderData,
);

/**
 * Templates data resource operation (Read method) globals logic.
 */
function templateDataResourceOperationGlobals(
  resourceContext: DataResourceContext,
): string {
  const globalFields = resourceContext.DataSource.GlobalFields;
  const result: string[] = [];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      result.push(
        ``,
        `if (data.${fieldName}.IsNull() || data.${fieldName}.IsUnknown()) && !r.${fieldName}.IsUnknown() {`,
        `data.${fieldName} = r.${fieldName}`,
        `}`,
      );
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateDataResourceOperationGlobals",
  templateDataResourceOperationGlobals,
);

/**
 * Templates data resource struct fields.
 */
function templateDataResourceStructFields(
  resourceContext: DataResourceContext,
): string {
  const globalFields = resourceContext.DataSource.GlobalFields;
  const sdkType = resourceContext.ProviderContext.SDKType;
  // Historically, the SDK field was named "client" and this was only preserved
  // to reduce code churn when globals were introduced. Since it is not
  // exported, it should be non-conflicting (although potentially confusing)
  // should a global be named "Client".
  const result: string[] = [
    `// Provider configured SDK client.`,
    `client ${sdkType}`,
  ];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      const fieldDef = globalFields[fieldName];
      const description = sanitizeComments(fieldDef.Comments?.Description)
        .split("\n")
        .map((line) => `// ${line}`)
        .join("\n");
      const structField =
        FrameworkTypeFromFieldDef(fieldDef).templateDataModelStructField(
          fieldName,
        );

      result.push(``, description, structField);
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateDataResourceStructFields",
  templateDataResourceStructFields,
);

/**
 * Templates ephemeral resource Configure() method provider data handling.
 */
function templateEphemeralResourceConfigureProviderData(
  resourceContext: EphemeralResourceContext,
): string {
  const result: string[] = [`r.client = providerData.SDKClient`];

  const globalFields = resourceContext.EphemeralResource.GlobalFields;

  Object.keys(globalFields).forEach((fieldName) => {
    result.push(`r.${fieldName} = providerData.${fieldName}`);
  });

  return result.sort().join("\n");
}

registerTemplateFunc(
  "templateEphemeralResourceConfigureProviderData",
  templateEphemeralResourceConfigureProviderData,
);

/**
 * Templates ephemeral resource operation (Read method) globals logic.
 */
function templateEphemeralResourceOperationGlobals(
  resourceContext: EphemeralResourceContext,
): string {
  const globalFields = resourceContext.EphemeralResource.GlobalFields;
  const result: string[] = [];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      result.push(
        ``,
        `if (data.${fieldName}.IsNull() || data.${fieldName}.IsUnknown()) && !r.${fieldName}.IsUnknown() {`,
        `data.${fieldName} = r.${fieldName}`,
        `}`,
      );
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateEphemeralResourceOperationGlobals",
  templateEphemeralResourceOperationGlobals,
);

/**
 * Templates ephemeral resource struct fields.
 */
function templateEphemeralResourceStructFields(
  resourceContext: EphemeralResourceContext,
): string {
  const globalFields = resourceContext.EphemeralResource.GlobalFields;
  const sdkType = resourceContext.ProviderContext.SDKType;
  // Historically, the SDK field was named "client" and this was only preserved
  // to reduce code churn when globals were introduced. Since it is not
  // exported, it should be non-conflicting (although potentially confusing)
  // should a global be named "Client".
  const result: string[] = [
    `// Provider configured SDK client.`,
    `client ${sdkType}`,
  ];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      const fieldDef = globalFields[fieldName];
      const description = sanitizeComments(fieldDef.Comments?.Description)
        .split("\n")
        .map((line) => `// ${line}`)
        .join("\n");
      const structField =
        FrameworkTypeFromFieldDef(fieldDef).templateDataModelStructField(
          fieldName,
        );

      result.push(``, description, structField);
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateEphemeralResourceStructFields",
  templateEphemeralResourceStructFields,
);

/**
 * Templates managed resource Configure() method provider data handling.
 */
function templateManagedResourceConfigureProviderData(
  resourceContext: ManagedResourceContext,
): string {
  const result: string[] = [`r.client = providerData.SDKClient`];

  const globalFields = resourceContext.Resource.GlobalFields;

  Object.keys(globalFields).forEach((fieldName) => {
    result.push(`r.${fieldName} = providerData.${fieldName}`);
  });

  return result.sort().join("\n");
}

registerTemplateFunc(
  "templateManagedResourceConfigureProviderData",
  templateManagedResourceConfigureProviderData,
);

/**
 * Templates managed resource operation (Create, Delete, ImportState, and Update
 * methods) globals logic.
 */
function templateManagedResourceOperationGlobals(
  resourceContext: ManagedResourceContext,
): string {
  const globalFields = resourceContext.Resource.GlobalFields;
  const result: string[] = [];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      result.push(
        ``,
        `if (data.${fieldName}.IsNull() || data.${fieldName}.IsUnknown()) && !r.${fieldName}.IsUnknown() {`,
        `data.${fieldName} = r.${fieldName}`,
        `}`,
      );
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateManagedResourceOperationGlobals",
  templateManagedResourceOperationGlobals,
);

/**
 * Templates managed resource struct fields.
 */
function templateManagedResourceStructFields(
  resourceContext: ManagedResourceContext,
): string {
  const globalFields = resourceContext.Resource.GlobalFields;
  const sdkType = resourceContext.ProviderContext.SDKType;
  // Historically, the SDK field was named "client" and this was only preserved
  // to reduce code churn when globals were introduced. Since it is not
  // exported, it should be non-conflicting (although potentially confusing)
  // should a global be named "Client".
  const result: string[] = [
    `// Provider configured SDK client.`,
    `client ${sdkType}`,
  ];

  Object.keys(globalFields)
    .sort()
    .forEach((fieldName) => {
      const fieldDef = globalFields[fieldName];
      const description = sanitizeComments(fieldDef.Comments?.Description)
        .split("\n")
        .map((line) => `// ${line}`)
        .join("\n");
      const structField =
        FrameworkTypeFromFieldDef(fieldDef).templateDataModelStructField(
          fieldName,
        );

      result.push(``, description, structField);
    });

  return result.join("\n");
}

registerTemplateFunc(
  "templateManagedResourceStructFields",
  templateManagedResourceStructFields,
);

function trace(obj1: TypeDef): string {
  if (obj1.Extensions.All["x-speakeasy-trace"]) {
    return `\ntrace=${JSON.stringify(
      Object.keys(obj1.Extensions.All["x-speakeasy-trace"]),
    )}`;
  }
  return "";
}

function handleTypeMismatch(
  hierarchy: string,
  obj1: TypeDef,
  obj2: TypeDef,
): TypeDef | undefined {
  const type1 = obj1.Type.toString();
  const type2 = obj2.Type.toString();
  if (type1 === type2) {
    return undefined;
  }
  if (
    (type1 === "class" && type2 === "union") ||
    (type1 === "union" && type2 === "class")
  ) {
    return undefined;
  }
  if (type1 === "enum") {
    return handleTypeMismatch(hierarchy + ".enum", obj1.Enum.Type, obj2);
  }
  if (type2 === "enum") {
    return handleTypeMismatch(hierarchy + ".enum", obj1, obj2.Enum.Type);
  }
  switch (type1) {
    case "float32":
      if (type2 === "integer" || type2 === "int32" || type2 === "bigint") {
        return obj2;
      }
      break;
    case "integer":
    case "int32":
    case "bigint":
      if (
        type2 === "integer" ||
        type2 === "int32" ||
        type2 === "bigint" ||
        type2 === "float32"
      ) {
        return obj1;
      }
      break;
    case "class":
      if (type2 === "any") {
        return obj1;
      }
      break;
  }
  throw new Error(
    `${hierarchy} impedance mismatch: ${prettyPrintType(
      obj1,
    )} != ${prettyPrintType(obj2)}${trace(obj1)}${trace(obj2)}`,
  );
}
