## cliCommands: 0.4.0 - 2026-10-23
### :bee: New Features
- declare a continuation recipe so a command re-sends its request with resume values from the response while output.outcome classifies the response state as continue, combining the responses (merge arrays, sum numbers) or streaming every response as one output, with --auto-continue and --max-continuations flags and CLI_CONTINUATION_LIMIT when the hop limit is reached *(commit by [@TristanSpeakEasy](https://github.com/TristanSpeakEasy))*
