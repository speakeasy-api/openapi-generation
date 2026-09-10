function checksum(s: string): number {
  var chk = 0x12345678;
  var len = s.length;
  for (var i = 0; i < len; i++) {
    chk += s.charCodeAt(i) * (i + 1);
  }

  return chk;
}

/**
 * Escapes Terraform interpolation and template directive sequences in an
 * already-quoted HCL string literal (i.e. the output of JSON.stringify).
 *
 * Terraform treats `${` as interpolation and `%{` as a template directive.
 * To include these literally, they must be doubled: `$${` and `%%{`.
 *
 * See: https://developer.hashicorp.com/terraform/language/expressions/strings
 */
function templateTerraformStringLiteral(value: string): string {
  // In JavaScript replaceAll, $$ in the replacement string produces a single
  // literal $. So "$$$$" produces "$$" and "$$$${" produces "$${".
  return value.replaceAll("${", "$$$${").replaceAll("%{", "%%{");
}

function templateTfExampleIntValue(value: number, typeDef: TypeDef) {
  switch (typeDef.Type.toString()) {
    case "int32":
    case "integer":
      return `${value}`;
    case "bigint":
      switch (typeDef.Format) {
        case "string":
          return `"${value}"`;
        default:
          return `${value}`;
      }
  }
}

function templateTFPrimitiveValue(field: AttributeField) {
  let fallback = `...my_${sanitizeTFStateName(field.name)}...`;
  const typeDef = field.type;

  if (!field.name) {
    const lastItem = field.hierarchy.split(".").pop();
    // Special case for sets of items
    if (
      lastItem.startsWith("[") &&
      lastItem.slice(1, -1) != "0" &&
      lastItem.endsWith("]")
    ) {
      fallback = `...${lastItem.slice(1, -1)}...`;
    } else {
      fallback = "...";
    }
  }
  let value = undefined;
  let example: any = undefined;
  let fakedExample: any = undefined;

  faker.seed(checksum(field.hierarchy));
  if (field.example && field.example.ToJSON) {
    example = getExampleValue(field.example);
  }
  if (!example && field.example && typeCompliant(field.type, field.example)) {
    example = field.example;
  } else if (
    !example &&
    field.example &&
    !typeCompliant(field.type, field.example)
  ) {
  }
  if (!example && typeDef.Examples?.length > 0) {
    const exampleIdx = faker.number.int(typeDef.Examples.length - 1);
    if (typeDef.Examples[exampleIdx].ToJSON) {
      example = getExampleValue(typeDef.Examples[exampleIdx]);
    }
  }
  if (!example && field.defaultValue && field.defaultValue.Value !== "null") {
    example = field.defaultValue.Value;
  }

  switch (typeDef.Type.toString()) {
    case "enum":
      if (!example) {
        example = faker.helpers.arrayElement(typeDef.Enum.Values);
      }
      value = templateTFPrimitiveValue({
        ...field,
        example,
        type: typeDef.Enum.Type,
      });
      break;
    case "string":
      if (typeDef.ContentMediaType === "application/json") {
        if (example) {
          const jsonStr =
            typeof example === "string" ? example : JSON.stringify(example);
          value = jsonToHCLExpression(jsonStr);
        } else {
          value = "jsonencode({})";
        }
        break;
      }
      if (!example && typeDef.Enum?.Values?.length) {
        example = faker.helpers.arrayElement(typeDef.Enum.Values);
      }
      if (typeDef.Format) {
        switch (typeDef.Format.toString()) {
          case "email":
            fakedExample = faker.internet.email();
            break;
          case "id":
            fakedExample = faker.string.uuid();
            break;
          case "uuid":
            fakedExample = faker.string.uuid();
            break;
          case "uri":
            fakedExample = faker.internet.url();
            break;
          case "hostname":
            fakedExample = faker.internet.domainName();
            break;
          case "ipv4":
            fakedExample = faker.internet.ip();
            break;
          case "ipv6":
            fakedExample = faker.internet.ipv6();
            break;
          case "name":
            fakedExample = faker.person.fullName();
            break;
          case "firstname":
            fakedExample = faker.person.firstName();
            break;
          case "lastname":
            fakedExample = faker.person.lastName();
            break;
          case "city":
            fakedExample = faker.location.city();
            break;
          case "country":
            fakedExample = faker.location.country();
            break;
          case "postalcode":
          case "postcode":
          case "zipcode":
            fakedExample = faker.location.zipCode();
            break;
          case "address":
          case "street":
            fakedExample = faker.location.streetAddress();
            break;
          case "countrycode":
            fakedExample = faker.location.countryCode();
            break;
          case "phone":
          case "mobile":
            fakedExample = faker.phone.number();
            break;
          case "company":
            fakedExample = faker.company.name();
            break;
          case "jobtitle":
            fakedExample = faker.person.jobTitle();
            break;
          case "username":
            fakedExample = faker.internet.userName();
            break;
          case "title":
            fakedExample = faker.person.prefix();
            break;
          case "gender":
          case "sex":
            fakedExample = faker.person.sex();
            break;
          case "json":
            // faker no longer has a built in method to generate a json string
            fakedExample = `{"foo":"mxz.v8ISij","bar":29154,"bike":8658,"a":"GxTlw$nuC:","b":40693,"name":"%'<FTou{7X","prop":"X(bd4iT>77"}`;
            break;
        }
      }

      value = example ?? fakedExample ?? fallback;
      if (
        typeDef.Validations?.MaxLength &&
        value.length > typeDef.Validations.MaxLength
      ) {
        value = value.trim(".").slice(0, typeDef.Validations.MaxLength);
      }
      value = templateTerraformStringLiteral(JSON.stringify(value));
      break;
    case "int32":
    case "bigint":
    case "integer": {
      if (!example && typeDef.Enum?.Values?.length) {
        example = faker.helpers.arrayElement(typeDef.Enum.Values);
      }
      const min = typeDef.Validations?.Minimum ?? 0;
      const max = typeDef.Validations?.Maximum ?? min + 10;
      fakedExample = faker.number.int({ min, max }).toString();
      value = templateTfExampleIntValue(example ?? fakedExample, typeDef);
      break;
    }
    case "float32":
    case "number": {
      if (!example && typeDef.Enum?.Values?.length) {
        example = faker.helpers.arrayElement(typeDef.Enum.Values);
      }
      const min = typeDef.Validations?.Minimum ?? 0;
      const max = typeDef.Validations?.Maximum ?? min + 10;
      fakedExample = faker.number
        .float({ min, max, fractionDigits: 2 })
        .toString();
      value = `${example ?? fakedExample}`;
      break;
    }
    case "boolean":
      fakedExample = faker.datatype.boolean();
      value = Boolean(example ?? fakedExample) ? "true" : "false";
      break;
    case "date":
      fakedExample = faker.date
        .past({
          years: faker.number.int({ min: 1, max: 3 }),
          refDate: "2023-01-01T00:00:00.000Z",
        })
        .toISOString()
        .split("T")[0];
      value = JSON.stringify(example ?? fakedExample);
      break;
    case "date-time":
      fakedExample = faker.date
        .past({
          years: faker.number.int({ min: 1, max: 3 }),
          refDate: "2023-01-01T00:00:00.000Z",
        })
        .toISOString();
      value = JSON.stringify(example ?? fakedExample);
      break;
    case "bytes":
      if (example) {
        value = JSON.stringify(example);
      } else {
        value = 'filebase64("${path.module}/example")';
      }
      break;
    case "any":
      fallback = `{ "see": "documentation" }`;
      value = templateTerraformStringLiteral(
        JSON.stringify(example ?? fallback),
      );
      break;
    default:
      throw new Error(`invalid type ${typeDef.Type.toString()}`);
  }

  if (value === undefined) {
    throw new Error(`no value received for ${typeDef.Type.toString()}`);
  }

  return value;
}

