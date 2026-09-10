const BLOCK_BOUNDARY_PREFIX = "<!--";
const BLOCK_BOUNDARY_SUFFIX = "-->";
const DEFAULT_TOC_MIN_DEPTH = 1;
const DEFAULT_TOC_MAX_DEPTH = 2;

function blockBoundary(keyword: string, title: string): string {
  return `${BLOCK_BOUNDARY_PREFIX} ${keyword} ${title} ${BLOCK_BOUNDARY_SUFFIX}`;
}

const NEW_SECTIONS_PLACEHOLDER = blockBoundary(
  "Placeholder",
  "for Future Speakeasy SDK Sections",
);

// TODO: deprecated remove once all templates are using the job system
function templateReadme() {
  const jobs = getReadmeJobs();

  for (const job of jobs) {
    switch (job.ID) {
      case "readme":
        templateReadmeJob(job);
        break;
      case "usage":
        break;
      default:
        templateFileJob(job as TemplateFileJob);
        break;
    }
  }
}

function getReadmeJobs(singleThreadedModelDocs?: boolean): Job[] {
  const readmeJob: Job = {
    ID: "readme",
    FileName: "README.md",
    Context: globallyRegisteredReadmeSections,
  };

  const jobs: Job[] = [readmeJob];
  const language = context.Global.Config.Language;

  if (language.startsWith("mcp-")) {
    return jobs;
  }

  jobs.push({
    ID: "usage",
    FileName: "usage",
    Context: null,
  });

  if (language === "terraform") {
    return jobs;
  }

  // Sub-SDK README docs (docs/sdks/<subSDK>/README.md) are generated as
  // separate jobs so they can run in parallel across worker engines.
  jobs.push(...getSDKDocsJobs(context.Global.AST.MainSDK));

  jobs.push(
    ...getModelDocsJobs(
      context.Global.AST.BucketedTypes,
      singleThreadedModelDocs,
    ),
  );

  return jobs;
}

function templateReadmeJob(job: Job) {
  const header = templateString("readme/header.stmpl", {});

  createOrEditReadme(context.Global.AST.MainSDK, header, job.Context);
}

function templateReadmeHeader(header: string, level: number = 2): string {
  return `${"#".repeat(level)} ${header}`;
}

function createOrEditReadme(
  sdk: SDK,
  header: string,
  sections: ReadmeSection[],
  template: string = "readme.stmpl",
  outFile: string = "README.md",
) {
  let tocMaxDepth = DEFAULT_TOC_MAX_DEPTH;
  let readme = readFile(outFile);
  const resolvedSections = resolveReadmeSections(sdk, sections);

  if (!readme) {
    readme = templateString(template, {
      Header: header,
      Sections: resolvedSections
        .filter((s) => s.Needed)
        .sort((a, b) => a.Weight - b.Weight),
    });
  } else {
    const tocSection = getReadmeTOCSection(readme, resolvedSections);
    if (tocSection !== undefined) {
      tocMaxDepth = getReadmeTOCMaxDepth(readme, tocSection);
    }

    readme = editReadmeContent(
      updateReadmeSectionsWeight(readme, resolvedSections),
      readme,
    );
  }

  readme = updateReadmeTOC(readme, resolvedSections, tocMaxDepth);

  if (!readme) {
    throw new Error("Updated readme content is empty!");
  }

  writeFile(outFile, readme, 0, false);
}

function editReadmeContent(
  sections: ResolvedReadmeSection[],
  readme: string,
): string {
  if (!readme.includes(NEW_SECTIONS_PLACEHOLDER)) {
    readme += `\n${NEW_SECTIONS_PLACEHOLDER}\n`;
  }

  for (const section of sections) {
    if (
      isReadmeSectionMarkedAsNoAction(
        BLOCK_BOUNDARY_PREFIX,
        section.LegacyHeader,
        section.ID,
        BLOCK_BOUNDARY_SUFFIX,
        readme,
      )
    ) {
      continue;
    }

    // replace or remove section
    let replacement = "";
    if (section.Needed) {
      if (section.Header != "") {
        replacement += templateReadmeHeader(section.Header);
      }

      if (!context.Global.Config.Readme?.HeadersOnly) {
        replacement += `\n\n${section.Content}`;
      }
    }

    const updatedReadme = replaceReadmeBlock(
      BLOCK_BOUNDARY_PREFIX,
      section.LegacyHeader,
      section.ID,
      BLOCK_BOUNDARY_SUFFIX,
      readme,
      section.Header,
      replacement,
    );

    if (updatedReadme !== undefined) {
      readme = updatedReadme;

      if (section.ID == "usage") {
        // Legacy: the Header for the SDK Example Usage section
        // used to be templated *above* the start block boundary.
        // TODO: deprecate if/when no longer relevant.
        readme = cleanSdkExampleUsagePrefix(readme);
      }
    } else if (section.Needed) {
      // insert new section
      const newSection = replaceReadmeBlock(
        BLOCK_BOUNDARY_PREFIX,
        "",
        section.ID,
        BLOCK_BOUNDARY_SUFFIX,
        "",
        section.Header,
        replacement,
      );

      // find insertion point based on section weight
      const followingSections = sections
        .filter((s) => {
          return s.Needed && s.Weight > section.Weight;
        })
        .map((s) => {
          const startBoundary = `${BLOCK_BOUNDARY_PREFIX} Start ${s.Header} [${s.ID}] ${BLOCK_BOUNDARY_SUFFIX}`;
          return readme.indexOf(startBoundary);
        })
        .filter((i) => i > 0);

      const insertionPoint =
        followingSections.length > 0
          ? followingSections[0]
          : readme.indexOf(NEW_SECTIONS_PLACEHOLDER);

      readme =
        readme.slice(0, insertionPoint) +
        `${newSection}\n\n` +
        readme.slice(insertionPoint);
    }
  }

  return readme;
}

function cleanSdkExampleUsagePrefix(readme: string): string {
  const prefix = "## SDK Example Usage";
  const index = readme.indexOf(`${prefix}
<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage`);

  if (index > 0) {
    return readme.slice(0, index - 1) + readme.slice(index + prefix.length);
  }
  return readme;
}

function generateStandaloneReadme() {
  if (!context.Global.Config.Readme?.FileName) {
    throw new Error("Standalone Readme Generation is misconfigured");
  }

  createOrEditReadme(
    context.Global.AST.MainSDK,
    "Standalone README",
    globallyRegisteredReadmeSections,
    "readme.stmpl",
    context.Global.Config.Readme.FileName,
  );
}
