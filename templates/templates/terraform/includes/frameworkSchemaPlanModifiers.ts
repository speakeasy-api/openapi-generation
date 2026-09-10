/**
 * Base class for all schema plan modifiers. Only abstract classes for schema
 * types, such as StringSchemaPlanModifier, Int32SchemaPlanModifier, etc.,
 * should extend this class. Concrete plan modifier implementations for specific
 * modification behaviors, such as StringSchemaPlanModifierRequiresReplaceIfConfigured,
 * should extend the abstract schema type plan modifier classes. This structure
 * ensures type safety and organization of plan modifiers by their applicable
 * schema types.
 */
abstract class SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (e.g., "Bool", "String"). Used to
   * generate the Go type reference `planmodifier.{PlanModifierTypeName}`.
   */
  abstract readonly planModifierTypeName: string;

  /**
   * Provides an optional description for inclusion in attribute descriptions.
   * Concrete plan modifiers should override this method to add a description of
   * the plan modification behavior (e.g., "Requires replacement if changed.")
   * should it be suitable for attribute descriptions.
   */
  description(): string {
    return "";
  }

  /**
   * Returns an array of Go imports required by the plan modifier. Concrete
   * plan modifiers should call super.imports() and spread the result to include
   * base imports alongside their own.
   */
  imports(): GoImport[] {
    return [
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier",
      },
    ];
  }

  /**
   * Returns the Go code for the plan modifier. Concrete plan modifiers must
   * override this method.
   */
  template(): string {
    throw new Error("SchemaPlanModifier template() not extended");
  }
}

/**
 * Configuration fields relevant for plan modifier creation.
 */
type PlanModifierConfig = {
  CustomPlanModifiers: string[];
  ForceNew: boolean;
  SuppressComputedDiff: "" | "Standard" | "ExplicitSuppress";
  PlanOnly: boolean;
  HoistedFrom: TerraformHoistedSource[] | null;
  Computed: boolean;
};

/**
 * Builds a path expression string for UseHoistedValue plan modifiers.
 * e.g., ["parent", "child"] => path.Root("parent").AtName("child")
 */
function buildHoistedValuePathExpression(
  pathPrefix: string[],
  ...additionalSegments: string[]
): string {
  const allSegments = [...pathPrefix, ...additionalSegments];
  if (allSegments.length === 0) {
    return 'path.Root("")';
  }
  let result = `path.Root("${allSegments[0]}")`;
  for (let i = 1; i < allSegments.length; i++) {
    result += `.AtName("${allSegments[i]}")`;
  }
  return result;
}

/**
 * Generates the Go template string for UseHoistedValue source arguments.
 */
function templateHoistedSources(sources: TerraformHoistedSource[]): string {
  return sources
    .map((source) => {
      const pathPrefix = source.PathPrefix || [];
      const associatedTypePath = buildHoistedValuePathExpression(
        pathPrefix,
        source.AssociatedTypeName,
      );
      const fieldPath = buildHoistedValuePathExpression(
        pathPrefix,
        source.AssociatedTypeName,
        source.FieldName,
      );
      return (
        `speakeasy_planmodifierutils.HoistedSource{` +
        `AssociatedTypePath: ${associatedTypePath}, ` +
        `FieldPath: ${fieldPath}}`
      );
    })
    .join(", ");
}

/**
 * Creates typed SchemaPlanModifier instances from config using the provided
 * constructor functions. Handles the interaction between UseHoistedValue and
 * SuppressDiff (UseHoistedValue subsumes drift suppression).
 */
