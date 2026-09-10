function validatePythonEventStreamClassName(
  value: unknown,
  fallback: string,
): string {
  const name = typeof value === "string" && value.length > 0 ? value : fallback;
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(name)) {
    throw new Error(
      `eventStreamClassNames values must be valid Python identifiers, got ${JSON.stringify(
        name,
      )}`,
    );
  }
  if (methodReservedKeywords.includes(name)) {
    throw new Error(
      `eventStreamClassNames values must not be Python keywords, got ${JSON.stringify(
        name,
      )}`,
    );
  }
  return name;
}

function pythonEventStreamSyncClassName(): string {
  const classNames = context.Global.Config.EventStreamClassNames ?? {};
  return validatePythonEventStreamClassName(
    classNames.sync ?? classNames.Sync,
    "EventStream",
  );
}

function pythonEventStreamAsyncClassName(): string {
  const classNames = context.Global.Config.EventStreamClassNames ?? {};
  const asyncClassName = validatePythonEventStreamClassName(
    classNames.async ?? classNames.Async,
    "EventStreamAsync",
  );
  const syncClassName = pythonEventStreamSyncClassName();
  if (asyncClassName === syncClassName) {
    throw new Error(
      `eventStreamClassNames.sync and eventStreamClassNames.async must be different, got ${JSON.stringify(
        asyncClassName,
      )}`,
    );
  }
  return asyncClassName;
}

function pythonEventStreamUsesDirectTypeRefs(): boolean {
  const classNames = context.Global.Config.EventStreamClassNames;
  if (classNames === undefined || classNames === null) {
    return false;
  }

  return (
    pythonEventStreamSyncClassName() !== "EventStream" ||
    pythonEventStreamAsyncClassName() !== "EventStreamAsync"
  );
}

function pythonEventStreamSyncTypeRef(): string {
  if (pythonEventStreamUsesDirectTypeRefs()) {
    return pythonEventStreamSyncClassName();
  }
  return `eventstreaming.${pythonEventStreamSyncClassName()}`;
}

function pythonEventStreamAsyncTypeRef(): string {
  if (pythonEventStreamUsesDirectTypeRefs()) {
    return pythonEventStreamAsyncClassName();
  }
  return `eventstreaming.${pythonEventStreamAsyncClassName()}`;
}

function pythonEventStreamMethodTypeRef(): string {
  return "__PYTHON_EVENT_STREAM_TYPE_REF__";
}

registerTemplateFunc(
  "pythonEventStreamSyncClassName",
  pythonEventStreamSyncClassName,
);
registerTemplateFunc(
  "pythonEventStreamAsyncClassName",
  pythonEventStreamAsyncClassName,
);
registerTemplateFunc(
  "pythonEventStreamUsesDirectTypeRefs",
  pythonEventStreamUsesDirectTypeRefs,
);
registerTemplateFunc(
  "pythonEventStreamSyncTypeRef",
  pythonEventStreamSyncTypeRef,
);
registerTemplateFunc(
  "pythonEventStreamAsyncTypeRef",
  pythonEventStreamAsyncTypeRef,
);
registerTemplateFunc(
  "pythonEventStreamMethodTypeRef",
  pythonEventStreamMethodTypeRef,
);
