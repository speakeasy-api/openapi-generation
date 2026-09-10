## nameOverrides: 2.81.6 - 2026-02-27
### :bug: Bug Fixes
- Add nil guard for nested `x-speakeasy-match` paths so intermediate struct pointers are checked before access, preventing nil pointer dereferences in generated Terraform SDK code *(commit by [@bradcypert](https://github.com/bradcypert))*
