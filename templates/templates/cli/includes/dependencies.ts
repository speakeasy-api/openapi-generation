// Load config.ts to make getTemplateDependencies available globally

// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

// @ts-ignore
function isInteractiveAuthEnabled(): boolean {
  const val = context.Global.Config.InteractiveAuth;
  if (val === undefined || val === null) return true; // default ON
  return val === true || val === "true";
}
registerTemplateFunc("isInteractiveAuthEnabled", isInteractiveAuthEnabled);

// @ts-ignore
function isInteractiveModeEnabled(): boolean {
  const val = context.Global.Config.InteractiveMode;
  if (val === undefined || val === null) return true; // default ON
  return val === true || val === "true";
}
registerTemplateFunc("isInteractiveModeEnabled", isInteractiveModeEnabled);

// interactiveByDefault controls runtime behavior only. interactiveMode and
// interactiveAuth still control which interactive capabilities are generated.
// @ts-ignore
function interactiveByDefault(): boolean {
  const val = context.Global.Config.InteractiveByDefault;
  if (val === undefined || val === null) return true; // compatibility default
  return val === true || val === "true";
}
registerTemplateFunc("interactiveByDefault", interactiveByDefault);

// agentEnvironmentDetection controls whether the generated runtime contains
// compatibility auto-detection for known agent environment variables.
// @ts-ignore
function agentEnvironmentDetection(): boolean {
  const val = context.Global.Config.AgentEnvironmentDetection;
  if (val === undefined || val === null) return true; // compatibility default
  return val === true || val === "true";
}
registerTemplateFunc("agentEnvironmentDetection", agentEnvironmentDetection);

// classifiedErrors controls whether explicit machine-rendering requests receive
// the same reason-first classification and actionable hints as agent mode.
// @ts-ignore
function classifiedErrors(): boolean {
  const val = context.Global.Config.ClassifiedErrors;
  if (val === undefined || val === null) return false; // compatibility default
  return val === true || val === "true";
}
registerTemplateFunc("classifiedErrors", classifiedErrors);

// @ts-ignore
function isAnyInteractiveEnabled(): boolean {
  return isInteractiveAuthEnabled() || isInteractiveModeEnabled();
}
registerTemplateFunc("isAnyInteractiveEnabled", isAnyInteractiveEnabled);

// interactiveFlagHelp keeps Cobra help and the KDL usage contract aligned
// across the independently generated prompt and auth capabilities.
function interactiveFlagHelp(): string {
  const mode = isInteractiveModeEnabled();
  const auth = isInteractiveAuthEnabled();
  if (mode && auth) {
    return "Prompt for missing inputs and open guided configure/auth forms (forms fall back to line prompts on stdin off-TTY)";
  }
  if (mode) {
    return "Prompt for missing inputs (requires a terminal)";
  }
  if (auth) {
    return "Open guided configure/auth forms (line prompts on stdin off-TTY)";
  }
  return "";
}
registerTemplateFunc("interactiveFlagHelp", interactiveFlagHelp);

function dryRunLocalMutationCommands(): string[] {
  const commands: string[] = [];
  if (isInteractiveAuthEnabled() && hasGlobalSecurity()) {
    commands.push("auth login", "auth logout");
  }
  commands.push("configure");
  return commands;
}

function joinNatural(items: string[]): string {
  if (items.length === 1) return items[0];
  return `${items.slice(0, -1).join(", ")} and ${items[items.length - 1]}`;
}

function dryRunFlagHelp(): string {
  const commands = dryRunLocalMutationCommands();
  const preview =
    "Preview API requests without sending them (no network, no OS keychain). Human preview on stderr; with -o json or --jq, one JSON object per request on stdout.";
  const machine = "(stderr, or one JSON object on stdout in the machine form)";
  if (commands.length === 1) {
    return `${preview} The local mutation command (${commands[0]}) makes no request: it skips prompts and writes and reports a no-op ${machine}`;
  }
  return `${preview} Local mutation commands (${joinNatural(
    commands,
  )}) make no request: they skip prompts and writes and report a no-op ${machine}`;
}
registerTemplateFunc("dryRunFlagHelp", dryRunFlagHelp);

function dryRunLocalCommandsMarkdown(): string {
  return joinNatural(dryRunLocalMutationCommands().map((c) => `\`${c}\``));
}
registerTemplateFunc(
  "dryRunLocalCommandsMarkdown",
  dryRunLocalCommandsMarkdown,
);

// @ts-ignore
function isCustomCommandsEnabled(): boolean {
  const val = context.Global.Config.EnableCustomCodeRegions;
  return val === true || val === "true";
}
registerTemplateFunc("isCustomCommandsEnabled", isCustomCommandsEnabled);

