# 2-edit-with-new-section

## Notes
- the README input already contains some section placeholders. Their content is expected to get updated.
- the section ordering differs from the default weights (see `sections.ts`) and is expected to be preserved.
- one of the endpoints uses pagination so a `pagination` section is expected to be inserted.
- no endpoint makes use of `file-upload`, this section is expected to be removed as it is not "Needed".
- the `retries` and `server` sections are marked as "No Action" so are expected to be skipped.
- `$toc-max-depth` is set to 3 so headers up to `###` are expected to be collected.

## Level 2
### Level 3
#### Level 4
## Level 2
### Level 3

<!-- Start SDK Installation [installation] -->
## SDK Installation
<!-- End SDK Installation [installation] -->
Any text outside section boundaries should be left untouched.
<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=3 -->
* [2-edit-with-new-section](#2-edit-with-new-section)
  * [Notes](#notes)
  * [Level 2](#level-2)
    * [Level 3](#level-3)
  * [Level 2](#level-2-1)
    * [Level 3](#level-3-1)
  * [SDK Installation](#sdk-installation)
  * [SDK Example Usage](#sdk-example-usage)
  * [Custom Error Handling](#custom-error-handling)
    * [Sub-Header](#sub-header)
  * [Pagination](#pagination)
  * [Custom HTTP Client](#custom-http-client)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Footer](#footer)

<!-- End Table of Contents [toc] -->
<!-- Start Summary [summary] -->
## Summary
<!-- End Summary [summary] -->


Empty lines between section boundaries should be preserved.


<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage
<!-- End SDK Example Usage [usage] -->

<!-- No Error Handling [errors] -->
## Custom Error Handling
This section should be left as is.

### Sub-Header
Markdown Headers should be collected whether they are within boundaries or not.
<!-- Block boundary with no effect [errors] -->

<!-- Start Pagination [pagination] -->
## Pagination
<!-- End Pagination [pagination] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client
<!-- End Custom HTTP Client [http-client] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations
<!-- End Available Resources and Operations [operations] -->

<!-- No Retries [retries] -->
<!-- Placeholder for Future Speakeasy SDK Sections -->
<!-- No Server Selection [server] -->

## Footer
This section should remain at the bottom of the file.
