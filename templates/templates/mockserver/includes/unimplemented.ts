// This file contains unimplemented functions that are called by common logic to
// appease the TypeScript compiler.

// @ts-ignore
function getTestFileName(_: string): string {
  throw new Error("getTestFileName not implemented by target");
}

// @ts-ignore
function sanitizeSDKAccess(_: SDK): string {
  throw new Error("sanitizeSDKAccess not implemented by target");
}

// @ts-ignore
function sanitizeMethodName(_: Operation): string {
  throw new Error("sanitizeMethodName not implemented by target");
}

// @ts-ignore
function sanitizeSecurityFieldName(_: string): string {
  throw new Error("sanitizeSecurityFieldName not implemented by target");
}

// @ts-ignore
function templateDefaultError(_?: boolean): string {
  throw new Error("templateDefaultError not implemented by target");
}

//@ts-ignore
function templateMethodParametersDocsSDKs(_: Operation): SDKDocType[] {
  throw new Error("templateMethodParametersDocsSDKs not implemented by target");
}

//@ts-ignore
function templateMethodResponseDocsSDKs(_: Operation): SDKDocType[] {
  throw new Error("templateMethodResponseDocsSDKs not implemented by target");
}

// @ts-ignore
function templateTypeMarkdown(_: TemplateTypeMarkdownParams): string {
  throw new Error("templateTypeMarkdown not implemented by target");
}

function templateServerVariableSetter(_: ServerVariable): string {
  throw new Error("templateServerVariableSetter not implemented by target");
}

//@ts-ignore
function getServerReadmeContext(): ServerReadmeContext {
  throw new Error("getServerReadmeContext not implemented by target");
}
