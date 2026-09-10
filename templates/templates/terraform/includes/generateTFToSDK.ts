function primitiveAccessor(
  def: TypeDef,
  pointer: boolean,
):
  | "ValueString()"
  | "ValueStringPointer()"
  | "ValueFloat32()"
  | "ValueFloat32Pointer()"
  | "ValueInt32()"
  | "ValueInt32Pointer()"
  | "ValueInt64()"
  | "ValueInt64Pointer()"
  | "ValueFloat64()"
  | "ValueFloat64Pointer()"
  | "ValueBool()"
  | "ValueBoolPointer()"
  | undefined {
  switch (String(def.Type)) {
    case "string":
      return `ValueString${pointer ? "Pointer" : ""}()`;
    case "float32":
      return `ValueFloat32${pointer ? "Pointer" : ""}()`;
    case "int32":
      return `ValueInt32${pointer ? "Pointer" : ""}()`;
    case "integer":
      return `ValueInt64${pointer ? "Pointer" : ""}()`;
    case "number":
      return `ValueFloat64${pointer ? "Pointer" : ""}()`;
    case "boolean":
      return `ValueBool${pointer ? "Pointer" : ""}()`;
    default:
      return undefined;
  }
}
function primitiveValuePrefix(
  def: TypeDef,
  pointer: boolean,
): string | undefined {
  switch (String(def.Type)) {
    case "string":
      return `types.String${pointer ? "Pointer" : ""}`;
    case "float32":
      return `types.Float32${pointer ? "Pointer" : ""}`;
    case "int32":
      return `types.Int32${pointer ? "Pointer" : ""}`;
    case "integer":
      return `types.Int64${pointer ? "Pointer" : ""}`;
    case "number":
      return `types.Float64${pointer ? "Pointer" : ""}`;
    case "boolean":
      return `types.Bool${pointer ? "Pointer" : ""}`;
    default:
      return undefined;
  }
}
function primitiveWrap(def: TypeDef): string | undefined {
  switch (String(def.Type)) {
    case "string":
      return "string";
    case "int32":
      return "int32";
    case "integer":
      return "int64";
    default:
      return undefined;
  }
}

function isGoArray(targetTypeDef: TypeDef) {
  return targetTypeDef.Type == "array" || targetTypeDef.Type == "set";
}

function findMatch(
  parentTypeDef: TypeDef,
  match: string,
): { accessor?: string[]; field?: FieldDef; error?: string } {
  const steps = match.split(".");
  let accessorSteps: string[] = [];
  let error = undefined;
  let fieldDef: FieldDef = undefined;
  for (const step of steps) {
    const matchField = parentTypeDef.Fields.find(
      (field) => sanitizeFieldName(field.Name) == sanitizeFieldName(step),
    );
    if (!matchField) {
      const allAttributes = new Set(
        parentTypeDef.Fields.filter((f) => !f.Type.Extensions?.MatchConfig).map(
          (f) => sanitizeTFStateName(f.Name),
        ),
      );
      error = `${JSON.stringify(match)} not in ${JSON.stringify(
        Array.from(allAttributes).join(","),
      )}. Any x-speakeasy-match used during update must also be used during creation.`;
      return { error };
    }
    accessorSteps.push(sanitizeFieldName(matchField.Name));
    fieldDef = matchField;
    parentTypeDef = matchField.Type;
  }
  if (!fieldDef) {
    error = `x-speakeasy-match: ${match} not found`;
    return { error };
  }
  return { accessor: accessorSteps, field: fieldDef };
}

function isAggregateType(field: TypeDef) {
  return (
    field.Type == "map" ||
    field.Type == "class" ||
    field.Type == "union" ||
    isGoArray(field)
  );
}

/** Helper to build patch semantics condition for primitive attributes */
function stateInequalityCheck(
  tfAccessor: string,
  hasPatchSemantics: boolean,
  stateNonNilGuard?: string,
): string {
  if (!hasPatchSemantics) return "";
  const tfAccessorParts = tfAccessor.split(".");
  tfAccessorParts[0] = "opts.State";
  const stateAccessor = tfAccessorParts.join(".");
  const guardClause = stateNonNilGuard ? `!(${stateNonNilGuard}) || ` : "";
  return `&& (opts == nil || opts.State == nil || ${guardClause}(!${stateAccessor}.Equal(${tfAccessor})))`;
}

/**
 * Helper to match union variants between target and source unions.
 * First we try to use the discriminator mapping name, if applicable.
 * If no match is found, we fallback to comparing model names.
 */
function findMatchingSourceUnionAssociatedType(
  sourceUnionType: TypeDef, // the source type containing variants to search
  targetUnionType: TypeDef, // the target type (used for discriminator lookup)
  targetAssociatedType: TypeDef, // the target variant to find a matching source for
): TypeDef | undefined {
  const sourceAssociatedTypes = sourceUnionType.AssociatedTypes;
  if (!sourceAssociatedTypes?.length) {
    return undefined;
  }

  // If it exists, prefer discriminator mapping name.
  const mapping = targetUnionType?.Discriminator?.Mapping.find(
    (mapping) =>
      targetAssociatedType.Name == mapping.Type.Name ||
      (targetAssociatedType.OriginalName &&
        targetAssociatedType.OriginalName == mapping.Type.OriginalName),
  );
  if (mapping) {
    const discriminatorKeyName = sanitizeClassName(
      getDiscriminatorDisplayName(mapping),
    );
    const match = sourceAssociatedTypes.find(
      (typ) =>
        sanitizeTFUnionTypeName(sourceUnionType, typ) === discriminatorKeyName,
    );
    if (match) {
      return match;
    }
  }

  // Falling back to comparing model names
  return sourceAssociatedTypes.find(
    (typ) =>
      sanitizeTFUnionTypeName(undefined, typ) ===
      sanitizeTFUnionTypeName(undefined, targetAssociatedType),
  );
}

