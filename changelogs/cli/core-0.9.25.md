## core: 0.9.25 - 2026-11-02
### :bug: Bug Fixes
- reject --server-url values without an http or https scheme and host, and non-positive --timeout durations, as usage errors. This also applies to --dry-run, which previously printed a preview for --timeout 0 *(commit by [@TristanSpeakEasy](https://github.com/TristanSpeakEasy))*