function createPlanModifiersFromConfig<T extends SchemaPlanModifier>(
  config: PlanModifierConfig,
  customClass: new (name: string) => T,
  requiresReplaceClass: new () => T,
  suppressDiffClass: new (mode: string) => T,
  useConfigValueClass: new () => T,
  useHoistedValueClass: new (sources: TerraformHoistedSource[]) => T,
): T[] {
  const modifiers: T[] = [];

  // Add custom plan modifiers
  if (config.CustomPlanModifiers && config.CustomPlanModifiers.length > 0) {
    config.CustomPlanModifiers.forEach((name) => {
      modifiers.push(new customClass(name));
    });
  }

  // Add ForceNew plan modifier
  if (config.ForceNew) {
    modifiers.push(new requiresReplaceClass());
  }

  const hasHoistedValue =
    config.HoistedFrom && config.HoistedFrom.length > 0 && config.Computed;

  // Add SuppressComputedDiff plan modifier (skipped when UseHoistedValue
  // applies, since UseHoistedValue handles drift suppression)
  if (config.SuppressComputedDiff && !hasHoistedValue) {
    modifiers.push(new suppressDiffClass(config.SuppressComputedDiff));
  }

  // Add UseConfigValue plan modifier for plan-only attributes
  if (config.PlanOnly) {
    modifiers.push(new useConfigValueClass());
  }

  // Add UseHoistedValue plan modifier for hoisted oneOf fields
  if (hasHoistedValue) {
    modifiers.push(new useHoistedValueClass(config.HoistedFrom!));
  }

  return modifiers;
}

/**
 * Creates typed SchemaPlanModifier instances for an enum TypeDef by dispatching
 * to the appropriate typed plan modifier class's fromConfig() method based on
 * the enum's underlying type.
 */
function createEnumPlanModifiersFromConfig(
  typeDef: TypeDef,
  config: PlanModifierConfig,
): SchemaPlanModifier[] {
  switch (typeDef.Enum?.Type?.Type?.toString()) {
    case "string":
      return StringSchemaPlanModifier.fromConfig(config);
    case "int32":
      return Int32SchemaPlanModifier.fromConfig(config);
    case "integer":
      return Int64SchemaPlanModifier.fromConfig(config);
    case "float32":
      return Float32SchemaPlanModifier.fromConfig(config);
    case "number":
      return Float64SchemaPlanModifier.fromConfig(config);
    default:
      return [];
  }
}

// ---------------------------------------------------------------------------
// Bool
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework planmodifier.Bool
 * implementations. Ensures type safety when used with BoolAttribute schema
 * types.
 */
abstract class BoolSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Bool).
   */
  readonly planModifierTypeName: "Bool" = "Bool";

  /**
   * Returns BoolSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): BoolSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      BoolSchemaPlanModifierCustom,
      BoolSchemaPlanModifierRequiresReplaceIfConfigured,
      BoolSchemaPlanModifierSuppressDiff,
      BoolSchemaPlanModifierUseConfigValue,
      BoolSchemaPlanModifierUseHoistedValue,
    );
  }
}

/**
 * Custom user-defined bool plan modifier. Generates boilerplate file if needed.
 */
class BoolSchemaPlanModifierCustom extends BoolSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("boolplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/boolplanmodifier`,
        Alias: "custom_boolplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_boolplanmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

/**
 * terraform-plugin-framework boolplanmodifier.RequiresReplaceIfConfigured().
 * Forces resource replacement when the attribute value changes.
 */
class BoolSchemaPlanModifierRequiresReplaceIfConfigured extends BoolSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier",
      },
    ];
  }

  template(): string {
    return "boolplanmodifier.RequiresReplaceIfConfigured()";
  }
}

/**
 * Internal speakeasy_boolplanmodifier.SuppressDiff(). Suppresses
 * "(known after changes)" messages by copying state to plan when plan is
 * unknown.
 */
class BoolSchemaPlanModifierSuppressDiff extends BoolSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/boolplanmodifier`,
        Alias: "speakeasy_boolplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_boolplanmodifier.SuppressDiff(speakeasy_boolplanmodifier.${this.mode})`;
  }
}

/**
 * Internal speakeasy_boolplanmodifier.UseConfigValue(). Ensures configuration
 * values (not prior state) are used in API requests for plan-only attributes.
 */
class BoolSchemaPlanModifierUseConfigValue extends BoolSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/boolplanmodifier`,
        Alias: "speakeasy_boolplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_boolplanmodifier.UseConfigValue()";
  }
}

/**
 * Internal speakeasy_boolplanmodifier.UseHoistedValue(). Copies value from
 * active oneOf variant to hoisted field to prevent false drift detection.
 */
class BoolSchemaPlanModifierUseHoistedValue extends BoolSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/boolplanmodifier`,
        Alias: "speakeasy_boolplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_boolplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Float32
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.Float32 implementations. Ensures type safety when used with
 * Float32Attribute schema types.
 */
