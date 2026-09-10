import { Faker } from "@faker-js/faker";

// @ts-ignore: Duplicate identifier
export {};

declare global {
  /* Record<Scope, Record<ModelName, TypeDef[]>> */
  type BucketedTypes = SequencedMap<string, SequencedMap<string, TypeDef[]>>;

  interface GlobalContext {
    AST: AST;
    Config: Record<string, any>;
    EnabledTemplateFeatures?: Record<string, boolean>;
  }

  type GlobalComputed = Record<string, any>;

  interface Context {
    Global: GlobalContext;
    Local: any;
    GlobalComputed: GlobalComputed;
    LocalComputed: any;
    RecursiveComputed: any;
    OutFile?: string;
  }

  type JsonPointer = {
    set: (obj: any, path: string, value: any) => any;
    get: (obj: any, path: string) => any;
  };

  var context: Context;
  var faker: Faker;
  var jsonpointer: JsonPointer;
  var withExclusions: (faker: Faker) => Faker;
  var registerTemplateFunc: (name: string, func: (...any) => void) => void;
  var unregisterTemplateFunc: (name: string) => void;
  var sanitizeFile: (name: string, replacement: string) => string;
  var sanitizeName: (name: string) => string;
  var pythonRelativeImport: (fromPkg: string, toPkg: string) => string;
  var quote: (value: string) => string;
  var debug: (...any: any[]) => string;
  var templateFile: (
    templateFile: string,
    targetFile: string,
    values: any,
  ) => void;
  var templateString: (templateFile: string, values: any) => string;
  var templateStringInput: (
    templateFile: string,
    targetFile: string,
    values: any,
  ) => string;
  var copy: (inFile: string, outFile: string) => void;

  // lists files in input directory
  var getDirectoryFiles: (dir: string) => string[];
  var directoryExists: (dir: string) => boolean;
  var render: (str: string) => void;
  var readFile: (fileName: string) => string;
  var fileExists: (fileName: string) => boolean;
  /**
   * True when a .speakeasy/patches patch file exists for the given generated
   * path. A patch pins its target to generator+patch ownership: templates
   * keep emitting a generated-once file whose content a patch owns, so
   * hermetic and incremental regeneration produce identical output.
   */
  var patchFileExists: (fileName: string) => boolean;
  var base64Encode: (str: string) => string;

  // lists files in output directory
  var listFiles: (folder: string) => string[];
  var listFolders: (folder: string) => string[];
  /** @deprecated Non-atomic with registerComment*; use resolveComment* which registers and returns the merged comment in one step. */
  var getComment: (commentName: string) => SimpleCommentDef;
  /** @deprecated Use resolveCommentDef. */
  var registerCommentDef: (
    source: CommentSource,
    namespace: string,
    name: string,
    comment: CommentDef | null,
  ) => string;
  var registerComment: (
    source: CommentSource,
    namespace: string,
    name: string,
    summary: string,
    description: string,
  ) => string;
  // Atomically registers a comment (empty input claims the source slot) and
  // returns the merged comment, so rendered output never depends on what
  // other render jobs stored under the same key.
  var resolveCommentDef: (
    source: CommentSource,
    namespace: string,
    kind: string,
    name: string,
    comment: CommentDef | null,
  ) => SimpleCommentDef;
  var resolveComment: (
    source: CommentSource,
    namespace: string,
    kind: string,
    name: string,
    summary: string,
    description: string,
  ) => SimpleCommentDef;
  var generateReadmeTOC: (
    contents: string,
    minLevel: number,
    maxLevel: number,
    exclude: string, // comma-separated list of headers to exclude
  ) => string;
  var generateReadmeTOCEntry: (header: string, level: number) => string;
  var isReadmeSectionMarkedAsNoAction: (
    prefix: string,
    title: string,
    id: string,
    suffix: string,
    contents: string,
  ) => boolean;
  var parseReadmeBlockIDs: (
    prefix: string,
    suffix: string,
    contents: string,
  ) => string;
  var parseReadmeBlock: (
    prefix: string,
    id: string,
    suffix: string,
    contents: string,
  ) => string;
  var replaceReadmeBlock: (
    prefix: string,
    title: string,
    id: string,
    suffix: string,
    contents: string,
    newTitle: string,
    newContents: string,
  ) => string | undefined;
  var selectExampleOperations: (
    sdk: SDK,
    limit: number,
    scopes: UsageExampleScope[],
    isMainExample: boolean,
  ) => UsageContext[];
  var fileUploadOpPredicate: (sdk: SDK, operation: Operation) => boolean;
  var adjustHeadings: (level: number, content: string) => string;
  var writeFile: (
    fileName: string,
    data: string,
    perm: number,
    checkExisting: boolean,
  ) => string;
  var replaceInstallation: (contents: string, installation: string) => string;
  var replaceUsage: (contents: string, usage: string) => string;
  var replaceOperations: (contents: string, operations: string) => string;
  var formatUsageSnippetOutput: (snippet: string) => string;
  var createUsageContext: (
    sdk: SDK,
    operation: Operation,
    usageExample: UsageExampleConfig,
  ) => UsageContext;
  var sortTypeDefFieldCount: (ls: TypeDefs) => TypeDefs;
  var sortTypeDefRequiredFieldsDescending: (ls: TypeDefs) => TypeDefs;
  var sortDiscriminatorMappingRequiredFieldsDescending: (
    ls: DiscriminatorMappings,
  ) => DiscriminatorMappings;
  var compareTypeDefsDescendingRequiredFields: (
    a: TypeDef,
    b: TypeDef,
  ) => number;
  var addUntrackedPattern: (goRegexPattern: string) => void;
  var addHeaderPattern: (goRegexPattern: string, dontMatch?: boolean) => void;
  var addTerraformResourceCount: (resourceCount: number) => void;

  interface Caser {
    ToGoPascal: (s: string) => string;
    ToGoPascalIgnoreAcronyms: (s: string) => string;
    ToPascal: (s: string) => string;
    ToSnake: (s: string) => string;
    ToSNAKE: (s: string) => string;
    ToGoCamel: (s: string) => string;
    ToCamel: (s: string) => string;
    ToKebab: (s: string) => string;
    ToKEBAB: (s: string) => string;
  }

  var caser: () => Caser;
  var interoptCallMethod: (
    ast: AST,
    language: string,
    method: string,
    ...args: any[]
  ) => any;
  var interoptTemplateString: (
    ast: AST,
    language: string,
    templateFile: string,
    context: any,
  ) => string;
  var interoptTemplateTarget: (
    target: string,
    outDir: string,
    ast: AST,
    targetCfg: Record<string, any>,
    enabledFeatures?: Record<string, boolean>,
  ) => void;
  var getTargetDefaultTemplateConfig: (target: string) => Record<string, any>;
  var getAllTypes: () => TypeDefs;
  var parseYaml: (file: string) => Record<string, any>;

  type LogField = {
    Key: string;
    Type: number; // See types here https://github.com/uber-go/zap/blob/master/zapcore/field.go#L37
    Integer: number;
    String: string;
    Interface: any;
  };

  interface Logger {
    Debug: (msg: string, fields?: LogField[]) => void;
    Info: (msg: string, fields?: LogField[]) => void;
    Warn: (msg: string, fields?: LogField[]) => void;
    Error: (msg: string, fields?: LogField[]) => void;
    Github: (msg: string) => void; // Prints only when running in GitHub Actions
    With: (fields?: LogField[]) => Logger;
  }

  interface MethodArguments {
    merged: FieldDef[];
    request: FieldDef[];
    security: FieldDef[];
    requestFieldName: string;
    flattenedRequest: boolean;
  }

  var logger: () => Logger;
  var logWarning: (msg: string, lineNumber: number) => void;
  var accountHasFeatureAccess: (feat: Feature) => boolean;
  var addIgnoreOverride: (ignoreRegex: string) => void;
  var getGenericMapTypeDef: () => TypeDef;
  var getGenericArrayTypeDef: () => TypeDef;
  var isRE2Regex: (pattern: string) => boolean;

  /** Compares two version strings using github.com/hashicorp/go-version.
   *  Returns -1 if a < b, 0 if a == b, 1 if a > b.
   *  Handles Go module versions including v prefix, pre-release,
   *  pseudoversions, and +incompatible suffixes. */
  var compareVersions: (a: string, b: string) => number;

  type SecurityUsageContext = {
    index?: number;
    example?: any;
  };

  interface FileDirective {
    path: string;
    overriddenFileName?: string;
  }

  interface ExamplesHelper {
    NewExamplesMap: () => Examples;
    NewOperationExamplesMap: () => SequencedMap<string, OperationExamples>;
    NewParameterExamples: () => ParameterExamples;
    NewResponsesExamplesMap: () => SequencedMap<
      string,
      SequencedMap<string, YamlNode> | null
    >;
    NewYamlNodeMap: () => SequencedMap<string, YamlNode>;
    NewExample: (name: string, value: YamlNode) => Example;
    YamlNodeFromString: (s: string) => YamlNode;
    GetDefaultExampleName: (op: Operation) => string;
    IsDefaultExample: (name: string) => boolean;
  }

  var examplesHelper: () => ExamplesHelper;

  var addTypeToBucket: (
    types: BucketedTypes,
    model: string,
    typeDef: TypeDef,
  ) => BucketedTypes;

  var getOperationModels: (
    operation: Operation,
    ignoreResponseSubTypes?: boolean,
    ignoreResponseTypes?: boolean,
    considerFlattening?: boolean,
  ) => BucketedTypes;

  var getOperationModelName: (operation: Operation) => string;
  var isDebug: () => boolean;

  var areCircular: (t1: TypeDef, t2: TypeDef) => boolean;
  var isPartOfCycle: (t: TypeDef) => boolean;
  var createBucket: (existingBucket: BucketedTypes, bucketName: string) => void;
  var typeDefToFieldDef: (
    typeDef: TypeDef,
    parentFieldDef?: FieldDef,
    optional?: boolean,
    nullable?: boolean,
  ) => FieldDef;
  var topologicalSortTypeDefs: (
    types: TypeDefs,
    hoistCyclicUnions?: boolean,
  ) => TypeDefs;
  var createZeroTypeDef: () => TypeDef;

  /** Converts a JSON string into an HCL jsonencode() expression. Falls back
   * to a quoted string literal if conversion fails. */
  var jsonToHCLExpression: (jsonString: string) => string;
}
