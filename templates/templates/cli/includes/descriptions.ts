// =============================================================================
// CLI Description Generation
// Generates Short/Long descriptions for commands, root, and group nodes.
// =============================================================================

/**
 * Build a short description from comments, with configurable sentence splitting and fallback.
 * Shared logic for command, root, and group short descriptions.
 */
function buildShortDescription(
  comments: { Summary?: string; Description?: string } | undefined,
  fallback: string,
  splitRegex: RegExp = /[.\n]/,
): string {
  if (comments?.Summary) {
    const summary = comments.Summary.trim();
    if (summary) {
      return `"${escapeGoString(summary)}"`;
    }
  }

  if (comments?.Description) {
    const desc = comments.Description.trim();
    if (desc) {
      const firstPart = desc.split(splitRegex)[0].trim();
      if (firstPart) {
        return `"${escapeGoString(firstPart)}"`;
      }
    }
  }

  return `"${escapeGoString(fallback)}"`;
}

/**
 * Get the Short description for a command from OpenAPI.
 * Uses summary, falls back to first sentence of description, then humanized operationId.
 */
function templateCmdShort(op: Operation): string {
  const pascal = caser().ToPascal(op.GetID());
  const humanized = pascal.replace(/([a-z])([A-Z])/g, "$1 $2");
  return buildShortDescription(op.Comments, humanized, /[.!?\n]/);
}
registerTemplateFunc("templateCmdShort", templateCmdShort);

/**
 * Get the Long description for a command from OpenAPI.
 * Uses description, adds deprecation notice if applicable.
 */
function templateCmdLong(op: Operation): string {
  let desc = "";

  if (op.Comments?.Description) {
    desc = op.Comments.Description.trim();
  } else if (op.Comments?.Summary) {
    desc = op.Comments.Summary.trim();
  }

  // Add deprecation notice
  if (op.Comments?.Deprecated) {
    const deprecationMsg =
      op.Comments?.DeprecationMessage || "This operation is deprecated.";
    const deprecationNotice = `DEPRECATED: ${deprecationMsg}`;
    desc = desc ? `${deprecationNotice}\n\n${desc}` : deprecationNotice;
  }

  if (!desc) {
    // Fallback: humanize the operation ID (e.g., "listResourcesByAssetIds" → "List Resources By Asset Ids")
    const pascal = caser().ToPascal(op.GetID());
    desc = pascal.replace(/([a-z])([A-Z])/g, "$1 $2");
  }

  // Add binary response guidance
  if (hasBinaryResponse(op)) {
    const binaryNotice = isOnlyBinaryResponse(op)
      ? `This operation returns binary data. Use --output-file <path> to save to a file, or pipe the output to another command.`
      : `This operation may return binary data. Use --output-file <path> to save binary responses to a file.`;
    desc = desc ? `${desc}\n\n${binaryNotice}` : binaryNotice;
  }

  const streamNotice = cliOperationStreamHelp(op);
  if (streamNotice) {
    desc = desc ? `${desc}\n\n${streamNotice}` : streamNotice;
  }

  return `"${escapeGoString(desc)}"`;
}
registerTemplateFunc("templateCmdLong", templateCmdLong);

// =============================================================================
// Per-Command Examples
// =============================================================================

/**
 * Get an example value for a field. Checks pre-calculated example, field-level
 * example, defaults, then falls back to type-appropriate placeholders.
 */
interface CLIExampleValue {
  Value: string;
  SynthesizedAnglePlaceholder: boolean;
}

function isGeneratorDefaultExample(example: any): boolean {
  try {
    return (
      typeof example?.Name === "function" &&
      String(example.Name()).startsWith("speakeasy-default-")
    );
  } catch (_err) {
    return false;
  }
}

