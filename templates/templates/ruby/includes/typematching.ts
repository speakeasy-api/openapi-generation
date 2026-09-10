// Ruby-specific override: lower "any" type priority in union matching
// so that more specific types (e.g. SimpleObject) are preferred over catch-all "any".
// We override matchTypeWithExample (the outer function) to post-process
// results from matchTypeWithExampleInner, avoiding function hoisting issues.
// @ts-ignore
function matchTypeWithExample(
  type: TypeDef,
  example: any,
  additionalContext?: { operation?: Operation; exampleName?: string },
): MatchResult | undefined {
  CURRENT_MATCH_OP_ID = additionalContext?.operation?.OriginalID ?? "";
  CURRENT_MATCH_EXAMPLE_NAME = additionalContext?.exampleName ?? "";

  debugMatchTypeLogging(
    `Attempting to match type ${type.Name} (${type.Type})`,
    example,
  );

  const res = matchTypeWithExampleInner(type, example, additionalContext);

  // Bare "any" (no associated types) should have lowest priority
  // so that specific class types win in union matching
  if (
    res &&
    isTypeDefType(res.type, "any") &&
    res.type.AssociatedTypes.length == 0
  ) {
    debugMatchTypeLogging(
      `MATCH (lowered rating) - example matches ${res.type.Name} (${res.type.Type}) with rating 0.01`,
      example,
    );
    return { type: res.type, rating: 0.01 };
  }

  if (res) {
    debugMatchTypeLogging(
      `MATCH - example matches ${res.type.Name} (${res.type.Type}) with rating ${res.rating}`,
      example,
    );
  } else {
    debugMatchTypeLogging(`NO MATCH - for example`, example);
  }
  return res;
}
