require("common/fieldDefs.ts");

const DEBUG_MATCH_TYPE_LOGGING = false; // Set this to true to enable very verbose logging of match results (don't leave this set to true outside local dev)
const DEBUG_MATCH_OP_FILTER = ""; // Set this to a specific operation ID to only log match results for that operation (don't leave this set outside local dev)
const DEBUG_MATCH_EXAMPLE_NAME = ""; // Set this to a specific example name to only log match results for that example name (don't leave this set outside local dev)
let CURRENT_MATCH_OP_ID = "";
let CURRENT_MATCH_EXAMPLE_NAME = "";

function debugMatchTypeLogging(message?: any, example?: any) {
  if (
    DEBUG_MATCH_TYPE_LOGGING &&
    (!DEBUG_MATCH_OP_FILTER || CURRENT_MATCH_OP_ID == DEBUG_MATCH_OP_FILTER) &&
    (!DEBUG_MATCH_EXAMPLE_NAME ||
      CURRENT_MATCH_EXAMPLE_NAME == DEBUG_MATCH_EXAMPLE_NAME)
  ) {
    console.log(`[matchTypeWithExample] ${message}`);
    if (example !== undefined) {
      console.log("  example:", JSON.stringify(example));
    }
  }
}

type MatchResult = {
  type: TypeDef;
  rating: number;
};

// @ts-ignore
function matchTypeWithExample(
  type: TypeDef,
  example: any,
  additionalContext?: { operation?: Operation; exampleName?: string },
): MatchResult | undefined {
  CURRENT_MATCH_OP_ID = additionalContext?.operation?.OriginalID ?? "";
  CURRENT_MATCH_EXAMPLE_NAME = additionalContext?.exampleName ?? "";

  debugMatchTypeLogging(
    `Attempting to match type ${type.Name} (${type.Type})`,
    example,
  );

  const res = matchTypeWithExampleInner(type, example, additionalContext);
  if (res) {
    debugMatchTypeLogging(
      `MATCH - example matches ${res.type.Name} (${res.type.Type}) with rating ${res.rating}`,
      example,
    );
  } else {
    debugMatchTypeLogging(`NO MATCH - for example`, example);
  }
  return res;
}

