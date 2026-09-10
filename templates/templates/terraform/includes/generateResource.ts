/** Returns templated resp.State.RemoveResource() checks for all top level
 *  attributes with x-speakeasy-soft-delete-property annotation. */
function generateResourceSoftDeletePropertiesRemovals(
  entity: TerraformManagedResource,
): string {
  const result: string[] = [];

  entity.ReadSoftDeleteProperties?.forEach((fieldDef) => {
    result.push(`if !data.${sanitizeFieldName(fieldDef.Name)}.IsNull() {`);
    result.push(`${templateIndent(1)}resp.State.RemoveResource(ctx)`);
    result.push(`${templateIndent(1)}return`);
    result.push(`}`);
  });

  return result.join(`\n${templateIndent(1)}`);
}

registerTemplateFunc(
  "generateResourceSoftDeletePropertiesRemovals",
  generateResourceSoftDeletePropertiesRemovals,
);
