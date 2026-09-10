type GoOptionalMethodArgumentsMode =
  | "pointers"
  | "shared-options"
  | "method-options";

type GoMethodArgumentOption = {
  field: FieldDef;
  name: string;
  storageKey: string;
  supportedOptionName: string;
  location: string;
};

type GoOperationOptionMetadata = {
  operation: Operation;
  mode: GoOptionalMethodArgumentsMode;
  prefix: string;
  optionType: string;
  optionsType: string;
  requiredArguments: FieldDef[];
  optionalArguments: GoMethodArgumentOption[];
  commonControls: string[];
};

const goOperationOptionMetadata = new Map<string, GoOperationOptionMetadata>();

function getGoOperationMetadataKey(operation: Operation): string {
  return [
    operation.OwningSDK?.Group ?? "",
    operation.OwningSDK?.Type?.Name ?? "",
    operation.GetID(),
  ].join(":");
}

function resolveGoOptionalMethodArgumentsMode(
  operation: Operation,
): GoOptionalMethodArgumentsMode {
  const configured = context.Global.Config.OptionalMethodArguments;
  let globalMode: GoOptionalMethodArgumentsMode;
  switch (configured) {
    case undefined:
    case null:
    case "":
    case "pointers":
      globalMode = "pointers";
      break;
    case "shared-options":
    case "method-options":
      globalMode = configured;
      break;
    default:
      throw new Error(
        `invalid go.optionalMethodArguments value ${JSON.stringify(
          configured,
        )}; expected pointers, shared-options, or method-options`,
      );
  }
  const operationMode = operation.Extensions.GoOptionalMethodArguments;
  if (!operationMode) {
    return globalMode;
  }

  switch (operationMode.toString()) {
    case "pointers":
      return "pointers";
    case "shared-options":
      return "shared-options";
    case "method-options":
      return "method-options";
    default:
      throw new Error(
        `invalid x-speakeasy-go-optional-method-arguments value ${operationMode.toString()} for operation ${
          operation.ID
        }`,
      );
  }
}

function getGoArgumentLocation(operation: Operation, field: FieldDef): string {
  if (operation.Security === field || field.Annotations?.Has("opSecurity")) {
    return "Security";
  }

  const parameterType = getParameterType(field);
  switch (parameterType) {
    case "queryParam":
      return "Query";
    case "pathParam":
      return "Path";
    case "header":
      return "Header";
    case "cookie":
      return "Cookie";
  }

  if (field.Annotations?.Has("request")) {
    return "Body";
  }

  return "Request";
}

function getGoCommonControlSuffixes(operation: Operation): string[] {
  const controls = ["ServerURL", "TemplatedServerURL", "Timeout", "SetHeaders"];
  if (operation.Extensions.Retries) controls.push("Retries");
  if (operation.Extensions.Polling) controls.push("Polling");
  if (operation.GetAcceptTypes().length > 1) {
    controls.push("AcceptHeaderOverride");
  }
  if (hasPaginationURL(operation)) controls.push("URLOverride");
  if (isSkipDeserializationEnabled()) controls.push("SkipDeserialization");
  return controls;
}

function getGoOwningSDKPrefix(operation: Operation): string {
  const group = operation.OwningSDK?.Group ?? "";
  if (group) {
    return group
      .split(".")
      .filter(Boolean)
      .map((part) => sanitizeClassName(part))
      .join("");
  }

  return sanitizeClassName(operation.OwningSDK?.Type?.Name ?? "SDK");
}