// @ts-ignore
function matchTypeWithExampleInner(
  typeDef: TypeDef,
  example: any,
  additionalContext?: { operation?: Operation },
): MatchResult | undefined {
  if (typeDef == undefined) {
    debugMatchTypeLogging("NO MATCH - typeDef is undefined");
    return undefined;
  }

  if (example === undefined) {
    debugMatchTypeLogging("NO MATCH - example is undefined");
    return undefined;
  }

  if (example === null) {
    debugMatchTypeLogging("NO MATCH - example is null");
    return undefined;
  }

  if (isTypeDefType(typeDef, "any")) {
    if (typeDef.AssociatedTypes.length == 0) {
      return { type: typeDef, rating: 1 };
    }

    const matches = typeDef.AssociatedTypes.filter((t) => t !== undefined)
      .map((t) => {
        const match = matchTypeWithExample(t, example, additionalContext);
        // Return the associated type itself if it matches, not the inner matched type
        return match ? { type: t, rating: match.rating } : undefined;
      })
      .filter((m) => m !== undefined);

    if (matches.length == 0) {
      debugMatchTypeLogging("NO MATCH - associated types (dataType: any)");
      return undefined;
    }

    return matches.sort((a, b) => b.rating - a.rating)[0];
  }

  if (isTypeDefType(typeDef, "union")) {
    if (typeDef.Discriminator) {
      const subType = example[typeDef.Discriminator.TypePropertyName];

      const foundType = subType
        ? typeDef.Discriminator.Mapping.find((t) => t.Name == subType)?.Type
        : undefined;

      if (!foundType) {
        debugMatchTypeLogging(
          `NO MATCH - discriminator types prop: ${typeDef.Discriminator.TypePropertyName} (dataType: union)`,
        );
        return undefined;
      }

      return { type: foundType, rating: 1 };
    }

    const matches = typeDef.AssociatedTypes.filter((t) => t !== undefined)
      .map((t) => {
        const match = matchTypeWithExample(t, example, additionalContext);
        // Return the associated type itself if it matches, not the inner matched type
        return match ? { type: t, rating: match.rating } : undefined;
      })
      .filter((m) => m !== undefined);

    if (matches.length == 0) {
      debugMatchTypeLogging("NO MATCH - associated types (dataType: union)");
      return undefined;
    }

    return matches.sort((a, b) => b.rating - a.rating)[0];
  }

  switch (typeof example) {
    case "string":
      const dateRe = /^\d{4}-\d{2}-\d{2}/;
      const dateTimeRe =
        /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)((-(\d{2}):(\d{2})|Z)?)$/;

      switch (true) {
        case isTypeDefType(typeDef, "date") && dateRe.test(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "date-time") && dateTimeRe.test(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "enum") &&
          typeDef.Enum?.Type.Type.toString() == "string" &&
          typeDef.Enum.Values.includes(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(
          typeDef,
          "bigint",
          "decimal",
          "integer",
          "int32",
          "number",
          "float32",
        ) &&
          typeDef.Format == "string" &&
          /^\d+(\.\d+)?$/.test(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "string"):
          return { type: typeDef, rating: 1 };
        default:
          // TODO: deal with bytes
          debugMatchTypeLogging(
            `NO MATCH - dataType: ${typeDef.Type}, typeof example: string`,
          );
          return undefined;
      }
    case "number":
      switch (true) {
        case isTypeDefType(typeDef, "number"):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "float32"):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "decimal") && typeDef.Format != "string":
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "integer") &&
          typeDef.Format != "string" &&
          Number.isInteger(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "int32") && Number.isInteger(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "bigint") &&
          typeDef.Format != "string" &&
          Number.isInteger(example):
          return { type: typeDef, rating: 1 };
        case isTypeDefType(typeDef, "enum") &&
          (typeDef.Enum?.Type.Type.toString() == "integer" ||
            typeDef.Enum?.Type.Type.toString() == "int32") &&
          Number.isInteger(example) &&
          typeDef.Enum.Values.includes(`${example}`):
          return { type: typeDef, rating: 1 };
        default:
          debugMatchTypeLogging(
            `NO MATCH - dataType: ${typeDef.Type}, typeof example: number`,
          );
          return undefined;
      }
    case "boolean":
      if (isTypeDefType(typeDef, "boolean")) {
        return { type: typeDef, rating: 1 };
      } else {
        debugMatchTypeLogging(
          `NO MATCH - dataType: ${typeDef.Type}, typeof example: boolean`,
        );
        return undefined;
      }
    case "object":
      // Is an array
      if (isTypeDefType(typeDef, "array") && Array.isArray(example)) {
        // If example is empty then we can match it
        if (example.length == 0) {
          return { type: typeDef, rating: 1 };
        }
        // Otherwise test for a match on element type
        const res = matchTypeWithExample(
          typeDef.ItemType,
          example[0],
          additionalContext,
        );
        if (res) {
          return { type: typeDef, rating: 1 };
        }
      }

      // Is a map
      if (isTypeDefType(typeDef, "map")) {
        // If example is empty then we can match it
        if (Object.keys(example).length == 0) {
          return { type: typeDef, rating: 1 };
        }
        // Otherwise test for a match on element type
        const res = matchTypeWithExample(
          typeDef.ItemType,
          example[Object.keys(example)[0]],
          additionalContext,
        );
        if (res) {
          return { type: typeDef, rating: 1 };
        }
      }

      // If its a custom class
      if (typeDef.IsTypeWithFields()) {
        const res = matchClass(typeDef, example, additionalContext);
        if (res) {
          return res;
        }
      }

      // Otherwise not an array, map or matches a class so likely an invalid example for this type
      debugMatchTypeLogging(
        `NO MATCH - dataType: ${typeDef.Type}, typeof example: ${
          Array.isArray(example) ? "array" : "object"
        }`,
      );
      return undefined;
    default:
      debugMatchTypeLogging(
        `NO MATCH - dataType: ${typeDef.Type}, typeof example: unknown`,
      );
      return undefined;
  }
}

