/**
 * Base class for all schema defaults. Only abstract classes for schema types,
 * such as StringSchemaDefault, BoolSchemaDefault, etc., should extend this
 * class. Concrete default implementations for specific default behaviors,
 * such as StringSchemaDefaultStatic, should extend the abstract schema type
 * default classes. This structure ensures type safety and organization of
 * defaults by their applicable schema types.
 */
abstract class SchemaDefault {
  /**
   * Go framework default type name (e.g., "Bool", "String"). Used to
   * generate the Go type reference `defaults.{DefaultTypeName}`.
   */
  abstract readonly defaultTypeName: string;

  /**
   * Provides an optional description for inclusion in attribute descriptions.
   * Concrete defaults should override this method to add a description of
   * the default value (e.g., "Default: true") should it be suitable for
   * attribute descriptions.
   */
  description(): string {
    return "";
  }

  /**
   * Returns an array of Go imports required by the default. Concrete
   * defaults should override this method with their specific imports.
   *
   * Unlike SchemaValidator which includes the base validator import
   * (github.com/hashicorp/terraform-plugin-framework/schema/validator)
   * because generated Go code references validator.{TypeName} types, the
   * defaults base import is not included here because generated Go code
   * does not reference defaults.{TypeName} types directly.
   */
  imports(): GoImport[] {
    return [];
  }

  /**
   * Returns the Go code for the default. Concrete defaults must override
   * this method.
   */
  template(): string {
    throw new Error("SchemaDefault template() not extended");
  }
}

/**
 * Creates a SchemaDefault instance from the given TypeDef, default value, and
 * configuration. Returns undefined if no default should be generated.
 *
 * This factory function handles all dispatch logic including guard conditions
 * (global fields, hoisted fields), custom default extensions, and type-based
 * dispatch to the appropriate concrete SchemaDefault class.
 */
function createSchemaDefault(
  typeDef: TypeDef,
  defaultValue: AnyValue | undefined,
  isGlobalField: boolean,
  fieldConfig: TypeDefConfig,
): SchemaDefault | undefined {
  // Global fields with default have the default value set at the provider
  // level. Including the default in the resource schema is invalid and prevents
  // the provider value from correctly being set in the resource logic.
  if (isGlobalField) {
    return undefined;
  }

  // Hoisted fields should not have a Default because they get their value from
  // the active oneOf variant via the UseHoistedValue plan modifier. If we set a
  // Default, it would override the plan value before the plan modifier runs,
  // preventing the variant's value from being used.
  if (fieldConfig.HoistedFrom && fieldConfig.HoistedFrom.length > 0) {
    return undefined;
  }

  const customDefaultConfig = typeDef.Extensions?.TerraformCustomDefault;

  if (customDefaultConfig) {
    return new CustomSchemaDefault(customDefaultConfig);
  }

  if (!defaultValue || defaultValue.Value === "null") {
    return undefined;
  }

  switch (typeDef.Type.toString()) {
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return typeof defaultValue.Value === "string"
        ? new StringSchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "boolean":
      return typeof defaultValue.Value === "boolean"
        ? new BoolSchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "enum":
      return createSchemaDefault(
        typeDef.Enum.Type,
        defaultValue,
        isGlobalField,
        fieldConfig,
      );
    case "float32":
      return typeof defaultValue.Value === "number"
        ? new Float32SchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "int32":
      return typeof defaultValue.Value === "number" &&
        Number.isInteger(defaultValue.Value)
        ? new Int32SchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "integer":
      return typeof defaultValue.Value === "number" &&
        Number.isInteger(defaultValue.Value)
        ? new Int64SchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "number":
      return typeof defaultValue.Value === "number"
        ? new Float64SchemaDefaultStatic(defaultValue.Value)
        : undefined;
    case "set": {
      // Exclude set-nested (class/union item types) which become
      // SetNestedAttribute and do not support defaults.
      const setItemType = typeDef.ItemType?.Type.toString();
      if (setItemType === "class" || setItemType === "union") {
        return undefined;
      }
      return Array.isArray(defaultValue.Value)
        ? SetSchemaDefault.fromTypeDef(typeDef, defaultValue.Value)
        : undefined;
    }
    case "array": {
      // Exclude list-nested (class/union item types) which become
      // ListNestedAttribute and do not support defaults.
      const listItemType = typeDef.ItemType?.Type.toString();
      if (listItemType === "class" || listItemType === "union") {
        return undefined;
      }
      return Array.isArray(defaultValue.Value)
        ? ListSchemaDefault.fromTypeDef(typeDef, defaultValue.Value)
        : undefined;
    }
    case "class":
    case "union":
      // Loose equality to match both null and undefined.
      return defaultValue.Value == null
        ? new SingleNestedSchemaDefaultNull(typeDef)
        : undefined;
    default:
      return undefined;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Bool
 * implementations. Ensures type safety when used with BoolAttribute schema
 * types.
 */
abstract class BoolSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Bool" = "Bool";
}

/**
 * terraform-plugin-framework resource/schema/booldefault.StaticBool(). Sets a
 * static boolean default value for a BoolAttribute.
 */
class BoolSchemaDefaultStatic extends BoolSchemaDefault {
  private readonly value: boolean;

  constructor(value: boolean) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault",
      },
    ];
  }

  template(): string {
    return `booldefault.StaticBool(${this.value})`;
  }
}

