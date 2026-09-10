require("includes/includes.ts");
require("features.ts");

context.GlobalComputed.CustomTypes = new Map<string, boolean>();

// @ts-ignore
function runJobs(jobs: Job[]) {
  for (const job of jobs) {
    runJob(job);
  }
}

// @ts-ignore
function runJob(job: Job) {
  if (!handleCommonJobs(job)) {
    switch (job.ID) {
      case "testProject":
        templateTestProject(
          (u: UsageContext) =>
            `${sanitizeFileName(
              u.Operation.OwningSDK.Type.Name,
            )}/${sanitizeFileName(u.Operation.ID)}.go`,
        );
        break;
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

// @ts-ignore
function getJobs(): Job[] {
  initializeGoOperationOptionMetadata();

  const jobs: Job[] = [];

  let templateOption = false;

  jobs.push(
    ...getSDKJobs(context.Global.AST.MainSDK, (operation) => {
      const optionMetadata = getGoOperationOptionMetadata(operation);
      if (
        (operation.Servers ||
          operation.Extensions.Retries ||
          operation.Extensions.Polling ||
          optionMetadata.mode === "method-options" ||
          optionMetadata.optionalArguments.length > 0) &&
        !templateOption
      ) {
        templateOption = true;

        const opsLocation = context.Global.Config.Imports.GetOperationsPath();

        let templatingPath = "docs";
        if (opsLocation) {
          templatingPath += `/${opsLocation}`;
        }

        jobs.push(
          createTemplateFileJob(
            `readme/options.stmpl`,
            `${templatingPath}/option.md`,
            null,
          ),
        );
      }
    }),
  );

  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (isTemplateFeatureEnabled("tests")) {
    if (context.Global.AST.MainSDK.OutputTests) {
      const testDirectory = getTestDirectory();
      templateDirectory(`tests/common`, testDirectory);

      const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
      if (directoryExists(groupTestPath)) {
        templateDirectory(groupTestPath, testDirectory);
      }
    }

    jobs.push(...getTestJobs(context.Global.AST, "go"));
    jobs.push(...getTestHelperJobs(context.Global.AST, "go"));
  }

  // Template usage snippets inside test project
  if (context.Global.Config.TestProject?.OutputDir) {
    jobs.push({
      ID: "testProject",
      FileName: "testProject",
      Context: null,
    });
  }

  if (context.Global.AST.MainSDK.Globals) {
    jobs.push(
      createTemplateFileJob(
        `modelfile.go.stmpl`,
        `${getInternalPackageName()}/globals/globals.go`,
        {
          DocGroup: "internal/globals",
          Name: "globals",
          Types: [context.Global.AST.MainSDK.Globals],
          OutputLocation: `${getInternalPackageName()}/globals`,
        },
      ),
    );
  }

  if (isTemplateFeatureEnabled("documentation")) {
    const hasCustomTypes = () => listCustomTypes().length > 0;

    registerReadmeSection("types", hasCustomTypes, (_sdk: SDK) =>
      templateString("readme/types.stmpl", {}),
    );

    jobs.push(...getReadmeJobs());
  }

  const opsLocation = context.Global.Config.Imports.GetOperationsPath();

  let path = "options.go";
  if (opsLocation) {
    path = opsLocation + "/" + path;
  }

  jobs.push(
    createTemplateFileJob(`options.go.stmpl`, path, context.Global.AST.MainSDK),
  );

  if (hasGoOptionalMethodArgumentOptions() || hasGoMethodSpecificOptions()) {
    let methodOptionsPath = "method_options.go";
    if (opsLocation) {
      methodOptionsPath = `${opsLocation}/${methodOptionsPath}`;
    }
    jobs.push(
      createTemplateFileJob(
        `method_options.go.stmpl`,
        methodOptionsPath,
        context.Global.AST.MainSDK,
      ),
    );
  }

  jobs.push(...getGitFilesJobs("go"));

  // Generated Dev Containers Templating
  if (context.Global.Config.hasOwnProperty("DevContainerSchemaPath")) {
    jobs.push(
      ...getDevContainerJobs({
        FileSuffix: "go",
        Language: "go",
        FileExecution: "go run",
        SchemaPath: context.Global.Config.DevContainerSchemaPath,
      }),
    );
  }

  jobs.push(...getHooksJobs());

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  jobs.push(...getModelJobs(context.Global.AST.BucketedTypes, "", "go"));

  return jobs;
}

function getSDKJobs(
  sdk: SDK,
  onOperation: (operation: Operation) => void = () => {},
): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob(
      `sdk.go.stmpl`,
      sanitizeFileName(sdk.Type.Name).toLowerCase() + ".go", // TODO: check if I got this name correct
      sdk,
    ),
  );

  jobs.push(
    createTemplateFileJob(
      `sdkconfiguration.go.stmpl`,
      `${getInternalPackageName()}/config/sdkconfiguration.go`,
      sdk,
    ),
  );

  for (const operation of sdk.Operations) {
    onOperation(operation);
  }

  for (const subSDK of sdk.SubSDKs) {
    jobs.push(...getSubSDKJobs(subSDK, onOperation));
  }

  return jobs;
}

function getSubSDKJobs(
  subSDK: SDK,
  onOperation: (operation: Operation) => void = () => {},
): Job[] {
  const jobs: Job[] = [];

  if (subSDK.Type.Name === "") return jobs;

  jobs.push(
    createTemplateFileJob(
      `subsdk.go.stmpl`,
      sanitizeFileName(subSDK.Type.Name).toLowerCase() + ".go",
      subSDK,
    ),
  );

  for (const operation of subSDK.Operations) {
    onOperation(operation);
  }

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk, onOperation));
  }

  return jobs;
}
