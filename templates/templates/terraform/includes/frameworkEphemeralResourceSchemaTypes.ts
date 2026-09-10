/**
 * Collection of all terraform-plugin-framework ephemeral/schema.Attribute types.
 */
type EphemeralResourceAttribute =
  | EphemeralResourceBoolAttribute
  | EphemeralResourceDynamicAttribute
  | EphemeralResourceFloat32Attribute
  | EphemeralResourceFloat64Attribute
  | EphemeralResourceInt32Attribute
  | EphemeralResourceInt64Attribute
  | EphemeralResourceListAttribute
  | EphemeralResourceListNestedAttribute
  | EphemeralResourceMapAttribute
  | EphemeralResourceMapNestedAttribute
  | EphemeralResourceNumberAttribute
  | EphemeralResourceObjectAttribute
  | EphemeralResourceSetAttribute
  | EphemeralResourceSetNestedAttribute
  | EphemeralResourceSingleNestedAttribute
  | EphemeralResourceStringAttribute;

/**
 * Create a EphemeralResourceAttribute from a FieldDef.
 */
function EphemeralResourceAttributeFromFieldDef(
  fieldDef: FieldDef,
  config: EphemeralResourceAttributeConfiguration,
): EphemeralResourceAttribute {
  return EphemeralResourceAttributeFromTypeDef(fieldDef.Type, config);
}

/**
 * Create a EphemeralResourceAttribute from a TypeDef.
 */
