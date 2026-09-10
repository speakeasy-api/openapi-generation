// =============================================================================
// README example selection
//
// The generic README sections (Output Formats, Diagnostics, Retries, Request
// Body Input, Pagination, Streaming, ...) illustrate CLI-wide behaviour, so
// they used to print `<command>` placeholders. Placeholders are not runnable
// and rot silently, so every section now renders a real command chosen from
// the generated command tree (and the declared intent manifest) by the
// deterministic pickers below. Every picker degrades explicitly: when no
// candidate satisfies its capability predicate the section either falls back
// to a weaker-but-real candidate or reports "" so the template can omit the
// example instead of inventing one.
//
// The README example contract test (readme_examples_test.go.stmpl) executes
// every rendered `<cli> ...` line under --dry-run, so a picker that produced
// an unrunnable invocation fails the generated test suite.
// =============================================================================

interface ReadmeOpEntry {
  op: Operation;
  path: string[]; // CLI command path without the CLI name
}

interface ReadmeBodyFlag {
  Flag: string;
  Value: string; // shell-ready value (quoted when needed)
  Arg: string; // "--flag value" / "--flag=true" as it appears on the command line
  Kind: string; // "string" | "number" | "boolean"
}

interface ReadmeBodyExample {
  Command: string; // "<cli> create-user" (path only, no inputs)
  BodyFlag: string; // flag that accepts the whole JSON body ("body", "body-param", ...)
  BodyJSON: string; // compact JSON body (from the spec example)
  Flags: ReadmeBodyFlag[]; // individual field flags with their example values
  FirstFlag: ReadmeBodyFlag | null; // the flag used to demonstrate precedence over --body / stdin
  FirstFlagOverrideJSON: string; // BodyJSON with that flag's value replaced
  OverrideValue: string; // the replacement value, rendered as it appears on the command line
  HasFlags: boolean;
  HasSchema: boolean;
}

interface ReadmeExampleContext {
  CliName: string;
  Showcase: string; // "<cli> agent list" — a runnable command with no required inputs
  ShowcasePath: string; // "agent list"
  ShowcaseArgs: string; // Showcase without the leading CLI name — runnable after global flags
  Paginated: string; // "<cli> agent list" for a paginated operation ("" = none)
  Streaming: string; // full runnable invocation of a streaming command, SSE first ("" = none)
  StreamingSSE: string; // ("" = no SSE command)
  StreamingJSONL: string; // ("" = no JSONL command)
  Body: ReadmeBodyExample | null;
  Schema: string; // "<cli> invite" — a command exposing --schema ("" = none)
  Intent: string; // first bound intent example command ("" = none)
  IntentName: string; // command display path of that intent
  IntentHasArtifact: boolean;
  Artifact: string; // first artifact-producing intent example ("" = none)
  ArtifactDefaultPath: string;
  Async: string; // first async intent example ("" = none)
  AsyncResume: string; // generated poll command prefix, without the handle
  StreamSelect: string; // first intent example with a streamed projection ("" = none)
  StreamSelectPointer: string;
  HasOperationStreamProjection: boolean;
  Hints: string; // first intent example that declares hints ("" = none)
  HintsJSON: string;
  HintsIntentName: string;
  HintsReasons: string; // comma-joined hint reasons of that intent
  HasIntents: boolean;
  UploadCommand: string; // "<cli> post-file" — first multipart upload command ("" = none)
  UploadFlag: string; // its file flag name ("upload")
}

let readmeExampleContextCache: ReadmeExampleContext | null = null;

// Every operation reachable from the CLI command tree, in document order.
function readmeAllOperations(): ReadmeOpEntry[] {
  const out: ReadmeOpEntry[] = [];
  const walk = (sdk: SDK) => {
    for (const op of sdk.Operations || []) {
      if (isOperationOverridden(op)) continue;
      out.push({ op, path: getCLICommandPath(op) });
    }
    for (const sub of sdk.SubSDKs || []) walk(sub);
  };
  walk(context.Global.AST.MainSDK);
  return out;
}

