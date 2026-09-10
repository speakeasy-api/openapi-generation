//  <!-- Start {Header} [ID] -->     < start boundary  |B
//  ## {Header}                                        |L
//                                                     |O
//  {Content}                                          |C
//  <!-- End {Header} [ID] -->       < end boundary    |K
type ReadmeSection = {
  ID: keyof AvailableReadmeSections; // [a-z\-]+ should be unique and stable
  Header: string; // [\w\s-]+  bound to change over time
  LegacyHeader: string; // kept for retro-compatibility
  Needed: () => boolean; // whether it must be rendered or possibly removed
  Weight: number; // lower weights float to the top
  Content: (sdk: SDK) => string;
};

type AvailableReadmeSection = {
  Header: string;
  Weight: number;
  LegacyHeader?: string;
  Feature?: Feature;
};

type AvailableReadmeSections = Partial<{
  /* keys must contain lowercase letters (a-z) and hyphens (-) only */
  [key in
    | "summary"
    | "toc"
    | "installation"
    | "requirements"
    | "idesupport"
    | "usage"
    | "async-support"
    | "security"
    | "operations"
    | "standalone-funcs"
    | "react-query"
    | "dev-containers"
    | "global-parameters"
    | "eventstream"
    | "pagination"
    | "file-upload"
    | "retries"
    | "errors"
    | "server"
    | "http-client"
    | "resource-management"
    | "jsonl"
    | "debug"
    | "types"
    | "dynamic-mode"
    | "jackson"]: AvailableReadmeSection;
}>;

// @ts-ignore
const AVAILABLE_README_SECTIONS: AvailableReadmeSections = {
  summary: {
    Header: "Summary",
    Weight: 0,
    Feature: "core",
  },
  toc: {
    Header: "Table of Contents",
    Weight: 1,
    Feature: "core",
  },
  installation: {
    Header: "SDK Installation",
    Weight: 2,
    Feature: "core",
  },
  requirements: {
    Header: "Requirements",
    Weight: 3,
    Feature: "core",
  },
  idesupport: {
    Header: "IDE Support",
    Weight: 4,
  },
  usage: {
    Header: "SDK Example Usage",
    Weight: 5,
    Feature: "core",
  },
  security: {
    Header: "Authentication",
    Weight: 10,
    Feature: "globalSecurity",
  },
  operations: {
    Header: "Available Resources and Operations",
    Weight: 20,
    Feature: "core",
    LegacyHeader: "SDK Available Operations",
  },
  "standalone-funcs": {
    Header: "Standalone functions",
    Weight: 21,
    Feature: "core",
  },
  "react-query": {
    Header: "React hooks with TanStack Query",
    Weight: 22,
    Feature: "reactQueryHooks",
  },
  "dev-containers": {
    Header: "Dev Containers",
    Weight: 30,
    Feature: "devContainers",
  },
  "global-parameters": {
    Header: "Global Parameters",
    Weight: 40,
    Feature: "globals",
  },
  eventstream: {
    Header: "Server-sent event streaming",
    Weight: 50,
    Feature: "serverEvents",
  },
  jsonl: {
    Header: "Json Streaming",
    Weight: 55,
    Feature: "jsonlResponses",
  },

  pagination: {
    Header: "Pagination",
    Weight: 60,
    Feature: "pagination",
  },
  "file-upload": {
    Header: "File uploads",
    Weight: 70,
    Feature: "uploadStreams",
  },
  retries: {
    Header: "Retries",
    Weight: 80,
    Feature: "retries",
  },
  errors: {
    Header: "Error Handling",
    Weight: 90,
    Feature: "errors",
  },
  server: {
    Header: "Server Selection",
    Weight: 100,
    Feature: "serverIDs",
  },
  "http-client": {
    Header: "Custom HTTP Client",
    Weight: 110,
    Feature: "core",
  },
  "resource-management": {
    Header: "Resource Management",
    Weight: 120,
  },
  "async-support": {
    Header: "Asynchronous Support",
    Weight: 6,
    Feature: "core",
  },
  debug: {
    Header: "Debugging",
    Weight: 130,
  },
  types: {
    Header: "Special Types",
    Weight: 140,
  },
  jackson: {
    Header: "Jackson Configuration",
    Weight: 145,
  },
};

const globallyRegisteredReadmeSections = new Array<ReadmeSection>();

function registerReadmeSection(
  id: keyof AvailableReadmeSections,
  neededCallback: () => boolean,
  contentCallback: (sdk: SDK) => string,
  availableSections: AvailableReadmeSections = AVAILABLE_README_SECTIONS,
) {
  const section = availableSections[id];

  if (section.Feature) {
    registerReadmeSectionTemplated(id, section.Feature);
  }

  globallyRegisteredReadmeSections.push({
    ID: id,
    Header: section.Header,
    Needed: neededCallback,
    Weight: section.Weight,
    Content: contentCallback,
    LegacyHeader: section.LegacyHeader || section.Header,
  });
}

type ResolvedReadmeSection = {
  ID: keyof AvailableReadmeSections;
  Header: string;
  LegacyHeader: string;
  Content: string;
  Needed: boolean;
  Weight: number;
};

function resolveReadmeSections(
  sdk: SDK,
  sections: ReadmeSection[],
): ResolvedReadmeSection[] {
  return sections.map((s) => {
    const needed = s.Needed();
    return {
      ID: s.ID,
      Header: s.Header,
      LegacyHeader: s.LegacyHeader,
      Content: needed ? s.Content(sdk).trim() : "",
      Needed: needed,
      Weight: s.Weight,
    };
  });
}

function updateReadmeSectionsWeight(
  contents: string,
  sections: ResolvedReadmeSection[],
): ResolvedReadmeSection[] {
  const parsedIDs = parseReadmeBlockIDs(
    BLOCK_BOUNDARY_PREFIX,
    BLOCK_BOUNDARY_SUFFIX,
    contents,
  )
    .split(",")
    .filter((id) => id in ({} as AvailableReadmeSections));

  const referenceWeights = sections
    .filter((s) => parsedIDs.includes(s.ID))
    .map((s) => s.Weight)
    .sort((a, b) => a - b);

  const updatedSections: Record<string, ResolvedReadmeSection> = {};
  parsedIDs.forEach((id) => {
    updatedSections[id] = {
      ...sections.find((s) => s.ID === id),
      Weight: referenceWeights.shift(),
    };
  });

  return sections
    .map((s) => updatedSections[s.ID] || s)
    .sort((a, b) => a.Weight - b.Weight);
}