function initializeGoOperationOptionMetadata(): void {
  goOperationOptionMetadata.clear();
  const operations = allOperations(context.Global.AST.MainSDK, {
    includeWebhooks: true,
    dedupe: false,
  });
  const basePrefixes = new Map<string, Operation[]>();

  for (const operation of operations) {
    const prefix = sanitizeMethodName(operation);
    const matches = basePrefixes.get(prefix) ?? [];
    matches.push(operation);
    basePrefixes.set(prefix, matches);
  }

  const usedPrefixes = new Map<string, Operation>();
  for (const operation of operations) {
    const basePrefix = sanitizeMethodName(operation);
    const collisions = basePrefixes.get(basePrefix) ?? [];
    const prefix =
      collisions.length > 1
        ? `${getGoOwningSDKPrefix(operation)}${basePrefix}`
        : basePrefix;
    const previous = usedPrefixes.get(prefix);
    if (previous && previous !== operation) {
      throw new Error(
        `unable to generate unique Go method option prefix ${prefix} for operations ${previous.ID} and ${operation.ID}; use x-speakeasy-name-override to disambiguate the generated method names`,
      );
    }
    usedPrefixes.set(prefix, operation);

    const mode = resolveGoOptionalMethodArgumentsMode(operation);
    const args = getArguments(operation);
    const requiredArguments =
      mode === "pointers"
        ? [...args.merged]
        : args.merged.filter((field) => !field.Optional);
    const optionalFields =
      mode === "pointers" ? [] : args.merged.filter((field) => field.Optional);
    if (mode !== "pointers" && !context.Global.Config.NullableOptionalWrapper) {
      const nullableField = optionalFields.find((field) => field.Nullable);
      if (nullableField) {
        throw new Error(
          `operation ${operation.ID} optional method argument ${nullableField.Name} is nullable; enable go.nullableOptionalWrapper or override this operation with x-speakeasy-go-optional-method-arguments: pointers`,
        );
      }
    }

    const commonControls = getGoCommonControlSuffixes(operation);
    const usedOptionNames = new Set<string>();
    const optionalArguments = optionalFields.map((field) => {
      const fieldSuffix = sanitizeClassName(field.Name);
      let suffix = fieldSuffix;
      if (commonControls.includes(fieldSuffix)) {
        suffix = `${getGoArgumentLocation(operation, field)}${fieldSuffix}`;
      }
      const name = `With${prefix}${suffix}`;
      if (usedOptionNames.has(name)) {
        throw new Error(
          `unable to generate unique Go method option ${name} for operation ${operation.ID}; use x-speakeasy-name-override to disambiguate the optional method arguments`,
        );
      }
      usedOptionNames.add(name);
      return {
        field,
        name,
        storageKey: `${prefix}:${sanitizePrivateFieldName(field.Name)}`,
        supportedOptionName: `SupportedOption${prefix}${suffix}`,
        location: getGoArgumentLocation(operation, field),
      };
    });

    goOperationOptionMetadata.set(getGoOperationMetadataKey(operation), {
      operation,
      mode,
      prefix,
      optionType: `${prefix}Option`,
      optionsType: `${prefix}Options`,
      requiredArguments,
      optionalArguments,
      commonControls,
    });
  }
}

function ensureGoOperationOptionMetadata(): void {
  if (goOperationOptionMetadata.size === 0) {
    initializeGoOperationOptionMetadata();
  }
}

function getGoOperationOptionMetadataList(): GoOperationOptionMetadata[] {
  ensureGoOperationOptionMetadata();
  return Array.from(goOperationOptionMetadata.values());
}

function hasGoOptionalMethodArgumentOptions(): boolean {
  return getGoOperationOptionMetadataList().some(
    (metadata) => metadata.optionalArguments.length > 0,
  );
}

function hasGoMethodSpecificOptions(): boolean {
  return getGoOperationOptionMetadataList().some(
    (metadata) => metadata.mode === "method-options",
  );
}

function getGoMethodArgumentOption(
  operation: Operation,
  field: FieldDef,
): GoMethodArgumentOption | undefined {
  return getGoOperationOptionMetadata(operation).optionalArguments.find(
    (option) => option.field.Name === field.Name,
  );
}

function getGoMethodOptionType(operation: Operation): string {
  const metadata = getGoOperationOptionMetadata(operation);
  return metadata.mode === "method-options" ? metadata.optionType : "Option";
}

