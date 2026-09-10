// Helpers for the `methodSignature: params-object` configuration: SDK class
// methods keep required path parameters positional and gather all remaining
// arguments into a generated `<Operation>Params` object, while the underlying
// standalone functions keep their stable positional shape.

type TSMethodParamsBodyVariant = {
  ParamsName: string;
  MemberName: string;
  MemberImportPath: string;
};

type TSMethodParamsState = {
  ParamsName: string;
  RequestName: string;
  RequestImportPath: string;
  OmitKeys: string[];
  SpreadBodyKey: string;
  // For union-typed spread bodies: one params type per union member
  // (`CreateModelInteractionParams`), mirroring the per-variant overloads
  // used by compatibility-focused SDK surfaces for narrowing and autocomplete.
  BodyVariants: TSMethodParamsBodyVariant[];
  // The spread body field is optional on the request: its fields intersect
  // as Partial<...> and an empty gathered body maps back to undefined at
  // runtime so the request omits the body like the positional shape does.
  BodyOptional: boolean;
  WrapperStyle: boolean;
  ParamsRequired: boolean;
  ParamMemberNames: string[];
  SSE: boolean;
};

// Relative import path from models/operations/method-params.ts to the file
// declaring the request model, which is a shared component model for
// fully-flattened bodies.
function tsMethodParamsRequestImportPath(reqType: TypeDef): string {
  return `${getImportPrefix(
    getModelsLocation(reqType.OutputLocation),
    getOperationsLocation(),
  )}/${resolveModelName(reqType)}.js`;
}

function tsParamsObjectEnabled(): boolean {
  return (
    (context.Global.Config.MethodSignature ?? "positional") === "params-object"
  );
}
registerTemplateFunc("tsParamsObjectEnabled", tsParamsObjectEnabled);

function tsOpUsesParamsObject(op: Operation): boolean {
  if (!tsParamsObjectEnabled() || op.Webhook) {
    return false;
  }
  const reqType = op.Request?.Field?.Type;
  if (!reqType || !reqType.IsCustomType()) {
    return false;
  }
  // The params object derives from the rendered request model via Omit, so
  // there must be at least one non-positional argument to gather.
  return (op.Arguments?.Sorted ?? []).length > 0;
}

function tsRequiredPathParamNames(op: Operation): Set<string> {
  const names = new Set<string>();
  for (const param of op.Request?.Params?.PathParams ?? []) {
    if (param.Field && !param.Field.Optional) {
      names.add(sanitizeFieldName(param.Field.Name));
    }
  }
  return names;
}

function tsMethodPositionalFields(op: Operation): FieldDef[] {
  const positional = tsRequiredPathParamNames(op);
  return (op.Arguments?.Sorted ?? []).filter((f) =>
    positional.has(sanitizeFieldName(f.Name)),
  );
}
registerTemplateFunc("tsMethodPositionalFields", tsMethodPositionalFields);

