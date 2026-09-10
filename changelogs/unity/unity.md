## core: 0.7.3 - 2023-10-20
### :bug: Bug Fixes
- retain configuration passed in on CLI for future gen runs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## downloadStreams: 0.0.2 - 2023-10-20
### :bug: Bug Fixes
- detection of download streams *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.7.2 - 2023-10-19
### :bug: Bug Fixes
- ensure complex allOfs are handle correctly with circular reference tracking *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.7.1 - 2023-10-19
### :bug: Bug Fixes
- handling of circular references in allOf using inline schemas *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## docs: 0.0.3 - 2023-10-19
### :bug: Bug Fixes
- replace generic faker strings with known valus *(commit by [@alexrozanski](https://github.com/alexrozanski))*


## core: 0.7.0 - 2023-10-18
### :bee: New Features
- adjust default max method params *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 0.5.5 - 2023-10-18
### :bug: Bug Fixes
- additionalProperties not currently supported for parameter serialization so fall back to maps *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## pagination: 0.1.1 - 2023-10-17
### :bug: Bug Fixes
- allow cursor field to be any type for pagination *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 0.0.3 - 2023-10-17
### :wrench: Chores
- remove multi word generated examples *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## docs: 0.0.2 - 2023-10-10
### :wrench: Chores
- refactor templateTypeMarkdown using object destructuring *(commit by [@2ynn](https://github.com/2ynn))*


## docs: 0.0.1 - 2023-10-10
### :bug: Bug Fixes
- prevent usage snippet generation from panicking when an example operation cannot be found *(commit by [@zostay](https://github.com/zostay))*


## docs: 0.0.0 - 2023-10-05
### :bee: New Features
- add a new global parameters section to the README *(commit by [@zostay](https://github.com/zostay))*


## core: 0.5.4 - 2023-10-06
### :bug: Bug Fixes
- better handling of option fields when generating usage snippets *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.5.3 - 2023-10-05
### :bug: Bug Fixes
- unreachable code when no error status codes are defined in OpenAPI def *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.5.2 - 2023-10-05
### :bug: Bug Fixes
- don't generate sub types for oneOf/anyOf if unions not support in language *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## groups: 0.0.2 - 2023-10-02
### :bug: Bug Fixes
- ensure x-speakeasy-groups and tags behave exactly the same in terms of operation namespacing *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.5.1 - 2023-10-01
### :bug: Bug Fixes
- fix handling of allOf circular references when arrays and non-referenced schemas are involved *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.5.0 - 2023-09-29
### :bee: New Features
- collapse primitive oneOf single sub schema type and fix handling of null and any types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.4.1 - 2023-09-29
### :bug: Bug Fixes
- Enum class name conflict with System *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.4.0 - 2023-09-26
### :bee: New Features
- added sdk package name to user-agent string in http requests *(commit by [@disintegrator](https://github.com/disintegrator))*


## core: 0.3.3 - 2023-09-26
### :bug: Bug Fixes
- field comments for CSharp that include markdown links *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.3.2 - 2023-09-26
### :bug: Bug Fixes
- titles for SDKs in Readmes *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.3.1 - 2023-09-22
### :bug: Bug Fixes
- enum deserialization *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.3.0 - 2023-09-21
### :bee: New Features
- template comments for concrete classes *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 0.2.0 - 2023-09-20
### :bee: New Features
- add descriptions for http response additions *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## methodSecurity: 0.1.0 - 2023-09-16
### :bee: New Features
- hoist most used security to the global level *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.1.3 - 2023-09-18
### :wrench: Chores
- add support for remote URLs *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 0.1.2 - 2023-09-15
### :bug: Bug Fixes
- handling of markdown comments in csharp and unity *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
- fix classNames incorrectly cased in ReadMe *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 0.1.1 - 2023-09-15
### :bug: Bug Fixes
- fixed comments on SDKs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## multiLevelTagging: 0.1.0 - 2023-09-15
### :bee: New Features
- add support for multi level tagging *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.1.0 - 2023-09-07
### :bee: New Features
- add Unity Serializable fields to models *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 0.0.2 - 2023-09-05
### :memo: Documentation Changes
- avoid duplicate examples in array and maps *(commit by [@idbentley](https://github.com/idbentley))*


## core: 0.0.4 - 2023-09-04
### :bug: Bug Fixes
- support for long and double types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 0.0.2 - 2023-08-31
### :bug: Bug Fixes
- allow schemas that use enums with base types speakeasy doesn't support to still generate fields based on the base type *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
