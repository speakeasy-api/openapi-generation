function generateSchema(
  entity: TerraformEntity,
  resourceType: TerraformResourceType,
): string {
  const iterator: IteratorFunction = (field, renderChildren) => {
    if (field.type.Extensions?.TerraformIgnore?.Schema) {
      return {};
    }

    let fieldAttributes = "";

    const fieldSchemaType = getSchemaType(field.type);
    fieldAttributes += `${templateIndent(field.indent)}"${
      field.name
    }": ${fieldSchemaType}{\n`;
    const sanitizedClassName = sanitizeClassName(field.name);

    // Check if this is a root-level field (entity.field)
    const isRootAttribute = field.hierarchy.split(".").length === 2;
    const isGlobalField =
      isRootAttribute && sanitizedClassName in entity.GlobalFields;
    const isPaginationInputField =
      isRootAttribute && sanitizedClassName in entity.PaginationInputFields;
    const isPaginationOutputField =
      isRootAttribute && sanitizedClassName in entity.PaginationOutputFields;

    const fieldConfig = configFromTypeDef(
      field.type,
      field.defaultValue,
      field.optional,
      resourceType,
      isGlobalField,
    );

    // Intentionally do not include x-speakeasy-soft-delete-property in managed
    // resource schema since these properties would provide no Terraform user
    // benefit (generally not configurable, no state value until deleted) and
    // would warrant additional handling to prevent unknowns during planning.
    //
    // Conversely, leave it in data resource schemas so Terraform users can
    // benefit from being able to client-side filter these.
    if (resourceType === "managed" && fieldConfig.SoftDeleteProperty) {
      return {};
    }

    // Do not generate schema for pagination fields. The underlying Go SDK
    // handles all pagination logic and conventionally in the Terraform Provider
    // ecosystem pagination details are not exposed (e.g. always paginate
    // all results automatically).
    if (isPaginationInputField || isPaginationOutputField) {
      return {};
    }

    fieldAttributes += generateCommonAttributes(
      fieldConfig,
      field.indent + 1,
      isGlobalField,
    );

    let temporaryAttributes = "";
    if (resourceType === "managed") {
      if (fieldConfig.Computed) {
        const schemaDefault = createSchemaDefault(
          field.type,
          field.defaultValue,
          isGlobalField,
          fieldConfig,
        );
        if (schemaDefault) {
          const defaultDescription = schemaDefault.description();
          if (defaultDescription) {
            field.extraDescription = [
              field.extraDescription,
              defaultDescription,
            ]
              .filter(Boolean)
              .join("; ");
          }
          schemaDefault.imports().forEach((imp) => {
            addGenImport(imp.Path, false, imp.Alias);
          });
          fieldAttributes += `${templateIndent(
            field.indent + 1,
          )}Default: ${schemaDefault.template()},\n`;
        }
      }
      temporaryAttributes += generateValidators(
        field,
        fieldConfig,
        field.indent + 1,
      );
      fieldAttributes += generatePlanModifiers(
        field,
        fieldConfig,
        field.indent + 1,
      );
    } else if (fieldConfig.Optional || fieldConfig.Required) {
      temporaryAttributes += generateValidators(
        field,
        fieldConfig,
        field.indent + 1,
      );
    }

    if (field.type.Type == "class" && field.type.Fields.length) {
      fieldAttributes += handleTypeWithAttributesChildren(
        renderChildren,
        field.indent,
      );
    } else if (field.type.AssociatedTypes?.length) {
      fieldAttributes += handleTypeWithAttributesChildren(
        renderChildren,
        field.indent,
      );
    } else if (
      field.type.ItemType?.Type.toString() === "class" ||
      field.type.ItemType?.Type.toString() === "union"
    ) {
      fieldAttributes += handleTypeWithNestedObject(
        resourceType,
        field.type.ItemType,
        renderChildren,
        field.indent,
        `${field.hierarchy}.[]`,
      );
    } else if (field.type.ItemType) {
      const elementType = FrameworkTypeFromTypeDef(field.type.ItemType);
      elementType
        .schemaTypeImports()
        .forEach((typeImport) => addGenImport(typeImport));
      fieldAttributes += `ElementType: ${elementType.templateInstantiation()},\n`;
    } else if (field.type.Type == "any") {
      field.extraDescription = [field.extraDescription, `Parsed as JSON.`]
        .filter(Boolean)
        .join("; ");
    }

    if (field.comments?.Deprecated || field.type.Comments?.Deprecated) {
      const deprecationMessage =
        field.comments?.DeprecationMessage ||
        field.type.Comments?.DeprecationMessage;

      fieldAttributes += `${templateIndent(
        field.indent,
      )}DeprecationMessage: ${templateBuiltinString(deprecationMessage)},\n`;
    }

    fieldAttributes += generateDescription(field);
    fieldAttributes += temporaryAttributes;
    fieldAttributes += `${templateIndent(field.indent)}},\n`;

    return { [field.name]: fieldAttributes };
  };

  const attributes = AttributeIterator(
    entity.SchemaTypeDef,
    iterator,
    3,
    entity.Name,
  );

  if (entity.OperationSecurity) {
    let securityAttributes:
      | DataResourceAttributes
      | EphemeralResourceAttributes
      | ManagedResourceAttributes;

    switch (resourceType) {
      case "action":
        securityAttributes = getActionSecurityAttributes(
          entity.OperationSecurity.Type,
        );
        break;
      case "data":
        securityAttributes = getDataResourceSecurityAttributes(
          entity.OperationSecurity.Type,
        );
        break;
      case "ephemeral":
        securityAttributes = getEphemeralResourceSecurityAttributes(
          entity.OperationSecurity.Type,
        );
        break;
      case "managed":
        securityAttributes = getManagedResourceSecurityAttributes(
          entity.OperationSecurity.Type,
        );
        break;
      default:
        throw new Error(`Unsupported resource type: ${resourceType}`);
    }

    Object.entries(securityAttributes).forEach(([attributeName, attribute]) => {
      attribute.imports().forEach((imp) => {
        addGenImport(imp.Path, false, imp.Alias);
      });

      attributes[
        attributeName
      ] = `"${attributeName}": ${attribute.template()},\n`;
    });
  }

  if (entity.Server) {
    const attributeName = sanitizeTFStateName(entity.Server.AttributeName);

    if (attributeName in attributes) {
      throw new Error(`Duplicate attribute name: ${attributeName}`);
    }

    let attribute:
      | ActionStringAttribute
      | DataResourceStringAttribute
      | EphemeralResourceStringAttribute
      | ManagedResourceStringAttribute;

    switch (resourceType) {
      case "action":
        attribute = new ActionStringAttribute({
          markdownDescription: entity.Server.Description,
          optional: true,
        });
        break;
      case "data":
        attribute = new DataResourceStringAttribute({
          markdownDescription: entity.Server.Description,
          optional: true,
        });
        break;
      case "ephemeral":
        attribute = new EphemeralResourceStringAttribute({
          markdownDescription: entity.Server.Description,
          optional: true,
        });
        break;
      case "managed":
        attribute = new ManagedResourceStringAttribute({
          markdownDescription: entity.Server.Description,
          optional: true,
        });
        break;
      default:
        throw new Error(`Unsupported resource type: ${resourceType}`);
    }

    attribute.imports().forEach((imp) => {
      addGenImport(imp.Path, false, imp.Alias);
    });

    attributes[
      attributeName
    ] = `"${attributeName}": ${attribute.template()},\n`;
  }

  return Object.keys(attributes)
    .sort((a, b) => a.localeCompare(b))
    .map((key) => attributes[key])
    .join("");
}

