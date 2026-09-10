// @ts-ignore (Ignore override of function implementation)
// TODO: deprecated remove when all templates are using the job systems
function templateModels(types: BucketedTypes, path: string, ext: string) {
  const jobs = getModelJobs(types, path, ext);

  for (const job of jobs) {
    templateFileJob(job);
  }
}

// @ts-ignore (Ignore override of function implementation)
function getModelJobs(
  types: BucketedTypes,
  path: string,
  ext: string,
): TemplateFileJob[] {
  let jobs: TemplateFileJob[] = [];

  for (const [rawOutputLocation, models] of sequencedMapEntries(types)) {
    const outputLocation = sanitizeOutputLocation(rawOutputLocation);
    for (const [model, types] of sequencedMapEntries(models)) {
      let modelPath = path;

      if (outputLocation) {
        modelPath = `${path}/${outputLocation}`;
      }

      const typesArray = types;

      let ctx: Model = {
        DocGroup: modelPath,
        Name: model,
        Types: typesArray,
        OutputLocation: outputLocation,
      };

      if (
        typesArray.find((t) => t.Scope.toString() == "operations") ||
        typesArray.length == 0
      ) {
        ctx.Servers = context.Global.AST.OperationServers.Get(model)[0];
      }

      if (typesArray.length == 0 && !ctx.Servers) {
        continue;
      }

      jobs.push(
        createTemplateFileJob(
          `modelfile.${ext}.stmpl`,
          `${modelPath}/${sanitizeFileName(model)}.${ext}`,
          ctx,
        ),
      );
    }
  }

  return jobs;
}

type TemplateFileJob = {
  ID: "templateFile";
  FileName: string;
  Context: {
    templateFile: string;
    context: any;
  };
};

function createTemplateFileJob(
  templateFile: string,
  fileName: string,
  context: any,
): TemplateFileJob {
  return {
    ID: "templateFile",
    FileName: fileName,
    Context: { templateFile, context },
  };
}

// @ts-ignore
function templateFileJob(job: TemplateFileJob) {
  templateFile(job.Context.templateFile, job.FileName, job.Context.context);
}

registerTemplateFunc("fileExists", fileExists);

function getTemplateAuxiliaryFilesJob(sourcePath: string = "auxiliary"): Job {
  return {
    ID: "auxiliary",
    FileName: "auxiliary",
    Context: { sourcePath },
  };
}

function getGeneratedLicenseJobs(): Job[] {
  if (context.Global.Config.GeneratedLicense !== "agpl") {
    return [];
  }

  return [
    createTemplateFileJob("generated-license/LICENSE", "LICENSE", {}),
    createTemplateFileJob("generated-license/NOTICE.stmpl", "NOTICE", {}),
  ];
}

// @ts-ignore (Ignore override of function implementation)
function templateAuxiliaryFiles(data: any, sourcePath: string = "auxiliary") {
  let files = getDirectoryFiles(sourcePath);

  for (const file of files) {
    let outFile = file.replace(`${sourcePath}/`, "");

    outFile = templateStringInput("auxPath:" + outFile, outFile, data);
    if (outFile.trim().length == 0) {
      continue;
    }

    if (file.endsWith(".stmpl")) {
      outFile = outFile.replace(".stmpl", "");

      templateFile(file, outFile, context.Global);
    } else {
      copy(file, outFile);
    }
  }
}

function getTemplateDirectoryJob(dir: string, outDir?: string): Job {
  return {
    ID: "templateDirectory",
    FileName: outDir ?? dir,
    Context: {
      dir,
      outDir,
    },
  };
}

function templateDirectoryJob(job: Job) {
  templateDirectory(job.Context.dir, job.Context.outDir);
}

// @ts-ignore (Ignore override of function implementation)
function templateDirectory(dir: string, outDir?: string) {
  let files = getDirectoryFiles(dir);

  for (const file of files) {
    let outFile = file;
    if (outDir) {
      outFile = file.replace(dir, outDir);
    }

    if (file.endsWith(".stmpl")) {
      outFile = outFile.replace(".stmpl", "");

      templateFile(file, outFile, context.Global);
      // This is for supporting .stmpl.skip files (which are used when you don't want templateDirectory
      // to template a file so that you can control it's templating seperately - either because to support
      // conditionally templating the file, or to allow the file to be generated on the first
      // sdk generation, but not overwrite them on subsequent generations)
    } else if (file.endsWith(".skip")) {
      continue;
    } else {
      copy(file, outFile);
    }
  }
}

