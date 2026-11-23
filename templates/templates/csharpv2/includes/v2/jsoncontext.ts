function getJsonContextBucketPrefix(
  outputLocation: string,
  modelNamespace: string = "",
): string {
  const prefix = outputLocation
    ? getJsonContextBucketPrefixes().get(outputLocation) ??
      leafBucketPrefix(outputLocation)
    : "";
  return prefix || modelNamespace.split(".").pop() || getScopePath("models");
}

function leafBucketPrefix(outputLocation: string): string {
  const lastSegment = outputLocation.split("/").pop();
  return lastSegment ? sanitizeClassName(lastSegment) : "";
}

// Every JsonTypeInfo name shares the JsonTypeInfos class, so output locations
// whose last segment sanitizes to the same prefix fall back to their full path.
let _jsonContextBucketPrefixes: Map<string, string> | undefined;
function getJsonContextBucketPrefixes(): Map<string, string> {
  if (_jsonContextBucketPrefixes) return _jsonContextBucketPrefixes;

  const locationsByLeaf = new Map<string, string[]>();
  for (const [outputLocation] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    if (!outputLocation) continue;
    const leaf = leafBucketPrefix(outputLocation);
    if (!leaf) continue;
    const locations = locationsByLeaf.get(leaf) ?? [];
    locations.push(outputLocation);
    locationsByLeaf.set(leaf, locations);
  }

  const prefixes = new Map<string, string>();
  for (const [leaf, locations] of locationsByLeaf) {
    for (const outputLocation of locations) {
      prefixes.set(
        outputLocation,
        locations.length > 1
          ? sanitizeClassName(
              outputLocation.split("/").filter(Boolean).join(" "),
            )
          : leaf,
      );
    }
  }
  _jsonContextBucketPrefixes = prefixes;
  return prefixes;
}

function getJsonContextClassName(outputLocation: string): string {
  return `${getJsonContextBucketPrefix(outputLocation)}JsonContext`;
}

// STJ source generator transitively discovers all types reachable from explicitly
// declared [JsonSerializable] types. When multiple types from different namespaces
// share the same short class name (e.g. Foo.Pet, Bar.Pet), the generator emits
// SYSLIB1031. Explicit [JsonSerializable] declarations with TypeInfoPropertyName
// win over transitive discovery and suppress the warning. This function walks the
// full type graph of typeDef, collecting every cross-namespace type reference into
// deps so the caller can add them as explicit entries in the context file, along
// with collection, Nullable<T> and OptionalNullable<T> wrappers whose default names
// collide (List<Foo.Pet>, List<Bar.Pet>).
function collectJsonContextCrossNsDeps(
  typeDef: TypeDef,
  outputLocation: string,
  deps: Map<string, { Type: string; Name: string }>,
  visited: Set<string>,
): void {
  const id = `${typeDef.OutputLocation ?? ""}::${typeDef.Name}`;
  if (visited.has(id) || typeDef.Truncated) return;
  visited.add(id);

  const visit = (typeRef: TypeDef) => {
    const t = recordCollectionWrappers(typeRef, deps);
    if (!t) return;

    if (t.OutputLocation && t.OutputLocation !== outputLocation) {
      const className = sanitizeClassName(t.Name);
      const fqn = `${getModelNamespace(t.OutputLocation)}.${className}`;
      if (!deps.has(fqn)) {
        deps.set(fqn, {
          Type: fqn,
          Name: `${getJsonContextBucketPrefix(t.OutputLocation)}${className}`,
        });
      }
    }

    collectJsonContextCrossNsDeps(t, outputLocation, deps, visited);
  };

  for (const field of typeDef.Fields ?? []) {
    visit(field.Type);
    recordFieldWrapperTypes(field, deps);
  }
  for (const assoc of typeDef.AssociatedTypes ?? []) visit(assoc);

  // Each JsonContext expands converters using its own resolver in isolation.
  // When a union flows into this bucket via cross-NS, the bucket also needs
  // every collection-wrapper variant the union's factory will request via
  // `options.GetTypeInfo` — otherwise STJ throws NotSupportedException at
  // expansion time. Register them locally even though their element types are
  // already covered by `visit` above.
  collectUnionVariantTypes(typeDef, deps);
}