// helpStyle resolves cli.helpStyle so generated templates never see auto:
//   compact — concise per-command help; inherited globals live behind the
//             root-local --help-global flag
//   full    — the established Cobra help output, byte-for-byte
//   auto    — compact only for a validated non-empty declared-command manifest
// @ts-ignore
function helpStyle(): "compact" | "full" {
  const val = context.Global.Config.HelpStyle;
  if (val === "compact" || val === "full") return val;
  const manifest = context.Global.AST.CLICommands;
  return manifest?.Commands?.length > 0 ? "compact" : "full";
}
registerTemplateFunc("helpStyle", helpStyle);

// retryFlagsVisibility controls the retry control flags (--no-retries,
// --retry-config, --retry-connection-errors, --retry-max-elapsed-time):
//   visible — registered and shown in help (default)
//   hidden  — registered (still parse; config/env behavior unchanged) but
//             omitted from help output
//   none    — not registered at all
// @ts-ignore
function retryFlagsVisibility(): string {
  const val = context.Global.Config.RetryFlagsVisibility;
  if (val === "hidden" || val === "none") return val;
  return "visible";
}
registerTemplateFunc("retryFlagsVisibility", retryFlagsVisibility);

// retryMethodPolicy controls whether root retry policies follow the document
// exactly or are inherited only by safe HTTP methods.
// @ts-ignore
function retryMethodPolicy(): string {
  return context.Global.Config.RetryMethodPolicy === "safe-methods"
    ? "safe-methods"
    : "spec";
}
registerTemplateFunc("retryMethodPolicy", retryMethodPolicy);

// serverSelectionIsMeaningful reports whether --server has an alternative
// value to select: multiple global servers, or multiple servers on any
// operation reachable through the SDK tree. Named (ServerMap) and indexed
// server collections both expose their entries through Servers.
function serverSelectionIsMeaningful(): boolean {
  const mainSDK = context.Global.AST.MainSDK;
  if ((mainSDK.Servers?.Servers?.length ?? 0) > 1) return true;

  const walk = (sdk: SDK): boolean => {
    for (const op of sdk.Operations || []) {
      if ((op.Servers?.Servers?.length ?? 0) > 1) return true;
    }
    for (const sub of sdk.SubSDKs || []) {
      if (walk(sub)) return true;
    }
    return false;
  };

  return walk(mainSDK);
}

// serverSelectionFlag resolves cli.serverSelectionFlag so templates never see
// auto:
//   visible — registered and shown in help
//   hidden  — registered (still parses) but omitted from help output
//   none    — not registered at all
//   auto    — visible when serverSelectionIsMeaningful is true, hidden
//             otherwise: the flag stays parseable so existing --server
//             invocations survive regeneration; only explicit none breaks it
// @ts-ignore
function serverSelectionFlag(): "visible" | "hidden" | "none" {
  const val = context.Global.Config.ServerSelectionFlag;
  if (val === "visible" || val === "hidden" || val === "none") return val;
  return serverSelectionIsMeaningful() ? "visible" : "hidden";
}
registerTemplateFunc("serverSelectionFlag", serverSelectionFlag);

// defaultColorMode is the gen.yaml default for the --color flag
// (auto|always|never; auto when unset or invalid).
// @ts-ignore
function defaultColorMode(): string {
  const val = context.Global.Config.DefaultColor;
  if (val === "always" || val === "never") return val;
  return "auto";
}
registerTemplateFunc("defaultColorMode", defaultColorMode);

// jqRawOutputDefault is the gen.yaml default for the --raw-output flag: when
// true, --jq string results are written raw (jq -r semantics).
// @ts-ignore
function jqRawOutputDefault(): boolean {
  const val = context.Global.Config.JqRawOutput;
  return val === true || val === "true";
}
registerTemplateFunc("jqRawOutputDefault", jqRawOutputDefault);

// cliDefaultTimeout is the gen.yaml cli.defaultTimeout value, the implicit
// per-operation network timeout, as the integer nanosecond count spliced into
// a time.Duration(...) expression in the generated Go, or "" when unset. The
// value is re-validated here against the anchored duration grammar (the
// config regex is the first line of defense) so that nothing but digits ever
// reaches Go source, and the sum is checked against int64 nanoseconds, which
// Go's time.ParseDuration would otherwise reject at runtime. A value of zero
// is treated as unset.
// @ts-ignore
function cliDefaultTimeout(): string {
  const val = context.Global.Config.DefaultTimeout;
  const raw = typeof val === "string" ? val.trim() : "";
  if (raw === "") return "";
  if (!/^([0-9]+(?:\.[0-9]+)?(?:ns|us|µs|ms|s|m|h))+$/.test(raw)) {
    throw new Error(
      `cli.defaultTimeout ${JSON.stringify(
        raw,
      )} is not a Go duration such as 60s or 1m30s`,
    );
  }
  const unitNs: Record<string, bigint> = {
    ns: 1n,
    us: 1000n,
    µs: 1000n,
    ms: 1000000n,
    s: 1000000000n,
    m: 60000000000n,
    h: 3600000000000n,
  };
  let totalNs = 0n;
  for (const [, num, unit] of raw.matchAll(
    /([0-9]+(?:\.[0-9]+)?)(ns|us|µs|ms|s|m|h)/g,
  )) {
    const [whole, frac = ""] = num.split(".");
    const scale = unitNs[unit];
    totalNs +=
      BigInt(whole) * scale +
      (frac ? (BigInt(frac) * scale) / 10n ** BigInt(frac.length) : 0n);
  }
  if (totalNs > 9223372036854775807n) {
    throw new Error(
      `cli.defaultTimeout ${raw} exceeds the maximum representable duration (about 2540400h)`,
    );
  }
  if (totalNs === 0n) return "";
  return `${totalNs}`;
}
registerTemplateFunc("cliDefaultTimeout", cliDefaultTimeout);

