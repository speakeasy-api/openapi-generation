/** Describes a generated data model file (e.g.
 *  internal/provider/types/example.go). These files are generated for class
 *  TypeDef and contain the Go struct type definition with tfsdk struct field
 *  tags necessary for terraform-plugin-framework. */
type GeneratedDataModelTypeFile = {
  /** Name of the generated type file, typically based on an assigned symbol in
   *  the symbol manager or inline field name from another class. */
  filename: string;

  /** The content of the generated type file. */
  content: string;

  /** Imports for the generated type file contents. */
  imports: string[];
};

function generatePrimitiveType(typedef: TypeDef) {
  switch (String(typedef.Type)) {
    case "string":
      return "types.String";
    case "float32":
      return "types.Float32";
    case "int32":
      return "types.Int32";
    case "integer":
      return "types.Int64";
    case "number":
      return "types.Float64";
    case "boolean":
      return "types.Bool";
    case "date":
    case "date-time":
      return "types.String";
  }
  return null;
}

/** Template data model field type and additional files for a TypeDef. */
function templateTypeDefDataModelFieldType(
  typedef: TypeDef,
  fieldName: string,
  optional: boolean,
  indent: number,
  packageName: string,
  hierarchy: string,
  getSymbol: (string, TypeDef) => string,
  isCollectionItem: boolean = false,
): {
  content: string;
  imports: string[];
  additionalFiles: GeneratedDataModelTypeFile[];
} {
  // TODO: Replace function with templateStructFieldValueType() and generated
  // tfTypes templating method when all types are converted to FrameworkType.
  switch (typedef.Type.toString()) {
    case "any":
    case "boolean":
    case "bytes":
    case "date":
    case "date-time":
    case "enum":
    case "float32":
    case "int32":
    case "integer":
    case "number":
    case "string":
      const primitiveFrameworkType = FrameworkTypeFromTypeDef(typedef);

      return {
        content: primitiveFrameworkType.templateStructFieldValueType(),
        imports: primitiveFrameworkType.valueTypeImports(),
        additionalFiles: [],
      };
    case "map":
      const customTypeConfig = TypeDefTerraformCustomTypeConfig(typedef);

      if (customTypeConfig) {
        const frameworkType = FrameworkTypeFromTypeDef(typedef);

        return {
          content: frameworkType.templateStructFieldValueType(),
          imports: frameworkType.valueTypeImports(),
          additionalFiles: [],
        };
      }

      const innerType = templateTypeDefDataModelFieldType(
        typedef.ItemType,
        "",
        false,
        indent,
        packageName,
        `${hierarchy}.[]`,
        getSymbol,
        true, // collection item
      );
      return {
        content: "map[string]" + innerType.content,
        imports: innerType.imports,
        additionalFiles: innerType.additionalFiles,
      };
  }

  if (typedef.Type == "class") {
    const selfName = getSymbol(fieldName, typedef);
    const additionalFile: GeneratedDataModelTypeFile = {
      filename: selfName,
      content: "",
      imports: [],
    };
    additionalFile.content += `type ${selfName} struct {\n`;
    // Class fields are pointers (for nil tracking), except collection items which are values
    let subBuilder = isCollectionItem ? "" : "*";
    let otherAdditionalFiles = [];
    for (const field of typedef.Fields) {
      // Skip fields with path aliases only
      if (field.Type.Extensions?.MatchConfig?.Path) {
        continue;
      }
      const innerType = templateTypeDefDataModelFieldType(
        field.Type,
        field.Name,
        field.Optional || field.Nullable,
        1,
        "types",
        `${hierarchy}.${fieldName}`,
        getSymbol,
      );
      otherAdditionalFiles = otherAdditionalFiles.concat(
        innerType.additionalFiles,
      );
      additionalFile.content +=
        templateIndent(1) +
        sanitizeFieldName(field.Name) +
        " " +
        innerType.content +
        ` ${generateStructFieldTag(field)}\n`;
      additionalFile.imports.push(...innerType.imports);
    }
    additionalFile.content += "}\n";
    if (packageName == "provider") {
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );
      subBuilder += "tfTypes.";
    }
    subBuilder += `${selfName}`;
    return {
      content: subBuilder,
      // TODO: Aliased imports, e.g. tfTypes
      imports: [],
      additionalFiles: otherAdditionalFiles.concat(additionalFile),
    };
  }
  if (isGoArray(typedef)) {
    const customTypeConfig = TypeDefTerraformCustomTypeConfig(typedef);

    if (customTypeConfig) {
      const frameworkType = FrameworkTypeFromTypeDef(typedef);

      return {
        content: frameworkType.templateStructFieldValueType(),
        imports: frameworkType.valueTypeImports(),
        additionalFiles: [],
      };
    }

    if (generatePrimitiveType(typedef.ItemType)) {
      return {
        content: `[]${generatePrimitiveType(typedef.ItemType)}`,
        imports: ["github.com/hashicorp/terraform-plugin-framework/types"],
        additionalFiles: [],
      };
    } else if (
      typedef.ItemType.Type == "class" &&
      typedef.ItemType.Fields &&
      typedef.ItemType.Fields.length
    ) {
      const selfName = getSymbol(fieldName, typedef.ItemType);
      const additionalFile: GeneratedDataModelTypeFile = {
        filename: selfName,
        content: "",
        imports: [],
      };
      additionalFile.content += `type ${selfName} struct {\n`;

      let otherAdditionalFiles = [];
      for (const field of typedef.ItemType.Fields) {
        // Skip fields with path aliases only
        if (field.Type.Extensions?.MatchConfig?.Path) {
          continue;
        }
        const innerType = templateTypeDefDataModelFieldType(
          field.Type,
          field.Name,
          field.Optional || field.Nullable,
          1,
          "types",
          `${hierarchy}.[].${fieldName}`,
          getSymbol,
        );
        otherAdditionalFiles = otherAdditionalFiles.concat(
          innerType.additionalFiles,
        );
        additionalFile.content +=
          templateIndent(1) +
          sanitizeFieldName(field.Name) +
          " " +
          innerType.content +
          ` ${generateStructFieldTag(field)}\n`;
        additionalFile.imports.push(...innerType.imports);
      }
      additionalFile.content += "}\n";
      let prefix = "";
      if (packageName == "provider") {
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
        prefix = "tfTypes.";
      }

      return {
        content: `[]${prefix}${selfName}`,
        // TODO: Aliased imports, e.g. tfTypes
        imports: [],
        additionalFiles: otherAdditionalFiles.concat(additionalFile),
      };
    } else if (typedef.ItemType.AssociatedTypes?.length) {
      const selfName = getSymbol(fieldName, typedef.ItemType);
      const additionalFile: GeneratedDataModelTypeFile = {
        filename: selfName,
        content: "",
        imports: [],
      };
      additionalFile.content += `type ${selfName} struct {\n`;
      let otherAdditionalFiles = [];
      for (const subtype of typedef.ItemType.AssociatedTypes) {
        const unionName = sanitizeTFUnionTypeName(typedef.ItemType, subtype);
        const innerType = templateTypeDefDataModelFieldType(
          subtype,
          unionName,
          true,
          1,
          "types",
          `${hierarchy}.[].${fieldName}`,
          getSymbol,
        );
        otherAdditionalFiles = otherAdditionalFiles.concat(
          innerType.additionalFiles,
        );
        additionalFile.content +=
          templateIndent(1) +
          sanitizeFieldName(unionName) +
          " " +
          innerType.content +
          ` ${generateUnionStructFieldTag(typedef.ItemType, subtype)}\n`;
        additionalFile.imports.push(...innerType.imports);
      }
      // Add hoisted fields for list/set item union types
      for (const field of typedef.ItemType.Fields) {
        // Skip fields with path aliases only
        if (field.Type.Extensions?.MatchConfig?.Path) {
          continue;
        }
        const innerType = templateTypeDefDataModelFieldType(
          field.Type,
          field.Name,
          field.Optional || field.Nullable,
          1,
          "types",
          `${hierarchy}.[].${fieldName}`,
          getSymbol,
        );
        otherAdditionalFiles = otherAdditionalFiles.concat(
          innerType.additionalFiles,
        );
        additionalFile.content +=
          templateIndent(1) +
          sanitizeFieldName(field.Name) +
          " " +
          innerType.content +
          ` ${generateStructFieldTag(field)}\n`;
        additionalFile.imports.push(...innerType.imports);
      }
      additionalFile.content += "}\n";
      let prefix = "";
      if (packageName == "provider") {
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
        prefix = "tfTypes.";
      }

      return {
        content: `[]${prefix}${selfName}`,
        // TODO: Aliased imports, e.g. tfTypes
        imports: [],
        additionalFiles: otherAdditionalFiles.concat(additionalFile),
      };
    } else if (isGoArray(typedef)) {
      const innerType = templateTypeDefDataModelFieldType(
        typedef.ItemType,
        fieldName,
        false,
        1,
        packageName,
        `${hierarchy}.[]`,
        getSymbol,
        true, // collection item
      );

      return {
        content: `[]${innerType.content}`,
        imports: innerType.imports,
        additionalFiles: innerType.additionalFiles,
      };
    } else {
      return {
        content: `[]types.Object`,
        imports: ["github.com/hashicorp/terraform-plugin-framework/types"],
        additionalFiles: [],
      };
    }
  }
  if (typedef.AssociatedTypes?.length) {
    const selfName = getSymbol(fieldName, typedef);

    const additionalFile: GeneratedDataModelTypeFile = {
      filename: selfName,
      content: "",
      imports: [],
    };
    additionalFile.content += `type ${selfName} struct {\n`;
    // Union fields are pointers (for nil tracking), except collection items which are values
    let subBuilder = isCollectionItem ? "" : "*";
    let otherAdditionalFiles = [];
    for (const subtype of typedef.AssociatedTypes) {
      const unionName = sanitizeTFUnionTypeName(typedef, subtype);
      const innerType = templateTypeDefDataModelFieldType(
        subtype,
        unionName,
        true,
        1,
        "types",
        `${hierarchy}.${fieldName}`,
        getSymbol,
      );
      otherAdditionalFiles = otherAdditionalFiles.concat(
        innerType.additionalFiles,
      );
      additionalFile.content +=
        templateIndent(1) +
        unionName +
        " " +
        innerType.content +
        ` ${generateUnionStructFieldTag(typedef, subtype)}\n`;
      additionalFile.imports.push(...innerType.imports);
    }
    // Add hoisted fields (fields that were merged to the parent union type)
    for (const field of typedef.Fields) {
      // Skip fields with path aliases only
      if (field.Type.Extensions?.MatchConfig?.Path) {
        continue;
      }
      const innerType = templateTypeDefDataModelFieldType(
        field.Type,
        field.Name,
        field.Optional || field.Nullable,
        1,
        "types",
        `${hierarchy}.${fieldName}`,
        getSymbol,
      );
      otherAdditionalFiles = otherAdditionalFiles.concat(
        innerType.additionalFiles,
      );
      additionalFile.content +=
        templateIndent(1) +
        sanitizeFieldName(field.Name) +
        " " +
        innerType.content +
        ` ${generateStructFieldTag(field)}\n`;
      additionalFile.imports.push(...innerType.imports);
    }
    additionalFile.content += "}\n";
    let prefix = "";
    if (packageName == "provider") {
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );
      prefix = "tfTypes.";
    }

    subBuilder += prefix + selfName;
    return {
      content: subBuilder,
      // TODO: Aliased imports, e.g. tfTypes
      imports: [],
      additionalFiles: otherAdditionalFiles.concat(additionalFile),
    };
  }

  return {
    content: `types.Object`,
    imports: ["github.com/hashicorp/terraform-plugin-framework/types"],
    additionalFiles: [],
  };
}

