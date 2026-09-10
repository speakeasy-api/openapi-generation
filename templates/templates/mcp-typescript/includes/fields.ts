type SecurityScheme = { type: string; value: string; fieldName?: string };
type SecurityRequirements = SecurityScheme[];
type SecurityOptions = SecurityRequirements[];

function genSecuritySpecs(
  securityDef: TypeDef | undefined,
  accessor: string,
  usageLocation: string,
  withEnv = false,
): SecurityOptions {
  // - Pushing an element into this array builds an OR condition.
  // - Pushing an element into one of its child arrays is building an AND
  //   condition among the members of that child array.
  const options: SecurityOptions = [];
  if (securityDef == null) {
    return options;
  }

  const reqs: SecurityRequirements = [];
  // We need to process basic auth fields in one go and produce an object with
  // "username" and "password" keys. This variable will hold the first occurence
  // of a basic auth field. When the second field comes along we do the work to
  // produce the security spec for the basic scheme.
  let basicKeyVal = "";

  securityDef.Fields.forEach((fieldDef) => {
    const ann = fieldDef.Annotations?.Get("security");

    if (ann == null || !isSecurityAnnotation(ann)) {
      return;
    }

    const name = ann.Option
      ? fieldDef.Name
      : sanitizeFieldName(originalFieldName(fieldDef));
    const property = sanitizeAccessor(accessor, name, true);

    // If we are on an option field, then we should descend and work through
    // each of its members.
    if (ann.Option) {
      options.push(
        ...genSecuritySpecs(fieldDef.Type, property, usageLocation, withEnv),
      );
      return;
    }

    // If a field is not a scheme field, e.g. the token url field in oauth2
    // client credentials, then skip over them.
    if (!ann.Scheme) {
      return;
    }

    const type = [ann.SecType, ann.SubType].filter(Boolean).join(":");

    const envAcc = usageLocation === "lib" ? "env()" : "env$()";
    const envSymbol = usageLocation === "lib" ? "env" : "env as env$";
    let envVarFallback = "";
    if (withEnv) {
      addInternalImport("env", envSymbol, usageLocation, typeImport);
      envVarFallback = ` || ${envAcc}.${templateSecurityEnvVars(fieldDef)}`;
    }

    const isBasicClass =
      type === "http:basic" && fieldDef.Type?.Type.toString() === "class";
    const isResourceOwnerPassword =
      type === "oauth2:password" && isFeatureUsed("oauth2Password");
    const isClientCredClass =
      type === "oauth2:client_credentials" &&
      fieldDef.Type?.Type.toString() === "class";
    const isBasicField =
      type === "http:basic" &&
      fieldDef.Type?.Type.toString() === "string" &&
      (name === "username" || name === "password");
    const isCustomHttpClass =
      type === "http:custom" && fieldDef.Type?.Type.toString() === "class";

    let value = "";
    if (isBasicClass || isClientCredClass || isCustomHttpClass) {
      // This branch is hit when we have one of the following:
      // - Basic HTTP class (either unflattened or along other security requirements)
      // - Client Credentials class (either unflattened or along other security requirements)
      // - Custom HTTP class (never flattened, even when it is the only security requirement*)
      //
      // In these cases we iterate over the scheme's fields to build the security spec with an
      // object value. Note that for `http:custom` we include all fields in the class, while for
      // `http:basic` and `oauth2:client_credentials` we filter out fields that don't have the
      // "security" annotation. For example we skip the `tokenURL` field in Client Credentials.
      //
      // *note: the schema defined under `x-speakeasy-custom-security-scheme`
      // must be an object with properties otherwise `getCustomSecurityScheme`
      // would throw a validation error.

      const kvs = fieldDef.Type.Fields.map((f) => {
        if (!isCustomHttpClass && !f.Annotations?.Get("security")) {
          return;
        }
        let val = sanitizeAccessor(
          property,
          sanitizeFieldName(originalFieldName(f)),
          true,
        );
        if (withEnv) {
          val += ` || ${envAcc}.${templateSecurityEnvVars(f)}`;
        }
        return `${sanitizeFieldName(originalFieldName(f))}: ${val}`;
      })
        .filter(Boolean)
        .join(", ");
      value = `{ ${kvs} }`;
    } else if (isBasicField) {
      // This branch is hit when basic auth class IS flattened. Usually when it
      // is the only security requirement. In this case we wait until we've
      // collected both the username and password fields before building the
      // security spec to have an object value containing these fields.

      const val = sanitizeAccessor(accessor, name, true);
      const kv = `${name}: ${val}` + envVarFallback;
      // If this is the first basic auth field we encountered, then we'll just
      // pluck it out and go around the loop again.
      if (!basicKeyVal) {
        basicKeyVal = kv;
        return;
      }

      value = `{ ${basicKeyVal}, ${kv} }`;
      basicKeyVal = "";
    } else if (isResourceOwnerPassword) {
      const defaults = generateOAuth2Defaults(
        fieldDef,
        "password",
        withEnv ? envAcc : "",
      );
      if (usageLocation !== "lib") {
        addInternalImport(
          "security",
          "resolveOAuth2Password",
          usageLocation,
          typeImport,
        );
      }
      value = `resolveOAuth2Password(${property}, { defaults: ${defaults} })`;
    } else {
      value = property + envVarFallback;
    }

    const opt: SecurityScheme = { type, value };
    if (type !== "http:basic") {
      opt.fieldName = ann.FieldName;
    }

    if (ann.SecurityOption) {
      options.push([opt]);
    } else {
      reqs.push(opt);
    }
  });

  if (reqs.length) {
    options.push(reqs);
  }

  return options;
}

