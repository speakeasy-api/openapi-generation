// Java 8 compatibility helpers.
//
// The generated SDK normally targets JDK 9+ APIs (List.of, Map.of, Optional.or,
// Files.writeString, ...). When LanguageVersion is 8, call sites route those
// static calls through utils.Java8Compat instead. resolveJava8Compat (config.ts)
// guarantees, on Java 8:
// - no java.net.http
// - OkHttp transport
// - async is CompletableFuture-based (the okhttp arm); the reactive
//   (Publisher-based) async surface exists only on the JDK arm (Java 11+)

function isJava8(): boolean {
  return Number(context.Global.Config.LanguageVersion) === 8;
}
registerTemplateFunc("isJava8", isJava8);

// JDK 9+ static calls mapped to [default-arm import path, Java8Compat method].
// Call sites swap the call prefix only, so the default (non-Java 8) arm stays
// byte-identical; the Java 8 arm always imports utils.Java8Compat.
const JAVA8_COMPAT_CALLS: Record<string, [string, string]> = {
  "List.of": ["java.util.List", "listOf"],
  "Map.of": ["java.util.Map", "mapOf"],
  "Set.of": ["java.util.Set", "setOf"],
  "Set.copyOf": ["java.util.Set", "setCopyOf"],
  "Map.entry": ["java.util.Map", "mapEntry"],
  "Map.ofEntries": ["java.util.Map", "mapOfEntries"],
  "Files.writeString": ["java.nio.file.Files", "writeString"],
  "CompletableFuture.failedFuture": [
    "java.util.concurrent.CompletableFuture",
    "failedFuture",
  ],
};

function lookupJava8CompatCall(call: string): [string, string] {
  const entry = JAVA8_COMPAT_CALLS[call];
  if (!entry) {
    throw new Error(`no Java 8 replacement registered for "${call}"`);
  }
  return entry;
}

// For .stmpl call sites: the enclosing file declares its own imports.
function jdkCompat(call: string): string {
  if (!isJava8()) {
    return call;
  }
  const [, compatMethod] = lookupJava8CompatCall(call);
  return `Java8Compat.${compatMethod}`;
}
registerTemplateFunc("jdkCompat", jdkCompat);

function templateJdkCall(call: string): string {
  const [importPath, compatMethod] = lookupJava8CompatCall(call);
  if (isJava8()) {
    return `${javaImportLocal("utils.Java8Compat")}.${compatMethod}`;
  }
  return `${javaImport(importPath)}.${call.split(".")[1]}`;
}

// Method-reference form of jdkCompat for .stmpl call sites: `List::of` on
// Java 9+, `Java8Compat::listOf` on Java 8. The enclosing file declares its
// own imports.
function jdkCompatRef(call: string): string {
  if (!isJava8()) {
    return call.replace(".", "::");
  }
  const [, compatMethod] = lookupJava8CompatCall(call);
  return `Java8Compat::${compatMethod}`;
}
registerTemplateFunc("jdkCompatRef", jdkCompatRef);

// Method reference to a JDK 9+ static method; routes through Java8Compat on
// Java 8. For generated files with managed imports (javaImport).
function templateJdkMethodRef(call: string): string {
  const [importPath, compatMethod] = lookupJava8CompatCall(call);
  if (isJava8()) {
    return `${javaImportLocal("utils.Java8Compat")}::${compatMethod}`;
  }
  return `${javaImport(importPath)}::${call.split(".")[1]}`;
}
registerTemplateFunc("templateJdkMethodRef", templateJdkMethodRef);

// Local variable declaration: `var` is Java 10+; Java 8 call sites declare
// the explicit type instead. The type is resolved through javaImport (only on
// the Java 8 arm), so pass a fully qualified name in files with managed
// imports.
function templateJdkVar(java8Type: string, varName: string): string {
  const type = isJava8() ? javaImport(java8Type) : "var";
  return `${type} ${varName}`;
}
registerTemplateFunc("templateJdkVar", templateJdkVar);

function javaHttpRequestType(): string {
  if (useOkHttp()) {
    return javaImportLocal("utils.transport.HttpRequest");
  }
  return javaImport("java.net.http.HttpRequest");
}
registerTemplateFunc("javaHttpRequestType", javaHttpRequestType);

// innerType: fully qualified generic parameter, "?" for a wildcard, or absent
// for the raw type. Both arms are generic: `java.net.http.HttpResponse<T>` on
// the JDK transport, the SDK-owned `utils.transport.HttpResponse<T>` on okhttp.
function javaHttpResponseType(innerType?: string): string {
  const base = useOkHttp()
    ? javaImportLocal("utils.transport.HttpResponse")
    : javaImport("java.net.http.HttpResponse");
  if (!innerType) {
    return base;
  }
  if (innerType === "?") {
    return `${base}<?>`;
  }
  return `${base}<${javaImport(innerType)}>`;
}
registerTemplateFunc("javaHttpResponseType", javaHttpResponseType);

