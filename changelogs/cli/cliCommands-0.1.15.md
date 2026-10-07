## cliCommands: 0.1.15 - 2026-10-06
### :bee: New Features
- An intent command flag can bind one level beneath an object preset (to $.format.width under a $.format preset); the field is resolved and typed in the union member the preset selects, written into the preset object, and merged into a partial --body or --body-param object, which is refused when it selects another member *(commit by [@2ynn](https://github.com/2ynn))*
