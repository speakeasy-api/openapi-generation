// @ts-ignore
declare var _: any;

unregisterTemplateFunc("templateLanguageName");

// @ts-ignore
function getSourceDirectory(): string {
  return `src/main/java/${templatePackageName().replaceAll(".", "/")}`;
}
registerTemplateFunc("getSourceDirectory", getSourceDirectory);

function getTestSourceDirectory(): string {
  return `src/test/java/${templatePackageName().replaceAll(".", "/")}`;
}

function getTestResourceDirectory(): string {
  return "src/test/resources";
}

function getBaseTestDirectory(): string {
  return `src/test`;
}
registerTemplateFunc("getBaseTestDirectory", getBaseTestDirectory);

function requiresDistinctAsyncType(type: TypeDef): boolean {
  return isResponse(type);
}

// @ts-ignore (Ignore override of function implementation)
function getModelJobs(
  types: BucketedTypes,
  path: string,
  ext: string,
): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];

  for (const [rawOutputLocation, models] of sequencedMapEntries(types)) {
    const outputLocation = sanitizeOutputLocation(rawOutputLocation);
    for (const [model, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        let modelName = t.Name;
        let modelPath = path;
        if (outputLocation) {
          modelPath += `/${outputLocation}`;
        }

        let servers = null;

        if (t.Scope.toString() == "operations" && t.IsRequest()) {
          servers = context.Global.AST.OperationServers.Get(model)[0];
        }

        let ctx = {
          Name: modelName,
          Type: t,
          Servers: servers,
          Scope: t.Scope,
          IsAsync: false,
        };
        jobs.push(
          createTemplateFileJob(
            `modelfile.${ext}.stmpl`,
            `${modelPath}/${sanitizeFileName(modelName)}.${ext}`,
            ctx,
          ),
        );

        // Generate additional files for discriminated unions (unknown type and resolver)
        if (isOneOfDiscriminated(t) && isForwardCompatibleUnions()) {
          const className = sanitizeClass(t, t.Scope, true);
          const resolverClassName = `${className}TypeIdResolver`;
          const unknownClassName = `Unknown${className}`;

          // Generate type ID resolver class
          jobs.push(
            createTemplateFileJob(
              `model-oneof-disc-resolver.${ext}.stmpl`,
              `${modelPath}/${sanitizeFileName(resolverClassName)}.${ext}`,
              ctx,
            ),
          );

          // Generate unknown type class
          jobs.push(
            createTemplateFileJob(
              `model-oneof-disc-unknown.${ext}.stmpl`,
              `${modelPath}/${sanitizeFileName(unknownClassName)}.${ext}`,
              ctx,
            ),
          );
        }

        /** Generate async implementation for response and error types when async is enabled.
         * This creates separate async model files in the async/ subdirectory to handle
         * asynchronous operations with distinct type handling for responses and errors.
         */
        if (asyncEnabled() && requiresDistinctAsyncType(t)) {
          let asyncCtx = {
            Name: modelName,
            Type: t,
            Servers: servers,
            Scope: t.Scope,
            IsAsync: true,
          };
          jobs.push(
            createTemplateFileJob(
              `modelfile.${ext}.stmpl`,
              `${modelPath}/async/${sanitizeFileName(modelName)}.${ext}`,
              asyncCtx,
            ),
          );
        }
      }
    }
  }

  return jobs;
}

// @ts-ignore
function templateSDKBuilderName(name: string): string {
  let reserved = ["client", "security", "serverURL", "server", "serverIndex"];

  let methodName = sanitizeFieldName(name);

  if (reserved.includes(methodName)) {
    methodName = sanitizeFieldName("global_" + name);
  }

  return methodName;
}

registerTemplateFunc("templateSDKBuilderName", templateSDKBuilderName);

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

  return templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    0,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    example,
    additionalContext,
  );
}

// @ts-ignore
function joinSDKOptions(options: string[]): string {
  if (options.length > 0) {
    options.unshift("");
    return indentLines(options, 4);
  }

  return "";
}

// @ts-ignore
function templateUsageSDKOptions(local: UsageContext): string {
  const options: string[] = [];
  let hasSecurity = local.Operation.Security != undefined;

  const ctx: TemplateValueContext = {
    isTest: local.Test && true,
    usageContext: local,
    operation: local.Operation,
    test: local.Test,
  };

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal && scope.Feature != "") {
        const feature = scope.Feature.toString();

        switch (feature) {
          case "security":
            if (!hasSecurity) {
              const globalSecurity = getGlobalSecurity(scope.Value, ctx);

              if (globalSecurity) {
                options.push(globalSecurity);
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            const hasOperationServers =
              local.Operation?.Servers?.Servers?.length > 0;
            if (!hasOperationServers) {
              options.push(
                `.serverURL\(${getUsageServerUrl(local, scope, ctx)}\)`,
              );
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              const sdk = context.Global.AST.MainSDK;
              options.push(
                `.server\(${sanitizeFileName(
                  sdk.Type.Name,
                )}.AvailableServers.${templateConstName(server.ID, "")}\)`,
              );
            } else if (server.Index !== undefined) {
              options.push(`.serverIndex\(${server.Index}\)`);
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const builderName = templateSDKBuilderName(v.Name);
                options.push(
                  `.${builderName}(${getUsageServerVariableValue(v)})`,
                );
              });
            }

            break;
          }
          case "retries": {
            options.push(templateString("usage/retries.stmpl", {}));
            break;
          }
          case "parameter": {
            const parameter = scope.Value as ParameterUsage;

            const value = templateValueAsRequired(
              parameter.field,
              getExampleValue(parameter.example),
              ctx,
            );

            if (value === "") {
              return;
            }

            options.push(
              `.${templateSDKBuilderName(parameter.field.Name)}(${value})`,
            );

            break;
          }
          case "http_client": {
            options.push(`.client(${templateHTTPClient(scope.Value)})`);
            break;
          }
        }
      }
    });
  }

  if (!hasSecurity && opUsesGlobalSecurity(local.Operation)) {
    const globalSecurity = getGlobalUsageSecurity(local);
    if (globalSecurity) {
      options.push(globalSecurity);
    }
  }

  return joinSDKOptions(options);
}

registerTemplateFunc("templateUsageSDKOptions", templateUsageSDKOptions);

function builderOf(typeDef: TypeDef) {
  let type = sanitizeTypeMandatory(typeDef);
  return `${javaImport(type)}.builder()`;
}

// @ts-ignore
function templateHTTPClient(value: HTTPClientUsage): string {
  switch (value.type) {
    case "test":
      return `testHttpClient`;
    default:
      throw new Error(`unsupported http client type: ${value.type}`);
  }
}

function templateCommentsForField(field: FieldDef, indent: number = 1): string {
  if (field.Comments) {
    const lines = [""];
    lines.push(templateCommentsStripped(field.Comments, null, indent, "field"));
    // May seem confusing to add the @deprecated tag here but it's convenient
    // to plug this here.
    if (field.Comments.Deprecated) {
      lines.push(`${templateIndent(indent)}@${javaImportDeprecated()}`);
    }
    return lines.join("\n");
  }

  return "";
}

registerTemplateFunc("templateCommentsForField", templateCommentsForField);

