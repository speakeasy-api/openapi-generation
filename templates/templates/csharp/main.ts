require("includes/includes.ts");
require("features.ts");

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
              sanitizeClassName(u.Operation.OwningSDK.Type.Name),
            )}/${sanitizeFileName(sanitizeClassName(u.Operation.ID))}.cs`,
        );
        break;
      case "nugetReadme":
        templateNugetReadme();
        break;
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));

  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    templateDirectory(`tests/Common`, `Tests`);
    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, `Tests`);
    }
  }

  // Template Arazzo test files
  if (isTemplateFeatureEnabled("tests")) {
    if (
      sdkHasTests(context.Global.AST) &&
      !context.Global.AST.MainSDK.OutputTests
    ) {
      templateDirectory(`tests/Common`, `Tests`);
    }
    jobs.push(...getTestJobs(context.Global.AST, "cs"));
    jobs.push(...getTestHelperJobs(context.Global.AST, "cs"));
  }

  // Template usage snippets inside test project
  if (context.Global.Config.TestProject?.OutputDir) {
    jobs.push({
      ID: "testProject",
      FileName: "testProject",
      Context: null,
    });
  }

  jobs.push(
    ...getModelJobs(context.Global.AST.BucketedTypes, getSDKTopLevelFolder()),
  );

  // registered sections are listed in includes/includes.ts
  jobs.push(...getReadmeJobs());

  if (context.Global.Config.Published) {
    jobs.push({
      ID: "nugetReadme",
      FileName: "nugetReadme",
      Context: null,
    });
  }

  jobs.push(...getGitFilesJobs("cs"));

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
    createTemplateFileJob(
      `sdk.cs.stmpl`,
      `${getSDKTopLevelFolder()}/${sanitizeFileName(
        sanitizeSDKName(context.Global.Config.SDKName),
      )}.cs`, // TODO: check if I got this name correct
      sdk,
    ),
    createTemplateFileJob(
      `sdkconfig.cs.stmpl`,
      `${getSDKTopLevelFolder()}/SDKConfig.cs`,
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
      `subsdk.cs.stmpl`,
      `${getSDKTopLevelFolder()}/${sanitizeFileName(
        sanitizeSDKName(subSDK.Type.Name),
      )}.cs`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
