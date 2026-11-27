## additionalProperties: 0.0.3 - 2026-11-25
### :bug: Bug Fixes
- spread additionalProperties at the top level of the JSON object on read and write behind the spreadAdditionalProperties fix flag, on by default for new SDKs. With the flag enabled, strict schema validation now rejects responses containing top-level additional-property values that cannot be converted to the declared value type; these values were previously ignored. *(commit by [@2ynn](https://github.com/2ynn))*
