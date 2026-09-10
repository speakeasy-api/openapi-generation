require("includes/includes.ts");

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
            `src/${sanitizeFileName(
              u.Operation.OwningSDK.Type.Name,
            )}/${sanitizeFileName(u.Operation.ID)}.php`,
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

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    templateDirectory(`Tests/Common`, `Tests`);

    const groupTestPath = `Tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, `Tests`);
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

  jobs.push(...getModelJobs(context.Global.AST.BucketedTypes, "src"));

  if (context.Global.Config.LaravelServiceProvider.enabled) {
    jobs.push(
      createTemplateFileJob(
        "PackageServiceProvider.php.stmpl",
        "src/PackageServiceProvider.php",
        context.Global.AST.MainSDK,
      ),
      createTemplateFileJob(
        "services.php.stmpl",
        "config/services.php",
        context.Global.AST.MainSDK,
      ),
    );
  }

  jobs.push(...getReadmeJobs());
  jobs.push(...getGitFilesJobs("php"));

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  jobs.push(...getHooksJobs());

  return jobs;
}

function getSDKJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob(
      "sdk.php.stmpl",
      `src/${sanitizeClassName(sdk.Type.Name)}.php`,
      sdk,
    ),
    createTemplateFileJob(
      "builder.php.stmpl",
      `src/${sanitizeClassName(sdk.Type.Name)}Builder.php`,
      sdk,
    ),
    createTemplateFileJob(
      "sdkconfiguration.php.stmpl",
      "src/SDKConfiguration.php",
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
      "subsdk.php.stmpl",
      `src/${sanitizeClassName(subSDK.Type.Name)}.php`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