abstract class Float32SchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Float32).
   */
  readonly planModifierTypeName: "Float32" = "Float32";

  /**
   * Returns Float32SchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): Float32SchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      Float32SchemaPlanModifierCustom,
      Float32SchemaPlanModifierRequiresReplaceIfConfigured,
      Float32SchemaPlanModifierSuppressDiff,
      Float32SchemaPlanModifierUseConfigValue,
      Float32SchemaPlanModifierUseHoistedValue,
    );
  }
}

class Float32SchemaPlanModifierCustom extends Float32SchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("float32planmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float32planmodifier`,
        Alias: "custom_float32planmodifier",
      },
    ];
  }

  template(): string {
    return `custom_float32planmodifier.${sanitizeClassName(
      this.modifierName,
    )}()`;
  }
}

class Float32SchemaPlanModifierRequiresReplaceIfConfigured extends Float32SchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier",
      },
    ];
  }

  template(): string {
    return "float32planmodifier.RequiresReplaceIfConfigured()";
  }
}

class Float32SchemaPlanModifierSuppressDiff extends Float32SchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float32planmodifier`,
        Alias: "speakeasy_float32planmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_float32planmodifier.SuppressDiff(speakeasy_float32planmodifier.${this.mode})`;
  }
}

class Float32SchemaPlanModifierUseConfigValue extends Float32SchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float32planmodifier`,
        Alias: "speakeasy_float32planmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_float32planmodifier.UseConfigValue()";
  }
}

class Float32SchemaPlanModifierUseHoistedValue extends Float32SchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float32planmodifier`,
        Alias: "speakeasy_float32planmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_float32planmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Float64
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.Float64 implementations. Ensures type safety when used with
 * Float64Attribute schema types.
 */
abstract class Float64SchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Float64).
   */
  readonly planModifierTypeName: "Float64" = "Float64";

  /**
   * Returns Float64SchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): Float64SchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      Float64SchemaPlanModifierCustom,
      Float64SchemaPlanModifierRequiresReplaceIfConfigured,
      Float64SchemaPlanModifierSuppressDiff,
      Float64SchemaPlanModifierUseConfigValue,
      Float64SchemaPlanModifierUseHoistedValue,
    );
  }
}

class Float64SchemaPlanModifierCustom extends Float64SchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("float64planmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float64planmodifier`,
        Alias: "custom_float64planmodifier",
      },
    ];
  }

  template(): string {
    return `custom_float64planmodifier.${sanitizeClassName(
      this.modifierName,
    )}()`;
  }
}

class Float64SchemaPlanModifierRequiresReplaceIfConfigured extends Float64SchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier",
      },
    ];
  }

  template(): string {
    return "float64planmodifier.RequiresReplaceIfConfigured()";
  }
}

class Float64SchemaPlanModifierSuppressDiff extends Float64SchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float64planmodifier`,
        Alias: "speakeasy_float64planmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_float64planmodifier.SuppressDiff(speakeasy_float64planmodifier.${this.mode})`;
  }
}

class Float64SchemaPlanModifierUseConfigValue extends Float64SchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float64planmodifier`,
        Alias: "speakeasy_float64planmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_float64planmodifier.UseConfigValue()";
  }
}

class Float64SchemaPlanModifierUseHoistedValue extends Float64SchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/float64planmodifier`,
        Alias: "speakeasy_float64planmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_float64planmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Int32
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.Int32 implementations. Ensures type safety when used with
 * Int32Attribute schema types.
 */