function cliExampleHasAnglePlaceholder(value: any, depth = 0): boolean {
  if (depth > 8 || value === undefined || value === null) return false;
  if (typeof value === "string") return /<[^<>]+>/.test(value);
  if (Array.isArray(value)) {
    return value.some((item) => cliExampleHasAnglePlaceholder(item, depth + 1));
  }
  if (typeof value === "object") {
    return Object.values(value).some((item) =>
      cliExampleHasAnglePlaceholder(item, depth + 1),
    );
  }
  return false;
}

function getCLIExampleValue(
  field: FieldDef,
  resolvedBodyExample?: Record<string, any>,
  resolvedBodyExampleIsGenerated = false,
): CLIExampleValue {
  // Priority 1: Pre-resolved body example (parsed from operation-level examples)
  if (resolvedBodyExample) {
    const key = originalFieldName(field);
    const val = resolvedBodyExample[key];
    if (val !== undefined && val !== null) {
      if (typeof val === "object") {
        return {
          Value: `'${JSON.stringify(val)}'`,
          SynthesizedAnglePlaceholder:
            resolvedBodyExampleIsGenerated &&
            cliExampleHasAnglePlaceholder(val),
        };
      }
      return {
        Value: String(val),
        SynthesizedAnglePlaceholder:
          resolvedBodyExampleIsGenerated && cliExampleHasAnglePlaceholder(val),
      };
    }
  }

  // Priority 2: Field-level examples (from preCalculateExamples pipeline)
  // @ts-ignore — Example is a dynamic Go proxy field not in TS type defs
  const fieldExample = field.Example;
  if (fieldExample?.Value !== undefined && fieldExample?.Value !== null) {
    const value = String(fieldExample.Value);
    return {
      Value: value,
      SynthesizedAnglePlaceholder:
        isGeneratorDefaultExample(fieldExample) &&
        cliExampleHasAnglePlaceholder(value),
    };
  }

  // Priority 3: Field defaults
  if (field.Default?.Value !== undefined && field.Default?.Value !== null) {
    return {
      Value: String(field.Default.Value),
      SynthesizedAnglePlaceholder: false,
    };
  }

  const typeDef = field.Type;

  // Enum: use first value
  if (typeDef.Type?.toString() === "enum" && typeDef.Enum?.Values?.length > 0) {
    return {
      Value: String(typeDef.Enum.Values[0]),
      SynthesizedAnglePlaceholder: false,
    };
  }

  // Type-appropriate placeholders
  switch (typeDef.Type?.toString()) {
    case "boolean":
      return { Value: "true", SynthesizedAnglePlaceholder: false };
    case "integer":
    case "int32":
      return { Value: "123", SynthesizedAnglePlaceholder: false };
    case "number":
    case "float32":
      return { Value: "3.14", SynthesizedAnglePlaceholder: false };
    default:
      return { Value: "<value>", SynthesizedAnglePlaceholder: true };
  }
}

/**
 * Render one "--flag value" example token. Boolean values are shown inline
 * ("--flag=false"): pflag parses "--flag false" as "--flag" plus a stray
 * positional, so a spaced example would teach an invocation that generated
 * commands reject (and that would set the flag to true).
 */
function exampleFlagPart(flagName: string, val: string): string {
  if (val === "true" || val === "false") {
    return `--${flagName}=${val}`;
  }
  return `--${flagName} ${val}`;
}

/**
 * Generate a Cobra Example string showing a CLI invocation with flags.
 * Uses the operation's required fields to build a realistic example.
 */
