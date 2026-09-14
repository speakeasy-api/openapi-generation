export {};

// Runtime globals inherited from csharp/includes at runtime but not visible to tsc.

declare global {
  // namespaces.ts
  function getScopePath(scope: string): string;
  function getAccessNamespace(
    outputLocation: string,
    full?: boolean,
    utils?: boolean,
    rootNamespace?: string,
  ): string;
  function getPackageId(): string;
  function getSDKNamespace(): string;
  function getSDKTopLevelFolder(): string;
  function getProjectFilePath(): string;
  function getGeneratorsAnalyzerReference(): string;
  function getDefaultHttpClient(): string;
  function getUsageSourceFilePath(sourceDir: string): string;

  // imports.ts
  function addImport(
    imports: Set<string>,
    imp: string,
    local: boolean,
  ): Set<string>;
  function addImportForType(imports: Set<string>, type: TypeDef): Set<string>;
  function getTypeImports(
    imports: Set<string>,
    typeDef: TypeDef,
    usageLocation: string,
    topLevel: boolean,
    visited?: Record<string, boolean>,
  ): Set<string>;
  function templateImports(imports: Set<string>, indent: number): string;

  // models.ts
  type CSharpModel = {
    Name: string;
    Type: TypeDef;
    OutputLocation: string;
  };

  // dependencies.ts
  var deps: Record<string, { version: string }>;

  // annotations.ts
  function getFieldAnnotations(fieldDef: FieldDef): Annotation[];
  function templateFieldAnnotation(
    fieldDef: FieldDef,
    annotation: Annotation,
  ): string;
  function templateDeprecationAnnotation(
    comments: CommentDef | null,
    indent: number,
    type: "method" | "field" | "class",
    addLeadingNewLine?: boolean,
  ): string;

  // errors.ts
  function templateDefaultError(full?: boolean): string;

  // responses.ts
  function templateErrorUnion(
    typeDef: TypeDef,
    responseFormat: string,
    output: string,
  ): string;
  function templateResponseBuilder(
    subResponse: SubResponse,
    responseFormat: string,
    pagination: boolean,
    responseType: string,
    content?: ResponseBodyContent,
  ): string;
  function templateEmptyResponseBuilder(
    op: Operation,
    responseFormat: string,
    pagination: boolean,
    responseType: string,
  ): string;
  function templateBodyHandler(
    op: Operation,
    subResponse: SubResponse,
    content: ResponseBodyContent,
  ): string;
  function templateEventStreamHandler(
    op: Operation,
    subResponse: SubResponse,
    content: ResponseBodyContent,
  ): string;
  function getResponseUnionName(
    typeDef: TypeDef,
    content: ResponseBodyContent,
  ): string;
  // documentation.ts
  function templateMethodParametersDocsSDKs(operation: Operation): SDKDocType[];
  function templateMethodResponseDocsSDKs(operation: Operation): SDKDocType[];
  function templateTypeMarkdown(params: TemplateTypeMarkdownParams): string;

  // serverusage.ts
  function getServerReadmeContext(): ServerReadmeContext;
  function templateServerVariableSetter(v: ServerVariable): string;

  // sanitization.ts
  function sanitizeFieldName(name: string, className?: string): string;
  function sanitizeClassName(name: string, isInterface?: boolean): string;
  function sanitizeUnionTypeName(typeDef: TypeDef): string;

  // templating.ts
  function getDataTypeFormat(request: RequestDef): string;
  function canTypeBeNull(typeDef: TypeDef): boolean;
  function canFieldBeNull(field: FieldDef): boolean;

  // models.ts (csharp v1, presence-aware serialization)
  function needsOptionalNullableWrapper(fieldDef: FieldDef): boolean;

  // sanitization.ts (csharp v1)
  function isValueType(typeDef: TypeDef): boolean;

  // form.ts (csharpv2 v2)
  function templateApplyFormBody(op: Operation): string;
  function jsonTypeInfoRef(typeDef: TypeDef): string;
  function csharpTypeForJsonContext(typeDef: TypeDef): string;

  // multipart.ts (csharpv2 v2)
  function templateApplyMultipartBody(op: Operation): string;
}