// Operation shapes get no cross-namespace pass (the root context is built from
// them alone): declare only the reachable types whose default names collide.
function collectJsonContextShapeCollisions(
  shape: TypeDef,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  const element = recordCollectionWrappers(shape, deps);
  if (!element) return;
  recordCollidingType(
    csharpTypeForJsonContext(element).replaceAll("global::", ""),
    jsonTypeInfoName(element),
    deps,
  );

  const { visited } = jsonContextCollisionState(deps);
  const id = `${element.OutputLocation ?? ""}::${element.Name}`;
  if (visited.has(id) || element.Truncated) return;
  visited.add(id);

  for (const field of element.Fields ?? []) {
    collectJsonContextShapeCollisions(field.Type, deps);
    recordFieldWrapperTypes(field, deps);
  }
  for (const assoc of element.AssociatedTypes ?? []) {
    collectJsonContextShapeCollisions(assoc, deps);
  }
}

function jsonContextElement(t: TypeDef | undefined): TypeDef | undefined {
  while (t?.Type?.toString() === "array" || t?.Type?.toString() === "map") {
    t = t.ItemType;
  }
  return t && SERIALIZABLE_TYPES.includes(t.Type?.toString() ?? "")
    ? t
    : undefined;
}

function recordCollectionWrappers(
  t: TypeDef,
  deps: Map<string, { Type: string; Name: string }>,
): TypeDef | undefined {
  const element = jsonContextElement(t);
  if (!element) return undefined;
  for (; t !== element; t = t.ItemType) {
    recordCollidingContextType(jsonContextPropertyType(t), deps);
  }
  return element;
}

function recordFieldWrapperTypes(
  field: FieldDef,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  if (!jsonContextElement(field.Type)) return;
  const fieldType = jsonContextPropertyType(field.Type);
  const nullableValue =
    field.Type.Type?.toString() === "enum" &&
    isValueType(field.Type) &&
    (field.Nullable || (!field.Const && field.Optional));
  const valueType = nullableValue ? `${fieldType}?` : fieldType;
  if (nullableValue) recordCollidingContextType(valueType, deps);
  if (needsOptionalNullableWrapper(field)) {
    recordCollidingContextType(
      `${getModelNamespace("")}.OptionalNullable<${valueType}>`,
      deps,
    );
  }
}

// csharpTypeForJsonContext renders nullable closed-enum collection items as `E`;
// the model property declares `E?`.
function jsonContextPropertyType(t: TypeDef): string {
  const kind = t.Type?.toString();
  if ((kind !== "array" && kind !== "map") || !t.ItemType) {
    return csharpTypeForJsonContext(t).replaceAll("global::", "");
  }
  const nullableItem = t.ContainsNull && isValueType(t.ItemType) ? "?" : "";
  const item = `${jsonContextPropertyType(t.ItemType)}${nullableItem}`;
  return kind === "array"
    ? `System.Collections.Generic.List<${item}>`
    : `System.Collections.Generic.Dictionary<string, ${item}>`;
}

function recordCollidingContextType(
  type: string,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  recordCollidingType(type, jsonTypeInfoNameForContextType(type), deps);
}

const jsonContextCollisionStates = new WeakMap<
  Map<string, { Type: string; Name: string }>,
  { groups: Map<string, Map<string, string>>; visited: Set<string> }
>();

function jsonContextCollisionState(
  deps: Map<string, { Type: string; Name: string }>,
): { groups: Map<string, Map<string, string>>; visited: Set<string> } {
  let state = jsonContextCollisionStates.get(deps);
  if (!state) {
    state = { groups: new Map(), visited: new Set() };
    jsonContextCollisionStates.set(deps, state);
  }
  return state;
}

