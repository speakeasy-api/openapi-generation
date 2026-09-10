function templateCrystalline(): TemplateFileJob[] {
  const jobs = [];
  copy("crystalline.rb", "lib/crystalline.rb");
  const crystallineSrcPath = "lib/crystalline/";
  let files = getDirectoryFiles("crystalline");

  for (const file of files) {
    let outFile = file.replace("crystalline/", crystallineSrcPath);
    if (file.endsWith("stmpl")) {
      const filename = outFile.substring(0, outFile.length - 6);
      jobs.push(createTemplateFileJob(file, filename, {}));
    } else {
      copy(file, outFile);
    }
  }

  return jobs;
}

type RubyType = {
  ModelName: string;
  Scope: string;
  Type: TypeDef;
};

type RubyModel = {
  Name: string;
  Types: RubyType[];
  Scope: string;
  Servers?: Servers;
};

// @ts-ignore
function getModelJobs(
  packageName: string,
  types: BucketedTypes,
  path: string,
  ext: string,
): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];

  const scopeKeys = ["shared", "operations", "errors", "callbacks", "webhooks"];
  const scopeConfigs = scopeKeys.map((key) => {
    const path = getScopePath(key);
    return { key, path, dirName: path.split("/").pop() };
  });

  // Map from scope key → set of model names belonging to that scope
  const modelsByScope = new Map<string, Set<string>>();
  for (const sc of scopeConfigs) {
    modelsByScope.set(sc.key, new Set<string>());
  }

  // Lookup: lowercase directory name → scope key (for matching output locations)
  const dirNameToScope = new Map<string, string>();
  for (const sc of scopeConfigs) {
    if (sc.dirName) {
      dirNameToScope.set(sc.dirName.toLowerCase(), sc.key);
    }
  }

  const standardScopes = scopeConfigs
    .map((sc) => sc.dirName?.toLowerCase())
    .filter(Boolean);

  // Track custom namespace models - Map from namespace name to {models, outputLocation}
  const customNamespaceModels = new Map<
    string,
    { models: Set<string>; outputLocation: string }
  >();

  for (const [rawOutputLocation, models] of sequencedMapEntries(types)) {
    const outputLocation = sanitizeOutputLocation(rawOutputLocation);
    let templatedModels = [];

    // Check if this is a custom namespace
    // A path is a custom namespace if:
    // 1. It's a non-empty path that doesn't match any standard scope path exactly
    // 2. For paths like "models/foo", check if "foo" is not a standard scope
    const parts = rawOutputLocation.split("/");
    const lastPart = parts.length > 0 ? parts[parts.length - 1] : "";
    const isCustomNamespace =
      parts.length > 0 &&
      parts[0] !== "" &&
      lastPart !== "" &&
      !standardScopes.includes(lastPart.toLowerCase());

    // For non-custom namespaces, determine the effective scope from the output
    // location. This handles x-speakeasy-model-namespace redirecting models into
    // standard scope directories that differ from their native scope.
    const lastPartLower = lastPart.toLowerCase();
    const outputScopeName = !isCustomNamespace
      ? dirNameToScope.get(lastPartLower) ?? null
      : null;

    for (const [model, tt] of sequencedMapEntries(models)) {
      let scope: string = undefined;

      tt.forEach((type) => {
        let modelPath;
        const modelName = type.Name;
        if (outputLocation) {
          modelPath = `${path}/${outputLocation}`;
        }

        if (!scope) {
          scope = type.Scope.toString();
        }
        // When a model is redirected into a standard scope directory via
        // x-speakeasy-model-namespace, use the output location's scope for
        // template rendering (module declarations and RBI class paths).
        const templateScope = outputScopeName || scope;
        const rubyModel = {
          Name: modelName,
          Scope: templateScope,
          Type: type,
          Servers: context.Global.AST.OperationServers.Get(model)[0],
          OutputLocation: rawOutputLocation,
          IsCustomNamespace: isCustomNamespace,
        };

        jobs.push(
          createTemplateFileJob(
            `modelfile.${ext}.stmpl`,
            `${modelPath}/${sanitizeFileName(modelName)}.${ext}`,
            rubyModel,
          ),
        );

        jobs.push(
          createTemplateFileJob(
            `model.rbi.stmpl`,
            `${modelPath}/${sanitizeFileName(modelName)}.rbi`,
            rubyModel,
          ),
        );

        if (isCustomNamespace) {
          // Track model for custom namespace module file
          // Use the last part of the path as the namespace name (e.g., "foo" from "models/foo")
          const namespaceName = lastPart;
          if (!customNamespaceModels.has(namespaceName)) {
            customNamespaceModels.set(namespaceName, {
              models: new Set<string>(),
              outputLocation: outputLocation,
            });
          }
          customNamespaceModels.get(namespaceName).models.add(modelName);
        } else {
          // Use the effective scope (from output location or native scope) to
          // determine which standard scope module should autoload this model.
          const effectiveScope = outputScopeName || scope;
          const scopeSet = modelsByScope.get(effectiveScope);
          if (scopeSet) {
            scopeSet.add(modelName);
          } else {
            modelsByScope.get("shared").add(modelName);
          }
        }
        templatedModels.push(modelName);
        return rubyModel;
      });
    }
  }
  modelsByScope.get("errors").add(getDefaultErrorClassName());

  // Create module files for custom namespaces
  for (const [namespaceName, namespaceData] of customNamespaceModels) {
    const moduleName = caser().ToPascal(namespaceName);
    const snakeNamespace = caser().ToSnake(namespaceName);
    // Get the parent directory from outputLocation (e.g., "models" from "models/foo")
    const outputParts = namespaceData.outputLocation.split("/");
    const outputParent = outputParts.slice(0, -1).join("/");
    // Namespace module should be peer to the namespace folder
    // e.g., for models/foo/*.rb files, the module is at models/foo.rb
    const moduleFilePath = outputParent
      ? `${path}/${outputParent}/${snakeNamespace}.rb`
      : `${path}/${snakeNamespace}.rb`;
    // PathString needs to include the full path to the namespace directory
    const pathString = outputParent
      ? `${path.slice(4)}/${outputParent}/`
      : `${path.slice(4)}/`;
    jobs.push(
      createTemplateFileJob(`namespace_module.rb.stmpl`, moduleFilePath, {
        PathString: pathString, // strip leading `lib/` - getModuleAutoloads adds the module name
        ModuleName: snakeNamespace, // use snake_case for path generation
        ModuleNamePascal: moduleName, // use PascalCase for module declaration
        OutputLocation: namespaceData.outputLocation, // full path like "models/foo"
        NamespaceDepth:
          getNamespaceModuleParts(namespaceData.outputLocation).length + 1,
        Models: Array.from(namespaceData.models).sort(),
      }),
    );
  }

  jobs.push(
    createTemplateFileJob(
      `{{.Global.Config.PackageName}}.rb.stmpl`,
      `lib/${sanitizeFileName(packageName)}.${ext}`,
      {},
    ),
  );

  jobs.push(
    ...scopeConfigs
      .filter(
        // Filter out the webhooks scope when it has no models. Normally there are no webhooks models
        // so the webhooks file doesn't get generated. However customers can still set a custom namespace
        // called "webhooks" via x-speakeasy-model-namespace.
        (sc) => sc.key !== "webhooks" || modelsByScope.get(sc.key).size > 0,
      )
      .map((sc) =>
        createTemplateFileJob(`module.rb.stmpl`, `${path}/${sc.path}.rb`, {
          PathString:
            [path.slice(4), parentDir(sc.path)].filter(Boolean).join("/") + "/", // strip leading `lib/`
          ModuleName: getScopeName(sc.key),
          ModuleDeclarations: templateModuleDeclarations(sc.path, 1),
          ModuleClosures: templateModuleClosures(sc.path, 1),
          ModuleDepth: getNamespaceModuleParts(sc.path).length + 1,
          Models: Array.from(modelsByScope.get(sc.key)).sort(),
        }),
      ),
  );

  return jobs;
}

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  title: string | null,
  indent: number,
  type: "method" | "field" | "class" | null = null,
  additionalNotes: string = "",
  omitDocs: boolean = false,
): string {
  if (comments || additionalNotes) {
    let lines = [];

    let firstLineParts = [];

    if (title) {
      firstLineParts.push(title);
    }

    let descriptionLines = [];
    if (comments?.Description) {
      descriptionLines = comments.Description.replaceAll("\r\n", "\n").split(
        "\n",
      );
    }

    if (comments?.Summary) {
      firstLineParts.push(comments.Summary);
    } else if (descriptionLines.length > 0) {
      firstLineParts.push(descriptionLines.shift());
    }

    const firstLine = firstLineParts.join(" - ");

    if (firstLine) {
      lines.push(...firstLine.replaceAll("\r\n", "\n").split("\n"));
    }

    if (descriptionLines.length > 0) {
      lines.push(...descriptionLines);
    }

    if (comments?.ExternalDocs && !omitDocs) {
      let externalDocs = comments.ExternalDocs.URL;
      if (comments.ExternalDocs.Description) {
        externalDocs += ` - ${comments.ExternalDocs.Description}`;
      }

      lines.push(...externalDocs.replaceAll("\r\n", "\n").split("\n"));
    }

    if (additionalNotes) {
      if (lines.length > 0) {
        lines.push("");
      }

      lines.push(...additionalNotes.split("\n"));
    }

    if (comments?.Deprecated) {
      if (lines.length > 0) {
        lines.push("");
      }

      let deprecated = `@deprecated ${type}: ${comments.DeprecationMessage}.`;

      if (comments.DeprecationReplacement) {
        const replacement = sanitizeDeprecationReplacement(
          comments.DeprecationReplacement,
          type,
        );
        if (replacement) {
          deprecated += ` Use ${replacement} instead.`;
        }
      }

      lines.push(...deprecated.split("\n"));
    }

    if (lines.length == 0) {
      return "";
    }

    lines = lines.map((line) => (line === "" ? "#" : `# ${line}`));

    return indentLines(lines, indent);
  }

  return "";
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map((f) => `\`${sanitizeFieldName(f.Name)}\``);

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set on the \`security\` parameter when initializing the SDK.`
      : `If set, this operation will use ${items} from the global security.`;

  if (fieldList.length === 1) {
    return remark(fieldList[0]);
  }

  const last = fieldList.pop();
  if (fieldList.length === 1) {
    return remark(`either ${fieldList[0]} or ${last}`);
  }

  return remark(`one of ${fieldList.join(", ")}, or ${last}`);
}

