//@ts-ignore
const supportedFeatures = {
  core: "6.1.2",
  getRequestBodies: "3.0.0",
  flattening: "3.1.1",
  flatRequests: "1.0.1",
  globalSecurity: "3.0.7",
  methodSecurity: "3.0.3",
  globalServerURLs: "3.2.1",
  methodServerURLs: "3.1.2",
  globals: "3.0.0",
  enums: "3.2.0",
  serverIDs: "3.0.0",
  nameOverrides: "3.0.3",
  includes: "3.0.0",
  docs: "1.5.3",
  examples: "3.0.4",
  groups: "3.0.1",
  deprecations: "3.0.2",
  retries: "3.1.2",
  pagination: "3.0.10",
  inputOutputModels: "3.0.0",
  ignores: "3.0.1",
  typeOverrides: "3.0.0",
  errors: "3.4.5",
  unions: "3.1.9",
  sliceUnions: "0.0.1",
  multiLevelTagging: "3.0.0",
  bigint: "1.0.0",
  decimal: "1.0.0",
  uuid: "0.1.1",
  duration: "0.1.1",
  devContainers: "3.0.0",
  downloadStreams: "1.0.1",
  constsAndDefaults: "1.0.7",
  additionalProperties: "1.0.2",
  serverEvents: "1.1.1",
  tests: "1.19.10",
  serverEventsSentinels: "0.1.0",
  webhooks: "2.0.0",
  callbacks: "2.0.0",
  oauth2ClientCredentials: "2.1.6",
  responseFormat: "1.1.0",
  hiddenGlobals: "1.0.0",
  errorUnions: "1.0.2",
  acceptHeaders: "3.0.0",
  stringNumberFormats: "1.0.0",
  methodArguments: "1.1.1",
  sdkHooks: "1.3.0",
  additionalDependencies: "1.1.0",
  nullables: "1.0.2",
  globalSecurityFlattening: "1.0.0",
  globalSecurityCallbacks: "1.0.0",
  openEnums: "1.0.4",
  openDiscriminatedUnions: "0.1.0",
  uploadStreams: "1.0.3",
  multipartFileContentType: "1.0.0",
  urlBasedPagination: "1.0.0",
  operationTimeout: "0.3.1",
  defaultEnabledRetries: "0.2.0",
  envVarSecurityUsage: "0.3.3",
  envVarGlobals: "0.3.0",
  deepObjectParams: "0.1.0",
  customSecuritySchemes: "0.1.0",
  mockServer: "0.1.4",
  enumUnions: "0.1.1",
  customCodeRegions: "0.1.2",
  jsonlResponses: "0.2.2",
  configurableModuleName: "0.2.0",
  modelNamespaces: "0.1.3",
  publicExports: "0.1.2",
};

//@ts-ignore
const implementedReadmeSections = [
  "installation",
  "usage",
  "operations",
  "dev-containers",
  "global-parameters",
  "errors",
  "file-uploads",
  "eventstream",
  "http-client",
  "pagination",
  "retries",
  "security",
  "server",
  "debug",
  "jsonl",
];

//@ts-ignore
function isFeatureSupported(feature: string): boolean {
  return feature in supportedFeatures;
}

//@ts-ignore
function isFeatureEnabled(
  feature: string,
  config: Record<string, any>,
): boolean {
  switch (feature) {
    case "envVarSecurityUsage":
      return config["EnvVarPrefix"] !== undefined;
    case "uuid": {
      const uuidFormat = config["UUIDFormat"] ?? config["UuidFormat"];
      return uuidFormat === true || uuidFormat === "true";
    }
    case "duration":
      return (
        config["DurationFormat"] === true || config["DurationFormat"] === "true"
      );
    default:
      return false;
  }
}

//@ts-ignore
function allSupportedFeatures(): string[] {
  const allFeatures = Object.keys(supportedFeatures);
  return allFeatures.filter((feature) => !isFeatureIgnored(feature));
}

//@ts-ignore
function isFeatureIgnored(feature: string): boolean {
  return [
    "disallowCircularRefs",
    "tagBasedOrdering",
    "sets",
    "reactQueryHooks",
  ].includes(feature);
}

//@ts-ignore
function isReadmeSectionImplemented(id: string): boolean {
  return implementedReadmeSections.includes(id);
}

//@ts-ignore
function isReadmeSectionIgnored(id: string): boolean {
  return false;
}

//@ts-ignore
function getStaticallyUsedFeatures(): string[] {
  return ["sdkHooks", "additionalDependencies", "globalSecurityCallbacks"];
}

