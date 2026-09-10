/** terraform-plugin-framework custom types, defined via
 *  x-speakeasy-terraform-custom-type extension.
 *  Reference: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom */
type CustomTypeConfig = {
  /** Go imports for the schemaType and valueType, if necessary. */
  imports?: string[];

  /** Schema type for custom type. Must extend framework Typable interface. */
  schemaType: string;

  /** Value type for custom type. Must extend framework Valuable interface. */
  valueType: string;
};

/** Returns the CustomType if defined by x-speakeasy-terraform-custom-type
 *  extension, otherwise undefined. */
function TypeDefTerraformCustomTypeConfig(
  typeDef: TypeDef,
): CustomTypeConfig | undefined {
  if (
    typeof typeDef.Extensions.All["x-speakeasy-terraform-custom-type"] !==
    "undefined"
  ) {
    return {
      ...typeDef.Extensions.All["x-speakeasy-terraform-custom-type"],
    };
  }

  if (typeDef.Type.toString() === "any") {
    return {
      imports: [
        "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
      ],
      schemaType: "jsontypes.NormalizedType{}",
      valueType: "jsontypes.Normalized",
    };
  }

  if (
    typeDef.Type.toString() === "string" &&
    typeDef.ContentMediaType === "application/json"
  ) {
    return {
      imports: [
        "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
      ],
      schemaType: "jsontypes.NormalizedType{}",
      valueType: "jsontypes.Normalized",
    };
  }

  if (typeDef.Type.toString() === "bytes") {
    return {
      imports: [
        "github.com/speakeasy-api/terraform-plugin-framework-base64types/base64types",
      ],
      schemaType: "base64types.StandardType{}",
      valueType: "base64types.Standard",
    };
  }

  return undefined;
}

/** Collection of all terraform-plugin-framework types/basetypes. */
type FrameworkType =
  | FrameworkBase64StandardType
  | FrameworkBoolType
  | FrameworkDynamicType
  | FrameworkFloat32Type
  | FrameworkFloat64Type
  | FrameworkInt32Type
  | FrameworkInt64Type
  | FrameworkJsonNormalizedType
  | FrameworkListType
  | FrameworkMapType
  | FrameworkNumberType
  | FrameworkObjectType
  | FrameworkSetType
  | FrameworkStringType
  | FrameworkTupleType;

/** Create a FrameworkType (e.g. types.StringType) from a FieldDef. */
function FrameworkTypeFromFieldDef(fieldDef: FieldDef): FrameworkType {
  return FrameworkTypeFromTypeDef(fieldDef.Type);
}

/** Create a FrameworkType (e.g. types.StringType) from a TypeDef. */
function FrameworkTypeFromTypeDef(
  typeDef: TypeDef,
  parentCustomTypeConfig?: CustomTypeConfig,
): FrameworkType {
  const customTypeConfig = parentCustomTypeConfig
    ? parentCustomTypeConfig
    : TypeDefTerraformCustomTypeConfig(typeDef);

  switch (typeDef.Type.toString()) {
    case "any":
    case "date":
    case "date-time":
      return new FrameworkStringType(customTypeConfig);
    case "string":
      if (
        !parentCustomTypeConfig &&
        typeDef.ContentMediaType === "application/json"
      ) {
        return new FrameworkJsonNormalizedType();
      }
      return new FrameworkStringType(customTypeConfig);
    case "bytes":
      return new FrameworkBase64StandardType();
    case "boolean":
      return new FrameworkBoolType(customTypeConfig);
    case "enum":
      return FrameworkTypeFromTypeDef(typeDef.Enum.Type, customTypeConfig);
    case "float32":
      return new FrameworkFloat32Type(customTypeConfig);
    case "integer":
      return new FrameworkInt64Type(customTypeConfig);
    case "int32":
      return new FrameworkInt32Type(customTypeConfig);
    case "number":
      return new FrameworkFloat64Type(customTypeConfig);
    case "map":
      return new FrameworkMapType(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        customTypeConfig,
      );
    case "array":
      return new FrameworkListType(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        customTypeConfig,
      );
    case "set":
      return new FrameworkSetType(
        FrameworkTypeFromTypeDef(typeDef.ItemType),
        customTypeConfig,
      );
    case "class":
      const classAttrTypes: FrameworkTypeAttributeTypes = {};

      for (const field of typeDef.Fields) {
        const attrType = FrameworkTypeFromFieldDef(field);
        classAttrTypes[sanitizeTFStateName(field.Name)] = attrType;
      }

      return new FrameworkObjectType(classAttrTypes, customTypeConfig);
    case "union":
      const unionAttrTypes: FrameworkTypeAttributeTypes = {};

      for (const associatedType of typeDef.AssociatedTypes) {
        const attrType = FrameworkTypeFromTypeDef(associatedType);
        unionAttrTypes[
          sanitizeTFStateName(sanitizeTFUnionTypeName(typeDef, associatedType))
        ] = attrType;
      }

      return new FrameworkObjectType(unionAttrTypes, customTypeConfig);
    default:
      throw new Error(`Unknown type: ${typeDef.Type}`);
  }
}

/** Mapping of (object) attribute names to types. */
type FrameworkTypeAttributeTypes = Record<string, FrameworkType>;

/** Collection of FrameworkType. */
type FrameworkTypeElementTypes = FrameworkType[];

type FrameworkTypeBaseTypeName =
  | "Bool"
  | "Dynamic"
  | "Float32"
  | "Float64"
  | "Int32"
  | "Int64"
  | "List"
  | "Map"
  | "Number"
  | "Object"
  | "Set"
  | "String"
  | "Tuple";

/** Base representation of a terraform-plugin-framework types/basetype. */
abstract class FrameworkTypeBase {
  protected customTypeConfig: CustomTypeConfig | undefined;
  protected typeName: FrameworkTypeBaseTypeName;