function getCasing(v: string): "upper" | "lower" | "mixed" {
  if (v.toUpperCase() == v) {
    return "upper";
  } else if (v.toLowerCase() == v) {
    return "lower";
  } else {
    return "mixed";
  }
}

function indentLines(lines: string[], indent: number) {
  const lineResult = [];
  lines.forEach((line, index) => {
    if (!line) {
      return;
    }

    let innerLines = line.split("\n");

    innerLines.forEach((innerLine, innerIndex) => {
      const trimmed = innerLine.trim();
      if (trimmed.length === 0) {
        innerLines[innerIndex] = trimmed;
      } else {
        innerLines[innerIndex] = templateIndent(indent) + innerLine;
      }
    });

    lineResult[index] = innerLines.join("\n");
  });

  return lineResult.join("\n");
}
registerTemplateFunc("indentLines", indentLines);

function indentString(string: string, level: number): string {
  return indentLines(string.split("\n"), level);
}
registerTemplateFunc("indentString", indentString);

function templateIndentedString(
  templateFile: string,
  context: any,
  indent: number,
): string {
  return templateString(templateFile, context)
    .split("\n")
    .map((line) => {
      if (line.trim() == "") {
        return "";
      }
      return templateIndent(indent) + line;
    })
    .join("\n");
}
registerTemplateFunc("templateIndentedString", templateIndentedString);

function templateBasicValue(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  // TODO this needs to go through extension resolution
  if (fieldDef.Type.Extensions?.ExampleUnset) {
    // to skip unset replacement and proceed normally
    // create templateUnsetValue function in your lang
    // and return undefined.
    //@ts-ignore
    if (typeof templateUnsetValue === "function") {
      //@ts-ignore
      const v = templateUnsetValue();
      if (v != undefined) {
        return v;
      } // else continue
    } else {
      return templateNullValue(fieldDef);
    }
  }

  let value = undefined;

  switch (fieldDef.Type.Type.toString()) {
    case "string":
      const envVarOverride = getEnvVarFromDirective(example);
      if (envVarOverride) {
        value = templateStringValue(
          languageSpecificEnvVarWrapping(envVarOverride, additionalContext),
          fieldDef,
          additionalContext,
        );
        break;
      }
      const fileDirective = parseFileDirectiveFromExample(
        example,
        additionalContext,
      );
      example = fileDirective.example;

      if (fileDirective.fileDirective) {
        if (
          fieldDef.Annotations?.Has("multipartForm") &&
          fieldDef.Name == "fileName"
        ) {
          value =
            fileDirective.fileDirective.overriddenFileName ??
            fileDirective.fileDirective.path.split("/").pop();
        } else {
          value = templateFileToStringValue(
            fieldDef,
            fileDirective.fileDirective.path,
            example,
            additionalContext,
          );
        }

        break;
      }

      value = templateStringValue(`${example}`, fieldDef, additionalContext);
      break;
    case "int32":
    case "bigint":
    case "integer":
      value = templateIntValue(example, fieldDef, additionalContext);
      break;
    case "float32":
    case "decimal":
    case "number":
      value = templateFloatValue(example, fieldDef, additionalContext);
      break;
    case "boolean":
      value = templateBoolValue(example, fieldDef, additionalContext);
      break;
    case "date":
      value = templateDateValue(`${example}`, fieldDef, additionalContext);
      break;
    case "date-time":
      value = templateDateTimeValue(`${example}`, fieldDef, additionalContext);
      break;
    case "uuid":
      value = templateUUIDValue(`${example}`, fieldDef, additionalContext);
      break;
    case "duration":
      value = templateDurationValue(`${example}`, fieldDef, additionalContext);
      break;
    case "bytes": {
      const fileDirective = parseFileDirectiveFromExample(
        example,
        additionalContext,
      );
      example = fileDirective.example;
      if (example === undefined) {
        // TODO this is still going to be a source of churn, need to revisit when generating x-file examples
        example = faker.string.hexadecimal({ length: 10 });
      }

      if (fileDirective.fileDirective) {
        value = templateFileToByteArrayValue(
          fieldDef,
          fileDirective.fileDirective.path,
          example,
          additionalContext,
        );
      } else {
        if (typeof example !== "string") {
          example = JSON.stringify(example);
        }
        value = templateByteValue(example, fieldDef, additionalContext);
      }
      break;
    }
    case "any":
      switch (typeof example) {
        case "string":
          value = templateStringValue(
            `${example}`,
            fieldDef,
            additionalContext,
          );
          break;
        case "number":
          if (Number.isInteger(example)) {
            value = templateIntValue(example, fieldDef, additionalContext);
          } else {
            value = templateFloatValue(example, fieldDef, additionalContext);
          }
          break;
        case "boolean":
          value = templateBoolValue(example, fieldDef, additionalContext);
          break;
        case "object":
          if (Array.isArray(example)) {
            value = templateArray(
              typeDefToFieldDef(getGenericArrayTypeDef(), fieldDef),
              example,
              additionalContext,
            );
          } else {
            value = templateMap(
              typeDefToFieldDef(getGenericMapTypeDef(), fieldDef),
              example,
              additionalContext,
            );
          }
          break;
        default:
          value = templateStringValue(
            JSON.stringify(example),
            fieldDef,
            additionalContext,
          );
          break;
      }
      break;
    default:
      throw new Error(`invalid type ${fieldDef.Type.Type.toString()}`);
  }

  if (value === undefined) {
    throw new Error(`no value received for ${fieldDef.Type.Type.toString()}`);
  }

  return postProcessBasicValue(fieldDef, value);
}

