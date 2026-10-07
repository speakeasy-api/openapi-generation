## cliCommands: 0.1.14 - 2026-10-06
### :bug: Bug Fixes
- An intent command's object preset now fills the missing keys of a partial object passed for the same key in --body or --body-param, so a nested discriminator such as a pinned union variant's type is no longer dropped; an object naming another variant is still sent as given *(commit by [@2ynn](https://github.com/2ynn))*
