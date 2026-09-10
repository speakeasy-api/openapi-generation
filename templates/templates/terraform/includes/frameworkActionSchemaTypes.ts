/**
 * Collection of all terraform-plugin-framework action/schema.Attribute types.
 */
type ActionAttribute =
  | ActionBoolAttribute
  | ActionDynamicAttribute
  | ActionFloat32Attribute
  | ActionFloat64Attribute
  | ActionInt32Attribute
  | ActionInt64Attribute
  | ActionListAttribute
  | ActionListNestedAttribute
  | ActionMapAttribute
  | ActionMapNestedAttribute
  | ActionNumberAttribute
  | ActionObjectAttribute
  | ActionSetAttribute
  | ActionSetNestedAttribute
  | ActionSingleNestedAttribute
  | ActionStringAttribute;

/**
 * Create a ActionAttribute from a FieldDef.
 */
function ActionAttributeFromFieldDef(
  fieldDef: FieldDef,
  config: ActionAttributeConfiguration,
): ActionAttribute {
  return ActionAttributeFromTypeDef(fieldDef.Type, config);
}

/**
 * Create a ActionAttribute from a TypeDef.
 */
function ActionAttributeFromTypeDef(
  typeDef: TypeDef,
  config: ActionAttributeConfiguration,
): ActionAttribute {
  // Preserve the TypeDef for example generation
  const configWithTypeDef = { ...config, typeDef };

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new ActionStringAttribute(configWithTypeDef);
    case "boolean":
      return new ActionBoolAttribute(configWithTypeDef);
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
      return ActionAttributeFromTypeDef(typeDef.Enum.Type, enumConfig);
    }
    case "float32":
      return new ActionFloat32Attribute(configWithTypeDef);
    case "integer":
      return new ActionInt64Attribute(configWithTypeDef);
    case "int32":
      return new ActionInt32Attribute(configWithTypeDef);
    case "number":
      return new ActionFloat64Attribute(configWithTypeDef);
    case "map":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ActionMapNestedAttribute(
          new ActionNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ActionAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ActionMapAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "array":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ActionListNestedAttribute(
          new ActionNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ActionAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ActionListAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "set":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ActionSetNestedAttribute(
          new ActionNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ActionAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ActionSetAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "class":
    case "union":
      // TODO: Finish the implementation for AssociatedTypes/Fields.
      throw new Error(`Unimplemented: ${typeDef.Type}`);
      const attributes: ActionAttributes = {};
      return new ActionSingleNestedAttribute(attributes, configWithTypeDef);
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/**
 * Mapping of attribute names to action attributes.
 */
type ActionAttributes = Record<string, ActionAttribute>;

/**
 * Base representation of a terraform-plugin-framework
 * action/schema.Attribute.
 */
abstract class ActionAttributeBase extends AttributeBase {
  /** Original OAS TypeDef, preserved for example generation. */
  typeDef?: TypeDef;

  constructor(
    typeName: AttributeBaseTypeName,
    config: ActionAttributeConfiguration,
  ) {
    super(typeName);
    // action attributes intentionally do not implement computed.
    // action attributes intentionally do not implement default.
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.optional = config.optional;
    this.markdownDescription = config.markdownDescription;
    // action attributes intentionally do not implement planModifiers.
    this.required = config.required;
    this.sensitive = config.sensitive;
    this.typeDef = config.typeDef;
    this.validators = config.validators;
    this.writeOnly = config.writeOnly;

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
      Path: "github.com/hashicorp/terraform-plugin-framework/action/schema",
    });
    return result;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * action/schema.Attribute with Attributes (e.g. SingleNested).
 */
abstract class ActionAttributeBaseWithAttributes extends ActionAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributes: ActionAttributes,
    config: ActionAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributes = attributes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * action/schema.Attribute with AttributeTypes (e.g. Object).
 */
abstract class ActionAttributeBaseWithAttributeTypes extends ActionAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ActionAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributeTypes = attributeTypes;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * action/schema.Attribute with ElementType (e.g. List, Map, Set).
 */
abstract class ActionAttributeBaseWithElementType extends ActionAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementType: FrameworkType,
    config: ActionAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementType = elementType;
  }
}

/**
 * Base representation of a terraform-plugin-framework
 * action/schema.Attribute with NestedObject (e.g. ListNested).
 */
abstract class ActionAttributeBaseWithNestedObject extends ActionAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    nestedObject: ActionNestedAttributeObject,
    config: ActionAttributeConfiguration,
  ) {
    super(typeName, config);
    this.nestedObject = nestedObject;
  }
}

/**
 * Available configuration for ActionAttribute.
 */