function generateParamAnnotationStructFieldTag(
  annotations: Annotations | undefined,
): string {
  // Only generate struct field tags for query parameters to reduce code churn,
  // such as ignoring path parameters.
  if (!annotations || typeof annotations.Get !== "function") {
    return "";
  }

  const annotation = annotations.Get("param");

  if (!annotation || annotation.ParamType !== "queryParam") {
    return "";
  }

  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.Serialization) {
    serialization = `serialization=${annotation.Serialization},`;
  } else if (annotation.Style) {
    serialization = `style=${annotation.Style},explode=${
      annotation.Explode ? "true" : "false"
    },`;
  }

  return `${annotation.ParamType}:"${serialization}name=${name}"`;
}

function generateStructFieldTag(fieldDef: FieldDef): string {
  if (
    fieldDef.Type.Extensions?.All["x-speakeasy-soft-delete-property"] ===
      true ||
    fieldDef.Type.Extensions?.TerraformIgnore?.Schema
  ) {
    return `\`tfsdk:"-"\``;
  }

  const structFieldTags: string[] = [
    generateParamAnnotationStructFieldTag(fieldDef.Annotations),
    `tfsdk:"${sanitizeTFStateName(fieldDef.Name)}"`,
  ];

  return `\`${structFieldTags.filter((tag) => tag !== "").join(" ")}\``;
}

