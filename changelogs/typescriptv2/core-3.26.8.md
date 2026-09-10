## core: 3.26.8 - 2025-11-18
### :bee: New Features
- Support for lax mode deserialization. Configurable via gen.yaml laxMode: lax | strict. Missing required fields will not throw zod response validation errors but instead fallback to a zero value. eg for a string the zero value is "". Lax mode also introduces non-lossy coercion where possible eg a boolean field will tolerate the string "true". *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*
