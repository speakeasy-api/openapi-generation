function templateUsage() {
  templateFile("usage.stmpl", "USAGE.md", {
    ExampleUsage: templateUsageOperations(
      selectExampleOperations(context.Global.AST.MainSDK, 0, [], true),
      "",
    )(context.Global.AST.MainSDK),
  });

  // Template TestUsage
  if (
    context.Global.Config.Usage?.OutputDir &&
    context.Global.Config.Language != "terraform"
  ) {
    templateDirectory("testusage", context.Global.Config.Usage.OutputDir);
  }
}

function templateUsageSection(
  opFilter: string,
  header?: string,
  feature?: string,
  isGlobal?: boolean,
  value?: any,
  asyncMode?: boolean,
): string {
  const scopes: UsageExampleScope[] = [
    {
      OpFilter: opFilter,
      Feature: feature || "",
      IsGlobal: isGlobal || false,
      Value: value,
    },
  ];

  const sdk = context.Global.AST.MainSDK;
  let ctxs: UsageContext[] = selectExampleOperations(sdk, 1, scopes, false)
    // Mutate async mode for all usage contexts
    .map((c) => (asyncMode ? { ...c, AsyncMode: true } : c));

  return templateUsageOperations(ctxs, header || "")(sdk);
}
registerTemplateFunc("templateUsageSection", templateUsageSection);

// this function exists as a workaround because calling $usageCtx.RenderFeature
// with args from a template is not working reliably (unknown reason for the moment)
function shouldRenderFeature(
  c: UsageContext,
  feature: string,
  defaultValue: boolean,
): boolean {
  return c.RenderFeature(feature, defaultValue);
}
registerTemplateFunc("shouldRenderFeature", shouldRenderFeature);

function templateUsageSectionWithContext(
  usageContext: UsageContext,
  header: string,
) {
  return templateUsageSectionWithContexts([usageContext], header);
}
registerTemplateFunc(
  "templateUsageSectionWithContext",
  templateUsageSectionWithContext,
);

function templateUsageSectionWithContexts(
  usageContexts: UsageContext[],
  header: string,
) {
  const sdk = context.Global.AST.MainSDK;
  return templateUsageOperations(usageContexts, header || "")(sdk);
}
registerTemplateFunc(
  "templateUsageSectionWithContexts",
  templateUsageSectionWithContexts,
);

function templateUsageExample(usageOperation: UsageContext, header: string) {
  return templateUsageOperations(
    [usageOperation],
    header,
  )(context.Global.AST.MainSDK);
}

registerTemplateFunc("templateUsageExample", templateUsageExample);

function templateUsageOperations(
  usageOperations: UsageContext[],
  header: string,
): (sdk: SDK) => string {
  return (sdk: SDK) => {
    let counterStr = "";
    let counter = 1;
    if (usageOperations.length > 1) {
      counterStr = " " + counter++;
    }
    return usageOperations
      .sort((usageA, usageB) => {
        const a = usageA.Config.Position;
        const b = usageB.Config.Position;
        if (a < b) {
          return -1;
        }
        if (a > b) {
          return 1;
        }
        return 0;
      })
      .map((ctx) => {
        let head = "";
        let desc = "";
        // custom title/description parsed from x-speakeasy-usage-example
        // are only used for the main `SDK Example Usage` section, i.e. when IsMainExample is set to true.
        if (ctx.IsMainExample) {
          if (ctx.Config.Title) {
            head = "### " + ctx.Config.Title;
          } else if (header != "") {
            head = header + counterStr; // fallback header
          }
          desc = ctx.Config.Description;
        } else if (header != "") {
          head = header;
        }
        counterStr = " " + counter++;

        return templateUsageBlock(ctx, head, desc);
      })

      .join("\n\n");
  };
}

function defaultModelDocLocation(
  outputLocation: string,
  type: TypeDef,
): string {
  let path = "docs";
  const sourceFileLocation = getModelsLocation(outputLocation);
  if (sourceFileLocation) {
    path += "/" + sourceFileLocation;
  }

  return `${path}/${sanitizeFileName(sanitizeClassName(type.Name))}.md`;
}