function recordCollidingType(
  type: string,
  name: string,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  const { groups } = jsonContextCollisionState(deps);
  const key = type.replace(/[\w.]+\.(\w+)/g, "$1");
  const group = groups.get(key) ?? new Map<string, string>();
  groups.set(key, group);
  group.set(type, name);
  if (group.size < 2) return;
  for (const [groupType, groupName] of group) {
    if (!deps.has(groupType)) {
      deps.set(groupType, { Type: groupType, Name: groupName });
    }
  }
}

// Union converter factories request `JsonTypeInfo<List<TVariant>>` (or
// `Dictionary<...>`) via `options.GetTypeInfo` when a variant is itself a
// collection. STJ source-generated contexts only return metadata for
// explicitly declared [JsonSerializable] types — without a wrapper entry, the
// runtime throws NotSupportedException. Walk a union's variants and emit the
// collection wrapper alongside the element type. Element type is already
// reachable via the existing crossNsDeps / bucket logic.
// Scalar variants are registered too: no model property may reference them.
function collectUnionVariantTypes(
  typeDef: TypeDef,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  if (typeDef.Type?.toString() !== "union") return;
  const variants: TypeDef[] = [];
  if (typeDef.Discriminator?.Mapping) {
    for (const m of typeDef.Discriminator.Mapping) variants.push(m.Type);
  } else {
    for (const at of typeDef.AssociatedTypes ?? []) variants.push(at);
  }
  for (const v of variants) {
    const kind = v.Type?.toString() ?? "";
    // Event-stream variants are read by the SSE reader, not by the converter.
    if (!SERIALIZABLE_TYPES.includes(kind) && kind !== "event-stream") {
      const cs = csharpTypeForJsonContext(v).replaceAll("global::", "");
      if (!deps.has(cs)) deps.set(cs, { Type: cs, Name: jsonTypeInfoName(v) });
    }
    let t = v;
    while (
      t &&
      (t.Type?.toString() === "array" || t.Type?.toString() === "map")
    ) {
      const cs = csharpTypeForJsonContext(t).replaceAll("global::", "");
      const name = jsonTypeInfoName(t);
      if (!deps.has(cs)) deps.set(cs, { Type: cs, Name: name });
      t = t.ItemType;
    }
  }
}

function collectFormJsonContextTypes(
  typeDef: TypeDef,
  outputLocation: string,
  deps: Map<string, { Type: string; Name: string }>,
  visited: Set<string>,
): void {
  const id = `${typeDef.OutputLocation ?? ""}::${typeDef.Name}`;
  if (visited.has(id) || typeDef.Truncated) return;
  visited.add(id);

  for (const field of typeDef.Fields ?? []) {
    const formAnnotation = getFormAnnotation(field);
    const multipartAnnotation = getMultipartFormAnnotation(field);
    if (
      (formAnnotation?.JSON || multipartAnnotation?.JSON) &&
      !SERIALIZABLE_TYPES.includes(field.Type.Type?.toString() ?? "")
    ) {
      const type = csharpTypeForJsonContext(field.Type);
      if (!deps.has(type)) {
        deps.set(type, {
          Type: type,
          Name: jsonTypeInfoName(field.Type),
        });
      }
    }

    if (SERIALIZABLE_TYPES.includes(field.Type.Type?.toString() ?? "")) {
      collectFormJsonContextTypes(field.Type, outputLocation, deps, visited);
    }
  }
}

// Walks operation request-body and response-variant TypeDefs that are
// serialized as JSON, collecting primitive / array / map / empty-location
// shapes that don't otherwise live in BucketedTypes. Result is keyed by
// target output-location (the bucket whose JsonContext should register the
// shape).
function collectOperationJsonShapes(
  sdk: SDK,
  perBucket: Map<string, Map<string, { Type: string; Name: string }>>,
  visited: Set<string>,
): void {
  for (const typeDef of collectOperationJsonTypeDefs(sdk)) {
    recordOperationShape(typeDef, perBucket, visited);
  }
}

