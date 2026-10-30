import fs from "node:fs";
import { Client } from "@modelcontextprotocol/client";
import {
  completable,
  InMemoryTransport,
  McpServer,
} from "@modelcontextprotocol/server";
import { test, expect, vi } from "vitest";
import { z } from "zod/v3";
import { SDKCore } from "../core.js";
import { createConsoleLogger } from "../mcp-server/console-logger.js";
import { createRegisterTool } from "../mcp-server/tools.js";
import { createRegisterPrompt } from "../mcp-server/prompts.js";

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

test("MCP v2 advertises Zod 3 inputs and preserves validation and transforms", async () => {
  const server = new McpServer({ name: "schema-test", version: "1.0.0" });
  const client = new Client({ name: "schema-client", version: "1.0.0" });
  const sdk = new SDKCore();
  const logger = createConsoleLogger("error");
  const registerTool = createRegisterTool(logger, server, sdk, new Set());
  const handler = vi.fn(
    (args: { count: number; label: string }, signal: AbortSignal) => {
      expect(signal).toBeInstanceOf(AbortSignal);
      return {
        content: [{ type: "text" as const, text: JSON.stringify(args) }],
      };
    },
  );
  registerTool({
    name: "convert-input",
    description: "Validate and transform an input",
    args: {
      count: z.string().regex(/^\d+$/).transform(Number),
      label: z.string().default("example"),
      choice: z.union([z.string(), z.instanceof(Date)]).optional(),
    },
    tool: (_sdk, args, ctx) => handler(args, ctx.mcpReq.signal),
  });
  registerTool({
    name: "no-input",
    description: "Accept no arguments",
    tool: (_sdk, ctx) => ({
      content: [{ type: "text", text: String(ctx.mcpReq.signal.aborted) }],
    }),
  });
  const registerPrompt = createRegisterPrompt(logger, server, sdk, new Set());
  registerPrompt({
    name: "greeting",
    args: { name: z.string() },
    prompt: (_sdk, args) => ({
      messages: [{ role: "user", content: { type: "text", text: args.name } }],
    }),
  });
  const [clientTransport, serverTransport] =
    InMemoryTransport.createLinkedPair();
  try {
    await server.connect(serverTransport);
    await client.connect(clientTransport);
    const { tools } = await client.listTools();
    expect(
      tools.find((tool) => tool.name === "convert-input")?.inputSchema,
    ).toMatchObject({
      type: "object",
      properties: {
        count: { type: "string" },
        label: { default: "example" },
        choice: { anyOf: [{ type: "string" }] },
      },
      required: ["count"],
    });
    expect(
      await client.callTool({
        name: "convert-input",
        arguments: { count: "42" },
      }),
    ).toMatchObject({
      content: [
        { type: "text", text: JSON.stringify({ count: 42, label: "example" }) },
      ],
    });
    expect(
      await client.callTool({
        name: "convert-input",
        arguments: { count: "invalid" },
      }),
    ).toMatchObject({ isError: true });
    expect(handler).toHaveBeenCalledTimes(1);
    expect(await client.callTool({ name: "no-input" })).toMatchObject({
      content: [{ type: "text", text: "false" }],
    });
    expect((await client.listPrompts()).prompts).toMatchObject([
      { name: "greeting", arguments: [{ name: "name", required: true }] },
    ]);
    expect(
      await client.getPrompt({
        name: "greeting",
        arguments: { name: "example" },
      }),
    ).toMatchObject({
      messages: [{ role: "user", content: { type: "text", text: "example" } }],
    });
  } finally {
    await client.close();
    await server.close();
  }
});

test("MCP v2 serializes date and bigint defaults without changing runtime values", async () => {
  const server = new McpServer({ name: "defaults-test", version: "1.0.0" });
  const client = new Client({ name: "defaults-client", version: "1.0.0" });
  const date = new Date("2025-01-01T00:00:00Z");
  const id = BigInt("12345678901234567890");
  const handler = vi.fn((args: { value: { date: Date; id: bigint } }) => {
    expect(args.value).toEqual({ date, id });
    return { content: [{ type: "text" as const, text: "defaults preserved" }] };
  });
  createRegisterTool(
    createConsoleLogger("error"),
    server,
    new SDKCore(),
    new Set(),
  )({
    name: "nested-defaults",
    description: "Preserve nested default values",
    args: {
      value: z.object({ date: z.date(), id: z.bigint() }).default({ date, id }),
    },
    tool: (_sdk, args) => handler(args),
  });
  const [clientTransport, serverTransport] =
    InMemoryTransport.createLinkedPair();
  try {
    await server.connect(serverTransport);
    await client.connect(clientTransport);
    const { tools } = await client.listTools();
    expect(tools[0]?.inputSchema).toMatchObject({
      properties: {
        value: { default: { date: date.toISOString(), id: id.toString() } },
      },
    });
    expect(
      await client.callTool({ name: "nested-defaults", arguments: {} }),
    ).toMatchObject({
      content: [{ type: "text", text: "defaults preserved" }],
    });
    expect(handler).toHaveBeenCalledOnce();
  } finally {
    await client.close();
    await server.close();
  }
});

test("prompt completion preserves input validation and transforms", async () => {
  const server = new McpServer({ name: "prompt-test", version: "1.0.0" });
  const client = new Client({ name: "prompt-client", version: "1.0.0" });
  const complete = vi.fn((value: string) =>
    ["typescript", "python"].filter((language) => language.startsWith(value)),
  );
  const handler = vi.fn(
    (args: {
      language: string;
      count: string;
      label?: string | undefined;
    }) => ({
      messages: [
        {
          role: "user" as const,
          content: { type: "text" as const, text: JSON.stringify(args) },
        },
      ],
    }),
  );
  createRegisterPrompt(
    createConsoleLogger("error"),
    server,
    new SDKCore(),
    new Set(),
  )({
    name: "review",
    args: {
      language: completable(z.string().describe("Language"), complete),
      count: z
        .string()
        .regex(/^\d+$/)
        .transform((value) => `count:${value}`),
      label: completable(z.string(), complete).optional(),
    },
    prompt: (_sdk, args) => handler(args),
  });
  const [clientTransport, serverTransport] =
    InMemoryTransport.createLinkedPair();
  try {
    await server.connect(serverTransport);
    await client.connect(clientTransport);
    expect((await client.listPrompts()).prompts[0]?.arguments).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          name: "language",
          description: "Language",
          required: true,
        }),
        expect.objectContaining({ name: "label", required: false }),
      ]),
    );
    expect(
      await client.complete({
        ref: { type: "ref/prompt", name: "review" },
        argument: { name: "language", value: "type" },
        context: { arguments: { count: "7" } },
      }),
    ).toMatchObject({ completion: { values: ["typescript"] } });
    expect(complete).toHaveBeenCalledWith("type", {
      arguments: { count: "7" },
    });
    expect(
      await client.complete({
        ref: { type: "ref/prompt", name: "review" },
        argument: { name: "label", value: "py" },
      }),
    ).toMatchObject({ completion: { values: ["python"] } });
    expect(
      await client.getPrompt({
        name: "review",
        arguments: { language: "typescript", count: "7" },
      }),
    ).toMatchObject({
      messages: [
        {
          content: {
            text: JSON.stringify({
              language: "typescript",
              count: "count:7",
            }),
          },
        },
      ],
    });
    await expect(
      client.getPrompt({
        name: "review",
        arguments: { language: "typescript", count: "invalid" },
      }),
    ).rejects.toThrow();
    expect(handler).toHaveBeenCalledTimes(1);
  } finally {
    await client.close();
    await server.close();
  }
});
