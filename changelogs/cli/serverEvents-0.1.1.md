## serverEvents: 0.1.1 - 2026-10-06
### :bug: Bug Fixes
- --jq on a command that streams by default now requests one complete response and filters it; pass --stream or -o json to keep filtering each streamed event *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
