import { createServer } from "node:http";
import { fileURLToPath } from "node:url";
import { StdioClientTransport } from "@modelcontextprotocol/client/stdio";
import { serveStdio } from "@modelcontextprotocol/server/stdio";
import {
  Client,
  StreamableHTTPClientTransport,
} from "@modelcontextprotocol/client";
import {
  completable,
  createMcpHandler,
  InMemoryTransport,
  McpServer,
  ResourceTemplate,
} from "@modelcontextprotocol/server";
import { describe, expect, test, vi } from "vitest";
import { z } from "zod";
import { bigint } from "../types/bigint.js";
import { mcpInputSchema } from "../mcp-server/shared.js";
import { SDKCore } from "../core.js";
import { createConsoleLogger } from "../mcp-server/console-logger.js";
import { createRegisterPrompt } from "../mcp-server/prompts.js";
import {
  createRegisterResource,
  createRegisterResourceTemplate,
} from "../mcp-server/resources.js";
import { MCPScope } from "../mcp-server/scopes.js";
import {
  createRegisterTool,
  registerDynamicTools,
} from "../mcp-server/tools.js";

const annotations = {
  title: "Example",
  readOnlyHint: true,
  destructiveHint: false,
  idempotentHint: true,
  openWorldHint: false,
};

function makeHarness(dynamic: boolean) {
  const server = new McpServer({ name: "registration-test", version: "1.0.0" });
  const logger = createConsoleLogger("error");
  const getSDK = () => new SDKCore();
  const scopes = new Set<MCPScope>();
  const [tool, , toolMap] = createRegisterTool(
    logger,
    server,
    getSDK,
    scopes,
    undefined,
    dynamic,
  );
  const stub = vi
    .fn()
    .mockResolvedValue({ content: [{ type: "text", text: "stub-result" }] });
  const noArgsStub = vi
    .fn()
    .mockResolvedValue({ content: [{ type: "text", text: "no-args-result" }] });
  tool({
    name: "stub-tool",
    description: "Accept a nested request",
    annotations,
    args: { request: z.object({ name: z.string() }) },
    tool: stub,
  });
  tool({
    name: "no-args-tool",
    description: "Accept no arguments",
    annotations,
    tool: noArgsStub,
  });
  tool({
    name: "transform-tool",
    description: "Transform input values",
    annotations,
    args: {
      amount: z.union([z.string(), z.number()]).transform(Number),
      price: z.string().transform((v) => v),
      when: z
        .union([z.date(), z.string().transform((v) => new Date(v))])
        .transform((v) => v.toISOString()),
    },
    tool: (_sdk, args) => ({
      content: [{ type: "text", text: JSON.stringify(args) }],
    }),
  });
  tool({
    name: "bigint-default-tool",
    description: "Preserve nested bigint defaults",
    annotations,
    args: {
      request: z
        .object({
          amount: bigint().default(12345678901234567890n),
          amounts: z.array(bigint()).default([12345678901234567890n]),
        })
        .default({
          amount: 12345678901234567890n,
          amounts: [12345678901234567890n],
        }),
    },
    tool: (_sdk, args) => ({
      content: [
        {
          type: "text",
          text: JSON.stringify(args, (_key, value: unknown) =>
            typeof value === "bigint" ? value.toString() : value,
          ),
        },
      ],
    }),
  });
  if (dynamic) registerDynamicTools(logger, server, getSDK, toolMap, scopes);
  return { server, logger, getSDK, scopes, stub, noArgsStub };
}

async function withClient(
  server: McpServer,
  run: (client: Client) => Promise<void>,
) {
  const client = new Client({ name: "registration-client", version: "1.0.0" });
  const [clientTransport, serverTransport] =
    InMemoryTransport.createLinkedPair();
  try {
    await server.connect(serverTransport);
    await client.connect(clientTransport);
    await run(client);
  } finally {
    await client.close();
    await server.close();
  }
}