/**
 * Custom user-defined default from x-speakeasy-terraform-custom-default
 * extension configuration. Can apply to any attribute type.
 */
class CustomSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "" = "";
  private readonly config: TerraformCustomDefault;

  constructor(config: TerraformCustomDefault) {
    super();
    this.config = config;
  }

  imports(): GoImport[] {
    const result: GoImport[] = [];

    if (this.config.Imports) {
      this.config.Imports.forEach((importPath) => {
        result.push({ Path: importPath });
      });
    }

    return result;
  }

  template(): string {
    return this.config.SchemaDefinition;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Float32
 * implementations. Ensures type safety when used with Float32Attribute schema
 * types.
 */
abstract class Float32SchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Float32" = "Float32";
}

/**
 * terraform-plugin-framework resource/schema/float32default.StaticFloat32().
 * Sets a static float32 default value for a Float32Attribute.
 */
class Float32SchemaDefaultStatic extends Float32SchemaDefault {
  private readonly value: number;

  constructor(value: number) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default",
      },
    ];
  }

  template(): string {
    return `float32default.StaticFloat32(${this.value})`;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Float64
 * implementations. Ensures type safety when used with Float64Attribute schema
 * types.
 */
abstract class Float64SchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Float64" = "Float64";
}

/**
 * terraform-plugin-framework resource/schema/float64default.StaticFloat64().
 * Sets a static float64 default value for a Float64Attribute.
 */
class Float64SchemaDefaultStatic extends Float64SchemaDefault {
  private readonly value: number;

  constructor(value: number) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default",
      },
    ];
  }

  template(): string {
    return `float64default.StaticFloat64(${this.value})`;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Int32
 * implementations. Ensures type safety when used with Int32Attribute schema
 * types.
 */
abstract class Int32SchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Int32" = "Int32";
}

/**
 * terraform-plugin-framework resource/schema/int32default.StaticInt32(). Sets
 * a static int32 default value for an Int32Attribute.
 */
class Int32SchemaDefaultStatic extends Int32SchemaDefault {
  private readonly value: number;

  constructor(value: number) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default",
      },
    ];
  }

  template(): string {
    return `int32default.StaticInt32(${this.value})`;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Int64
 * implementations. Ensures type safety when used with Int64Attribute schema
 * types.
 */
abstract class Int64SchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Int64" = "Int64";
}

/**
 * terraform-plugin-framework resource/schema/int64default.StaticInt64(). Sets
 * a static int64 default value for an Int64Attribute.
 */
class Int64SchemaDefaultStatic extends Int64SchemaDefault {
  private readonly value: number;

  constructor(value: number) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
      },
    ];
  }

  template(): string {
    return `int64default.StaticInt64(${this.value})`;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.List
 * implementations. Ensures type safety when used with ListAttribute schema
 * types.
 */
abstract class ListSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "List" = "List";

  /**
   * Returns a ListSchemaDefault instance for the given TypeDef and values.
   * Supports primitive element types with values and empty arrays for any
   * element type.
   */
  static fromTypeDef(
    typeDef: TypeDef,
    values: unknown[],
  ): ListSchemaDefault | undefined {
    const elementType = FrameworkTypeFromTypeDef(typeDef.ItemType);

    switch (typeDef.ItemType.Type.toString()) {
      case "boolean":
      case "enum":
      case "float32":
      case "integer":
      case "int32":
      case "number":
      case "string":
        return new ListSchemaDefaultStaticValue(elementType, values);
      default:
        if (values.length === 0) {
          return new ListSchemaDefaultStaticValue(elementType, []);
        }
        return undefined;
    }
  }
}

/**
 * terraform-plugin-framework resource/schema/listdefault.StaticValue(). Sets a
 * static list default value using types.ListValueMust().
 */
class ListSchemaDefaultStaticValue extends ListSchemaDefault {
  private readonly elementType: FrameworkType;
  private readonly values: unknown[];

