// @ts-ignore
function seralizeFieldDefAsJSON(fieldDef: FieldDef) {
  return JSON.stringify(parseValue(fieldDef), null, 2);
}

// @ts-ignore
function parseValue(fieldDef: FieldDef) {
  switch (fieldDef.Type?.Type.toString()) {
    case "class":
      return parseObject(fieldDef);
    case "enum":
      return parseEnum(fieldDef);
    case "array":
      return parseArray(fieldDef);
    case "map":
      return parseMap(fieldDef);
    default:
      return parseBasicValue(fieldDef);
  }
}

function parseBasicValue(fieldDef: FieldDef) {
  if (!faker) {
    throw new Error("faker hasn't been injected into runtime");
  }

  let example = fieldDef.Default?.Value;

  if (example === undefined && fieldDef.Type.Examples?.length > 0) {
    example = getExampleValue(
      faker.helpers.arrayElement(fieldDef.Type.Examples),
    );
  }

  if (example !== undefined && example === null) {
    return templateNullValue(fieldDef);
  }

  let value = undefined;

  switch (fieldDef.Type.Type.toString()) {
    case "string":
      if (example !== undefined) {
        if (typeof example != "string") {
          value = JSON.stringify(example);
        } else {
          value = example;
        }
      } else if (fieldDef.Type.Format) {
        value = fakeString(fieldDef.Type.Format.toString());
      } else {
        value = templatePlaceholderString();
      }
      break;
    case "int32":
    case "bigint":
    case "integer":
      value = faker.number.int(1000000);
      break;
    case "float32":
    case "decimal":
    case "number":
      value = faker.number.float(10000).toFixed(2);
      break;
    case "boolean":
      value = faker.datatype.boolean();
      break;
    case "date":
      value = fakeDateTime().split("T")[0];
      break;
    case "date-time":
      value = fakeDateTime();
      break;
    case "bytes":
      value = faker.string.hexadecimal({ length: 10 });
      break;
    case "any":
      if (example != undefined) {
        value = example;
      } else {
        value = templatePlaceholderString();
      }
      break;
    default:
      throw new Error(`invalid type ${fieldDef.Type.Type.toString()}`);
  }

  if (value === undefined) {
    throw new Error(`no value received for ${fieldDef.Type.Type.toString()}`);
  }

  return value;
}

// @ts-ignore
function parseObject(fieldDef: FieldDef) {
  let output = new Map<string, any>();
  let example = undefined;

  if (!example && fieldDef.Type.Examples?.length > 0) {
    example = getExampleValue(
      faker.helpers.arrayElement(fieldDef.Type.Examples),
    );
  }

  if (example !== undefined && example === null) {
    return null;
  }

  for (const field of fieldDef.Type.Fields) {
    let fieldExample = example?.[field.OriginalName];

    output.set(field.OriginalName, fieldExample);
  }

  return Object.fromEntries(output);
}

// @ts-ignore
function parseEnum(fieldDef: FieldDef) {
  if (fieldDef.Type.Enum.Values.length > 0) {
    return faker.helpers.arrayElement(fieldDef.Type.Enum.Values);
  } else {
    return templatePlaceholderString();
  }
}

// @ts-ignore
function parseArray(fieldDef: FieldDef) {
  if (fieldDef.Type.Examples?.length > 0) {
    return getExampleValue(faker.helpers.arrayElement(fieldDef.Type.Examples));
  }

  const itemMap = new Map<string, any>();

  if (!isPartOfCycle(fieldDef.Type)) {
    if (fieldDef.Type.ItemType.Fields.length > 0) {
      for (const field of fieldDef.Type.ItemType.Fields) {
        itemMap.set(field.Name, parseValue(field));
      }
    }
  }

  return [Object.fromEntries(itemMap)];
}

// @ts-ignore
function parseMap(fieldDef: FieldDef) {
  let example = undefined;

  if (!example && fieldDef.Type.Examples?.length > 0) {
    example = getExampleValue(
      faker.helpers.arrayElement(fieldDef.Type.Examples),
    );
  }

  if (!example) {
    example = {
      [templatePlaceholderMapKeyValue()]: undefined,
    };
  }

  if (example !== undefined && example === null) {
    return null;
  }

  let output = new Map<string, any>();

  if (example !== undefined && example !== null) {
    for (const key in example) {
      output.set(key, example[key]);
    }
    return output;
  }

  if (!isPartOfCycle(fieldDef.Type)) {
    for (const field of fieldDef.Type.Fields) {
      output.set(field.Name, parseValue(field));
    }
  }

  return output;
}

// @ts-ignore
function templateAnyValue(fieldDef: FieldDef) {
  if (!faker) {
    throw new Error("faker hasn't been injected into runtime");
  }

  return parseValue(fieldDef);
}