registerTemplateFunc("generateSchema", generateSchema);

function generateCommonAttributes(
  config: TypeDefConfig,
  indent: number,
  isGlobalField: boolean = false,
): string {
  let attributes = "";

  if (config.CustomType?.schemaType) {
    if (config.CustomType?.imports?.length) {
      config.CustomType.imports.forEach((importPath) => {
        addGenImport(importPath);
      });
    }

    attributes += `${templateIndent(indent)}CustomType: ${
      config.CustomType.schemaType
    },\n`;
  }

  if (config.Computed || isGlobalField)
    attributes += `${templateIndent(indent)}Computed: true,\n`;
  if (config.Optional)
    attributes += `${templateIndent(indent)}Optional: true,\n`;
  if (config.Required)
    attributes += `${templateIndent(indent)}Required: true,\n`;
  if (config.Sensitive)
    attributes += `${templateIndent(indent)}Sensitive: true,\n`;
  if (config.WriteOnly)
    attributes += `${templateIndent(indent)}WriteOnly: true,\n`;

  return attributes;
}

function generateDescription(field: AttributeField): string {
  const comments = field.comments;
  const description = comments?.Description || comments?.Summary || "";
  let extraDescription = field.extraDescription;

  if (description || extraDescription) {
    return `${templateIndent(field.indent)}${formatDescription(
      description,
      field.indent + 1,
      extraDescription,
    )},\n`;
  }
  return "";
}

