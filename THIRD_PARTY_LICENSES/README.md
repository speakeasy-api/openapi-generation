# Third-party license texts

This directory bundles the standard license texts identified by the top-level
artifact inventory in the repository-root [`NOTICE`](../NOTICE).

| File                                         | Applies to known top-level material                                                             | Authoritative text source                                                  |
| -------------------------------------------- | ----------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| [`Apache-2.0.txt`](Apache-2.0.txt)           | `dprint-plugin-java.wasm`, the Apache option for Mago, and the Apache portion of LLVM's license | <https://www.apache.org/licenses/LICENSE-2.0.txt>                          |
| [`LLVM-22.1.0-rc3.txt`](LLVM-22.1.0-rc3.txt) | `clangfmt.wasm`, built from LLVM 22.1.0-rc3                                                     | <https://github.com/llvm/llvm-project/blob/llvmorg-22.1.0-rc3/LICENSE.TXT> |
| [`MIT.txt`](MIT.txt)                         | Rubyfmt and the MIT-licensed dprint plugin components identified in `NOTICE`                    | <https://opensource.org/license/mit>                                       |
| [`MPL-2.0.txt`](MPL-2.0.txt)                 | The copied Terraform formatter subset in `internal/format/tfFormat.go`                          | <https://www.mozilla.org/MPL/2.0/>                                         |

These are unmodified standard license texts, not a statement that every
copyright or transitive dependency notice embedded in the formatter artifacts
has been collected. The component inventory, exact artifact hashes, source
revisions, and remaining closure requirements are recorded in `NOTICE`.