// @ts-ignore
function templateCommentsStripped(
  comments: CommentDef | null,
  title: string | null,
  indent: number,
  type: "field" | "method" | "class" | "enum" | null = null,
  omitDocs: boolean = false,
  operation: Operation | null = null,
  operationParams: JavaParam[] | null = null,
  returnsComment: string = "The response from the API call",
  methodThrows: boolean = true,
): string {
  return templateComments(
    comments,
    title,
    indent,
    type,
    omitDocs,
    operation,
    operationParams,
    returnsComment,
    methodThrows,
    true,
  );
}
registerTemplateFunc("templateCommentsStripped", templateCommentsStripped);

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  title: string | null,
  indent: number,
  type: "field" | "method" | "class" | "enum" | null = null,
  omitDocs: boolean = false,
  operation: Operation | null = null,
  operationParams: JavaParam[] | null = null,
  returnsComment: string = "The response from the API call",
  methodThrows: boolean = true,
  omitNewLine: boolean = false,
  sentencesPerParagraph: number = 3,
  additionalNotes: string = "",
): string {
  if (comments) {
    // collect lines to write to the javadoc method description section
    let descriptionLines: string[] = [];

    if (title) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      const sanitizedTitle = sanitizeComments(title.trim());
      const wrappedTitle = wrapTextAtWordBoundaries(
        sanitizedTitle,
        100,
        sentencesPerParagraph,
      );
      descriptionLines.push(...wrappedTitle);
    }

    if (comments.Summary) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      const sanitizedSummary = sanitizeComments(comments.Summary.trim());
      const wrappedSummary = wrapTextAtWordBoundaries(
        sanitizedSummary,
        100,
        sentencesPerParagraph,
      );
      descriptionLines.push(...wrappedSummary);
    }

    if (comments.Description) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      const sanitizedDescription = sanitizeComments(
        comments.Description.trim(),
      );
      const wrappedDescription = wrapTextAtWordBoundaries(
        sanitizedDescription,
        100,
        sentencesPerParagraph,
      );
      descriptionLines.push(...wrappedDescription);
    }

    if (comments.ExternalDocs && !omitDocs) {
      let externalDocs;
      if (comments.ExternalDocs.Description) {
        const sanitizedDescription = sanitizeComments(
          comments.ExternalDocs.Description,
        );
        externalDocs = `<a href="${comments.ExternalDocs.URL}">${sanitizedDescription}</a>`;
      } else {
        externalDocs = `<a href="${comments.ExternalDocs.URL}">${comments.ExternalDocs.URL}</a>`;
      }
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      descriptionLines.push(externalDocs);
    }

    if (additionalNotes) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      const wrappedRemark = wrapTextAtWordBoundaries(
        sanitizeComments(additionalNotes),
        100,
        sentencesPerParagraph,
      );
      descriptionLines.push(...wrappedRemark);
    }

    // now ensure that secondary paragraphs start with <p> tags (no closing tag required by javadoc).
    // If a blank line encountered then that is a paragraph separator.

    // the aim is to turn this
    // ```
    // hello there
    // you
    //
    // blah blah
    // more blah
    // ```
    // into these lines
    // ```
    // hello there
    // you
    //
    // <p>blah blah
    // more blah
    // ```
    const linesWithPrTags: string[] = [];
    let firstPara = true;
    let paraStart = true;
    for (const item of descriptionLines) {
      if (item.trim().length == 0) {
        paraStart = true;
      } else if (paraStart) {
        if (firstPara) {
          linesWithPrTags.push(item);
          firstPara = false;
        } else if (paraStart) {
          linesWithPrTags.push("");
          // note that javadoc does not close <p> tags
          linesWithPrTags.push("<p>" + item);
        }
        paraStart = false;
      } else {
        linesWithPrTags.push(item);
      }
    }

    let lines: string[] = [...linesWithPrTags];

    if (lines.length == 0 && (!comments || !comments.Deprecated)) {
      return "";
    }

    // add javadoc tags (@*)
    // we use first to ensure there is a blank line before these @ items (convention)
    let first: boolean = true;
    if (operation) {
      for (const param of operationParams) {
        if (!param.IncludedInSignature) {
          continue;
        }

        let fieldComments: string[] = [];

        if (param.Comment) {
          fieldComments = param.Comment.split("\n");
        } else {
          fieldComments.push("");
        }
        if (first) {
          lines.push("");
          first = false;
        }
        fieldComments = fieldComments.map((value, index) =>
          index == 0 ? `@param ${param.Name} ${value}` : `        ${value}`,
        );
        lines.push(...fieldComments);
      }
      if (first) {
        lines.push("");
        first = false;
      }
      // Check if returnsComment contains a type that should be wrapped in {@code ...}
      let processedReturnsComment = returnsComment;

      // If the comment contains angle brackets (indicating a generic type), wrap the entire type in {@code ...}
      const typeMatch = returnsComment.match(
        /^(.*?)([A-Za-z_][A-Za-z0-9_]*<[^>]*>)(.*)$/,
      );
      if (typeMatch) {
        const [, beforeType, fullTypeWithGenerics, afterType] = typeMatch;
        processedReturnsComment = `${beforeType}{@code ${fullTypeWithGenerics
          .replace(/&lt;/g, "<")
          .replace(/&gt;/g, ">")}}${afterType}`;
      }

      // Now sanitize the processed comment (this will leave {@code ...} intact)
      const sanitizedReturnsComment = sanitizeComments(processedReturnsComment);

      // Split long @return comments into multiple lines for better readability
      if (sanitizedReturnsComment.length > 60) {
        const parts = sanitizedReturnsComment.split(". ");
        lines.push(`@return ${parts[0]}${parts.length > 1 ? "." : ""}`);
        for (let i = 1; i < parts.length; i++) {
          lines.push(`${parts[i]}${i < parts.length - 1 ? "." : ""}`);
        }
      } else {
        lines.push(`@return ${sanitizedReturnsComment}`);
      }
      if (methodThrows) {
        if (first) {
          lines.push("");
          first = false;
        }
        // TODO make this more precise based on operation error responses
        lines.push(`@throws RuntimeException subclass if the API call fails`);
      }
    }

    if (comments.Deprecated) {
      let deprecated = `@deprecated ${type}: ${comments.DeprecationMessage}.`;

      if (comments.DeprecationReplacement) {
        const replacement = sanitizeDeprecationReplacement(
          comments.DeprecationReplacement,
          type as "field" | "method" | "class",
        );
        if (replacement) {
          deprecated += ` Use ${replacement} instead.`; // TODO: use @link javadoc annotation eventually
        }
      }
      if (first) {
        lines.push("");
        first = false;
      }
      lines.push(deprecated);
    }

    lines = lines.map((line, _) => ` * ${line}`);

    lines.unshift("/**");
    lines.push(" */");

    return cleanJavadoc(indentLines(lines, indent) + (omitNewLine ? "" : "\n"));
  }

  return "";
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map(
    (f) => `{@link Security#${sanitizeFieldName(f.Name)}}`,
  );

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set via the {@code security} builder method when initializing the SDK.`
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
function templateCommentsWithParagraphs(
  comments: CommentDef | null,
  indent: number,
  sentencesPerParagraph: number = 3,
): string {
  return templateComments(
    comments,
    null, // title
    indent,
    "class", // type
    false, // omitDocs
    null, // operation
    null, // operationParams
    "The response from the API call", // returnsComment
    true, // methodThrows
    false, // omitNewLine
    sentencesPerParagraph,
  );
}

registerTemplateFunc(
  "templateCommentsWithParagraphs",
  templateCommentsWithParagraphs,
);

// @ts-ignore
function templateBlockComment(
  comments: CommentDef | null,
  title: string | null,
  indent: number,
): string {
  if (comments) {
    // collect lines to write to the block comment
    let descriptionLines: string[] = [];

    if (title) {
      descriptionLines.push(...sanitizeComments(title.trim()).split("\n"));
    }

    if (comments.Summary) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      descriptionLines.push(
        ...sanitizeComments(comments.Summary.trim()).split("\n"),
      );
    }

    if (comments.Description) {
      if (descriptionLines.length > 0) {
        descriptionLines.push("");
      }
      descriptionLines.push(
        ...sanitizeComments(comments.Description.trim()).split("\n"),
      );
    }

    if (descriptionLines.length == 0) {
      return "";
    }

    let lines: string[] = [...descriptionLines];
    lines = lines.map((line, _) => ` * ${line}`);

    lines.unshift("/*");
    lines.push(" */");

    return cleanJavadoc(indentLines(lines, indent));
  }

  return "";
}

registerTemplateFunc("templateBlockComment", templateBlockComment);

function cleanJavadoc(doc: string): string {
  // recursive template rendering will think double braces are
  // templating commands so we render them as html entities instead
  return doc
    .replaceAll("{{", "&lbrace;&lbrace;")
    .replaceAll("}}", "&rbrace;&rbrace;");
}

function wrapTextAtWordBoundaries(
  text: string,
  maxLength: number,
  sentencesPerParagraph: number = 3,
): string[] {
  // Step 1: Split by line breaks to identify paragraphs (both single \n and double \n\n)
  const paragraphs = text.split(/\n/);
  const processedLines: string[] = [];

  for (
    let paragraphIndex = 0;
    paragraphIndex < paragraphs.length;
    paragraphIndex++
  ) {
    const paragraph = paragraphs[paragraphIndex].trim();

    if (paragraph.length === 0) {
      // Preserve empty lines as paragraph separators
      processedLines.push("");
      continue;
    }

    // Split paragraph into sentences
    const sentences = paragraph
      .split(/([.!?]+)(?=\s|$)/)
      .filter((s) => s.trim().length > 0);
    const completeSentences: string[] = [];

    // Reconstruct complete sentences (text + punctuation)
    for (let i = 0; i < sentences.length; i += 2) {
      const sentence = sentences[i]?.trim();
      const punctuation = sentences[i + 1] || "";
      if (sentence && sentence.length > 0) {
        completeSentences.push(sentence + punctuation);
      }
    }

    // Group sentences into sub-paragraphs based on sentencesPerParagraph
    const subParagraphs: string[] = [];
    for (let i = 0; i < completeSentences.length; i += sentencesPerParagraph) {
      const group = completeSentences.slice(i, i + sentencesPerParagraph);
      if (i === 0) {
        // First sub-paragraph
        subParagraphs.push(group.join(" "));
      } else {
        // Subsequent sub-paragraphs - add blank line for separation
        subParagraphs.push("", group.join(" "));
      }
    }

    // Line wrap each sub-paragraph
    for (const subParagraph of subParagraphs) {
      if (subParagraph.trim().length === 0) {
        processedLines.push("");
        continue;
      }

      if (subParagraph.length <= maxLength) {
        processedLines.push(subParagraph);
        continue;
      }

      // Wrap long sub-paragraphs
      let currentLine = "";
      const words = subParagraph.split(" ");

      for (const word of words) {
        if (currentLine.length + word.length + 1 > maxLength) {
          if (currentLine.length > 0) {
            processedLines.push(currentLine.trim());
            currentLine = word;
          } else {
            processedLines.push(word);
          }
        } else {
          currentLine += (currentLine.length > 0 ? " " : "") + word;
        }
      }

      if (currentLine.length > 0) {
        processedLines.push(currentLine.trim());
      }
    }
  }

  return processedLines;
}

function isRequiredArg(field: FieldDef | JavaParam): boolean {
  if (context.Global.Config.NullFriendlyParameters) {
    return !(field.Optional || field.Nullable || field.Default);
  }

  return !(field.Optional || field.Nullable);
}

function isBinaryStream(typeDef: TypeDef): boolean {
  const typeStr = typeDef.Type?.toString();
  return typeStr == "request-stream";
}
registerTemplateFunc("isBinaryStream", isBinaryStream);

function templateConstructorRequiredArgs(
  fields: FieldDef[],
  isAsync: boolean = false,
): string {
  function render(field: FieldDef): string {
    return `${javaImportConstructorParam(field, isAsync)} ${sanitizeFieldName(
      field.Name,
    )}`;
  }
  const requiredFields = fields.filter(isRequiredArg);
  const out = [];
  for (const field of requiredFields) {
    if (!isRequiredArg(field)) {
      continue;
    }
    out.push(`\n${templateIndent(2)}${render(field)}`);
  }

  return out.join(",");
}

function templateConstructorParams(
  fields: FieldDef[],
  {
    jsonAnnot = true,
    isAsync = false,
    useBytes = false,
  }: {
    jsonAnnot?: boolean;
    isAsync?: boolean;
    useBytes?: boolean;
  } = {},
): string {
  const hasJsonProperty =
    fields
      .filter((field) => field.Annotations)
      .map((field) => field.Annotations)
      .filter((a) => a.Has("json")).length > 0;

  const annotation = (fieldName: String) =>
    jsonAnnot && hasJsonProperty
      ? `@${javaImportJacksonAnn("JsonProperty")}(\"${fieldName}\") `
      : "";

  function render(field: FieldDef): string {
    let fieldType = javaImportConstructorParam(field, isAsync);
    if (isBinaryStream(field.Type) && useBytes) {
      // hack to force Blob -> byte[]
      fieldType = fieldType.replace(
        `${javaImportLocal("utils.Blob")}`,
        "byte[]",
      );
    }
    return `${annotation(propertyName(field))}${fieldType} ${sanitizeFieldName(
      field.Name,
    )}`;
  }
  const result = [];
  for (const field of fields) {
    result.push(`\n${templateIndent(3)}${render(field)}`);
  }

  return result.join(",");
}

function propertyName(field: FieldDef): string {
  // note that in a custom Error object the rawResponse field is optional
  // and is never present at deserialization time so its name is irrelevant.
  // field.OriginalName is blank for rawResponse so we default to field.Name
  // when blank
  // TODO when field.OriginalName is fixed so is always present then this method
  // can just return field.OriginalName
  return escapeString(field.OriginalName ? field.OriginalName : field.Name);
}

registerTemplateFunc("propertyName", propertyName);

function constructorPreconditions(fields: FieldDef[]): string {
  const lines = [];
  for (const field of fields) {
    const sanitizedFieldName = sanitizeFieldName(field.Name);
    if (field.Type.Type.toString() === "map" && isRequiredArg(field)) {
      lines.push(
        `${sanitizedFieldName} = ${javaImportUtils()}.emptyMapIfNull(${sanitizedFieldName});`,
      );
    }
    if (context.Global.Config.NullFriendlyParameters) {
      continue;
    }
    lines.push(
      `${javaImportUtils()}.checkNotNull(${sanitizedFieldName}, "${sanitizedFieldName}");`,
    );
  }

  return lines.join("\n        ");
}