function generateUnionStructFieldTag(
  unionTypeDef: TypeDef,
  associatedTypeDef: TypeDef,
): string {
  if (
    associatedTypeDef.Extensions?.All["x-speakeasy-soft-delete-property"] ===
    true
  ) {
    return `\`tfsdk:"-"\``;
  }

  const structFieldTags: string[] = [
    `queryParam:"inline"`,
    `tfsdk:"${sanitizeTFStateName(
      sanitizeTFUnionTypeName(unionTypeDef, associatedTypeDef),
    )}"`,
  ];

  return `\`${structFieldTags.filter((tag) => tag !== "").join(" ")}\``;
}

/** Templates entity data model Go struct fields. */
function templateEntityDataModelFields(entity: TerraformEntity): string {
  function getSymbol(fieldName: string, typedef: TypeDef): string {
    if (typedef.Extensions?.All["Symbol"]) {
      return typedef.Extensions.All["Symbol"];
    }
    const name = sanitizeClassName(typedef.Name || fieldName);

    throw new Error("couldn't get symbol for: " + name);
  }

  const structFields: Record<string, string> = {};

  entity.SchemaTypeDef.AssociatedTypes.forEach((associatedTypeDef) => {
    const unionName = sanitizeTFUnionTypeName(
      entity.SchemaTypeDef,
      associatedTypeDef,
    );
    const innerType = templateTypeDefDataModelFieldType(
      associatedTypeDef,
      unionName,
      true,
      1,
      "provider",
      `${entity.Name}.${unionName}`,
      getSymbol,
    );

    innerType.imports.forEach((imp) => {
      addGenImport(imp);
    });

    const structFieldName = sanitizeFieldName(unionName);
    const structField =
      structFieldName +
      " " +
      innerType.content +
      ` ${generateUnionStructFieldTag(
        entity.SchemaTypeDef,
        associatedTypeDef,
      )}`;

    structFields[structFieldName] = structField;
  });

  entity.SchemaTypeDef.Fields.filter(
    (f) => !f.Type.Extensions?.MatchConfig?.Path,
  ).forEach((fieldDef) => {
    const sanitizedClassName = sanitizeClassName(fieldDef.Name);

    // Skip pagination input fields, as they should not be included.
    if (
      sanitizedClassName in entity.PaginationInputFields ||
      sanitizedClassName in entity.PaginationOutputFields
    ) {
      return;
    }

    const innerType = templateTypeDefDataModelFieldType(
      fieldDef.Type,
      fieldDef.Name,
      fieldDef.Optional || fieldDef.Nullable,
      1,
      "provider",
      `${entity.Name}.${fieldDef.Name}`,
      getSymbol,
    );

    innerType.imports.forEach((imp) => {
      addGenImport(imp);
    });

    const structFieldName = sanitizeFieldName(fieldDef.Name);
    const structField =
      structFieldName +
      " " +
      innerType.content +
      ` ${generateStructFieldTag(fieldDef)}`;

    structFields[structFieldName] = structField;
  });

  if (entity.OperationSecurity) {
    flattenSecurityObject(entity.OperationSecurity.Type).forEach((fieldDef) => {
      const fieldName = sanitizeFieldName(fieldDef.Name);
      const structField = FrameworkTypeFromFieldDef(
        fieldDef,
      ).templateDataModelStructField(fieldDef.Name);

      structFields[fieldName] = structField;
    });
  }

  if (entity.Server) {
    const structFieldName = sanitizeFieldName(entity.Server.AttributeName);
    const structField = new FrameworkStringType().templateDataModelStructField(
      entity.Server.AttributeName,
    );

    structFields[structFieldName] = structField;
  }

  return Object.keys(structFields)
    .sort((a, b) => a.localeCompare(b))
    .map((structFieldName) => structFields[structFieldName])
    .join("\n");
}

