// =============================================================================
// README "Commands" section
//
// Mirrors the root help page: when the OpenAPI document declares a command
// surface (x-speakeasy-cli-commands), the section lists the declared help
// categories in manifest order — bound intents with their authored examples,
// planned placeholders with their availability note, and existing groups /
// built-ins tagged into a category — followed by everything the manifest
// leaves untagged under "Additional commands". Without a manifest it renders
// the plain operation tree (the previous "Available Commands" listing).
// =============================================================================

interface ReadmeCommandCatalog {
  cliName: string;
  manifest: IntentManifestCtx;
  rootOps: Map<string, Operation>; // root-level operation command name → op
  groups: Map<string, SDK>; // root-level group name → SubSDK
  rootAliasOwners: Map<string, string>; // alias → canonical root command name
  boundByKey: Map<string, IntentCmdCtx>; // "parent path|name" → bound intent
  plannedByKey: Map<string, PlannedCmdCtx>;
  groupTagByName: Map<string, GroupTagCtx>;
}

function readmeCommandKey(parentPath: string[], name: string): string {
  return `${parentPath.join(" ")}|${name}`;
}

function readmeBuiltinShort(name: string): string {
  switch (name) {
    case "configure":
      return templateConfigureShort();
    case "whoami":
      return templateWhoamiShort();
    case "auth":
      return "Manage authentication credentials";
    case "version":
      return "Print the CLI version";
    case "explore":
      return "Interactively browse and run commands";
    case "completion":
      return "Generate the autocompletion script for the specified shell";
  }
  return "";
}

function readmeDocPath(cliName: string, path: string[]): string {
  return `docs/${[cliName, ...path].join("_")}.md`;
}

function readmeCommandCatalog(sdk: SDK): ReadmeCommandCatalog {
  const cliName = sanitizeCliName();
  const manifest = collectIntentManifest();
  const rootOps = new Map<string, Operation>();
  const rootAliasOwners = new Map<string, string>();
  for (const op of sdk.Operations || []) {
    if (isOperationOverridden(op)) continue;
    const name = sanitizeCLICommand(op.GetID());
    rootOps.set(name, op);
    for (const alias of getOperationCommandAliases(op)) {
      rootAliasOwners.set(alias, name);
    }
  }
  const groups = new Map<string, SDK>();
  const groupAliases = getGroupAliases(sdk, "");
  for (const sub of sdk.SubSDKs || []) {
    const name = sanitizeCLICommand(getSDKGroupName(sub));
    groups.set(name, sub);
    for (const alias of groupAliases.get(name) || []) {
      rootAliasOwners.set(alias, name);
    }
  }
  const boundByKey = new Map<string, IntentCmdCtx>();
  for (const b of manifest.Bound) {
    boundByKey.set(readmeCommandKey(b.ParentPath, firstUseWord(b.Use)), b);
  }
  const plannedByKey = new Map<string, PlannedCmdCtx>();
  for (const p of manifest.Planned) {
    plannedByKey.set(readmeCommandKey(p.ParentPath, p.Name), p);
  }
  const groupTagByName = new Map<string, GroupTagCtx>();
  for (const g of manifest.GroupTags) {
    groupTagByName.set(g.Name, g);
  }
  return {
    cliName,
    manifest,
    rootOps,
    groups,
    rootAliasOwners,
    boundByKey,
    plannedByKey,
    groupTagByName,
  };
}

function readmeRootOwnerName(
  cat: ReadmeCommandCatalog,
  name: string,
): string | null {
  if (cat.groups.has(name) || cat.rootOps.has(name)) return name;
  return cat.rootAliasOwners.get(name) || null;
}

// Manifest examples render as an indented fenced block right under the
// command bullet: `# summary` (when authored) followed by the command line.
function readmeIntentExampleBlock(
  cat: ReadmeCommandCatalog,
  intent: IntentCmdCtx,
  indent: string,
): string {
  const lines = readmeIntentExamples(intent);
  if (lines.length === 0) return "";
  const summaries = readmeIntentExampleSummaries(intent);
  let out = `\n${indent}\`\`\`bash\n`;
  lines.forEach((line, i) => {
    const summary = summaries[i] || "";
    if (summary) out += `${indent}# ${summary}\n`;
    out += `${indent}${line}\n`;
  });
  out += `${indent}\`\`\`\n\n`;
  return out;
}

// Example summaries are not carried on IntentCmdCtx (help renders only the
// command lines); read them from the manifest by command ID.
function readmeIntentExampleSummaries(intent: IntentCmdCtx): string[] {
  const manifest = context.Global.AST.CLICommands;
  const cmd = (manifest?.Commands || []).find((c: any) => c.ID === intent.ID);
  return ((cmd?.Examples || []) as any[])
    .slice(0, 3)
    .map((e: any) => (e.Summary || "").trim());
}