// readmeOpRequiresInputs reports whether invoking the operation without any
// flags fails validation: required parameters, or a request body that is not
// optional.
function readmeOpRequiresInputs(op: Operation): boolean {
  if (!op.Request) return false;
  if (intentOperationHasRequiredParams(op)) return true;
  if (op.Request.IsRequestBody) {
    return !op.Request.RequestBody?.Optional;
  }
  const bodyField = (op.Request.Field?.Type?.Fields || []).find(
    (f: FieldDef) => f.Annotations?.Has("request"),
  );
  return Boolean(bodyField) && !bodyField.Optional;
}

function readmeOpIsJSONBody(op: Operation): boolean {
  if (!op.Request) return false;
  const ser = op.SerializationMethod?.toString() || "";
  if (ser === "multipart" || ser === "form") return false;
  if (op.Request.IsRequestBody) {
    if (!op.Request.RequestBody) return false;
    const t = op.Request.RequestBody.Type?.Type?.toString() || "";
    return t !== "bytes" && t !== "string";
  }
  if (isMultipartMixedOp(op)) return false;
  return getBodyFieldPath(op) !== "";
}

// A body example is unsafe for a README when it carries credential-looking
// keys (they would be copied verbatim into shell history and issue trackers).
function readmeExampleHasSecrets(value: any, depth: number = 0): boolean {
  if (depth > 8 || value === null || typeof value !== "object") return false;
  if (Array.isArray(value)) {
    return value.some((v) => readmeExampleHasSecrets(v, depth + 1));
  }
  for (const key of Object.keys(value)) {
    if (
      /password|secret|token|api[_-]?key|credential|private[_-]?key/i.test(key)
    ) {
      return true;
    }
    if (readmeExampleHasSecrets(value[key], depth + 1)) return true;
  }
  return false;
}

// Generated examples fall back to `<value>`-style placeholders for fields
// without a spec example; a README example must never show one.
function readmeExampleHasPlaceholders(value: any, depth: number = 0): boolean {
  if (depth > 8 || value === null || value === undefined) return false;
  if (typeof value === "string") return /^<[^<>]+>$/.test(value.trim());
  if (typeof value !== "object") return false;
  if (Array.isArray(value)) {
    return value.some((v) => readmeExampleHasPlaceholders(v, depth + 1));
  }
  return Object.keys(value).some((k) =>
    readmeExampleHasPlaceholders(value[k], depth + 1),
  );
}

function readmeCompareEntries(a: ReadmeOpEntry, b: ReadmeOpEntry): number {
  return a.path.join(" ").localeCompare(b.path.join(" "));
}

// Deterministic best-of: highest score wins, ties broken by command path so
// regeneration is stable regardless of document order changes elsewhere.
function readmePickBest(
  entries: ReadmeOpEntry[],
  score: (e: ReadmeOpEntry) => number,
): ReadmeOpEntry | null {
  let best: ReadmeOpEntry | null = null;
  let bestScore = -Infinity;
  for (const e of entries) {
    const s = score(e);
    if (s < 0) continue;
    if (
      best === null ||
      s > bestScore ||
      (s === bestScore && readmeCompareEntries(e, best) < 0)
    ) {
      best = e;
      bestScore = s;
    }
  }
  return best;
}

// The showcase command illustrates flags that apply to every command
// (--output-format, --jq, --dry-run, --debug, retries, server selection...).
// It must run with no inputs at all: a read-only operation without required
// parameters or body, preferring list-style operations.
function readmeShowcaseEntry(): ReadmeOpEntry | null {
  return readmePickBest(readmeAllOperations(), (e) => {
    const op = e.op;
    if (readmeOpRequiresInputs(op)) return -1;
    if (op.Comments?.Deprecated) return -1;
    if (hasBinaryResponse(op)) return -1;
    let s = 0;
    const method = (op.Method || "").toUpperCase();
    if (method === "GET") s += 4;
    if (hasPagination(op)) s += 2;
    if (!op.Request) s += 1;
    if (/^list/i.test(op.GetID())) s += 2;
    else if (/^(get|search|describe)/i.test(op.GetID())) s += 1;
    if (e.path.length <= 2) s += 1;
    return s;
  });
}

