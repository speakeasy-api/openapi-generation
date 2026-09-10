function attemptGenerateImportState(entity: TerraformManagedResource): {
  out: string;
  success: boolean;
  importType: TypeDef | undefined;
  importAccessor: string | undefined;
} {
  if (entity.Operations.Read && entity.ImportStateTypeDef) {
    const found = entity.ImportStateTypeDef;
    {
      const explicitlyTriggered = entity.Operations.Read.find((op) => {
        if (
          op.APIOperation.Extensions?.All[
            "x-speakeasy-terraform-import-strategy"
          ] == "json"
        ) {
          return true;
        }
        return false;
      });

      const idRefs = getSimpleIDFields("", entity, found);
      if (!idRefs || explicitlyTriggered || idRefs.length > 1) {
        return handleComplexImport(entity, found);
      } else if (idRefs.length == 0) {
        return {
          out: `resp.Diagnostics.AddError("Not Implemented", "No available import state operation is available for resource ${sanitizeTFStateName(
            entity.Name,
          )}. Reason: no ID fields found")`,
          success: false,
          importType: undefined,
          importAccessor: undefined,
        };
      }
      let Out = "";

      Object.keys(entity.GlobalFields)
        .sort()
        .forEach((fieldName) => {
          const path = `path.Root("${sanitizeTFStateName(fieldName)}")`;
          const value = `r.${fieldName}`;
          const setAttribute = `resp.State.SetAttribute(ctx, ${path}, ${value})`;

          Out += `resp.Diagnostics.Append(${setAttribute}...)\n`;
        });

      const idRef = idRefs[0];
      const symbolManager = makeSymbolMananger();
      symbolManager["r"] = true;

      let value = `req.ID`;

      // Most straightforward to compare enum values as strings, since req.ID
      // starts as a string and since the AST stores enum values as strings.
      if (idRef.type.Type.toString() === "enum") {
        addGenImport("slices");

        const enumValues = idRef.type.Enum.Values.map((v) => JSON.stringify(v));
        const valuesVar = getPluralizedVarSymbolName(
          symbolManager,
          idRef.accessor,
          "Values",
        );

        Out += `${valuesVar} := []string{\n${enumValues.map(
          (v) => `${v},\n`,
        )}}\n`;
        Out += `\n`;
        Out += `if !slices.Contains(${valuesVar}, ${value}) {\n`;
        Out += `resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("ID must be one of %v but was %s", ${valuesVar}, ${value}))\n`;
        Out += `return\n`;
        Out += `}\n`;
        Out += `\n`;
      }

      const valueType =
        idRef.type.Enum?.Type.Type.toString() ?? idRef.type.Type.toString();

      if (["integer", "int32"].includes(valueType)) {
        addGenImport("strconv");

        const tmpVar = getPluralizedVarSymbolName(
          symbolManager,
          idRef.accessor,
        );

        Out += `${tmpVar}, err := strconv.Atoi(req.ID)\n`;
        Out += `if err != nil {\n`;
        Out += `resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("ID must be an integer but was %s", req.ID))\n`;
        Out += `return\n`;
        Out += `}\n`;
        Out += `\n`;

        if (valueType === "int32") {
          addGenImport("math");

          Out += `if ${tmpVar} < math.MinInt32 || ${tmpVar} > math.MaxInt32 {\n`;
          Out += `resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("ID must be an int32 but was %d", ${tmpVar}))\n`;
          Out += `return\n`;
          Out += `}\n`;
          Out += `\n`;
        }

        value = tmpVar;
      }

      addGenImport("github.com/hashicorp/terraform-plugin-framework/path");

      const hierarchy = idRef.accessor.split(".");
      const path = hierarchy
        .map((part, index) => {
          const quotedAttributeName = JSON.stringify(sanitizeTFStateName(part));

          return index === 0
            ? `path.Root(${quotedAttributeName})`
            : `.AtName(${quotedAttributeName})`;
        })
        .join("");

      Out += `resp.Diagnostics.Append(resp.State.SetAttribute(ctx, ${path}, ${value})...)`;

      return {
        out: Out,
        success: true,
        importType: idRef.type,
        importAccessor: idRef.accessor,
      };
    }
  }

  return {
    out: `resp.Diagnostics.AddError("Not Implemented", "No available import state operation is available for resource ${sanitizeTFStateName(
      entity.Name,
    )}.")`,
    success: false,
    importType: undefined,
    importAccessor: undefined,
  };
}

