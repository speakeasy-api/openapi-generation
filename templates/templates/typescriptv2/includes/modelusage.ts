// @ts-ignore
function templateFileValue(
  path: string,
  additionalContext?: TemplateValueContext,
): string {
  addUsageImport("node:fs", "openAsBlob");
  return `await openAsBlob("${path}")`;
}
