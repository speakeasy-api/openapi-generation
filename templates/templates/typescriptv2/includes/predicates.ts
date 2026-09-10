function reservedErrorFields(): string[] {
  // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error#instance_properties
  const base = [
    getConstants().errorFields.name,
    getConstants().errorFields.message,
    getConstants().errorFields.stack,
    getConstants().errorFields.cause,
    getConstants().errorFields.data$,
  ];

  if (isErrorEnvelope()) {
    return [...base, getConstants().errorFields.httpMeta];
  }

  return [
    ...base,
    getConstants().errorFields.statusCode,
    getConstants().errorFields.contentType,
    getConstants().errorFields.body,
    getConstants().errorFields.rawResponse,
    getConstants().errorFields.headers,
  ];
}

// @ts-ignore
function isReservedErrorField(fieldName: string): boolean {
  return reservedErrorFields().includes(fieldName);
}
registerTemplateFunc("isReservedErrorField", isReservedErrorField);

function usesSerialization(op: Operation, match: string): boolean {
  return op.SerializationMethod?.toString() === match;
}
registerTemplateFunc("usesSerialization", usesSerialization);

function isSecurityModel(typeDef: TypeDef): boolean {
  return typeDef.Fields.findIndex((fd) => fd.Annotations?.Has("security")) >= 0;
}
registerTemplateFunc("isSecurityModel", isSecurityModel);

function isMultipartRequest(op: Operation): boolean {
  return getRequestMediaType(op).startsWith("multipart/");
}
registerTemplateFunc("isMultipartRequest", isMultipartRequest);

/**
 * Returns whether or not an operation yields a valid response backed by
 * generated models. This means the response has at least one success (2XX)
 * response or an error response with a body defined.
 */
function hasConcreteResponses(op: Operation): boolean {
  const res = op.Response;
  if (!res) {
    return false;
  }

  if (!res.Responses.length) {
    return false;
  }

  return (
    res.Responses.findIndex((r) => {
      // A success response, whether it has a defined body or not, means the
      // operation will have a concrete model defined for its response that we
      // can use.
      if (!r.Error) {
        return true;
      }

      return (
        r.Content.findIndex((c) => {
          return c.Content.Type != null;
        }) >= 0
      );
    }) >= 0
  );
}

registerTemplateFunc("hasConcreteResponses", hasConcreteResponses);

/**
 * Returns the set of built-in SDKOptions property names that could collide
 * with a flattened global security field name.
 */
function sdkOptionsBuiltinNames(): Set<string> {
  const c = getConstants();
  return new Set([
    c.httpClient,
    c.serverIdx,
    c.serverURL,
    c.security,
    c.server,
    c.userAgent,
    c.timeoutMs,
    c.debugLogger,
    c.retryConfig,
  ]);
}

/**
 * Returns whether global security can be safely flattened into SDKOptions
 * without causing duplicate property names.
 */
function canFlattenGlobalSecurity(): boolean {
  const security = context.Global.AST.MainSDK.Security;
  if (!security || security.Type.Fields.length !== 1) return false;
  if (!context.Global.Config.FlattenGlobalSecurity) return false;

  const fieldName = sanitizeFieldName(security.Type.Fields[0].Name);
  return !sdkOptionsBuiltinNames().has(fieldName);
}
registerTemplateFunc("canFlattenGlobalSecurity", canFlattenGlobalSecurity);

/**
 * Returns whether SDK method request options should expose request extras
 * such as the extraQuery map.
 */
function requestExtrasEnabled(): boolean {
  return context.Global.Config.RequestExtras === true;
}
registerTemplateFunc("requestExtrasEnabled", requestExtrasEnabled);

/**
 * Returns whether the generated APIPromise should expose `.withResponse()`
 * and `.asResponse()` helpers and be re-exported from the package root.
 */
function withAPIPromiseHelpers(): boolean {
  return context.Global.Config.APIPromiseHelpers === true;
}
registerTemplateFunc("withAPIPromiseHelpers", withAPIPromiseHelpers);

/**
 * Returns the npm install command URL. When there's a subdirectory,
 * uses git+URL.git?subdir=DIR format (requires npm with pacote >= 21.2.0).
 */
function npmInstallURL(): string {
  const url = context.Global.Config.InstallationURL || "<UNSET>";
  const subDir = context.Global.Config.RepoSubDirectory;
  if (!subDir) {
    return url;
  }
  const gitUrl = url.endsWith(".git") ? url : url + ".git";
  return `git+${gitUrl}?subdir=${subDir}`;
}
registerTemplateFunc("npmInstallURL", npmInstallURL);

/**
 * Returns the pnpm install command URL. When there's a subdirectory,
 * uses URL#path:DIR format.
 */
function pnpmInstallURL(): string {
  const url = context.Global.Config.InstallationURL || "<UNSET>";
  const subDir = context.Global.Config.RepoSubDirectory;
  if (!subDir) {
    return url;
  }
  return `${url}#path:${subDir}`;
}
registerTemplateFunc("pnpmInstallURL", pnpmInstallURL);