// Interactive theme color helpers.
// Reads hex colors from interactiveTheme config and returns Go lipgloss expressions.

const themeDefaults: Record<string, string> = {
  accentColor: "#38BDF8",
  dimmedColor: "#64748B",
  subtleColor: "#475569",
  errorColor: "#F87171",
  successColor: "#4ADE80",
};

// @ts-ignore
function getThemeColor(key: string): string {
  const theme = context.Global.Config.InteractiveTheme;
  if (theme && typeof theme === "object" && theme[key]) {
    return String(theme[key]);
  }
  return themeDefaults[key] || "#FFFFFF";
}

registerTemplateFunc(
  "themeAccent",
  () => `lipgloss.Color("${getThemeColor("accentColor")}")`,
);
registerTemplateFunc(
  "themeDimmed",
  () => `lipgloss.Color("${getThemeColor("dimmedColor")}")`,
);
registerTemplateFunc(
  "themeSubtle",
  () => `lipgloss.Color("${getThemeColor("subtleColor")}")`,
);
registerTemplateFunc(
  "themeError",
  () => `lipgloss.Color("${getThemeColor("errorColor")}")`,
);
registerTemplateFunc(
  "themeSuccess",
  () => `lipgloss.Color("${getThemeColor("successColor")}")`,
);

function getDefaultDependencies(): Record<string, string> {
  const defaultDependencies: Record<string, string> = {
    "github.com/spyzhov/ajson": deps["github.com/spyzhov/ajson"].version,
    "github.com/spf13/cobra": deps["github.com/spf13/cobra"].version,
    "gopkg.in/yaml.v3": deps["gopkg.in/yaml.v3"].version,
    "golang.org/x/term": deps["golang.org/x/term"].version,
  };

  if (includeDecimal()) {
    defaultDependencies["github.com/ericlagergren/decimal"] =
      deps["github.com/ericlagergren/decimal"].version;
  }

  if (
    context.Global.AST.MainSDK.OutputTests ||
    sdkHasTests(context.Global.AST)
  ) {
    defaultDependencies["github.com/stretchr/testify"] =
      deps["github.com/stretchr/testify"].version;
  }

  if (hasClientCredentials() || hasOAuth2PasswordFlow()) {
    defaultDependencies["golang.org/x/sync"] =
      deps["golang.org/x/sync"].version;
  }

  // gojq is always included for --jq flag support
  defaultDependencies["github.com/itchyny/gojq"] =
    deps["github.com/itchyny/gojq"].version;

  // gotoon is always included for --output-format=toon support
  defaultDependencies["github.com/alpkeskin/gotoon"] =
    deps["github.com/alpkeskin/gotoon"].version;

  if (hasGlobalSecurity()) {
    defaultDependencies["github.com/zalando/go-keyring"] =
      deps["github.com/zalando/go-keyring"].version;
  }

  // Interactive mode dependencies (Charm ecosystem)
  if (isInteractiveAuthEnabled() || isInteractiveModeEnabled()) {
    defaultDependencies["github.com/charmbracelet/huh"] =
      deps["github.com/charmbracelet/huh"].version;
    defaultDependencies["github.com/charmbracelet/lipgloss"] =
      deps["github.com/charmbracelet/lipgloss"].version;
  }
  if (isInteractiveModeEnabled()) {
    defaultDependencies["github.com/charmbracelet/bubbletea"] =
      deps["github.com/charmbracelet/bubbletea"].version;
    defaultDependencies["github.com/charmbracelet/bubbles"] =
      deps["github.com/charmbracelet/bubbles"].version;
  }

  return defaultDependencies;
}

// @ts-ignore
function templateDependencies(): string {
  return renderDependencies(
    getDefaultDependencies(),
    context.Global.Config.AdditionalDependencies,
  );
}
registerTemplateFunc("templateDependencies", templateDependencies);

// @ts-ignore
function renderDependencies(
  defaultDependencies: Record<string, string>,
  additionalDependencies: Record<string, string>,
): string {
  const dependencies = {
    ...defaultDependencies,
    ...additionalDependencies,
  };

  let renderedDependencies = "";

  for (const dependency of Object.keys(dependencies).sort()) {
    renderedDependencies += `${dependency} ${dependencies[dependency]}\n`;
  }
  renderedDependencies = renderedDependencies.replace(/\n$/, "");

  return renderedDependencies;
}
