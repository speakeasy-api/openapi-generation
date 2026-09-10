type AvailableSDK = {
  SDK: SDK;
  Name: string;
  Deprecated: boolean;
};

type AvailableGroups = { [group: string]: AvailableSDK[] };

type SDKGroup = {
  Path: string;
  Name: string;
  SubGroups: SDKGroup[];
  SDKs: AvailableSDK[];
};

function templateOperationSummary(
  operation: Operation,
  replacementLink?: string,
): string {
  if (
    !operation.Comments ||
    (!operation.Comments.Summary && !operation.Comments.Deprecated)
  ) {
    return "";
  }

  let summary = " -";

  if (operation.Comments.Summary) {
    summary += ` ${operation.Comments.Summary}`;
  }

  if (operation.Comments.Deprecated) {
    summary += ` :warning: **Deprecated**`;

    if (operation.Comments.DeprecationReplacement) {
      if (!replacementLink && operation.Comments.DeprecationReplacement) {
        const replacementOp = context.Global.AST.MainSDK.FindOperation(
          operation.Comments.DeprecationReplacement,
        );
        if (replacementOp) {
          const slug = sanitizeMarkdownSlug(sanitizeMethodName(replacementOp));
          replacementLink = `[${sanitizeMethodName(
            replacementOp,
          )}](${getSDKReadmeFileName(replacementOp.OwningSDK)}#${slug})`;
        }
      }

      if (replacementLink) {
        summary += ` Use ${replacementLink} instead.`;
      }
    }
  }

  return summary;
}

registerTemplateFunc("templateOperationSummary", templateOperationSummary);

function collectAvailableSDKGroups(sdk: SDK): AvailableGroups {
  const availableGroups: AvailableGroups = {};

  for (const subSDK of sdk.SubSDKs) {
    for (const sub of flattenSubSDKs(subSDK)) {
      if (!(sub.Group in availableGroups)) {
        availableGroups[sub.Group] = [];
      }

      availableGroups[sub.Group].push({
        SDK: sub,
        Name: sanitizeClassName(sub.Type.Name),
        Deprecated:
          sub.Operations.every((op) => op.Comments?.Deprecated) &&
          sub.SubSDKs.every((s) =>
            s.Operations.every((op) => op.Comments?.Deprecated),
          ),
      });
    }
  }

  return availableGroups;
}

/**
 * Builds a hierarchical tree structure for available SDK groups
 * (e.g., "Group", "Group.Parent.Child", etc.) with parent-child
 * relationships established based on path prefixes.
 */
function buildAvailableSDKGroupsTree(sdk: SDK): SDKGroup {
  let sdkName = sanitizeClassName(sdk.Type.Name);
  if (!sdkName.endsWith("SDK")) {
    sdkName += " SDK";
  }

  const root: SDKGroup = {
    Path: "",
    Name: sdkName,
    SubGroups: [],
    SDKs: [
      {
        SDK: sdk,
        Name: sdkName,
        Deprecated: false,
      },
    ],
  };

  const groupMap: { [group: string]: SDKGroup } = { "": root };
  const groups = collectAvailableSDKGroups(sdk);
  const paths = Object.keys(groups).sort((a, b) => a.localeCompare(b));

  function findNearestAncestor(path: string): SDKGroup {
    const parentPath = path.split(".").slice(0, -1).join(".");
    // When `multiLevelTagging` feature is enabled, a `parentPath` entry will always exist since:
    //  - multi-level tagging creates all intermediate subSDKs
    //  - alphabetical sort ensures parents are processed before children
    //  - root group ("") is pre-seeded in groupMap
    if (groupMap[parentPath]) {
      return groupMap[parentPath];
    }
    // Otherwise, we recursively search for the nearest existing ancestor to nest under.
    const parts = path.split(".");
    if (parts.length === 0) return root;
    return findNearestAncestor(parts.slice(0, -1).join("."));
  }

  for (const path of paths) {
    const sanitizedPath = path
      .split(".")
      .map((p) => caser().ToPascal(sanitizeName(p)))
      .join(".");
    const sdkGroup: SDKGroup = {
      Path: path,
      Name: sanitizedPath,
      SubGroups: [],
      SDKs: groups[path] || [],
    };
    groupMap[path] = sdkGroup;

    const parent = findNearestAncestor(path);
    parent.SubGroups.push(sdkGroup);
  }

  return root;
}