describe("MCP registration", () => {
  test("lists and invokes tools through the v2 protocol", async () => {
    const { server, stub, noArgsStub } = makeHarness(false);
    await withClient(server, async (client) => {
      const { tools } = await client.listTools();
      expect(tools.map((tool) => tool.name)).toEqual(
        expect.arrayContaining(["stub-tool", "no-args-tool", "transform-tool"]),
      );
      const result = await client.callTool({
        name: "stub-tool",
        arguments: { request: { name: "Example" } },
      });
      expect(result.isError).toBeFalsy();
      expect(stub.mock.calls[0]?.[1]).toEqual({ request: { name: "Example" } });
      expect(stub.mock.calls[0]?.[2].mcpReq.signal).toBeInstanceOf(AbortSignal);
      const invalid = await client.callTool({
        name: "stub-tool",
        arguments: { request: { name: 42 } },
      });
      expect(invalid.isError).toBe(true);
      expect(stub).toHaveBeenCalledTimes(1);
      await client.callTool({ name: "no-args-tool" });
      expect(noArgsStub.mock.calls[0]?.[1].mcpReq.signal).toBeInstanceOf(
        AbortSignal,
      );
      const transformed = await client.callTool({
        name: "transform-tool",
        arguments: {
          amount: "12.5",
          price: "8",
          when: "2026-01-01T00:00:00.000Z",
        },
      });
      expect(transformed.content).toEqual([
        {
          type: "text",
          text: JSON.stringify({
            amount: 12.5,
            price: "8",
            when: "2026-01-01T00:00:00.000Z",
          }),
        },
      ]);
    });
  });

  test("registers prompts and static and templated resources", async () => {
    const { server, logger, getSDK, scopes } = makeHarness(false);
    const prompt = createRegisterPrompt(logger, server, getSDK, scopes);
    prompt({
      name: "greeting",
      args: { name: z.string() },
      prompt: (_sdk, args, ctx) => {
        expect(ctx.mcpReq.signal).toBeInstanceOf(AbortSignal);
        return {
          messages: [
            { role: "user", content: { type: "text", text: args.name } },
          ],
        };
      },
    });
    prompt({
      name: "empty",
      prompt: (_sdk, ctx) => {
        expect(ctx.mcpReq.signal).toBeInstanceOf(AbortSignal);
        return { messages: [] };
      },
    });
    createRegisterResource(
      logger,
      server,
      getSDK,
      scopes,
    )({
      name: "static",
      resource: "test://static",
      read: (_sdk, uri, ctx) => {
        expect(ctx.mcpReq.signal).toBeInstanceOf(AbortSignal);
        return { contents: [{ uri: uri.href, text: "static" }] };
      },
    });
    createRegisterResourceTemplate(
      logger,
      server,
      getSDK,
      scopes,
    )({
      name: "template",
      description: "Named resource",
      resource: new ResourceTemplate("test://items/{name}", {
        list: undefined,
      }),
      read: (_sdk, uri, vars, ctx) => {
        expect(ctx.mcpReq.signal).toBeInstanceOf(AbortSignal);
        return { contents: [{ uri: uri.href, text: String(vars["name"]) }] };
      },
    });
    await withClient(server, async (client) => {
      expect((await client.listPrompts()).prompts.map((p) => p.name)).toEqual([
        "greeting",
        "empty",
      ]);
      expect(
        (
          await client.getPrompt({
            name: "greeting",
            arguments: { name: "Example" },
          })
        ).messages[0]?.content,
      ).toEqual({ type: "text", text: "Example" });
      expect((await client.getPrompt({ name: "empty" })).messages).toEqual([]);
      expect((await client.listResources()).resources[0]?.uri).toBe(
        "test://static",
      );
      expect(
        (await client.listResourceTemplates()).resourceTemplates[0]
          ?.uriTemplate,
      ).toBe("test://items/{name}");
      expect(
        (await client.readResource({ uri: "test://static" })).contents,
      ).toEqual([{ uri: "test://static", text: "static" }]);
      expect(
        (await client.readResource({ uri: "test://items/example" })).contents,
      ).toEqual([{ uri: "test://items/example", text: "example" }]);
    });
  });
});