// @ts-ignore
function postProcessBasicValue(field: FieldDef, value: string): string {
  return value;
}
registerTemplateFunc("postProcessBasicValue", postProcessBasicValue);

// @ts-ignore
function templateStream(
  fieldDef: FieldDef,
  filePath: string,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateStream not implemented");
}

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateFileToByteArrayValue not implemented");
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateFileToStringValue not implemented");
}

function splitOnNewline(input: string): string[] {
  return input.split("\n");
}
registerTemplateFunc("splitOnNewline", splitOnNewline);

// @ts-ignore
function templateLanguageName(lang: string): string {
  if (lang === "cli") return "bash";
  return lang;
}
registerTemplateFunc("templateLanguageName", templateLanguageName);

function templateSecurityEnvVars(
  fieldDef: FieldDef,
  addDirective = false,
): string {
  let envVarPrefix = context.Global.Config.EnvVarPrefix?.toUpperCase() || "";
  let name = caser().ToSNAKE(fieldDef.Name).toUpperCase();
  let evar = envVarPrefix ? `${envVarPrefix}_${name}` : name;
  const security = fieldDef.Annotations?.Get("security");
  if (!security?.SecurityOption) {
    if (addDirective) {
      return `${ENV_VAR_DIRECTIVE} ${evar}`;
    } else {
      return evar;
    }
  }

  let id = `${security.SecType}:${security.SubType}`;
  if (security.FieldName) {
    id += `:${security.FieldName.toLowerCase()}`;
  }

  switch (id) {
    case "oauth2:client_credentials:clientid":
      evar = name.includes("CLIENT_ID") ? evar : `${evar}_CLIENT_ID`;
      break;
    case "oauth2:client_credentials:clientsecret":
      evar = name.includes("CLIENT_SECRET") ? evar : `${evar}_CLIENT_SECRET`;
      break;
    case "http:basic:username":
      evar = name.includes("USERNAME") ? evar : `${evar}_USERNAME`;
      break;
    case "http:basic:password":
      evar = name.includes("PASSWORD") ? evar : `${evar}_PASSWORD`;
      break;
  }

  if (addDirective) {
    return `${ENV_VAR_DIRECTIVE} ${evar}`;
  } else {
    return evar;
  }
}
registerTemplateFunc("templateSecurityEnvVars", templateSecurityEnvVars);

function templateDebugEnvVar(): string {
  let envVarPrefix = context.Global.Config.EnvVarPrefix?.toUpperCase() || "";
  return envVarPrefix ? `${envVarPrefix}_DEBUG` : "DEBUG";
}
registerTemplateFunc("templateDebugEnvVar", templateDebugEnvVar);

