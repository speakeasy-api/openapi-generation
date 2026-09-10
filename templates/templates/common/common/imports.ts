// @ts-ignore
function getModelsLocation(outputLocation: string): string {
  return sanitizeOutputLocation(outputLocation);
}

registerTemplateFunc("getModelsLocation", getModelsLocation);
