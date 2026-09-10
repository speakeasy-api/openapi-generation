/**
 * Creates typed SchemaValidator instances for an enum TypeDef by dispatching
 * to the appropriate typed validator class's fromTypeDef() method based on
 * the enum's underlying type.
 *
 * Returns an empty array if the TypeDef has no enum or no values.
 */
function createEnumValidatorsFromTypeDef(typeDef: TypeDef): SchemaValidator[] {
  if (!typeDef.Enum?.Values?.length) return [];

  switch (typeDef.Enum.Type.Type.toString()) {
    case "string":
      return StringSchemaValidator.fromTypeDef(typeDef);
    case "int32":
      return Int32SchemaValidator.fromTypeDef(typeDef);
    case "integer":
      return Int64SchemaValidator.fromTypeDef(typeDef);
    case "float32":
      return Float32SchemaValidator.fromTypeDef(typeDef);
    case "number":
      return Float64SchemaValidator.fromTypeDef(typeDef);
    default:
      return [];
  }
}

/**
 * Creates typed SchemaValidator instances for path validators on an enum TypeDef
 * by dispatching to the appropriate typed validator class's
 * fromPathValidatorConfig() method based on the enum's underlying type.
 */
function createEnumValidatorsFromPathValidatorConfig(
  typeDef: TypeDef,
  config: PathValidatorConfig,
): SchemaValidator[] {
  switch (typeDef.Enum?.Type?.Type?.toString()) {
    case "string":
      return StringSchemaValidator.fromPathValidatorConfig(config);
    case "int32":
      return Int32SchemaValidator.fromPathValidatorConfig(config);
    case "integer":
      return Int64SchemaValidator.fromPathValidatorConfig(config);
    case "float32":
      return Float32SchemaValidator.fromPathValidatorConfig(config);
    case "number":
      return Float64SchemaValidator.fromPathValidatorConfig(config);
    default:
      return [];
  }
}

/**
 * Base class for all schema validators. Only abstract classes for schema types,
 * such as StringSchemaValidator, Int32SchemaValidator, etc., should extend this
 * class. Concrete validator implementations for specific validation constraints,
 * such as StringSchemaValidatorOneOf, should extend the abstract schema type
 * validator classes. This structure ensures type safety and organization of
 * validators by their applicable schema types.
 */
abstract class SchemaValidator {
  /**
   * Go framework validator type name (e.g., "Bool", "String"). Used to
   * generate the Go type reference `validator.{ValidatorTypeName}`.
   */
  abstract readonly validatorTypeName: string;

  /**
   * Provides an optional description for inclusion in attribute descriptions.
   * Concrete validators should override this method to add a description of
   * the validation constraint (e.g., "must be one of [...]") should it be
   * suitable for attribute descriptions.
   */
  description(): string {
    return "";
  }

  /**
   * Returns an array of Go imports required by the validator. Concrete
   * validators should call super.imports() and spread the result to include
   * base imports alongside their own.
   */
  imports(): GoImport[] {
    return [
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/schema/validator",
      },
    ];
  }

  /**
   * Returns the Go code for the validator. Concrete validators must override
   * this method.
   */
  template(): string {
    throw new Error("SchemaValidator template() not extended");
  }
}

/**
 * Builds a path expression string for a single path value. Splits the value on
 * "/" and maps each segment to either .AtParent() (for "..") or
 * .AtName("segment_name") (with TF state name sanitization).
 */
function pathMatchRelativeExpression(value: string): string {
  return (
    `path.MatchRelative().AtParent()` +
    value
      .split("/")
      .map((segment) =>
        segment == ".."
          ? `.AtParent()`
          : `.AtName("${sanitizeTFStateName(segment)}")`,
      )
      .join("")
  );
}

/**
 * Configuration fields relevant for path validators.
 */
type PathValidatorConfig = {
  ConflictsWith: string[];
  XorWith: string[];
  RequiredWith: string[];
};

/**
 * Creates typed path validator instances (AlsoRequires, ConflictsWith,
 * ExactlyOneOf) from config using the provided constructor functions.
 */
function createPathValidators<T extends SchemaValidator>(
  config: PathValidatorConfig,
  alsoRequires: new (values: string[]) => T,
  conflictsWith: new (values: string[]) => T,
  exactlyOneOf: new (values: string[]) => T,
): T[] {
  const validators: T[] = [];
  if (config.ConflictsWith.length > 0) {
    validators.push(new conflictsWith(config.ConflictsWith));
  }
  if (config.XorWith.length > 0) {
    validators.push(new exactlyOneOf(config.XorWith));
  }
  if (config.RequiredWith.length > 0) {
    validators.push(new alsoRequires(config.RequiredWith));
  }
  return validators;
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Bool
 * implementations. Ensures type safety when used with BoolAttribute schema
 * types.
 */
abstract class BoolSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Bool).
   */
  readonly validatorTypeName: "Bool" = "Bool";

  /**
   * Returns BoolSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): BoolSchemaValidator[] {
    return createPathValidators(
      config,
      BoolSchemaValidatorAlsoRequires,
      BoolSchemaValidatorConflictsWith,
      BoolSchemaValidatorExactlyOneOf,
    );
  }
}

class BoolSchemaValidatorAlsoRequires extends BoolSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `boolvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class BoolSchemaValidatorConflictsWith extends BoolSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `boolvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined bool validator. Generates boilerplate file if needed.
 */