function collectOperationJsonTypeDefs(
  sdk: SDK,
  typeDefs: TypeDef[] = [],
): TypeDef[] {
  for (const op of sdk.Operations ?? []) {
    if (
      `${op.SerializationMethod ?? ""}` === "json" &&
      op.Request?.RequestBody?.Type
    ) {
      typeDefs.push(op.Request.RequestBody.Type);
    }
    for (const response of op.Response?.Responses ?? []) {
      for (const content of response.Content ?? []) {
        const method = `${content.SerializationMethod ?? ""}`;
        if (method === "json" && content.Content?.Type) {
          typeDefs.push(content.Content.Type);
        } else if (
          method === "eventstream" &&
          content.Content?.Type?.ItemType
        ) {
          typeDefs.push(content.Content.Type.ItemType);
        }
      }
    }
  }
  for (const subSDK of sdk.SubSDKs ?? []) {
    collectOperationJsonTypeDefs(subSDK, typeDefs);
  }
  return typeDefs;
}

function bucketKeyForShape(typeDef: TypeDef): string {
  const kind = typeDef.Type?.toString() ?? "";
  if (kind === "array" || kind === "map") {
    return typeDef.ItemType?.OutputLocation || "";
  }
  return typeDef.OutputLocation || "";
}

function recordOperationShape(
  typeDef: TypeDef,
  perBucket: Map<string, Map<string, { Type: string; Name: string }>>,
  visited: Set<string>,
): void {
  if (!typeDef) return;
  // Dedupe keyed off the C# type expression — uniquely identifies arrays /
  // maps / nested generics by structure rather than relying on TypeDef.Name
  // which is empty for collection wrappers and would collide.
  const id = csharpTypeForJsonContext(typeDef);
  if (visited.has(id)) return;
  visited.add(id);

  const kind = typeDef.Type?.toString() ?? "";

  const record = (bucket: string) => {
    const csharpType = csharpTypeForJsonContext(typeDef);
    const name = jsonTypeInfoName(typeDef);
    let m = perBucket.get(bucket);
    if (!m) {
      m = new Map();
      perBucket.set(bucket, m);
    }
    m.set(csharpType, { Type: csharpType, Name: name });
    collectJsonContextShapeCollisions(typeDef, m);
  };

  // Class / enum / union / error already in BucketedTypes when they have an
  // OutputLocation. Their JsonContext registration is emitted by the main
  // loop. Walk fields anyway so any nested primitive / array / map shapes
  // surface here.
  if (
    kind === "class" ||
    kind === "enum" ||
    kind === "union" ||
    kind === "error"
  ) {
    if (!typeDef.OutputLocation) {
      record("");
    }
    for (const field of typeDef.Fields ?? []) {
      recordOperationShape(field.Type, perBucket, visited);
    }
    return;
  }

  if (kind === "array" || kind === "map") {
    record(bucketKeyForShape(typeDef));
    if (typeDef.ItemType) {
      recordOperationShape(typeDef.ItemType, perBucket, visited);
    }
    return;
  }

  // Primitive scalar (string, long, decimal, bool, BigInteger, DateTime, …).
  // Bucket = "" so it lands on the top-level ModelsJsonContext.
  record("");
}

function getObjectJsonConverterType(): string {
  return `global::${getJsonContextNamespace()}.ObjectJsonConverter`;
}

// Source-gen contexts need our runtime converters baked into their `.Default`
// options so each generated JsonTypeInfo<T> resolves property-level converters
// against them. In particular, `object`-typed properties pick up
// ObjectJsonConverter and round-trip to native CLR types instead of
// JsonElement. The same list is mirrored at runtime in JsonOptions.SerializerOptions
// so direct callsites and the resolver chain agree on conversion behavior.
function getJsonContextConverters(): string[] {
  const utilsNs = getScopeNamespace("utils", true);
  return [
    getObjectJsonConverterType(),
    `global::${utilsNs}.BigIntegerJsonConverterFactory`,
    `global::${utilsNs}.DecimalJsonConverterFactory`,
    `global::${utilsNs}.DateTimeJsonConverterFactory`,
    ...(requiresNodatime()
      ? [`global::${utilsNs}.LocalDateJsonConverter`]
      : []),
  ];
}

