## core: 2.93.3 - 2023-10-20
### :bug: Bug Fixes
- retain configuration passed in on CLI for future gen runs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## downloadStreams: 0.1.1 - 2023-10-20
### :bug: Bug Fixes
- detection of download streams *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.93.2 - 2023-10-19
### :bug: Bug Fixes
- ensure complex allOfs are handle correctly with circular reference tracking *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.93.1 - 2023-10-19
### :bug: Bug Fixes
- handling of circular references in allOf using inline schemas *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## docs: 0.3.4 - 2023-10-19
### :bug: Bug Fixes
- replace generic faker strings with known valus *(commit by [@alexrozanski](https://github.com/alexrozanski))*


## core: 2.93.0 - 2023-10-18
### :bee: New Features
- adjust default max method params *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*
  

## core: 2.91.6 - 2023-10-18
### :bug: Bug Fixes
- additionalProperties not currently supported for parameter serialization so fall back to maps *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## unions: 2.84.0 - 2023-10-18
### :bee: New Features
- support complex types in unions *(commit by [@idbentley](https://github.com/idbentley))*
## errors: 2.81.7 - 2023-10-18
### :bug: Bug Fixes
- avoid generating utility/shared classes used by custom errors as errors themselves *(commit by [@disintegrator](https://github.com/disintegrator))*


## pagination: 2.81.2 - 2023-10-17
### :bug: Bug Fixes
- allow cursor field to be any type for pagination *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## examples: 2.81.3 - 2023-10-17
### :wrench: Chores
- remove multi word generated examples *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## errors: 2.81.6 - 2023-10-16
### :bug: Bug Fixes
- unions converted to errors no longer pollute the type register and generate unique types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.5 - 2023-10-13
### :bug: Bug Fixes
- catch literal nil derefs when mashaling models to JSON *(commit by [@disintegrator](https://github.com/disintegrator))*


## devContainers: 2.90.0 - 2023-10-10
### :bee: New Features
- make devContainers readme instructions more clear *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## docs: 0.3.3 - 2023-10-10
### :wrench: Chores
- refactor templateTypeMarkdown using object destructuring *(commit by [@2ynn](https://github.com/2ynn))*


## docs: 0.3.2 - 2023-10-10
### :bug: Bug Fixes
- prevent usage snippet generation from panicking when an example operation cannot be found *(commit by [@zostay](https://github.com/zostay))*


## errors: 2.81.5 - 2023-10-10
### :bug: Bug Fixes
- handling of sub types within errors *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## docs: 0.3.1 - 2023-10-09
### :memo: Documentation Changes
- fix readme generation for globalSecurityFlattening *(commit by [@idbentley](https://github.com/idbentley))*


## additionalProperties: 0.1.1 - 2023-10-09
### :bug: Bug Fixes
- handle additionalProperties: false as if it is a NOOP *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## docs: 0.3.0 - 2023-10-05
### :bee: New Features
- add a new global parameters section to the README *(commit by [@zostay](https://github.com/zostay))*


## globalSecurity: 2.82.0 - 2023-09-20
### :bee: New Features
- global security flattening in go *(commit by [@idbentley](https://github.com/idbentley))*


## core: 2.91.4 - 2023-10-06
### :bug: Bug Fixes
- better handling of option fields when generating usage snippets *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.3 - 2023-10-05
### :bug: Bug Fixes
- don't generate sub types for oneOf/anyOf if unions not support in language *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.2 // globalSecurity: 2.82.2 - 2023-10-04
### :bug: Bug Fixes
- invalid withSecurity generation *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## constsAndDefaults: 0.1.1 // core: 2.91.2 - 2023-10-04
### :bug: Bug Fixes
- ensure single value enums don't generate as consts *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## globalSecurity: 2.82.1 // methodSecurity: 2.82.1 - 2023-10-02
### :bug: Bug Fixes
- fixed query params being added twice to url if security is used *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## groups: 2.81.2 - 2023-10-02
### :bug: Bug Fixes
- ensure x-speakeasy-groups and tags behave exactly the same in terms of operation namespacing *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.1 - 2023-10-01
### :bug: Bug Fixes
- fix handling of allOf circular references when arrays and non-referenced schemas are involved *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## unions: 2.83.1 - 2023-09-30
### :bug: Bug Fixes
- consts and defaults when dealing with unions *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## errors: 2.81.4 - 2023-09-30
### :bug: Bug Fixes
- consts and defaults when dealing with error models *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.91.0 - 2023-09-29
### :bee: New Features
- collapse primitive oneOf single sub schema type and fix handling of null and any types *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## downloadStreams: 0.1.0 - 2023-09-29
### :bee: New Features
- enable go download streaming *(commit by [@disintegrator](https://github.com/disintegrator))*


## globalSecurity: 2.82.0 - 2023-09-28
### :bee: New Features
- oAuth support through function callbacks *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## core: 2.90.0 - 2023-09-26
### :bee: New Features
- added sdk package name to user-agent string in http requests *(commit by [@disintegrator](https://github.com/disintegrator))*


## core: 2.89.3 - 2023-09-26
### :bug: Bug Fixes
- deserialization of zero-valued date and time pointers in Go *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.89.2 - 2023-09-26
### :bug: Bug Fixes
- nullable serialization *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
- titles for SDKs in Readmes *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.89.1 - 2023-09-25
### :bug: Bug Fixes
- nullable request body serialization *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## devContainers: 2.89.0 - 2023-09-22
### :bee: New Features
- clarify dev containers ReadMe *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.89.0 - 2023-09-20
### :bee: New Features
- add descriptions for http response additions *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## docs: 0.2.1 - 2023-09-21
### :bug: Bug Fixes
- correct typo in readme generation *(commit by [@zostay](https://github.com/zostay))*


## methodSecurity: 2.82.0 - 2023-09-16
### :bee: New Features
- hoist most used security to the global level *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.88.4 - 2023-09-21
### :bug: Bug Fixes
- fixes for regressions when dealing with consts and defaults *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## docs: 0.2.0 - 2023-09-20
### :bee: New Features
- more flexibility for README files and improvements to content *(commit by [@zostay](https://github.com/zostay))*


## core: 2.88.3 - 2023-09-18
### :wrench: Chores
- add support for remote URLs *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.88.2 - 2023-09-15
### :bug: Bug Fixes
- fix classNames incorrectly cased in ReadMe *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## devContainers: 2.89.0 - 2023-09-14
### :bee: New Features
- readme badges for dev containers *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## unions: 2.83.0 - 2023-09-15
### :bee: New Features
- add support for nullable unions *(commit by [@2ynn](https://github.com/2ynn))*


## multiLevelTagging: 2.87.3 - 2023-09-14
### :bug: Bug Fixes
- handle special characters for readme examples *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.88.1 - 2023-09-13
### :bug: Bug Fixes
- fixed handling of import paths when major version bumped *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## additionalProperties: 0.1.0 - 2023-09-13
### :bee: New Features
- add support for additional properties *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## devContainers: 2.89.0 - 2023-09-12
### :bee: New Features
- out of the box dev containers for go this provides an easy to use sandbox for your SDKs *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## constsAndDefaults: 0.1.0 - 2023-09-08
### :bee: New Features
- add support for json-schema defaults and consts *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## retries: 2.82.1 - 2023-09-07
### :bug: Bug Fixes
- sdk class global retry config *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## multiLevelTagging: 2.87.2 - 2023-09-07
### :bug: Bug Fixes
- multi level tagging now supports symbols in tag names *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## bigint: 0.0.2 - 2023-09-06
### :bug: Bug Fixes
- serialization of bigints to path and query params *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## decimal: 0.1.0 - 2023-09-06
### :bee: New Features
- add support for decimals *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## multiLevelTagging: 2.87.1 - 2023-09-06
### :bug: Bug Fixes
- improve ReadMe formatting nested groups *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.88.0 - 2023-09-06
### :bee: New Features
- add support for nullable+required object properties *(commit by [@2ynn](https://github.com/2ynn))*


## docs: 0.1.0 - 2023-08-31
### :bee: New Features
- adding the x-speakeasy-docs extension and per-SDK documentation customization to Go, Python, and TypeScript *(commit by [@zostay](https://github.com/zostay))*


## retries: 2.82.0 - 2023-09-05
### :bee: New Features
- add global retry configuration *(commit by [@idbentley](https://github.com/idbentley))*


## examples: 2.81.2 - 2023-09-05
### :memo: Documentation Changes
- avoid duplicate examples in array and maps *(commit by [@idbentley](https://github.com/idbentley))*


## unions: 2.82.0 - 2023-09-04
### :bee: New Features
- create input and output unions when appropriate *(commit by [@idbentley](https://github.com/idbentley))*


## core: 2.86.4 - 2023-09-04
### :bug: Bug Fixes
- fixes for passing null as an optional request body *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.86.2 - 2023-08-31
### :bug: Bug Fixes
- allow schemas that use enums with base types speakeasy doesn't support to still generate fields based on the base type *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.86.3 - 2023-09-01
### :bug: Bug Fixes
- fixed bug in readme badge formatting *(commit by [@chase-crumbaugh](https://github.com/chase-crumbaugh))*


## multiLevelTagging: 2.87.0 - 2023-08-31
### :bee: New Features
- multi-level tagging *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.86.2 - 2023-09-01
### :bug: Bug Fixes
- allow a non-string example to be considered a string for string schemas *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## methodSecurity: 2.81.2 - 2023-08-31
### :bug: Bug Fixes
- fix for missing imports in usage snippets when using operation level security *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.86.1 - 2023-08-30
### :bug: Bug Fixes
- undefined accept header options in go methods *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.86.0 - 2023-08-29
### :bee: New Features
- speakeasy readme badges *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## core: 2.85.0 - 2023-08-29
### :bee: New Features
- adds .gitattributes file to generated SDKs *(commit by [@ryan-timothy-albert](https://github.com/ryan-timothy-albert))*


## errors: 2.81.3 - 2023-08-29
### :bug: Bug Fixes
- fixed handling of errors for schemas that can't be converted to a error class *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.84.2 - 2023-08-28
### :bug: Bug Fixes
- accept headers in go *(commit by [@anuraagnalluri](https://github.com/anuraagnalluri))*


## core: 2.84.1 - 2023-08-25
### :bug: Bug Fixes
- correct problems with genignore *(commit by [@zostay](https://github.com/zostay))*


## core: 2.84.0 - 2023-08-24
### :bee: New Features
- bring back .genignore for all languages for monkey patch support *(commit by [@zostay](https://github.com/zostay))*


## core: 2.83.1 - 2023-08-23
### :bug: Bug Fixes
- edge case which could cause conflicted names under some circumstances *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*


## errors: 2.81.2 - 2023-08-22
### :bug: Bug Fixes
- handle shared error types that are used directly as responses better *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.83.0 - 2023-08-14
### :bee: New Features
- selectable accept headers in go *(commit by [@idbentley](https://github.com/idbentley))*


## globalServerURLs: 2.82.0 // methodServerURLs: 2.82.0 - 2023-08-09
### :bee: New Features
- allow parameterised protocols in server URLs *(commit by [@alexrozanski](https://github.com/alexrozanski))*


## unions: 2.81.2 - 2023-08-07
### :bug: Bug Fixes
- fixes variable naming in union marshalling *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*


## core: 2.82.0 - 2023-08-07
### :bee: New Features
- implement granular versioning to aid in generating relevant changelogs *(commit by [@TristanSpeakeasy](https://github.com/TristanSpeakeasy))*