function getGoCommonControlOptionName(
  operation: Operation,
  control: string,
): string {
  const metadata = getGoOperationOptionMetadata(operation);
  if (metadata.mode === "method-options") {
    return `With${metadata.prefix}${control}`;
  }

  const optionNames = getOptionNames();
  const sharedControlNames: Record<string, string> = {
    ServerURL: optionNames.ServerURL,
    TemplatedServerURL: optionNames.TemplatedServerURL,
    Timeout: optionNames.OperationTimeout,
    SetHeaders: optionNames.SetHeaders,
    Retries: optionNames.Retries,
    Polling: optionNames.Polling,
    AcceptHeaderOverride: optionNames.AcceptHeaderOverride,
    URLOverride: optionNames.URLOverride,
    SkipDeserialization: optionNames.SkipDeserialization,
  };
  const name = sharedControlNames[control];
  if (!name) {
    throw new Error(`missing shared Go option name for ${control}`);
  }
  return name;
}

function getGoMethodOptionConstructorNames(operation: Operation): string[] {
  const metadata = getGoOperationOptionMetadata(operation);
  const names = metadata.optionalArguments.map((argument) => argument.name);
  if (metadata.mode !== "pointers") {
    names.push(
      ...metadata.commonControls.map((control) =>
        getGoCommonControlOptionName(operation, control),
      ),
    );
  }
  return names;
}