/** Returns a stable, single union associated type based on the parent hierarchy. */
function filterUnionAssociatedTypes(
  parent: AttributeField,
  child: AttributeField,
): boolean {
  if (parent.type.Type.toString() !== "union") {
    return true;
  }

  const associatedTypeIndex = parent.type.AssociatedTypes.findIndex(
    (t) => t === child.type,
  );

  // Do not render if the child unexpectedly not match any of the associated types.
  if (associatedTypeIndex === -1) {
    return false;
  }

  const associatedTypeIndexToRender =
    checksum(parent.hierarchy) % parent.type.AssociatedTypes.length;

  return associatedTypeIndex === associatedTypeIndexToRender;
}

function templateTfResource(
  entity: TerraformEntity,
  resourceType: TerraformResourceType,
) {
  const iterator: IteratorFunction = (
    field: AttributeField,
    renderChildren: ({ indent, filter }: RenderChildrenOptions) => string,
  ) => {
    if (field.type.Extensions?.TerraformIgnore?.Schema) {
      return {};
    }

    const sanitizedFieldName = sanitizeFieldName(field.name);

    // Check if this is a root-level field (entity.field)
    const isRootAttribute = field.hierarchy.split(".").length === 2;
    const isPaginationInputField =
      isRootAttribute && sanitizedFieldName in entity.PaginationInputFields;

    if (isPaginationInputField) {
      return {};
    }

    const isGlobalField =
      isRootAttribute && sanitizedFieldName in entity.GlobalFields;

    const fieldConfig = configFromTypeDef(
      field.type,
      field.defaultValue,
      field.optional,
      resourceType,
      isGlobalField,
    );
    if (!fieldConfig.Required && !fieldConfig.Optional) {
      return {};
    }

    let children = "";
    const result: string[] = [];

    switch (field.type.Type.toString()) {
      case "class":
      case "union":
        result.push(`${field.name} = {`);

        children = renderChildren({
          indent: field.indent + 1,
          filter: (child) => filterUnionAssociatedTypes(field, child),
        });

        if (children.length == 0) {
          children = `${templateIndent(1)}# ...`;
        }

        result.push(children.trimEnd());
        result.push(`}`);
        break;
      case "array":
      case "set":
        result.push(`${field.name} = [`);

        switch (field.type.ItemType.Type.toString()) {
          case "class":
          case "union":
            result.push(`{`);

            children = renderChildren({
              indent: field.indent + 2,
              filter: (child) =>
                filterUnionAssociatedTypes(
                  unionAttributeFieldFromTypeDef(
                    field.type.ItemType,
                    sanitizeTFStateName(
                      sanitizeTFUnionTypeName(field.type, field.type.ItemType),
                    ),
                    field.hierarchy,
                  ),
                  child,
                ),
            });

            if (children.length == 0) {
              children = `# ...`;
            }

            result.push(children.trimEnd());
            result.push(`}`);
            break;
          case "array":
          case "set":
            result.push(`[`);
            children = renderChildren({ indent: field.indent + 2 });

            if (children.length == 0) {
              children = `# ...`;
            }

            result.push(children.trimEnd());
            result.push(`]`);
            break;
          case "map":
            result.push(`{`);
            children = renderChildren({ indent: field.indent + 2 });

            if (children.length == 0) {
              children = `# ...`;
            }

            result.push(children.trimEnd());
            result.push(`}`);
            break;
          default:
            let example: any = undefined;

            if (field.example && field.example.ToJSON) {
              example = getExampleValue(field.example);
            }

            if (
              !example &&
              field.example &&
              typeCompliant(field.type, field.example)
            ) {
              example = field.example;
            }

            if (!example && field.type.Examples?.length > 0) {
              faker.seed(checksum(field.hierarchy));

              const exampleIdx = faker.number.int(
                field.type.Examples.length - 1,
              );

              if (field.type.Examples[exampleIdx].ToJSON) {
                example = getExampleValue(field.type.Examples[exampleIdx]);
              }
            }

            if (example && Array.isArray(example)) {
              example.forEach((item, index) => {
                const itemValue = templateTFPrimitiveValue({
                  hierarchy: field.hierarchy + `.[${index}]`,
                  indent: field.indent + 1,
                  name: "",
                  optional: false,
                  nullable: field.nullable,
                  type: field.type.ItemType,
                  example: item,
                } as AttributeField);

                result.push(`${itemValue},`);
              });

              break;
            }

            const count = field.type.Validations.MinItems ?? 1;
            for (let i = 0; i < count; i++) {
              const innerValue = templateTFPrimitiveValue({
                hierarchy: field.hierarchy + `.[${i}]`,
                indent: field.indent + 1,
                name: "",
                optional: false,
                nullable: field.nullable,
                type: field.type.ItemType,
              } as AttributeField);

              result.push(
                `${templateIndent(1)}${innerValue}${i < count - 1 ? "," : ""}`,
              );
            }
        }

        result.push(`]`);
        break;
      case "map":
        result.push(`${field.name} = {`);

        switch (field.type.ItemType.Type.toString()) {
          case "class":
          case "map":
          case "union":
            result.push(`key = {`);

            children = renderChildren({ indent: field.indent + 1 });

            if (children.length == 0) {
              children = `${templateIndent(1)}# ...`;
            }

            result.push(children.trimEnd(), "}");
            break;
          case "array":
          case "set":
            result.push(`key = [`);

            children = renderChildren({ indent: field.indent + 1 });

            if (children.length == 0) {
              children = `${templateIndent(1)}# ...`;
            }

            result.push(children.trimEnd(), "]");
            break;
          default:
            // TODO: Support field example
            if (field.example) {
              result.push(`${templateIndent(1)}# ...`);
            } else {
              switch (field.type.ItemType.Type.toString()) {
                case "any":
                  result.push(`${templateIndent(1)}key = jsonencode("value")`);
                  break;
                case "boolean":
                  result.push(`${templateIndent(1)}key = true`);
                  break;
                case "float32":
                case "number":
                  result.push(`${templateIndent(1)}key = 1.23`);
                  break;
                case "int32":
                case "integer":
                  result.push(`${templateIndent(1)}key = 123`);
                  break;
                case "string":
                  result.push(`${templateIndent(1)}key = "value"`);
                  break;
                default:
                  result.push(`${templateIndent(1)}# ...`);
              }
            }
        }

        result.push(`}`);
        break;
      default:
        if (field.type.IsTerraformPrimitiveType()) {
          result.push(`${field.name} = ${templateTFPrimitiveValue(field)}`);
        } else {
          result.push(`# ...`);
        }
    }
    return {
      [field.name]: result
        .map((line) => `${templateIndent(field.indent)}${line}\n`)
        .join(""),
    };
  };

  const attributes = AttributeIterator(
    entity.SchemaTypeDef,
    iterator,
    1,
    entity.Name,
    {
      indent: 1,
      filter: (field) =>
        filterUnionAssociatedTypes(
          attributeFieldFromTerraformEntity(entity),
          field,
        ),
    },
  );

  if (entity.Server) {
    const attributeName = sanitizeTFStateName(entity.Server.AttributeName);
    attributes[attributeName] = `${attributeName} = "${entity.Server.URL}"\n`;
  }

  return Object.keys(attributes)
    .sort((a, b) => a.localeCompare(b))
    .map((key) => attributes[key])
    .join("");
}

registerTemplateFunc("templateTfResource", templateTfResource);
