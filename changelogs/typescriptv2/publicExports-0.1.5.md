## publicExports: 0.1.5 - 2026-10-08
### :bug: Bug Fixes
- export streaming and per-variant params types for SSE operations with union request bodies, add union-level streaming overloads to union-body SSE methods (union-typed params with a literal stream now resolve to the narrowed return type instead of the full response), and let explicit exports take precedence over generated SSE params aliases on name collisions *(commit by [@2ynn](https://github.com/2ynn))*
