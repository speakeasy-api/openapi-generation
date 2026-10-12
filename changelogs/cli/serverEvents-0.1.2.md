## serverEvents: 0.1.2 - 2026-10-09
### :bug: Bug Fixes
- an output_format saved in the CLI config file (for example by configure) is now treated as a default rather than an explicit choice, so it no longer disables the stream projection or the complete-response selection for --jq on default-streaming commands; pass --output-format or set the output format env var for per-event frames *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
