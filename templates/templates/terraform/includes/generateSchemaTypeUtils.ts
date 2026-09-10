function formatDescription(
  Description: string,
  indent: number,
  extraComment: string = "",
) {
  let lines = Description.trim().split("\n");
  if (extraComment) {
    if (lines.length == 1) {
      if (lines[0].length > 0 && lines[0][lines[0].length - 1] != ".") {
        lines = [(lines[0] + ". " + extraComment).trim()];
      } else {
        lines = [(lines[0] + " " + extraComment).trim()];
      }
    } else {
      lines = lines.concat(extraComment);
    }
  }
  let builder = lines.length > 1 ? "MarkdownDescription: " : "Description: ";
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const lastLine = i == lines.length - 1;
    const firstLine = i == 0;
    if (!firstLine) {
      builder += templateIndent(indent);
    }
    builder +=
      "`" +
      line.replace(/(`+)/g, '` + "$1" +`').replace(/(({{)|(}}))/g, '{{"$1"}}');
    if (!lastLine) {
      builder += "` + " + '"\\n" +\n';
    } else {
      builder += "`";
    }
  }
  return builder;
}

type TerraformAttributeType =
  | "schema.StringAttribute"
  | "schema.Float32Attribute"
  | "schema.Float64Attribute"
  | "schema.Int32Attribute"
  | "schema.Int64Attribute"
  | "schema.BoolAttribute"
  | "schema.MapAttribute"
  | "schema.MapNestedAttribute"
  | "schema.ListAttribute"
  | "schema.ListNestedAttribute"
  | "schema.SetAttribute"
  | "schema.SetNestedAttribute"
  | "schema.SingleNestedAttribute";

function boilerplateCustomPlanModification(
  boilerplateFolder: string,
  planModifier: string,
) {
  const fileName = sanitizeTFStateName(planModifier);
  const filePath = `internal/planmodifiers/${boilerplateFolder}/${fileName}.go`;
  addUntrackedPattern(escapeRegExp(filePath));

  let existingData = readFile(filePath);

  // If we don't already have a file, create it
  if (!existingData) {
    templateFile(
      `boilerplate/${boilerplateFolder}/planmodifier.go.stmpl`,
      filePath,
      {
        PlanModifierName: sanitizeClassName(planModifier),
      },
    );
  }
}

function boilerplateCustomValidation(
  boilerplateFolder: string,
  validator: string,
) {
  const fileName = sanitizeTFStateName(validator);
  const filePath = `internal/validators/${boilerplateFolder}/${fileName}.go`;
  addUntrackedPattern(escapeRegExp(filePath));

  let existingData = readFile(filePath);

  // If we don't already have a file, create it
  if (!existingData) {
    templateFile(
      `boilerplate/${boilerplateFolder}/validator.go.stmpl`,
      filePath,
      {
        ValidatorName: sanitizeClassName(validator),
      },
    );
  }
}

type TypeDefConfig = {
  Alias: boolean;
  NotNull: boolean;
  Computed: boolean;
  ForceNew: boolean;
  Sensitive: boolean;
  Optional: boolean;
  Required: boolean;
  SuppressComputedDiff: "" | "Standard" | "ExplicitSuppress";
  ConflictsWith: string[];
  XorWith: string[];
  RequiredWith: string[];
  CustomPlanModifiers: string[];
  CustomType?: CustomTypeConfig;
  CustomValidators: string[];

  /** Enabled if the x-speakeasy-soft-delete-property annotation is present. */
  SoftDeleteProperty: boolean;

  /**
   * Terraform SDK WriteOnly flag. Only enabled if the
   * x-speakeasy-terraform-write-only extension is enabled. Only valid for
   * certain Terraform resource types, e.g. NOT data resources.
   */
  WriteOnly: boolean;

  /**
   * Plan-only flag. When true, adds the UseConfigValue plan modifier to ensure
   * configuration values (not prior state) are used in API requests.
   * Set via x-speakeasy-terraform-plan-only extension.
   */
  PlanOnly: boolean;

  /**
   * HoistedFrom tracks source associated types for hoisted oneOf fields.
   * When set, generates UseHoistedValue plan modifier to prevent false drift.
   * Uses PascalCase to match Go struct field names.
   */
  HoistedFrom: TerraformHoistedSource[] | null;
};