registerTemplateFunc("genSecuritySpecs", genSecuritySpecs);

function generateOAuth2Defaults(
  securityField: FieldDef,
  flow: "password",
  envAccessor: string,
): string {
  const vars = envAccessor
    ? mapOAuth2EnvVariables(securityField, flow)
    : undefined;

  let defaultTokenURL = "";
  const credType = getOAuth2UnionMember(securityField, "credentials");
  const tokenURLField = credType.Fields.find((f) => f.Name === "TokenURL");
  if (typeof tokenURLField?.Default?.Value === "string") {
    const serialized = JSON.stringify(tokenURLField.Default.Value);
    defaultTokenURL =
      envAccessor && vars?.credentials.TokenURL
        ? `${sanitizeAccessor(
            envAccessor,
            vars?.credentials.TokenURL,
          )} || ${serialized}`
        : serialized;
  } else {
    throw new Error(
      "oauth2 password flow requires a default token URL to be specified",
    );
  }

  const defaults = {
    token: vars?.token ? sanitizeAccessor(envAccessor, vars?.token) : undefined,
    clientID: vars?.credentials.ClientID
      ? sanitizeAccessor(envAccessor, vars?.credentials.ClientID)
      : undefined,
    clientSecret: vars?.credentials.ClientSecret
      ? sanitizeAccessor(envAccessor, vars?.credentials.ClientSecret)
      : undefined,
    username: vars?.credentials.Username
      ? sanitizeAccessor(envAccessor, vars?.credentials.Username)
      : undefined,
    password: vars?.credentials.Password
      ? sanitizeAccessor(envAccessor, vars?.credentials.Password)
      : undefined,
    tokenURL: defaultTokenURL,
  };

  let code = "";
  for (const [key, value] of Object.entries(defaults)) {
    if (value != null) {
      code += `${key}: ${value}, `;
    }
  }
  return `{ ${code} }`;
}

/**
 * Resolves the field name of a field as it is defined in the API spec. It will
 * fall back to the computed name if a canonical name cannot be found.
 */
//@ts-ignore
function originalFieldName(field: FieldDef): string {
  // If the field is a parameter, typically use original name in the API spec.
  // However, if the parameter name has been suffixed due to a conflict, we
  // use the computed name instead. Ideally, the AST would have a field to
  // denote the intentional renaming, but currently it does not, so this logic
  // is hardcoded to match the AST behavior.
  // Reference: internal/utils.paramTypeToSuffix
  if (
    hasAnnotation(field, "param") &&
    (field.Name.endsWith("PathParameter") ||
      field.Name.endsWith("QueryParameter"))
  ) {
    return field.Name;
  }

  return field.OriginalName || field.Name;
}

