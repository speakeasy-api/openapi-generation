require("includes/index.ts");

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
    templateDirectory(
      `tests/Project`,
      `../sdk-unity-${context.Global.AST.MainSDK.TestGroup}-tests`,
    );

    const groupTestPath = `tests/Tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(
        groupTestPath,
        `../sdk-unity-${context.Global.AST.MainSDK.TestGroup}-tests/Assets/Tests`,
      );
    }
  }

  jobs.push(
    ...getModelJobs(context.Global.AST.BucketedTypes, getSDKTopLevelFolder()),
  );

  // registered sections are listed in includes/index.ts
  jobs.push(...getReadmeJobs());

  jobs.push(...getGitFilesJobs("cs"));

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
      )}.cs`,
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