type ActionAttributeConfiguration = {
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
   * this attribute or not. Required and Optional cannot both be true.
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
   * generation in action configuration examples.
   */
  typeDef?: TypeDef;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];

  /**
   * WriteOnly indicates that Terraform will not store this attribute value
   * in the plan or state artifacts. If WriteOnly is true, either Optional or
   * Required must also be true.
   *
   * This functionality is only supported in Terraform 1.11 and later.
   * Practitioners that choose a value for this attribute with older
   * versions of Terraform will receive an error.
   */
  writeOnly?: boolean;
};

/**
 * Representation of terraform-plugin-framework action/schema.BoolAttribute.
 */
class ActionBoolAttribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    super("Bool", config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.DynamicAttribute.
 */
class ActionDynamicAttribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    super("Dynamic", config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.Float32Attribute.
 */
class ActionFloat32Attribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    const autoValidators = Float32SchemaValidator.fromTypeDef(config.typeDef);
    super("Float32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.Float64Attribute.
 */
class ActionFloat64Attribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    const autoValidators = Float64SchemaValidator.fromTypeDef(config.typeDef);
    super("Float64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.Int32Attribute.
 */
class ActionInt32Attribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    const autoValidators = Int32SchemaValidator.fromTypeDef(config.typeDef);
    super("Int32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.Int64Attribute.
 */
class ActionInt64Attribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    const autoValidators = Int64SchemaValidator.fromTypeDef(config.typeDef);
    super("Int64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.ListAttribute.
 */
class ActionListAttribute extends ActionAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ActionAttributeConfiguration,
  ) {
    const autoValidators = ListSchemaValidator.fromTypeDef(config.typeDef);
    super("List", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.ListNestedAttribute.
 */
class ActionListNestedAttribute extends ActionAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ActionNestedAttributeObject,
    config: ActionAttributeConfiguration,
  ) {
    super("ListNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.MapAttribute.
 */
class ActionMapAttribute extends ActionAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ActionAttributeConfiguration,
  ) {
    const autoValidators = MapSchemaValidator.fromTypeDef(config.typeDef);
    super("Map", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.MapNestedAttribute.
 */
class ActionMapNestedAttribute extends ActionAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ActionNestedAttributeObject,
    config: ActionAttributeConfiguration,
  ) {
    super("MapNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.NumberAttribute.
 */
class ActionNumberAttribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    super("Number", config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.ObjectAttribute.
 */
class ActionObjectAttribute extends ActionAttributeBaseWithAttributeTypes {
  constructor(
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ActionAttributeConfiguration,
  ) {
    super("Object", attributeTypes, config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.SetAttribute.
 */
class ActionSetAttribute extends ActionAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ActionAttributeConfiguration,
  ) {
    const autoValidators = SetSchemaValidator.fromTypeDef(config.typeDef);
    super("Set", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.SetNestedAttribute.
 */
class ActionSetNestedAttribute extends ActionAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ActionNestedAttributeObject,
    config: ActionAttributeConfiguration,
  ) {
    super("SetNested", nestedObject, config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.SingleNestedAttribute.
 */
class ActionSingleNestedAttribute extends ActionAttributeBaseWithAttributes {
  constructor(
    attributes: ActionAttributes,
    config: ActionAttributeConfiguration,
  ) {
    super("SingleNested", attributes, config);
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.StringAttribute.
 */
class ActionStringAttribute extends ActionAttributeBase {
  constructor(config: ActionAttributeConfiguration) {
    const autoValidators = StringSchemaValidator.fromTypeDef(config.typeDef);
    super("String", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }
}

/**
 * Representation of terraform-plugin-framework action/schema.NestedAttributeObject.
 */
class ActionNestedAttributeObject extends NestedAttributeObjectBase {
  constructor(config: ActionNestedAttributeObjectConfiguration) {
    super();
    this.attributes = config.attributes;
    this.validators = config.validators;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/action/schema",
    });
    return result;
  }
}

/**
 * Available configuration for ActionNestedAttributeObject.
 */
type ActionNestedAttributeObjectConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   */
  attributes: Record<string, ActionAttribute>;

  /**
   * Validators define value validation functionality for the attribute.
   */
  validators?: SchemaValidator[];
};

/**
 * Representation of terraform-plugin-framework action/schema.Schema.
 */
class ActionSchema extends SchemaBase {
  constructor(config: ActionSchemaConfiguration) {
    super();
    this.attributes = config.attributes;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.markdownDescription = config.markdownDescription;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/action/schema",
    });
    return result;
  }
}

/**
 * Available configuration for ActionSchema.
 */
type ActionSchemaConfiguration = {
  /**
   * Attributes is the mapping of underlying attribute names to attribute
   * definitions.
   *
   * Names must only contain lowercase letters, numbers, and underscores.
   */
  attributes: Record<string, ActionAttribute>;

  /**
   * DeprecationMessage defines warning diagnostic details to display when
   * practitioner configurations use this action.
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
