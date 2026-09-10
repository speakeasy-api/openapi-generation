// @ts-ignore
type securityParam = {
  Type: string;
  Name: string;
  Source: string;
  FlattenedFieldName: string;
  Optional: boolean;
};

// @ts-ignore
function getSecurityParam(sdk: SDK): securityParam {
  if (!sdk.Security) {
    return null;
  }

  if (
    sdk.Security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    const secField = sdk.Security.Type.Fields[0];
    const fieldName = sanitizeSecurityFieldName(secField.Name);
    const paramName = sanitizePrivateFieldName(fieldName);
    return {
      Type: sanitizeType(secField.Type, false, ""),
      Name: paramName,
      Source: `${paramName}Source`,
      FlattenedFieldName: fieldName,
      Optional: secField.Optional,
    };
  }

  return {
    Type: getGlobalSecurityClass(),
    Name: "security",
    Source: "securitySource",
    FlattenedFieldName: "",
    Optional: sdk.Security.Optional,
  };
}

registerTemplateFunc("getSecurityParam", getSecurityParam);

// @ts-ignore
function getGlobalSecurityClass(): string {
  const security = context.Global.AST.MainSDK.Security;
  if (!security) {
    return "";
  }

  return `${getModelNamespace(security.Type.OutputLocation, true)}.Security`;
}

registerTemplateFunc("getGlobalSecurityClass", getGlobalSecurityClass);

// @ts-ignore
interface SecurityInit {
  fieldName: string;
  value: string;
  isOptional: boolean;
}

function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): SecurityInit | undefined {
  if (!context.Global.AST.MainSDK.Security) {
    return;
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example, additionalContext);
    }
  }

  const security = context.Global.AST.MainSDK.Security;
  const isOptional =
    security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme" &&
    index === undefined &&
    example === undefined;

  if (isOptional) {
    return;
  }

  let securityFieldName = "security";
  addUsageImportForType(security.Type);

  const flattenedSecurity = context.Global.Config.FlattenGlobalSecurity;

  if (security.Type.Fields.length == 1 && flattenedSecurity) {
    securityFieldName = sanitizeGlobalSecurityFieldName(
      security.Type.Fields[0].Name,
    );
  }

  const securityValue = templateSecurityUsage(
    security.Type,
    0,
    true,
    flattenedSecurity,
    index,
    example,
    additionalContext,
  ).trim();

  return {
    fieldName: securityFieldName,
    value: securityValue,
    isOptional: security.Optional,
  };
}

// Override common templateScheme — C# needs the parent type passed to
// templateFieldDeclaration so sanitizeFieldName can detect when a field
// name collides with its enclosing class name (CS0542).
// The parent type is passed via additionalContext.schemeParent, set by
// templateOption in the common security usage code.
// @ts-ignore
function templateScheme(
  fieldDef: FieldDef,
  schemeAnno: SecurityAnnotation,
  optionalPrefix: string,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  const schemeParent = additionalContext?.schemeParent;

  if (fieldDef.Type.IsPrimitive()) {
    return indentLines(
      [
        `${templateFieldDeclaration(
          fieldDef,
          schemeParent,
        )}${templateSecurityFieldValue(
          fieldDef,
          schemeAnno,
          optionalPrefix != "",
          example,
          additionalContext,
        )}${templateFieldDelimiter()}`,
      ],
      indent,
    );
  } else {
    let fields = [];

    for (const field of fieldDef.Type.Fields) {
      const fieldExample =
        example !== undefined ? example[field.OriginalName] : undefined;

      fields.push(
        `${templateFieldDeclaration(field, fieldDef.Type)}${templateSchemeValue(
          field,
          schemeAnno,
          fieldExample,
          additionalContext,
        )}${templateFieldDelimiter()}`,
      );
    }

    return (
      indentLines(
        [
          `${templateFieldDeclaration(fieldDef)}${optionalPrefix}${templateType(
            fieldDef.Type,
            { security: true },
          )}${templateBracket(fieldDef.Type, true, { security: true })}`,
          indentLines(fields, indent),
          templateBracket(fieldDef.Type, false, { security: true }),
        ],
        indent,
      ) + templateFieldDelimiter()
    );
  }
}

function templateSchemeValue(
  field: FieldDef,
  schemeAnno: SecurityAnnotation,
  fieldExample: any,
  additionalContext?: TemplateValueContext,
): string {
  if (field.Type.Type.toString() === "array") {
    const itemPlaceholder = getSecurityExample(
      fieldExample,
      field,
      schemeAnno.SecType,
      schemeAnno.SubType,
    );
    return templateValue(field, [itemPlaceholder], false, additionalContext);
  }
  return templateSecurityFieldValue(
    field,
    schemeAnno,
    false,
    fieldExample,
    additionalContext,
  );
}

// Override common stub — C# returns SecurityInit instead of string
// @ts-ignore
function getGlobalUsageSecurity(local: UsageContext): SecurityInit | undefined {
  return getGlobalSecurity(
    { index: local.Operation?.HoistedSecurityConfig?.Fields?.[0]?.Index },
    {
      isTest: local.Test && true,
      usageContext: local,
    },
  );
}

// When op-level security was hoisted and it is a strict subset of global security,
// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const securityClass = getGlobalSecurityClass();
  const fieldList = fields.map((f) =>
    securityClass
      ? `<see cref="${securityClass}.${sanitizeFieldName(f.Name)}"/>`
      : `<c>${sanitizeFieldName(f.Name)}</c>`,
  );

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set in the security parameter when initializing the SDK.`
      : `If set, this operation will use ${items} from the global security.`;

  if (fieldList.length === 1) {
    return remark(fieldList[0]);
  }

  // Hoisted strict subsets are always OR alternatives (whether security classes or option wrappers).
  // When global security is a flattened composite scheme (AND), then for an hoisted op-level security
  // to be a subset it would have to be the same scheme (not a strict subset), so no remark is needed.

  const last = fieldList.pop();
  if (fieldList.length === 1) {
    return remark(`either ${fieldList[0]} or ${last}`);
  }

  return remark(`one of ${fieldList.join(", ")}, or ${last}`);
}
