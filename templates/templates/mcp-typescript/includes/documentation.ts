function templateEnumDocs(type: TypeDef): string {
  const enumFormat = "union"; // mcp-typescript only uses const objects with named properties
  const isOpen = type.Enum?.Open;
  const isString = type.Enum?.Type.Type.toString() === "string";
  const tsType = isString ? "string" : "number";
  const fallbackType = `Unrecognized<${tsType}>`;

  let preamble = "";
  if (isOpen) {
    preamble = `This is an open enum. Unrecognized values will be captured as the \`Unrecognized<${tsType}>\` branded type.\n\n`;
  }

  if (enumFormat === "union") {
    const members = type.Enum?.Values.map((v) => {
      return isString ? `"${v}"` : `${v}`;
    });

    if (isOpen) {
      members.push(fallbackType);
    }

    let output = "```typescript\n";
    output += members.join(" | ") + "\n";
    output += "```";

    return preamble + output;
  } else {
    const contents: string[][] = [["Name", "Value"]];

    let enumNames = getEnumNames(type);

    type.Enum.Values.forEach((value, index) => {
      contents.push(["`" + enumNames[index] + "`", value]);
    });

    if (type.Enum?.Open) {
      contents.push(["-", `\`${fallbackType}\``]);
    }

    return preamble + createMarkdownTable(contents);
  }
}

registerTemplateFunc("templateEnumDocs", templateEnumDocs);

function templateUnionDocs(type: TypeDef): string {
  let unionDocs = "";

  let types: TypeDef[] = type.AssociatedTypes;
  if (type.Discriminator) {
    types = type.Discriminator.Mapping.map((m) => m.Type);
  }

  for (const subType of types) {
    const typename = sanitizeType(subType, false);
    seedFaker(typename);
    const example = templateModelUsage(typeDefToFieldDef(subType), 0);

    unionDocs += `### \`${typename}\`\n\n`;
    unionDocs += "```typescript\n";
    unionDocs +=
      formatUsageSnippetOutput(
        `const value: ${typename} = ${example};`,
      ).trimEnd() + `\n`;
    unionDocs += "```\n\n";
  }

  return unionDocs;
}

registerTemplateFunc("templateUnionDocs", templateUnionDocs);

function isRemoteMcp(): boolean {
  return context.Global?.Config?.CloudflareEnabled ?? false;
}

registerTemplateFunc("isRemoteMcp", isRemoteMcp);
