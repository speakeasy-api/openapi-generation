const UNKNOWN_FIELDS_MEMBER_NAME = "UnknownFields";
const UNKNOWN_FIELDS_OWNER_KINDS = ["class", "union", "error"];

// @ts-ignore
function preserveUnknownFields(): boolean {
  const value: unknown = context.Global.Config.PreserveUnknownFields;
  return value === true || value === "true";
}

unregisterTemplateFunc("preserveUnknownFields");
registerTemplateFunc("preserveUnknownFields", preserveUnknownFields);

let _unknownFieldsTypes: Set<string> | undefined;

function getUnknownFieldsTypes(): Set<string> {
  if (_unknownFieldsTypes) return _unknownFieldsTypes;

  const definitions = new Map<string, TypeDef>();
  for (const [, buckets] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, typeDefs] of sequencedMapEntries(buckets)) {
      for (const typeDef of typeDefs) {
        if (
          UNKNOWN_FIELDS_OWNER_KINDS.includes(typeDef.Type?.toString() ?? "") &&
          !typeDef.Truncated
        ) {
          definitions.set(csharpTypeForJsonContext(typeDef), typeDef);
        }
      }
    }
  }

  const roots: TypeDef[] = collectOperationJsonTypeDefs(
    context.Global.AST.MainSDK,
  );
  const webhooks = getWebhookOperations();
  roots.push(
    ...collectOperationJsonTypeDefs({
      Operations: webhooks,
      SubSDKs: [],
    } as unknown as SDK),
  );
  const envelopes = new Set<string>();
  const operations = collectUnknownFieldsOperations(context.Global.AST.MainSDK);
  for (const op of [...operations, ...webhooks]) {
    collectUnknownFieldsOperationRoots(op, roots, envelopes);
  }
  for (const op of operations) {
    for (const callback of op.Callbacks ?? []) {
      if (callback.ResponseEnvelope) {
        for (const field of callback.Fields ?? []) {
          if (!getJsonAnnotation(field)?.Ignore) {
            roots.push(field.Type);
          }
        }
      } else {
        roots.push(callback);
      }
    }
  }

  const eligible = new Set<string>();
  const visited = new Set<string>();
  const visit = (typeDef: TypeDef | undefined) => {
    while (
      typeDef &&
      (typeDef.Type?.toString() === "array" ||
        typeDef.Type?.toString() === "map")
    ) {
      typeDef = typeDef.ItemType;
    }
    const kind = typeDef?.Type?.toString() ?? "";
    if (!typeDef || !UNKNOWN_FIELDS_OWNER_KINDS.includes(kind)) return;

    const key = csharpTypeForJsonContext(typeDef);
    if (visited.has(key)) return;
    visited.add(key);

    const definition = definitions.get(key) ?? typeDef;
    if (kind === "union") {
      for (const associated of definition.AssociatedTypes ?? []) {
        visit(associated);
      }
      for (const mapping of definition.Discriminator?.Mapping ?? []) {
        visit(mapping.Type);
      }
      return;
    }

    const fields = definition.Fields ?? [];
    if (!envelopes.has(key) && !fields.some((f) => f.IsAdditionalProperties)) {
      eligible.add(key);
    }
    for (const field of fields) {
      if (getJsonAnnotation(field)?.Ignore && !field.IsAdditionalProperties) {
        continue;
      }
      visit(field.Type);
    }
  };
  for (const root of roots) {
    visit(root);
  }

  _unknownFieldsTypes = eligible;
  return eligible;
}

function collectUnknownFieldsOperations(
  sdk: SDK,
  operations: Operation[] = [],
): Operation[] {
  operations.push(...(sdk.Operations ?? []));
  for (const subSDK of sdk.SubSDKs ?? []) {
    collectUnknownFieldsOperations(subSDK, operations);
  }
  return operations;
}

function collectUnknownFieldsOperationRoots(
  op: Operation,
  roots: TypeDef[],
  envelopes: Set<string>,
): void {
  for (const field of op.Request?.Field?.Type?.Fields ?? []) {
    if (getParamAnnotation(field)?.Serialization === "json") {
      roots.push(field.Type);
    }
  }

  const method = `${op.SerializationMethod ?? ""}`;
  if (
    (method === "form" || method === "multipart") &&
    op.Request?.RequestBody?.Type
  ) {
    collectJsonEncodedFieldRoots(op.Request.RequestBody.Type, roots, new Set());
  }

  for (const response of op.Response?.Responses ?? []) {
    for (const content of response.Content ?? []) {
      if (`${content.SerializationMethod ?? ""}` !== "eventstream") continue;
      const envelope = content.Content?.Type?.ItemType;
      if (!envelope) continue;
      const envelopeClasses =
        envelope.Type?.toString() === "union"
          ? envelope.AssociatedTypes ?? []
          : [envelope];
      for (const envelopeClass of envelopeClasses) {
        if (envelopeClass.Type?.toString() === "class") {
          envelopes.add(csharpTypeForJsonContext(envelopeClass));
        }
      }
    }
  }
}

function collectJsonEncodedFieldRoots(
  typeDef: TypeDef,
  roots: TypeDef[],
  visited: Set<string>,
): void {
  if (typeDef.Type?.toString() !== "class") return;
  const key = csharpTypeForJsonContext(typeDef);
  if (visited.has(key)) return;
  visited.add(key);

  for (const field of typeDef.Fields ?? []) {
    if (
      getFormAnnotation(field)?.JSON ||
      getMultipartFormAnnotation(field)?.JSON
    ) {
      roots.push(field.Type);
    } else {
      collectJsonEncodedFieldRoots(field.Type, roots, visited);
    }
  }
}

function unknownFieldsMemberName(typeDef: TypeDef): string {
  const className = sanitizeClassName(typeDef.Name, false);
  const taken = new Set<string>([
    typeDef.Type?.toString() === "error" ? `${className}Payload` : className,
  ]);
  if (typeDef.Extensions?.Pagination) taken.add("Next");

  for (const field of typeDef.Fields ?? []) {
    const fieldName = sanitizeFieldName(field.Name, className);
    taken.add(fieldName);
    const wireName = getJsonAnnotation(field)?.FieldName;
    if (wireName) taken.add(wireName);
    if (field.Type?.Type?.toString() === "union") {
      for (const mapping of field.Type.Discriminator?.Mapping ?? []) {
        taken.add(
          `Get${fieldName}${sanitizeFieldName(mapping.Name, className)}`,
        );
      }
    }
  }

  let name = UNKNOWN_FIELDS_MEMBER_NAME;
  for (let i = 1; taken.has(name); i++) {
    name = `${UNKNOWN_FIELDS_MEMBER_NAME}${i}`;
  }
  return name;
}

// @ts-ignore
function templateUnknownFieldsMember(typeDef: TypeDef, indent: number): string {
  if (
    !preserveUnknownFields() ||
    !getUnknownFieldsTypes().has(csharpTypeForJsonContext(typeDef))
  ) {
    return "";
  }

  const member = `/// <summary>
/// JSON fields received for this object that its schema does not declare. They are written back when the object is serialized.
/// </summary>
[System.Text.Json.Serialization.JsonExtensionData]
public System.Collections.Generic.Dictionary<string, System.Text.Json.JsonElement>? ${unknownFieldsMemberName(
    typeDef,
  )} { get; set; }`;
  return `\n\n${indentString(member, indent)}`;
}

unregisterTemplateFunc("templateUnknownFieldsMember");
registerTemplateFunc(
  "templateUnknownFieldsMember",
  templateUnknownFieldsMember,
);