function generateImportDocs(
  entity: TerraformManagedResource,
  importType: TypeDef,
) {
  let importCommandID = "";
  let importConfigID = "";

  if (importType.Type.toString() === "class") {
    importCommandID = `'${generateJSONExample(importType)}'`;
    importConfigID = generateImportBlockIDExample(importType);
  } else {
    const frameworkType = FrameworkTypeFromTypeDef(importType);
    importCommandID = frameworkType.templateExampleJSONValue(importType);
    importConfigID = importCommandID;
  }

  const templateContext = {
    ResourceName: entity.TerraformTypeName,
    ImportCommandID: importCommandID,
    ImportConfigID: importConfigID,
  };

  templateFile(
    "examples/import.sh.stmpl",
    `examples/resources/${entity.TerraformTypeName}/import.sh`,
    templateContext,
  );

  templateFile(
    "examples/import-by-string-id.tf.stmpl",
    `examples/resources/${entity.TerraformTypeName}/import-by-string-id.tf`,
    templateContext,
  );
}

function generateImportState(entity: TerraformManagedResource) {
  const attempt = attemptGenerateImportState(entity);
  if (attempt.success && attempt.importType) {
    // Enrich the import type with examples from the entity TypeDef if the
    // import type has no examples. This is needed because the import type
    // comes from the readRequestShard (path parameters) which may not have
    // examples, while the entity TypeDef (merged from request + response)
    // may have examples from the response schema.
    if (
      attempt.importAccessor &&
      entity.SchemaTypeDef &&
      (!attempt.importType.Examples || attempt.importType.Examples.length === 0)
    ) {
      const { field } = findMatch(entity.SchemaTypeDef, attempt.importAccessor);
      if (field && field.Type.Examples && field.Type.Examples.length > 0) {
        attempt.importType.Examples = field.Type.Examples;
      }
    }
    generateImportDocs(entity, attempt.importType);
  }
  return attempt.out;
}

registerTemplateFunc("generateImportState", generateImportState);

function genIsZeroValue(
  accessor:
    | "ValueBigFloat().Float64()"
    | "ValueBool()"
    | "ValueBoolPointer()"
    | "ValueFloat32()"
    | "ValueFloat32Pointer()"
    | "ValueFloat64()"
    | "ValueFloat64Pointer()"
    | "ValueInt32()"
    | "ValueInt32Pointer()"
    | "ValueInt64()"
    | "ValueInt64Pointer()"
    | "ValueString()"
    | "ValueStringPointer()",
  symbol: string,
): string | undefined {
  switch (accessor) {
    case "ValueString()":
      return `len(${symbol}) == 0`;
  }
  return undefined;
}

function validateAndSet(
  valSymbol: string,
  hierarchy: string[],
  requiredAttributes: TypeDef,
  globalFields: Record<string, FieldDef>,
): string {
  const result: string[] = [];

  for (const field of requiredAttributes.Fields) {
    const matchConfig = field.Type.Extensions?.MatchConfig;
    const matchPath = matchConfig?.Path;
    const fieldName = matchPath ?? field.Name;
    const sanitizedFieldName = sanitizeFieldName(fieldName);
    const curSymbol = `${valSymbol}.${sanitizedFieldName}`;
    const curHierarchy = matchPath
      ? matchPath.split(".")
      : hierarchy.concat(fieldName);

    if (field.Type.Type.toString() === "class") {
      result.push(
        validateAndSet(curSymbol, curHierarchy, field.Type, globalFields),
      );

      continue;
    }

    const accessorType = primitiveAccessor(field.Type, false);

    if (!accessorType && field.Type.Type.toString() !== "enum") {
      throw new Error(
        `unsupported in import validateAndSet ${field.Type.Type} at location ${valSymbol}`,
      );
    }

    addGenImport("github.com/hashicorp/terraform-plugin-framework/path");

    const check =
      field.Optional || field.Nullable
        ? `${curSymbol} == nil`
        : genIsZeroValue(accessorType, curSymbol);
    const frameworkType = FrameworkTypeFromFieldDef(field);
    const isGlobalField =
      curHierarchy.length === 1 && sanitizedFieldName in globalFields;
    const symbolManager = makeSymbolMananger();
    let path = `path.Root("${sanitizeTFStateName(curHierarchy[0])}")`;

    for (let i = 1; i < curHierarchy.length; i++) {
      path += `.AtName("${sanitizeTFStateName(curHierarchy[i])}")`;
    }

    if (check) {
      result.push(`if ${check} {`);

      if (isGlobalField) {
        frameworkType
          .templateTerraformToSDKImports(
            field.Type,
            true,
            field.Optional || field.Nullable,
          )
          .forEach((importStr) => {
            addGenImport(importStr);
          });

        result.push(
          frameworkType.templateTerraformToSDK(
            symbolManager,
            sanitizedFieldName,
            field.Type,
            true,
            field.Optional || field.Nullable,
            curSymbol,
            `r.${sanitizedFieldName}`,
            false,
            "diags",
          ),
        );

        result.push(`if ${check} {`);
      }

      // Only include example hint if there's a real OAS-defined example
      const hasExample = field.Type.Examples?.length > 0;
      const fieldName = sanitizeTFStateName(curHierarchy);
      if (hasExample) {
        const exampleValue = FrameworkTypeFromTypeDef(
          field.Type,
        ).templateExampleJSONValue(field.Type);
        result.push(
          `resp.Diagnostics.AddError("Missing required field", \`The field ${fieldName} is required but was not found in the json encoded ID. It's expected to be a value alike '${exampleValue}'\`)`,
        );
      } else {
        result.push(
          `resp.Diagnostics.AddError("Missing required field", \`The field ${fieldName} is required but was not found in the json encoded ID.\`)`,
        );
      }
      result.push(`return`);

      if (isGlobalField) {
        result.push(`}`);
      }

      result.push(`}`);
    }

    result.push(
      `resp.Diagnostics.Append(resp.State.SetAttribute(ctx, ${path}, ${curSymbol})...)`,
    );
  }

  return result.join("\n");
}

