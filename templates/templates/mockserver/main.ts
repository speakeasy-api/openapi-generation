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
  switch (job.ID) {
    case "templateFile":
      templateFileJob(job as TemplateFileJob);
      break;
    case "auxiliary":
      // Template/copy the auxiliary files
      templateAuxiliaryFiles(null);
      break;
  }
}

// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  const handlers = collectMockServerHandlers(context.Global.AST);

  // Template/copy static files
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  for (const handler of handlers) {
    jobs.push(
      createTemplateFileJob(
        "handlerfile.go.stmpl",
        `internal/handler/${handler.FileName}`,
        handler,
      ),
    );
  }

  jobs.push(
    createTemplateFileJob(
      "generated_handlers.go.stmpl",
      `internal/handler/generated_handlers.go`,
      handlers,
    ),
  );

  jobs.push(
    ...getModelJobs(context.Global.AST.BucketedTypes, "internal/sdk", "go"),
  );

  return jobs;
}
