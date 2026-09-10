function getReadmeTOCSection(
  readme: string,
  sections: ResolvedReadmeSection[],
): ResolvedReadmeSection | undefined {
  const tocSection = sections.find((s) => s.ID == "toc");

  if (
    tocSection == undefined ||
    !tocSection.Needed ||
    isReadmeSectionMarkedAsNoAction(
      BLOCK_BOUNDARY_PREFIX,
      tocSection.LegacyHeader,
      tocSection.ID,
      BLOCK_BOUNDARY_SUFFIX,
      readme,
    )
  ) {
    return undefined;
  }

  return tocSection;
}

function getReadmeTOCMaxDepth(
  readme: string,
  tocSection: ResolvedReadmeSection,
): number {
  const minDepth = DEFAULT_TOC_MIN_DEPTH;
  let maxDepth = DEFAULT_TOC_MAX_DEPTH;

  const toc = parseReadmeBlock(
    BLOCK_BOUNDARY_PREFIX,
    tocSection.ID,
    BLOCK_BOUNDARY_SUFFIX,
    readme,
  );

  if (toc) {
    const maxDepthRegex = new RegExp(
      `<!-- \\$${tocSection.ID}-max-depth=(\\d+) -->`,
    );
    const maxDepthMatch = toc.match(maxDepthRegex);
    if (maxDepthMatch && maxDepthMatch[1]) {
      maxDepth = parseInt(maxDepthMatch[1]);
      if (maxDepth < minDepth) {
        maxDepth = minDepth;
      }
    }
  }

  return maxDepth;
}

function updateReadmeTOC(
  contents: string,
  sections: ResolvedReadmeSection[],
  maxDepth: number,
): string {
  const tocSection = getReadmeTOCSection(contents, sections);

  if (tocSection == undefined) {
    return contents;
  }

  const sectionsExcludedFromTOC = ["toc", "summary"];
  const excludedHeaders = sections
    .filter((s) => !s.Needed || sectionsExcludedFromTOC.includes(s.ID))
    .map((s) => templateReadmeHeader(s.Header));

  const toc = generateReadmeTOC(
    contents,
    DEFAULT_TOC_MIN_DEPTH,
    maxDepth,
    excludedHeaders.join(","),
  );

  const maxDepthSetting = `<!-- $${tocSection.ID}-max-depth=${maxDepth} -->`;

  return replaceReadmeBlock(
    BLOCK_BOUNDARY_PREFIX,
    tocSection.LegacyHeader,
    tocSection.ID,
    BLOCK_BOUNDARY_SUFFIX,
    contents,
    tocSection.Header,
    `${templateReadmeHeader(tocSection.Header)}\n${maxDepthSetting}\n${toc}`,
  );
}
