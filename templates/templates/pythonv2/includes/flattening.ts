interface BodyVariant {
  // Display/source name from the member TypeDef. Used for docstrings and
  // debug; not part of the emitted signature.
  name: string;

  // Field carrying the discriminator value for this variant, if the union has
  // a discriminator. Null when the union has no discriminator.
  discriminatorField: FieldDef | null;

  // The literal value of the discriminator for this variant (e.g. "model" or
  // "agent"). Empty when discriminatorField is null.
  discriminatorValue: string;

  // Unpacked fields of this variant's class. These become the typed kwargs of
  // the overload signature. Order matches the member's class definition.
  fields: FieldDef[];

  // The original member TypeDef. Kept so callers can reach the underlying
  // class type when constructing pydantic models at impl time.
  type: TypeDef;
}

// Returns the union member TypeDefs of the operation's request body, or null
// if the body is not a flattenable oneOf-of-classes shape.
// @ts-ignore
function getOneOfClassBodyMembers(op: Operation): TypeDef[] | null {
  if (!op || !op.Request || !op.Request.RequestBody) {
    return null;
  }
  const requestBody = op.Request.RequestBody;
  const bodyType = requestBody.Type;
  if (!bodyType) {
    return null;
  }
  if (bodyType.Type.toString() !== "union") {
    return null;
  }

  // Collect the member TypeDefs once per concrete class.
  let members: TypeDef[] = [];
  if (bodyType.Discriminator && bodyType.Discriminator.Mapping) {
    // Mapping can list the same target Type under multiple discriminator values, e.g.
    // {"active": Status1, "pending": Status1, "inactive": Status2, "archived": Status2}
    // skip repeat references so each concrete class contributes exactly one member.
    const seen = new Set<string>();
    for (const mapping of bodyType.Discriminator.Mapping) {
      if (!mapping || !mapping.Type) {
        return null;
      }
      if (seen.has(mapping.Type.Name)) {
        continue;
      }
      seen.add(mapping.Type.Name);
      members.push(mapping.Type);
    }
  } else if (bodyType.AssociatedTypes && bodyType.AssociatedTypes.length > 0) {
    // AssociatedTypes can also list the same Type twice when the union has
    // both `oneOf` and `anyOf` siblings collapsed during preprocessing.
    const seen = new Set<string>();
    for (const associated of bodyType.AssociatedTypes) {
      if (!associated || seen.has(associated.Name)) {
        continue;
      }
      seen.add(associated.Name);
      members.push(associated);
    }
  } else {
    return null;
  }

  for (const member of members) {
    if (
      !member ||
      member.Type.toString() !== "class" ||
      !member.Fields ||
      member.Fields.length === 0
    ) {
      return null;
    }
  }

  // Collapse members whose emitted overload signatures are structurally
  // identical. Pyright treats TypedDicts structurally, so distinct schemas with
  // nested fields of the same shape still overlap even when their type names
  // differ. Keep the first occurrence to preserve OpenAPI ordering semantics
  // ("first wins").
  const fingerprintSeen = new Set<string>();
  const unique: TypeDef[] = [];
  for (const member of members) {
    const fingerprint = bodyVariantTypeFingerprint(member);
    if (fingerprintSeen.has(fingerprint)) {
      continue;
    }
    fingerprintSeen.add(fingerprint);
    unique.push(member);
  }

  return unique;
}

function bodyVariantTypeFingerprint(
  typeDef: TypeDef | null | undefined,
): string {
  if (!typeDef) {
    return "unknown";
  }

  if (typeDef.Fields && typeDef.Fields.length > 0) {
    const parts = typeDef.Fields.map((field: FieldDef) => {
      return [
        field.Name,
        field.Optional ? "optional" : "required",
        bodyVariantTypeFingerprint(field.Type),
      ].join("|");
    });
    return `class(${parts.sort().join(",")})`;
  }

  if (typeDef.AssociatedTypes && typeDef.AssociatedTypes.length > 0) {
    const parts = typeDef.AssociatedTypes.map((associated: TypeDef) =>
      bodyVariantTypeFingerprint(associated),
    );
    return `${typeDef.Type?.toString() ?? "associated"}(${parts
      .sort()
      .join(",")})`;
  }

  return `${typeDef.Type?.toString() ?? "unknown"}:${typeDef.Name ?? ""}`;
}