function configFromTypeDef(
  typedef: TypeDef,
  defaultValue: any,
  optional: boolean,
  resourceType: TerraformResourceType,
  isGlobalField: boolean,
): TypeDefConfig {
  const config: TypeDefConfig = {
    Alias: false,
    Computed: false,
    NotNull: false,
    ForceNew: false,
    Sensitive: false,
    Optional: false,
    Required: false,
    SuppressComputedDiff: "",
    ConflictsWith: [],
    XorWith: [],
    RequiredWith: [],
    CustomPlanModifiers: [],
    CustomValidators: [],
    SoftDeleteProperty: false,
    WriteOnly: false,
    PlanOnly: false,
    HoistedFrom: null,
  };

  if (typedef.Extensions?.MatchConfig) {
    config.Alias = true;
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-computed"] !== "undefined"
  ) {
    config.Computed = typedef.Extensions.All["x-speakeasy-param-computed"];
  } else {
    config.Computed = false;
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-computed-override"] !==
    "undefined"
  ) {
    config.Computed =
      typedef.Extensions.All["x-speakeasy-param-computed-override"];
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-force-new"] !==
    "undefined"
  ) {
    config.ForceNew = typedef.Extensions.All["x-speakeasy-param-force-new"];
  } else {
    config.ForceNew = false; // typedef.IsParameter
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-sensitive"] !==
    "undefined"
  ) {
    config.Sensitive = typedef.Extensions.All["x-speakeasy-param-sensitive"];
  } else {
    config.Sensitive = false;
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-optional"] !== "undefined"
  ) {
    config.Optional = typedef.Extensions.All["x-speakeasy-param-optional"];
    config.Required = !config.Optional;
  } else {
    config.Optional = optional;
    config.Required = !config.Optional;
  }

  if (
    (config.Computed &&
      typedef.Extensions?.All["x-speakeasy-param-suppress-computed-diff"]) ||
    (config.Computed &&
      config.ForceNew &&
      typedef.Extensions?.All["x-speakeasy-param-suppress-computed-diff"] ===
        undefined)
  ) {
    config.SuppressComputedDiff = "ExplicitSuppress";
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-param-readonly"] !==
      "undefined" &&
    typedef.Extensions.All["x-speakeasy-param-readonly"]
  ) {
    config.Computed = true;
    config.ForceNew = false;
    config.Required = false;
    config.Optional = false;
  }

  if (
    config.Required &&
    typedef.Extensions?.All["x-speakeasy-parent-require-to-not-null"]
  ) {
    config.Required = false;
    config.Optional = true;
    config.NotNull = true;
  }

  // Track hoisting sources for UseHoistedValue plan modifier generation
  // Must be set before createSchemaDefault check below
  if (
    typedef.Extensions?.TerraformHoistedFrom &&
    typedef.Extensions.TerraformHoistedFrom.length > 0
  ) {
    config.HoistedFrom = typedef.Extensions.TerraformHoistedFrom;
  }

  if (
    resourceType === "managed" &&
    createSchemaDefault(typedef, defaultValue, isGlobalField, config)
  ) {
    // Implicitly make it computed when there's a default
    config.Computed = true;
    // if it's required, it's now optional
    if (config.Required) {
      config.Optional = true;
      config.Required = false;
    }
  }

  if (
    typeof typedef.Extensions?.All["x-speakeasy-terraform-plan-only"] !==
      "undefined" &&
    typedef.Extensions.All["x-speakeasy-terraform-plan-only"] &&
    (config.Optional || config.Required)
  ) {
    // plan-only, where mutable can never be computed
    config.Computed = false;
    // Enable UseConfigValue plan modifier to ensure config values are used
    config.PlanOnly = true;
  }

  if (
    (resourceType === "managed" || resourceType === "action") &&
    typedef.Extensions?.TerraformWriteOnly
  ) {
    // Computed MUST be disabled.
    config.Computed = false;
    config.WriteOnly = typedef.Extensions.TerraformWriteOnly;
  }

  // cannot set both Computed and Required.
  if (config.Required) {
    config.Computed = false;
  }

  config.CustomType = TypeDefTerraformCustomTypeConfig(typedef);

  if (
    typeof typedef.Extensions?.All["x-speakeasy-plan-modifiers"] !== "undefined"
  ) {
    if (Array.isArray(typedef.Extensions.All["x-speakeasy-plan-modifiers"])) {
      config.CustomPlanModifiers =
        typedef.Extensions.All["x-speakeasy-plan-modifiers"];
    } else {
      config.CustomPlanModifiers = [
        typedef.Extensions.All["x-speakeasy-plan-modifiers"],
      ];
    }
    config.CustomPlanModifiers = config.CustomPlanModifiers.map((i) =>
      i.toString(),
    );
  }

  function handleValidator(
    extensionName: string,
    configValue: "ConflictsWith" | "XorWith" | "RequiredWith",
  ) {
    const val = typedef.Extensions?.All[extensionName];
    if (typeof val === "undefined" || (config.Computed && config.Required)) {
      return;
    }
    if (Array.isArray(val)) {
      config[configValue] = val;
    } else if (val != null && typeof val === "object") {
      // GoJa Go slices don't pass Array.isArray() but are iterable
      config[configValue] = Array.from(val);
    } else {
      config[configValue] = [val];
    }
    config[configValue] = config[configValue].map((i) => i.toString());
  }

  handleValidator("x-speakeasy-conflicts-with", "ConflictsWith");
  handleValidator("x-speakeasy-xor-with", "XorWith");
  handleValidator("x-speakeasy-required-with", "RequiredWith");

  if (config.Required || config.Optional) {
    if (
      typeof typedef.Extensions?.All["x-speakeasy-plan-validators"] !==
      "undefined"
    ) {
      if (
        Array.isArray(typedef.Extensions.All["x-speakeasy-plan-validators"])
      ) {
        config.CustomValidators =
          typedef.Extensions.All["x-speakeasy-plan-validators"];
      } else {
        config.CustomValidators = [
          typedef.Extensions.All["x-speakeasy-plan-validators"],
        ];
      }
      config.CustomValidators = config.CustomValidators.map((i) =>
        i.toString(),
      );
    }
  }

  if (
    resourceType === "managed" &&
    typeof typedef.Extensions?.All["x-speakeasy-soft-delete-property"] !==
      "undefined"
  ) {
    config.SoftDeleteProperty =
      typedef.Extensions.All["x-speakeasy-soft-delete-property"];
  }

  return config;
}

registerTemplateFunc("formatDescription", formatDescription);