//@ts-ignore
function getFeatureVersion(feature: string): string {
  return supportedFeatures[feature];
}

//@ts-ignore
function isTestSkipped(test: string): boolean {
  return [
    "event-stream-timeout-still-bounds-slow-stream",
    "event-stream-with-timeout-streams-to-completion",
    "smart-union-nullable-collection-item",
    "smart-union-wrapped-complex-fields",
    "servers-select-global-server-by-name-valid-using-builder",
    "request-bodies-file-upload-extract-content-type",
    "retries-retry-connection-errors-reject",
    "retries-retry-connection-errors-reset",
    "pagination-body-wrapped-request",
    "pagination-body-flattened-with-security",
    "pagination-body-flattened-optional-security",
    "pagination-cursor-non-numeric-with-limit",
    "pagination-ambiguous-input",
    "pagination-encapsulated-parameter",
    "hooks-access-retry-config",
    "hooks-terminate-retry-loop",
    "hooks-trigger-retries",
    "webhooks-consume",
    "webhooks-consume-custom-security",
    "webhooks-consume-bad-data",
    "webhooks-consume-bad-signature",
    "hooks-oauth2-password-bad-credentials",
    "hooks-oauth2-password-operation-scope",
    "hooks-oauth2-password-operation-security-option",
    "hooks-oauth2-password-token-renewal",
    "hooks-oauth2-password-with-credentials",
    "hooks-oauth2-password-with-token",
    "hooks-oauth2-password-not-required",
    "parameters-path-encoding",
    "parameters-query-encoding",
    "pagination-cursor-response-envelope",
    "errors-additional-properties",
    "errors-optional-nullable-error-message",
    "errors-optional-nullable-error-message-nested",
    "unions-optional-union-map",
    "unions-mixed-union-types",
    "union-enum-nested-in-array-in-union",
    "servers-select-global-server-by-id-default-using-builder",
    "servers-select-global-server-by-name-broken-using-builder",
    "servers-select-global-server-by-name-default-using-builder",
    "servers-select-global-server-by-name-with-templates-broken-using-builder",
    "servers-select-global-server-by-name-with-templates-defaults-using-builder",
    "servers-select-global-server-by-name-with-templates-valid-using-builder",
    "servers-select-global-server-by-id-valid-using-builder",
    "servers-select-global-server-by-id-invalid-using-builder",
    "servers-select-global-server-by-id-broken-using-builder",
    "custom-code-region-model-method-with-imports",
    "parameters-open-enum",
    "parameters-form-query-params-unions",
    "parameters-ordering-body-first",
    "parameters-ordering-parameters-first",
    "parameters-ordering-with-legacy-flattening-order",
    "parameters-ordering-using-optional-request-body-body-first",
    "parameters-ordering-using-optional-request-body-parameters-first",
    "parameters-ordering-using-optional-request-body-with-legacy-flattening-order",
    "pagination-cursor-params-snake",
    "pagination-cursor-snake",
    "pagination-limit-offset-params-snake",
    "pagination-limit-offset-snake",
    "pagination-simple-object-snake",
    "pagination-limit-offset-page-params-const",
    "pagination-params-wrapped-request",
    "cancellation-token-cancelled-before-request",
    "cancellation-token-cancelled-during-request",
    "cancellation-token-cancelled-inside-hook",
    "cancellation-token-no-cancellation",
    "customclient-request-parameters-retained-legacy",
    "unions-strongly-typed-nullable-one-of-post",
    "unions-one-of-boolean-and-string-enum-with-response-boolean",
    "unions-one-of-boolean-and-string-enum-with-response-enum",
    "object-with-optional-false-nullable-false-field-absent",
    "object-with-optional-false-nullable-false-field-null",
    "object-with-optional-false-nullable-false-full",
    "object-with-optional-false-nullable-true-field-null",
    "object-with-optional-false-nullable-true-full",
    "object-with-optional-nullable-field-absent",
    "object-with-optional-nullable-field-null",
    "object-with-optional-nullable-full",
    "object-with-optional-true-nullable-false-field-absent",
    "object-with-optional-true-nullable-false-field-null",
    "object-with-optional-true-nullable-false-full",
    "object-with-optional-false-nullable-true-field-absent",
    "transform-jq-request-response",
    "polling-delay-seconds-override",
    "polling-delay-seconds",
    "polling-failure-criteria-response-body-nested",
    "polling-failure-criteria-response-body-optional",
    "polling-failure-criteria-response-body-regex",
    "polling-failure-criteria-response-body",
    "polling-failure-criteria-status-code-regex",
    "polling-failure-criteria-status-code",
    "polling-interval-seconds-override",
    "polling-interval-seconds",
    "polling-limit-count-override",
    "polling-limit-count",
    "polling-success-criteria-response-body-nested",
    "polling-success-criteria-response-body-optional",
    "polling-success-criteria-response-body-regex",
    "polling-success-criteria-response-body",
    "polling-success-criteria-status-code-error-known",
    "polling-success-criteria-status-code-error-regex",
    "polling-success-criteria-status-code-error-unknown",
    "polling-success-criteria-status-code-regex",
    "polling-success-criteria-status-code",
    "request-bodies-form-encoded-string-array",
    "union-enum-array-deserialization",
    "event-stream-with-abort-signal",
    "custom-client-http-proxy",
    "open-union-smart-union-interop",
    "unions-discriminated-multiple-memberships",
    "flattening-nullable-body-with-required-param-ordering",
    "retries-timeout-fresh-signal",
    "event-stream-chat-json",
    "unions-typed-object-one-of-post-body-less-throws",
    "unions-strongly-typed-one-of-discriminated-post",
    "request-extras-extra-body-merged",
    "request-extras-extra-body-overrides",
    "request-extras-extra-body-rejects-non-json",
    "request-extras-extra-query-appended",
    "request-extras-extra-query-merged",
    "request-extras-extra-query-nullish-arrays",
    "request-extras-extra-query-object-json",
    "request-extras-extra-query-overrides",
    "request-extras-extra-query-preserves-server-url-query",
    "request-extras-extra-query-security-overrides",
    "pagination-cursor-deep-nested-outputs",
    "pagination-cursor-deep-nested-outputs-iterator",
    "react-query-builders-key-distinct-bodies",
    "react-query-key-includes-parameters-and-request-body",
    "react-query-key-includes-request-body",
    "react-query-key-opt-in-infers-query-type",
    "smart-union-nested-union-vs-nested-union",
    "request-bodies-put-multipart-optional-nullable",
    "request-bodies-post-form-optional-nullable",
    "request-bodies-post-json-optional-nullable",
    "smart-union-nullable-union-nullable-fields",
    "request-bodies-post-form-optional-nullable-json-shared",
    "request-bodies-put-multipart-optional-nullable-json-shared",
    "request-bodies-post-json-optional-nullable-body-and-param",
    "pagination-limit-offset-page-body-optional-nullable",
    "pagination-cursor-nullable-results",
    "retries-binary-request-body",
    "retries-header-http-date-beyond-max-interval",
    "retries-header-rate-limit-reset",
    "errors-error-body-validation-lenient",
    "agent-mode-does-not-leak-between-invocations",
    "catalog-lists-values",
    "catalog-rejects-positional-argument",
    "catalog-structured-output",
    "cli-exit-code-bare-intent",
    "cli-exit-code-connection-and-agent-block",
    "cli-exit-code-mapping-tables",
    "cli-exit-code-reason-first",
    "cli-exit-code-root-unknown-message",
    "cli-exit-code-single-envelope",
    "cli-exit-code-success-and-discovery",
    "cli-exit-codes-rendering-modes",
    "client-credentials-dry-run-token-and-api",
    "clierrors-agent-mode-explicitly-disabled",
    "clierrors-agent-mode-from-environment",
    "clierrors-agent-mode-matrix",
    "clierrors-api-error-not-double-wrapped",
    "clierrors-declared-hints-fire",
    "clierrors-flag-values-not-rendering-flags",
    "clierrors-invalid-values-do-not-stop-preparse",
    "clierrors-json-output-envelope",
    "clierrors-plain-mode-classified",
    "clierrors-typed-reasons",
    "error-taxonomy-declared-rules-and-ordering",
    "error-taxonomy-every-closed-type-has-a-hint",
    "error-taxonomy-fallback-message-omits-raw-body",
    "error-taxonomy-hint-precedence-and-dedupe",
    "error-taxonomy-jq-uses-structured-envelope",
    "error-taxonomy-odd-bodies-use-fallback-envelope",
    "error-taxonomy-odd-reflection-fields-do-not-panic",
    "error-taxonomy-origin-semantics",
    "error-taxonomy-pretty-http-status-without-reason",
    "error-taxonomy-pretty-json-agent-parity",
    "error-taxonomy-pretty-residual-details",
    "error-taxonomy-reserved-keys-and-rendered-classification",
    "event-stream-connection-refused-stays-connection-error",
    "event-stream-decode-failure-is-protocol-error",
    "event-stream-decode-failure-without-json-error-type",
    "event-stream-error-event-agent-mode",
    "event-stream-error-event-agent-mode-fallback",
    "event-stream-error-event-exits-non-zero",
    "event-stream-error-event-pretty",
    "event-stream-numbers-keep-precision",
    "event-stream-server-error-mentioning-decode-is-not-protocol-error",
    "event-stream-timeout-still-bounds-slow-stream",
    "event-stream-undiscriminated-error-shaped-variant-is-data",
    "event-stream-union-events-render-wire-shape",
    "intent-dispatch-resolves-stdin-once",
    "intent-dispatch-select-matrix",
    "intent-interactive-piped-body-skips-prompts",
    "intent-route-dispatch-agent-envelope",
    "intent-route-dispatch-bare-help",
    "intent-route-dispatch-body-evidence",
    "intent-route-dispatch-conflicting-evidence",
    "intent-route-dispatch-default-required",
    "intent-route-dispatch-flag-paths",
    "intent-route-dispatch-help-usage",
    "intent-route-dispatch-interactive-plan",
    "intent-route-dispatch-override-once",
    "intents-both-body-surfaces-merged",
    "intents-discriminated-partial-body",
    "intents-foreign-selector-agent-envelope",
    "intents-foreign-selector-usage-error",
    "intents-non-object-body-falls-through",
    "intents-partial-body-keeps-presets",
    "intents-partial-body-user-keys-win",
    "intents-preset-from-args",
    "intents-preset-merge-matrix",
    "intents-undeclared-nested-group-order",
    "pre-request-connection-classifier-typed-table",
    "pre-request-discriminated-root-union-messages",
    "pre-request-dry-run-unresolvable-host",
    "pre-request-generated-required-union-paths-and-optionality",
    "pre-request-generated-required-union-positive-echoes",
    "pre-request-real-connection-failures",
    "pre-request-silent-stdin-times-out-on-root-union",
    "pre-request-validation-envelope-matrix",
    "request-bodies-complex-number-types-bigint-overflow",
    "request-shape-body-agent-variant",
    "request-shape-body-and-whole-body-flag-rejected",
    "request-shape-body-both-selectors-rejected",
    "request-shape-body-default-selector-merges",
    "request-shape-body-null-optional-key-allowed",
    "request-shape-body-typo-key-rejected",
    "request-shape-expanded-body-typo-key-lenient-warning",
    "request-shape-expanded-body-typo-key-rejected",
    "request-shape-intent-bare-usage",
    "request-shape-intent-empty-stdin-usage",
    "request-shape-intent-flag-and-body-conflict",
    "request-shape-intent-flag-merged-into-body",
    "request-shape-intent-positional-and-body-conflict",
    "request-shape-intent-positional-merged-into-body",
    "request-shape-intent-positional-merged-into-whole-body-flag",
    "request-shape-intent-stdin-runs-request",
    "request-shape-intent-stdin-typo-key-rejected",
    "request-shape-intent-stdin-with-positional",
    "request-shape-intent-synthesized-body",
    "request-shape-intent-zero-value-flag-merged",
    "request-shape-nested-union-agent-selector-no-default",
    "request-shape-nested-union-body-default-selector-merges",
    "request-shape-nested-union-own-flag-default-selector-merges",
    "request-shape-nested-union-stdin-default-selector-merges",
    "request-shape-open-schema-keeps-unknown-keys",
    "request-shape-stdin-typo-key-rejected",
    "request-shape-whole-body-flag-typo-key-rejected",
    "request-shape-whole-body-flag-wins-over-stdin",
    "security-ranking-both-flags",
    "security-ranking-env-only",
    "security-ranking-env-tie",
    "security-ranking-flag-beats-env",
    "security-ranking-operation-restriction",
    "security-ranking-pick-matrix",
    "security-ranking-source-rank",
    "event-stream-with-operation-timeout-streams-to-completion",
    "jsonl-stream-timeout-still-bounds-slow-stream",
    "jsonl-stream-with-timeout-streams-to-completion",
  ].includes(test);
}

//@ts-ignore
function getAutoGeneratedHeader(
  generatedId: string,
  config: Record<string, any>,
): string {
  let header = `"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""\n`;
  if (generatedId) {
    header += `# @generated-id: ${generatedId}\n`;
  }
  if (config.GeneratedLicense === "agpl") {
    header +=
      "# Generated under the AGPL-3.0-only license.\n" +
      "# SPDX-License-Identifier: AGPL-3.0-only\n";
  }
  header += "\n";
  return header;
}

// @ts-ignore
function collectGlobalsAndServersModels(): boolean {
  return true;
}