  constructor(
    typeName: FrameworkTypeBaseTypeName,
    customTypeConfig: CustomTypeConfig | undefined,
  ) {
    this.customTypeConfig = customTypeConfig;
    this.typeName = typeName;
  }

  /**
   * JSON-encoded string example value. If present, a TypeDef example is
   * returned, otherwise an example value matching the type is returned.
   */
  templateExampleJSONValue(typeDef: TypeDef, exampleName?: string): string {
    switch (typeDef.Examples.length) {
      case 0:
        if (
          typeDef.Type.toString() === "enum" &&
          typeDef.Enum.Values.length > 0
        ) {
          return typeDef.Enum.Type.Type.toString() === "string"
            ? JSON.stringify(typeDef.Enum.Values[0])
            : typeDef.Enum.Values[0];
        }
        // TODO: Replace the templateExampleZeroValue method with a faker
        // variant similar to templateTFPrimitiveValue. This is okay for now
        // since this is only replacing the previous import example logic which
        // was not using faker.
        return this.templateExampleZeroValue();
      case 1:
        return typeDef.Examples[0].ToJSON();
      default:
        if (exampleName) {
          const example = getExampleByName(typeDef.Examples, exampleName);

          if (example) {
            return example.ToJSON();
          }
        }

        // If no specific example was found, return the first one.
        return typeDef.Examples[0].ToJSON();
    }
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    throw new Error(
      `templateExampleZeroValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /**
   * HCL-formatted example value for Terraform configuration files.
   * Used for provider example configuration (provider.tf).
   *
   * Priority:
   * 1. OAS-defined example value
   * 2. Enum value (if enum type)
   * 3. Zero value fallback
   *
   * @param typeDef - The TypeDef containing type info and examples
   * @param sensitive - If true, uses placeholder like "<YOUR_API_KEY>" when no example
   * @param sensitivePlaceholder - Custom placeholder text for sensitive fields
   */
  templateTerraformConfigValue(
    typeDef: TypeDef,
    sensitive?: boolean,
    sensitivePlaceholder?: string,
  ): string {
    // Check for OAS examples first
    if (typeDef.Examples.length > 0) {
      return typeDef.Examples[0].ToJSON();
    }

    // Handle enum types
    if (typeDef.Type.toString() === "enum" && typeDef.Enum.Values.length > 0) {
      return typeDef.Enum.Type.Type.toString() === "string"
        ? JSON.stringify(typeDef.Enum.Values[0])
        : typeDef.Enum.Values[0];
    }

    // For sensitive fields without examples, use placeholder
    if (sensitive) {
      const placeholder = sensitivePlaceholder || "...";
      return `"<${placeholder}>"`;
    }

    // Fallback to zero value
    return this.templateExampleZeroValue();
  }

  /**
   * Go syntax import statement strings for the schema type. For base framework
   * types, this will be github.com/hashicorp/terraform-plugin-framework/types.
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  schemaTypeImports(): string[] {
    return this.customTypeConfig
      ? this.customTypeConfig.imports
      : ["github.com/hashicorp/terraform-plugin-framework/types"];
  }

  /** Go syntax type instantiation string, e.g. types.BoolType, for schemas. */
  templateInstantiation(): string {
    return this.customTypeConfig
      ? this.customTypeConfig.schemaType
      : `types.${this.typeName}Type`;
  }

  /** Go syntax struct field string, e.g. Name types.Bool `tfsdk:"name"`,
   *  for Terraform data models. */
  templateDataModelStructField(name: string): string {
    const attributeName = sanitizeTFStateName(name);
    const fieldName = sanitizeFieldName(name);

    return `${fieldName} ${this.templateStructFieldValueType()} \`tfsdk:"${attributeName}"\``;
  }

  /**
   * Go syntax struct field string for PrivateDataModel, e.g.
   * Name string `json:"name"`.
   * Returns undefined if the type is not supported in PrivateDataModel.
   */
  templatePrivateDataModelStructField(name: string): string | undefined {
    const goType = this.templatePrivateDataModelStructFieldType();

    if (goType === undefined) {
      return undefined;
    }

    const attributeName = sanitizeTFStateName(name);
    const fieldName = sanitizeFieldName(name);

    return `${fieldName} ${goType} \`json:"${attributeName}"\``;
  }

  /**
   * Go syntax templating of raw value (e.g. from OAS default or example) to
   * framework type value, e.g. types.StringValue("example").
   */
  templateFrameworkValue(value: any): string {
    return `types.${this.typeName}Value(${value})`;
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion.
   *
   * For base framework types, this will be the types package Value functions.
   * For custom types, this will also include the Typable interface ValueFrom
   * method logic.
   *
   * If terraformAccessor is undefined, the result will be the value which may
   * contain new lines and other Go syntax.
   */
  templateSDKToTerraform(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    _terraformLegacyTypes: boolean,
    operation: TerraformOperation,
  ): string {
    const terraformValue = this.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );

    const lastNewlineIndex = terraformValue.lastIndexOf("\n");

    if (lastNewlineIndex === -1) {
      return `${terraformAccessor} = ${terraformValue}`;
    }

    return (
      terraformValue.substring(0, lastNewlineIndex + 1) +
      `${terraformAccessor} = ` +
      terraformValue.substring(lastNewlineIndex + 1, terraformValue.length)
    );
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion of the value.
   *
   * For base framework types, this will be the types package Value functions.
   * For custom types, this will also include the Typable interface ValueFrom
   * method logic.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    _sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    const terraformValue = sdkTypeOptional
      ? `types.${this.typeName}PointerValue(${sdkAccessor})`
      : `types.${this.typeName}Value(${sdkAccessor})`;

    if (!this.customTypeConfig) {
      return terraformValue;
    }

    const diagsVariable = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Diags",
    );
    const valuableVariable = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Valuable",
    );

    const terraformValuable = `${this.customTypeConfig.schemaType}.ValueFrom${this.typeName}(ctx, ${terraformValue})`;