function matchClass(
  classType: TypeDef,
  example: any,
  additionalContext?: { operation?: Operation },
): MatchResult | undefined {
  if (classType.Fields.length == 0) {
    if (Object.keys(example).length == 0) {
      return { type: classType, rating: 1 };
    }

    debugMatchTypeLogging(
      `NO CLASS MATCH - class with no field has example with fields`,
    );
    return undefined;
  }

  let exampleCopy = { ...example };

  let matchedFields = 0;
  let additionalPropertiesField: FieldDef | undefined = undefined;

  for (const field of classType.Fields) {
    // Remaining fields in example will be matched with this field
    if (field.IsAdditionalProperties) {
      additionalPropertiesField = field;
      continue;
    }

    const jsonAnno = field.Annotations?.Get("json");
    if (jsonAnno?.Ignore) {
      continue;
    }

    const fieldExample = example[originalFieldName(field)];

    if (fieldExample === undefined) {
      // A required field is not present so no match
      if (!field.Optional) {
        debugMatchTypeLogging(
          `NO CLASS MATCH - field ${field.Name} is required but example is missing`,
        );
        return undefined;
      }

      delete exampleCopy[originalFieldName(field)];
      continue;
    }

    if (fieldExample === null) {
      if (!field.Nullable) {
        debugMatchTypeLogging(
          `NO CLASS MATCH - field ${field.Name} is not nullable but example is null`,
        );
        return undefined;
      }

      matchedFields++;
      delete exampleCopy[originalFieldName(field)];
      continue;
    }

    const match = matchTypeWithExample(
      field.Type,
      fieldExample,
      additionalContext,
    );

    // No match on this field so no match overall
    if (match === undefined) {
      debugMatchTypeLogging(
        `NO CLASS MATCH - field ${field.Name} has example but does not match type`,
      );
      return undefined;
    }

    matchedFields++;
    delete exampleCopy[originalFieldName(field)];
  }

  if (additionalPropertiesField) {
    const match = matchTypeWithExample(
      additionalPropertiesField.Type,
      exampleCopy,
      additionalContext,
    );

    if (match === undefined) {
      debugMatchTypeLogging(
        `NO CLASS MATCH - class had additional properties but example does not match type`,
        exampleCopy,
      );
      return undefined;
    }

    matchedFields += 0.5; // We will score an additional properties match as half of that of a full field match so that a full field match is considered first
    exampleCopy = {};
  }

  if (matchedFields == 0 && Object.keys(exampleCopy).length > 0) {
    debugMatchTypeLogging(
      `NO CLASS MATCH - no relevant fields found but example has fields`,
    );
    return undefined;
  }

  // If we didn't match any fields and there are still fields in the example then no match or there is still fields in the example (TODO: that might be too strict)
  if (Object.keys(exampleCopy).length > 0) {
    debugMatchTypeLogging(
      `NO CLASS MATCH - example had extra fields but class doesn't support additional properties`,
      exampleCopy,
    );
    return undefined;
  }

  const res = {
    type: classType,
    rating: matchedFields / classType.Fields.length,
  };

  debugMatchTypeLogging(
    `CLASS MATCH - ${res.type.Name} (${res.rating} ${matchedFields}/${classType.Fields.length})`,
    example,
  );

  return res;
}

function isTypeDefType(typeDef: TypeDef, ...type: DataType[]): boolean {
  return type.some((t) => t.toString() == typeDef.Type.toString());
}