function initializeField(field: FieldDef, receiver: string = "this"): string {
  if (context.Global.Config.NullFriendlyParameters) {
    const sanitizedFieldName = sanitizeFieldName(field.Name);
    if (hasConstValue(field)) {
      if (field.Nullable) {
        return `${receiver}.${sanitizedFieldName} = ${javaImportJsonNullable()}.of(Builder.${singletonValueGet(
          sanitizedFieldName,
        )});`;
      }
      return `${receiver}.${sanitizedFieldName} = Builder.${singletonValueGet(
        sanitizedFieldName,
      )};`;
    }
    if (field.IsAdditionalProperties) {
      return `${receiver}.${sanitizedFieldName} = new ${javaImportHashMap()}<>();`;
    }
    const { body } = getConstructorMetadata(parameterForField(field));
    return `${receiver}.${sanitizeFieldName(field.Name)} = ${body.join(
      "\n            ",
    )};`;
  }

  const sanitizedFieldName = sanitizeFieldName(field.Name);
  if (hasConstValue(field)) {
    return `${receiver}.${sanitizedFieldName} = Builder.${singletonValueGet(
      sanitizedFieldName,
    )};`;
  }
  if (field.IsAdditionalProperties) {
    return `${receiver}.${sanitizedFieldName} = new ${javaImportHashMap()}<>();`;
  }

  return `${receiver}.${sanitizedFieldName} = ${sanitizedFieldName};`;
}

function isOneOf(type: TypeDef): boolean {
  return type.Type == "union" && type.AssociatedTypes.length > 0;
}

function isOneOfDiscriminated(type: TypeDef): boolean {
  return isOneOf(type) && type.Discriminator != undefined;
}

registerTemplateFunc("isOneOfDiscriminated", isOneOfDiscriminated);

function isOneOfNonDiscriminated(type: TypeDef): boolean {
  return isOneOf(type) && type.Discriminator == undefined;
}

registerTemplateFunc("isOneOfNonDiscriminated", isOneOfNonDiscriminated);

function isForwardCompatibleUnions(): boolean {
  // Post-normalize the value is a boolean. The string checks cover paths where
  // resolveConfig is skipped or fails (non-fatal): raw gen.yaml values may be
  // the quoted string "true" or the legacy materialized "tagged-and-untagged".
  const mode = context.Global.Config.ForwardCompatibleUnionsByDefault;
  return mode === true || mode === "true" || mode === "tagged-and-untagged";
}

registerTemplateFunc("isForwardCompatibleUnions", isForwardCompatibleUnions);

function checkArgument(
  b: boolean,
  message: string = "failed precondition",
): void {
  if (!b) {
    throw new Error(message);
  }
}

function discriminatorFieldName(type: TypeDef): string {
  checkArgument(isOneOfDiscriminated(type));
  return sanitizeFieldName(type.Discriminator.TypePropertyName);
}

registerTemplateFunc("discriminatorFieldName", discriminatorFieldName);

function oneOfDiscriminatedInterfaceMethod(type: TypeDef): string {
  checkArgument(isOneOfDiscriminated(type));
  const fieldName = discriminatorFieldName(type);
  return `    ${javaImportString()} ${fieldName}();`;
}

registerTemplateFunc(
  "oneOfDiscriminatedInterfaceMethod",
  oneOfDiscriminatedInterfaceMethod,
);

namespace DiscriminatedOneOf {
  // for a typeDef the map contains the oneOf TypeDefs that it is a member of
  // key is outputLocation:type.Name
  let map: Map<string, Set<TypeDef>> = null;
  const emptySet = new Set<TypeDef>();

  export function membership(type: TypeDef): Set<TypeDef> {
    if (map == null) {
      map = createMembershipMap();
    }
    const v = map.get(type.OutputLocation + ":" + type.Name);
    if (v == undefined) {
      return emptySet;
    } else {
      return v;
    }
  }

  export function discriminatorFieldNames(type: TypeDef): string[] {
    return Array.from(membership(type)).map((x) => discriminatorFieldName(x));
  }

  // creates a map from a type name to the discriminated oneOf TypeDefs it
  // is a member of. We choose a string as the key here because of the possibility
  // of truncated types (circular references).
  function createMembershipMap(): Map<string, Set<TypeDef>> {
    const map = new Map<string, Set<TypeDef>>();
    const types = context.Global.AST.BucketedTypes;

    for (const [outputLocation, models] of sequencedMapEntries(types)) {
      for (const [, types] of sequencedMapEntries(models)) {
        types
          .filter((x) => x.Type == "union" && x.Discriminator != undefined)
          .forEach((x) => {
            x.AssociatedTypes.forEach((y) => {
              // we assume that if truncated that the outputLocation is
              // the same as the original type. TODO What else can I do here?
              const location = y.Truncated ? outputLocation : y.OutputLocation;
              const key = location + ":" + y.Name;
              let v = map.get(key);
              if (v == undefined) {
                v = new Set<TypeDef>();
                map.set(key, v);
              }
              v.add(x);
            });
          });
      }
    }
    return map;
  }
}

function isDiscriminatorFieldName(type: TypeDef, name: string): boolean {
  const names = DiscriminatedOneOf.discriminatorFieldNames(type);
  return names.includes(sanitizeFieldName(name));
}

registerTemplateFunc("isDiscriminatorFieldName", isDiscriminatorFieldName);

function discriminatedInterfaceExtends(type: TypeDef, scope: string): string {
  const set = DiscriminatedOneOf.membership(type);
  const items = Array.from(set).map((typ) => importedClass(typ));

  if (items.length > 0) {
    return " extends " + items.join(", ");
  } else {
    return "";
  }
}

registerTemplateFunc(
  "discriminatedInterfaceExtends",
  discriminatedInterfaceExtends,
);

function modelImplements(
  type: TypeDef,
  scope: string,
  isAsync: boolean = false,
): string {
  const items: string[] = [];
  if (isResponse(type)) {
    items.push(
      isAsync
        ? javaImportLocal(`utils.AsyncResponse`)
        : javaImportLocal(`utils.Response`),
    );
  }
  const set = DiscriminatedOneOf.membership(type);
  Array.from(set).forEach((typ) => {
    items.push(importedClass(typ));
  });
  const hasSecurity = hasSecurityAnnotation(type);
  if (hasSecurity) {
    items.push(javaImportLocal(`utils.HasSecurity`));
  }
  if (items.length > 0) {
    return " implements " + items.join(", ");
  } else {
    return "";
  }
}

registerTemplateFunc("modelImplements", modelImplements);

function modelSimpleClassName(type: TypeDef): string {
  return type.Type == "error" ? "Data" : sanitizeClassName(type.Name);
}
registerTemplateFunc("modelSimpleClassName", modelSimpleClassName);

function hasSecurityAnnotation(type: TypeDef): boolean {
  return (
    undefined !=
    type.Fields.find((f) => {
      for (const a of f.Annotations) {
        if (a.Type().toString() == "security") {
          return true;
        }
      }
    })
  );
}

function oneOfDiscriminatedClassAnnotations(
  type: TypeDef,
  scope: string,
): string {
  if (isForwardCompatibleUnions()) {
    return oneOfDiscriminatedClassAnnotationsWithUnknownHandling(type, scope);
  } else {
    return oneOfDiscriminatedClassAnnotationsLegacy(type, scope);
  }
}

function oneOfDiscriminatedClassAnnotationsWithUnknownHandling(
  type: TypeDef,
  scope: string,
): string {
  const className = sanitizeClass(type, scope, true);
  const resolverClass = `${className}TypeIdResolver`;
  const unknownClassName = `Unknown${className}`;

  // Use custom type ID resolver that handles unknown union sub-types for forward compatibility
  let result = `@${javaImportJacksonAnn("JsonTypeInfo")}(
        use = ${javaImportJacksonAnn("JsonTypeInfo.Id")}.CUSTOM,
        property = "${type.Discriminator.TypePropertyName}",
        include = ${javaImportJacksonAnn("JsonTypeInfo.As")}.EXISTING_PROPERTY,
        visible = true,
        defaultImpl = ${unknownClassName}.class
)`;

  result += `\n@${javaImport(
    "com.fasterxml.jackson.databind.annotation.JsonTypeIdResolver",
  )}(${resolverClass}.class)`;
  return result;
}

function oneOfDiscriminatedClassAnnotationsLegacy(
  type: TypeDef,
  scope: string,
): string {
  // Use the old approach with JsonSubTypes annotations
  let result = `@${javaImportJacksonAnn(
    "JsonTypeInfo",
  )}(use = ${javaImportJacksonAnn("JsonTypeInfo.Id")}.NAME, property = "${
    type.Discriminator.TypePropertyName
  }", include = ${javaImportJacksonAnn(
    "JsonTypeInfo.As",
  )}.EXISTING_PROPERTY, visible = true)`;

  // Generate annotations by iterating through discriminator mappings first
  // This ensures each mapping gets its own annotation, even if the same type appears multiple times
  const members = type.Discriminator.Mapping.map((m) => {
    // Find the associated type for this mapping
    const associatedType = type.AssociatedTypes.find(
      (x) => x.OriginalName === m.Type.Name,
    );
    if (!associatedType) {
      // This shouldn't happen in a well-formed discriminator, but handle gracefully
      return null;
    }

    const cls = importedClass(associatedType);
    return `\n    @${javaImportJacksonAnn(
      "JsonSubTypes.Type",
    )}(value = ${cls}.class, name="${m.Name}")`;
  })
    .filter(Boolean)
    .join(",");

  result += `\n@${javaImportJacksonAnn("JsonSubTypes")}({${members}})`;
  return result;
}

registerTemplateFunc(
  "oneOfDiscriminatedClassAnnotations",
  oneOfDiscriminatedClassAnnotations,
);

function collectConstructorFields(scope: string, type: TypeDef): FieldDef[] {
  const fields: FieldDef[] = [];
  // if is oneOf then we add a field
  if (isOneOfNonDiscriminated(type)) {
    const objectType = createZeroTypeDef();
    objectType.Name = "Object";
    objectType.OriginalName = "Object";
    objectType.Type = dataType("any");
    objectType.Scope = <Scope>scope;

    const fieldDef: FieldDef = {
      Name: "value",
      OriginalName: "value",
      Nullable: false,
      Optional: false,
      IsAdditionalProperties: false,
      IsResponseHeaders: false,
      IsResponseMetadata: false,
      ErrorMessage: false,
      Type: objectType,
      Clone: (): FieldDef => {
        return { ...fieldDef };
      },
      GetID: () => `fieldName:value ${objectType.GetRegistrationIDOrType()}`,
    };
    fields.push(fieldDef);
    if (!type.Fields) {
      logger().Warn(
        "property fields next to a oneOf type are not supported yet",
      );
    }
  } else {
    fields.push(...removeDerivedAndUndesiredFields(type));
  }
  return fields;
}

