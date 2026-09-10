/** Collection of all terraform-plugin-framework provider/schema.Attribute
 *  types. */
type ProviderAttribute =
  | ProviderBoolAttribute
  | ProviderDynamicAttribute
  | ProviderFloat32Attribute
  | ProviderFloat64Attribute
  | ProviderInt32Attribute
  | ProviderInt64Attribute
  | ProviderListAttribute
  | ProviderListNestedAttribute
  | ProviderMapAttribute
  | ProviderMapNestedAttribute
  | ProviderNumberAttribute
  | ProviderObjectAttribute
  | ProviderSetAttribute
  | ProviderSetNestedAttribute
  | ProviderSingleNestedAttribute
  | ProviderStringAttribute
  | ProviderTupleAttribute;

/** Create a ProviderAttribute from a FieldDef. */
function ProviderAttributeFromFieldDef(
  fieldDef: FieldDef,
  config: ProviderAttributeConfiguration,
): ProviderAttribute {
  return ProviderAttributeFromTypeDef(fieldDef.Type, config);
}

/** Create a ProviderAttribute from a TypeDef. */
function ProviderAttributeFromTypeDef(
  typeDef: TypeDef,
  config: ProviderAttributeConfiguration,
): ProviderAttribute {
  // Preserve the TypeDef for example generation
  const configWithTypeDef = { ...config, typeDef };

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new ProviderStringAttribute(configWithTypeDef);
    case "boolean":
      return new ProviderBoolAttribute(configWithTypeDef);
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
      return ProviderAttributeFromTypeDef(typeDef.Enum.Type, enumConfig);
    }
    case "float32":
      return new ProviderFloat32Attribute(configWithTypeDef);
    case "integer":
      return new ProviderInt64Attribute(configWithTypeDef);
    case "int32":
      return new ProviderInt32Attribute(configWithTypeDef);
    case "number":
      return new ProviderFloat64Attribute(configWithTypeDef);
    case "map":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ProviderMapNestedAttribute(
          new ProviderNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ProviderAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ProviderMapAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "array":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ProviderListNestedAttribute(
          new ProviderNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ProviderAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ProviderListAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "set":
      if (
        typeDef.ItemType?.Type.toString() === "class" ||
        typeDef.ItemType?.Type.toString() === "union"
      ) {
        return new ProviderSetNestedAttribute(
          new ProviderNestedAttributeObject({
            attributes: Object.fromEntries(
              Object.entries(typeDef.ItemType.Fields).map(
                ([name, fieldDef]) => [
                  name,
                  ProviderAttributeFromFieldDef(fieldDef, configWithTypeDef),
                ],
              ),
            ),
          }),
          configWithTypeDef,
        );
      }
      return new ProviderSetAttribute(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        configWithTypeDef,
      );
    case "class":
    case "union":
      // TODO: Finish the implementation for AssociatedTypes/Fields.
      throw new Error(`Unimplemented: ${typeDef.Type}`);
      const attributes: ProviderAttributes = {};
      return new ProviderSingleNestedAttribute(attributes, configWithTypeDef);
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/** Mapping of attribute names to provider attributes. */
type ProviderAttributes = Record<string, ProviderAttribute>;

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute. */
abstract class ProviderAttributeBase extends AttributeBase {
  /** Original OAS TypeDef, preserved for example generation. */
  typeDef?: TypeDef;

  constructor(
    typeName: AttributeBaseTypeName,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName);
    // Provider attributes intentionally do not implement computed.
    // Provider attributes intentionally do not implement default.
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.optional = config.optional;
    this.markdownDescription = config.markdownDescription;
    // Provider attributes intentionally do not implement planModifiers.
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
      Path: "github.com/hashicorp/terraform-plugin-framework/provider/schema",
    });
    return result;
  }

  /** Returns an HCL-formatted zero/default value for this attribute type.
   *  Used for provider example generation when no OAS example is available. */
  abstract templateExampleZeroValue(): string;
}

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute with Attributes (e.g. SingleNested). */
abstract class ProviderAttributeBaseWithAttributes extends ProviderAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributes: ProviderAttributes,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributes = attributes;
  }
}

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute with AttributeTypes (e.g. Object). */
abstract class ProviderAttributeBaseWithAttributeTypes extends ProviderAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName, config);
    this.attributeTypes = attributeTypes;
  }
}

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute with ElementType (e.g. List, Map, Set). */
abstract class ProviderAttributeBaseWithElementType extends ProviderAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementType: FrameworkType,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementType = elementType;
  }
}

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute with ElementTypes (e.g. Tuple). */
abstract class ProviderAttributeBaseWithElementTypes extends ProviderAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    elementTypes: FrameworkTypeElementTypes,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName, config);
    this.elementTypes = elementTypes;
  }
}