registerTemplateFunc(
  "templateEntityDataModelFields",
  templateEntityDataModelFields,
);

/**
 * Returns the set of field names required by close operations.
 * These are fields present in close request shards.
 */
function getCloseRequiredFieldNames(entity: TerraformEntity): Set<string> {
  const fieldNames = new Set<string>();

  const closeRequestShard =
    "CloseRequestShard" in entity.Operations
      ? entity.Operations.CloseRequestShard
      : undefined;

  if (!closeRequestShard?.Fields) {
    return fieldNames;
  }

  for (const field of closeRequestShard.Fields) {
    fieldNames.add(field.Name);
  }

  return fieldNames;
}

/**
 * Returns the private data model fields: field defs from the entity TypeDef
 * that are required by close operations and have supported primitive types.
 */
function getEphemeralResourcePrivateDataModelFields(
  entity: TerraformEntity,
): { fieldDef: FieldDef; frameworkType: FrameworkType }[] {
  if (!("HasCloseOperations" in entity) || !entity.HasCloseOperations()) {
    return [];
  }

  const closeFieldNames = getCloseRequiredFieldNames(entity);

  if (closeFieldNames.size === 0) {
    return [];
  }

  const result: { fieldDef: FieldDef; frameworkType: FrameworkType }[] = [];

  if (!entity.SchemaTypeDef?.Fields) {
    return result;
  }

  for (const fieldDef of entity.SchemaTypeDef.Fields) {
    if (!closeFieldNames.has(fieldDef.Name)) {
      continue;
    }

    const frameworkType = FrameworkTypeFromFieldDef(fieldDef);

    if (frameworkType.templatePrivateDataModelStructFieldType() !== undefined) {
      result.push({ fieldDef, frameworkType });
    }
  }

  return result;
}