function collectOpShapesForMainSDK(): Map<
  string,
  Map<string, { Type: string; Name: string }>
> {
  const opShapesPerBucket = new Map<
    string,
    Map<string, { Type: string; Name: string }>
  >();
  collectOperationJsonShapes(
    context.Global.AST.MainSDK,
    opShapesPerBucket,
    new Set<string>(),
  );
  return opShapesPerBucket;
}

const JSON_NUMBER_HANDLING = "AllowReadingFromString";

interface JsonContextRegistration {
  OutputLocation: string;
  Types: { Type: string; Name: string }[];
}

function uniqueByType<T extends { Type: string }>(refs: T[]): T[] {
  const seen = new Set<string>();
  return refs.filter((ref) => {
    if (seen.has(ref.Type)) return false;
    seen.add(ref.Type);
    return true;
  });
}

function collectJsonContextRegistrations(): JsonContextRegistration[] {
  const registrations: JsonContextRegistration[] = [];

  const opShapesPerBucket = collectOpShapesForMainSDK();

  const bucketedTypes = sequencedMapEntries(context.Global.AST.BucketedTypes);
  const bucketsWithRefs = new Set<string>();

  for (const [outputLocation, buckets] of bucketedTypes) {
    const typeRefs: { Type: string; Name: string }[] = [];
    const crossNsDeps = new Map<string, { Type: string; Name: string }>();
    const formJsonDeps = new Map<string, { Type: string; Name: string }>();
    const visited = new Set<string>();
    const formVisited = new Set<string>();

    for (const [, typeDefs] of sequencedMapEntries(buckets)) {
      for (const typeDef of typeDefs) {
        if (!SERIALIZABLE_TYPES.includes(typeDef.Type?.toString() ?? ""))
          continue;
        // csharpTypeForJsonContext / jsonTypeInfoName handle the error → Payload
        // DTO substitution. Registering the wrapper directly would force STJ
        // source-gen to walk RUC-annotated System.Exception surface (TargetSite,
        // etc.) and produce IL2026/IL3050 warnings under AOT.
        const cs = csharpTypeForJsonContext(typeDef).replaceAll("global::", "");
        const name = jsonTypeInfoName(typeDef);
        typeRefs.push({ Type: cs, Name: name });
        collectJsonContextCrossNsDeps(
          typeDef,
          outputLocation,
          crossNsDeps,
          visited,
        );
        collectFormJsonContextTypes(
          typeDef,
          outputLocation,
          formJsonDeps,
          formVisited,
        );
        collectUnionVariantTypes(typeDef, crossNsDeps);
        collectAdditionalPropertiesValueTypes(typeDef, crossNsDeps);
      }
    }

    for (const dep of crossNsDeps.values()) {
      typeRefs.push(dep);
    }
    for (const dep of formJsonDeps.values()) {
      typeRefs.push(dep);
    }

    // Pull operation-driven primitive / array / map shapes that should live
    // on this scope's context. Dedupe against existing typeRefs by Type.
    for (const dep of opShapesPerBucket.get(outputLocation)?.values() ?? []) {
      typeRefs.push(dep);
    }

    registrations.push({
      OutputLocation: outputLocation,
      Types: uniqueByType(typeRefs),
    });

    bucketsWithRefs.add(outputLocation);
  }

  // Emit per-bucket contexts for operation-driven shapes whose target bucket
  // had no class/enum/union/error of its own (so the main loop above didn't
  // produce a context for it). The empty bucket "" is the canonical case —
  // primitives belong to ModelsJsonContext and there's typically no
  // BucketedTypes entry at the empty location.
  for (const [outputLocation, deps] of opShapesPerBucket) {
    if (bucketsWithRefs.has(outputLocation)) continue;
    if (deps.size === 0) continue;
    registrations.push({
      OutputLocation: outputLocation,
      Types: [...deps.values()],
    });
  }

  return registrations;
}

