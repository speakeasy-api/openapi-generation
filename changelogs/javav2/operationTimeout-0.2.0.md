## operationTimeout: 0.2.0 - 2026-10-30
### :bee: New Features
- add an SDK-wide and per-call timeout that bounds the whole call including the response body and streams, on both HTTP clients, and honour x-speakeasy-timeout behind the enforceOperationTimeout option, which is on for new SDKs and off for existing ones *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
