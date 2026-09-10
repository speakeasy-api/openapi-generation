interface CLIErrorReasonTableEntry {
  Reason: string;
  Type: string;
  HasHints: boolean;
  Hints: string[];
}

interface CLIErrorTypeTableEntry {
  Type: string;
  Hints: string[];
}

interface CLIErrorTableContext {
  Reasons: CLIErrorReasonTableEntry[];
  Types: CLIErrorTypeTableEntry[];
}

function cliErrorMapRules(
  rules: CLIErrorRule[] | undefined,
): CLIErrorReasonTableEntry[] {
  return (rules || []).map((rule) => ({
    Reason: rule.Reason || "",
    Type: rule.Type || "",
    HasHints: rule.HasHints || false,
    Hints: rule.Hints || [],
  }));
}

function cliErrorTable(): CLIErrorTableContext {
  const manifest = context.Global.AST.CLIErrors;
  return {
    Reasons: cliErrorMapRules(manifest?.Reasons),
    Types: (manifest?.Types || []).map((rule) => ({
      Type: rule.Type || "",
      Hints: rule.Hints || [],
    })),
  };
}
registerTemplateFunc("cliErrorTable", cliErrorTable);

interface CLIErrorCarrierSegment {
  Field: string;
  IsWild: boolean;
}

interface CLIErrorReasonCarrier {
  Declared: boolean; // pointer explicitly declared in x-speakeasy-cli-errors
  Display: string; // JSONPath spelling shown in docs and generated comments
  Segments: CLIErrorCarrierSegment[];
}

interface CLIErrorProbeInfo extends CLIErrorReasonCarrier {
  Rules: CLIErrorReasonTableEntry[];
}

function cliErrorMapSegments(
  segments: CLIErrorPointerSegment[] | undefined,
): CLIErrorCarrierSegment[] {
  return (segments || []).map((segment) => ({
    Field: segment.Field || "",
    IsWild: segment.IsWild || false,
  }));
}

// cliErrorReasonCarrier resolves the primary reason carrier: the declared
// x-speakeasy-cli-errors reasonPointer when present, otherwise the default
// top-level $.reason member of the error object. Declaring the pointer also
// opts the carrier into promoting unmatched reason codes verbatim.
function cliErrorReasonCarrier(): CLIErrorReasonCarrier {
  const manifest = context.Global.AST.CLIErrors;
  if (manifest?.ReasonPointer) {
    return {
      Declared: true,
      Display: manifest.ReasonPointer,
      Segments: cliErrorMapSegments(manifest.ReasonSegments),
    };
  }
  return {
    Declared: false,
    Display: "$.reason",
    Segments: [{ Field: "reason", IsWild: false }],
  };
}
registerTemplateFunc("cliErrorReasonCarrier", cliErrorReasonCarrier);

// cliErrorProbes lists every reason carrier compiled into this build, in
// classification order: the primary carrier first (when it has rules to match
// or is explicitly declared), then each declared probe. Runtime-owned CLI_*
// rules are excluded — they classify local failures, not body carriers.
function cliErrorProbes(): CLIErrorProbeInfo[] {
  const manifest = context.Global.AST.CLIErrors;
  const probes: CLIErrorProbeInfo[] = [];
  const primary = cliErrorReasonCarrier();
  const primaryRules = cliErrorTable().Reasons.filter(
    (rule) => !rule.Reason.startsWith("CLI_"),
  );
  if (primary.Declared || primaryRules.length > 0) {
    probes.push({ ...primary, Rules: primaryRules });
  }
  for (const probe of manifest?.Probes || []) {
    probes.push({
      Declared: true,
      Display: probe.Pointer || "",
      Segments: cliErrorMapSegments(probe.Segments),
      Rules: cliErrorMapRules(probe.Reasons),
    });
  }
  return probes;
}
registerTemplateFunc("cliErrorProbes", cliErrorProbes);

// cliErrorCarrierDisplayList renders the carrier pointers for prose docs:
// "`$.reason`" or "`$.details[*].reason`, then `$.status`". Empty when the
// build compiles no carriers.
function cliErrorCarrierDisplayList(): string {
  return cliErrorProbes()
    .map((probe) => "`" + probe.Display + "`")
    .join(", then ");
}
registerTemplateFunc("cliErrorCarrierDisplayList", cliErrorCarrierDisplayList);

// cliErrorAnyCarrierDeclared reports whether any carrier pointer is
// explicitly declared — the opt-in for surfacing unmatched reason codes
// verbatim as error_reason.
function cliErrorAnyCarrierDeclared(): boolean {
  return cliErrorProbes().some((probe) => probe.Declared);
}
registerTemplateFunc("cliErrorAnyCarrierDeclared", cliErrorAnyCarrierDeclared);

