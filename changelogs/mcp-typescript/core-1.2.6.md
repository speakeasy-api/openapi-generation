## core: 1.2.6 - 2026-02-11
### :bee: New Features
- Introduced `--mode dynamic` to generated MCP servers. In this mode, MCP servers no longer register all tools upfront and instead expose only four tools: `list_tools`, `describe_tool_input`, `list_scopes`. This enables progressively revealing tools to LLM agents which can be useful for minimizing the impact of large MCP servers on LLM context windows. *(commit by [@disintegrator](https://github.com/disintegrator))*