function getGoMethodOptionsDescription(operation: Operation): string {
  const constructorNames = getGoMethodOptionConstructorNames(operation);
  if (constructorNames.length === 0) {
    return "The options for this request.";
  }

  return `The options for this request. Available options are ${constructorNames
    .map((name) => `\`${name}\``)
    .join(", ")}.`;
}

function getGoMethodAdditionalNotes(operation: Operation): string {
  const notes = [getHoistedSecurityRemark(operation)];
  const metadata = getGoOperationOptionMetadata(operation);
  if (metadata.optionalArguments.length > 0) {
    const namespace = getAccessNamespace("", "operations");
    notes.push(
      `Optional arguments may be provided with ${metadata.optionalArguments
        .map((argument) => `[${namespace}${argument.name}]`)
        .join(", ")}.`,
    );
  }
  return notes.filter(Boolean).join("\n\n");
}

function getGoMethodArgumentOptionDescription(
  option: GoMethodArgumentOption,
): string {
  return templateDocumentationComments(option.field.Comments)
    .replace(/\s+/g, " ")
    .trim();
}

function shouldDocumentGoMethodOptions(operation: Operation): boolean {
  const metadata = getGoOperationOptionMetadata(operation);
  return (
    metadata.mode !== "pointers" ||
    !!operation.Servers ||
    !!operation.Extensions.Retries ||
    !!operation.Extensions.Polling
  );
}

function getGoCommonOptionsExpression(operation: Operation): string {
  const metadata = getGoOperationOptionMetadata(operation);
  return metadata.mode === "method-options" ? "o.Options" : "o";
}

function getGoMethodArgumentInputType(option: GoMethodArgumentOption): string {
  if (option.field.Nullable) {
    addImport("optionalnullable");
    return `*${sanitizeType(option.field.Type, false, "")}`;
  }
  return sanitizeType(option.field.Type, false, "");
}

function getGoMethodArgumentStoredType(option: GoMethodArgumentOption): string {
  return sanitizeFieldType(option.field, "");
}

function templateGoMethodArgumentStoredValue(
  option: GoMethodArgumentOption,
  valueName: string,
): string {
  if (option.field.Nullable) {
    addImport("optionalnullable");
    return `optionalnullable.From(${valueName})`;
  }
  if (isPointerType(option.field.Type)) {
    return valueName;
  }
  return `&${valueName}`;
}

function templateGoPaginationOptionValue(
  option: GoMethodArgumentOption,
  valueName: string,
): string {
  return option.field.Nullable ? `&${valueName}` : valueName;
}

function templateGoMethodArgumentInputValue(
  option: GoMethodArgumentOption,
  valueName: string,
): string {
  if (option.field.Nullable || isPointerType(option.field.Type)) {
    return valueName;
  }
  return `*${valueName}`;
}

function getGoOperationOptionMetadata(
  operation: Operation,
): GoOperationOptionMetadata {
  ensureGoOperationOptionMetadata();
  const metadata = goOperationOptionMetadata.get(
    getGoOperationMetadataKey(operation),
  );
  if (!metadata) {
    throw new Error(
      `Go method option metadata was not initialized for operation ${operation.ID}`,
    );
  }
  return metadata;
}

registerTemplateFunc(
  "getGoOperationOptionMetadata",
  getGoOperationOptionMetadata,
);
registerTemplateFunc(
  "getGoOperationOptionMetadataList",
  getGoOperationOptionMetadataList,
);
registerTemplateFunc(
  "hasGoOptionalMethodArgumentOptions",
  hasGoOptionalMethodArgumentOptions,
);
registerTemplateFunc("hasGoMethodSpecificOptions", hasGoMethodSpecificOptions);
registerTemplateFunc("getGoMethodArgumentOption", getGoMethodArgumentOption);
registerTemplateFunc("getGoMethodOptionType", getGoMethodOptionType);
registerTemplateFunc(
  "getGoMethodOptionsDescription",
  getGoMethodOptionsDescription,
);
registerTemplateFunc("getGoMethodAdditionalNotes", getGoMethodAdditionalNotes);
registerTemplateFunc(
  "getGoMethodArgumentOptionDescription",
  getGoMethodArgumentOptionDescription,
);
registerTemplateFunc(
  "shouldDocumentGoMethodOptions",
  shouldDocumentGoMethodOptions,
);
registerTemplateFunc(
  "getGoCommonOptionsExpression",
  getGoCommonOptionsExpression,
);
registerTemplateFunc(
  "getGoMethodArgumentInputType",
  getGoMethodArgumentInputType,
);
registerTemplateFunc(
  "getGoMethodArgumentStoredType",
  getGoMethodArgumentStoredType,
);
registerTemplateFunc(
  "templateGoMethodArgumentStoredValue",
  templateGoMethodArgumentStoredValue,
);
registerTemplateFunc(
  "templateGoPaginationOptionValue",
  templateGoPaginationOptionValue,
);
registerTemplateFunc(
  "templateGoMethodArgumentInputValue",
  templateGoMethodArgumentInputValue,
);

type OptionDoc = {
  Name: string;
  Description: string;
  Example: string;
  ExtraDescription: string;
  ExtraExample: string;
};

// getOptionNames returns the resolved function names for all With*
// options.  Both options.go.stmpl / sdk.go.stmpl (which generate real Go
// code) and readme/options.stmpl (which generates docs) reference these
// via the getOptionNames() template function.
//
// Per-operation ServerURL and TemplatedServerURL include the "Method"
// prefix when there is no separate operations package.
function getOptionNames() {
  const prefix = context.Global.Config.Imports.GetOperationsPath()
    ? ""
    : "Method";
  return {
    // Global SDK options (sdk.go.stmpl)
    GlobalServerURL: "WithServerURL",
    GlobalTemplatedServerURL: "WithTemplatedServerURL",
    GlobalClient: "WithClient",
    GlobalSecurity: "WithSecurity",
    GlobalSecuritySource: "WithSecuritySource",
    GlobalRetryConfig: "WithRetryConfig",
    GlobalTimeout: "WithTimeout",
    GlobalServer: "WithServer",
    GlobalServerIndex: "WithServerIndex",
    // Per-operation options (options.go.stmpl)
    ServerURL: `With${prefix}ServerURL`,
    TemplatedServerURL: `With${prefix}TemplatedServerURL`,
    Retries: "WithRetries",
    OperationTimeout: "WithOperationTimeout",
    SetHeaders: "WithSetHeaders",
    URLOverride: "WithURLOverride",
    AcceptHeaderOverride: "WithAcceptHeaderOverride",
    Polling: "WithPolling",
    PollingDelaySecondsOverride: "WithDelaySecondsOverride",
    PollingIntervalSecondsOverride: "WithIntervalSecondsOverride",
    PollingLimitCountOverride: "WithLimitCountOverride",
    SkipDeserialization: "WithSkipDeserialization",
  } as const;
}

registerTemplateFunc("getOptionNames", getOptionNames);

// getGlobalAcceptTypes collects all accept types from the global AST.
// Unlike getAllAcceptTypes() this doesn't rely on context.Local so it
// works in any template context (including readme templates rendered
// with null context).
function getGlobalAcceptTypes(): string[] {
  const allAcceptTypes = new Set<string>();
  for (const operation of flattenOperationsPerSDK(context.Global.AST.MainSDK)) {
    for (const type of operation.GetAcceptTypes()) {
      allAcceptTypes.add(type.split(";")[0]);
    }
  }
  return Array.from(allAcceptTypes);
}

// getGlobalOptionDocs returns documentation for global SDK options
// (those passed to New()).
function getGlobalOptionDocs(): OptionDoc[] {
  const opt = getOptionNames();
  const sdkPkg = sanitizeSDKPackageName(true);
  const docs: OptionDoc[] = [];

  docs.push({
    Name: opt.GlobalServerURL,
    Description: `${opt.GlobalServerURL} allows providing an alternative server URL.`,
    Example: `${sdkPkg}.${opt.GlobalServerURL}("https://api.example.com")`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.GlobalTemplatedServerURL,
    Description: `${opt.GlobalTemplatedServerURL} allows providing an alternative server URL with templated parameters.`,
    Example: `${sdkPkg}.${opt.GlobalTemplatedServerURL}("https://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  const sdk = context.Global.AST.MainSDK;
  const hasServers = sdk.Servers && sdk.Servers.HasAbsoluteURL();
  if (hasServers) {
    if (sdk.Servers.ServerMap) {
      docs.push({
        Name: opt.GlobalServer,
        Description: `${opt.GlobalServer} allows the overriding of the default server by name.`,
        Example: `${sdkPkg}.${opt.GlobalServer}("my-server")`,
        ExtraDescription: "",
        ExtraExample: "",
      });
    } else {
      docs.push({
        Name: opt.GlobalServerIndex,
        Description: `${opt.GlobalServerIndex} allows the overriding of the default server by index.`,
        Example: `${sdkPkg}.${opt.GlobalServerIndex}(1)`,
        ExtraDescription: "",
        ExtraExample: "",
      });
    }

    for (const v of sdk.Servers.GetVariables()) {
      const name = templateSDKOptionName(v.Name);
      docs.push({
        Name: name,
        Description: `${name} allows setting the ${v.Name} variable for url substitution.`,
        Example: `${sdkPkg}.${name}(/* ... */)`,
        ExtraDescription: "",
        ExtraExample: "",
      });
    }
  }

  docs.push({
    Name: opt.GlobalClient,
    Description: `${opt.GlobalClient} allows the overriding of the default HTTP client used by the SDK.`,
    Example: `${sdkPkg}.${opt.GlobalClient}(httpClient)`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  if (sdk.Security) {
    docs.push({
      Name: opt.GlobalSecurity,
      Description: `${opt.GlobalSecurity} configures the SDK to use the provided security details.`,
      Example: `${sdkPkg}.${opt.GlobalSecurity}(/* ... */)`,
      ExtraDescription: "",
      ExtraExample: "",
    });

    docs.push({
      Name: opt.GlobalSecuritySource,
      Description: `${opt.GlobalSecuritySource} configures the SDK to invoke the provided function on each method call to determine authentication.`,
      Example: `${sdkPkg}.${opt.GlobalSecuritySource}(/* ... */)`,
      ExtraDescription: "",
      ExtraExample: "",
    });
  }

  if (sdk.Globals) {
    for (const field of sdk.Globals.Fields) {
      const name = templateSDKOptionName(field.Name);
      docs.push({
        Name: name,
        Description: `${name} allows setting the ${sanitizeFieldName(
          field.Name,
        )} parameter for all supported operations.`,
        Example: `${sdkPkg}.${name}(/* ... */)`,
        ExtraDescription: "",
        ExtraExample: "",
      });
    }
  }

  docs.push({
    Name: opt.GlobalRetryConfig,
    Description: `${opt.GlobalRetryConfig} allows setting the default retry configuration used by the SDK for all supported operations.`,
    Example: `${sdkPkg}.${opt.GlobalRetryConfig}(retry.Config{
    Strategy: "backoff",
    Backoff: retry.BackoffStrategy{
        InitialInterval: 500 * time.Millisecond,
        MaxInterval: 60 * time.Second,
        Exponent: 1.5,
        MaxElapsedTime: 5 * time.Minute,
    },
    RetryConnectionErrors: true,
})`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.GlobalTimeout,
    Description: `${opt.GlobalTimeout} sets the default request timeout for all operations.`,
    Example: `${sdkPkg}.${opt.GlobalTimeout}(30 * time.Second)`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  return docs;
}

registerTemplateFunc("getGlobalOptionDocs", getGlobalOptionDocs);

// getOperationOptionDocs returns documentation for per-operation options
// (those passed as the last argument to individual SDK methods).
function getOperationOptionDocs(): OptionDoc[] {
  const ns = getAccessNamespace("usage", "operations");
  const opt = getOptionNames();
  const docs: OptionDoc[] = [];

  docs.push({
    Name: opt.ServerURL,
    Description: `${opt.ServerURL} allows providing an alternative server URL for a single request.`,
    Example: `${ns}${opt.ServerURL}("http://api.example.com")`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.TemplatedServerURL,
    Description: `${opt.TemplatedServerURL} allows providing an alternative server URL with templated parameters for a single request.`,
    Example: `${ns}${opt.TemplatedServerURL}("http://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.Retries,
    Description: `${opt.Retries} allows customizing the default retry configuration for a single request.`,
    Example: `${ns}${opt.Retries}(retry.Config{
    Strategy: "backoff",
    Backoff: retry.BackoffStrategy{
        InitialInterval: 500 * time.Millisecond,
        MaxInterval: 60 * time.Second,
        Exponent: 1.5,
        MaxElapsedTime: 5 * time.Minute,
    },
    RetryConnectionErrors: true,
})`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.OperationTimeout,
    Description: `${opt.OperationTimeout} allows setting the request timeout for a single request.`,
    Example: `${ns}${opt.OperationTimeout}(30 * time.Second)`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.SetHeaders,
    Description: `${opt.SetHeaders} allows setting custom headers on a per-request basis. If the request already contains headers matching the provided keys, they will be overwritten.`,
    Example: `${ns}${opt.SetHeaders}(map[string]string{
    "X-Cache-TTL": "60",
})`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  docs.push({
    Name: opt.URLOverride,
    Description: `${opt.URLOverride} allows overriding the default URL for an operation.`,
    Example: `${ns}${opt.URLOverride}("/custom/path")`,
    ExtraDescription: "",
    ExtraExample: "",
  });

  const acceptTypes = getGlobalAcceptTypes();
  if (acceptTypes.length > 1) {
    const exampleEnum = sanitizeAcceptEnumKey(acceptTypes[0]);
    docs.push({
      Name: opt.AcceptHeaderOverride,
      Description: `${opt.AcceptHeaderOverride} allows overriding the \`Accept\` header for operations that support multiple response content types.`,
      Example: `${ns}${opt.AcceptHeaderOverride}(${ns}${exampleEnum})`,
      ExtraDescription: "",
      ExtraExample: "",
    });
  }

  if (isFeatureUsed("operationPolling")) {
    const pollingNs = getAccessNamespace("usage", "polling");
    docs.push({
      Name: opt.Polling,
      Description: `${opt.Polling} enables method-specific polling configurations, such as waiting for a particular HTTP status code or response body content. Only usable with methods that implement polling support.`,
      Example: `${ns}${opt.Polling}(client.ExampleOperationWaitForSuccess())`,
      ExtraDescription: `There are separate polling options available in the \`${getPollingLocation()}\` package for overriding polling behaviors, such as the request count limit. Provide any number of these polling options.`,
      ExtraExample: `${ns}${opt.Polling}(
    client.ExampleOperationWaitForSuccess(),
    ${pollingNs}${opt.PollingDelaySecondsOverride}(5),
    ${pollingNs}${opt.PollingIntervalSecondsOverride}(2),
    ${pollingNs}${opt.PollingLimitCountOverride}(10),
)`,
    });
  }

  return docs;
}

registerTemplateFunc("getOperationOptionDocs", getOperationOptionDocs);
