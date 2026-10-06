/* Isomorphic Zod helpers to support both v3, v4 and v4-mini */

function templateSmartUnionFunc(): string {
  return /*ts*/ `
/**
 * Smart union parser that tries all schemas and returns the best match
 * based on the number of populated fields.
 */
export function smartUnion<
  Options extends readonly [${z.ZodType()}, ${z.ZodType()}, ...${z.ZodType()}[]],
>(
  options: Options,
): ${z.ZodType("z.output<Options[number]>", "z.input<Options[number]>")} {
  return ${z.transform(
    z.unknown(),
    `(input, ctx) => {
    const candidates: Candidate[] = [];
    const errors: ${isZodV4() ? "z.core.$ZodRawIssue" : z.zodIssue()}[][] = [];

    const parentUnrecognizedCtr = startCountingUnrecognized();
    ${
      isLaxMode()
        ? `const parentZeroDefaultCtr = startCountingDefaultToZeroValue();`
        : ""
    }

    // Filter out invalid options
    for (const option of options) {
      const unrecognizedCtr = startCountingUnrecognized();
      ${
        isLaxMode()
          ? `const zeroDefaultCtr = startCountingDefaultToZeroValue();`
          : ""
      }
      ${
        isZodV4()
          ? `const result = option._zod.run({ value: input, issues: [] }, { async: false });
      if (result instanceof Promise) {
        throw new z.core.$ZodAsyncError();
      }`
          : `const result = option.safeParse(input);`
      }
      const inexactCount = unrecognizedCtr.end();
      const zeroDefaultCount = ${isLaxMode() ? `zeroDefaultCtr.end();` : "0"};
      if (${isZodV4() ? "result.issues.length === 0" : "result.success"}) {
        candidates.push({
          data: result.${isZodV4() ? "value" : "data"},
          inexactCount,
          zeroDefaultCount,
          fieldCount: -1, // We'll count this later if needed
        });
        continue;
      }
      errors.push(result.${isZodV4() ? "issues" : "error.issues"});
    }

    // No valid options
    if (candidates.length === 0) {
      parentUnrecognizedCtr.end(0); ${
        isLaxMode() ? `parentZeroDefaultCtr.end(0);` : ""
      }
      ${z.addIssue({
        code: "invalid_union",
        input: "input",
        errors: isZodV4()
          ? "errors.map(issues => issues.map(issue => z.core.util.finalizeIssue(issue, {}, z.core.config())))"
          : "errors",
      })};
      return ${isZodV4() ? "undefined" : z.NEVER()};
    }

    let best = candidates[0]!;

    // Find the best option
    for (const candidate of candidates) {
      // Minor optimization to avoid counting fields if there's only one candidate
      if (candidates.length > 1) candidate.fieldCount = countFieldsRecursive(candidate.data);
      best = better(candidate, best);
    }

    // The cost of this union should be the cost of the best candidate not all the candidates
    parentUnrecognizedCtr.end(best.inexactCount);
    ${isLaxMode() ? `parentZeroDefaultCtr.end(best.zeroDefaultCount);` : ""}

    return best.data;
  }`,
  )} as any;
}`;
}

registerTemplateFunc("templateSmartUnionFunc", templateSmartUnionFunc);
