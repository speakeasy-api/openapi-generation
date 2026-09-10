require("includes/includes.ts");
require("features.ts");

// @ts-ignore
function runJobs(jobs: Job[]) {
  for (const job of jobs) {
    runJob(job);
  }
}

// @ts-ignore
function postJobs() {
  templateSnippets();
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

  if (!context.Global.Config.GroupID.includes(".")) {
    throw new ValidationError(
      `the groupID in gen.yaml must be in reverse domain name notation, e.g. com.example`,
    );
  }

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    // copy test group resources
    if (
      directoryExists(`tests/resources/${context.Global.AST.MainSDK.TestGroup}`)
    ) {
      templateDirectory(
        `tests/resources/${context.Global.AST.MainSDK.TestGroup}`,
        `${getTestResourceDirectory()}`,
      );
    }

    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, getTestSourceDirectory());
    }

    templateDirectory(`tests/common`, getTestSourceDirectory());
  }

  // Template tests directory
  if (isTemplateFeatureEnabled("tests")) {
    jobs.push(...getTestJobs(context.Global.AST, "java"));
    jobs.push(...getTestHelperJobs(context.Global.AST, "java"));
  }

  jobs.push(
    ...getModelJobs(
      context.Global.AST.BucketedTypes,
      getSourceDirectory(),
      "java",
    ),
  );

  jobs.push(...getReadmeJobs());
  jobs.push(
    ...getGitFilesJobs(
      "java",
      `#
# https://help.github.com/articles/dealing-with-line-endings/
#
# Linux start script should use lf
/gradlew        text eol=lf

`,
    ),
  );
  jobs.push(...getHooksJobs());

  const buildExtrasJob = getBuildExtrasJob();
  if (buildExtrasJob) {
    jobs.push(buildExtrasJob);
  }

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  return jobs;
}

function templateSnippets() {
  // this method extracts all java snippets from markdown files in the
  // generated project and creates a java class for each in the file SnippetNNNN.java in
  // src/test/java with package name snippets. The snippet is modified a little so that
  // the snippet class name is SnippetNNNN rather than Application.
  if (context.Global.AST.MainSDK.TestGroup) {
    console.debug("templating snippets");
    const n = templateSnippetsInFolder(".", 0);
    console.debug("snippetCount=" + n);
  }
}

function templateSnippetsInFolder(folder: string, snippetNo: number): number {
  const start = "\n```java";
  const finish = "\n```";
  if (folder == "./build" || folder == "./.git") {
    // temporary directories for gradle
    return snippetNo;
  }
  for (const file of listFiles(folder)) {
    if (file.endsWith(".md")) {
      const filename = folder + "/" + file;
      const text = readFile(filename);
      if (text == undefined) {
        throw new Error("text not readable for " + filename);
      }
      let i = 0;
      while (i < text.length) {
        const j = text.indexOf(start, i);
        if (j == -1) {
          break;
        }
        const k = text.indexOf(finish, j + start.length);
        if (k == -1) {
          break;
        }
        const extract = text.substring(j + start.length, k);
        // ensure we ignore fragments that are not full classes
        if (extract.includes("public class")) {
          snippetNo++;
          const className = "Snippet" + pad(snippetNo, 4);
          const snippet =
            extract
              .trim()
              .replace("package hello.world;", "package snippets;")
              .replace(
                "public class Application",
                `public class ${className}`,
              ) + "\n";
          const outFile = `src/test/java/snippets/${className}.java`;
          writeFile(outFile, snippet, 0, false);
        }
        i = k + finish.length;
      }
    }
  }
  for (const file of listFolders(folder)) {
    snippetNo = templateSnippetsInFolder(folder + "/" + file, snippetNo);
  }
  return snippetNo;
}

function pad(value: number, size: number) {
  let num = value.toString();
  while (num.length < size) num = "0" + num;
  return num;
}

function getSDKJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob(
      "sdk.java.stmpl",
      `${getSourceDirectory()}/${sanitizeClassName(
        context.Global.AST.MainSDK.Type.Name,
      )}.java`,
      context.Global.AST.MainSDK,
    ),
  );

  if (asyncEnabled()) {
    jobs.push(
      createTemplateFileJob(
        "async-sdk.java.stmpl",
        `${getSourceDirectory()}/Async${sanitizeClassName(
          context.Global.AST.MainSDK.Type.Name,
        )}.java`,
        context.Global.AST.MainSDK,
      ),
    );
  }

  const operationsPath = `/${getScopePath("operations")}/`;

  for (const op of getUniqueOperations(sdk)) {
    jobs.push(
      createTemplateFileJob(
        context.Global.Config.NullFriendlyParameters
          ? "null-friendly/methodrequestbuilder.java.stmpl"
          : "methodrequestbuilder.java.stmpl",
        `${getSourceDirectory()}${operationsPath}${templateRequestBuilderSimpleClassName(
          op,
        )}.java`,
        sdkOperation(sdk, op),
      ),
    );
    if (asyncEnabled()) {
      jobs.push(
        createTemplateFileJob(
          context.Global.Config.NullFriendlyParameters
            ? "null-friendly/methodrequestbuilder.java.stmpl"
            : "methodrequestbuilder.java.stmpl",
          `${getSourceDirectory()}${operationsPath}/async/${templateRequestBuilderSimpleClassName(
            op,
          )}.java`,
          sdkOperation(sdk, op, true),
        ),
      );
    }
    jobs.push(
      createTemplateFileJob(
        "operation.java.stmpl",
        `${getSourceDirectory()}/operations/${templateOperationSimpleClassName(
          op,
        )}.java`,
        sdkOperation(sdk, op),
      ),
    );
  }

  jobs.push(
    createTemplateFileJob(
      "sdkconfiguration.java.stmpl",
      `${getSourceDirectory()}/SDKConfiguration.java`,
      sdk,
    ),
  );

  jobs.push(
    createTemplateFileJob(
      "securitysource.java.stmpl",
      `${getSourceDirectory()}/SecuritySource.java`,
      sdk,
    ),
  );

  if (!hasSecurity()) {
    jobs.push(
      createTemplateFileJob(
        "securityempty.java.stmpl",
        `${getSourceDirectory()}/${getScopePath("shared")}/Security.java`,
        sdk,
      ),
    );
  }

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
      "subsdk.java.stmpl",
      `${getSourceDirectory()}/${sanitizeClassName(subSDK.Type.Name)}.java`,
      subSDK,
    ),
  );

  if (asyncEnabled()) {
    jobs.push(
      createTemplateFileJob(
        "async-subsdk.java.stmpl",
        `${getSourceDirectory()}/Async${sanitizeClassName(
          subSDK.Type.Name,
        )}.java`,
        subSDK,
      ),
    );
  }

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