class BoolSchemaValidatorCustom extends BoolSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("boolvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/boolvalidators`,
        Alias: "custom_boolvalidators",
      },
    ];
  }

  template(): string {
    return `custom_boolvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class BoolSchemaValidatorExactlyOneOf extends BoolSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `boolvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_boolvalidators.NotNull(). Validates that a bool
 * attribute value is not null.
 */
class BoolSchemaValidatorNotNull extends BoolSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/boolvalidators`,
        Alias: "speakeasy_boolvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_boolvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Float32
 * implementations. Ensures type safety when used with Float32Attribute schema
 * types.
 */
abstract class Float32SchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Float32).
   */
  readonly validatorTypeName: "Float32" = "Float32";

  /**
   * Returns Float32SchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for enum constraints and minimum/maximum validations,
   * returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): Float32SchemaValidator[] {
    if (!typeDef) return [];

    const validators: Float32SchemaValidator[] = [];

    if (typeDef.Enum?.Values?.length && !typeDef.Enum?.Open) {
      validators.push(new Float32SchemaValidatorOneOf(typeDef.Enum.Values));
    }

    const hasMinimum = typeDef.Validations?.Minimum;
    const hasMaximum = typeDef.Validations?.Maximum;

    if (hasMinimum && hasMaximum) {
      validators.push(
        new Float32SchemaValidatorBetween(
          typeDef.Validations!.Minimum!,
          typeDef.Validations!.Maximum!,
        ),
      );
    } else if (hasMinimum) {
      validators.push(
        new Float32SchemaValidatorAtLeast(typeDef.Validations!.Minimum!),
      );
    } else if (hasMaximum) {
      validators.push(
        new Float32SchemaValidatorAtMost(typeDef.Validations!.Maximum!),
      );
    }

    return validators;
  }

  /**
   * Returns Float32SchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): Float32SchemaValidator[] {
    return createPathValidators(
      config,
      Float32SchemaValidatorAlsoRequires,
      Float32SchemaValidatorConflictsWith,
      Float32SchemaValidatorExactlyOneOf,
    );
  }
}

class Float32SchemaValidatorAlsoRequires extends Float32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float32validator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * terraform-plugin-framework-validators float32validator.AtLeast(). Validates
 * that a float32 attribute value is at least the specified minimum.
 */
class Float32SchemaValidatorAtLeast extends Float32SchemaValidator {
  private minimum: number;

  constructor(minimum: number) {
    super();
    this.minimum = minimum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
    ];
  }

  template(): string {
    return `float32validator.AtLeast(${this.minimum})`;
  }
}

/**
 * terraform-plugin-framework-validators float32validator.AtMost(). Validates
 * that a float32 attribute value is at most the specified maximum.
 */
class Float32SchemaValidatorAtMost extends Float32SchemaValidator {
  private maximum: number;

  constructor(maximum: number) {
    super();
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
    ];
  }

  template(): string {
    return `float32validator.AtMost(${this.maximum})`;
  }
}

/**
 * terraform-plugin-framework-validators float32validator.Between(). Validates
 * that a float32 attribute value is between the specified minimum and maximum.
 */
class Float32SchemaValidatorBetween extends Float32SchemaValidator {
  private minimum: number;
  private maximum: number;

  constructor(minimum: number, maximum: number) {
    super();
    this.minimum = minimum;
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
    ];
  }

  template(): string {
    return `float32validator.Between(${this.minimum}, ${this.maximum})`;
  }
}

class Float32SchemaValidatorConflictsWith extends Float32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float32validator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined float32 validator. Generates boilerplate file if needed.
 */
class Float32SchemaValidatorCustom extends Float32SchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("float32validators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/float32validators`,
        Alias: "custom_float32validators",
      },
    ];
  }

  template(): string {
    return `custom_float32validators.${sanitizeClassName(
      this.validatorName,
    )}()`;
  }
}

class Float32SchemaValidatorExactlyOneOf extends Float32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float32validator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_float32validators.NotNull(). Validates that a float32
 * attribute value is not null.
 */
class Float32SchemaValidatorNotNull extends Float32SchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/float32validators`,
        Alias: "speakeasy_float32validators",
      },
    ];
  }

  template(): string {
    return `speakeasy_float32validators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators float32validator.OneOf(). Validates
 * that a float32 attribute value matches one of the specified enum values.
 */
class Float32SchemaValidatorOneOf extends Float32SchemaValidator {
  private values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float32validator",
      },
    ];
  }

  template(): string {
    const formatted = this.values;

    if (formatted.join(", ").length > 10) {
      return `float32validator.OneOf(\n${formatted
        .map((v) => `${v},`)
        .join("\n")}\n)`;
    }

    return `float32validator.OneOf(${formatted.join(", ")})`;
  }

  /**
   * Provides an optional description for inclusion in attribute
   * descriptions.
   */
  description(): string {
    const values = this.values.join(", ");

    if (this.values.length > 1) {
      return `must be one of [${values}]`;
    }

    return `must be ${values}`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Float64
 * implementations. Ensures type safety when used with Float64Attribute schema
 * types.
 */
abstract class Float64SchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Float64).
   */
  readonly validatorTypeName: "Float64" = "Float64";

  /**
   * Returns Float64SchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for enum constraints and minimum/maximum validations,
   * returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): Float64SchemaValidator[] {
    if (!typeDef) return [];

    const validators: Float64SchemaValidator[] = [];

    if (typeDef.Enum?.Values?.length && !typeDef.Enum?.Open) {
      validators.push(new Float64SchemaValidatorOneOf(typeDef.Enum.Values));
    }

    const hasMinimum = typeDef.Validations?.Minimum;
    const hasMaximum = typeDef.Validations?.Maximum;

    if (hasMinimum && hasMaximum) {
      validators.push(
        new Float64SchemaValidatorBetween(
          typeDef.Validations!.Minimum!,
          typeDef.Validations!.Maximum!,
        ),
      );
    } else if (hasMinimum) {
      validators.push(
        new Float64SchemaValidatorAtLeast(typeDef.Validations!.Minimum!),
      );
    } else if (hasMaximum) {
      validators.push(
        new Float64SchemaValidatorAtMost(typeDef.Validations!.Maximum!),
      );
    }

    return validators;
  }

  /**
   * Returns Float64SchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): Float64SchemaValidator[] {
    return createPathValidators(
      config,
      Float64SchemaValidatorAlsoRequires,
      Float64SchemaValidatorConflictsWith,
      Float64SchemaValidatorExactlyOneOf,
    );
  }
}

class Float64SchemaValidatorAlsoRequires extends Float64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float64validator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * terraform-plugin-framework-validators float64validator.AtLeast(). Validates
 * that a float64 attribute value is at least the specified minimum.
 */