type RenderChildrenOptions = {
  indent: number;
  filter?: (child: AttributeField) => boolean;
};

type IteratorFunction = (
  field: AttributeField,
  renderChildren: ({ indent, filter }: RenderChildrenOptions) => string,
) => Record<string, string>;

type AttributeField = {
  name: string;
  type: TypeDef;
  extraDescription: string;
  example?: Example;
  optional: boolean;
  nullable: boolean;
  defaultValue?: AnyValue;
  comments: CommentDef;
  indent: number;
  hierarchy: string;
};

/**
 * Returns an AttributeField from a FieldDef.
 */
function attributeFieldFromFieldDef(
  fieldDef: FieldDef,
  name: string,
  hierarchy: string,
): AttributeField {
  return {
    comments: fieldDef.Comments || fieldDef.Type.Comments,
    defaultValue: fieldDef.Default,
    example: fieldDef.Type.Examples.filter((e) =>
      typeCompliant(fieldDef.Type, JSON.parse(e.ToJSON())),
    )[0],
    extraDescription: "",
    hierarchy: `${hierarchy}.${name}`,
    indent: 0,
    name: name,
    nullable: fieldDef.Nullable,
    optional: fieldDef.Optional || fieldDef.Nullable,
    type: fieldDef.Type,
  };
}

/**
 * Returns an AttributeField from a TerraformEntity.
 */
function attributeFieldFromTerraformEntity(
  entity: TerraformEntity,
): AttributeField {
  return {
    comments: entity.SchemaTypeDef.Comments,
    example: entity.SchemaTypeDef.Examples.filter((e) =>
      typeCompliant(entity.SchemaTypeDef, JSON.parse(e.ToJSON())),
    )[0],
    extraDescription: "",
    hierarchy: entity.Name,
    indent: 0,
    name: entity.Name,
    nullable: false,
    optional: false,
    type: entity.SchemaTypeDef,
  };
}

/** Given a TypeDef returns an AttributeField. */
function unionAttributeFieldFromTypeDef(
  typeDef: TypeDef,
  name: string,
  hierarchy: string,
): AttributeField {
  return {
    comments: typeDef.Comments,
    example: typeDef.Examples.filter((e) =>
      typeCompliant(typeDef, JSON.parse(e.ToJSON())),
    )[0],
    extraDescription: "",
    hierarchy: hierarchy,
    indent: 0,
    name: name,
    nullable: true,
    optional: true,
    type: typeDef,
  };
}

function typeCompliant(field: TypeDef, value: any): boolean {
  switch (getSchemaType(field)) {
    case "schema.StringAttribute":
      return typeof value == "string";
    case "schema.Int32Attribute":
    case "schema.Int64Attribute":
      if (typeof value == "number" && Number.isInteger(value)) {
        return true;
      }
      if (typeof value == "string") {
        return /^\d+$/.test(value);
      }
      return false;
    case "schema.Float32Attribute":
    case "schema.Float64Attribute":
      if (typeof value == "number") {
        return true;
      }
      if (typeof value == "string") {
        return /^\d+(\.\d+)?$/.test(value);
      }
      return false;
    case "schema.BoolAttribute":
      if (typeof value == "boolean") {
        return true;
      }
      if (typeof value == "string") {
        return /^(true|false)$/.test(value);
      }
    case "schema.MapAttribute":
      return false;
    case "schema.SetAttribute":
    case "schema.ListAttribute": {
      const isArray = Array.isArray(value);
      if (!isArray) {
        return false;
      }
      return value.every((v) => typeCompliant(field.ItemType, v));
    }
    case "schema.MapNestedAttribute":
    case "schema.ListNestedAttribute":
    case "schema.SetNestedAttribute":
    case "schema.SingleNestedAttribute":
      return false;
  }
}

