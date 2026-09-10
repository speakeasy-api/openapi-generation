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
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  jobs.push(...getSDKFunctionsJobs(context.Global.AST.MainSDK));
  jobs.push(...getMCPServerJobs(context.Global.AST.MainSDK));
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    const testDirectory = "src/__tests__";
    templateDirectory("tests/common", testDirectory);

    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, testDirectory);
    }

    templateDirectory("tests/lib", `${testDirectory}/lib`);
    templateFile("tests/vitest.config.mts", "vitest.config.mts", {});
  }

  const testJobs = getTestJobs(context.Global.AST, "ts");
  jobs.push(...testJobs);
  jobs.push(...getTestHelperJobs(context.Global.AST, "ts"));

  // // Template usage snippets inside test project
  // if (context.Global.Config.TestProject?.OutputDir) {
  //   jobs.push({
  //     ID: "testProject",
  //     FileName: "testProject",
  //     Context: null,
  //   });
  // }

  jobs.push(...getModelJobs(context.Global.AST.BucketedTypes, "src"));
  jobs.push(...getReadmeJobs());
  jobs.push(...getGitFilesJobs("ts"));
  jobs.push(...getHooksJobs());

  // const contributingFileJob = getTemplateContributingFileJob();
  // if (contributingFileJob) {
  //   jobs.push(contributingFileJob);
  // }

  // Has a dependency on context.Global so best to just template it while they are available
  templateFile("package.json.stmpl", "package.json", context.Global);

  return jobs;
}

function getSDKJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(createTemplateFileJob("sdk.ts.stmpl", "src/sdk/sdk.ts", sdk));

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
      "subsdk.ts.stmpl",
      `src/sdk/${sanitizeFileName(subSDK.Type.Name).toLowerCase()}.ts`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