/** Base representation of a terraform-plugin-framework
 *  provider/schema.Attribute with NestedObject (e.g. ListNested). */
abstract class ProviderAttributeBaseWithNestedObject extends ProviderAttributeBase {
  constructor(
    typeName: AttributeBaseTypeName,
    nestedObject: ProviderNestedAttributeObject,
    config: ProviderAttributeConfiguration,
  ) {
    super(typeName, config);
    this.nestedObject = nestedObject;
  }
}

/** Available configuration for ProviderAttribute. */
type ProviderAttributeConfiguration = {
  /** DeprecationMessage defines warning diagnostic details to display when
   *  practitioner configurations use this Attribute. The warning diagnostic
   *  summary is automatically set to "Attribute Deprecated" along with
   *  configuration source file and line information. */
  deprecationMessage?: string;

  /** Description is used in various tooling, like the language server, to
   *  give practitioners more information about what this attribute is,
   *  what it's for, and how it should be used. It should be written as
   *  plain text, with no special formatting. */
  description?: string;

  /** Optional indicates whether the practitioner can choose to enter a value
   *  for this attribute or not. Optional and Required cannot both be true. */
  optional?: boolean;

  /** MarkdownDescription is used in various tooling, like the
   *  documentation generator, to give practitioners more information
   *  about what this attribute is, what it's for, and how it should be
   *  used. It should be formatted using Markdown. */
  markdownDescription?: string;

  /** Required indicates whether the practitioner must enter a value for
   *  this attribute or not. Required and Optional cannot both be true,
   *  and Required and Computed cannot both be true. */
  required?: boolean;

  /** Sensitive indicates whether the value of this attribute should be
   *  considered sensitive data. Setting it to true will obscure the value
   *  in CLI output. */
  sensitive?: boolean;

  /** TypeDef is the original OAS type definition, preserved for example
   *  generation in provider configuration examples. */
  typeDef?: TypeDef;

  /** Validators define value validation functionality for the attribute. */
  validators?: SchemaValidator[];
};

/** Representation of terraform-plugin-framework provider/schema.BoolAttribute. */
class ProviderBoolAttribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    super("Bool", config);
  }

  templateExampleZeroValue(): string {
    return "false";
  }
}

/** Representation of terraform-plugin-framework provider/schema.DynamicAttribute. */
class ProviderDynamicAttribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    super("Dynamic", config);
  }

  templateExampleZeroValue(): string {
    return `"..."`;
  }
}

/** Representation of terraform-plugin-framework provider/schema.Float32Attribute. */
class ProviderFloat32Attribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    const autoValidators = Float32SchemaValidator.fromTypeDef(config.typeDef);
    super("Float32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "0.0";
  }
}

/** Representation of terraform-plugin-framework provider/schema.Float64Attribute. */
class ProviderFloat64Attribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    const autoValidators = Float64SchemaValidator.fromTypeDef(config.typeDef);
    super("Float64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "0.0";
  }
}

/** Representation of terraform-plugin-framework provider/schema.Int32Attribute. */
class ProviderInt32Attribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    const autoValidators = Int32SchemaValidator.fromTypeDef(config.typeDef);
    super("Int32", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "0";
  }
}

/** Representation of terraform-plugin-framework provider/schema.Int64Attribute. */
class ProviderInt64Attribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    const autoValidators = Int64SchemaValidator.fromTypeDef(config.typeDef);
    super("Int64", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "0";
  }
}

/** Representation of terraform-plugin-framework provider/schema.ListAttribute. */
class ProviderListAttribute extends ProviderAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ProviderAttributeConfiguration,
  ) {
    const autoValidators = ListSchemaValidator.fromTypeDef(config.typeDef);
    super("List", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "[]";
  }
}

/** Representation of terraform-plugin-framework provider/schema.ListNestedAttribute. */
class ProviderListNestedAttribute extends ProviderAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ProviderNestedAttributeObject,
    config: ProviderAttributeConfiguration,
  ) {
    super("ListNested", nestedObject, config);
  }

  templateExampleZeroValue(): string {
    return "[]";
  }
}

/** Representation of terraform-plugin-framework provider/schema.MapAttribute. */
class ProviderMapAttribute extends ProviderAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ProviderAttributeConfiguration,
  ) {
    const autoValidators = MapSchemaValidator.fromTypeDef(config.typeDef);
    super("Map", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "{}";
  }
}

/** Representation of terraform-plugin-framework provider/schema.MapNestedAttribute. */
class ProviderMapNestedAttribute extends ProviderAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ProviderNestedAttributeObject,
    config: ProviderAttributeConfiguration,
  ) {
    super("MapNested", nestedObject, config);
  }

  templateExampleZeroValue(): string {
    return "{}";
  }
}