describe("dynamic MCP tools", () => {
  test("advertises schemas and preserves nested arguments during execution", async () => {
    const { server, stub, noArgsStub } = makeHarness(true);
    await withClient(server, async (client) => {
      const { tools } = await client.listTools();
      const execute = tools.find((tool) => tool.name === "execute_tool");
      expect(execute?.inputSchema.properties).toHaveProperty("name");
      expect(execute?.inputSchema.properties).toHaveProperty("arguments");
      expect(tools.some((tool) => tool.name === "stub-tool")).toBe(false);
      const result = await client.callTool({
        name: "execute_tool",
        arguments: {
          name: "stub-tool",
          arguments: { request: { name: "Example" } },
        },
      });
      expect(result.isError).toBeFalsy();
      expect(stub.mock.calls[0]?.[1]).toEqual({ request: { name: "Example" } });
      expect(stub.mock.calls[0]?.[2].mcpReq.signal).toBeInstanceOf(AbortSignal);
      await client.callTool({
        name: "execute_tool",
        arguments: { name: "no-args-tool" },
      });
      expect(noArgsStub.mock.calls[0]?.[1].mcpReq.signal).toBeInstanceOf(
        AbortSignal,
      );
    });
  });

  test("rejects invalid nested input and unknown tools without invoking handlers", async () => {
    const { server, stub } = makeHarness(true);
    await withClient(server, async (client) => {
      const invalid = await client.callTool({
        name: "execute_tool",
        arguments: { name: "stub-tool", arguments: { request: { name: 42 } } },
      });
      expect(invalid.isError).toBe(true);
      expect(invalid.content).toEqual([
        expect.objectContaining({
          text: expect.stringContaining("Invalid input"),
        }),
      ]);
      const unknown = await client.callTool({
        name: "execute_tool",
        arguments: { name: "unknown" },
      });
      expect(unknown.isError).toBe(true);
      expect(stub).not.toHaveBeenCalled();
    });
  });

  test("describes transform-bearing inputs and dispatches their parsed values", async () => {
    const { server } = makeHarness(true);
    await withClient(server, async (client) => {
      const description = await client.callTool({
        name: "describe_tool_input",
        arguments: { tool_names: ["transform-tool"] },
      });
      expect(description.isError).toBeFalsy();
      expect(description.content).toEqual([
        expect.objectContaining({ text: expect.stringContaining('"amount"') }),
      ]);
      const result = await client.callTool({
        name: "execute_tool",
        arguments: {
          name: "transform-tool",
          arguments: {
            amount: "12.5",
            price: "8",
            when: "2026-01-01T00:00:00.000Z",
          },
        },
      });
      expect(result.content).toEqual([
        {
          type: "text",
          text: JSON.stringify({
            amount: 12.5,
            price: "8",
            when: "2026-01-01T00:00:00.000Z",
          }),
        },
      ]);
    });
  });
});

describe("MCP protocol eras", () => {
  test.each(["legacy", "modern"] as const)(
    "HTTP serves %s clients",
    async (era) => {
      const handler = createMcpHandler(() => makeHarness(false).server);
      const client = new Client(
        { name: "http-test", version: "1.0.0" },
        {
          versionNegotiation: {
            mode: era === "modern" ? { pin: "2026-07-28" } : "legacy",
          },
        },
      );
      const transport = new StreamableHTTPClientTransport(
        new URL("http://localhost/mcp"),
        {
          fetch: (input, init) => handler.fetch(new Request(input, init)),
        },
      );
      try {
        await client.connect(transport);
        expect(
          (await client.listTools()).tools.some(
            (tool) => tool.name === "stub-tool",
          ),
        ).toBe(true);
        expect(
          (await client.callTool({ name: "no-args-tool" })).content,
        ).toEqual([{ type: "text", text: "no-args-result" }]);
      } finally {
        await client.close();
        await handler.close();
      }
    },
  );

  test.each(["legacy", "modern"] as const)(
    "stdio entry serves %s clients",
    async (era) => {
      const [clientTransport, serverTransport] =
        InMemoryTransport.createLinkedPair();
      const handle = serveStdio(() => makeHarness(false).server, {
        transport: serverTransport,
      });
      const client = new Client(
        { name: "stdio-test", version: "1.0.0" },
        {
          versionNegotiation: {
            mode: era === "modern" ? { pin: "2026-07-28" } : "legacy",
          },
        },
      );
      try {
        await client.connect(clientTransport);
        expect(
          (await client.listTools()).tools.some(
            (tool) => tool.name === "stub-tool",
          ),
        ).toBe(true);
        expect(
          (await client.callTool({ name: "no-args-tool" })).content,
        ).toEqual([{ type: "text", text: "no-args-result" }]);
      } finally {
        await client.close();
        await handle.close();
      }
    },
  );
});

