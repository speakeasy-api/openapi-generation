## core: 2.92.0 - 2023-10-20
### :bee: New Features
- add RepositoryUrl to csproj file if present *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.3 - 2023-10-20
### :bug: Bug Fixes
- retain configuration passed in on CLI for future gen runs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.2 - 2023-10-19
### :bug: Bug Fixes
- ensure complex allOfs are handle correctly with circular reference tracking *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.1 - 2023-10-19
### :bug: Bug Fixes
- handling of circular references in allOf using inline schemas *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.5 - 2023-10-18
### :bug: Bug Fixes
- additionalProperties not currently supported for parameter serialization so fall back to maps *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## docs: 0.0.3 - 2023-10-19
### :bug: Bug Fixes
- replace generic faker strings with known valus *(commit by [@alexrozanski](https://github.com/alexrozanski))*


## core: 2.91.0 - 2023-10-18
### :bee: New Features
- adjust default max method params *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## pagination: 0.1.1 - 2023-10-17
### :bug: Bug Fixes
- allow cursor field to be any type for pagination *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 2.81.3 - 2023-10-17
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


## core: 2.89.4 - 2023-10-06
### :bug: Bug Fixes
- better handling of option fields when generating usage snippets *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.3 - 2023-10-05
### :bug: Bug Fixes
- unreachable code when no error status codes are defined in OpenAPI def *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.2 - 2023-10-05
### :bug: Bug Fixes
- don't generate sub types for oneOf/anyOf if unions not support in language *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## groups: 2.81.2 - 2023-10-02
### :bug: Bug Fixes
- ensure x-speakeasy-groups and tags behave exactly the same in terms of operation namespacing *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.1 - 2023-10-01
### :bug: Bug Fixes
- fix handling of allOf circular references when arrays and non-referenced schemas are involved *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.0 - 2023-09-29
### :bee: New Features
- collapse primitive oneOf single sub schema type and fix handling of null and any types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.88.0 - 2023-09-26
### :bee: New Features
- added sdk package name to user-agent string in http requests *(commit by [@disintegrator](https://github.com/disintegrator))*


## core: 2.87.3 - 2023-09-26
### :bug: Bug Fixes
- field comments for CSharp that include markdown links *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.87.2 - 2023-09-26
### :bug: Bug Fixes
- titles for SDKs in Readmes *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.87.1 - 2023-09-22
### :bug: Bug Fixes
- enum deserialization *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.87.0 - 2023-09-21
### :bee: New Features
- template comments for concrete classes *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.86.0 - 2023-09-20
### :bee: New Features
- add descriptions for http response additions *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## methodSecurity: 2.82.0 - 2023-09-16
### :bee: New Features
- hoist most used security to the global level *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.85.9 - 2023-09-18
### :wrench: Chores
- add support for remote URLs *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.85.8 - 2023-09-15
### :bug: Bug Fixes
- handling of markdown comments in csharp and unity *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
- fix classNames incorrectly cased in ReadMe *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.85.7 - 2023-09-15
### :bug: Bug Fixes
- fixed comments on SDKs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## multiLevelTagging: 2.86.1 - 2023-09-14
### :bug: Bug Fixes
- handle special characters for readme examples *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## multiLevelTagging: 2.86.0 - 2023-09-07
### :bee: New Features
- multilevel tagging *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.85.6 - 2023-09-07
### :bug: Bug Fixes
- various fixes to serialization logic and warnings cleanup *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.85.5 - 2023-09-07
### :bug: Bug Fixes
- to model serialization in non-default namespace names *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## bigint: 0.1.0 - 2023-09-06
### :bee: New Features
- add support for bigints *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 2.81.2 - 2023-09-05
### :memo: Documentation Changes
- avoid duplicate examples in array and maps *(commit by [@idbentley](https://github.com/idbentley))*


## core: 2.85.4 - 2023-09-04
### :bug: Bug Fixes
- support for long and double types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## decimal: 0.0.1  - 2023-09-04
### :bee: New Features
- added support for decimal types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.85.1 - 2023-08-31
### :bug: Bug Fixes
- allow schemas that use enums with base types speakeasy doesn't support to still generate fields based on the base type *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.85.3 - 2023-09-01
### :bug: Bug Fixes
- fixed bug in readme badge formatting *(commit by [@chase-crumbaugh](https://github.com/chase-crumbaugh))*


## core: 2.85.2 - 2023-09-01
### :bug: Bug Fixes
- allow a non-string example to be considered a string for string schemas *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.85.1 - 2023-08-30
### :recycle: Refactors
- rebuilt C# sdk to match standard Speakeasy structure *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.85.0 - 2023-08-29
### :bee: New Features
- speakeasy readme badges *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.84.0 - 2023-08-29
### :bee: New Features
- adds .gitattributes file to generated SDKs *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.83.1 - 2023-08-25
### :bug: Bug Fixes
- correct problems with genignore *(commit by [@zostay](https://github.com/zostay))*


## core: 2.83.0 - 2023-08-24
### :bee: New Features
- bring back .genignore for all languages for monkey patch support *(commit by [@zostay](https://github.com/zostay))*


## core: 2.82.1 - 2023-08-23
### :bug: Bug Fixes
- edge case which could cause conflicted names under some circumstances *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## globalServerURLs: 2.82.0 // methodServerURLs: 2.82.0 - 2023-08-09
### :bee: New Features
- allow parameterised protocols in server URLs *(commit by [@alexrozanski](https://github.com/alexrozanski))*


## core: 2.82.0 - 2023-08-07
### :bee: New Features
- implement granular versioning to aid in generating relevant changelogs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