class Float64SchemaValidatorAtLeast extends Float64SchemaValidator {
  private minimum: number;

  constructor(minimum: number) {
    super();
    this.minimum = minimum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
    ];
  }

  template(): string {
    return `float64validator.AtLeast(${this.minimum})`;
  }
}

/**
 * terraform-plugin-framework-validators float64validator.AtMost(). Validates
 * that a float64 attribute value is at most the specified maximum.
 */
class Float64SchemaValidatorAtMost extends Float64SchemaValidator {
  private maximum: number;

  constructor(maximum: number) {
    super();
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
    ];
  }

  template(): string {
    return `float64validator.AtMost(${this.maximum})`;
  }
}

/**
 * terraform-plugin-framework-validators float64validator.Between(). Validates
 * that a float64 attribute value is between the specified minimum and maximum.
 */
class Float64SchemaValidatorBetween extends Float64SchemaValidator {
  private minimum: number;
  private maximum: number;

  constructor(minimum: number, maximum: number) {
    super();
    this.minimum = minimum;
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
    ];
  }

  template(): string {
    return `float64validator.Between(${this.minimum}, ${this.maximum})`;
  }
}

class Float64SchemaValidatorConflictsWith extends Float64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float64validator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined float64 validator. Generates boilerplate file if needed.
 */
class Float64SchemaValidatorCustom extends Float64SchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("float64validators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/float64validators`,
        Alias: "custom_float64validators",
      },
    ];
  }

  template(): string {
    return `custom_float64validators.${sanitizeClassName(
      this.validatorName,
    )}()`;
  }
}