// @ts-ignore
function shouldEmitBodyVariantOverloads(op: Operation): boolean {
  if (context?.Global?.Config?.BodyVariantOverloads !== true) {
    return false;
  }
  const requestBody = op?.Request?.RequestBody;
  if (!requestBody) {
    return false;
  }
  if (requestBody.Optional || requestBody.Nullable) {
    return false;
  }

  const members = getOneOfClassBodyMembers(op);
  return members !== null && members.length > 0;
}
registerTemplateFunc(
  "shouldEmitBodyVariantOverloads",
  shouldEmitBodyVariantOverloads,
);

// A field merged across all variants of a flattenable union body. Used to
// generate the impl signature and dispatch logic for overload-compatible
// union bodies. Each merged field is Optional at the impl level so the impl
// signature is callable with any subset of variant kwargs; per-variant
// required-ness is enforced by the @overload signatures above.
interface MergedBodyField {
  // Sanitized Python parameter name (`type` → `type_`, etc.). Computed
  // once at merge time so downstream callers can rely on stable kwarg names.
  paramName: string;

  // Underlying FieldDef from a representative variant. Used for the
  // body-assembly dict literal and as the source of `Name` / `Optional`
  // flags. Type rendering uses `fieldsByVariant` so unioned variants
  // surface as `Union[...]` rather than a single variant's type.
  field: FieldDef;

  // All FieldDefs that contributed to this merged field (one per variant
  // in which the field appears). Iterated when rendering the impl
  // signature so cross-variant type differences turn into `Union[...]`
  // and the impl signature stays a strict supertype of every overload.
  fieldsByVariant: FieldDef[];

  // Variants that declare this field. Length == 1 → variant-exclusive
  // (candidate discriminator). Length == numVariants → shared field.
  variantNames: string[];

  // Literal discriminator values when this is a discriminator field with a
  // const value across variants (e.g. `type: Literal["obj1"]`). Empty
  // otherwise. At impl level we render as `Literal["a"] | Literal["b"]` so
  // callers can pass any allowed literal; overloads above narrow per
  // variant.
  literalValues: string[];

  // True when the impl signature should keep this field required (i.e.
  // every variant requires it). False when at least one variant treats it
  // as optional. Currently always false to preserve the all-optional impl
  // pattern, but tracked so future tightening is easy.
  requiredEverywhere: boolean;
}

// Returns the merged field list for the union body, preserving the order in
// which fields first appear across variant declarations. This is the
// canonical kwarg ordering for both the impl signature and the body-assembly
// dict literal.
// @ts-ignore
function getMergedBodyFields(op: Operation): MergedBodyField[] {
  const variants = getBodyVariants(op);
  if (variants.length === 0) {
    return [];
  }

  const order: string[] = [];
  const byName: Record<string, MergedBodyField> = {};

  for (const variant of variants) {
    const discriminatorFieldName = variant.discriminatorField?.Name ?? "";
    for (const field of variant.fields) {
      if (!byName[field.Name]) {
        order.push(field.Name);
        byName[field.Name] = {
          paramName: sanitizeParameterName(field.Name),
          field,
          fieldsByVariant: [field],
          variantNames: [variant.name],
          literalValues: [],
          requiredEverywhere: !field.Optional,
        };
      } else {
        const merged = byName[field.Name];
        merged.fieldsByVariant.push(field);
        if (!merged.variantNames.includes(variant.name)) {
          merged.variantNames.push(variant.name);
        }
        if (field.Optional) {
          merged.requiredEverywhere = false;
        }
      }

      if (
        field.Name === discriminatorFieldName &&
        variant.discriminatorValue &&
        !byName[field.Name].literalValues.includes(variant.discriminatorValue)
      ) {
        byName[field.Name].literalValues.push(variant.discriminatorValue);
      }
    }
  }

  // Impl args are all-optional to match the most permissive overload's
  // defaultability. Override here so subsequent typing helpers
  // emit `| Omit = omit` (Speakeasy uses `OptionalNullable[T] = UNSET`).
  for (const name of order) {
    byName[name].requiredEverywhere = false;
  }

  return order.map((name) => byName[name]);
}