/** Templates an entity data model to Go SDK request type method body. */
function generateModelToSDKMethodBody(
  symbolManager: Record<string, boolean>,
  targetVar: string,
  targetTypeDef: TypeDef,
  fromTypeDef: TypeDef,
  targetIsNullable: boolean,
  targetIsOptional: boolean,
  fromIsOptional: boolean,
  scope: Scope,
  withinType: boolean,
  fieldName: string,
  sdkFieldName: string,
  accessor: string,
  entity: TerraformEntity,
  indent: number,
  hierarchy: string,
  operation: TerraformOperation,
  priorMatch: boolean = false,
  changedVar?: string,
  stateNonNilGuard?: string,
  isAdditionalProperties: boolean = false,
  declarationHoisted: boolean = false,
) {
  const innerSymbols: Record<string, { Name: string; Optional: boolean }> = {};

  if (targetTypeDef == undefined || fromTypeDef == undefined) {
    throw new Error(
      `\n\nUnexpected type detected in Entity "${
        entity.Name
      }" ${hierarchy}\n${JSON.stringify(targetTypeDef)}\n${JSON.stringify(
        fromTypeDef,
      )}\n`,
    );
  }
  try {
    if (targetTypeDef.IsTerraformPrimitiveType()) {
      handleTypeMismatch(hierarchy, targetTypeDef, fromTypeDef);
    }
  } catch (e) {
    let subBuilder = `${templateIndent(indent)}// ${hierarchy}${`${
      e instanceof Error ? e.message : e
    }`.replace("\n", "")}\n`;

    if (!declarationHoisted) {
      subBuilder += `${templateIndent(indent)}var ${targetVar} *${templateType(
        targetTypeDef,
        { scope: "tf" },
      )}\n`;
    }
    return subBuilder;
  }

  function getWithinType() {
    return (
      withinType ||
      (entity?.SchemaTypeDef?.HasEntityName(entity.Name) &&
        targetTypeDef.HasEntityName(entity.Name))
    );
  }

  let tfAccessor = fieldName ? `${accessor}.${fieldName}` : accessor;

  // Calculate patch semantics condition early so it can be used for all types
  const hasPatchSemantics =
    entity.IncludeSDKMethodOptions &&
    operation.Options?.Patch?.Style == "only-send-changed-attributes";

  // Handle path aliasing via x-speakeasy-match
  // (usePriorState without path falls through to normal processing below)
  const matchPath = targetTypeDef.Extensions?.MatchConfig?.Path;
  if (matchPath && !priorMatch) {
    const { field, accessor: newAccessor } = findMatch(
      entity.SchemaTypeDef,
      matchPath,
    );
    const lastField = newAccessor.pop();
    const newHierarchy =
      hierarchy + `=${accessor}>${sanitizeTFStateName(field.Name)}`;
    // debug(`${hierarchy}=${accessor}>${sanitizeTFStateName(field.Name)}`, targetVar, targetTypeDef, fromTypeDef, field.Type, targetIsOptional, field.Optional || field.Nullable, scope, withinType, "", sdkFieldName, [accessor, newAccessor].join("."), entity, indent)
    // When the entity field is an enum, resolve to its underlying primitive type.
    // In TF state, enum values are stored as their base type (e.g., string).
    // Note: String() wrapper required for Go DataType comparison in Goja runtime.
    const resolvedFieldType =
      String(field.Type.Type) === "enum" && field.Type.Enum?.Type
        ? field.Type.Enum.Type
        : field.Type;

    // Build nil guards for intermediate objects in nested match paths.
    // For a match path like "safety_settings.enforce_approval",
    // newAccessor contains intermediate steps (e.g. ["SafetySettings"])
    // that may be nil pointers on the TF model. Each needs a nil check.
    const matchNilGuards: string[] = [];
    for (let i = 0; i < newAccessor.length; i++) {
      const intermediatePath = [accessor, ...newAccessor.slice(0, i + 1)].join(
        ".",
      );
      matchNilGuards.push(`${intermediatePath} != nil`);
    }

    const innerBody = generateModelToSDKMethodBody(
      symbolManager,
      targetVar,
      targetTypeDef,
      resolvedFieldType,
      targetIsNullable,
      targetIsOptional,
      field.Optional || field.Nullable,
      scope,
      withinType,
      lastField,
      sdkFieldName,
      [accessor, ...newAccessor].join("."),
      entity,
      indent + matchNilGuards.length,
      newHierarchy,
      operation,
      true,
      changedVar,
      stateNonNilGuard,
      false,
      matchNilGuards.length > 0,
    );

    if (matchNilGuards.length === 0) {
      return innerBody;
    }

    // Hoist variable declaration outside the nil guard to avoid Go scoping
    // issues. A "var x string" declared inside "if r.Parent != nil { }" is
    // inaccessible to code after the if block.
    const goType = templateType(targetTypeDef, { scope: "tf" });
    const isPointerType = targetIsNullable || targetIsOptional;
    const declType = isPointerType ? `*${goType}` : goType;
    const hoistedDecl = `${templateIndent(
      indent,
    )}var ${targetVar} ${declType}\n`;

    // Remove or convert the declaration from innerBody to avoid duplication.
    let guardedBody = innerBody;
    const varDeclMarker = `var ${targetVar} `;

    if (guardedBody.includes(varDeclMarker)) {
      // Pattern: "var X type\nX = ..." — strip the var declaration line
      const startIdx = guardedBody.indexOf(varDeclMarker);
      const lineStart = guardedBody.lastIndexOf("\n", startIdx - 1) + 1;
      const lineEnd = guardedBody.indexOf("\n", startIdx);
      guardedBody =
        guardedBody.substring(0, lineStart) +
        guardedBody.substring(lineEnd + 1);
    } else {
      // Pattern: "X := expr" or "X, _ := expr" — convert to assignment
      guardedBody = guardedBody
        .replace(`${targetVar}, _ := `, `${targetVar}, _ = `)
        .replace(`${targetVar} := `, `${targetVar} = `);
    }

    let result = hoistedDecl;
    for (let i = 0; i < matchNilGuards.length; i++) {
      result += `${templateIndent(indent + i)}if ${matchNilGuards[i]} {\n`;
    }
    result += guardedBody;
    for (let i = matchNilGuards.length - 1; i >= 0; i--) {
      result += `${templateIndent(indent + i)}}\n`;
    }
    return result;
  }

  if (primitiveAccessor(targetTypeDef, false)) {
    const isCreateOp = operation.EntityOperation?.includes("#create");
    const isUpdateOp = operation.EntityOperation?.includes("#update");
    const useConfig =
      targetTypeDef.Extensions?.TerraformWriteOnly &&
      (isCreateOp || isUpdateOp);
    const usePriorState =
      targetTypeDef.Extensions?.MatchConfig?.UsePriorState && isUpdateOp;

    let sourceAccessor = tfAccessor;

    if (entity.IncludeSDKMethodOptions) {
      if (useConfig) {
        const tfAccessorParts = tfAccessor.split(".");
        tfAccessorParts[0] = "opts.Config";
        sourceAccessor = tfAccessorParts.join(".");
      } else if (usePriorState) {
        const tfAccessorParts = tfAccessor.split(".");
        tfAccessorParts[0] = "opts.State";
        sourceAccessor = tfAccessorParts.join(".");
      }
    }

    let accessor = `${sourceAccessor}.${primitiveAccessor(fromTypeDef, false)}`;
    if (targetTypeDef.Type == "int32") {
      accessor = `int(${accessor})`;
    }
    if (!targetIsNullable && !targetIsOptional) {
      if (declarationHoisted) {
        return `${templateIndent(indent)}${targetVar} = ${accessor}\n`;
      }
      return (
        `${templateIndent(indent)}var ${targetVar} ${templateType(
          targetTypeDef,
          { scope: "tf" },
        )}\n` + `${targetVar} = ${accessor}\n\n`
      );
    }

    const patchSemantics = stateInequalityCheck(
      tfAccessor,
      hasPatchSemantics,
      stateNonNilGuard,
    );

    const assignOp = declarationHoisted ? "=" : ":=";
    return (
      `${templateIndent(indent)}${targetVar} ${assignOp} new(${templateType(
        targetTypeDef,
        { scope: "tf" },
      )})\n` +
      `${templateIndent(indent)}if !${
        useConfig ? sourceAccessor : tfAccessor
      }.IsUnknown() && !${
        useConfig ? sourceAccessor : tfAccessor
      }.IsNull()${patchSemantics} {\n` +
      `${templateIndent(indent + 1)}*${targetVar} = ${accessor}\n` +
      (changedVar
        ? `${templateIndent(indent + 1)}${changedVar} = true\n`
        : "") +
      `${templateIndent(indent)}} else {\n` +
      `${targetVar} = nil\n` +
      `${templateIndent(indent)}}\n`
    );
  }
  if (targetTypeDef.Type == "bytes") {
    // []byte is nil-able without being a pointer type in Go. Both required
    // and optional fields use []byte (not *[]byte). ValueBytes() returns
    // ([]byte, diag.Diagnostics) — nil when null/unknown.
    const bytesVarName = getPluralizedVarSymbolName(
      symbolManager,
      sdkFieldName || fieldName,
      "Bytes",
    );
    const diagsVarName = getPluralizedVarSymbolName(
      symbolManager,
      sdkFieldName || fieldName,
      "Diags",
    );

    if (!targetIsNullable && !targetIsOptional) {
      const bytesAssignOp = declarationHoisted ? "=" : ":=";
      return (
        `${templateIndent(
          indent,
        )}${bytesVarName}, ${diagsVarName} := ${tfAccessor}.ValueBytes()\n` +
        `${templateIndent(indent)}diags.Append(${diagsVarName}...)\n` +
        `${templateIndent(
          indent,
        )}${targetVar} ${bytesAssignOp} ${bytesVarName}\n`
      );
    }

    const patchSemantics = stateInequalityCheck(
      tfAccessor,
      hasPatchSemantics,
      stateNonNilGuard,
    );
    return (
      `${templateIndent(indent)}var ${targetVar} ${templateType(targetTypeDef, {
        scope: "tf",
      })}\n` +
      `${templateIndent(
        indent,
      )}if !${tfAccessor}.IsUnknown() && !${tfAccessor}.IsNull()${patchSemantics} {\n` +
      `${templateIndent(
        indent + 1,
      )}${bytesVarName}, ${diagsVarName} := ${tfAccessor}.ValueBytes()\n` +
      `${templateIndent(indent + 1)}diags.Append(${diagsVarName}...)\n` +
      `${templateIndent(indent + 1)}${targetVar} = ${bytesVarName}\n` +
      (changedVar
        ? `${templateIndent(indent + 1)}${changedVar} = true\n`
        : "") +
      `${templateIndent(indent)}}\n`
    );
  }
  if (targetTypeDef.Type == "date-time") {
    addGenImport("time");
    if (!targetIsNullable && !targetIsOptional) {
      const timeAssignOp = declarationHoisted ? "=" : ":=";
      return `${templateIndent(
        indent,
      )}${targetVar}, _ ${timeAssignOp} time.Parse(time.RFC3339Nano, ${tfAccessor}.ValueString())\n`;
    }
    const patchSemantics = stateInequalityCheck(
      tfAccessor,
      hasPatchSemantics,
      stateNonNilGuard,
    );
    const dtAssignOp = declarationHoisted ? "=" : ":=";
    return (
      `${templateIndent(indent)}${targetVar} ${dtAssignOp} new(${templateType(
        targetTypeDef,
        { scope: "tf" },
      )})\n` +
      `${templateIndent(
        indent,
      )}if !${tfAccessor}.IsUnknown() && !${tfAccessor}.IsNull()${patchSemantics} {\n` +
      `${templateIndent(
        indent + 1,
      )}*${targetVar}, _ = time.Parse(time.RFC3339Nano, ${tfAccessor}.ValueString())\n` +
      (changedVar
        ? `${templateIndent(indent + 1)}${changedVar} = true\n`
        : "") +
      `${templateIndent(indent)}} else {\n` +
      `${templateIndent(indent + 1)}${targetVar} = nil\n` +
      `${templateIndent(indent)}}\n`
    );
  }
  if (targetTypeDef.Type == "date") {
    addGenImport(getSDKPackage() + "/types", false, "customTypes");
    if (!targetIsNullable && !targetIsOptional) {
      const dateAssignOp = declarationHoisted ? "=" : ":=";
      return `${templateIndent(
        indent,
      )}${targetVar} ${dateAssignOp} customTypes.MustDateFromString(${tfAccessor}.ValueString())\n`;
    }
    const patchSemantics = stateInequalityCheck(
      tfAccessor,
      hasPatchSemantics,
      stateNonNilGuard,
    );
    const dateNullAssignOp = declarationHoisted ? "=" : ":=";
    return (
      `${templateIndent(
        indent,
      )}${targetVar} ${dateNullAssignOp} new(customTypes.Date)\n` +
      `${templateIndent(
        indent,
      )}if !${tfAccessor}.IsUnknown() && !${tfAccessor}.IsNull()${patchSemantics} {\n` +
      `${templateIndent(
        indent + 1,
      )}${targetVar} = customTypes.MustNewDateFromString(${tfAccessor}.ValueString())\n` +
      (changedVar
        ? `${templateIndent(indent + 1)}${changedVar} = true\n`
        : "") +
      `${templateIndent(indent)}} else {\n` +
      `${templateIndent(indent + 1)}${targetVar} = nil\n` +
      `${templateIndent(indent)}}\n`
    );
  }

  if (targetTypeDef.Type == "map") {
    if (targetTypeDef.Scope && String(targetTypeDef.Scope).length) {
      addGenImport(`${getSDKPackage()}/models/${targetTypeDef.Scope}`);
    }

    const customTypeConfig = TypeDefTerraformCustomTypeConfig(fromTypeDef);

    if (customTypeConfig) {
      const frameworkType = FrameworkTypeFromTypeDef(fromTypeDef);

      return (
        frameworkType.templateTerraformToSDK(
          symbolManager,
          fieldName,
          targetTypeDef,
          false,
          targetIsNullable || targetIsOptional,
          targetVar,
          tfAccessor,
          false,
        ) + "\n"
      );
    }

    let subBuilder = "";

    if (targetIsNullable) {
      subBuilder += `var ${targetVar} ${templateType(targetTypeDef, {
        scope: "tf",
      })}\n`;
      subBuilder += `if ${tfAccessor} != nil {\n`;
    }

    const assignmentOperator = targetIsNullable ? "=" : ":=";

    subBuilder += `${targetVar} ${assignmentOperator} make(${templateType(
      targetTypeDef,
      {
        scope: "tf",
      },
    )})\n`;

    if (fromTypeDef.Type != "map") {
      return (
        subBuilder +
        `${templateIndent(
          indent,
        )}// Warning. This is a map, but the source tf var is a ${
          fromTypeDef.Type
        }. This might indicate a bug.\n`
      );
    }
    const innerSymbol = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Inst",
    );
    const keySymbol = getPluralizedVarSymbolName(
      symbolManager,
      fieldName,
      "Key",
    );
    subBuilder += `${templateIndent(
      indent,
    )}for ${keySymbol} := range ${tfAccessor} {\n`;
    tfAccessor = `${tfAccessor}[${keySymbol}]`;
    // subBuilder += `${templateIndent(indent)}// Generate ${hierarchy} target=${JSON.stringify(targetTypeDef)}\n`
    // subBuilder += `${templateIndent(indent)}// Generate ${hierarchy} from=${JSON.stringify(fromTypeDef)}\n`

    subBuilder += generateModelToSDKMethodBody(
      symbolManager,
      innerSymbol,
      targetTypeDef.ItemType,
      fromTypeDef.ItemType,
      targetTypeDef.ContainsNull,
      false,
      false,
      targetTypeDef.ItemType.Scope,
      getWithinType(),
      "",
      "",
      tfAccessor,
      entity,
      indent + 1,
      hierarchy + ".[]",
      operation,
      false,
      changedVar,
      stateNonNilGuard,
    );
    subBuilder += `${templateIndent(
      indent + 1,
    )}${targetVar}[${keySymbol}] = ${innerSymbol}\n`;
    subBuilder += `${templateIndent(indent)}}\n`;

    if (targetIsNullable) {
      subBuilder += `}\n`;
    }

    return subBuilder;
  }
  if (targetTypeDef.Type == "class") {
    if (targetTypeDef.Scope && String(targetTypeDef.Scope).length) {
      addGenImport(`${getSDKPackage()}/models/${targetTypeDef.Scope}`);
    }
    const stateAccessor = tfAccessor;
    let subBuilder = "";
    const childNonZeroChecks: Record<string, FieldDef> = {};
    const onRoot = Boolean(fromTypeDef?.Extensions?.All["x-speakeasy-root"]);
    const pagination = operation.APIOperation.Extensions?.Pagination;

    if (
      targetTypeDef.HasEntityName(entity.Name) &&
      // recursion prevention
      !onRoot &&
      fromTypeDef.Scope.toString() !== "shared"
    ) {
      const dataToSDKMethodName = sanitizeClassName(
        `To_` + templateType(targetTypeDef),
      );
      const diagsVariable = getPluralizedVarSymbolName(
        symbolManager,
        targetVar,
        "Diags",
      );
      let extraArgument = "";
      if (entity.IncludeSDKMethodOptions) {
        extraArgument = `, opts`;
      }
      if (!targetIsNullable && !targetIsOptional) {
        const pointerVariable = getPluralizedVarSymbolName(
          symbolManager,
          targetVar,
          "Ptr",
        );

        return (
          `${pointerVariable}, ${diagsVariable} := ${accessor}.${dataToSDKMethodName}(ctx${extraArgument})\n` +
          `diags.Append(${diagsVariable}...)\n\n` +
          `if diags.HasError() {\n` +
          `return nil, diags\n` +
          `}\n\n` +
          `${targetVar} := *${pointerVariable}\n`
        );
      }

      return (
        `${targetVar}, ${diagsVariable} := ${accessor}.${dataToSDKMethodName}(ctx${extraArgument})\n` +
        `diags.Append(${diagsVariable}...)\n\n` +
        `if diags.HasError() {\n` +
        `return nil, diags\n` +
        `}\n\n`
      );
    }

    if (targetIsNullable || targetIsOptional) {
      subBuilder += `${templateIndent(indent)}var ${targetVar} *${templateType(
        targetTypeDef,
        { scope: "tf" },
      )}\n`;
      // Not if we're on the root -- that's an object type
      if (!onRoot && fromIsOptional) {
        subBuilder += `${templateIndent(indent)}if ${stateAccessor} != nil {\n`;
        indent++;
      }
    }

    for (let i = 0; i < targetTypeDef.Fields.length; i++) {
      const field = targetTypeDef.Fields[i];
      if (field.Type.Extensions?.Ignore) {
        continue;
      }
      if (field.Const) {
        continue;
      }

      if (
        onRoot &&
        pagination?.Inputs.find((input) => {
          return input.Name === field.Name;
        })
      ) {
        continue;
      }

      const [equivalentField] = field.FindTerraformEquivalentField(
        fromTypeDef?.Fields ?? [],
        false,
      );
      const equivalent = equivalentField
        ? { field: equivalentField }
        : undefined;

      if (!equivalent) {
        // Ensure all fields on the path to the entity TypeDef are templated.
        // Reference: internal issue reference
        if (field.Type?.FindEntityTypeDef(entity.Name)) {
          const symbolName = getPluralizedVarSymbolName(
            symbolManager,
            sanitizeVariableName(field.Name),
            "",
          );

          if (field.Type.ItemType) {
            innerSymbols[
              `${templateType(field.Type)}{${
                field.Type.ContainsNull ? "" : "*"
              }${symbolName}}`
            ] = field;
          } else {
            innerSymbols[
              `${field.Optional || field.Nullable ? "" : "*"}${symbolName}`
            ] = field;
          }

          const dataToSDKMethodName = sanitizeClassName(
            `To_` + templateType(field.Type),
          );
          let extraArgument = "";
          if (entity.IncludeSDKMethodOptions) {
            extraArgument = `, opts`;
          }
          const diagsVariable = getPluralizedVarSymbolName(
            symbolManager,
            symbolName,
            "Diags",
          );

          subBuilder +=
            `${symbolName}, ${diagsVariable} := ${accessor}.${dataToSDKMethodName}(ctx${extraArgument})\n` +
            `diags.Append(${diagsVariable}...)\n\n` +
            `if diags.HasError() {\n` +
            `return nil, diags\n` +
            `}\n\n`;
        }

        continue;
      }

      const equivalentType = equivalent.field.Type;
      const equivalentName = sanitizeFieldName(equivalent.field.Name);
      const equivalentOptionalNullable =
        equivalent.field.Optional || equivalent.field.Nullable;

      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        sanitizeVariableName(field.Name),
        "",
      );
      innerSymbols[symbolName] = field;
      subBuilder += generateModelToSDKMethodBody(
        symbolManager,
        symbolName,
        field.Type,
        equivalentType,
        field.Nullable,
        field.Optional || field.Nullable,
        equivalentOptionalNullable,
        targetTypeDef.Scope,
        getWithinType(),
        equivalentName,
        field.Name,
        stateAccessor,
        entity,
        indent,
        hierarchy + `.${sanitizeTFStateName(field.Name)}`,
        operation,
        false,
        changedVar,
        stateNonNilGuard,
        field.IsAdditionalProperties,
      );
    }

    if (
      (targetIsNullable || targetIsOptional) &&
      Object.keys(childNonZeroChecks).length
    ) {
      const condition = [];
      for (const [symbolName, fieldDef] of Object.entries(childNonZeroChecks)) {
        condition.push(
          `${symbolName} != ${templateZero(
            fieldDef.Type,
            fieldDef.Optional,
            scope,
          )}`,
        );
      }
      subBuilder += `${templateIndent(indent)}if ${condition.join(" && ")} {\n`;
      indent++;
    }

    if (targetIsNullable || targetIsOptional) {
      subBuilder += `${templateIndent(indent)}${targetVar} = &${templateType(
        targetTypeDef,
        { scope: "tf" },
      )}{\n`;
    } else {
      subBuilder += `${templateIndent(indent)}${targetVar} := ${templateType(
        targetTypeDef,
        { scope: "tf" },
      )}{\n`;
    }

    for (const symbolName of Object.keys(innerSymbols)) {
      const field = innerSymbols[symbolName];
      subBuilder += `${templateIndent(indent + 1)}${sanitizeFieldName(
        field.Name,
      )}: ${symbolName},\n`;
    }

    subBuilder += templateIndent(indent) + "}\n";
    if ((targetIsNullable || targetIsOptional) && !onRoot && fromIsOptional) {
      indent--;
      subBuilder += `${templateIndent(indent)}}\n`;
    }
    if (
      (targetIsNullable || targetIsOptional) &&
      Object.keys(childNonZeroChecks).length
    ) {
      indent--;
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    // subBuilder += `${templateIndent(indent)}// End Generate(${optional}, ${JSON.stringify(targetTypeDef.Name)}, ${accessor}, ${withinType})\n`
    return subBuilder;
  }
  if (isGoArray(targetTypeDef)) {
    if (!isGoArray(fromTypeDef) && withinType) {
      throw new Error(
        `\nUnexpected type detected in Entity "${
          entity.Name
        }" ${hierarchy}\n\nThe terraform attribute is ${JSON.stringify(
          targetTypeDef.Type,
        )} but the API request type is ${JSON.stringify(fromTypeDef.Type)}\n` +
          `\nThis is a fatal error.\n\nWorkarounds are:\n using "x-speakeasy-name-override: NewAttributeName" to rename the request or response types\n using "x-speakeasy-ignore: true" to ignore the response attribute\n`,
      );
    }
    if (
      targetTypeDef.ItemType.Scope &&
      String(targetTypeDef.ItemType.Scope).length
    ) {
      addGenImport(`${getSDKPackage()}/models/${targetTypeDef.ItemType.Scope}`);
    }

    const customTypeConfig = TypeDefTerraformCustomTypeConfig(fromTypeDef);

    if (customTypeConfig) {
      const frameworkType = FrameworkTypeFromTypeDef(fromTypeDef);

      frameworkType
        .templateTerraformToSDKImports(
          targetTypeDef,
          false,
          targetIsNullable || targetIsOptional,
        )
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      return (
        frameworkType.templateTerraformToSDK(
          symbolManager,
          fieldName,
          targetTypeDef,
          false,
          targetIsNullable || targetIsOptional,
          targetVar,
          tfAccessor,
          false,
        ) + "\n"
      );
    }

    if (generatePrimitiveType(targetTypeDef.ItemType)) {
      let subBuilder = "";

      // Build patch semantics condition for lists
      // Only apply patch semantics to optional/nullable lists (required lists should always be sent)
      let listPatchCondition = "";
      if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
        const tfAccessorParts = tfAccessor.split(".");
        tfAccessorParts[0] = "opts.State";
        const stateAccessor = tfAccessorParts.join(".");
        // For lists, we check if opts is nil or if the list has changed (different lengths or different elements)
        const listChangedVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Changed",
        );
        subBuilder += `${templateIndent(
          indent,
        )}${listChangedVar} := opts == nil || opts.State == nil || len(${stateAccessor}) != len(${tfAccessor})\n`;
        subBuilder += `${templateIndent(indent)}if !${listChangedVar} {\n`;
        const checkIterVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "CheckIdx",
        );
        subBuilder += `${templateIndent(
          indent + 1,
        )}for ${checkIterVar} := range ${tfAccessor} {\n`;
        subBuilder += `${templateIndent(
          indent + 2,
        )}if !${stateAccessor}[${checkIterVar}].Equal(${tfAccessor}[${checkIterVar}]) {\n`;
        subBuilder += `${templateIndent(indent + 3)}${listChangedVar} = true\n`;
        subBuilder += `${templateIndent(indent + 3)}break\n`;
        subBuilder += `${templateIndent(indent + 2)}}\n`;
        subBuilder += `${templateIndent(indent + 1)}}\n`;
        subBuilder += `${templateIndent(indent)}}\n`;
        listPatchCondition = listChangedVar;
        // Set the changed accumulator if the list has changed
        if (changedVar) {
          subBuilder += `${templateIndent(indent)}if ${listChangedVar} {\n`;
          subBuilder += `${templateIndent(indent + 1)}${changedVar} = true\n`;
          subBuilder += `${templateIndent(indent)}}\n`;
        }
      }

      if (targetIsNullable) {
        subBuilder += `var ${targetVar} ${templateType(targetTypeDef, {
          scope: "tf",
        })}\n`;
        const conditions = [`${tfAccessor} != nil`];
        if (listPatchCondition) {
          conditions.push(listPatchCondition);
        }
        subBuilder += `if ${conditions.join(" && ")} {\n`;
      } else if (listPatchCondition) {
        subBuilder += `var ${targetVar} ${templateType(targetTypeDef, {
          scope: "tf",
        })}\n`;
        subBuilder += `if ${listPatchCondition} {\n`;
      }

      subBuilder += `${targetVar} ${
        targetIsNullable || listPatchCondition ? "=" : ":="
      } make(${templateType(targetTypeDef, {
        scope: "tf",
      })}, 0, len(${tfAccessor}))\n`;

      {
        const iterVariable = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Index",
        );
        subBuilder += `for ${iterVariable} := range ${tfAccessor} {\n`;
        tfAccessor = `${tfAccessor}[${iterVariable}]`;
      }
      if (targetTypeDef.ItemType.Type.toString() === "int32") {
        if (targetTypeDef.ContainsNull) {
          const tmpVariable = getPluralizedVarSymbolName(
            symbolManager,
            fieldName,
            "Tmp",
          );
          subBuilder += `if ${tfAccessor}.IsNull() {\n`;
          subBuilder += `${targetVar} = append(${targetVar}, nil)\n`;
          subBuilder += `} else {\n`;
          subBuilder += `${tmpVariable} := int(${tfAccessor}.${primitiveAccessor(
            targetTypeDef.ItemType,
            false,
          )})\n`;
          subBuilder += `${targetVar} = append(${targetVar}, &${tmpVariable})\n`;
          subBuilder += `}\n`;
        } else {
          subBuilder += `${targetVar} = append(${targetVar}, int(${tfAccessor}.${primitiveAccessor(
            targetTypeDef.ItemType,
            targetTypeDef.ContainsNull,
          )}))\n`;
        }
      } else {
        subBuilder += `${targetVar} = append(${targetVar}, ${tfAccessor}.${primitiveAccessor(
          targetTypeDef.ItemType,
          targetTypeDef.ContainsNull,
        )})\n`;
      }
      subBuilder += `}\n`;

      if (targetIsNullable || listPatchCondition) {
        subBuilder += `${templateIndent(indent)}}\n`;
      }

      return subBuilder;
    } else if (targetTypeDef.ItemType.Type.toString() === "enum") {
      const frameworkType = FrameworkTypeFromTypeDef(fromTypeDef);

      frameworkType
        .templateTerraformToSDKImports(targetTypeDef, false, targetIsNullable)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      return (
        frameworkType.templateTerraformToSDK(
          symbolManager,
          fieldName,
          targetTypeDef,
          false,
          targetIsNullable,
          targetVar,
          tfAccessor,
          true,
        ) + "\n"
      );
    } else if (targetTypeDef.ItemType.Type.toString() === "class") {
      if (!targetTypeDef.ItemType.Fields.length) {
        return `${targetVar} := make(${templateType(targetTypeDef, {
          scope: "tf",
        })}, len(${tfAccessor}))\n`;
      }

      if (!withinType) {
        let subBuilder = "";
        // subBuilder += `${templateIndent(
        //   indent,
        // )}// ${hierarchy} from=${JSON.stringify(fromTypeDef)}\n`;
        const singletonVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Singleton",
        );
        // subBuilder += `${templateIndent(
        //   indent,
        // )}// fun case -- we make a singleton array\n`;

        subBuilder += generateModelToSDKMethodBody(
          symbolManager,
          singletonVar,
          targetTypeDef.ItemType,
          fromTypeDef,
          false,
          false,
          fromIsOptional,
          targetTypeDef.Scope,
          getWithinType(),
          fieldName,
          sdkFieldName,
          accessor,
          entity,
          indent,
          hierarchy + `.[]`,
          operation,
          false,
          changedVar,
          stateNonNilGuard,
        );

        subBuilder += `${templateIndent(indent)}${targetVar} := ${templateType(
          targetTypeDef,
          { scope: "tf" },
        )}{${singletonVar}}\n`;

        return subBuilder;
      }

      let subBuilder = "";
      // subBuilder += `${templateIndent(indent)}// ${hierarchy} target=${JSON.stringify(targetTypeDef)}\n`
      // subBuilder += `${templateIndent(indent)}// ${hierarchy} from=${JSON.stringify(fromTypeDef)}\n`
      if (targetIsNullable) {
        subBuilder += `var ${targetVar} ${templateType(targetTypeDef, {
          scope: "tf",
        })}\n`;
        subBuilder += `if ${tfAccessor} != nil {\n`;
      }

      // Build patch semantics with per-item changed accumulator for classes
      let listChangedVar: string | undefined;
      let stateAccessor: string | undefined;
      let itemChangedVar: string | undefined;
      let itemStatePresentVar: string | undefined;
      if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
        const tfAccessorParts = tfAccessor.split(".");
        tfAccessorParts[0] = "opts.State";
        stateAccessor = tfAccessorParts.join(".");
        listChangedVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Changed",
        );
        // Initialize with length check
        subBuilder += `${templateIndent(
          indent,
        )}${listChangedVar} := opts == nil || opts.State == nil || len(${stateAccessor}) != len(${tfAccessor})\n`;
      }

      subBuilder += `${targetVar} ${
        targetIsNullable ? "=" : ":="
      } make(${templateType(targetTypeDef, {
        scope: "tf",
      })}, 0, len(${tfAccessor}))\n`;

      const iterVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName,
        "Index",
      );
      subBuilder += `${templateIndent(
        indent,
      )}for ${iterVariable} := range ${tfAccessor} {\n`;

      // For patch semantics, declare per-item changed tracker and state presence guard
      if (listChangedVar && stateAccessor) {
        itemChangedVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName + "_ItemChanged",
          "",
        );
        itemStatePresentVar = getPluralizedVarSymbolName(
          symbolManager,
          fieldName + "_ItemStatePresent",
          "",
        );
        subBuilder += `${templateIndent(
          indent + 1,
        )}var ${itemChangedVar} bool\n`;
        subBuilder += `${templateIndent(
          indent + 1,
        )}${itemStatePresentVar} := opts != nil && opts.State != nil && ${iterVariable} < len(${stateAccessor})\n`;
      }

      tfAccessor = `${tfAccessor}[${iterVariable}]`;
      for (let i = 0; i < targetTypeDef.ItemType.Fields.length; i++) {
        const field = targetTypeDef.ItemType.Fields[i];
        const equivalent =
          fromTypeDef.ItemType?.Fields.find(
            (f) => sanitizeFieldName(field.Name) == sanitizeFieldName(f.Name),
          ) || fromTypeDef.ItemType?.Fields[i];
        if (!equivalent) {
          continue;
        }
        const symbolName = getPluralizedVarSymbolName(
          symbolManager,
          field.Name,
          "",
        );
        innerSymbols[symbolName] = field;
        subBuilder += generateModelToSDKMethodBody(
          symbolManager,
          symbolName,
          field.Type,
          equivalent.Type,
          field.Nullable,
          field.Optional,
          equivalent.Optional || equivalent.Nullable,
          field.Type.Scope,
          getWithinType(),
          sanitizeFieldName(equivalent.Name),
          field.Name,
          tfAccessor,
          entity,
          indent + 2,
          hierarchy + `.${sanitizeTFStateName(field.Name)}`,
          operation,
          false,
          itemChangedVar,
          itemStatePresentVar,
          field.IsAdditionalProperties,
        );
      }

      // Check if this item changed and update list changed flag
      if (listChangedVar && itemChangedVar && itemStatePresentVar) {
        subBuilder += `${templateIndent(
          indent + 1,
        )}if ${itemChangedVar} || (opts != nil && opts.State != nil && !${itemStatePresentVar}) {\n`;
        subBuilder += `${templateIndent(indent + 2)}${listChangedVar} = true\n`;
        subBuilder += `${templateIndent(indent + 1)}}\n`;
      }

      subBuilder += `${templateIndent(
        indent + 1,
      )}${targetVar} = append(${targetVar}, ${templateType(
        targetTypeDef.ItemType,
        { scope: "tf" },
      )}{\n`;
      for (const symbolName of Object.keys(innerSymbols)) {
        const field = innerSymbols[symbolName];
        subBuilder += `${templateIndent(indent + 2)}${sanitizeFieldName(
          field.Name,
        )}: ${symbolName},\n`;
      }
      subBuilder += `${templateIndent(indent + 1)}})\n`;
      subBuilder += `${templateIndent(indent)}}\n`;

      // After the loop, handle patch semantics: only use list if changed
      if (listChangedVar) {
        subBuilder += `${templateIndent(indent)}if !${listChangedVar} {\n`;
        subBuilder += `${templateIndent(indent + 1)}${targetVar} = nil\n`;
        subBuilder += `${templateIndent(indent)}}\n`;
        if (changedVar) {
          subBuilder += `${templateIndent(indent)}if ${listChangedVar} {\n`;
          subBuilder += `${templateIndent(indent + 1)}${changedVar} = true\n`;
          subBuilder += `${templateIndent(indent)}}\n`;
        }
      }
      if (targetIsNullable) {
        subBuilder += `}\n`;
      }

      return subBuilder;
    } else if (
      targetTypeDef.ItemType.Type == "union" &&
      targetTypeDef.ItemType.AssociatedTypes?.length
    ) {
      let subBuilder = "";
      // subBuilder += `${templateIndent(indent)}// ${hierarchy} target=${JSON.stringify(targetTypeDef)}\n`
      // subBuilder += `${templateIndent(indent)}// ${hierarchy} from=${JSON.stringify(fromTypeDef)}\n`

      if (targetIsNullable) {
        subBuilder += `var ${targetVar} ${sanitizeType(
          targetTypeDef,
          true,
          "",
        )}\n`;
        subBuilder += `if ${tfAccessor} != nil {\n`;
      }

      subBuilder += `${targetVar} ${
        targetIsNullable ? "=" : ":="
      } make(${sanitizeType(
        targetTypeDef,
        targetIsNullable || targetIsOptional,
        "",
      )}, 0, len(${tfAccessor}))\n`;

      {
        const iterVariable = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Item",
        );
        subBuilder += `${templateIndent(
          indent,
        )}for ${iterVariable} := range ${tfAccessor} {\n`;
        tfAccessor = `${tfAccessor}[${iterVariable}]`;
      }
      for (const targetAssociatedType of targetTypeDef.ItemType
        .AssociatedTypes) {
        const targetAssociatedTypeName = sanitizeTFUnionTypeName(
          targetTypeDef.ItemType,
          targetAssociatedType,
        );
        const targetSDKFieldName =
          targetAssociatedType.Name || targetAssociatedTypeName;
        const matchingSourceAssociatedType =
          findMatchingSourceUnionAssociatedType(
            fromTypeDef.ItemType,
            targetTypeDef.ItemType,
            targetAssociatedType,
          );
        if (!matchingSourceAssociatedType) {
          continue;
        }

        const sourceUnionFieldName = sanitizeTFUnionTypeName(
          fromTypeDef.ItemType,
          matchingSourceAssociatedType,
        );
        const symbolName = getPluralizedVarSymbolName(
          symbolManager,
          targetAssociatedType.Name || targetAssociatedTypeName,
          "",
        );

        if (isAggregateType(targetAssociatedType)) {
          subBuilder += `if ${tfAccessor}.${sourceUnionFieldName} != nil {\n`;
        } else {
          subBuilder += `if !${tfAccessor}.${sourceUnionFieldName}.IsUnknown() && !${tfAccessor}.${sourceUnionFieldName}.IsNull() {\n`;
        }

        subBuilder += generateModelToSDKMethodBody(
          symbolManager,
          symbolName,
          targetAssociatedType,
          matchingSourceAssociatedType,
          false,
          false,
          false,
          targetAssociatedType.Scope,
          getWithinType(),
          sourceUnionFieldName,
          targetSDKFieldName,
          tfAccessor,
          entity,
          indent + 2,
          hierarchy + `.${sourceUnionFieldName}`,
          operation,
          false,
          changedVar,
          stateNonNilGuard,
        );
        subBuilder += `${targetVar} = append(${targetVar}, ${
          targetTypeDef.ContainsNull ? "&" : ""
        }${templateType(targetTypeDef.ItemType, { scope: "tf" })}{\n`;
        subBuilder += `${sanitizeFieldName(
          targetSDKFieldName,
        )}: &${symbolName},\n`;
        subBuilder += `})\n`;
        subBuilder += `}\n`;
      }

      subBuilder += `}\n`;

      if (targetIsNullable) {
        subBuilder += `}\n`;
      }

      return subBuilder;
    } else if (isGoArray(targetTypeDef)) {
      let subBuilder = "";

      if (targetIsNullable) {
        subBuilder += `var ${targetVar} ${templateType(targetTypeDef, {
          scope: "tf",
        })}\n`;
        subBuilder += `if ${tfAccessor} != nil {\n`;
      }

      subBuilder += `${targetVar} ${
        targetIsNullable ? "=" : ":="
      } make(${templateType(targetTypeDef, {
        scope: "tf",
      })}, 0, len(${tfAccessor}))\n`;

      {
        const iterVariable = getPluralizedVarSymbolName(
          symbolManager,
          fieldName,
          "Index",
        );
        subBuilder += `${templateIndent(
          indent,
        )}for ${iterVariable} := range ${tfAccessor} {\n`;
        tfAccessor = `${tfAccessor}[${iterVariable}]`;
      }
      const tmpVariable = getPluralizedVarSymbolName(
        symbolManager,
        fieldName,
        "Tmp",
      );
      subBuilder += generateModelToSDKMethodBody(
        symbolManager,
        tmpVariable,
        targetTypeDef.ItemType,
        fromTypeDef.ItemType,
        targetTypeDef.ContainsNull,
        false,
        false,
        targetTypeDef.ItemType.Scope,
        getWithinType(),
        "",
        "",
        tfAccessor,
        entity,
        indent + 1,
        hierarchy + `.${sanitizeTFStateName(fieldName)}`,
        operation,
        false,
        changedVar,
        stateNonNilGuard,
      );
      subBuilder += `${templateIndent(
        indent + 1,
      )}${targetVar} = append(${targetVar}, ${tmpVariable})\n`;
      subBuilder += `${templateIndent(indent)}}\n`;

      if (targetIsNullable) {
        subBuilder += `}\n`;
      }

      return subBuilder;
    } else {
      throw new Error(
        `\n\nUnexpected type detected in Entity "${
          entity.Name
        }" ${hierarchy}\n${JSON.stringify(targetTypeDef)}\n`,
      );
    }
  }
  if (targetTypeDef.AssociatedTypes?.length) {
    if (targetTypeDef.Scope && String(targetTypeDef.Scope).length) {
      addGenImport(`${getSDKPackage()}/models/${targetTypeDef.Scope}`);
    }
    let subBuilder = `${templateIndent(indent)}var ${targetVar} ${sanitizeType(
      targetTypeDef,
      targetIsNullable || targetIsOptional,
      "",
    )}\n`;
    if (fromIsOptional) {
      subBuilder += `${templateIndent(indent)}if ${tfAccessor} != nil {\n`;
      indent++;
    }
    // subBuilder += `${templateIndent(indent)}// Generate(${fromIsOptional}, ${targetIsOptional} ${JSON.stringify(targetTypeDef)}, ${JSON.stringify(fromTypeDef)}, ${withinType})\n`

    // Precompute state presence for all union variants for patch semantics
    const statePresenceVars: Record<string, string> = {};
    if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
      const tfAccessorParts = tfAccessor.split(".");
      tfAccessorParts[0] = "opts.State";
      const stateUnionAccessor = tfAccessorParts.join(".");
      const statePresenceChecks: string[] = [];

      for (const targetAssociatedType of targetTypeDef.AssociatedTypes) {
        const targetAssociatedTypeName = sanitizeTFUnionTypeName(
          targetTypeDef,
          targetAssociatedType,
        );
        const matchingSourceAssociatedType =
          findMatchingSourceUnionAssociatedType(
            fromTypeDef,
            targetTypeDef,
            targetAssociatedType,
          );
        if (!matchingSourceAssociatedType) {
          continue;
        }

        const sourceUnionFieldName = sanitizeTFUnionTypeName(
          fromTypeDef,
          matchingSourceAssociatedType,
        );
        const presenceVar = getPluralizedVarSymbolName(
          symbolManager,
          (targetAssociatedType.Name || targetAssociatedTypeName) +
            "_StatePresent",
          "",
        );
        statePresenceVars[sourceUnionFieldName] = presenceVar;
        statePresenceChecks.push(presenceVar);

        // Determine presence check based on whether it's an aggregate type
        if (isAggregateType(targetAssociatedType)) {
          subBuilder += `${templateIndent(
            indent,
          )}${presenceVar} := opts != nil && opts.State != nil && ${stateUnionAccessor}.${sourceUnionFieldName} != nil\n`;
        } else {
          subBuilder += `${templateIndent(
            indent,
          )}${presenceVar} := opts != nil && opts.State != nil && !${stateUnionAccessor}.${sourceUnionFieldName}.IsUnknown() && !${stateUnionAccessor}.${sourceUnionFieldName}.IsNull()\n`;
        }
      }

      if (statePresenceChecks.length > 0) {
        const stateAnyPresent = getPluralizedVarSymbolName(
          symbolManager,
          "stateAnyVariantPresent",
          "",
        );
        subBuilder += `${templateIndent(
          indent,
        )}${stateAnyPresent} := ${statePresenceChecks.join(" || ")}\n`;
        statePresenceVars["__any__"] = stateAnyPresent;
      }
    }

    for (const targetAssociatedType of targetTypeDef.AssociatedTypes) {
      const targetAssociatedTypeName = sanitizeTFUnionTypeName(
        targetTypeDef,
        targetAssociatedType,
      );
      const targetSDKFieldName = sanitizeClassName(
        targetAssociatedType.Name || targetAssociatedTypeName,
      );
      const matchingSourceAssociatedType =
        findMatchingSourceUnionAssociatedType(
          fromTypeDef,
          targetTypeDef,
          targetAssociatedType,
        );

      if (!matchingSourceAssociatedType) {
        continue;
      }

      const sourceUnionFieldName = sanitizeTFUnionTypeName(
        fromTypeDef,
        matchingSourceAssociatedType,
      );
      const symbolName = getPluralizedVarSymbolName(
        symbolManager,
        targetAssociatedType.Name || targetAssociatedTypeName,
        "",
      );
      innerSymbols[symbolName] = {
        Name: targetSDKFieldName,
        Optional: false,
      };
      const stateAccessor = `${tfAccessor}`;

      // For patch semantics, create changed accumulator and get state presence
      let variantChangedVar: string | undefined;
      let variantStatePresent: string | undefined;
      if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
        variantChangedVar = getPluralizedVarSymbolName(
          symbolManager,
          (targetAssociatedType.Name || targetAssociatedTypeName) + "_Changed",
          "",
        );
        variantStatePresent = statePresenceVars[sourceUnionFieldName];
        subBuilder += `${templateIndent(
          indent,
        )}var ${variantChangedVar} bool\n`;
      }

      subBuilder += generateModelToSDKMethodBody(
        symbolManager,
        symbolName,
        targetAssociatedType,
        matchingSourceAssociatedType,
        true,
        true,
        true,
        targetAssociatedType.Scope,
        getWithinType(),
        sourceUnionFieldName,
        targetSDKFieldName,
        stateAccessor,
        entity,
        indent,
        hierarchy + `.${sanitizeTFStateName(targetAssociatedTypeName)}`,
        operation,
        false,
        variantChangedVar,
        variantStatePresent,
      );

      // Compute type change condition: type switched from a different variant
      let typeChangedCondition = "";
      if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
        const stateAnyPresent = statePresenceVars["__any__"];
        if (stateAnyPresent && variantStatePresent) {
          typeChangedCondition = ` || (opts != nil && opts.State != nil && ${stateAnyPresent} && !${variantStatePresent})`;
        }
      }

      if (hasPatchSemantics && (targetIsOptional || targetIsNullable)) {
        subBuilder += `${templateIndent(
          indent,
        )}if ${symbolName} != nil && (opts == nil || opts.State == nil${
          variantChangedVar ? ` || ${variantChangedVar}` : ""
        }${typeChangedCondition}) {\n`;
      } else {
        subBuilder += `${templateIndent(indent)}if ${symbolName} != nil {\n`;
      }
      // subBuilder += `${templateIndent(indent + 1)} // "any type" ${JSON.stringify(field)}\n`
      if (targetTypeDef.Type == "union") {
        subBuilder += `${templateIndent(indent + 1)}${targetVar} = ${
          targetIsNullable || targetIsOptional ? "&" : ""
        }${templateType(targetTypeDef, { scope: "tf" })}{\n`;
        subBuilder += `${templateIndent(
          indent + 2,
        )}${targetSDKFieldName}: ${symbolName},\n`;
        subBuilder += `${templateIndent(indent + 1)}}\n`;
      } else {
        subBuilder += `${templateIndent(
          indent + 1,
        )}${targetVar} = ${symbolName}\n`;
      }
      subBuilder += `${templateIndent(indent)}}\n`;
    }

    if (fromIsOptional) {
      indent--;
      subBuilder += `${templateIndent(indent)}}\n`;
    }
    return subBuilder;
  }
  if (targetTypeDef.Enum?.Values && targetTypeDef.Enum?.Values.length) {
    let validatorType = "";
    switch (targetTypeDef.Enum.Type.Type.toString()) {
      case "string":
        validatorType = "String";
        break;
      case "float32":
        validatorType = "Float32";
        break;
      case "int32":
        validatorType = "Int32";
        break;
      case "integer":
        validatorType = "Int64";
        break;
      case "number":
        validatorType = "Float64";
        break;
      default:
        throw new Error(
          `Unknown enum detected in ${hierarchy}\n\n${JSON.stringify(
            targetTypeDef,
          )}\n`,
        );
    }
    if (!targetIsNullable && !targetIsOptional) {
      const enumAssignOp = declarationHoisted ? "=" : ":=";
      return `${templateIndent(
        indent,
      )}${targetVar} ${enumAssignOp} ${templateType(targetTypeDef, {
        scope: "tf",
      })}(${tfAccessor}.Value${validatorType}())\n`;
    }
    if (targetTypeDef.Scope && String(targetTypeDef.Scope).length) {
      addGenImport(`${getSDKPackage()}/models/${targetTypeDef.Scope}`);
    }
    const patchSemantics = stateInequalityCheck(
      tfAccessor,
      hasPatchSemantics,
      stateNonNilGuard,
    );
    const enumNullAssignOp = declarationHoisted ? "=" : ":=";
    return (
      `${templateIndent(
        indent,
      )}${targetVar} ${enumNullAssignOp} new(${templateType(targetTypeDef, {
        scope: "tf",
      })})\n` +
      `${templateIndent(
        indent,
      )}if !${tfAccessor}.IsUnknown() && !${tfAccessor}.IsNull()${patchSemantics} {\n` +
      `${templateIndent(indent + 1)}*${targetVar} = ${templateType(
        targetTypeDef,
        { scope: "tf" },
      )}(${tfAccessor}.Value${validatorType}())\n` +
      (changedVar
        ? `${templateIndent(indent + 1)}${changedVar} = true\n`
        : "") +
      `${templateIndent(indent)}} else {\n` +
      `${templateIndent(indent + 1)}${targetVar} = nil\n` +
      `${templateIndent(indent)}}\n`
    );
  }
  if (targetTypeDef.Type == "any") {
    addGenImport("encoding/json");
    const goType = isAdditionalProperties
      ? "map[string]any"
      : templateType(targetTypeDef, { scope: "tf" });
    const varDecl = declarationHoisted
      ? ""
      : `${templateIndent(indent)}var ${targetVar} ${goType}\n`;
    if (targetIsNullable || targetIsOptional) {
      return (
        varDecl +
        `${templateIndent(
          indent,
        )}if !${tfAccessor}.IsUnknown() && !${tfAccessor}.IsNull() {\n` +
        `${templateIndent(
          indent + 1,
        )}_ = json.Unmarshal([]byte(${tfAccessor}.ValueString()), &${targetVar})\n` +
        `${templateIndent(indent)}}\n`
      );
    } else {
      return (
        varDecl +
        `${templateIndent(
          indent,
        )}_ = json.Unmarshal([]byte(${tfAccessor}.ValueString()), &${targetVar})\n`
      );
    }
  }

  throw new Error(
    `\n\nUnexpected type detected in Entity "${
      entity.Name
    }" ${hierarchy}\n${JSON.stringify(targetTypeDef)}\n`,
  );
}