function EphemeralResourceAttributeFromTypeDef(
  typeDef: TypeDef,
  config: EphemeralResourceAttributeConfiguration,
): EphemeralResourceAttribute {
  // Preserve the TypeDef for example generation
  const configWithTypeDef = { ...config, typeDef };

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new EphemeralResourceStringAttribute(configWithTypeDef);
    case "boolean":
      return new EphemeralResourceBoolAttribute(configWithTypeDef);
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
      return EphemeralResourceAttributeFromTypeDef(
        typeDef.Enum.Type,
        enumConfig,
      );
    }
    case "float32":
      return new EphemeralResourceFloat32Attribute(configWithTypeDef);
    case "integer":
      return new EphemeralResourceInt64Attribute(configWithTypeDef);
    case "int32":
      return new EphemeralResourceInt32Attribute(configWithTypeDef);
    case "number":
      return new EphemeralResourceFloat64Attribute(configWithTypeDef);
    case "map":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new EphemeralResourceMapNestedAttribute(
          new EphemeralResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  EphemeralResourceAttributeFromFieldDef(
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
      return new EphemeralResourceMapAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "array":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new EphemeralResourceListNestedAttribute(
          new EphemeralResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  EphemeralResourceAttributeFromFieldDef(
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
      return new EphemeralResourceListAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "set":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new EphemeralResourceSetNestedAttribute(
          new EphemeralResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  EphemeralResourceAttributeFromFieldDef(
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
      return new EphemeralResourceSetAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "class":
    case "union":
      // TODO: Finish the implementation for AssociatedTypes/Fields.
      throw new Error(`Unimplemented: ${typeDef.Type}`);
      const attributes: EphemeralResourceAttributes = {};
      return new EphemeralResourceSingleNestedAttribute(
        attributes,
        configWithTypeDef,
      );
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/**
 * Mapping of attribute names to ephemeral resource attributes.
 */
type EphemeralResourceAttributes = Record<string, EphemeralResourceAttribute>;

/**
 * Base representation of a terraform-plugin-framework
 * ephemeral/schema.Attribute.
 */
abstract class EphemeralResourceAttributeBase extends AttributeBase {
  /** Original OAS TypeDef, preserved for example generation. */
  typeDef?: TypeDef;

  constructor(
    typeName: AttributeBaseTypeName,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super(typeName);
    this.computed = config.computed;
    // ephemeral resource attributes intentionally do not implement default.
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.optional = config.optional;
    this.markdownDescription = config.markdownDescription;
    // ephemeral resource attributes intentionally do not implement planModifiers.
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
      Path: "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema",
    });
    return result;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * ephemeral/schema.Attribute with Attributes (e.g. SingleNested).
 */
abstract class EphemeralResourceAttributeBaseWithAttributes extends EphemeralResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributes: EphemeralResourceAttributes,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributes = attributes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * ephemeral/schema.Attribute with AttributeTypes (e.g. Object).
 */
abstract class EphemeralResourceAttributeBaseWithAttributeTypes extends EphemeralResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributeTypes: FrameworkTypeAttributeTypes,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributeTypes = attributeTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * ephemeral/schema.Attribute with ElementType (e.g. List, Map, Set).
 */
abstract class EphemeralResourceAttributeBaseWithElementType extends EphemeralResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementType: FrameworkType,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementType = elementType;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * ephemeral/schema.Attribute with NestedObject (e.g. ListNested).
 */
abstract class EphemeralResourceAttributeBaseWithNestedObject extends EphemeralResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    nestedObject: EphemeralResourceNestedAttributeObject,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.nestedObject = nestedObject;
  }
}

/**
 * Available configuration for EphemeralResourceAttribute.
 */
type EphemeralResourceAttributeConfiguration = {
  /**
   * Computed indicates whether the provider may return its own value for
   * this Attribute or not. Required and Computed cannot both be true. If
   * Required and Optional are both false, Computed must be true, and the
   * attribute will be considered "read only" for the practitioner, with
   * only the provider able to set its value.
   */
  computed?: boolean;

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
   * generation in ephemeral resource configuration examples.
   */
  typeDef?: TypeDef;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework ephemeral/schema.BoolAttribute.
 */
class EphemeralResourceBoolAttribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    super("Bool", config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.DynamicAttribute.
 */
class EphemeralResourceDynamicAttribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    super("Dynamic", config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.Float32Attribute.
 */
class EphemeralResourceFloat32Attribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    const autoValidators = Float32SchemaValidator.fromTypeDef(config.typeDef);
    super("Float32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.Float64Attribute.
 */
class EphemeralResourceFloat64Attribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    const autoValidators = Float64SchemaValidator.fromTypeDef(config.typeDef);
    super("Float64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.Int32Attribute.
 */
class EphemeralResourceInt32Attribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    const autoValidators = Int32SchemaValidator.fromTypeDef(config.typeDef);
    super("Int32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.Int64Attribute.
 */
class EphemeralResourceInt64Attribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    const autoValidators = Int64SchemaValidator.fromTypeDef(config.typeDef);
    super("Int64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.ListAttribute.
 */
class EphemeralResourceListAttribute extends EphemeralResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    const autoValidators = ListSchemaValidator.fromTypeDef(config.typeDef);
    super("List", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.ListNestedAttribute.
 */
class EphemeralResourceListNestedAttribute extends EphemeralResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: EphemeralResourceNestedAttributeObject,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super("ListNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.MapAttribute.
 */
class EphemeralResourceMapAttribute extends EphemeralResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    const autoValidators = MapSchemaValidator.fromTypeDef(config.typeDef);
    super("Map", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.MapNestedAttribute.
 */
class EphemeralResourceMapNestedAttribute extends EphemeralResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: EphemeralResourceNestedAttributeObject,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super("MapNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.NumberAttribute.
 */
class EphemeralResourceNumberAttribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    super("Number", config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.ObjectAttribute.
 */
class EphemeralResourceObjectAttribute extends EphemeralResourceAttributeBaseWithAttributeTypes {
  constructor(
    attributeTypes: FrameworkTypeAttributeTypes,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super("Object", attributeTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.SetAttribute.
 */
class EphemeralResourceSetAttribute extends EphemeralResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    const autoValidators = SetSchemaValidator.fromTypeDef(config.typeDef);
    super("Set", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.SetNestedAttribute.
 */
class EphemeralResourceSetNestedAttribute extends EphemeralResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: EphemeralResourceNestedAttributeObject,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super("SetNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.SingleNestedAttribute.
 */
class EphemeralResourceSingleNestedAttribute extends EphemeralResourceAttributeBaseWithAttributes {
  constructor(
    attributes: EphemeralResourceAttributes,
    config: EphemeralResourceAttributeConfiguration,
  ) {
    super("SingleNested", attributes, config);
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.StringAttribute.
 */
class EphemeralResourceStringAttribute extends EphemeralResourceAttributeBase {
  constructor(config: EphemeralResourceAttributeConfiguration) {
    const autoValidators = StringSchemaValidator.fromTypeDef(config.typeDef);
    super("String", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework ephemeral/schema.NestedAttributeObject.
 */
class EphemeralResourceNestedAttributeObject extends NestedAttributeObjectBase {
  constructor(config: EphemeralResourceNestedAttributeObjectConfiguration) {
    super();
    this.attributes = config.attributes;
    this.validators = config.validators;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema",
    });
    return result;
  }
}

/**
 * Available configuration for EphemeralResourceNestedAttributeObject.
 */
type EphemeralResourceNestedAttributeObjectConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   */
  attributes: Record<string, EphemeralResourceAttribute>;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework ephemeral/schema.Schema.
 */
class EphemeralResourceSchema extends SchemaBase {
  constructor(config: EphemeralResourceSchemaConfiguration) {
    super();
    this.attributes = config.attributes;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.markdownDescription = config.markdownDescription;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema",
    });
    return result;
  }
}

/**
 * Available configuration for EphemeralResourceSchema.
 */
type EphemeralResourceSchemaConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   *
   * Names must only contain lowercase letters, numbers, and underscores.
   */
  attributes: Record<string, EphemeralResourceAttribute>;

  /**
   * DeprecationMessage defines warning diagnostic details to display when
   * practitioner configurations use this ephemeral resource.
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