function templateConstructorAll(
  scope: string,
  type: TypeDef,
  isAsync: boolean = false,
): string {
  const className = modelSimpleClassName(type);
  const fields = collectConstructorFields(scope, type);

  const fieldsWithoutAdditionalProperties =
    removeAdditionalPropertiesField(fields);
  if (!fieldsWithoutAdditionalProperties) {
    // noArgsConstructor
    return `public ${className}(){}`;
  }

  if (exceedsJvmParamLimit(fields)) {
    const initLines = fields
      .filter((f) => hasConstValue(f) || f.IsAdditionalProperties)
      .map((f) => initializeField(f))
      .join("\n        ");
    if (initLines.length === 0) {
      return `public ${className}() {}`;
    }
    return [
      `${templateIndent(1)}public ${className}() {`,
      `${templateIndent(2)}${initLines}`,
      `    }`,
    ].join("\n");
  }

  const nonConstantFields = nonConstFields(fieldsWithoutAdditionalProperties);
  const args = templateConstructorParams(nonConstantFields, { isAsync });
  const preconditions = constructorPreconditions(nonConstantFields);

  let inits = fields.map((f) => initializeField(f)).join("\n        ");

  const constructor = [
    `@${javaImportJacksonAnn("JsonCreator")}`,
    `${templateIndent(1)}public ${className}(${args}) {`,
    preconditions.length > 0
      ? `${templateIndent(2)}${preconditions}`
      : undefined,
    inits.length > 0 ? `${templateIndent(2)}${inits}` : undefined,
    `    }`,
  ];
  let constructors = constructor.filter(Boolean).join("\n");

  // if any of the params are optional then add a constructor overload
  // with just required fields that delegates to the primary constructor
  const hasOptionalParams = nonConstantFields.some((f) => !isRequiredArg(f));
  if (hasOptionalParams) {
    const params = templateConstructorParams(
      nonConstantFields.filter(isRequiredArg),
      {
        isAsync,
        jsonAnnot: false,
      },
    );
    const args = groupArgs(
      nonConstantFields.map((v) => getFalseyValue(v)),
    ).join(",\n            ");
    constructors += `
    
    public ${className}(${params}) {
        this(${args});
    }`;
  }
  return constructors;
}

registerTemplateFunc("templateConstructorAll", templateConstructorAll);

function templateBuilderBuild(
  scope: string,
  type: TypeDef,
  indent: number = 3,
): string {
  const className = modelSimpleClassName(type);
  const fields = collectConstructorFields(scope, type);
  const nonConstantFields = nonConstFields(
    removeAdditionalPropertiesField(fields),
  );

  const nullFriendly = !!context.Global.Config.NullFriendlyParameters;
  const additionalPropertiesChain = templateSetAdditionalProperties(fields, 4);
  const setDefaults = nullFriendly ? "" : `\n${templateSetDefaults(fields)}`;

  if (exceedsJvmParamLimit(fields)) {
    // Instantiate via no-args ctor and assign fields directly.
    const receiver = "_instance"; // local var that initializeField writes to
    const lines: string[] = [];
    if (setDefaults) {
      lines.push(setDefaults);
    }
    lines.push(`${className} ${receiver} = new ${className}();`);
    for (const f of nonConstantFields) {
      const name = sanitizeFieldName(f.Name);
      if (!nullFriendly) {
        lines.push(`${javaImportUtils()}.checkNotNull(${name}, "${name}");`);
      }
      lines.push(initializeField(f, receiver));
    }

    lines.push(`return ${receiver}${additionalPropertiesChain};`);
    return indentLines(lines, indent);
  }

  const args = templateFieldsIndentedArgs(nonConstantFields, indent + 1);
  return `${setDefaults}
${templateIndent(indent)}return new ${className}(\n${templateIndent(
    indent + 1,
  )}${args})${additionalPropertiesChain};`;
}

registerTemplateFunc("templateBuilderBuild", templateBuilderBuild);

function errorMessageExpression(type: TypeDef): string {
  if (type.Type.toString() != "error") {
    return "";
  }
  const defaultErrorMessage = "API error occurred";
  var result: string;
  // if the type is an error then we need to call super(message) in
  // the constructor and honour the x-speakeasy-error-message extension
  const errorMessageFields = type.Fields.filter((f) => f.ErrorMessage);
  if (errorMessageFields.length == 0) {
    result = `"${defaultErrorMessage}"`;
  } else {
    if (errorMessageFields.length > 1) {
      logger().Warn(
        type.Name +
          " has more than one field annotated with x-speakeasy-error-message. The first one will be used to construct the super call.",
      );
    }
    const messageField = errorMessageFields[0];
    // we wrap with Utils.valueOrElse to unwrap Optional or JSONNullable
    result = `${javaImportUtils()}.valueOrElse(data == null ? null : data.${sanitizeFieldName(
      messageField.Name,
    )}(), "${defaultErrorMessage}")`;
  }
  return result;
}

registerTemplateFunc("errorMessageExpression", errorMessageExpression);

function groupArgs(args: string[], max: number = 3): string[] {
  const groupedArgs = [];
  for (let i = 0; i < args.length; i += max) {
    groupedArgs.push(args.slice(i, i + max));
  }
  return groupedArgs.map((group) => group.join(", "));
}
registerTemplateFunc("groupArgs", groupArgs);

function getFalseyValue(field: FieldDef): string {
  if (context.Global.Config.NullFriendlyParameters) {
    if (isRequiredArg(field)) {
      return `${sanitizeFieldName(field.Name)}`;
    }
    return "null";
  }

  if (field.Optional && field.Nullable) {
    return `${javaImportJsonNullable()}.undefined()`;
  }

  if (field.Optional || field.Nullable) {
    return `${javaImportOptional()}.empty()`;
  }

  return `${sanitizeFieldName(field.Name)}`;
}

function removeAdditionalPropertiesField(fields: FieldDef[]): FieldDef[] {
  return fields.filter((f) => !f.IsAdditionalProperties);
}
registerTemplateFunc(
  "removeAdditionalPropertiesField",
  removeAdditionalPropertiesField,
);

const RESERVED_ERROR_CLASS_FIELD_NAMES = new Set<string>([
  "message",
  "code",
  "body",
  "rawResponse",
  "data",
  "fillInStackTrace",
  "getCause",
  "getLocalizedMessage",
  "getMessage",
  "getStackTrace",
  "getSuppressed",
  "printStackTrace",
]);

function errorClassFields(type: TypeDef): FieldDef[] {
  return nonConstFields(removeDerivedAndUndesiredFields(type)).filter(
    (f) => !RESERVED_ERROR_CLASS_FIELD_NAMES.has(sanitizeFieldName(f.Name)),
  );
}
registerTemplateFunc("errorClassFields", errorClassFields);

function removeDerivedAndUndesiredFields(type: TypeDef): FieldDef[] {
  // For example in a response of type `text/event-stream` the response
  // is a derived function of the `rawResponse` (InputStream) field.
  return (
    type.Fields.filter((f) => f.Type.Type.toString() !== "event-stream")
      // don't want HttpResponse field appearing in error data class (is in parent object) which is used in errors v2
      .filter(
        (f) => !(type.Type == "error" && f.Type.Type.toString() == "response"),
      )
  );
}
registerTemplateFunc(
  "removeDerivedAndUndesiredFields",
  removeDerivedAndUndesiredFields,
);

function templateSetAdditionalProperties(
  fields: FieldDef[],
  indent: number = 4,
): string {
  return fields
    .filter((f) => f.IsAdditionalProperties)
    .map(
      (f) =>
        `\n${templateIndent(indent)}.with${sanitizeClassName(
          f.Name,
        )}(${sanitizeFieldName(f.Name)})`,
    )
    .join("");
}
registerTemplateFunc(
  "templateSetAdditionalProperties",
  templateSetAdditionalProperties,
);

function templateFieldsEqualityExpression(fields: FieldDef[]): string {
  return fields
    .map((field) => {
      const fieldName = sanitizeFieldName(field.Name);
      return `${javaImportUtils()}.enhancedDeepEquals(this.${fieldName}, other.${fieldName})`;
    })
    .join(" &&\n            ");
}

registerTemplateFunc(
  "templateFieldsEqualityExpression",
  templateFieldsEqualityExpression,
);

function templateFieldsIndentedArgs(
  fields: FieldDef[],
  indent: number,
): string {
  return groupArgs(
    fields.map((field) => `${sanitizeFieldName(field.Name)}`),
    3,
  ).join(",\n" + "    ".repeat(indent));
}

registerTemplateFunc("templateFieldsIndentedArgs", templateFieldsIndentedArgs);

function templateToStringExpression(
  className: String,
  fields: FieldDef[],
): string {
  const args = fields
    .map((field) => {
      const name = sanitizeFieldName(field.Name);
      return `,\n                "${name}", ${name}`;
    })
    .join("");
  return `${javaImportUtils()}.toString(${className}.class${args})`;
}

registerTemplateFunc("templateToStringExpression", templateToStringExpression);

function getUniqueOperations(sdk: SDK): Operation[] {
  const operations: Operation[] = [];
  // could use a Set for operationIds but my
  // editor doesn't recognize it yet (hehe).
  // numbers are small so O(N) lookups ok
  const operationIds: String[] = [];
  const result: Operation[] = [];

  operations.push(...sdk.Operations);
  sdk.SubSDKs.forEach((sub) => {
    operations.push(...getUniqueOperations(sub));
  });
  operations.forEach((op) => {
    if (operationIds.indexOf(op.ID) == -1) {
      result.push(op);
      operationIds.push(op.ID);
    }
  });
  return result;
}

// @ts-ignore
function templateLanguageName(lang: string): string {
  // this function used to type code blocks (snippets) in markdown
  // files so that code gets rendered optimally by GitHub
  if (lang === "javav2") {
    return "java";
  } else {
    return lang;
  }
}
registerTemplateFunc("templateLanguageName", templateLanguageName);

function templateOptional(val: string, optional: boolean): string {
  return optional ? `${javaImportOptional()}.ofNullable(${val})` : val;
}
registerTemplateFunc("templateOptional", templateOptional);

function templateNullableOptional(
  val: string,
  optional: boolean,
  nullable: boolean,
): string {
  if (optional && nullable) {
    return `${javaImportJsonNullable()}.of(${val})`;
  } else if (optional || nullable) {
    return `${javaImportOptional()}.ofNullable(${val})`;
  } else {
    return val;
  }
}
registerTemplateFunc("templateNullableOptional", templateNullableOptional);

function templateOneOfNonDiscFactoryMethods(
  classSimpleName: string,
  type: TypeDef,
): string {
  const methodNames = associatedTypeFactoryMethodNames(type);
  const lines = type.AssociatedTypes.map((x) => {
    const t = sanitizeTypeMandatory(x);
    const methodName = methodNames.get(t);
    const typ = javaImport(t);
    const shape = hasShapeString(x) ? "STRING" : "DEFAULT";
    // Only check null for non-primitive types
    const nullCheck = isPrimitive(typ)
      ? ""
      : `${javaImportUtils()}.checkNotNull(value, "value");\n    `;
    // Java 8 cannot use the diamond operator on anonymous classes.
    const typeRefArg = isJava8() ? javaImportBoxed(t) : "";
    return `
public static ${classSimpleName} ${methodName}(${typ} value) {
    ${nullCheck}return new ${classSimpleName}(${javaImportLocal(
      "utils.TypedObject",
    )}.of(value, ${javaImportLocal(
      "utils.Utils.JsonShape",
    )}.${shape}, new ${javaImport(
      "com.fasterxml.jackson.core.type.TypeReference",
    )}<${typeRefArg}>(){}));
}`;
  });
  return indentLines(lines, 1);
}