function containsFileDirective(example: any): boolean {
  return (
    example !== undefined &&
    typeof example === "string" &&
    example.startsWith(FILE_DIRECTIVE_KEY)
  );
}

function parseFileDirectiveFromExample(
  example: any,
  additionalContext: TemplateValueContext,
): {
  example: any;
  fileDirective?: FileDirective;
} {
  if (!containsFileDirective(example)) {
    return {
      example: example,
      fileDirective: undefined,
    };
  }

  const parts = example.split(";");
  const path = parts[0].substring(FILE_DIRECTIVE_KEY.length).trim();
  let fileName = undefined;
  if (parts.length > 1) {
    const fileNameParts = parts[1].split("=");
    if (fileNameParts.length > 1 && fileNameParts[0].trim() == "fileName") {
      fileName = fileNameParts[1].trim();
    }
  }

  return {
    example: example,
    fileDirective: {
      path: path,
      overriddenFileName: fileName,
    },
  };
}

function templateGlobalEnvVars(field: FieldDef): string {
  let envVarPrefix = context.Global.Config.EnvVarPrefix
    ? context.Global.Config.EnvVarPrefix.toUpperCase()
    : "";
  let fieldEnvVar = caser().ToSNAKE(field.Name).toUpperCase();
  if (envVarPrefix) {
    return `${envVarPrefix}_${fieldEnvVar}`;
  }

  return fieldEnvVar;
}
registerTemplateFunc("templateGlobalEnvVars", templateGlobalEnvVars);

// TODO: deprecated remove when all templates are using the job system
function templateContributingFile() {
  const job = getTemplateContributingFileJob();

  if (job) {
    templateFileJob(job);
  }
}

function getTemplateContributingFileJob(): TemplateFileJob | undefined {
  const contributingFile = "CONTRIBUTING.md";
  // CONTRIBUTING.md is automatically untracked in filetracking package, similar
  // to README.md and .gitignore.

  const contributingData = readFile(contributingFile);
  if (!contributingData) {
    return createTemplateFileJob(
      "contributing.md.stmpl",
      contributingFile,
      null,
    );
  }

  return;
}

// @ts-ignore
function templateConstOrDefaultValue(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let value = fieldDef.Default;
  let typeOfValue = "default";
  if (fieldDef.Const) {
    value = fieldDef.Const;
    typeOfValue = "const";
  }

  if (value?.Value === null) {
    return templateNullValue(fieldDef);
  }

  switch (fieldDef.Type.Type.toString()) {
    case "enum":
      if (getEnumFormat(fieldDef.Type) == "union") {
        return sanitizeEnumValue(value?.Value, fieldDef.Type.Enum.Type.Type);
      } else {
        return templateEnumValue(
          fieldDef,
          fieldDef.Type.Enum.Values.indexOf(value?.Value.toString()),
          additionalContext,
        );
      }
    case "string":
      return templateStringValue(`${value?.Value}`, fieldDef);
    case "date":
      return templateDateValue(value?.Value, fieldDef);
    case "date-time":
      return templateDateTimeValue(value?.Value, fieldDef);
    case "bigint":
    case "int32":
    case "integer":
      return templateIntValue(value?.Value, fieldDef);
    case "number":
    case "float32":
    case "decimal":
      return templateFloatValue(value?.Value, fieldDef);
    case "boolean":
      return templateBoolValue(value?.Value, fieldDef);
    case "bytes":
    default:
      throw new Error(
        `unsupported ${typeOfValue} type: '${fieldDef.Type.Type.toString()}' of value '${value?.Value}' for field '${
          fieldDef.OriginalName
        }'`,
      );
  }
}
registerTemplateFunc(
  "templateConstOrDefaultValue",
  templateConstOrDefaultValue,
);

// @ts-ignore
function templateGlobalType(field: FieldDef): string {
  // TODO: Migrate all sanitizeType to FieldDef similar to pythonv2?
  return sanitizeType(field.Type);
}

registerTemplateFunc("templateGlobalType", templateGlobalType);