// Returns the atomic Python type tokens for a variant's field, ready to
// feed into a flat top-level `Union[...]` at the impl signature.
//
// sanitizeInputParameterType wraps Class types in `Union[Class, TypedDict]`
// to surface the dict-shaped overload; merging those across variants would
// produce `Union[Union[A, A_TypedDict], Union[B, B_TypedDict]]`. We bypass
// that wrapper by calling sanitizeType twice — once for the pydantic
// class, once for the TypedDict variant — and returning the two strings
// directly. The caller (templateBodyVariantImplArguments) accumulates
// these across variants and emits a single flat Union.
// @ts-ignore
function variantFieldAtomicTypes(field: FieldDef): string[] {
  const cls = sanitizeType(field, { optional: false, nullable: false });
  const td = sanitizeType(
    field,
    { optional: false, nullable: false },
    "",
    null,
    true,
    false,
    "",
    DEFAULT_TYPED_DICT_SUFFIX,
  );
  if (cls === td) {
    return [cls];
  }
  return [cls, td];
}

function getBodyVariantNonBodyParamFields(op: Operation): FieldDef[] {
  return op?.Arguments?.ParamFields ?? [];
}

function getBodyVariantNonBodyMethodFields(op: Operation): FieldDef[] {
  const bodyFieldNames = new Set<string>();
  for (const field of op?.Arguments?.BodyFields ?? []) {
    bodyFieldNames.add(field.Name);
  }
  if (op?.Arguments?.BodyField?.Name) {
    bodyFieldNames.add(op.Arguments.BodyField.Name);
  }
  if (op?.Request?.Field?.Name) {
    bodyFieldNames.add(op.Request.Field.Name);
  }
  for (const merged of getMergedBodyFields(op)) {
    bodyFieldNames.add(merged.field.Name);
  }

  return (op?.Arguments?.Sorted ?? []).filter(
    (field) => !bodyFieldNames.has(field.Name),
  );
}

function getBodyVariantKeywordOnlyNonBodyParamFields(
  op: Operation,
): FieldDef[] {
  return getBodyVariantNonBodyMethodFields(op).filter(
    (field) => !isPythonPositionalOperationArgument(field),
  );
}

function hasBodyVariantKeywordOnlyNonBodyParams(op: Operation): boolean {
  return getBodyVariantKeywordOnlyNonBodyParamFields(op).length > 0;
}
registerTemplateFunc(
  "hasBodyVariantKeywordOnlyNonBodyParams",
  hasBodyVariantKeywordOnlyNonBodyParams,
);

function templateBodyVariantNonBodyOverloadArguments(op: Operation): string {
  return getBodyVariantKeywordOnlyNonBodyParamFields(op)
    .map((field) => {
      const paramName = sanitizeParameterName(field.Name);
      const paramType = sanitizeInputParameterType(field, {
        optional: isArgumentOptional(op, field),
        nullable: false,
        iterableCollections: true,
      });
      return `${paramName}: ${paramType}${templateInputParamDefault(
        op,
        field,
      )}`;
    })
    .join(",\n        ");
}
registerTemplateFunc(
  "templateBodyVariantNonBodyOverloadArguments",
  templateBodyVariantNonBodyOverloadArguments,
);