/** Representation of terraform-plugin-framework provider/schema.NumberAttribute. */
class ProviderNumberAttribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    super("Number", config);
  }

  templateExampleZeroValue(): string {
    return "0.0";
  }
}

/** Representation of terraform-plugin-framework provider/schema.ObjectAttribute. */
class ProviderObjectAttribute extends ProviderAttributeBaseWithAttributeTypes {
  constructor(
    attributeTypes: FrameworkTypeAttributeTypes,
    config: ProviderAttributeConfiguration,
  ) {
    super("Object", attributeTypes, config);
  }

  templateExampleZeroValue(): string {
    return "{}";
  }
}

/** Representation of terraform-plugin-framework provider/schema.SetAttribute. */
class ProviderSetAttribute extends ProviderAttributeBaseWithElementType {
  constructor(
    elementType: FrameworkType,
    config: ProviderAttributeConfiguration,
  ) {
    const autoValidators = SetSchemaValidator.fromTypeDef(config.typeDef);
    super("Set", elementType, {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return "[]";
  }
}

/** Representation of terraform-plugin-framework provider/schema.SetNestedAttribute. */
class ProviderSetNestedAttribute extends ProviderAttributeBaseWithNestedObject {
  constructor(
    nestedObject: ProviderNestedAttributeObject,
    config: ProviderAttributeConfiguration,
  ) {
    super("SetNested", nestedObject, config);
  }

  templateExampleZeroValue(): string {
    return "[]";
  }
}

/** Representation of terraform-plugin-framework provider/schema.SingleNestedAttribute. */
class ProviderSingleNestedAttribute extends ProviderAttributeBaseWithAttributes {
  constructor(
    attributes: ProviderAttributes,
    config: ProviderAttributeConfiguration,
  ) {
    super("SingleNested", attributes, config);
  }

  templateExampleZeroValue(): string {
    return "{}";
  }
}

/** Representation of terraform-plugin-framework provider/schema.StringAttribute. */
class ProviderStringAttribute extends ProviderAttributeBase {
  constructor(config: ProviderAttributeConfiguration) {
    const autoValidators = StringSchemaValidator.fromTypeDef(config.typeDef);
    super("String", {
      ...config,
      validators: [...(config.validators || []), ...autoValidators],
    });
  }

  templateExampleZeroValue(): string {
    return `"..."`;
  }
}

/** Representation of terraform-plugin-framework provider/schema.TupleAttribute. */
class ProviderTupleAttribute extends ProviderAttributeBaseWithElementTypes {
  constructor(
    elementTypes: FrameworkTypeElementTypes,
    config: ProviderAttributeConfiguration,
  ) {
    super("Tuple", elementTypes, config);
  }

  templateExampleZeroValue(): string {
    return "[]";
  }
}

/** Representation of terraform-plugin-framework provider/schema.NestedAttributeObject. */
class ProviderNestedAttributeObject extends NestedAttributeObjectBase {
  constructor(config: ProviderNestedAttributeObjectConfiguration) {
    super();
    this.attributes = config.attributes;
    this.validators = config.validators;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/provider/schema",
    });
    return result;
  }
}

/** Available configuration for ProviderNestedAttributeObject. */
type ProviderNestedAttributeObjectConfiguration = {
  /** Attributes is the mapping of underlying attribute names to attribute
   *  definitions. */
  attributes: Record<string, ProviderAttribute>;

  /** Validators define value validation functionality for the attribute. */
  validators?: SchemaValidator[];
};

/** Representation of terraform-plugin-framework provider/schema.Schema. */
class ProviderSchema extends SchemaBase {
  constructor(config: ProviderSchemaConfiguration) {
    super();
    this.attributes = config.attributes;
    this.deprecationMessage = config.deprecationMessage;
    this.description = config.description;
    this.markdownDescription = config.markdownDescription;
  }

  imports(): GoImport[] {
    let result = super.imports();
    result.push({
      Path: "github.com/hashicorp/terraform-plugin-framework/provider/schema",
    });
    return result;
  }
}

/** Available configuration for ProviderSchema. */
type ProviderSchemaConfiguration = {
  /** Attributes is the mapping of underlying attribute names to attribute
   *  definitions.
   *
   *  Names must only contain lowercase letters, numbers, and underscores. */
  attributes: Record<string, ProviderAttribute>;

  /** DeprecationMessage defines warning diagnostic details to display when
   *  practitioner configurations use this provider. */
  deprecationMessage?: string;

  /** Description is used in various tooling, like the language server, to
   *  give practitioners more information about what this provider is,
   *  what it's for, and how it should be used. It should be written as
   *  plain text, with no special formatting. */
  description?: string;

  /** MarkdownDescription is used in various tooling, like the
   *  documentation generator, to give practitioners more information
   *  about what this provider is, what it's for, and how it should be
   *  used. It should be formatted using Markdown. */
  markdownDescription?: string;
};