// Raw per-operation computation. Templates resolve through
// tsMethodParamsState below, which returns the canonical (file-level
// deduplicated) state so emitted types and overloads always agree.
function tsMethodParamsStateRaw(
  op: Operation,
): TSMethodParamsState | undefined {
  if (!tsOpUsesParamsObject(op)) {
    return undefined;
  }

  const reqType = op.Request!.Field!.Type!;
  const requestName = sanitizeClassName(reqType.Name);
  const paramsName = requestName.endsWith("Request")
    ? requestName.slice(0, -"Request".length) + "Params"
    : requestName + "Params";

  const positionalFields = tsMethodPositionalFields(op);
  const positional = new Set(
    positionalFields.map((f) => sanitizeFieldName(f.Name)),
  );
  // Omit keys and member access go through the request model, so they use
  // the model property naming rather than the method argument naming.
  const omitKeys: string[] = positionalFields
    .map((f) => sanitizeModelField(f))
    .sort();
  let spreadBodyKey = "";
  let bodyVariants: TSMethodParamsBodyVariant[] = [];
  let bodyOptional = false;
  let wrapperStyle = false;
  let bodyAsMember = false;
  const memberNames: string[] = [];
  let paramsRequired = false;

  for (const field of op.Arguments?.Sorted ?? []) {
    const name = sanitizeFieldName(field.Name);
    if (positional.has(name)) {
      continue;
    }
    // Per-operation security arguments are not part of the request model the
    // params object derives from; such methods keep the positional shape.
    if (isSecurityClassField(field)) {
      return undefined;
    }
    if (!field.Optional) {
      paramsRequired = true;
    }
    if (
      hasAnnotation(field, "request") ||
      hasAnnotation(field, "requestWrapper")
    ) {
      if (
        field.Type === reqType ||
        sanitizeClassName(field.Type.Name) === requestName
      ) {
        // The argument is the request model itself (unflattened request or
        // fully-flattened body): the params object is the request model.
        wrapperStyle = true;
        continue;
      }
      // Spreading the body into the params object is only sound for object
      // shapes: classes, and unions whose members are all classes. Anything
      // else (string/binary/array bodies, mixed unions) stays a nested
      // member.
      const bodyKind = field.Type.Type.toString();
      const unionMembers =
        bodyKind === "union" ? field.Type.AssociatedTypes ?? [] : [];
      const allClassUnion =
        unionMembers.length > 1 &&
        unionMembers.every(
          (member) =>
            member.IsCustomType() && member.Type.toString() === "class",
        );
      if (bodyKind === "class" || allClassUnion) {
        spreadBodyKey = sanitizeModelField(field);
        bodyOptional = Boolean(field.Optional);
        omitKeys.push(spreadBodyKey);
        if (allClassUnion) {
          bodyVariants = unionMembers.map((member) => ({
            ParamsName: `${sanitizeClassName(member.Name)}Params`,
            MemberName: sanitizeClassName(member.Name),
            MemberImportPath: `${getImportPrefix(
              getModelsLocation(member.OutputLocation),
              getOperationsLocation(),
            )}/${resolveModelName(member)}.js`,
          }));
        }
        continue;
      }
      bodyAsMember = true;
      memberNames.push(sanitizeModelField(field));
      continue;
    }
    memberNames.push(sanitizeModelField(field));
  }

  if (wrapperStyle) {
    if (memberNames.length > 0 || spreadBodyKey !== "" || omitKeys.length > 0) {
      // Mixed wrapper + exploded arguments are not expressible as a single
      // params object; fall back to the positional signature for this method.
      return undefined;
    }
    if (opHasSSEOverload(op)) {
      // The SSE overload variants intersect a top-level `stream` member,
      // which a wrapper-style request does not necessarily have; fall back
      // to the positional signature.
      return undefined;
    }
    return {
      ParamsName: paramsName,
      RequestName: requestName,
      RequestImportPath: tsMethodParamsRequestImportPath(reqType),
      OmitKeys: [],
      SpreadBodyKey: "",
      BodyVariants: [],
      BodyOptional: false,
      WrapperStyle: true,
      ParamsRequired: paramsRequired,
      ParamMemberNames: [],
      SSE: false,
    };
  }

  if (memberNames.length === 0 && spreadBodyKey === "") {
    return undefined;
  }

  const sse = opHasSSEOverload(op);
  if (sse && bodyAsMember) {
    // The SSE overload variants intersect a top-level `stream` member; with
    // the body kept as a nested member, `stream` lives inside it, so fall
    // back to the positional signature.
    return undefined;
  }

  return {
    ParamsName: paramsName,
    RequestName: requestName,
    RequestImportPath: tsMethodParamsRequestImportPath(reqType),
    OmitKeys: omitKeys,
    SpreadBodyKey: spreadBodyKey,
    BodyVariants: bodyVariants,
    BodyOptional: bodyOptional,
    WrapperStyle: false,
    ParamsRequired: paramsRequired,
    ParamMemberNames: memberNames,
    SSE: sse,
  };
}
function tsMethodParamsState(op: Operation): TSMethodParamsState | undefined {
  const raw = tsMethodParamsStateRaw(op);
  if (!raw) {
    return undefined;
  }
  for (const state of tsMethodParamsFileStates()) {
    if (state.ParamsName === raw.ParamsName) {
      return state;
    }
  }
  return raw;
}
registerTemplateFunc("tsMethodParamsState", tsMethodParamsState);

function tsIsSafeIdentifier(name: string): boolean {
  return /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(name);
}

function tsParamsMemberLocal(op: Operation, modelName: string): string {
  // Locals mirror the method-argument naming; model property keys that are
  // not valid identifiers destructure through a computed key.
  for (const field of op.Arguments?.Sorted ?? []) {
    if (sanitizeModelField(field) === modelName) {
      return sanitizeFieldName(field.Name);
    }
  }
  return modelName;
}

// Destructuring prelude for the method body: splits param-only members from
// the request body spread when the body was gathered into the params object.
// The spread body must stay undefined when the caller omitted the params
// object entirely.
function tsMethodCallPrelude(op: Operation): string {
  const state = tsMethodParamsState(op);
  if (!state || state.SpreadBodyKey === "") {
    return "";
  }
  const members = state.ParamMemberNames.map((modelName) => {
    const local = tsParamsMemberLocal(op, modelName);
    if (modelName === local && tsIsSafeIdentifier(modelName)) {
      return `${modelName},`;
    }
    return `["${modelName}"]: ${local},`;
  }).join(" ");
  const rest = `${state.SpreadBodyKey}$body`;
  // An optional body that gathered no fields maps back to undefined so the
  // request omits the body exactly like the positional shape would.
  const emptyIsUndefined = state.BodyOptional
    ? ` || Object.keys(${rest}).length === 0`
    : "";
  if (state.ParamsRequired) {
    if (!state.BodyOptional) {
      return `const { ${members} ...${state.SpreadBodyKey} } = params;`;
    }
    return [
      `const { ${members} ...${rest} } = params;`,
      `const ${state.SpreadBodyKey} = Object.keys(${rest}).length === 0 ? undefined : ${rest};`,
    ].join("\n    ");
  }
  return [
    `const { ${members} ...${rest} } = params ?? {};`,
    `const ${state.SpreadBodyKey} = params === undefined${emptyIsUndefined} ? undefined : ${rest};`,
  ].join("\n    ");
}
registerTemplateFunc("tsMethodCallPrelude", tsMethodCallPrelude);