class Float64SchemaValidatorExactlyOneOf extends Float64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `float64validator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_float64validators.NotNull(). Validates that a float64
 * attribute value is not null.
 */
class Float64SchemaValidatorNotNull extends Float64SchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/float64validators`,
        Alias: "speakeasy_float64validators",
      },
    ];
  }

  template(): string {
    return `speakeasy_float64validators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators float64validator.OneOf(). Validates
 * that a float64 attribute value matches one of the specified enum values.
 */
class Float64SchemaValidatorOneOf extends Float64SchemaValidator {
  private values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/float64validator",
      },
    ];
  }

  template(): string {
    const formatted = this.values;

    if (formatted.join(", ").length > 10) {
      return `float64validator.OneOf(\n${formatted
        .map((v) => `${v},`)
        .join("\n")}\n)`;
    }

    return `float64validator.OneOf(${formatted.join(", ")})`;
  }

  /**
   * Provides an optional description for inclusion in attribute
   * descriptions.
   */
  description(): string {
    const values = this.values.join(", ");

    if (this.values.length > 1) {
      return `must be one of [${values}]`;
    }

    return `must be ${values}`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Int32
 * implementations. Ensures type safety when used with Int32Attribute schema
 * types.
 */
abstract class Int32SchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Int32).
   */
  readonly validatorTypeName: "Int32" = "Int32";

  /**
   * Returns Int32SchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for enum constraints and minimum/maximum validations,
   * returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): Int32SchemaValidator[] {
    if (!typeDef) return [];

    const validators: Int32SchemaValidator[] = [];

    if (typeDef.Enum?.Values?.length && !typeDef.Enum?.Open) {
      validators.push(new Int32SchemaValidatorOneOf(typeDef.Enum.Values));
    }

    const hasMinimum =
      typeDef.Validations?.Minimum != null &&
      Number.isInteger(+typeDef.Validations?.Minimum);
    const hasMaximum =
      typeDef.Validations?.Maximum != null &&
      Number.isInteger(+typeDef.Validations?.Maximum);

    if (hasMinimum && hasMaximum) {
      validators.push(
        new Int32SchemaValidatorBetween(
          typeDef.Validations!.Minimum!,
          typeDef.Validations!.Maximum!,
        ),
      );
    } else if (hasMinimum) {
      validators.push(
        new Int32SchemaValidatorAtLeast(typeDef.Validations!.Minimum!),
      );
    } else if (hasMaximum) {
      validators.push(
        new Int32SchemaValidatorAtMost(typeDef.Validations!.Maximum!),
      );
    }

    return validators;
  }

  /**
   * Returns Int32SchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): Int32SchemaValidator[] {
    return createPathValidators(
      config,
      Int32SchemaValidatorAlsoRequires,
      Int32SchemaValidatorConflictsWith,
      Int32SchemaValidatorExactlyOneOf,
    );
  }
}

class Int32SchemaValidatorAlsoRequires extends Int32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int32validator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * terraform-plugin-framework-validators int32validator.AtLeast(). Validates
 * that an int32 attribute value is at least the specified minimum.
 */
class Int32SchemaValidatorAtLeast extends Int32SchemaValidator {
  private minimum: number;

  constructor(minimum: number) {
    super();
    this.minimum = minimum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
    ];
  }

  template(): string {
    return `int32validator.AtLeast(${this.minimum})`;
  }
}

/**
 * terraform-plugin-framework-validators int32validator.AtMost(). Validates
 * that an int32 attribute value is at most the specified maximum.
 */
class Int32SchemaValidatorAtMost extends Int32SchemaValidator {
  private maximum: number;

  constructor(maximum: number) {
    super();
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
    ];
  }

  template(): string {
    return `int32validator.AtMost(${this.maximum})`;
  }
}

/**
 * terraform-plugin-framework-validators int32validator.Between(). Validates
 * that an int32 attribute value is between the specified minimum and maximum.
 */
class Int32SchemaValidatorBetween extends Int32SchemaValidator {
  private minimum: number;
  private maximum: number;

  constructor(minimum: number, maximum: number) {
    super();
    this.minimum = minimum;
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
    ];
  }

  template(): string {
    return `int32validator.Between(${this.minimum}, ${this.maximum})`;
  }
}

class Int32SchemaValidatorConflictsWith extends Int32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int32validator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined int32 validator. Generates boilerplate file if needed.
 */
class Int32SchemaValidatorCustom extends Int32SchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("int32validators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/int32validators`,
        Alias: "custom_int32validators",
      },
    ];
  }

  template(): string {
    return `custom_int32validators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class Int32SchemaValidatorExactlyOneOf extends Int32SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int32validator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_int32validators.NotNull(). Validates that an int32
 * attribute value is not null.
 */
class Int32SchemaValidatorNotNull extends Int32SchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/int32validators`,
        Alias: "speakeasy_int32validators",
      },
    ];
  }

  template(): string {
    return `speakeasy_int32validators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators int32validator.OneOf(). Validates that
 * an int32 attribute value matches one of the specified enum values.
 */
class Int32SchemaValidatorOneOf extends Int32SchemaValidator {
  private values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int32validator",
      },
    ];
  }

  template(): string {
    const formatted = this.values;

    if (formatted.join(", ").length > 10) {
      return `int32validator.OneOf(\n${formatted
        .map((v) => `${v},`)
        .join("\n")}\n)`;
    }

    return `int32validator.OneOf(${formatted.join(", ")})`;
  }

  /**
   * Provides an optional description for inclusion in attribute
   * descriptions.
   */
  description(): string {
    const values = this.values.join(", ");

    if (this.values.length > 1) {
      return `must be one of [${values}]`;
    }

    return `must be ${values}`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Int64
 * implementations. Ensures type safety when used with Int64Attribute schema
 * types.
 */
abstract class Int64SchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Int64).
   */
  readonly validatorTypeName: "Int64" = "Int64";

  /**
   * Returns Int64SchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for enum constraints and minimum/maximum validations,
   * returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): Int64SchemaValidator[] {
    if (!typeDef) return [];

    const validators: Int64SchemaValidator[] = [];

    if (typeDef.Enum?.Values?.length && !typeDef.Enum?.Open) {
      validators.push(new Int64SchemaValidatorOneOf(typeDef.Enum.Values));
    }

    const hasMinimum =
      typeDef.Validations?.Minimum != null &&
      Number.isInteger(+typeDef.Validations?.Minimum);
    const hasMaximum =
      typeDef.Validations?.Maximum != null &&
      Number.isInteger(+typeDef.Validations?.Maximum);

    if (hasMinimum && hasMaximum) {
      validators.push(
        new Int64SchemaValidatorBetween(
          typeDef.Validations!.Minimum!,
          typeDef.Validations!.Maximum!,
        ),
      );
    } else if (hasMinimum) {
      validators.push(
        new Int64SchemaValidatorAtLeast(typeDef.Validations!.Minimum!),
      );
    } else if (hasMaximum) {
      validators.push(
        new Int64SchemaValidatorAtMost(typeDef.Validations!.Maximum!),
      );
    }

    return validators;
  }

  /**
   * Returns Int64SchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): Int64SchemaValidator[] {
    return createPathValidators(
      config,
      Int64SchemaValidatorAlsoRequires,
      Int64SchemaValidatorConflictsWith,
      Int64SchemaValidatorExactlyOneOf,
    );
  }
}

class Int64SchemaValidatorAlsoRequires extends Int64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int64validator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * terraform-plugin-framework-validators int64validator.AtLeast(). Validates
 * that an int64 attribute value is at least the specified minimum.
 */
class Int64SchemaValidatorAtLeast extends Int64SchemaValidator {
  private minimum: number;

  constructor(minimum: number) {
    super();
    this.minimum = minimum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
    ];
  }

  template(): string {
    return `int64validator.AtLeast(${this.minimum})`;
  }
}

/**
 * terraform-plugin-framework-validators int64validator.AtMost(). Validates
 * that an int64 attribute value is at most the specified maximum.
 */
class Int64SchemaValidatorAtMost extends Int64SchemaValidator {
  private maximum: number;

  constructor(maximum: number) {
    super();
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
    ];
  }

  template(): string {
    return `int64validator.AtMost(${this.maximum})`;
  }
}

/**
 * terraform-plugin-framework-validators int64validator.Between(). Validates
 * that an int64 attribute value is between the specified minimum and maximum.
 */
class Int64SchemaValidatorBetween extends Int64SchemaValidator {
  private minimum: number;
  private maximum: number;

  constructor(minimum: number, maximum: number) {
    super();
    this.minimum = minimum;
    this.maximum = maximum;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
    ];
  }

  template(): string {
    return `int64validator.Between(${this.minimum}, ${this.maximum})`;
  }
}

class Int64SchemaValidatorConflictsWith extends Int64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int64validator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined int64 validator. Generates boilerplate file if needed.
 */
class Int64SchemaValidatorCustom extends Int64SchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("int64validators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/int64validators`,
        Alias: "custom_int64validators",
      },
    ];
  }

  template(): string {
    return `custom_int64validators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class Int64SchemaValidatorExactlyOneOf extends Int64SchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `int64validator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_int64validators.NotNull(). Validates that an int64
 * attribute value is not null.
 */
class Int64SchemaValidatorNotNull extends Int64SchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/int64validators`,
        Alias: "speakeasy_int64validators",
      },
    ];
  }

  template(): string {
    return `speakeasy_int64validators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators int64validator.OneOf(). Validates that
 * an int64 attribute value matches one of the specified enum values.
 */
class Int64SchemaValidatorOneOf extends Int64SchemaValidator {
  private values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/int64validator",
      },
    ];
  }

  template(): string {
    const formatted = this.values;

    if (formatted.join(", ").length > 10) {
      return `int64validator.OneOf(\n${formatted
        .map((v) => `${v},`)
        .join("\n")}\n)`;
    }

    return `int64validator.OneOf(${formatted.join(", ")})`;
  }

  /**
   * Provides an optional description for inclusion in attribute
   * descriptions.
   */
  description(): string {
    const values = this.values.join(", ");

    if (this.values.length > 1) {
      return `must be one of [${values}]`;
    }

    return `must be ${values}`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.List
 * implementations. Ensures type safety when used with ListAttribute schema
 * types.
 */
abstract class ListSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.List).
   */
  readonly validatorTypeName: "List" = "List";

  /**
   * Returns ListSchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for size constraints, unique items, and item type
   * validations, returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): ListSchemaValidator[] {
    if (!typeDef) return [];

    const validators: ListSchemaValidator[] = [];

    if (typeDef.Validations?.MinItems) {
      validators.push(
        new ListSchemaValidatorSizeAtLeast(
          Math.abs(typeDef.Validations.MinItems),
        ),
      );
    }

    if (typeDef.Validations?.MaxItems) {
      validators.push(
        new ListSchemaValidatorSizeAtMost(
          Math.abs(typeDef.Validations.MaxItems),
        ),
      );
    }

    if (typeDef.Validations?.UniqueItems) {
      validators.push(new ListSchemaValidatorUniqueValues());
    }

    if (typeDef.ItemType?.Type == "any") {
      validators.push(new ListSchemaValidatorValueStringsAreValidJSON());
    }

    return validators;
  }

  /**
   * Returns ListSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): ListSchemaValidator[] {
    return createPathValidators(
      config,
      ListSchemaValidatorAlsoRequires,
      ListSchemaValidatorConflictsWith,
      ListSchemaValidatorExactlyOneOf,
    );
  }
}

class ListSchemaValidatorAlsoRequires extends ListSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `listvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class ListSchemaValidatorConflictsWith extends ListSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `listvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined list validator. Generates boilerplate file if needed.
 */