test("bigint defaults remain precise in listed and dynamically described schemas", async () => {
  const { server } = makeHarness(false);
  await withClient(server, async (client) => {
    const listed = (await client.listTools()).tools.find(
      (tool) => tool.name === "bigint-default-tool",
    );
    expect(JSON.stringify(listed?.inputSchema)).toContain(
      '"12345678901234567890"',
    );
    expect(listed?.inputSchema.properties?.["request"]).toMatchObject({
      default: {
        amount: "12345678901234567890",
        amounts: ["12345678901234567890"],
      },
      properties: {
        amount: { default: "12345678901234567890" },
        amounts: { default: ["12345678901234567890"] },
      },
    });
    const result = await client.callTool({ name: "bigint-default-tool" });
    expect(result.content).toEqual([
      {
        type: "text",
        text: '{"request":{"amount":"12345678901234567890","amounts":["12345678901234567890"]}}',
      },
    ]);
  });
  const dynamic = makeHarness(true);
  await withClient(dynamic.server, async (client) => {
    const result = await client.callTool({
      name: "describe_tool_input",
      arguments: { tool_names: ["bigint-default-tool"] },
    });
    expect(result.isError).toBeFalsy();
    expect(result.content).toEqual([
      expect.objectContaining({
        text: expect.stringContaining('"12345678901234567890"'),
      }),
    ]);
  });
});

test("schema conversion preserves recursive references and runtime bigint defaults", async () => {
  type Item = { amount: bigint | string; children?: Item[] | undefined };
  const item: z.ZodType<Item> = z
    .lazy(() =>
      z.object({
        amount: bigint().default(12345678901234567890n),
        children: z.array(item).optional(),
      }),
    )
    .describe("Recursive item");
  const schema = mcpInputSchema({ item });
  const json = schema["~standard"].jsonSchema.input({
    target: "draft-2020-12",
  });
  expect(JSON.stringify(json)).toContain('"12345678901234567890"');
  expect(JSON.stringify(json)).toContain('"$ref"');
  expect(JSON.stringify(json)).toContain("Recursive item");
  expect(await schema["~standard"].validate({ item: {} })).toEqual({
    value: { item: { amount: 12345678901234567890n } },
  });
  expect(item.parse({}).amount).toBe(12345678901234567890n);
});

test("schema conversion does not advertise string defaults for bigint-only inputs", () => {
  const schema = mcpInputSchema({
    amount: z.bigint().default(12345678901234567890n),
  });
  const json = schema["~standard"].jsonSchema.input({
    target: "draft-2020-12",
  });
  expect(json).toMatchObject({ properties: { amount: {} } });
  expect(JSON.stringify(json)).not.toContain('"default"');
});

test("schema conversion advertises ISO dates and keeps original runtime defaults", async () => {
  const date = new Date("2026-01-01T00:00:00.000Z");
  const shape = {
    request: z
      .object({
        when: z
          .union([z.date(), z.string().transform((value) => new Date(value))])
          .default(date),
        amount: bigint().default(12345678901234567890n),
      })
      .default({ when: date, amount: 12345678901234567890n }),
  };
  const schema = mcpInputSchema(shape);
  const json = schema["~standard"].jsonSchema.input({
    target: "draft-2020-12",
  });
  expect(json).toMatchObject({
    properties: {
      request: {
        default: {
          when: "2026-01-01T00:00:00.000Z",
          amount: "12345678901234567890",
        },
        properties: {
          when: { default: "2026-01-01T00:00:00.000Z" },
          amount: { default: "12345678901234567890" },
        },
      },
    },
  });
  expect(JSON.parse(JSON.stringify(json))).toEqual(json);
  const parsed = await schema["~standard"].validate({});
  expect(parsed).toEqual({
    value: { request: { when: date, amount: 12345678901234567890n } },
  });
  expect(shape.request.parse(undefined).when).toBeInstanceOf(Date);
});