    return (
      `${valuableVariable}, ${diagsVariable} := ${terraformValuable}\n` +
      `diags.Append(${diagsVariable}...)\n` +
      `${valuableVariable}.(${this.customTypeConfig.valueType})`
    );
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion. For base framework types, this will be
   * github.com/hashicorp/terraform-plugin-framework/types.
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
  ): string[] {
    return this.schemaTypeImports();
  }

  /** Go syntax struct field value type string, e.g. types.Bool, for Terraform
   *  data models. */
  templateStructFieldValueType(): string {
    return this.customTypeConfig
      ? this.customTypeConfig.valueType
      : `types.${this.typeName}`;
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields
   * (e.g. "string", "bool", "int64").
   * Returns undefined for unsupported types (collections, complex types).
   * Subclasses for supported primitive types override this.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return undefined;
  }

  /**
   * Go syntax for extracting the Go builtin value from a framework type value,
   * e.g. "ValueString()".
   * Returns undefined if the type does not support directly extracting value.
   */
  templateValueMethod(pointer: boolean = false): string | undefined {
    if (pointer) {
      return `Value${this.typeName}Pointer()`;
    }

    return `Value${this.typeName}()`;
  }

  /**
   * Go syntax for constructing a framework type value from a Go variable
   * reference, e.g. "types.StringValue(x)".
   * Returns undefined if the type is not supported in PrivateDataModel.
   *
   * Note: This intentionally does not delegate to templateFrameworkValue()
   * because that method is designed for literal values (OAS defaults/examples)
   * and some types like FrameworkStringType add quoting around the value.
   */
  templatePrivateDataModelFrameworkValue(value: string): string | undefined {
    if (this.templatePrivateDataModelStructFieldType() === undefined) {
      return undefined;
    }

    return `types.${this.typeName}Value(${value})`;
  }

  /** Go syntax Terraform to SDK string. */
  templateTerraformToSDK(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    terraformLegacyTypes: boolean,
    diagsAccessor: string = "diags",
  ): string {
    const sdkValue = this.templateTerraformToSDKValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeEnumTypeCast,
      sdkTypeOptional,
      terraformAccessor,
    );
    const result: string[] = [
      `if !${terraformAccessor}.IsUnknown() {`,
      `${sdkAccessor} = ${sdkValue}`,
      `}`,
    ];

    return result.join("\n");
  }

  /**
   * Go syntax import statement strings for templating Terraform to Go SDK
   * conversion.
   */
  templateTerraformToSDKImports(
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    _sdkTypeOptional: boolean,
  ): string[] {
    return sdkTypeEnumTypeCast && sdkTypeDef.Type.toString() === "enum"
      ? [getImportScope(sdkTypeDef)]
      : [];
  }

  /**
   * Go syntax templating of Terraform to Go SDK to conversion of the value.
   */
  templateTerraformToSDKValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    terraformAccessor: string,
  ): string {
    let result = `${terraformAccessor}.${this.templateValueMethod(
      sdkTypeOptional,
    )}`;

    if (sdkTypeEnumTypeCast && sdkTypeDef.Type.toString() === "enum") {
      const enumType = templateType(sdkTypeDef);
      const enumTypeCast = sdkTypeOptional ? `(*${enumType})` : enumType;

      result = `${enumTypeCast}(${result})`;
    }

    return result;
  }

  /**
   * Go syntax import statement strings for the value type. For base framework
   * types, this will be github.com/hashicorp/terraform-plugin-framework/types.
   * For custom types, this will be a subset of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  valueTypeImports(): string[] {
    return this.customTypeConfig
      ? this.customTypeConfigValueTypeImports()
      : ["github.com/hashicorp/terraform-plugin-framework/types"];
  }

  /**
   * In the x-speakeasy-terraform-custom-type extension, the imports are a list
   * of all the imports needed for the schema type and value type to simplify
   * the configuration for customers. Value type imports are usually a subset of
   * the schema type imports. This function returns the imports needed for the
   * value type.
   */
  private customTypeConfigValueTypeImports(): string[] {
    if (!this.customTypeConfig) {
      return [];
    }

    const imports = this.customTypeConfig.imports;

    if (!imports) {
      return [];
    }

    // If there is only one import, assume it is the import necessary for the
    // value type and return it early.
    if (imports.length === 1) {
      return imports;
    }

    const valueType = this.customTypeConfig.valueType;
    const valueTypeParts = valueType.split(".");

    if (valueTypeParts.length < 2) {
      return [];
    }

    // e.g. customtypes of customtypes.Example
    const valueTypePackage = valueTypeParts[0];
    let result: string[] = [];

    imports.forEach((importStr) => {
      const importParts = importStr.split("/");

      if (importParts.length < 2) {
        return;
      }

      // NOTE: Custom type configurations do not currently support import
      // aliases. If import aliases are ever needed, this logic will need to be
      // updated to handle them anyways. Consider adding new configuration of
      // valueTypeAliasImports: Record<string, string> or similar.
      const importPackage = importParts[importParts.length - 1];

      if (importPackage === valueTypePackage) {
        result.push(importStr);
      }
    });

    return result;
  }
}

/** Base representation of a terraform-plugin-framework types/basetype with
 *  attribute types (mapping of string names to types). */
abstract class FrameworkTypeBaseWithAttrTypes extends FrameworkTypeBase {
  attrTypes: FrameworkTypeAttributeTypes;

  constructor(
    typeName: FrameworkTypeBaseTypeName,
    attrTypes: FrameworkTypeAttributeTypes,
    customTypeConfig: CustomTypeConfig | undefined,
  ) {
    super(typeName, customTypeConfig);
    this.attrTypes = attrTypes;
  }