class ListSchemaValidatorCustom extends ListSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("listvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/listvalidators`,
        Alias: "custom_listvalidators",
      },
    ];
  }

  template(): string {
    return `custom_listvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class ListSchemaValidatorExactlyOneOf extends ListSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `listvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_listvalidators.NotNull(). Validates that a list
 * attribute value is not null.
 */
class ListSchemaValidatorNotNull extends ListSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/listvalidators`,
        Alias: "speakeasy_listvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_listvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators listvalidator.SizeAtLeast(). Validates
 * that a list attribute has at least the specified number of elements.
 */
class ListSchemaValidatorSizeAtLeast extends ListSchemaValidator {
  private minItems: number;

  constructor(minItems: number) {
    super();
    this.minItems = minItems;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
    ];
  }

  template(): string {
    return `listvalidator.SizeAtLeast(${this.minItems})`;
  }
}

/**
 * terraform-plugin-framework-validators listvalidator.SizeAtMost(). Validates
 * that a list attribute has at most the specified number of elements.
 */
class ListSchemaValidatorSizeAtMost extends ListSchemaValidator {
  private maxItems: number;

  constructor(maxItems: number) {
    super();
    this.maxItems = maxItems;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
    ];
  }

  template(): string {
    return `listvalidator.SizeAtMost(${this.maxItems})`;
  }
}

/**
 * terraform-plugin-framework-validators listvalidator.UniqueValues(). Validates
 * that all elements in a list attribute are unique.
 */
class ListSchemaValidatorUniqueValues extends ListSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
    ];
  }

  template(): string {
    return `listvalidator.UniqueValues()`;
  }
}

/**
 * terraform-plugin-framework-validators listvalidator.ValueStringsAre() with
 * validators.IsValidJSON(). Validates that all string elements in a list
 * attribute contain valid JSON.
 */
class ListSchemaValidatorValueStringsAreValidJSON extends ListSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
      },
      { Path: `${getRootPackage()}/internal/validators` },
    ];
  }

  template(): string {
    return `listvalidator.ValueStringsAre(validators.IsValidJSON())`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Map
 * implementations. Ensures type safety when used with MapAttribute schema
 * types.
 */
abstract class MapSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Map).
   */
  readonly validatorTypeName: "Map" = "Map";

  /**
   * Returns MapSchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for item type validations, returning validators for
   * any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): MapSchemaValidator[] {
    if (!typeDef) return [];

    const validators: MapSchemaValidator[] = [];

    if (typeDef.ItemType?.Type == "any") {
      validators.push(new MapSchemaValidatorValueStringsAreValidJSON());
    }

    return validators;
  }

  /**
   * Returns MapSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): MapSchemaValidator[] {
    return createPathValidators(
      config,
      MapSchemaValidatorAlsoRequires,
      MapSchemaValidatorConflictsWith,
      MapSchemaValidatorExactlyOneOf,
    );
  }
}

class MapSchemaValidatorAlsoRequires extends MapSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `mapvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class MapSchemaValidatorConflictsWith extends MapSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `mapvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined map validator. Generates boilerplate file if needed.
 */
class MapSchemaValidatorCustom extends MapSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("mapvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/mapvalidators`,
        Alias: "custom_mapvalidators",
      },
    ];
  }

  template(): string {
    return `custom_mapvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class MapSchemaValidatorExactlyOneOf extends MapSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `mapvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_mapvalidators.NotNull(). Validates that a map
 * attribute value is not null.
 */
class MapSchemaValidatorNotNull extends MapSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/mapvalidators`,
        Alias: "speakeasy_mapvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_mapvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators mapvalidator.ValueStringsAre() with
 * validators.IsValidJSON(). Validates that all string values in a map
 * attribute contain valid JSON.
 */
class MapSchemaValidatorValueStringsAreValidJSON extends MapSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator",
      },
      { Path: `${getRootPackage()}/internal/validators` },
    ];
  }

  template(): string {
    return `mapvalidator.ValueStringsAre(validators.IsValidJSON())`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Object
 * implementations. Ensures type safety when used with ObjectAttribute and
 * SingleNestedAttribute schema types.
 */
abstract class ObjectSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Object).
   */
  readonly validatorTypeName: "Object" = "Object";

  /**
   * Returns ObjectSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): ObjectSchemaValidator[] {
    return createPathValidators(
      config,
      ObjectSchemaValidatorAlsoRequires,
      ObjectSchemaValidatorConflictsWith,
      ObjectSchemaValidatorExactlyOneOf,
    );
  }
}

class ObjectSchemaValidatorAlsoRequires extends ObjectSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `objectvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class ObjectSchemaValidatorConflictsWith extends ObjectSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `objectvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined object validator. Generates boilerplate file if needed.
 */
class ObjectSchemaValidatorCustom extends ObjectSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("objectvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/objectvalidators`,
        Alias: "custom_objectvalidators",
      },
    ];
  }

  template(): string {
    return `custom_objectvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class ObjectSchemaValidatorExactlyOneOf extends ObjectSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `objectvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_objectvalidators.NotNull(). Validates that an object
 * attribute value is not null.
 */
class ObjectSchemaValidatorNotNull extends ObjectSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/objectvalidators`,
        Alias: "speakeasy_objectvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_objectvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.Set
 * implementations. Ensures type safety when used with SetAttribute schema
 * types.
 */
abstract class SetSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.Set).
   */
  readonly validatorTypeName: "Set" = "Set";

  /**
   * Returns SetSchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for size constraints and item type validations,
   * returning validators for any that are found. uniqueItems needs no
   * validator: Terraform set elements are unique by construction.
   */
  static fromTypeDef(typeDef?: TypeDef): SetSchemaValidator[] {
    if (!typeDef) return [];

    const validators: SetSchemaValidator[] = [];

    if (typeDef.Validations?.MinItems) {
      validators.push(
        new SetSchemaValidatorSizeAtLeast(
          Math.abs(typeDef.Validations.MinItems),
        ),
      );
    }

    if (typeDef.Validations?.MaxItems) {
      validators.push(
        new SetSchemaValidatorSizeAtMost(
          Math.abs(typeDef.Validations.MaxItems),
        ),
      );
    }

    if (typeDef.ItemType?.Type == "any") {
      validators.push(new SetSchemaValidatorValueStringsAreValidJSON());
    }

    return validators;
  }

  /**
   * Returns SetSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): SetSchemaValidator[] {
    return createPathValidators(
      config,
      SetSchemaValidatorAlsoRequires,
      SetSchemaValidatorConflictsWith,
      SetSchemaValidatorExactlyOneOf,
    );
  }
}

class SetSchemaValidatorAlsoRequires extends SetSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `setvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class SetSchemaValidatorConflictsWith extends SetSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `setvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined set validator. Generates boilerplate file if needed.
 */
class SetSchemaValidatorCustom extends SetSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("setvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/setvalidators`,
        Alias: "custom_setvalidators",
      },
    ];
  }

  template(): string {
    return `custom_setvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class SetSchemaValidatorExactlyOneOf extends SetSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `setvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal speakeasy_setvalidators.NotNull(). Validates that a set
 * attribute value is not null.
 */
class SetSchemaValidatorNotNull extends SetSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/setvalidators`,
        Alias: "speakeasy_setvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_setvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators setvalidator.SizeAtLeast(). Validates
 * that a set attribute has at least the specified number of elements.
 */
class SetSchemaValidatorSizeAtLeast extends SetSchemaValidator {
  private minItems: number;

  constructor(minItems: number) {
    super();
    this.minItems = minItems;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
    ];
  }

  template(): string {
    return `setvalidator.SizeAtLeast(${this.minItems})`;
  }
}

/**
 * terraform-plugin-framework-validators setvalidator.SizeAtMost(). Validates
 * that a set attribute has at most the specified number of elements.
 */
class SetSchemaValidatorSizeAtMost extends SetSchemaValidator {
  private maxItems: number;

  constructor(maxItems: number) {
    super();
    this.maxItems = maxItems;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
    ];
  }

  template(): string {
    return `setvalidator.SizeAtMost(${this.maxItems})`;
  }
}

/**
 * terraform-plugin-framework-validators setvalidator.ValueStringsAre() with
 * validators.IsValidJSON(). Validates that all string elements in a set
 * attribute contain valid JSON.
 */
class SetSchemaValidatorValueStringsAreValidJSON extends SetSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator",
      },
      { Path: `${getRootPackage()}/internal/validators` },
    ];
  }

  template(): string {
    return `setvalidator.ValueStringsAre(validators.IsValidJSON())`;
  }
}

/**
 * Abstract typed validator for terraform-plugin-framework validator.String
 * implementations. Ensures type safety when used with StringAttribute schema
 * types.
 */
abstract class StringSchemaValidator extends SchemaValidator {
  /**
   * Go framework validator type name (validator.String).
   */
  readonly validatorTypeName: "String" = "String";

  /**
   * Returns StringSchemaValidator instances applicable to the given TypeDef.
   * Inspects the TypeDef for enum constraints, length constraints, and pattern
   * validations, returning validators for any that are found.
   */
  static fromTypeDef(typeDef?: TypeDef): StringSchemaValidator[] {
    if (!typeDef) return [];

    const validators: StringSchemaValidator[] = [];

    if (typeDef.Enum?.Values?.length && !typeDef.Enum?.Open) {
      validators.push(new StringSchemaValidatorOneOf(typeDef.Enum.Values));
    }

    switch (String(typeDef.Type)) {
      case "date":
        validators.push(new StringSchemaValidatorIsValidDate());
        break;
      case "date-time":
        validators.push(new StringSchemaValidatorIsRFC3339());
        break;
      case "string":
        const hasMinLength = typeDef.Validations?.MinLength != null;
        const hasMaxLength = typeDef.Validations?.MaxLength != null;

        if (hasMinLength && hasMaxLength) {
          validators.push(
            new StringSchemaValidatorUTF8LengthBetween(
              typeDef.Validations!.MinLength!,
              typeDef.Validations!.MaxLength!,
            ),
          );
        } else if (hasMinLength) {
          validators.push(
            new StringSchemaValidatorUTF8LengthAtLeast(
              typeDef.Validations!.MinLength!,
            ),
          );
        } else if (hasMaxLength) {
          validators.push(
            new StringSchemaValidatorUTF8LengthAtMost(
              typeDef.Validations!.MaxLength!,
            ),
          );
        }

        if (
          typeDef.Validations?.Pattern &&
          isRE2Regex(typeDef.Validations.Pattern)
        ) {
          validators.push(
            new StringSchemaValidatorRegexMatches(typeDef.Validations.Pattern),
          );
        }

        break;
    }

    return validators;
  }

  /**
   * Returns StringSchemaValidator instances for path validators derived from
   * config.
   */
  static fromPathValidatorConfig(
    config: PathValidatorConfig,
  ): StringSchemaValidator[] {
    return createPathValidators(
      config,
      StringSchemaValidatorAlsoRequires,
      StringSchemaValidatorConflictsWith,
      StringSchemaValidatorExactlyOneOf,
    );
  }
}

class StringSchemaValidatorAlsoRequires extends StringSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `stringvalidator.AlsoRequires(path.Expressions{\n${expressions},\n}...)`;
  }
}

class StringSchemaValidatorConflictsWith extends StringSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `stringvalidator.ConflictsWith(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Custom user-defined string validator. Generates boilerplate file if needed.
 */
class StringSchemaValidatorCustom extends StringSchemaValidator {
  private validatorName: string;

  constructor(validatorName: string) {
    super();
    this.validatorName = validatorName;
    boilerplateCustomValidation("stringvalidators", validatorName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/stringvalidators`,
        Alias: "custom_stringvalidators",
      },
    ];
  }

  template(): string {
    return `custom_stringvalidators.${sanitizeClassName(this.validatorName)}()`;
  }
}

class StringSchemaValidatorExactlyOneOf extends StringSchemaValidator {
  readonly values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const expressions = this.values
      .map((v) => pathMatchRelativeExpression(v))
      .join(",\n");
    return `stringvalidator.ExactlyOneOf(path.Expressions{\n${expressions},\n}...)`;
  }
}

/**
 * Internal validators.IsRFC3339(). Validates that a string attribute value is a
 * valid RFC 3339 date-time.
 */
class StringSchemaValidatorIsRFC3339 extends StringSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      { Path: `${getRootPackage()}/internal/validators` },
    ];
  }

  template(): string {
    return `validators.IsRFC3339()`;
  }
}

/**
 * Internal validators.IsValidDate(). Validates that a string attribute value is
 * a valid date.
 */
class StringSchemaValidatorIsValidDate extends StringSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      { Path: `${getRootPackage()}/internal/validators` },
    ];
  }

  template(): string {
    return `validators.IsValidDate()`;
  }
}

/**
 * Internal speakeasy_stringvalidators.NotNull(). Validates that a string
 * attribute value is not null.
 */
class StringSchemaValidatorNotNull extends StringSchemaValidator {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/validators/stringvalidators`,
        Alias: "speakeasy_stringvalidators",
      },
    ];
  }

  template(): string {
    return `speakeasy_stringvalidators.NotNull()`;
  }

  description(): string {
    return "Not Null";
  }
}

/**
 * terraform-plugin-framework-validators stringvalidator.OneOf(). Validates that
 * a string attribute value matches one of the specified enum values.
 */
class StringSchemaValidatorOneOf extends StringSchemaValidator {
  private values: string[];

  constructor(values: string[]) {
    super();
    this.values = values;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
    ];
  }

  template(): string {
    const formatted = this.values.map((v) => `"${v}"`);

    if (formatted.join(", ").length > 10) {
      return `stringvalidator.OneOf(\n${formatted
        .map((v) => `${v},`)
        .join("\n")}\n)`;
    }

    return `stringvalidator.OneOf(${formatted.join(", ")})`;
  }

  /**
   * Provides an optional description for inclusion in attribute
   * descriptions.
   */
  description(): string {
    const values = this.values.map((v) => JSON.stringify(v)).join(", ");

    if (this.values.length > 1) {
      return `must be one of [${values}]`;
    }

    return `must be ${values}`;
  }
}

