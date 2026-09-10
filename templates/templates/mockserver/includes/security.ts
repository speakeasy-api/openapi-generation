/** Describes a request header security context within a HTTP handler. */
type RequestHeaderSecurityContext = {
  /** HTTP Authorization scheme for an Authorization header. */
  HttpAuthorizationScheme?: string;

  /** Whether this security context is required or not. */
  Optional: boolean;
};

/** Returns the collection of global and operation request header security contexts. */
function collectRequestHeaderSecurity(
  usageContext: UsageContext,
): Record<string, RequestHeaderSecurityContext> {
  // Determine which security fields this specific test is using (if any).
  // This allows correct assertions for operations with multiple security options (OR).
  const testedFields = getTestedSecurityFieldNames(usageContext);

  const globalSecurityContexts = getRequestHeaderSecurityContexts(
    usageContext.Operation.GlobalSecurity?.Type?.Fields,
    testedFields,
  );
  const operationSecurityContexts = getRequestHeaderSecurityContexts(
    usageContext.Operation.Security?.Type?.Fields,
    testedFields,
  );

  return { ...globalSecurityContexts, ...operationSecurityContexts };
}

/**
 * Extract the set of security field names being tested from the usage context's
 * test security scopes (x-speakeasy-test-security in Arazzo workflows).
 * Returns null if no specific security option is indicated.
 */
function getTestedSecurityFieldNames(
  usageContext: UsageContext,
): Set<string> | null {
  let securityValue: any = null;

  if (usageContext.Scopes) {
    for (const scope of usageContext.Scopes) {
      if (scope.Feature === "security" && scope.Value) {
        securityValue = scope.Value;
        break;
      }
    }
  }

  if (!securityValue && usageContext.Test?.Security) {
    securityValue = usageContext.Test.Security;
  }

  if (!securityValue) return null;

  // Resolve to a plain object: may be an Example with ToJSON or a plain object
  let obj = securityValue;
  if (obj && typeof obj === "object" && "ToJSON" in obj) {
    try {
      obj = JSON.parse(obj.ToJSON());
    } catch {
      return null;
    }
  }

  if (typeof obj !== "object" || obj === null) return null;

  return new Set(Object.keys(obj));
}

/** Returns the request header security contexts for the given security fields. */
function getRequestHeaderSecurityContexts(
  fieldDefs: FieldDef[],
  testedFields: Set<string> | null,
): Record<string, RequestHeaderSecurityContext> {
  let result: Record<string, RequestHeaderSecurityContext> = {};

  if (!fieldDefs) {
    return result;
  }

  for (const fieldDef of fieldDefs) {
    // When we know which security option the test is using, skip unrelated fields.
    if (testedFields) {
      const name = fieldDef.Name;
      const originalName = fieldDef.OriginalName || name;
      if (!testedFields.has(name) && !testedFields.has(originalName)) {
        continue;
      }
    }

    const securityAnnotation = fieldDef.Annotations?.Get(
      "security",
    ) as SecurityAnnotation;

    switch (securityAnnotation?.SecType) {
      case "apiKey":
        if (securityAnnotation.SubType != "header") {
          break;
        }

        if (!result[securityAnnotation.FieldName]) {
          result[securityAnnotation.FieldName] = {
            Optional: fieldDef.Optional,
          };
        }

        break;
      case "http":
        if (!result["Authorization"]) {
          result["Authorization"] = {
            HttpAuthorizationScheme: normalizeHttpAuthorizationScheme(
              securityAnnotation.SubType,
            ),
            Optional: fieldDef.Optional,
          };
        }

        break;
      case "oauth2":
      case "openIdConnect":
        if (!result["Authorization"]) {
          result["Authorization"] = {
            HttpAuthorizationScheme: "Bearer",
            Optional: fieldDef.Optional,
          };
        }

        break;
    }
  }

  return result;
}

/** Returns a given HTTP Authorization scheme in IANA registered form. While
 * strict casing is not required in many IETF RFCs, IANA values should
 * guarantee the most compatibility. For example, the scheme value may be from
 * the OpenAPI specification, which allows for lowercased values such as
 * "bearer" instead of IANA registered "Bearer".
 *
 * Reference: https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
 * */
function normalizeHttpAuthorizationScheme(original: string): string {
  switch (original.toLowerCase()) {
    case "basic":
      return "Basic";
    case "bearer":
      return "Bearer";
    case "digest":
      return "Digest";
    case "mutual":
      return "Mutual";
    case "privatetoken":
      return "PrivateToken";
    // Otherwise, this was best effort.
    default:
      return original;
  }
}