function collectJsonTypeInfos(
  registrations: JsonContextRegistration[],
): { Type: string; Name: string }[] {
  const byName = new Map<string, { Type: string; Name: string }>();
  for (const registration of registrations) {
    for (const ref of registration.Types) {
      const existing = byName.get(ref.Name);
      if (!existing) {
        byName.set(ref.Name, ref);
        continue;
      }
      const normalize = (t: string) => t.replaceAll("global::", "");
      if (normalize(existing.Type) !== normalize(ref.Type)) {
        throw new Error(
          `JsonTypeInfo name ${ref.Name} maps to both ${existing.Type} and ${ref.Type}`,
        );
      }
    }
  }
  return [...byName.values()];
}

interface NullCheckedType {
  Type: string;
  Fields: { JsonName: string; Required: boolean }[];
}

function forEachJsonObjectType(
  visitObject: (t: TypeDef, type: string) => void,
): void {
  const visited = new Set<string>();

  const visit = (t: TypeDef | undefined) => {
    while (t?.Type?.toString() === "array" || t?.Type?.toString() === "map") {
      t = t.ItemType;
    }
    if (!t || t.Truncated) return;
    const kind = t.Type?.toString() ?? "";
    if (!SERIALIZABLE_TYPES.includes(kind)) return;
    const type = csharpTypeForJsonContext(t);
    if (visited.has(type)) return;
    visited.add(type);

    if (kind === "class" || kind === "error") visitObject(t, type);

    for (const field of t.Fields ?? []) visit(field.Type);
    for (const assoc of t.AssociatedTypes ?? []) visit(assoc);
    for (const m of t.Discriminator?.Mapping ?? []) visit(m.Type);
  };

  for (const [, buckets] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, typeDefs] of sequencedMapEntries(buckets)) {
      for (const typeDef of typeDefs) visit(typeDef);
    }
  }
  for (const typeDef of collectOperationJsonTypeDefs(
    context.Global.AST.MainSDK,
  )) {
    visit(typeDef);
  }
}

function collectNullCheckedTypes(): NullCheckedType[] {
  const result: NullCheckedType[] = [];
  forEachJsonObjectType((t, type) => {
    const fields = (t.Fields ?? [])
      .filter((f) => {
        const state = requiredState(f);
        return (
          (state === "DisallowNull" ||
            (state === "Always" && !isValueType(f.Type))) &&
          v2RequiredEligible(f)
        );
      })
      .map((f) => ({
        JsonName: jsonFieldNameLiteral(getJsonAnnotation(f)!),
        Required: requiredState(f) === "Always",
      }));
    if (fields.length > 0) result.push({ Type: type, Fields: fields });
  });
  return result;
}

interface ConstFieldsType {
  Type: string;
  Fields: { JsonName: string; Value: string; IntegralToken: string }[];
}

const STRING_WIRE_CONST_TYPES = ["string", "date", "date-time"];
const NUMBER_WIRE_CONST_TYPES = [
  "integer",
  "int32",
  "bigint",
  "number",
  "float32",
  "decimal",
];
const INTEGRAL_CONST_TYPES = ["integer", "int32", "bigint"];
const JSON_NUMBER = /^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$/;

