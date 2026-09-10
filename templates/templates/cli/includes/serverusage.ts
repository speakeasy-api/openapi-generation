// @ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  return {
    ServerVariableSetter: "flag",
    ServerByName: "--server-url <url>",
    ServerByIndex: "--server-url <url>",
    ServerByUrl: "--server-url <url>",
  };
}

// @ts-ignore
function templateServerVariableSetter(v: ServerVariable): string {
  return `--${caser().ToKebab(sanitizeName(v.Name))} <value>`;
}

registerTemplateFunc(
  "templateServerVariableSetter",
  templateServerVariableSetter,
);