abstract class Int32SchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Int32).
   */
  readonly planModifierTypeName: "Int32" = "Int32";

  /**
   * Returns Int32SchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): Int32SchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      Int32SchemaPlanModifierCustom,
      Int32SchemaPlanModifierRequiresReplaceIfConfigured,
      Int32SchemaPlanModifierSuppressDiff,
      Int32SchemaPlanModifierUseConfigValue,
      Int32SchemaPlanModifierUseHoistedValue,
    );
  }
}

class Int32SchemaPlanModifierCustom extends Int32SchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("int32planmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int32planmodifier`,
        Alias: "custom_int32planmodifier",
      },
    ];
  }

  template(): string {
    return `custom_int32planmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

class Int32SchemaPlanModifierRequiresReplaceIfConfigured extends Int32SchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier",
      },
    ];
  }

  template(): string {
    return "int32planmodifier.RequiresReplaceIfConfigured()";
  }
}

class Int32SchemaPlanModifierSuppressDiff extends Int32SchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int32planmodifier`,
        Alias: "speakeasy_int32planmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_int32planmodifier.SuppressDiff(speakeasy_int32planmodifier.${this.mode})`;
  }
}

class Int32SchemaPlanModifierUseConfigValue extends Int32SchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int32planmodifier`,
        Alias: "speakeasy_int32planmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_int32planmodifier.UseConfigValue()";
  }
}

class Int32SchemaPlanModifierUseHoistedValue extends Int32SchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int32planmodifier`,
        Alias: "speakeasy_int32planmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_int32planmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Int64
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.Int64 implementations. Ensures type safety when used with
 * Int64Attribute schema types.
 */
abstract class Int64SchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Int64).
   */
  readonly planModifierTypeName: "Int64" = "Int64";

  /**
   * Returns Int64SchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): Int64SchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      Int64SchemaPlanModifierCustom,
      Int64SchemaPlanModifierRequiresReplaceIfConfigured,
      Int64SchemaPlanModifierSuppressDiff,
      Int64SchemaPlanModifierUseConfigValue,
      Int64SchemaPlanModifierUseHoistedValue,
    );
  }
}

class Int64SchemaPlanModifierCustom extends Int64SchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("int64planmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int64planmodifier`,
        Alias: "custom_int64planmodifier",
      },
    ];
  }

  template(): string {
    return `custom_int64planmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

class Int64SchemaPlanModifierRequiresReplaceIfConfigured extends Int64SchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier",
      },
    ];
  }

  template(): string {
    return "int64planmodifier.RequiresReplaceIfConfigured()";
  }
}

class Int64SchemaPlanModifierSuppressDiff extends Int64SchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int64planmodifier`,
        Alias: "speakeasy_int64planmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_int64planmodifier.SuppressDiff(speakeasy_int64planmodifier.${this.mode})`;
  }
}

class Int64SchemaPlanModifierUseConfigValue extends Int64SchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int64planmodifier`,
        Alias: "speakeasy_int64planmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_int64planmodifier.UseConfigValue()";
  }
}

class Int64SchemaPlanModifierUseHoistedValue extends Int64SchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/int64planmodifier`,
        Alias: "speakeasy_int64planmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_int64planmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework planmodifier.List
 * implementations. Ensures type safety when used with ListAttribute and
 * ListNestedAttribute schema types.
 */
abstract class ListSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.List).
   */
  readonly planModifierTypeName: "List" = "List";

  /**
   * Returns ListSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): ListSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      ListSchemaPlanModifierCustom,
      ListSchemaPlanModifierRequiresReplaceIfConfigured,
      ListSchemaPlanModifierSuppressDiff,
      ListSchemaPlanModifierUseConfigValue,
      ListSchemaPlanModifierUseHoistedValue,
    );
  }
}

class ListSchemaPlanModifierCustom extends ListSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("listplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/listplanmodifier`,
        Alias: "custom_listplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_listplanmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

class ListSchemaPlanModifierRequiresReplaceIfConfigured extends ListSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier",
      },
    ];
  }

  template(): string {
    return "listplanmodifier.RequiresReplaceIfConfigured()";
  }
}

class ListSchemaPlanModifierSuppressDiff extends ListSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/listplanmodifier`,
        Alias: "speakeasy_listplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_listplanmodifier.SuppressDiff(speakeasy_listplanmodifier.${this.mode})`;
  }
}

class ListSchemaPlanModifierUseConfigValue extends ListSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/listplanmodifier`,
        Alias: "speakeasy_listplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_listplanmodifier.UseConfigValue()";
  }
}

class ListSchemaPlanModifierUseHoistedValue extends ListSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/listplanmodifier`,
        Alias: "speakeasy_listplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_listplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Map
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework planmodifier.Map
 * implementations. Ensures type safety when used with MapAttribute and
 * MapNestedAttribute schema types.
 */
