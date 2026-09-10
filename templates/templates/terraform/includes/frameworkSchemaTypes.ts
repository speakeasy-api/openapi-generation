type AttributeBaseTypeName =
  | "Bool"
  | "Dynamic"
  | "Float32"
  | "Float64"
  | "Int32"
  | "Int64"
  | "List"
  | "ListNested"
  | "Map"
  | "MapNested"
  | "Number"
  | "Object"
  | "Set"
  | "SetNested"
  | "SingleNested"
  | "String"
  | "Tuple";

/** Base representation of terraform-plugin-framework fwschema.Attribute. */
abstract class AttributeBase {
  protected typeName: AttributeBaseTypeName;

  constructor(typeName: AttributeBaseTypeName) {
    this.typeName = typeName;
  }

  attributes?: Record<string, AttributeBase>;
  attributeTypes?: FrameworkTypeAttributeTypes;
  elementType?: FrameworkType;
  elementTypes?: FrameworkTypeElementTypes;
  nestedObject?: NestedAttributeObjectBase;

  computed?: boolean;
  customType?: CustomTypeConfig;
  default?: SchemaDefault;
  deprecationMessage?: string;
  description?: string;
  markdownDescription?: string;
  optional?: boolean;
  planModifiers?: SchemaPlanModifier[];
  required?: boolean;
  sensitive?: boolean;
  validators?: SchemaValidator[];
  writeOnly?: boolean; // NOTE: Terraform SDK WriteOnly, not OAS writeOnly.

  imports(): GoImport[] {
    const result: GoImport[] = [];

    if (this.attributeTypes) {
      result.push(
        { Path: "github.com/hashicorp/terraform-plugin-framework/attr" },
        ...Object.values(this.attributeTypes)
          .map((at) => at.schemaTypeImports())
          .flat()
          .map((p) => ({ Path: p })),
      );
    }

    if (this.customType?.imports?.length) {
      this.customType.imports.forEach((importPath) => {
        if (!result.some((r) => r.Path === importPath)) {
          result.push({ Path: importPath });
        }
      });
    }

    if (this.default) {
      result.push(...this.default.imports());
    }

    if (this.elementType) {
      result.push(
        ...this.elementType.schemaTypeImports().map((p) => ({ Path: p })),
      );
    }

    if (this.elementTypes) {
      result.push(
        ...this.elementTypes
          .map((et) => et.schemaTypeImports())
          .flat()
          .map((p) => ({ Path: p })),
      );
    }

    if (this.planModifiers?.length > 0) {
      result.push(
        {
          Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier",
        },
        ...this.planModifiers.map((pm) => pm.imports()).flat(),
      );
    }

    if (this.validators?.length > 0 && (this.optional || this.required)) {
      result.push(
        {
          Path: "github.com/hashicorp/terraform-plugin-framework/schema/validator",
        },
        ...this.validators?.map((v) => v.imports()).flat(),
      );
    }

    return result;
  }

  template(): string {
    const result: string[] = [`schema.${this.typeName}Attribute{`];

    if (this.attributes) {
      result.push(
        `Attributes: map[string]schema.Attribute{`,
        ...Object.keys(this.attributes)
          .sort()
          .map(
            (attrName) =>
              `"${sanitizeTFStateName(attrName)}": ${this.attributes[
                attrName
              ].template()},`,
          ),
        `},`,
      );
    }

    if (this.attributeTypes) {
      result.push(
        `AttributeTypes: map[string]attr.Type{`,
        ...Object.keys(this.attributeTypes)
          .sort()
          .map(
            (attrName) =>
              `"${sanitizeTFStateName(attrName)}": ${this.attributeTypes[
                attrName
              ].templateInstantiation()},`,
          ),
        `},`,
      );
    }

    if (this.computed) {
      result.push(`Computed: true,`);
    }

    if (this.customType?.schemaType) {
      result.push(`CustomType: ${this.customType.schemaType},`);
    }

    if (this.default) {
      result.push(`Default: ${this.default.template()},`);
    }

    if (this.deprecationMessage) {
      result.push(
        `DeprecationMessage: ${templateBuiltinString(
          this.deprecationMessage,
        )},`,
      );
    }

    if (this.description) {
      result.push(`Description: ${templateBuiltinString(this.description)},`);
    }

    if (this.elementType) {
      result.push(`ElementType: ${this.elementType.templateInstantiation()},`);
    }

    if (this.elementTypes) {
      result.push(
        `ElementTypes: []schema.Type{`,
        ...this.elementTypes.map((et) => `${et.templateInstantiation()},`),
        `},`,
      );
    }

    if (this.markdownDescription) {
      result.push(
        `MarkdownDescription: ${templateBuiltinString(
          this.markdownDescription,
        )},`,
      );
    }

    if (this.nestedObject) {
      result.push(`NestedObject: ${this.nestedObject.template()},`);
    }

    if (this.optional) {
      result.push(`Optional: true,`);
    }

    if (this.planModifiers?.length > 0) {
      result.push(
        `PlanModifiers: []planmodifier.${this.typeName}{`,
        ...this.planModifiers.map((pm) => `${pm.template()},`),
        `},`,
      );
    }

    if (this.required) {
      result.push(`Required: true,`);
    }

    if (this.sensitive) {
      result.push(`Sensitive: true,`);
    }

    if (this.validators?.length > 0 && (this.optional || this.required)) {
      result.push(
        `Validators: []validator.${this.typeName}{`,
        ...this.validators.map((v) => `${v.template()},`),
        `},`,
      );
    }

    if (this.writeOnly) {
      result.push(`WriteOnly: true,`);
    }

    result.push(`}`);

    return result.join("\n");
  }
}