unregisterTemplateFunc("originalFieldName");
registerTemplateFunc("originalFieldName", originalFieldName);

function sortOptionalFieldDefs(fields: FieldDef[]): FieldDef[] {
  fields = fields.slice();
  return fields.sort((a, b) => {
    const aRequired = a.Optional || a.Type?.Type.toString() === "any";
    const bRequired = b.Optional || b.Type?.Type.toString() === "any";

    if (aRequired && !bRequired) {
      return 1;
    } else if (!aRequired && bRequired) {
      return -1;
    } else {
      return originalFieldName(a).localeCompare(originalFieldName(b));
    }
  });
}

registerTemplateFunc("sortOptionalFieldDefs", sortOptionalFieldDefs);

/**
 * This function digs out all the fields from a security TypeDef that can be set
 * using environment variables. It traverses security options and unpacks
 * schemes which take an object for security values such as Basic and OAuth2
 * client credentials.
 */
function unnestSecurityEnvFields(
  security: TypeDef,
  seen: Set<string> = new Set(),
): FieldDef[] {
  const fields: FieldDef[] = [];

  security.Fields.forEach((field) => {
    const ev = templateSecurityEnvVars(field);

    const ann = field.Annotations?.Get("security");

    const isAllowedEnvType =
      field.Type.IsPrimitive() ||
      (field.Type.Type.toString() === "array" &&
        field.Type.ItemType.IsPrimitive());

    if (ann?.Option || field.Type.Type.toString() === "class") {
      fields.push(...unnestSecurityEnvFields(field.Type, seen));
    } else if (
      ann?.SecType === "oauth2" &&
      ann?.SubType === "password" &&
      field.Type.Type.toString() === "union" &&
      !seen.has(ev)
    ) {
      const len = field.Type.AssociatedTypes.length;
      if (field.Type.AssociatedTypes.length !== 2) {
        throw new Error(
          `oauth password type is expected to be a union of 2 members. got: ${len} members.`,
        );
      }
      const memberFields = field.Type.AssociatedTypes.flatMap((t) => {
        const dataType = t.Type.toString();
        if (dataType === "class") {
          return unnestSecurityEnvFields(t, seen);
        } else if (dataType === "string") {
          const tokenField = typeDefToFieldDef(t);
          tokenField.Name = "Token";
          const tokenEnv = templateSecurityEnvVars(tokenField);
          seen.add(tokenEnv);
          return [tokenField];
        } else {
          throw new Error(
            `oauth password scheme union has unrecognized member type: ${dataType}`,
          );
        }
      });
      fields.push(...memberFields);
    } else if (isAllowedEnvType && !seen.has(ev)) {
      fields.push(field);
      seen.add(ev);
    }
  });

  return fields;
}

registerTemplateFunc("unnestSecurityEnvFields", unnestSecurityEnvFields);

type RemoteServerHeaders = {
  fieldName: string;
  headerName: string;
  type: string;
  isString: boolean;
};

function getRemoteServerHeaders(securityDef: TypeDef | undefined) {
  if (!securityDef) return [];

  const options: RemoteServerHeaders[][] = [];
  const flatHeaders = getRemoteServerHeadersInner(securityDef);
  if (flatHeaders.length) {
    options.push(flatHeaders);
  }

  for (const option of securityDef.Fields) {
    const ann = option.Annotations?.Get("security");
    if (!ann || !isSecurityAnnotation(ann) || !ann.Option) {
      continue;
    }
    options.push(getRemoteServerHeadersInner(option.Type));
  }

  return options;
}
registerTemplateFunc("remoteServerHeaders", getRemoteServerHeaders);