function readmeBoundIntentEntry(
  cat: ReadmeCommandCatalog,
  intent: IntentCmdCtx,
  depth: number,
): string {
  const indent = "  ".repeat(depth);
  const name = firstUseWord(intent.Use);
  const path = [...intent.ParentPath, name];
  let line = `${indent}* [\`${name}\`](${readmeDocPath(cat.cliName, path)})`;
  if (intent.Summary) line += ` - ${intent.Summary}`;
  line += "\n";
  line += readmeIntentExampleBlock(cat, intent, indent + "  ");
  return line;
}

function readmePlannedEntry(
  cat: ReadmeCommandCatalog,
  planned: PlannedCmdCtx,
  depth: number,
): string {
  const indent = "  ".repeat(depth);
  const path = [...planned.ParentPath, planned.Name];
  let line = `${indent}* [\`${planned.Name}\`](${readmeDocPath(
    cat.cliName,
    path,
  )})`;
  const summary = [planned.Summary, planned.Tagline].filter((s) => s).join(" ");
  if (summary) line += ` - ${summary}`;
  line += ` — _not in this build_`;
  if (planned.Note) line += `: ${planned.Note}`;
  line += "\n";
  return line;
}

function readmeOpEntry(
  cat: ReadmeCommandCatalog,
  op: Operation,
  parentPath: string[],
  depth: number,
  sdkGroupName?: string,
): string {
  const indent = "  ".repeat(depth);
  const cmdName =
    sdkGroupName && !hasNameOverride(op)
      ? getDeStutteredCommandName(sdkGroupName, op.GetID())
      : sanitizeCLICommand(op.GetID());
  const summary = op.Comments?.Summary || "";
  const deprecated = op.Comments?.Deprecated;
  const displayName = deprecated ? `~~${cmdName}~~` : cmdName;
  let line = `${indent}* [\`${displayName}\`](${readmeDocPath(cat.cliName, [
    ...parentPath,
    cmdName,
  ])})`;
  if (summary) line += ` - ${summary}`;
  if (deprecated) line += ` :warning: **Deprecated**`;
  line += "\n";
  return line;
}

// A group renders as one bullet (its help summary) with its commands nested:
// declared intents/planned commands for that parent path first (manifest
// order), then the generated operations, then nested groups.
function readmeGroupEntry(
  cat: ReadmeCommandCatalog,
  group: SDK,
  parentPath: string[],
  depth: number,
): string {
  const indent = "  ".repeat(depth);
  const sdkGroupName = getSDKGroupName(group);
  const groupName = sanitizeCLICommand(sdkGroupName);
  const path = [...parentPath, groupName];

  let promotedOp: Operation | null = null;
  for (const op of group.Operations || []) {
    if (isOperationOverridden(op)) continue;
    if (
      !hasNameOverride(op) &&
      getStutterKind(sdkGroupName, op.GetID()) === "exact"
    ) {
      promotedOp = op;
      break;
    }
  }
  const visibleOps = (group.Operations || []).filter(
    (op: Operation) =>
      !isOperationOverridden(op) &&
      (hasNameOverride(op) ||
        getStutterKind(sdkGroupName, op.GetID()) !== "exact"),
  );

  const groupSummary =
    promotedOp?.Comments?.Summary || usageGroupHelp(group) || "";
  let out = `${indent}* [\`${groupName}\`](${readmeDocPath(
    cat.cliName,
    path,
  )})`;
  if (groupSummary) out += ` - ${groupSummary}`;
  out += "\n";

  // Declared commands mounted under this group, in manifest order.
  for (const entry of cat.manifest.Order) {
    if (entry.ParentPath.join(" ") !== path.join(" ")) continue;
    const key = readmeCommandKey(entry.ParentPath, entry.Name);
    const bound = cat.boundByKey.get(key);
    if (bound) {
      out += readmeBoundIntentEntry(cat, bound, depth + 1);
      continue;
    }
    const planned = cat.plannedByKey.get(key);
    if (planned) {
      const owned = visibleOps.some((op: Operation) => {
        const name = hasNameOverride(op)
          ? sanitizeCLICommand(op.GetID())
          : getDeStutteredCommandName(sdkGroupName, op.GetID());
        return (
          name === planned.Name ||
          getOperationCommandAliases(op).includes(planned.Name)
        );
      });
      if (!owned) out += readmePlannedEntry(cat, planned, depth + 1);
    }
  }
  for (const op of visibleOps) {
    out += readmeOpEntry(cat, op, path, depth + 1, sdkGroupName);
  }
  for (const sub of group.SubSDKs || []) {
    out += readmeGroupEntry(cat, sub, path, depth + 1);
  }
  return out;
}

