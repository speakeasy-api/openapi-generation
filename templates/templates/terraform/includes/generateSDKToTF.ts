/** Options for generateSDKToModelMethodBody that change during recursion */
interface SDKToModelOptions {
  /** Source of prior data for write-only field preservation (e.g., "itemPriorData") */
  priorDataAccessor?: string;
  /** Whether we're processing an item inside an array or map */
  isCollectionItem?: boolean;
}

/** A resolved response filter: pre-computed field name, value accessor, and SDK pointer info. */
interface ResponseFilterPair {
  /** PascalCase Go field name (sanitized, shared between TF model and SDK type). */
  fieldName: string;
  /** Terraform value accessor method (e.g., "ValueString()"). */
  tfValueMethod: string;
  /** Whether the SDK field is a pointer type (determines nil check + deref). */
  sdkFieldIsPointer: boolean;
}

/** Builds a lookup map from sanitized field names to their FieldDefs. */
function buildFieldMap(fields: FieldDef[]): Map<string, FieldDef> {
  const map = new Map<string, FieldDef>();
  for (const f of fields) {
    map.set(sanitizeFieldName(f.Name), f);
  }
  return map;
}

/**
 * Collects response filter info by walking the entity's SchemaTypeDef
 * fields for ResponseFilter and cross-referencing with the SDK item type
 * fields for Go type info (pointer vs value).
 *
 * Only top-level primitive fields (string, number, bool, integer) are
 * supported — fields where FrameworkType.templateValueMethod returns
 * undefined are silently skipped. Fields with no corresponding SDK field
 * are also skipped (the generated Go code would not compile).
 */
function collectResponseFilters(
  entity: TerraformEntity,
  sdkItemType: TypeDef,
): ResponseFilterPair[] {
  const filters: ResponseFilterPair[] = [];

  if (!entity.SchemaTypeDef?.Fields) {
    return filters;
  }

  const sdkFieldMap = sdkItemType?.Fields
    ? buildFieldMap(sdkItemType.Fields)
    : new Map<string, FieldDef>();

  for (const schemaField of entity.SchemaTypeDef.Fields) {
    if (!schemaField.Type?.Extensions?.ResponseFilter) {
      continue;
    }

    const tfValueMethod =
      FrameworkTypeFromFieldDef(schemaField).templateValueMethod();
    if (!tfValueMethod) {
      continue;
    }

    const fieldName = sanitizeFieldName(schemaField.Name);
    const sdkField = sdkFieldMap.get(fieldName);
    if (!sdkField) {
      continue;
    }

    filters.push({
      fieldName,
      tfValueMethod,
      sdkFieldIsPointer: sdkField.Optional || sdkField.Nullable,
    });
  }

  return filters;
}

/**
 * Generates a Go client-side filter loop that narrows an SDK response array
 * down to items matching user-provided filter values on the TF data model.
 *
 * Returns the filter loop code string, or empty string if no filters exist.
 * When code is returned, the caller should use `filteredItems[0]` instead of
 * the original array accessor.
 */
function generateResponseFilterLoop(
  filters: ResponseFilterPair[],
  arrayAccessor: string,
  sdkItemTypeName: string,
): string {
  if (filters.length === 0) {
    return "";
  }

  let code = "";
  code += `{\n`;
  code += `filteredItems := make([]${sdkItemTypeName}, 0, len(${arrayAccessor}))\n`;
  code += `for _, filterItem := range ${arrayAccessor} {\n`;

  for (const filter of filters) {
    code += `if !r.${filter.fieldName}.IsNull() && !r.${filter.fieldName}.IsUnknown() {\n`;
    if (filter.sdkFieldIsPointer) {
      code += `if filterItem.${filter.fieldName} == nil || *filterItem.${filter.fieldName} != r.${filter.fieldName}.${filter.tfValueMethod} {\n`;
    } else {
      code += `if filterItem.${filter.fieldName} != r.${filter.fieldName}.${filter.tfValueMethod} {\n`;
    }
    code += `continue\n`;
    code += `}\n`;
    code += `}\n`;
  }

  code += `filteredItems = append(filteredItems, filterItem)\n`;
  code += `}\n\n`;

  return code;
}

/**
 * Generates an in-place filter preamble for the "entity-above-array" pattern.
 *
 * When the entity wraps an array field whose items have response-filter fields,
 * the hoisted filter fields exist at the entity level. This function generates
 * Go code that filters the array in-place (mutating the response) before the
 * normal array→list mapping happens.
 *
 * Returns the filter preamble code string, or empty string if no array filter
 * applies.
 */
function generateArrayFieldFilterPreamble(
  entity: TerraformEntity,
  fromType: TypeDef,
  valueAccessor: string,
): string {
  if (!entity.SchemaTypeDef?.Fields || !fromType?.Fields) {
    return "";
  }

  // Collect entity-level response filter fields that are hoisted from array
  // items (entity-above-array pattern). Fields that exist in the SDK type are
  // handled by the normal entity-on-items filter path.
  const sdkFieldMap = buildFieldMap(fromType.Fields);

  interface HoistedFilter {
    fieldName: string;
    tfValueMethod: string;
  }

  const hoistedFilters: HoistedFilter[] = [];
  for (const schemaField of entity.SchemaTypeDef.Fields) {
    if (!schemaField.Type?.Extensions?.ResponseFilter) {
      continue;
    }
    const fieldName = sanitizeFieldName(schemaField.Name);
    if (sdkFieldMap.has(fieldName)) {
      continue;
    }
    const tfValueMethod =
      FrameworkTypeFromFieldDef(schemaField).templateValueMethod();
    if (!tfValueMethod) {
      continue;
    }
    hoistedFilters.push({ fieldName, tfValueMethod });
  }

  if (hoistedFilters.length === 0) {
    return "";
  }

  // Find the array field in the SDK type whose items contain matching fields.
  // Skip arrays whose items ARE the entity — those are handled by the existing
  // entity-on-items filter path (generateResponseFilterLoop).
  for (const fromField of fromType.Fields) {
    if (!isGoArray(fromField.Type) || !fromField.Type.ItemType?.Fields) {
      continue;
    }

    if (fromField.Type.ItemType.FindEntityTypeDef(entity.Name)) {
      continue;
    }

    const itemFieldMap = buildFieldMap(fromField.Type.ItemType.Fields);

    // Check if the item type has fields matching ALL the hoisted filters.
    const matchedFilters: ResponseFilterPair[] = [];
    for (const filter of hoistedFilters) {
      const itemField = itemFieldMap.get(filter.fieldName);
      if (itemField) {
        matchedFilters.push({
          fieldName: filter.fieldName,
          tfValueMethod: filter.tfValueMethod,
          sdkFieldIsPointer: itemField.Optional || itemField.Nullable,
        });
      }
    }

    if (matchedFilters.length !== hoistedFilters.length) {
      continue;
    }

    // Reuse generateResponseFilterLoop for the filter logic, then reassign
    // the filtered result back to the array field in-place.
    const sanitizedFieldName = sanitizeFieldName(fromField.Name);
    const arrayAccessor = `${valueAccessor}.${sanitizedFieldName}`;
    const sdkItemTypeName = sanitizeType(fromField.Type.ItemType, false, "");

    addImportScope(fromField.Type.ItemType);

    const filterCode = generateResponseFilterLoop(
      matchedFilters,
      arrayAccessor,
      sdkItemTypeName,
    );
    return filterCode + `${arrayAccessor} = filteredItems\n` + `}\n`;
  }

  return "";
}

function primitiveReflectCall(def: TypeDef): string {
  switch (String(def.Type)) {
    case "string":
      return "String()";
    case "float32":
      return "Float32()";
    case "int32":
      return "Int32()";
    case "integer":
      return "Int64()";
    case "boolean":
      return "Bool()";
    default:
      return undefined;
  }
}

