## cliCommands: 0.3.0 - 2026-10-22
### :bee: New Features
- declare terminal failure states with output.outcome so a command exits 1 when a successful response reports a declared failure state, for discrete JSON responses (statePointer) and streamed events (streamStatePointer), with the failure detail surfaced from errorPointer / streamErrorPointer and reported as CLI_OUTCOME_FAILED *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