// returns a map of java type to method name
function associatedTypeFactoryMethodNames(type: TypeDef): Map<string, string> {
  const m = new Map<string, string>();
  const counts = new Map<string, number>();
  type.AssociatedTypes.map((x) => sanitizeTypeMandatory(x)).forEach((x) => {
    x = javaTypeBase(x);
    const count = counts.get(x);
    if (count == undefined) {
      counts.set(x, 1);
    } else {
      counts.set(x, count + 1);
    }
  });
  type.AssociatedTypes.forEach((x, index) => {
    const t = sanitizeTypeMandatory(x);
    const key = javaTypeBase(t);
    var methodName: string;
    const count = counts.get(key);
    if (count != 1) {
      methodName = "of" + caser().ToPascal(subTypeName(x, index));
    } else {
      methodName = "of";
    }
    m.set(t, methodName);
  });
  return m;
}

function subTypeName(t: TypeDef, index: number): string {
  // To name an anonymous subtype of oneOf that has erasure conflicts
  // we prefer the overriden name, then the type name, then itemType name and fallback to the
  // one-based index
  const nameOverride: string = t.Extensions?.All
    ? t.Extensions.All["x-speakeasy-name-override"]
    : undefined;
  if (nameOverride != undefined) {
    return nameOverride;
  }
  if (t.Name) {
    return t.Name;
  }
  if (t.ItemType?.Name) {
    return t.ItemType.Name;
  }
  // use 1-based name
  return index + 1 + "";
}

function subTypeAccessorName(t: TypeDef): string {
  // Similar to subTypeName but optimized for generating accessor method names
  // Handles container types (map, list, etc.) with discriminant labels
  const nameOverride: string = t.Extensions?.All
    ? t.Extensions.All["x-speakeasy-name-override"]
    : undefined;
  if (nameOverride != undefined) {
    return nameOverride;
  }
  if (t.Name) {
    return sanitizeName(t.Name);
  }
  if (t.ItemType) {
    // Include container type discriminant for well-formed names
    // Use sanitized type name to get the language-specific type (e.g., integer -> long)
    const containerType = t.Type.toString();
    const sanitizedItemType = sanitizeTypeMandatory(t.ItemType);
    // Strip generic params before extracting simple class name to avoid
    // leaking angle brackets (e.g., Map<String, Object> -> Object> bug)
    const simpleItemTypeName = simpleClassName(javaTypeBase(sanitizedItemType));
    return `${containerType}_of_${simpleItemTypeName}`;
  }
  // For simple types without a name, use the sanitized type name
  // This ensures integer -> long, etc.
  const sanitizedType = sanitizeTypeMandatory(t);
  return simpleClassName(javaTypeBase(sanitizedType));
}

function javaTypeBase(typeName: string): string {
  const i = typeName.indexOf("<");
  if (i > -1) {
    return typeName.substring(0, i);
  } else {
    return typeName;
  }
}

registerTemplateFunc(
  "templateOneOfNonDiscFactoryMethods",
  templateOneOfNonDiscFactoryMethods,
);

function templateMemberTypeReferencesCommaDelimited(type: TypeDef): string {
  let types = type.AssociatedTypes;
  if (!isUnionStrategyPopulatedFields()) {
    types = sortTypeDefFieldCount(type.AssociatedTypes);
    // to enable the application of a heuristic to help with
    // matching weak unions (oneOf with multiple matches) we
    // sort by field count descending
    types = [...sortTypeDefFieldCount(type.AssociatedTypes)].reverse();
  }
  return types
    .map(
      (x) =>
        `,\n${templateIndent(4)}  ${javaImportLocal(
          "utils.Utils.TypeReferenceWithShape",
        )}.of(new ${javaImport(
          "com.fasterxml.jackson.core.type.TypeReference",
        )}<${javaImport(
          toNonPrimitive(sanitizeTypeMandatory(x)),
        )}>() {}, ${javaImportLocal("utils.Utils.JsonShape")}.${
          hasShapeString(x) ? "STRING" : "DEFAULT"
        })`,
    )
    .join("");
}

registerTemplateFunc(
  "templateMemberTypeReferencesCommaDelimited",
  templateMemberTypeReferencesCommaDelimited,
);

function templateOneOfValueJavadoc(type: TypeDef, scope: string): string {
  const types = type.AssociatedTypes.map(
    (x) => `\n * <li>{@code ${sanitizeTypeMandatory(x)}}</li>`,
  ).join("");
  return indentLines(
    [
      `/**
 * Returns an instance of one of these types:
 * <ul>${types}
 * </ul>
 * 
 * <p>Use {@code instanceof} to determine what type is returned. For example:
 * 
 * <pre>
 * if (obj.value() instanceof String) {
 *     String answer = (String) obj.value();
 *     System.out.println("answer=" + answer);
 * }
 * </pre>
 * 
 * @return value of oneOf type
 **/ 
`,
    ],
    1,
  );
}

registerTemplateFunc("templateOneOfValueJavadoc", templateOneOfValueJavadoc);

/**
 * Returns metadata for each union member to support optional accessor generation.
 * Each entry contains the type and accessor method name derived from the type name.
 */
function getUnionMemberMetadata(type: TypeDef): Array<{
  javaType: string;
  nonPrimitiveType: string;
  rawType: string;
  accessorMethodName: string;
  needsSuppressWarnings: boolean;
}> {
  return type.AssociatedTypes.map((x) => {
    const t = sanitizeTypeMandatory(x);
    const javaType = javaImport(t);

    // Get the type name for the accessor - use sanitized type to ensure
    // language-specific types (e.g., integer -> long) are reflected in the name
    const typeName = subTypeAccessorName(x);

    // Generate accessor name:
    // - For primitive types: use "as<Type>" pattern (e.g., asLong)
    // - For complex types: use just the type name (e.g., myClass)
    const baseType = javaTypeBase(javaType);
    const isPrimitiveType = isPrimitive(baseType);
    const accessorMethodName = isPrimitiveType
      ? caser().ToCamel("as_" + typeName)
      : caser().ToCamel(typeName);

    // Check if type is generic (contains < and >)
    const isGeneric = javaType.includes("<");

    // For generic types, extract the raw type for instanceof checks
    const rawType = toNonPrimitive(
      isGeneric ? javaTypeBase(javaType) : javaType,
    );

    return {
      javaType,
      nonPrimitiveType: toNonPrimitive(javaType),
      rawType,
      accessorMethodName,
      needsSuppressWarnings: isGeneric,
    };
  });
}

registerTemplateFunc("getUnionMemberMetadata", getUnionMemberMetadata);

function defaultValueJavaExpression(
  name: string,
  defaultValue: any,
  enhancedType: string,
): string {
  // note that enhancedType is base type wrapped with Optional or JsonNullable if applicable
  const ind = 6;
  const json = JSON.stringify(defaultValue)
    .replaceAll("\n", "")
    .replaceAll("\\", "\\\\")
    .replaceAll('"', '\\"');
  return `${javaImportUtils()}.readDefaultOrConstValue(\n${templateIndent(
    ind,
  )}"${name}",\n${templateIndent(ind)}"${json}",\n${templateIndent(
    ind,
  )}new ${javaImport(
    "com.fasterxml.jackson.core.type.TypeReference<" + enhancedType + ">",
  )}() {})`;
}
registerTemplateFunc("defaultValueJavaExpression", defaultValueJavaExpression);

function hasDefaultValue(field: FieldDef): boolean {
  return (
    !hasConstValue(field) &&
    field.Default &&
    typeof field.Default.Value !== "undefined"
  );
}

registerTemplateFunc("hasDefaultValue", hasDefaultValue);

function hasConstValue(field: FieldDef): boolean {
  return field.Const && typeof field.Const.Value !== "undefined";
}
registerTemplateFunc("hasConstValue", hasConstValue);

function wrapOptional(expression: string): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return `${javaImportOptional()}.ofNullable(${expression})`;
  }

  return expression;
}

registerTemplateFunc("wrapOptional", wrapOptional);

function wrapNullable(expression: string): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return `${javaImportJsonNullable()}.of(${expression})`;
  }

  return expression;
}

registerTemplateFunc("wrapNullable", wrapNullable);

function templateModelBuilderInitialAssignment(field: FieldDef, scope: string) {
  if (context.Global.Config.NullFriendlyParameters) {
    return "";
  }

  if (field.Optional && field.Nullable) {
    if (hasDefaultValue(field)) {
      return "";
    } else {
      return ` = ${javaImportJsonNullable()}.undefined()`;
    }
  } else if (field.Optional || field.Nullable) {
    if (hasDefaultValue(field)) {
      return "";
    } else {
      return ` = ${javaImportOptional()}.empty()`;
    }
  } else {
    return "";
  }
}

registerTemplateFunc(
  "templateModelBuilderInitialAssignment",
  templateModelBuilderInitialAssignment,
);

function templateModelDefaultValueSuppliers(
  fields: FieldDef[],
  scope: string,
): string {
  const lines = [];
  for (const field of fields) {
    if (hasDefaultValue(field) || hasConstValue(field)) {
      let typ = sanitizeType(field.Type, field.Optional, field.Nullable);
      if (context.Global.Config.NullFriendlyParameters) {
        typ = sanitizeType(field.Type);
      }
      const v = hasDefaultValue(field)
        ? field.Default.Value
        : field.Const.Value;
      lines.push(modelSingletonValueSupplier(field.Name, v, typ));
    }
  }
  if (lines.length > 0) {
    lines.unshift("");
  }
  return indentLines(lines, 2);
}

function objectToJavaJsonStringContent(o: Object): string {
  return JSON.stringify(o)
    .replaceAll("\n", "")
    .replaceAll("\\", "\\\\")
    .replaceAll('"', '\\"')
    .replaceAll("{{", `{{"{{"}}`);
}

function modelSingletonValueSupplier(
  fieldName: string,
  fieldDefaultOrConstValue: any,
  enhancedType: string,
  indent: number = 0,
) {
  const json = objectToJavaJsonStringContent(fieldDefaultOrConstValue);
  const typ = javaImport(toNonPrimitive(enhancedType));
  return `
${templateIndent(indent)}private static final ${javaImportLocal(
    "utils.LazySingletonValue",
  )}<${typ}> _SINGLETON_VALUE_${sanitizeClassName(fieldName)} =
        ${templateIndent(indent)}new ${javaImportLocal(
          "utils.LazySingletonValue",
        )}<>(
                ${templateIndent(indent)}"${fieldName}",
                ${templateIndent(indent)}"${json}",
                ${templateIndent(indent)}new ${javaImport(
                  "com.fasterxml.jackson.core.type.TypeReference",
                )}<${typ}>() {});`;
}

registerTemplateFunc(
  "modelSingletonValueSupplier",
  modelSingletonValueSupplier,
);

registerTemplateFunc(
  "templateModelDefaultValueSuppliers",
  templateModelDefaultValueSuppliers,
);

const primitiveTypeMapping = {
  boolean: "java.lang.Boolean",
  int: "java.lang.Integer",
  long: "java.lang.Long",
  float: "java.lang.Float",
  double: "java.lang.Double",
};

function toNonPrimitive(type: string): string {
  return primitiveTypeMapping[type] || type;
}

registerTemplateFunc("toNonPrimitive", toNonPrimitive);

function isPrimitive(type: string): boolean {
  return primitiveTypeMapping[type] !== undefined;
}

registerTemplateFunc("isPrimitive", isPrimitive);

