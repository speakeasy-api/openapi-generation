## cliCommands: 0.1.13 - 2026-10-07
### :bug: Bug Fixes
- An intent command now merges its presets into a request body piped on stdin, as it does for --body, so a piped partial body keeps the command's pinned model and request variant, and a piped body naming another variant is a validation error instead of being sent *(commit by [@2ynn](https://github.com/2ynn))*
