## core: 3.13.43 - 2026-05-25
### :wrench: Chores
- bump staticcheck pin from v0.6.1 to v0.7.0 for Go 1.26 support; exclude SA5008 (mirrors mockserver) since unexported fields with json tags are an intentional generator pattern *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