function csharpStringLiteralContent(value: string): string {
  return value.replace(/[\\"\u0000-\u001f\u0085\u2028\u2029]/g, (c) =>
    c === "\\" || c === '"'
      ? `\\${c}`
      : `\\u${c.charCodeAt(0).toString(16).padStart(4, "0")}`,
  );
}

function constWireToken(field: FieldDef): { json: string; integral: string } {
  const value = field.Const!.Value;
  if (value === null || value === undefined)
    return { json: "null", integral: "" };

  let kind = field.Type.Type?.toString() ?? "";
  if (kind === "enum") kind = field.Type.Enum?.Type?.Type?.toString() ?? "";
  if (kind === "any") {
    kind =
      typeof value === "boolean"
        ? "boolean"
        : typeof value === "number"
        ? "number"
        : "string";
  }
  const stringWire =
    STRING_WIRE_CONST_TYPES.includes(kind) ||
    ((kind === "bigint" || kind === "decimal") &&
      field.Type.Format === "string");

  if (stringWire && typeof value !== "object") {
    return { json: JSON.stringify(`${value}`), integral: "" };
  }
  if (kind === "boolean" && typeof value === "boolean") {
    return { json: `${value}`, integral: "" };
  }
  if (NUMBER_WIRE_CONST_TYPES.includes(kind) && typeof value !== "object") {
    const json = `${value}`;
    if (JSON_NUMBER.test(json)) {
      const integral =
        INTEGRAL_CONST_TYPES.includes(kind) && /^-?\d+$/.test(json)
          ? json.replace(/^-0$/, "0")
          : "";
      return { json, integral };
    }
  }
  throw new Error(
    `unsupported const value '${JSON.stringify(
      value,
    )}' of type '${kind}' for field '${field.OriginalName}'`,
  );
}

function collectConstFieldsTypes(): ConstFieldsType[] {
  const result: ConstFieldsType[] = [];
  forEachJsonObjectType((t, type) => {
    const fields = (t.Fields ?? [])
      .filter((f) => {
        const annotation = getJsonAnnotation(f);
        return (
          f.Const !== undefined &&
          f.Const !== null &&
          annotation !== undefined &&
          !annotation.Ignore &&
          !f.IsAdditionalProperties
        );
      })
      .map((f) => {
        const token = constWireToken(f);
        return {
          JsonName: jsonFieldNameLiteral(getJsonAnnotation(f)!),
          Value: csharpStringLiteralContent(token.json),
          IntegralToken: token.integral ? `"${token.integral}"` : "null",
        };
      });
    if (fields.length > 0) result.push({ Type: type, Fields: fields });
  });
  return result;
}

function getJsonContextJobs(registrations: JsonContextRegistration[]): Job[] {
  return registrations.map(({ OutputLocation, Types }) => {
    const className = getJsonContextClassName(OutputLocation);
    return createTemplateFileJob(
      "json-context.cs.stmpl",
      `${getSDKTopLevelFolder()}/${getJsonContextLocation(
        OutputLocation,
      )}/${className}.cs`,
      {
        Types,
        Namespace: getJsonContextNamespace(OutputLocation),
        ClassName: className,
        Converters: getJsonContextConverters(),
        NumberHandling: JSON_NUMBER_HANDLING,
      },
    );
  });
}

function getJsonOptionsJobs(registrations: JsonContextRegistration[]): Job[] {
  const resolvers = registrations.map(
    ({ OutputLocation }) =>
      `${getJsonContextNamespace(OutputLocation)}.${getJsonContextClassName(
        OutputLocation,
      )}.Default`,
  );

  const rootLocation = getJsonContextLocation("");
  const rootNamespace = getJsonContextNamespace();

  return [
    createTemplateFileJob(
      "json-options.cs.stmpl",
      `${getSDKTopLevelFolder()}/${rootLocation}/JsonOptions.cs`,
      {
        Resolvers: resolvers.join(",\n            "),
        Namespace: rootNamespace,
        Converters: getJsonContextConverters().map((c) =>
          c === getObjectJsonConverterType()
            ? "_objectConverter"
            : `new ${c}()`,
        ),
        NumberHandling: JSON_NUMBER_HANDLING,
        NullCheckedTypes: collectNullCheckedTypes(),
        ConstFieldsTypes: collectConstFieldsTypes(),
        AdditionalPropertiesTypes: collectAdditionalPropertiesTypes(),
        HasOpenUnion: hasAnyOpenUnion(),
      },
    ),
    createTemplateFileJob(
      "json-type-infos.cs.stmpl",
      `${getSDKTopLevelFolder()}/${rootLocation}/JsonTypeInfos.cs`,
      {
        Types: collectJsonTypeInfos(registrations),
        Namespace: rootNamespace,
      },
    ),
    createTemplateFileJob(
      "object-json-converter.cs.stmpl",
      `${getSDKTopLevelFolder()}/${rootLocation}/ObjectJsonConverter.cs`,
      {
        Namespace: rootNamespace,
      },
    ),
    createTemplateFileJob(
      "json-serialization.cs.stmpl",
      `${getSDKTopLevelFolder()}/${rootLocation}/JsonSerialization.cs`,
      {
        Namespace: rootNamespace,
      },
    ),
  ];
}

interface UnionVariant {
  FieldName: string; // PascalCase property name (+ "Payload" for errors)
  FieldType: string; // C# type expression for property declaration
  CsType: string; // FQN for JsonTypeInfo<T> generic arg
  TypeInfo: string; // JsonTypeInfo<CsType> expression over the factory's `options`
  DiscriminatorValues: string[]; // disc-mapping names that resolve to this variant; empty for undiscriminated
}

// A union variant has no property to carry a string-wire converter attribute, so
// `format: string` bigint/decimal variants get a JsonTypeInfo built over the
// string-wire converter instead of the options' default number-wire one.
function unionVariantTypeInfo(variantType: TypeDef, csType: string): string {
  const factory = getStjConverterFactory(variantType);
  if (!factory.endsWith("StringJsonConverterFactory")) {
    return `(System.Text.Json.Serialization.Metadata.JsonTypeInfo<${csType}>)options.GetTypeInfo(typeof(${csType}))!`;
  }
  return `System.Text.Json.Serialization.Metadata.JsonMetadataServices.CreateValueInfo<${csType}>(options, new ${factory}().CreateConverter(typeof(${csType}), options)!)`;
}

// Single source of truth for union-variant derivation. Consumed by
// union-discr.cs.stmpl / union-undiscr.cs.stmpl for property + factory-param
// emission.
//
// csharpTypeForJsonContext returns class/union/error names qualified with
// their model namespace (no `global::`) but emits non-aliased primitives
// bare (e.g. `BigInteger`, `DateTime`). The JsonOptions template adds
// matching `using` directives so bare names resolve. Variant types keep
// nullable reference items (`List<Pet?>`) to match the model property type.
// @ts-ignore
function getUnionVariants(typeDef: TypeDef): UnionVariant[] {
  const isErr = isUnionOfErrors(typeDef);
  const result: UnionVariant[] = [];
  const byField = new Map<string, UnionVariant>();

  const addOrMerge = (variantType: TypeDef, discName: string | null) => {
    let fieldName = sanitizeFieldName(sanitizeUnionTypeName(variantType));
    let fieldType = sanitizeType(
      variantType,
      false,
      variantType.OutputLocation,
    );
    if (isErr) {
      fieldName += "Payload";
      fieldType += "Payload";
    }
    let entry = byField.get(fieldName);
    if (!entry) {
      const csType = csharpTypeForJsonContext(variantType, true);
      entry = {
        FieldName: fieldName,
        FieldType: fieldType,
        CsType: csType,
        TypeInfo: unionVariantTypeInfo(variantType, csType),
        DiscriminatorValues: [],
      };
      byField.set(fieldName, entry);
      result.push(entry);
    }
    if (discName !== null) entry.DiscriminatorValues.push(discName);
  };

  if (typeDef.Discriminator?.Mapping) {
    for (const m of typeDef.Discriminator.Mapping) addOrMerge(m.Type, m.Name);
  } else {
    for (const at of typeDef.AssociatedTypes ?? []) addOrMerge(at, null);
  }
  return result;
}
registerTemplateFunc("getUnionVariants", getUnionVariants);