  /**
   * Go syntax import statement strings for the schema type.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/attr
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the attribute types
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  schemaTypeImports(): string[] {
    const result = super.schemaTypeImports();
    result.push("github.com/hashicorp/terraform-plugin-framework/attr");
    Object.values(this.attrTypes).forEach((attrType) => {
      attrType.schemaTypeImports().forEach((importStr) => {
        if (!result.includes(importStr)) {
          result.push(importStr);
        }
      });
    });
    return result;
  }

  /**
   * Go syntax templating of raw value (e.g. from OAS default or example) to
   * framework type value, e.g. types.StringValue("example").
   */
  templateFrameworkValue(value: any): string {
    throw new Error(
      `templateFrameworkValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax type instantiation string, e.g. types.TupleType{...} */
  templateInstantiation(): string {
    const result: string[] = [
      `types.${this.typeName}Type{`,
      `AttrTypes: map[string]attr.Type{`,
    ];
    Object.keys(this.attrTypes)
      .sort()
      .forEach((attrName) => {
        result.push(
          `${templateBuiltinString(attrName)}: ${this.attrTypes[
            attrName
          ].templateInstantiation()},`,
        );
      });
    result.push(`},`, `}`);

    return result.join("\n");
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/attr
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the attribute types
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateSDKToTerraformImports(
      sdkTypeDef,
      sdkTypeOptional,
    );

    result.push("github.com/hashicorp/terraform-plugin-framework/attr");

    Object.keys(this.attrTypes).forEach((attrName) => {
      const sdkFieldDef = sdkTypeDef.Fields.find(
        (field) => sanitizeTFStateName(field.Name) === attrName,
      );

      if (!sdkFieldDef) {
        return;
      }

      this.attrTypes[attrName]
        .templateSDKToTerraformImports(
          sdkFieldDef.Type,
          sdkFieldDef.Optional || sdkFieldDef.Nullable,
        )
        .forEach((importStr) => {
          if (!result.includes(importStr)) {
            result.push(importStr);
          }
        });
    });

    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion of the value.
   */
  templateSDKToTerraformValue(
    _symbolManager: Record<string, boolean>,
    _fieldName: string,
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
    _sdkAccessor: string,
  ): string {
    throw new Error(
      `templateSDKToTerraformValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax Terraform to SDK string. */
  templateTerraformToSDK(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    terraformLegacyTypes: boolean,
    diagsAccessor: string = "diags",
  ): string {
    throw new Error(
      `templateTerraformToSDK not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /**
   * Returns undefined as AttrTypes types do not support directly extracting value.
   */
  templateValueMethod(pointer: boolean = false): string | undefined {
    return undefined;
  }
}

/** Base representation of a terraform-plugin-framework types/basetype with
 *  element type (collection of single type). */
abstract class FrameworkTypeBaseWithElemType extends FrameworkTypeBase {
  elemType: FrameworkType;

  constructor(
    typeName: FrameworkTypeBaseTypeName,
    elemType: FrameworkType,
    customTypeConfig: CustomTypeConfig | undefined,
  ) {
    super(typeName, customTypeConfig);
    this.elemType = elemType;
  }

  /**
   * Go syntax import statement strings for the schema type.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the element type
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  schemaTypeImports(): string[] {
    const result = super.schemaTypeImports();
    this.elemType.schemaTypeImports().forEach((importStr) => {
      if (!result.includes(importStr)) {
        result.push(importStr);
      }
    });
    return result;
  }

  /**
   * Go syntax templating of raw value (e.g. from OAS default or example) to
   * framework type value, e.g. types.StringValue("example").
   */
  templateFrameworkValue(value: any): string {
    throw new Error(
      `templateFrameworkValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax type instantiation string, e.g. types.ListType{...} */
  templateInstantiation(): string {
    return (
      `types.${this.typeName}Type{\n` +
      `ElemType: ${this.elemType.templateInstantiation()},\n` +
      `}`
    );
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion.
   *
   * For base framework types, this will be the types package Value functions.
   * For custom types, this will also include the Typable interface ValueFrom
   * method logic.
   *
   * If terraformLegacyTypes is true, this will use Go slices and append() for
   * the Terraform type handling instead of the framework type.
   */
  templateSDKToTerraform(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    terraformLegacyTypes: boolean,
    operation: TerraformOperation,
  ): string {
    if (terraformLegacyTypes) {
      const elemType = this.elemType.templateStructFieldValueType();
      const elemValue = this.elemType.templateSDKToTerraformValue(
        symbolManager,
        fieldName,
        sdkTypeDef.ItemType,
        sdkTypeDef.ContainsNull,
        `v`,
      );
      const result: string[] = [];

      // Sets must always be reinitialized to prevent duplicates.
      // Lists should only use nil check for paginated operations.
      const isPaginated = operation.APIOperation.Extensions?.Pagination;

      if (!isPaginated) {
        // Sets and non-paginated lists: always reinitialize
        result.push(
          `${terraformAccessor} = make([]${elemType}, 0, len(${sdkAccessor}))`,
          `for _, v := range ${sdkAccessor} {`,
          `${terraformAccessor} = append(${terraformAccessor}, ${elemValue})`,
          `}`,
        );
      } else {
        // Paginated Lists: Initialize to empty slice if nil. This allows:
        // 1. Clearing stale data when set to nil before pagination loops
        // 2. Preserving accumulated data during multi-page pagination
        result.push(
          `if ${terraformAccessor} == nil {`,
          `${terraformAccessor} = make([]${elemType}, 0, len(${sdkAccessor}))`,
          `}`,
          `for _, v := range ${sdkAccessor} {`,
          `${terraformAccessor} = append(${terraformAccessor}, ${elemValue})`,
          `}`,
        );
      }

      return result.join("\n");
    }

    const diagsVariable = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Diags",
    );
    const valueVariable = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Value",
    );

    const terraformValue = `types.${
      this.typeName
    }ValueFrom(ctx, ${this.elemType.templateInstantiation()}, ${sdkAccessor})`;

    const result: string[] = [
      `${valueVariable}, ${diagsVariable} := ${terraformValue}`,
      `diags.Append(${diagsVariable}...)`,
    ];

    if (!this.customTypeConfig) {
      result.push(`${terraformAccessor} = ${valueVariable}`);
      return result.join("\n");
    }

    const valuableVariable = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Valuable",
    );

    const terraformValuable = `${this.customTypeConfig.schemaType}.ValueFrom${this.typeName}(ctx, ${valueVariable})`;

    result.push(
      `${valuableVariable}, ${diagsVariable} := ${terraformValuable}`,
      `diags.Append(${diagsVariable}...)`,
      `${terraformAccessor}, _ = ${valuableVariable}.(${this.customTypeConfig.valueType})`,
    );

    return result.join("\n");
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the element type
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateSDKToTerraformImports(
      sdkTypeDef,
      sdkTypeOptional,
    );

    const elemTypeDef = sdkTypeDef.ItemType;
    const elemTypeOptional = sdkTypeDef.ContainsNull;

    this.elemType
      .templateSDKToTerraformImports(elemTypeDef, elemTypeOptional)
      .forEach((importStr) => {
        if (!result.includes(importStr)) {
          result.push(importStr);
        }
      });

    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion of the value.
   */
  templateSDKToTerraformValue(
    _symbolManager: Record<string, boolean>,
    _fieldName: string,
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
    _sdkAccessor: string,
  ): string {
    throw new Error(
      `templateSDKToTerraformValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax Terraform to SDK string. */
  templateTerraformToSDK(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    terraformLegacyTypes: boolean,
    diagsAccessor: string = "diags",
  ): string {
    const sdkTypeName = templateType(sdkTypeDef, { scope: "tf" });
    const result: string[] = [];

    if (terraformLegacyTypes) {
      if (sdkTypeOptional) {
        result.push(
          `var ${sdkAccessor} ${sdkTypeName}`,
          `if ${terraformAccessor} != nil {`,
          `${sdkAccessor} = make(${sdkTypeName}, 0, len(${terraformAccessor}))`,
        );
      } else {
        result.push(
          `${sdkAccessor} := make(${sdkTypeName}, 0, len(${terraformAccessor}))`,
        );
      }

      let elemType = templateType(sdkTypeDef.ItemType, { scope: "tf" });

      if (sdkTypeDef.ContainsNull) {
        elemType = `(*${elemType})`;
      }

      const elemVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName,
        "Item",
      );
      const elemValue = this.elemType.templateTerraformToSDKValue(
        symbolManager,
        fieldName,
        sdkTypeDef.ItemType,
        false,
        sdkTypeDef.ContainsNull,
        elemVariable,
      );

      result.push(
        `for _, ${elemVariable} := range ${terraformAccessor} {`,
        `${sdkAccessor} = append(${sdkAccessor}, ${elemType}(${elemValue}))`,
        `}`,
      );

      if (sdkTypeOptional) {
        result.push(`}`);
      }

      return result.join("\n");
    }

    const templateAccessorDeclaration = (varName: string) =>
      sdkTypeOptional
        ? `var ${varName} ${sdkTypeName}`
        : `${varName} := ${sdkTypeName}{}`;

    if (!sdkAccessor.includes(".")) {
      result.push(
        templateAccessorDeclaration(sdkAccessor),
        `if !${terraformAccessor}.IsUnknown() && !${terraformAccessor}.IsNull() {`,
        `${diagsAccessor}.Append(${terraformAccessor}.ElementsAs(ctx, &${sdkAccessor}, true)...)`,
        "}",
      );
    } else {
      // `sdkAccessor` cannot be used as variable name since it is nested. Using `fieldName` instead.
      result.push(
        `if !${terraformAccessor}.IsUnknown() && !${terraformAccessor}.IsNull() {`,
        templateAccessorDeclaration(fieldName),
        `${diagsAccessor}.Append(${terraformAccessor}.ElementsAs(ctx, &${fieldName}, true)...)`,
        `${sdkAccessor} = ${fieldName}`,
        "}",
      );
    }

    return result.join("\n");
  }

  /**
   * Go syntax import statement strings for templating Terraform to Go SDK
   * conversion.
   */
  templateTerraformToSDKImports(
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateTerraformToSDKImports(
      sdkTypeDef,
      sdkTypeEnumTypeCast,
      sdkTypeOptional,
    );

    const elemTypeDef = sdkTypeDef.ItemType;
    const elemTypeOptional = sdkTypeDef.ContainsNull;

    this.elemType
      .templateTerraformToSDKImports(
        elemTypeDef,
        sdkTypeEnumTypeCast,
        elemTypeOptional,
      )
      .forEach((importStr) => {
        if (!result.includes(importStr)) {
          result.push(importStr);
        }
      });

    return result;
  }

  /**
   * Returns undefined as ElemType types do not support directly extracting value.
   */
  templateValueMethod(pointer: boolean = false): string | undefined {
    return undefined;
  }
}

/** Base representation of a terraform-plugin-framework types/basetype with
 *  element types (collection of array indexed types). */
abstract class FrameworkTypeBaseWithElemTypes extends FrameworkTypeBase {
  elemTypes: FrameworkTypeElementTypes;

  constructor(
    typeName: FrameworkTypeBaseTypeName,
    elemTypes: FrameworkTypeElementTypes,
    customTypeConfig: CustomTypeConfig | undefined,
  ) {
    super(typeName, customTypeConfig);
    this.elemTypes = elemTypes;
  }

  /**
   * Go syntax import statement strings for the schema type.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the element types
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  schemaTypeImports(): string[] {
    const result = super.schemaTypeImports();
    this.elemTypes.forEach((elemType) => {
      elemType.schemaTypeImports().forEach((importStr) => {
        if (!result.includes(importStr)) {
          result.push(importStr);
        }
      });
    });
    return result;
  }

  /**
   * Go syntax templating of raw value (e.g. from OAS default or example) to
   * framework type value, e.g. types.StringValue("example").
   */
  templateFrameworkValue(value: any): string {
    throw new Error(
      `templateFrameworkValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax type instantiation string, e.g. types.TupleType{...} */
  templateInstantiation(): string {
    return (
      `types.${this.typeName}Type{\n` +
      `ElemTypes: [\n` +
      this.elemTypes.map((t) => t.templateInstantiation() + ",\n") +
      `],\n` +
      `}`
    );
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - Any imports needed for the element types
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
  ): string[] {
    throw new Error(
      `templateSDKToTerraformImports not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /**
   * Go syntax templating of Go SDK to Terraform conversion of the value.
   */
  templateSDKToTerraformValue(
    _symbolManager: Record<string, boolean>,
    _fieldName: string,
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
    _sdkAccessor: string,
  ): string {
    throw new Error(
      `templateSDKToTerraformValue not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /** Go syntax Terraform to SDK string. */
  templateTerraformToSDK(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
    terraformAccessor: string,
    terraformLegacyTypes: boolean,
    diagsAccessor: string = "diags",
  ): string {
    throw new Error(
      `templateTerraformToSDK not implemented for FrameworkType: ${this.typeName}`,
    );
  }

  /**
   * Returns undefined as ElemTypes types do not support directly extracting value.
   */
  templateValueMethod(pointer: boolean = false): string | undefined {
    return undefined;
  }
}

/** Representation of terraform-plugin-framework basetypes.BoolType. */
class FrameworkBoolType extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Bool", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "bool";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "false";
  }
}

/** Representation of terraform-plugin-framework basetypes.DynamicType. */
class FrameworkDynamicType extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Dynamic", customTypeConfig);
  }
}

/** Representation of terraform-plugin-framework basetypes.Float32Type. */
class FrameworkFloat32Type extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Float32", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "float32";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "0.0";
  }

  /**
   * Go syntax templating of Go SDK to Terraform value conversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    if (sdkTypeDef.Type.toString() === "enum") {
      // Unwrap enum type
      sdkAccessor = sdkTypeOptional
        ? `(*float32)(${sdkAccessor})`
        : `float32(${sdkAccessor})`;
    }

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }
}

/** Representation of terraform-plugin-framework basetypes.Float64Type. */
class FrameworkFloat64Type extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Float64", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "float64";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "0.0";
  }

  /**
   * Go syntax templating of Go SDK to Terraform value conversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    if (sdkTypeDef.Type.toString() === "enum") {
      // Unwrap enum type
      sdkAccessor = sdkTypeOptional
        ? `(*float64)(${sdkAccessor})`
        : `float64(${sdkAccessor})`;
    }

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }
}

/** Representation of terraform-plugin-framework basetypes.Int32Type. */
class FrameworkInt32Type extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Int32", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "int32";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "0";
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - If sdkTypeOptional is true, {ROOT_PACKAGE}/internal/provider/typeconvert
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateSDKToTerraformImports(
      sdkTypeDef,
      sdkTypeOptional,
    );

    if (sdkTypeOptional) {
      result.push(`${getRootPackage()}/internal/provider/typeconvert`);
    }

    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform value conversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    // Go SDK uses int/*int/*int64 instead of matching int32/*int32
    // First, unwrap enum type
    if (sdkTypeDef.Type.toString() === "enum") {
      const enumType = sdkTypeDef.Enum.Type.Type.toString();

      switch (enumType) {
        case "int32":
          sdkAccessor = sdkTypeOptional
            ? `typeconvert.IntPointerToInt32Pointer((*int)(${sdkAccessor}))`
            : `int32(int(${sdkAccessor}))`;
          break;
        case "integer":
          // Either of these may panic. The generation should be updated to
          // always pick the largest integer type necessary between all requests
          // and responses to avoid the potential panics (and this code will go
          // away).
          sdkAccessor = sdkTypeOptional
            ? `typeconvert.Int64PointerToInt32Pointer((*int64)(${sdkAccessor}))`
            : `int32(${sdkAccessor})`;
          break;
        default:
          throw new Error(
            `Unsupported enum type for ${this.typeName} Go SDK to Terraform conversion: ${enumType}`,
          );
      }

      return super.templateSDKToTerraformValue(
        symbolManager,
        fieldName,
        sdkTypeDef,
        sdkTypeOptional,
        sdkAccessor,
      );
    }

    switch (sdkTypeDef.Type.toString()) {
      case "int32":
        sdkAccessor = sdkTypeOptional
          ? `typeconvert.IntPointerToInt32Pointer(${sdkAccessor})`
          : `int32(${sdkAccessor})`;
        break;
      case "integer":
        // Either of these may panic. The generation should be updated to
        // always pick the largest integer type necessary between all requests
        // and responses to avoid the potential panics (and this code will go
        // away).
        sdkAccessor = sdkTypeOptional
          ? `typeconvert.Int64PointerToInt32Pointer(${sdkAccessor})`
          : `int32(${sdkAccessor})`;
        break;
    }

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }

  /**
   * Go syntax templating of Terraform to Go SDK to conversion of the value.
   */
  templateTerraformToSDKValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
    terraformAccessor: string,
  ): string {
    let sdkValue = super.templateTerraformToSDKValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      false, // handled below
      sdkTypeOptional,
      terraformAccessor,
    );

