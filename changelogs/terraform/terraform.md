## core: 2.92.5 - 2023-10-20
### :bug: Bug Fixes
- retain configuration passed in on CLI for future gen runs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.92.4 - 2023-10-20
### :wrench: Chores
- remove special x-speakeasy-entity scoping rules to get ready for import scope changes *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.92.3 - 2023-10-19
### :bug: Bug Fixes
- ensure complex allOfs are handle correctly with circular reference tracking *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.92.2 - 2023-10-19
### :bug: Bug Fixes
- handling of circular references in allOf using inline schemas *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.92.2 - 2023-10-18
### :bug: Bug Fixes
- additionalProperties not currently supported for parameter serialization so fall back to maps *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*

## examples: 2.81.5 - 2023-10-17
### :wrench: Chores
- remove multi word generated examples *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## additionalProperties: 0.1.1 - 2023-10-09
### :bug: Bug Fixes
- handle additionalProperties: false as if it is a NOOP *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.92.1 - 2023-10-05
### :bug: Bug Fixes
- don't generate sub types for oneOf/anyOf if unions not support in language *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 2.81.4 - 2023-10-02
### :bug: Bug Fixes
- state names with underscores get trimmed *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## groups: 2.81.2 - 2023-10-02
### :bug: Bug Fixes
- ensure x-speakeasy-groups and tags behave exactly the same in terms of operation namespacing *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.3 - 2023-10-01
### :bug: Bug Fixes
- fix handling of allOf circular references when arrays and non-referenced schemas are involved *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.2 - 2023-09-29
### :bug: Bug Fixes
- some edge cases with request/response bodies not defined as explicit components *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.91.1 - 2023-09-29
### :bug: Bug Fixes
- edge case when generating type definition mappings for nullable, yet required properties  *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*

## core: 2.92.0 - 2023-09-29
### :bee: New Features
- reduce the need for x-speakeasy-entity on create/update request bodies with more inference logic *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.91.0 - 2023-09-29
### :bee: New Features
- collapse primitive oneOf single sub schema type and fix handling of null and any types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.90.0 - 2023-09-26
### :bee: New Features
- added sdk package name to user-agent string in http requests *(commit by [@disintegrator](https://github.com/disintegrator))*


## core: 2.89.1 - 2023-09-26
### :bug: Bug Fixes
- titles for SDKs in Readmes *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.0 - 2023-09-20
### :bee: New Features
- add descriptions for http response additions *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## methodSecurity: 2.82.0 - 2023-09-16
### :bee: New Features
- hoist most used security to the global level *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.88.2 - 2023-09-18
### :wrench: Chores
- add support for remote URLs *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.88.1 - 2023-09-15
### :bug: Bug Fixes
- fix classNames incorrectly cased in ReadMe *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.88.0 - 2023-09-06
### :bee: New Features
- add support for nullable+required object properties *(commit by [@2ynn](https://github.com/2ynn))*


## examples: 2.81.3 - 2023-09-05
### :memo: Documentation Changes
- avoid duplicate examples in array and maps *(commit by [@idbentley](https://github.com/idbentley))*


## core: 2.86.3 - 2023-09-04
### :bug: Bug Fixes
- fixes for passing null as an optional request body *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.86.1 - 2023-08-31
### :bug: Bug Fixes
- allow schemas that use enums with base types speakeasy doesn't support to still generate fields based on the base type *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.86.2 - 2023-09-01
### :bug: Bug Fixes
- fixed bug in readme badge formatting *(commit by [@chase-crumbaugh](https://github.com/chase-crumbaugh))*


## core: 2.86.1 - 2023-09-01
### :bug: Bug Fixes
- allow a non-string example to be considered a string for string schemas *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.86.0 - 2023-08-29
### :bee: New Features
- speakeasy readme badges *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.85.0 - 2023-08-29
### :bee: New Features
- adds .gitattributes file to generated SDKs *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.84.1 - 2023-08-25
### :bug: Bug Fixes
- correct problems with genignore *(commit by [@zostay](https://github.com/zostay))*


## core: 2.84.0 - 2023-08-24
### :bee: New Features
- bring back .genignore for all languages for monkey patch support *(commit by [@zostay](https://github.com/zostay))*


## core: 2.83.4 - 2023-08-23
### :bug: Bug Fixes
- edge case which could cause conflicted names under some circumstances *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## examples: 2.81.2 - 2023-08-23
### :bug: Bug Fixes
- apply terraform formatting to generated examples without a second step being necessary *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.83.3 - 2023-08-23
### :bug: Bug Fixes
- edge case generating an incorrect identifier lookup for resource get without hoisting *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.83.2 - 2023-08-18
### :wrench: Chores
- upgrade to terraform-plugin-framework 1.3.5 _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## unions: 2.81.5 - 2023-08-10
### :bug: Bug Fixes
- improve generation of invocations from arrays of union types _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_
- support for mandatory hoisted entities _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## unions: 2.81.5 - 2023-08-10
### :bug: Bug Fixes
- improve generation of invocations from arrays of union types _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## core: 2.83.1 - 2023-08-10
### :wrench: Chores
- improve error logs to also include request body types _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## globalServerURLs: 2.82.0 // methodServerURLs: 2.82.0 - 2023-08-09
### :bee: New Features
- allow parameterised protocols in server URLs _(commit by [@alexrozanski](https://github.com/alexrozanski))_


## unions: 2.81.4 - 2023-08-08
### :bug: Bug Fixes
- improve union handling when request / response optionality is variant _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## unions: 2.81.2 - 2023-08-07
### :bug: Bug Fixes
- fixes variable naming in union marshalling _(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))_


## unions: 2.81.3 - 2023-08-08
### :bug: Bug Fixes
- improve union handling when merging multiple subtypes with differing oneOf type ordering _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## core: 2.83.0 - 2023-08-08
### :bee: New Features
- improve documentation for enums _(commit by [@ThomasRooney](https://github.com/ThomasRooney))_


## core: 2.82.0 - 2023-08-07
### :bee: New Features
- implement granular versioning to aid in generating relevant changelogs _(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))_