abstract class MapSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Map).
   */
  readonly planModifierTypeName: "Map" = "Map";

  /**
   * Returns MapSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): MapSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      MapSchemaPlanModifierCustom,
      MapSchemaPlanModifierRequiresReplaceIfConfigured,
      MapSchemaPlanModifierSuppressDiff,
      MapSchemaPlanModifierUseConfigValue,
      MapSchemaPlanModifierUseHoistedValue,
    );
  }
}

class MapSchemaPlanModifierCustom extends MapSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("mapplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/mapplanmodifier`,
        Alias: "custom_mapplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_mapplanmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

class MapSchemaPlanModifierRequiresReplaceIfConfigured extends MapSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier",
      },
    ];
  }

  template(): string {
    return "mapplanmodifier.RequiresReplaceIfConfigured()";
  }
}

class MapSchemaPlanModifierSuppressDiff extends MapSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/mapplanmodifier`,
        Alias: "speakeasy_mapplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_mapplanmodifier.SuppressDiff(speakeasy_mapplanmodifier.${this.mode})`;
  }
}

class MapSchemaPlanModifierUseConfigValue extends MapSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/mapplanmodifier`,
        Alias: "speakeasy_mapplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_mapplanmodifier.UseConfigValue()";
  }
}

class MapSchemaPlanModifierUseHoistedValue extends MapSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/mapplanmodifier`,
        Alias: "speakeasy_mapplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_mapplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Object
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.Object implementations. Ensures type safety when used with
 * SingleNestedAttribute schema types.
 */
abstract class ObjectSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Object).
   */
  readonly planModifierTypeName: "Object" = "Object";

  /**
   * Returns ObjectSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): ObjectSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      ObjectSchemaPlanModifierCustom,
      ObjectSchemaPlanModifierRequiresReplaceIfConfigured,
      ObjectSchemaPlanModifierSuppressDiff,
      ObjectSchemaPlanModifierUseConfigValue,
      ObjectSchemaPlanModifierUseHoistedValue,
    );
  }
}

class ObjectSchemaPlanModifierCustom extends ObjectSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("objectplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/objectplanmodifier`,
        Alias: "custom_objectplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_objectplanmodifier.${sanitizeClassName(
      this.modifierName,
    )}()`;
  }
}

class ObjectSchemaPlanModifierRequiresReplaceIfConfigured extends ObjectSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier",
      },
    ];
  }

  template(): string {
    return "objectplanmodifier.RequiresReplaceIfConfigured()";
  }
}

class ObjectSchemaPlanModifierSuppressDiff extends ObjectSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/objectplanmodifier`,
        Alias: "speakeasy_objectplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_objectplanmodifier.SuppressDiff(speakeasy_objectplanmodifier.${this.mode})`;
  }
}

class ObjectSchemaPlanModifierUseConfigValue extends ObjectSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/objectplanmodifier`,
        Alias: "speakeasy_objectplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_objectplanmodifier.UseConfigValue()";
  }
}

class ObjectSchemaPlanModifierUseHoistedValue extends ObjectSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/objectplanmodifier`,
        Alias: "speakeasy_objectplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_objectplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// Set
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework planmodifier.Set
 * implementations. Ensures type safety when used with SetAttribute and
 * SetNestedAttribute schema types.
 */
