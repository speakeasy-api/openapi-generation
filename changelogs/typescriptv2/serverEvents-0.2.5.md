## serverEvents: 0.2.5 - 2026-11-05
### :bug: Bug Fixes
- event streams type-check for consumers whose TypeScript and Node.js type definitions declare an async-disposable ReadableStream iterator, including when several generated SDKs are used together, and the fallback iterator supports async disposal *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