/** Templates ephemeral resource PrivateDataModel Go struct fields with json tags. */
function templateEphemeralResourcePrivateDataModelFields(
  entity: TerraformEntity,
): string {
  const fields = getEphemeralResourcePrivateDataModelFields(entity);

  if (fields.length === 0) {
    return "";
  }

  addGenImport("encoding/json");

  const structFields: string[] = [];

  for (const { fieldDef, frameworkType } of fields) {
    const structField = frameworkType.templatePrivateDataModelStructField(
      fieldDef.Name,
    );

    if (structField !== undefined) {
      structFields.push(`${templateIndent(1)}${structField}`);
    }
  }

  return structFields.sort().join("\n");
}

registerTemplateFunc(
  "templateEphemeralResourcePrivateDataModelFields",
  templateEphemeralResourcePrivateDataModelFields,
);

/**
 * Templates a Go method on the DataModel that converts to PrivateDataModel.
 * Generates: func (r *XModel) toPrivateDataModel() *XPrivateDataModel { ... }
 */
function templateDataModelToPrivateDataModelMethod(
  entity: TerraformEntity,
): string {
  const fields = getEphemeralResourcePrivateDataModelFields(entity);

  if (fields.length === 0) {
    return "";
  }

  const privateModelTypeName = entity.GoPrivateDataModelTypeName;

  const fieldAssignments: string[] = [];

  for (const { fieldDef, frameworkType } of fields) {
    const fieldName = sanitizeFieldName(fieldDef.Name);
    const valueMethod = frameworkType.templateValueMethod();

    if (valueMethod !== undefined) {
      fieldAssignments.push(
        `${templateIndent(2)}${fieldName}: r.${fieldName}.${valueMethod},`,
      );
    }
  }

  const lines: string[] = [
    `func (r *${entity.GoDataModelTypeName}) toPrivateDataModel() *${privateModelTypeName} {`,
    `${templateIndent(1)}return &${privateModelTypeName}{`,
    ...fieldAssignments.sort(),
    `${templateIndent(1)}}`,
    `}`,
  ];

  return lines.join("\n");
}

registerTemplateFunc(
  "templateDataModelToPrivateDataModelMethod",
  templateDataModelToPrivateDataModelMethod,
);

/**
 * Templates a Go method on the PrivateDataModel that converts to DataModel.
 * Generates: func (r *XPrivateDataModel) toDataModel() *XModel { ... }
 */
function templatePrivateDataModelToDataModelMethod(
  entity: TerraformEntity,
): string {
  const fields = getEphemeralResourcePrivateDataModelFields(entity);

  if (fields.length === 0) {
    return "";
  }

  const privateModelTypeName = entity.GoPrivateDataModelTypeName;

  const fieldAssignments: string[] = [];

  for (const { fieldDef, frameworkType } of fields) {
    const fieldName = sanitizeFieldName(fieldDef.Name);
    const frameworkValue = frameworkType.templatePrivateDataModelFrameworkValue(
      `r.${fieldName}`,
    );

    if (frameworkValue !== undefined) {
      fieldAssignments.push(
        `${templateIndent(2)}${fieldName}: ${frameworkValue},`,
      );
    }
  }

  const lines: string[] = [
    `func (r *${privateModelTypeName}) toDataModel() *${entity.GoDataModelTypeName} {`,
    `${templateIndent(1)}return &${entity.GoDataModelTypeName}{`,
    ...fieldAssignments.sort(),
    `${templateIndent(1)}}`,
    `}`,
  ];

  return lines.join("\n");
}

registerTemplateFunc(
  "templatePrivateDataModelToDataModelMethod",
  templatePrivateDataModelToDataModelMethod,
);
