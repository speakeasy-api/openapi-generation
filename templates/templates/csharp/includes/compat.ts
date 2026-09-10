// The compile target framework moniker emitted as <TargetFramework>.
// targetFramework > dotnetVersion is rejected in getCompileDependencies.
function getTargetFramework(): string {
  const tfm = context.Global.Config.TargetFramework;
  if (!tfm || tfm == "dotnetVersion") {
    return context.Global.Config.DotnetVersion;
  }
  return tfm;
}
registerTemplateFunc("getTargetFramework", getTargetFramework);

function isNetstandard2(): boolean {
  return getTargetFramework() == "netstandard2.0";
}
registerTemplateFunc("isNetstandard2", isNetstandard2);

// Executable and test projects must target the build SDK (dotnetVersion)
function getExecutableFramework(): string {
  return context.Global.Config.DotnetVersion;
}
registerTemplateFunc("getExecutableFramework", getExecutableFramework);

// Returns the C# language version when target is decoupled from dotnetVersion.
function getLangVersion(): string {
  if (getTargetFramework() == context.Global.Config.DotnetVersion) {
    return "";
  }
  switch (context.Global.Config.DotnetVersion) {
    case "net10.0":
      return "14.0";
    case "net8.0":
      return "12.0";
    case "net6.0":
      return "10.0";
    case "net5.0":
      return "9.0";
    default:
      return "latest";
  }
}
registerTemplateFunc("getLangVersion", getLangVersion);

// NodaTime is required when the target lacks DateOnly/TimeOnly
function useNodatime(): boolean {
  return (
    context.Global.Config.UseNodatime ||
    getTargetFramework() == "net5.0" ||
    getTargetFramework() == "netstandard2.0"
  );
}
registerTemplateFunc("useNodatime", useNodatime);

// Whether to pass a CancellationToken to TextReader.ReadLineAsync.
function isReadLineAsyncCancellable(): boolean {
  if (!context.Global.Config.EnableCancellationToken) {
    return false;
  }
  const tfm = getTargetFramework();
  return tfm != "net5.0" && tfm != "net6.0" && !isNetstandard2();
}
registerTemplateFunc("isReadLineAsyncCancellable", isReadLineAsyncCancellable);

// Emits the HttpRequestMessage constructor for an operation.
// HttpMethod.Patch is net5.0/netstandard2.1+; netstandard2.0 lacks it.
function templateHttpRequest(op): string {
  const verb = sanitizeFieldName(op.Method);
  const method =
    isNetstandard2() && verb == "Patch"
      ? `new HttpMethod("PATCH")`
      : `HttpMethod.${verb}`;
  return `new HttpRequestMessage(${method}, urlString)`;
}
registerTemplateFunc("templateHttpRequest", templateHttpRequest);

// String.Split(char)/Split(string) and String.Contains(char) are netstandard2.1+;
// netstandard2.0 exposes only Split(char[]) and Contains(string).
function compatStringCall(
  varName: string,
  method: string,
  arg: string,
): string {
  if (isNetstandard2()) {
    const inner = arg.slice(1, -1);
    if (method == "Split") {
      return `${varName}.Split(new char[] { '${inner}' })`;
    }
    if (method == "Contains") {
      return `${varName}.Contains("${inner}")`;
    }
  }
  return `${varName}.${method}(${arg})`;
}
registerTemplateFunc("compatStringCall", compatStringCall);
