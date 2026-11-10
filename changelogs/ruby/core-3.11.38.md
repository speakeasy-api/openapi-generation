## core: 3.11.38 - 2026-11-09
### :bug: Bug Fixes
- qualify Ruby core constants so tags such as File, Time or Class no longer break type checking, and add an SDK suffix to sub-SDKs named Utils, SDKHooks or SDKConfiguration (accessor sdk_configuration becomes sdk_configuration_) so SDKs with those tags load *(commit by [@2ynn](https://github.com/2ynn))*