  constructor(elementType: FrameworkType, values: unknown[]) {
    super();
    this.elementType = elementType;
    this.values = values;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.values)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/attr" },
      ...this.elementType.schemaTypeImports().map((p) => ({ Path: p })),
    ];
  }

  template(): string {
    const elementTypeValues = this.values.map((value) => {
      return this.elementType.templateFrameworkValue(value);
    });

    const templateValues =
      this.values.length > 1
        ? "\n" + elementTypeValues.map((value) => `${value},\n`).join("")
        : elementTypeValues.join(", ");

    return `listdefault.StaticValue(types.ListValueMust(${this.elementType.templateInstantiation()}, []attr.Value{${templateValues}}))`;
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Object
 * implementations. Ensures type safety when used with SingleNestedAttribute
 * schema types.
 */
abstract class ObjectSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Object" = "Object";
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.Set
 * implementations. Ensures type safety when used with SetAttribute schema
 * types.
 */
abstract class SetSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "Set" = "Set";

  /**
   * Returns a SetSchemaDefault instance for the given TypeDef and values.
   * Supports primitive element types with values and empty arrays for any
   * element type.
   */
  static fromTypeDef(
    typeDef: TypeDef,
    values: unknown[],
  ): SetSchemaDefault | undefined {
    const elementType = FrameworkTypeFromTypeDef(typeDef.ItemType);

    switch (typeDef.ItemType.Type.toString()) {
      case "boolean":
      case "enum":
      case "float32":
      case "integer":
      case "int32":
      case "number":
      case "string":
        return new SetSchemaDefaultStaticValue(elementType, values);
      default:
        if (values.length === 0) {
          return new SetSchemaDefaultStaticValue(elementType, []);
        }
        return undefined;
    }
  }
}

/**
 * terraform-plugin-framework resource/schema/setdefault.StaticValue(). Sets a
 * static set default value using types.SetValueMust().
 */
class SetSchemaDefaultStaticValue extends SetSchemaDefault {
  private readonly elementType: FrameworkType;
  private readonly values: unknown[];

  constructor(elementType: FrameworkType, values: unknown[]) {
    super();
    this.elementType = elementType;
    this.values = values;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.values)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/attr" },
      ...this.elementType.schemaTypeImports().map((p) => ({ Path: p })),
    ];
  }

  template(): string {
    const elementTypeValues = this.values.map((value) => {
      return this.elementType.templateFrameworkValue(value);
    });

    const templateValues =
      this.values.length > 1
        ? "\n" + elementTypeValues.map((value) => `${value},\n`).join("")
        : elementTypeValues.join(", ");

    return `setdefault.StaticValue(types.SetValueMust(${this.elementType.templateInstantiation()}, []attr.Value{${templateValues}}))`;
  }
}

/**
 * terraform-plugin-framework resource/schema/objectdefault.StaticValue(). Sets
 * a static null object default value using types.ObjectNull() for
 * SingleNestedAttribute schema types.
 */
class SingleNestedSchemaDefaultNull extends ObjectSchemaDefault {
  private readonly typeDef: TypeDef;

  constructor(typeDef: TypeDef) {
    super();
    this.typeDef = typeDef;
  }

  imports(): GoImport[] {
    const result: GoImport[] = [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/attr" },
    ];

    // Collect imports from child field types.
    this.iterateChildren((frameworkType) => {
      frameworkType.schemaTypeImports().forEach((p) => {
        result.push({ Path: p });
      });
    });

    return result;
  }

  template(): string {
    const children: string[] = [];

    this.iterateChildren((frameworkType, name) => {
      children.push(`"${name}": ${frameworkType.templateInstantiation()},\n`);
    });

    children.sort();

    return (
      `objectdefault.StaticValue(types.ObjectNull(map[string]attr.Type{\n` +
      children.join("") +
      `}))`
    );
  }

  /**
   * Iterates over child fields of the TypeDef, calling the callback for each
   * with the FrameworkType and sanitized field name.
   */
  private iterateChildren(
    callback: (frameworkType: FrameworkType, name: string) => void,
  ): void {
    const iterator: IteratorFunction = (
      field: AttributeField,
      renderChildren: ({ indent, filter }: RenderChildrenOptions) => string,
    ) => {
      const frameworkType = FrameworkTypeFromTypeDef(field.type);
      callback(frameworkType, field.name);
      return { [field.name]: "" };
    };

    AttributeIterator(this.typeDef, iterator, 0, "");
  }
}

/**
 * Abstract typed default for terraform-plugin-framework defaults.String
 * implementations. Ensures type safety when used with StringAttribute schema
 * types.
 */
abstract class StringSchemaDefault extends SchemaDefault {
  readonly defaultTypeName: "String" = "String";
}

/**
 * terraform-plugin-framework resource/schema/stringdefault.StaticString(). Sets
 * a static string default value for a StringAttribute.
 */
class StringSchemaDefaultStatic extends StringSchemaDefault {
  private readonly value: string;

  constructor(value: string) {
    super();
    this.value = value;
  }

  description(): string {
    return `Default: ${JSON.stringify(this.value)}`;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
      },
    ];
  }

  template(): string {
    return `stringdefault.StaticString(${templateBuiltinString(this.value)})`;
  }
}
