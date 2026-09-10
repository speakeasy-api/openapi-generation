## docs: 0.4.1 - 2026-05-15
### :bug: Bug Fixes
- mintlify mode now escapes `<` whenever the following run isn't a valid HTML/JSX tag name -- catches `Record<string,...>`, `array<string,...>`, bare `<https://...>` autolinks, and similar MDX parse hazards across every target's generated docs *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
