interface UnionConversion {
  Member: TypeDef;
  ParamType: string;
  SignatureKey: string;
  IsValueType: boolean;
  MutatesArgument: boolean;
  FactoryCall: (arg: string) => string;
}

const nonConvertibleUnionMemberTypes = [
  "any",
  "event-stream",
  "request",
  "response",
];

const nonConvertibleUnionParamTypes = ["object", "dynamic"];

// @ts-ignore
function typeDefKey(typeDef: TypeDef): string {
  return `${typeDef.Type}:${typeDef.OutputLocation}/${typeDef.Name}`;
}

// @ts-ignore
function isValueType(typeDef: TypeDef): boolean {
  return !canTypeBeNull(typeDef) && !typeDef.Enum?.Open;
}

// @ts-ignore
function factorySetsDiscriminator(
  member: TypeDef,
  discriminatorProperty: string,
): boolean {
  if (!["class", "error"].includes(member.Type.toString())) {
    return false;
  }
  const field = member.Fields?.find((f) => f.Name === discriminatorProperty);
  return !field?.Const;
}

// C# signatures ignore nullable reference annotations (List<string?> collides with List<string>), but int? and int stay distinct.
// @ts-ignore
function conversionSignatureKey(
  typeDef: TypeDef,
  usageLocation: string,
): string {
  const type = typeDef.Type.toString();
  if (type !== "array" && type !== "map") {
    return sanitizeType(typeDef, false, usageLocation);
  }
  const item = conversionSignatureKey(typeDef.ItemType, usageLocation);
  const nullable =
    typeDef.ContainsNull && isValueType(typeDef.ItemType) ? "?" : "";
  return type === "array"
    ? `List<${item}${nullable}>`
    : `Dictionary<string, ${item}${nullable}>`;
}

// @ts-ignore
function incrementCount(counts: Map<string, number>, key: string, by: number) {
  counts.set(key, (counts.get(key) ?? 0) + by);
}

// Filters avoid CS0553 (object source), CS1964 (dynamic source), CS0555 (self source) and CS0557 (duplicate source).
// @ts-ignore
function directUnionConversions(typeDef: TypeDef): {
  candidates: UnionConversion[];
  sourceCounts: Map<string, number>;
  conversions: UnionConversion[];
} {
  const discriminator = typeDef.Discriminator;
  const members = discriminator?.Mapping
    ? discriminator.Mapping.map((m) => ({
        Type: m.Type,
        Name: m.Name,
        MutatesArgument: factorySetsDiscriminator(
          m.Type,
          discriminator.TypePropertyName,
        ),
      }))
    : (typeDef.AssociatedTypes ?? []).map((t) => ({
        Type: t,
        Name: sanitizeUnionTypeName(t),
        MutatesArgument: false,
      }));

  const candidates: UnionConversion[] = members
    .filter(
      (m) => !nonConvertibleUnionMemberTypes.includes(m.Type.Type.toString()),
    )
    .map((m) => {
      const factory = `Create${sanitizeClassName(m.Name)}`;
      return {
        Member: m.Type,
        ParamType: sanitizeType(m.Type, false, m.Type.OutputLocation),
        SignatureKey: conversionSignatureKey(m.Type, m.Type.OutputLocation),
        IsValueType: isValueType(m.Type),
        MutatesArgument: m.MutatesArgument,
        FactoryCall: (arg: string) => `${factory}(${arg})`,
      };
    });

  const sourceCounts = new Map<string, number>();
  for (const c of candidates) {
    incrementCount(sourceCounts, c.SignatureKey, 1);
  }
  const selfKey = typeDefKey(typeDef);

  return {
    candidates,
    sourceCounts,
    conversions: candidates.filter(
      (c) =>
        sourceCounts.get(c.SignatureKey) === 1 &&
        typeDefKey(c.Member) !== selfKey &&
        !nonConvertibleUnionParamTypes.includes(c.ParamType),
    ),
  };
}

// C# applies at most one user-defined conversion, so nested union members need their own operators.
// @ts-ignore
function collectUnionConversions(
  typeDef: TypeDef,
  visited: Set<string>,
): { conversions: UnionConversion[]; sourceCounts: Map<string, number> } {
  const selfKey = typeDefKey(typeDef);
  if (visited.has(selfKey) || isUnionOfErrors(typeDef)) {
    return { conversions: [], sourceCounts: new Map() };
  }
  visited = new Set(visited).add(selfKey);

  const direct = directUnionConversions(typeDef);

  const nested: UnionConversion[] = [];
  const nestedSourceCounts = new Map<string, number>();
  for (const candidate of direct.candidates) {
    if (candidate.Member.Type.toString() !== "union") {
      continue;
    }
    const inner = collectUnionConversions(candidate.Member, visited);
    for (const [key, count] of inner.sourceCounts) {
      incrementCount(nestedSourceCounts, key, count);
    }
    if (!direct.conversions.includes(candidate)) {
      continue;
    }
    for (const conversion of inner.conversions) {
      nested.push({
        ...conversion,
        MutatesArgument:
          candidate.MutatesArgument || conversion.MutatesArgument,
        FactoryCall: (arg: string) =>
          candidate.FactoryCall(
            `${candidate.ParamType}.${conversion.FactoryCall(arg)}`,
          ),
      });
    }
  }

  const sourceCounts = new Map(direct.sourceCounts);
  for (const [key, count] of nestedSourceCounts) {
    if (!sourceCounts.has(key)) {
      sourceCounts.set(key, count);
    }
  }

  return {
    conversions: [
      ...direct.conversions,
      ...nested.filter(
        (n) =>
          sourceCounts.get(n.SignatureKey) === 1 &&
          !direct.sourceCounts.has(n.SignatureKey) &&
          typeDefKey(n.Member) !== selfKey,
      ),
    ],
    sourceCounts,
  };
}

// @ts-ignore
function getUnionImplicitConversions(typeDef: TypeDef): UnionConversion[] {
  if (
    context.Global.Config.UnionImplicitConversions !== true ||
    typeDef.Type.toString() !== "union"
  ) {
    return [];
  }
  return collectUnionConversions(typeDef, new Set()).conversions;
}

// @ts-ignore
function templateUnionImplicitConversions(typeDef: TypeDef): string {
  const className = sanitizeClassName(typeDef.Name);

  return getUnionImplicitConversions(typeDef)
    .map((c) => {
      const call = c.FactoryCall("value");
      const lines = [
        "/// <summary>",
        `/// Implicitly converts a value to <see cref="${className}"/>. Equivalent to <c>${call}</c>.`,
      ];
      if (c.MutatesArgument) {
        lines.push(
          "/// Like the factory, this sets the discriminator property on the passed instance.",
        );
      }
      if (!c.IsValueType) {
        lines.push("/// A null value converts to a null reference.");
      }
      lines.push("/// </summary>");
      if (c.IsValueType) {
        lines.push(
          `public static implicit operator ${className}(${c.ParamType} value) => ${call};`,
        );
      } else {
        if (!isNetstandard2()) {
          lines.push(
            `[return: System.Diagnostics.CodeAnalysis.NotNullIfNotNull("value")]`,
          );
        }
        lines.push(
          `public static implicit operator ${className}?(${c.ParamType}? value) => value is null ? null : ${call};`,
        );
      }
      return "\n\n" + lines.map((line) => `        ${line}`).join("\n");
    })
    .join("");
}
registerTemplateFunc(
  "templateUnionImplicitConversions",
  templateUnionImplicitConversions,
);