function templateSetDefaults(fields: FieldDef[]): string {
  const lines = fields
    .filter((field) => hasDefaultValue(field))
    .flatMap((field) => {
      const name = sanitizeFieldName(field.Name);
      return templateSetDefault(name);
    });
  if (lines.length == 0) {
    return "";
  } else {
    return indentLines(lines, 3) + "\n";
  }
}

function templateSetDefault(fieldName: string): string[] {
  return `if (${fieldName} == null) {
    ${fieldName} = ${singletonValueGet(fieldName)};
}`.split("\n");
}

function singletonValueGet(fieldName: string): string {
  return `_SINGLETON_VALUE_${sanitizeClassName(fieldName)}.value()`;
}

registerTemplateFunc("templateSetDefaults", templateSetDefaults);

// @ts-ignore
function templateUnsetValue(): string {
  return undefined;
}

function templateEventStreamGetter(
  type: TypeDef,
  isAsync: boolean = false,
): string {
  if (isAsync || !isEventStreamResponse(type)) {
    return "";
  }
  const field = type.Fields.find((f) => isEventStreamField(f));
  const t = sanitizeEventStream(field.Type);
  const itemType = javaImport(sanitizeEventStreamItemType(field.Type));
  const sseDataRequired = isSSEDataRequired(field.Type.ItemType);
  const dataRequiredArg = sseDataRequired ? "" : ",\n            false";
  // TODO obtain value of x-speakeasy-sse-sentinel to pass
  // as last parameter below.
  return `
    // value overriden by x-speakeasy-sse-sentinel extension via reflection at runtime
    private ${javaImportOptional()}<${javaImportString()}> _eventSentinel = ${javaImportOptional()}.empty();

    public ${t} events() {
        return new ${t}(
            rawResponse.body(),
            new ${javaImport(
              "com.fasterxml.jackson.core.type.TypeReference",
            )}<${itemType}>() {},
            ${javaImportUtils()}.mapper(),
            _eventSentinel${dataRequiredArg});
    }
`;
}

registerTemplateFunc("templateEventStreamGetter", templateEventStreamGetter);

function templateJsonLStreamGetter(
  type: TypeDef,
  isAsync: boolean = false,
): string {
  if (isAsync || !isJsonLStreamResponse(type)) {
    return "";
  }
  const field = type.Fields.find((f) => isJsonLStreamField(f));
  const t = sanitizeJsonLStream(field.Type);
  const itemType = javaImport(sanitizeJsonLStreamItemType(field.Type));
  return `
    public ${t} events() {
        return new ${t}(
            rawResponse.body(), 
            new ${javaImport(
              "com.fasterxml.jackson.core.type.TypeReference",
            )}<${itemType}>() {}, 
            ${javaImportUtils()}.mapper());
  }
`;
}

registerTemplateFunc("templateJsonLStreamGetter", templateJsonLStreamGetter);

function isEventStreamResponse(responseType: TypeDef): boolean {
  return (
    isResponse(responseType) &&
    responseType.Fields.findIndex((x) => isEventStreamField(x)) != -1
  );
}

registerTemplateFunc("isEventStreamResponse", isEventStreamResponse);

function isJsonLStreamResponse(responseType: TypeDef): boolean {
  return (
    isResponse(responseType) &&
    responseType.Fields.findIndex((x) => isJsonLStreamField(x)) != -1
  );
}

registerTemplateFunc("isJsonLStreamResponse", isJsonLStreamResponse);

function isJsonLStreamField(field: FieldDef): boolean {
  return field.Type.Type.toString() == "jsonl";
}

registerTemplateFunc("isJsonLStreamField", isJsonLStreamField);

function isEventStreamField(field: FieldDef): boolean {
  return field.Type.Type.toString() == "event-stream";
}
registerTemplateFunc("isEventStreamField", isEventStreamField);

function eventStreamInnerType(responseType: TypeDef, scope: string): string {
  const field = responseType.Fields.find((f) => isEventStreamField(f));
  return sanitizeEventStreamItemType(field.Type);
}

registerTemplateFunc("eventStreamInnerType", eventStreamInnerType);

function jsonLStreamInnerType(responseType: TypeDef, scope: string): string {
  const field = responseType.Fields.find((f) => isJsonLStreamField(f));
  return sanitizeJsonLStreamItemType(field.Type);
}

registerTemplateFunc("jsonLStreamInnerType", jsonLStreamInnerType);

// @ts-ignore
function getBuildExtrasJob(): Job | undefined {
  const file = "build-extras.gradle";

  // If we don't already have a build-extras.gradle file then create it
  if (!readFile(file)) {
    return createTemplateFileJob("build-extras.gradle", file, {});
  }

  return;
}

// @ts-ignore
function getProjectIdentifier(): string {
  return caser().ToKebab(context.Global.Config.ProjectName).toLowerCase();
}

// @ts-ignore
function templateReadmeTitle(): string {
  return context.Global.Config.ProjectName;
}
unregisterTemplateFunc("templateReadmeTitle");
registerTemplateFunc("templateReadmeTitle", templateReadmeTitle);

function hasSecurity(): boolean {
  return sdkHasSecurity(context.Global.AST.MainSDK);
}
registerTemplateFunc("hasSecurity", hasSecurity);

function sdkHasSecurity(sdk: SDK): boolean {
  return (
    sdk.Security != undefined ||
    sdk.Operations.find((op) => op.Security != undefined) != undefined ||
    sdk.SubSDKs.find((s) => sdkHasSecurity(s)) != undefined
  );
}

function sdkHasServerMap(sdk: SDK): boolean {
  return sdk.Servers != undefined && sdk.Servers.ServerMap;
}
registerTemplateFunc("sdkHasServerMap", sdkHasServerMap);

function templateServerUrlInitialization(operation: Operation): string {
  const servers = operation.Servers;
  const operationId = operation.ID;

  const serverVariables = servers
    .GetVariables()
    .map((v) => `"${v.Name}", "${v.Default}"`);

  const serverConst = templateConstName(
    "SERVERS",
    sanitizeClassName(operationId),
  );
  const serverTemplate = servers.ServerMap
    ? `${serverConst}.get(${sanitizeClassName(
        operationId,
      )}Servers.${templateConstName(servers.Default, "")})`
    : `${serverConst}[0]`;

  return `${javaImportUtils()}.templateUrl(
    ${templateIndent(5)}${serverTemplate}, 
    ${templateIndent(5)}${templateJdkCall("Map.of")}(${serverVariables.join(
      ", ",
    )}))`;
}
registerTemplateFunc(
  "templateServerUrlInitialization",
  templateServerUrlInitialization,
);

const MAX_THROWS_IN_MAIN_METHOD = 6;

function usageMainThrowsList(
  operation: Operation,
  isAsync: boolean = false,
): string {
  if (isAsync) {
    return "";
  }
  let throwing = operation.Response.Responses.filter((x) => x.Error)
    .flatMap((x) => x.Content)
    .map((x) => javaImportTypeMandatory(x.Content.Type));
  if (throwing.length > MAX_THROWS_IN_MAIN_METHOD - 1) {
    throwing = [];
  }
  throwing.push(javaImportException());
  // remove duplicates (async Error copies?)
  var result = throwing.filter(
    (item, index) => throwing.indexOf(item) === index,
  );
  if (result.length == 0) {
    return "";
  } else {
    return ` throws ${result.join(", ")}`;
  }
}
registerTemplateFunc("usageMainThrowsList", usageMainThrowsList);

// @ts-ignore
function templateStream(
  fieldDef: FieldDef,
  filePath: string,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (filePath) {
    if (additionalContext?.isTest) {
      filePath = `.speakeasy/testfiles/${filePath}`;
    }
    if (additionalContext?.isResponse) {
      return `${javaImportUtils()}.readBytes("${filePath}")`;
    }
    if (useBlob()) {
      return `${javaImportLocal("utils.Blob")}.from(${javaImport(
        "java.nio.file.Paths",
      )}.get("${filePath}"))`;
    }
    return `${javaImportUtils()}.readBytesAndClose(new ${javaImport(
      "java.io.FileInputStream",
    )}("${filePath}"))`;
  } else {
    if (typeof example !== "string") {
      example = JSON.stringify(example);
    }
    const byteValue = templateByteValue(example, fieldDef, additionalContext);

    if (additionalContext?.isResponse) {
      return byteValue;
    } else {
      return byteValue;
    }
  }
}

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

  return `${javaImportUtils()}.readBytes("${filePath}")`;
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

  return `${javaImportUtils()}.readString("${filePath}")`;
}

function templateRequestParams(
  fields: FieldDef[],
  pageInput: PaginationInputs,
  offsetInput: PaginationInputs,
  cursorInput: PaginationInputs,
  indent: number = 6,
): string {
  const lines = [];
  for (const field of nonConstFields(fields)) {
    if (
      pageInput != null &&
      pageInput.Name == field.Name &&
      pageInput.In == "requestBody"
    ) {
      lines.push(`${javaImportOptional()}.ofNullable(_newPage)`);
      continue;
    }
    if (
      offsetInput != null &&
      offsetInput.Name == field.Name &&
      offsetInput.In == "requestBody"
    ) {
      lines.push(`${javaImportOptional()}.ofNullable(_newOffset)`);
      continue;
    }
    if (cursorInput != null && cursorInput.Name == field.Name) {
      if (!!field.Optional || !!field.Nullable) {
        lines.push(`${javaImportJsonNullable()}.of(_nextCursor)`);
      } else {
        lines.push(`_nextCursor`);
      }
      continue;
    }
    lines.push(`request.${sanitizeFieldName(field.Name)}()`);
  }
  // indent lines
  const indentedLines = [];
  if (lines.length > 0) {
    indentedLines.push(lines[0]);
    for (const line of lines.slice(1)) {
      indentedLines.push(templateIndent(indent) + line);
    }
  }
  return indentedLines.join(",\n");
}

registerTemplateFunc("templateRequestParams", templateRequestParams);

function templatePackageName(): string {
  if (context.Global.Config.PackageName) {
    return context.Global.Config.PackageName;
  } else {
    return `${context.Global.Config.GroupID.replaceAll(
      "-",
      "_",
    )}.${context.Global.Config.ArtifactID.replaceAll("-", "_")}`;
  }
}
registerTemplateFunc("templatePackageName", templatePackageName);

function templateProjectName(): string {
  let projectName = context.Global.Config.ProjectName || "lib";
  // For the new method of publishing (Sonatype central portal), the project name must be the artifactID
  if (context.Global.Config.Published && !context.Global.Config.OSSHURL) {
    projectName = context.Global.Config.ArtifactID;
  }
  return projectName;
}
registerTemplateFunc("templateProjectName", templateProjectName);

function sdkUsesPagination(sdk: SDK): boolean {
  let index = sdk.Operations.findIndex((op) => op.Extensions.Pagination);
  if (index != -1) {
    return true;
  } else {
    return sdk.SubSDKs.findIndex((s) => sdkUsesPagination(s)) != -1;
  }
}