function getSimpleIDFields(
  accessor: string,
  entity: TerraformManagedResource,
  type: TypeDef,
): { accessor: string; type: TypeDef }[] | undefined {
  if (type.Type.toString() === "class") {
    const ids: { accessor: string; type: TypeDef }[] = [];
    for (let i = 0; i < type.Fields.length; i++) {
      let field = type.Fields[i];

      const matchPath = field.Type.Extensions?.MatchConfig?.Path;

      // Handle path aliasing (x-speakeasy-match with a path)
      if (matchPath) {
        const {
          field: newField,
          accessor: newAccessor,
          error,
        } = findMatch(entity.SchemaTypeDef, matchPath);
        if (error) {
          throw new Error(
            `unable to resolve x-speakeasy-match path "${matchPath}" for import state: ${error}`,
          );
        }
        field = newField;
        accessor = newAccessor.join(".");
        ids.push({
          accessor: accessor,
          type: field.Type,
        });
        continue;
      }

      // usePriorState without path falls through to normal field processing

      const hasIDField = getSimpleIDFields(
        (accessor.length ? `${accessor}.` : "") + field.Name,
        entity,
        field.Type,
      );
      if (!hasIDField) {
        continue;
      }
      ids.push(...hasIDField);
    }
    return ids;
  }

  return type.TerraformHasInvalidImportTypes()
    ? undefined
    : [{ accessor, type }];
}

/**
 * Templates the struct for receiving a JSON-encoded import string.
 */
function templateImportJSONStruct(requiredAttributes: TypeDef): string {
  let subBuilder = "";
  if (requiredAttributes.Type.toString() !== "class") {
    throw new Error(
      `unsupported in import inlineJSONStruct ${requiredAttributes.Type}`,
    );
  }
  subBuilder += `struct {\n`;
  for (const field of requiredAttributes.Fields) {
    // Short-circuit if the field is a class, as it will be handled recursively.
    if (field.Type.Type.toString() === "class") {
      subBuilder += templateImportJSONStruct(field.Type);

      continue;
    }

    addImportScope(field.Type);

    const fieldName = field.Type.Extensions?.MatchConfig?.Path ?? field.Name;
    const attributeName = sanitizeTFStateName(fieldName);
    const structFieldName = sanitizeFieldName(fieldName);
    const structFieldTag = `\`json:"${attributeName}"\``;
    const structFieldType = sanitizeType(
      field.Type,
      field.Optional || field.Nullable,
      "",
    );

    subBuilder += `${structFieldName} ${structFieldType} ${structFieldTag}\n`;
  }
  subBuilder += `}\n`;
  return subBuilder;
}