function getBadgeURL(): string {
  if (context.Global.Config.RepoURL) {
    const repoURL = context.Global.Config.RepoURL; // "https://github.com/speakeasy-sdks/example.git
    const repoURLWithoutSuffix = repoURL.replace(/\.git$/, ""); // "https://github.com/speakeasy-sdks/example
    const urlPath = repoURLWithoutSuffix.replace(/.*github\.com/, ""); // "/speakeasy-sdks/example"
    const urlPathParts = urlPath.split("/").filter(Boolean); // ["speakeasy-sdks", "example"]
    const constructedURL = `https://img.shields.io/github/actions/workflow/status/${urlPathParts.join(
      "/",
    )}/speakeasy_sdk_generation.yml?style=for-the-badge`;
    return JSON.stringify(constructedURL);
  }
  return '""';
}
registerTemplateFunc("getBadgeURL", getBadgeURL);

type getTargetFileFunc = (scope: string, type: TypeDef) => string;

function templateModelDocs(collectedTypes: BucketedTypes) {
  const jobs = getModelDocsJobs(collectedTypes);

  for (const job of jobs) {
    templateFileJob(job as TemplateFileJob);
  }
}

function getModelDocsJobs(
  collectedTypes: BucketedTypes,
  singleThreadedModelDocs?: boolean,
): Job[] {
  const seen = new Set<string>();
  const jobs = [];

  if (singleThreadedModelDocs) {
    return [
      {
        ID: "templateModelDocs",
        FileName: "",
        Context: {
          collectedTypes,
        },
      },
    ];
  }

  for (const [outputLocation, models] of sequencedMapEntries(collectedTypes)) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const type of types) {
        if (
          type.Scope.toString() == "utils" ||
          type.Scope.toString() == "sdk" ||
          seen.has(type.GetRegistrationID())
        ) {
          continue;
        }

        seen.add(type.GetRegistrationID());

        jobs.push(
          createTemplateFileJob(
            `readme/model.stmpl`,
            defaultModelDocLocation(outputLocation, type),
            type,
          ),
        );
      }
    }
  }

  return jobs;
}

function templateUsageSnippet(local: UsageContext): string {
  const output = templateString("usage/snippet.stmpl", local);
  return formatUsageSnippetOutput(output);
}

registerTemplateFunc("templateUsageSnippet", templateUsageSnippet);

function templateUsageBlock(
  context: UsageContext,
  header: string,
  description: string,
): string {
  if (!context || !context.Operation) {
    return "";
  }

  faker.seed(context.Operation.GetExampleSeed());

  if (header == "" && description == "") {
    return templateString("readme/snippet.stmpl", context);
  } else {
    return templateString("readme/usage.stmpl", {
      Header: header,
      Description: description,
      UsageContext: context,
    });
  }
}

function createUsageContextTS(
  sdk: SDK,
  operation: Operation,
  usageExample?: UsageExampleConfig,
): UsageContext {
  if (operation) {
    faker.seed(operation.GetExampleSeed());
  }
  return createUsageContext(sdk, operation, usageExample);
}

registerTemplateFunc("createUsageContext", createUsageContextTS);

/**
 * Creates UsageContext objects for an operation - one for each named example.
 * Named examples are paired by name between request and response.
 * If no named examples exist, returns a single context with default behavior.
 */
function createUsageContextsForOperation(
  sdk: SDK,
  operation: Operation,
  usageExample?: UsageExampleConfig,
): UsageContext[] {
  if (operation) {
    faker.seed(operation.GetExampleSeed());
  }

  const requestExamples = getRequestExampleNames(operation);
  const responseExamples = getResponseExampleNames(operation);
  const exampleNames = getUniqueExampleNames(requestExamples, responseExamples);

  if (exampleNames.length === 0) {
    return [createUsageContext(sdk, operation, usageExample)];
  }

  return exampleNames.map((name) => {
    const ctx = createUsageContext(sdk, operation, usageExample);
    ctx.ExampleName = name;
    return ctx;
  });
}

