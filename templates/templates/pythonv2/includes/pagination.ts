function isAsyncPaginationSep2025Fix(): boolean {
  return (
    context.Global.Config.FixFlags?.["asyncPaginationSep2025"] ||
    context.Global.Config.AsyncMode === "split"
  );
}
registerTemplateFunc(
  "isAsyncPaginationSep2025Fix",
  isAsyncPaginationSep2025Fix,
);

function isClassThatContainsPaginationFields(
  requestBodyType: TypeDef,
  pageInput?: PaginationInputs,
  offsetInput?: PaginationInputs,
  cursorInput?: PaginationInputs,
): boolean {
  if (requestBodyType.Type?.toString() !== "class") {
    return false;
  }

  const fieldsToCheckFor: string[] = [];
  if (pageInput && pageInput.In.toString() === "requestBody") {
    fieldsToCheckFor.push(pageInput.Name.toString());
  }
  if (offsetInput && offsetInput.In.toString() === "requestBody") {
    fieldsToCheckFor.push(offsetInput.Name.toString());
  }
  if (cursorInput && cursorInput.In.toString() === "requestBody") {
    fieldsToCheckFor.push(cursorInput.Name.toString());
  }
  return (requestBodyType.Fields ?? []).some((field) =>
    fieldsToCheckFor.includes(field.Name?.toString()),
  );
}
registerTemplateFunc(
  "isClassThatContainsPaginationFields",
  isClassThatContainsPaginationFields,
);