function templateCmdExample(op: Operation): string {
  const cliName = sanitizeCliName();
  const groupName = context.Local?.GroupCommandName || "";
  const promoted = context.Local?.PromotedToParent === true;
  // Use OpCommandName if available (may be de-stuttered), else compute from operation ID
  const cmdName =
    context.Local?.OpCommandName || sanitizeCLICommand(op.GetID());

  const parts: {
    Value: string;
    SynthesizedPlaceholder: boolean;
  }[] = [];
  const pushPart = (value: string, synthesizedPlaceholder = false) => {
    parts.push({
      Value: value,
      SynthesizedPlaceholder: synthesizedPlaceholder,
    });
  };

  // Build the base command string
  let fullCmd: string;
  if (promoted) {
    // Promoted operation: invoked as just "<cli> <group>"
    fullCmd = groupName ? `${cliName} ${groupName}` : `${cliName} ${cmdName}`;
  } else {
    fullCmd = groupName
      ? `${cliName} ${groupName} ${cmdName}`
      : `${cliName} ${cmdName}`;
  }

  if (!op.Request) {
    return `"  ${escapeGoString(fullCmd)}"`;
  }

  // Pre-resolve body examples from operation-level examples (if available).
  // Each named request example becomes a candidate invocation; the first one
  // also seeds field-level values (spec-provided UUIDs, realistic names, etc.).
  const bodyExamples: {
    Value: Record<string, any>;
    Synthesized: boolean;
  }[] = [];
  try {
    const examples = op.Request.Examples;
    if (examples && examples.length > 0) {
      for (const example of examples) {
        const value = getExampleValue(example);
        if (
          value &&
          typeof value === "object" &&
          Object.keys(value).length > 0
        ) {
          bodyExamples.push({
            Value: value,
            Synthesized:
              isGeneratorDefaultExample(example) &&
              cliExampleHasAnglePlaceholder(value),
          });
        }
        if (bodyExamples.length >= 3) break;
      }
    }
  } catch (e) {
    // Fall through — use field-level examples and placeholders
  }
  const bodyExample: Record<string, any> | undefined = bodyExamples[0]?.Value;
  const bodyExampleIsGenerated = bodyExamples[0]?.Synthesized || false;

  // Collect required param flags
  if (op.Request.Params) {
    const allParams = [
      ...(op.Request.Params.PathParams || []),
      ...(op.Request.Params.QueryParams || []),
      ...(op.Request.Params.HeaderParams || []),
    ];
    for (const param of allParams) {
      if (param.Field.Optional) continue;
      if (param.Field.Const) continue;
      const flagName = sanitizeFlagNameWithReserved(param.Field.Name);
      // Try param-level examples first, then fall back
      try {
        if (param.Examples?.length > 0) {
          const pex = findExampleByName(param.Examples, "");
          if (pex) {
            const pval = getExampleValue(pex);
            if (pval !== undefined && pval !== null) {
              pushPart(
                exampleFlagPart(flagName, String(pval)),
                isGeneratorDefaultExample(pex) &&
                  cliExampleHasAnglePlaceholder(pval),
              );
              continue;
            }
          }
        }
      } catch (e) {
        // Fall through
      }
      const val = getCLIExampleValue(param.Field);
      pushPart(
        exampleFlagPart(flagName, val.Value),
        val.SynthesizedAnglePlaceholder,
      );
    }
  }

  // Add body flags to example. Non-expandable bodies (e.g. unions) render one
  // full invocation per named request example so each variant is runnable.
  let bodyJSONVariants: {
    Value: string;
    SynthesizedPlaceholder: boolean;
  }[] = [];
  if (op.Request.IsRequestBody) {
    if (isRequestBodyExpandable(op)) {
      // Expandable body: show individual field flags with spec examples
      const bodyField = op.Request.RequestBody;
      if (bodyField?.Type?.Fields) {
        for (const field of bodyField.Type.Fields) {
          if (field.Optional) continue;
          if (field.Const) continue;
          const flagName = sanitizeFlagNameWithReserved(field.Name);
          const val = getCLIExampleValue(
            field,
            bodyExample,
            bodyExampleIsGenerated,
          );
          pushPart(
            exampleFlagPart(flagName, val.Value),
            val.SynthesizedAnglePlaceholder,
          );
        }
      }
    } else if (op.SerializationMethod?.toString() === "multipart") {
      // Multipart bodies register per-part flags (file parts take paths);
      // no whole-body JSON flag exists on these commands, so a JSON example
      // would document an invocation that cannot run.
      const bodyField = op.Request.RequestBody;
      for (const field of bodyField?.Type?.Fields || []) {
        if (field.Optional || field.Const) continue;
        const flagName = sanitizeFlagNameWithReserved(field.Name);
        if (isMultipartFileField(field)) {
          pushPart(`--${flagName} <path/to/file>`, true);
        } else {
          const val = getCLIExampleValue(
            field,
            bodyExample,
            bodyExampleIsGenerated,
          );
          pushPart(
            exampleFlagPart(flagName, val.Value),
            val.SynthesizedAnglePlaceholder,
          );
        }
      }
    } else {
      // Non-expandable body: show realistic JSON examples if available
      const bodyField = op.Request.RequestBody;
      const flagName = getRequestBodyFlagName(bodyField);
      if (bodyExamples.length > 0) {
        bodyJSONVariants = bodyExamples.map((ex) => ({
          Value: `--${flagName} '${JSON.stringify(ex.Value)}'`,
          SynthesizedPlaceholder: ex.Synthesized,
        }));
      } else {
        pushPart(`--${flagName} '{"key": "value"}'`, true);
      }
    }
  } else if (op.Request.RequestBody) {
    // Mixed params + body: find the body field and show its expanded sub-fields
    const reqFields = op.Request.Field?.Type?.Fields || [];
    const bodyField = reqFields.find(
      (f: FieldDef) => f.Annotations?.Has("request"),
    );
    if (bodyField && shouldExpandNestedField(bodyField)) {
      const bodyPrefix = getBodyFlagPrefix(bodyField, reqFields, "");
      const bodySubFields = bodyField.Type.Fields || [];
      for (const subField of bodySubFields) {
        if (subField.Optional) continue;
        if (subField.Const) continue;
        const subFlagName = bodyPrefix
          ? `${bodyPrefix}.${sanitizeFlagNameWithReserved(subField.Name)}`
          : sanitizeFlagNameWithReserved(subField.Name);
        const val = getCLIExampleValue(
          subField,
          bodyExample,
          bodyExampleIsGenerated,
        );
        pushPart(
          exampleFlagPart(subFlagName, val.Value),
          val.SynthesizedAnglePlaceholder,
        );
      }
    } else if (bodyField && !isMultipartMixedOp(op)) {
      // Non-expandable body (e.g. a union): fall back to whole-body JSON.
      // Prefer the plain --body flag when the command registers one.
      const flagName = templateHasBodyFlag(op)
        ? "body"
        : getRequestBodyFlagName(bodyField);
      if (bodyExamples.length > 0) {
        bodyJSONVariants = bodyExamples.map((ex) => ({
          Value: `--${flagName} '${JSON.stringify(ex.Value)}'`,
          SynthesizedPlaceholder: ex.Synthesized,
        }));
      } else {
        pushPart(`--${flagName} '{"key": "value"}'`, true);
      }
    } else if (bodyField && isMultipartMixedOp(op)) {
      // Multipart bodies have no whole-body JSON flag. Keep examples runnable
      // by rendering the required part flags with the same names as metadata.
      const globalFlags = getGlobalFlagNames();
      for (const subField of bodyField.Type.Fields || []) {
        if (subField.Optional || subField.Const) continue;
        const flagName = sanitizeFlagNameWithReserved(subField.Name);
        if (globalFlags.has(flagName)) continue;

        const fieldTypeStr = subField.Type?.Type?.toString() || "";
        const isFileArray =
          fieldTypeStr === "array" &&
          subField.Type?.ItemType?.Type?.toString() === "class" &&
          (() => {
            const ann = subField.Annotations?.Get("multipartForm");
            return Boolean(ann && isMultipartFormAnnotation(ann) && ann.File);
          })();
        if (isMultipartFileField(subField) || isFileArray) {
          pushPart(exampleFlagPart(flagName, "./path/to/file"));
        } else {
          const val = getCLIExampleValue(
            subField,
            bodyExample,
            bodyExampleIsGenerated,
          );
          pushPart(
            exampleFlagPart(flagName, val.Value),
            val.SynthesizedAnglePlaceholder,
          );
        }
      }
    }
  }

  // Limit to first 5 flags to keep the example readable
  const flagParts = parts.slice(0, 5);
  const flagStr =
    flagParts.length > 0
      ? " " + flagParts.map((part) => part.Value).join(" ")
      : "";
  const baseLine = fullCmd + flagStr;

  const compact = helpStyle() === "compact";
  if (compact && flagParts.some((part) => part.SynthesizedPlaceholder)) {
    return `""`;
  }

  // Distinct named request examples can serialize to identical payloads;
  // documenting the same invocation twice reads as an error.
  {
    const seenVariantValues = new Set<string>();
    bodyJSONVariants = bodyJSONVariants.filter((variant) => {
      if (seenVariantValues.has(variant.Value)) return false;
      seenVariantValues.add(variant.Value);
      return true;
    });
  }
  if (bodyJSONVariants.length > 0) {
    const visibleVariants = compact
      ? bodyJSONVariants.filter((variant) => !variant.SynthesizedPlaceholder)
      : bodyJSONVariants;
    if (visibleVariants.length === 0) return `""`;
    // One runnable line per body example (shared params prefix on each)
    const lines = visibleVariants.map((v) => `  ${baseLine} ${v.Value}`);
    return `"${escapeGoString(lines.join("\n"))}"`;
  }

  return `"  ${escapeGoString(baseLine)}"`;
}
registerTemplateFunc("templateCmdExample", templateCmdExample);