function getDevContainerJobs(opts: {
  FileSuffix: string;
  Language: string;
  FileExecution: string;
  SchemaPath: string;
}): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob(`containerReadme.stmpl`, ".devcontainer/README.md", {
      FileSuffix: opts.FileSuffix,
      Language: opts.Language,
      FileExecution: opts.FileExecution,
      SchemaPath: opts.SchemaPath,
    }),
  );

  jobs.push(
    createTemplateFileJob(
      "devcontainers/config.stmpl",
      ".devcontainer/devcontainer.json",
      null,
    ),
  );

  jobs.push(
    createTemplateFileJob("containerSetup.stmpl", ".devcontainer/setup.sh", {
      FileSuffix: opts.FileSuffix,
      Language: opts.Language,
      SchemaPath: opts.SchemaPath,
    }),
  );

  return jobs;
}

function handleCommonJobs(job: Job): boolean {
  switch (job.ID) {
    case "templateFile":
      templateFileJob(job as TemplateFileJob);
      break;
    case "auxiliary":
      // Template/copy the auxiliary files
      templateAuxiliaryFiles(undefined, job.Context.sourcePath);
      break;
    case "readme":
      // @ts-ignore
      templateReadmeJob(job);
      break;
    case "usage":
      // @ts-ignore
      templateUsage();
      break;
    case "gitignore":
      templateGitignore();
      break;
    case "templateDirectory":
      templateDirectoryJob(job);
      break;
    case "templateModelDocs":
      // @ts-ignore
      templateModelDocs(job.Context.collectedTypes);
      break;
    default:
      return false;
  }

  return true;
}

// @ts-ignore
function templateIndent(indent: number): string {
  throw new Error("templateIndent not implemented");
}

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  throw new Error("templateNullValue not implemented");
}

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateFieldDeclaration not implemented");
}

// @ts-ignore
function templateFieldDelimiter(): string {
  throw new Error("templateFieldDelimiter not implemented");
}

// @ts-ignore
function templateOptionalSymbol(): string {
  throw new Error("templateOptionalSymbol not implemented");
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateType not implemented");
}

// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  opening: boolean,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateBracket not implemented");
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateEnumValue not implemented");
}

// @ts-ignore
function templateArrayValue(val: any): string {
  throw new Error("templateArrayValue not implemented");
}

// @ts-ignore
function templateMapValue(key: string, val: any): string {
  throw new Error("templateMapValue not implemented");
}

// @ts-ignore
function templateStringValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateStringValue not implemented");
}

// @ts-ignore
function templateIntValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateIntValue not implemented");
}

// @ts-ignore
function templateFloatValue(
  value: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateFloatValue not implemented");
}
// @ts-ignore
function templateBoolValue(
  val: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateBoolValue not implemented");
}

// @ts-ignore
function templateDateValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateDateValue not implemented");
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateDateTimeValue not implemented");
}

// @ts-ignore
function templateUUIDValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateUUIDValue not implemented");
}

// @ts-ignore
function templateDurationValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateDurationValue not implemented");
}

// @ts-ignore
function templateByteValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateByteValue not implemented");
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  throw new Error("sanitizeEnumValue not implemented");
}

// @ts-ignore
function templateOmittedValue(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

// @ts-ignore
function sanitizeFieldName(name: string): string {
  throw new Error("sanitizeFieldName not implemented");
}

// @ts-ignore
function sanitizeSecurityFieldName(name: string): string {
  throw new Error("sanitizeSecurityFieldName not implemented");
}

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  throw new Error("sanitizeMethodName not implemented");
}

function whereByProperty<T>(list: T[], property: string): T[] {
  return list.filter((item) => {
    if (!item) return false;
    const prop = (item as any)[property];
    return !!prop;
  });
}

function whereByPredicate<T>(list: T[], predicateName: string): T[] {
  const fn = (globalThis as any)[predicateName];
  if (typeof fn !== "function") {
    throw new Error(
      `Predicate function '${predicateName}' not found on globalThis`,
    );
  }
  return list.filter((item) => {
    if (!item) return false;
    return !!fn(item);
  });
}

registerTemplateFunc("whereByProperty", whereByProperty);
registerTemplateFunc("whereByPredicate", whereByPredicate);