    // Go SDK uses int/*int instead of matching int32/*int32
    if (sdkTypeDef.Type.toString() === "enum") {
      const enumType = sdkTypeDef.Enum.Type.Type.toString();

      switch (enumType) {
        case "int32":
        case "integer":
          sdkValue = sdkTypeOptional
            ? `typeconvert.Int32PointerToIntPointer(${sdkValue})`
            : `int(${sdkValue})`;
          break;
        default:
          throw new Error(
            `Unsupported enum type for ${this.typeName} Terraform to Go SDK conversion: ${enumType}`,
          );
      }

      if (sdkTypeEnumTypeCast) {
        const enumType = templateType(sdkTypeDef);
        const enumTypeCast = sdkTypeOptional ? `(*${enumType})` : enumType;

        sdkValue = `${enumTypeCast}(${sdkValue})`;
      }

      return sdkValue;
    }

    switch (sdkTypeDef.Type.toString()) {
      case "int32":
        sdkValue = sdkTypeOptional
          ? `typeconvert.Int32PointerToIntPointer(${sdkValue})`
          : `int(${sdkValue})`;
        break;
    }

    return sdkValue;
  }

  /**
   * Go syntax import statement strings for templating Terraform to Go SDK
   * conversion.
   */
  templateTerraformToSDKImports(
    sdkTypeDef: TypeDef,
    sdkTypeEnumTypeCast: boolean,
    sdkTypeOptional: boolean,
  ): string[] {
    if (sdkTypeOptional) {
      return [`${getRootPackage()}/internal/provider/typeconvert`];
    }

    return [];
  }
}

