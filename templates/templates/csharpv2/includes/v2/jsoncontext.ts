function getJsonContextBucketPrefix(
  outputLocation: string,
  modelNamespace: string = "",
): string {
  if (outputLocation) {
    const lastSegment = outputLocation.split("/").pop();
    if (lastSegment) {
      return sanitizeClassName(lastSegment);
    }
  }

  return modelNamespace.split(".").pop() || getScopePath("models");
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
// deps so the caller can add them as explicit entries in the context file.
function collectJsonContextCrossNsDeps(
  typeDef: TypeDef,
  outputLocation: string,
  deps: Map<string, { Type: string; Name: string }>,
  visited: Set<string>,
): void {
  const id = `${typeDef.OutputLocation ?? ""}::${typeDef.Name}`;
  if (visited.has(id) || typeDef.Truncated) return;
  visited.add(id);

  const visit = (t: TypeDef) => {
    while (t.Type === "array" || t.Type === "map") {
      t = t.ItemType;
      if (!t) return;
    }
    if (!SERIALIZABLE_TYPES.includes(t.Type?.toString() ?? "")) return;

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

  for (const field of typeDef.Fields ?? []) visit(field.Type);
  for (const assoc of typeDef.AssociatedTypes ?? []) visit(assoc);

  // Each JsonContext expands converters using its own resolver in isolation.
  // When a union flows into this bucket via cross-NS, the bucket also needs
  // every collection-wrapper variant the union's factory will request via
  // `options.GetTypeInfo` — otherwise STJ throws NotSupportedException at
  // expansion time. Register them locally even though their element types are
  // already covered by `visit` above.
  collectUnionVariantCollectionWrappers(typeDef, deps);
}

// Union converter factories request `JsonTypeInfo<List<TVariant>>` (or
// `Dictionary<...>`) via `options.GetTypeInfo` when a variant is itself a
// collection. STJ source-generated contexts only return metadata for
// explicitly declared [JsonSerializable] types — without a wrapper entry, the
// runtime throws NotSupportedException. Walk a union's variants and emit the
// collection wrapper alongside the element type. Element type is already
// reachable via the existing crossNsDeps / bucket logic.
function collectUnionVariantCollectionWrappers(
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
// shape), so the registration ends up where `jsonTypeInfoRef` will look for
// it at the call site.
function collectOperationJsonShapes(
  sdk: SDK,
  perBucket: Map<string, Map<string, { Type: string; Name: string }>>,
  visited: Set<string>,
): void {
  for (const op of sdk.Operations ?? []) {
    if (
      `${op.SerializationMethod ?? ""}` === "json" &&
      op.Request?.RequestBody?.Type
    ) {
      recordOperationShape(op.Request.RequestBody.Type, perBucket, visited);
    }
    for (const response of op.Response?.Responses ?? []) {
      for (const content of response.Content ?? []) {
        const method = `${content.SerializationMethod ?? ""}`;
        if (method === "json" && content.Content?.Type) {
          recordOperationShape(content.Content.Type, perBucket, visited);
        } else if (
          method === "eventstream" &&
          content.Content?.Type?.ItemType
        ) {
          recordOperationShape(
            content.Content.Type.ItemType,
            perBucket,
            visited,
          );
        }
      }
    }
  }
  for (const subSDK of sdk.SubSDKs ?? []) {
    collectOperationJsonShapes(subSDK, perBucket, visited);
  }
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

// Source-gen contexts need our runtime converters baked into their `.Default`
// options so each generated JsonTypeInfo<T> resolves property-level converters
// against them. In particular, `object`-typed properties pick up
// ObjectJsonConverter and round-trip to native CLR types instead of
// JsonElement. The same list is mirrored at runtime in JsonOptions.SerializerOptions
// so direct callsites and the resolver chain agree on conversion behavior.
function getJsonContextConverters(): string[] {
  const rootNs = getJsonContextNamespace();
  const utilsNs = getScopeNamespace("utils", true);
  return [
    `global::${rootNs}.ObjectJsonConverter`,
    `global::${utilsNs}.BigIntegerJsonConverterFactory`,
    `global::${utilsNs}.DecimalJsonConverterFactory`,
    `global::${utilsNs}.DateTimeJsonConverterFactory`,
  ];
}

// @ts-ignore
function getJsonContextJobs(): Job[] {
  const jobs: Job[] = [];

  const pushContextJob = (
    outputLocation: string,
    types: { Type: string; Name: string }[],
  ) => {
    const className = getJsonContextClassName(outputLocation);
    jobs.push(
      createTemplateFileJob(
        "json-context.cs.stmpl",
        `${getSDKTopLevelFolder()}/${getJsonContextLocation(
          outputLocation,
        )}/${className}.cs`,
        {
          Types: types,
          Namespace: getJsonContextNamespace(outputLocation),
          ClassName: className,
          Converters: getJsonContextConverters(),
        },
      ),
    );
  };

  const opShapesPerBucket = new Map<
    string,
    Map<string, { Type: string; Name: string }>
  >();
  const opShapesVisited = new Set<string>();
  collectOperationJsonShapes(
    context.Global.AST.MainSDK,
    opShapesPerBucket,
    opShapesVisited,
  );

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
        collectUnionVariantCollectionWrappers(typeDef, crossNsDeps);
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
    const existingTypes = new Set(typeRefs.map((r) => r.Type));
    for (const dep of opShapesPerBucket.get(outputLocation)?.values() ?? []) {
      if (!existingTypes.has(dep.Type)) typeRefs.push(dep);
    }

    pushContextJob(outputLocation, typeRefs);

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
    pushContextJob(outputLocation, [...deps.values()]);
  }

  return jobs;
}

// @ts-ignore
function getJsonOptionsJobs(): Job[] {
  const contexts: { Namespace: string; ClassName: string }[] = [];
  const bucketedTypes = sequencedMapEntries(context.Global.AST.BucketedTypes);

  const seenLocations = new Set<string>();
  for (const [outputLocation] of bucketedTypes) {
    contexts.push({
      Namespace: getJsonContextNamespace(outputLocation),
      ClassName: getJsonContextClassName(outputLocation),
    });
    seenLocations.add(outputLocation);
  }

  // Operation-driven primitives / array / map shapes may produce additional
  // contexts (most notably the top-level ModelsJsonContext for empty bucket).
  // Register them as resolvers too so JsonOptions.SerializerOptions can find
  // their JsonTypeInfo<T> entries at runtime.
  const opShapesPerBucket = new Map<
    string,
    Map<string, { Type: string; Name: string }>
  >();
  collectOperationJsonShapes(
    context.Global.AST.MainSDK,
    opShapesPerBucket,
    new Set<string>(),
  );
  for (const [outputLocation, deps] of opShapesPerBucket) {
    if (deps.size === 0 || seenLocations.has(outputLocation)) continue;
    contexts.push({
      Namespace: getJsonContextNamespace(outputLocation),
      ClassName: getJsonContextClassName(outputLocation),
    });
    seenLocations.add(outputLocation);
  }

  const rootLocation = getJsonContextLocation("");
  const rootNamespace = getJsonContextNamespace();

  return [
    createTemplateFileJob(
      "json-options.cs.stmpl",
      `${getSDKTopLevelFolder()}/${rootLocation}/JsonOptions.cs`,
      {
        Contexts: contexts,
        Resolvers: contexts
          .map((ctx) => `${ctx.Namespace}.${ctx.ClassName}.Default`)
          .join(",\n            "),
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
  ];
}

interface UnionVariant {
  FieldName: string; // PascalCase property name (+ "Payload" for errors)
  FieldType: string; // C# type expression for property declaration
  CsType: string; // FQN for JsonTypeInfo<T> generic arg
  DiscriminatorValues: string[]; // disc-mapping names that resolve to this variant; empty for undiscriminated
}

// Single source of truth for union-variant derivation. Consumed by
// union-discr.cs.stmpl / union-undiscr.cs.stmpl for property + factory-param
// emission.
//
// csharpTypeForJsonContext returns class/union/error FQNs already
// qualified with `global::` but emits non-aliased primitives bare
// (e.g. `BigInteger`, `DateTime`). The JsonOptions template adds
// matching `using` directives so bare names resolve.
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
      entry = {
        FieldName: fieldName,
        FieldType: fieldType,
        CsType: csharpTypeForJsonContext(variantType),
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