function generateJSONExample(requiredAttributes: TypeDef): string {
  const jsonProperties: Record<string, string> = {};

  requiredAttributes.Fields.forEach((fieldDef) => {
    const fieldName =
      fieldDef.Type.Extensions?.MatchConfig?.Path ?? fieldDef.Name;
    const attributeName = sanitizeTFStateName(fieldName);

    if (fieldDef.Type.Type.toString() === "class") {
      jsonProperties[attributeName] = generateJSONExample(fieldDef.Type);

      return;
    }

    const frameworkType = FrameworkTypeFromFieldDef(fieldDef);
    const jsonExample = frameworkType.templateExampleJSONValue(fieldDef.Type);

    jsonProperties[attributeName] = jsonExample;
  });

  let result = "{";

  // Format the JSON properties into a single line string with whitespace
  result += Object.keys(jsonProperties)
    .sort()
    .map((key) => {
      return `"${key}": ${jsonProperties[key]}`;
    })
    .join(", ");

  result += "}";

  // Format the JSON string with indentation
  // return JSON.stringify(jsonProperties, null, 1).replaceAll(`\n`, ``);
  return result;
}

/**
 * Returns a Terraform configuration language import block ID example using
 * the jsonencode() function.
 */
function generateImportBlockIDExample(requiredAttributes: TypeDef): string {
  const jsonProperties: Record<string, string> = {};

  requiredAttributes.Fields.forEach((fieldDef) => {
    const fieldName =
      fieldDef.Type.Extensions?.MatchConfig?.Path ?? fieldDef.Name;
    const attributeName = sanitizeTFStateName(fieldName);

    if (fieldDef.Type.Type.toString() === "class") {
      jsonProperties[attributeName] = generateImportBlockIDExample(
        fieldDef.Type,
      );

      return;
    }

    const frameworkType = FrameworkTypeFromFieldDef(fieldDef);
    const jsonExample = frameworkType.templateExampleJSONValue(fieldDef.Type);

    jsonProperties[attributeName] = jsonExample;
  });

  const result: string[] = ["jsonencode({"];

  result.push(
    ...Object.keys(jsonProperties)
      .sort()
      .map((key) => {
        // 4 space indentation is intentional for containing configuration
        return `    ${key} = ${jsonProperties[key]}`;
      }),
  );

  // 2 space indentation is intentional for containing configuration
  result.push("  })");

  return result.join("\n");
}

function handleComplexImport(
  entity: TerraformManagedResource,
  requiredAttributes: TypeDef,
): {
  out: string;
  success: boolean;
  importType: TypeDef;
  importAccessor: undefined;
} {
  const invalidTypes = requiredAttributes.TerraformInvalidImportTypes();

  if (invalidTypes && invalidTypes.length > 0) {
    const resourceName = sanitizeTFStateName(entity.Name);
    const invalidTypesOut = invalidTypes
      .map((invalidType) => {
        const details =
          `No available import state operation is available for resource ${resourceName} ` +
          `attribute ${invalidType.Hierarchy}. ` +
          `Reason: unhandled type ${invalidType.TypeName} for a required field`;
        return `resp.Diagnostics.AddError("Not Implemented", "${details}")\n`;
      })
      .join("");

    return {
      out: invalidTypesOut,
      success: false,
      importType: requiredAttributes,
      importAccessor: undefined,
    };
  }

  addGenImport("encoding/json");
  addGenImport("bytes");

  let subBuilder = "";

  subBuilder += `${templateIndent(
    1,
  )}dec := json.NewDecoder(bytes.NewReader([]byte(req.ID)))\n`;
  subBuilder += `${templateIndent(1)}dec.DisallowUnknownFields()\n`;
  subBuilder += `${templateIndent(1)}var data ${templateImportJSONStruct(
    requiredAttributes,
  )}\n`;

  subBuilder += `${templateIndent(2)}\n`;
  subBuilder += `${templateIndent(
    1,
  )}if err := dec.Decode(&data); err != nil {\n`;
  subBuilder += `${templateIndent(
    2,
  )}resp.Diagnostics.AddError("Invalid ID", \`The import ID is not valid. It is expected to be a JSON object string with the format: '${generateJSONExample(
    requiredAttributes,
  )}': \` + err.Error())\n`;
  subBuilder += `${templateIndent(2)}return\n`;
  subBuilder += `${templateIndent(1)}}\n\n`;
  subBuilder += validateAndSet(
    "data",
    [],
    requiredAttributes,
    entity.GlobalFields,
  );

  return {
    out: subBuilder,
    success: true,
    importType: requiredAttributes,
    importAccessor: undefined,
  };
}
