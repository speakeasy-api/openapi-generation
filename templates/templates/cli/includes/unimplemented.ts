// This file contains unimplemented functions that are called by common logic to
// appease the TypeScript compiler.

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