/**
 * terraform-plugin-framework-validators stringvalidator.RegexMatches(). Validates
 * that a string attribute value matches the specified regular expression pattern.
 */
class StringSchemaValidatorRegexMatches extends StringSchemaValidator {
  private pattern: string;

  constructor(pattern: string) {
    super();
    this.pattern = pattern;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
      { Path: "regexp" },
    ];
  }

  template(): string {
    const escapedPattern = this.pattern.replaceAll("`", '` + "`" + `');
    return `stringvalidator.RegexMatches(regexp.MustCompile(\`${escapedPattern}\`), "must match pattern " + regexp.MustCompile(\`${escapedPattern}\`).String())`;
  }
}

/**
 * terraform-plugin-framework-validators stringvalidator.UTF8LengthAtLeast().
 * Validates that a string attribute value has at least the specified UTF-8
 * length.
 */
class StringSchemaValidatorUTF8LengthAtLeast extends StringSchemaValidator {
  private minLength: number;

  constructor(minLength: number) {
    super();
    this.minLength = minLength;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
    ];
  }

  template(): string {
    return `stringvalidator.UTF8LengthAtLeast(${this.minLength})`;
  }
}

/**
 * terraform-plugin-framework-validators stringvalidator.UTF8LengthAtMost().
 * Validates that a string attribute value has at most the specified UTF-8
 * length.
 */