/** Representation of terraform-plugin-framework basetypes.Int64Type. */
class FrameworkInt64Type extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Int64", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "int64";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "0";
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - If sdkTypeDef is enum, enum type is int32, and sdkTypeOptional is true,
   *    {ROOT_PACKAGE}/internal/provider/typeconvert
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateSDKToTerraformImports(
      sdkTypeDef,
      sdkTypeOptional,
    );

    if (
      sdkTypeDef.Type.toString() === "enum" &&
      sdkTypeDef.Enum.Type.Type.toString() === "int32" &&
      sdkTypeOptional
    ) {
      result.push(`${getRootPackage()}/internal/provider/typeconvert`);
    }

    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform valueconversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    if (sdkTypeDef.Type.toString() === "enum") {
      // Unwrap enum type
      const enumType = sdkTypeDef.Enum.Type.Type.toString();

      switch (enumType) {
        case "int32":
          sdkAccessor = sdkTypeOptional
            ? `typeconvert.IntPointerToInt64Pointer((*int)(${sdkAccessor}))`
            : `int64(int(${sdkAccessor}))`;
          break;
        case "integer":
          sdkAccessor = sdkTypeOptional
            ? `(*int64)(${sdkAccessor})`
            : `int64(${sdkAccessor})`;
          break;
        default:
          throw new Error(
            `Unsupported enum type for ${this.typeName} Go SDK to Terraform conversion: ${enumType}`,
          );
      }

      return super.templateSDKToTerraformValue(
        symbolManager,
        fieldName,
        sdkTypeDef,
        sdkTypeOptional,
        sdkAccessor,
      );
    }

    switch (sdkTypeDef.Type.toString()) {
      case "int32":
        sdkAccessor = sdkTypeOptional
          ? `typeconvert.IntPointerToInt64Pointer(${sdkAccessor})`
          : `int64(int(${sdkAccessor}))`;
        break;
    }

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }
}