test("schema conversion is lazy and observes current default factories", async () => {
  let amount = 12345678901234567890n;
  const defaultValue = vi.fn(() => amount);
  const shape = { amount: bigint().default(defaultValue) };
  const schema = mcpInputSchema(shape);
  expect(await schema["~standard"].validate({ amount: "42" })).toEqual({
    value: { amount: 42n },
  });
  expect(defaultValue).not.toHaveBeenCalled();

  const options = { target: "draft-2020-12" as const };
  expect(schema["~standard"].jsonSchema.input(options)).toMatchObject({
    properties: { amount: { default: "12345678901234567890" } },
  });
  amount = 12345678901234567891n;
  expect(schema["~standard"].jsonSchema.input(options)).toMatchObject({
    properties: { amount: { default: "12345678901234567891" } },
  });
  expect(schema["~standard"].jsonSchema.output(options)).toMatchObject({
    properties: { amount: { default: "12345678901234567891" } },
  });

  defaultValue.mockClear();
  expect(await schema["~standard"].validate({})).toEqual({ value: { amount } });
  expect(defaultValue).toHaveBeenCalledTimes(1);
});

test("prompt completion preserves input validation, transforms, and defaults", async () => {
  const server = new McpServer({ name: "prompt-test", version: "1.0.0" });
  const client = new Client({ name: "prompt-client", version: "1.0.0" });
  const complete = vi.fn((value: string) =>
    ["typescript", "python"].filter((language) => language.startsWith(value)),
  );
  const handler = vi.fn(
    (args: { language: string; count: number; label: string }) => ({
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
    () => new SDKCore(),
    new Set(),
  )({
    name: "review",
    args: {
      language: completable(z.string().describe("Language"), complete),
      count: z.string().regex(/^\d+$/).transform(Number),
      label: z.string().default("example"),
      suggestion: completable(z.string(), complete).optional(),
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
        argument: { name: "suggestion", value: "py" },
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
              count: 7,
              label: "example",
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

test.each([false, true])(
  "async tool validation preserves static/dynamic parity (dynamic=%s)",
  async (dynamic) => {
    const server = new McpServer({ name: "async-tool-test", version: "1.0.0" });
    const logger = createConsoleLogger("error");
    const getSDK = () => new SDKCore();
    const scopes = new Set<MCPScope>();
    const [tool, , toolMap] = createRegisterTool(
      logger,
      server,
      getSDK,
      scopes,
      undefined,
      dynamic,
    );
    const handler = vi.fn((args: { count: number; name: string }) => ({
      content: [{ type: "text" as const, text: JSON.stringify(args) }],
    }));
    tool({
      name: "async-input",
      description: "Validate and transform input asynchronously",
      annotations,
      args: {
        count: z.number().refine(async (value) => value % 2 === 0),
        name: z.string().transform(async (value) => value.toUpperCase()),
      },
      tool: (_sdk, args) => handler(args),
    });
    if (dynamic) registerDynamicTools(logger, server, getSDK, toolMap, scopes);
    await withClient(server, async (client) => {
      const call = (count: number) =>
        client.callTool({
          name: dynamic ? "execute_tool" : "async-input",
          arguments: dynamic
            ? { name: "async-input", arguments: { count, name: "example" } }
            : { count, name: "example" },
        });
      expect((await call(4)).isError).toBeFalsy();
      expect(handler).toHaveBeenLastCalledWith({ count: 4, name: "EXAMPLE" });
      expect((await call(3)).isError).toBe(true);
      expect(handler).toHaveBeenCalledTimes(1);
    });
  },
);

test.each([false, true])(
  "BigInt catch fallbacks remain serializable and preserve runtime values (dynamic=%s)",
  async (dynamic) => {
    const server = new McpServer({ name: "catch-test", version: "1.0.0" });
    const logger = createConsoleLogger("error");
    const getSDK = () => new SDKCore();
    const scopes = new Set<MCPScope>();
    const [tool, , toolMap] = createRegisterTool(
      logger,
      server,
      getSDK,
      scopes,
      undefined,
      dynamic,
    );
    const amount = 12345678901234567890n;
    const fallback = vi.fn(() => ({ amount }));
    const shape = {
      request: z.object({ amount: bigint() }).catch(fallback),
      bigintOnly: z.bigint().catch(amount),
    };
    tool({
      name: "catch-input",
      description: "Recover from invalid input",
      annotations,
      args: shape,
      tool: (_sdk, args) => {
        expect(args).toEqual({ request: { amount }, bigintOnly: amount });
        return {
          content: [{ type: "text", text: String(args.request.amount) }],
        };
      },
    });
    expect(fallback).not.toHaveBeenCalled();
    if (dynamic) registerDynamicTools(logger, server, getSDK, toolMap, scopes);
    await withClient(server, async (client) => {
      const json = mcpInputSchema(shape)["~standard"].jsonSchema.input({
        target: "draft-2020-12",
      });
      expect(json).toMatchObject({
        properties: { request: { default: { amount: String(amount) } } },
      });
      expect(JSON.parse(JSON.stringify(json))).toEqual(json);
      expect(
        (json["properties"] as Record<string, unknown>)["bigintOnly"],
      ).not.toHaveProperty("default");
      if (dynamic) {
        const described = await client.callTool({
          name: "describe_tool_input",
          arguments: { tool_names: ["catch-input"] },
        });
        expect(described.isError).toBeFalsy();
        expect(JSON.stringify(described)).toContain(String(amount));
      } else {
        const listed = await client.listTools();
        expect(listed.tools[0]?.inputSchema).toMatchObject(json);
      }
      expect(
        await client.callTool({
          name: dynamic ? "execute_tool" : "catch-input",
          arguments: dynamic
            ? {
                name: "catch-input",
                arguments: { request: null, bigintOnly: null },
              }
            : { request: null, bigintOnly: null },
        }),
      ).toMatchObject({ content: [{ text: String(amount) }] });
    });
  },
);

test.each([false, true])(
  "context-dependent catch factories preserve discovery and runtime fallback (dynamic=%s)",
  async (dynamic) => {
    const server = new McpServer({
      name: "context-catch-test",
      version: "1.0.0",
    });
    const logger = createConsoleLogger("error");
    const getSDK = () => new SDKCore();
    const scopes = new Set<MCPScope>();
    const [tool, , toolMap] = createRegisterTool(
      logger,
      server,
      getSDK,
      scopes,
      undefined,
      dynamic,
    );
    const field = z.string().catch((ctx) => `fallback:${String(ctx.value)}`);
    const shape = { value: field };
    tool({
      name: "context-catch",
      description: "Recover using the invalid input",
      annotations,
      args: shape,
      tool: (_sdk, args) => ({ content: [{ type: "text", text: args.value }] }),
    });
    if (dynamic) registerDynamicTools(logger, server, getSDK, toolMap, scopes);
    await withClient(server, async (client) => {
      const standard = mcpInputSchema(shape)["~standard"];
      for (const direction of ["input", "output"] as const) {
        const json = standard.jsonSchema[direction]({
          target: "draft-2020-12",
        });
        expect(
          (json["properties"] as Record<string, unknown>)["value"],
        ).not.toHaveProperty("default");
        expect(JSON.parse(JSON.stringify(json))).toEqual(json);
      }
      const listed = await client.listTools();
      expect(listed.tools.length).toBeGreaterThan(0);
      if (dynamic) {
        const described = await client.callTool({
          name: "describe_tool_input",
          arguments: { tool_names: ["context-catch"] },
        });
        expect(described.isError).toBeFalsy();
        expect(JSON.stringify(described)).not.toContain("default");
      } else {
        expect(listed.tools[0]?.inputSchema).toMatchObject({
          properties: { value: { type: "string" } },
        });
        expect(
          (listed.tools[0]?.inputSchema.properties as Record<string, unknown>)[
            "value"
          ],
        ).not.toHaveProperty("default");
      }
      for (const value of [42, "valid"]) {
        expect(
          await client.callTool({
            name: dynamic ? "execute_tool" : "context-catch",
            arguments: dynamic
              ? { name: "context-catch", arguments: { value } }
              : { value },
          }),
        ).toMatchObject({ content: [{ text: field.parse(value) }] });
      }
      expect(field.parse(42)).toBe("fallback:42");
    });
  },
);

test.each([
  ["legacy", "static"],
  ["legacy", "dynamic"],
  ["modern", "static"],
  ["modern", "dynamic"],
] as const)(
  "generated stdio supports concurrent calls and cancellation (%s, %s)",
  async (era, mode) => {
    let requestCount = 0;
    let signalStarted!: () => void;
    let signalClosed!: () => void;
    const started = new Promise<void>((resolve) => {
      signalStarted = resolve;
    });
    const closed = new Promise<void>((resolve) => {
      signalClosed = resolve;
    });
    const upstream = createServer((request, response) => {
      requestCount++;
      response.setHeader("content-type", "application/json");
      if (request.url?.includes("/slow/")) {
        signalStarted();
        const timer = setTimeout(() => response.end("{}"), 5000);
        response.on("close", () => {
          clearTimeout(timer);
          signalClosed();
        });
      } else {
        const url = request.url;
        setTimeout(
          () => response.end(JSON.stringify({ url })),
          requestCount % 7,
        );
      }
    });
    await new Promise<void>((resolve) =>
      upstream.listen(0, "127.0.0.1", resolve),
    );
    const address = upstream.address();
    if (typeof address !== "object" || address === null)
      throw new Error("No upstream address");
    const toolName = "parameters-duplicate-path-param";
    const transport = new StdioClientTransport({
      command: process.execPath,
      args: [
        fileURLToPath(new URL("../../bin/mcp-server.js", import.meta.url)),
        "start",
        "--transport",
        "stdio",
        "--server-url",
        "http://127.0.0.1:" + address.port,
        "--tool",
        toolName,
        ...(mode === "dynamic" ? ["--mode", "dynamic"] : []),
      ],
      stderr: "pipe",
    });
    let stderr = "";
    transport.stderr?.on("data", (chunk: Buffer) => {
      stderr += chunk.toString();
    });
    const client = new Client(
      { name: "transport-test", version: "1.0.0" },
      {
        versionNegotiation: {
          mode: era === "modern" ? { pin: "2026-07-28" } : "legacy",
        },
      },
    );
    const toolCall = (param1: unknown) => {
      const arguments_ = { request: { param1, param2: "tail" } };
      return mode === "dynamic"
        ? {
            name: "execute_tool",
            arguments: { name: toolName, arguments: arguments_ },
          }
        : { name: toolName, arguments: arguments_ };
    };
    try {
      await client.connect(transport);
      const { tools } = await client.listTools();
      expect(tools.map((tool) => tool.name)).toEqual(
        mode === "dynamic"
          ? expect.arrayContaining([
              "list_tools",
              "describe_tool_input",
              "execute_tool",
            ])
          : [toolName],
      );
      const results = await Promise.all(
        Array.from({ length: 20 }, (_, index) =>
          client.callTool(toolCall("item-" + index)),
        ),
      );
      for (const [index, result] of results.entries()) {
        expect(result.isError, stderr).not.toBe(true);
        const content = result.content?.[0];
        expect(content?.type).toBe("text");
        if (content?.type !== "text") throw new Error("Expected text result");
        expect(JSON.parse(content.text)).toEqual({
          url: "/anything/params/item-" + index + "/tail/item-" + index,
        });
      }
      const beforeInvalid = requestCount;
      expect((await client.callTool(toolCall(42))).isError).toBe(true);
      expect(requestCount).toBe(beforeInvalid);

      const controller = new AbortController();
      const pending = client
        .callTool(toolCall("slow"), { signal: controller.signal })
        .then(
          () => false,
          () => true,
        );
      let startTimer: ReturnType<typeof setTimeout> | undefined;
      try {
        await Promise.race([
          started,
          pending.then(() => {
            throw new Error("Tool call completed before reaching upstream");
          }),
          new Promise<never>((_, reject) => {
            startTimer = setTimeout(
              () => reject(new Error("Tool call did not reach upstream")),
              2000,
            );
          }),
        ]);
      } finally {
        clearTimeout(startTimer);
      }
      controller.abort();
      expect(await pending).toBe(true);
      let timer: ReturnType<typeof setTimeout> | undefined;
      try {
        await Promise.race([
          closed,
          new Promise<never>((_, reject) => {
            timer = setTimeout(
              () => reject(new Error("Upstream request was not cancelled")),
              1500,
            );
          }),
        ]);
      } finally {
        clearTimeout(timer);
      }
      expect(
        (await client.callTool(toolCall("after-cancel"))).isError,
      ).not.toBe(true);
    } finally {
      await client.close();
      upstream.closeAllConnections();
      await new Promise<void>((resolve) => upstream.close(() => resolve()));
    }
  },
  15000,
);