class StringSchemaValidatorUTF8LengthAtMost extends StringSchemaValidator {
  private maxLength: number;

  constructor(maxLength: number) {
    super();
    this.maxLength = maxLength;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
    ];
  }

  template(): string {
    return `stringvalidator.UTF8LengthAtMost(${this.maxLength})`;
  }
}

/**
 * terraform-plugin-framework-validators stringvalidator.UTF8LengthBetween().
 * Validates that a string attribute value has a UTF-8 length between the
 * specified minimum and maximum values.
 */
class StringSchemaValidatorUTF8LengthBetween extends StringSchemaValidator {
  private minLength: number;
  private maxLength: number;

  constructor(minLength: number, maxLength: number) {
    super();
    this.minLength = minLength;
    this.maxLength = maxLength;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
      },
    ];
  }

  template(): string {
    return `stringvalidator.UTF8LengthBetween(${this.minLength}, ${this.maxLength})`;
  }
}

/**
 * Creates a NotNull SchemaValidator for the given TypeDef type.
 */
function createNotNullValidator(typeDef: TypeDef): SchemaValidator | null {
  switch (String(typeDef.Type)) {
    case "boolean":
      return new BoolSchemaValidatorNotNull();
    case "float32":
      return new Float32SchemaValidatorNotNull();
    case "int32":
      return new Int32SchemaValidatorNotNull();
    case "integer":
      return new Int64SchemaValidatorNotNull();
    case "number":
      return new Float64SchemaValidatorNotNull();
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new StringSchemaValidatorNotNull();
    case "array":
      return new ListSchemaValidatorNotNull();
    case "set":
      return new SetSchemaValidatorNotNull();
    case "map":
      return new MapSchemaValidatorNotNull();
    case "class":
    case "union":
      return new ObjectSchemaValidatorNotNull();
    case "enum":
      return createNotNullValidator(typeDef.Enum.Type);
    default:
      return null;
  }
}

/**
 * Creates a custom SchemaValidator for the given TypeDef type and validator
 * name. Generates boilerplate file if needed.
 */
function createCustomValidator(
  typeDef: TypeDef,
  validatorName: string,
): SchemaValidator | null {
  switch (String(typeDef.Type)) {
    case "boolean":
      return new BoolSchemaValidatorCustom(validatorName);
    case "float32":
      return new Float32SchemaValidatorCustom(validatorName);
    case "int32":
      return new Int32SchemaValidatorCustom(validatorName);
    case "integer":
      return new Int64SchemaValidatorCustom(validatorName);
    case "number":
      return new Float64SchemaValidatorCustom(validatorName);
    case "any":
    case "bytes":
    case "date":
    case "date-time":
    case "string":
      return new StringSchemaValidatorCustom(validatorName);
    case "array":
      return new ListSchemaValidatorCustom(validatorName);
    case "set":
      return new SetSchemaValidatorCustom(validatorName);
    case "map":
      return new MapSchemaValidatorCustom(validatorName);
    case "class":
    case "union":
      return new ObjectSchemaValidatorCustom(validatorName);
    case "enum":
      return createCustomValidator(typeDef.Enum.Type, validatorName);
    default:
      return null;
  }
}