function AttributeIterator(
  typedef: TypeDef,
  iterator: IteratorFunction,
  indent: number,
  hierarchy: string,
  options: RenderChildrenOptions = {
    indent: indent,
    filter: () => true,
  },
): Record<string, string> {
  if (!options.filter) {
    options.filter = () => true;
  }
  const fields: Record<string, string> = {};

  function iterateField(field: AttributeField) {
    const renderChildren = (options: RenderChildrenOptions): string => {
      const childFields = AttributeIterator(
        field.type,
        iterator,
        options.indent,
        `${hierarchy}.${field.name}`,
        options,
      );

      return Object.keys(childFields)
        .sort((a, b) => a.localeCompare(b))
        .map((key) => childFields[key])
        .join("");
    };

    const result = iterator(field, renderChildren);
    Object.assign(fields, result);
  }

  function handleFields(typedef: TypeDef) {
    typedef.AssociatedTypes.forEach((associatedTypeDef) => {
      const name = sanitizeTFStateName(
        sanitizeTFUnionTypeName(typedef, associatedTypeDef),
      );
      const fieldContext = unionAttributeFieldFromTypeDef(
        associatedTypeDef,
        name,
        `${hierarchy}.${name}`,
      );

      if (options.filter(fieldContext)) {
        iterateField(fieldContext);
      }
    });

    typedef.Fields.forEach((fieldDef) => {
      // Skip fields with path aliases (they're represented by other fields in the schema)
      // but include fields that only have usePriorState (they're normal fields that use prior state in updates)
      if (fieldDef.Type.Extensions?.MatchConfig?.Path) {
        return;
      }

      const name = sanitizeTFStateName(fieldDef.Name);
      const fieldContext = attributeFieldFromFieldDef(
        fieldDef,
        name,
        hierarchy,
      );

      if (options.filter(fieldContext)) {
        iterateField(fieldContext);
      }
    });
  }
  handleFields(typedef);

  if (isComplexType(typedef.ItemType)) {
    handleFields(typedef.ItemType);
  }

  return fields;
}

// This handles "Complex" List/Set/Map child type which need recursion into children
// returns the NestedObject attribute, alongside validators, plan modifiers and adds required imports
function handleTypeWithNestedObject(
  resourceType: TerraformResourceType,
  childType: TypeDef,
  renderChildren: ({ indent }: { indent: number }) => string,
  indent: number,
  hierarchy: string,
): string {
  let attributes = `${templateIndent(
    indent,
  )}NestedObject: schema.NestedAttributeObject{\n`;

  if (resourceType === "managed") {
    const subConfig = configFromTypeDef(
      childType,
      undefined,
      false,
      resourceType,
      false,
    );
    const field = {
      name: "",
      type: childType,
      extraDescription: "",
      optional: true,
      nullable: true,
      defaultValue: undefined,
      comments: childType.Comments,
      indent: indent + 1,
      hierarchy,
    };
    attributes += generateValidators(field, subConfig, indent + 1);
    attributes += generatePlanModifiers(field, subConfig, indent + 1);
  }

  attributes += `${templateIndent(
    indent + 1,
  )}Attributes: map[string]schema.Attribute{\n`;
  attributes += renderChildren({ indent: indent + 2 });
  attributes += `${templateIndent(indent + 1)}},\n`;
  attributes += `${templateIndent(indent)}},\n`;
  return attributes;
}

