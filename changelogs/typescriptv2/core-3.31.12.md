## core: 3.31.12 - 2026-11-12
### :bug: Bug Fixes
- behaviour change: operation path parameter values that form . or .. URL segments are now rejected with an InvalidRequestError (an MCP tool error for MCP servers) instead of sending the request to a different route *(commit by [@TristanSpeakEasy](https://github.com/TristanSpeakEasy))*
