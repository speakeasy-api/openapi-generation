## serverEvents: 0.3.0 - 2026-09-08
### :bee: New Features
- implement IAsyncEnumerable and IAsyncDisposable on EventStream for await foreach iteration; disposing an EventStream now closes the underlying response stream *(commit by [@2ynn](https://github.com/2ynn))*
