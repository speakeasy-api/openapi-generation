// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "None";
}

// @ts-ignore
function sanitizeClassName(name: string): string {
  return name;
}

// @ts-ignore
function sanitizeFileName(name: string): string {
  name = sanitizeFile(name, "");

  return name.toLowerCase();
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => sanitizeFile(segment, "").toLowerCase())
    .join("/");
}