/** Representation of terraform-plugin-framework basetypes.ListType. */
class FrameworkListType extends FrameworkTypeBaseWithElemType {
  constructor(elemType: FrameworkType, customTypeConfig?: CustomTypeConfig) {
    super("List", elemType, customTypeConfig);
  }
}

/** Representation of terraform-plugin-framework basetypes.MapType. */
class FrameworkMapType extends FrameworkTypeBaseWithElemType {
  constructor(elemType: FrameworkType, customTypeConfig?: CustomTypeConfig) {
    super("Map", elemType, customTypeConfig);
  }
}

/** Representation of terraform-plugin-framework basetypes.NumberType. */
class FrameworkNumberType extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("Number", customTypeConfig);
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    return "0.0";
  }

  /**
   * Go syntax import statement strings for the schema type.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - math/big
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  schemaTypeImports(): string[] {
    if (this.customTypeConfig) {
      return super.schemaTypeImports();
    }

    const result = super.schemaTypeImports();
    result.push("math/big");
    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform value conversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    // Terraform SDK uses *big.Float for high precision numbers
    sdkAccessor = sdkTypeOptional
      ? `Float32PointerToBigFloatPointer(${sdkAccessor})`
      : `big.NewFloat(*${sdkAccessor})`;

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }
}

/** Representation of terraform-plugin-framework basetypes.ObjectType. */
class FrameworkObjectType extends FrameworkTypeBaseWithAttrTypes {
  constructor(
    attrTypes: FrameworkTypeAttributeTypes,
    customTypeConfig?: CustomTypeConfig,
  ) {
    super("Object", attrTypes, customTypeConfig);
  }
}

/** Representation of terraform-plugin-framework basetypes.SetType. */
class FrameworkSetType extends FrameworkTypeBaseWithElemType {
  constructor(elemType: FrameworkType, customTypeConfig?: CustomTypeConfig) {
    super("Set", elemType, customTypeConfig);
  }
}

/** Representation of terraform-plugin-framework basetypes.StringType. */
class FrameworkStringType extends FrameworkTypeBase {
  constructor(customTypeConfig?: CustomTypeConfig) {
    super("String", customTypeConfig);
  }

  /**
   * Go builtin type name for use in PrivateDataModel struct fields.
   */
  templatePrivateDataModelStructFieldType(): string | undefined {
    return "string";
  }

  /**
   * String example zero value for the type. This is used when there are no
   * examples defined for the type.
   */
  templateExampleZeroValue(): string {
    /*
     * As a special case, prevent "" as Terraform validation will return errors
     * in import block configurations, such as:
     * │ Error: Invalid import id argument
     * │
     * |  on import-by-string-id.tf line 3, in import:
     * |   3:  id = ""
     * │
     * The import ID value evaluates to an empty string, please provide a non-empty value.
     */
    return `"..."`;
  }

  /**
   * Go syntax templating of raw value (e.g. from OAS default or example) to
   * framework type value, e.g. types.StringValue("example").
   */
  templateFrameworkValue(value: any): string {
    if (typeof value === "string" && value.charAt(0) !== '"') {
      value = `"${value}"`;
    }

    return `types.${this.typeName}Value(${value})`;
  }