function templateBodyVariantNonBodyImplArguments(op: Operation): string {
  return getBodyVariantKeywordOnlyNonBodyParamFields(op)
    .map((field) => {
      const paramName = sanitizeParameterName(field.Name);
      const paramType = sanitizeInputParameterType(field, {
        optional: true,
        nullable: true,
        iterableCollections: true,
      });
      if (field.Name === "security") {
        return `${paramName}: ${paramType}${templateInputParamDefault(
          op,
          field,
        )}`;
      }
      addImport("types", "UNSET", true);
      return `${paramName}: ${paramType} = UNSET`;
    })
    .join(",\n        ");
}
registerTemplateFunc(
  "templateBodyVariantNonBodyImplArguments",
  templateBodyVariantNonBodyImplArguments,
);

// Renders the impl signature's catch-all `**body_kwargs: Any` parameter.
// The per-variant @overloads above use `**kwargs: Unpack[<VariantTypedDict>]`
// to narrow types at the call site; the impl receives the body fields as a
// single dict and routes them through the body-assembly dict literal. This
// also keeps the impl signature strictly compatible with every overload
// (pyright's reportInconsistentOverload requires impl to accept any kwarg
// keys the overloads accept).
// @ts-ignore
function templateBodyVariantImplArguments(op: Operation): string {
  const merged = getMergedBodyFields(op);
  if (merged.length === 0) {
    return "";
  }
  addImport("typing", "Any");
  return "**body_kwargs: Any";
}
registerTemplateFunc(
  "templateBodyVariantImplArguments",
  templateBodyVariantImplArguments,
);

// Renders runtime XOR guards for mutually-exclusive variant-exclusive
// fields. Emitted at the top of the impl body so misuse fails fast with a
// clear ValueError. Returns an empty string when there are no incompatible
// pairs.
// @ts-ignore
function templateBodyVariantXorGuards(op: Operation): string {
  const pairs = getMutuallyExclusiveFieldPairs(op);
  if (pairs.length === 0) {
    return "";
  }

  // Body fields arrive via the impl's `**body_kwargs` catch-all (keyed by
  // their original OAS property name). Presence check is membership; missing
  // keys mean the caller omitted the field.
  const lines: string[] = [];
  for (const [a, b] of pairs) {
    lines.push(
      `        if "${a}" in body_kwargs and "${b}" in body_kwargs:\n` +
        `            raise ValueError("Cannot supply both '${a}' and '${b}'.")`,
    );
  }
  return lines.join("\n");
}
registerTemplateFunc(
  "templateBodyVariantXorGuards",
  templateBodyVariantXorGuards,
);

// Returns the FieldDef list for a single variant overload, in the order
// they should appear as kwargs. Used to feed getMethodParamDescriptions so
// each overload's docstring lists per-kwarg `:param ...:` lines instead of
// the stale single `:param request:` line.
// @ts-ignore
function getBodyVariantOverloadFieldDefs(
  op: Operation,
  variantIdx: number,
): FieldDef[] {
  const variants = getBodyVariants(op);
  const variant = variants[variantIdx];
  return variant
    ? getBodyVariantKeywordOnlyNonBodyParamFields(op).concat(variant.fields)
    : [];
}
registerTemplateFunc(
  "getBodyVariantOverloadFieldDefs",
  getBodyVariantOverloadFieldDefs,
);

// Returns the FieldDef list for the impl signature, one entry per merged
// kwarg. Uses the representative field captured at merge time, which keeps
// the docstring stable and matches the kwarg order in the signature.
// @ts-ignore
function getBodyVariantImplFieldDefs(op: Operation): FieldDef[] {
  return getBodyVariantKeywordOnlyNonBodyParamFields(op).concat(
    getMergedBodyFields(op).map((m) => m.field),
  );
}
registerTemplateFunc(
  "getBodyVariantImplFieldDefs",
  getBodyVariantImplFieldDefs,
);

