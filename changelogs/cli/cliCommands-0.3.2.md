## cliCommands: 0.3.2 - 2026-11-02
### :bug: Bug Fixes
- report WAV content written to a .pcm artifact --out path as an extension mismatch; raw PCM responses (audio/pcm, audio/l16) now resolve to .pcm instead of .wav, so a genuine PCM payload at .pcm and a WAV payload at .wave stay warning-free *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
