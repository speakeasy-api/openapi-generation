import { Client } from "@modelcontextprotocol/client";
import { InMemoryTransport, McpServer } from "@modelcontextprotocol/server";
import { AjvJsonSchemaValidator } from "@modelcontextprotocol/server/validators/ajv";
import { ResourceTemplate } from "@modelcontextprotocol/sdk/server/mcp.js";
import {
  ErrorCode,
  ListRootsResultSchema,
} from "@modelcontextprotocol/sdk/types.js";
import {
  mcpSchema,
  mcpServerContext,
  type MCPServerContext,
} from "../mcp-server/shared.js";
import { expect, test } from "vitest";
import { SDKCore } from "../core.js";
import { createConsoleLogger } from "../mcp-server/console-logger.js";
import {
  createRegisterResource,
  createRegisterResourceTemplate,
} from "../mcp-server/resources.js";
import { createRegisterTool } from "../mcp-server/tools.js";
import { createRegisterPrompt } from "../mcp-server/prompts.js";
import { z } from "zod/v3";

async function withClient(
  server: McpServer,
  run: (client: Client) => Promise<void>,
) {
  const client = new Client(
    { name: "compatibility-client", version: "1.0.0" },
    { capabilities: { roots: {} }, versionNegotiation: { mode: "legacy" } },
  );
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

test("v1 resource templates retain discovery, completion, and reads", async () => {
  const server = new McpServer({
    name: "compatibility-server",
    version: "1.0.0",
  });
  createRegisterResourceTemplate(
    createConsoleLogger("error"),
    server,
    new SDKCore(),
    new Set(),
  )({
    name: "widget",
    description: "Read a widget",
    resource: new ResourceTemplate("widget://{name}", {
      list: (extra) => {
        expect(extra.signal).toBeInstanceOf(AbortSignal);
        expect(extra.requestId).toBeDefined();
        return { resources: [{ name: "example", uri: "widget://example" }] };
      },
      complete: { name: (value) => [value + "ample"] },
    }),
    read: (_sdk, uri, vars, extra) => {
      expect(extra.signal).toBe(extra.mcpReq.signal);
      expect(extra.requestId).toBe(extra.mcpReq.id);
      return {
        contents: [{ uri: uri.toString(), text: String(vars["name"]) }],
      };
    },
  });
  await withClient(server, async (client) => {
    expect(await client.listResources()).toMatchObject({
      resources: [{ uri: "widget://example" }],
    });
    expect(await client.listResourceTemplates()).toMatchObject({
      resourceTemplates: [{ uriTemplate: "widget://{name}" }],
    });
    expect(
      await client.complete({
        ref: { type: "ref/resource", uri: "widget://{name}" },
        argument: { name: "name", value: "ex" },
      }),
    ).toMatchObject({ completion: { values: ["example"] } });
    expect(
      await client.readResource({ uri: "widget://example" }),
    ).toMatchObject({ contents: [{ text: "example" }] });
  });
});

test("legacy tool context preserves cancellation", async () => {
  const server = new McpServer({
    name: "compatibility-server",
    version: "1.0.0",
  });
  const logger = createConsoleLogger("error");
  const tool = createRegisterTool(logger, server, new SDKCore(), new Set());
  let started!: () => void;
  const running = new Promise<void>((resolve) => {
    started = resolve;
  });
  let cancelled!: () => void;
  const aborted = new Promise<void>((resolve) => {
    cancelled = resolve;
  });
  let context: MCPServerContext | undefined;
  tool({
    name: "held",
    description: "Wait for cancellation",
    tool: async (_sdk, extra) => {
      context = extra;
      expect(extra.signal).toBe(extra.mcpReq.signal);
      expect(extra.requestId).toBe(extra.mcpReq.id);
      expect(extra.sendNotification).toBeTypeOf("function");
      extra.signal.addEventListener("abort", cancelled, { once: true });
      started();
      await aborted;
      return { content: [{ type: "text", text: "cancelled" }] };
    },
  });
  await withClient(server, async (client) => {
    const controller = new AbortController();
    const pending = client
      .callTool({ name: "held" }, { signal: controller.signal })
      .catch((error: unknown) => error);
    await running;
    controller.abort();
    await aborted;
    await expect(
      context!.sendNotification({
        method: "notifications/tools/list_changed",
      }),
    ).resolves.toBeUndefined();
    await expect(
      context!.sendRequest({ method: "roots/list" }, ListRootsResultSchema),
    ).rejects.toMatchObject({ code: ErrorCode.ConnectionClosed });
    expect(await pending).toBeInstanceOf(Error);
  });
}, 10000);

test.each([false, true])(
  "legacy sendRequest validates v1 result schemas (invalid=%s)",
  async (invalid) => {
    const server = new McpServer({
      name: "compatibility-server",
      version: "1.0.0",
    });
    server.registerTool("roots", {}, async (ctx) => {
      const schema = invalid
        ? ListRootsResultSchema.refine(() => false)
        : ListRootsResultSchema;
      const result = await mcpServerContext(ctx).sendRequest(
        { method: "roots/list" },
        schema,
      );
      return { content: [{ type: "text", text: JSON.stringify(result) }] };
    });
    await withClient(server, async (client) => {
      client.setRequestHandler("roots/list", async () => ({
        roots: [{ uri: "file:///example", name: "Example" }],
      }));
      const result = await client.callTool({ name: "roots" });
      if (invalid) expect(result.isError).toBe(true);
      else {
        expect(result.isError).not.toBe(true);
        expect(result.content).toEqual([
          {
            type: "text",
            text: JSON.stringify({
              roots: [{ uri: "file:///example", name: "Example" }],
            }),
          },
        ]);
      }
    });
  },
);

test("optional MCP prompts expose legacy context with and without arguments", async () => {
  const server = new McpServer({ name: "prompt-context", version: "1.0.0" });
  const register = createRegisterPrompt(
    createConsoleLogger("error"),
    server,
    new SDKCore(),
    new Set(),
  );
  const result = (extra: MCPServerContext) => {
    expect(extra.signal).toBe(extra.mcpReq.signal);
    expect(extra.requestId).toBe(extra.mcpReq.id);
    expect(extra.sendRequest).toBeTypeOf("function");
    expect(extra.sendNotification).toBeTypeOf("function");
    expect(extra._meta).toEqual({ example: "metadata" });
    return { messages: [] };
  };
  register({
    name: "with-args",
    args: { value: z.string() },
    prompt: (_sdk, _args, extra) => result(extra),
  });
  register({ name: "without-args", prompt: (_sdk, extra) => result(extra) });
  await withClient(server, async (client) => {
    await client.getPrompt({
      name: "with-args",
      arguments: { value: "example" },
      _meta: { example: "metadata" },
    });
    await client.getPrompt({
      name: "without-args",
      _meta: { example: "metadata" },
    });
  });
});

test("MCP schemas distinguish pipeline input and output while preserving parsing", async () => {
  const shape = {
    count: z.string().transform(Number).pipe(z.number().int().min(1)),
  };
  const schema = mcpSchema(shape)["~standard"];
  const input = schema.jsonSchema.input({ target: "draft-2020-12" });
  const output = schema.jsonSchema.output({ target: "draft-2020-12" });
  expect(input).toMatchObject({ properties: { count: { type: "string" } } });
  expect(output).toMatchObject({
    properties: { count: { type: "integer", minimum: 1 } },
  });
  expect(await schema.validate({ count: "7" })).toEqual({
    value: { count: 7 },
  });
  expect(await schema.validate({ count: "0" })).toHaveProperty("issues");
  const server = new McpServer({ name: "pipeline-test", version: "1.0.0" });
  server.registerTool(
    "pipeline",
    { inputSchema: mcpSchema(shape), outputSchema: mcpSchema(shape) },
    async (args) => ({
      content: [{ type: "text", text: String(args.count) }],
      structuredContent: { count: args.count },
    }),
  );
  await withClient(server, async (client) => {
    const listed = (await client.listTools()).tools[0];
    expect(listed?.inputSchema).toMatchObject(input);
    expect(listed?.outputSchema).toMatchObject(output);
  });
});

test("MCP output schemas do not claim input types for transforms or discard unknown union outputs", async () => {
  const schema = mcpSchema({
    length: z.string().transform((value) => value.length),
    choice: z.union([
      z.string().transform((value) => value.length),
      z.boolean(),
    ]),
  })["~standard"];
  const input = schema.jsonSchema.input({ target: "draft-2020-12" });
  expect(input).toMatchObject({
    properties: {
      length: { type: "string" },
      choice: { anyOf: [{ type: "string" }, { type: "boolean" }] },
    },
  });
  const output = schema.jsonSchema.output({ target: "draft-2020-12" });
  expect(output["properties"]).toEqual({
    length: {},
    choice: { anyOf: [{}, { type: "boolean" }] },
  });
  expect(await schema.validate({ length: "four", choice: "six" })).toEqual({
    value: { length: 4, choice: 3 },
  });
});

test.each(["input", "output"] as const)(
  "MCP %s schemas honor JSON Schema dialects for tuples",
  (direction) => {
    const converter = mcpSchema({
      fixed: z.tuple([z.string(), z.number()]),
      rest: z.tuple([z.string()]).rest(z.number()),
      empty: z.tuple([]),
    })["~standard"].jsonSchema[direction];
    for (const target of ["draft-07", "draft-2020-12"] as const) {
      const schema = converter({ target });
      expect(schema["$schema"]).toBe(
        target === "draft-07"
          ? "http://json-schema.org/draft-07/schema#"
          : "https://json-schema.org/draft/2020-12/schema",
      );
      const validate = new AjvJsonSchemaValidator().getValidator(schema);
      const valid = {
        fixed: ["example", 1],
        rest: ["example", 2, 3],
        empty: [],
      };
      expect(validate(valid).valid).toBe(true);
      expect(validate({ ...valid, fixed: ["example", "wrong"] }).valid).toBe(
        false,
      );
      expect(validate({ ...valid, fixed: ["example", 1, 2] }).valid).toBe(
        false,
      );
      expect(validate({ ...valid, rest: ["example", "wrong"] }).valid).toBe(
        false,
      );
      expect(validate({ ...valid, empty: [1] }).valid).toBe(false);
    }
  },
);

test("MCP draft 2020-12 schemas preserve recursive references relocated inside tuples", () => {
  type Node = { value: string; children?: Node[] | undefined };
  const node: z.ZodType<Node> = z.lazy(() =>
    z.object({ value: z.string(), children: z.array(node).optional() }),
  );
  const schema = mcpSchema({ pair: z.tuple([node, node]), same: node })[
    "~standard"
  ].jsonSchema.input({ target: "draft-2020-12" });
  const validate = new AjvJsonSchemaValidator().getValidator(schema);
  const valid = {
    pair: [{ value: "one", children: [{ value: "nested" }] }, { value: "two" }],
    same: { value: "three" },
  };
  expect(validate(valid).valid).toBe(true);
  expect(validate({ ...valid, same: { value: 3 } }).valid).toBe(false);
  expect(
    validate({
      ...valid,
      pair: [{ value: "one", children: [{ value: 1 }] }, { value: "two" }],
    }).valid,
  ).toBe(false);
});

test("MCP dialect conversion leaves default values untouched and rejects unsupported dialects", () => {
  const value = {
    items: [{ type: "string" }],
    additionalItems: false,
    $ref: "#/properties/example/items/0",
  };
  const converter = mcpSchema({ example: z.any().default(value) })["~standard"]
    .jsonSchema;
  expect(converter.input({ target: "draft-2020-12" })).toMatchObject({
    properties: { example: { default: value } },
  });
  expect(() => converter.input({ target: "openapi-3.0" })).toThrow(
    "Unsupported JSON Schema target",
  );
  expect(() => converter.output({ target: "openapi-3.0" })).toThrow(
    "Unsupported JSON Schema target",
  );
});

test.each(["a/b", "a~b"])(
  "MCP draft 2020-12 references escape property name %s",
  (name) => {
    const value = z.object({ value: z.string() });
    const schema = mcpSchema({ [name]: value, same: value })[
      "~standard"
    ].jsonSchema.input({ target: "draft-2020-12" });
    const validate = new AjvJsonSchemaValidator().getValidator(schema);
    expect(
      validate({ [name]: { value: "one" }, same: { value: "two" } }).valid,
    ).toBe(true);
    expect(
      validate({ [name]: { value: "one" }, same: { value: 2 } }).valid,
    ).toBe(false);
  },
);

test("tools with arguments and static resources preserve legacy context", async () => {
  const server = new McpServer({ name: "callback-context", version: "1.0.0" });
  const logger = createConsoleLogger("error");
  const sdk = new SDKCore();
  const check = (extra: MCPServerContext) => {
    expect(extra.signal).toBe(extra.mcpReq.signal);
    expect(extra.requestId).toBe(extra.mcpReq.id);
    expect(extra._meta).toEqual({ example: "metadata" });
    expect(extra.sendRequest).toBeTypeOf("function");
    expect(extra.sendNotification).toBeTypeOf("function");
  };
  createRegisterTool(
    logger,
    server,
    sdk,
    new Set(),
  )({
    name: "length",
    description: "Return input length",
    args: { value: z.string().transform((value) => value.length) },
    tool: (_sdk, args, extra) => {
      check(extra);
      expect(args.value).toBe(4);
      return { content: [{ type: "text", text: String(args.value) }] };
    },
  });
  createRegisterResource(
    logger,
    server,
    sdk,
    new Set(),
  )({
    name: "example",
    resource: "example://static",
    read: (_sdk, uri, extra) => {
      check(extra);
      return { contents: [{ uri: uri.toString(), text: "example" }] };
    },
  });
  await withClient(server, async (client) => {
    expect(
      await client.callTool({
        name: "length",
        arguments: { value: "four" },
        _meta: { example: "metadata" },
      }),
    ).toMatchObject({ content: [{ text: "4" }] });
    expect(
      await client.readResource({
        uri: "example://static",
        _meta: { example: "metadata" },
      }),
    ).toMatchObject({ contents: [{ text: "example" }] });
  });
});