// Fully qualified raw-response field type for the sanitization mapper
// (per-model rawResponse fields, method-level response typing).
function javaRawResponseType(isAsync: boolean = false): string {
  const base = useOkHttp()
    ? `${templatePackageName()}.utils.transport.HttpResponse`
    : "java.net.http.HttpResponse";
  // The reactive (JDK arm) async body is a Blob over a Publisher; the okhttp
  // arm's CompletableFuture-based async body is a plain InputStream, as in
  // sync.
  return isAsync && !useOkHttp()
    ? `${base}<${templatePackageName()}.utils.Blob>`
    : `${base}<java.io.InputStream>`;
}
registerTemplateFunc("javaRawResponseType", javaRawResponseType);

// Async raw body type as seen by the operation seams (doRequest/onError/
// onSuccess): plain InputStream on the okhttp (CompletableFuture) arm,
// Blob-over-Publisher on the JDK reactive arm.
function javaAsyncBodyType(): string {
  return useOkHttp()
    ? javaImport("java.io.InputStream")
    : javaImportLocal("utils.Blob");
}
registerTemplateFunc("javaAsyncBodyType", javaAsyncBodyType);

// `HttpResponse<InputStream|Blob>` for async operation seam signatures.
function javaAsyncHttpResponseType(): string {
  return `${javaImport("java.net.http.HttpResponse")}<${javaAsyncBodyType()}>`;
}
registerTemplateFunc("javaAsyncHttpResponseType", javaAsyncHttpResponseType);

// `CompletableFuture<HttpResponse<InputStream|Blob>>` — the async transport
// stage type threaded through doRequest/onError/onSuccess.
function javaAsyncResponseFutureType(): string {
  return `${javaImport(
    "java.util.concurrent.CompletableFuture",
  )}<${javaAsyncHttpResponseType()}>`;
}
registerTemplateFunc(
  "javaAsyncResponseFutureType",
  javaAsyncResponseFutureType,
);

// Wildcard raw-response type for var-style declarations in usage snippets:
// the base error type stores the response as `HttpResponse<?>`.
function javaRawResponseWildcardType(): string {
  const base = useOkHttp()
    ? `${templatePackageName()}.utils.transport.HttpResponse`
    : "java.net.http.HttpResponse";
  return `${base}<?>`;
}
registerTemplateFunc(
  "javaRawResponseWildcardType",
  javaRawResponseWildcardType,
);

// Fully qualified names for the transport-independent reactive/async types.
// Used both for literal `import <fqn>;` lines in auxiliary templates and for
// the javaImport remap below. The default (JDK transport, Java 11+) arm stays
// byte-identical.
function javaHttpResponseFqn(): string {
  return useOkHttp()
    ? `${templatePackageName()}.utils.transport.HttpResponse`
    : "java.net.http.HttpResponse";
}
registerTemplateFunc("javaHttpResponseFqn", javaHttpResponseFqn);

function javaHttpRequestFqn(): string {
  return useOkHttp()
    ? `${templatePackageName()}.utils.transport.HttpRequest`
    : "java.net.http.HttpRequest";
}
registerTemplateFunc("javaHttpRequestFqn", javaHttpRequestFqn);

// `java.util.concurrent.Flow` (Java 9+) backs the reactive async surface,
// which only exists on the JDK arm (Java 11+). The former Java 8
// `utils.reactive.Flow` backport was removed along with the okhttp reactive
// pump; nothing on the Java 8 (okhttp, CompletableFuture) arm may reference
// this type.
function javaFlowFqn(): string {
  return "java.util.concurrent.Flow";
}
registerTemplateFunc("javaFlowFqn", javaFlowFqn);

// `org.reactivestreams.FlowAdapters` requires Java 9+ (it adapts
// `java.util.concurrent.Flow`); reactive-only, JDK arm only — see javaFlowFqn.
function javaFlowAdaptersFqn(): string {
  return "org.reactivestreams.FlowAdapters";
}
registerTemplateFunc("javaFlowAdaptersFqn", javaFlowAdaptersFqn);

// Remaps JDK-transport / Java 9+ types to their SDK-owned counterparts for
// javaImport call sites (generated operations, methods.ts, pagination).
// Auxiliary templates with literal import lines use the *Fqn functions above.
function mapCompatJavaType(javaType: string): string {
  switch (javaType) {
    case "java.net.http.HttpResponse":
      return javaHttpResponseFqn();
    case "java.net.http.HttpRequest":
      return javaHttpRequestFqn();
    case "java.util.concurrent.Flow":
      return javaFlowFqn();
    case "org.reactivestreams.FlowAdapters":
      return javaFlowAdaptersFqn();
    default:
      return javaType;
  }
}
