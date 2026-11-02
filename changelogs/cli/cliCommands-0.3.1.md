## cliCommands: 0.3.1 - 2026-10-14
### :bug: Bug Fixes
- write an artifact --out path that already has a media extension exactly as given, reporting a MIME mismatch on stderr and as extension_mismatch in the result envelope instead of rewriting it; a path without one still gets the returned extension appended, so a suffix such as .v1 is kept, and MIME subtypes are normalized (image/svg+xml writes .svg, video/x-msvideo writes .avi) *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
