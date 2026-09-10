import { expect, test } from "vitest";

import { SDKCore } from "../core.js";
import {
  buildGenerationReactHookBodyAndParamsInQueryKeyQuery,
  queryKeyGenerationReactHookBodyAndParamsInQueryKey,
} from "../react-query/generationReactHookBodyAndParamsInQueryKey.core.js";
import {
  buildGenerationReactHookBodyInQueryKeyQuery,
  queryKeyGenerationReactHookBodyInQueryKey,
} from "../react-query/generationReactHookBodyInQueryKey.core.js";
import { queryKeyGenerationReactHookBodyInQueryKeyInferredType } from "../react-query/generationReactHookBodyInQueryKeyInferredType.core.js";
import { recordTest } from "./common_helpers.js";

test("React query key includes request body when opted in", () => {
  recordTest("react-query-key-includes-request-body");

  const body = { chatId: "chat-1", prompt: "hello" };

  expect(queryKeyGenerationReactHookBodyInQueryKey(body)).toEqual([
    "openapi",
    "generation",
    "reactHookBodyInQueryKey",
    body,
  ]);
});

test("React query key opt-in infers query hook type on POST", () => {
  recordTest("react-query-key-opt-in-infers-query-type");

  const body = { prompt: "hello" };

  expect(queryKeyGenerationReactHookBodyInQueryKeyInferredType(body)).toEqual([
    "openapi",
    "generation",
    "reactHookBodyInQueryKeyInferredType",
    body,
  ]);
});

test("React query key includes both parameters and request body", () => {
  recordTest("react-query-key-includes-parameters-and-request-body");

  const body = { prompt: "hello", messageTypes: ["user"] };

  expect(
    queryKeyGenerationReactHookBodyAndParamsInQueryKey({ scope: "s1" }, body),
  ).toEqual([
    "openapi",
    "generation",
    "reactHookBodyAndParamsInQueryKey",
    { scope: "s1" },
    body,
  ]);
});

test("React query builders key distinct request bodies separately", () => {
  recordTest("react-query-builders-key-distinct-bodies");

  const client = new SDKCore();
  const first = buildGenerationReactHookBodyInQueryKeyQuery(client, {
    prompt: "first",
  });
  const second = buildGenerationReactHookBodyInQueryKeyQuery(client, {
    prompt: "second",
  });

  expect(first.queryKey).toEqual(
    queryKeyGenerationReactHookBodyInQueryKey({ prompt: "first" }),
  );
  expect(first.queryKey).not.toEqual(second.queryKey);

  const withParams = buildGenerationReactHookBodyAndParamsInQueryKeyQuery(
    client,
    { prompt: "first" },
    "s1",
  );
  expect(withParams.queryKey).toEqual(
    queryKeyGenerationReactHookBodyAndParamsInQueryKey(
      { scope: "s1" },
      { prompt: "first" },
    ),
  );
});
