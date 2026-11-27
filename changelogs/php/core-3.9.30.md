## core: 3.9.30 - 2026-11-18
### :bug: Bug Fixes
- send enum, date, date-time, bigint and decimal header parameters instead of dropping them, and keep headers whose value serializes to "0" *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
