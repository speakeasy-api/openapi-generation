type CSharpModel = {
  Name: string;
  Type: TypeDef;
  OutputLocation: string;
};

// Returns true when a field must be emitted as OptionalNullable<T> rather than
// plain T?. Wrapping is required only when the spec distinguishes three
// states — absent / explicit null / set — which happens for optional+nullable
// schema fields under presenceAwareJsonSerialization.
//
// Excluded even when optional+nullable:
//   - PresenceAwareJSONSerialization disabled: legacy T? semantics, no tri-state needed.
//   - Const fields: value is fixed by the schema, absent vs null is meaningless.
//   - fields never used in a JSON body/response (no "json" annotation): the tri-state is
//     only observable through JSON serialization, so a form/multipart-only body field has
//     no absent/null/set distinction to preserve and stays plain T?.
//
// Note: OptionalNullable<T> has no meaning outside the JSON serializer. A component reached
// by more than one serializer (e.g. a JSON body field that is also a form/multipart field
// or a query/header parameter) carries the "json" annotation, so its fields stay wrapped to
// keep the JSON tri-state; the form/multipart serializers and the URL/header serializers
// then unwrap OptionalNullable<T> at serialization time (see Utilities.UnwrapValue() in
// RequestBodySerializer.cs, URLBuilder.cs and HeaderSerializer.cs). Path parameters are
// forced required upstream, so they are never optional and never wrapped.
// @ts-ignore
function needsOptionalNullableWrapper(fieldDef: FieldDef): boolean {
  return (
    context.Global.Config.PresenceAwareJSONSerialization &&
    fieldDef.Optional &&
    fieldDef.Nullable &&
    !fieldDef.Const &&
    !isScalarOptionalNullableDecimal(fieldDef) &&
    Array.from(fieldDef.Annotations).some((a) => a.Type().toString() === "json")
  );
}
registerTemplateFunc(
  "needsOptionalNullableWrapper",
  needsOptionalNullableWrapper,
);

// NB: A scalar (non-array/non-map) bare-number `format: decimal` field is emitted as
// plain `decimal?` rather than OptionalNullable<decimal?>, even when optional+nullable.
//
// Root cause is Newtonsoft-specific: to dispatch to any JsonConverter, Json.NET first
// performs a generic reader.Read() that parses a scalar JSON number token as `double`
// (FloatParseHandling.Double) *before* the converter runs. So OptionalNullable<decimal?>
// would round-trip a JSON number at ~15 significant digits (silent precision loss), e.g.
// (decimal)(double)3.141592653589793 == 3.14159265358979
// @ts-ignore
function isScalarOptionalNullableDecimal(fieldDef: FieldDef): boolean {
  return (
    fieldDef.Type.Type.toString() === "decimal" &&
    fieldDef.Type.Format != "string"
  );
}

// Returns the C# expression reading `prop`'s value off `varName` for non-JSON
// (form/multipart) serialization. Under presenceAwareJsonSerialization the value is a
// JSON-only OptionalNullable<T> that must be unwrapped before it reaches the form/multipart
// serializers (see Utilities.UnwrapValue); otherwise it is read as-is.
// @ts-ignore
function templateGetPropValue(varName: string): string {
  const getter = `prop.GetValue(${varName})`;
  return context.Global.Config.PresenceAwareJSONSerialization
    ? `Utilities.UnwrapValue(${getter})`
    : getter;
}
registerTemplateFunc("templateGetPropValue", templateGetPropValue);

// Smart-union scoring (unionStrategy: populated-fields) relies on the Required
// JsonProperty attribute and OptionalNullable wrapper type, which are only emitted
// when presenceAwareJsonSerialization is enabled.
// @ts-ignore
function useSmartUnion(typeDef?: TypeDef): boolean {
  if (!context.Global.Config.PresenceAwareJSONSerialization) return false;
  if (context.Global.Config.UnionStrategy !== "populated-fields") return false;
  if (!typeDef) return true;
  const discriminated = !!typeDef.Discriminator?.TypePropertyName;
  if (discriminated) return false;
  return true;
}
registerTemplateFunc("useSmartUnion", useSmartUnion);

// @ts-ignore
function templateFieldType(
  fieldDef: FieldDef,
  fallbackOptional: boolean,
  usageLocation: string,
): string {
  if (!context.Global.Config.PresenceAwareJSONSerialization) {
    return sanitizeType(fieldDef.Type, fallbackOptional, usageLocation);
  }
  if (fieldDef.Const) {
    return sanitizeType(fieldDef.Type, fallbackOptional, usageLocation);
  }
  if (needsOptionalNullableWrapper(fieldDef)) {
    const inner = sanitizeType(fieldDef.Type, true, usageLocation);
    return `OptionalNullable<${inner}>`;
  }
  return sanitizeType(fieldDef.Type, fallbackOptional, usageLocation);
}
registerTemplateFunc("templateFieldType", templateFieldType);

