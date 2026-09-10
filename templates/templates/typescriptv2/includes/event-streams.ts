const eventStreamReservedNames = new Set([
  "ReadableStream",
  "Uint8Array",
  "TextDecoder",
  "IteratorResult",
  "AsyncIterableIterator",
  "Symbol",
  "SseMessage",
  "WrapSseOptions",
]);

function eventStreamClassName(): string {
  const raw = context.Global.Config.EventStreamClassName;
  const name = sanitizeClassName(raw);
  if (eventStreamReservedNames.has(name)) {
    throw new Error(
      `when sanitized, eventStreamClassName ${JSON.stringify(
        raw,
      )} collides with the reserved name ${JSON.stringify(name)}`,
    );
  }
  return name;
}
registerTemplateFunc("eventStreamClassName", eventStreamClassName);

function eventStreamTypeRef(eventType: string): string {
  return `${eventStreamClassName()}<${eventType}>`;
}
registerTemplateFunc("eventStreamTypeRef", eventStreamTypeRef);

function addEventStreamImport(
  usageLocation: string,
  type: TSImportType = typeImport,
): string {
  return addInternalImport(
    "event-streams",
    eventStreamClassName(),
    usageLocation,
    type,
  );
}
registerTemplateFunc("addEventStreamImport", addEventStreamImport);

// @ts-ignore
function addEventStreamUsageImport(): string {
  const prefix = context.Global.Config.PackageName;
  const path = `${prefix}/lib/event-streams.js`;
  return addImport(path, eventStreamClassName(), typeImport, usageImports);
}
registerTemplateFunc("addEventStreamUsageImport", addEventStreamUsageImport);