// =============================================================================
// Root Command Descriptions
// =============================================================================

/**
 * Get the CLI display name used in root fallback descriptions.
 * Prefer the configured cliName exactly as provided (e.g. "example-cli").
 */
function getHumanCliName(): string {
  return sanitizeCliName();
}

/**
 * Generate the Short description for the root command.
 * Uses the OpenAPI info.title or info.summary.
 */
function templateRootShort(): string {
  const comments = context.Global.AST.MainSDK.Comments;
  return buildShortDescription(
    comments,
    `${getHumanCliName()} command-line interface`,
  );
}
registerTemplateFunc("templateRootShort", templateRootShort);

/**
 * Generate the Long description for the root command.
 * Uses the OpenAPI info.description.
 */
function templateRootLong(): string {
  const comments = context.Global.AST.MainSDK.Comments;

  let desc = "";

  if (comments?.Description) {
    desc = comments.Description.trim();
  } else if (comments?.Summary) {
    desc = comments.Summary.trim();
  }

  if (!desc) {
    desc = `Command-line interface for ${getHumanCliName()}`;
  }

  return `"${escapeGoString(desc)}"`;
}
registerTemplateFunc("templateRootLong", templateRootLong);

// =============================================================================
// Group (SubSDK) Command Descriptions
// =============================================================================

/**
 * Generate the Short description for a group (subroot) command.
 * Uses the SubSDK's Comments from OpenAPI tag descriptions.
 */
function templateGroupShort(sdk: SDK): string {
  const name = getSDKGroupName(sdk) || "commands";
  return buildShortDescription(
    sdk?.Comments,
    `Operations for ${sanitizeCLICommand(name)}`,
  );
}
registerTemplateFunc("templateGroupShort", templateGroupShort);

/**
 * Generate the Long description for a group (subroot) command.
 * Uses the SubSDK's Comments from OpenAPI tag descriptions.
 */
function templateGroupLong(sdk: SDK): string {
  const comments = sdk?.Comments;

  let desc = "";

  if (comments?.Description) {
    desc = comments.Description.trim();
  } else if (comments?.Summary) {
    desc = comments.Summary.trim();
  }

  if (!desc) {
    const name = getSDKGroupName(sdk) || "commands";
    desc = `Operations for ${sanitizeCLICommand(name)}`;
  }

  return `"${escapeGoString(desc)}"`;
}
registerTemplateFunc("templateGroupLong", templateGroupLong);