abstract class SetSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.Set).
   */
  readonly planModifierTypeName: "Set" = "Set";

  /**
   * Returns SetSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): SetSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      SetSchemaPlanModifierCustom,
      SetSchemaPlanModifierRequiresReplaceIfConfigured,
      SetSchemaPlanModifierSuppressDiff,
      SetSchemaPlanModifierUseConfigValue,
      SetSchemaPlanModifierUseHoistedValue,
    );
  }
}

class SetSchemaPlanModifierCustom extends SetSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("setplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/setplanmodifier`,
        Alias: "custom_setplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_setplanmodifier.${sanitizeClassName(this.modifierName)}()`;
  }
}

class SetSchemaPlanModifierRequiresReplaceIfConfigured extends SetSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier",
      },
    ];
  }

  template(): string {
    return "setplanmodifier.RequiresReplaceIfConfigured()";
  }
}

class SetSchemaPlanModifierSuppressDiff extends SetSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/setplanmodifier`,
        Alias: "speakeasy_setplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_setplanmodifier.SuppressDiff(speakeasy_setplanmodifier.${this.mode})`;
  }
}

class SetSchemaPlanModifierUseConfigValue extends SetSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/setplanmodifier`,
        Alias: "speakeasy_setplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_setplanmodifier.UseConfigValue()";
  }
}

class SetSchemaPlanModifierUseHoistedValue extends SetSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/setplanmodifier`,
        Alias: "speakeasy_setplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_setplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

/**
 * Abstract typed plan modifier for terraform-plugin-framework
 * planmodifier.String implementations. Ensures type safety when used with
 * StringAttribute schema types.
 */
abstract class StringSchemaPlanModifier extends SchemaPlanModifier {
  /**
   * Go framework plan modifier type name (planmodifier.String).
   */
  readonly planModifierTypeName: "String" = "String";

  /**
   * Returns StringSchemaPlanModifier instances derived from config.
   */
  static fromConfig(config: PlanModifierConfig): StringSchemaPlanModifier[] {
    return createPlanModifiersFromConfig(
      config,
      StringSchemaPlanModifierCustom,
      StringSchemaPlanModifierRequiresReplaceIfConfigured,
      StringSchemaPlanModifierSuppressDiff,
      StringSchemaPlanModifierUseConfigValue,
      StringSchemaPlanModifierUseHoistedValue,
    );
  }
}

class StringSchemaPlanModifierCustom extends StringSchemaPlanModifier {
  private modifierName: string;

  constructor(modifierName: string) {
    super();
    this.modifierName = modifierName;
    boilerplateCustomPlanModification("stringplanmodifier", modifierName);
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/stringplanmodifier`,
        Alias: "custom_stringplanmodifier",
      },
    ];
  }

  template(): string {
    return `custom_stringplanmodifier.${sanitizeClassName(
      this.modifierName,
    )}()`;
  }
}

class StringSchemaPlanModifierRequiresReplaceIfConfigured extends StringSchemaPlanModifier {
  description(): string {
    return "Requires replacement if changed.";
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier",
      },
    ];
  }

  template(): string {
    return "stringplanmodifier.RequiresReplaceIfConfigured()";
  }
}

class StringSchemaPlanModifierSuppressDiff extends StringSchemaPlanModifier {
  private mode: string;

  constructor(mode: string) {
    super();
    this.mode = mode;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/stringplanmodifier`,
        Alias: "speakeasy_stringplanmodifier",
      },
    ];
  }

  template(): string {
    return `speakeasy_stringplanmodifier.SuppressDiff(speakeasy_stringplanmodifier.${this.mode})`;
  }
}

class StringSchemaPlanModifierUseConfigValue extends StringSchemaPlanModifier {
  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/stringplanmodifier`,
        Alias: "speakeasy_stringplanmodifier",
      },
    ];
  }

  template(): string {
    return "speakeasy_stringplanmodifier.UseConfigValue()";
  }
}

class StringSchemaPlanModifierUseHoistedValue extends StringSchemaPlanModifier {
  private sources: TerraformHoistedSource[];

  constructor(sources: TerraformHoistedSource[]) {
    super();
    this.sources = sources;
  }

  imports(): GoImport[] {
    return [
      ...super.imports(),
      {
        Path: `${getRootPackage()}/internal/planmodifiers/stringplanmodifier`,
        Alias: "speakeasy_stringplanmodifier",
      },
      {
        Path: `${getRootPackage()}/internal/planmodifiers/utils`,
        Alias: "speakeasy_planmodifierutils",
      },
      { Path: "github.com/hashicorp/terraform-plugin-framework/path" },
    ];
  }

  template(): string {
    const sources = templateHoistedSources(this.sources);
    return `speakeasy_stringplanmodifier.UseHoistedValue([]speakeasy_planmodifierutils.HoistedSource{${sources}})`;
  }
}
