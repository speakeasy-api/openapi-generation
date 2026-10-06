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
    ${
      isZodV4()
        ? `const rawIssues: z.core.$ZodRawIssue[][] = options.map(() => []);
    const parseCtx = { async: false } as const;`
        : `const errors: ${z.zodIssue()}[][] = options.map(() => []);`
    }

    const parentUnrecognizedCtr = startCountingUnrecognized();
    ${
      isLaxMode()
        ? `const parentZeroDefaultCtr = startCountingDefaultToZeroValue();`
        : ""
    }

    // Filter out invalid options
    for (const [i, option] of options.entries()) {
      const unrecognizedCtr = startCountingUnrecognized();
      ${
        isLaxMode()
          ? `const zeroDefaultCtr = startCountingDefaultToZeroValue();`
          : ""
      }
      ${
        isZodV4()
          ? `const result = option._zod.run({ value: input, issues: [] }, parseCtx);
      if (result instanceof Promise) {
        throw new z.core.$ZodAsyncError();
      }`
          : `const result = option.safeParse(input);`
      }
      const inexactCount = unrecognizedCtr.end();
      const zeroDefaultCount = ${isLaxMode() ? `zeroDefaultCtr.end();` : "0"};
      if (${isZodV4() ? "result.issues.length === 0" : "result.success"}) {
        candidates.push({
          data: ${isZodV4() ? "result.value" : "result.data"},
          inexactCount,
          zeroDefaultCount,
          fieldCount: -1, // We'll count this later if needed
        });
        continue;
      }
      ${
        isZodV4()
          ? `rawIssues[i] = result.issues;`
          : `errors[i]!.push(...result.error.issues);`
      }
    }

    // No valid options
    if (candidates.length === 0) {
      parentUnrecognizedCtr.end(0); ${
        isLaxMode() ? `parentZeroDefaultCtr.end(0);` : ""
      }
      ${
        isZodV4()
          ? `const config = z.core.config();
      const errors = rawIssues.map((issues) =>
        issues.map((iss) => z.core.util.finalizeIssue(iss, parseCtx, config)),
      );`
          : ""
      }
      ${z.addIssue({
        code: "invalid_union",
        input: "input",
        errors: "errors",
      })};
      return ${z.NEVER()};
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
