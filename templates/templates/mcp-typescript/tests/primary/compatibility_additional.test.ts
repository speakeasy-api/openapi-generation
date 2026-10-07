import { Client } from "@modelcontextprotocol/client";
import { InMemoryTransport, McpServer } from "@modelcontextprotocol/server";
import { ResourceTemplate } from "@modelcontextprotocol/sdk/server/mcp.js";
import {
  ErrorCode,
  ListRootsResultSchema,
} from "@modelcontextprotocol/sdk/types.js";
import {
  mcpServerContext,
  type MCPServerContext,
} from "../mcp-server/shared.js";
import { expect, test } from "vitest";
import { SDKCore } from "../core.js";
import { createConsoleLogger } from "../mcp-server/console-logger.js";
import { createRegisterResourceTemplate } from "../mcp-server/resources.js";
import {
  createRegisterTool,
  registerDynamicTools,
} from "../mcp-server/tools.js";

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
    () => new SDKCore(),
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

test.each([false, true])(
  "legacy tool context preserves cancellation (dynamic=%s)",
  async (dynamic) => {
    const server = new McpServer({
      name: "compatibility-server",
      version: "1.0.0",
    });
    const logger = createConsoleLogger("error");
    const getSDK = () => new SDKCore();
    const scopes = new Set<never>();
    const [tool, , toolMap] = createRegisterTool(
      logger,
      server,
      getSDK,
      scopes,
      undefined,
      dynamic,
    );
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
      annotations: {
        title: "Held",
        readOnlyHint: true,
        destructiveHint: false,
        idempotentHint: true,
        openWorldHint: false,
      },
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
    if (dynamic) registerDynamicTools(logger, server, getSDK, toolMap, scopes);
    await withClient(server, async (client) => {
      const controller = new AbortController();
      const pending = client
        .callTool(
          dynamic
            ? {
                name: "execute_tool",
                arguments: { name: "held", arguments: {} },
              }
            : { name: "held" },
          { signal: controller.signal },
        )
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
  },
  10000,
);

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