function tsMethodCallArg(op: Operation, field: FieldDef): string {
  const state = tsMethodParamsState(op);
  const argName = sanitizeFieldName(field.Name);
  if (!state) {
    return argName;
  }
  if (state.WrapperStyle) {
    return "params";
  }
  if (tsRequiredPathParamNames(op).has(argName)) {
    return argName;
  }
  const modelName = sanitizeModelField(field);
  if (state.SpreadBodyKey !== "") {
    // the prelude destructured every member into a local
    return tsParamsMemberLocal(op, modelName);
  }
  if (!tsIsSafeIdentifier(modelName)) {
    return state.ParamsRequired
      ? `params["${modelName}"]`
      : `params?.["${modelName}"]`;
  }
  return state.ParamsRequired ? `params.${modelName}` : `params?.${modelName}`;
}
registerTemplateFunc("tsMethodCallArg", tsMethodCallArg);

// Collects the params-object states for every operation, for the generated
// models/operations/method-params.ts module.
function tsMethodParamsFileStates(): TSMethodParamsState[] {
  if (!tsParamsObjectEnabled()) {
    return [];
  }
  if (context.GlobalComputed?._tsMethodParamsFileStates !== undefined) {
    return context.GlobalComputed._tsMethodParamsFileStates;
  }
  context.GlobalComputed ??= {};

  const byName = new Map<string, TSMethodParamsState>();
  for (const op of iterateOperations()) {
    const state = tsMethodParamsStateRaw(op);
    if (!state) {
      continue;
    }
    const existing = byName.get(state.ParamsName);
    if (!existing) {
      byName.set(state.ParamsName, state);
    } else if (state.SSE && !existing.SSE) {
      // Multiple operations can share a request type (e.g. content-type
      // splits); keep the streaming variants if any of them needs them.
      existing.SSE = true;
    }
  }

  // Per-variant aliases share the flat name space of the module: a variant
  // colliding with an operation's params name (or another op's variant)
  // would emit duplicate declarations, so that operation falls back to the
  // union-only overloads.
  const claimed = new Set<string>();
  for (const state of byName.values()) {
    claimed.add(state.ParamsName);
    if (state.SSE) {
      claimed.add(`${state.ParamsName}NonStreaming`);
      claimed.add(`${state.ParamsName}Streaming`);
    }
  }
  for (const state of byName.values()) {
    if (state.BodyVariants.length === 0) {
      continue;
    }
    const variantNames = state.BodyVariants.flatMap((variant) =>
      state.SSE
        ? [
            variant.ParamsName,
            `${variant.ParamsName}NonStreaming`,
            `${variant.ParamsName}Streaming`,
          ]
        : [variant.ParamsName],
    );
    if (variantNames.some((name) => claimed.has(name))) {
      state.BodyVariants = [];
      continue;
    }
    variantNames.forEach((name) => claimed.add(name));
  }

  const states = [...byName.values()].sort((a, b) =>
    a.ParamsName.localeCompare(b.ParamsName),
  );
  context.GlobalComputed._tsMethodParamsFileStates = states;
  return states;
}
registerTemplateFunc("tsMethodParamsFileStates", tsMethodParamsFileStates);

// Unique union-member imports needed by the method-params module for the
// per-variant params types.
function tsMethodParamsVariantImports(): TSMethodParamsBodyVariant[] {
  const byName = new Map<string, TSMethodParamsBodyVariant>();
  for (const state of tsMethodParamsFileStates()) {
    for (const variant of state.BodyVariants) {
      byName.set(variant.MemberName, variant);
    }
  }
  return [...byName.values()].sort((a, b) =>
    a.MemberName.localeCompare(b.MemberName),
  );
}
registerTemplateFunc(
  "tsMethodParamsVariantImports",
  tsMethodParamsVariantImports,
);

// Registers type-only imports for the params object types used by a method,
// importing them directly from the generated method-params module so the
// signature works with and without index modules.
function tsImportMethodParamsTypes(
  state: TSMethodParamsState,
  usageLocation: string,
): string {
  const path = `${getImportPrefix(
    getOperationsLocation(),
    usageLocation,
  )}/method-params.js`;
  addImport(path, state.ParamsName, typeImport);
  if (state.SSE) {
    // With per-variant overloads the union NonStreaming/Streaming aliases
    // are not referenced by the method.
    if (state.BodyVariants.length > 0) {
      for (const variant of state.BodyVariants) {
        addImport(path, `${variant.ParamsName}NonStreaming`, typeImport);
        addImport(path, `${variant.ParamsName}Streaming`, typeImport);
      }
    } else {
      addImport(path, `${state.ParamsName}NonStreaming`, typeImport);
      addImport(path, `${state.ParamsName}Streaming`, typeImport);
    }
  }
  return "";
}
registerTemplateFunc("tsImportMethodParamsTypes", tsImportMethodParamsTypes);
