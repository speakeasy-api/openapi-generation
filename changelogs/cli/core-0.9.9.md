## core: 0.9.9 - 2026-10-19
### :bug: Bug Fixes
- list the flags each command registers in --usage: the whole-body flag of body-only operations under its registered name, --no-descriptions on completion shells, and flags added to generated commands at runtime; --help and --usage mark a whole-body flag [required] only when the body is neither optional nor nullable *(commit by [@2ynn](https://github.com/2ynn))*