// Renders the typed `request: Union[Model, ModelTypedDict]` kwarg used by
// the back-compat overload. Existing call sites in pythonv2 test fixtures
// (and customer code that hasn't migrated to flat kwargs yet) keep working
// by passing `request=`. The impl signature also accepts `request` as an
// optional kwarg; when provided it bypasses the merged-kwarg path entirely.
// @ts-ignore
function templateBodyVariantRequestKwarg(
  op: Operation,
  optional: boolean,
): string {
  const requestField = op?.Request?.Field;
  if (!requestField) {
    return "";
  }
  const requestType = sanitizeInputParameterType(requestField, {
    optional: false,
    nullable: false,
    iterableCollections: true,
  });
  if (optional) {
    addImport("types", "OptionalNullable", true);
    addUtilsImport();
    return `request: OptionalNullable[${requestType}] = UNSET`;
  }
  return `request: ${requestType}`;
}
registerTemplateFunc(
  "templateBodyVariantRequestKwarg",
  templateBodyVariantRequestKwarg,
);

// Renders the impl-body expression that assembles the unmarshal payload.
// With **body_kwargs as the catch-all impl param, the dict already carries
// the OAS-keyed body fields; pass it through directly.
// @ts-ignore
function templateBodyVariantImplBodyDict(op: Operation): string {
  return "dict(body_kwargs)";
}
registerTemplateFunc(
  "templateBodyVariantImplBodyDict",
  templateBodyVariantImplBodyDict,
);

function templateBodyVariantImplRequestAssignment(op: Operation): string {
  const requestField = op?.Request?.Field;
  if (!requestField) {
    return "";
  }

  const requestType = sanitizeType(requestField, {
    optional: false,
    nullable: false,
  });

  if (op.Request.IsRequestBody) {
    return `            request_ = cast(${requestType}, utils_.unmarshal(_body_kwargs, ${requestType}))`;
  }

  const paramFields = getBodyVariantNonBodyParamFields(op);
  const lines: string[] = [];
  if (paramFields.length > 0) {
    const paramEntries = paramFields.map((field) => {
      const fieldName = field.Name.replaceAll("\\", "\\\\").replaceAll(
        '"',
        '\\"',
      );
      const paramName = sanitizeParameterName(field.Name);
      return `"${fieldName}": ${paramName}`;
    });

    lines.push(
      `            _request_kwargs: dict[str, Any] = {${paramEntries.join(
        ", ",
      )}}`,
      `            _request_kwargs = {k: v for k, v in _request_kwargs.items() if v is not UNSET}`,
    );
  } else {
    lines.push(`            _request_kwargs: dict[str, Any] = {}`);
  }

  const bodyField = op?.Arguments?.BodyField;
  if (bodyField) {
    const bodyFieldName = bodyField.Name.replaceAll("\\", "\\\\").replaceAll(
      '"',
      '\\"',
    );
    lines.push(
      `            _request_kwargs["${bodyFieldName}"] = _body_kwargs`,
    );
  } else {
    lines.push(`            _request_kwargs.update(_body_kwargs)`);
  }

  lines.push(
    `            request_ = cast(${requestType}, utils_.unmarshal(_request_kwargs, ${requestType}))`,
  );
  return lines.join("\n");
}
registerTemplateFunc(
  "templateBodyVariantImplRequestAssignment",
  templateBodyVariantImplRequestAssignment,
);

// Returns the runtime expression used by SSE overload composition to choose
// the Accept header after a flattened union body has been marshalled. When
// path/query params force a request wrapper, the stream flag lives on the
// wrapped request body field rather than the top-level request object.
// @ts-ignore
function templateBodyVariantSSEStreamValue(op: Operation): string {
  const streamFieldName = sseStreamFieldName(op);
  if (op?.Request?.IsRequestBody) {
    return `getattr(request_, "${streamFieldName}", False) is True`;
  }

  const bodyFieldName = op?.Request?.RequestBody?.Name
    ? sanitizeFieldName(op.Request.RequestBody.Name)
    : "";
  if (!bodyFieldName) {
    return `getattr(request_, "${streamFieldName}", False) is True`;
  }

  return `getattr(getattr(request_, "${bodyFieldName}", None), "${streamFieldName}", False) is True`;
}
registerTemplateFunc(
  "templateBodyVariantSSEStreamValue",
  templateBodyVariantSSEStreamValue,
);