// Check if an example name is auto-generated (not explicitly defined in the spec)
function isAutoGeneratedExampleName(
  name: string,
  operationId: string,
): boolean {
  // Auto-generated names follow patterns like:
  // - "speakeasy-default-*"
  // - "something[N]" (e.g., "postTest2[0]", "user-lifecycle[1]", "userSetup[0]")
  if (name.startsWith("speakeasy-default-")) {
    return true;
  }
  // Check for any name ending with [N] - these are auto-generated array indices
  if (/\[\d+\]$/.test(name)) {
    return true;
  }
  return false;
}

function getRequestExampleNames(operation: Operation): string[] {
  if (!operation.Request?.Examples || operation.Request.Examples.length === 0) {
    return [];
  }
  const operationId = operation.ID || "";
  // Filter out examples that have references (require test context to resolve)
  // and auto-generated example names
  return operation.Request.Examples.filter((ex) => !ex.Reference)
    .map((ex) => ex.Name())
    .filter(
      (name) => name !== "" && !isAutoGeneratedExampleName(name, operationId),
    );
}

function getResponseExampleNames(operation: Operation): string[] {
  if (!operation.Response?.Responses) {
    return [];
  }

  const operationId = operation.ID || "";
  const names = new Set<string>();
  for (const resp of operation.Response.Responses) {
    // Only consider successful responses, not errors
    if (resp.Error) continue;
    for (const content of resp.Content) {
      if (content.Examples) {
        for (const ex of content.Examples) {
          // Skip examples that have references (require test context to resolve)
          if (ex.Reference) continue;
          const name = ex.Name();
          // Skip auto-generated example names
          if (name !== "" && !isAutoGeneratedExampleName(name, operationId)) {
            names.add(name);
          }
        }
      }
    }
  }
  return Array.from(names).sort();
}

function getUniqueExampleNames(
  requestExamples: string[],
  responseExamples: string[],
): string[] {
  const allNames = new Set([...requestExamples, ...responseExamples]);
  return Array.from(allNames).sort();
}

registerTemplateFunc(
  "createUsageContextsForOperation",
  createUsageContextsForOperation,
);

/**
 * Extracts the appropriate response field name from an operation's responses.
 * Prioritizes responses with usage examples, falls back to first non-error response.
 *
 * @param operation - The operation containing response definitions
 * @param serializationMethod - Optional filter for specific serialization method
 * @returns The sanitized field name of the selected response
 * @throws Error if a specific serialization method is required but not found
 */
function getResponseFieldName(
  operation: Operation,
  serializationMethod = "",
): string {
  // Primary field name (prefers responses with usage examples)
  let resField = "";
  // Backup field name (first valid response found)
  let backupResField = "";

  for (const response of operation.Response.Responses) {
    for (const content of response.Content) {
      // Skip if we need a specific serialization method and this isn't it
      if (
        serializationMethod &&
        content.SerializationMethod !== serializationMethod
      ) {
        continue;
      }

      // Skip error responses
      if (response.Error) continue;

      if (!backupResField) {
        backupResField = content.Content.Name;
      }

      // Prefer responses with usage examples for the primary field
      if (content.UsageExample && !resField) {
        resField = content.Content.Name;
      }
    }
  }

  if (serializationMethod && !resField && !backupResField) {
    throw new Error(
      `No response field found for operation "${operation.ID}" and serialization method "${serializationMethod}"`,
    );
  }

  if (!resField) {
    resField = backupResField;
  }

  return sanitizeFieldName(resField);
}

registerTemplateFunc("getResponseFieldName", getResponseFieldName);

function collectAllUsageExamples(): UsageContext[] {
  return selectExampleOperations(context.Global.AST.MainSDK, 0, [], true);
}

registerTemplateFunc("collectAllUsageExamples", collectAllUsageExamples);
