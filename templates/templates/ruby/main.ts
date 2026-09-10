require("includes/includes.ts");

// Previously, tapioca generated RBI files were bundled in the generator. This
// added a large amount of size to the generator binary. Now, tapioca commands
// dynamically handle these files. Adding this pattern prevents the generator
// from deleting prior RBI files after the first generation with this change.
addUntrackedPattern(`sorbet/rbi/(annotations|gems)/.+\.rbi$`);

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
            )}/${sanitizeFileName(u.Operation.ID)}.rb`,
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

  jobs.push(...templateCrystalline());

  let srcPath = getSourcePath(context.Global.Config.Module);

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));

  const serverModels = new Set();
  for (const serverVariable of context.Global.AST.MainSDK.Servers.GetVariables()) {
    if (serverVariable.Type.Type == "enum") {
      const fileName = sanitizeFileName(serverVariable.Type.Name);
      jobs.push(
        createTemplateFileJob(
          "modelfile.rb.stmpl",
          `${srcPath}/${getScopePath("serverVariables")}/${fileName}.rb`,
          {
            Type: serverVariable.Type,
            Scope: "serverVariables",
          },
        ),
      );
      serverModels.add(serverVariable.Type.Name);
    }
  }
  if (serverModels.size > 0) {
    const svPath = getScopePath("serverVariables");
    jobs.push(
      createTemplateFileJob("module.rb.stmpl", `${srcPath}/${svPath}.rb`, {
        PathString:
          [srcPath.slice(4), parentDir(svPath)].filter(Boolean).join("/") + "/",
        ModuleName: "ServerVariables",
        ModuleDeclarations: templateModuleDeclarations(svPath, 1),
        ModuleClosures: templateModuleClosures(svPath, 1),
        ModuleDepth: getNamespaceModuleParts(svPath).length + 1,
        Models: serverModels,
      }),
    );
  }

  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  jobs.push(
    ...getModelJobs(
      context.Global.Config.PackageName,
      context.Global.AST.BucketedTypes,
      `${srcPath}`,
      "rb",
    ),
  );

  jobs.push(...getReadmeJobs());
  jobs.push(...getGitFilesJobs("rb"));

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    templateDirectory(`tests/common`, `test`);

    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, `test`);
    }
  }

  jobs.push(...getTestJobs(context.Global.AST, "rb"));
  jobs.push(...getTestHelperJobs(context.Global.AST, "rb"));

  // Template usage snippets inside test project
  if (context.Global.Config.TestProject?.OutputDir) {
    jobs.push({
      ID: "testProject",
      FileName: "testProject",
      Context: null,
    });
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

  let srcPath = getSourcePath(context.Global.Config.Module);

  const sdkFileName = sanitizeFileName(context.Global.Config.SDKName);

  jobs.push(
    createTemplateFileJob("sdk.rb.stmpl", `${srcPath}/${sdkFileName}.rb`, sdk),
    createTemplateFileJob(
      "sdkconfiguration.rb.stmpl",
      `${srcPath}/sdkconfiguration.rb`,
      sdk,
    ),
    createTemplateFileJob(
      "sdkconfiguration.rbi.stmpl",
      `${srcPath}/sdkconfiguration.rbi`,
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
      "subsdk.rb.stmpl",
      `${getSourcePath(context.Global.Config.Module)}/${sanitizeFileName(
        subSDK.Type.Name,
      ).toLowerCase()}.rb`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
