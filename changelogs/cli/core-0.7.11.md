## core: 0.7.11 - 2026-10-06
### :bug: Bug Fixes
- stop a locked OS keychain from hanging commands; keychain calls now time out after 2s and fall back to the config file, flag and env credentials skip the keychain, and --no-keyring turns it off. whoami --dry-run no longer reads the keychain, like every other command under --dry-run *(commit by [@2ynn](https://github.com/2ynn))*