function templateDryRunShowcaseTestEnabled(): boolean {
  return readmeShowcaseEntry() !== null;
}
registerTemplateFunc(
  "templateDryRunShowcaseTestEnabled",
  templateDryRunShowcaseTestEnabled,
);

function templateDryRunShowcaseTestArgs(): string {
  const entry = readmeShowcaseEntry();
  if (!entry) return "";
  return entry.path.map((part) => goStringLiteral(part)).join(", ");
}
registerTemplateFunc(
  "templateDryRunShowcaseTestArgs",
  templateDryRunShowcaseTestArgs,
);

// Content for a single-quoted shell literal ('...'): the only character that
// needs care is the single quote itself.
function readmeShellSingleQuoted(s: string): string {
  return s.replace(/'/g, "'\\''");
}

// A body field flag with its example value as it appears on the command
// line: numbers bare, booleans in the `--flag=value` form (a bool flag takes
// no separate argument), strings single-quoted.
function readmeBodyFlag(flag: string, value: any): ReadmeBodyFlag {
  if (typeof value === "boolean") {
    return {
      Flag: flag,
      Value: String(value),
      Arg: `--${flag}=${value}`,
      Kind: "boolean",
    };
  }
  if (typeof value === "number") {
    return {
      Flag: flag,
      Value: String(value),
      Arg: `--${flag} ${value}`,
      Kind: "number",
    };
  }
  const quoted = `'${readmeShellSingleQuoted(String(value))}'`;
  return {
    Flag: flag,
    Value: quoted,
    Arg: `--${flag} ${quoted}`,
    Kind: "string",
  };
}

function readmeInvocation(cliName: string, path: string[]): string {
  return [cliName, ...path].join(" ");
}

// Runnable invocation for an operation that needs inputs: rendered through
// the usage-example machinery (spec examples, then generated examples).
function readmeUsageInvocation(op: Operation): string {
  try {
    const ctx = createUsageContextTS(context.Global.AST.MainSDK, op);
    const cmd = templateCLIUsageCommand(ctx);
    return cmd || "";
  } catch (_e) {
    return "";
  }
}

// Manifest examples are authored as full command lines ("<cli> generate ...").
function readmeIntentExamples(intent: IntentCmdCtx): string[] {
  return (intent.Example || "")
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
}

function readmeBodyExample(): ReadmeBodyExample | null {
  const cliName = sanitizeCliName();
  const entries = readmeAllOperations();
  const candidates: {
    entry: ReadmeOpEntry;
    payload: Record<string, any> | undefined;
    flags: ReadmeBodyFlag[];
    score: number;
  }[] = [];
  for (const entry of entries) {
    const op = entry.op;
    if (!readmeOpIsJSONBody(op)) continue;
    if (op.Comments?.Deprecated) continue;
    // The section is about body input; a command whose other inputs are
    // required would need extra flags on every line.
    if (intentOperationHasRequiredParams(op)) continue;
    let payload: Record<string, any> | undefined;
    try {
      payload = extractBodyExamplePayload(op, undefined);
    } catch (_e) {
      payload = undefined;
    }
    if (payload && (typeof payload !== "object" || Array.isArray(payload))) {
      payload = undefined;
    }
    if (payload && readmeExampleHasSecrets(payload)) continue;
    // A placeholder-bearing example is worse than none; without a usable
    // example the command cannot illustrate the section.
    if (!payload || readmeExampleHasPlaceholders(payload)) continue;

    const flags: ReadmeBodyFlag[] = [];
    if (op.Request.IsRequestBody && isRequestBodyExpandable(op)) {
      const bodyFields = (op.Request.RequestBody.Type?.Fields || []).filter(
        (field: FieldDef) => !field.Const,
      );
      const renderFlag = (field: FieldDef): ReadmeBodyFlag | null => {
        const key = field.OriginalName || field.Name;
        const value = payload[key];
        if (value === undefined || value === null) return null;
        const ft = field.Type?.Type?.toString() || "";
        // Scalars only: they read naturally as `--flag value`.
        if (typeof value === "object") return null;
        if (fieldNeedsJSONFormat(field)) return null;
        if (ft === "bytes") return null;
        return readmeBodyFlag(sanitizeFlagNameWithReserved(field.Name), value);
      };

      // A runnable individual-flags example must provide every required body
      // field. Required fields are never capped; after them, add optional
      // scalar fields until the former two-flag display target is reached.
      let requiredRenderable = true;
      for (const field of bodyFields.filter(
        (field: FieldDef) => !field.Optional,
      )) {
        const flag = renderFlag(field);
        if (!flag) {
          requiredRenderable = false;
          break;
        }
        flags.push(flag);
      }
      if (!requiredRenderable) {
        flags.length = 0;
      } else {
        for (const field of bodyFields.filter(
          (field: FieldDef) => field.Optional,
        )) {
          if (flags.length >= 2) break;
          const flag = renderFlag(field);
          if (flag) flags.push(flag);
        }
      }
    }
    let score = 4;
    if (flags.length >= 2) score += 3;
    else if (flags.length === 1) score += 1;
    if (JSON.stringify(payload).length <= 160) score += 2;
    if (!readmeOpRequiresInputs(op)) score += 1;
    candidates.push({ entry, payload, flags, score });
  }
  if (candidates.length === 0) return null;
  candidates.sort(
    (a, b) => b.score - a.score || readmeCompareEntries(a.entry, b.entry),
  );
  const pick = candidates[0];
  const op = pick.entry.op;
  const bodyFlag = templateHasBodyFlag(op)
    ? "body"
    : op.Request.IsRequestBody
    ? getRequestBodyFlagName(op.Request.RequestBody)
    : wholeBodyFlagName(op) || "body";
  const payload = pick.payload;
  // Precedence demo: a number or boolean reads most naturally when changed
  // (30 → 31, true → false); fall back to the first string flag.
  const first =
    pick.flags.find((f) => f.Kind !== "string") || pick.flags[0] || null;
  let overrideJSON = "";
  let overrideValue = "";
  if (first) {
    const field = (op.Request.RequestBody.Type?.Fields || []).find(
      (f: FieldDef) => sanitizeFlagNameWithReserved(f.Name) === first.Flag,
    );
    const key = field ? field.OriginalName || field.Name : first.Flag;
    const current = payload[key];
    let replacement: any;
    if (typeof current === "number") replacement = current + 1;
    else if (typeof current === "boolean") replacement = !current;
    else replacement = `${String(current)} (updated)`;
    overrideJSON = readmeShellSingleQuoted(
      JSON.stringify({ ...payload, [key]: replacement }),
    );
    overrideValue = readmeBodyFlag(first.Flag, replacement).Arg;
  }
  return {
    Command: readmeInvocation(cliName, pick.entry.path),
    BodyFlag: bodyFlag,
    BodyJSON: readmeShellSingleQuoted(JSON.stringify(payload)),
    Flags: pick.flags,
    FirstFlag: first,
    FirstFlagOverrideJSON: overrideJSON,
    OverrideValue: overrideValue,
    HasFlags: pick.flags.length > 0,
    HasSchema: hasBodySchemaForOp(op),
  };
}

// Runnable invocation of a command whose response streams as `kind`
// ("sse" | "jsonl"): a declared intent that streams by construction (preset
// stream flag or a declared --stream flag) with an authored example, else the
// first streaming operation. "" when the CLI has no such command.
function readmeStreamingPick(kind: string): string {
  const cliName = sanitizeCliName();
  const manifest = collectIntentManifest();
  for (const intent of manifest.Bound) {
    if (intent.StreamKind !== kind) continue;
    const examples = readmeIntentExamples(intent);
    if (examples.length === 0) continue;
    const presetStreams = /"stream"\s*:\s*true/.test(intent.PresetJSON || "");
    const streamFlag = intent.Flags.find((f) => f.Key === "stream");
    if (!presetStreams && !streamFlag) continue;
    let line =
      examples.find((l) => streamFlag && l.includes(`--${streamFlag.Name}`)) ||
      examples[0];
    if (
      !presetStreams &&
      streamFlag &&
      !line.includes(`--${streamFlag.Name}`)
    ) {
      line = `${line} --${streamFlag.Name}`;
    }
    return line;
  }
  const entry = readmePickBest(readmeAllOperations(), (e) => {
    if (intentStreamKind(e.op).kind !== kind) return -1;
    if (e.op.Comments?.Deprecated) return -1;
    let s = 1;
    if (!readmeOpRequiresInputs(e.op)) s += 4;
    return s;
  });
  if (!entry) return "";
  return readmeOpRequiresInputs(entry.op)
    ? readmeUsageInvocation(entry.op)
    : readmeInvocation(cliName, entry.path);
}

function readmeExampleContext(): ReadmeExampleContext {
  if (readmeExampleContextCache) return readmeExampleContextCache;
  const cliName = sanitizeCliName();
  const manifest = collectIntentManifest();

  const showcaseEntry = readmeShowcaseEntry();
  let showcasePath = showcaseEntry ? showcaseEntry.path.join(" ") : "";
  let showcase = showcaseEntry
    ? readmeInvocation(cliName, showcaseEntry.path)
    : "";
  if (!showcase) {
    // No input-free operation: fall back to the first intent example, then to
    // the first operation rendered with example values.
    const intentWithExample = manifest.Bound.find(
      (i) => readmeIntentExamples(i).length > 0,
    );
    if (intentWithExample) {
      showcase = readmeIntentExamples(intentWithExample)[0];
      showcasePath = intentWithExample.CommandDisplay;
    } else {
      const first = readmeAllOperations()[0];
      if (first) {
        showcase =
          readmeUsageInvocation(first.op) ||
          readmeInvocation(cliName, first.path);
        showcasePath = first.path.join(" ");
      }
    }
  }
  if (!showcase) {
    showcase = `${cliName} version`;
    showcasePath = "version";
  }
  const showcaseArgs = showcase.startsWith(`${cliName} `)
    ? showcase.slice(cliName.length + 1)
    : showcasePath;

  const paginatedEntry = readmePickBest(readmeAllOperations(), (e) => {
    if (!hasPagination(e.op)) return -1;
    if (e.op.Comments?.Deprecated) return -1;
    let s = 1;
    if (!readmeOpRequiresInputs(e.op)) s += 4;
    return s;
  });
  let paginated = "";
  if (paginatedEntry) {
    paginated = readmeOpRequiresInputs(paginatedEntry.op)
      ? readmeUsageInvocation(paginatedEntry.op)
      : readmeInvocation(cliName, paginatedEntry.path);
  }

  const streamingSSE = readmeStreamingPick("sse");
  const streamingJSONL = readmeStreamingPick("jsonl");
  const body = readmeBodyExample();

  // --schema lives on every body-bearing command with a bundled schema;
  // prefer an intent (agents meet those first), then the body example's op.
  let schema = "";
  const schemaIntent = manifest.Bound.find((i) => i.HasSchema);
  if (schemaIntent) {
    schema = `${cliName} ${schemaIntent.CommandDisplay}`;
  } else if (body && body.HasSchema) {
    schema = body.Command;
  } else {
    const entry = readmePickBest(readmeAllOperations(), (e) =>
      hasBodySchemaForOp(e.op) ? (readmeOpRequiresInputs(e.op) ? 1 : 2) : -1,
    );
    if (entry) schema = readmeInvocation(cliName, entry.path);
  }

  // File uploads: the first multipart operation with a file field and no
  // other required inputs (a missing file is the only thing --dry-run cannot
  // tolerate, so the README marks this block as not executed by the contract
  // test).
  let uploadCommand = "";
  let uploadFlag = "";
  const uploadEntry = readmePickBest(readmeAllOperations(), (e) => {
    const op = e.op;
    if (op.SerializationMethod?.toString() !== "multipart") return -1;
    if (!op.Request?.IsRequestBody || !op.Request.RequestBody) return -1;
    if (op.Comments?.Deprecated) return -1;
    const fields = op.Request.RequestBody.Type?.Fields || [];
    const fileField = fields.find((f: FieldDef) => isMultipartFileField(f));
    if (!fileField) return -1;
    const otherRequired = fields.some(
      (f: FieldDef) => f !== fileField && !f.Const && !f.Optional,
    );
    let s = 1;
    if (!otherRequired && !intentOperationHasRequiredParams(op)) s += 4;
    return s;
  });
  if (uploadEntry) {
    const fields = uploadEntry.op.Request.RequestBody.Type?.Fields || [];
    const fileField = fields.find((f: FieldDef) => isMultipartFileField(f));
    uploadCommand = readmeInvocation(cliName, uploadEntry.path);
    uploadFlag = sanitizeFlagNameWithReserved(fileField.Name);
  }

  const firstIntent = manifest.Bound.find(
    (i) => readmeIntentExamples(i).length > 0,
  );
  const artifactIntent = manifest.Bound.find(
    (i) => i.ArtifactJSON && readmeIntentExamples(i).length > 0,
  );
  const asyncIntent = manifest.Bound.find(
    (i) => i.AsyncJSON && readmeIntentExamples(i).length > 0,
  );
  const streamSelectIntent = manifest.Bound.find(
    (i) => i.StreamSelect && readmeIntentExamples(i).length > 0,
  );
  const hintsIntent = manifest.Bound.find(
    (i) => i.HintsJSON && readmeIntentExamples(i).length > 0,
  );

  readmeExampleContextCache = {
    CliName: cliName,
    Showcase: showcase,
    ShowcasePath: showcasePath,
    ShowcaseArgs: showcaseArgs,
    Paginated: paginated,
    Streaming: streamingSSE || streamingJSONL,
    StreamingSSE: streamingSSE,
    StreamingJSONL: streamingJSONL,
    Body: body,
    Schema: schema,
    Intent: firstIntent ? readmeIntentExamples(firstIntent)[0] : "",
    IntentName: firstIntent ? firstIntent.CommandDisplay : "",
    IntentHasArtifact: Boolean(firstIntent && firstIntent.ArtifactJSON),
    Artifact: artifactIntent ? readmeIntentExamples(artifactIntent)[0] : "",
    ArtifactDefaultPath: artifactIntent
      ? artifactIntent.ArtifactDefaultPath
      : "",
    Async: asyncIntent ? readmeIntentExamples(asyncIntent)[0] : "",
    AsyncResume: asyncIntent
      ? String(JSON.parse(asyncIntent.AsyncJSON).resume || "")
      : "",
    StreamSelect: streamSelectIntent
      ? readmeIntentExamples(streamSelectIntent)[0]
      : "",
    StreamSelectPointer: streamSelectIntent
      ? streamSelectIntent.StreamSelect
      : "",
    HasOperationStreamProjection: Boolean(
      (context.Global.AST.CLICommands as any)?.Operations?.some(
        (op: any) => op.Output?.Stream,
      ),
    ),
    Hints: hintsIntent ? readmeIntentExamples(hintsIntent)[0] : "",
    HintsJSON: hintsIntent ? hintsIntent.HintsJSON : "",
    HintsIntentName: hintsIntent ? hintsIntent.CommandDisplay : "",
    HintsReasons: hintsIntent
      ? Object.keys(JSON.parse(hintsIntent.HintsJSON)).join(", ")
      : "",
    HasIntents: manifest.Bound.length > 0 || manifest.Planned.length > 0,
    UploadCommand: uploadCommand,
    UploadFlag: uploadFlag,
  };
  return readmeExampleContextCache;
}

// Template accessors ---------------------------------------------------------

function templateReadmeShowcaseCommand(): string {
  return readmeExampleContext().Showcase;
}
registerTemplateFunc(
  "templateReadmeShowcaseCommand",
  templateReadmeShowcaseCommand,
);

function templateReadmeShowcasePath(): string {
  return readmeExampleContext().ShowcasePath;
}
registerTemplateFunc("templateReadmeShowcasePath", templateReadmeShowcasePath);