function isOptionalType(def: TypeDef): boolean {
  if (def.Enum?.Type) {
    return isOptionalType(def.Enum.Type);
  }
  switch (String(def.Type)) {
    case "string":
    case "integer":
    case "int32":
    case "number":
    case "float32":
    case "boolean":
    case "bytes":
      return true;
    default:
      return false;
  }
}

function prettyPrintType(type: TypeDef) {
  if (type.Enum) {
    return `${type.Enum.Type.Type}(enum=${type.Enum.Values.map((i) =>
      JSON.stringify(i),
    ).join(",")})`;
  }
  return type.Type;
}

/** Templates a Go SDK response type to entity data model type method body. */
function generateSDKToModelMethodBody(
  symbolManager: Record<string, boolean>,
  entity: TerraformEntity,
  targetAccessor: string,
  toType: TypeDef,
  toIsOptional: boolean,
  toIsNullable: boolean,
  fromIsOptional: boolean,
  fieldName: string,
  valueAccessor: string,
  fromType: TypeDef,
  indent: number,
  hierarchy: string,
  operation: TerraformOperation,
  opts: SDKToModelOptions = {},
): string {
  const { priorDataAccessor, isCollectionItem = false } = opts;
  // if (targetAccessor.toLowerCase().includes("debug")) {
  //   debug("generateOverride", hierarchy, fromType, toType, entity);
  // }
  if (
    fromType.Extensions?.TerraformIgnore?.DataModel ||
    toType.Extensions?.TerraformIgnore?.DataModel
  ) {
    return "";
  }

  const pagination = operation.APIOperation.Extensions?.Pagination;
  const subAccessor = sanitizeFieldName(fieldName)
    ? "." + sanitizeFieldName(fieldName)
    : "";
  const associatedName =
    sanitizeVariableName(fieldName) ||
    sanitizeVariableName(targetAccessor.split(".").pop()) ||
    sanitizeVariableName(valueAccessor.split(".").pop());

  function calcHierarchy(obj: string) {
    return `${hierarchy}.${sanitizeTFStateName(obj)}`;
  }

  if (toType.Extensions?.All[`x-speakeasy-wrapped-${hierarchy}`]) {
    const wrappedField = sanitizeFieldName(
      toType.Extensions?.All[`x-speakeasy-wrapped-${hierarchy}`],
    );
    const intoField = toType.Fields.find(
      (field) => sanitizeFieldName(field.Name) == wrappedField,
    );
    if (!intoField) {
      throw new Error(
        `PANIC(${JSON.stringify(wrappedField)} not in ${JSON.stringify(
          fromType,
        )})`,
      );
    }
    let subBuilder = ``;
    // Allocate the wrapped field pointer to prevent nil dereference. When
    // fromIsOptional, the allocation is placed inside the value nil check
    // so the field is only instantiated when there is response data.
    const wrappedFieldSymbol = intoField.Type.Extensions?.All["Symbol"];
    const needsWrappedAllocation = wrappedFieldSymbol && !isCollectionItem;

    if (needsWrappedAllocation) {
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );
      if (fromIsOptional) {
        subBuilder += `if ${valueAccessor} != nil {\n`;
      }
      subBuilder += `if ${targetAccessor}.${wrappedField} == nil {\n`;
      subBuilder += `${targetAccessor}.${wrappedField} = &tfTypes.${wrappedFieldSymbol}{}\n`;
      subBuilder += `}\n`;
    }
    subBuilder += generateSDKToModelMethodBody(
      symbolManager,
      entity,
      targetAccessor + "." + wrappedField,
      intoField.Type,
      intoField.Optional || intoField.Nullable,
      intoField.Nullable,
      needsWrappedAllocation ? false : fromIsOptional,
      "",
      valueAccessor,
      fromType,
      indent,
      calcHierarchy(entity.Name),
      operation,
      opts,
    );
    if (needsWrappedAllocation && fromIsOptional) {
      subBuilder += `}\n`;
    }
    // Now make all fields, pulling them from the constructed object.
    return subBuilder;
  }

  if (fromType == undefined || toType == undefined) {
    throw new Error(
      `PANIC(${JSON.stringify(toType)} !== ${JSON.stringify(fromType)})`,
    );
  }

  // Special case for response schema of inline array items entity. This logic
  // potentially not comprehensive for non-class item types, but this covers at
  // least one known problematic case when the SDK response handling was
  // adjusted to use data methods for all schema levels instead of just entity.
  if (
    toType.Extensions?.All["x-speakeasy-root"] &&
    isGoArray(fromType) &&
    // NOTE: This may need fromType.ItemType.Type.toString() === "class" check
    // here and/or other function changes to cover other array item type cases.
    fromType.ItemType.FindEntityTypeDef(entity.Name)
  ) {
    // Check if the item type has a compatible data model conversion target.
    const arrayItemTarget = operation.FindResponseSDKMethodTarget(
      fromType.ItemType,
    );

    if (!arrayItemTarget) {
      return "";
    }

    const dataRefreshFromSDKMethodName =
      arrayItemTarget.SDKResponseMethod().MethodName;

    // Check for response filter fields to generate client-side filtering.
    const filters = collectResponseFilters(entity, fromType.ItemType);

    if (filters.length > 0) {
      const sdkItemTypeName = sanitizeType(fromType.ItemType, false, "");
      addImportScope(fromType.ItemType);
      const filterCode = generateResponseFilterLoop(
        filters,
        valueAccessor,
        sdkItemTypeName,
      );
      const refreshFromAccessor = fromType.ContainsNull
        ? `filteredItems[0]`
        : `&filteredItems[0]`;

      return (
        filterCode +
        `if len(filteredItems) == 0 {\n` +
        `if len(${valueAccessor}) == 0 {\n` +
        `diags.AddError("Unexpected response from API", "Missing response body array data.")\n` +
        `} else {\n` +
        `diags.AddError("No matching items found",\n` +
        `"No items in the API response matched the configured filter criteria.")\n` +
        `}\n` +
        `return diags\n` +
        `}\n\n` +
        `diags.Append(${targetAccessor}.${dataRefreshFromSDKMethodName}(ctx, ${refreshFromAccessor})...)\n\n` +
        `if diags.HasError() {\n` +
        `return diags\n` +
        `}\n` +
        `}\n`
      );
    }

    const refreshFromAccessor = fromType.ContainsNull
      ? `${valueAccessor}[0]`
      : `&${valueAccessor}[0]`;

    return (
      `if len(${valueAccessor}) == 0 {\n` +
      `diags.AddError("Unexpected response from API", "Missing response body array data.")\n` +
      `return diags\n` +
      `}\n\n` +
      `diags.Append(${targetAccessor}.${dataRefreshFromSDKMethodName}(ctx, ${refreshFromAccessor})...)\n\n` +
      `if diags.HasError() {\n` +
      `return diags\n` +
      `}\n`
    );
  }

  try {
    handleTypeMismatch(hierarchy, toType, fromType);
  } catch (e) {
    logger().Warn(
      `Incompatible types detected in ${hierarchy} (state[${prettyPrintType(
        toType,
      )}] != response[${prettyPrintType(
        fromType,
      )}]): the request type does not match the response type. No drift detection or import behaviour possible. To fix this, adjust the API signature of the returned attribute to match the request type.`,
    );
    return "";
  }

  switch (toType.Type.toString()) {
    case "boolean":
    case "bytes":
    case "date":
    case "date-time":
    case "float32":
    case "int32":
    case "integer":
    case "number":
    case "string":
      const frameworkType = FrameworkTypeFromTypeDef(toType);

      if (
        toType.Extensions.All["x-speakeasy-terraform-in-entity"] &&
        !fromType.Extensions.All["x-speakeasy-terraform-in-entity"]
      ) {
        return "";
      }

      frameworkType
        .templateSDKToTerraformImports(fromType, fromIsOptional)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      let result =
        frameworkType.templateSDKToTerraform(
          symbolManager,
          fieldName ||
            targetAccessor.split(".").pop() ||
            valueAccessor.split(".").pop(),
          fromType,
          fromIsOptional,
          `${valueAccessor}${subAccessor}`,
          `${targetAccessor}${subAccessor}`,
          false,
          operation,
        ) + "\n";

      if (toType.Extensions?.TerraformAliasTo) {
        const relativeLocation = applyPath(
          `${targetAccessor}${subAccessor}`,
          toType.Extensions.TerraformAliasTo,
        );
        result += `${templateIndent(
          indent,
        )}${relativeLocation} = ${targetAccessor}${subAccessor}\n`;
      }

      return result;
  }

  if (toType.Type == "map") {
    const customTypeConfig = TypeDefTerraformCustomTypeConfig(toType);

    if (customTypeConfig) {
      const frameworkType = FrameworkTypeFromTypeDef(toType);

      frameworkType
        .templateSDKToTerraformImports(fromType, fromIsOptional)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      return (
        frameworkType.templateSDKToTerraform(
          symbolManager,
          fieldName ||
            targetAccessor.split(".").pop() ||
            valueAccessor.split(".").pop(),
          fromType,
          fromIsOptional,
          `${valueAccessor}${subAccessor}`,
          `${targetAccessor}${subAccessor}`,
          false,
          operation,
        ) + "\n"
      );
    }

    let result: string[] = [];

    if (toIsNullable) {
      result.push(`if ${valueAccessor}${subAccessor} != nil {`);
    } else {
      // NOTE: This check is technically incorrect, as a known map of 0 elements
      // should always be created for non-nullable types. However, this logic has
      // been left in place to maintain compatibility with existing code. Given a
      // customer bug report though, this should be revisited and removing this
      // should likely be considered a fix.
      result.push(`if len(${valueAccessor}${subAccessor}) > 0 {`);
    }

    if (primitiveValuePrefix(toType.ItemType, toType.ContainsNull)) {
      const keySymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Key",
      );
      const valueSymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Value",
      );

      const subValueAccessor =
        toType.ItemType.Type.toString() === "int32"
          ? `int32(${valueSymbol})`
          : valueSymbol;

      result.push(
        `${targetAccessor}${subAccessor} = make(map[string]${primitiveValuePrefix(
          toType.ItemType,
          false,
        )}, len(${valueAccessor}${subAccessor}))`,
        `for ${keySymbol}, ${valueSymbol} := range ${valueAccessor}${subAccessor} {`,
        `${targetAccessor}${subAccessor}[${keySymbol}] = ${primitiveValuePrefix(
          fromType.ItemType,
          fromType.ContainsNull,
        )}Value(${subValueAccessor})`,
        `}`,
        `}`,
      );
      return result
        .map((line) => `${templateIndent(indent)}${line}\n`)
        .join("");
    }

    if (
      toType.ItemType.Type.toString() === "array" ||
      toType.ItemType.Type.toString() === "map" ||
      toType.ItemType.Type.toString() === "set"
    ) {
      const keySymbol = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Key",
      );
      const valueSymbol = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Value",
      );
      const resultSymbol = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Result",
      );

      const innerType = templateTypeDefDataModelFieldType(
        toType.ItemType,
        "",
        false,
        indent,
        "provider",
        `${hierarchy}.[]`,
        (fieldName: string, typedef: TypeDef) => {
          if (typedef.Extensions?.All["Symbol"]) {
            return typedef.Extensions.All["Symbol"];
          }
        },
      );

      result.push(
        `${targetAccessor}${subAccessor} = make(map[string]${innerType.content}, len(${valueAccessor}${subAccessor}))`,
        `for ${keySymbol}, ${valueSymbol} := range ${valueAccessor}${subAccessor} {`,
        `var ${resultSymbol} ${innerType.content}`,
        generateSDKToModelMethodBody(
          symbolManager,
          entity,
          resultSymbol,
          toType.ItemType,
          false,
          false,
          false,
          "",
          valueSymbol,
          fromType.ItemType,
          indent + 1,
          `${hierarchy}.[]`,
          operation,
          { isCollectionItem: true },
        ),
        `${targetAccessor}${subAccessor}[${keySymbol}] = ${resultSymbol}`,
        `}`,
        `}`,
      );
      return result
        .map((line) => `${templateIndent(indent)}${line}\n`)
        .join("");
    }
    if (toType.ItemType.Type == "any") {
      addGenImport("encoding/json");
      addGenImport(
        "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
      );

      const keySymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Key",
      );
      const valueSymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Value",
      );
      const tmpSymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Result",
      );

      result.push(
        `${targetAccessor}${subAccessor} = make(map[string]jsontypes.Normalized, len(${valueAccessor}${subAccessor}))`,
        `for ${keySymbol}, ${valueSymbol} := range ${valueAccessor}${subAccessor} {`,
        `${tmpSymbol}, _ := json.Marshal(${valueSymbol})`,
        `${targetAccessor}${subAccessor}[${keySymbol}] = jsontypes.NewNormalizedValue(string(${tmpSymbol}))`,
        `}`,
        `}`,
      );
      return result
        .map((line) => `${templateIndent(indent)}${line}\n`)
        .join("");
    } else if (toType.ItemType.Extensions?.All["Symbol"]) {
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );

      // Reference: internal issue reference
      const itemTypeNeedsValueSymbol =
        toType.ItemType.AssociatedTypes?.length > 0 ||
        toType.ItemType.Fields?.length > 0 ||
        toType.ItemType.ItemType;

      // Check if map values have write-only fields that need prior data preservation
      const itemsNeedPriorData =
        toType.ItemType.TerraformHasNestedWriteOnlyFields(fromType.ItemType);
      const priorMapVar = itemsNeedPriorData
        ? getPluralizedVarSymbolName(
            symbolManager,
            toType.ItemType.Name,
            "PriorMap",
          )
        : undefined;

      const keySymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Key",
      );
      const valueSymbol = itemTypeNeedsValueSymbol
        ? getPluralizedVarSymbolName(
            symbolManager,
            toType.ItemType.Name,
            "Value",
          )
        : "_";
      const tmpSymbol = getPluralizedVarSymbolName(
        symbolManager,
        toType.ItemType.Name,
        "Result",
      );
      const priorValueVar = priorMapVar
        ? getPluralizedVarSymbolName(
            symbolManager,
            toType.ItemType.Name,
            "PriorValue",
          )
        : undefined;
      const typeSymbol = "tfTypes." + toType.ItemType.Extensions.All["Symbol"];

      // Capture prior map before reinitializing (for write-only field preservation).
      // When a parent was reallocated, its captured prior data is the only
      // remaining source; the captured root pointer can be nil on first create.
      if (priorMapVar) {
        if (priorDataAccessor) {
          const priorDataRootVar = priorDataAccessor.split(".")[0];
          result.push(
            `var ${priorMapVar} map[string]${typeSymbol}`,
            `if ${priorDataRootVar} != nil {`,
            `${priorMapVar} = ${priorDataAccessor}${subAccessor}`,
            `}`,
          );
        } else {
          result.push(`${priorMapVar} := ${targetAccessor}${subAccessor}`);
        }
      }

      result.push(
        `${targetAccessor}${subAccessor} = make(map[string]${typeSymbol}, len(${valueAccessor}${subAccessor}))`,
        `for ${keySymbol}, ${valueSymbol} := range ${valueAccessor}${subAccessor} {`,
        `var ${tmpSymbol} ${typeSymbol}`,
      );

      // Look up prior value by key if we're preserving write-only fields
      if (priorMapVar && priorValueVar) {
        result.push(
          `var ${priorValueVar} *${typeSymbol}`,
          `if ${priorMapVar} != nil {`,
          `if priorVal, ok := ${priorMapVar}[${keySymbol}]; ok {`,
          `${priorValueVar} = &priorVal`,
          `}`,
          `}`,
        );
      }

      result.push(
        generateSDKToModelMethodBody(
          symbolManager,
          entity,
          `${tmpSymbol}`,
          toType.ItemType,
          false,
          false,
          false,
          "",
          `${valueSymbol}`,
          fromType.ItemType,
          2,
          `${hierarchy}.[]`,
          operation,
          { priorDataAccessor: priorValueVar, isCollectionItem: true },
        ),
        `${targetAccessor}${subAccessor}[${keySymbol}] = ${tmpSymbol}`,
        `}`,
        `}`,
      );
      return result
        .map((line) => `${templateIndent(indent)}${line}\n`)
        .join("");
    }
  }
  if (toType.Type == "class") {
    // TODO: Migrate or remove x-speakeasy-root usage.
    const onEntity =
      toType.HasEntityName(entity.Name) ||
      toType.Extensions?.All["x-speakeasy-root"];

    let subBuilder = "";
    let conditionalStack: string[] = [];
    // subBuilder += `\n/*\n${JSON.stringify(
    //   {
    //     hierarchy,
    //     fromIsOptional,
    //     toIsOptional,
    //     toType,
    //     fromType,
    //     targetAccessor,
    //   },
    //   null,
    //   2,
    // )}\n*/\n`;

    // Ensure that any class properties that are not present in the Go SDK type,
    // such as writeOnly properties, are not forcibly set to null in the
    // Terraform state by preserving the prior data value.
    // Reference: internal issue reference
    let targetNeedsPriorData = false;

    // Ensure this is not the entity TypeDef to reduce code output since root
    // attributes are fine without this logic.
    if (!onEntity) {
      for (const toTypeField of toType.Fields) {
        if (
          !fromType.Fields.find(
            (fromTypeField) =>
              sanitizeFieldName(fromTypeField.Name) ==
              sanitizeFieldName(toTypeField.Name),
          )
        ) {
          targetNeedsPriorData = true;
          break;
        }
      }
    }

    const hasNestedFieldsNeedingPrior =
      toType.TerraformHasNestedWriteOnlyFields(fromType);

    // Capture prior data if this type or any nested type has write-only fields.
    // For collection items, the parent loop already provides priorDataAccessor, so we don't capture again.
    // We only need to capture if we don't already have a priorDataAccessor from the parent.
    const needsToCapturePriorData =
      !isCollectionItem &&
      (targetNeedsPriorData || (hasNestedFieldsNeedingPrior && !onEntity));

    const targetPriorDataAccessor: string | undefined = needsToCapturePriorData
      ? getPluralizedVarSymbolName(
          symbolManager,
          fieldName || associatedName,
          "PriorData",
        )
      : undefined;

    // For collection items, priorDataAccessor is already the prior item pointer from the parent loop.
    // For non-collection items, we capture from the target before reallocation.
    // The priorDataAccessorForFields is what we use when assigning write-only fields.
    const priorDataAccessorForFields = isCollectionItem
      ? priorDataAccessor
      : targetPriorDataAccessor || priorDataAccessor;

    // If parent was reallocated, use the captured prior data; otherwise use current target
    const priorDataCaptureSource = priorDataAccessor
      ? `${priorDataAccessor}${subAccessor}`
      : `${targetAccessor}${subAccessor}`;

    // Parent prior data could be nil on first create, so we need nil-safe capture
    const needsNilSafePriorDataCapture =
      priorDataAccessor && targetPriorDataAccessor;

    const generatePriorDataCapture = (): string => {
      if (!targetPriorDataAccessor) return "";

      const typeSymbol = toType.Extensions?.All["Symbol"];
      if (needsNilSafePriorDataCapture && typeSymbol && priorDataAccessor) {
        // Check the root variable, not the full path (e.g., check "parentPriorData" not "parentPriorData.Child")
        const priorDataRootVar = priorDataAccessor.split(".")[0];
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
        return (
          `var ${targetPriorDataAccessor} *tfTypes.${typeSymbol}\n` +
          `if ${priorDataRootVar} != nil {\n` +
          `${targetPriorDataAccessor} = ${priorDataCaptureSource}\n` +
          `}\n`
        );
      }
      return `${targetPriorDataAccessor} := ${priorDataCaptureSource}\n`;
    };

    const childpriorDataAccessor = targetPriorDataAccessor || priorDataAccessor;

    // Class/union fields are pointers (for nil tracking), except collection items which are values
    const needsAllocation =
      !onEntity && !isCollectionItem && toType.Extensions?.All["Symbol"];

    const generatePriorDataCaptureAndAllocation = (): string => {
      let result = generatePriorDataCapture();
      if (needsAllocation) {
        result += `${targetAccessor}${subAccessor} = &tfTypes.${toType.Extensions?.All["Symbol"]}{}\n`;
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
      }
      return result;
    };

    if (fromIsOptional && toIsOptional) {
      subBuilder += `if ${valueAccessor}${subAccessor} == nil {\n`;
      subBuilder += `${targetAccessor}${subAccessor} = nil\n`;
      subBuilder += `} else {\n`;
      subBuilder += generatePriorDataCaptureAndAllocation();
      indent++;
    } else if (toIsOptional) {
      subBuilder += generatePriorDataCapture();
      if (needsAllocation) {
        subBuilder += `if ${targetAccessor}${subAccessor} == nil {\n`;
        subBuilder += `${targetAccessor}${subAccessor} = &tfTypes.${toType.Extensions?.All["Symbol"]}{}\n`;
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
        subBuilder += `}\n`;
      }
    } else if (fromIsOptional) {
      conditionalStack.push(`if ${valueAccessor}${subAccessor} != nil {\n`);
      subBuilder += conditionalStack[conditionalStack.length - 1];
      subBuilder += generatePriorDataCaptureAndAllocation();
      indent++;
    } else {
      subBuilder += generatePriorDataCaptureAndAllocation();
    }

    // Entity-above-array filter preamble: when the entity wraps an array
    // whose items have response-filter fields, filter the array in-place
    // before the normal field processing maps it to a TF list.
    if (onEntity) {
      subBuilder += generateArrayFieldFilterPreamble(
        entity,
        fromType,
        `${valueAccessor}${subAccessor}`,
      );
    }

    // Handle via fromType fields to ensure piblings (from hoisting) are
    // included as necessary. Sorting to prevent code churn. Spread operator to
    // prevent sorting from mutating original (.toSorted() not available in Goja
    // environment).
    [...fromType.Fields]
      .sort((a, b) => a.Name.localeCompare(b.Name))
      .forEach((fromTypeField) => {
        const sanitizedFromTypeFieldName = sanitizeFieldName(
          fromTypeField.Name,
        );

        if (
          onEntity &&
          pagination &&
          (sanitizePaginationOutputName(pagination.Outputs.NextURL) ===
            sanitizedFromTypeFieldName ||
            sanitizePaginationOutputName(pagination.Outputs.NumPages) ===
              sanitizedFromTypeFieldName)
        ) {
          return;
        }

        const equivalent = toType.Fields.find(
          (f) => sanitizeFieldName(f.Name) == sanitizedFromTypeFieldName,
        );

        if (equivalent) {
          const fieldPriorDataAccessor = childpriorDataAccessor
            ? `${childpriorDataAccessor}.${sanitizedFromTypeFieldName}`
            : undefined;
          subBuilder += generateSDKToModelMethodBody(
            symbolManager,
            entity,
            `${targetAccessor}${subAccessor}.${sanitizedFromTypeFieldName}`,
            equivalent.Type,
            equivalent.Optional || equivalent.Nullable,
            equivalent.Nullable,
            fromTypeField.Optional || fromTypeField.Nullable,
            "",
            `${valueAccessor}${subAccessor}.${sanitizedFromTypeFieldName}`,
            fromTypeField.Type,
            indent,
            calcHierarchy(equivalent.Name),
            operation,
            { priorDataAccessor: fieldPriorDataAccessor },
          );

          return;
        }

        // From (SDK) type still above the entity, recurse using same to
        // (entity data model) type.
        if (fromTypeField.Type.FindEntityTypeDef(entity.Name)) {
          if (
            isGoArray(fromTypeField.Type) &&
            !fromTypeField.Type.HasEntityName(entity.Name)
          ) {
            // Check if the item type has a compatible data model conversion target.
            const arrayFieldTarget = operation.FindResponseSDKMethodTarget(
              fromTypeField.Type.ItemType,
            );
            const fromFieldAccessor = `${valueAccessor}.${sanitizedFromTypeFieldName}`;

            if (arrayFieldTarget) {
              const dataRefreshFromSDKMethodName =
                arrayFieldTarget.SDKResponseMethod().MethodName;

              // Check for response filter fields to generate client-side filtering.
              const filters = collectResponseFilters(
                entity,
                fromTypeField.Type.ItemType,
              );

              if (filters.length > 0) {
                const sdkItemTypeName = sanitizeType(
                  fromTypeField.Type.ItemType,
                  false,
                  "",
                );
                addImportScope(fromTypeField.Type.ItemType);
                const filterCode = generateResponseFilterLoop(
                  filters,
                  fromFieldAccessor,
                  sdkItemTypeName,
                );
                const refreshFromAccessor = fromTypeField.Type.ContainsNull
                  ? `filteredItems[0]`
                  : `&filteredItems[0]`;

                subBuilder +=
                  filterCode +
                  `if len(filteredItems) == 0 {\n` +
                  `if len(${fromFieldAccessor}) == 0 {\n` +
                  `diags.AddError("Unexpected response from API", "Missing response body array data.")\n` +
                  `} else {\n` +
                  `diags.AddError("No matching items found",\n` +
                  `"No items in the API response matched the configured filter criteria.")\n` +
                  `}\n` +
                  `return diags\n` +
                  `}\n\n` +
                  `diags.Append(${targetAccessor}.${dataRefreshFromSDKMethodName}(ctx, ${refreshFromAccessor})...)\n\n` +
                  `if diags.HasError() {\n` +
                  `return diags\n` +
                  `}\n` +
                  `}\n\n`;
              } else {
                const refreshFromAccessor = fromTypeField.Type.ContainsNull
                  ? `${fromFieldAccessor}[0]`
                  : `&${fromFieldAccessor}[0]`;

                subBuilder +=
                  `if len(${fromFieldAccessor}) == 0 {\n` +
                  `diags.AddError("Unexpected response from API", "Missing response body array data.")\n` +
                  `return diags\n` +
                  `}\n\n` +
                  `diags.Append(${targetAccessor}.${dataRefreshFromSDKMethodName}(ctx, ${refreshFromAccessor})...)\n\n` +
                  `if diags.HasError() {\n` +
                  `return diags\n` +
                  `}\n\n`;
              }
            }

            return;
          }

          const refreshFromAccessor =
            fromTypeField.Optional || fromTypeField.Nullable
              ? `${valueAccessor}.${sanitizedFromTypeFieldName}`
              : `&${valueAccessor}.${sanitizedFromTypeFieldName}`;
          // Check if the field type has a compatible data model conversion target.
          const nonArrayFieldTarget = operation.FindResponseSDKMethodTarget(
            fromTypeField.Type,
          );

          if (nonArrayFieldTarget) {
            const dataRefreshFromSDKMethodName =
              nonArrayFieldTarget.SDKResponseMethod().MethodName;
            subBuilder +=
              `diags.Append(${targetAccessor}.${dataRefreshFromSDKMethodName}(ctx, ${refreshFromAccessor})...)\n\n` +
              `if diags.HasError() {\n` +
              `return diags\n` +
              `}\n\n`;
          }

          return;
        }

        // Reference: internal issue reference
        // If there was no equivalent field in the to (entity data model) type,
        // no entity match, and its an object type, then its possible that this
        // data is an object pibling where its properties were hoisted to the
        // parent object. Try recursing the fromTypeField with the same toType.
        if (
          fromTypeField.Type.Type.toString() === "class" &&
          !fromTypeField?.Type?.Extensions.TerraformIgnore?.DataModel
        ) {
          subBuilder += generateSDKToModelMethodBody(
            symbolManager,
            entity,
            `${targetAccessor}${subAccessor}`,
            toType,
            false,
            false,
            fromTypeField.Optional || fromTypeField.Nullable,
            "",
            `${valueAccessor}${subAccessor}.${sanitizedFromTypeFieldName}`,
            fromTypeField.Type,
            indent,
            calcHierarchy(fromTypeField.Name),
            operation,
            { priorDataAccessor: childpriorDataAccessor, isCollectionItem },
          );
        }
      });

    // Handle writeOnly properties not present in the from (SDK) type.
    // For collection items, priorDataAccessorForFields is the prior item from parent loop.
    // For non-collection items, it's the captured prior data accessor.
    if (priorDataAccessorForFields) {
      const writeOnlyFields = toType.Fields.filter(
        (toTypeField) =>
          !fromType.Fields.find(
            (fromTypeField) =>
              sanitizeFieldName(fromTypeField.Name) ===
              sanitizeFieldName(toTypeField.Name),
          ),
      ).sort((a, b) => a.Name.localeCompare(b.Name));

      if (writeOnlyFields.length > 0) {
        // Prior data can be nil (e.g. after terraform import), so always nil-check
        if (!isCollectionItem) {
          conditionalStack.push(`if ${priorDataAccessorForFields} != nil {\n`);
          subBuilder += conditionalStack[conditionalStack.length - 1];
        }

        // For collection items, the prior data is a pointer that may be nil
        if (isCollectionItem) {
          subBuilder += `if ${priorDataAccessorForFields} != nil {\n`;
        }

        writeOnlyFields.forEach((toTypeField) => {
          const sanitizeToTypeFieldName = sanitizeFieldName(toTypeField.Name);
          subBuilder += `${targetAccessor}${subAccessor}.${sanitizeToTypeFieldName} = ${priorDataAccessorForFields}.${sanitizeToTypeFieldName}\n`;
        });

        if (isCollectionItem) {
          subBuilder += `}\n`;
        }

        if (!isCollectionItem) {
          const lastConditional = conditionalStack.pop();
          if (subBuilder.endsWith(lastConditional)) {
            // drop last conditional from subBuilder
            subBuilder = subBuilder.slice(0, -lastConditional.length);
          } else {
            subBuilder += `}\n`;
          }
        }
      }
    }

    const associatedTypes = toType.AssociatedTypes;
    for (let i = 0; i < associatedTypes.length; i++) {
      const subTypeDef = associatedTypes[i];
      const subTypeDefName = sanitizeTFUnionTypeName(toType, subTypeDef);
      let equivalent = fromType.AssociatedTypes?.find(
        (associated) =>
          sanitizeTFUnionTypeName(fromType, associated) == subTypeDefName,
      );
      if (!equivalent) {
        continue;
      }
      const newTarget = `${targetAccessor}${subAccessor}.${subTypeDefName}`;
      const maybeAccessor = `${valueAccessor}${subAccessor}.${sanitizeFieldName(
        equivalent.Name || sanitizeTFUnionTypeName(fromType, equivalent),
      )}`;
      conditionalStack.push(
        `${templateIndent(indent)}if ${maybeAccessor} != nil {\n`,
      );
      subBuilder += conditionalStack[conditionalStack.length - 1];
      // When the recursive call enters the class handler for this variant, it
      // computes needsAllocation = !onEntity && !isCollectionItem && symbol.
      // If onEntity is true for the variant, the recursive call won't allocate,
      // so we must do it here.
      const variantOnEntity =
        subTypeDef.HasEntityName(entity.Name) ||
        subTypeDef.Extensions?.All["x-speakeasy-root"];
      if (
        variantOnEntity &&
        !isCollectionItem &&
        subTypeDef.Extensions?.All["Symbol"]
      ) {
        subBuilder += `${templateIndent(indent + 1)}${newTarget} = &tfTypes.${
          subTypeDef.Extensions.All["Symbol"]
        }{}\n`;
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
      }
      const unionpriorDataAccessor = childpriorDataAccessor
        ? `${childpriorDataAccessor}.${subTypeDefName}`
        : undefined;
      subBuilder += generateSDKToModelMethodBody(
        symbolManager,
        entity,
        newTarget,
        subTypeDef,
        Boolean(isOptionalType(subTypeDef)),
        Boolean(isOptionalType(subTypeDef)),
        Boolean(isOptionalType(subTypeDef)),
        "",
        maybeAccessor,
        equivalent,
        indent,
        calcHierarchy(subTypeDefName),
        operation,
        { priorDataAccessor: unionpriorDataAccessor },
      );
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    if (fromIsOptional) {
      indent--;
      const lastConditional = conditionalStack.pop();
      if (lastConditional && subBuilder.endsWith(lastConditional)) {
        subBuilder = subBuilder.slice(0, -lastConditional.length);
      } else {
        subBuilder += `${templateIndent(indent)}}\n`;
      }
    }
    subBuilder += ""; //`${templateIndent(indent)}// End Generate(${optional}, ${JSON.stringify(toType.Name)}, ${targetAccessor})\n`
    return subBuilder;
  }
  if (isGoArray(toType)) {
    const customTypeConfig = TypeDefTerraformCustomTypeConfig(toType);

    if (customTypeConfig) {
      const frameworkType = FrameworkTypeFromTypeDef(toType);

      frameworkType
        .templateSDKToTerraformImports(fromType, fromIsOptional)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      return (
        frameworkType.templateSDKToTerraform(
          symbolManager,
          fieldName ||
            targetAccessor.split(".").pop() ||
            valueAccessor.split(".").pop(),
          fromType,
          fromIsOptional,
          `${valueAccessor}${subAccessor}`,
          `${targetAccessor}${subAccessor}`,
          false,
          operation,
        ) + "\n"
      );
    }

    let result: string[] = [];

    if (toIsNullable) {
      result.push(`if ${valueAccessor}${subAccessor} != nil {`);
      indent++;
    }

    if (
      primitiveValuePrefix(toType.ItemType, toType.ContainsNull) ||
      toType.ItemType.Type.toString() === "enum"
    ) {
      const frameworkType = FrameworkTypeFromTypeDef(toType);

      frameworkType
        .templateSDKToTerraformImports(fromType, fromIsOptional)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      result.push(
        frameworkType.templateSDKToTerraform(
          symbolManager,
          fieldName ||
            targetAccessor.split(".").pop() ||
            valueAccessor.split(".").pop(),
          fromType,
          fromIsOptional,
          `${valueAccessor}${subAccessor}`,
          `${targetAccessor}${subAccessor}`,
          true,
          operation,
        ),
      );
    } else if (
      toType.ItemType?.Extensions?.All["Symbol"] &&
      ((toType.ItemType.Type == "union" &&
        toType.ItemType.AssociatedTypes?.length) ||
        (toType.ItemType.Fields && toType.ItemType.Fields.length))
    ) {
      // Only use nil check for paginated operations to preserve data
      // across multiple pages. For non-paginated operations, always reinitialize
      // to prevent stale data accumulation.
      const isPaginated = operation.APIOperation.Extensions?.Pagination;

      // Check if array items have write-only fields that need prior data preservation
      const itemsNeedPriorData =
        toType.ItemType.TerraformHasNestedWriteOnlyFields(fromType.ItemType);
      const priorSliceVar = itemsNeedPriorData
        ? getPluralizedVarSymbolName(
            symbolManager,
            fieldName || associatedName,
            "PriorSlice",
          )
        : undefined;
      const typeSymbol = toType.ItemType?.Extensions?.All["Symbol"];

      // Capture prior slice before reinitializing (for write-only field preservation).
      // When a parent was reallocated, its captured prior data is the only
      // remaining source; the captured root pointer can be nil on first create.
      if (priorSliceVar) {
        if (priorDataAccessor) {
          const priorDataRootVar = priorDataAccessor.split(".")[0];
          result.push(
            `var ${priorSliceVar} []tfTypes.${typeSymbol}`,
            `if ${priorDataRootVar} != nil {`,
            `${priorSliceVar} = ${priorDataAccessor}${subAccessor}`,
            `}`,
          );
        } else {
          result.push(`${priorSliceVar} := ${targetAccessor}${subAccessor}`);
        }
      }

      if (isPaginated) {
        result.push(
          // Initialize to empty slice if nil. This allows:
          // 1. Clearing stale data when set to nil before pagination loops
          // 2. Preserving accumulated data during multi-page pagination
          `if ${targetAccessor}${subAccessor} == nil {`,
          `${targetAccessor}${subAccessor} = []tfTypes.${typeSymbol}{}`,
          `}`,
        );
      } else {
        result.push(
          // Always reinitialize for non-paginated operations
          `${targetAccessor}${subAccessor} = []tfTypes.${typeSymbol}{}`,
        );
      }
      const indexVariable = priorSliceVar
        ? getPluralizedVarSymbolName(
            symbolManager,
            fieldName || associatedName,
            "Idx",
          )
        : undefined;
      const iterVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Item",
      );
      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "",
      );
      const priorItemVar = priorSliceVar
        ? getPluralizedVarSymbolName(
            symbolManager,
            fieldName || associatedName,
            "PriorItem",
          )
        : undefined;
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );

      // Use indexed iteration when we need prior data for positional matching
      const forLoop = priorSliceVar
        ? `for ${indexVariable}, ${iterVariable} := range ${valueAccessor}${subAccessor} {`
        : `for _, ${iterVariable} := range ${valueAccessor}${subAccessor} {`;

      result.push(``, forLoop, `var ${symbolName} tfTypes.${typeSymbol}`, ``);

      // Look up prior item by index if we're preserving write-only fields
      if (priorSliceVar && priorItemVar) {
        result.push(
          `var ${priorItemVar} *tfTypes.${typeSymbol}`,
          `if ${indexVariable} < len(${priorSliceVar}) {`,
          `${priorItemVar} = &${priorSliceVar}[${indexVariable}]`,
          `}`,
          ``,
        );
      }

      result.push(
        generateSDKToModelMethodBody(
          symbolManager,
          entity,
          symbolName,
          toType.ItemType,
          false,
          false,
          false,
          "",
          iterVariable,
          fromType.ItemType,
          indent + 1,
          `${hierarchy}.[]`,
          operation,
          { priorDataAccessor: priorItemVar, isCollectionItem: true },
        ),
        `${targetAccessor}${subAccessor} = append(${targetAccessor}${subAccessor}, ${symbolName})`,
        `}`,
      );
    } else if (toType.ItemType.Type == "any") {
      addGenImport(
        "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
      );

      const iterVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Item",
      );
      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "",
      );

      result.push(
        `${targetAccessor}${subAccessor} = make([]jsontypes.Normalized, 0, len(${valueAccessor}${subAccessor}))`,
        `for _, ${iterVariable} := range ${valueAccessor}${subAccessor} {`,
        `var ${symbolName} jsontypes.Normalized`,
        ``,
        generateSDKToModelMethodBody(
          symbolManager,
          entity,
          symbolName,
          toType.ItemType,
          false,
          false,
          false,
          "",
          iterVariable,
          fromType.ItemType,
          indent + 1,
          `${hierarchy}.[]`,
          operation,
          { isCollectionItem: true },
        ),
        `${targetAccessor}${subAccessor} = append(${targetAccessor}${subAccessor}, ${symbolName})`,
        `}`,
      );
    } else if (toType.ItemType.Type == "map") {
      let subBuilder = "";
      subBuilder += `${templateIndent(
        indent,
      )}${targetAccessor}${subAccessor} = nil\n`;
      const iterVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Item",
      );
      subBuilder += `${templateIndent(
        indent,
      )}for _, ${iterVariable} := range ${valueAccessor}${subAccessor} {\n`;
      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "",
      );
      const innerType = templateTypeDefDataModelFieldType(
        toType.ItemType,
        "",
        false,
        indent,
        "provider",
        `${hierarchy}.[]`,
        (fieldName: string, typedef: TypeDef) => {
          if (typedef.Extensions?.All["Symbol"]) {
            return typedef.Extensions.All["Symbol"];
          }
        },
      );
      subBuilder += `${templateIndent(indent + 1)}var ${symbolName} ${
        innerType.content
      }\n`;
      subBuilder += generateSDKToModelMethodBody(
        symbolManager,
        entity,
        symbolName,
        toType.ItemType,
        false,
        false,
        false,
        "",
        iterVariable,
        fromType.ItemType,
        indent + 1,
        `${hierarchy}.[]`,
        operation,
        { isCollectionItem: true },
      );
      subBuilder += `${templateIndent(
        indent + 1,
      )}${targetAccessor}${subAccessor} = append(${targetAccessor}${subAccessor}, ${symbolName})\n`;
      subBuilder += `${templateIndent(indent)}}\n`;
      return subBuilder;
    } else if (isGoArray(toType.ItemType)) {
      let subBuilder = "";
      subBuilder += `${templateIndent(
        indent,
      )}${targetAccessor}${subAccessor} = nil\n`;
      const iterVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "Item",
      );
      subBuilder += `${templateIndent(
        indent,
      )}for _, ${iterVariable} := range ${valueAccessor}${subAccessor} {\n`;
      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        fieldName || associatedName,
        "",
      );
      const innerType = templateTypeDefDataModelFieldType(
        toType.ItemType,
        "",
        false,
        indent,
        "provider",
        `${hierarchy}.[]`,
        (fieldName: string, typedef: TypeDef) => {
          if (typedef.Extensions?.All["Symbol"]) {
            return typedef.Extensions.All["Symbol"];
          }
        },
      );
      subBuilder += `${templateIndent(indent + 1)}var ${symbolName} ${
        innerType.content
      }\n`;
      subBuilder += generateSDKToModelMethodBody(
        symbolManager,
        entity,
        symbolName,
        toType.ItemType,
        false,
        false,
        false,
        "",
        iterVariable,
        fromType.ItemType,
        indent + 1,
        `${hierarchy}.[]`,
        operation,
        { isCollectionItem: true },
      );
      subBuilder += `${templateIndent(
        indent + 1,
      )}${targetAccessor}${subAccessor} = append(${targetAccessor}${subAccessor}, ${symbolName})\n`;
      subBuilder += `${templateIndent(indent)}}`;
      result.push(subBuilder);
    } else {
      // Not Implemented
    }

    if (toIsNullable) {
      indent--;
      result.push(`${templateIndent(indent)}} else {`);
      // When API response has null/absent nullable list, clear the TF state
      // to enable drift detection. Without this, the state would preserve its old value.
      result.push(
        `${templateIndent(indent + 1)}${targetAccessor}${subAccessor} = nil`,
      );
      result.push(`${templateIndent(indent)}}`);
    }

    return result.map((line) => `${line}\n`).join("");
  }
  if (toType.Type == "union" && toType.AssociatedTypes?.length) {
    const associatedTypes = toType.AssociatedTypes;
    let subBuilder = "";
    // subBuilder += `${templateIndent(indent)}..(${hierarchy}, ${fromIsOptional}, ${toIsOptional}, ${subAccessor}, ${targetAccessor})\n`
    // Union fields are pointers (for nil tracking), except collection items which are values
    const needsUnionAllocation =
      toType.Extensions?.All["Symbol"] && !isCollectionItem;

    // Check if any variant actually needs prior data (has write-only fields)
    const unionHasVariantsNeedingPriorData =
      toType.TerraformHasNestedWriteOnlyFields(fromType);

    // Capture union-level prior data when:
    // 1. The union is a pointer (needsUnionAllocation=true)
    // 2. We're not a collection item
    // 3. At least one variant actually needs prior data (has write-only fields)
    // Prior data must be captured BEFORE allocation overwrites it.
    // This applies for both optional and required unions because the model's
    // union pointer can be nil during terraform import (only `id` is in state).
    const shouldCaptureUnionPriorData =
      needsUnionAllocation &&
      !isCollectionItem &&
      unionHasVariantsNeedingPriorData;

    // Generate prior data variable name for the union (to capture before allocation)
    const unionPriorDataAccessor = shouldCaptureUnionPriorData
      ? getPluralizedVarSymbolName(
          symbolManager,
          fieldName || associatedName,
          "PriorData",
        )
      : undefined;

    // Determine the source for prior data capture
    const unionPriorDataCaptureSource = priorDataAccessor
      ? `${priorDataAccessor}${subAccessor}`
      : `${targetAccessor}${subAccessor}`;

    if (fromIsOptional) {
      subBuilder += `${templateIndent(
        indent,
      )}if ${valueAccessor}${subAccessor} != nil {\n`;
      // Capture prior data BEFORE allocation
      if (unionPriorDataAccessor) {
        subBuilder += `${templateIndent(
          indent + 1,
        )}${unionPriorDataAccessor} := ${unionPriorDataCaptureSource}\n`;
      }
      // Then allocate
      if (needsUnionAllocation) {
        subBuilder += `${templateIndent(
          indent + 1,
        )}${targetAccessor}${subAccessor} = &tfTypes.${
          toType.Extensions.All["Symbol"]
        }{}\n`;
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
      }
      indent++;
    } else if (needsUnionAllocation) {
      // For required unions that are pointers: the model's union pointer can
      // be nil during terraform import (only `id` is in state). Capture prior
      // data before allocation, then ensure the pointer is initialized.
      if (unionPriorDataAccessor) {
        subBuilder += `${templateIndent(
          indent,
        )}${unionPriorDataAccessor} := ${unionPriorDataCaptureSource}\n`;
      }
      subBuilder += `${templateIndent(
        indent,
      )}if ${targetAccessor}${subAccessor} == nil {\n`;
      subBuilder += `${templateIndent(
        indent + 1,
      )}${targetAccessor}${subAccessor} = &tfTypes.${
        toType.Extensions.All["Symbol"]
      }{}\n`;
      addGenImport(
        `${getRootPackage()}/internal/provider/types`,
        false,
        "tfTypes",
      );
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    for (let i = 0; i < associatedTypes.length; i++) {
      const subTypeDef = associatedTypes[i];
      const subTypeDefName = sanitizeTFUnionTypeName(toType, subTypeDef);
      let equivalent = fromType.AssociatedTypes?.find(
        (associated) =>
          sanitizeTFUnionTypeName(fromType, associated) == subTypeDefName,
      );
      if (!equivalent) {
        continue;
      }
      const newTarget = `${targetAccessor}${subAccessor}.${subTypeDefName}`;
      const maybeAccessor = `${valueAccessor}${subAccessor}.${sanitizeFieldName(
        equivalent.Name || sanitizeTFUnionTypeName(fromType, equivalent),
      )}`;
      subBuilder += `${templateIndent(indent)}if ${maybeAccessor} != nil {\n`;
      // Allocate variants that won't self-allocate when recursing.
      // Union-type variants always need explicit allocation (their own handler
      // doesn't allocate). Class-type variants only need it when onEntity=true
      // in the recursive call (which sets needsAllocation=false).
      const variantIsUnion =
        subTypeDef.Type == "union" && subTypeDef.AssociatedTypes?.length;
      const variantClassOnEntity =
        subTypeDef.Type == "class" &&
        (subTypeDef.HasEntityName(entity.Name) ||
          subTypeDef.Extensions?.All["x-speakeasy-root"]);
      const variantNeedsAllocation =
        (variantIsUnion || variantClassOnEntity) &&
        subTypeDef.Extensions?.All["Symbol"];
      // Entity-level unions have no union-level prior data capture (the model
      // itself is the union), so variants with write-only fields must capture
      // their own prior data before the allocation below overwrites it.
      const variantPriorDataAccessor =
        variantNeedsAllocation &&
        !unionPriorDataAccessor &&
        !priorDataAccessor &&
        subTypeDef.TerraformHasNestedWriteOnlyFields(equivalent)
          ? getPluralizedVarSymbolName(
              symbolManager,
              subTypeDefName,
              "PriorData",
            )
          : undefined;
      if (variantNeedsAllocation) {
        if (variantPriorDataAccessor) {
          subBuilder += `${templateIndent(
            indent + 1,
          )}${variantPriorDataAccessor} := ${newTarget}\n`;
        }
        subBuilder += `${templateIndent(indent + 1)}${newTarget} = &tfTypes.${
          subTypeDef.Extensions.All["Symbol"]
        }{}\n`;
        addGenImport(
          `${getRootPackage()}/internal/provider/types`,
          false,
          "tfTypes",
        );
      }
      // Use captured prior data if available, otherwise fall back to priorDataAccessor
      const unionVariantpriorDataAccessor = variantPriorDataAccessor
        ? variantPriorDataAccessor
        : unionPriorDataAccessor
        ? `${unionPriorDataAccessor}.${subTypeDefName}`
        : priorDataAccessor
        ? `${priorDataAccessor}${subAccessor}.${subTypeDefName}`
        : undefined;
      // Union members in Go SDK are always pointers for primitive types (e.g., *time.Time, *string).
      // For date/date-time types, we need to use pointer-aware conversion functions.
      // isOptionalType returns false for date-time, so we check explicitly.
      const isPrimitivePointerInSDK =
        equivalent.Type.toString() === "date" ||
        equivalent.Type.toString() === "date-time";
      const unionMemberFromIsOptional =
        isPrimitivePointerInSDK || Boolean(isOptionalType(subTypeDef));
      subBuilder += generateSDKToModelMethodBody(
        symbolManager,
        entity,
        newTarget,
        subTypeDef,
        Boolean(isOptionalType(subTypeDef)),
        Boolean(isOptionalType(subTypeDef)),
        unionMemberFromIsOptional,
        "",
        maybeAccessor,
        equivalent,
        indent,
        calcHierarchy(subTypeDefName),
        operation,
        { priorDataAccessor: unionVariantpriorDataAccessor },
      );
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    if (fromIsOptional) {
      indent--;
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    return subBuilder;
  }
  if (toType.Enum && toType.Enum.Values.length) {
    addGenImport("github.com/hashicorp/terraform-plugin-framework/types");

    let enumResult: string;
    switch (toType.Enum.Type.Type.toString()) {
      case "string":
        if (!fromIsOptional) {
          enumResult = `${templateIndent(
            indent,
          )}${targetAccessor}${subAccessor} = types.StringValue(string(${valueAccessor}${subAccessor}))\n`;
        } else {
          enumResult =
            `${templateIndent(
              indent,
            )}if ${valueAccessor}${subAccessor} != nil {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.StringValue(string(*${valueAccessor}${subAccessor}))\n` +
            `${templateIndent(indent)}} else {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.StringNull()\n` +
            `${templateIndent(indent)}}\n`;
        }
        break;
      case "float32":
        if (!fromIsOptional) {
          enumResult = `${templateIndent(
            indent,
          )}${targetAccessor}${subAccessor} = types.Float32Value(float32(${valueAccessor}${subAccessor}))\n`;
        } else {
          enumResult =
            `${templateIndent(
              indent,
            )}if ${valueAccessor}${subAccessor} != nil {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Float32Value(float32(*${valueAccessor}${subAccessor}))\n` +
            `${templateIndent(indent)}} else {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Float32Null()\n` +
            `${templateIndent(indent)}}\n`;
        }
        break;
      case "int32":
        if (!fromIsOptional) {
          enumResult = `${templateIndent(
            indent,
          )}${targetAccessor}${subAccessor} = types.Int32Value(int32(${valueAccessor}${subAccessor}))\n`;
        } else {
          enumResult =
            `${templateIndent(
              indent,
            )}if ${valueAccessor}${subAccessor} != nil {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Int32Value(int32(*${valueAccessor}${subAccessor}))\n` +
            `${templateIndent(indent)}} else {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Int32Null()\n` +
            `${templateIndent(indent)}}\n`;
        }
        break;
      case "integer":
        if (!fromIsOptional) {
          enumResult = `${templateIndent(
            indent,
          )}${targetAccessor}${subAccessor} = types.Int64Value(int64(${valueAccessor}${subAccessor}))\n`;
        } else {
          enumResult =
            `${templateIndent(
              indent,
            )}if ${valueAccessor}${subAccessor} != nil {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Int64Value(int64(*${valueAccessor}${subAccessor}))\n` +
            `${templateIndent(indent)}} else {\n` +
            `${templateIndent(
              indent + 1,
            )}${targetAccessor}${subAccessor} = types.Int64Null()\n` +
            `${templateIndent(indent)}}\n`;
        }
        break;
      default:
        throw new Error("Unknown enum type: " + toType.Enum.Type.Type);
    }

    if (toType.Extensions?.TerraformAliasTo) {
      const relativeLocation = applyPath(
        `${targetAccessor}${subAccessor}`,
        toType.Extensions.TerraformAliasTo,
      );
      enumResult += `${templateIndent(
        indent,
      )}${relativeLocation} = ${targetAccessor}${subAccessor}\n`;
    }

    return enumResult;
  }
  if (toType.Type == "any") {
    addGenImport("encoding/json");
    addGenImport(
      "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes",
    );

    const tmpSymbol = getPluralizedVarSymbolName(
      symbolManager,
      toType.Name || associatedName,
      "Result",
    );

    const result: string[] = [];
    if (fromIsOptional) {
      result.push(
        `if ${valueAccessor}${subAccessor} == nil {`,
        `${targetAccessor}${subAccessor} = jsontypes.NewNormalizedNull()`,
        `} else {`,
        `${tmpSymbol}, _ := json.Marshal(${valueAccessor}${subAccessor})`,
        `${targetAccessor}${subAccessor} = jsontypes.NewNormalizedValue(string(${tmpSymbol}))`,
        `}`,
      );
    } else {
      result.push(
        `${tmpSymbol}, _ := json.Marshal(${valueAccessor}${subAccessor})`,
        `${targetAccessor}${subAccessor} = jsontypes.NewNormalizedValue(string(${tmpSymbol}))`,
      );
    }
    let anyResult = result.map((line) => `${line}\n`).join("");

    if (toType.Extensions?.TerraformAliasTo) {
      const relativeLocation = applyPath(
        `${targetAccessor}${subAccessor}`,
        toType.Extensions.TerraformAliasTo,
      );
      anyResult += `${templateIndent(
        indent,
      )}${relativeLocation} = ${targetAccessor}${subAccessor}\n`;
    }

    return anyResult;
  }

  throw new Error(
    `\n\nUnsupported Type. To continue, apply escape hatches [x-speakeasy-terraform-ignore: true or x-speakeasy-type-override: any] to ${hierarchy}\n\n${JSON.stringify(
      toType,
    )}`,
  );
}
