// Zod test schema generation functions using isomorphic helpers

// @ts-ignore
function templateTestSchemaPersonObject(): string {
  return z.catchall(
    z.object(`{
      name: ${z.string()},
      age: ${z.number()},
    }`),
    z.unknown(),
  );
}

// @ts-ignore
function templateTestSchemaPersonObjectWithArrayCatchall(): string {
  return z.catchall(
    z.object(`{
      name: ${z.string()},
      age: ${z.number()},
    }`),
    z.array(z.number()),
  );
}

registerTemplateFunc(
  "templateTestSchemaPersonObject",
  templateTestSchemaPersonObject,
);
registerTemplateFunc(
  "templateTestSchemaPersonObjectWithArrayCatchall",
  templateTestSchemaPersonObjectWithArrayCatchall,
);
