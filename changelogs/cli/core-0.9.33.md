## core: 0.9.33 - 2026-11-16
### :bug: Bug Fixes
- the generated install.sh now quotes the temporary download directory in its EXIT trap, so cleanup removes only that directory and never runs path text as shell code *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
