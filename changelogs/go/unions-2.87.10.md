## unions: 2.87.10 - 2026-09-01
### :bug: Bug Fixes
- union values reset any previously decoded member before re-decoding, so a reused value cannot marshal a stale variant *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