function readmeBuiltinEntry(
  cat: ReadmeCommandCatalog,
  name: string,
  depth: number,
): string {
  const indent = "  ".repeat(depth);
  const short = readmeBuiltinShort(name);
  let line = `${indent}* [\`${name}\`](${readmeDocPath(cat.cliName, [name])})`;
  if (short) line += ` - ${short}`;
  line += "\n";
  return line;
}

// Render one root-level command by name: an existing group, a root
// operation, or a built-in.
function readmeExistingRootEntry(
  cat: ReadmeCommandCatalog,
  name: string,
): string {
  const ownerName = readmeRootOwnerName(cat, name) || name;
  const group = cat.groups.get(ownerName);
  if (group) return readmeGroupEntry(cat, group, [], 0);
  const op = cat.rootOps.get(ownerName);
  if (op) return readmeOpEntry(cat, op, [], 0);
  return readmeBuiltinEntry(cat, name, 0);
}

function readmeLegacyCommandTree(cat: ReadmeCommandCatalog, sdk: SDK): string {
  let result = "";
  for (const op of sdk.Operations || []) {
    result += readmeOpEntry(cat, op, [], 0);
  }
  for (const sub of sdk.SubSDKs || []) {
    result += readmeGroupEntry(cat, sub, [], 0);
  }
  return result;
}

function templateCLICommandsSection(sdk: SDK): string {
  const cat = readmeCommandCatalog(sdk);
  const m = cat.manifest;
  const declared =
    m.Bound.length > 0 || m.Planned.length > 0 || m.GroupTags.length > 0;

  if (!declared) {
    return (
      `<details open>\n<summary>Available commands</summary>\n\n` +
      readmeLegacyCommandTree(cat, sdk) +
      `\n</details>\n`
    );
  }

  let out =
    `Commands are grouped the way \`${cat.cliName} --help\` shows them. ` +
    `Every command accepts \`--help\`; body-bearing commands also accept ` +
    `\`--schema\` (exact request JSON Schema) and \`--dry-run\` (preview the ` +
    `request without sending it) — see [For AI agents](#for-ai-agents).\n\n`;

  // Root-level names consumed by a category (so they are not repeated below).
  const consumed = new Set<string>();
  const categoryOf = new Map<string, string>();
  for (const b of m.Bound) {
    if (b.ParentPath.length === 0)
      categoryOf.set(firstUseWord(b.Use), b.Category);
  }
  for (const p of m.Planned) {
    if (p.ParentPath.length === 0) categoryOf.set(p.Name, p.Category);
  }
  for (const g of m.GroupTags) categoryOf.set(g.Name, g.Category);

  for (const category of m.Categories) {
    let body = "";
    for (const entry of m.Order) {
      if (entry.ParentPath.length !== 0) continue;
      if (categoryOf.get(entry.Name) !== category) continue;
      const key = readmeCommandKey([], entry.Name);
      const bound = cat.boundByKey.get(key);
      if (bound) {
        body += readmeBoundIntentEntry(cat, bound, 0);
        consumed.add(entry.Name);
        continue;
      }
      const planned = cat.plannedByKey.get(key);
      if (planned) {
        // A planned placeholder yields to a real command that owns the name
        // (the runtime only re-tags the owner into the category).
        const ownerName = readmeRootOwnerName(cat, entry.Name);
        if (ownerName) {
          body += readmeExistingRootEntry(cat, entry.Name);
          consumed.add(ownerName);
        } else {
          body += readmePlannedEntry(cat, planned, 0);
        }
        consumed.add(entry.Name);
        continue;
      }
      if (cat.groupTagByName.has(entry.Name)) {
        body += readmeExistingRootEntry(cat, entry.Name);
        consumed.add(entry.Name);
        const ownerName = readmeRootOwnerName(cat, entry.Name);
        if (ownerName) consumed.add(ownerName);
      }
    }
    if (!body) continue;
    out += `### ${category}\n\n${body.trimEnd()}\n\n`;
  }

  // Everything the manifest leaves untagged (mirrors "Additional Commands").
  let rest = "";
  for (const op of sdk.Operations || []) {
    const name = sanitizeCLICommand(op.GetID());
    if (consumed.has(name)) continue;
    rest += readmeOpEntry(cat, op, [], 0);
  }
  for (const sub of sdk.SubSDKs || []) {
    const name = sanitizeCLICommand(getSDKGroupName(sub));
    if (consumed.has(name)) continue;
    rest += readmeGroupEntry(cat, sub, [], 0);
  }
  if (rest) {
    out += `### Additional commands\n\n${rest}\n`;
  }
  return out.trimEnd() + "\n";
}
