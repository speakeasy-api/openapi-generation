interface UnionConversion {
  Member: TypeDef;
  ParamType: string;
  IsValueType: boolean;
  MutatesArgument: boolean;
  Call: (arg: string) => string;
}

const nonConvertibleUnionMemberTypes = [
  "any",
  "event-stream",
  "request",
  "response",
];

const nonConvertibleUnionParamTypes = ["object", "dynamic"];

// @ts-ignore
function templatePartial(): string {
  return context.Global.Config.PartialClasses === true ? "partial " : "";
}
registerTemplateFunc("templatePartial", templatePartial);

// @ts-ignore
function unionTypeKey(typeDef: TypeDef): string {
  return `${typeDef.Type}:${typeDef.OutputLocation}/${typeDef.Name}`;
}

// @ts-ignore
function isUnionMemberValueType(typeDef: TypeDef): boolean {
  if (typeDef.Type.toString() === "enum") {
    return !typeDef.Enum?.Open;
  }
  return isValueType(typeDef);
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

// Filters avoid CS0553 (object source), CS1964 (dynamic source), CS0555 (self source) and CS0557 (duplicate source).
// @ts-ignore
function directUnionConversions(typeDef: TypeDef): {
  all: string[];
  kept: UnionConversion[];
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
        IsValueType: isUnionMemberValueType(m.Type),
        MutatesArgument: m.MutatesArgument,
        Call: (arg: string) => `${factory}(${arg})`,
      };
    });

  const counts = new Map<string, number>();
  for (const c of candidates) {
    counts.set(c.ParamType, (counts.get(c.ParamType) ?? 0) + 1);
  }
  const selfKey = unionTypeKey(typeDef);

  return {
    all: candidates.map((c) => c.ParamType),
    kept: candidates.filter(
      (c) =>
        counts.get(c.ParamType) === 1 &&
        unionTypeKey(c.Member) !== selfKey &&
        !nonConvertibleUnionParamTypes.includes(c.ParamType),
    ),
  };
}

// C# applies at most one user-defined conversion, so nested union members need their own operators.
// @ts-ignore
function collectUnionConversions(
  typeDef: TypeDef,
  visited: Set<string>,
): UnionConversion[] {
  const selfKey = unionTypeKey(typeDef);
  if (visited.has(selfKey) || isUnionOfErrors(typeDef)) {
    return [];
  }
  visited = new Set(visited).add(selfKey);

  const { all, kept } = directUnionConversions(typeDef);

  const nested: UnionConversion[] = [];
  for (const direct of kept) {
    if (direct.Member.Type.toString() !== "union") {
      continue;
    }
    for (const inner of collectUnionConversions(direct.Member, visited)) {
      nested.push({
        ...inner,
        MutatesArgument: direct.MutatesArgument || inner.MutatesArgument,
        Call: (arg: string) =>
          direct.Call(`${direct.ParamType}.${inner.Call(arg)}`),
      });
    }
  }

  const nestedCounts = new Map<string, number>();
  for (const n of nested) {
    nestedCounts.set(n.ParamType, (nestedCounts.get(n.ParamType) ?? 0) + 1);
  }

  return [
    ...kept,
    ...nested.filter(
      (n) =>
        !all.includes(n.ParamType) &&
        nestedCounts.get(n.ParamType) === 1 &&
        unionTypeKey(n.Member) !== selfKey,
    ),
  ];
}

// @ts-ignore
function getUnionImplicitConversions(typeDef: TypeDef): UnionConversion[] {
  if (
    context.Global.Config.UnionImplicitConversions !== true ||
    typeDef.Type.toString() !== "union"
  ) {
    return [];
  }
  return collectUnionConversions(typeDef, new Set());
}

// @ts-ignore
function templateUnionImplicitConversions(typeDef: TypeDef): string {
  const className = sanitizeClassName(typeDef.Name);

  return getUnionImplicitConversions(typeDef)
    .map((c) => {
      const call = c.Call("value");
      const lines = [
        "/// <summary>",
        `/// Implicitly converts a value to <see cref="${className}"/>. Equivalent to <c>${call}</c>.`,
      ];
      if (c.MutatesArgument) {
        lines.push(
          "/// Like the factory, this sets the discriminator property on the passed instance.",
        );
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
