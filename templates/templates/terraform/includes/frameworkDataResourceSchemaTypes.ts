/**
 * Collection of all terraform-plugin-framework datasource/schema.Attribute types.
 */
type DataResourceAttribute =
  | DataResourceBoolAttribute
  | DataResourceDynamicAttribute
  | DataResourceFloat32Attribute
  | DataResourceFloat64Attribute
  | DataResourceInt32Attribute
  | DataResourceInt64Attribute
  | DataResourceListAttribute
  | DataResourceListNestedAttribute
  | DataResourceMapAttribute
  | DataResourceMapNestedAttribute
  | DataResourceNumberAttribute
  | DataResourceObjectAttribute
  | DataResourceSetAttribute
  | DataResourceSetNestedAttribute
  | DataResourceSingleNestedAttribute
  | DataResourceStringAttribute
  | DataResourceTupleAttribute;

/**
 * Create a DataResourceAttribute from a FieldDef.
 */
function DataResourceAttributeFromFieldDef(
  fieldDef: FieldDef,
  config: DataResourceAttributeConfiguration,
): DataResourceAttribute {
  return DataResourceAttributeFromTypeDef(fieldDef.Type, config);
}

/**
 * Create a DataResourceAttribute from a TypeDef.
 */
function DataResourceAttributeFromTypeDef(
  typeDef: TypeDef,
  config: DataResourceAttributeConfiguration,
): DataResourceAttribute {
  // Preserve the TypeDef for example generation
  const configWithTypeDef = { ...config, typeDef };

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new DataResourceStringAttribute(configWithTypeDef);
    case "boolean":
      return new DataResourceBoolAttribute(configWithTypeDef);
    case "enum": {
      // Create validators from the enum TypeDef before recursing, since the
      // recursive call overwrites config.typeDef with the underlying type.
      const enumValidators = createEnumValidatorsFromTypeDef(typeDef);
      const enumConfig = {
        ...configWithTypeDef,
        validators: [
          ...(configWithTypeDef.validators || []),
          ...enumValidators,
        ],
      };
      return DataResourceAttributeFromTypeDef(typeDef.Enum.Type, enumConfig);
    }
    case "float32":
      return new DataResourceFloat32Attribute(configWithTypeDef);
    case "integer":
      return new DataResourceInt64Attribute(configWithTypeDef);
    case "int32":
      return new DataResourceInt32Attribute(configWithTypeDef);
    case "number":
      return new DataResourceFloat64Attribute(configWithTypeDef);
    case "map":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new DataResourceMapNestedAttribute(
          new DataResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  DataResourceAttributeFromFieldDef(
                    fieldDef,
                    configWithTypeDef,
                  ),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new DataResourceMapAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "array":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new DataResourceListNestedAttribute(
          new DataResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  DataResourceAttributeFromFieldDef(
                    fieldDef,
                    configWithTypeDef,
                  ),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new DataResourceListAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "set":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new DataResourceSetNestedAttribute(
          new DataResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  DataResourceAttributeFromFieldDef(
                    fieldDef,
                    configWithTypeDef,
                  ),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new DataResourceSetAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "class":
    case "union":
      // TODO: Finish the implementation for AssociatedTypes/Fields.
      throw new Error(`Unimplemented: ${typeDef.Type}`);
      const attributes: DataResourceAttributes = {};
      return new DataResourceSingleNestedAttribute(
        attributes,
        configWithTypeDef,
      );
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/**
 * Mapping of attribute names to data resource attributes.
 */
type DataResourceAttributes = Record<string, DataResourceAttribute>;

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute.
 */
abstract class DataResourceAttributeBase extends AttributeBase {
  /** Original OAS TypeDef, preserved for example generation. */
  typeDef?: TypeDef;

  constructor(
    typeName: AttributeBaseTypeName,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName);
    this.computed = config.computed;
    // Data resource attributes intentionally do not implement default.
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.optional = config.optional;
    this.markdownDescription = config.markdownDescription;
    // Data resource attributes intentionally do not implement planModifiers.
    this.required = config.required;
    this.sensitive = config.sensitive;
    this.typeDef = config.typeDef;
    this.validators = config.validators;

    // Extend attribute descriptions with validator descriptions.
    if (this.validators?.length) {
      const descriptions = this.validators
        .map((v) => v.description())
        .filter(Boolean);

      if (descriptions.length > 0) {
        const combined = descriptions.join("; ");

        if (this.markdownDescription) {
          this.markdownDescription += `; ${combined}`;
        } else if (this.description) {
          this.description += `; ${combined}`;
        } else {
          this.markdownDescription = combined;
        }
      }
    }
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/datasource/schema",
    });
    return result;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute with Attributes (e.g. SingleNested).
 */
abstract class DataResourceAttributeBaseWithAttributes extends DataResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributes: DataResourceAttributes,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributes = attributes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute with AttributeTypes (e.g. Object).
 */
abstract class DataResourceAttributeBaseWithAttributeTypes extends DataResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributeTypes: FrameworkTypeAttributeTypes,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributeTypes = attributeTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute with ElementType (e.g. List, Map, Set).
 */
abstract class DataResourceAttributeBaseWithElementType extends DataResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementType: FrameworkType,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementType = elementType;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute with ElementTypes (e.g. Tuple).
 */
abstract class DataResourceAttributeBaseWithElementTypes extends DataResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementTypes: FrameworkTypeElementTypes,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementTypes = elementTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * datasource/schema.Attribute with NestedObject (e.g. ListNested).
 */
abstract class DataResourceAttributeBaseWithNestedObject extends DataResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    nestedObject: DataResourceNestedAttributeObject,
    config: DataResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.nestedObject = nestedObject;
  }
}

/**
 * Available configuration for DataResourceAttribute.
 */
type DataResourceAttributeConfiguration = {
  /**
   * Computed indicates whether the provider may return its own value for
   * this Attribute or not. Required and Computed cannot both be true. If
   * Required and Optional are both false, Computed must be true, and the
   * attribute will be considered "read only" for the practitioner, with
   * only the provider able to set its value.
   */
  computed?: boolean;

  /**
   * Default defines a proposed new state (plan) value for the attribute
   * if the configuration value is null. Default prevents the framework
   * from automatically marking the value as unknown during planning when
   * other proposed new state changes are detected. If the attribute is
   * computed and the value could be altered by other changes then a default
   * should be avoided and a plan modifier should be used instead.
   */
  default?: SchemaDefault;

  /**
   * DeprecationMessage defines warning diagnostic details to display when
   * practitioner configurations use this Attribute. The warning diagnostic
   * summary is automatically set to "Attribute Deprecated" along with
   * configuration source file and line information.
   */
  deprecationMessage?: string;

  /**
   * Description is used in various tooling, like the language server, to
   * give practitioners more information about what this attribute is,
   * what it's for, and how it should be used. It should be written as
   * plain text, with no special formatting.
   */
  description?: string;

  /**
   * Optional indicates whether the practitioner can choose to enter a value
   * for this attribute or not. Optional and Required cannot both be true.
   */
  optional?: boolean;

  /**
   * MarkdownDescription is used in various tooling, like the
   * documentation generator, to give practitioners more information
   * about what this attribute is, what it's for, and how it should be
   * used. It should be formatted using Markdown.
   */
  markdownDescription?: string;

  /**
   * Plan modifiers defines a sequence of modifiers for this attribute at
   * plan time. Schema-based plan modifications occur before any
   * resource-level plan modifications.
   */
  planModifiers?: SchemaPlanModifier[];

  /**
   * Required indicates whether the practitioner must enter a value for
   * this attribute or not. Required and Optional cannot both be true,
   * and Required and Computed cannot both be true.
   */
  required?: boolean;

  /**
   * Sensitive indicates whether the value of this attribute should be
   * considered sensitive data. Setting it to true will obscure the value
   * in CLI output.
   */
  sensitive?: boolean;

  /**
   * TypeDef is the original OAS type definition, preserved for example
   * generation in data resource configuration examples.
   */
  typeDef?: TypeDef;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework datasource/schema.BoolAttribute.
 */
class DataResourceBoolAttribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    super("Bool", config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.DynamicAttribute.
 */
class DataResourceDynamicAttribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    super("Dynamic", config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.Float32Attribute.
 */
class DataResourceFloat32Attribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    const autoValidators = Float32SchemaValidator.fromTypeDef(config.typeDef);
    super("Float32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.Float64Attribute.
 */
class DataResourceFloat64Attribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    const autoValidators = Float64SchemaValidator.fromTypeDef(config.typeDef);
    super("Float64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.Int32Attribute.
 */
class DataResourceInt32Attribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    const autoValidators = Int32SchemaValidator.fromTypeDef(config.typeDef);
    super("Int32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.Int64Attribute.
 */
class DataResourceInt64Attribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    const autoValidators = Int64SchemaValidator.fromTypeDef(config.typeDef);
    super("Int64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.ListAttribute.
 */
class DataResourceListAttribute extends DataResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: DataResourceAttributeConfiguration,
  ) {
    const autoValidators = ListSchemaValidator.fromTypeDef(config.typeDef);
    super("List", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.ListNestedAttribute.
 */
class DataResourceListNestedAttribute extends DataResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: DataResourceNestedAttributeObject,
    config: DataResourceAttributeConfiguration,
  ) {
    super("ListNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.MapAttribute.
 */
class DataResourceMapAttribute extends DataResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: DataResourceAttributeConfiguration,
  ) {
    const autoValidators = MapSchemaValidator.fromTypeDef(config.typeDef);
    super("Map", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.MapNestedAttribute.
 */
class DataResourceMapNestedAttribute extends DataResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: DataResourceNestedAttributeObject,
    config: DataResourceAttributeConfiguration,
  ) {
    super("MapNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.NumberAttribute.
 */
class DataResourceNumberAttribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    super("Number", config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.ObjectAttribute.
 */
class DataResourceObjectAttribute extends DataResourceAttributeBaseWithAttributeTypes {
  constructor(
    attributeTypes: FrameworkTypeAttributeTypes,
    config: DataResourceAttributeConfiguration,
  ) {
    super("Object", attributeTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.SetAttribute.
 */
class DataResourceSetAttribute extends DataResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: DataResourceAttributeConfiguration,
  ) {
    const autoValidators = SetSchemaValidator.fromTypeDef(config.typeDef);
    super("Set", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.SetNestedAttribute.
 */
class DataResourceSetNestedAttribute extends DataResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: DataResourceNestedAttributeObject,
    config: DataResourceAttributeConfiguration,
  ) {
    super("SetNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.SingleNestedAttribute.
 */
class DataResourceSingleNestedAttribute extends DataResourceAttributeBaseWithAttributes {
  constructor(
    attributes: DataResourceAttributes,
    config: DataResourceAttributeConfiguration,
  ) {
    super("SingleNested", attributes, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.StringAttribute.
 */
class DataResourceStringAttribute extends DataResourceAttributeBase {
  constructor(config: DataResourceAttributeConfiguration) {
    const autoValidators = StringSchemaValidator.fromTypeDef(config.typeDef);
    super("String", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.TupleAttribute.
 */
class DataResourceTupleAttribute extends DataResourceAttributeBaseWithElementTypes {
  constructor(
    elementTypes: FrameworkTypeElementTypes,
    config: DataResourceAttributeConfiguration,
  ) {
    super("Tuple", elementTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework datasource/schema.NestedAttributeObject.
 */
class DataResourceNestedAttributeObject extends NestedAttributeObjectBase {
  constructor(config: DataResourceNestedAttributeObjectConfiguration) {
    super();
    this.attributes = config.attributes;
    this.validators = config.validators;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/datasource/schema",
    });
    return result;
  }
}

/**
 * Available configuration for DataResourceNestedAttributeObject.
 */
type DataResourceNestedAttributeObjectConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   */
  attributes: Record<string, DataResourceAttribute>;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework datasource/schema.Schema.
 */
class DataResourceSchema extends SchemaBase {
  constructor(config: DataResourceSchemaConfiguration) {
    super();
    this.attributes = config.attributes;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.markdownDescription = config.markdownDescription;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/datasource/schema",
    });
    return result;
  }
}

/**
 * Available configuration for DataResourceSchema.
 */
type DataResourceSchemaConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   *
   * Names must only contain lowercase letters, numbers, and underscores.
   */
  attributes: Record<string, DataResourceAttribute>;

  /**
   * DeprecationMessage defines warning diagnostic details to display when
   * practitioner configurations use this data resource.
   */
  deprecationMessage?: string;

  /**
   * Description is used in various tooling, like the language server, to
   * give practitioners more information about what this resource is,
   * what it's for, and how it should be used. It should be written as
   * plain text, with no special formatting.
   */
  description?: string;

  /**
   * MarkdownDescription is used in various tooling, like the
   * documentation generator, to give practitioners more information
   * about what this resource is, what it's for, and how it should be
   * used. It should be formatted using Markdown.
   */
  markdownDescription?: string;
};
