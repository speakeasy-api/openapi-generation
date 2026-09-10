## core: 3.13.15 - 2026-02-19
### :bug: Bug Fixes
- defer res.Body.Close in ConsumeRawBody to prevent resource leaks on error *(commit by [@vishalg0wda](https://github.com/vishalg0wda))*