abstract class NestedAttributeObjectBase {
  constructor() {}

  attributes: Record<string, AttributeBase>;
  customType?: CustomTypeConfig;
  planModifiers?: SchemaPlanModifier[];
  validators?: SchemaValidator[];

  imports(): GoImport[] {
    const result: GoImport[] = Object.values(this.attributes)
      .map((attr) => attr.imports())
      .flat();

    if (this.customType?.imports?.length) {
      this.customType.imports.forEach((importPath) => {
        if (!result.some((r) => r.Path === importPath)) {
          result.push({ Path: importPath });
        }
      });
    }

    if (this.planModifiers?.length > 0) {
      result.push(
        {
          Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier",
        },
        ...this.planModifiers.map((pm) => pm.imports()).flat(),
      );
    }

    if (this.validators?.length > 0) {
      result.push(
        {
          Path: "github.com/hashicorp/terraform-plugin-framework/schema/validator",
        },
        ...this.validators.map((v) => v.imports()).flat(),
      );
    }

    return result;
  }

  template(): string {
    const result: string[] = [`schema.NestedAttributeObject{`];

    result.push(
      `Attributes: map[string]schema.Attribute{`,
      ...Object.keys(this.attributes)
        .sort()
        .map(
          (attrName) =>
            `"${sanitizeTFStateName(attrName)}": ${this.attributes[
              attrName
            ].template()},`,
        ),
      `},`,
    );

    if (this.customType?.schemaType) {
      result.push(`CustomType: ${this.customType.schemaType},`);
    }

    if (this.planModifiers?.length > 0) {
      result.push(
        `PlanModifiers: []planmodifier.Object{`,
        ...this.planModifiers.map((pm) => `${pm.template()},`),
        `},`,
      );
    }

    if (this.validators?.length > 0) {
      result.push(
        `Validators: []validator.Object{`,
        ...this.validators.map((v) => `${v.template()},`),
        `},`,
      );
    }

    result.push(`}`);

    return result.join("\n");
  }
}

/** Base representation of terraform-plugin-framework fwschema.Schema. */
abstract class SchemaBase {
  constructor() {}

  attributes?: Record<string, AttributeBase>;
  // Block not supported nor needed currently.
  // blocks?: Record<string, BlockBase>;
  deprecationMessage?: string;
  description?: string;
  markdownDescription?: string;
  version?: number;

  imports(): GoImport[] {
    let result: GoImport[] = [];

    if (Object.keys(this.attributes).length > 0) {
      result.push(
        ...Object.values(this.attributes)
          .map((a) => a.imports())
          .flat(),
      );
    }

    return result;
  }

  template(): string {
    let result: string[] = [`schema.Schema{`];

    if (this.attributes) {
      result.push(
        `Attributes: map[string]schema.Attribute{`,
        ...Object.keys(this.attributes)
          .sort()
          .map(
            (attrName) =>
              `"${sanitizeTFStateName(attrName)}": ${this.attributes[
                attrName
              ].template()},`,
          ),
        `},`,
      );
    }

    if (this.deprecationMessage) {
      result.push(
        `DeprecationMessage: ${templateBuiltinString(
          this.deprecationMessage,
        )},`,
      );
    }

    if (this.description) {
      result.push(`Description: ${templateBuiltinString(this.description)},`);
    }

    if (this.markdownDescription) {
      result.push(
        `MarkdownDescription: ${templateBuiltinString(
          this.markdownDescription,
        )},`,
      );
    }

    if (this.version) {
      result.push(`Version: ${this.version},`);
    }

    result.push(`}`);

    return result.join("\n");
  }
}
