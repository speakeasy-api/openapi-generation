import fs from "node:fs";
import { test, expect } from "vitest";

import { tool$requestBodiesRequestBodyPostApplicationJsonMapOfMapOfPrimitive } from "../mcp-server/tools/requestBodiesRequestBodyPostApplicationJsonMapOfMapOfPrimitive.js";
import { tool$resourceGetResource } from "../mcp-server/tools/resourceGetResource.js";
import { tool$resourceUpdateResource } from "../mcp-server/tools/resourceUpdateResource.js";
import { tool$resourceDeleteResource } from "../mcp-server/tools/resourceDeleteResource.js";

test("parametersDeepObjectQueryParamsObject should be disabled", () => {
  expect(
    fs.existsSync("src/funcs/parametersDeepObjectQueryParamsObject.ts"),
  ).toBe(true);

  expect(
    fs.existsSync(
      "src/mcp-server/tools/parametersDeepObjectQueryParamsObject.ts",
    ),
  ).toBe(false);
});

test("requestBodiesRequestBodyPostApplicationJsonMapOfMapOfPrimitive has custom description", () => {
  const { description } =
    tool$requestBodiesRequestBodyPostApplicationJsonMapOfMapOfPrimitive;

  expect(description).toContain(
    "This endpoint is used to test MCP server generation",
  );

  expect(description).not.toContain(
    "The OpenAPI description for this operation",
  );
});

test("operations have custom scopes attached", () => {
  expect(tool$resourceGetResource.scopes).toEqual(["read"]);
  expect(tool$resourceUpdateResource.scopes).toEqual(["write"]);
  expect(tool$resourceDeleteResource.scopes).toEqual(["destructive", "write"]);
});
