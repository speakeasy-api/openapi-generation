function templateTestSchemaPersonObject(): string {
  return zodCatchall(
    zodObject({
      name: zodString(),
      age: zodNumber(),
    }),
    zodAny(),
  );
}

function templateTestSchemaPersonObjectWithArrayCatchall(): string {
  return zodCatchall(
    zodObject({
      name: zodString(),
      age: zodNumber(),
    }),
    zodArray(zodNumber()),
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
