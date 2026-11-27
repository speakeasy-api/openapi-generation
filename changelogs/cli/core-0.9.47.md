## core: 0.9.47 - 2026-11-25
### :bug: Bug Fixes
- --jq halt and halt_error keep earlier output and follow jq exit semantics, so halt and halt_error(0) succeed and other halt_error calls fail with the halted value *(commit by [@TristanSpeakEasy](https://github.com/TristanSpeakEasy))*
