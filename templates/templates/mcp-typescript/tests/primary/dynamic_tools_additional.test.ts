import { describe, expect, test, vi } from "vitest";
import { z } from "zod";
import { SDKCore } from "../core.js";
import { createConsoleLogger } from "../mcp-server/console-logger.js";
import { MCPScope } from "../mcp-server/scopes.js";
import { registerDynamicTools, ToolDefinition } from "../mcp-server/tools.js";

// Regression tests: dynamic mode `execute_tool` must accept tool
// arguments under a nested `arguments` object, matching the MCP spec's
// canonical `tools/call` shape (`{ name, arguments }`). Previously, args were
// wrapped under an `input` field typed as an opaque `record<string, unknown>`,
// which produced JSON Schema that LLMs could not reliably populate.

type CapturedHandler = {
  config: { inputSchema?: unknown };
  handler: (
    args: Record<string, unknown>,
    ctx: { signal: AbortSignal },
  ) => Promise<{
    isError?: boolean;
    content: Array<{ type: string; text: string }>;
  }>;
};

function makeHarness() {
  const handlers = new Map<string, CapturedHandler>();
  const mockServer = {
    registerTool(
      name: string,
      config: { inputSchema?: unknown },
      handler: CapturedHandler["handler"],
    ) {
      handlers.set(name, { config, handler });
    },
  };

  const stub = vi.fn().mockResolvedValue({
    content: [{ type: "text", text: "stub-result" }],
  });

  const argsShape = { request: z.object({ name: z.string() }) };
  const stubToolDef: ToolDefinition<typeof argsShape> = {
    name: "stub-tool",
    description: "Stub tool for execute_tool unit tests",
    annotations: {
      title: "Stub",
      readOnlyHint: true,
      destructiveHint: false,
      idempotentHint: true,
      openWorldHint: false,
    },
    args: argsShape,
    // deno-lint-ignore no-explicit-any
    tool: stub as any,
  };

  const noArgsStub = vi.fn().mockResolvedValue({
    content: [{ type: "text", text: "no-args-result" }],
  });

  const noArgsToolDef: ToolDefinition = {
    name: "no-args-tool",
    description: "Tool that takes no arguments",
    annotations: {
      title: "NoArgs",
      readOnlyHint: true,
      destructiveHint: false,
      idempotentHint: true,
      openWorldHint: false,
    },
    // deno-lint-ignore no-explicit-any
    tool: noArgsStub as any,
  };

  // deno-lint-ignore no-explicit-any
  const toolMap = new Map<string, ToolDefinition<any>>();
  toolMap.set("stub-tool", stubToolDef);
  toolMap.set("no-args-tool", noArgsToolDef);

  registerDynamicTools(
    createConsoleLogger("error"),
    // deno-lint-ignore no-explicit-any
    mockServer as any,
    () => ({}) as SDKCore,
    // deno-lint-ignore no-explicit-any
    toolMap as any,
    new Set<MCPScope>(),
  );

  return { handlers, stub, noArgsStub };
}

