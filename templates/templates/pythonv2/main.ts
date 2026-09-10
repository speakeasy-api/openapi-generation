require("includes/includes.ts");

function updateGlobals() {
  const globals = context.Global.AST.MainSDK.Globals;
  if (globals) {
    const globalsOutputLocation = "models/internal";
    globals.OutputLocation = globalsOutputLocation;
    context.Global.AST.MainSDK.Globals = { ...globals };
  }
}

// @ts-ignore
function runJobs(jobs: Job[]) {
  updateGlobals();

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
            `src/${sanitizeFileName(
              u.Operation.OwningSDK.Type.Name,
            )}/${sanitizeFileName(u.Operation.ID)}.py`,
        );
        break;
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}
// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  updateGlobals();

  if (context.Global.AST.MainSDK.Globals) {
    context.Global.AST.BucketedTypes = addTypeToBucket(
      context.Global.AST.BucketedTypes,
      "globals",
      context.Global.AST.MainSDK.Globals,
    );
  }

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    const testDirectory = getTestDirectory();
    templateDirectory(`tests/common`, testDirectory);
    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, testDirectory);
    }
  }

  // Template usage snippets inside test project
  if (context.Global.Config.TestProject?.OutputDir) {
    jobs.push({
      ID: "testProject",
      FileName: "testProject",
      Context: null,
    });
  }

  jobs.push(...getTestJobs(context.Global.AST, "py"));
  jobs.push(...getTestHelperJobs(context.Global.AST, "py"));

  jobs.push(...getModelJobs(context.Global.AST.BucketedTypes, getSourcePath()));
  jobs.push(...getPythonPublicExportJobs(getSourcePath()));

  if (shouldTemplateOpenDiscriminatedUnion()) {
    jobs.push(
      createTemplateFileJob(
        "open-union-utils.py.stmpl",
        `${getSourcePath()}/utils/unions.py`,
        {},
      ),
    );
  }

  // registered sections are listed in includes/includes.ts
  registerReadmeSection(
    "idesupport",
    () => true,
    (sdk: SDK) => templateString("readme/idesupport.stmpl", {}),
  );

  registerReadmeSection(
    "resource-management",
    () => true,
    (sdk: SDK) => {
      const results = selectExampleOperations(
        context.Global.AST.MainSDK,
        0,
        [],
        true,
      );
      let usageContext: UsageContext | null = null;
      if (results.length) {
        usageContext = results[0];
      }

      return templateString("readme/resource-management.stmpl", {
        UsageContext: usageContext,
      });
    },
  );

  jobs.push(...getReadmeJobs());
  jobs.push(...getGitFilesJobs("py"));

  // Generated Dev Containers Templating
  if (context.Global.Config.hasOwnProperty("DevContainerSchemaPath")) {
    jobs.push(
      ...getDevContainerJobs({
        FileSuffix: "py",
        Language: "python",
        FileExecution: "python",
        SchemaPath: context.Global.Config.DevContainerSchemaPath,
      }),
    );
  }

  jobs.push(...getHooksJobs());

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  return jobs;
}

function getSDKJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob("sdk.py.stmpl", `${getSourcePath()}/sdk.py`, sdk),
  );

  jobs.push(
    createTemplateFileJob(
      "sdkconfiguration.py.stmpl",
      `${getSourcePath()}/sdkconfiguration.py`,
      sdk,
    ),
  );

  for (const subSDK of sdk.SubSDKs) {
    jobs.push(...getSubSDKJobs(subSDK));
  }

  return jobs;
}

function getSubSDKJobs(subSDK: SDK): Job[] {
  const jobs: Job[] = [];

  if (subSDK.Type.Name === "") return jobs;

  jobs.push(
    createTemplateFileJob(
      "subsdk.py.stmpl",
      `${getSourcePath()}/${sanitizeSDKFileName(subSDK.Type.Name, true)}.py`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