// @ts-ignore
function templateFieldDefault(fieldDef: FieldDef): string {
  if (!context.Global.Config.PresenceAwareJSONSerialization) {
    return templateDefaultValue(fieldDef);
  }
  if (fieldDef.Const || fieldDef.Default) {
    return templateDefaultValue(fieldDef);
  }
  if (needsOptionalNullableWrapper(fieldDef)) {
    return "";
  }
  if (
    fieldDef.Type.Type.toString() === "class" &&
    !fieldDef.Optional &&
    !fieldDef.Nullable
  ) {
    return " = new();";
  }
  return templateDefaultValue(fieldDef);
}
registerTemplateFunc("templateFieldDefault", templateFieldDefault);

// @ts-ignore
function templateShouldSerialize(
  fieldDef: FieldDef,
  className: string,
): string {
  if (!needsOptionalNullableWrapper(fieldDef)) return "";
  const name = sanitizeFieldName(fieldDef.Name, className);
  return `\n        public bool ShouldSerialize${name}() => ${name}.IsSet;`;
}
registerTemplateFunc("templateShouldSerialize", templateShouldSerialize);

// @ts-ignore
function getModelJobs(types: BucketedTypes, path: string): Job[] {
  const jobs: Job[] = [];

  for (const [outputLocation, models] of sequencedMapEntries(types)) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const typeDef of types) {
        let modelName = typeDef.Name;

        const ctx: CSharpModel = {
          Name: modelName,
          Type: typeDef,
          OutputLocation: outputLocation,
        };

        jobs.push(
          createTemplateFileJob(
            `modelfile.cs.stmpl`,
            `${path}/${getModelsLocation(outputLocation)}/${sanitizeFileName(
              modelName,
            )}.cs`,
            ctx,
          ),
        );
      }
    }
  }

  return jobs;
}

const cachedTypeConflicts = new Map<string, boolean>();

// @ts-ignore
function doesTypeConflict(className: string, outputLocation: string): boolean {
  let cacheKey = className;

  function cache(conflicts: boolean): boolean {
    cachedTypeConflicts.set(cacheKey, conflicts);
    return conflicts;
  }

  if (cachedTypeConflicts.has(cacheKey)) {
    return cachedTypeConflicts.get(cacheKey)!;
  }

  if (cSharpReservedClasses.includes(className)) {
    return cache(true);
  }

  if (doesTypeConflictWithSDK(className, context.Global.AST.MainSDK)) {
    return cache(true);
  }

  if (outputLocation != "") {
    cacheKey = `${outputLocation}/${className}`;
    if (cachedTypeConflicts.has(cacheKey)) {
      return cachedTypeConflicts.get(cacheKey)!;
    }
  }

  const typeBuckets: Record<string, TypeDef[]> = {};

  for (const [outputLocation, models] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        if (!typeBuckets[outputLocation]) {
          typeBuckets[outputLocation] = [];
        }

        typeBuckets[outputLocation].push(t);
      }
    }
  }

  for (const bucketName in typeBuckets) {
    if (bucketName == outputLocation) {
      continue;
    }

    const bucket = typeBuckets[bucketName];

    for (const i in bucket) {
      const t = bucket[i];
      if (sanitizeClassName(t.Name, false) == className) {
        return cache(true);
      }
    }
  }

  return cache(false);
}

function doesTypeConflictWithSDK(className: string, sdk: SDK): boolean {
  if (className == sanitizeSDKName(sdk.Type.Name)) {
    return true;
  }

  for (const subSDK of sdk.SubSDKs) {
    if (doesTypeConflictWithSDK(className, subSDK)) {
      return true;
    }
  }

  return false;
}

// Detects two namespace collision scenarios that require fully-qualified
// references in C#:
//
// 1. Segment conflict: a type name matches the first segment of a namespace
//    path (e.g. a class "Models" shadows "Models.Shared.SomeType"), making
//    qualified references ambiguous.
//
// 2. Parent duplication: a model output location produces a namespace path
//    where a child segment duplicates its parent (e.g. Models/Models from
//    x-speakeasy-model-namespace: models), creating ambiguous `using`
//    directives.
let _hasNamespaceConflict: boolean | undefined;
function hasNamespaceConflict(): boolean {
  if (_hasNamespaceConflict !== undefined) return _hasNamespaceConflict;

  const rootSegments = new Set<string>();

  for (const [outputLocation] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    const modelLocation = getModelsLocation(outputLocation);
    const segments = modelLocation.split("/");

    if (segments[0]) {
      rootSegments.add(segments[0]);
    }

    // Check for parent duplication (e.g. Models/Models)
    const seen = new Set<string>();
    seen.add(segments[0] ?? "");
    for (let i = 1; i < segments.length; i++) {
      if (seen.has(segments[i])) {
        _hasNamespaceConflict = true;
        return true;
      }
      seen.add(segments[i]);
    }
  }

  // Check for type names that shadow a root namespace segment
  for (const [, models] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        if (rootSegments.has(sanitizeClassName(t.Name, false))) {
          _hasNamespaceConflict = true;
          return true;
        }
      }
    }
  }

  _hasNamespaceConflict = false;
  return false;
}