function templateAvailableOperations(sdk: SDK): string {
  const root = buildAvailableSDKGroupsTree(sdk);

  return `<details open>
<summary>Available methods</summary>

${renderAvailableSDKGroupsTree(root, 3)}</details>
`;
}

/**
 * Collects templateFile jobs for sub-SDK README docs (e.g. docs/sdks/<subSDK>/README.md).
 * These are generated as separate jobs so they can run in parallel rather than
 * being generated as a side effect of the operations table Content callback.
 */
function getSDKDocsJobs(sdk: SDK): Job[] {
  const root = buildAvailableSDKGroupsTree(sdk);
  const jobs: Job[] = [];
  collectSDKDocsJobs(root, jobs);
  return jobs;
}

function collectSDKDocsJobs(group: SDKGroup, jobs: Job[]): void {
  const sdksWithOps = group.SDKs.filter((sdk) => sdk.SDK.Operations.length > 0);

  for (const availableSdk of sdksWithOps) {
    let title = group.Name;
    if (availableSdk.Deprecated) {
      title = `~~${title}~~`;
    }

    const sdkReadmeFileName = getSDKReadmeFileName(availableSdk.SDK);
    jobs.push(
      createTemplateFileJob("readme/sdk.stmpl", sdkReadmeFileName, {
        Title: title,
        ...availableSdk,
      }),
    );
  }

  for (const subGroup of group.SubGroups) {
    collectSDKDocsJobs(subGroup, jobs);
  }
}

/**
 * Recursively renders a markdown-formatted list for a given SDK group tree.
 * Heading levels increment only when a group has operations (except root).
 */
function renderAvailableSDKGroupsTree(
  group: SDKGroup,
  headingLevel: number,
): string {
  let result = "";

  // Groups without operations are skipped to avoid empty sections.
  const sdksWithOps = group.SDKs.filter((sdk) => sdk.SDK.Operations.length > 0);

  if (sdksWithOps.length > 0) {
    for (const availableSdk of sdksWithOps) {
      let title = group.Name;
      if (availableSdk.Deprecated) {
        title = `~~${title}~~`;
      }

      const sdkReadmeFileName = getSDKReadmeFileName(availableSdk.SDK);

      // Add group heading
      const sdkLink = `[${title}](${sdkReadmeFileName})`;
      result += `${"#".repeat(headingLevel)} ${sdkLink}\n\n`;

      // List available operations
      for (const op of availableSdk.SDK.Operations) {
        const deprecated = op.Comments && op.Comments.Deprecated;
        const opLinkName = deprecated
          ? `~~${sanitizeMethodName(op)}~~`
          : sanitizeMethodName(op);

        const opLinkTarget = `${sdkReadmeFileName}#${sanitizeMarkdownSlug(
          sanitizeMethodName(op),
        )}`;

        result += `* [${opLinkName}](${opLinkTarget})${templateOperationSummary(
          op,
        )}\n`;
      }
      result += "\n";
    }
  }

  for (const subGroup of group.SubGroups) {
    // Increment heading level only if the current group is root or has operations.
    // Otherwise keep same level to skip invisible intermediate groups.
    const isRoot = group.Path === "";
    const nextLevel =
      sdksWithOps.length > 0 && !isRoot ? headingLevel + 1 : headingLevel;
    result += renderAvailableSDKGroupsTree(subGroup, nextLevel);
  }

  return result;
}