// @ts-ignore
function templateSDKInitFieldName(name: string, suffix: string): string {
  let reserved = ["security", "server", "server_url", "url_params", "client"];

  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reserved.push("server");
  } else {
    reserved.push("server_idx");
  }

  let fieldName = sanitizeFieldName(name);

  if (reserved.includes(name)) {
    fieldName += "_" + suffix;
  }

  return fieldName;
}

registerTemplateFunc("templateSDKInitFieldName", templateSDKInitFieldName);

// @ts-ignore
function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example, additionalContext);
    }
  }

  if (
    index === undefined &&
    example === undefined &&
    context.Global.AST.MainSDK.Security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme"
  ) {
    return "";
  }

  let securityFieldName = "security";
  if (
    context.Global.AST.MainSDK.Security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    securityFieldName = caser().ToSnake(
      context.Global.AST.MainSDK.Security.Type.Fields[0].Name,
    );
  }

  const securityValue = templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    additionalContext.indent || 1,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    example,
    additionalContext,
  ).trim();
  // Strip trailing commas before closing parens in security model constructors
  const cleaned = securityValue.replace(/,(\n\s*\))/g, "$1");
  return `${securityFieldName}: ${cleaned},`;
}

// @ts-ignore
function joinSDKParams(params: string[]): string {
  if (params.length > 0) {
    // Strip trailing comma from last param
    const last = params[params.length - 1];
    if (last.endsWith(",")) {
      params[params.length - 1] = last.slice(0, -1);
    }

    params.unshift("");

    return params.join(`\n${"  ".repeat(1)}`) + "\n";
  }

  return "";
}

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${sanitizeFieldName(fieldDef.Name)}: `;
}

function templateSDKConfigurationConstructor(sdk: SDK): string {
  const serverVariables = sdk.Servers && sdk.Servers.GetVariables().length > 0;

  const params = ["client", "hooks", "retry_config", "timeout_ms"];

  let securityName;
  if (sdk.Security) {
    if (
      sdk.Security.Type.Fields.length == 1 &&
      context.Global.Config.FlattenGlobalSecurity
    ) {
      securityName = sanitizeFieldName(sdk.Security.Type.Fields[0].Name);
    } else {
      securityName = "security";
    }
    params.push(securityName);
    params.push("security_source");
  }
  params.push("server_url");
  if (sdk.Servers && sdk.Servers.ServerMap) {
    params.push("server");
  } else {
    params.push("server_idx");
  }
  if (serverVariables) {
    params.push("server_params");
  }
  if (sdk.Globals) {
    params.push("globals");
  }
  return `SDKConfiguration.new(
${indentString(params.join(",\n"), 4)}
      )`;
}
registerTemplateFunc(
  "templateSDKConfigurationConstructor",
  templateSDKConfigurationConstructor,
);

/** Templates AST status codes, including wildcards, as individual status codes
 *  for faraday-retry retry_statuses. */
function templateFaradayRetryStatuses(statuses: string[]): string {
  const result: string[] = [];

  for (const status of statuses) {
    if (status.toUpperCase().includes("4XX")) {
      result.push(
        "400",
        "401",
        "402",
        "403",
        "404",
        "405",
        "406",
        "407",
        "408",
        "409",
        "410",
        "411",
        "412",
        "413",
        "414",
        "415",
        "416",
        "417",
        "418",
        "419",
        "421",
        "422",
        "423",
        "426",
        "428",
        "429",
      );
    } else if (status.toUpperCase().includes("5XX")) {
      result.push("500", "501", "502", "503", "504", "505");
    } else {
      result.push(status);
    }
  }

  return result.join(", ");
}

registerTemplateFunc(
  "templateFaradayRetryStatuses",
  templateFaradayRetryStatuses,
);

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "nil";
  }

  return `[${scopes.map((s) => `'${s}'`).join(", ")}]`;
}

registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

function templateStatusCodeCheck(varName: string, statusCodes: string[]) {
  const $templatedCodes = statusCodes.map((c) => `'${c}'`);
  return `Utils.match_status_code(${varName}, [${$templatedCodes.join(", ")}])`;
}
registerTemplateFunc("templateStatusCodeCheck", templateStatusCodeCheck);
function templateDefaultValue(
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  if (fieldDef.Optional) {
    return "nil";
  }

  switch (fieldDef.Type.Type.toString()) {
    case "union":
      return templateDefaultTypeValue(
        fieldDef.Type.AssociatedTypes[0],
        usageLocation,
      );
    default:
      return templateDefaultTypeValue(fieldDef.Type, usageLocation);
  }
}

function templateDefaultTypeValue(
  typeDef: TypeDef,
  usageLocation: string,
): string {
  switch (typeDef.Type.toString()) {
    case "enum":
      return `${sanitizeClass(typeDef, usageLocation, false)}::${
        getEnumNames(typeDef)[0]
      }`;
    case "class":
      return `${sanitizeClass(typeDef, usageLocation, false)}.new`;
    case "string":
      return `''`;
    case "date":
      return "::Date.new";
    case "date-time":
      return "::DateTime.new";
    case "integer":
    case "int32":
    case "number":
    case "float32":
      return "0";
    case "boolean":
      return "false";
    case "bytes":
      return `''`;
    case "map":
      return "{}";
    case "array":
      return "[]";
    case "any":
      return "nil";
    case "response":
      return `::Faraday::Response.new`;
    case "request":
      return `::Faraday::Request.new`;
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("templateDefaultValue", templateDefaultValue);

/**
 * This function sorts method params so that required fields with no default value are at the front
 * @param params
 * @returns
 */
function sortFields(fields: FieldDef[]): FieldDef[] {
  return fields.slice().sort((a, b) => {
    const aCount =
      (a.Optional ? 1 : 0) +
      (a.Nullable ? 1 : 0) +
      (a.Default ? 1 : 0) +
      (a.Const ? 1 : 0);
    const bCount =
      (b.Optional ? 1 : 0) +
      (b.Nullable ? 1 : 0) +
      (b.Default ? 1 : 0) +
      (b.Const ? 1 : 0);
    return aCount - bCount;
  });
}

registerTemplateFunc("sortFields", sortFields);

function filterFields(fields: FieldDef[], mode: string): FieldDef[] {
  if (mode === "const") {
    return fields.filter((f) => f.Const);
  }
  return fields.filter((f) => !f.Const);
}

registerTemplateFunc("filterFields", filterFields);

function templateCompareConst(field: FieldDef, fields?: FieldDef[]): string {
  const constValue = templateConstValue(field);
  const fieldName = fields
    ? deduplicatedFieldName(field.Name, fields)
    : sanitizeFieldName(field.Name);

  if (constValue === "nil") {
    return `unless ${fieldName}.nil?`;
  }

  switch (field.Type.Type.toString()) {
    case "decimal":
      return `unless !${fieldName}.nil? && ${fieldName}.to_d == ${constValue}.to_d`;
    case "number":
    case "float32":
    case "date-time":
      return `if !${fieldName}.nil? && (${constValue} - ${fieldName}).abs >= 0.1`;
    case "integer":
    case "int32":
      if (constValue === "0") {
        return `unless ${fieldName}&.zero?`;
      }
      return `unless ${fieldName} == ${constValue}`;
    default:
      return `unless ${fieldName} == ${constValue}`;
  }
}
registerTemplateFunc("templateCompareConst", templateCompareConst);

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    filePath = `.speakeasy/testfiles/${filePath}`;
  }

  return `File.binread('${filePath}')`;
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    filePath = `.speakeasy/testfiles/${filePath}`;
  }

  return `File.read("${filePath}")`;
}