  /**
   * Go syntax import statement strings for templating Go SDK to Terraform
   * conversion.
   *
   * For base framework types this will be:
   *  - github.com/hashicorp/terraform-plugin-framework/types
   *  - If sdkTypeDef is date or date-time,
   *    {ROOT_PACKAGE}/internal/provider/typeconvert
   *
   * For custom types, this will be all of the imports defined in the
   * x-speakeasy-terraform-custom-type extension.
   */
  templateSDKToTerraformImports(
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
  ): string[] {
    const result = super.templateSDKToTerraformImports(
      sdkTypeDef,
      sdkTypeOptional,
    );

    if (
      sdkTypeDef.Type.toString() === "date" ||
      sdkTypeDef.Type.toString() === "date-time"
    ) {
      result.push(`${getRootPackage()}/internal/provider/typeconvert`);
    }

    return result;
  }

  /**
   * Go syntax templating of Go SDK to Terraform value conversion.
   */
  templateSDKToTerraformValue(
    symbolManager: Record<string, boolean>,
    fieldName: string,
    sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    if (sdkTypeDef.Type.toString() === "date") {
      // Convert SDK types.Date
      sdkAccessor = sdkTypeOptional
        ? `typeconvert.DatePointerToStringPointer(${sdkAccessor})`
        : `typeconvert.DateToString(${sdkAccessor})`;
    } else if (sdkTypeDef.Type.toString() === "date-time") {
      // Convert time.Time
      sdkAccessor = sdkTypeOptional
        ? `typeconvert.TimePointerToStringPointer(${sdkAccessor})`
        : `typeconvert.TimeToString(${sdkAccessor})`;
    } else if (sdkTypeDef.Type.toString() === "enum") {
      // Unwrap enum type
      sdkAccessor = sdkTypeOptional
        ? `(*string)(${sdkAccessor})`
        : `string(${sdkAccessor})`;
    }

    return super.templateSDKToTerraformValue(
      symbolManager,
      fieldName,
      sdkTypeDef,
      sdkTypeOptional,
      sdkAccessor,
    );
  }
}

/**
 * Representation of jsontypes.NormalizedType from
 * github.com/hashicorp/terraform-plugin-framework-jsontypes.
 *
 * Used for contentMediaType: application/json string fields. The custom type
 * provides plan-time JSON syntax validation, semantic equality (normalizes
 * whitespace/key ordering), and NewNormalizedValue() /
 * NewNormalizedPointerValue() for TF<->SDK conversion.
 */
class FrameworkJsonNormalizedType extends FrameworkStringType {
  constructor() {
    super({
      imports: [
        "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
      ],
      schemaType: "jsontypes.NormalizedType{}",
      valueType: "jsontypes.Normalized",
    });
  }

  templatePrivateDataModelStructFieldType(): string | undefined {
    return undefined;
  }

  templateExampleZeroValue(): string {
    return `"{}"`;
  }

  templateFrameworkValue(value: any): string {
    if (typeof value === "string" && value.charAt(0) !== '"') {
      value = `"${value}"`;
    }

    return `jsontypes.NewNormalizedValue(${value})`;
  }

  templateSDKToTerraformValue(
    _symbolManager: Record<string, boolean>,
    _fieldName: string,
    _sdkTypeDef: TypeDef,
    sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    return sdkTypeOptional
      ? `jsontypes.NewNormalizedPointerValue(${sdkAccessor})`
      : `jsontypes.NewNormalizedValue(${sdkAccessor})`;
  }

  templateTerraformToSDKImports(
    _sdkTypeDef: TypeDef,
    _sdkTypeEnumTypeCast: boolean,
    _sdkTypeOptional: boolean,
  ): string[] {
    return [];
  }
}

/**
 * Representation of base64types.StandardType from
 * github.com/speakeasy-api/terraform-plugin-framework-base64types.
 *
 * Used for format: binary fields. The custom type provides plan-time base64
 * validation, semantic equality (normalizes padding), and ValueBytes() /
 * NewStandardValueFromBytes() for TF<->SDK conversion.
 */
class FrameworkBase64StandardType extends FrameworkStringType {
  constructor() {
    super({
      imports: [
        "github.com/speakeasy-api/terraform-plugin-framework-base64types/base64types",
      ],
      schemaType: "base64types.StandardType{}",
      valueType: "base64types.Standard",
    });
  }

  templatePrivateDataModelStructFieldType(): string | undefined {
    return undefined;
  }

  templateExampleZeroValue(): string {
    return `""`;
  }

  templateFrameworkValue(value: any): string {
    if (typeof value === "string" && value.charAt(0) !== '"') {
      value = `"${value}"`;
    }

    return `base64types.NewStandardValue(${value})`;
  }

  templateSDKToTerraformValue(
    _symbolManager: Record<string, boolean>,
    _fieldName: string,
    _sdkTypeDef: TypeDef,
    _sdkTypeOptional: boolean,
    sdkAccessor: string,
  ): string {
    // []byte is always nil-able in Go (not a pointer type), so
    // NewStandardValueFromBytes handles nil -> null for both required
    // and optional fields without needing pointer dereference.
    return `base64types.NewStandardValueFromBytes(${sdkAccessor})`;
  }

  templateTerraformToSDKImports(
    _sdkTypeDef: TypeDef,
    _sdkTypeEnumTypeCast: boolean,
    _sdkTypeOptional: boolean,
  ): string[] {
    return [];
  }
}

/** Representation of terraform-plugin-framework basetypes.TupleType. */
class FrameworkTupleType extends FrameworkTypeBaseWithElemTypes {
  constructor(
    elemTypes: FrameworkTypeElementTypes,
    customTypeConfig?: CustomTypeConfig,
  ) {
    super("Tuple", elemTypes, customTypeConfig);
  }
}