// cliErrorUnwrapDeclared reports whether the manifest opts into normalizing
// the single-element array-wrapped error body shape ([{"error": {...}}]).
function cliErrorUnwrapDeclared(): boolean {
  return context.Global.AST.CLIErrors?.UnwrapErrorArray === true;
}
registerTemplateFunc("cliErrorUnwrapDeclared", cliErrorUnwrapDeclared);

function cliErrorTestBodyForSegments(
  segments: CLIErrorCarrierSegment[],
  reason: string,
): string {
  let value: unknown = reason;
  for (const segment of [...segments].reverse()) {
    value = segment.IsWild ? [value] : { [segment.Field]: value };
  }
  return JSON.stringify({ error: value });
}

// cliErrorCarrierTestBody builds a JSON error body carrying the given reason
// code at this build's primary reason carrier, for generated harness tests.
// The carrier is resolved against the nested error object, and its first
// segment is always a named member (the decoder rejects a leading fan-out).
function cliErrorCarrierTestBody(reason: string): string {
  return cliErrorTestBodyForSegments(cliErrorReasonCarrier().Segments, reason);
}
registerTemplateFunc("cliErrorCarrierTestBody", cliErrorCarrierTestBody);

interface CLIErrorProbeRuleContext {
  Display: string;
  Body: string;
  Reason: string;
  Type: string;
}

// cliErrorFirstProbeTypedRule returns the first additional probe's first
// typed rule together with a body carrying that reason at the probe's
// pointer, or null when the document declares no typed probe rules beyond
// the primary carrier. Generated harness tests use it to prove declared
// probes classify.
function cliErrorFirstProbeTypedRule(): CLIErrorProbeRuleContext | null {
  const manifest = context.Global.AST.CLIErrors;
  const probes = cliErrorProbes();
  const skipPrimary = probes.length > (manifest?.Probes || []).length ? 1 : 0;
  for (const probe of probes.slice(skipPrimary)) {
    const rule = probe.Rules.find((candidate) => candidate.Type !== "");
    if (rule) {
      return {
        Display: probe.Display,
        Body: cliErrorTestBodyForSegments(probe.Segments, rule.Reason),
        Reason: rule.Reason,
        Type: rule.Type,
      };
    }
  }
  return null;
}
registerTemplateFunc(
  "cliErrorFirstProbeTypedRule",
  cliErrorFirstProbeTypedRule,
);

// cliErrorRuntimeRules returns the declared hint rules for runtime-owned
// CLI_* reasons; they classify local failures rather than body carriers.
function cliErrorRuntimeRules(): CLIErrorReasonTableEntry[] {
  return cliErrorTable().Reasons.filter((rule) =>
    rule.Reason.startsWith("CLI_"),
  );
}
registerTemplateFunc("cliErrorRuntimeRules", cliErrorRuntimeRules);

// cliErrorFirstTypedRule returns the first declared server-reason rule on the
// primary carrier that maps to an error type (the rule generated harness
// tests exercise), or null when the document declares none.
function cliErrorFirstTypedRule(): CLIErrorReasonTableEntry | null {
  return (
    cliErrorTable().Reasons.find(
      (rule) => rule.Type !== "" && !rule.Reason.startsWith("CLI_"),
    ) || null
  );
}
registerTemplateFunc("cliErrorFirstTypedRule", cliErrorFirstTypedRule);

// cliErrorTypeDeclared reports whether the x-speakeasy-cli-errors manifest
// maps any reason (on any carrier) to the given error type or declares hints
// for it.
function cliErrorTypeDeclared(errorType: string): boolean {
  return (
    cliErrorProbes().some((probe) =>
      probe.Rules.some((rule) => rule.Type === errorType),
    ) || cliErrorTable().Types.some((rule) => rule.Type === errorType)
  );
}
registerTemplateFunc("cliErrorTypeDeclared", cliErrorTypeDeclared);

// The account-state error types are part of this build's surface only when
// the document's declared error rules reference them: a document declaring
// only one account-state type emits (and documents) only that one.
function cliEmitsServiceDisabledErrorType(): boolean {
  return cliErrorTypeDeclared("service_disabled");
}
registerTemplateFunc(
  "cliEmitsServiceDisabledErrorType",
  cliEmitsServiceDisabledErrorType,
);

function cliEmitsBillingDisabledErrorType(): boolean {
  return cliErrorTypeDeclared("billing_disabled");
}
registerTemplateFunc(
  "cliEmitsBillingDisabledErrorType",
  cliEmitsBillingDisabledErrorType,
);
