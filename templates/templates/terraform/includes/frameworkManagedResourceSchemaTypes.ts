/**
 * Collection of all terraform-plugin-framework resource/schema.Attribute types.
 */
type ManagedResourceAttribute =
  | ManagedResourceBoolAttribute
  | ManagedResourceDynamicAttribute
  | ManagedResourceFloat32Attribute
  | ManagedResourceFloat64Attribute
  | ManagedResourceInt32Attribute
  | ManagedResourceInt64Attribute
  | ManagedResourceListAttribute
  | ManagedResourceListNestedAttribute
  | ManagedResourceMapAttribute
  | ManagedResourceMapNestedAttribute
  | ManagedResourceNumberAttribute
  | ManagedResourceObjectAttribute
  | ManagedResourceSetAttribute
  | ManagedResourceSetNestedAttribute
  | ManagedResourceSingleNestedAttribute
  | ManagedResourceStringAttribute
  | ManagedResourceTupleAttribute;

/**
 * Create a ManagedResourceAttribute from a FieldDef.
 */
function ManagedResourceAttributeFromFieldDef(
  fieldDef: FieldDef,
  config: ManagedResourceAttributeConfiguration,
): ManagedResourceAttribute {
  return ManagedResourceAttributeFromTypeDef(fieldDef.Type, config);
}

/**
 * Create a ManagedResourceAttribute from a TypeDef.
 */