describe("dynamic mode execute_tool", () => {
  test("accepts tool arguments under nested `arguments` field", async () => {
    const { handlers, stub } = makeHarness();
    const execute = handlers.get("execute_tool");
    expect(execute).toBeDefined();

    const result = await execute!.handler(
      { name: "stub-tool", arguments: { request: { name: "Alice" } } },
      { signal: new AbortController().signal },
    );

    expect(stub).toHaveBeenCalledTimes(1);
    // tool(client, validatedInput, ctx) — validated input is second arg
    expect(stub.mock.calls[0]?.[1]).toEqual({ request: { name: "Alice" } });
    expect(result.isError).toBeFalsy();
  });

  test("rejects invalid inner arguments with a schema error", async () => {
    const { handlers, stub } = makeHarness();
    const execute = handlers.get("execute_tool")!;

    const result = await execute.handler(
      {
        name: "stub-tool",
        arguments: { request: { name: 42 } },
      },
      { signal: new AbortController().signal },
    );

    expect(stub).not.toHaveBeenCalled();
    expect(result.isError).toBe(true);
    expect(result.content[0]?.text).toContain("Invalid input");
  });

  test("inputSchema advertises `name` and `arguments` meta-fields", () => {
    const { handlers } = makeHarness();
    const { config } = handlers.get("execute_tool")!;
    const shape = config.inputSchema as Record<string, z.ZodTypeAny>;

    expect(shape["name"]).toBeDefined();
    const argumentsSchema = shape["arguments"];
    expect(argumentsSchema).toBeDefined();

    // `arguments` must accept an arbitrary object payload (loose) so the
    // nested target-tool args survive validation.
    const parsed = argumentsSchema!.parse({ request: { name: "Alice" } });
    expect(parsed).toEqual({ request: { name: "Alice" } });
  });

  test("reports unknown tool name without invoking stub", async () => {
    const { handlers, stub } = makeHarness();
    const execute = handlers.get("execute_tool")!;

    const result = await execute.handler(
      { name: "does-not-exist" },
      { signal: new AbortController().signal },
    );

    expect(stub).not.toHaveBeenCalled();
    expect(result.isError).toBe(true);
    expect(result.content[0]?.text).toContain("Unknown tool: does-not-exist");
  });

  test("executes a no-args tool when `arguments` is omitted", async () => {
    const { handlers, noArgsStub } = makeHarness();
    const execute = handlers.get("execute_tool")!;

    const result = await execute.handler(
      { name: "no-args-tool" },
      { signal: new AbortController().signal },
    );

    expect(noArgsStub).toHaveBeenCalledTimes(1);
    expect(result.isError).toBeFalsy();
    expect(result.content[0]?.text).toBe("no-args-result");
  });
});

// Regression test: dynamic mode `describe_tool_input` must not
// fail for tools whose generated schemas contain Zod `.transform()` pipelines
// (e.g. numeric coercion, decimals, date-time fields). Previously the default
// `z.toJSONSchema` call threw "Transforms cannot be represented in JSON Schema"
// and the tool was returned with `isError: true`, making dynamic mode unusable.
describe("dynamic mode describe_tool_input", () => {
  function makeTransformHarness() {
    const handlers = new Map<string, CapturedHandler>();
    const mockServer = {
      registerTool(
        name: string,
        config: { inputSchema?: unknown },
        handler: CapturedHandler["handler"],
      ) {
        handlers.set(name, { config, handler });
      },
    };

    // Mirrors the kinds of pipelines emitted by the generator: numeric coercion
    // from string, a decimal-style string transform, and a date-time union.
    const argsShape = {
      amount: z
        .union([z.string(), z.number()])
        .transform((v) => (typeof v === "string" ? parseFloat(v) : v)),
      price: z.string().transform((v) => v),
      when: z
        .union([z.date(), z.string().transform((v) => new Date(v))])
        .transform((v) => v.toISOString()),
    };
    const transformToolDef: ToolDefinition<typeof argsShape> = {
      name: "transform-tool",
      description: "Tool whose schema relies on Zod transforms",
      annotations: {
        title: "Transform",
        readOnlyHint: true,
        destructiveHint: false,
        idempotentHint: true,
        openWorldHint: false,
      },
      args: argsShape,
      // deno-lint-ignore no-explicit-any
      tool: vi.fn() as any,
    };

    // deno-lint-ignore no-explicit-any
    const toolMap = new Map<string, ToolDefinition<any>>();
    toolMap.set("transform-tool", transformToolDef);

    registerDynamicTools(
      createConsoleLogger("error"),
      // deno-lint-ignore no-explicit-any
      mockServer as any,
      () => ({}) as SDKCore,
      // deno-lint-ignore no-explicit-any
      toolMap as any,
      new Set<MCPScope>(),
    );

    return { handlers };
  }

  test("returns JSON Schema for transform-bearing tools without erroring", async () => {
    const { handlers } = makeTransformHarness();
    const describe = handlers.get("describe_tool_input");
    expect(describe).toBeDefined();

    const result = await describe!.handler(
      { tool_names: ["transform-tool"] },
      { signal: new AbortController().signal },
    );

    expect(result.isError).toBeFalsy();
    const text = result.content[0]?.text ?? "";
    expect(text).not.toContain("Transforms cannot be represented");
    expect(text).toContain('<input_schema tool="transform-tool">');
    // The input side of the coercion union is surfaced for the LLM.
    expect(text).toContain("amount");
    expect(text).toContain("price");
    expect(text).toContain("when");
  });
});