// @ts-ignore (Ignore override of function implementation)
function templateAuxiliaryFiles(data: any, sourcePath: string = "auxiliary") {
  let files = getDirectoryFiles(sourcePath);
  for (const file of files) {
    if (
      file.includes("pagination") &&
      !sdkUsesPagination(context.Global.AST.MainSDK)
    ) {
      continue;
    }
    if (!asyncEnabled()) {
      if (
        [
          "AsyncRetries.java.stmpl",
          "AsyncRetryableException.java.stmpl",
          "AsyncResponse.java.stmpl",
          "AsyncHook.java.stmpl",
          "AsyncHooks.java.stmpl",
          "FlatteningSubscriber.java.stmpl",
          "AsyncPaginator.java.stmpl",
          "NonRetryableException.java.stmpl",
          "RetryableException.java.stmpl",
          "Unchecked{{getDefaultErrorClassName}}.java.stmpl",
          "reactive/EventStream.java.stmpl",
          "HookAdapters.java.stmpl",
        ].find((f) => file.endsWith(f))
      ) {
        continue;
      }
    }
    if (!asyncEnabled() && !useBlob()) {
      if (file.endsWith("Blob.md.stmpl")) {
        continue;
      }
    }
    if (useOkHttp()) {
      // java.net.http-specific utilities with no okhttp counterpart:
      // ResponseWithBody adapts the JDK async response, while the okhttp path
      // constructs its SDK-owned transport.HttpResponse directly. Helpers
      // exposes JDK HttpRequest.Builder copying
      // (transport.HttpRequest.toBuilder() replaces it).
      if (
        file.endsWith("ResponseWithBody.java.stmpl") ||
        file.endsWith("Helpers.java.stmpl")
      ) {
        continue;
      }
      // The okhttp arm's async surface is CompletableFuture-based: the
      // reactive types (utils.reactive.*) and the Blob-typed async error
      // class back the reactive async surface, which exists only on the JDK
      // arm (where ReactiveUtils is also used for BodyPublisher composition).
      if (
        file.endsWith("reactive/EventStream.java.stmpl") ||
        file.endsWith("reactive/ReactiveUtils.java.stmpl") ||
        file.endsWith("Async{{getDefaultErrorClassName}}.java.stmpl")
      ) {
        continue;
      }
    }
    if (!springBootStarterEnabled()) {
      if (
        file.includes("spring-boot-autoconfigure") ||
        file.includes("spring-boot-starter")
      ) {
        continue;
      }
    }
    if (!isForwardCompatibleUnions()) {
      if (
        file.endsWith("GenericTypeIdResolver.java.stmpl") ||
        file.endsWith("UnknownType.java.stmpl")
      ) {
        continue;
      }
    }

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

// @ts-ignore (Ignore override of function implementation)
function templateDirectory(dir: string, outDir?: string) {
  let files = getDirectoryFiles(dir);

  for (const file of files) {
    if (
      file.includes("pagination") &&
      !sdkUsesPagination(context.Global.AST.MainSDK)
    ) {
      continue;
    }
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

function filterOutParamByName(params: JavaParam[], name: string): JavaParam[] {
  return params.filter((p) => p.Name !== name);
}

registerTemplateFunc("filterOutParamByName", filterOutParamByName);

// Helper function to detect if an operation has SSE response for async operations
function isSSEOperation(operation: Operation): boolean {
  if (!operation || !operation.Response || !operation.Response.Type) {
    return false;
  }

  // Check if the response type has event stream fields
  return isEventStreamResponse(operation.Response.Type);
}

registerTemplateFunc("isSSEOperation", isSSEOperation);

// Consolidated helper function to detect if an operation has streaming response (SSE or JSONL) for async operations
function isStreamingOperation(operation: Operation): boolean {
  if (!operation || !operation.Response || !operation.Response.Type) {
    return false;
  }

  // Check if the response type has event stream or JSONL stream fields
  return (
    isEventStreamResponse(operation.Response.Type) ||
    isJsonLStreamResponse(operation.Response.Type)
  );
}

registerTemplateFunc("isStreamingOperation", isStreamingOperation);

// Helper function to get the SSE item type for an operation
function getSSEItemType(operation: Operation): string {
  if (!isSSEOperation(operation)) {
    return "";
  }

  const responseType = operation.Response.Type;
  const field = responseType.Fields.find((f) => isEventStreamField(f));
  if (!field) {
    return "";
  }

  return javaImport(sanitizeEventStreamItemType(field.Type));
}

registerTemplateFunc("getSSEItemType", getSSEItemType);

// Consolidated helper function to get the streaming item type for an operation (SSE or JSONL)
function getStreamingItemType(operation: Operation): string {
  if (!isStreamingOperation(operation)) {
    return "";
  }

  const responseType = operation.Response.Type;

  // Check for SSE field first
  const sseField = responseType.Fields.find((f) => isEventStreamField(f));
  if (sseField) {
    return javaImport(sanitizeEventStreamItemType(sseField.Type));
  }

  // Check for JSONL field
  const jsonlField = responseType.Fields.find((f) => isJsonLStreamField(f));
  if (jsonlField) {
    return javaImport(sanitizeJsonLStreamItemType(jsonlField.Type));
  }

  return "";
}

registerTemplateFunc("getStreamingItemType", getStreamingItemType);

// `new TypeReference<>() {}` (anonymous class with diamond) needs Java 9+;
// Java 8 call sites must name the streaming item type explicitly. Returns ""
// on Java 9+ so the default arm stays byte-identical.
function templateStreamingTypeRefParam(operation: Operation): string {
  return isJava8() ? getStreamingItemType(operation) : "";
}
registerTemplateFunc(
  "templateStreamingTypeRefParam",
  templateStreamingTypeRefParam,
);

// Helper function to get the SSE sentinel value for an operation
function getSSESentinel(operation: Operation): string {
  if (!isSSEOperation(operation)) {
    return "null";
  }

  // Check for sentinel in response content
  for (const response of operation.Response.Responses) {
    for (const content of response.Content) {
      if (
        content.SerializationMethod === "eventstream" &&
        content.SSESentinel
      ) {
        return `"${content.SSESentinel}"`;
      }
    }
  }

  // Check for sentinel in the event stream type
  const responseType = operation.Response.Type;
  if (responseType.EventStreamSentinel) {
    return `"${responseType.EventStreamSentinel}"`;
  }

  return "null";
}

registerTemplateFunc("getSSESentinel", getSSESentinel);

// Helper function to check if SSE data is required for an operation
function getSSEDataRequired(operation: Operation): boolean {
  if (!isSSEOperation(operation)) {
    return true;
  }
  const responseType = operation.Response.Type;
  const sseField = responseType.Fields?.find((f: FieldDef) =>
    isEventStreamField(f),
  );
  if (sseField) {
    return isSSEDataRequired(sseField.Type.ItemType);
  }
  return true;
}

registerTemplateFunc("getSSEDataRequired", getSSEDataRequired);

// Returns a code snippet (string) illustrating pattern matching on SSE event types
// using asEnum() + switch (open enums) or a direct switch (native enums).
// Returns "" if the SSE item type has no enum event field.
function getSSEPatternMatchingExample(operation: Operation): string {
  if (!isSSEOperation(operation)) return "";

  const responseType = operation.Response.Type;
  const sseField = responseType.Fields.find((f) => isEventStreamField(f));
  if (!sseField || !sseField.Type.ItemType) return "";

  const itemType = sseField.Type.ItemType;
  if (
    itemType.Type.toString() !== "class" ||
    !itemType.Fields ||
    itemType.Fields.length === 0
  ) {
    return "";
  }

  // Prefer an open-enum field (e.g. "event" field on PostMessageSseResponse);
  // fall back to a native enum field.
  const openEnumField = itemType.Fields.find(
    (f) => f.Type.Type.toString() === "enum" && isOpenEnum(f.Type),
  );
  const nativeEnumField = openEnumField
    ? undefined
    : itemType.Fields.find(
        (f) => f.Type.Type.toString() === "enum" && !isOpenEnum(f.Type),
      );
  const enumFieldDef = openEnumField ?? nativeEnumField;
  if (!enumFieldDef) return "";

  const enumNames = getEnumNames(enumFieldDef.Type);
  if (!enumNames || enumNames.length === 0) return "";

  const itemTypeName = javaImport(sanitizeEventStreamItemType(sseField.Type));
  const fieldMethodName = sanitizeFieldName(enumFieldDef.Name);
  const optional = isAlwaysOptionalGetter();

  // Optional getters unwrap the field to `eventValue`; otherwise read inline.
  const printValue = optional
    ? `System.out.println(eventValue.value());`
    : `System.out.println(event.${fieldMethodName}().value());`;

  const indent = (lines: string[]): string[] =>
    lines.map((line) => (line ? `    ${line}` : line));

  const cases = enumNames.flatMap((name) => [
    `case ${name}:`,
    `    ${printValue}`,
    `    break;`,
  ]);

  const switchSubject = openEnumField
    ? "eventType"
    : optional
    ? "eventValue"
    : `event.${fieldMethodName}()`;

  let block = [
    `switch (${switchSubject}) {`,
    ...indent(cases),
    `    default:`,
    `        break;`,
    `}`,
  ];

  // Open enums pattern-match via asEnum().ifPresent(eventType -> { ... }).
  if (openEnumField) {
    const receiver = optional ? "eventValue" : `event.${fieldMethodName}()`;
    block = [
      `${receiver}.asEnum().ifPresent(eventType -> {`,
      ...indent(block),
      `});`,
    ];
  }

  // Optional getters wrap the whole switch in ifPresent(eventValue -> { ... }).
  if (optional) {
    block = [
      `event.${fieldMethodName}().ifPresent(eventValue -> {`,
      ...indent(block),
      `});`,
    ];
  }

  return indentLines(
    [
      `// Option 3: Filter by event type using pattern matching`,
      `for (${itemTypeName} event : events) {`,
      ...indent(block),
      `}`,
    ],
    3,
  );
}

registerTemplateFunc(
  "getSSEPatternMatchingExample",
  getSSEPatternMatchingExample,
);

// Generate complete body field getter method - moved from model.java.stmpl
function templateBodyFieldGetter(
  field: FieldDef,
  isAsync: boolean = false,
): string {
  const fieldName = sanitizeFieldName(field.Name);

  if (isAsync) {
    // Async body field handling - return the body from rawResponse.body():
    // a plain InputStream on the okhttp (CompletableFuture) arm, a Blob on
    // the reactive JDK arm.
    const asyncBodyType = useOkHttp()
      ? javaImport("java.io.InputStream")
      : javaImportLocal("utils.Blob");
    return `
    public ${asyncBodyType} ${fieldName}() {
        return rawResponse.body();
    }`;
  }

  // Response body getters cache the raw body and have hand-rolled bodies
  // above; they always keep presence-aware types regardless of getterStyle.
  const bytesReturn = getGetterMetadata(
    parameterForField(field),
    "presence-aware",
  ).returnType;

  if (context.Global.Config.NullFriendlyParameters) {
    return `
    public ${bytesReturn} ${fieldName}() {
        this.${fieldName} = ${wrapOptional("body")}
            .orElseGet(() -> {
                try {
                    return ${javaImportUtils()}.extractByteArrayFromBody(rawResponse);
                } catch (${javaImport("java.io.IOException")} e) {
                    throw new ${javaImport("java.io.UncheckedIOException")}(e);
                }
            });
        return ${wrapOptional("this.body")};
    }`;
  }

  // Optional.or is Java 9+; route through Java8Compat.or on Java 8.
  const recoverBody = `{
                try {
                    return ${javaImportOptional()}.of(${javaImportUtils()}.extractByteArrayFromBody(rawResponse));
                } catch (${javaImport("java.io.IOException")} e) {
                    throw new ${javaImport("java.io.UncheckedIOException")}(e);
                }
            }`;
  const bodyOrExpr = isJava8()
    ? `${javaImportLocal("utils.Java8Compat")}.or(${wrapOptional(
        "body",
      )}, () -> ${recoverBody})`
    : `${wrapOptional("body")}
            .or(() -> ${recoverBody})`;
  return `
    public ${bytesReturn} ${fieldName}() {
        this.${fieldName} = ${bodyOrExpr};
        return ${coerceByte(
          castToRemoveWildcard(field, false),
        )} this.${fieldName};
    }`;
}

registerTemplateFunc("templateBodyFieldGetter", templateBodyFieldGetter);

/**
 * Builds type parameters, method parameters, and arguments for Java request building methods.
 *
 * @param op - The operation to analyze
 * @param reqParam - Optional request parameter (generates args array when provided)
 * @returns Object with typeParams, params, and optional args arrays, or undefined if no request
 */
function getBuildRequestParams(
  op: Operation,
  reqParam?: JavaParam,
):
  | {
      typeParams: string[];
      params: string[];
      args?: string[];
    }
  | undefined {
  const request = op.Request;
  if (!request) {
    return;
  }
  const typeParams = [];
  const params = [];
  const args = [];

  typeParams.push("T");
  params.push("T request");
  if (reqParam) {
    args.push("request");
  }
  if (
    request.Params &&
    (request.Params.HasPathParams() || request.Params.HasQueryParams())
  ) {
    params.push("Class<T> klass");
    if (reqParam) {
      args.push(`${javaImportRequest(reqParam, true)}.class`);
    }
  }

  if (op.SerializationMethod) {
    typeParams.push("U");
    params.push(
      `${javaImport(
        "com.fasterxml.jackson.core.type.TypeReference",
      )}<U> typeReference`,
    );
    if (reqParam) {
      args.push(
        `new ${javaImport(
          "com.fasterxml.jackson.core.type.TypeReference",
        )}<${javaImportRequest(reqParam)}>() {}`,
      );
    }
  }

  return { typeParams, params, args };
}

registerTemplateFunc("getBuildRequestParams", getBuildRequestParams);

function coerceByte(type: string): string {
  const blobType = `${templatePackageName()}.utils.Blob`;
  let out = type.replace(blobType, "byte[]");
  out = out.replace("Blob", "byte[]");

  return out;
}

function useBlob(): boolean {
  // Use Blob for request-streams when EnableStreamingUploads is enabled
  return context.Global.Config.EnableStreamingUploads;
}
registerTemplateFunc("useBlob", useBlob);

function asyncEnabled(): boolean {
  return context.Global.Config.AsyncMode === "enabled";
}
registerTemplateFunc("asyncEnabled", asyncEnabled);

function slf4jLoggingEnabled(): boolean {
  return context.Global.Config.EnableSlf4jLogging === true;
}
registerTemplateFunc("slf4jLoggingEnabled", slf4jLoggingEnabled);

function hasFileUpload(op: Operation): boolean {
  return fileUploadOpPredicate(context.Global.AST.MainSDK, op);
}

function springBootStarterEnabled(): boolean {
  // Spring Boot 3.x requires Java 17+
  return (
    context.Global.Config.GenerateSpringBootStarter === true &&
    context.Global.Config.LanguageVersion >= 17
  );
}
registerTemplateFunc("springBootStarterEnabled", springBootStarterEnabled);

function useOkHttp(): boolean {
  return context.Global.Config.HTTPClient === "okhttp";
}
registerTemplateFunc("useOkHttp", useOkHttp);

function templatePackageNameClean(): string {
  return templatePackageName().replaceAll(".", "");
}
registerTemplateFunc("templatePackageNameClean", templatePackageNameClean);

function templateJacksonModuleClassName(): string {
  return sanitizeClassName(templateProjectName()) + "JacksonModule";
}
registerTemplateFunc(
  "templateJacksonModuleClassName",
  templateJacksonModuleClassName,
);

function templateAutoConfigClassName(): string {
  // Create a proper class name from the project name or SDK name
  return sanitizeClassName(templateProjectName()) + "AutoConfig";
}
registerTemplateFunc(
  "templateAutoConfigClassName",
  templateAutoConfigClassName,
);

function templateServerVarEnum(sdk: SDK, variable: ServerVariable): string {
  return javaImportStatic(
    `${sanitizeClassName(sdk.Type.Name)}.Builder.${sanitizeClass(
      variable.Type,
      variable.Type.Scope,
      true,
    )}`,
    true,
  );
}
registerTemplateFunc("templateServerVarEnum", templateServerVarEnum);

function toStringSelect(fieldName: string, type: TypeDef): string {
  if (type.Type == "string") {
    return fieldName;
  }

  return `${fieldName}.toString()`;
}
registerTemplateFunc("toStringSelect", toStringSelect);

function errorAccessorUsesFlatMap(field: FieldDef): boolean {
  if (field.IsAdditionalProperties) {
    return false;
  }
  // Under "raw" every getter returns a plain (possibly null) value: map
  // yields an empty Optional for null. Under "always-optional" every getter
  // returns Optional<T>: always flatMap.
  if (isRawGetter()) {
    return false;
  }
  if (isAlwaysOptionalGetter()) {
    return true;
  }
  if (field.Optional && !field.Nullable) {
    return true;
  }
  if (
    field.Nullable &&
    !field.Optional &&
    !context.Global.Config.NullFriendlyParameters
  ) {
    return true;
  }
  return false;
}

/**
 * Called when templating a custom error in order to unpack nested data field.
 * Depending on whether the data field is optional, nullable - we return
 * map or flatMap to extract the inner value.
 */
function templateDataFieldTransform(field: FieldDef): string {
  return errorAccessorUsesFlatMap(field) ? "flatMap" : "map";
}
registerTemplateFunc("templateDataFieldTransform", templateDataFieldTransform);

// Helper function to generate discriminator type mappings for resolver classes
function generateDiscriminatorMappings(type: TypeDef, scope: string): string[] {
  if (!type.Discriminator || !type.Discriminator.Mapping) {
    return [];
  }

  const mappings = type.Discriminator.Mapping.map((mapping) => {
    // Find the associated type for this mapping
    const associatedType = type.AssociatedTypes.find(
      (x) =>
        x.OriginalName === mapping.Type.Name || x.Name === mapping.Type.Name,
    );

    if (!associatedType) {
      // This shouldn't happen in a well-formed discriminator, but handle gracefully
      return null;
    }

    const className = importedClass(associatedType);
    return `registerType("${mapping.Name}", ${className}.class);`;
  }).filter(Boolean);

  return mappings;
}

registerTemplateFunc(
  "generateDiscriminatorMappings",
  generateDiscriminatorMappings,
);

// Helper function to generate a safe "UNKNOWN" fallback discriminator value that doesn't conflict with existing mappings
function generateSafeUnknownDiscriminator(type: TypeDef): string {
  if (!type.Discriminator || !type.Discriminator.Mapping) {
    return DEFAULT_UNKNOWN_DISCRIMINATOR_VALUE;
  }
  return findUniqueUnknownValue(type, type.Discriminator.Mapping);
}

registerTemplateFunc(
  "generateSafeUnknownDiscriminator",
  generateSafeUnknownDiscriminator,
);

// Helper function to find the hardcoded discriminator value for a model type
function getHardcodedDiscriminatorValue(
  modelType: TypeDef,
  discriminatedUnionType: TypeDef,
): string | null {
  if (
    !discriminatedUnionType.Discriminator ||
    !discriminatedUnionType.Discriminator.Mapping
  ) {
    return null;
  }

  // Find the mapping that corresponds to this model type
  const mapping = discriminatedUnionType.Discriminator.Mapping.find(
    (m) =>
      m.Type.Name === modelType.OriginalName || m.Type.Name === modelType.Name,
  );

  return mapping ? mapping.Name : null;
}

registerTemplateFunc(
  "getHardcodedDiscriminatorValue",
  getHardcodedDiscriminatorValue,
);

// Helper function to generate Javadoc for discriminator getters
function generateDiscriminatorGetterJavadoc(
  modelType: TypeDef,
  discriminatedUnionType: TypeDef,
): string {
  const unionName = sanitizeClass(
    discriminatedUnionType,
    discriminatedUnionType.Scope,
    true,
  );

  // Get all possible discriminator values for context
  const allValues =
    discriminatedUnionType.Discriminator?.Mapping?.map(
      (m) => `"${m.Name}"`,
    ).join(", ") || "";

  const valuesDoc = allValues
    ? ` Valid discriminator values: ${allValues}`
    : "";

  return `
    /**
     * Returns the discriminator value for this ${unionName} implementation.
     * 
     * <p>
     * The value is determined from the actual field content, allowing this model
     * to be used both as part of a discriminated union and as a standalone model.
     *
     * @return the discriminator value from the field content
     * @see ${unionName}${valuesDoc}
     */`;
}

registerTemplateFunc(
  "generateDiscriminatorGetterJavadoc",
  generateDiscriminatorGetterJavadoc,
);

// Helper function to generate all discriminator getters for a type
function generateAllDiscriminatorGetters(type: TypeDef): string {
  const memberships = DiscriminatedOneOf.membership(type);
  const getters: string[] = [];

  for (const unionType of memberships) {
    const fieldName = discriminatorFieldName(unionType);
    const javadoc = generateDiscriminatorGetterJavadoc(type, unionType);

    // Find the hardcoded discriminator value for this model type
    const hardcodedValue = getHardcodedDiscriminatorValue(type, unionType);
    const returnValue = hardcodedValue
      ? `"${hardcodedValue}"`
      : `"${generateSafeUnknownDiscriminator(unionType)}"`;

    const getter = `
    ${javadoc}
    @${javaImportOverride()}
    public ${javaImportString()} ${fieldName}() {
        return ${returnValue};
    }`;

    getters.push(getter);
  }

  return getters.join("\n");
}

registerTemplateFunc(
  "generateAllDiscriminatorGetters",
  generateAllDiscriminatorGetters,
);

function templateGlobalParamAccess(param: FieldDef): string {
  const paramAnnot = param.Annotations?.Get("param");
  const paramType = paramAnnot?.ParamType;
  const paramName = getOperationParameterName(param);

  return `this.sdkConfiguration.globals.getParam("${paramType}", "${paramName}")
                .ifPresent(param -> operationGlobals.putParam("${paramType}", "${paramName}", param));`;
}
registerTemplateFunc("templateGlobalParamAccess", templateGlobalParamAccess);

function templateGlobalsReference(operation: Operation): string {
  if (!operation.Globals) {
    return "null";
  }

  if (context.Global.Config.OperationScopedParams) {
    return "this.operationGlobals";
  }

  return "this.sdkConfiguration.globals";
}
registerTemplateFunc("templateGlobalsReference", templateGlobalsReference);