function ManagedResourceAttributeFromTypeDef(
  typeDef: TypeDef,
  config: ManagedResourceAttributeConfiguration,
): ManagedResourceAttribute {
  // Preserve the TypeDef for example generation
  const configWithTypeDef = { ...config, typeDef };

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new ManagedResourceStringAttribute(configWithTypeDef);
    case "boolean":
      return new ManagedResourceBoolAttribute(configWithTypeDef);
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
      return ManagedResourceAttributeFromTypeDef(typeDef.Enum.Type, enumConfig);
    }
    case "float32":
      return new ManagedResourceFloat32Attribute(configWithTypeDef);
    case "integer":
      return new ManagedResourceInt64Attribute(configWithTypeDef);
    case "int32":
      return new ManagedResourceInt32Attribute(configWithTypeDef);
    case "number":
      return new ManagedResourceFloat64Attribute(configWithTypeDef);
    case "map":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ManagedResourceMapNestedAttribute(
          new ManagedResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ManagedResourceAttributeFromFieldDef(
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
      return new ManagedResourceMapAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "array":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ManagedResourceListNestedAttribute(
          new ManagedResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ManagedResourceAttributeFromFieldDef(
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
      return new ManagedResourceListAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "set":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ManagedResourceSetNestedAttribute(
          new ManagedResourceNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ManagedResourceAttributeFromFieldDef(
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
      return new ManagedResourceSetAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "class":
    case "union":
      // TODO: Finish the implementation for AssociatedTypes/Fields.
      throw new Error(`Unimplemented: ${typeDef.Type}`);
      const attributes: ManagedResourceAttributes = {};
      return new ManagedResourceSingleNestedAttribute(
        attributes,
        configWithTypeDef,
      );
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/**
 * Mapping of attribute names to managed resource attributes.
 */
type ManagedResourceAttributes = Record<string, ManagedResourceAttribute>;

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute.
 */
abstract class ManagedResourceAttributeBase extends AttributeBase {
  /** Original OAS TypeDef, preserved for example generation. */
  typeDef?: TypeDef;

  constructor(
    typeName: AttributeBaseTypeName,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName);
    this.computed = config.computed;
    this.default = config.default;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.optional = config.optional;
    this.markdownDescription = config.markdownDescription;
    this.planModifiers = config.planModifiers;
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
      Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema",
    });
    return result;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute with Attributes (e.g. SingleNested).
 */
abstract class ManagedResourceAttributeBaseWithAttributes extends ManagedResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributes: ManagedResourceAttributes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributes = attributes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute with AttributeTypes (e.g. Object).
 */
abstract class ManagedResourceAttributeBaseWithAttributeTypes extends ManagedResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributeTypes = attributeTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute with ElementType (e.g. List, Map, Set).
 */
abstract class ManagedResourceAttributeBaseWithElementType extends ManagedResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementType: FrameworkType,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementType = elementType;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute with ElementTypes (e.g. Tuple).
 */
abstract class ManagedResourceAttributeBaseWithElementTypes extends ManagedResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementTypes: FrameworkTypeElementTypes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementTypes = elementTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * resource/schema.Attribute with NestedObject (e.g. ListNested).
 */
abstract class ManagedResourceAttributeBaseWithNestedObject extends ManagedResourceAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    nestedObject: ManagedResourceNestedAttributeObject,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super(typeName, config);
    this.nestedObject = nestedObject;
  }
}

/**
 * Available configuration for ManagedResourceAttribute.
 */
type ManagedResourceAttributeConfiguration = {
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
   * generation in managed resource configuration examples.
   */
  typeDef?: TypeDef;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];

  /**
   * WriteOnly indicates that Terraform will not store this attribute value
   * in the plan or state artifacts. If WriteOnly is true, either Optional or
   * Required must also be true. WriteOnly cannot be set with Computed.
   *
   * This functionality is only supported in Terraform 1.11 and later.
   * Practitioners that choose a value for this attribute with older
   * versions of Terraform will receive an error.
   */
  writeOnly?: boolean;
};

/**
 * Representation of terraform-plugin-framework resource/schema.BoolAttribute.
 */
class ManagedResourceBoolAttribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    super("Bool", config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.DynamicAttribute.
 */
class ManagedResourceDynamicAttribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    super("Dynamic", config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.Float32Attribute.
 */
class ManagedResourceFloat32Attribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    const autoValidators = Float32SchemaValidator.fromTypeDef(config.typeDef);
    super("Float32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.Float64Attribute.
 */
class ManagedResourceFloat64Attribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    const autoValidators = Float64SchemaValidator.fromTypeDef(config.typeDef);
    super("Float64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.Int32Attribute.
 */
class ManagedResourceInt32Attribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    const autoValidators = Int32SchemaValidator.fromTypeDef(config.typeDef);
    super("Int32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.Int64Attribute.
 */
class ManagedResourceInt64Attribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    const autoValidators = Int64SchemaValidator.fromTypeDef(config.typeDef);
    super("Int64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.ListAttribute.
 */
class ManagedResourceListAttribute extends ManagedResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ManagedResourceAttributeConfiguration,
  ) {
    const autoValidators = ListSchemaValidator.fromTypeDef(config.typeDef);
    super("List", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.ListNestedAttribute.
 */
class ManagedResourceListNestedAttribute extends ManagedResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ManagedResourceNestedAttributeObject,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("ListNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.MapAttribute.
 */
class ManagedResourceMapAttribute extends ManagedResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ManagedResourceAttributeConfiguration,
  ) {
    const autoValidators = MapSchemaValidator.fromTypeDef(config.typeDef);
    super("Map", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.MapNestedAttribute.
 */
class ManagedResourceMapNestedAttribute extends ManagedResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ManagedResourceNestedAttributeObject,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("MapNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.NumberAttribute.
 */
class ManagedResourceNumberAttribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    super("Number", config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.ObjectAttribute.
 */
class ManagedResourceObjectAttribute extends ManagedResourceAttributeBaseWithAttributeTypes {
  constructor(
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("Object", attributeTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.SetAttribute.
 */
class ManagedResourceSetAttribute extends ManagedResourceAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ManagedResourceAttributeConfiguration,
  ) {
    const autoValidators = SetSchemaValidator.fromTypeDef(config.typeDef);
    super("Set", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.SetNestedAttribute.
 */
class ManagedResourceSetNestedAttribute extends ManagedResourceAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ManagedResourceNestedAttributeObject,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("SetNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.SingleNestedAttribute.
 */
class ManagedResourceSingleNestedAttribute extends ManagedResourceAttributeBaseWithAttributes {
  constructor(
    attributes: ManagedResourceAttributes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("SingleNested", attributes, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.StringAttribute.
 */
class ManagedResourceStringAttribute extends ManagedResourceAttributeBase {
  constructor(config: ManagedResourceAttributeConfiguration) {
    const autoValidators = StringSchemaValidator.fromTypeDef(config.typeDef);
    super("String", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.TupleAttribute.
 */
class ManagedResourceTupleAttribute extends ManagedResourceAttributeBaseWithElementTypes {
  constructor(
    elementTypes: FrameworkTypeElementTypes,
    config: ManagedResourceAttributeConfiguration,
  ) {
    super("Tuple", elementTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework resource/schema.NestedAttributeObject.
 */
class ManagedResourceNestedAttributeObject extends NestedAttributeObjectBase {
  constructor(config: ManagedResourceNestedAttributeObjectConfiguration) {
    super();
    this.attributes = config.attributes;
    this.validators = config.validators;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema",
    });
    return result;
  }
}

/**
 * Available configuration for ManagedResourceNestedAttributeObject.
 */
type ManagedResourceNestedAttributeObjectConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   */
  attributes: Record<string, ManagedResourceAttribute>;

  /**
   * Plan modifiers defines a sequence of modifiers for this attribute at
   * plan time. Schema-based plan modifications occur before any
   * resource-level plan modifications.
   */
  planModifiers?: SchemaPlanModifier[];

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework resource/schema.Schema.
 */
class ManagedResourceSchema extends SchemaBase {
  constructor(config: ManagedResourceSchemaConfiguration) {
    super();
    this.attributes = config.attributes;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.markdownDescription = config.markdownDescription;
    this.version = config.version;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema",
    });
    return result;
  }
}

/**
 * Available configuration for ManagedResourceSchema.
 */
type ManagedResourceSchemaConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   *
   * Names must only contain lowercase letters, numbers, and underscores.
   */
  attributes: Record<string, ManagedResourceAttribute>;

  /**
   * DeprecationMessage defines warning diagnostic details to display when
   * practitioner configurations use this managed resource.
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

  /**
   * Version indicates the current version of the resource schema. Resource
   * schema versioning enables state upgrades in conjunction with the
   * [resource.ResourceWithStateUpgrades] interface. Versioning is only
   * required if there is a breaking change involving existing state data,
   * such as changing an attribute or block type in a manner that is
   * incompatible with the Terraform type.
   */
  version?: number;
};