function getRemoteServerHeadersInner(securityDef: TypeDef | undefined) {
  if (!securityDef) return;

  const ret: RemoteServerHeaders[] = [];

  for (const field of securityDef.Fields) {
    const ann = field.Annotations?.Get("security");
    if (!ann || !isSecurityAnnotation(ann) || ann.Option) {
      continue;
    }
    const secType = [ann.SecType, ann.SubType].filter(Boolean).join(":");

    // default to kebab-case
    let headerName = caser().ToKebab(field.Name);
    if (!headerName.startsWith("x-")) {
      headerName = `x-${headerName}`;
    }
    if (secType === "apiKey:header") {
      headerName = ann.FieldName;
    }
    if (secType === "http:bearer") {
      headerName = "authorization";
    }

    ret.push({
      fieldName: sanitizeFieldName(originalFieldName(field)),
      headerName,
      type: secType,
      isString: field.Type.Type.toString() === "string",
    });
  }
  return ret;
}

function templateRemoteServerHeaders(securityDef: TypeDef | undefined) {
  const headers = getRemoteServerHeaders(securityDef);
  if (!headers.length) return;

  const ret = [];
  const usePrefix = headers.length > 1;

  for (const [i, option] of headers.entries()) {
    const prefix = usePrefix ? `Option${i + 1}` : "";
    if (prefix) {
      ret.push(`${prefix}: {`);
    }
    for (const header of option) {
      let v = `getHeader("${header.headerName}")`;
      if (!header.isString) {
        v = `JSON.parse(${v})`;
      }
      ret.push(`${header.fieldName}: ${v},`);
    }
    if (prefix) ret.push(`},`);
  }
  return `{ ${ret.join("\n")} }`;
}
registerTemplateFunc(
  "templateRemoteServerHeaders",
  templateRemoteServerHeaders,
);

function getMCPServerHeaders(securityDef: TypeDef | undefined) {
  const headers = getRemoteServerHeaders(securityDef);
  if (!headers || headers.length === 0) return null;

  // For MCP server headers, we need to flatten all options and deduplicate by header name
  const byHeaderName = new Map<string, RemoteServerHeaders>();
  for (const option of headers) {
    for (const header of option) {
      byHeaderName.set(header.headerName, header);
    }
  }

  return Array.from(byHeaderName.values());
}
registerTemplateFunc("getMCPServerHeaders", getMCPServerHeaders);

function templateMCPServerHeaders(
  securityDef: TypeDef | undefined,
  pkgNameKebab: string,
  pkgNameUpper: string,
) {
  const headers = getRemoteServerHeaders(securityDef);
  if (!headers || headers.length === 0) return "";

  const byHeaderName = new Map<string, RemoteServerHeaders>();
  for (const option of headers) {
    for (const header of option) {
      byHeaderName.set(header.headerName, header);
    }
  }

  const lines = [];
  for (const [headerName, header] of byHeaderName.entries()) {
    const envVar = `${sanitizeEnvVarName(pkgNameUpper)}_${caser().ToSNAKE(
      header.fieldName,
    )}`;
    lines.push(`${headerName}: \${${envVar}}`);
  }

  return lines.join("\n");
}
registerTemplateFunc("templateMCPServerHeaders", templateMCPServerHeaders);

function templateHeaderNames(securityDef: TypeDef | undefined) {
  const headers = getRemoteServerHeaders(securityDef);
  if (!headers || headers.length === 0) return [];

  const byHeaderName = new Map<string, RemoteServerHeaders>();
  for (const option of headers) {
    for (const header of option) {
      byHeaderName.set(header.headerName, header);
    }
  }

  return Array.from(byHeaderName.keys());
}
registerTemplateFunc("templateHeaderNames", templateHeaderNames);

/**
 * Generates security object for Gram deployment, reading from process.env.
 * Similar to templateMCPSecurityAccess but uses env vars instead of CLI flags.
 */
function templateGramSecurityFromEnv(secField: FieldDef): string {
  const fields = unnestSecurityEnvFields(secField.Type);
  if (fields.length === 0) return "{}";

  const fieldAssignments = fields.map((field) => {
    const fieldName = sanitizeFieldName(originalFieldName(field));
    const envVar = templateSecurityEnvVars(field);
    return `${fieldName}: process.env["${envVar}"]`;
  });

  return `{ ${fieldAssignments.join(", ")} }`;
}
registerTemplateFunc(
  "templateGramSecurityFromEnv",
  templateGramSecurityFromEnv,
);