// Returns pairs of merged-field names that are mutually exclusive across
// variants. Two fields conflict when no variant declares both: passing both
// at runtime is therefore a usage error and must be guarded with a
// `raise ValueError(...)`.
// @ts-ignore
function getMutuallyExclusiveFieldPairs(
  op: Operation,
): Array<[string, string]> {
  const variants = getBodyVariants(op);
  if (variants.length < 2) {
    return [];
  }

  // Build per-variant set of field names.
  const variantFieldNames: Array<Set<string>> = variants.map(
    (v) => new Set(v.fields.map((f) => f.Name)),
  );

  // Collect all field names that appear in at least one but not all variants.
  // Only variant-exclusive fields can be the basis of a mutually-exclusive
  // pair: shared fields are always compatible.
  const exclusiveNames: string[] = [];
  const merged = getMergedBodyFields(op);
  for (const m of merged) {
    if (m.variantNames.length === 1) {
      exclusiveNames.push(m.field.Name);
    }
  }

  const pairs: Array<[string, string]> = [];
  for (let i = 0; i < exclusiveNames.length; i++) {
    for (let j = i + 1; j < exclusiveNames.length; j++) {
      const a = exclusiveNames[i];
      const b = exclusiveNames[j];
      // a,b mutually exclusive iff no variant has both.
      const compatible = variantFieldNames.some(
        (names) => names.has(a) && names.has(b),
      );
      if (!compatible) {
        pairs.push([a, b]);
      }
    }
  }

  return pairs;
}

// Renders the body-variant kwarg as a single `**kwargs: Unpack[...]` entry
// pointing at the variant's existing TypedDict. The TypedDict already
// carries the typed field list with proper Required/NotRequired markers,
// so the overload signature stays a single line and schema changes
// propagate via regen without touching the template.
//
// Emitted *alongside* the expanded overload (see
// templateBodyVariantExpandedOverloadArguments). Together they cover:
//  - expanded: drives hover/Sphinx/REPL display of the full field list
//    (Pylance only inlines `Unpack[TypedDict]` in hover when at least one
//    expanded overload exists in the set), gives per-field completion
//    inside parens.
//  - Unpack: matches `**dict` spread callers, e.g.
//    `client.create(**stored_params)` where the caller built the params
//    elsewhere as the TypedDict.
// @ts-ignore
function templateBodyVariantUnpackOverloadArguments(
  op: Operation,
  variantIdx: number,
): string {
  const variants = getBodyVariants(op);
  const variant = variants[variantIdx];
  if (!variant) {
    return "";
  }

  // pythonv2 min target is 3.10; typing.Unpack arrived in 3.11. Always import
  // from typing_extensions to keep the baseline compatible.
  addImport("typing_extensions", "Unpack");

  const typedDictRef = sanitizeTypedDictType(
    typeDefToFieldDef(variant.type),
    { optional: false, nullable: false },
    "",
    null,
    true,
    false,
  );

  return `**kwargs: Unpack[${typedDictRef}]`;
}
registerTemplateFunc(
  "templateBodyVariantUnpackOverloadArguments",
  templateBodyVariantUnpackOverloadArguments,
);

// Renders the expanded per-field kwarg list for a single variant overload.
// Each TypedDict field becomes an explicit `name: Type` entry; optional
// fields take `= ...` (Ellipsis) as the stub default, matching the
// convention pyright/mypy use for `NotRequired[T]` in protocol stubs.
// The discriminator field, when present, is narrowed to its `Literal["..."]`
// value so the overload selects on that key.
// @ts-ignore
function templateBodyVariantExpandedOverloadArguments(
  op: Operation,
  variantIdx: number,
  streamType: "false" | "true" | "" = "",
): string {
  const variants = getBodyVariants(op);
  const variant = variants[variantIdx];
  if (!variant) {
    return "";
  }

  const lines: string[] = [];
  for (const field of variant.fields) {
    const paramName = sanitizeParameterName(field.Name);

    let paramType: string;
    if (streamType !== "" && field.Name === sseStreamFieldName(op)) {
      lines.push(
        templateStreamParameterOverloadType(field, op, paramName, streamType),
      );
      continue;
    } else if (
      variant.discriminatorField &&
      field === variant.discriminatorField &&
      variant.discriminatorValue
    ) {
      addImport("typing", "Literal");
      paramType = `Literal["${variant.discriminatorValue}"]`;
    } else {
      paramType = sanitizeTypedDictType(
        field,
        { optional: false, nullable: false },
        "",
        null,
        true,
        false,
      );
    }

    // Required fields: no default. Optional (`NotRequired[T]` in the
    // TypedDict): `= ...` matches the stub-style default Pylance + Sphinx
    // expect for overloaded protocol declarations and prevents the
    // overload from accidentally accepting `None` where the TypedDict
    // would not.
    const defaultClause = field.Optional ? " = ..." : "";
    lines.push(`${paramName}: ${paramType}${defaultClause}`);
  }

  return lines.join(",\n        ");
}
registerTemplateFunc(
  "templateBodyVariantExpandedOverloadArguments",
  templateBodyVariantExpandedOverloadArguments,
);