function handleTypeWithAttributesChildren(
  renderChildren: ({ indent }: { indent: number }) => string,
  indent: number,
): string {
  let attributes = `${templateIndent(
    indent,
  )}Attributes: map[string]schema.Attribute{\n`;
  attributes += renderChildren({ indent: indent + 1 });
  attributes += `${templateIndent(indent)}},\n`;
  return attributes;
}

function generateValidators(
  field: AttributeField,
  config: TypeDefConfig,
  indent: number,
): string {
  // Configuration validation is only necessary for configurable attributes.
  if (!config.Optional && !config.Required) {
    return "";
  }

  const typedef = field.type;
  const schemaValidators: SchemaValidator[] = [];

  if (config.NotNull) {
    const notNullValidator = createNotNullValidator(typedef);
    if (notNullValidator) {
      schemaValidators.push(notNullValidator);
    }
  }

  // Add type-specific and path validators
  switch (String(typedef.Type)) {
    case "boolean":
      schemaValidators.push(
        ...BoolSchemaValidator.fromPathValidatorConfig(config),
      );
      break;
    case "float32":
      schemaValidators.push(
        ...Float32SchemaValidator.fromPathValidatorConfig(config),
        ...Float32SchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "int32":
      schemaValidators.push(
        ...Int32SchemaValidator.fromPathValidatorConfig(config),
        ...Int32SchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "integer":
      schemaValidators.push(
        ...Int64SchemaValidator.fromPathValidatorConfig(config),
        ...Int64SchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "number":
      schemaValidators.push(
        ...Float64SchemaValidator.fromPathValidatorConfig(config),
        ...Float64SchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      schemaValidators.push(
        ...StringSchemaValidator.fromPathValidatorConfig(config),
        ...StringSchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "array":
      schemaValidators.push(
        ...ListSchemaValidator.fromPathValidatorConfig(config),
        ...ListSchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "set":
      schemaValidators.push(
        ...SetSchemaValidator.fromPathValidatorConfig(config),
        ...SetSchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "map":
      schemaValidators.push(
        ...MapSchemaValidator.fromPathValidatorConfig(config),
        ...MapSchemaValidator.fromTypeDef(typedef),
      );
      break;
    case "class":
    case "union":
      schemaValidators.push(
        ...ObjectSchemaValidator.fromPathValidatorConfig(config),
      );
      break;
    case "enum":
      schemaValidators.push(
        ...createEnumValidatorsFromPathValidatorConfig(typedef, config),
        ...createEnumValidatorsFromTypeDef(typedef),
      );
      // Open enums skip the OneOf runtime validator (in fromTypeDef), but
      // we still document the known values for provider docs.
      if (typedef.Enum?.Open && typedef.Enum?.Values?.length) {
        const isString = typedef.Enum.Type.Type.toString() === "string";
        const values = isString
          ? typedef.Enum.Values.map((v: any) => JSON.stringify(v)).join(", ")
          : typedef.Enum.Values.join(", ");
        const desc =
          typedef.Enum.Values.length > 1
            ? `possible known values include one of [${values}]`
            : `possible known value is ${values}`;
        field.extraDescription = [desc, field.extraDescription]
          .filter(Boolean)
          .join("; ");
      }
      break;
  }

  // Add custom validators
  if (config.CustomValidators && config.CustomValidators.length > 0) {
    config.CustomValidators.forEach((validatorName) => {
      const customValidator = createCustomValidator(typedef, validatorName);
      if (customValidator) {
        schemaValidators.push(customValidator);
      }
    });
  }

  if (schemaValidators.length > 0) {
    const validatorType = schemaValidators[0].validatorTypeName;
    addGenImport(
      "github.com/hashicorp/terraform-plugin-framework/schema/validator",
    );

    for (const v of schemaValidators) {
      v.imports().forEach((imp) => addGenImport(imp.Path, false, imp.Alias));
      const desc = v.description();
      if (desc) {
        field.extraDescription = [field.extraDescription, desc]
          .filter(Boolean)
          .join("; ");
      }
    }

    return (
      `${templateIndent(indent)}Validators: []validator.${validatorType}{\n` +
      schemaValidators
        .map((v) => `${templateIndent(indent + 1)}${v.template()},\n`)
        .join("") +
      `${templateIndent(indent)}},\n`
    );
  }

  return "";
}

function generatePlanModifiers(
  field: AttributeField,
  config: TypeDefConfig,
  indent: number,
): string {
  const typedef = field.type;
  const schemaPlanModifiers: SchemaPlanModifier[] = [];

  const planModifierConfig: PlanModifierConfig = {
    CustomPlanModifiers: config.CustomPlanModifiers,
    ForceNew: config.ForceNew,
    SuppressComputedDiff: config.SuppressComputedDiff,
    PlanOnly: config.PlanOnly,
    HoistedFrom: config.HoistedFrom,
    Computed: config.Computed,
  };

  switch (String(typedef.Type)) {
    case "boolean":
      schemaPlanModifiers.push(
        ...BoolSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "float32":
      schemaPlanModifiers.push(
        ...Float32SchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "int32":
      schemaPlanModifiers.push(
        ...Int32SchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "integer":
      schemaPlanModifiers.push(
        ...Int64SchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "number":
      schemaPlanModifiers.push(
        ...Float64SchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      schemaPlanModifiers.push(
        ...StringSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "array":
      schemaPlanModifiers.push(
        ...ListSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "set":
      schemaPlanModifiers.push(
        ...SetSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "map":
      schemaPlanModifiers.push(
        ...MapSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "class":
    case "union":
      schemaPlanModifiers.push(
        ...ObjectSchemaPlanModifier.fromConfig(planModifierConfig),
      );
      break;
    case "enum":
      schemaPlanModifiers.push(
        ...createEnumPlanModifiersFromConfig(typedef, planModifierConfig),
      );
      break;
  }

  if (schemaPlanModifiers.length > 0) {
    const planModifierType = schemaPlanModifiers[0].planModifierTypeName;

    for (const pm of schemaPlanModifiers) {
      pm.imports().forEach((imp) => addGenImport(imp.Path, false, imp.Alias));
      const desc = pm.description();
      if (desc) {
        field.extraDescription = [field.extraDescription, desc]
          .filter(Boolean)
          .join("; ");
      }
    }

    return (
      `${templateIndent(
        indent,
      )}PlanModifiers: []planmodifier.${planModifierType}{\n` +
      schemaPlanModifiers
        .map((pm) => `${templateIndent(indent + 1)}${pm.template()},\n`)
        .join("") +
      `${templateIndent(indent)}},\n`
    );
  }

  return "";
}

function isComplexType(type: TypeDef): boolean {
  return (
    type &&
    (type.Type == "class" ||
      type.Type == "map" ||
      type.Type == "array" ||
      type.Type == "set" ||
      (type.AssociatedTypes?.length ?? 0) > 0)
  );
}

function getSchemaType(typedef: TypeDef): TerraformAttributeType {
  switch (String(typedef.Type)) {
    case "any":
    case "bytes":
    case "date-time":
    case "date":
    case "string":
      // NOTE: Current implementation details
      //   any: JSON escape hatch (internal issue reference)
      //   date/date-time: Date validators
      return "schema.StringAttribute";
    case "float32":
      return "schema.Float32Attribute";
    case "int32":
      return "schema.Int32Attribute";
    case "integer":
      return "schema.Int64Attribute";
    case "number":
      return "schema.Float64Attribute";
    case "boolean":
      return "schema.BoolAttribute";
    case "array":
      return typedef.ItemType?.Type.toString() === "class" ||
        typedef.ItemType?.Type.toString() === "union"
        ? "schema.ListNestedAttribute"
        : "schema.ListAttribute";
    case "set":
      return typedef.ItemType?.Type.toString() === "class" ||
        typedef.ItemType?.Type.toString() === "union"
        ? "schema.SetNestedAttribute"
        : "schema.SetAttribute";
    case "map":
      return typedef.ItemType?.Type.toString() === "class" ||
        typedef.ItemType?.Type.toString() === "union"
        ? "schema.MapNestedAttribute"
        : "schema.MapAttribute";
    case "class":
    case "union":
      return "schema.SingleNestedAttribute";
    case "enum":
      return getSchemaType(typedef.Enum.Type);
    default:
      throw new Error(`Unsupported type: ${typedef.Type}`);
  }
}

/**
 * Returns action attributes for the flattened security TypeDef.
 */
function getActionSecurityAttributes(security: TypeDef): ActionAttributes {
  const attributes: ActionAttributes = {};

  if (!security) {
    return attributes;
  }

  const fieldDefs = flattenSecurityObject(security);

  for (const fieldDef of fieldDefs) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    const optional = fieldDef.Optional || fieldDef.Default?.Value;
    const attribute = ActionAttributeFromFieldDef(fieldDef, {
      markdownDescription: getSecurityAttributeDescription(fieldDef, undefined),
      optional: optional,
      required: !optional,
      sensitive: true,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}

/**
 * Returns data resource attributes for the flattened security TypeDef.
 */
function getDataResourceSecurityAttributes(
  security: TypeDef,
): DataResourceAttributes {
  const attributes: DataResourceAttributes = {};

  if (!security) {
    return attributes;
  }

  const fieldDefs = flattenSecurityObject(security);

  for (const fieldDef of fieldDefs) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    // TODO: Consider environment variables per-resource. If being implemented,
    // this feature should be designed carefully to avoid configuration
    // collisions, e.g. consider requiring the resource type and resource name
    // as part of the environmentVariables generation configuration.
    // const environmentVariable = getEnvironmentVariable(attributeName);
    // const optional =
    //   fieldDef.Optional || fieldDef.Default?.Value || environmentVariable;
    const optional = fieldDef.Optional || fieldDef.Default?.Value;
    const attribute = DataResourceAttributeFromFieldDef(fieldDef, {
      markdownDescription: getSecurityAttributeDescription(fieldDef, undefined),
      optional: optional,
      required: !optional,
      sensitive: true,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}

/**
 * Returns ephemeral resource attributes for the flattened security TypeDef.
 */
function getEphemeralResourceSecurityAttributes(
  security: TypeDef,
): EphemeralResourceAttributes {
  const attributes: EphemeralResourceAttributes = {};

  if (!security) {
    return attributes;
  }

  const fieldDefs = flattenSecurityObject(security);

  for (const fieldDef of fieldDefs) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    // TODO: Consider environment variables per-resource. If being implemented,
    // this feature should be designed carefully to avoid configuration
    // collisions, e.g. consider requiring the resource type and resource name
    // as part of the environmentVariables generation configuration.
    // const environmentVariable = getEnvironmentVariable(attributeName);
    // const optional =
    //   fieldDef.Optional || fieldDef.Default?.Value || environmentVariable;
    const optional = fieldDef.Optional || fieldDef.Default?.Value;
    const attribute = EphemeralResourceAttributeFromFieldDef(fieldDef, {
      markdownDescription: getSecurityAttributeDescription(fieldDef, undefined),
      optional: optional,
      required: !optional,
      sensitive: true,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}

/**
 * Returns managed resource attributes for the flattened security TypeDef.
 */
function getManagedResourceSecurityAttributes(
  security: TypeDef,
): ManagedResourceAttributes {
  const attributes: ManagedResourceAttributes = {};

  if (!security) {
    return attributes;
  }

  const fieldDefs = flattenSecurityObject(security);

  for (const fieldDef of fieldDefs) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    // TODO: Consider environment variables per-resource. If being implemented,
    // this feature should be designed carefully to avoid configuration
    // collisions, e.g. consider requiring the resource type and resource name
    // as part of the environmentVariables generation configuration.
    // const environmentVariable = getEnvironmentVariable(attributeName);
    // const optional =
    //   fieldDef.Optional || fieldDef.Default?.Value || environmentVariable;
    const optional = fieldDef.Optional || fieldDef.Default?.Value;
    const attribute = ManagedResourceAttributeFromFieldDef(fieldDef, {
      markdownDescription: getSecurityAttributeDescription(fieldDef, undefined),
      optional: optional,
      required: !optional,
      sensitive: true,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}