// Returns the BodyVariant list for an operation whose body is a flattenable
// oneOf-of-classes. Callers should gate via shouldEmitBodyVariantOverloads
// before invoking. The discriminator field is identified by looking up the
// Discriminator.TypePropertyName on each member.
// @ts-ignore
function getBodyVariants(op: Operation): BodyVariant[] {
  const members = getOneOfClassBodyMembers(op);
  if (!members) {
    return [];
  }

  const bodyType = op.Request.RequestBody.Type;
  const discriminator = bodyType.Discriminator;
  const discriminatorPropName = discriminator?.TypePropertyName ?? "";

  // Build a name → DiscriminatorValue map so we can recover the const value
  // each variant declares for the discriminator field.
  const discriminatorValueByMemberName: Record<string, string> = {};
  if (discriminator?.Mapping) {
    for (const mapping of discriminator.Mapping) {
      if (mapping?.Type?.Name && mapping?.Name) {
        discriminatorValueByMemberName[mapping.Type.Name] = mapping.Name;
      }
    }
  }

  const variants: BodyVariant[] = [];
  const memberOrder = new Map<string, number>();
  members.forEach((member: TypeDef, index: number) => {
    memberOrder.set(member.Name, index);
  });

  for (const member of members) {
    let discriminatorField: FieldDef | null = null;
    let discriminatorValue = "";
    if (discriminatorPropName && member.Fields) {
      const field =
        member.Fields.find((f: FieldDef) => f.Name === discriminatorPropName) ??
        null;
      discriminatorField = field;
      discriminatorValue =
        discriminatorValueByMemberName[member.Name] ??
        member.DiscriminatorPreApplied ??
        "";
    }

    variants.push({
      name: member.Name,
      discriminatorField,
      discriminatorValue,
      fields: (member.Fields ?? []).slice(),
      type: member,
    });
  }

  variants.sort((a, b) => {
    const specificityDiff =
      bodyVariantSpecificityScore(b.type) - bodyVariantSpecificityScore(a.type);
    if (specificityDiff !== 0) {
      return specificityDiff;
    }
    return (memberOrder.get(a.name) ?? 0) - (memberOrder.get(b.name) ?? 0);
  });

  return variants;
}

function bodyVariantSpecificityScore(
  typeDef: TypeDef | null | undefined,
): number {
  if (!typeDef) {
    return 0;
  }

  if (typeDef.Fields && typeDef.Fields.length > 0) {
    return typeDef.Fields.reduce((score: number, field: FieldDef) => {
      const fieldScore = field.Optional ? 1 : 2;
      return score + fieldScore + bodyVariantSpecificityScore(field.Type);
    }, 0);
  }

  if (typeDef.AssociatedTypes && typeDef.AssociatedTypes.length > 0) {
    return Math.max(
      ...typeDef.AssociatedTypes.map((associated: TypeDef) =>
        bodyVariantSpecificityScore(associated),
      ),
    );
  }

  return 0;
}
registerTemplateFunc("getBodyVariants", getBodyVariants);
